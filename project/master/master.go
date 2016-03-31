package master

import (
	"../com"
	"../network"
	"../driver"
	"time"
	"fmt"
)

const slaveTimeoutPeriod = 5 * time.Second
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

		case packet := <- events.FromSlave:
			_, err := com.DecodeSlaveData(packet.Data)
			if err != nil {
				break
			}

			if(packet.Address == myID && selfAsBackup) || (packet.Address != myID) {
				if packet.Address == myID {
					fmt.Println("Using self as backup")
				}
				queue, slaves = masterLoop(events, packet.Address, queue, slaves)
			}
		}
	}
}

func masterLoop(events com.MasterEvent,
		backup network.ID,
		initialQueue []com.Order,
		initialSlaves map[network.ID]com.Slave) ([]com.Order, map[network.ID]com.Slave) {

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
			go listenForTimeout(s, slaveTimedOut)
		}
	}

	// ADD LOOP HERE
}
