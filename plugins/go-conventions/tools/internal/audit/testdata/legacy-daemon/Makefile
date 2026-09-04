PKG := example.com/daemon
VERSION ?= $(shell git describe --tags --always)

build:
	go build -ldflags "-s -w -X $(PKG)/internal/build.Version=$(VERSION)" -o bin/daemon .

test:
	go test ./...
