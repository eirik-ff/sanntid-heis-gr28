package bcast

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"reflect"
	"strings"
	"time"

	"../conn"
)

//Unique ID appended to every message to differentiate out messages from other groups
const UniqueId string = "4242"

//Times to resend a network package
const timesToResendMessage = 10

const networkLogFile = "network.log"

var logFile *os.File
var logger *log.Logger

//function for finding the first null termination in a byte array
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

//Function for initalizeing the
func InitLogger() {
	fmt.Printf("Entering init logger %v\n", logger)
	if logger == nil {
		fmt.Println("Initializing logger")
		usr, err := user.Current()
		if err != nil {
			fmt.Println("Couldn't create user object")
			return
		}
		logDirPath := usr.HomeDir + "/sanntid-heis-gr28/logs/"
		logFilePath := logDirPath + networkLogFile

		err = os.MkdirAll(logDirPath, 0755)
		if err != nil {
			fmt.Printf("Error creating log directory at %s\n", logDirPath)
			return
		}
		logFile, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0655)
		if err != nil {
			fmt.Printf("Error opening info log file at %s\n", logFilePath)
			return
		}
		_ = logFile
		logger = log.New(logFile, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	}
}

func isDuplicate(timestamp string, timestampMap *map[reflect.Type]string, msg string, Type reflect.Type) bool {
	//check if same timestamp is the same as the last received message for this type of message
	//if current timestamp is equal to last time stamp, don't decode message
	if v, ok := (*timestampMap)[Type]; ok {
		if v != timestamp {

			(*timestampMap)[Type] = timestamp
			logger.Println("Received message: " + msg)
		} else {

			logger.Println("Dumped message due to duplicate: " + msg)
			return true
		}

	} else {
		logger.Println("First message of type: " + msg)
		(*timestampMap)[Type] = timestamp
	}
	return false
}

//Network receive routine which can receive JSONs and output them on the correct channel
//based on the type it received
//
// Received message format:
// | UniqueId | TimeStamp | Struct type | Message |
//
func Receiver(port int, outputChans ...interface{}) {
	//the end position of the timestamp in the received message
	const timestampLength = 20

	//open connection
	conn := conn.DialBroadcastUDP(port)

	//create map for storing timestamp of the different types of received messages
	timestampMap := make(map[reflect.Type]string)

	for {
		var buf [1024]byte //receive buffer

		conn.ReadFrom(buf[0:])           //read from network
		for _, ch := range outputChans { //check outputChans against the prefix to check which type of message was received

			//Type of channel
			Type := reflect.TypeOf(ch).Elem()
			typeName := Type.String()

			//prefix to search for
			prefix := UniqueId + typeName

			//extract timestamp
			nanoTimeStamp := string(buf[len(UniqueId) : timestampLength+len(UniqueId)])

			//remove the timestamp from the message
			msg := fmt.Sprintf("%s\n", string(buf[:len(UniqueId)])+string(buf[len(UniqueId)+timestampLength:]))

			//terminate the string correctly
			terminatedMsg := msg[:clen([]byte(msg))]

			if strings.HasPrefix(terminatedMsg[:clen([]byte(msg))], prefix) {

				//if message is duplicate, don't decode the message
				if isDuplicate(nanoTimeStamp, &timestampMap, terminatedMsg, Type) {
					break
				}

				v := reflect.New(Type)

				//convert from json to correct struct type
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

//Takes in a struct and adds a uniqueID and the type of the struct as a prefix.
//Used before transmitting a message over the network
func convertToJsonMsg(msg interface{}) (encodedMsg string) {

	// logger.Println("CONVERT: ", msg)
	json, err := json.Marshal(msg)

	if err != nil {
		logger.Println("Network TX - convertToJsonMsg:", err)
	}

	nanoTime := fmt.Sprintf("%020d", time.Now().UTC().UnixNano())

	prefixedMsg := UniqueId + nanoTime + reflect.TypeOf(msg).String() + string(json) //add uniqueId and type prefix to json
	return prefixedMsg
}

//Routine used to transmit message sent into txChan as a struct
//Adds unique ID and typePrefix
func Transmitter(port int, txChan <-chan interface{}) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	for {
		//wait for msg
		select {
		case msg := <-txChan:

			//convert received struct to json with prefix
			jsonMsg := convertToJsonMsg(msg)

			for i := 0; i < timesToResendMessage; i++ {
				//transmit msg
				conn.WriteTo([]byte(jsonMsg), addr)
				// time.Sleep(1 * time.Millisecond)
			}
		}
	}
}

//  LocalWords:  JSON
