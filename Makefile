.PHONY: build test lint clean install

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/hop-cli/hop/internal/cmd.Version=$(VERSION) \
	-X github.com/hop-cli/hop/internal/cmd.Commit=$(COMMIT) \
	-X github.com/hop-cli/hop/internal/cmd.BuildDate=$(BUILD_DATE)"

# Default target
all: build

# Build the binary
build:
	go build $(LDFLAGS) -o bin/hop ./cmd/hop

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/hop

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run the application
run:
	go run ./cmd/hop $(ARGS)
