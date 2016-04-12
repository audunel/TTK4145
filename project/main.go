package main

import (
	"flag"
	"./driver"
	"./network"
	"./master"
	"./slave"
	"./logger"
)

const (
	masterPort	= 20001
	slavePor	= 20002

func main() {
	var startAsMaster bool
	flag.BoolVar(&startAsMaster, "master", false, "Start as master")
	flag.Parse()

	var elevatorEvents com.ElevatorEvent
	elevatorEvents.FloorReached		:= make(chan int)
	elevatorEvents.NewTargetFloor	:= make(chan int)

	var slaveEvents com.SlaveEvent
	slaveEvents.CompletedFloor	:= make(chan int)
	slaveEvents.MissedDeadline	:= make(chan bool)
	slaveEvents.ButtonPressed	:= make(chan order.OrderButton)
	slaveEvents.FromMaster		:= make(chan network.UDPMessage)
	slaveEvents.ToMaster		:= make(chan network.UDPMessage)

	var masterEvents com.MasterEvent
	masterEvents.ToSlaves	:= make(chan network.UDPMessage)
	masterEvents.FromSlaves	:= make(chan network.UDPMessage)

	driver.ElevInit()

	driver.EventListener(
				slaveEvents.ButtonPressed,
				elevatorEvents.FloorReached)

	elevator.Init(
				slaveEvents.CompletedFloor,
				slaveEvents.MissedDeadline,
				elevatorEvents.FloorReached,
				elevatorEvents.NewTargetFloor)

	if startAsMaster {
		go network.UDPInit(masterPort, slavePort, masterEvents.ToSlaves, masterEvents.FromSlaves)
		go master.InitMaster(masterEvents, nil, nil)
	}
	go network.UDPInit(slavePort, masterPort, slaveEvents.FromMaster, slaveEvents.ToMaster)
	slave.InitSlave(slaveEvents, masterEvents, elevatorEvents)
}
