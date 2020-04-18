package control

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"../driver"
	"../elevTypes/elevator"
	"../elevTypes/order"
	"../filebackup"
	"../network"
	"../request"
	"../watchdog"
)

// NOTE: timer durations must be different. If they're equal, one of the timers
// might not fire.
const (
	// How often to check if an order has timed out.
	checkTimestampInterval time.Duration = 100 * time.Millisecond
	// How ofter to write the current state to a backup file.
	writeToFileInterval time.Duration = 90 * time.Millisecond
)

var (
	// orderTimer is used to wait before an order is accepted.
	orderTimer *time.Timer
	// will be formatted in main
	backupFileName string = "logs/elevBackupFile_%d.log"

	Nfloors  int
	Nbuttons int = 3
)

var (
	mainElevatorChan chan elevator.Elevator
	orderChan        chan order.Order
	buttonPressChan  chan order.Order

	txChan           chan interface{}
	networkOrderChan chan order.Order
)

// startOrderTimer selects delay based on distance to order and a random backoff
// interval. Similar to 802.11 protocol.
func startOrderTimer(newElev elevator.Elevator, nextOrder order.Order) {
	// find wait duration based on distance
	distPenalty := 250.0 // how much to space distance in time
	dist := math.Abs(float64(nextOrder.Floor) - float64(newElev.Floor))
	orderWaitInterval := time.Duration(distPenalty*dist) * time.Millisecond

	// add random number to avoid time collisions
	backoffInterval := 100000
	lower := -backoffInterval
	upper := backoffInterval
	duration := lower + rand.Intn(upper-lower)
	orderWaitInterval += (time.Duration(duration) * time.Microsecond)

	orderTimer.Reset(orderWaitInterval) // resets starts the timer again
}

// startNextOrder checks if nextOrder is still not taken and executes it. This
// function is run when orderTimer times out.
func startNextOrder(
	elev elevator.Elevator,
	nextOrder order.Order,
	orderChan chan<- order.Order,
	txChan chan<- interface{}) {
	// Check if next order to execute is already taken
	if nextOrder.Status != order.Invalid &&
		elev.Orders[nextOrder.Floor][nextOrder.Type].Status == order.NotTaken {
		orderChan <- nextOrder

		// tell the other elevators that the last active order
		// is no longer active and someone else can take it
		if elev.ActiveOrder.Type != order.Cab &&
			!order.CompareFloorAndType(nextOrder, elev.ActiveOrder) &&
			elev.ActiveOrder.Status == order.Taken {
			log.Println("Next order different from current active. " +
				"Sending NotTaken.")
			o := elev.ActiveOrder
			o.Status = order.NotTaken
			txChan <- o
		}
	}
}

// updatedElevatorState handles when a new Elevator object is received from the
// driver.
func updatedElevatorState(
	newElev elevator.Elevator,
	elev elevator.Elevator,
	txChan chan<- interface{}) (elevator.Elevator, order.Order) {

	log.Println(newElev.ToString())
	log.Println(newElev.OrderMatrixToString())

	nextOrder := request.FindNextOrder(newElev)
	if nextOrder.Status != order.Invalid {
		startOrderTimer(newElev, nextOrder)
	}

	if newElev.ActiveOrder.Status == order.Finished ||
		newElev.ActiveOrder.Status == order.Taken {
		// Only transmit if active order changed, and not cab order
		if !order.CompareEq(elev.ActiveOrder, newElev.ActiveOrder) &&
			newElev.ActiveOrder.Type != order.Cab {
			txChan <- newElev.ActiveOrder
		}
	}

	enteredErrorState := newElev.State == elevator.Error && elev.State != newElev.State
	if enteredErrorState {
		o := newElev.ActiveOrder
		o.Status = order.NotTaken
		if o.Type != order.Cab {
			txChan <- o
			log.Println("Entered error state. Sending active order on network.")
		}
	}

	return newElev, nextOrder
}

func newButtonPress(ord order.Order, txChan chan interface{}) {
	if ord.Type != order.Cab {
		txChan <- ord
		log.Printf("Sending order on network: %s\n", ord.ToString())
	}
}

func newNetworkMessage(ord order.Order, orderChan chan<- order.Order) {
	log.Printf("Received order from network: %s\n", ord.ToString())
	orderChan <- ord
}

// Setup initializes the driver and network module as well as sets up other
// variables for controlling the elevator.
func Setup(elevIOport, nfloors int, readFile bool) {
	rand.Seed(time.Now().UnixNano())
	Nfloors = nfloors

	var elev elevator.Elevator = elevator.NewElevator(Nfloors, Nbuttons)
	mainElevatorChan = make(chan elevator.Elevator, 100)
	orderChan = make(chan order.Order, 100)
	buttonPressChan = make(chan order.Order)
	backupFileName = fmt.Sprintf(backupFileName, elevIOport)
	if readFile {
		elev = filebackup.Read(backupFileName, Nfloors, Nbuttons)
		o := elev.ActiveOrder
		o.Status = order.Execute
		orderChan <- o
	}
	go driver.Driver(elevIOport, Nfloors, Nbuttons, mainElevatorChan,
		orderChan, buttonPressChan, elev)

	txChan = make(chan interface{})
	networkOrderChan = make(chan order.Order)
	logID := "port" + strconv.Itoa(elevIOport)
	go network.Network(20028, logID, txChan, networkOrderChan)

	orderTimer = time.NewTimer(time.Second) // this init time doesn't matter
	orderTimer.Stop()
}

// Loop starts a for-select loop that runs the control logic for the elevator.
func Loop(sigs chan os.Signal) {
	elev := <-mainElevatorChan // halt program until driver is initialized
	var nextOrder order.Order
	for {
		select {
		case newElev := <-mainElevatorChan:
			elev, nextOrder = updatedElevatorState(newElev, elev, txChan)

		case <-orderTimer.C:
			startNextOrder(elev, nextOrder, orderChan, txChan)

		case ord := <-buttonPressChan:
			newButtonPress(ord, txChan)

		case ord := <-networkOrderChan:
			newNetworkMessage(ord, orderChan)

		case <-time.After(checkTimestampInterval):
			timeoutChan := make(chan order.Order, elev.Nfloors*elev.Nbuttons)
			elev.CheckOrderTimestamp(timeoutChan)
			for len(timeoutChan) > 0 {
				o := <-timeoutChan
				o.Status = order.NotTaken
				orderChan <- o
			}

			filebackup.Write(backupFileName, elev)

		case <-watchdog.Hungry():
			watchdog.Feed()

		case sig := <-sigs:
			log.Printf("Received signal: %s. Exiting...\n", sig.String())
			return
		}
	}
}
