// Package driver is for interfacing with low level hardware
// and to handle switching lights and acting on orders.
package driver

import (
	"log"
	"time"

	"./elevio"
)

// OrderType is a typedef of int
type OrderType elevio.ButtonType

const (
	O_HallUp   = OrderType(elevio.BT_HallUp)
	O_HallDown = OrderType(elevio.BT_HallDown)
	O_Cab      = OrderType(elevio.BT_Cab)
)

// MotorDirection is a typedef of elevio.MotorDirection to be able
// to use it in packages that include driver.
type MotorDirection elevio.MotorDirection

const (
	MD_Up   = MotorDirection(elevio.MD_Up)
	MD_Down = MotorDirection(elevio.MD_Down)
	MD_Stop = MotorDirection(elevio.MD_Stop)
)

// Order is a struct with necessary information to execute an order.
type Order struct {
	TargetFloor int
	Type        OrderType
	ForMe       bool // default: false
}

// ElevState is a struct with the current position and active order of
// the elevator.
type ElevState struct {
	Order        Order
	CurrentFloor int
	Direction    MotorDirection
}

const (
	floorChangeTimeout time.Duration = 3 * time.Second // TODO: Measure suitable value for floorChangeTimeout
	doorTimeout        time.Duration = 3 * time.Second
)

var drvButtons chan elevio.ButtonEvent
var drvFloors chan int
var drvObstr chan bool
var drvStop chan bool
var floorMonitorChan chan ElevState

// Initialized driver channels for low level communication
// and starts goroutines for polling hardware.
func driverInit() {
	elevio.Init("localhost:15657", 4) // TODO: CHANGE CHANGE CHANGE

	drvButtons = make(chan elevio.ButtonEvent)
	drvFloors = make(chan int)
	drvObstr = make(chan bool)
	drvStop = make(chan bool)
	floorMonitorChan = make(chan ElevState)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)
	go monitorFloor(floorMonitorChan)

	log.Printf("Driver initialized")
}

// Function for checking if at target floor.
func monitorFloor(floorMonitorChan <-chan ElevState) {
	var d elevio.MotorDirection
	floorChangeTimer := time.NewTimer(floorChangeTimeout)
	floorChangeTimer.Stop()
	for {
		select {
		case state := <-floorMonitorChan:
			if state.CurrentFloor == state.Order.TargetFloor {
				d = elevio.MD_Stop
				log.Println("Arrived at floor, stopping motor")
				floorChangeTimer.Stop()

				elevio.SetButtonLamp(elevio.ButtonType(O_Cab), state.CurrentFloor, false)
				elevio.SetButtonLamp(elevio.ButtonType(state.Order.Type), state.CurrentFloor, false)

			} else if state.CurrentFloor < state.Order.TargetFloor {
				d = elevio.MD_Up
			} else {
				d = elevio.MD_Down
			}

			elevio.SetMotorDirection(d)
			log.Printf("Setting motor in direction %#v to get to target floor %d\n", d, state.Order.TargetFloor)
			elevio.SetFloorIndicator(state.CurrentFloor)

			if d != elevio.MD_Stop {
				floorChangeTimer.Stop()
				floorChangeTimer.Reset(floorChangeTimeout)
			} else {
				elevio.SetDoorOpenLamp(true)
				<-time.After(doorTimeout)
				elevio.SetDoorOpenLamp(false)
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
	var state ElevState

	for {
		select {
		case btnEvent := <-drvButtons:
			order := Order{
				TargetFloor: btnEvent.Floor,
				Type:        OrderType(btnEvent.Button),
			}
			getOrderChan <- order
			// floorMonitorChan <- state                                   // Start monitorFloor

			log.Printf("Received button press: %#v\n", order)

		case newFloor := <-drvFloors:
			state.CurrentFloor = newFloor
			elevio.SetFloorIndicator(state.CurrentFloor) // Set floor indicator to current floor
			floorMonitorChan <- state                    // Start monitorFloor

			log.Printf("Arrived at new floor: %#v\n", state.CurrentFloor)

		case order := <-execOrderChan:
			elevio.SetButtonLamp(elevio.ButtonType(order.Type), order.TargetFloor, true) // turn on button lamp

			if order.ForMe {
				state.Order = order
				floorMonitorChan <- state // Start monitorFloor
				log.Printf("Received new order for ME: %#v\n", order)
			} else {
				log.Println("Received new order NOT for me")
			}
		}
	}
}
