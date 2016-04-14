package order

import (
	"../driver"
	"../network"
)

const InvalidFloor = -1

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

func PrioritizeOrders(orders []Order, ip network.IP, lastPassedFloor int) {
	targetFloor 	:= InvalidFloor
	currentPriority := -1
	for i, o := range(orders) {
		if o.TakenBy == ip && o.Priority {
			targetFloor		= o.Button.Floor
			currentPriority = i
		}
	}

	betterPriority := closestOrder(ip, orders, lastPassedFloor, targetFloor)

	if betterPriority >= 0 {
		if currentPriority >= 0 {
			orders[currentPriority].Priority = false
		}
		orders[betterPriority].Priority = true
	}
}

	
func closestOrder(ip network.IP, orders []Order, floor, targetFloor int) int {
	currentIndex	:= -1
	currentDistance	:= -1

	var distance int
	for i, o := range(orders) {
		if o.TakenBy != ip {
			continue
		}

		if targetFloor == -1 { // No target floor, find closest
			distance = distanceSquared(o.Button.Floor, floor)
		} else {
		  if !(floor < o.Button.Floor && o.Button.Floor <= targetFloor)/* || (floor >= o.Button.Floor && o.Button.Floor > targetFloor))*/ {
				continue
			}

			dirUp	:= targetFloor - floor > 0
			dirDown	:= targetFloor - floor < 0

			orderUp		 := o.Button.Type == driver.ButtonCallUp
			orderDown	 := o.Button.Type == driver.ButtonCallDown
			orderCommand := o.Button.Type == driver.ButtonCallCommand

			if orderCommand || ((orderDown && dirDown) || (orderUp && dirUp)) {
				distance = distanceSquared(o.Button.Floor, floor)
			}/* else if (orderUp && dirDown) {
				distance = distanceSquared(floor, 0) + distanceSquared(0, o.Button.Floor); 
				fmt.Printf("Order up, dir down, distance = %d\n", distance)
			} else if (orderDown && dirUp) {
				distance = distanceSquared(floor, driver.NumFloors - 1) + distanceSquared(driver.NumFloors - 1, o.Button.Floor);
				fmt.Printf("Order down, dir up, distance = %d\n", distance)
			}*/
		  
		}

		if currentIndex == -1 || distance < currentDistance {
			currentIndex	= i
			currentDistance = distance
		}
	}
	return currentIndex
}

func distanceSquared(x, y int) int {
	return (x - y) * (x - y)
}
