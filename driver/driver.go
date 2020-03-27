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
	floorChangeTimeout time.Duration = 5 * time.Second
	doorTimeout        time.Duration = 3 * time.Second
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
	switch ord.Status {
	case order.Taken:
		ord.LocalTimeStamp = time.Now().Unix() + order.OrderTimeout

	case order.Execute:
		ord.LocalTimeStamp = time.Now().Unix() + order.OrderTimeout
		if elev.ActiveOrder.Status != order.Finished && !order.CompareEq(ord, elev.ActiveOrder) {
			oldActive := elev.ActiveOrder
			oldActive.Status = order.NotTaken
			elev.AssignOrderToMatrix(oldActive)
		}

		ord.Status = order.Taken
		elev.ActiveOrder = ord
	}
	elev.AssignOrderToMatrix(ord)

	return elev, true
}

// can only happen in state elevator.Moving
func floorChange(elev elevator.Elevator, newFloor int,
	motorTimer, doorTimer *time.Timer) (elevator.Elevator, bool) {
	elevio.SetFloorIndicator(newFloor)
	motorTimer.Reset(floorChangeTimeout)

	elev.Floor = newFloor
	if newFloor == elev.ActiveOrder.Floor {
		if elev.ActiveOrder.Type != order.Cab {
			elev.Orders[elev.ActiveOrder.Floor][order.Cab].Status = order.Finished
		}
		elev, _ = arrivedAtTarget(elev, motorTimer, doorTimer)
	} else {
		elev.State = elevator.Moving
	}
	return elev, true
}

func arrivedAtTarget(
	elev elevator.Elevator,
	motorTimer, doorTimer *time.Timer) (elevator.Elevator, bool) {
	log.Println("Arrived at target floor.")

	elevio.SetMotorDirection(elevio.MD_Stop)
	motorTimer.Stop()
	elev.Direction = elevator.Stop

	elev.ActiveOrder.Status = order.Finished
	elev.Orders[elev.ActiveOrder.Floor][elev.ActiveOrder.Type].Status = order.Finished

	elevio.SetDoorOpenLamp(true)
	doorTimer.Reset(doorTimeout)
	elev.State = elevator.DoorOpen
	log.Println("Door opening")

	return elev, true
}

func buttonPress(
	elev elevator.Elevator,
	press elevio.ButtonEvent) (elevator.Elevator, bool, order.Order) {
	f := press.Floor
	t := order.Type(press.Button)
	o := order.Order{Floor: f, Type: t, Status: order.NotTaken}
	elev.Orders[f][t] = o
	return elev, true, o
}

func doorClose(elev elevator.Elevator, doorTimer *time.Timer) (elevator.Elevator, bool) {
	doorTimer.Stop()
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

func setDirection(elev elevator.Elevator,
	motorTimer, doorTimer *time.Timer) (elevator.Elevator, bool) {
	var updateElev bool = false
	var d elevio.MotorDirection
	if elev.ActiveOrder.Floor > elev.Floor {
		d = elevio.MD_Up
	} else if elev.ActiveOrder.Floor < elev.Floor {
		d = elevio.MD_Down
	} else {
		elev, updateElev = arrivedAtTarget(elev, motorTimer, doorTimer)
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
func driverInit(port int, drvButtons chan elevio.ButtonEvent, drvFloors chan int) (*time.Timer, *time.Timer) {
	elevio.Init(fmt.Sprintf("localhost:%d", port), 4)
	motorTimer := time.NewTimer(floorChangeTimeout)
	doorTimer := time.NewTimer(doorTimeout)
	motorTimer.Stop()
	doorTimer.Stop()

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)

	return motorTimer, doorTimer
}

// Driver is the main function of the package. It reads the low level channels
// and sends the information to a higher level.
func Driver(
	elevIOport int,
	nfloors, nbuttons int,
	mainElevatorChan chan<- elevator.Elevator,
	orderChan <-chan order.Order,
	buttonPressChan chan<- order.Order,
	initElev elevator.Elevator) {
	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	motorTimer, doorTimer := driverInit(elevIOport, drvButtons, drvFloors)

	var elev elevator.Elevator = initElev
	mainElevatorChan <- elev
	elevio.SetMotorDirection(elevio.MotorDirection(elev.Direction))

	var updateElev bool = true
	for {
		select {
		case press := <-drvButtons:
			var o order.Order
			elev, updateElev, o = buttonPress(elev, press)
			buttonPressChan <- o

		case newFloor := <-drvFloors:
			elev, updateElev = floorChange(elev, newFloor, motorTimer, doorTimer)

		case o := <-orderChan:
			elev, updateElev = orderFromMain(elev, o)

		case <-doorTimer.C:
			elev, updateElev = doorClose(elev, doorTimer)

		case <-motorTimer.C:
			elev, updateElev = motorTimeout(elev)

		case <-time.After(1 * time.Millisecond):
			if updateElev {
				setLamps(elev)

				mainElevatorChan <- elev
				updateElev = false
			}

			switch elev.State {
			case elevator.Idle:
				if elev.ActiveOrder.Status == order.Taken {
					// will come into effect at next iteration
					elev.State = elevator.Moving
					updateElev = true
				}
			case elevator.Moving:
				elev, updateElev = setDirection(elev, motorTimer, doorTimer)
			case elevator.DoorOpen:
				// do nothing, everything happens in transition/on events
			case elevator.Error:
				// do nothing, will exit when motors start working again
			}
		}
	}
}
