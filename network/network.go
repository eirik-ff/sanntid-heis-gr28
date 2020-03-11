package network

import (
	"./bcast"
)

// Network starts the transmitter and receiver threads used for sending and
// receiving orders.
func Network(port int, txChan chan interface{}, rxChans ...interface{}) {
	bcast.InitLogger()

	go bcast.Transmitter(port, txChan)
	go bcast.Receiver(port, rxChans...)
}

//  LocalWords:  JSON
