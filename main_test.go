package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

var testPrograms = []string{
	//"for_range_map_value",
	"hello",
	"multiple",
	"if",
	"contains",
	"prefix",
	"var",
	"switch",
	"goto",
	"for_regular",
	"for_range_single",
	"for_range_list",
	"for_range_both",
	"for_range_map_key",
}

// Programs with unordered words as the output
var unorderedOutput = []string{
	"for_range_map_key",
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
		return outputStdout.String(), outputStderr.String(), err
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
		fmt.Println("Compiling and running " + gofile + " with go2cpp and g++...")
		Run("./go2cpp " + gofile + " -o testdata/" + program)
		stdoutTgc, stderrTgc, err := Run("testdata/" + program)
		if err != nil {
			cmd := "./go2cpp " + gofile + " -O"
			if explanation, _, err := Run(cmd); err != nil {
				fmt.Println("TRANSPIPLATION FAILED:", cmd)
				if strings.Contains(explanation, ": error: ") {

				}
				fmt.Println(explanation)
			} else {
				t.Fatal("go2cpp should not first fail and then succeed! Something is wrong.")
			}
			t.Fatal(errors.New("transpiling failed: " + gofile))
		}
		Run("rm testdata/" + program)

		// For some test-programs, assume the order of the outputted words are random
		// And only check stdout.
		if has(unorderedOutput, program) {
			words1 := strings.Split(strings.TrimSpace(stdoutGo), " ")
			words2 := strings.Split(strings.TrimSpace(stdoutTgc), " ")
			for _, word := range words1 {
				if !has(words2, word) {
					fmt.Println(words2, "DOES NOT HAVE", word)
					assertEqual(t, stdoutGo, stdoutTgc, "go2cpp and go run should produce the same list of words on stdout")
				}
			}
			for _, word := range words2 {
				if !has(words1, word) {
					fmt.Println(words1, "DOES NOT HAVE", word)
					assertEqual(t, stdoutGo, stdoutTgc, "go2cpp and go run should produce the same list of words on stdout")
				}
			}
			return
		}

		// Check if they are equal
		assertEqual(t, stdoutGo, stdoutTgc, "go2cpp and go run should produce the same output on stdout")
		assertEqual(t, stderrGo, stderrTgc, "go2cpp and go run should produce the same output on stderr")
	}
}
