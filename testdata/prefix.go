package main

import (
	"fmt"
	"strings"
)

// Test strings.HasPrefix

func main() {
	fmt.Println(strings.HasPrefix("asdfqwerty", "asdf"))
	fmt.Println(strings.HasPrefix("asdfqwerty", "qwerty"))
}
