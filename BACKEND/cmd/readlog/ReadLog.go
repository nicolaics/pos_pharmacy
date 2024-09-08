package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	var fileName string

	fmt.Print("input full file name: ")
	fmt.Scanf("%s", &fileName)

	jsonData, err :=  os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	
	var data string

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println(data)
}