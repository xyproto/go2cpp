package main

import (
	"fmt"
)

func main() {
	l := []string{"a", "b"}
	for i, e := range l {
		fmt.Println(i, e)
	}
}
