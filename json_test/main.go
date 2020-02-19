package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"reflect"
)

type Message1 struct {
	Name   string
	Number int
}

type Message2 struct {
	Age      int
	LastName string
}

func printJSON(ch <-chan string) {

	var m1 Message1
	var m2 Message2

	for {
		select {
		case msg := <-ch:

			//Check prefix
			fmt.Printf("Started printJSON\n\n")
			if strings.HasPrefix(string(msg), "Message1") {
				fmt.Println("Recieved: Message1 type")

				//convert from JSON to struct
				err := json.Unmarshal([]byte(msg[len("Message1"):]), &m1)
				if err != nil {
					fmt.Println("Error in JSON unmarshal")
				}
				fmt.Println(m1)
			} else if strings.HasPrefix(string(msg), "Message2") {
				fmt.Println("Recieved: Message2 type")

				//convert from JSON to struct
				err := json.Unmarshal([]byte(msg[len("Message2"):]), &m2)
				if err != nil {
					fmt.Println("Error in JSON unmarshal")
				}
				fmt.Println(m2)
			} else {
				fmt.Println("Error: Not correct prefix")
			}

		}
	}
}

// Encode to JSON and add type prefix
// param: msg interface{} : The message to send as an empty interface. Empty interface allows
//                          generic data types to be used
func encodeJson(msg interface{}){


	fmt.Printf("\n\nStarted encodejson\n")

	//Encode message as JSON 
	msgJson, err := json.Marshal(msg)

	if(err != nil){
		fmt.Println("Error: encodeJson failed at Marshal")
	}

	//Check which data type the message is
	msgType := string(reflect.TypeOf(msg).Name())
	fmt.Printf("EncodeJson: Type: %s, length: %d\n\n", msgType, len(msgType))

	msgJson = []byte(msgType+string(msgJson))
	fmt.Printf("Encoded Json with prefix:\n%s\n\n", string(msgJson))
	//Add prefix based on datatype
	
}

// Decode JSON to correct type based on prefix
// param msg interface - Prefixed json message to decode
// Might need a list of all types available to decode to
func decodeJson(msg interface{}, typeMap interface{}){
	fmt.Printf("\n\nStarted decodeJSON\n\n")

	//Assert that msg is a byte array
	if msg1, ok := msg.([]byte); ok{
		//Read prefix to check which type

		msgType := strings.SplitN(string(msg1), "{", 1)
		fmt.Printf("Decoded JSON:\nmsg:%s\n", string(msg1))
		fmt.Printf("Decoded JSON:\nmsgType:%s\n", msgType)
	}
	// if strings.HasPrefix(string(msg), "Message1") {	
	//Unmarshal to that type
	//return correct datatype
}

func main() {
	fmt.Println("JSON conversion test")
	jsonChan := make(chan string) //channel for sending JSON objects as string


	//Create two messages of different types
	message1 := Message1{Name: "Daniel", Number: 42}
	message2 := Msg{Age: 100, LastName: "Vedå"}

	go printJSON(jsonChan) //start print function


	//send both messages as strings
	json1, _ := json.Marshal(message1)
	json2, _ := json.Marshal(message2)

	//Add type prefix to json object
	json1 = []byte("Message1" + string(json1))
	json2 = []byte("Message2" + string(json2))
	fmt.Println(string(json1))
	fmt.Println(string(json2))

	jsonChan <- string(json1)
	jsonChan <- string(json2)

	fmt.Scanln()	 
}

//  LocalWords:  JSON
