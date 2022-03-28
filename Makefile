test:
	go test -v -count=1 -race -timeout 10s ./internal/...

test-all: test

run:
	go run ./cmd/httpserver
