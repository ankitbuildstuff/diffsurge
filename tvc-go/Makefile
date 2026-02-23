.PHONY: build build-cli build-proxy build-api test lint clean dev fmt vet

BINARY_DIR := bin
CLI_BINARY := $(BINARY_DIR)/surge
PROXY_BINARY := $(BINARY_DIR)/tvc-proxy
API_BINARY := $(BINARY_DIR)/tvc-api

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/tvc-org/tvc/internal/config.Version=$(VERSION)"

build: build-cli build-proxy build-api

build-cli:
	@echo "Building CLI..."
	@go build $(LDFLAGS) -o $(CLI_BINARY) ./cmd/cli

build-proxy:
	@echo "Building Proxy..."
	@go build $(LDFLAGS) -o $(PROXY_BINARY) ./cmd/proxy

build-api:
	@echo "Building API..."
	@go build $(LDFLAGS) -o $(API_BINARY) ./cmd/api

test:
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

lint:
	@golangci-lint run ./...

fmt:
	@gofmt -s -w .

vet:
	@go vet ./...

clean:
	@rm -rf $(BINARY_DIR) coverage.out

dev:
	@docker compose up -d
	@echo "Dev services started (Postgres + Redis)"

dev-down:
	@docker compose down

build-all: clean
	@echo "Building for all platforms..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_DIR)/surge-darwin-amd64 ./cmd/cli
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_DIR)/surge-darwin-arm64 ./cmd/cli
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_DIR)/surge-linux-amd64 ./cmd/cli
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_DIR)/surge-linux-arm64 ./cmd/cli
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_DIR)/surge-windows-amd64.exe ./cmd/cli
	@echo "Done. Binaries in $(BINARY_DIR)/"
