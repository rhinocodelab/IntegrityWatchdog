# FIM Tool Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=fim

# Build flags for static linking
LDFLAGS=-ldflags "-s -w"

# Default target
all: test build

# Build for current platform
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v

# Build for Linux/Unix (static binary)
build-unix:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v
	mv $(BINARY_NAME) $(BINARY_NAME)_linux_amd64

# Build for MacOS ARM64 (Apple Silicon)
build-mac-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v
	mv $(BINARY_NAME) $(BINARY_NAME)_darwin_arm64

# Build for MacOS AMD64 (Intel)
build-mac-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v
	mv $(BINARY_NAME) $(BINARY_NAME)_darwin_amd64

# Build all platforms
build-all: build-unix build-mac-arm64 build-mac-amd64

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)_linux_amd64
	rm -f $(BINARY_NAME)_darwin_arm64
	rm -f $(BINARY_NAME)_darwin_amd64

# Install dependencies
deps:
	$(GOGET) -v -t -d ./...

# Run linter
lint:
	golangci-lint run

# Help target
help:
	@echo "Available targets:"
	@echo "  all          - Run tests and build for current platform"
	@echo "  build        - Build for current platform"
	@echo "  build-unix   - Build static binary for Linux/Unix"
	@echo "  build-mac-arm64  - Build static binary for MacOS ARM64"
	@echo "  build-mac-amd64  - Build static binary for MacOS AMD64"
	@echo "  build-all    - Build for all platforms"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  lint         - Run linter"
	@echo "  help         - Show this help message"

.PHONY: all build build-unix build-mac-arm64 build-mac-amd64 build-all test clean deps lint help 