package order

// Type is a typedef of int
type Type int

const (
	HallUp   Type = 0
	HallDown      = 1
	Cab           = 2
)

// Status is typedef of int
type Status int

const (
	// Abort status is when an elevator on the network sends a message
	// with better cost and you should abort the currently active order.
	Abort Status = -1
	// Invalid exists because the default value of int is zero,
	// but the status field must be set to the desired status manually.
	Invalid Status = 0
	// InitialBroadcast status is for when an order comes locally and is
	// broadcast to the other elevators.
	InitialBroadcast Status = 1
	// NotTaken status is set if an order is received, but no one has
	// broadcasted that they are taking this order
	NotTaken Status = 2
	// Taken status is set if someone has broadcasted that they are taking
	// this order
	Taken Status = 3
	// Execute status is set when the order should be executed by this local
	// elevator
	Execute Status = 4
	// Finished status is assigned to orders which have been executed and
	// are finisehd (duh).
	Finished Status = 5
)

// Order is a struct with necessary information to execute an order.
type Order struct {
	Floor  int
	Type   int
	Status Status
}
