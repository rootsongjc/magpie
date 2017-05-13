all: build
.PHONY: build 

export GOOS= linux

export GOARCH=amd64

build:
	go build -o magpie main.go
