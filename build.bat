go build -tags "sqlite_math_functions" -o bin\server.exe ./cmd/server
go build -tags "sqlite_math_functions" -o bin\querytool.exe ./tools/querytool
go build -tags "sqlite_math_functions" -o bin\migrate.exe ./tools/buckets_to_db