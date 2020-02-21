package main

import (
	"fmt"

	"../driver"
)

var maxFloor = 3
var minFloor = 0

const ( // TODO: change to motor direction
	dirDown int = -1
	dirUp       = 1
	dirStop     = 0
)

// Cost calculate the cost of the given order.
// The cost is calculated based on the current position
func Cost(order driver.Order, state driver.ElevState) int {
	// Functions used to calculate cost if directions doesn't match
	updown := func(c, t int) int { return (maxFloor - c) + (maxFloor - t) }
	downup := func(c, t int) int { return (c - minFloor) + (t - minFloor) }

	// Elevator state variables
	current := state.CurrentFloor
	currentDir := state.Direction // TODO: change to motor direction

	target := order.TargetFloor
	targetDir := order.Type

	targetAbove := target > current
	targetBelow := target < current
	atTarget := target == current

	if currentDir == dirStop {
		cost := target - current
		if cost < 0 {
			return -1 * cost
		}
		return cost
	} else if atTarget {
		if currentDir == dirUp {
			return updown(current, target)
		} else if currentDir == dirDown {
			return downup(current, target)
		}

	} else if targetAbove {
		if currentDir == dirDown {
			return downup(current, target)
		}
		// else <=> currentDir == dirUp
		if targetDir == driver.O_HallDown {
			return updown(current, target)
		}
		// else <=> targetDir == driver.O_HallUp || targetDir == driver.O_Cab
		return target - current

	} else if targetBelow {
		if currentDir == dirUp {
			return updown(current, target)
		}
		// else <=> currentDir == dirDown
		if targetDir == driver.O_HallUp {
			return downup(current, target)
		}
		// else <=> targetDir == driver.O_HallUp || targetDir == driver.O_Cab
		return current - target
	}

	return -1 // invalid
}

func main() {
	c1 := Cost(
		driver.Order{
			TargetFloor: 2,
			Type:        driver.O_HallUp,
		},
		driver.ElevState{
			CurrentFloor: 0,
			Direction:    dirUp,
		},
	)
	c1ans := 2

	c2 := Cost(
		driver.Order{
			TargetFloor: 2,
			Type:        driver.O_HallDown,
		},
		driver.ElevState{
			CurrentFloor: 0,
			Direction:    dirUp,
		},
	)
	c2ans := 4

	c3 := Cost(
		driver.Order{
			TargetFloor: 2,
			Type:        driver.O_HallUp,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    dirUp,
		},
	)
	c3ans := 1

	c4 := Cost(
		driver.Order{
			TargetFloor: 2,
			Type:        driver.O_HallDown,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    dirUp,
		},
	)
	c4ans := 3

	c5 := Cost(
		driver.Order{
			TargetFloor: 2,
			Type:        driver.O_Cab,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    dirUp,
		},
	)
	c5ans := 1

	c6 := Cost(
		driver.Order{
			TargetFloor: 2,
			Type:        driver.O_Cab,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    dirDown,
		},
	)
	c6ans := 3

	c7 := Cost(
		driver.Order{
			TargetFloor: 2,
			Type:        driver.O_HallUp,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    dirDown,
		},
	)
	c7ans := 3

	c8 := Cost(
		driver.Order{
			TargetFloor: 1,
			Type:        driver.O_HallUp,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    dirUp,
		},
	)
	c8ans := 3

	c9 := Cost(
		driver.Order{
			TargetFloor: 1,
			Type:        driver.O_HallDown,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    dirDown,
		},
	)
	c9ans := 1

	c10 := Cost(
		driver.Order{
			TargetFloor: 1,
			Type:        driver.O_HallUp,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    dirDown,
		},
	)
	c10ans := 3

	c11 := Cost(
		driver.Order{
			TargetFloor: 1,
			Type:        driver.O_Cab,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    dirDown,
		},
	)
	c11ans := 1

	c12 := Cost(
		driver.Order{
			TargetFloor: 2,
			Type:        driver.O_Cab,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    dirDown,
		},
	)
	c12ans := 4

	c13 := Cost(
		driver.Order{
			TargetFloor: 2,
			Type:        driver.O_Cab,
		},
		driver.ElevState{
			CurrentFloor: 2,
			Direction:    dirUp,
		},
	)
	c13ans := 2

	c14 := Cost(
		driver.Order{
			TargetFloor: 3,
			Type:        driver.O_HallDown,
		},
		driver.ElevState{
			CurrentFloor: 1,
			Direction:    dirDown,
		},
	)
	c14ans := 4

	c15 := Cost(
		driver.Order{
			TargetFloor: 3,
			Type:        driver.O_HallDown,
		},
		driver.ElevState{
			CurrentFloor: 3,
			Direction:    dirStop,
		},
	)
	c15ans := 0

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

}
