package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"./network/bcast"
	"./network/localip"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

type Order struct {
	Floor     int
	Direction int
}

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	generalTx := make(chan string)
	generalRx := make(chan string)
	go bcast.Transmitter(54321, generalTx)
	go bcast.Receiver(54321, generalRx)

	// The example message. We just send one of these every second.
	go func() {
		var iter int
		for {
			iter++
			generalTx <- "Hello from " + id + "   " + strconv.Itoa(iter)
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case a := <-generalRx:
			fmt.Printf("Received: %#v\n", a)
		}
	}
}
