package main

import (
	"fmt"
)

func main() {
	l := []string{"a", "b", "c"}
	for i := range l {
		fmt.Println(i)
	}
}
