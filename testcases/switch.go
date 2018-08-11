package main

import (
	"fmt"
)

func main() {
	fruit := "apple"
	switch fruit {
	case "banana":
		fmt.Println("it's a banana")
	case "apple":
		fallthrough
	default:
		fmt.Println("probably apple")
	}
}
