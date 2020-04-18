package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"./control"
	"./watchdog"
)

func setupLog() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(os.Stdout)

	os.Mkdir("logs", 0755)
}

func getPID() int {
	pid := os.Getpid()
	log.Printf("PID: %d\n", pid)
	return pid
}

func setupSignals() chan os.Signal {
	// Capture signals to exit more gracefully
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	return sigs
}

func parseFlags() (port, nfloors, wdPort int, wdMsg string, readFile bool) {
	portF := flag.Int("port", 15657, "Port for connecting to ElevatorServer/SimElevatorServer")
	nfloorsF := flag.Int("floors", 4, "Number of floors per elevator")
	readFileF := flag.Bool("fromfile", false, "Read Elevator struct from file if this flag is passed")
	wdPortF := flag.Int("wd", 57005, "Port to communicate with watchdog program")
	wdMsgF := flag.String("wdmsg", "28-IAmAlive",
		"String to send to watchdog to indicate the program is up and running")
	flag.Parse()

	port = *portF
	nfloors = *nfloorsF
	wdPort = *wdPortF
	wdMsg = *wdMsgF
	readFile = *readFileF
	return
}

func main() {
	elevIOport, nfloors, wdPort, wdMsg, readFile := parseFlags()
	setupLog()
	pid := getPID()
	watchdog.Setup(fmt.Sprintf("%s:%d", wdMsg, pid), wdPort)
	sigs := setupSignals()

	control.Setup(elevIOport, nfloors, readFile)
	control.Loop(sigs)
}
