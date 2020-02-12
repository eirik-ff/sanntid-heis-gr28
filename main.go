package main

import (
	"fmt"

	"./driver"
)

func main() {
	orderChan := make(chan driver.Order)
	execOrderChan := make(chan driver.Order)

	go driver.Driver(orderChan, execOrderChan)

	for {
		select {
		case order := <-orderChan:
			fmt.Println(order)
			execOrderChan <- order
		}
	}
}
