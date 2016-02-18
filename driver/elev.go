package driver

/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "elev.h"
*/
import "C"

type MotorDirection int
type ButtonType int

const (
	DirnDown MotorDirection = -1
	DirnStop MotorDirection = 0
	DirnUp   MotorDirection = 1

	ButtonCallUp      ButtonType = 0
	ButtonCallDown    ButtonType = 1
	ButtonCallCommand ButtonType = 2

	NumFloors  = int(C.N_FLOORS)
	NumButtons = int(C.N_BUTTONS)
)

func ElevInit() { C.elev_init() }

func SetMotorDirection(dirn MotorDirection) { C.elev_set_motor_direction(C.elev_motor_direction_t(dirn)) }
func SetButtonLamp(button ButtonType, floor, value int) { C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor), C.int(value)) }
func SetFloorIndicator(floor int) { C.elev_set_floor_indicator(C.int(floor)) }
func SetDoorOpenLamp(value int)   { C.elev_set_door_open_lamp(C.int(value)) }
func SetStopLamp(value int)       { C.elev_set_stop_lamp(C.int(value)) }

func GetButtonSignal(button ButtonType, floor int) int { return int(C.elev_get_button_signal(C.elev_button_type_t(button), C.int(floor))) }
func GetFloorSignal() int       { return int(C.elev_get_floor_sensor_signal()) }
func GetStopSignal() int        { return int(C.elev_get_stop_signal()) }
func GetObstructionSignal() int { return int(C.elev_get_obstruction_signal()) }
