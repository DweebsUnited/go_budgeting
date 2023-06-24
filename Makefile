.PHONY: clean all
all: bin/server

bin/server:
	go build -o bin/server ./cmd/server

clean:
	rm -rf bin/*