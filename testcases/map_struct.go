package main

import (
	"fmt"
)

type Creature struct {
	X    float64
	Y    float64
	name string
}

func main() {
	o := map[int]*Creature{
		0: {1.2, 3.4, "Bob"},
		1: {4.5, 6.7, "Alice"},
	}
	fmt.Println(o[1])
}
