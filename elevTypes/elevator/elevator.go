package elevator

import (
	"fmt"
	"time"

	"../order"
)

// Direction is the valid directions the elevator can move.
type Direction int

const (
	Up   Direction = 1
	Down Direction = -1
	Stop Direction = 0
)

// State is the valid states the elevator can be in.
type State int

const (
	Init     State = 0
	Idle     State = 1
	Moving   State = 2
	DoorOpen State = 3
	Error    State = 4
)

// Elevator is a struct of variables key to controlling the elevator.
type Elevator struct {
	ActiveOrder order.Order
	Floor       int
	Direction   Direction
	State       State

	Nfloors  int
	Nbuttons int
	Orders   [][]order.Order
}

// CheckOrderTimestamp checks the Orders matrix for orders where the current
// time is passed the stored timeout time, and pushes all timed out orders onto
// timeoutChan channel.
func (elev *Elevator) CheckOrderTimestamp(timeoutChan chan<- order.Order) {
	//Loop through all orders in matrix
	for f := 0; f < len(elev.Orders); f++ {
		for t := range elev.Orders[f] {

			currentTime := time.Now().Unix() //get current time

			//check if order is taken
			if elev.Orders[f][t].Status == order.Taken {
				//Check if order is timed out
				if currentTime >= elev.Orders[f][t].LocalTimeStamp {
					fmt.Printf("Order %s timed out. Current time :'%d' \n", elev.Orders[f][t].ToString(), currentTime)
					timeoutChan <- elev.Orders[f][t] //send the timeout order to main
				}
			}
		}
	}
}

// NewElevator creates a new elevator object and initializes its order matrix.
// This is the prefered way of creating a new elevator object.
func NewElevator(nfloors, nbuttons int) Elevator {
	var elev Elevator
	elev.Nfloors = nfloors
	elev.Nbuttons = nbuttons

	elev.Orders = make([][]order.Order, nfloors)
	for i := range elev.Orders {
		elev.Orders[i] = make([]order.Order, nbuttons)
	}

	elev.State = Idle
	elev.ActiveOrder.Status = order.Finished
	return elev
}

// ToString creates a string representation of an elevator object.
func (elev *Elevator) ToString() string {
	dirStr := fmt.Sprintf("Invalid (%d)", elev.Direction)
	switch elev.Direction {
	case Up:
		dirStr = "Up"
	case Down:
		dirStr = "Down"
	case Stop:
		dirStr = "Stop"
	}

	stateStr := ""
	switch elev.State {
	case Init:
		stateStr = "Init"
	case Idle:
		stateStr = "Idle"
	case Moving:
		stateStr = "Moving"
	case DoorOpen:
		stateStr = "DoorOpen"
	case Error:
		stateStr = "Error"
	}

	return fmt.Sprintf("Elevator:{%s floor:%d dir:'%s' state:'%s'}",
		elev.ActiveOrder.ToString(), elev.Floor, dirStr, stateStr)
}

// OrderMatrixToString creates a string representation of the order matrix. The
// format is XXX XXX XXX XXX where each X is a digit representing that orders
// status. The first batch is floor 0 with hall up, hall down, cab order of the
// digits. The second batch is floor 1, and so on.
func (elev *Elevator) OrderMatrixToString() string {
	s := ""
	for f := 0; f < elev.Nfloors; f++ {
		for i := range elev.Orders[f] {
			s += fmt.Sprintf("%d", elev.Orders[f][i].Status)
		}
		s += " "
	}
	return s
}

// AssignOrderToMatrix modifies the order matrix to set the argument order `ord`
// to that orders floor and type in the matrix.
func (elev *Elevator) AssignOrderToMatrix(ord order.Order) {
	elev.Orders[ord.Floor][ord.Type] = ord
}
