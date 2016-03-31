package main

import (
    "./network"
    "./driver"
    "./communication"
    "./elevator"
    "fmt"

)


func main() {





    var elevator_event communication.ElevatorEvent
    elevator_event.FloorReached = make(chan int)
    elevator_event.NewTargetFloor = make(chan int)
    elevator_event.StopButton = make(chan bool)

    var slave_event communication.SlaveEvent
    slave_event.CompletedFloor = make(chan int)
    slave_event.MissedDeadline = make(chan bool)
    slave_event.ButtonPressed = make(chan driver.OrderButton)
    slave_event.FromMaster = make(chan network.UDPMessage)
    slave_event.ToMaster = make(chan network.UDPMessage)

    var master_event communication.MasterEvent
    master_event.ToSlaves = make(chan network.UDPMessage)
    master_event.FromSlaves = make(chan network.UDPMessage)

    go fmt.Printf("%d\n", 1) 

    driver.ElevInit()

    go driver.GetSignals(
        slave_event.ButtonPressed,
        elevator_event.FloorReached,
        elevator_event.StopButton)

    go elevator.Init(
        slave_event.CompletedFloor,
        slave_event.MissedDeadline,
        elevator_event.FloorReached,
        elevator_event.NewTargetFloor,
        elevator_event.StopButton)
    
    for{
        
        
        for f := 0; f < 4; f++{
            if(driver.GetButtonSignal(2,f) == 1){
                elevator_event.NewTargetFloor <- f
            }
        }

        



    }

}

