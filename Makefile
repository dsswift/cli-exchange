VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y%m%d-%H%M)
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: build build-all build-linux build-darwin build-windows test lint install clean tidy hooks help

build: ## Build for current platform
	go build $(LDFLAGS) -o bin/exchange ./cmd/exchange

build-all: build-linux build-darwin build-windows ## Build for all platforms

build-linux: ## Build for linux/amd64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/exchange-linux-amd64 ./cmd/exchange

build-darwin: ## Build for darwin (amd64 + arm64)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/exchange-darwin-amd64 ./cmd/exchange
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/exchange-darwin-arm64 ./cmd/exchange

build-windows: ## Build for windows/amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/exchange-windows-amd64.exe ./cmd/exchange

test: ## Run tests with race detection and coverage
	go test -race -coverprofile=coverage.out ./...

lint: ## Run linter
	$(shell go env GOPATH)/bin/golangci-lint run

install: build ## Install to ~/.local/bin
	mkdir -p ~/.local/bin
	cp bin/exchange ~/.local/bin/exchange

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out

hooks: ## Install git pre-commit hook
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

tidy: ## Tidy module dependencies
	go mod tidy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
