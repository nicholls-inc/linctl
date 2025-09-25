# linctl Makefile

.PHONY: build clean test install lint fmt deps help

# Build variables
BINARY_NAME=linctl
GO_FILES=$(shell find . -type f -name '*.go' | grep -v vendor/)
VERSION?=$(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X github.com/dorkitude/linctl/cmd.version=$(VERSION)"

# Default target
all: build

# Build the binary
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	go clean

# Run unit tests only
test-unit:
	@echo "🧪 Running unit tests..."
	@go test ./...

# Run unit tests with verbose output
test-unit-verbose:
	@echo "🧪 Running unit tests (verbose)..."
	@go test -v ./...

# Run OAuth integration tests
test-oauth:
	@echo "🧪 Running OAuth integration tests..."
	@./test_oauth_integration.sh

# Run smoke tests (requires authentication)
test-smoke:
	@echo "🧪 Running smoke tests..."
	@./smoke_test.sh

# Run smoke tests with verbose output
test-smoke-verbose:
	@echo "🧪 Running smoke tests (verbose)..."
	@bash -x ./smoke_test.sh

# Run all tests (unit + OAuth integration)
test:
	@echo "🧪 Running all tests..."
	@$(MAKE) test-unit
	@$(MAKE) test-oauth

# Run all tests including smoke tests (requires authentication)
test-all:
	@echo "🧪 Running all tests including smoke tests..."
	@$(MAKE) test-unit
	@$(MAKE) test-oauth
	@$(MAKE) test-smoke

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run

# Install binary to system
install: build
	@echo "📦 Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo mv $(BINARY_NAME) /usr/local/bin/

# Development installation (symlink)
dev-install: build
	@echo "🔗 Creating development symlink..."
	sudo ln -sf $(PWD)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# Cross-compile for multiple platforms
build-all:
	@echo "🌍 Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

# Create release directory
release: clean
	@echo "🚀 Preparing release..."
	mkdir -p dist
	$(MAKE) build-all

# Run the binary
run: build
	./$(BINARY_NAME)

# Run everything: build, format, lint, test, and install
everything: build fmt lint test install
	@echo "✅ Everything complete!"

# Show help
help:
	@echo "📖 Available targets:"
	@echo "  build            - Build the binary"
	@echo "  clean            - Clean build artifacts"
	@echo "  test             - Run unit tests and OAuth integration tests"
	@echo "  test-unit        - Run unit tests only"
	@echo "  test-oauth       - Run OAuth integration tests"
	@echo "  test-smoke       - Run smoke tests (requires authentication)"
	@echo "  test-all         - Run all tests including smoke tests"
	@echo "  deps             - Install dependencies"
	@echo "  fmt              - Format code"
	@echo "  lint             - Lint code"
	@echo "  install          - Install binary to system"
	@echo "  dev-install      - Create development symlink"
	@echo "  build-all        - Cross-compile for all platforms"
	@echo "  release          - Prepare release builds"
	@echo "  run              - Build and run the binary"
	@echo "  everything       - Run build, fmt, lint, test, and install"
	@echo "  help             - Show this help"
