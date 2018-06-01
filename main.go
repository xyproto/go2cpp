package main

// Plan:
// 1. Read in the source code
// 2. Convert it to C++17
// 3. Compile it

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const tupleType = "std::tuple"

func LiteralStrings(source string) (output string) {
	output = source
	replacements := map[string]string{
		"\")": "\"s)",
		"\";": "\"s;",
		"\",": "\"s,",
	}
	hasLiteral := false
	for k, v := range replacements {
		if strings.Contains(output, k) {
			output = strings.Replace(output, k, v, -1)
			hasLiteral = true
		}
	}
	if hasLiteral {
		output = "\nusing namespace std::string_literals;\n" + output
	}
	return output
}

func WholeProgramReplace(source string) (output string) {
	output = source
	replacements := map[string]string{
		" string ": " std::string ",
		"(string ": "(std::string ",
	}
	for k, v := range replacements {
		output = strings.Replace(output, k, v, -1)
	}
	return output
}

func between(s, a, b string) string {
	apos := strings.Index(s, a)
	if apos == -1 {
		return s
	}
	bpos := strings.Index(s, b)
	if bpos == -1 {
		return s
	}
	return s[apos+len(a) : bpos]
}

func FunctionArguments(source string) (output string) {
	output = source
	if strings.Contains(output, ",") {
		currentName := ""
		currentType := ""
		args := strings.Split(output, ",")
		for i := len(args) - 1; i >= 0; i-- {
			strippedArg := strings.TrimSpace(args[i])
			//fmt.Println(i, strippedArg)
			if strings.Contains(strippedArg, " ") {
				elems := strings.SplitN(strippedArg, " ", 2)
				currentName = elems[0]
				currentType = elems[1]
			} else {
				currentName = strippedArg
			}
			newArgs := " " + currentType + " " + currentName
			output = strings.Replace(output, args[i], newArgs, -1)
		}
	} else if strings.Contains(output, " ") {
		words := strings.Split(output, " ")
		output = strings.TrimSpace(words[1]) + " " + strings.TrimSpace(words[0])
	}
	return strings.TrimSpace(output)
}

func FunctionRetvals(source string) (output string) {
	if len(strings.TrimSpace(source)) == 0 {
		return source
	}
	output = source
	if strings.Contains(output, "(") {
		s := between(output, "(", ")")
		output = "(" + FunctionArguments(s) + ")"
	}
	return strings.TrimSpace(output)
}

// Picks out the types given a list of C++ arguments with name and type
func CPPTypes(args string) string {
	words := strings.Split(between(args, "(", ")"), ",")
	var atypes []string
	for _, word := range words {
		elems := strings.Split(strings.TrimSpace(word), " ")
		atypes = append(atypes, elems[0])
	}
	return strings.Join(atypes, ", ")
}

func FunctionSignature(source string) (output, returntype, name string) {
	if len(strings.TrimSpace(source)) == 0 {
		return source, "", ""
	}
	output = source
	args := FunctionArguments(between(output, "(", ")"))
	rets := FunctionRetvals(between(output, ")", "{"))
	if strings.Contains(rets, ",") {
		// Multiple return
		rets = tupleType + "<" + CPPTypes(rets) + ">"
	}
	name = between(output, "func ", "(")
	if name == "main" {
		rets = "int"
	}
	output = "auto " + name + "(" + args + ") -> " + rets + " {"
	return strings.TrimSpace(output), rets, name
}

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

var endings = []string{
	"{", ",", "}",
}

func hasInt(ints []int, x int) bool {
	for _, z := range ints {
		if z == x {
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

func PrintStatement(source string) (output string) {
	if !strings.Contains(source, "(") {
		// Invalid print line, no function call
		return output
	}
	elems := strings.SplitN(source, "(", 2)
	name := strings.TrimSpace(elems[0])
	args := strings.TrimSpace(elems[1])
	if strings.HasSuffix(args, ")") {
		args = args[:len(args)-1]
	}
	output = "std::cout << "
	// Don't split on commas that are within paranthesis or quotes
	withinPar := 0
	withinQuot := false
	commaPos := []int{}
	for i, c := range args {
		if c == '(' {
			withinPar++
		} else if c == ')' {
			withinPar--
		} else if c == '"' {
			withinQuot = !withinQuot
		} else if c == ',' && (withinPar == 0) && (!withinQuot) {
			commaPos = append(commaPos, i)
		}
	}
	//fmt.Println(args)
	//fmt.Println(commaPos)
	if len(commaPos) > 0 {
		parts := splitAtAndTrim(args, commaPos)
		//fmt.Println(parts)
		s := strings.Join(parts, " << \" \" << ")
		//fmt.Println(s)
		output += s
	} else {
		output += args
	}
	// Println, println, Fprintln etc should end with << std::endl
	if strings.HasSuffix(name, "ln") {
		output += " << std::endl"
	}
	return output
}

func AddIncludes(source string) (output string) {
	output = source
	includes := map[string]string{
		"std::tuple":  "tuple",
		"std::endl":   "iostream",
		"std::cout":   "iostream",
		"std::string": "string",
	}
	includeString := ""
	for k, v := range includes {
		if strings.Contains(output, k) {
			newInclude := "#include <" + v + ">\n"
			if !strings.Contains(includeString, newInclude) {
				includeString += newInclude
			}
		}
	}
	return includeString + "\n" + output
}

func go2cpp(source string) string {
	lines := []string{}
	currentReturnType := ""
	currentFunctionName := ""
	inImport := false
	for _, line := range strings.Split(source, "\n") {
		newLine := line
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) == 0 {
			lines = append(lines, newLine)
			continue
		}
		if inImport && strings.Contains(trimmedLine, ")") {
			inImport = false
			continue
		} else if inImport {
			continue
		}
		if strings.HasPrefix(trimmedLine, "func") {
			newLine, currentReturnType, currentFunctionName = FunctionSignature(trimmedLine)
		} else if strings.HasPrefix(trimmedLine, "return") {
			if strings.HasPrefix(currentReturnType, tupleType) {
				elems := strings.SplitN(newLine, "return ", 2)
				newLine = "return " + currentReturnType + "{" + elems[1] + "}"
			} else {
				// Just use the standard tuple
			}
		} else if strings.HasPrefix(trimmedLine, "fmt.Print") || strings.HasPrefix(trimmedLine, "print") {
			newLine = PrintStatement(line)
		} else if strings.Contains(trimmedLine, "=") {
			elem := strings.Split(trimmedLine, "=")
			left := strings.TrimSpace(elem[0])
			declarationAssignment := false
			if strings.HasSuffix(left, ":") {
				declarationAssignment = true
				left = left[:len(left)-1]
			}
			right := strings.TrimSpace(elem[1])
			if strings.Contains(left, ",") {
				newLine = "auto [" + left + "] = " + right
			} else if declarationAssignment {
				newLine = "auto " + left + " = " + right
			} else {
				newLine = left + " = " + right
			}
		} else if strings.HasPrefix(trimmedLine, "package") {
			continue
		} else if strings.HasPrefix(trimmedLine, "import") {
			inImport = true
			if strings.Contains(trimmedLine, ")") {
				inImport = false
			}
			continue
		}
		if currentFunctionName == "main" && trimmedLine == "}" {
			newLine = strings.Replace(line, "}", "return 0;\n}", 1)
		}
		if !has(endings, lastchar(trimmedLine)) && !strings.HasPrefix(trimmedLine, "//") {
			newLine += ";"
		}
		lines = append(lines, newLine)
	}
	output := strings.Join(lines, "\n")

	// The order matters
	output = LiteralStrings(output)
	output = WholeProgramReplace(output)
	output = AddIncludes(output)

	return output
}

func main() {
	debug := false
	compile := true

	inputFilename := ""
	if len(os.Args) > 1 {
		inputFilename = os.Args[1]
	}

	var sourceData []byte
	var err error
	if inputFilename != "" {
		sourceData, err = ioutil.ReadFile(inputFilename)
	} else {
		sourceData, err = ioutil.ReadAll(os.Stdin)
	}
	if err != nil {
		log.Fatal(err)
	}
	if debug {
		fmt.Println(go2cpp(string(sourceData)))
		return
	}
	cmd := exec.Command("clang-format", "-style=Webkit")
	cmd.Stdin = strings.NewReader(go2cpp(string(sourceData)))
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Println("clang-format must be available")
		log.Fatal(err)
	}
	formattedCPP := out.String()
	if !compile {
		fmt.Println(formattedCPP)
		return
	}
	// Compile the string in formattedCPP
	cmd2 := exec.Command("g++", "-x", "c++", "-std=c++17", "-Os", "-o", "/dev/stdout", "-")
	cmd2.Stdin = strings.NewReader(formattedCPP)
	var compiled bytes.Buffer
	var errors bytes.Buffer
	cmd2.Stdout = &compiled
	cmd2.Stderr = &errors
	err = cmd2.Run()
	if err != nil {
		fmt.Println("Failed to compile this with g++:")
		fmt.Println(formattedCPP)
		fmt.Println("Errors:")
		fmt.Println(errors.String())
		log.Fatal(err)
	}
	//defaultOutputFilename := filepath.Base(os.Getenv("PWD"))
	if len(os.Args) > 2 {
		if os.Args[2] != "-o" {
			log.Fatal("The second argument must be -o")
		}
	}
	outputFilename := ""
	if len(os.Args) > 3 {
		outputFilename = os.Args[3]
	}
	if outputFilename != "" {
		err = ioutil.WriteFile(outputFilename, compiled.Bytes(), 0755)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println(formattedCPP)
	}
}
