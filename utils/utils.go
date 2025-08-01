package utils

import (
	"encoding/json"
	"fmt"
)

func PrintJSon(data interface{}) {
	b, bError := json.Marshal(data)
	if bError != nil {
		fmt.Printf("error marshalling data: %v\n", bError)
		return
	}
	fmt.Println(string(b))
}
