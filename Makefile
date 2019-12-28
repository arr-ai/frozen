.PHONY: all test lint

all: lint test

test:
	go test ./...

lint:
	golangci-lint run
