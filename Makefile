# Cut the BS Makefile
# Build automation for Go + Wails v2 + Svelte + Bun development workflow

.PHONY: help dev build build-debug release-macos clean test test-race test-verbose test-ats vet lint lint-go lint-frontend fmt fmt-go fmt-frontend fmt-check fmt-frontend-check frontend-build frontend-check frontend-install install setup doctor ci kill-dev atscheck atscheck-resume atscheck-cover-letter

# Bun binary - check common locations
BUN := $(shell command -v bun 2>/dev/null || echo ~/.bun/bin/bun)

# Wails binary
WAILS := $(shell command -v wails 2>/dev/null || echo ~/go/bin/wails)

# Release artifact path used by Homebrew cask workflow
RELEASE_ARCHIVE := dist/cut-the-bs-macos-universal.tar.gz

# Default target
help: ## Show this help message
	@echo "Cut the BS Makefile - Build Automation"
	@echo ""
	@echo "Getting Started:"
	@echo "  make setup        Install all dependencies and verify toolchain"
	@echo "  make dev          Start Wails dev mode with hot reload"
	@echo ""
	@echo "Development:"
	@echo "  make dev          Start Wails dev mode (Go + Svelte hot reload)"
	@echo "  make build        Build production macOS application"
	@echo "  make build-debug  Build with debug symbols"
	@echo "  make release-macos Build universal macOS archive + checksum"
	@echo "  make clean        Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test         Run all Go tests"
	@echo "  make test-race    Run Go tests with race detector"
	@echo "  make test-verbose Run Go tests with verbose output"
	@echo "  make test-ats     Run ATS-focused integration tests"
	@echo ""
	@echo "Quality:"
	@echo "  make lint         Run all linters (Go + frontend)"
	@echo "  make lint-go      Run golangci-lint on Go code"
	@echo "  make lint-frontend Run ESLint on Svelte/TS code"
	@echo "  make fmt          Format all code (Go + frontend)"
	@echo "  make fmt-check    Check formatting without modifying files"
	@echo ""
	@echo "Frontend:"
	@echo "  make frontend-build    Build frontend for production"
	@echo "  make frontend-check    Run svelte-check (TypeScript)"
	@echo "  make frontend-install  Install frontend dependencies"
	@echo ""
	@echo "CI:"
	@echo "  make ci           Run full CI pipeline locally"
	@echo ""
	@echo "Utility:"
	@echo "  make doctor       Run wails doctor to check dependencies"
	@echo "  make install      Install Go + frontend dependencies"
	@echo "  make kill-dev     Terminate stray dev processes"
	@echo "  make atscheck PDF=/path/file.pdf           Run ATS checker"
	@echo "  make atscheck-resume PDF=/path/file.pdf    Resume ATS preset"
	@echo "  make atscheck-cover-letter PDF=/path/file.pdf Cover letter ATS preset"

# ============================================================
# Setup & Installation
# ============================================================

setup: install check-prereqs doctor ## Full first-time setup: install deps, verify toolchain
	@echo ""
	@echo "Setup complete. Run 'make dev' to start developing."

install: ## Install Go and frontend dependencies
	@echo "Installing Go dependencies..."
	@go mod tidy
	@echo "Installing frontend dependencies..."
	@cd frontend && $(BUN) install
	@echo "Dependencies installed."

frontend-install: ## Install frontend dependencies only
	@cd frontend && $(BUN) install

check-prereqs: ## Verify all required tools are installed
	@echo "Checking prerequisites..."
	@command -v go >/dev/null 2>&1 || (echo "ERROR: Go is not installed. Install from https://go.dev/dl/" && exit 1)
	@echo "  Go $$(go version | awk '{print $$3}' | sed 's/go//')"
	@(command -v $(WAILS) >/dev/null 2>&1) || (echo "ERROR: Wails is not installed. Run: go install github.com/wailsapp/wails/v2/cmd/wails@latest" && exit 1)
	@echo "  Wails $$($(WAILS) version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')"
	@(command -v bun >/dev/null 2>&1 || test -x "$(HOME)/.bun/bin/bun") || (echo "ERROR: Bun is not installed. Install from https://bun.sh" && exit 1)
	@echo "  Bun $$($(BUN) --version)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "  golangci-lint $$(golangci-lint --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)"; \
	else \
		echo "  golangci-lint (not installed - 'make lint-go' will fail)"; \
	fi
	@echo "All prerequisites OK."

# ============================================================
# Development
# ============================================================

dev: ## Start Wails dev mode (Go backend + Svelte frontend with hot reload)
	@echo "Starting Wails development mode..."
	@echo "  Backend:  Go with live rebuild"
	@echo "  Frontend: Vite dev server with HMR"
	@echo ""
	@$(WAILS) dev -forcebuild

# ============================================================
# Build
# ============================================================

build: ## Build production macOS application (output: build/bin/cut-the-bs.app)
	@echo "Building production application..."
	@$(WAILS) build
	@echo ""
	@echo "Build complete: build/bin/cut-the-bs.app"

build-debug: ## Build with debug symbols and dev tools
	@echo "Building debug application..."
	@$(WAILS) build -debug
	@echo ""
	@echo "Debug build complete: build/bin/cut-the-bs.app"

release-macos: ## Build universal macOS app archive for GitHub Release/Homebrew cask
	@echo "Building universal macOS application..."
	@$(WAILS) build -platform darwin/universal -clean
	@echo "Packaging app bundle..."
	@mkdir -p dist
	@tar -czf $(RELEASE_ARCHIVE) -C build/bin cut-the-bs.app
	@shasum -a 256 $(RELEASE_ARCHIVE) | tee $(RELEASE_ARCHIVE).sha256
	@echo ""
	@echo "Release artifacts created:"
	@echo "  $(RELEASE_ARCHIVE)"
	@echo "  $(RELEASE_ARCHIVE).sha256"

clean: ## Remove build artifacts and caches
	@echo "Removing build artifacts..."
	@rm -rf build/bin
	@rm -f cut-the-bs
	@rm -rf dist
	@rm -rf frontend/dist
	@rm -rf frontend/node_modules/.vite
	@echo "Clean complete."

# ============================================================
# Testing
# ============================================================

test: ## Run all Go tests
	@echo "Running Go tests..."
	@go test ./...

test-race: ## Run Go tests with race detector
	@echo "Running Go tests with race detector..."
	@go test -race ./...

test-verbose: ## Run Go tests with verbose output
	@echo "Running Go tests (verbose)..."
	@go test -v -count=1 ./...

test-ats: ## Run ATS-focused integration tests
	@echo "Running ATS integration tests..."
	@go test ./tests/integration -run ATS -count=1

# ============================================================
# Quality - Linting
# ============================================================

lint: lint-go lint-frontend ## Run all linters

lint-go: ## Run golangci-lint on Go code
	@echo "Linting Go code..."
	@golangci-lint run
	@echo "Go lint passed."

lint-frontend: ## Run ESLint on frontend code
	@echo "Linting frontend code..."
	@cd frontend && $(BUN) run lint
	@echo "Frontend lint passed."

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# ============================================================
# Quality - Formatting
# ============================================================

fmt: fmt-go fmt-frontend ## Format all code

fmt-go: ## Format Go code with gofmt
	@echo "Formatting Go code..."
	@gofmt -w .
	@echo "Go formatting complete."

fmt-frontend: ## Format frontend code with Prettier
	@echo "Formatting frontend code..."
	@cd frontend && $(BUN) run format
	@echo "Frontend formatting complete."

fmt-check: ## Check formatting without modifying files
	@echo "Checking Go formatting..."
	@test -z "$$(gofmt -l . 2>/dev/null | grep -v vendor)" || (echo "Go files need formatting:" && gofmt -l . && exit 1)
	@echo "Go formatting OK."
	@$(MAKE) fmt-frontend-check

fmt-frontend-check: ## Check frontend formatting without modifying
	@echo "Checking frontend formatting..."
	@cd frontend && $(BUN) run format:check
	@echo "Frontend formatting OK."

# ============================================================
# Frontend
# ============================================================

frontend-build: ## Build frontend for production
	@echo "Building frontend..."
	@cd frontend && $(BUN) run build

frontend-check: ## Run svelte-check (TypeScript type checking)
	@echo "Running svelte-check..."
	@cd frontend && $(BUN) run check

# ============================================================
# CI - Run the full pipeline locally
# ============================================================

ci: ## Run the full CI pipeline locally (mirrors GitHub Actions)
	@echo "=========================================="
	@echo "  Running CI pipeline locally"
	@echo "=========================================="
	@echo ""
	@echo "[1/7] Installing frontend dependencies..."
	@cd frontend && $(BUN) install --frozen-lockfile 2>/dev/null || cd frontend && $(BUN) install
	@echo ""
	@echo "[2/7] Building frontend..."
	@cd frontend && $(BUN) run build
	@echo ""
	@echo "[3/7] Running go vet..."
	@go vet ./...
	@echo ""
	@echo "[4/7] Building Go..."
	@go build ./...
	@echo ""
	@echo "[5/7] Running Go tests with race detector..."
	@go test -race ./...
	@echo ""
	@echo "[6/7] Linting frontend..."
	@cd frontend && $(BUN) run lint
	@echo ""
	@echo "[7/7] Checking frontend formatting..."
	@cd frontend && $(BUN) run format:check
	@echo ""
	@echo "=========================================="
	@echo "  CI pipeline passed"
	@echo "=========================================="

# ============================================================
# Utility
# ============================================================

doctor: ## Run wails doctor to check system dependencies
	@$(WAILS) doctor

atscheck: ## Run ATS checker against one PDF (set PDF=/path/file.pdf)
	@if [ -z "$(PDF)" ]; then \
		echo "ERROR: PDF is required. Usage: make atscheck PDF=\"/path/to/generated.pdf\""; \
		exit 1; \
	fi
	@go run ./cmd/atscheck --pdf "$(PDF)" $(ATS_ARGS)

atscheck-resume: ## Resume ATS preset (set PDF=/path/file.pdf)
	@if [ -z "$(PDF)" ]; then \
		echo "ERROR: PDF is required. Usage: make atscheck-resume PDF=\"/path/to/resume.pdf\""; \
		exit 1; \
	fi
	@go run ./cmd/atscheck \
		--pdf "$(PDF)" \
		--require "Experience" \
		--require "Skills" \
		--order "Experience>Skills" \
		$(ATS_ARGS)

atscheck-cover-letter: ## Cover-letter ATS preset (set PDF=/path/file.pdf)
	@if [ -z "$(PDF)" ]; then \
		echo "ERROR: PDF is required. Usage: make atscheck-cover-letter PDF=\"/path/to/cover-letter.pdf\""; \
		exit 1; \
	fi
	@go run ./cmd/atscheck \
		--pdf "$(PDF)" \
		--require "Dear" \
		--require "Sincerely" \
		--order "Dear>Sincerely" \
		$(ATS_ARGS)

kill-dev: ## Terminate stray development processes
	@echo "Terminating stray development processes..."
	@pkill -f "wails dev" 2>/dev/null || echo "  No wails dev processes"
	@pkill -f "vite" 2>/dev/null || echo "  No Vite processes"
	@echo "Done."
