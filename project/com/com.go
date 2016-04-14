package com

import (
	"../network"
	"../order"
	"../driver"
	"time"
	"encoding/json"
	"log"
)

type SlaveData struct {
	LastPassedFloor		int
	Requests			[]order.Order
}

type MasterData struct {
	AssignedBackup	network.IP
	Orders			[]order.Order
	Slaves			map[network.IP]Slave
}

type Slave struct {
	IP				network.IP
	LastPassedFloor	int
	HasTimedOut		bool
	AliveTimer		*time.Timer `json:"-"`
}

func EncodeMasterData(m MasterData) []byte {
	result, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func DecodeMasterMessage(b []byte) (MasterData, error) {
	var result MasterData
	err := json.Unmarshal(b, &result)
	return result, err
}

func EncodeSlaveData(s SlaveData) []byte {
	result, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func DecodeSlaveMessage(b []byte) (SlaveData, error) {
	var result SlaveData
	err := json.Unmarshal(b, &result)
	return result, err
}

type ElevatorEvent struct {
	FloorReached	chan int
	NewTargetFloor	chan int
}

type SlaveEvent struct {
	CompletedFloor	chan int
	MissedDeadline	chan bool
	ButtonPressed	chan driver.OrderButton
	FromMaster		chan network.UDPMessage
	ToMaster		chan network.UDPMessage
}

type MasterEvent struct {
	ToSlaves	chan network.UDPMessage
	FromSlaves	chan network.UDPMessage
}
