package main

import (
	"fmt"
)

func main() {
	m := map[string]string{"first": "hi", "second": "you", "third": "there"}
	first := true
	for k := range m {
		if first {
			first = false
		} else {
			fmt.Print(" ")
		}
		fmt.Print(k)
	}
	fmt.Println()
	first = true
	for _, v := range m {
		if first {
			first = false
		} else {
			fmt.Print(" ")
		}
		fmt.Print(v)
	}
	fmt.Println()
	for k, v := range m {
		fmt.Print(k + v)
	}
	fmt.Println()
}
