package elevator

import (
	"../driver"
	"time"
	"log"
)

const deadlinePeriod	= time.Duration(5 * driver.NumFloors) * time.Second
const doorPeriod		= 3 * time.Second

var lastPassedFloor int

type state int
const (
	idle state = iota
	doorOpen
	moving
)

func GetLastPassedFloor() int {
	return lastPassedFloor
}

func Init(
		completedFloor	chan <- int,
		missedDeadline	chan <- bool,
		floorReached	<- chan int,
		newTargetFloor	<- chan int,
		elevLogger		log.Logger) {

	deadlineTimer := time.NewTimer(deadlinePeriod)
	deadlineTimer.Stop()

	doorTimer := time.NewTimer(doorPeriod)
	doorTimer.Stop()

	state := idle
	lastPassedFloor := 0
	targetFloor := -1

	for {
		select {
		case <- deadlineTimer.C:
			missedDeadline <- true

		case <- doorTimer.C:
			switch state {
				case doorOpen:
					elevLogger.Print("Door timer, state at doorOpen")
					driver.SetDoorOpenLamp(0)
					state = idle
					completedFloor <- targetFloor
					targetFloor = -1
					deadlineTimer.Stop()
				case idle:
					elevLogger.Print("Door timer, state at idle")
				case moving:
					elevLogger.Print("Door timer, state at moving")
			}

		case floor := <- newTargetFloor:
			if targetFloor != floor {
				deadlineTimer.Reset(deadlinePeriod)
			}
			targetFloor = floor
			switch state {
				case idle:
					elevLogger.Printf("New order for floor %d, state at idle", targetFloor+1)
					if targetFloor == -1 {
						break
					} else if targetFloor > lastPassedFloor {
						state = moving
						driver.MotorUp()
					} else if targetFloor < lastPassedFloor {
						state = moving
						driver.MotorDown()
					} else {
						doorTimer.Reset(doorPeriod)
						driver.SetDoorOpenLamp(1)
						driver.MotorStop()
						state = doorOpen
					}
				case moving:
					elevLogger.Printf("New order for floor %d, state at moving", targetFloor+1)
				case doorOpen:
					elevLogger.Printf("New order for floor %d, state at doorOpen", targetFloor+1)
			}

		case floor := <- floorReached:
			lastPassedFloor = floor
			switch state {
				case moving:
					elevLogger.Printf("Reached floor %d, state at moving", floor+1)
					driver.SetFloorIndicator(floor)
					if targetFloor == -1 {
						break
					} else if targetFloor > lastPassedFloor {
						state = moving
						driver.MotorUp()
					} else if targetFloor < lastPassedFloor {
						state = moving
						driver.MotorDown()
					} else {
						doorTimer.Reset(doorPeriod)
						driver.SetDoorOpenLamp(1)
						driver.MotorStop()
						state = doorOpen
					}
				case idle:
					elevLogger.Printf("Reached floor %d, state at idle", floor+1)
				case doorOpen:
					elevLogger.Printf("Reached floor %d, state at doorOpen", floor+1)
			}
		}
	}
}
