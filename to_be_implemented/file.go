package to_be_implemented

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"../driver"
)

var (
	newFile *os.File
	err     error
)

func writeToFile(matrix driver.Elevator) {

	os.Remove("elevBackupFile.log")

	file, _ := os.OpenFile("elevBackupFile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	defer file.Close()

	msg, _ := json.Marshal(matrix)

	if _, err := file.Write([]byte(msg)); err != nil {
		log.Fatal(err)
	}

	file.Close()

}

//This has been moved to main.go
func readFromFile() driver.Elevator {

	file, _ := os.Open("currentMatrix.txt")

	data, _ := ioutil.ReadAll(file)

	var elev driver.Elevator

	json.Unmarshal([]byte(data), &elev)

	file.Close()

	return elev

}
