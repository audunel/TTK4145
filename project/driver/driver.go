package driver

/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "elev.h"
*/
import "C"

type motorDirection int
type ButtonType int

const (
	dirnDown motorDirection = -1
	dirnStop motorDirection = 0
	dirnUp   motorDirection = 1

	ButtonCallUp      ButtonType = 0
	ButtonCallDown    ButtonType = 1
	ButtonCallCommand ButtonType = 2

	NumFloors  = int(C.N_FLOORS)
	NumButtons = int(C.N_BUTTONS)
)

func ElevInit() {
	C.elev_init()

	ClearAllButtonLamps()
	SetStopLamp(0)
	SetDoorOpenLamp(0)

	MotorDown()
	for GetFloorSignal() != 0 {}
	MotorStop()
}

func ClearAllButtonLamps() {
	for floor := 0; floor < NumFloors; floor++ {
		if floor < NumFloors - 1 {
			SetButtonLamp(ButtonCallDown, floor, 0)
		}
		if floor > 0 {
			SetButtonLamp(ButtonCallUp, floor, 0)
		}
		SetButtonLamp(ButtonCallCommand, floor, 0)
	}
}

func setMotorDirection(dirn motorDirection) {
	C.elev_set_motor_direction(C.elev_motor_direction_t(dirn))
}

func MotorDown() {
	setMotorDirection(dirnDown)
}

func MotorStop() {
	setMotorDirection(dirnStop)
}

func MotorUp() {
	setMotorDirection(dirnUp)
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
