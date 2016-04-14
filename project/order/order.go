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
