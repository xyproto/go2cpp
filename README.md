# tinygocompiler

Compiles Go to native executables via C++17.

## Known problems

* Only works with extremely simple code samples, for now.

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

```
// Multiple return
package main

import (
	"fmt"
)

func addsub(x int) (a int, b int) {
	return x+2, x-2
}

func main() {
	y,z := addsub(4)
	fmt.Println("y =", y)
	fmt.Println("z =", z)
}
```

**C++ output:**

```
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

## General info

* Version: 0.0.0
* License: MIT
