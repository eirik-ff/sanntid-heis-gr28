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

//States for MAIN
type State int
const (
	Init     State = 0
	Normal   State = 1
	Error    State = 2
)


func readElevatorFromFile() driver.Elevator {
	// TODO: implement this function
	file, _ := os.Open("elevBackupFile.txt")

	data, _ := ioutil.ReadAll(file)

	var elev driver.Elevator

	json.Unmarshal([]byte(data), &elev)

	file.Close()

	return elev
	//return driver.Elevator{}
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
// | e driver.Elevator | The updated elevator state |
// | s State           | The main state             |
//
// | Returns | Description               |
// |---------+---------------------------|
// | State   | The updated state of Main |
func updatedElevatorState(e driver.Elevator, s State) State {

	log.Printf("New elevator: %#v f:%#v d:%#v s:%#v\n",
		elev.ActiveOrder, elev.Floor, elev.Direction, elev.State)
	log.Println(driver.OrderMatrixToString(elev))

	
	//Currently both if statements broadcasts the active order on the txChan
	//I kept it this way, if we need to do additional stuff before broadcasting in error state.
	//If we do not need to change anything in the case of error. The txChan <- e.ActiveOrder should be moved out of the ifs.
	var state State
	if e.ActiveOrder.Status == order.Finished {
		//If active order is finished - Broadcast order to network
		txChan <- e.ActiveOrder
		state = s
		
	} else if e.State == Error {
		//If elevator is in error state - send active order
		//(is the order set to notTaken in the driver? - if not, need to set it here)
		//e.ActiveOrder = order.NotTaken
		txChan <- e.ActiveOrder
		state = Error
	}
	return state
}

//Function for handling 'new button press'
//
//*****Why did we call this function*************
// | Causes for event |
// |------------------|
// | New button press |

// *****What should be done*************
// | Need to do                                |
// |-------------------------------------------|
// | Broadcast new order on the network        |
//
// | Parameters      | Description        |
// |-----------------+--------------------|
// | ord order.Order | The new order      |
func newButtonPress(ord order.Order) {
	
	//Broadcast new order
	txChan <- ord	
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
// | e driver.Elevator      | The elevator state |
//
// | Return          | Description                |
// |-----------------+----------------------------|
// | driver.Elevator | The updated elevator state |
func newNetworkMessage(ord order.Order, e driver.Elevator) driver.Elevator {
	
	log.Printf("Received order from network: %#v\n", ord)
	
	//If finished - remove
	if ord.Status == order.Finished {
		//remove from matrix
	} else {
		//update matrix with new order
	}
	
	return driver.Elevator	
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
	nbuttons := flag.Int("buttons", 3, "Number of button types per elevator (e.g. all up, hall down, cab call)")
	readFile := flag.Bool("fromfile", false, "Read Elevator struct from file if this flag is passed")
	flag.Parse()

	Nfloors = *nfloors
	Nbuttons = *nbuttons

	mainElevatorChan := make(chan driver.Elevator, 100)
	orderChan := make(chan order.Order, 100)
	buttonPressChan := make(chan order.Order)
	go driver.Driver(*port, Nfloors, Nbuttons, mainElevatorChan, orderChan, buttonPressChan)

	var elev driver.Elevator
	elev = <-mainElevatorChan // hang program untill driver is initialized

	// Combine network and driver
	txChan := make(chan interface{})
	networkOrderChan := make(chan order.Order)
	go network.Network(20028, txChan, networkOrderChan)

	if *readFile {
		elev = readElevatorFromFile() // TODO: implement this function
	}
	var lastOrder order.Order


	state = Normal //Set the state of main to Normal
	for {
		select {
		case elev = <-mainElevatorChan:
			// log.Printf("New elevator: %#v f:%#v d:%#v s:%#v\n",
			// 	elev.ActiveOrder, elev.Floor, elev.Direction, elev.State)
			// log.Println(driver.OrderMatrixToString(elev))

			/**********************************************
			OLD CODE

			// if elev.ActiveOrder.Status == order.Finished {
			// 	txChan <- elev.ActiveOrder
			// }
            **********************************************/


			/////////
			// FSM //
			/////////			
			state = updatedElevatorState(elev, state)			

		case ord := <-buttonPressChan:


			/**********************************************
			OLD CODE
			
			// if cab order, no need to broadcast
			if ord.Type != order.Cab {
				// status is set to NotTaken in buttonPress in driver
				ord.Status = order.InitialBroadcast
				txChan <- ord // this also sends to myself
			}

			// TODO: might not be necessary since my own packets are received
			// 		 and matrix is updated in buttonPress in driver for cab calls
			// orderChan <- ord

		    **********************************************/

			/////////////
			// FSM 	   //
			/////////////
			newButtonPress(ord)

		case ord := <-networkOrderChan:

			// orderChan <- ord

			// if message has status not taken, add that to the matrix
			// if message has status finished, add that to matrix and handle
			// the same must happen either way

			//TODO: write to file?

			/////////////
			// FSM 	   //
			/////////////
			elev = newNetworkMessage(ord, elev)
			
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
				fmt.Printf("Order to exec: %#v\n", o)
				lastOrder = o
				orderChan <- o
			}

		case <-time.After(100 * time.Millisecond):
			//////////////////
			// Evaluate FSM //
			//////////////////
			
			switch state {
			case Init:
				//Init mode
				//Do nothing?

				//Init is done before entering the for/select main loop
			case Normal:
				//Normal mode
				//Do nothing?
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

func orderBelow(elev driver.Elevator) (int, int) {
	for f := elev.Floor - 1; f >= 0; f-- {
		for i := range elev.Orders[f] {
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, i
			}
		}
	}

	return -1, -1
}

func orderAbove(elev driver.Elevator) (int, int) {
	for f := elev.Floor + 1; f < 4; f++ {
		for i := range elev.Orders[f] {
			if elev.Orders[f][i].Status == order.NotTaken {
				return f, i
			}
		}
	}

	return -1, -1
}

func orderAtFloor(elev driver.Elevator) (int, int) {
	for i := range elev.Orders[elev.Floor] {
		if elev.Orders[elev.Floor][i].Status == order.NotTaken {
			return elev.Floor, i
		}
	}
	return -1, -1
}

func findNextOrder(elev driver.Elevator) order.Order {
	// TODO: re-write this to use one of the algorithms on github

	// this is currently a simple, dumb implementation that simply looks if
	// there are orders above, go up. if below, go down.
	f, t := -1, -1
	if elev.Direction == driver.MD_Down {
		f, t = orderBelow(elev)
	} else if elev.Direction == driver.MD_Up {
		f, t = orderAbove(elev)
	} else {
		f, t = orderAtFloor(elev)
	}

	o := order.Order{Floor: f, Type: t, Status: order.Execute}
	if f < 0 || t < 0 {
		// no orders exist
		o.Status = order.Invalid
	}

	return o
}
