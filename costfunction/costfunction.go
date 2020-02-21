package main

import (
	"fmt"

	"../driver"
)

var maxFloor = 4

// Cost calculate the cost of the given order.
// The cost is calculated based on the current position
func Cost(order driver.Order, state driver.ElevState) int {

	return -1
}

func main() {
	c1 := Cost(
		driver.Order{
			TargetFloor: 0,
			Type:        driver.OrderType(0),
		},
		driver.ElevState{
			Order: driver.Order{
				TargetFloor: 3,
				Type:        driver.OrderType(0),
			},
			CurrentFloor: 2,
		},
	)
	fmt.Println(c1)
}
