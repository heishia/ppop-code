.PHONY: build install uninstall clean test help

# Detect OS
ifeq ($(OS),Windows_NT)
    BINARY_NAME := ppopcode.exe
    INSTALL_SCRIPT := .\install.ps1
    UNINSTALL_SCRIPT := .\install.ps1 -Uninstall
else
    BINARY_NAME := ppopcode
    INSTALL_SCRIPT := ./install.sh
    UNINSTALL_SCRIPT := ./install.sh uninstall
endif

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go mod tidy
	go build -o $(BINARY_NAME) ./cmd/ppopcode
	@echo "✓ Build complete: $(BINARY_NAME)"

install: build ## Build and install globally
	@echo "Installing ppopcode..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File install.ps1
else
	chmod +x install.sh
	./install.sh
endif

uninstall: ## Uninstall ppopcode
	@echo "Uninstalling ppopcode..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File install.ps1 -Uninstall
else
	./install.sh uninstall
endif

clean: ## Remove built binaries
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	@echo "✓ Clean complete"

test: ## Run tests
	@echo "Running tests..."
	go test ./...

run: build ## Build and run locally
	./$(BINARY_NAME)

.DEFAULT_GOAL := help
