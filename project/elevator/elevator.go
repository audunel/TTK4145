package elevator

import (
	"../driver"
	"time"
	"fmt"
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
		newTargetFloor	<- chan int) {

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
					fmt.Println("Door timer, state at doorOpen")
					driver.SetDoorOpenLamp(0)
					state = idle
					completedFloor <- targetFloor
					targetFloor = -1
					deadlineTimer.Stop()
				case idle:
					fmt.Println("Door timer, state at idle")
				case moving:
					fmt.Println("Door timer, state at moving")
			}

		case floor := <- newTargetFloor:
			if targetFloor != floor {
				deadlineTimer.Reset(deadlinePeriod)
			}
			targetFloor = floor
			switch state {
				case idle:
					fmt.Println("New order, state at idle")
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
					fmt.Println("New order, state at moving")
				case doorOpen:
					fmt.Println("New order, state at doorOpen")
			}

		case floor := <- floorReached:
			lastPassedFloor = floor
			switch state {
				case moving:
					fmt.Printf("Reached floor %d, state at moving\n", floor)
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
					}
				case idle:
					fmt.Printf("Reached floor %d, state at idle\n", floor)
				case doorOpen:
					fmt.Println("Reached floor %d, state at doorOpen\n", floor)
			}
		}
	}
}
