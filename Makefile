# Makefile for gydnc

.PHONY: all build test fmt lint clean version-info help

BINARY_NAME=gydnc
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "v0.0.0-dev")
COMMIT_SHA ?= $(shell git rev-parse HEAD)
SHORT_SHA ?= $(shell git rev-parse --short=7 HEAD)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT_SHA) -X main.buildTime=$(BUILD_TIME)

all: build

help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

version-info: ## Display version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT_SHA)"
	@echo "Short SHA: $(SHORT_SHA)"
	@echo "Build Time: $(BUILD_TIME)"

build: fmt lint ## Build the binary for current platform
	@echo "Building $(BINARY_NAME)..."
	@PLATFORM=$$(go env GOOS)-$$(go env GOARCH); \
	VERSION_WITH_BUILD="$(VERSION)+sha.$(SHORT_SHA).$${PLATFORM}"; \
	echo "$${VERSION_WITH_BUILD}" > cmd/version.txt; \
	go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "Binary: $(BINARY_NAME)"
	@./$(BINARY_NAME) version

test: ## Run unit tests
	@echo "Running unit tests..."
	go test ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@if [ -n "$(DIR)" ]; then \
		echo "Filtering integration tests to directory: $(DIR)"; \
		GYDNC_TEST_SUITE_DIR=$(DIR) go test ./tests -v -tags=integration; \
	else \
		go test ./tests -v -tags=integration; \
	fi

.PHONY: test-session
test-session: build ## Creates a temporary directory with a specific test sample and the built gydnc binary for manual debugging.
	@if [ -z "$(DIR)" ]; then \
		echo "Error: DIR variable must be set. Example: make test-session DIR=cmd_samples/create/01_create_simple"; \
		exit 1; \
	fi
	@SESSION_DIR=$$(mktemp -d); \
	echo "Setting up test session in: $${SESSION_DIR}"; \
	cp -r tests/$(DIR)/* "$${SESSION_DIR}/"; \
	cp $(BINARY_NAME) "$${SESSION_DIR}/"; \
	echo "Test session environment created."; \
	echo "To start debugging, run:"; \
	echo "  cd $${SESSION_DIR}"; \
	echo "  # Start your preferred shell, e.g., zsh or bash"; \
	echo "  zsh"; \
	echo "  # You can now run ./$(BINARY_NAME) commands against the test files."

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Lint code
	@echo "Linting code... (requires golangci-lint)"
	@golangci-lint run || echo "Warning: golangci-lint not found or lint errors detected."

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	go clean

install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(shell go env GOPATH)/bin..."
	cp $(BINARY_NAME) $(shell go env GOPATH)/bin/