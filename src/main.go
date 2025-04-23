package main

import (
	"fmt"
)

func main() {
	element := "Egg"
	recipes := findRecipe(element)
	for _, recipe := range recipes {
		fmt.Println(element, ":", recipe[0], "+", recipe[1])
	}
}
