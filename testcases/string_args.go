package main

import "fmt"

func threecat(a, b, c string) string {
	return a + b + c
}

func main() {
	threecat("in", "cred", "ible")
	fmt.Println("hi")
}
