setup:
	go mod download

build:
	CGO_ENABLED=0 go build -ldflags="-s -w"

lint:
	golangci-lint run

.DEFAULT_GOAL := build
