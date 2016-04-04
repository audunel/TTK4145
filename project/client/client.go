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
var myID                    = network.GetMyID()

func InitClient(clientEvents    com.ClientEvents,
                masterEvents    com.MasterEvents,
                elevatorEvents  com.ElevatorEvents) {

    fmt.Println("Awaiting master")
    sendTicker := time.NewTicker(sendInterval)

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
                handleOrders(orders, nil, myID, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)
            }

        case <- clientEvents.MissedDeadline:
            driver.MotorStop()
            fmt.Println("Failed to complete order within deadline")

        case floor := <- clientEvents.CompletedFloor:
            fmt.Printf("Completed floor %d\n", floor)
            for i := 0; i < len(orders); i++ {
                order := orders[i]
                if order.TakenBy == myID && order.Button.Floor == floor {
                    orders = append(orders[:i], orders[i+1:]...)
                }
            }
            handleOrders(orders, myID, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)

        case button := <- clientEvents.ButtonPressed:
            if button.Type == driver.ButtonCallCommand {
                orders = append(orders, com.Order {Button: button, TakenBy: myID})
                handleOrders(orders, nil, myID, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)
            }

        case <- sendTicker.C:
            fmt.Println("Pinging")
            data := com.ClientData {
                LastPassedFloor: elevator.GetLastPassedFloor(),
            }
            clientEvents.ToMaster <- network.message {
                Data: com.EncodeClientData(data),
            }
        }
    }
}

func clientLoop(clientEvents    com.ClientEvents,
                masterEvents    com.MasterEvents,
                elevatorEvents  com.ElevatorEvents) []com.Order {

    sendTicker := time.NewTicker(sendInterval)

    clients     := make(map[network.ID]com.Client)
    orders      := make([]com.Order, 0)
    requests    := make([]com.Order, 0)

    isBackup := false

    for {
        select {
        case <- timer.After(masterTimeout):
            fmt.Println("Master timed out")
            if isBackup {
                go master.InitMaster(masterEvents, orders, clients)
            }
            return orders

        case <- clientEvents.MissedDeadline:
            driver.SetMotorDirection(Driver.DirnStop)
            fmt.Println("Failed to complete order within deadline")

        case <- sendTicker.C:
            fmt.Println("Sending..")
            data := com.ClientData {
                LastPassedFloor:    elevator.GetLastPassedFloor(),
                requests:           requests,
            }
            clientEvents.ToMaster <- network.Message {
                Data: com.EncodeClientData(data),
            }

        case button := <- clientEvents.ButtonPressed:
            fmt.Println("Button pressed")
            requests = append(requests, com.Order {Button: button})

        case floor := <- clientEvents.CompletedFloor:
            fmt.Println("Completed floor")
            for _, order := range(orders) {
                if order.TakenBy == myID && order.Button.Floor == floor {
                    order.Done = true
                    requests = append(requests, order)
                }
            }

        case message := <- clientEvents.FromMaster:
            data, err := com.DecodeMasterData(message)
            if err != nil {
                break
            }
            fmt.Println("Message received")
            clients = data.Clients
            orders = data.Orders
            isBackup = (data.AssignedBackup == myID)
            handleOrders(orders, requests, myID, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)

func handleOrders(orders, requests []com.Order, ID network.ID, lastPassedFloor int, newTargetFloor chan int) {
    delegation.PrioritizeForSingleElevator(orders, myID, lastPassedFloor)
    // TODO: Lamp control
    priority := queue.GetPriority(orders, myID)
    if priority != nil && !queue.OrderDone(*priority, requests) {
        newTargetFloor <- priority.Button.Floor
    }
}
