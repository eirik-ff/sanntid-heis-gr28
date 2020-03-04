package queue

import (
	"container/list"
	"log"

	"../elevTypes/order"
)

// Enqueues new order
//
// return true: inserted before at front
// return false: didn't insert before at front
func enqueue(queue *list.List, ord order.Order) bool {
	for e := queue.Front(); e != nil; e = e.Next() {
		if e.Value.(order.Order).Type == ord.Type &&
			e.Value.(order.Order).TargetFloor == ord.TargetFloor {
			return false
		}
	}

	for e := queue.Front(); e != nil; e = e.Next() {
		if e.Value.(order.Order).Cost > ord.Cost {
			// fmt.Println(e.Value.(order.Order).Cost)
			queue.InsertBefore(ord, e)
			if e == queue.Front() {
				return true
			}
		} else if e == queue.Back() {
			queue.InsertAfter(ord, e)
			return false
		}
	}

	// the list is empty
	queue.PushFront(ord)
	return true
}

// Dequeues first order
func dequeue(queue *list.List, ord order.Order) {
	toDelete := []*list.Element{}

	for e := queue.Front(); e != nil; e = e.Next() {
		if ord.Status == order.LowerCostReply && e.Value.(order.Order).ID == ord.ID {
			// Another elevator have better cost, remove your entry in the queue so
			// not both elevators execute the same order.
			toDelete = append(toDelete, e)
		} else if ord.Status == order.Finished {
			if e.Value.(order.Order).TargetFloor == ord.TargetFloor &&
				(e.Value.(order.Order).Type == order.Cab || ord.Type == e.Value.(order.Order).Type) {
				// Order is executed locally and is finished
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
	return queue.Front().Value.(order.Order)

}

// Queue listens for incoming orders or signals on the channels and acts accordingly.
// * OrderEnqueue: enqueues sent order
// * OrderDequeue: if anything is sent on channel, dequeue/delete first element
// * NextOrder: the order at the front of the queue
func Queue(
	OrderEnqueue <-chan order.Order,
	OrderDequeue <-chan order.Order,
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

		case ord := <-OrderDequeue:
			dequeue(queue, ord)
			if queue.Len() > 0 {
				NextOrder <- getNextOrder(queue)
			} else { // order from network had better cost, abort current active order
				ord.Status = order.Abort
				NextOrder <- ord
			}

			printQueue(queue)
		}
	}
}
