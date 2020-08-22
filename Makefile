# The main purpose of this Makefile is to make Travis CI
# NOT build the testcase/*.go files.

.PHONY: clean test

SRC := $(wildcard *.go)

all: $(SRC)
	go build && go test

clean:
	rm -f go2cpp
