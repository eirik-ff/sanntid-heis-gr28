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
			//			l := len(elev.Orders[f])
			//			i := l - t - 1
			if elev.Orders[f][t].Status == order.NotTaken {
				return f, order.Type(t), true
			}
		}
	}

	return -1, -1, false
}

func orderAboveFromElev(elev elevator.Elevator) (int, order.Type, bool) {
	for f := elev.Floor + 1; f < 4; f++ {
		for t := range elev.Orders[f] {
			//			l := len(elev.Orders[f])
			//			i := l - t - 1
			if elev.Orders[f][t].Status == order.NotTaken {
				return f, order.Type(t), true
			}
		}
	}

	return -1, -1, false
}

func orderAboveFromTop(elev elevator.Elevator) (int, order.Type, bool) {
	for f := elev.Nfloors - 1; f >= elev.Floor + 1; f-- {
		for t := range elev.Orders[f] {
			//			l := len(elev.Orders[f])
			//			i := l - t - 1
			if elev.Orders[f][t].Status == order.NotTaken {
				return f, order.Type(t), true
			}
		}
	}

	return -1, -1, false
}

func orderAtFloor(elev elevator.Elevator) (int, order.Type, bool) {
	for t := range elev.Orders[elev.Floor] {
		//		l := len(elev.Orders[elev.Floor])
		//		i := l - t - 1
		if elev.Orders[elev.Floor][t].Status == order.NotTaken {
			return elev.Floor, order.Type(t), true
		}
	}
	return -1, -1, false
}

func orderBetween(elev elevator.Elevator) (int, order.Type, bool) {
	// checks if there is NotTaken order between current pos and active order
	if elev.Direction == elevator.Up {
		for f := elev.Floor; f < elev.ActiveOrder.Floor; f++ {
			for t := range elev.Orders[f] {
				//				l := len(elev.Orders[f])
				//				i := l - t - 1
				if elev.Orders[f][t].Status == order.NotTaken {
					return f, order.Type(t), true
				}
			}
		}
	} else if elev.Direction == elevator.Down {
		for f := elev.Floor; f > elev.ActiveOrder.Floor; f-- {
			for t := range elev.Orders[f] {
				//				l := len(elev.Orders[f])
				//				i := l - t - 1
				if elev.Orders[f][t].Status == order.NotTaken {
					return f, order.Type(t), true
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
			if f, t, ok = orderAboveFromElev(elev); t == order.HallDown {
				ok = false
			}
		}
		if !ok {
			f, t, ok = orderAboveFromElev(elev)
		}
		if !ok {
			f, t, ok = orderAtFloor(elev)
		}
		if !ok {
			f, t, ok = orderBelow(elev)
		}

		if elev.ActiveOrder.Status != order.Finished {
			ok = false
		}
		
	case order.HallDown:
		if f, t, ok = orderAtFloor(elev); !ok || t == order.HallUp  {
			ok = false
		}
		if !ok {
			fmt.Println("from top")
			if f, t, ok = orderAboveFromTop(elev); !ok {
				ok = false
			}
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
			 f, t, ok = orderAtFloor(elev) 
				ok = false
			
		}
		if elev.ActiveOrder.Status != order.Finished {
			ok = false
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
