.PHONY: build test_all

build:
	go build ./...

test_all:
	go test -count=1 -v ./...
