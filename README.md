# tinygocompiler

Compiles Go to native executables via C++17.

Aims to produce small executables, by default.

## Known issues

* Only works with extremely simple code samples, for now.
* Does not use an AST, deals mainly with strings, for now.

Pull requests are warmly welcome!

## Features

* Pretty fast.
* Simple to use.
* Only uses to Go standard library, no external packages (but depends on `g++` and `clang-format`).
* Low complexity.
* Short source code.

## Usage

Compile to executable:

    tinygocompiler main.go -o main

Output what the intermediate C++17 code looks like:

    tinygocompiler main.go

## Requirements

* `g++` with support for C++17
* `clang-format`

## Example transformation

**Go input:**

```go
// Multiple return
package main

import (
	"fmt"
)

func addsub(x int) (a, b int) {
	return x + 2, x - 2
}

func main() {
	y, z := addsub(4)
	fmt.Println("y =", y)
	fmt.Println("z =", z)
}
```

**C++ output:**

```c++
#include <iostream>
#include <tuple>

// Multiple return

auto addsub(int x) -> std::tuple<int, int>
{
    return std::tuple<int, int>{ x + 2, x - 2 };
}

auto main() -> int
{
    auto [y, z] = addsub(4);
    std::cout << "y ="
              << " " << y << std::endl;
    std::cout << "z ="
              << " " << z << std::endl;
    return 0;
}
```

# General info

* Version: 0.0.0
* License: MIT

# TODO

## Keywords

- [ ] `break`
- [ ] `case`
- [ ] `chan`
- [ ] `const`
- [ ] `continue`
- [ ] `default`
- [ ] `defer`
- [x] `else`
- [ ] `fallthrough`
- [ ] `for`
- [x] `func`
- [ ] `go`
- [ ] `goto`
- [x] `if`
- [x] `import` (partially)
- [ ] `interface`
- [ ] `map`
- [x] `package` (partially)
- [ ] `range`
- [x] `return`
- [ ] `select`
- [ ] `struct`
- [ ] `switch`
- [ ] `type`
- [x] `var`

## Standard library

- [x] `fmt.Println`
- [x] `fmt.Print`
- [ ] `fmt.Printf`
- [ ] `fmt.Sprintf`
- [x] `strings.Contains`
- [x] `strings.HasPrefix`
- [ ] `strings.HasSuffix`
- [ ] `strings.Index`
- [ ] `strings.Join`
- [ ] `strings.NewReader`
- [ ] `strings.Replace`
- [ ] `strings.Split`
- [ ] `strings.SplitN`
- [ ] `strings.TrimSpace`
- [ ] All the rest

## Robustness

- [ ] Rewrite the broken semicolon-inserter.
- [ ] The addition of `s` to string literals needs to be more robust, and check if within quotes or not.
- [ ] Add a `main_test.go` that builds and calls the executables, then compare the output with the output from the executables built with the go compiler.
