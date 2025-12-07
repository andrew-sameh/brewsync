# BrewSync Makefile
# Provides common commands for building, testing, and development

.PHONY: help build install test test-coverage test-verbose clean run fmt lint vet dev doctor release

# Variables
BINARY_NAME=brewsync
BUILD_DIR=./bin
CMD_DIR=./cmd/brewsync
INSTALL_PATH=$(shell go env GOPATH)/bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/andrew-sameh/brewsync/pkg/version.Version=$(VERSION) \
                  -X github.com/andrew-sameh/brewsync/pkg/version.Commit=$(COMMIT) \
                  -X github.com/andrew-sameh/brewsync/pkg/version.Date=$(BUILD_DATE)"

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

##@ General

help: ## Display this help message
	@echo "$(COLOR_BOLD)BrewSync - Development Commands$(COLOR_RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(COLOR_BLUE)<target>$(COLOR_RESET)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(COLOR_BLUE)%-15s$(COLOR_RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(COLOR_BOLD)%s$(COLOR_RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Building

build: ## Build the binary
	@echo "$(COLOR_GREEN)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

build-local: ## Build binary in current directory
	@echo "$(COLOR_GREEN)Building $(BINARY_NAME) locally...$(COLOR_RESET)"
	@go build $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Built: ./$(BINARY_NAME)$(COLOR_RESET)"

install: ## Install binary to GOPATH/bin
	@echo "$(COLOR_GREEN)Installing $(BINARY_NAME) to $(INSTALL_PATH)...$(COLOR_RESET)"
	@go install $(LDFLAGS) $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Installed: $(INSTALL_PATH)/$(BINARY_NAME)$(COLOR_RESET)"

release: ## Build optimized release binary
	@echo "$(COLOR_GREEN)Building release binary...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Release built: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

##@ Testing

test: ## Run all tests
	@echo "$(COLOR_YELLOW)Running tests...$(COLOR_RESET)"
	@go test ./...

test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_YELLOW)Running tests with coverage...$(COLOR_RESET)"
	@go test ./... -cover

test-coverage-detail: ## Run tests with detailed coverage
	@echo "$(COLOR_YELLOW)Running tests with detailed coverage...$(COLOR_RESET)"
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: coverage.html$(COLOR_RESET)"

test-verbose: ## Run tests with verbose output
	@echo "$(COLOR_YELLOW)Running tests (verbose)...$(COLOR_RESET)"
	@go test ./... -v

test-race: ## Run tests with race detector
	@echo "$(COLOR_YELLOW)Running tests with race detector...$(COLOR_RESET)"
	@go test ./... -race

test-bench: ## Run benchmark tests
	@echo "$(COLOR_YELLOW)Running benchmark tests...$(COLOR_RESET)"
	@go test ./... -bench=. -benchmem

test-specific: ## Run specific package tests (usage: make test-specific PKG=./internal/brewfile)
	@echo "$(COLOR_YELLOW)Running tests for $(PKG)...$(COLOR_RESET)"
	@go test $(PKG) -v

##@ Code Quality

fmt: ## Format code with gofmt
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	@go fmt ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

vet: ## Run go vet
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	@go vet ./...
	@echo "$(COLOR_GREEN)✓ Vet passed$(COLOR_RESET)"

lint: ## Run golangci-lint (requires golangci-lint installed)
	@echo "$(COLOR_BLUE)Running golangci-lint...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(COLOR_GREEN)✓ Lint passed$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed. Install with: brew install golangci-lint$(COLOR_RESET)"; \
	fi

check: fmt vet ## Run formatting and vet checks

##@ Development

dev: build-local ## Build and show version
	@echo ""
	@./$(BINARY_NAME) --version

run: build-local ## Build and run with args (usage: make run ARGS="status")
	@echo "$(COLOR_GREEN)Running $(BINARY_NAME) $(ARGS)...$(COLOR_RESET)"
	@./$(BINARY_NAME) $(ARGS)

doctor: build-local ## Build and run doctor command
	@echo "$(COLOR_GREEN)Running brewsync doctor...$(COLOR_RESET)"
	@./$(BINARY_NAME) doctor

debug: ## Build with debug symbols and run
	@echo "$(COLOR_YELLOW)Building with debug symbols...$(COLOR_RESET)"
	@go build -gcflags="all=-N -l" -o $(BINARY_NAME)-debug $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Debug binary: ./$(BINARY_NAME)-debug$(COLOR_RESET)"

##@ Dependencies

deps: ## Download dependencies
	@echo "$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)"
	@go mod download
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

deps-tidy: ## Tidy dependencies
	@echo "$(COLOR_BLUE)Tidying dependencies...$(COLOR_RESET)"
	@go mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies tidied$(COLOR_RESET)"

deps-verify: ## Verify dependencies
	@echo "$(COLOR_BLUE)Verifying dependencies...$(COLOR_RESET)"
	@go mod verify
	@echo "$(COLOR_GREEN)✓ Dependencies verified$(COLOR_RESET)"

deps-update: ## Update all dependencies
	@echo "$(COLOR_BLUE)Updating dependencies...$(COLOR_RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"

##@ Cleanup

clean: ## Remove build artifacts
	@echo "$(COLOR_YELLOW)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-debug
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

clean-all: clean ## Remove build artifacts and test cache
	@echo "$(COLOR_YELLOW)Cleaning all...$(COLOR_RESET)"
	@go clean -testcache
	@echo "$(COLOR_GREEN)✓ All cleaned$(COLOR_RESET)"

##@ Manual Testing

test-setup: build-local ## Setup test environment for manual testing
	@echo "$(COLOR_BLUE)Setting up test environment...$(COLOR_RESET)"
	@mkdir -p ~/brewsync-test-dotfiles/_brew_test1
	@mkdir -p ~/brewsync-test-dotfiles/_brew_test2
	@if [ -d ~/.config/brewsync ]; then \
		echo "$(COLOR_YELLOW)⚠ Backing up existing config to ~/.config/brewsync.backup$(COLOR_RESET)"; \
		mv ~/.config/brewsync ~/.config/brewsync.backup 2>/dev/null || true; \
	fi
	@mkdir -p ~/.config/brewsync/profiles
	@echo "$(COLOR_GREEN)✓ Test environment ready$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Next steps:$(COLOR_RESET)"
	@echo "  1. Run: ./$(BINARY_NAME) config init"
	@echo "  2. Follow: MANUAL_TEST_GUIDE.md"

test-cleanup: ## Cleanup test environment
	@echo "$(COLOR_YELLOW)Cleaning up test environment...$(COLOR_RESET)"
	@rm -rf ~/brewsync-test-dotfiles
	@rm -rf ~/.config/brewsync
	@if [ -d ~/.config/brewsync.backup ]; then \
		echo "$(COLOR_BLUE)Restoring original config...$(COLOR_RESET)"; \
		mv ~/.config/brewsync.backup ~/.config/brewsync; \
	fi
	@echo "$(COLOR_GREEN)✓ Test environment cleaned$(COLOR_RESET)"

##@ Git & Release

tag: ## Create a new git tag (usage: make tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(COLOR_YELLOW)Usage: make tag VERSION=v1.0.0$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BLUE)Creating tag $(VERSION)...$(COLOR_RESET)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "$(COLOR_GREEN)✓ Tag created: $(VERSION)$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Push tag with: git push origin $(VERSION)$(COLOR_RESET)"

status: ## Show git status and build info
	@echo "$(COLOR_BOLD)Project Status$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BLUE)Version:$(COLOR_RESET)     $(VERSION)"
	@echo "$(COLOR_BLUE)Commit:$(COLOR_RESET)      $(COMMIT)"
	@echo "$(COLOR_BLUE)Build Date:$(COLOR_RESET)  $(BUILD_DATE)"
	@echo ""
	@echo "$(COLOR_BOLD)Git Status:$(COLOR_RESET)"
	@git status -s
	@echo ""
	@echo "$(COLOR_BOLD)Recent Commits:$(COLOR_RESET)"
	@git log --oneline -5

##@ Quick Actions

all: clean build test ## Clean, build, and test
	@echo "$(COLOR_GREEN)✓ All tasks completed$(COLOR_RESET)"

quick: build-local test ## Quick build and test
	@echo "$(COLOR_GREEN)✓ Quick build and test completed$(COLOR_RESET)"

ci: fmt vet test ## Run CI checks (format, vet, test)
	@echo "$(COLOR_GREEN)✓ CI checks passed$(COLOR_RESET)"

pre-commit: fmt vet test-coverage ## Run pre-commit checks
	@echo "$(COLOR_GREEN)✓ Pre-commit checks passed$(COLOR_RESET)"

##@ Information

version: ## Show current version
	@echo "$(VERSION)"

info: ## Show project information
	@echo "$(COLOR_BOLD)BrewSync Project Information$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BLUE)Binary Name:$(COLOR_RESET)      $(BINARY_NAME)"
	@echo "$(COLOR_BLUE)Build Directory:$(COLOR_RESET)  $(BUILD_DIR)"
	@echo "$(COLOR_BLUE)Install Path:$(COLOR_RESET)     $(INSTALL_PATH)"
	@echo "$(COLOR_BLUE)Version:$(COLOR_RESET)          $(VERSION)"
	@echo "$(COLOR_BLUE)Commit:$(COLOR_RESET)           $(COMMIT)"
	@echo "$(COLOR_BLUE)Build Date:$(COLOR_RESET)       $(BUILD_DATE)"
	@echo ""
	@echo "$(COLOR_BOLD)Go Environment:$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Go Version:$(COLOR_RESET)       $$(go version)"
	@echo "$(COLOR_BLUE)GOPATH:$(COLOR_RESET)           $$(go env GOPATH)"
	@echo "$(COLOR_BLUE)GOOS:$(COLOR_RESET)             $$(go env GOOS)"
	@echo "$(COLOR_BLUE)GOARCH:$(COLOR_RESET)           $$(go env GOARCH)"

list: ## List all available targets
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@$(MAKE) -pRrq -f $(firstword $(MAKEFILE_LIST)) : 2>/dev/null | \
		awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | \
		sort | grep -E -v -e '^[^[:alnum:]]' -e '^$@$$'
