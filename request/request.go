package request

import (
	"fmt"

	"../elevTypes/elevator"
	"../elevTypes/order"
)

var lastHallCall order.Order

func orderBelow(elev elevator.Elevator) (int, order.Type, bool) {
	for f := elev.Floor - 1; f >= 0; f-- {
		for t := range elev.Orders[f] {
			l := len(elev.Orders[f])
			i := l - t - 1
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, order.Type(i), true
			}
		}
	}

	return -1, -1, false
}

func orderAbove(elev elevator.Elevator) (int, order.Type, bool) {
	for f := elev.Floor + 1; f < 4; f++ {
		for t := range elev.Orders[f] {
			l := len(elev.Orders[f])
			i := l - t - 1
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, order.Type(i), true
			}
		}
	}

	return -1, -1, false
}

func orderAtFloor(elev elevator.Elevator) (int, order.Type, bool) {
	for t := range elev.Orders[elev.Floor] {
		l := len(elev.Orders[elev.Floor])
		i := l - t - 1
		if elev.Orders[elev.Floor][i].Status == order.NotTaken {
			return elev.Floor, order.Type(i), true
		}
	}
	return -1, -1, false
}

func orderBetween(elev elevator.Elevator) (int, order.Type, bool) {
	// checks if there is NotTaken order between current pos and active order
	if elev.Direction == elevator.Up {
		for f := elev.Floor; f < elev.ActiveOrder.Floor; f++ {
			for t := range elev.Orders[f] {
				l := len(elev.Orders[f])
				i := l - t - 1
				if elev.Orders[f][i].Status == order.NotTaken {
					return f, order.Type(i), true
				}
			}
		}
	} else if elev.Direction == elevator.Down {
		for f := elev.Floor; f > elev.ActiveOrder.Floor; f-- {
			for t := range elev.Orders[f] {
				l := len(elev.Orders[f])
				i := l - t - 1
				if elev.Orders[f][i].Status == order.NotTaken {
					return f, order.Type(i), true
				}
			}
		}
	}
	return -1, -1, false
}

// FindNextOrder evaluates all NotTaken orders and selects the best next order.
func FindNextOrder(elev elevator.Elevator) order.Order {
	if elev.ActiveOrder.Type == order.HallUp || elev.ActiveOrder.Type == order.HallDown {
		lastHallCall = elev.ActiveOrder
	}

	var f int
	var t order.Type
	var ok bool = false
	switch lastHallCall.Type {
	case order.HallUp:
		if f, t, ok = orderAtFloor(elev); !ok || t == order.HallDown {
			ok = false
		}
		if !ok {
			if f, t, ok = orderAbove(elev); t == order.HallDown {
				ok = false
			}
		}
		if !ok {
			f, t, ok = orderAbove(elev)
		}
		if !ok {
			f, t, ok = orderAtFloor(elev)
		}
		if !ok {
			f, t, ok = orderBelow(elev)
		}
	case order.HallDown:
		if f, t, ok = orderAtFloor(elev); !ok ||
			t == order.HallUp || f < elev.ActiveOrder.Floor {
			ok = false
		}
		if !ok {
			if f, t, ok = orderBelow(elev); !ok || t == order.HallUp {
				ok = false
			}
		}
		if !ok {
			f, t, ok = orderBelow(elev)
		}
		if !ok {
			if f, t, ok = orderAtFloor(elev); f < elev.ActiveOrder.Floor {
				ok = false
			}

		}
		if !ok {
			fmt.Println("Before")
			if f, t, ok = orderAbove(elev); f < elev.ActiveOrder.Floor {
				fmt.Println("After")
				ok = false
			}
		}
	}
	// if !ok {
	// 	f, t, ok = orderBetween(elev)
	// }

	// f, t, ok := orderAtFloor(elev)
	// if !ok {
	// 	f, t, ok = orderAbove(elev)
	// }
	// if !ok {
	// 	f, t, ok = orderBelow(elev)
	// }
	// if !ok {
	// 	f, t, ok = orderBetween(elev)
	// }

	o := order.Order{Floor: f, Type: t, Status: order.Execute}
	if !ok {
		// no orders exist
		o.Status = order.Invalid
	}

	return o
}
