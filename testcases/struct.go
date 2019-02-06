package main

import (
	"fmt"
)

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

type V3 Vec3

func main() {
	v := &V3{1.2, 3.4, 5.6}
	fmt.Println(v)
}
