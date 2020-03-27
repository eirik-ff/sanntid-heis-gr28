# Elevator project
Elevator project in TTK4145 Real-time programming. 

## Diagrams
### Class diagram
![class_diagram](docs/diagrams/class-diagram.svg)

## Modules
### Control
Includes the main control logic for the elevators. 

### Driver


### Elevator
Defines elevator object containing necessary information about the elevator. Also implements methods for the elevator object. 

### Order
Defines order object and order statuses. Also implements methods for the order object.

### Network
#### Bcast
Slightly modified version of the given [Network-go](https://github.com/TTK4145/Network-go) driver.

### Request
Implements functions to select the next order to execute.

### Watchdog
Implements functions to send a message to the watchdog program [Watchdog-go](https://github.com/eirik-ff/watchdog-go/) which monitors this process and respawns it if it dies.

### Main
Runs initial setup and starts the control module. 
