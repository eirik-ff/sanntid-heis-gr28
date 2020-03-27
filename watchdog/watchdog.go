package watchdog

import (
	"time"

	"../network/bcast"
)

const (
	// How often to send message to watchdog.
	wdTimerInterval time.Duration = 500 * time.Millisecond
)

var (
	wdChan  chan interface{}
	wdTimer *time.Timer
	message string
)

// Setup initializes the broadcaster and timer. Parameter msg is what is sent
// to the watchdog program.
func Setup(msg string, port int) {
	message = msg
	wdChan = make(chan interface{})
	wdTimer = time.NewTimer(wdTimerInterval)
	go bcast.Transmitter(port, wdChan)
}

// Hungry returns a channel which is filled when timer times out.
func Hungry() <-chan time.Time {
	return wdTimer.C
}

// Feed sends the I'm alive message and restarts the timer.
func Feed() {
	wdChan <- message
	wdTimer.Reset(wdTimerInterval)
}
