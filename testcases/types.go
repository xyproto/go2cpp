package main

import (
	"fmt"
)

type u8 uint8

type (
	f64 float64
	f32 float32
	i16 int16
)

type snakestring string

func main() {
	var sss snakestring = "sssssss"
	var (
		n u8 = 255
	)
	var i int = int(n)
	fmt.Println(sss, i)
}
