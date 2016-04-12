package delegation

import (
	"../driver"
	"../network"
	"../com"
	"../order"
	"fmt"
)

const (
	InvalidFloor	= -1
	InvalidIP		= network.IP("")
)

func DelegateWork(slaves map[network.IP]com.Slave, orders []order.Order) error {
	for i, order := range(orders) {
		if	(order.Button.Type != driver.ButtonCallCommand) &&
			(order.TakenBy == InvalidIP ||
			slaves[order.TakenBy].HasTimedOut) {

			closest := closestElevator(slaves, order.Button.Floor)
			if closest == InvalidIP {
				return fmt.Errorf("No active elevators")
			}
			order.TakenBy = closest
			orders[i] = order
		}
	}

	for id, slave := range(slaves) {
		PrioritizeForSingleElevator(orders, id, slave.LastPassedFloor)
	}

	return nil
}

func PrioritizeForSingleElevator(orders []order.Order, id network.IP, lastPassedFloor int) {
	targetFloor 	:= InvalidFloor
	currentPriority := -1
	for i, order := range(orders) {
		if order.TakenBy == id && order.Priority {
			targetFloor		= order.Button.Floor
			currentPriority = i
		}
	}

	betterPriority := closestOrder(id, orders, lastPassedFloor, targetFloor)

	if betterPriority >= 0 {
		if currentPriority >= 0 {
			orders[currentPriority].Priority = false
		}
		orders[betterPriority].Priority = true
	}
}

func distanceSquared(x, y int) int {
	return (x - y) * (x - y)
}

func closestElevator(slaves map[network.IP]com.Slave, floor int) network.IP {
	currentDistance	:= driver.NumFloors * driver.NumFloors
	currentIP		:= InvalidIP

	var distance int
	for ip, slave := range(slaves) {
		if slave.HasTimedOut {
			continue
		}
		distance = distanceSquared(slave.LastPassedFloor, floor)
		if distance < currentDistance {
			currentDistance = distance
			currentIP		= ip
		}
	}
	return currentIP
}


func closestOrder(owner network.IP, orders []order.Order, floor, targetFloor int) int {
	currentIndex	:= -1
	currentDistance	:= -1

	var distance int
	for i, order := range(orders) {
		if order.TakenBy != owner {
			continue
		}

		if targetFloor == -1 { // No target floor, find closest
			distance = distanceSquared(order.Button.Floor, floor)
		} else {
			if !(floor < order.Button.Floor && order.Button.Floor < targetFloor) {
				continue
			}

			dirUp	:= targetFloor - floor > 0
			dirDown	:= !dirUp

			orderUp		 := order.Button.Type == driver.ButtonCallUp
			orderDown	 := order.Button.Type == driver.ButtonCallDown
			orderCommand := order.Button.Type == driver.ButtonCallCommand

			if orderCommand || ((orderUp && dirUp) || (orderDown && dirDown)) {
				distance = distanceSquared(order.Button.Floor, floor)
			}
		}

		if currentIndex == -1 || distance < currentDistance {
			currentIndex	= i
			currentDistance = distance
		}
	}
	return currentIndex
}
