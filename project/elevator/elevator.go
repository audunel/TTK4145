package elevator

import (
	"../driver"
	"fmt"
	"log"
	"time"
)

const deadlinePeriod = time.Duration(5*driver.NumFloors) * time.Second
const doorPeriod = 3 * time.Second

type state int

const (
	idle state = iota
	doorOpen
	moving
)

type ElevData struct {
	LastPassedFloor  int
	CurrentDirection driver.MotorDirection
	Busy             bool
}

var elevData ElevData

func GetElevData() ElevData {
	return elevData
}

func Init(
	completedFloor chan<- int,
	missedDeadline chan<- bool,
	floorReached <-chan int,
	newTargetFloor <-chan int,
	elevLogger log.Logger) {

	deadlineTimer := time.NewTimer(deadlinePeriod)
	deadlineTimer.Stop()

	doorTimer := time.NewTimer(doorPeriod)
	doorTimer.Stop()

	state := idle
	targetFloor := -1

	for {
		select {
		case <-deadlineTimer.C:
			missedDeadline <- true

		case <-doorTimer.C:
			switch state {
			case doorOpen:
				elevLogger.Print("Door timer, state at doorOpen")
				driver.SetDoorOpenLamp(0)
				elevData.Busy = false
				state = idle
				completedFloor <- elevData.LastPassedFloor
				targetFloor = -1
				deadlineTimer.Stop()
			case idle:
				elevLogger.Print("Door timer, state at idle")
			case moving:
				elevLogger.Print("Door timer, state at moving")
			}

		case floor := <-newTargetFloor:
			if targetFloor != floor {
				elevLogger.Printf("New target floor %d\n", floor+1)
				deadlineTimer.Reset(deadlinePeriod)
			}
			targetFloor = floor
			switch state {
			case idle:
				elevLogger.Printf("New order for floor %d, state at idle", targetFloor+1)
				if targetFloor == -1 {
					break
				} else if targetFloor > elevData.LastPassedFloor {
					state = moving
					driver.SetMotorDirection(driver.DirnUp)
					elevData.CurrentDirection = driver.DirnUp
					elevData.Busy = true
				} else if targetFloor < elevData.LastPassedFloor {
					state = moving
					driver.SetMotorDirection(driver.DirnDown)
					elevData.CurrentDirection = driver.DirnDown
					elevData.Busy = true
				} else {
					doorTimer.Reset(doorPeriod)
					driver.SetDoorOpenLamp(1)
					driver.SetMotorDirection(driver.DirnStop)
					elevData.Busy = true
					state = doorOpen
				}
			case moving:
			case doorOpen:
			}

		case floor := <-floorReached:
			if (floor == driver.NumFloors-1) || (floor == 0) {
				elevData.CurrentDirection = driver.DirnStop
			}
			elevData.LastPassedFloor = floor
			switch state {
			case moving:
				elevLogger.Printf("Reached floor %d, target floor %d state at moving", floor+1, targetFloor+1)
				driver.SetFloorIndicator(floor)
				if targetFloor == -1 {
					break
				} else if targetFloor > elevData.LastPassedFloor {
					driver.SetMotorDirection(driver.DirnUp)
					elevData.CurrentDirection = driver.DirnUp
					elevData.Busy = true
				} else if targetFloor < elevData.LastPassedFloor {
					driver.SetMotorDirection(driver.DirnDown)
					elevData.CurrentDirection = driver.DirnDown
					elevData.Busy = true
				} else {
					fmt.Printf("Stopping at floor %d\n", floor)
					doorTimer.Reset(doorPeriod)
					driver.SetDoorOpenLamp(1)
					driver.SetMotorDirection(driver.DirnStop)
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
