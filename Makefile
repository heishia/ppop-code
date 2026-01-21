.PHONY: build install uninstall clean test help

# Detect OS
ifeq ($(OS),Windows_NT)
    BINARY_NAME := ppopcode.exe
    INSTALL_SCRIPT := .\scripts\install.ps1
    UNINSTALL_SCRIPT := .\scripts\install.ps1 -Uninstall
else
    BINARY_NAME := ppopcode
    INSTALL_SCRIPT := ./scripts/install.sh
    UNINSTALL_SCRIPT := ./scripts/install.sh uninstall
endif

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: clean ## Build the binary (auto-cleans previous build)
	@echo "Building $(BINARY_NAME)..."
	go mod tidy
	go build -o $(BINARY_NAME) ./cmd/ppopcode
	@echo "✓ Build complete: $(BINARY_NAME)"

install: build ## Build and install globally
	@echo "Installing ppopcode..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/install.ps1
else
	chmod +x scripts/install.sh
	./scripts/install.sh
endif

uninstall: ## Uninstall ppopcode
	@echo "Uninstalling ppopcode..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/install.ps1 -Uninstall
else
	./scripts/install.sh uninstall
endif

clean: ## Remove built binaries
	@echo "Cleaning..."
ifeq ($(OS),Windows_NT)
	@if exist $(BINARY_NAME) del /f $(BINARY_NAME)
else
	rm -f $(BINARY_NAME)
endif
	go clean
	@echo "✓ Clean complete"

test: ## Run tests (clean output)
	@echo "Running tests..."
	@echo ""
ifeq ($(OS),Windows_NT)
	@powershell -Command "$$results = go test ./... 2>&1; $$passed = ($$results | Select-String 'ok').Count; $$failed = ($$results | Select-String 'FAIL').Count; $$results | ForEach-Object { if ($$_ -match '^ok') { Write-Host \"  PASS: $$($$_ -replace 'ok\s+', '' -replace '\s+[\d.]+s.*', '')\" -ForegroundColor Green } elseif ($$_ -match '^FAIL') { Write-Host \"  FAIL: $$($$_ -replace 'FAIL\s+', '' -replace '\s+[\d.]+s.*', '')\" -ForegroundColor Red } }; echo ''; echo \"Summary: $$passed passed, $$failed failed\""
else
	@go test ./... 2>&1 | awk '/^ok/ {print "  \033[32mPASS:\033[0m " $$2} /^FAIL/ {print "  \033[31mFAIL:\033[0m " $$2} END {print ""}'
endif

test-v: ## Run tests with verbose output
	go test ./... -v

test-cover: ## Run tests with coverage
	go test ./... -cover

run: build ## Build and run locally
	./$(BINARY_NAME)

.DEFAULT_GOAL := help
