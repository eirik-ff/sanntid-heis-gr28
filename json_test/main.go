package main

import (
	"fmt"
	"encoding/json"
	"strings"
)

type Message1 struct {
	Name string
	Number int
}

type Message2 struct {
	Age int
	LastName string
}

func printJSON(ch <- chan string){

	var m1 Message1
	var m2 Message2
	
	for{
		select{
			case msg := <- ch:

			//Check prefix
			fmt.Printf("Started printJSON\n\n")
			if strings.HasPrefix(string(msg), "Message1") {
				fmt.Println("Recieved: Message1 type")

				//convert from JSON to struct
				err := json.Unmarshal([]byte(msg[len("Message1"):]), &m1);
				if(err != nil){
					fmt.Println("Error in JSON unmarshal")
				}
				fmt.Println(m1)
			} else if strings.HasPrefix(string(msg), "Message2") {
				fmt.Println("Recieved: Message2 type")

				//convert from JSON to struct
				err := json.Unmarshal([]byte(msg[len("Message2"):]), &m2);
				if(err != nil){
					fmt.Println("Error in JSON unmarshal")
				}
				fmt.Println(m2)
			} else {
				fmt.Println("Error: Not correct prefix")
			}
			

		}
	}
}

// Should have a function to add correct prefix to json object.
// func addPrefix

func main() {
	fmt.Println("JSON conversion test")
	jsonChan := make(chan string) //channel for sending JSON objects as string


	//Create two messages of different types
	message1 := Message1{Name: "Daniel", Number: 42}
	message2 := Message2{Age: 100, LastName: "Vedå"}

	go printJSON(jsonChan) //start print function


	//send both messages as strings
	json1, _ := json.Marshal(message1)
	json2, _ := json.Marshal(message2)

	//Add type prefix to json object
	json1 = []byte("Message1"+string(json1))
	json2 = []byte("Message2"+string(json2))
	fmt.Println(string(json1))
	fmt.Println(string(json2))
	
	jsonChan <- string(json1)
	jsonChan <- string(json2)

	fmt.Scanln()
}

//  LocalWords:  JSON
