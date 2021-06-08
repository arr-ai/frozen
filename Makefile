.PHONY: all test lint

all: lint test

test:
	go test -coverprofile cover.prof $(TESTFLAGS) ./...

lint:
	golangci-lint run
