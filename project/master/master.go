package master

import (
	"../com"
	"../network"
	"../order"
	"../driver"
	"../delegation"
	"time"
	"log"
)

const (
	slaveTimeoutPeriod		= 5 * time.Second
	sendInterval			= 100 * time.Millisecond
	backupDeadline			= 10 * time.Second
)
var myIP = network.GetOwnIP()

func InitMaster(events				com.MasterEvent,
				initialOrders		[]order.Order,
				initialSlaves		map[network.IP]com.Slave,
				masterLogger		log.Logger,
				networkKillswitch	chan bool) {

	backupDeadlineTimer	:= time.NewTimer(backupDeadline)

	orders	:= initialOrders
	slaves	:= initialSlaves

	masterLogger.Print("Waiting for backup")
	for {
		select {
		case <- backupDeadlineTimer.C:
			masterLogger.Print("Not contacted by external slave within deadline. Terminating master.")
			networkKillswitch <- true
			return

		case message := <- events.FromSlaves:
			_, err := com.DecodeSlaveMessage(message.Data)
			if err != nil {
				break
			}

			if(message.Address != myIP) {
				orders, slaves = masterLoop(events, message.Address, orders, slaves, masterLogger)
				masterLogger.Print("Waiting for new backup")
				backupDeadlineTimer.Reset(backupDeadline)
			}
		}
	}
}

func masterLoop(events			com.MasterEvent,
				backup			network.IP,
				initialOrders	[]order.Order,
				initialSlaves	map[network.IP]com.Slave,
				masterLogger	log.Logger) ([]order.Order, map[network.IP]com.Slave) {

	sendTicker		:= time.NewTicker(sendInterval)
	slaveTimedOut	:= make(chan network.IP)

	orders := make([]order.Order, 0)
	if initialOrders != nil {
		orders = initialOrders
	}

	slaves := make(map[network.IP]com.Slave)
	if initialSlaves != nil {
		for _, slave := range(initialSlaves) {
			slave.AliveTimer = time.NewTimer(slaveTimeoutPeriod)
			slaves[slave.IP] = slave
			go listenForTimeout(slave.IP, slave.AliveTimer, slaveTimedOut)
		}
	}

	masterLogger.Printf("Initiating master with backup %s", backup)
	for {
		select {
		case message := <- events.FromSlaves:
			senderIP := message.Address
			data, err := com.DecodeSlaveMessage(message.Data)
			if err != nil {
				break
			}

			slave, exists := slaves[senderIP]
			if !exists {
				masterLogger.Printf("Adding new slave %s", senderIP)
				aliveTimer := time.NewTimer(slaveTimeoutPeriod)
				slave = com.Slave {
					IP:			senderIP,
					AliveTimer:	aliveTimer,
				}
				go listenForTimeout(slave.IP, aliveTimer, slaveTimedOut)
			}

			slave.AliveTimer.Reset(slaveTimeoutPeriod)
			slave.HasTimedOut	= false
			slave.ElevData		= data.ElevData
			slaves[senderIP]	= slave

			orders = updateOrders(data.Requests, orders, senderIP)

		case <- sendTicker.C:
			err := delegation.DelegateWork(slaves, orders)
			if err != nil {
				masterLogger.Print(err)
			}

			data := com.MasterData {
				AssignedBackup: backup,
				Orders:			orders,
				Slaves:			slaves,
			}

			events.ToSlaves <- network.UDPMessage {
				Address:	myIP,
				Data:		com.EncodeMasterData(data),
			}

		case slaveIP := <- slaveTimedOut:
			masterLogger.Printf("Slave %s timed out", slaveIP)
			slave, exists := slaves[slaveIP]
			if exists {
				slave.HasTimedOut = true
				slaves[slaveIP] = slave
				err := delegation.DelegateWork(slaves, orders)
				if err != nil {
					masterLogger.Println(err)
				}
			}
			if slaveIP == backup {
				return orders, slaves // Return current state and await new backup
			}
		}
	}
}

func listenForTimeout(ip network.IP, timer *time.Timer, timeout chan network.IP) {
	for {
		select {
		case <- timer.C:
			timeout <- ip
		}
	}
}

func updateOrders(requests, orders []order.Order, sender network.IP) []order.Order {
	orders = addNewOrders(requests, orders, sender)
	orders = removeDoneOrders(requests, orders)
	return orders
}

func addNewOrders(requests, orders []order.Order, sender network.IP) []order.Order {
	for _, request := range(requests) {
		if request.Button.Type == driver.ButtonCallCommand {
			request.TakenBy = sender
		}
		if order.OrderNew(request, orders) {
			orders = append(orders, request)
		}
	}
	return orders
}

func removeDoneOrders(requests, orders []order.Order) []order.Order {
	for i := 0; i < len(orders); i++ {
		for _, request := range(requests) {
			if order.OrdersEqual(orders[i], request) && request.Done {
				orders[i].Done = true
			}
		}
		if orders[i].Done {
			orders = append(orders[:i], orders[i+1:]...)
			i--
		}
	}
	return orders
}
