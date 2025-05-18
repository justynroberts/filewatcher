# Makefile for FileWatcher

# Variables
BINARY_NAME=watcher
VERSION=1.0.0
BUILD_DIR=dist
GO_FILES=$(wildcard *.go)

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	@echo "Building for current platform..."
	@go build -o $(BINARY_NAME) -ldflags "-X main.Version=$(VERSION)" main.go
	@echo "Build complete: $(BINARY_NAME)"

# Run the application
.PHONY: run
run: build
	@echo "Running application..."
	@./$(BINARY_NAME)

# Build for all platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@go run tools/build.go -output $(BUILD_DIR) -version $(VERSION)

# Build only for current platform using the build script
.PHONY: build-current
build-current:
	@echo "Building for current platform using build script..."
	@go run tools/build.go -output $(BUILD_DIR) -version $(VERSION) -current

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@echo "Dependencies installed"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Help target
.PHONY: help
help:
	@echo "FileWatcher Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build for current platform"
	@echo "  make run          Build and run the application"
	@echo "  make build-all    Build for all platforms"
	@echo "  make build-current Build only for current platform using build script"
	@echo "  make clean        Clean build artifacts"
	@echo "  make deps         Install dependencies"
	@echo "  make test         Run tests"
	@echo "  make help         Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  BINARY_NAME       Binary name (default: watcher)"
	@echo "  VERSION           Version number (default: 1.0.0)"
	@echo "  BUILD_DIR         Build directory (default: dist)"