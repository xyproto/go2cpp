package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

var testPrograms = []string{
	"hello",
	"multiple",
	"if",
	"contains",
	"prefix",
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

// Run a command and return what has been written to stdout,
// or an empty string.
func Run(cmdString string) string {
	var output bytes.Buffer
	fields := strings.Split(cmdString, " ")
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		return ""
	}
	return output.String()
}

func TestPrograms(t *testing.T) {
	Run("go build")
	for _, program := range testPrograms {
		gofile := "testdata/" + program + ".go"

		// Program output when running with "go run"
		fmt.Println("Compiling and running " + gofile + " with go...")
		output_go := Run("go run " + gofile)

		// Program output when compiling with tinygocompiler and running the executable
		fmt.Println("Compiling and running " + gofile + " with tinygocompiler...")
		Run("./tinygocompiler " + gofile + " -o testdata/" + program)
		output_tgc := Run("testdata/" + program)
		Run("rm testdata/" + program)

		// Check if they are equal
		assertEqual(t, output_go, output_tgc, "tinygocompiler and go run should produce the same output")
	}
}
