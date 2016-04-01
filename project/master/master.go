package master

import (
	"../com"
	"../network"
	"../driver"
	"time"
	"fmt"
	"log"
)

const slaveTimeoutPeriod = 5 * time.Second
const sendInterval = 250 * time.Millisecond
const selfAsBackupDeadline = 10 * time.Second
var myID = network.GetOwnID()

func InitMaster(events com.MasterEvent,
		initialQueue []com.Order,
		initialSlaves map[network.ID]com.Slave) {

	selfAsBackup := false

	queue	:= initialQueue
	slaves	:= initialSlaves

	for {
		fmt.Printf("Waiting for backup on machine %s\n", myID)
		select {
		case <- time.After(selfAsBackupDeadline):
			selfAsBackup = true

		case message := <- events.FromSlave:
			_, err := com.DecodeSlaveData(message.Data)
			if err != nil {
				break
			}

			if(message.Address == myID && selfAsBackup) || (message.Address != myID) {
				if message.Address == myID {
					fmt.Println("Using self as backup")
				}
				queue, slaves = masterLoop(events, message.Address, queue, slaves)
			}
		}
	}
}

func masterLoop(events com.MasterEvent,
		backup network.ID,
		initialQueue []com.Order,
		initialSlaves map[network.ID]com.Slave) ([]com.Order, map[network.ID]com.Slave) {

	sendTicker := time.NewTicker(sendInterval)
	slaveTimedOut := make(chan network.ID)

	if initialQueue == nil {
		orders := make([]com.Order, 0)
	} else {
		orders := initialQueue
	}

	slaves := make(map[network.ID]com.Slave)
	if initialSlaves != nil {
		for _, s := range(intialSlaves) {
			s.AliveTimer = time.NewTimer(slaveTimeoutPeriod)
			slaves[s.ID] = s
			go listenForTimeout(s.ID, s.AliveTimer, slaveTimedOut)
		}
	}

	fmt.Printf("Initiating master with backup %s\n", backup)
	for {
		select {
		case message := <- events.FromSlave:
			senderID := message.Address
			data, err := com.DecodeSlaveData(message.Data)
			if err != nil {
				break
			}

			if backup == myID && senderID != myID {
				backup = senderID
				fmt.Printf("Changed backup to remote unit %s", senderID)
			}

			slave, exists := slaves[senderID]
			if !exists {
				fmt.Printf("Adding new slave %s\n", senderID)
				aliveTimer := time.NewTimer(slaveTimeoutPeriod)
				slave := com.Slave {
					ID:		senderID,
					AliveTimer:	aliveTimer,
				}
				go listenForTimeout(slave, aliveTimer, slaveTimedOut)
			}

			slave.AliveTimer.Reset(slaveTimeoutPeriod)
			slave.HasTimedOut = false
			slave.LastPassedFloor = data.LastPassedFloor
			slaves[senderID] = slave

			orders = updateOrders(data.Requests, orders, senderID)

		case sendTicker.C:
			fmt.Println("Sending to slaves")
			err := delegation.DelegateWork(slaves, orders)
			if err != nil {
				log.Fatal(err)
			}

			data := com.MasterData {
				AssignedBackup: backup,
				Orders:		orders,
				Slaves:		slaves,
			}

			events.ToSlaves <- network.Message {
				Address:	myID
				Data:		com.EncodeMasterData(data)
			}

		case slaveID := <- slaveTimedOut:
			fmt.Printf("Slave %s timed out\n", slaveID)
			slave, exists := slaves[slaveID]
			if exists {
				slave.HasTimedOut = true
				slave[slaveID] = slave
				err := delegation.DelegateWork(slaves, orders)
				if err != nil {
					log.Fatal(err)
				}
			}
			if slaveID == backup {
				return orders, slaves // Return current state and await new backup
			}
		}
}

func listenForTimeout(id network.ID, timer *time.Timer, timeout chan network.ID) {
	for {
		select {
		case <- timer.C:
			timeout <- id
		}
	}
}

func updateOrders(requests, orders []com.Order) []com.Order {
	orders = addNewOrders(requests, orders)
	orders = removeDoneOrders(requests, orders)
	return orders
}

func addNewOrders(requests, orders []com.Order) []com.Order {
	for _, r := range(requests) {
		if queue.NewOrder(r, orders) {
			orders = append(orders, r)
		}
	}
	return orders
}

func removeDoneOrders(requests, orders []com.Order) []com.Order {
	for i := 0; i < len(orders); i++ {
		for _, r := range(requests) {
			if queue.SameOrder(orders[i], r) && r.Done {
				orders[i].Done = true
			}
		}
		if orders[i].Done {
			orders = append(orders[:i], orders[i+1]...)
			i--
		}
	}
	return orders
}
