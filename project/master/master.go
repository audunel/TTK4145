package master

import (
	"../com"
	"../network"
	"../order"
	"../driver"
	"../delegation"
	"time"
	"log"
	"fmt"
)

const (
	slaveTimeoutPeriod		= 5 * time.Second
	sendInterval			= 100 * time.Millisecond
	selfAsBackupDeadline	= 1 * time.Second
)
var myIP = network.GetOwnIP()

func InitMaster(events			com.MasterEvent,
				initialOrders	[]order.Order,
				initialSlaves	map[network.IP]com.Slave,
				masterLogger	log.Logger) {

	selfAsBackup		:= false
	selfAsBackupTimer	:= time.NewTimer(selfAsBackupDeadline)

	orders	:= initialOrders
	slaves	:= initialSlaves

	masterLogger.Print("Waiting for backup")
	for {
		select {
		case <- selfAsBackupTimer.C:
			selfAsBackup = true

		case message := <- events.FromSlaves:
			_, err := com.DecodeSlaveMessage(message.Data)
			if err != nil {
				break
			}

			if(message.Address == myIP && selfAsBackup) || (message.Address != myIP) {
				if message.Address == myIP {
					masterLogger.Println("Using self as backup")
				}
				orders, slaves = masterLoop(events, message.Address, orders, slaves, masterLogger)
				masterLogger.Print("Waiting for new backup")
				selfAsBackup = false
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

			if backup == myIP && senderIP != myIP {
				backup = senderIP
				masterLogger.Printf("Changed backup to remote machine %s", senderIP)
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
			slave.HasTimedOut = false
			slave.LastPassedFloor = data.LastPassedFloor
			slaves[senderIP] = slave

			orders = updateOrders(data.Requests, orders, senderIP)

		case <- sendTicker.C:
			err := delegation.DelegateWork(slaves, orders)
			if err != nil {
				log.Fatal(err)
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
					log.Fatal(err)
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
			fmt.Printf("New order added, floor %d\n", request.Button.Floor)
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
			fmt.Printf("Order deleted, floor %d\n", orders[i].Button.Floor)
			orders = append(orders[:i], orders[i+1:]...)
			i--
		}
	}
	return orders
}
