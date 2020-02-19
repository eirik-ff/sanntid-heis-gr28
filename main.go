package main

import (
	"fmt"
	"log"
	"os"

	"./driver"
)

var (
	logDirPath  string = "/home/student/sanntid-heis-gr28/logs/"
	logFileName string = "elev.log"
	logFilePath string = logDirPath + logFileName
)

func main() {
	err := os.MkdirAll(logDirPath, 0755)
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

	orderChan := make(chan driver.Order)
	execOrderChan := make(chan driver.Order)

	go driver.Driver(orderChan, execOrderChan)

	for {
		select {
		case order := <-orderChan:
			execOrderChan <- order
		}
	}
}
