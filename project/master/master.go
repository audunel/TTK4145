package master

import (
	"../communication"
	"../network"
	"../driver"
	"time"
	"fmt"
)

const clientTimeout = 5 * time.Second

func WaitForBackup(events communication.MasterEvent, initialQueue []communication.Order,
					initialSlaves map[network.ID]communication.Slave) {

	myID := network.GetOwnID()
	fmt.Printf("Waiting for backup on machine %s\n", myID.String())

	queue	:= initialQueue
	slaves	:= initialSlaves

	for {
		packet := <- events.FromSlave
		_, err := communication.DecodeSlaveData
		if err != nil {
			break
		}
