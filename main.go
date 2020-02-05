package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
	go bcast.Transmitter(54321, "28", generalTx)
	go bcast.Receiver(54321, "28", generalRx)

	// The example message. We just send one of these every second.
	go func() {
		msg := HelloMsg{Message: "Hello from " + id, Iter: 0}
		for {
			msg.Iter++

			// will be its own function
			jsonbyte, _ := json.Marshal(msg)
			send := string(jsonbyte)

			generalTx <- string(send)
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case a := <-generalRx:
			var recv HelloMsg
			json.Unmarshal([]byte(a), &recv)
			fmt.Printf("Received: %s %d\n", recv.Message, recv.Iter)
		}
	}
}
