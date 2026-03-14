BIN_PATH = bin/hexlet-url-shortener

build:
	go build -ldflags="-w -s" -gcflags=all="-l -B" -o $(BIN_PATH) .

lint:
	golangci-lint run ./...

test:
	go test -v ./...

sqlc:
	sqlc generate

clean:
	rm bin/* || true # Ignore errors
