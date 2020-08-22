.PHONY: clean test

SRC := $(wildcard *.go)

all: $(SRC)
	go build

test: $(SRC)
	go test

clean:
	rm -f go2cpp
