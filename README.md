# Elevator project
Elevator project in TTK4145 Real-time programming. 

## How to run 
**Note**: This only works on Linux. If you're on Windows, good luck figuring out how to do this. 

To compile the necessary programs, run `make buildall`. 

To run the elevator without the watchdog, run `make runN` where `N` is `1`, `2`, or `3`. 

To run the elevator _with_ the watchdog, run `make startN` where `N` is `1`, `2`, or `3`. This will start the watchdog which in turn will start the elevator after around 5 seconds.

To see the output of the elevator when running with the watchdog, use `tail -f logs/heisM.log` where `M` is `57005` for `start1`, `57006` for `start2` and `57007` for `start3`. 

## Modules
### Control
Includes the main control logic for the elevators. 

### Driver
Handles the communication with the elevator server (or simulator) and take care of the floor lights.

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
Implements functions to send a message to the watchdog program (added as git submodule to this repository) [Watchdog-go](./watchdog-go-submod/README.md) which monitors this process and respawns it if it dies.

### Main
Runs initial setup and starts the control module. 
