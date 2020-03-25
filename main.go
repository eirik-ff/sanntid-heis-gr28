package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
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
	orderTimer     *time.Timer //Timer used in updatedElevatorState
	Nfloors        int
	Nbuttons       int
	backupFileName string = "elevBackupFile_%d.log" // will be formatted in main
)

//States for MAIN
type State int

const (
	Init   State = 0
	Normal State = 1
	Error  State = 2
)

func readElevatorFromFile(fileName string) elevator.Elevator {

	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("No file with path '%s' exists\n", fileName)
		return elevator.NewElevator(Nfloors, Nbuttons)
	}
	defer file.Close()

	var elev elevator.Elevator
	data, _ := ioutil.ReadAll(file)
	json.Unmarshal([]byte(data), &elev)

	return elev
}

func writeElevatorToFile(fileName string, elev elevator.Elevator) {
	os.Remove(fileName)

	file, _ := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()

	msg, _ := json.Marshal(elev)
	if _, err := file.Write([]byte(msg)); err != nil {
		log.Fatal(err)
	}

	file.Close()
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

//Function for handling 'updated elevator event'
//
//*****Why did we call this function*************
// | Causes for event                           |
// |--------------------------------------------|
// | Order matrix changed                       |
// | Floor changed                              |
// | DIrection changed                          |
// | State changed  (this includes error state) |
//
//*****What should be done*************
// | Condition      | Need to do                                                                                                            |
// |----------------+-----------------------------------------------------------------------------------------------------------------------|
// | ActiveOrder.Status = Finished | Broadcast the active order on the network. (The driver should have removed the order from the matrix.) |
// | State = Error                 | Send active order with status notTaken on the network and go to error state                            |
//
//
// | Parameters        | Description                |
// |-------------------+----------------------------|
// | e elevator.Elevator | The updated elevator state |
// | s State           | The main state             |
//
// | Returns             | Description               |
// |---------------------+---------------------------|
// | State               | The updated state of Main |
// | elevator.Elevator   | New elevator              |
func updatedElevatorState(newElev elevator.Elevator, elev elevator.Elevator, s State, txChan chan<- interface{}) (State, elevator.Elevator, order.Order) {
	log.Println(newElev.ToString())
	log.Println(newElev.OrderMatrixToString())

	//Currently both if statements broadcasts the active order on the txChan
	//I kept it this way, if we need to do additional stuff before broadcasting in error state.
	//If we do not need to change anything in the case of error. The txChan <- newElev.ActiveOrder should be moved out of the ifs.
	var state State = s
	var nextOrder order.Order

	//Evaluate if a another order should be taken
	if newElev.ActiveOrder.Status == order.Finished ||
		elev.Floor != newElev.Floor {
		nextOrder = findNextOrder(newElev)

		if nextOrder.Status != order.Invalid {
			fmt.Printf("Order to exec: %s\n", nextOrder.ToString())

			//Generate random number
			a := 100000
			lower := -a
			upper := a
			d := lower + rand.Intn(upper-lower)

			// select delay based on distance to order
			dist := math.Abs(float64(nextOrder.Floor) - float64(newElev.Floor))
			orderWaitInterval := time.Duration(250*dist) * time.Millisecond
			orderWaitInterval += (time.Duration(d) * time.Microsecond)
			log.Printf("orderWaitInterval: %d\n", orderWaitInterval) 
			orderTimer.Reset(orderWaitInterval) //Start timer
		}
	}

	///////////////////////////////
	// SEND ORDERS ON NETWORK	 //
	///////////////////////////////

	//Filter what to send over the network
	if newElev.ActiveOrder.Status == order.Finished ||
		newElev.ActiveOrder.Status == order.Taken ||
		newElev.ActiveOrder.Status == order.NotTaken {

		//Only transmit if active order changed, and not cab order
		if newElev.ActiveOrder.Type != order.Cab && !order.CompareEq(elev.ActiveOrder, newElev.ActiveOrder) {
			txChan <- newElev.ActiveOrder //Send active order
		}
	}

	if newElev.State == elevator.Error && elev.State != newElev.State {
		// TODO: need last check to not send a lot of updates when in error state.
		//       but why is it sending so much?

		//If elevator entered error state - send active order
		//(is the order set to notTaken in the driver? - if not, need to set it here)
		o := newElev.ActiveOrder
		o.Status = order.NotTaken
		if o.Type != order.Cab {
			txChan <- o
			log.Printf("Elev.State = Error => sent order on network: %s\n", o.ToString())
		}

		state = Error //Go to error state
	}

	return state, newElev, nextOrder
}

//Function for handling 'new button press'
//
//*****Why did we call this function*************
// | Causes for event |
// |------------------|
// | New button press |

// *****What should be done*************
// | Need to do                                            |
// |-------------------------------------------------------|
// | Broadcast new order on the network if not cab call    |
//
// | Parameters      | Description        |
// |-----------------+--------------------|
// | ord order.Order | The new order      |
func newButtonPress(ord order.Order, txChan chan interface{}) {
	if ord.Type != order.Cab {
		txChan <- ord //transmit new button order
		log.Printf("Sending order on network: %s\n", ord.ToString())
	}
}

//Function for handling 'new button press'
//
// *****Why did we call this function*************
// | Causes for event          |
// |---------------------------|
// | New order                 |
// | Someone started an order  |
// | Someone finished an order |
//
// *****What should be done*************
// | Condition                    | Need to do                                |
// |------------------------------+-------------------------------------------|
// | New order                    | Update the matrix with the received order |
// | Someone stated an order      | Update the matrix with the received order |
// | Someone finished a new order | Remove the order from the matrix          |
//
// | Parameters             | Description        |
// |------------------------+--------------------|
// | ord order.Order        | The received order |
// | e elevator.Elevator    | The elevator state |
//
// | Return            | Description                |
// |-------------------+----------------------------|
// | elevator.Elevator | The updated elevator state |
func newNetworkMessage(ord order.Order, elev elevator.Elevator,
	orderChan chan<- order.Order) elevator.Elevator {
	log.Printf("Received order from network: %s\n", ord.ToString())
	orderChan <- ord
	// if message has status not taken, add that to the matrix
	// if message has status finished, add that to matrix and handle
	// the same must happen either way

	return elev
}

func main() {

	state := Init //Set the state of main to Init

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

	backupFileName = fmt.Sprintf(backupFileName, *port)

	mainElevatorChan := make(chan elevator.Elevator, 100)
	orderChan := make(chan order.Order, 100)
	buttonPressChan := make(chan order.Order)

	var elev elevator.Elevator = elevator.NewElevator(Nfloors, Nbuttons)
	if *readFile {
		elev = readElevatorFromFile(backupFileName) //Read orders from file
		log.Println("Read old configuration from file")
		log.Println(elev.ToString())
		log.Println(elev.OrderMatrixToString())
		o := elev.ActiveOrder
		o.Status = order.Execute
		orderChan <- o
	}

	go driver.Driver(*port, Nfloors, Nbuttons, mainElevatorChan, orderChan, buttonPressChan, elev)
	elev = <-mainElevatorChan // hang program untill driver is initialized

	// Combine network and driver
	txChan := make(chan interface{})
	networkOrderChan := make(chan order.Order)
	go network.Network(20028, *port, txChan, networkOrderChan)

	var lastOrder order.Order
	_ = lastOrder

	state = Normal //Set the state of main to Normal

	var nextOrder order.Order
	orderTimer = time.NewTimer(1 * time.Second) // this init time doesn't matter
	orderTimer.Stop()

	rand.Seed(time.Now().UnixNano())

	for {
		select {
		case newElev := <-mainElevatorChan:
			state, elev, nextOrder = updatedElevatorState(newElev, elev, state, txChan)

		case <-orderTimer.C: //Order timer started in updatedElevatorState timed out
			log.Println(nextOrder.ToString())
			//Check if next order to execute is already taken
			if nextOrder.Status != order.Invalid && elev.Orders[nextOrder.Floor][nextOrder.Type].Status == order.NotTaken {

				orderChan <- nextOrder

				//Check if not cab, different than current order and current order is 'taken' (currently executed)
				if elev.ActiveOrder.Type != order.Cab && nextOrder.Floor != elev.ActiveOrder.Floor && nextOrder.Type != elev.ActiveOrder.Type && elev.ActiveOrder.Status == order.Taken {
					// tell the other elevators that the last active order
					// is no longer active and someone else can take it
					o := elev.ActiveOrder
					o.Status = order.NotTaken

					txChan <- o //Send new active order
					log.Printf("Sending order on network: %s\n", o.ToString())
					log.Println("Next order different from active, send NotTaken")
				}
			}

		case ord := <-buttonPressChan:
			newButtonPress(ord, txChan)

		case ord := <-networkOrderChan:
			elev = newNetworkMessage(ord, elev, orderChan)

		case <-wdTimer.C:
			wdChan <- "28-IAmAlive"
			wdTimer.Reset(wdTimerInterval)

		case sig := <-sigs:
			log.Printf("Received signal: %s. Exiting...\n", sig.String())
			return

		case <-time.After(100 * time.Millisecond):

			//////////////////
			// Evaluate FSM //
			//////////////////
			// log.Println("Evaluate FSM")

			switch state {
			case Init:
				//Init mode
				//Do nothing?

				//Init is done before entering the for/select main loop
			case Normal:
				//Normal mode
				//Do nothing?
				timeoutChan := make(chan order.Order, elev.Nfloors*elev.Nbuttons)

				//Check if any taken orders are timed out
				elev.CheckOrderTimestamp(timeoutChan)

				for len(timeoutChan) > 0 {
					o := <-timeoutChan //get timedout order
					o.Status = order.NotTaken
					orderChan <- o //send order to driver
				}

				writeElevatorToFile(backupFileName, elev) //write orders to file
			case Error:
				//Error mode

				//Do something to check if you should still be in error mode
				//F.ex. Move the motor to check if a new floor change is detected.

				//////////////////
				// PSUDOCODE    //
				//////////////////
				// if stillError() {
				// 	state = error
				// }
				// else {
				// 	state = normal or init
				// }

			default: // unknown state

			}
		}
	}
}

func orderBelow(elev elevator.Elevator) (int, int, bool) {
	for f := elev.Floor - 1; f >= 0; f-- {
		for t := range elev.Orders[f] {
			l := len(elev.Orders[f])
			i := l - t - 1
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, i, true
			}
		}
	}

	return -1, -1, false
}

func orderAbove(elev elevator.Elevator) (int, int, bool) {
	for f := elev.Floor + 1; f < 4; f++ {
		for t := range elev.Orders[f] {
			l := len(elev.Orders[f])
			i := l - t - 1
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, i, true
			}
		}
	}

	return -1, -1, false
}

func orderAtFloor(elev elevator.Elevator) (int, int, bool) {
	for t := range elev.Orders[elev.Floor] {
		l := len(elev.Orders[elev.Floor])
		i := l - t - 1
		if elev.Orders[elev.Floor][i].Status == order.NotTaken {
			return elev.Floor, i, true
		}
	}
	return -1, -1, false
}

func orderBetween(elev elevator.Elevator) (int, int, bool) {
	// checks if there is NotTaken order between current pos and active order
	if elev.Direction == elevator.Up {
		for f := elev.Floor; f < elev.ActiveOrder.Floor; f++ {
			for t := range elev.Orders[f] {
				l := len(elev.Orders[f])
				i := l - t - 1
				if elev.Orders[f][i].Status == order.NotTaken {
					return f, i, true
				}
			}
		}
	} else if elev.Direction == elevator.Down {
		for f := elev.Floor; f > elev.ActiveOrder.Floor; f-- {
			for t := range elev.Orders[f] {
				l := len(elev.Orders[f])
				i := l - t - 1
				if elev.Orders[f][i].Status == order.NotTaken {
					return f, i, true
				}
			}
		}
	}
	return -1, -1, false
}

func findNextOrder(elev elevator.Elevator) order.Order {
	// TODO: re-write this to use one of the algorithms on github

	// this is currently a simple, dumb implementation that simply looks if
	// there are orders above, go up. if below, go down.
	f, t, ok := orderAtFloor(elev)
	if !ok {
		f, t, ok = orderAbove(elev)
	}
	if !ok {
		f, t, ok = orderBelow(elev)
	}
	if !ok {
		f, t, ok = orderBetween(elev)
	}

	o := order.Order{Floor: f, Type: order.Type(t), Status: order.Execute}
	if !ok {
		// no orders exist
		o.Status = order.Invalid
	}

	return o
}
