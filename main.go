package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"time"

	"./driver"
	"./network/bcast"
)

func setupLog() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Couldn't create user object")
		return
	}
	logDirPath := usr.HomeDir + "/sanntid-heis-gr28/logs/"
	logFileName := "elev.log"
	logFilePath := logDirPath + logFileName

	err = os.MkdirAll(logDirPath, 0755)
	if err != nil {
		fmt.Printf("Error creating log directory at %s\n", logDirPath)
		return
	}
	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0655)
	if err != nil {
		fmt.Printf("Error opening info log file at %s\n", logFilePath)
		return
	}
	_ = logFile // logFile will be used later, but for now, stdout is easier

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

func main() {
	setupLog()

	pid := os.Getpid()
	exec.Command("/bin/bash", "-c", fmt.Sprintf("echo %d > /home/eirik/sanntid-heis-gr28/pid.txt", pid)).Run()

	orderChan := make(chan driver.Order)
	execOrderChan := make(chan driver.Order)

	go driver.Driver(orderChan, execOrderChan)

	wdChan := make(chan string)
	go bcast.Transmitter(57005, "", wdChan)
	wdTimerInterval := 500 * time.Millisecond
	wdTimer := time.NewTimer(wdTimerInterval)

	for {
		select {
		case order := <-orderChan:
			execOrderChan <- order
		case <-wdTimer.C:
			wdChan <- "28-IAmAlive"
			wdTimer.Reset(wdTimerInterval)
		}
	}
}
