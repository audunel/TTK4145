package driver

/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "elev.h"
*/
import "C"

type MotorDirection int

const (
	DirnDown MotorDirection = -1
	DirnStop MotorDirection = 0
	DirnUp   MotorDirection = 1
)

type ButtonType int

const (
	ButtonCallUp      ButtonType = 0
	ButtonCallDown    ButtonType = 1
	ButtonCallCommand ButtonType = 2
)

type OrderButton struct {
	Type  ButtonType
	Floor int
}

const (
	NumFloors  = int(C.N_FLOORS)
	NumButtons = int(C.N_BUTTONS)
)

func ElevInit() {
	C.elev_init()

	ClearAllButtonLamps()
	SetStopLamp(0)
	SetDoorOpenLamp(0)
	SetFloorIndicator(0)

	SetMotorDirection(DirnDown)
	for GetFloorSignal() == -1 {
	}
	SetMotorDirection(DirnStop)
}

func EventListener(buttonEvent chan OrderButton, floorEvent chan int) {
	buttonWasActive := make(map[ButtonType][NumFloors]bool)
	var buttonSignal, floorSignal int

	lastPassedFloor := -1

	for {
		floorSignal = GetFloorSignal()
		if floorSignal != lastPassedFloor && floorSignal != -1 {
			floorEvent <- floorSignal
			lastPassedFloor = floorSignal
		}

		for floor := 0; floor < NumFloors; floor++ {
			for button := ButtonCallUp; int(button) < NumButtons; button++ {
				if (floor == 0) && (button == ButtonCallDown) {
					continue
				}
				if (floor == NumFloors-1) && (button == ButtonCallUp) {
					continue
				}
				buttonSignal = GetButtonSignal(button, floor)
				if (buttonSignal == 1) && !buttonWasActive[button][floor] {
					buttonEvent <- OrderButton{Type: button, Floor: floor}
				}
				floorList := buttonWasActive[button]
				floorList[floor] = (buttonSignal == 1)
				buttonWasActive[button] = floorList
			}
		}
	}
}

func ClearAllButtonLamps() {
	for floor := 0; floor < NumFloors; floor++ {
		if floor < NumFloors-1 {
			SetButtonLamp(ButtonCallUp, floor, 0)
		}
		if floor > 0 {
			SetButtonLamp(ButtonCallDown, floor, 0)
		}
		SetButtonLamp(ButtonCallCommand, floor, 0)
	}
}

func SetMotorDirection(dirn MotorDirection) {
	C.elev_set_motor_direction(C.elev_motor_direction_t(dirn))
}

func SetButtonLamp(button ButtonType, floor, value int) {
	C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor), C.int(value))
}

func SetFloorIndicator(floor int) {
	C.elev_set_floor_indicator(C.int(floor))
}

func SetDoorOpenLamp(value int) {
	C.elev_set_door_open_lamp(C.int(value))
}

func SetStopLamp(value int) {
	C.elev_set_stop_lamp(C.int(value))
}

func GetButtonSignal(button ButtonType, floor int) int {
	return int(C.elev_get_button_signal(C.elev_button_type_t(button), C.int(floor)))
}

func GetFloorSignal() int {
	return int(C.elev_get_floor_sensor_signal())
}

func GetStopSignal() int {
	return int(C.elev_get_stop_signal())
}

func GetObstructionSignal() int {
	return int(C.elev_get_obstruction_signal())
}
