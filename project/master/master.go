package master

import (
	"../com"
	"../network"
	"../driver"
	"time"
	"fmt"
)

const clientTimeout = 5 * time.Second

func WaitForBackup(events com.MasterEvent, initialQueue []com.Order,
					initialSlaves map[network.ID]com.Slave) {

	myID := network.GetOwnID()
	fmt.Printf("Waiting for backup on machine %s\n", myID.String())

	queue	:= initialQueue
	slaves	:= initialSlaves

	for {
		packet := <- events.FromSlave
		_, err := com.DecodeSlaveData
		if err != nil {
			break
		}
