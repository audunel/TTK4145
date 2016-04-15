package order

import (
	"../driver"
	"../network"
)

type Order struct {
	Button	 driver.OrderButton
	TakenBy	 network.IP
	Done	 bool
	Priority bool
}

func OrdersEqual(order1, order2 Order) bool {
	return	order1.Button.Floor == order2.Button.Floor &&
			order1.Button.Type == order2.Button.Type
}

func OrderNew(request Order, orders []Order) bool {
	for _, o := range(orders) {
		if OrdersEqual(request, o) {
			return false
		}
	}
	return true
}

func OrderDone(order Order, orders []Order) bool {
	for _, o := range(orders) {
		if OrdersEqual(o, order) && o.Done {
			return true
		}
	}
	return false
}

func GetPriority(orders []Order, ip network.IP) *Order {
	for _, o := range(orders) {
		if o.TakenBy == ip && o.Priority {
			return &o
		}
	}
	return nil
}

func PrioritizeOrders(orders []Order, ip network.IP, lastPassedFloor int, currentDirection driver.MotorDirection) {
	targetFloor 	:= -1
	currentPriority := -1
	for i, o := range(orders) {
		if o.TakenBy == ip && o.Priority {
			targetFloor		= o.Button.Floor
			currentPriority = i
		}
	}

	betterPriority := closestOrder(ip, orders, lastPassedFloor, targetFloor, currentDirection)

	if betterPriority >= 0 {
		if currentPriority >= 0 {
			orders[currentPriority].Priority = false
		}
		orders[betterPriority].Priority = true
	}
}

func closestOrder(ip network.IP, orders []Order, lastPassedFloor, currentTargetFloor int, currentDirection driver.MotorDirection) int {
	currentIndex	:= -1
	currentDistance	:= -1

	var distance int
	for i, o := range(orders) {
		if o.TakenBy != ip {
			continue
		}

		orderAbove	:= o.Button.Floor - lastPassedFloor > 0
		orderBelow	:= o.Button.Floor - lastPassedFloor < 0

		orderUp		 := o.Button.Type == driver.ButtonCallUp
		orderDown	 := o.Button.Type == driver.ButtonCallDown

		movingUp	:= currentDirection == driver.DirnUp
		movingDown	:= currentDirection == driver.DirnDown

		if (orderAbove && movingUp && !orderDown) || (orderBelow && movingUp && !orderUp) {
			distance = distanceSquared(lastPassedFloor, o.Button.Floor)
		}
		if (orderAbove || orderUp) && movingDown {
			distance = distanceSquared(lastPassedFloor, 0) + distanceSquared(0, o.Button.Floor)
		}
		if (orderBelow || orderDown) && movingUp {
			distance = distanceSquared(lastPassedFloor, driver.NumFloors - 1) + distanceSquared(driver.NumFloors, o.Button.Floor)
		}

		if currentIndex == -1 || distance < currentDistance {
			currentIndex = i
			currentDistance = distance
		}
	}
	return currentIndex
}

func distanceSquared(x, y int) int {
	return (x - y) * (x - y)
}
