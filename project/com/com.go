package com

import (
	"../driver"
	"../network"
	"../order"
	"time"
	"encoding/json"
	"log"
)

type SlaveData struct {
	LastPassedFloor		int
	CurrentDirection	driver.MotorDirection
	Orders				[]order.Order
}

type MasterData struct {
	AssignedBackup	network.ID
	Orders			order.Order
	Slaves			map[network.ID]Slave
}

type Slave struct {
	ID				network.ID
	LastPassedFloor	int
	HastTimedOut	bool
	AliveTimer		*time.Timer
}

func EncodeMasterData(m MasterData) b []byte {
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
	err := json.Unmasrhsal(b, &result)
	return result, err
}

type ElevatorEvent struct {
	FloorReached	chan int
	NewTargetFloor	chan int
}

type SlaveEvent struct {
	CompletedFloor	chan int
	MissedDeadline	chan bool
	ButtonPressed	chan order.OrderButton
	FromMaster		chan network.UDPMessage
	ToMaster		chan network.UDPMessage
}

type MasterEvent struct {
	ToSlaves	chan network.UDPMessage
	FromSlaves	chan network.UDPMessage
}
