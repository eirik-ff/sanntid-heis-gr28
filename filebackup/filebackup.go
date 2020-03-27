package filebackup

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"../elevTypes/elevator"
)

func Read(fileName string, Nfloors, Nbuttons int) elevator.Elevator {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("No file with path '%s' exists\n", fileName)
		return elevator.NewElevator(Nfloors, Nbuttons)
	}
	defer file.Close()

	var elev elevator.Elevator
	data, _ := ioutil.ReadAll(file)
	err = json.Unmarshal([]byte(data), &elev)
	if err != nil {
		log.Println("Error converting backup JSON to elevator object.")
		return elevator.NewElevator(Nfloors, Nbuttons)
	}

	log.Println("Read old configuration from file")
	log.Println(elev.ToString())
	log.Println(elev.OrderMatrixToString())
	return elev
}

func Write(fileName string, elev elevator.Elevator) {
	os.Remove(fileName)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error creating backup file")
		return
	}
	defer file.Close()

	msg, err := json.Marshal(elev)
	if err != nil {
		log.Println("Error converting elevator object to JSON.")
		return
	}
	if _, err := file.Write([]byte(msg)); err != nil {
		log.Println("Error writing elevator JSON to backup file.")
		return
	}
}
