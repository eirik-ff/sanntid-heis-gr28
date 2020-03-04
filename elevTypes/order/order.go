package order

// Type is a typedef of int
type Type int

const (
	HallUp   Type = 0
	HallDown      = 1
	Cab           = 2
)

// Order is a struct with necessary information to execute an order.
type Order struct {
	TargetFloor int
	Type        Type
	Finished    bool
}
