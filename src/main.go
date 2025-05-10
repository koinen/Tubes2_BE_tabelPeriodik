package main

import (
	"fmt"
)

func main() {
	jsonBytes, err := convertToJson(scrape())
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		return
	}
	serve(jsonBytes)
}
