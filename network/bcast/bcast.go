package bcast

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"reflect"
	"strconv"
	"strings"
	"time"

	"../conn"
)

// UniqueID appended to every message to differentiate out messages from other groups
const UniqueID string = "4242"

// Times to resend a network package
const timesToResendMessage = 10

const networkLogFile = "network.log"

var logFile *os.File
var logger *log.Logger

func logReceive(s string) {
	//Filter out IAmAlive messages
	if !strings.Contains(s, "IAmAlive") {
		logger.Println("Received message: " + s)
	}
}

func logSending(s string) {
	//Filter out IAmAlive messages
	if !strings.Contains(s, "IAmAlive") {
		logger.Println("Sending message: " + s)
	}
}

// function for finding the first null termination in a byte array
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

//Function for initalizeing the
func InitLogger(port int) {
	fmt.Printf("Entering init logger %v\n", logger)
	if logger == nil {
		fmt.Println("Initializing logger")
		usr, err := user.Current()
		if err != nil {
			fmt.Println("Couldn't create user object")
			return
		}
		logDirPath := usr.HomeDir + "/sanntid-heis-gr28/logs/"
		logFilePath := logDirPath + fmt.Sprintf("port_%d_", port) + networkLogFile

		err = os.MkdirAll(logDirPath, 0755)
		if err != nil {
			fmt.Printf("Error creating log directory at %s\n", logDirPath)
			return
		}
		logFile, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0655)
		if err != nil {
			fmt.Printf("Error opening info log file at %s\n", logFilePath)
			return
		}
		_ = logFile
		logger = log.New(logFile, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	}
}

// check if same timestamp is the same as the last received message for this type of message
// if current timestamp is equal to last time stamp, don't decode message
func isDuplicate(timestamp string, timestampMap *map[reflect.Type]string, msg string, Type reflect.Type) bool {
	if v, ok := (*timestampMap)[Type]; ok {
		if v != timestamp {

			(*timestampMap)[Type] = timestamp
			logReceive(msg)
		} else {
			return true
		}

	} else {
		logReceive(msg)
		(*timestampMap)[Type] = timestamp
	}
	return false
}

// Receiver routine which can receive JSONs over the network and output them on
// the correct channel based on the type it received.
//
// Received message format:
// | UniqueID | TimeStamp | PID | Struct type | Message |
//
// Note: the PID is of the sending process. It's used to filter out messages so
// 	     they are not sent to the sending process.
func Receiver(port int, outputChans ...interface{}) {
	// the end position of the timestamp in the received message
	const timestampLength = 20
	const pidLength = 6
	pid := os.Getpid()

	// open connection
	conn := conn.DialBroadcastUDP(port)

	// create map for storing timestamp of the different types of received messages
	timestampMap := make(map[reflect.Type]string)

	for {
		var buf [1024]byte // receive buffer

		conn.ReadFrom(buf[0:])           // read from network
		for _, ch := range outputChans { // check outputChans against the prefix to check which type of message was received
			Type := reflect.TypeOf(ch).Elem() // Type of channel
			typeName := Type.String()

			prefix := UniqueID + typeName                                               // prefix to search for
			nanoTimeStamp := string(buf[len(UniqueID) : timestampLength+len(UniqueID)]) // extract timestamp
			pidStr := string(buf[timestampLength+len(UniqueID) : timestampLength+len(UniqueID)+pidLength])
			recvPid, err := strconv.Atoi(pidStr)
			if err != nil {
				logger.Printf("Received pid string '%s' couldn't be converted\n", pidStr)
			} else if recvPid == pid {
				break // do not receive your own messages
			}

			// remove the timestamp and pid from the message
			msg := string(buf[:len(UniqueID)]) + string(buf[len(UniqueID)+timestampLength+pidLength:])
			terminatedMsg := msg[:clen([]byte(msg))] // remove trailing zero bytes

			if strings.HasPrefix(terminatedMsg[:clen([]byte(msg))], prefix) {
				if isDuplicate(nanoTimeStamp, &timestampMap, terminatedMsg, Type) {
					break // if message is duplicate, don't decode the message
				}

				// convert from json to correct struct type
				v := reflect.New(Type)
				json.Unmarshal([]byte(terminatedMsg[len(prefix):clen([]byte(msg))]), v.Interface())

				reflect.Select([]reflect.SelectCase{{
					Dir:  reflect.SelectSend,
					Chan: reflect.ValueOf(ch),
					Send: reflect.Indirect(v),
				}})
			}
		}
	}
}

// Takes in a struct and adds a uniqueID and the type of the struct as a prefix.
// Used before transmitting a message over the network
func convertToJSONMsg(msg interface{}) string {

	json, err := json.Marshal(msg)

	if err != nil {
		logger.Println("Network TX - convertToJSONMsg:", err)
	}

	return reflect.TypeOf(msg).String() + string(json)
}

func prefixMsg(msg string) string {
	nanoTime := fmt.Sprintf("%020d", time.Now().UTC().UnixNano())
	pid := fmt.Sprintf("%06d", os.Getpid())

	prefixedMsg := UniqueID + nanoTime + pid + msg
	return prefixedMsg
}

// Transmitter routine used to transmit message sent into txChan as a struct
// Adds unique ID and typePrefix.
func Transmitter(port int, txChan <-chan interface{}) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	for {
		// wait for msg
		select {
		case msg := <-txChan:

			// convert received struct to json with prefix
			jsonMsg := convertToJSONMsg(msg)
			logSending(jsonMsg)
			jsonMsg = prefixMsg(jsonMsg)

			for i := 0; i < timesToResendMessage; i++ {
				// transmit msg
				conn.WriteTo([]byte(jsonMsg), addr)
				// time.Sleep(1 * time.Millisecond)
			}
		}
	}
}
