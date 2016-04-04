package cleitn

import (
    "../driver"
    "../network"
    "../com"
    "../master"
    "../elevator"
    "time"
    "fmt"
)

const masterTimeoutPeriod   = 5 * time.Second
const sendInterval          = 200 * time.Millisecond

func InitClient(clientEvents    com.ClientEvents,
                masterEvents    com.MasterEvents,
                elevatorEvents  com.ElevatorEvents) {

    fmt.Println("Awaiting master")
    sendTicker := time.NewTicker(sendInterval)

    myID    := network.GetMyID()
    orders  := make([]com.Order, 0)

    for {
        select {
        case message := <- clientEvents.FromMaster:
            fmt.Println("Contacted by master")
            if len(orders) == 0 {
                fmt.Printf("Initiating client, master %s\n", message.Address)
                remainingOrders := clientLoop(clientEvents, masterEvents, elevatorEvents)

                fmt.Println("Waiting for new master")
                for _, order := range(remainingOrders) {
                    if order.TakenBy == myID {
                        orders = append(orders, order)
                    }
                }

                delegation.PrioritizeForSingleElevator(orders, myID, elevator.GetLastPassedFloor())
                setButtonLamps(orders, myID)
                priority := queue.GetPriority(orders, myID)
                if priority != nil {
                    elevatorEvents.NewTargetFloor <- priority.Button.
