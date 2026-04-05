.PHONY: all build test lint format clean help

# Binary name
BINARY_NAME=challenge-deps
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOFMT=gofmt
GOVET=$(GOCMD) vet

all: format lint test build ## Run format, lint, test, and build

build: ## Build the binary
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/challenge-deps
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run tests with coverage
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Coverage report:"
	$(GOCMD) tool cover -func=coverage.out | tail -1

test-coverage: test ## Run tests and show coverage in browser
	$(GOCMD) tool cover -html=coverage.out

lint: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

format: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -w .
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install it with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOGET) -v ./...

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest

run: build ## Build and run with test data
	@echo "Running with test data..."
	./$(BUILD_DIR)/$(BINARY_NAME) --repo ./testdata/sample-repo --base main --head HEAD

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
