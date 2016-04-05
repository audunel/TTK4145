package main

import (
	"../driver"
	"time"
	"testing"
	"fmt"
)

const deadline = time.Duration(5 * driver.NumFloors) * time.Second

func TestDriver(t *testing.T) {
	driver.ElevInit()

	driver.SetMotorDirection(driver.DirnUp)
	
	go func() {
		time.Sleep(deadline)
		driver.SetMotorDirection(driver.DirnStop)
		t.Fatalf("Elevator test failed to finish in %d seconds", deadline)
	}()

	fmt.Printf("%d\n", driver.GetFloorSignal())
	for driver.GetFloorSignal() != driver.NumFloors - 1 {}
	driver.SetMotorDirection(driver.DirnDown)
	for driver.GetFloorSignal() != 0 {}
	driver.SetMotorDirection(driver.DirnStop)
}
