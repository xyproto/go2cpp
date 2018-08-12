# go2cpp

Compiles Go to native executables via C++17.

One of the goals is for the compiler to be able to compile itself.

## Known issues

* Only works with simple code samples, for now.
* Very few functions from the Go standard library are implemented.
* A plan for how to do "import" is needed.
* printing a byte is outputted as the number in Go, but as the character in C++ converted from Go:
  `std::uint8_t b = 65; std::cout << b << std::endl;` vs `var b: byte = 64; fmt.Println(b)`
  I found no good way to convince `std::out` to output the number of a byte instead of the character.

## Features and limitations

* Pretty fast.
* Simple to use.
* Few dependencies (for compiling `go2cpp`, only the go compiler is needed).
* Low complexity.
* Short source code.
* `g++` is used for compiling the generated C++ code.
* `clang-format` is used for formatting the generated C++ code.


## Usage

Compile to executable:

    go2cpp main.go -o main

Output what the intermediate C++17 code looks like:

    go2cpp main.go

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

- [x] `break`
- [x] `case`
- [ ] `chan`
- [x] `const`
- [x] `continue`
- [x] `default`
- [ ] `defer`
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
- [ ] `struct`
- [x] `switch`
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
