package communication

import (
	"../driver"
	"../network"
	"time"
	"encoding/json"
	"log"
)

type OrderButton struct {
	Type	driver.ButtonType
	Floor	int
}

type Order struct {
	Button	driver.OrderButton
	TakenBy	network.ID
	Done	bool
}

type SlaveData struct {
	LastPasserFloor		int
	CurrentDirection	driver.MotorDirection
	Orders			[]Order
}

type MasterData struct {
	AssignedBackup	network.ID
	Orders		[]Order
	Slaves		map[network.ID]Slave
}

type Slave struct {
	ID		network.ID
	LastPassedFloor	int
	HastTimedOut	bool
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

type ElevatorEvent struct {
	FloorReached	chan int
	NewTargetFloor	chan int
	StopButton	chan bool
}

type SlaveEvent struct {
	CompletedFloor	chan int
	MissedDeadline	chan bool
	ButtonPressed	chan driver.OrderButton
	FromMaster	chan network.UDPMessage
	ToMaster	chan network.UDPMessage
}

type MasterEvent struct {
	ToSlaves 	chan network.UDPMessage
	FromSlaves	chan network.UDPMessage
}
