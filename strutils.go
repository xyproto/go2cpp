package main

import (
	"strings"
)

func lastchar(line string) string {
	if len(line) > 0 {
		return string(line[len(line)-1])
	}
	return ""
}

func has(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

func splitAtAndTrim(s string, poss []int) []string {
	l := make([]string, len(poss)+1)
	startpos := 0
	for i, pos := range poss {
		l[i] = strings.TrimSpace(s[startpos:pos])
		startpos = pos + 1
	}
	l[len(poss)] = strings.TrimSpace(s[startpos:])
	return l
}
