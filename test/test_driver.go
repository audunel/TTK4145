package main

import (
	"../driver"
	"time"
)

func main() {
	driver.ElevInit()

	go func() {
		for {
			for i := 0; i < driver.NumFloors; i++ {
				driver.SetButtonLamp(driver.ButtonCallCommand, i, 1)
				time.Sleep(1 * time.Second)
			}
			for i := 0; i < driver.NumFloors; i++ {
				driver.SetButtonLamp(driver.ButtonCallCommand, i, 0)
				time.Sleep(1 * time.Second)
			}
		}
	}()	

	driver.SetMotorDirection(driver.DirnUp)
	for {
		if driver.GetFloorSignal() == driver.NumFloors - 1 {
			driver.SetMotorDirection(driver.DirnDown)
		} else if driver.GetFloorSignal() == 0 {
			driver.SetMotorDirection(driver.DirnUp)
		}
	}
}
