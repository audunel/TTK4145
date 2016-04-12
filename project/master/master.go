package master

import (
	"../com"
	"../network"
	"../order"
	"../driver"
	"../delegation"
	"time"
	"fmt"
	"log"
)

const slaveTimeoutPeriod = 5 * time.Second
const sendInterval = 250 * time.Millisecond
const selfAsBackupDeadline = 10 * time.Second
var myIP = network.GetOwnIP()

func InitMaster(events com.MasterEvent,
		initialQueue []order.Order,
		initialSlaves map[network.IP]com.Slave) {

	selfAsBackup := false

	queue	:= initialQueue
	slaves	:= initialSlaves

	for {
		fmt.Printf("Waiting for backup on machine %s\n", myIP)
		select {
		case <- time.After(selfAsBackupDeadline):
			selfAsBackup = true

		case message := <- events.FromSlaves:
			_, err := com.DecodeSlaveMessage(message.Data)
			if err != nil {
				break
			}

			if(message.Address == myIP && selfAsBackup) || (message.Address != myIP) {
				if message.Address == myIP {
					fmt.Println("Using self as backup")
				}
				queue, slaves = masterLoop(events, message.Address, queue, slaves)
			}
		}
	}
}

func masterLoop(events com.MasterEvent,
		backup network.IP,
		initialQueue []order.Order,
		initialSlaves map[network.IP]com.Slave) ([]order.Order, map[network.IP]com.Slave) {

	sendTicker := time.NewTicker(sendInterval)
	slaveTimedOut := make(chan network.IP)

	orders := make([]order.Order, 0)
	if initialQueue != nil {
		orders = initialQueue
	}

	slaves := make(map[network.IP]com.Slave)
	if initialSlaves != nil {
		for _, s := range(initialSlaves) {
			s.AliveTimer = time.NewTimer(slaveTimeoutPeriod)
			slaves[s.IP] = s
			go listenForTimeout(s.IP, s.AliveTimer, slaveTimedOut)
		}
	}

	fmt.Printf("Initiating master with backup %s\n", backup)
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
				fmt.Printf("Changed backup to remote unit %s", senderIP)
			}

			slave, exists := slaves[senderIP]
			if !exists {
				fmt.Printf("Adding new slave %s\n", senderIP)
				aliveTimer := time.NewTimer(slaveTimeoutPeriod)
				slave := com.Slave {
					IP:		senderIP,
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
			fmt.Println("Sending to slaves")
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
			fmt.Printf("Slave %s timed out\n", slaveIP)
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
