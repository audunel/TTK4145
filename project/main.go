package main

import (
    "./network"
    "./driver"
    "./com"
    "./elevator"
    "fmt"

)


func main() {
    var elevatorEvent com.ElevatorEvent
    elevatorEvent.FloorReached = make(chan int)
    elevatorEvent.NewTargetFloor = make(chan int)
    elevatorEvent.StopButton = make(chan bool)

    var slaveEvent com.SlaveEvent
    slaveEvent.CompletedFloor = make(chan int)
    slaveEvent.MissedDeadline = make(chan bool)
    slaveEvent.ButtonPressed = make(chan driver.OrderButton)
    slaveEvent.FromMaster = make(chan network.UDPMessage)
    slaveEvent.ToMaster = make(chan network.UDPMessage)

    var masterEvent com.MasterEvent
    masterEvent.ToSlaves = make(chan network.UDPMessage)
    masterEvent.FromSlaves = make(chan network.UDPMessage)

    go fmt.Printf("%d\n", 1) 

    driver.ElevInit()

    go driver.GetSignals(
        slaveEvent.ButtonPressed,
        elevatorEvent.FloorReached,
        elevatorEvent.StopButton)

    go elevator.Init(
        slaveEvent.CompletedFloor,
        slaveEvent.MissedDeadline,
        elevatorEvent.FloorReached,
        elevatorEvent.NewTargetFloor,
        elevatorEvent.StopButton)
 
    for {
        for f := 0; f < 4; f++{
            if(driver.GetButtonSignal(2,f) == 1){
                elevatorEvent.NewTargetFloor <- f
            }
        }
    }

}

