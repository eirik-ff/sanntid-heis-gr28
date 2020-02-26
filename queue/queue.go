package queue

import (
	"container/list"
	"fmt"
)

type Order struct {
	TargetFloor int
	Cost        int
}

// Enqueues new order
func enqueue(queue *list.List, order Order) bool {
	for e := queue.Front(); e != nil; e = e.Next() {
		if e.Value.(Order).Cost > order.Cost {
			fmt.Println(e.Value.(Order).Cost)
			queue.InsertBefore(order, e)
			if e == queue.Front() {
				return true
			} else {
				return false
			}
		}

	}
}

// Dequeues first order
func dequeue(queue *list.List) {
	queue.Remove(queue.Front())
}

// Prints queue
func printQueue(queue *list.List) {
	var j int = 1
	for p := queue.Front(); p != nil; p = p.Next() {
		fmt.Println("Order nr:", j)
		fmt.Println(p.Value)
		j++
	}
	j = 0
}

func getNextOrder(queue *list.List) Order {
	return queue.Front().Value.(Order)

}

func Queue(OrderEnqueue <-chan Order, OrderDequeue <-chan Order, NextOrder chan<- Order) {

	//Queue Init
	queue := list.New()

	for {
		select {
		case newOrder := <-OrderEnqueue:
			insertedAtFront := enqueue(queue, newOrder)
			if insertedAtFront {
				NextOrder <- getNextOrder(queue)
			}

		case <-OrderDequeue:
			dequeue(queue)
			if queue.Len() != 0 {
				NextOrder <- getNextOrder(queue)
			}

		}

	}

	//OrderEnqueue := make(chan Order)
	//OrderDequeue := make(chan Order)

	// Test queue
	// e4 := queue.PushBack(Order{1, 10})
	// e1 := queue.PushFront(Order{3, 4})
	// queue.InsertBefore(Order{5, 8}, e4)
	// queue.InsertAfter(Order{7, 6}, e1)

	// New accepted order
	// var testOrder = Order{1, 2}

	// // Enqueue
	// enqueue(queue, testOrder)

	// // Print queue (after adding order)
	// printQueue(queue)

	// // Dequeue
	// dequeue(queue)

	// // Print queue (after removing order)
	// printQueue(queue)

}
