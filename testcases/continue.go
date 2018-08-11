package main

import (
	"fmt"
)

func main() {
	for i := 0; i <= 42; i++ {
		if i < 42 {
			continue
		}
		fmt.Println("THE ANSWER =", i)
	}
}
