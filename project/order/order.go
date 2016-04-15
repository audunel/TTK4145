package order

import (
	"../driver"
	"../network"
	"fmt"
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

/*
func TestClosestOrder() {
	ip := network.IP("123")

	orders := make([]Order, 0)
	order1 := Order{Button: driver.OrderButton{Type: driver.ButtonCallCommand, Floor: 1}, TakenBy: ip, Done: false, Priority: true}
	order2 := Order{Button: driver.OrderButton{Type: driver.ButtonCallCommand, Floor: 3}, TakenBy: ip, Done: false, Priority: false}
	order3 := Order{Button: driver.OrderButton{Type: driver.ButtonCallCommand, Floor: 0}, TakenBy: ip, Done: false, Priority: false}
	orders = append(orders, order1, order2, order3)

	lastPassedFloor := 0
	currentTargetFloor := 3
	currentDirection := driver.DirnStop

	priority := closestOrder(ip, orders, lastPassedFloor, currentTargetFloor, currentDirection)
	fmt.Printf("Priority floor: %d\n", orders[priority].Button.Floor+1)
}
*/

func closestOrder(ip network.IP, orders []Order, lastPassedFloor, currentTargetFloor int, currentDirection driver.MotorDirection) int {
	currentIndex	:= -1
	currentDistance	:= -1

	var distance int
	for i, o := range(orders) {
		if o.TakenBy != ip {
			continue
		}
		
		orderUp		 := o.Button.Type == driver.ButtonCallUp
		orderDown	 := o.Button.Type == driver.ButtonCallDown

		movingUp	:= currentDirection == driver.DirnUp
		movingDown	:= currentDirection == driver.DirnDown

		if !orderDown && movingUp {
			if o.Button.Floor > lastPassedFloor {
				fmt.Printf("ONE Last passed floor: %d\nOld target floor: %d\nNew target floor: %d\n\n", lastPassedFloor, currentTargetFloor, o.Button.Floor)
				if (currentTargetFloor == -1) || (o.Button.Floor <= currentTargetFloor) {
					distance = distanceSquared(lastPassedFloor, o.Button.Floor)
				}
			}
		} else if !orderUp && movingDown {
			if o.Button.Floor < lastPassedFloor {
				fmt.Printf("TWO Last passed floor: %d\nOld target floor: %d\nNew target floor: %d", lastPassedFloor, currentTargetFloor, o.Button.Floor)
				if (currentTargetFloor == -1 ) || (o.Button.Floor >= currentTargetFloor) {
					distance = distanceSquared(lastPassedFloor, o.Button.Floor)
				}
			}
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
