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

	// Outputs to stderr
	println(a, x, y, z)
}
