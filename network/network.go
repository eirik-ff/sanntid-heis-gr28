package main

import (
	"fmt"
	"time"
	"./bcast"
)

func Network(port int, txChan chan interface{}, rxChans ...interface{}) {
	bcast.InitLogger()

	go bcast.Transmitter(port, txChan)
	go bcast.Receiver(port, rxChans...)
}

//  LocalWords:  JSON
