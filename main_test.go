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
	"var",
	"switch",
	"goto",
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
// or an empty string, and what has been written to stderr,
// or an empty string.
func Run(cmdString string) (string, string, error) {
	var outputStdout, outputStderr bytes.Buffer
	fields := strings.Split(cmdString, " ")
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stdout = &outputStdout
	cmd.Stderr = &outputStderr
	err := cmd.Run()
	if err != nil {
		return "", "", err
	}
	return outputStdout.String(), outputStderr.String(), nil
}

func TestPrograms(t *testing.T) {
	Run("go build")
	for _, program := range testPrograms {
		gofile := "testdata/" + program + ".go"

		// Program output when running with "go run"
		fmt.Println("Compiling and running " + gofile + " with go...")
		stdoutGo, stderrGo, err := Run("go run " + gofile)
		if err != nil {
			t.Fatal(err)
		}

		// Program output when compiling with go2cpp and running the executable
		fmt.Println("Compiling and running " + gofile + " with go2cpp...")
		Run("./go2cpp " + gofile + " -o testdata/" + program)
		stdoutTgc, stderrTgc, err := Run("testdata/" + program)
		if err != nil {
			t.Fatal(err)
		}
		Run("rm testdata/" + program)

		// Check if they are equal
		assertEqual(t, stdoutGo, stdoutTgc, "go2cpp and go run should produce the same output on stdout")
		assertEqual(t, stderrGo, stderrTgc, "go2cpp and go run should produce the same output on stderr")
	}
}
