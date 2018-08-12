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

type string snakestring

func main() {
	var sss snakestring = "sssssss"
	var (
		n u8 = 255
	)
	fmt.Println(n, snakestring)
}
