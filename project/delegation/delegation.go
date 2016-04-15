package delegation

import (
	"../driver"
	"../network"
	"../com"
	"../order"
	"fmt"
)

const (
	invalidIP	= network.IP("")
	maxDistance	= driver.NumFloors * driver.NumFloors
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

	for ip, slave := range(slaves) {
		order.PrioritizeOrders(orders, ip, slave.Data.LastPassedFloor, slave.Data.CurrentDirection)
	}

	return nil
}

func closestElevator(slaves map[network.IP]com.Slave, floor int) network.IP {
	currentDistance	:= maxDistance
	currentIP		:= InvalidIP

	var distance int
	for ip, slave := range(slaves) {
		if slave.HasTimedOut || slave.ElevData.busy {
			continue
		}
		distance = distanceSquared(slave.LastPassedFloor, floor)
		if distance < currentDistance {
			currentDistance = distance
			currentIP		= ip
		}
	}

	if currentDistance == maxDistance { // All elevators busy
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
	}
	return currentIP
}

func distanceSquared(x, y int) int {
	return (x - y) * (x - y)
}
