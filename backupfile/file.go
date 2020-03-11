package backupfile

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"../elevTypes/order"
)

var (
	newFile *os.File
	err     error
)

func writeToFile(matrix order.Elevator) {

	os.Remove("currentMatrix.txt")

	file, _ := os.OpenFile("currentMatrix.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	defer file.Close()

	msg, _ := json.Marshal(matrix)

	if _, err := file.Write([]byte(msg)); err != nil {
		log.Fatal(err)
	}

	file.Close()

}

func readFromFile() order.Elevator {

	file, _ := os.Open("currentMatrix.txt")

	data, _ := ioutil.ReadAll(file)

	var matrix order.Elevator

	json.Unmarshal([]byte(data), &matrix)

	file.Close()

	return matrix

}
