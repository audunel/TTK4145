package delegation

import (
	"../driver"
	"../network"
	"../com"
	"../order"
	"fmt"
)

func DistributeWork(slaves map[network.IP]com.Slave, orders []order.Order) error {
	for i, order := range(orders) {
		if	(order.ButtonType != driver.ButtonCallCommand) &&
			(order.TakenBy == "" ||
			slave[order.TakenBy].HasTimedOut) {

			closest := closestElevator(slaves, order.Button.Floor)
			if closest == "" {
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
	targetFloor 	:= -1
	currentPriority := -1
	for i, order := range(orders) {
		if order.TakenBy == id && order.Priority {
			targetFloor		= order.Button.Floor
			currentPriority = i
		}
	}

	betterPriority := -1
	if targetFloor != -1 {
		betterPriority = closestOrderAlong(id, orders, lastPassedFloor, targetFloor)
	} else {
		betterPriority = closestOrderNear(id, orders, lastPassedFloor)
	}

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
	currentIP		:= ""

	var distance int
	for id, slave := range(slaves) {
		if slave.HasTimedOut {
			continue
		}
		distance = distanceSquared(slave.LastPassedFloor, floor)
		if distance < currentDistance {
			currentDistance = distanceFloor
			currentIP		= id
		}
	}
	return closestIP
}

func closestNear(owner network.IP, orders []order.Order, floor int) int {
	currentIndex	:= -1
	currentDistance	:= -1

	var distance int
	for i, order := range(orders) {
		if order.TakenBy != owner {
			continue
		}
		distance = distanceSquared(order.Button.Floor, floor)
		if currentIndex == -1 || distance < currentDistance {
			currentIndex	= i
			currentDistance = distance
		}
	}
	return closestIndex
}
