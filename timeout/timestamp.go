package timeout

import (
	"fmt"
	"time"

	"../elevTypes/order"
)

//Uncertain where to put this in main so it is placed here in its own file

//Maximum waiting time for elevator to execute order
const (
	maxExecutionTime time.Duration = 5 * time.Second
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
//
//Param ordChan  - order to check
//      timeoutchan - sends the order that timedout
func CheckOrderTimestamp(ordChan <- chan order.Order , timeoutChan chan <- order.Order) {

	for {
		select {
		case ord := <-ordChan:
			currentTime := time.Now()
			executionTimer := currentTime.Sub(ord.LocalTimeStamp)

			if executionTimer > maxExecutionTime {
				fmt.Println("Order timed out.") 
				timeoutChan <- ord
			}
		}
	}
}

//Simple test
// func main() {

// 	ord1 := order.Order{Floor: 1, Type: 2, Timestamp: time.Now()}

// 	ord1.Status = order.Taken

// 	ordChan		:= make(chan order.Order)
// 	timeoutChan := make(chan order.Order)

// 	fmt.Printf("Status of order is %v. \n", ord1.Status)

// 	go CheckOrderTimestamp(ordChan, timeoutChan)


// 	func (){
// 		for {
// 			//Send order for checking
// 			ordChan <- ord1

// 			sleeping()
// 		}
// 	}()

// 	for{
// 		select{
// 			case timeoutOrd := <-timeoutChan:

// 			fmt.Printf("Timed out order %v. \n", timeoutOrd)
// 		}
// 	}
// }
