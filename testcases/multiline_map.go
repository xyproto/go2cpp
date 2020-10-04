package main

import (
	"fmt"
	"strings"
)

func main() {
	m := map[string]string{
		"a": `asdf`,
		"b": `a
s
d
f`,
	}
	lines := strings.Split(m["b"], "\n")
	fmt.Println(lines[1]) // should output "s"
}
