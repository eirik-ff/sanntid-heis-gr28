// Package driver is for interfacing with low level hardware
// and to handle switching lights and acting on orders.
package driver

import (
	"log"
	"time"

	"../elevTypes/order"
	"./elevio"
)

// MotorDirection is a typedef of elevio.MotorDirection to be able
// to use it in packages that include driver.
type MotorDirection elevio.MotorDirection

const (
	MD_Up   = MotorDirection(elevio.MD_Up)
	MD_Down = MotorDirection(elevio.MD_Down)
	MD_Stop = MotorDirection(elevio.MD_Stop)
)

// ElevState is a struct with the current position and active order of
// the elevator.
type ElevState struct {
	Order        order.Order
	CurrentFloor int
	Direction    MotorDirection
}

const (
	floorChangeTimeout time.Duration = 5 * time.Second // TODO: Measure suitable value for floorChangeTimeout
	doorTimeout        time.Duration = 3 * time.Second
)

var (
	drvButtons chan elevio.ButtonEvent
	drvFloors  chan int
	drvObstr   chan bool
	drvStop    chan bool

	floorMonitorChan chan ElevState
	updatedStateChan chan ElevState
)

// Initialized driver channels for low level communication
// and starts goroutines for polling hardware.
func driverInit() {
	elevio.Init("localhost:15657", 4) // TODO: CHANGE CHANGE CHANGE

	drvButtons = make(chan elevio.ButtonEvent, 10)
	drvFloors = make(chan int)
	drvObstr = make(chan bool)
	drvStop = make(chan bool)
	floorMonitorChan = make(chan ElevState)
	updatedStateChan = make(chan ElevState, 1) // This have caused problems if unbuffered

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)
	go monitorFloor(floorMonitorChan, updatedStateChan)

	log.Printf("Driver initialized")
}

// Function for checking if at target floor.
func monitorFloor(floorMonitorChan <-chan ElevState, stateChan chan<- ElevState) {
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

				elevio.SetButtonLamp(order.Cab, state.CurrentFloor, false)
				elevio.SetButtonLamp(elevio.ButtonType(state.Order.Type), state.CurrentFloor, false)

				state.Order.Finished = true

			} else if state.CurrentFloor < state.Order.TargetFloor {
				d = elevio.MD_Up
			} else {
				d = elevio.MD_Down
			}

			elevio.SetMotorDirection(d)
			log.Printf("Setting motor in direction %#v to get to target floor %d\n", d, state.Order.TargetFloor)
			elevio.SetFloorIndicator(state.CurrentFloor)

			state.Direction = MotorDirection(d)
			stateChan <- state // this blocks if channel is unbuffered

			if d != elevio.MD_Stop {
				floorChangeTimer.Stop()
				floorChangeTimer.Reset(floorChangeTimeout)
			} else {
				elevio.SetDoorOpenLamp(true)
				log.Println("Opening door")
				// <-time.After(doorTimeout)

				// for now, door is not active as it causes some head ache. if more
				// than one order comes when the door is open, they do no light up
				// before the door closes again (given that drvButtons are buffered)
				// this must be fixed later by maybe separating motor and light
				// activation code, but for now, we just ignore it and handle it
				// later
				// TODO: OBS, dette må vi gjøre noe med! :O

				elevio.SetDoorOpenLamp(false)
				log.Println("Closing door")
			}

		case <-floorChangeTimer.C:
			log.Println("floorChangeTimer timed out")
			// TODO: tell someone else about this (report an error or something)
		}
	}
}

// Driver is the main function of the package. It reads the low level channels
// and sends the information to a higher level.
func Driver(
	getOrderChan chan<- order.Order,
	execOrderChan <-chan order.Order,
	externalStateChan chan<- ElevState) {

	driverInit()
	var state ElevState

	for {
		select {
		case btnEvent := <-drvButtons:
			o := order.Order{
				TargetFloor: btnEvent.Floor,
				Type:        order.Type(btnEvent.Button),
			}
			getOrderChan <- o
			elevio.SetButtonLamp(btnEvent.Button, btnEvent.Floor, true) // turn on button lamp

			log.Printf("Received button press: %#v\n", o)

		case newFloor := <-drvFloors:
			state.CurrentFloor = newFloor
			elevio.SetFloorIndicator(state.CurrentFloor) // Set floor indicator to current floor
			floorMonitorChan <- state                    // Start monitorFloor

			log.Printf("Arrived at new floor: %#v\n", state.CurrentFloor)

		case o := <-execOrderChan:
			elevio.SetButtonLamp(elevio.ButtonType(o.Type), o.TargetFloor, true) // turn on button lamp

			if o.ForMe {
				state.Order = o
				floorMonitorChan <- state // Start monitorFloor
				log.Printf("Received new order for ME: %#v\n", o)
			} else {
				log.Println("Received new order NOT for me")
			}

			state.Order = o
			floorMonitorChan <- state // Start monitorFloor

			log.Printf("Received new order: %#v\n", o)

		case state = <-updatedStateChan:
			externalStateChan <- state
		}
	}
}
