# Makefile
.PHONY: fmt vet lint test build

fmt:
	@test -z "$$(gofmt -l .)"

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -race ./...

build:
	go build ./...
