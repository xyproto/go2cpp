# tinygocompiler

Compiles Go to native executables via C++17.

Only works for extremely basic code samples, for now.

Work in progress! It's very early days.

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
