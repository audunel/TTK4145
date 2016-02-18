package main

import (
	"../driver"
	"time"
	"testing"
)

const deadline = 20

func TestDriver(t *testing.T) {
	driver.ElevInit()

	driver.SetMotorDirection(driver.DirnUp)
	go func() {
		time.Sleep(deadline * time.Second)
		driver.SetMotorDirection(driver.DirnStop)
		t.Fatalf("Elevator test failed to finish in %d seconds", deadline)
	}()
	for driver.GetFloorSignal() != driver.NumFloors - 1 {}
	driver.SetMotorDirection(driver.DirnDown)
	for driver.GetFloorSignal() != 0 {}
	driver.SetMotorDirection(driver.DirnStop)
}
