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

const (
	hashMapSuffix = "_h__"
	keysSuffix    = "_k__"
	switchPrefix  = "_s__"
	labelPrefix   = "_l__"
)

var endings = []string{"{", ",", "}", ":"}

var (
	switchExpressionCounter = -1
	firstCase               bool
	switchLabel             string
	labelCounter            int
)

// between returns the string between two given strings, or the original string
// front can be set to false to search for b from the end of s
func between(s, a, b string, front bool) string {
	apos := strings.Index(s, a)
	if apos == -1 {
		return s
	}
	var bpos int
	if front {
		bpos = strings.Index(s, b)
	} else {
		bpos = strings.LastIndex(s, b)
	}
	if bpos == -1 {
		return s
	}
	return s[apos+len(a) : bpos]
}

// TODO: Make more robust, this easily breaks
func LiteralStrings(source string) string {

	// Temporary
	return source

	output := source
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
		"strings.Contains":  `inline auto stringsContains(std::string const& a, std::string const& b) -> bool { return a.find(b) != std::string::npos; }`,
		"strings.HasPrefix": `inline auto stringsHasPrefix(std::string const& givenString, std::string const& prefix) -> auto { return 0 == givenString.find(prefix); }`,
		"_format_output":    `template <typename T> void _format_output(std::ostream& out, T x) { if constexpr (std::is_same<T, bool>::value) { out << std::boolalpha << x << std::noboolalpha; } else if constexpr (std::is_integral<T>::value) { out << static_cast<int>(x); } else if constexpr (std::is_object<T>::value && std::is_pointer<T>::value) { out << "&" << x->_str(); } else if constexpr (std::is_object<T>::value && !std::is_pointer<T>::value) { out << x._str(); } else { out << x; } }`,
		"strings.TrimSpace": `inline auto stringsTrimSpace(std::string const& s) -> std::string { std::string news {}; for (auto l : s) { if (l != ' ' && l != '\n' && l != '\t' && l != '\v' && l != '\f' && l != '\r') { news += l; } } return news; }`,
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
func FunctionArguments(source string) string {
	output := source
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
		s := between(output, "(", ")", false)
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
	words := strings.Split(between(args, "(", ")", true), ",")
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
	args := FunctionArguments(between(output, "(", ")", true))
	rets := FunctionRetvals(between(output, ")", "{", true))
	if strings.Contains(rets, ",") {
		// Multiple return
		rets = tupleType + "<" + CPPTypes(rets) + ">"
	}
	name = between(output, "func ", "(", true)
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

// Split arguments. Handles quoting 1 level deep.
func SplitArgs(s string) []string {
	inQuote := false
	inSingleQuote := false
	inPar := false
	inCurly := false
	var args []string
	word := ""
	for _, letter := range s {
		switch letter {
		case '"':
			inQuote = !inQuote
		case '\'':
			inSingleQuote = !inSingleQuote
		}
		if letter == '(' && !inQuote && !inSingleQuote && !inPar && !inCurly {
			inPar = true
		}
		if letter == ')' && !inQuote && !inSingleQuote {
			inPar = false
		}
		if letter == '{' && !inQuote && !inSingleQuote && !inPar && !inCurly {
			inCurly = true
		}
		if letter == '}' && !inQuote && !inSingleQuote {
			inCurly = false
		}
		if letter == ',' && !inQuote && !inSingleQuote && !inPar && !inCurly {
			args = append(args, strings.TrimSpace(word))
			word = ""
		} else {
			word += string(letter)
		}
	}
	args = append(args, strings.TrimSpace(word))
	return args
}

func PrintStatement(source string) (string, bool) {

	// Pick out and trim all arguments given to the print functon
	args := SplitArgs(between(strings.TrimSpace(source), "(", ")", false))

	// Identify the print function
	if !strings.Contains(source, "(") {
		// Not a function call
		return source, false
	}

	fname := strings.TrimSpace(source[:strings.Index(source, "(")])
	//fmt.Println("FNAME", fname)

	// Check if the function call ends with "ln" (println, fmt.Println)
	addNewline := strings.HasSuffix(fname, "ln")
	//fmt.Println("NEWLINE", addNewline)

	// Check if the function call starts with "print" (as opposed to "Print")
	lowercasePrint := strings.HasPrefix(fname, "print")
	//fmt.Println("LOWERCASE PRINT", lowercasePrint)

	// Check if all the arguments are literal strings
	allLiteralStrings := true
	for _, arg := range args {
		if !strings.HasPrefix(arg, "\"") {
			allLiteralStrings = false
		}
	}
	mayNeedPrettyPrint := !allLiteralStrings

	//fmt.Println("ALL LITERAL STRINGS", allLiteralStrings)

	// --- enough information gathered, it's time to build the output code ---

	if strings.HasSuffix(fname, "rintf") {
		output := source
		// TODO: Also support fmt.Fprintf, and format %v values differently.
		//       Converting to an iostream expression is one possibility.
		output = strings.Replace(output, "fmt.Printf", "printf", 1)
		output = strings.Replace(output, "fmt.Fprintf", "fprintf", 1)
		output = strings.Replace(output, "fmt.Sprintf", "sprintf", 1)
		if strings.Contains(output, "%v") {
			panic("support for %v is not implemented yet")
		}
		return output, mayNeedPrettyPrint
	}

	outputName := "std::cout"
	if lowercasePrint {
		// print and println outputs to stderr
		outputName = "std::cerr"
	}
	//fmt.Println("OUTPUT NAME", outputName)

	// Useful values
	pipe := " << "
	blank := "\" \""
	nl := "std::endl"

	// Silence pipeNewline if the print function does not end with "ln"
	pipeNewline := pipe + nl
	if !addNewline {
		pipeNewline = ""
	}

	// No arguments given?
	if len(args) == 0 {
		// Just output a newline
		if addNewline {
			return outputName + pipeNewline, false
		}
	}

	// Only one argument given?
	if len(args) == 1 {
		if strings.TrimSpace(args[0]) == "" {
			// Just output a newline
			if addNewline {
				return outputName + pipeNewline, false
			}
		}
		if allLiteralStrings {
			return outputName + pipe + args[0] + pipeNewline, false
		}
		output := "_format_output(" + outputName + ", " + args[0] + ")"
		if addNewline {
			output += ";\n" + outputName + pipeNewline
		}
		return output, mayNeedPrettyPrint
	}

	// Several arguments given
	//fmt.Println("SEVERAL ARGUMENTS", args)

	// HINT: Almost everything should start with "pipe" and almost nothing should end with "pipe"
	output := outputName
	lastIndex := len(args) - 1
	for i, arg := range args {
		//fmt.Println("ARGUMENT", i, arg)
		// Is it a literal string?
		if strings.HasPrefix(arg, "\"") {
			output += pipe + arg
		} else {
			if i == 0 {
				output = ""
			} else {
				output += ";\n"
			}
			output += "_format_output(" + outputName + ", " + arg + ");\n" + outputName
		}
		if i < lastIndex {
			output += pipe + blank
		} else {
			output += pipeNewline
		}
	}

	//fmt.Println("GENERATED OUTPUT", output)

	return output, mayNeedPrettyPrint
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
		"std::int8_t":        "cinttypes",
		"std::int16_t":       "cinttypes",
		"std::int32_t":       "cinttypes",
		"std::int64_t":       "cinttypes",
		"std::uint8_t":       "cinttypes",
		"std::uint16_t":      "cinttypes",
		"std::uint32_t":      "cinttypes",
		"std::uint64_t":      "cinttypes",
		"printf":             "cstdio",
		"fprintf":            "cstdio",
		"sprintf":            "cstdio",
		"snprintf":           "cstdio",
		"std::stringstream":  "sstream",
		// TODO: complex64, complex128
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
	expression := strings.TrimSpace(between(source, "if", "{", true))
	return "if (" + expression + ") {"
}

func ElseIfSentence(source string) (output string) {
	output = source
	expression := strings.TrimSpace(between(source, "} else if", "{", true))
	return "} else if (" + expression + ") {"
}

func TypeReplace(source string) string {
	// TODO: uintptr, complex64 and complex128
	switch strings.TrimSpace(source) {
	case "string":
		return "std::string"
	case "float64":
		return "double"
	case "float32":
		return "float"
	case "uint64":
		return "std::uint64_t"
	case "uint32":
		return "std::uint32_t"
	case "uint16":
		return "std::uint16_t"
	case "uint8":
		return "std::uint8_t"
	case "int64":
		return "std::int64_t"
	case "int32":
		return "std::int32_t"
	case "int16":
		return "std::int16_t"
	case "int8":
		return "std::int8_t"
	case "byte":
		return "std::uint8_t"
	case "rune":
		return "std::int32_t"
	case "uint":
		return "unsigned int"
	default:
		return source
	}
}

// TODO: Make sure all variations are covered:
// * [_] for _ = range list {
// * [_] for i := range list {
// * [_] for i, v := range list {
// * [_] for _, v := range list {
// * [_] for i, _ := range list {
// * [_] for _, _ = range list {
// * [_] for _ = range map {
// * [_] for k := range map {
// * [_] for k, v := range map {
// * [_] for _, v := range map {
// * [_] for k, _ := range map {
// * [_] for _, _ = range map {
// * [_] for i := 0; i < 10; i++ {
// * [_] for {
// * [_] for ;; {
// * not possible: for x := range 10 {
func ForLoop(source string, encounteredHashMaps []string) string {
	expression := strings.TrimSpace(between(source, "for", "{", true))
	if expression == "" {
		// endless loop
		return "for (;;) {"
	}
	// for range, with no comma
	if strings.Count(expression, ",") == 0 && strings.Contains(expression, "range") {
		fields := strings.Split(expression, " ")
		varName := fields[0]
		listName := fields[len(fields)-1]

		// for i := range l {
		// -->
		// for (auto i = 0; i < std::size(l); i++) {

		hashMapName := listName
		if has(encounteredHashMaps, hashMapName) {
			// looping over the key of a hash map, not over the index of a list
			return "for (const auto & [" + varName + ", " + varName + "__" + "] : " + hashMapName + ") {"
		} else if varName == "_" {
			return "for (const auto & [" + varName + "__" + ", " + varName + "___" + "] : " + hashMapName + ") {"
		} else {
			// looping over the index of a list
			return "for (std::size_t " + varName + " = 0; " + varName + " < std::size(" + listName + "); " + varName + "++) {"
		}
	}
	// for range, over index and element, or key and value
	if strings.Count(expression, ",") == 1 && strings.Contains(expression, "range") && strings.Contains(expression, ":=") {
		fields := strings.Split(expression, ":=")
		varnames := strings.Split(fields[0], ",")

		indexvar := varnames[0]
		elemvar := varnames[1]

		fields = strings.Split(expression, " ")
		listName := fields[len(fields)-1]
		hashMapName := listName

		if has(encounteredHashMaps, hashMapName) {
			if indexvar == "_" {
				// looping over the values of a hash map
				hashMapHashKey := hashMapName + hashMapSuffix
				return "for (const auto & " + hashMapHashKey + " : " + hashMapName + ") {" + "\n" + "auto " + elemvar + " = " + hashMapHashKey + ".second"
			}
			// for k, v := range m
			keyvar := indexvar
			//hashMapHashKey := keyvar + hashMapSuffix + keysSuffix
			return "for (const auto & [" + keyvar + ", " + elemvar + "] : " + hashMapName + ") {"
			//return "for (auto " + hashMapHashKey + " : " + hashMapName + keysSuffix + ") {" + "\n" + "auto " + keyvar + " = " + hashMapHashKey + ".second;\nauto " + elemvar + " = " + hashMapName + ".at(" + hashMapHashKey + ".first)"
		}

		if indexvar == "_" {
			return "for (auto " + elemvar + " : " + listName + ") {"
		}
		return "for (std::size_t " + indexvar + " = 0; " + indexvar + " < std::size(" + listName + "); " + indexvar + "++) {" + "\n" + "auto " + elemvar + " = " + listName + "[" + indexvar + "]"
	}
	// not "for" + "range"
	if strings.Contains(expression, ":=") {
		if strings.HasPrefix(expression, "_,") && strings.Contains(expression, "range") {
			// For each, no index
			varname := between(expression, ",", ":", true)
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
	return switchPrefix + strconv.Itoa(switchExpressionCounter)
}

func LabelName() string {
	return labelPrefix + strconv.Itoa(labelCounter)
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
	s := between(output, " ", ":", true)
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

// Return transformed line and the variable name
func VarDeclaration(source string) (string, string) {
	if strings.Contains(source, "=") {
		parts := strings.SplitN(strings.TrimSpace(source), "=", 2)
		left := parts[0]
		right := strings.TrimSpace(parts[1])
		fields := strings.Split(strings.TrimSpace(left), " ")
		if fields[0] == "var" {
			fields = fields[1:]
		}
		if len(fields) > 2 {
			return TypeReplace(fields[1]) + " " + fields[0] + " " + strings.Join(fields[2:], " ") + " = " + right, fields[0]
		} else if len(fields) == 2 {
			return TypeReplace(fields[1]) + " " + fields[0] + " = " + right, fields[0]
		} else {
			return "auto" + " " + fields[0] + " = " + right, fields[0]
		}
	}
	fields := strings.Fields(source)
	if fields[0] == "var" {
		fields = fields[1:]
	}
	if len(fields) == 2 {
		return TypeReplace(fields[1]) + " " + fields[0], fields[0]
	}
	// Unrecognized
	panic("Unrecognized var declaration: " + source)
}

// TypeDeclaration returns a transformed string (from Go to C++),
// and a bool if a struct is opened (with {).
func TypeDeclaration(source string) (string, bool) {
	fields := strings.Split(strings.TrimSpace(source), " ")
	if fields[0] == "type" {
		fields = fields[1:]
	}
	left := strings.TrimSpace(fields[0])
	right := strings.TrimSpace(fields[1])
	words := strings.Split(left, " ")
	if len(fields) == 2 {
		// Type alias
		return "using " + left + " = " + TypeReplace(right), false
	} else if len(words) == 2 {
		// Type alias
		return "using " + words[1] + " " + words[0] + " = " + TypeReplace(right), false
	} else if strings.Contains(right, "struct") {
		// type Vec3 struct {
		// to
		// class Vec3 { public:
		// also the closing bracket must end with a semicolon
		return "class " + left + " { public:", true
	} else if len(words) == 1 {
		// Type alias
		return "using " + left + " = " + TypeReplace(right), false
	}
	// Unrecognized
	panic("Unrecognized type declaration: " + source)
}

func ConstDeclaration(source string) (output string) {
	output = source
	fields := strings.SplitN(source, "=", 2)
	left := strings.TrimSpace(fields[0])
	right := strings.TrimSpace(fields[1])
	words := strings.Split(left, " ")
	if len(words) == 1 {
		// No type
		return "const auto " + left + " = " + right
	} else if len(words) == 2 {
		if words[0] == "const" {
			return "const auto " + " " + words[1] + " = " + right
		}
		return "const " + TypeReplace(words[1]) + " " + words[0] + " = " + right
	}
	// Unrecognized
	panic("Unrecognized const expression: " + source)
}

// HashElements transforms the contents of a map in Go to the contents of an unordered_map in C++
// keyType is the type of the key, in C++, for instance "std::string"
// if keyForBoth is true, a hash(key)->key map is created,
// if not, a hash(key)->value map is created.
func HashElements(source, keyType string, keyForBoth bool) string {
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
		pairElements := strings.SplitN(pair, ":", 2)
		if len(pairElements) != 2 {
			panic("This should be two elements, separated by a colon: " + pair)
		}
		output += "{ " + pairElements[0] + ", " + pairElements[1] + " }"
	}
	return output + "}"
}

func CreateStrMethod(varNames []string) string {
	s := "std::string _str() {\n"
	s += `  std::stringstream ss; ss << "{" `;
	for i, varName := range varNames {
		if i > 0 {
			s += ` << " " `
		}
		s += " << " + varName
	}
	s += ` << "}";`
	s += "\n  return ss.str(); }"
	return s
}

func go2cpp(source string) string {
	lines := []string{}
	currentReturnType := ""
	currentFunctionName := ""
	structVarNames := []string{}
	inImport := false
	inVar := false
	inType := false
	inConst := false
	curlyCount := 0
	// Keep track of encountered hash maps
	// TODO: Use reflection instead to loop either one way or the other. The hash map may be defined in another package.
	encounteredHashMaps := []string{}
	inStruct := false
	for _, line := range strings.Split(source, "\n") {
		newLine := line
		trimmedLine := strings.TrimSpace(line)
		// TODO: A multiline string could have lines starting with //, make sure to support this
		if strings.HasPrefix(trimmedLine, "//") {
			lines = append(lines, trimmedLine)
			continue
		}
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
		} else if inType && strings.Contains(trimmedLine, ")") {
			inType = false
			continue
		} else if inConst && strings.Contains(trimmedLine, ")") {
			inConst = false
			continue
		} else if inVar || (inStruct && trimmedLine != "}"){
			name := ""
			newLine, name = VarDeclaration(line)
			if inStruct {
				// Gathering variable names from this struct
				structVarNames = append(structVarNames, name)
			}
		} else if inType {
			prevInStruct := inStruct
			newLine, inStruct = TypeDeclaration(line)
			if !prevInStruct && inStruct {
				// Entering struct, reset the slice that is used to gather variable names
				structVarNames = []string{}
			}
		} else if inConst {
			newLine = ConstDeclaration(line)
		} else if strings.HasPrefix(trimmedLine, "func") {
			newLine, currentReturnType, currentFunctionName = FunctionSignature(trimmedLine)
		} else if strings.HasPrefix(trimmedLine, "for") {
			newLine = ForLoop(line, encounteredHashMaps)
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
			// The ignored return value is if "pretty print" functionality may be needed, for structs and classes
			newLine, _ = PrintStatement(line)
		} else if strings.Contains(trimmedLine, "=") && !strings.HasPrefix(trimmedLine, "var ") && !strings.HasPrefix(trimmedLine, "if ") && !strings.HasPrefix(trimmedLine, "const ") && !strings.HasPrefix(trimmedLine, "type ") {
			elem := strings.Split(trimmedLine, "=")
			left := strings.TrimSpace(elem[0])
			declarationAssignment := false
			if strings.HasSuffix(left, ":") {
				declarationAssignment = true
				left = left[:len(left)-1]
			}
			right := strings.TrimSpace(elem[1])
			if strings.HasPrefix(right, "&") && strings.Contains(right, "{") && strings.Contains(right, "}") {
				right = "new " + right[1:]
			}
			if strings.Contains(left, ",") {
				newLine = "auto [" + left + "] = " + right
			} else if declarationAssignment {
				if strings.HasPrefix(right, "[]") {
					if !strings.Contains(right, "{") {
						fmt.Fprintln(os.Stderr, "UNRECOGNIZED LINE: "+trimmedLine)
						newLine = line

					}
					theType := TypeReplace(between(right, "]", "{", true))
					fields := strings.SplitN(right, "{", 2)
					newLine = theType + " " + strings.TrimSpace(left) + "[] {" + fields[1]
				} else if strings.HasPrefix(right, "map[") {
					keyType := TypeReplace(between(right, "map[", "]", true))
					valueType := TypeReplace(between(right, "]", "{", true))
					elements := between(right, "{", "}", true)
					hashName := strings.TrimSpace(left)
					newLine = "std::unordered_map<" + keyType + ", " + valueType + "> " + hashName + " " + HashElements(elements, keyType, false)
					encounteredHashMaps = append(encounteredHashMaps, hashName)
				} else {
					newLine = "auto " + strings.TrimSpace(left) + " = " + strings.TrimSpace(right)
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
		} else if trimmedLine == "type (" {
			inType = true
			continue
		} else if trimmedLine == "const (" {
			inConst = true
			continue
		} else if strings.HasPrefix(trimmedLine, "var ") {
			// Ignore variable name since it's not in a struct
			newLine, _ = VarDeclaration(line)
		} else if strings.HasPrefix(trimmedLine, "type ") {
			newLine, inStruct = TypeDeclaration(line)
		} else if strings.HasPrefix(trimmedLine, "const ") {
			newLine = ConstDeclaration(line)
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
			// If the struct is being closed, add a semicolon
			if inStruct {
				// Create a _str() method for this struct
				newLine = CreateStrMethod(structVarNames) + newLine + ";"

				inStruct = false;
			}
			newLine += "\n"
		}
		if (!strings.HasSuffix(newLine, ";") && !has(endings, lastchar(trimmedLine)) || strings.Contains(trimmedLine, "=")) && !strings.HasPrefix(trimmedLine, "//") && (!has(endings, lastchar(newLine)) && !strings.Contains(newLine, "//")) {
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
			log.Fatal("The second argument must be -o (format sources with clang-format) or -O (don't format sources with clang-format)")
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
		cmd := exec.Command("clang-format", "-style={BasedOnStyle: Webkit, ColumnLimit: 99}")
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
		//fmt.Println("Failed to compile this with g++:")
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
