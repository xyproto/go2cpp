package main

import (
	"fmt"
)

func d1() {
	fmt.Println("defer 1, outputted last")
}

func main() {
	fmt.Println("start of main function")
	defer d1()
	defer func() {
		fmt.Println("defer 2, outputted second")
	}()
	defer func() { fmt.Println("defer 3, outputted last") }()
	fmt.Println("end of main function")
}
