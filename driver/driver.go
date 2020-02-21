// Package driver is for interfacing with low level hardware
// and to handle switching lights and acting on orders.
package driver

import (
	"log"
	"time"

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

const floorChangeTimeout time.Duration = 3 * time.Second // TODO: Measure suitable value for floorChangeTimeout

var drvButtons chan elevio.ButtonEvent
var drvFloors chan int
var drvObstr chan bool
var drvStop chan bool
var floorMonitorChan chan elevState

// Initialized driver channels for low level communication
// and starts goroutines for polling hardware.
func driverInit() {
	elevio.Init("localhost:15657", 4) // TODO: CHANGE CHANGE CHANGE

	drvButtons = make(chan elevio.ButtonEvent)
	drvFloors = make(chan int)
	drvObstr = make(chan bool)
	drvStop = make(chan bool)
	floorMonitorChan = make(chan elevState)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)
	go monitorFloor(floorMonitorChan)

	log.Printf("Driver initialized")
}

// Function for checking if at target floor.
func monitorFloor(floorMonitorChan <-chan elevState) {
	var d elevio.MotorDirection
	floorChangeTimer := time.NewTimer(floorChangeTimeout)
	floorChangeTimer.Stop()
	for {
		select {
		case state := <-floorMonitorChan:
			if state.currentFloor == state.order.TargetFloor {
				d = elevio.MD_Stop
				log.Println("Arrived at floor, stopping motor")
				floorChangeTimer.Stop()

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

			if d != elevio.MD_Stop {
				floorChangeTimer.Reset(floorChangeTimeout)
			}

		case <-floorChangeTimer.C:
			log.Println("floorChangeTimer timed out")
			// TODO: tell someone else about this (report an error or something)
		}
	}
}

// Driver is the main function of the package. It reads the low level channels
// and sends the information to a higher level.
func Driver(getOrderChan chan<- Order, execOrderChan <-chan Order) {
	driverInit()
	var state elevState

	for {
		select {
		case btnEvent := <-drvButtons:
			order := Order{btnEvent.Floor, OrderType(btnEvent.Button)}
			getOrderChan <- order
			elevio.SetButtonLamp(btnEvent.Button, btnEvent.Floor, true) // turn on button lamp
			floorMonitorChan <- state                                   // Start monitorFloor

			log.Printf("Received button press: %#v\n", order)

		case newFloor := <-drvFloors:
			state.currentFloor = newFloor
			elevio.SetFloorIndicator(state.currentFloor) // Set floor indicator to current floor
			floorMonitorChan <- state                    // Start monitorFloor

			log.Printf("Arrived at new floor: %#v\n", state.currentFloor)

		case order := <-execOrderChan:
			state.order = order
			floorMonitorChan <- state // Start monitorFloor

			log.Printf("Received new order: %#v\n", order)
		}
	}
}
