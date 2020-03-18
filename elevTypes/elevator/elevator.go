package elevator

import (
	"fmt"
	"time"
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


func (elev *Elevator) CheckOrderTimestamp(timeoutChan chan <- order.Order) {

	fmt.Println(elev.OrderMatrixToString())
	
	//Loop through all orders in matrix
    for f := 0; f < len(elev.Orders); f++ {
        for t := range elev.Orders[f] {

			currentTime := time.Now().Unix() //get current time

			//check if order is taken
			if elev.Orders[f][t].Status == order.Taken {
				//Check if order is timed out
				if  currentTime  >= elev.Orders[f][t].LocalTimeStamp {
					fmt.Printf("Order %s timed out. Current time :'%d' \n", elev.Orders[f][t].ToString(), currentTime) 
					timeoutChan <- elev.Orders[f][t] //send the timeout order to main
				}
			}
        }
    }
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

// SameOrderInMatrix checks if the same order with
func (elev *Elevator) SameOrderInMatrix(o order.Order) bool {
	for f := range elev.Orders {
		for t := range elev.Orders[f] {
			curr := elev.Orders[f][t]
			if order.CompareEq(curr, o) {
				return true
			}
		}
	}
	return false
}

// AssignOrderToMatrix modifies the order matrix to set the argument order `ord`
// to that orders floor and type in the matrix.
func (elev *Elevator) AssignOrderToMatrix(ord order.Order) {
	elev.Orders[ord.Floor][ord.Type] = ord
}
