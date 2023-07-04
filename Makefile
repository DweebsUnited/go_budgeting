.PHONY: clean all bin/server bin/querytool

all: bin/server bin/querytool

bin/server:
	go build -o bin/server ./cmd/server

bin/querytool:
	go build -o bin/querytool ./tools/querytool

clean:
	rm -rf bin/*