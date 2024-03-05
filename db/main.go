package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// FOR WHATEVER REEASON MY SECTOR WASNT IN HERE YET. SO I DIDNT FINISH THIS CODE

func main() {
	// this comes from api.startraders.io/v2/systems.json
	data, err := os.ReadFile("./db/fixtures/systems.json")
	if err != nil {
		log.Fatal(err)
	}

	var x []map[string]interface{}
	json.Unmarshal(data, &x)

	for _, sector := range x {
		// if sector["symbol"] != "X1" {
		// 	continue
		// }
		fmt.Println(sector["symbol"])
	}

}
