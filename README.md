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

## General info

* Version: 0.0.0
* License: MIT
