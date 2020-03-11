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

	Nfloors  int
	Nbuttons int
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
	Orders      [][]order.Order
	// TODO: bounds check on index when accessing? if two elevators have
	// 		 different number of floors this will be necessary.
	// 		 maybe need bound check to be fault tolerant?
}

func setLamps(elev Elevator) {
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

func orderFromMain(elev Elevator, ord order.Order) (Elevator, bool) {
	log.Printf("New order from main: %#v\n", ord)

	updateElev := false
	if ord.Status == order.InitialBroadcast {
		ord.Status = order.NotTaken
		updateElev = true
	} else if ord.Status == order.Execute {
		elev.ActiveOrder = ord
		if elev.State == Idle {
			elev.State = Moving
		}
		updateElev = true
	} else if ord.Status == order.Finished {
		updateElev = false
	}
	// else just keep status

	if updateElev {
		elev.Orders[ord.Floor][ord.Type] = ord
	}
	return elev, updateElev
}

func arrivedAtTarget(elev Elevator) (Elevator, bool) {
	log.Println("Arrived at target floor")

	elevio.SetMotorDirection(elevio.MD_Stop)
	motorTimer.Stop() // TODO: look into motor timer
	elev.Direction = MD_Stop

	elev.ActiveOrder.Status = order.Finished
	elev.Orders[elev.ActiveOrder.Floor][elev.ActiveOrder.Type].Status = order.Finished

	elevio.SetDoorOpenLamp(true)
	doorTimer.Reset(doorTimeout) // TODO: look into door timer
	elev.State = DoorOpen
	log.Println("Door opening")

	return elev, true
}

// can only happen in state Moving
func floorChange(elev Elevator, newFloor int) (Elevator, bool) {
	elevio.SetFloorIndicator(newFloor)
	motorTimer.Reset(floorChangeTimeout) // TODO: look into motor timer

	elev.Floor = newFloor
	log.Printf("New floor: %d\n", newFloor)

	if newFloor == elev.ActiveOrder.Floor {
		elev, _ = arrivedAtTarget(elev)
	} else {
		elev.State = Moving
	}
	return elev, true
}

func buttonPress(elev Elevator, press elevio.ButtonEvent) (Elevator, bool, order.Order) {
	f := press.Floor
	t := int(press.Button)
	o := order.Order{Floor: f, Type: t, Status: order.NotTaken}
	elev.Orders[f][t] = o

	log.Printf("Button press: %#v\n", o)

	return elev, true, o
}

// can only happen in state DoorOpen
func doorClose(elev Elevator) (Elevator, bool) {
	doorTimer.Stop() // TODO: look into door timer
	elevio.SetDoorOpenLamp(false)
	elev.State = Idle

	log.Println("Door closing")

	return elev, true
}

func motorTimeout(elev Elevator) (Elevator, bool) {
	log.Println("Motor timed out!!")

	elevio.SetMotorDirection(elevio.MD_Stop)
	elev.State = Error

	return elev, true
}

func setDirection(elev Elevator) (Elevator, bool) {
	var updateElev bool = false
	var d elevio.MotorDirection
	if elev.ActiveOrder.Floor > elev.Floor {
		d = elevio.MD_Up
	} else if elev.ActiveOrder.Floor < elev.Floor {
		d = elevio.MD_Down
	} else {
		elev, updateElev = arrivedAtTarget(elev)
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
func Driver(port int, nfloors, nbuttons int, mainElevatorChan chan<- Elevator,
	orderChan <-chan order.Order, buttonPressChan chan<- order.Order) {

	Nfloors = nfloors
	Nbuttons = nbuttons

	var elev Elevator
	elev.State = Init
	elev.Orders = make([][]order.Order, Nfloors)
	for i := range elev.Orders {
		elev.Orders[i] = make([]order.Order, Nbuttons)
	}

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	driverInit(port, drvButtons, drvFloors)

	elev.State = Idle
	mainElevatorChan <- elev // to not make it crash on default in main
	var updateElev bool = false
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

		default:
			// do nothing
		}

		// Send new elevator object to main
		if updateElev {
			mainElevatorChan <- elev
			updateElev = false

			setLamps(elev) // TODO: is this a fitting place to update lights?
		}

		// Act according to new state
		switch elev.State {
		case Idle:
			if !(elev.ActiveOrder.Status == order.Finished || elev.ActiveOrder.Status == order.Abort) {
				// will come into effect at next iteration
				elev.State = Moving
				updateElev = true
			}

		case Moving:
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

func OrderMatrixToString(elev Elevator) string {
	s := ""
	for f := 0; f < Nfloors; f++ {
		for i := range elev.Orders[f] {
			s += fmt.Sprintf("%d", elev.Orders[f][i].Status)
		}
		s += " "
	}
	return s
}
