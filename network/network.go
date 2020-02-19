package network

import (
	"fmt"
	"encoding/json"
	"strings"
	"reflect"
	"net"
	"./conn"
)


const UniqueId string = "4242"

// type Message1 struct {
// 	Name string
// 	Number int
// }

// type Message2 struct {
// 	Age int
// 	LastName string
// }

//Network receive routine which can receive JSONs and output them on the correct channel
//based on the type it received
func Receiver(port int, outputChans ...interface{}){

	//open connection
	conn := conn.DialBroadcastUDP(port)

	var buf [1024]byte //receive buffer
	for{
		n, _, _ := conn.ReadFrom(buf[0:]) //read from network
		// fmt.Printf("Network Received: got %s \n", buf)
		for _, ch := range outputChans { //check outputChans against the prefix to check which type of message was received
			Type := reflect.TypeOf(ch).Elem()
			typeName := Type.String()
			prefix := UniqueId + typeName
			if strings.HasPrefix(string(buf[0:n]), prefix) {
				v := reflect.New(Type)

				json.Unmarshal([]byte(buf[len(prefix):n]), v.Interface())
				
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
func convertToJsonMsg (msg interface{}) (encodedMsg string){

	// fmt.Println("CONVERT: ", msg)
	json, err := json.Marshal(msg)

	if err != nil {
		fmt.Println("Network TX - convertToJsonMsg:", err)
	}
	
	// fmt.Println("CONVERT JSON: ", string(json))
	prefixedMsg := UniqueId +reflect.TypeOf(msg).String() + string(json) //add uniqueId and type prefix to json
	// fmt.Println("CONVERT JSON prefix: ", prefixedMsg)
	return prefixedMsg
}



//Routine used to transmit message sent into txChan as a struct
//Adds unique ID and typePrefix
func Transmitter(port int, txChan <- chan interface{}){
	
		conn := conn.DialBroadcastUDP(port)
		addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
		// conn, err := net.DialUDP("udp", nil, &send)
		// if err != nil {
		// 	panic(err)
		// }
		// defer conn.Close()

		for{
			//wait for msg
			select{
				case msg := <- txChan:
				// fmt.Println("TX: ", msg)
				//convert received struct to json with prefix
				jsonMsg := convertToJsonMsg(msg)
				// fmt.Println("TX JSON: ", jsonMsg)
				//transmit msg
				conn.WriteTo([]byte(jsonMsg), addr)
				break
			}
			
		}
	}

//  LocalWords:  JSON
