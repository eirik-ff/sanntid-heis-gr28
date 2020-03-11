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


// Dequeues and returns first order
func dequeueFirst(queue *list.List) order.Order{

	//get first element and remove it from the queue
	e:= queue.Front()
	queue.Remove(e)
	
	return e.Value.(order.Order)
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
// * OrderEnqueue	: enqueues sent order
// * OrderDequeue	: if anything is sent on channel, dequeue/delete first element
// * NextOrder		: the order at the front of the queue
// * FlushQueue		: If true the entire queue will be sent out on NextOrder
func Queue(
	OrderEnqueue <-chan order.Order,
	OrderDequeue <-chan order.Order,
	NextOrder chan<- order.Order,
	FlushQueue <- chan bool) {

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

		case flush := <-FlushQueue:

			//If received flush == true, the whole queue will be flushed to getNextOrder
			if flush == true {
				queueLen := queue.Len()				//store length of queue before dequeing the queue
				for i := 0; i < (queueLen); i++ {
					
					NextOrder <-dequeueFirst(queue) //Remove one order and sent it to main
				}
			}
			log.Println("****Flushed queue****")
			printQueue(queue)
		}
	}
}
