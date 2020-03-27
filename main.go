package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"./driver"
	"./elevTypes/elevator"
	"./elevTypes/order"
	"./filebackup"
	"./network"
	"./request"
	"./watchdog"
)

const (
	// How often to check if an order has timed out.
	checkTimestampInterval time.Duration = 100 * time.Millisecond
	// How ofter to write the current state to a backup file.
	writeToFileInterval time.Duration = 100 * time.Millisecond
)

var (
	// orderTimer is used to wait before an order is accepted.
	orderTimer *time.Timer
	// Nfloors is the number of floor per elevator
	Nfloors int
	// Nbuttons is the number of button types per elevator
	Nbuttons int
	// will be formatted in main
	backupFileName string = "elevBackupFile_%d.log"
)

func setupLog() (*os.File, error) {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Couldn't create user object")
		return nil, err
	}
	logDirPath := usr.HomeDir + "/sanntid-heis-gr28/logs/"
	logFileName := "elev.log"
	logFilePath := logDirPath + logFileName

	err = os.MkdirAll(logDirPath, 0755)
	if err != nil {
		fmt.Printf("Error creating log directory at %s\n", logDirPath)
		return nil, err
	}
	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0655)
	if err != nil {
		fmt.Printf("Error opening info log file at %s\n", logFilePath)
		return nil, err
	}
	_ = logFile // logFile will be used later, but for now, stdout is easier

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(os.Stdout)

	return logFile, nil
}

func logPID() {
	// Save current PID to file to be able to kill program
	pid := os.Getpid()
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user")
		return
	}
	exec.Command("/bin/bash", "-c", fmt.Sprintf("echo %d > %s/sanntid-heis-gr28/logs/pid.txt", pid, usr.HomeDir)).Run()
	log.Printf("PID: %d\n", pid)
}

func setupSignals() chan os.Signal {
	// Capture signals to exit more gracefully
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	return sigs
}

func setupFlags() (int, bool) {
	port := flag.Int("port", 15657, "Port for connecting to ElevatorServer/SimElevatorServer")
	nfloors := flag.Int("floors", 4, "Number of floors per elevator")
	readFile := flag.Bool("fromfile", false, "Read Elevator struct from file if this flag is passed")
	flag.Parse()

	Nfloors = *nfloors
	Nbuttons = 3 // must be constant
	return *port, *readFile
}

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

func main() {
	rand.Seed(time.Now().UnixNano())

	logFile, err := setupLog()
	if err != nil {
		fmt.Println("Error setting up log")
		return
	}
	defer logFile.Close()
	logPID()

	watchdog.Setup("28-IAmAlive")
	sigs := setupSignals()
	elevIOport, readFile := setupFlags()

	backupFileName = fmt.Sprintf(backupFileName, elevIOport)

	mainElevatorChan := make(chan elevator.Elevator, 100)
	orderChan := make(chan order.Order, 100)
	buttonPressChan := make(chan order.Order)

	var elev elevator.Elevator = elevator.NewElevator(Nfloors, Nbuttons)
	if readFile {
		elev = filebackup.Read(backupFileName, Nfloors, Nbuttons)
		o := elev.ActiveOrder
		o.Status = order.Execute
		orderChan <- o
	}

	go driver.Driver(elevIOport, Nfloors, Nbuttons, mainElevatorChan, orderChan, buttonPressChan, elev)
	elev = <-mainElevatorChan // halt program untill driver is initialized

	txChan := make(chan interface{})
	networkOrderChan := make(chan order.Order)
	go network.Network(20028, strconv.Itoa(elevIOport), txChan, networkOrderChan)

	orderTimer = time.NewTimer(time.Second) // this init time doesn't matter
	orderTimer.Stop()

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

		case <-time.After(writeToFileInterval):
			filebackup.Write(backupFileName, elev)

		case <-watchdog.Hungry():
			watchdog.Feed()

		case sig := <-sigs:
			log.Printf("Received signal: %s. Exiting...\n", sig.String())
			return
		}
	}
}
