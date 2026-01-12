# Makefile for omc (OpenShift Must-Gather CLI)

# Project variables
PROJECT_NAME := omc
BINARY_NAME := omc
PACKAGE := github.com/gmeghnag/omc

# Version information
VERSION_TAG ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
VERSION_HASH ?= $(shell git log -n1 --pretty=format:%h 2>/dev/null || echo "unknown")
LDFLAGS := -X '$(PACKAGE)/vars.OMCVersionTag=$(VERSION_TAG)' -X '$(PACKAGE)/vars.OMCVersionHash=$(VERSION_HASH)'

# Build variables
GO := go
GOFLAGS := -v
BUILD_FLAGS := -ldflags "$(LDFLAGS)"
CGO_ENABLED ?= 0

# Directories
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage

# Platforms for cross-compilation
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

# Test variables
TEST_TIMEOUT := 10m
COVERAGE_PROFILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html

.PHONY: all
all: clean deps build test ## Build and test the project (default target)

.PHONY: help
help: ## Display this help message
	@echo "$(PROJECT_NAME) - Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-20s %s\n", $$1, $$2 } /^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build targets

.PHONY: build
build: ## Build the binary for the current platform
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-all
build-all: ## Build binaries for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@$(foreach platform,$(PLATFORMS), \
		$(call build_platform,$(platform)); \
	)
	@echo "All binaries built in $(DIST_DIR)/"

.PHONY: install
install: build ## Install the binary to GOPATH/bin or /usr/local/bin
	@echo "Installing $(BINARY_NAME)..."
	@if [ -n "$(GOPATH)" ] && [ -d "$(GOPATH)/bin" ]; then \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/; \
		echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"; \
	else \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/; \
		echo "Installed to /usr/local/bin/$(BINARY_NAME)"; \
	fi

.PHONY: uninstall
uninstall: ## Uninstall the binary from GOPATH/bin or /usr/local/bin
	@echo "Uninstalling $(BINARY_NAME)..."
	@if [ -n "$(GOPATH)" ] && [ -f "$(GOPATH)/bin/$(BINARY_NAME)" ]; then \
		rm -f $(GOPATH)/bin/$(BINARY_NAME); \
		echo "Removed from $(GOPATH)/bin/$(BINARY_NAME)"; \
	elif [ -f "/usr/local/bin/$(BINARY_NAME)" ]; then \
		sudo rm -f /usr/local/bin/$(BINARY_NAME); \
		echo "Removed from /usr/local/bin/$(BINARY_NAME)"; \
	else \
		echo "No installed binary found"; \
	fi

##@ Test targets

.PHONY: test
test: ## Run unit tests
	@echo "Running tests..."
	$(GO) test -v -timeout $(TEST_TIMEOUT) ./...
	@echo "All tests passed"

.PHONY: test-short
test-short: ## Run unit tests (short mode)
	@echo "Running tests (short mode)..."
	$(GO) test -short -v ./...

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	$(GO) test -race -v -timeout $(TEST_TIMEOUT) ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -v -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_PROFILE) -covermode=atomic ./...
	@echo "Coverage profile generated: $(COVERAGE_PROFILE)"

.PHONY: coverage-html
coverage-html: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	@echo "HTML coverage report: $(COVERAGE_HTML)"

.PHONY: coverage-func
coverage-func: test-coverage ## Show function-level coverage
	@$(GO) tool cover -func=$(COVERAGE_PROFILE)

.PHONY: bench
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

##@ Code quality targets

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted"

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "All files are formatted"

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "No issues found"

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
		echo "Linting completed"; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: staticcheck
staticcheck: ## Run staticcheck (requires staticcheck to be installed)
	@echo "Running staticcheck..."
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
		echo "Staticcheck completed"; \
	else \
		echo "staticcheck not installed. Install with:"; \
		echo "  go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

.PHONY: check
check: fmt-check vet ## Run all code quality checks

##@ Dependency targets

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded"

.PHONY: deps-tidy
deps-tidy: ## Tidy and verify dependencies
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	$(GO) mod verify
	@echo "Dependencies tidied"

.PHONY: deps-vendor
deps-vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	$(GO) mod vendor
	@echo "Dependencies vendored"

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "Dependencies updated"

##@ Release targets

.PHONY: release
release: clean check test build-all dist ## Create a release (build, test, package)
	@echo "Release build complete"

.PHONY: dist
dist: ## Create distribution archives
	@echo "Creating distribution archives..."
	@mkdir -p $(DIST_DIR)
	@cd $(DIST_DIR) && \
	for binary in omc-*; do \
		if [ -f "$$binary" ]; then \
			platform=$$(echo $$binary | sed 's/omc-//'); \
			if echo "$$binary" | grep -q windows; then \
				zip "$${binary}.zip" "$$binary"; \
			else \
				tar -czf "$${binary}.tar.gz" "$$binary"; \
			fi; \
		fi; \
	done
	@echo "Distribution archives created in $(DIST_DIR)/"

.PHONY: checksums
checksums: ## Generate checksums for release files
	@echo "Generating checksums..."
	@cd $(DIST_DIR) && \
	find . -type f \( -name "*.tar.gz" -o -name "*.zip" \) -exec md5sum {} \; | tee checksums.txt
	@echo "Checksums generated: $(DIST_DIR)/checksums.txt"

##@ Cleanup targets

.PHONY: clean
clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@rm -f $(BINARY_NAME) omc-test
	@echo "Clean complete"

.PHONY: clean-all
clean-all: clean ## Remove all generated files including vendor
	@echo "Removing vendor directory..."
	@rm -rf vendor

##@ Development targets

.PHONY: run
run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: dev
dev: ## Run in development mode (with race detector)
	@echo "Running in development mode..."
	$(GO) run -race .

.PHONY: watch
watch: ## Watch for changes and rebuild (requires entr)
	@if command -v entr >/dev/null 2>&1; then \
		echo "Watching for changes..."; \
		find . -name "*.go" | entr -r make build; \
	else \
		echo "entr not installed. Install with your package manager."; \
	fi

##@ Completion targets

.PHONY: completion-bash
completion-bash: build ## Generate bash completion script
	@echo "Generating bash completion..."
	@$(BUILD_DIR)/$(BINARY_NAME) completion bash > $(BUILD_DIR)/completion.bash
	@echo "Bash completion: $(BUILD_DIR)/completion.bash"
	@echo "  Install with: source $(BUILD_DIR)/completion.bash"

.PHONY: completion-zsh
completion-zsh: build ## Generate zsh completion script
	@echo "Generating zsh completion..."
	@$(BUILD_DIR)/$(BINARY_NAME) completion zsh > $(BUILD_DIR)/completion.zsh
	@echo "Zsh completion: $(BUILD_DIR)/completion.zsh"
	@echo "  Install with: source $(BUILD_DIR)/completion.zsh"

.PHONY: completion-fish
completion-fish: build ## Generate fish completion script
	@echo "Generating fish completion..."
	@$(BUILD_DIR)/$(BINARY_NAME) completion fish > $(BUILD_DIR)/completion.fish
	@echo "Fish completion: $(BUILD_DIR)/completion.fish"

##@ Information targets

.PHONY: version
version: ## Display version information
	@echo "Version Tag:  $(VERSION_TAG)"
	@echo "Version Hash: $(VERSION_HASH)"
	@echo "Go Version:   $$($(GO) version)"

.PHONY: info
info: ## Display project information
	@echo "Project Information:"
	@echo "  Name:         $(PROJECT_NAME)"
	@echo "  Package:      $(PACKAGE)"
	@echo "  Version Tag:  $(VERSION_TAG)"
	@echo "  Version Hash: $(VERSION_HASH)"
	@echo "  Go Version:   $$($(GO) version)"
	@echo "  Build Dir:    $(BUILD_DIR)"
	@echo "  Dist Dir:     $(DIST_DIR)"

.PHONY: list-tests
list-tests: ## List all test files
	@echo "Test files:"
	@find . -name "*_test.go" -not -path "./vendor/*" | sed 's|^\./||'

# Helper function to build for a specific platform
define build_platform
	$(eval GOOS := $(word 1,$(subst /, ,$(1))))
	$(eval GOARCH := $(word 2,$(subst /, ,$(1))))
	$(eval OUTPUT := $(DIST_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,))
	@echo "Building for $(GOOS)/$(GOARCH)..."
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 $(GO) build $(BUILD_FLAGS) -o $(OUTPUT) .
endef

# Default goal when no target is specified
.DEFAULT_GOAL := all
