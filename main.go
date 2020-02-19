package main

import (
	"fmt"

	"./queue"
)

func main() {
	OrderEnqueue := make(chan queue.Order)
	//OrderDenqueue := make(chan queue.Order)

	go queue.Queue(OrderEnqueue)

	for {
		select {
		case order := <-OrderEnqueue:
			fmt.Println(order)
			//OrderDenqueue <- order
		}
	}
}
