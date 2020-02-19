// Package driver is for interfacing with low level hardware
// and to handle switching lights and acting on orders.
package driver

import (
	"fmt"
	"log"

	"./elevio"
)

// OrderType is a typedef of int
type OrderType int

const (
	O_HallUp   OrderType = 0
	O_HallDown           = 1
	O_Cab                = 2
)

// Order is a struct with necessary information to execute an order.
type Order struct {
	TargetFloor int
	Type        OrderType
}
type elevState struct {
	order        Order
	currentFloor int
}

var drvButtons chan elevio.ButtonEvent
var drvFloors chan int
var drvObstr chan bool
var drvStop chan bool
var floorMonitorChan chan bool
var state elevState

// Initialized driver channels for low level communication
// and starts goroutines for polling hardware.
func driverInit() {
	elevio.Init("localhost:15657", 4) // TODO: CHANGE CHANGE CHANGE

	drvButtons = make(chan elevio.ButtonEvent)
	drvFloors = make(chan int)
	drvObstr = make(chan bool)
	drvStop = make(chan bool)
	floorMonitorChan = make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)
	go monitorFloor(floorMonitorChan)

	log.Printf("Driver initialized")
}

// Function for checking if at target floor.
func monitorFloor(floorMonitorChan <-chan bool) {
	var d elevio.MotorDirection
	for {
		select {
		case <-floorMonitorChan:
			fmt.Printf("current: %d		target: %d\n", state.currentFloor, state.order.TargetFloor)
			if state.currentFloor == state.order.TargetFloor {
				d = elevio.MD_Stop
				log.Println("Arrived at floor, stopping motor")

				elevio.SetButtonLamp(O_Cab, state.currentFloor, false)
				elevio.SetButtonLamp(elevio.ButtonType(state.order.Type), state.currentFloor, false)

			} else if state.currentFloor < state.order.TargetFloor {
				d = elevio.MD_Up
			} else {
				d = elevio.MD_Down
			}

			elevio.SetMotorDirection(d)
			log.Printf("Setting motor in direction %#v to get to target floor %d\n", d, state.order.TargetFloor)
			elevio.SetFloorIndicator(state.currentFloor)
		}
	}
}

// Driver is the main function of the package. It reads the low level channels
// and sends the information to a higher level.
func Driver(getOrderChan chan<- Order, execOrderChan <-chan Order) {
	driverInit()

	for {
		select {
		case btnEvent := <-drvButtons:
			order := Order{btnEvent.Floor, OrderType(btnEvent.Button)}
			getOrderChan <- order
			elevio.SetButtonLamp(btnEvent.Button, btnEvent.Floor, true) // turn on button lamp
			floorMonitorChan <- true                                    // Start monitorFloor

			log.Printf("Received button press: %#v\n", order)

		case newFloor := <-drvFloors:
			state.currentFloor = newFloor
			elevio.SetFloorIndicator(state.currentFloor) // Set floor indicator to current floor
			floorMonitorChan <- true                     // Start monitorFloor

			log.Printf("Arrived at new floor: %#v\n", state.currentFloor)

		case order := <-execOrderChan:
			state.order = order
			floorMonitorChan <- true //Start monitorFloor

			log.Printf("Received new order: %#v\n", order)
		}
	}
}
