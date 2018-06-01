package main

// Tests var and println

var x = 1

var (
	y = 2
	z = 3
)

func main() {
	var (
		a string = "hi"
	)

	var n string

	n = a

	// Outputs to stderr
	println(n, x, y, z)
}
