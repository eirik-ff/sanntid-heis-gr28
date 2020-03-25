package request

import (
	"../elevTypes/elevator"
	"../elevTypes/order"
)

func orderBelow(elev elevator.Elevator) (int, int, bool) {
	for f := elev.Floor - 1; f >= 0; f-- {
		for t := range elev.Orders[f] {
			l := len(elev.Orders[f])
			i := l - t - 1
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, i, true
			}
		}
	}

	return -1, -1, false
}

func orderAbove(elev elevator.Elevator) (int, int, bool) {
	for f := elev.Floor + 1; f < 4; f++ {
		for t := range elev.Orders[f] {
			l := len(elev.Orders[f])
			i := l - t - 1
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, i, true
			}
		}
	}

	return -1, -1, false
}

func orderAtFloor(elev elevator.Elevator) (int, int, bool) {
	for t := range elev.Orders[elev.Floor] {
		l := len(elev.Orders[elev.Floor])
		i := l - t - 1
		if elev.Orders[elev.Floor][i].Status == order.NotTaken {
			return elev.Floor, i, true
		}
	}
	return -1, -1, false
}

func orderBetween(elev elevator.Elevator) (int, int, bool) {
	// checks if there is NotTaken order between current pos and active order
	if elev.Direction == elevator.Up {
		for f := elev.Floor; f < elev.ActiveOrder.Floor; f++ {
			for t := range elev.Orders[f] {
				l := len(elev.Orders[f])
				i := l - t - 1
				if elev.Orders[f][i].Status == order.NotTaken {
					return f, i, true
				}
			}
		}
	} else if elev.Direction == elevator.Down {
		for f := elev.Floor; f > elev.ActiveOrder.Floor; f-- {
			for t := range elev.Orders[f] {
				l := len(elev.Orders[f])
				i := l - t - 1
				if elev.Orders[f][i].Status == order.NotTaken {
					return f, i, true
				}
			}
		}
	}
	return -1, -1, false
}

// FindNextOrder evaluates all NotTaken orders and selects the best next order.
func FindNextOrder(elev elevator.Elevator) order.Order {
	f, t, ok := orderAtFloor(elev)
	if !ok {
		f, t, ok = orderAbove(elev)
	}
	if !ok {
		f, t, ok = orderBelow(elev)
	}
	if !ok {
		f, t, ok = orderBetween(elev)
	}

	o := order.Order{Floor: f, Type: order.Type(t), Status: order.Execute}
	if !ok {
		// no orders exist
		o.Status = order.Invalid
	}

	return o
}
