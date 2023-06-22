clean: .PHONY
	rm -rf bin/*

all: server

server:
	go build -o bin/server ./cmd/server