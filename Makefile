.PHONY: build test test-unit test-integration test-all clean lint fmt vet coverage help install

# Default target
all: fmt vet test build

# Build the binary
build:
	@echo "Building gh-issue-dependency..."
	go build -o gh-issue-dependency .
	@echo "Build complete!"

# Install the binary
install: build
	@echo "Installing gh-issue-dependency..."
	go install .
	@echo "Installation complete!"

# Run all tests
test: test-unit test-integration

# Run only unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v ./...

# Run unit tests with coverage
test-coverage:
	@echo "Running unit tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run unit tests with coverage and show in terminal
coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run integration tests
test-integration: build
	@echo "Running integration tests..."
	@if command -v bash >/dev/null 2>&1; then \
		bash tests/integration_test.sh; \
	else \
		@echo "Warning: bash not found, skipping integration tests"; \
	fi

# Run all tests (unit + integration)
test-all: test-unit test-integration

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f gh-issue-dependency gh-issue-dependency.exe
	rm -f coverage.out coverage.html
	go clean

# Run tests in watch mode (requires entr)
test-watch:
	@echo "Running tests in watch mode..."
	@if command -v find >/dev/null 2>&1 && command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -c make test-unit; \
	else \
		echo "find and/or entr not found"; \
		echo "Install entr for watch mode: http://eradman.com/entrproject/"; \
	fi

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Tidy go modules
tidy:
	@echo "Tidying go modules..."
	go mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	go mod verify

# Update dependencies
update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/gh-issue-dependency-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o dist/gh-issue-dependency-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o dist/gh-issue-dependency-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o dist/gh-issue-dependency-windows-amd64.exe .
	@echo "Cross-platform builds complete in dist/"

# Development workflow - format, vet, test, build
dev: fmt vet test-unit build

# CI workflow - comprehensive testing and validation
ci: fmt vet lint test-coverage test-integration

# Quick test - just unit tests without verbose output
quick-test:
	@echo "Running quick unit tests..."
	go test ./...

# Security scan (requires gosec)
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found, skipping security scan"; \
		echo "Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate mocks (if using mockery)
mocks:
	@echo "Generating mocks..."
	@if command -v mockery >/dev/null 2>&1; then \
		mockery --all; \
	else \
		echo "mockery not found, skipping mock generation"; \
		echo "Install with: go install github.com/vektra/mockery/v2@latest"; \
	fi

# Help target
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  install         - Build and install the binary"
	@echo "  test            - Run all tests (unit + integration)"
	@echo "  test-unit       - Run only unit tests"
	@echo "  test-integration - Run only integration tests"
	@echo "  test-all        - Run all tests (alias for test)"
	@echo "  test-coverage   - Run unit tests with coverage report"
	@echo "  coverage        - Show test coverage in terminal"
	@echo "  test-watch      - Run tests in watch mode"
	@echo "  bench           - Run benchmarks"
	@echo "  quick-test      - Run unit tests without verbose output"
	@echo "  fmt             - Format code"
	@echo "  vet             - Run go vet"
	@echo "  lint            - Run linter (requires golangci-lint)"
	@echo "  security        - Run security scan (requires gosec)"
	@echo "  clean           - Remove build artifacts"
	@echo "  tidy            - Tidy go modules"
	@echo "  deps            - Download dependencies"
	@echo "  verify          - Verify dependencies"
	@echo "  update          - Update dependencies"
	@echo "  build-all       - Build for multiple platforms"
	@echo "  mocks           - Generate mocks (requires mockery)"
	@echo "  dev             - Development workflow (fmt, vet, test, build)"
	@echo "  ci              - CI workflow (comprehensive testing)"
	@echo "  help            - Show this help message"