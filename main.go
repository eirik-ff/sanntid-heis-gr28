package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"./costfunction"
	"./driver"
	"./elevTypes/order"
	"./network"
	"./network/bcast"
	"./queue"
)

const (
	wdTimerInterval time.Duration = 500 * time.Millisecond
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

func main() {
	// costfunction.TestCost()
	// return

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
	flag.Parse()

	buttonOrderChan := make(chan order.Order)
	execOrderChan := make(chan order.Order)
	stateChan := make(chan driver.ElevState)
	go driver.Driver(*port, buttonOrderChan, execOrderChan, stateChan)

	var state driver.ElevState
	var lowerCostReplySentID int64

	// Combine network and driver
	txChan := make(chan interface{})
	networkOrderChan := make(chan order.Order)
	go network.Network(20028, txChan, networkOrderChan)

	// Init order
	localOrderEnqueueChan := make(chan order.Order)
	localOrderDequeueChan := make(chan order.Order)
	localNextOrderChan := make(chan order.Order)
	go queue.Queue(localOrderEnqueueChan, localOrderDequeueChan, localNextOrderChan)

	for {
		select {
		case state = <-stateChan:
			log.Printf("New state: %v\n", state)

			if state.Order.Status == order.Finished {
				localOrderDequeueChan <- state.Order
			}

		case ord := <-buttonOrderChan:
			// TODO: move formatting to function
			ord.ID = time.Now().UnixNano()
			ord.Status = order.InitialBroadcast
			ord.Cost = costfunction.Cost(ord, state)

			localOrderEnqueueChan <- ord

			// if cab order, no need to broadcast
			if ord.Type != order.Cab {
				txChan <- ord
			}

		case nextOrder := <-localNextOrderChan:
			log.Printf("Next order to execute: %#v\n", nextOrder)
			execOrderChan <- nextOrder

		case ord := <-networkOrderChan:
			log.Printf("Received order from network: %#v\n", ord)

			switch ord.Status {
			case order.InitialBroadcast:
				cost := costfunction.Cost(ord, state)
				log.Printf("Own cost: %d\n", cost)
				if cost < ord.Cost {
					ord.Status = order.LowerCostReply
					ord.Cost = cost

					// if lower cost, save ID so you won't abort your own order
					lowerCostReplySentID = ord.ID

					localOrderEnqueueChan <- ord
					txChan <- ord
					log.Printf("Sending lower cost reply with cost: %d\n", cost)
				}
			case order.LowerCostReply:
				if ord.ID != lowerCostReplySentID {
					localOrderDequeueChan <- ord
				}
			}
		case <-wdTimer.C:
			wdChan <- "28-IAmAlive"
			wdTimer.Reset(wdTimerInterval)

		case sig := <-sigs:
			log.Printf("Received signal: %s. Exiting...\n", sig.String())
			return
		}
	}
}
