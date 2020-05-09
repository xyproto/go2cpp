#!/bin/sh
#
# examine.sh depends on go, cxx (https://github.com/xyproto/cxx) and bat
#
go build
fn=$(go test | grep "Compiling and running" | cut -d"]" -f2- | cut -d" " -f5 | tail -1)
if [ ! -e "$fn" ]; then
  echo "No issues."
  exit 0
fi
echo 'Go program:'
bat "$fn" || exit 1
echo 'C++20 program:'
mkdir -p test
./go2cpp "$fn" -o | tee test/main.cpp | bat -l cpp
echo "Go output:"
go run "$fn"
grep -q main test/main.cpp && sm -C test >/dev/null || exit 1
echo "C++ output:"
sm -C test run
