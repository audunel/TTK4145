package main

import (
	"../driver"
	"time"
	"testing"
	"fmt"
)

const deadline = 5 * driver.NumFloors

func TestDriver(t *testing.T) {
	driver.ElevInit()

	driver.SetMotorDirection(driver.DirnUp)

	timeout := make(chan bool)
	success := make(chan bool)

	go func() {
		time.Sleep(time.Duration(deadline) * time.Second)
		driver.SetMotorDirection(driver.DirnStop)
		timeout <- true
	}()

	go func() {
		fmt.Printf("%d\n", driver.GetFloorSignal())
		for driver.GetFloorSignal() != driver.NumFloors - 1 {}
		driver.SetMotorDirection(driver.DirnDown)
		for driver.GetFloorSignal() != 0 {}
		driver.SetMotorDirection(driver.DirnStop)
		success <- true
	}()

	select {
	case <- timeout:
		t.Fatalf("Elevator test failed to finish within %d seconds", deadline)
	case <- success:
		t.Logf("Elevator test succeeded within %d seconds", deadline)
	}
}
