.PHONY: all test build fmt clean

.DEFAULT_GOAL := test

all: fmt build test

test:
	go test ./...

build:
	go build ./...

fmt:
	go fmt ./...

clean:
	go clean
