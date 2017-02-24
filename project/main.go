package main

import (
	"./com"
	"./driver"
	"./elevator"
	"./logger"
	"./master"
	"./network"
	"./slave"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
)

func main() {
	mainLogger := logger.NewLogger("MAIN")

	fmt.Println("Hello world!")

	var startAsMaster, recoverBackup bool
	flag.BoolVar(&startAsMaster, "master", false, "Start as master")
	flag.BoolVar(&recoverBackup, "recover", false, "Recover backup data from disk (Must be master!)")
	flag.Parse()

	if !startAsMaster && recoverBackup {
		mainLogger.Fatal("FATAL: Can only recover as master or in single elevator mode")
	}

	var elevatorEvents com.ElevatorEvent
	elevatorEvents.FloorReached = make(chan int)
	elevatorEvents.NewTargetFloor = make(chan int)

	var slaveEvents com.SlaveEvent
	slaveEvents.CompletedFloor = make(chan int)
	slaveEvents.MissedDeadline = make(chan bool)
	slaveEvents.ButtonPressed = make(chan driver.OrderButton)
	slaveEvents.FromMaster = make(chan network.UDPMessage)
	slaveEvents.ToMaster = make(chan network.UDPMessage)

	var masterEvents com.MasterEvent
	masterEvents.ToSlaves = make(chan network.UDPMessage)
	masterEvents.FromSlaves = make(chan network.UDPMessage)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		<-c
		driver.SetMotorDirection(driver.DirnStop)
		mainLogger.Fatal("Program terminated")
	}()

	mainLogger.Print("Initializing elevator")
	driver.ElevInit()

	go driver.EventListener(
		slaveEvents.ButtonPressed,
		elevatorEvents.FloorReached)

	go elevator.Init(
		slaveEvents.CompletedFloor,
		slaveEvents.MissedDeadline,
		elevatorEvents.FloorReached,
		elevatorEvents.NewTargetFloor,
		logger.NewLogger("ELEV"))

	if startAsMaster {
		go network.UDPInit(true, masterEvents.ToSlaves, masterEvents.FromSlaves, logger.NewLogger("NETWORK"))
		masterLogger := logger.NewLogger("MASTER")
		var initialData com.MasterData
		if recoverBackup {
			file, err := os.Open("backupData.json")
			buf := make([]byte, 1024)
			n, err := file.Read(buf)
			err = json.Unmarshal(buf[:n], &initialData)
			if err != nil {
				masterLogger.Print(err)
			}
		}
		go master.InitMaster(masterEvents, initialData.Orders, initialData.Slaves, masterLogger)
	}
	go network.UDPInit(false, slaveEvents.ToMaster, slaveEvents.FromMaster, logger.NewLogger("NETWORK"))
	slave.InitSlave(slaveEvents, masterEvents, elevatorEvents, logger.NewLogger("SLAVE"))
}
