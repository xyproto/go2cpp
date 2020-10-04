package main

import "fmt"

const s1 = `
a
b
c
`

var s2 = `
a
b
c
`

const s3 = `abc`
var s4 = `abc`

func main() {
	s5 := `abc`
	s6 := `
	a
	b
	c
	`
	fmt.Println(s1, s2, s3, s4, s5, s6)
}
