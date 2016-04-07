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

	driver.MotorUp()

	timeout := make(chan bool)
	success := make(chan bool)

	go func() {
		time.Sleep(time.Duration(deadline) * time.Second)
		driver.MotorStop()
		timeout <- true
	}()

	go func() {
		fmt.Printf("%d\n", driver.GetFloorSignal())
		for driver.GetFloorSignal() != driver.NumFloors - 1 {}
		driver.MotorDown()
		for driver.GetFloorSignal() != 0 {}
		driver.MotorStop()
		success <- true
	}()

	select {
	case <- timeout:
		t.Fatalf("Elevator test failed to finish within %d seconds", deadline)
	case <- success:
		t.Logf("Elevator test succeeded within %d seconds", deadline)
	}
}
