package queue

import (
	"container/list"
	"fmt"
	"log"

	"../elevTypes/order"
)

// QueueOrder wraps an order.Order and its associated cost.
type QueueOrder struct {
	Order order.Order
	Cost  int
}

// Enqueues new order
//
// return true: inserted before at front
// return false: didn't insert before at front
func enqueue(queue *list.List, order QueueOrder) bool {
	for e := queue.Front(); e != nil; e = e.Next() {
		if e.Value.(QueueOrder).Cost > order.Cost {
			fmt.Println(e.Value.(QueueOrder).Cost)
			queue.InsertBefore(order, e)
			if e == queue.Front() {
				return true
			}
		} else if e == queue.Back() {
			queue.InsertAfter(order, e)
			return false
		}
	}

	// the list is empty
	queue.PushFront(order)
	return true
}

// Dequeues first order
func dequeue(queue *list.List) {
	front := queue.Front()

	if front == nil { // nil means empty queue
		return
	}
	log.Printf("Removed order(s) from queue: %#v\n", *front)

	toDelete := []*list.Element{front}

	for e := queue.Front().Next(); e != nil; e = e.Next() {
		if front.Value.(QueueOrder).Order.TargetFloor == e.Value.(QueueOrder).Order.TargetFloor {
			if e.Value.(QueueOrder).Order.Type == order.Cab ||
				front.Value.(QueueOrder).Order.Type == e.Value.(QueueOrder).Order.Type {
				// queue.Remove(e)
				toDelete = append(toDelete, e)
			}
		}
	}

	for _, e := range toDelete {
		queue.Remove(e)
	}
}

// Prints queue
func printQueue(queue *list.List) {
	log.Println("********** QUEUEUEUEUUEUE *********")
	var j int = 1
	for p := queue.Front(); p != nil; p = p.Next() {
		log.Printf("\tQueueOrder nr %d: %#v\n", j, p.Value)
		j++
	}
	log.Println("******* QUEUUE FINISHED *******")
}

func getNextOrder(queue *list.List) order.Order {
	fmt.Println("Running getNextOrder")
	return queue.Front().Value.(QueueOrder).Order

}

// Queue listens for incoming orders or signals on the channels and acts accordingly.
// * OrderEnqueue: enqueues sent order
// * OrderDequeue: if anything is sent on channel, dequeue/delete first element
// * NextOrder: the order at the front of the queue
func Queue(
	OrderEnqueue <-chan QueueOrder,
	OrderDequeue <-chan bool,
	NextOrder chan<- order.Order) {

	//Queue Init
	queue := list.New()

	for {
		select {
		case newOrder := <-OrderEnqueue:
			insertedAtFront := enqueue(queue, newOrder)
			log.Printf("Add order to queue: %#v\n", newOrder)
			printQueue(queue)

			if insertedAtFront {
				NextOrder <- getNextOrder(queue)
			}

		case <-OrderDequeue:
			dequeue(queue)
			if queue.Len() > 0 {
				NextOrder <- getNextOrder(queue)
			}

			printQueue(queue)
		}
	}
}
