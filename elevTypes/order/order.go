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
	// UndefinedStatus exists because the default value of int is zero,
	// but the status field must be set to the desired status manually.
	UndefinedStatus Status = 0
	// InitialBroadcast status is for when an order comes locally and is
	// broadcast to the other elevators.
	InitialBroadcast = 1
	// LowerCostReply status is a received order message from an elevator
	// that have lower cost than the elevator which received the original
	// order.
	LowerCostReply = 2
	// Finished status is assigned to orders which have been executed and
	// are finisehd (duh).
	Finished = 3
	// Abort status is when an elevator on the network sends a message
	// with better cost and you should abort the currently active order.
	Abort = 4
)

// Order is a struct with necessary information to execute an order.
type Order struct {
	ID          int64 // time.Time.Unix() return type
	TargetFloor int
	Type        Type
	Cost        int
	Status      Status
}
