# MArrayCRDT Makefile
.PHONY: help test test-verbose test-race test-edge test-basic test-all coverage bench clean fmt vet lint install-tools test-specific

# Default target
.DEFAULT_GOAL := help

# Variables
PACKAGE := github.com/caslun/MArrayCRDT/marraycrdt
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Help target
help: ## Show this help message
	@echo "$(BLUE)MArrayCRDT Test Suite$(NC)"
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make [target]"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# Basic test commands
test: ## Run all tests
	@echo "$(BLUE)Running all tests...$(NC)"
	@go test ./marraycrdt

test-verbose: ## Run all tests with verbose output
	@echo "$(BLUE)Running all tests (verbose)...$(NC)"
	@go test -v ./marraycrdt 2>&1 | grep "^---"

test-race: ## Run tests with race detector
	@echo "$(BLUE)Running tests with race detector...$(NC)"
	@go test -race ./marraycrdt

# Specific test files
test-basic: ## Run only basic tests
	@echo "$(BLUE)Running basic tests...$(NC)"
	@go test -v ./marraycrdt -run "^Test(Concurrent|Swap|All).*$$"

test-edge: ## Run only edge case tests
	@echo "$(BLUE)Running edge case tests...$(NC)"
	@go test -v ./marraycrdt -run "^TestAllEdgeCases$$"

test-edge-verbose: ## Run edge cases with detailed output
	@echo "$(BLUE)Running edge case tests (verbose)...$(NC)"
	@go test -v ./marraycrdt -run "Test.*" | grep -E "(Running|Initial|After|convergence|PASS|FAIL)"

# Individual edge case tests
test-same-item: ## Test concurrent moves of same item
	@go test -v ./marraycrdt -run "^TestConcurrentMoveSameItemMultipleReplicas$$"

test-same-position: ## Test multiple items to same position
	@go test -v ./marraycrdt -run "^TestMultipleItemsToSamePosition$$"

test-overlapping: ## Test overlapping swaps
	@go test -v ./marraycrdt -run "^TestOverlappingSwaps$$"

test-circular: ## Test circular moves
	@go test -v ./marraycrdt -run "^TestCircularMoves$$"

test-delayed: ## Test delayed merge with multiple moves
	@go test -v ./marraycrdt -run "^TestDelayedMergeWithMultipleMoves$$"

test-stress: ## Run stress tests
	@echo "$(BLUE)Running stress tests...$(NC)"
	@go test -v ./marraycrdt -run "^Test.*Stress.*$$"

# Test a specific function by name
test-specific: ## Run specific test (usage: make test-specific TEST=TestName)
	@if [ -z "$(TEST)" ]; then \
		echo "$(RED)Error: TEST variable not set$(NC)"; \
		echo "Usage: make test-specific TEST=TestName"; \
		exit 1; \
	fi
	@echo "$(BLUE)Running test: $(TEST)$(NC)"
	@go test -v ./marraycrdt -run "^$(TEST)$$"

# Coverage
coverage: ## Generate test coverage report
	@echo "$(BLUE)Generating coverage report...$(NC)"
	@go test -coverprofile=$(COVERAGE_FILE) ./marraycrdt
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_HTML)$(NC)"
	@echo "Coverage summary:"
	@go tool cover -func=$(COVERAGE_FILE) | tail -1

coverage-terminal: ## Show coverage in terminal
	@echo "$(BLUE)Coverage by function:$(NC)"
	@go test -coverprofile=$(COVERAGE_FILE) ./marraycrdt > /dev/null 2>&1
	@go tool cover -func=$(COVERAGE_FILE)

# Benchmarks
bench: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./marraycrdt

bench-compare: ## Run benchmarks and save for comparison
	@echo "$(BLUE)Running benchmarks for comparison...$(NC)"
	@go test -bench=. -benchmem ./marraycrdt > bench_new.txt
	@if [ -f bench_old.txt ]; then \
		echo "$(YELLOW)Comparing with previous benchmark...$(NC)"; \
		benchstat bench_old.txt bench_new.txt; \
	else \
		echo "$(YELLOW)No previous benchmark found. Run 'make bench-save' after making changes.$(NC)"; \
	fi

bench-save: ## Save current benchmark results
	@echo "$(BLUE)Saving benchmark results...$(NC)"
	@go test -bench=. -benchmem ./marraycrdt > bench_old.txt
	@echo "$(GREEN)Benchmark saved to bench_old.txt$(NC)"

# Code quality
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)Vet complete$(NC)"

lint: ## Run golangci-lint (requires golangci-lint installed)
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Run 'make install-tools' to install.$(NC)"; \
	fi

# Testing utilities
test-loop: ## Run tests in a loop (useful for finding race conditions)
	@echo "$(BLUE)Running tests in loop (press Ctrl+C to stop)...$(NC)"
	@for i in $$(seq 1 100); do \
		echo "$(YELLOW)Run $$i:$(NC)"; \
		go test -race ./marraycrdt > /dev/null 2>&1 || (echo "$(RED)Test failed on run $$i$(NC)" && exit 1); \
	done
	@echo "$(GREEN)All 100 runs passed!$(NC)"

test-watch: ## Watch for changes and run tests (requires entr)
	@if command -v entr >/dev/null 2>&1; then \
		echo "$(BLUE)Watching for changes...$(NC)"; \
		find . -name "*.go" | entr -c make test-verbose; \
	else \
		echo "$(RED)entr not installed. Install with: apt-get install entr (or brew install entr on Mac)$(NC)"; \
	fi

# Build and clean
build: ## Build the package
	@echo "$(BLUE)Building package...$(NC)"
	@go build ./marraycrdt
	@echo "$(GREEN)Build complete$(NC)"

clean: ## Clean test cache and coverage files
	@echo "$(BLUE)Cleaning...$(NC)"
	@go clean -testcache
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -f bench_old.txt bench_new.txt
	@echo "$(GREEN)Clean complete$(NC)"

# Installation
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/perf/cmd/benchstat@latest
	@echo "$(GREEN)Tools installed$(NC)"

# Composite targets
check: fmt vet test ## Run fmt, vet, and tests
	@echo "$(GREEN)All checks passed!$(NC)"

test-all: clean test-race test-edge coverage ## Run all tests with race detection and coverage
	@echo "$(GREEN)All tests complete!$(NC)"

ci: check test-race coverage ## Target for CI/CD pipelines
	@echo "$(GREEN)CI checks passed!$(NC)"

# Quick test targets
quick: ## Run quick smoke test
	@echo "$(BLUE)Running quick smoke test...$(NC)"
	@go test -short ./marraycrdt

test-move: ## Test all move-related operations
	@echo "$(BLUE)Testing move operations...$(NC)"
	@go test -v ./marraycrdt -run ".*Move.*"

test-swap: ## Test all swap operations
	@echo "$(BLUE)Testing swap operations...$(NC)"
	@go test -v ./marraycrdt -run ".*Swap.*"

test-concurrent: ## Test all concurrent operations
	@echo "$(BLUE)Testing concurrent operations...$(NC)"
	@go test -v ./marraycrdt -run ".*Concurrent.*"

# Debug targets
debug: ## Run tests with delve debugger (requires dlv)
	@if command -v dlv >/dev/null 2>&1; then \
		echo "$(BLUE)Starting debugger...$(NC)"; \
		dlv test ./marraycrdt; \
	else \
		echo "$(RED)Delve not installed. Install with: go install github.com/go-delve/delve/cmd/dlv@latest$(NC)"; \
	fi

# Performance analysis
profile-cpu: ## Generate CPU profile
	@echo "$(BLUE)Generating CPU profile...$(NC)"
	@go test -cpuprofile=cpu.prof -bench=. ./marraycrdt
	@echo "$(GREEN)CPU profile saved to cpu.prof$(NC)"
	@echo "View with: go tool pprof cpu.prof"

profile-mem: ## Generate memory profile
	@echo "$(BLUE)Generating memory profile...$(NC)"
	@go test -memprofile=mem.prof -bench=. ./marraycrdt
	@echo "$(GREEN)Memory profile saved to mem.prof$(NC)"
	@echo "View with: go tool pprof mem.prof"