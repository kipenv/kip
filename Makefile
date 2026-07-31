.PHONY: build test lint server dev clean

BINARY=bin/kip
SERVER_BINARY=bin/kip-server

build:
	go build -o $(BINARY) ./cmd/kip/

server-build:
	go build -o $(SERVER_BINARY) ./server/cmd/server/

server: server-build
	$(SERVER_BINARY)

test:
	go test -race -cover ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

dev:
	docker compose -f deploy/docker-compose.dev.yml up

clean:
	rm -rf bin/ coverage.out coverage.html
