package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "    hello\t   \n "
	fmt.Println("|" + strings.TrimSpace(s) + "|")
}
