.PHONY: clean all bin/server bin/querytool

all: bin/server bin/querytool

bin/server:
	go build -tags "sqlite_math_functions" -o bin/server ./cmd/server

bin/querytool:
	go build -tags "sqlite_math_functions" -o bin/querytool ./tools/querytool

clean:
	rm -rf bin/*