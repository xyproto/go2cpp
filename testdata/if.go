package main

import "fmt"

// Tests if, else if, else, semicolons in Go, imports without parenthesis and fmt.Print

func main() {
	x := 4
	fmt.Print("The number + 3 is... ")
	if x+3 > 7 {
		fmt.Print("larger than seven and ")
	} else {
		fmt.Print("smaller or equal to seven and ")
	}
	if x+3 > 7 {
		fmt.Println("larger than seven.")
	} else if x < 7 {
		fmt.Println("smaller than seven.")
	} else {
		fmt.Println("smaller or equal to seven.")
	}
}
