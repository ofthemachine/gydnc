# Basic Makefile for Agent Guidance Service

.PHONY: all build test fmt lint clean

BINARY_NAME=gydnc

all: build

build: fmt lint
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "Binary: $(BINARY_NAME)"

# Remove old server and cli targets
# server: fmt lint ...
# cli: fmt lint ...

test:
	@echo "Running unit tests..."
	go test ./...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@if [ -n "$(DIR)" ]; then \
		echo "Filtering integration tests to directory: $(DIR)"; \
		GYDNC_TEST_SUITE_DIR=$(DIR) go test ./tests -v -tags=integration; \
	else \
		go test ./tests -v -tags=integration; \
	fi

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Linting code... (requires golangci-lint)"
	@golangci-lint run || echo "Warning: golangci-lint not found or lint errors detected."

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME) $(BINARY_NAME)-server-tmp $(BINARY_NAME)-cli-tmp # Keep cleaning old temp files just in case
	rm -f $(BINARY_NAME) # Clean the new binary
	go clean