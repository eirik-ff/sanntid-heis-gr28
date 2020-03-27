package driver

import (
	"fmt"
	"log"
	"time"

	"../elevTypes/elevator"
	"../elevTypes/order"
	"./elevio"
)

const (
	floorChangeTimeout time.Duration = 5 * time.Second // TODO: Measure suitable value for floorChangeTimeout
	doorTimeout        time.Duration = 3 * time.Second
)

var ( // TODO: look at making these local in Driver
	doorTimer  *time.Timer
	motorTimer *time.Timer

	Nfloors  int
	Nbuttons int
)

func setLamps(elev elevator.Elevator) {
	for i := range elev.Orders {
		for j := range elev.Orders[i] {
			set := false
			status := elev.Orders[i][j].Status
			if status == order.NotTaken ||
				status == order.Taken ||
				status == order.Execute {
				set = true
			}
			elevio.SetButtonLamp(elevio.ButtonType(j), i, set)
		}
	}
}

func orderFromMain(elev elevator.Elevator, ord order.Order) (elevator.Elevator, bool) {
	log.Printf("New order from main: %s\n", ord.ToString())

	switch ord.Status {
	case order.Abort:
		// shouldn't happen

	case order.Invalid:
		// shouldn't happen

	case order.Finished:
		// this is when you get an order from the network telling you that the
		// received order is finished and you should not care about it anymore
		elev.AssignOrderToMatrix(ord)

	case order.NotTaken:
		// this is when you get a message from the network
		elev.AssignOrderToMatrix(ord)

	case order.Taken:
		// this is when you get a message from the network saying that another
		// elevator is taking this specific order

		// if elev.Orders[ord.Floor][ord.Type].Status != order.Taken {
		ord.LocalTimeStamp = time.Now().Unix() + order.OrderTimeout
		elev.AssignOrderToMatrix(ord)
		// }

	case order.Execute:
		// this is when main tells you that this is the order you should execute
		// now
		ord.LocalTimeStamp = time.Now().Unix() + order.OrderTimeout
		if elev.ActiveOrder.Status != order.Finished && !order.CompareEq(ord, elev.ActiveOrder) {
			// new order, set old to NotTaken
			elev.ActiveOrder.Status = order.NotTaken
			elev.AssignOrderToMatrix(elev.ActiveOrder)

			log.Printf("Reset old active order: %s\n", elev.ActiveOrder.ToString())
		}
		// Set new active order
		ord.Status = order.Taken
		elev.ActiveOrder = ord
		elev.AssignOrderToMatrix(ord)
	}

	return elev, true
}

func arrivedAtTarget(elev elevator.Elevator) (elevator.Elevator, bool) {
	log.Println("Arrived at target floor")

	elevio.SetMotorDirection(elevio.MD_Stop)
	motorTimer.Stop() // TODO: look into motor timer
	elev.Direction = elevator.Stop

	elev.ActiveOrder.Status = order.Finished
	elev.Orders[elev.ActiveOrder.Floor][elev.ActiveOrder.Type].Status = order.Finished

	elevio.SetDoorOpenLamp(true)
	doorTimer.Reset(doorTimeout) // TODO: look into door timer
	elev.State = elevator.DoorOpen
	log.Println("Door opening")

	return elev, true
}

// can only happen in state elevator.Moving
func floorChange(elev elevator.Elevator, newFloor int) (elevator.Elevator, bool) {
	elevio.SetFloorIndicator(newFloor)
	motorTimer.Reset(floorChangeTimeout) // TODO: look into motor timer

	elev.Floor = newFloor
	log.Printf("New floor: %d\n", newFloor)

	if newFloor == elev.ActiveOrder.Floor {

		if elev.ActiveOrder.Type != order.Cab {
			elev.Orders[elev.ActiveOrder.Floor][order.Cab].Status = order.Finished
		}

		elev, _ = arrivedAtTarget(elev)

	} else {
		elev.State = elevator.Moving
	}
	return elev, true
}

func buttonPress(elev elevator.Elevator, press elevio.ButtonEvent) (elevator.Elevator, bool, order.Order) {
	f := press.Floor
	t := order.Type(press.Button)
	o := order.Order{Floor: f, Type: t, Status: order.NotTaken}
	elev.Orders[f][t] = o

	log.Printf("Button press: %s\n", o.ToString())

	return elev, true, o
}

// can only happen in state elevator.DoorOpen
func doorClose(elev elevator.Elevator) (elevator.Elevator, bool) {
	doorTimer.Stop() // TODO: look into door timer
	elevio.SetDoorOpenLamp(false)
	elev.State = elevator.Idle

	log.Println("Door closing")

	return elev, true
}

func motorTimeout(elev elevator.Elevator) (elevator.Elevator, bool) {
	log.Println("Motor timed out!!")

	elev.State = elevator.Error

	return elev, true
}

func setDirection(elev elevator.Elevator) (elevator.Elevator, bool) {
	var updateElev bool = false
	var d elevio.MotorDirection
	if elev.ActiveOrder.Floor > elev.Floor {
		d = elevio.MD_Up
	} else if elev.ActiveOrder.Floor < elev.Floor {
		d = elevio.MD_Down
	} else {
		elev, updateElev = arrivedAtTarget(elev)
	}

	if elev.Direction != elevator.Direction(d) {
		elevio.SetMotorDirection(d)
		motorTimer.Reset(floorChangeTimeout)

		elev.Direction = elevator.Direction(d)
		updateElev = true
	}

	return elev, updateElev
}

// Initialized driver channels for low level communication
// and starts goroutines for polling hardware.
func driverInit(port int, drvButtons chan elevio.ButtonEvent, drvFloors chan int) {
	elevio.Init(fmt.Sprintf("localhost:%d", port), 4)

	motorTimer = time.NewTimer(floorChangeTimeout)
	doorTimer = time.NewTimer(doorTimeout)
	motorTimer.Stop()
	doorTimer.Stop()

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)

	log.Printf("Driver initialized")
}

// Driver is the main function of the package. It reads the low level channels
// and sends the information to a higher level.
// TODO: re-write this
func Driver(port int, nfloors, nbuttons int, mainElevatorChan chan<- elevator.Elevator,
	orderChan <-chan order.Order, buttonPressChan chan<- order.Order, initElev elevator.Elevator) {

	Nfloors = nfloors
	Nbuttons = nbuttons

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	driverInit(port, drvButtons, drvFloors)

	var elev elevator.Elevator = initElev
	//	setLamps(elev)

	mainElevatorChan <- elev // to not make it crash on default in main

	elevio.SetMotorDirection(elevio.MotorDirection(elev.Direction))

	var updateElev bool = true
	for {
		// Capture events
		select {
		case press := <-drvButtons:
			var o order.Order
			elev, updateElev, o = buttonPress(elev, press)
			buttonPressChan <- o
			log.Println("elev update from drvButton")

		case newFloor := <-drvFloors:
			elev, updateElev = floorChange(elev, newFloor)
			log.Println("elev update from drvFloors")

		case o := <-orderChan:
			elev, updateElev = orderFromMain(elev, o)
			log.Println("elev update from orderChan")

		case <-doorTimer.C:
			elev, updateElev = doorClose(elev)
			log.Println("elev update from doorTimer")

		case <-motorTimer.C:
			elev, updateElev = motorTimeout(elev)
			log.Println("elev update from motorTimer")

		case <-time.After(1 * time.Millisecond):
			// Send new elevator object to main
			if updateElev {
				setLamps(elev)

				mainElevatorChan <- elev
				updateElev = false
			}

			// Act according to new state
			switch elev.State {
			case elevator.Idle:
				if elev.ActiveOrder.Status == order.Taken {
					// will come into effect at next iteration
					elev.State = elevator.Moving
					updateElev = true
					log.Println("elev update from idle")
				}

			case elevator.Moving:
				elev, updateElev = setDirection(elev)

			case elevator.DoorOpen:
				// do nothing, everything happens in transition/on events
			case elevator.Error:
				fallthrough
			default: // unknown state
				// TODO: something that happend should not have happened, send
				// 		 error to main
			}
		}

	}
}
