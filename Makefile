.PHONY: build run test test-race integration-test

build:
	go build -o server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

test-race:
	go test -race ./...

integration-test:
	go test ./tests/integration/...
