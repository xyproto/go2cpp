# go2cpp

Compiles Go to native executables via C++20.

One of the goals is for the compiler to be able to compile itself.

The intended use is not to convert entire existing Go programs to C++, but to help port parts of it to C++, or perhaps write programs from scratch and continually check that the program can be converted and compiled as C++.

## Known issues

* Only works with simple code samples, for now.
* Very few functions from the Go standard library are implemented. The ideal would be to be able to compile the official Go standard library.
* A good plan for how to implement `import` is needed.

## Features and limitations

* Pretty fast.
* Simple to use.
* Few dependencies (for compiling `go2cpp`, only the go compiler is needed).
* Low complexity.
* Short source code.

## Required dependencies

* `g++` with support for C++20 is used for compiling the generated C++ code.
* `clang-format` is used for formatting the generated C++ code.

## Installation

    go install github.com/xyproto/go2cpp@latest

Then `~/go/bin/go2cpp` should be available (unless GOPATH points somewhere else).

## Usage

Compile to executable:

    go2cpp main.go -o main

Output what the intermediate C++20 code looks like:

    go2cpp main.go

## Example transformations

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

**Go input:**

```go
package main

import (
    "fmt"
)

func main() {
    m := map[string]string{"first": "hi", "second": "you", "third": "there"}
    first := true
    for k, v := range m {
        if first {
            first = false
        } else {
            fmt.Print(" ")
        }
        fmt.Print(k + v)
    }
    fmt.Println()
}
```

**C++ output:**

```c++
#include <iostream>
#include <string>
#include <unordered_map>

template <typename T> void _format_output(std::ostream& out, T x)
{
    if constexpr (std::is_same<T, bool>::value) {
        out << std::boolalpha << x << std::noboolalpha;
    } else if constexpr (std::is_integral<T>::value) {
        out << static_cast<int>(x);
    } else {
        out << x;
    }
}

auto main() -> int
{
    std::unordered_map<std::string, std::string> m{ { "first", "hi" }, { "second", "you" },
        { "third", "there" } };
    auto first = true;
    for (const auto& [k, v] : m) {
        if (first) {
            first = false;
        } else {
            std::cout << " ";
        }

        _format_output(std::cout, k + v);
    }

    std::cout << std::endl;
    return 0;
}
```

# General info

* Version: 0.4.0
* License: MIT

# TODO

## Syntactic elements

- [x] backtick quoted strings: <code>`</code> (one level deep only)
- [ ] `iota`

## Keywords

- [x] `break`
- [x] `case`
- [ ] `chan`
- [x] `const`
- [x] `continue`
- [x] `default`
- [x] `defer`
- [x] `else`
- [x] `fallthrough`
- [x] `for`
- [x] `func`
- [ ] `go`
- [x] `goto`
- [x] `if`
- [x] `import` (partially)
- [ ] `interface`
- [x] `map` (needs more testing)
- [x] `package` (partially)
- [x] `range`
- [x] `return`
- [ ] `select`
- [x] `struct` (needs more testing)
- [x] `switch`
- [x] `type` (needs more testing)
- [x] `var`

## Standard library

- [x] `fmt.Println`
- [x] `fmt.Print`
- [ ] `fmt.Printf` (partially)
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
- [x] `strings.TrimSpace`
- [ ] All the rest

One goal is that all code in the standard library should transpile correctly to C++20.
