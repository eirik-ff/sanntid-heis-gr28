package to_be_implemented

import (
	"fmt"
	"time"

	"../elevTypes/order"
)

//Uncertain where to put this in main so it is placed here in its own file

//Maximum waiting time for elevator to execute order
const (
	maxExecutionTime time.Duration = 30 * time.Second
)

//This was just used for testing
func sleeping() {
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		fmt.Println("sleeping.....")
	}
}

// Checks for how long the elevator has had status = Taken on an order
// and compares it with the maximum waiting time
func checkOrderTimestamp(ord order.Order, c chan int) {

	for {
		currentTime := time.Now()
		executionTimer := currentTime.Sub(ord.Timestamp)

		if executionTimer > maxExecutionTime {
			println("Did not complete within maximum execution time.")
			c <- 1
		}
	}

}

//Simple test
func test() {

	ord1 := order.Order{Floor: 1, Type: 2, Timestamp: time.Now()}

	ord1.Status = order.Taken

	c := make(chan int)

	fmt.Printf("Status of order is %v. \n", ord1.Status)

	go checkOrderTimestamp(ord1, c)
	sleeping()

	x := <-c

	if x == 1 {
		ord1.Status = order.NotTaken
	}

	fmt.Printf("Status of order is %v. \n", ord1.Status)

}
