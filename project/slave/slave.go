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
var myIP                    = network.GetMyIP()

func InitSlave(slaveEvents    com.SlaveEvents,
				masterEvents    com.MasterEvents,
				elevatorEvents  com.ElevatorEvents) {

    fmt.Println("Awaiting master")
    sendTicker := time.NewTicker(sendInterval)

    orders  := make([]order.Order, 0)

    for {
        select {
        case message := <- slaveEvents.FromMaster:
            fmt.Println("Contacted by master")
            if len(orders) == 0 {
                fmt.Printf("Initiating slave, master %s\n", message.Address)
                remainingOrders := slaveLoop(slaveEvents, masterEvents, elevatorEvents)

                fmt.Println("Waiting for new master")
                for _, order := range(remainingOrders) {
                    if order.TakenBy == myIP {
                        orders = append(orders, order)
                    }
                }
                handleOrders(orders, nil, myIP, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)
            }

        case <- slaveEvents.MissedDeadline:
            driver.MotorStop()
            fmt.Println("Failed to complete order within deadline")

        case floor := <- slaveEvents.CompletedFloor:
            fmt.Printf("Completed floor %d\n", floor)
            for i := 0; i < len(orders); i++ {
                order := orders[i]
                if order.TakenBy == myIP && order.Button.Floor == floor {
                    orders = append(orders[:i], orders[i+1:]...)
                }
            }
            handleOrders(orders, myIP, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)

        case button := <- slaveEvents.ButtonPressed:
            if button.Type == driver.ButtonCallCommand {
                orders = append(orders, order.Order {Button: button, TakenBy: myIP})
                handleOrders(orders, nil, myIP, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)
            }

        case <- sendTicker.C:
            fmt.Println("Pinging")
            data := com.SlaveData {
                LastPassedFloor: elevator.GetLastPassedFloor(),
            }
            slaveEvents.ToMaster <- network.message {
                Data: com.EncodeSlaveData(data),
            }
        }
    }
}

func slaveLoop(slaveEvents    com.SlaveEvents,
                masterEvents    com.MasterEvents,
                elevatorEvents  com.ElevatorEvents) []order.Order {

    sendTicker := time.NewTicker(sendInterval)

    slaves     := make(map[network.IP]com.Slave)
    orders      := make([]order.Order, 0)
    requests    := make([]order.Order, 0)

    isBackup := false

    for {
        select {
        case <- timer.After(masterTimeout):
            fmt.Println("Master timed out")
            if isBackup {
                go master.InitMaster(masterEvents, orders, slaves)
            }
            return orders

        case <- slaveEvents.MissedDeadline:
            driver.MotorStop()
            fmt.Println("Failed to complete order within deadline")

        case <- sendTicker.C:
            fmt.Println("Sending..")
            data := com.SlaveData {
                LastPassedFloor:    elevator.GetLastPassedFloor(),
                requests:           requests,
            }
            slaveEvents.ToMaster <- network.Message {
                Data: com.EncodeSlaveData(data),
            }

        case button := <- slaveEvents.ButtonPressed:
            fmt.Println("Button pressed")
            requests = append(requests, order.Order {Button: button})

        case floor := <- slaveEvents.CompletedFloor:
            fmt.Println("Completed floor")
            for _, order := range(orders) {
                if order.TakenBy == myIP && order.Button.Floor == floor {
                    order.Done = true
                    requests = append(requests, order)
                }
            }

        case message := <- slaveEvents.FromMaster:
            data, err := com.DecodeMasterMessage(message)
            if err != nil {
                break
            }
            fmt.Println("Message received")
            slaves = data.Slaves
            orders = data.Orders
            isBackup = (data.AssignedBackup == myIP)
            handleOrders(orders, requests, myIP, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)

func handleOrders(orders, requests []order.Order, IP network.IP, lastPassedFloor int, newTargetFloor chan int) {
    delegation.PrioritizeForSingleElevator(orders, myIP, lastPassedFloor)
    // TODO: Lamp control
    priority := queue.GetPriority(orders, myIP)
    if priority != nil && !queue.OrderDone(*priority, requests) {
        newTargetFloor <- priority.Button.Floor
    }
}
