test:
	go test -v -race -timeout 10s ./internal/...

test-all: test

run:
	go run ./cmd/httpserver
