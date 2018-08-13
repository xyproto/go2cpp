#!/bin/sh
#
# examine.sh depends on go, sakemake and bat
#
go build
fn=$(go test | grep "FAILED:" | cut -d" " -f4)
echo 'Go program:'
bat "$fn"
echo 'C++17 program:'
mkdir -p test
./go2cpp "$fn" -o | tee test/main.cpp | bat -l cpp
go run "$fn"
grep -q main test/main.cpp && sm -C test/ run || echo 'No C++ program to run.'
