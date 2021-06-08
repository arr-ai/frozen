.PHONY: all test lint

all: lint test

test:
	go test $(TESTFLAGS) ./...

lint:
	golangci-lint run
