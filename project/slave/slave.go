package slave

import (
    "../driver"
    "../network"
    "../com"
    "../master"
    "../elevator"
	"../order"
	"../delegation"
	"../logger"
    "time"
    "log"
)

const (
	masterTimeout	= 5 * time.Second
	sendInterval	= 200 * time.Millisecond
)
var myIP = network.GetOwnIP()

func InitSlave(
		slaveEvents		com.SlaveEvent,
		masterEvents    com.MasterEvent,
		elevatorEvents  com.ElevatorEvent,
		slaveLogger		log.Logger) {

    slaveLogger.Print("Awaiting master")
    sendTicker := time.NewTicker(sendInterval)

    orders  := make([]order.Order, 0)

    for {
        select {
        case message := <- slaveEvents.FromMaster:
            slaveLogger.Print("Contacted by master")
            if len(orders) == 0 {
                slaveLogger.Printf("Initiating slave, master %s", message.Address)
                remainingOrders := slaveLoop(slaveEvents, masterEvents, elevatorEvents, slaveLogger)

                slaveLogger.Print("Waiting for new master")
                for _, order := range(remainingOrders) {
                    if order.TakenBy == myIP {
                        orders = append(orders, order)
                    }
                }
                handleOrders(orders, nil, myIP, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)
            }

        case <- slaveEvents.MissedDeadline:
            driver.MotorStop()
            slaveLogger.Fatal("Failed to complete order within deadline")

        case floor := <- slaveEvents.CompletedFloor:
            slaveLogger.Printf("Completed floor %d", floor)
            for i := 0; i < len(orders); i++ {
                order := orders[i]
                if order.TakenBy == myIP && order.Button.Floor == floor {
                    orders = append(orders[:i], orders[i+1:]...)
                }
            }
            handleOrders(orders, nil, myIP, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)

        case button := <- slaveEvents.ButtonPressed:
            if button.Type == driver.ButtonCallCommand {
                orders = append(orders, order.Order {Button: button, TakenBy: myIP})
                handleOrders(orders, nil, myIP, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)
            }

        case <- sendTicker.C:
            slaveLogger.Print("Pinging")
            data := com.SlaveData {
                LastPassedFloor: elevator.GetLastPassedFloor(),
            }
            slaveEvents.ToMaster <- network.UDPMessage {
                Data: com.EncodeSlaveData(data),
            }
        }
    }
}

func slaveLoop(
		slaveEvents		com.SlaveEvent,
		masterEvents    com.MasterEvent,
		elevatorEvents  com.ElevatorEvent,
		slaveLogger		log.Logger) []order.Order {

    sendTicker := time.NewTicker(sendInterval)

    slaves		:= make(map[network.IP]com.Slave)
    orders      := make([]order.Order, 0)
    requests    := make([]order.Order, 0)

    isBackup := false

    for {
        select {
        case <- time.After(masterTimeout):
            slaveLogger.Println("Master timed out")
            if isBackup {
                go master.InitMaster(masterEvents, orders, slaves, logger.NewLogger("MASTER"))
            }
            return orders

        case <- slaveEvents.MissedDeadline:
            driver.MotorStop()
            slaveLogger.Fatalf("Failed to complete order within deadline")

        case <- sendTicker.C:
            slaveLogger.Print("Sending..")
            data := com.SlaveData {
                LastPassedFloor:    elevator.GetLastPassedFloor(),
                Requests:           requests,
            }
            slaveEvents.ToMaster <- network.UDPMessage {
                Data: com.EncodeSlaveData(data),
            }

        case button := <- slaveEvents.ButtonPressed:
            slaveLogger.Print("Button pressed")
            requests = append(requests, order.Order {Button: button})

        case floor := <- slaveEvents.CompletedFloor:
            slaveLogger.Printf("Completed floor %d", floor)
            for _, order := range(orders) {
                if order.TakenBy == myIP && order.Button.Floor == floor {
                    order.Done = true
                    requests = append(requests, order)
                }
            }

        case message := <- slaveEvents.FromMaster:
            data, err := com.DecodeMasterMessage(message.Data)
            if err != nil {
                break
            }
            slaveLogger.Print("Message received")
            slaves = data.Slaves
            orders = data.Orders
            isBackup = (data.AssignedBackup == myIP)
            handleOrders(orders, requests, myIP, elevator.GetLastPassedFloor(), elevatorEvents.NewTargetFloor)
		}
	}
}

func handleOrders(orders, requests []order.Order, IP network.IP, lastPassedFloor int, newTargetFloor chan int) {
    delegation.PrioritizeForSingleElevator(orders, myIP, lastPassedFloor)
    // TODO: Lamp control
    priority := order.GetPriority(orders, myIP)
    if priority != nil {
		if requests == nil {
	        newTargetFloor <- priority.Button.Floor
		} else {
			for _, request := range(requests) {
				if order.OrdersEqual(*priority, request) {
					newTargetFloor <- priority.Button.Floor
					break
				}
			}
		}
	}
}
