package elevator

import (
	"fmt"

	"../order"
)

// MotorDirection is a typedef of elevio.MotorDirection to be able
// to use it in packages that include driver.
type Direction int

const (
	Up   Direction = 1
	Down Direction = -1
	Stop Direction = 0
)

type State int

const (
	Init     State = 0
	Idle     State = 1
	Moving   State = 2
	DoorOpen State = 3
	Error    State = 4
)

type Elevator struct {
	ActiveOrder order.Order
	Floor       int
	Direction   Direction
	State       State

	Nfloors  int
	Nbuttons int
	Orders   [][]order.Order
	// TODO: bounds check on index when accessing? if two elevators have
	// 		 different number of floors this will be necessary.
	// 		 maybe need bound check to be fault tolerant?
}

func NewElevator(nfloors, nbuttons int) Elevator {
	var elev Elevator
	elev.Nfloors = nfloors
	elev.Nbuttons = nbuttons

	elev.Orders = make([][]order.Order, nfloors)
	for i := range elev.Orders {
		elev.Orders[i] = make([]order.Order, nbuttons)
	}

	elev.State = Idle
	return elev
}

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
