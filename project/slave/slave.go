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
	sendInterval	= 100 * time.Millisecond
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
            if len(orders) == 0 {
                slaveLogger.Printf("Initiating slave, master %s", message.Address)
                remainingOrders := slaveLoop(slaveEvents, masterEvents, elevatorEvents, slaveLogger)

                slaveLogger.Print("Waiting for new master")
                for _, order := range(remainingOrders) {
                    if order.TakenBy == myIP {
                        orders = append(orders, order)
                    }
                }
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
			delegation.PrioritizeForSingleElevator(orders, myIP, elevator.GetLastPassedFloor())
			driver.ClearAllButtonLamps()
			for _, o := range(orders) {
				if o.Button.Type == driver.ButtonCallCommand && o.TakenBy != myIP {
					continue
				}
				driver.SetButtonLamp(o.Button.Type, o.Button.Floor, 1)
			}
			priority := order.GetPriority(orders, myIP)
			if priority != nil {
				elevatorEvents.NewTargetFloor <- priority.Button.Floor
			}

        case button := <- slaveEvents.ButtonPressed:
            if button.Type == driver.ButtonCallCommand {
                orders = append(orders, order.Order {Button: button, TakenBy: myIP})

				delegation.PrioritizeForSingleElevator(orders, myIP, elevator.GetLastPassedFloor())
				driver.ClearAllButtonLamps()
				for _, o := range(orders) {
					if o.Button.Type == driver.ButtonCallCommand && o.TakenBy != myIP {
						continue
					}
					driver.SetButtonLamp(o.Button.Type, o.Button.Floor, 1)
				}
				priority := order.GetPriority(orders, myIP)
				if priority != nil {
					elevatorEvents.NewTargetFloor <- priority.Button.Floor
				}
            }

        case <- sendTicker.C:
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
            slaveLogger.Printf("Completed floor %d", floor+1)
            for _, o := range(orders) {
                if o.TakenBy == myIP && o.Button.Floor == floor {
                    o.Done = true
                    requests = append(requests, o)
                }
            }

        case message := <- slaveEvents.FromMaster:
            data, err := com.DecodeMasterMessage(message.Data)
            if err != nil {
                break
            }
            slaves = data.Slaves
            orders = data.Orders
            isBackup = (data.AssignedBackup == myIP)

			driver.ClearAllButtonLamps()
			for _, o := range(orders) {
				if o.Button.Type == driver.ButtonCallCommand && o.TakenBy != myIP {
					continue
				}
				driver.SetButtonLamp(o.Button.Type, o.Button.Floor, 1)
			}

			priority := order.GetPriority(orders, myIP)
			if priority != nil && !order.OrderDone(*priority, requests) {
				elevatorEvents.NewTargetFloor <- priority.Button.Floor
			}
			// Remove acknowledged orders
			for i := 0; i < len(requests); i++ {
				r := requests[i]
				sentToMaster := false
				acknowledged := false
				for _, o := range(orders) {
					if order.OrdersEqual(r, o) {
						sentToMaster = true
						if r.Done == o.Done {
							acknowledged = true
						}
					}
				}
				if !sentToMaster && r.Done {
					acknowledged = true
				}
				if acknowledged {
					requests = append(requests[:i], requests[i+1:]...)
					i--
				}
			}
		}
	}
}
