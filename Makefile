test:
	go test -v -race -timeout 10s ./internal/...

run:
	go run ./cmd/httpserver
