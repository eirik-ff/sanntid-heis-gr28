package driver

import (
	"fmt"
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

const (
	floorChangeTimeout time.Duration = 5 * time.Second // TODO: Measure suitable value for floorChangeTimeout
	doorTimeout        time.Duration = 3 * time.Second
)

var ( // TODO: look at making these local in Driver
	doorTimer  *time.Timer
	motorTimer *time.Timer
)

type State int

const (
	Init     State = 0
	Idle     State = 1
	Moving   State = 2
	DoorOpen State = 3
	Error    State = 4
)

type Elevator struct {
	ActiveOrder order.Order
	Floor       int
	Direction   MotorDirection
	State       State
}

func orderFromMain(elev Elevator, ord order.Order) Elevator {
	log.Printf("New order from main: %#v\n", ord)

	if ord.Status == order.LightChange {
		elevio.SetButtonLamp(elevio.ButtonType(ord.Type), ord.TargetFloor, true)
		log.Printf("Set light on order with status LightChange. Floor %d Type %d\n", ord.TargetFloor, ord.Type)
		return elev // no change in state, but still needs to return something
	} else if ord.Status == order.Finished {
		// another elevator on the network has finished an order
		elevio.SetButtonLamp(elevio.ButtonType(ord.Type), ord.TargetFloor, false)
		log.Printf("Clear light on order with status Finished. Floor %d Type %d\n", ord.TargetFloor, ord.Type)
		return elev
	} else if ord.Status == order.Abort {
		elevio.SetMotorDirection(elevio.MD_Stop)
		elev.Direction = MD_Stop
		elev.ActiveOrder = ord
		elev.State = Idle
		return elev
	}

	elev.ActiveOrder = ord
	log.Printf("Set new active order with ID: %d\n", ord.ID)
	if elev.State == Idle {
		elev.State = Moving
	}

	return elev
}

func arrivedAtTarget(elev Elevator) Elevator {
	elevio.SetMotorDirection(elevio.MD_Stop)
	motorTimer.Stop() // TODO: look into motor timer
	elev.Direction = MD_Stop

	log.Println("Arrived at target floor")

	elev.ActiveOrder.Status = order.Finished
	elevio.SetButtonLamp(elevio.ButtonType(elev.ActiveOrder.Type),
		elev.ActiveOrder.TargetFloor, false)
	elevio.SetButtonLamp(elevio.BT_Cab, elev.ActiveOrder.TargetFloor, false)

	elevio.SetDoorOpenLamp(true)
	doorTimer.Reset(doorTimeout) // TODO: look into door timer
	elev.State = DoorOpen

	return elev
}

// can only happen in state Moving
func floorChange(elev Elevator, newFloor int) Elevator {
	elevio.SetFloorIndicator(newFloor)
	motorTimer.Reset(floorChangeTimeout) // TODO: look into motor timer

	elev.Floor = newFloor
	log.Printf("New floor: %d\n", newFloor)

	if newFloor == elev.ActiveOrder.TargetFloor {
		elev = arrivedAtTarget(elev)
	} else {
		elev.State = Moving
	}
	return elev
}

// state independent
func buttonPress(press elevio.ButtonEvent) order.Order {
	elevio.SetButtonLamp(press.Button, press.Floor, true)

	o := order.NewOrder(order.Type(press.Button), press.Floor, order.InitialBroadcast)

	log.Printf("Button press: %#v\n", o)

	return o
}

// can only happen in state DoorOpen
func doorClose(elev Elevator) Elevator {
	doorTimer.Stop() // TODO: look into door timer
	elevio.SetDoorOpenLamp(false)
	elev.State = Idle

	log.Println("Door closing")

	return elev
}

func motorTimeout(elev Elevator) Elevator {
	log.Println("Motor timed out!!")

	elevio.SetMotorDirection(elevio.MD_Stop)
	elev.State = Error

	return elev
}

func setDirection(elev Elevator) (Elevator, bool) {
	var updateElev bool = false
	var d elevio.MotorDirection
	if elev.ActiveOrder.TargetFloor > elev.Floor {
		d = elevio.MD_Up
	} else if elev.ActiveOrder.TargetFloor < elev.Floor {
		d = elevio.MD_Down
	} else {
		elev = arrivedAtTarget(elev)
		updateElev = true
	}

	if elev.Direction != MotorDirection(d) {
		elevio.SetMotorDirection(d)

		elev.Direction = MotorDirection(d)
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
func Driver(port int, mainElevatorChan chan<- Elevator,
	execOrderChan <-chan order.Order, buttonPressChan chan<- order.Order) {

	var elev Elevator
	elev.State = Init

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	driverInit(port, drvButtons, drvFloors)

	// ButtonPress creates an order and sends it to main
	go func() {
		for {
			press := <-drvButtons
			o := buttonPress(press)
			buttonPressChan <- o
		}
	}()

	elev.State = Idle
	var updateElev bool = false
	for {
		// Capture events
		select {
		case newFloor := <-drvFloors:
			elev = floorChange(elev, newFloor)
			updateElev = true

		case o := <-execOrderChan:
			elev = orderFromMain(elev, o)
			updateElev = true

		case <-doorTimer.C:
			elev = doorClose(elev)
			updateElev = true

		case <-motorTimer.C:
			elev = motorTimeout(elev)
			updateElev = true

		default:
			// do nothing
		}

		// Send new elevator object to main
		if updateElev {
			log.Printf("Elevator object update: %#v\n", elev)

			mainElevatorChan <- elev
			updateElev = false
		}

		// Act according to new state
		switch elev.State {
		case Idle:
			if !(elev.ActiveOrder.Status == order.Finished || elev.ActiveOrder.Status == order.Abort) {
				// log.Println("In IDLE - active order not finished - going to moving")
				// will come into effect at next iteration
				elev.State = Moving
				updateElev = true
			}
		case Moving:
			// log.Println("In MOVING")
			// log.Println("Active order:  ", elev.ActiveOrder)
			elev, updateElev = setDirection(elev)

		case DoorOpen:
			// do nothing, everything happens in transition/on events
		case Error:
			fallthrough
		default: // unknown state
			// TODO: something that happend should not have happened, send
			// 		 error to main
		}
	}
}
