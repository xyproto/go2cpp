package main

import (
	"fmt"
)

const a = 128

const (
	b = 256
	c = 386
)

const d = 486

func main() {
	const e = 586
	const (
		f int     = 42
		g float64 = 3.14
	)
	fmt.Println(a, b, c, d, e, f, g)
}
