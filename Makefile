# tf-safe Makefile
# Cross-platform build system for tf-safe CLI tool

# Build variables
BINARY_NAME=tf-safe
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Directories
BUILD_DIR=build
DIST_DIR=dist

# Supported platforms
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build clean test lint deps help package-all

# Default target
all: clean deps test build

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping lint check"; \
	fi

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for all platforms
build-all: clean
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "Building $$output_name..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) -o $(DIST_DIR)/$$output_name .; \
		if [ $$? -ne 0 ]; then \
			echo "Failed to build $$output_name"; \
			exit 1; \
		fi; \
	done

# Create release archives
package: build-all
	@echo "Creating release packages..."
	@cd $(DIST_DIR) && \
	for file in $(BINARY_NAME)-*; do \
		if [[ $$file == *".exe" ]]; then \
			zip $$file.zip $$file; \
		else \
			tar -czf $$file.tar.gz $$file; \
		fi; \
		echo "Created package for $$file"; \
	done

# Install locally (requires sudo on some systems)
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(BINARY_NAME) installed successfully"

# Uninstall
uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled successfully"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)

# Check binary sizes
check-size: build-all
	@echo "Checking binary sizes..."
	@for file in $(DIST_DIR)/$(BINARY_NAME)-*; do \
		size=$$(stat -f%z "$$file" 2>/dev/null || stat -c%s "$$file" 2>/dev/null); \
		size_mb=$$(echo "scale=2; $$size/1024/1024" | bc -l 2>/dev/null || echo "$$((size/1024/1024))"); \
		echo "$$file: $${size_mb}MB"; \
		if [ $$size -gt 15728640 ]; then \
			echo "WARNING: $$file exceeds 15MB size limit"; \
		fi; \
	done

# Development build with race detection
dev:
	@echo "Building development version with race detection..."
	@mkdir -p $(BUILD_DIR)
	go build -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-dev .

# Create installation packages
package-all: build-all
	@echo "Creating installation packages..."
	@./scripts/package-all.sh $(VERSION)

# Show help
help:
	@echo "tf-safe Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  all         - Clean, install deps, test, and build"
	@echo "  deps        - Install Go dependencies"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter (requires golangci-lint)"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all supported platforms"
	@echo "  package     - Create release archives"
	@echo "  package-all - Create all installation packages (DEB, Chocolatey, Homebrew)"
	@echo "  install     - Install binary to /usr/local/bin"
	@echo "  uninstall   - Remove binary from /usr/local/bin"
	@echo "  clean       - Clean build artifacts"
	@echo "  check-size  - Check binary sizes against 15MB limit"
	@echo "  dev         - Build development version with race detection"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION     - Version string (default: git describe or 'dev')"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=v1.0.0"
	@echo "  make build-all"
	@echo "  make package"
	@echo "  make package-all VERSION=v1.0.0"