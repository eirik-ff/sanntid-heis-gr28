package order

import (
	"fmt"
)

const (
	// OrderTimeout is how long an order can be Taken before a Finished message
	// must be received.
	OrderTimeout int64 = 10 // seconds
)

// Type is a typedef of int
type Type int

const (
	HallUp   Type = 0
	HallDown Type = 1
	Cab      Type = 2
)

// Status is typedef of int
type Status int

const (
	// Invalid exists because the default value of int is zero,
	// but the status field must be set to the desired status manually.
	Invalid Status = 0
	// NotTaken status is set if an order is received, but no one has
	// broadcasted that they are taking this order
	NotTaken Status = 1
	// Taken status is set if someone has broadcasted that they are taking
	// this order
	Taken Status = 2
	// Execute status is set when the order should be executed by this local
	// elevator
	Execute Status = 3
	// Finished status is assigned to orders which have been executed and
	// are finisehd (duh).
	Finished Status = 4
)

// Order is a struct with necessary information to execute an order.
type Order struct {
	// Floor is target floor of order.
	Floor int
	// Type is which type of order, can be HallUp, HallDown, or Cab.
	Type Type
	// Status is status of order, see status defines..
	Status Status
	// LocalTimeStamp is used to check if an order that is marked as taken is
	// not forgotten about.
	LocalTimeStamp int64
}

// ToString converts the order to a readable string.
func (o *Order) ToString() string {
	typeStr := ""
	switch o.Type {
	case 0:
		typeStr = "HallUp"
	case 1:
		typeStr = "HallDown"
	case 2:
		typeStr = "Cab"
	}

	statusStr := ""
	switch o.Status {
	case Invalid:
		statusStr = "Invalid"
	case NotTaken:
		statusStr = "NotTaken"
	case Taken:
		statusStr = "Taken"
	case Execute:
		statusStr = "Execute"
	case Finished:
		statusStr = "Finished"
	}

	return fmt.Sprintf("Order:{floor:%d type:'%s' status:'%s' timeout:'%d'}", o.Floor, typeStr, statusStr, o.LocalTimeStamp)
}

// CompareEq checks if o1 == o2 but doesn't check LocalTimeStamp
func CompareEq(o1, o2 Order) bool {
	return o1.Floor == o2.Floor && o1.Type == o2.Type && o1.Status == o2.Status
}

// CompareFloorAndType checks if only floor and type of o1 and o2 are equal
func CompareFloorAndType(o1, o2 Order) bool {
	return o1.Floor == o2.Floor && o1.Type == o2.Type
}
