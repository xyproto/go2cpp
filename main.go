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
	"strconv"
	"strings"
)

const tupleType = "std::tuple"

var endings = []string{"{", ",", "}", ":"}

var (
	switchExpressionCounter int = -1
	firstCase               bool
	switchLabel             string
	labelCounter            int
)

// between returns the string between two given strings, or the original string
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

// TODO: Make more robust, this easily breaks
func LiteralStrings(source string) (output string) {
	output = source
	replacements := map[string]string{
		"\")":  "\"s)",
		"\";":  "\"s;",
		"\",":  "\"s,",
		"\"}":  "\"s}",
		"\" }": "\"s }",
		"\" )": "\"s )",
		"\":":  "\"s:",
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

// TODO: Avoid whole-program replacements, if possible
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

func AddFunctions(source string) (output string) {
	output = source
	replacements := map[string]string{
		"strings.Contains":  "inline auto stringsContains(std::string const& a, std::string const& b) -> bool { return a.find(b) != std::string::npos; }",
		"strings.HasPrefix": "inline auto stringsHasPrefix(std::string const& givenString, std::string const& prefix) -> auto { return 0 == givenString.find(prefix); }",
	}
	for k, v := range replacements {
		if strings.Contains(output, k) {
			output = strings.Replace(output, k, strings.Replace(k, ".", "", -1), -1)
			output = v + "\n" + output
		}
	}
	return output
}

// FunctionArguments transforms the arguments given to a function
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

// FunctionRetvals transforms the return values from a function
func FunctionRetvals(source string) (output string) {
	if len(strings.TrimSpace(source)) == 0 {
		return source
	}
	output = source
	if strings.Contains(output, "(") {
		s := between(output, "(", ")")
		retvals := FunctionArguments(s)
		if strings.Contains(retvals, ",") {
			output = "(" + retvals + ")"
		} else {
			output = retvals
		}
	}
	return strings.TrimSpace(output)
}

// CPPTypes picks out the types given a list of C++ arguments with name and type
func CPPTypes(args string) string {
	words := strings.Split(between(args, "(", ")"), ",")
	var atypes []string
	for _, word := range words {
		elems := strings.Split(strings.TrimSpace(word), " ")
		atypes = append(atypes, elems[0])
	}
	return strings.Join(atypes, ", ")
}

// FunctionSignature transforms a function signature that spans one line
// Will change the "func main" signature to a main function that returns an int.
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
	// fmt.Println, fmt.Print
	if strings.HasPrefix(name, "print") {
		output = "std::cerr"
	} else {
		output = "std::cout"
	}
	if len(args) == 0 {
		// fmt.Println() or fmt.Print()
		return output + " << std::endl"
	}
	// Output booleans as "true" and "false" instead of as numbers
	output += " << std::boolalpha << "
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
		"std::tuple":         "tuple",
		"std::endl":          "iostream",
		"std::cout":          "iostream",
		"std::string":        "string",
		"std::size":          "iterator",
		"std::unordered_map": "unordered_map",
		"std::hash":          "functional",
		"std::size_t":        "cstddef",
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

func IfSentence(source string) (output string) {
	output = source
	expression := strings.TrimSpace(between(source, "if", "{"))
	return "if (" + expression + ") {"
}

func ElseIfSentence(source string) (output string) {
	output = source
	expression := strings.TrimSpace(between(source, "} else if", "{"))
	return "} else if (" + expression + ") {"
}

func TypeReplace(source string) string {
	output := source
	output = strings.Replace(output, "string", "std::string", -1)
	return output
}

func ForLoop(source string) string {
	expression := strings.TrimSpace(between(source, "for", "{"))
	// for range, with no comma
	if strings.Count(expression, ",") == 0 && strings.Contains(expression, "range") {
		fields := strings.Split(expression, " ")
		varname := fields[0]
		listname := fields[len(fields)-1]

		// for i := range l {
		// -->
		// for (auto i = 0; i < std::size(l); i++) {

		return "for (std::size_t " + varname + " = 0; " + varname + " < std::size(" + listname + "); " + varname + "++) {"
	}
	// for range, over index and element, or key and value
	if strings.Count(expression, ",") == 1 && strings.Contains(expression, "range") && strings.Contains(expression, ":=") {
		fields := strings.Split(expression, ":=")
		varnames := strings.Split(fields[0], ",")

		indexvar := varnames[0]
		elemvar := varnames[1]

		fields = strings.Split(expression, " ")
		listname := fields[len(fields)-1]

		return "for (std::size_t " + indexvar + " = 0; " + indexvar + " < std::size(" + listname + "); " + indexvar + "++) {" + "\n" + "auto " + elemvar + " = " + listname + "[" + indexvar + "]"
	}
	// not "for" + "range"
	if strings.Contains(expression, ":=") {
		if strings.HasPrefix(expression, "_,") && strings.Contains(expression, "range") {
			// For each, no index
			varname := between(expression, ",", ":")
			fields := strings.SplitN(expression, "range ", 2)
			listname := fields[1]
			// C++11 and later for each loop
			expression = "auto &" + varname + " : " + listname
		} else {
			expression = "auto " + strings.Replace(expression, ":=", "=", 1)
		}
	}
	return "for (" + expression + ") {"
}

func SwitchExpressionVariable() string {
	return "_s__" + strconv.Itoa(switchExpressionCounter)
}

func LabelName() string {
	return "_l__" + strconv.Itoa(labelCounter)
}

func Switch(source string) (output string) {
	output = strings.TrimSpace(source)[len("switch "):]
	if strings.HasSuffix(output, "{") {
		output = strings.TrimSpace(output[:len(output)-1])
	}
	switchExpressionCounter++
	firstCase = true
	return "auto " + SwitchExpressionVariable() + " = " + output + "; // switch on " + output
}

func Case(source string) (output string) {
	output = source
	s := between(output, " ", ":")
	if firstCase {
		firstCase = false
		output = "if ("
	} else {
		output = "} else if ("
	}
	output += SwitchExpressionVariable() + " == " + s + ") { // case " + s
	if switchLabel != "" {
		output += "\n" + switchLabel + ":"
		switchLabel = ""
	}
	return output
}

func VarDeclaration(source string) (output string) {
	output = source
	if strings.HasPrefix(output, "var ") {
		output = output[4:]
	}
	if strings.Contains(output, "=") {
		parts := strings.Split(output, " ")
		if len(parts) == 4 {
			output = parts[0] + " " + parts[2] + " " + parts[3]
		}
		output = "auto " + output
	} else {
		parts := strings.Split(output, " ")
		if len(parts) == 3 {
			output = TypeReplace(parts[2]) + " " + parts[1]
		}
	}
	return output
}

// shouldHash decides if the given type, as a key in an unordered_map, should be hashed
func shouldHash(keyType string) bool {
	// TODO: Check if always using std::hash makes sense, or only for some types (then which ones?)
	return strings.Contains(keyType, "std::")
}

// HashElements transforms the contents of a map in Go to the contents of an unordered_map in C++
// keyType is the type of the key, for instance "std::string"
func HashElements(source string, keyType string) string {
	if !strings.Contains(source, ",") {
		return source
	}
	pairs := strings.Split(source, ",")
	output := "{"
	first := true
	for _, pair := range pairs {
		if !first {
			output += ","
		} else {
			first = false
		}
		pair_elements := strings.SplitN(pair, ":", 2)
		if len(pair_elements) != 2 {
			panic("This should be two elements, separated by a colon: " + pair)
		}
		if shouldHash(keyType) {
			output += "{std::hash<" + keyType + ">{}(" + pair_elements[0] + "), " + pair_elements[1] + "}"
		} else {
			output += "{" + pair_elements[0] + ", " + pair_elements[1] + "}"
		}
	}
	output += "}"
	return output
}

func go2cpp(source string) string {
	lines := []string{}
	currentReturnType := ""
	currentFunctionName := ""
	inImport := false
	curlyCount := 0
	inVar := false
	for _, line := range strings.Split(source, "\n") {
		newLine := line
		trimmedLine := strings.TrimSpace(line)
		if strings.HasSuffix(trimmedLine, ";") {
			trimmedLine = trimmedLine[:len(trimmedLine)-1]
		}
		if len(trimmedLine) == 0 {
			lines = append(lines, newLine)
			continue
		}
		// Keep track of how deep we are into curly brackets
		curlyCount += (strings.Count(trimmedLine, "{") - strings.Count(trimmedLine, "}"))
		if inImport && strings.Contains(trimmedLine, ")") {
			inImport = false
			continue
		} else if inImport {
			continue
		} else if inVar && strings.Contains(trimmedLine, ")") {
			inVar = false
			continue
		} else if inVar {
			newLine = VarDeclaration(line)
		} else if strings.HasPrefix(trimmedLine, "func") {
			newLine, currentReturnType, currentFunctionName = FunctionSignature(trimmedLine)
		} else if strings.HasPrefix(trimmedLine, "for") {
			newLine = ForLoop(line)
		} else if strings.HasPrefix(trimmedLine, "switch") {
			newLine = Switch(line)
		} else if strings.HasPrefix(trimmedLine, "case") {
			newLine = Case(line)
		} else if strings.HasPrefix(trimmedLine, "return") {
			if strings.HasPrefix(currentReturnType, tupleType) {
				elems := strings.SplitN(newLine, "return ", 2)
				newLine = "return " + currentReturnType + "{" + elems[1] + "};"
			} else {
				// Just use the standard tuple
			}
		} else if strings.HasPrefix(trimmedLine, "fmt.Print") || strings.HasPrefix(trimmedLine, "print") {
			newLine = PrintStatement(line)
		} else if strings.Contains(trimmedLine, "=") && !strings.HasPrefix(trimmedLine, "var ") {
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
				if strings.HasPrefix(right, "[]") {
					if !strings.Contains(right, "{") {
						panic("UNRECOGNIZED LINE: " + trimmedLine)
					}
					theType := TypeReplace(between(right, "]", "{"))
					fields := strings.SplitN(right, "{", 2)
					newLine = theType + " " + left + "[] {" + fields[1]
				} else if strings.HasPrefix(right, "map[") {
					keyType := TypeReplace(between(right, "map[", "]"))
					valueType := TypeReplace(between(right, "]", "{"))
					elements := between(right, "{", "}")
					if shouldHash(keyType) {
						newLine = "std::unordered_map<std::size_t, " + valueType + "> " + left + HashElements(elements, keyType)
					} else {
						newLine = "std::unordered_map<" + keyType + ", " + valueType + "> " + left + HashElements(elements, keyType)
					}
				} else {
					newLine = "auto " + left + " = " + right
				}
			} else {
				newLine = left + " = " + right
			}
		} else if strings.HasPrefix(trimmedLine, "package ") {
			continue
		} else if strings.HasPrefix(trimmedLine, "import") {
			if strings.Contains(trimmedLine, "(") {
				inImport = true
			}
			if strings.Contains(trimmedLine, ")") {
				inImport = false
			}
			continue
		} else if strings.HasPrefix(trimmedLine, "if ") {
			newLine = IfSentence(line)
		} else if strings.HasPrefix(trimmedLine, "} else if ") {
			newLine = ElseIfSentence(line)
		} else if trimmedLine == "var (" {
			inVar = true
			continue
		} else if strings.HasPrefix(trimmedLine, "var ") {
			newLine = VarDeclaration(line)
		} else if trimmedLine == "fallthrough" {
			newLine = "goto " + LabelName() + "; // fallthrough"
			switchLabel = LabelName()
			labelCounter++
		} else if trimmedLine == "default:" {
			newLine = "} else { // default case"
			if switchLabel != "" {
				newLine += "\n" + switchLabel + ":"
				switchLabel = ""
			}
		}
		if currentFunctionName == "main" && trimmedLine == "}" && curlyCount == 0 { // curlyCount has already been decreased for this line
			newLine = strings.Replace(line, "}", "return 0;\n}", 1)
		}
		if strings.HasSuffix(trimmedLine, "}") {
			newLine += "\n"
		}
		if (!has(endings, lastchar(trimmedLine)) || strings.Contains(trimmedLine, "=")) && !strings.HasPrefix(trimmedLine, "//") && (!has(endings, lastchar(newLine)) && !strings.Contains(newLine, "//")) {
			newLine += ";"
		}
		lines = append(lines, newLine)
	}
	output := strings.Join(lines, "\n")

	// The order matters
	output = LiteralStrings(output)
	output = WholeProgramReplace(output)
	output = AddFunctions(output)
	output = AddIncludes(output)

	return output
}

func main() {

	// TODO: Use https://github.com/docopt/docopt.go for parsing arguments

	debug := false
	compile := true
	clangFormat := true

	inputFilename := ""
	if len(os.Args) > 1 {
		if os.Args[1] == "--help" {
			fmt.Println("supported arguments:")
			fmt.Println(" a .go file as the first argument")
			fmt.Println("supported options:")
			fmt.Println(" -o : Format with clang format")
			fmt.Println(" -O : Don't format with clang format")
			return
		}
		inputFilename = os.Args[1]
	}
	if len(os.Args) > 2 {
		if os.Args[2] == "-o" {
			clangFormat = true
		} else if os.Args[2] == "-O" {
			clangFormat = false
		} else if os.Args[2] != "-o" {
			log.Fatal("The second argument must be -o (don't prepare sources with clang-format) or -O (prepare sources with clang-format)")
		}
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

	cppSource := ""
	if clangFormat {
		cmd := exec.Command("clang-format", "-style={BasedOnStyle: Webkit, ColumnLimit: 90}")
		cmd.Stdin = strings.NewReader(go2cpp(string(sourceData)))
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			log.Println("clang-format is not available, the output will look ugly!")
			cppSource = go2cpp(string(sourceData))
		} else {
			cppSource = out.String()
		}
	} else {
		cppSource = go2cpp(string(sourceData))
	}

	if !compile {
		fmt.Println(cppSource)
		return
	}

	// Compile the string in cppSource
	cmd2 := exec.Command("g++", "-x", "c++", "-std=c++17", "-O2", "-pipe", "-fPIC", "-Wfatal-errors", "-s", "-o", "/dev/stdout", "-")
	cmd2.Stdin = strings.NewReader(cppSource)
	var compiled bytes.Buffer
	var errors bytes.Buffer
	cmd2.Stdout = &compiled
	cmd2.Stderr = &errors
	err = cmd2.Run()
	if err != nil {
		fmt.Println("Failed to compile this with g++:")
		fmt.Println(cppSource)
		fmt.Println("Errors:")
		fmt.Println(errors.String())
		log.Fatal(err)
	}
	//defaultOutputFilename := filepath.Base(os.Getenv("PWD"))
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
		fmt.Println(cppSource)
	}
}
