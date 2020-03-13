package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"./driver"
	"./elevTypes/elevator"
	"./elevTypes/order"
	"./network"
	"./network/bcast"
)

const (
	wdTimerInterval time.Duration = 500 * time.Millisecond
)

var (
	Nfloors  int
	Nbuttons int
)

func readElevatorFromFile() elevator.Elevator {
	// TODO: implement this function
	file, _ := os.Open("elevBackupFile.txt")

	data, _ := ioutil.ReadAll(file)

	var elev elevator.Elevator

	json.Unmarshal([]byte(data), &elev)

	file.Close()

	return elev
	//return elevator.Elevator{}
}

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

func main() {
	logFile, err := setupLog()
	if err != nil {
		fmt.Println("Error setting up log")
		return
	}
	defer logFile.Close()
	logPID()

	// Watchdog setup
	wdChan := make(chan interface{})
	wdTimer := time.NewTimer(wdTimerInterval)
	go bcast.Transmitter(57005, wdChan)

	// Capture signals to exit more gracefully
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Init driver
	port := flag.Int("port", 15657, "Port for connecting to ElevatorServer/SimElevatorServer")
	nfloors := flag.Int("floors", 4, "Number of floors per elevator")
	readFile := flag.Bool("fromfile", false, "Read Elevator struct from file if this flag is passed")
	flag.Parse()

	Nfloors = *nfloors
	Nbuttons = 3 // must be constant

	mainElevatorChan := make(chan elevator.Elevator, 100)
	orderChan := make(chan order.Order, 100)
	buttonPressChan := make(chan order.Order)
	go driver.Driver(*port, Nfloors, Nbuttons, mainElevatorChan, orderChan, buttonPressChan)

	var elev elevator.Elevator
	elev = <-mainElevatorChan // hang program untill driver is initialized

	// Combine network and driver
	txChan := make(chan interface{})
	networkOrderChan := make(chan order.Order)
	go network.Network(20028, txChan, networkOrderChan)

	if *readFile {
		elev = readElevatorFromFile() // TODO: implement this function
	}
	var lastOrder order.Order
	for {
		select {
		case newElev := <-mainElevatorChan:
			log.Printf("New elevator: %s f:%#v d:%#v s:%#v\n",
				newElev.ActiveOrder.ToString(), newElev.Floor,
				newElev.Direction, newElev.State)
			log.Println(newElev.OrderMatrixToString())

			if elev.ActiveOrder.Status != newElev.ActiveOrder.Status &&
				newElev.ActiveOrder.Status == order.Finished {

				txChan <- newElev.ActiveOrder
			}
			elev = newElev

		case ord := <-buttonPressChan:
			// if cab order, no need to broadcast
			if ord.Type != order.Cab {
				// status is set to NotTaken in buttonPress in driver
				// ord.Status = order.InitialBroadcast
				txChan <- ord // this also sends to myself
			}

			// TODO: might not be necessary since my own packets are received
			// 		 and matrix is updated in buttonPress in driver for cab calls
			// orderChan <- ord

		case ord := <-networkOrderChan:
			log.Printf("Received order from network: %s\n", ord.ToString())
			orderChan <- ord

			// if message has status not taken, add that to the matrix
			// if message has status finished, add that to matrix and handle
			// the same must happen either way

			//TODO: write to file?

		case <-wdTimer.C:
			wdChan <- "28-IAmAlive"
			wdTimer.Reset(wdTimerInterval)

		case sig := <-sigs:
			log.Printf("Received signal: %s. Exiting...\n", sig.String())
			return

		case <-time.After(50 * time.Millisecond):
			// My computer spun up a lot if this is runs every time there is no
			// other event.

			// send next order if not currently active order
			o := findNextOrder(elev)
			if o.Status != order.Invalid && o != lastOrder {
				fmt.Printf("Order to exec: %s\n", o.ToString())
				lastOrder = o
				orderChan <- o
			}
		}
	}
}

func orderBelow(elev elevator.Elevator) (int, int) {
	for f := elev.Floor - 1; f >= 0; f-- {
		for i := range elev.Orders[f] {
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, i
			}
		}
	}

	return -1, -1
}

func orderAbove(elev elevator.Elevator) (int, int) {
	for f := elev.Floor + 1; f < 4; f++ {
		for i := range elev.Orders[f] {
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, i
			}
		}
	}

	return -1, -1
}

func orderAtFloor(elev elevator.Elevator) (int, int) {
	for i := range elev.Orders[elev.Floor] {
		if elev.Orders[elev.Floor][i].Status == order.NotTaken {
			return elev.Floor, i
		}
	}
	return -1, -1
}

func findNextOrder(elev elevator.Elevator) order.Order {
	// TODO: re-write this to use one of the algorithms on github

	// this is currently a simple, dumb implementation that simply looks if
	// there are orders above, go up. if below, go down.
	f, t := -1, -1
	if elev.Direction == elevator.Down {
		f, t = orderBelow(elev)
	} else if elev.Direction == elevator.Up {
		f, t = orderAbove(elev)
	} else {
		f, t = orderAtFloor(elev)
	}

	o := order.Order{Floor: f, Type: order.Type(t), Status: order.Execute}
	if f < 0 || t < 0 {
		// no orders exist
		o.Status = order.Invalid
	}

	return o
}
