package costfunction

import (
	"fmt"

	"../driver"
	"../elevTypes/order"
)

var maxFloor = 3
var minFloor = 0

// Cost calculate the cost of the given order ord.
// The cost is calculated based on the current position
func Cost(ord order.Order, state driver.ElevState) int {
	// Functions used to calculate cost if directions doesn't match
	updown := func(c, t int) int { return (maxFloor - c) + (maxFloor - t) }
	downup := func(c, t int) int { return (c - minFloor) + (t - minFloor) }

	// Elevator state variables
	current := state.CurrentFloor
	currentDir := state.Direction

	target := ord.TargetFloor
	targetDir := ord.Type

	targetAbove := target > current
	targetBelow := target < current
	atTarget := target == current

	//Set orderDir based on cab call or not
	orderDir := state.Order.Type
	if state.Order.Type == order.Cab {
		if state.Direction == driver.MD_Up {
			orderDir = order.HallUp
		} else if state.Direction == driver.MD_Down {
			orderDir = order.HallDown
		} else {
			//If cab order and target above the direction is up,
			// else direction is down
			if targetAbove {
				orderDir = order.HallUp
			} else {
				orderDir = order.HallDown
			}
		}
	}

	if state.Order.Finished { // elevator is stationary and no active order
		fmt.Printf("############ Active finished: %#v	Target order: %#v\n", state.Order, ord)

		cost := target - current
		if cost < 0 {
			return -1 * cost
		}
		return cost
	} else if atTarget {
		fmt.Printf("############ At target with order: %#v\n", ord)
		if orderDir == order.HallUp {
			if currentDir == driver.MD_Down {
				return downup(current, target)
			}
			return updown(current, target)
		} else if orderDir == order.HallDown {
			if currentDir == driver.MD_Up {
				return updown(current, target)
			}
			return downup(current, target)
		}

	} else if targetAbove {
		fmt.Printf("############ Target above with order: %#v\n", ord)
		if orderDir == order.HallDown {
			if currentDir == driver.MD_Up {
				return updown(current, target)
			}
			// else <=> currentDir == driver.MD_Down
			return downup(current, target)
		}
		// else <=> orderDir == order.HallUp
		if targetDir == order.HallDown {
			return updown(current, target)
		}
		// else <=> targetDir == order.HallUp || targetDir == order.Cab
		return target - current

	} else if targetBelow {
		fmt.Printf("############ Target below with order: %#v\n", ord)
		if orderDir == order.HallUp {
			if currentDir == driver.MD_Down {
				return downup(current, target)
			}
			return updown(current, target)
		}
		// else <=> orderDir == order.HallDown
		if targetDir == order.HallUp {
			return downup(current, target)
		}
		// else <=> targetDir == order.HallUp || targetDir == order.Cab
		return current - target
	}

	return -1 // invalid
}

func TestCost() {
	c1 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.HallUp,
		},
		driver.ElevState{
			CurrentFloor: 0,
			Direction:    driver.MD_Up,
			Order:        order.Order{Type: order.HallUp},
		},
	)
	c1ans := 2

	c2 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.HallDown,
		},
		driver.ElevState{
			CurrentFloor: 0,
			Direction:    driver.MD_Up,
			Order:        order.Order{Type: order.HallUp},
		},
	)
	c2ans := 4

	c3 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.HallUp,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    driver.MD_Up,
			Order:        order.Order{Type: order.HallUp},
		},
	)
	c3ans := 1

	c4 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.HallDown,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    driver.MD_Up,
			Order:        order.Order{Type: order.HallUp},
		},
	)
	c4ans := 3

	c5 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.Cab,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    driver.MD_Up,
			Order:        order.Order{Type: order.HallUp},
		},
	)
	c5ans := 1

	c6 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.Cab,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    driver.MD_Down,
			Order:        order.Order{Type: order.HallDown},
		},
	)
	c6ans := 3

	c7 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.HallUp,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    driver.MD_Down,
			Order:        order.Order{Type: order.HallDown},
		},
	)
	c7ans := 3

	c8 := Cost(
		order.Order{
			TargetFloor: 1,
			Type:        order.HallUp,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    driver.MD_Up,
			Order:        order.Order{Type: order.HallUp},
		},
	)
	c8ans := 3

	c9 := Cost(
		order.Order{
			TargetFloor: 1,
			Type:        order.HallDown,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    driver.MD_Down,
			Order:        order.Order{Type: order.HallDown},
		},
	)
	c9ans := 1

	c10 := Cost(
		order.Order{
			TargetFloor: 1,
			Type:        order.HallUp,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    driver.MD_Down,
			Order:        order.Order{Type: order.HallDown},
		},
	)
	c10ans := 3

	c11 := Cost(
		order.Order{
			TargetFloor: 1,
			Type:        order.Cab,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    driver.MD_Down,
			Order:        order.Order{Type: order.HallDown},
		},
	)
	c11ans := 1

	c12 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.Cab,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    driver.MD_Down,
			Order:        order.Order{Type: order.HallDown},
		},
	)
	c12ans := 4

	c13 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.Cab,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    driver.MD_Up,
			Order:        order.Order{Type: order.HallUp},
		},
	)
	c13ans := 2

	c14 := Cost(
		order.Order{
			TargetFloor: 3,
			Type:        order.HallDown,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    driver.MD_Down,
			Order:        order.Order{Type: order.HallDown},
		},
	)
	c14ans := 4

	c15 := Cost(
		order.Order{
			TargetFloor: 3,
			Type:        order.HallDown,
		},
		driver.ElevState{
			CurrentFloor: 3,
			Direction:    driver.MD_Stop,
			Order:        order.Order{Type: order.Cab, Finished: true},
		},
	)
	c15ans := 0

	c16 := Cost(
		order.Order{
			TargetFloor: 0,
			Type:        order.HallDown,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    driver.MD_Stop,
			Order:        order.Order{Type: order.HallUp, TargetFloor: 3, Finished: false},
		},
	)
	c16ans := 5

	c17 := Cost(
		order.Order{
			TargetFloor: 0,
			Type:        order.Cab,
		},
		driver.ElevState{
			CurrentFloor: 0,
			Direction:    driver.MD_Up,
			Order:        order.Order{Type: order.HallDown, TargetFloor: 2, Finished: false},
		},
	)
	c17ans := 6

	c18 := Cost(
		order.Order{
			TargetFloor: 2,
			Type:        order.HallDown,
		},
		driver.ElevState{
			CurrentFloor: 0,
			Direction:    driver.MD_Up,
			Order:        order.Order{Type: order.HallDown, TargetFloor: 3, Finished: false},
		},
	)
	c18ans := 4

	c19 := Cost(
		order.Order{
			TargetFloor: 3,
			Type:        order.Cab,
		},
		driver.ElevState{
			CurrentFloor: 3,
			Direction:    driver.MD_Down,
			Order:        order.Order{Type: order.HallUp, TargetFloor: 1, Finished: false},
		},
	)
	c19ans := 6

	fmt.Printf("c1  = %d == %d ?   %t\n", c1, c1ans, c1 == c1ans)
	fmt.Printf("c2  = %d == %d ?   %t\n", c2, c2ans, c2 == c2ans)
	fmt.Printf("c3  = %d == %d ?   %t\n", c3, c3ans, c3 == c3ans)
	fmt.Printf("c4  = %d == %d ?   %t\n", c4, c4ans, c4 == c4ans)
	fmt.Printf("c5  = %d == %d ?   %t\n", c5, c5ans, c5 == c5ans)
	fmt.Printf("c6  = %d == %d ?   %t\n", c6, c6ans, c6 == c6ans)
	fmt.Printf("c7  = %d == %d ?   %t\n", c7, c7ans, c7 == c7ans)
	fmt.Printf("c8  = %d == %d ?   %t\n", c8, c8ans, c8 == c8ans)
	fmt.Printf("c9  = %d == %d ?   %t\n", c9, c9ans, c9 == c9ans)
	fmt.Printf("c10 = %d == %d ?   %t\n", c10, c10ans, c10 == c10ans)
	fmt.Printf("c11 = %d == %d ?   %t\n", c11, c11ans, c11 == c11ans)
	fmt.Printf("c12 = %d == %d ?   %t\n", c12, c12ans, c12 == c12ans)
	fmt.Printf("c13 = %d == %d ?   %t\n", c13, c13ans, c13 == c13ans)
	fmt.Printf("c14 = %d == %d ?   %t\n", c14, c14ans, c14 == c14ans)
	fmt.Printf("c15 = %d == %d ?   %t\n", c15, c15ans, c15 == c15ans)
	fmt.Printf("c16 = %d == %d ?   %t\n", c16, c16ans, c16 == c16ans)
	fmt.Printf("c17 = %d == %d ?   %t\n", c17, c17ans, c17 == c17ans)
	fmt.Printf("c18 = %d == %d ?   %t\n", c18, c18ans, c18 == c18ans)
	fmt.Printf("c19 = %d == %d ?   %t\n", c19, c19ans, c19 == c19ans)

}
