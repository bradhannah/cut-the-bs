# Quickstart: Cut the BS - Resume Manager

**Date**: 2026-02-19 | **Branch**: `001-resume-manager`

## Prerequisites

- **Go 1.26+**: [go.dev/dl](https://go.dev/dl/)
- **Bun 1.0+**: JavaScript runtime and package manager for the Svelte frontend ([bun.sh](https://bun.sh/))
- **Wails CLI v2.11.0**: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0`
- **macOS**: Xcode Command Line Tools (`xcode-select --install`)

Verify installations:
```bash
go version          # go1.26 or later
bun --version       # 1.0 or later
wails doctor        # checks all dependencies
```

## Project Setup

```bash
# Clone and checkout feature branch
git clone <repo-url> cut-the-bs
cd cut-the-bs
git checkout 001-resume-manager

# Initialize Wails project (first time only)
wails init -n cut-the-bs -t svelte-ts

# Install frontend dependencies
cd frontend && bun install && cd ..

# Install Go dependencies
go mod tidy
```

## Development

```bash
# Start development mode (Go backend + Svelte frontend with hot reload)
wails dev

# The app opens as a native macOS window
# Frontend changes: hot reload (instant)
# Go changes: auto-rebuild + relaunch (~2-3s)
# Browser debug: http://localhost:34115
```

## Project Structure

```
main.go                    # Wails entry point
app.go                     # Binding methods (thin wrappers)
wails.json                 # Wails project config
internal/
├── domain/                # Entities, validation (zero dependencies)
├── service/               # Business logic, interfaces
├── infra/sqlite/          # Database persistence
├── infra/pdf/             # PDF generation
└── config/                # App configuration
frontend/src/              # Svelte + TypeScript UI
tests/                     # Integration tests
```

## Testing

```bash
# Run all Go tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run tests for a specific package
go test ./internal/domain/...
go test ./internal/service/...

# Run with verbose output
go test -v ./...
```

## Linting

```bash
# Format Go code
gofmt -w .

# Vet Go code
go vet ./...

# Run golangci-lint (install: https://golangci-lint.run/usage/install/)
golangci-lint run

# Frontend linting
cd frontend && bun run lint && cd ..
```

## Building

```bash
# Build production binary for macOS
wails build

# Output: build/bin/cut-the-bs.app (macOS application bundle)

# Build with debug info
wails build -debug
```

## Database

The SQLite database is stored in the user's configured data directory.
Default location:

- **macOS**: `~/Library/Application Support/cut-the-bs/data.db`

The database is created automatically on first launch with the schema
from `data-model.md`. Migrations run automatically via
`PRAGMA user_version`.

### Manual Database Inspection

```bash
# Open the database with sqlite3
sqlite3 ~/Library/Application\ Support/cut-the-bs/data.db

# Check schema version
PRAGMA user_version;

# List tables
.tables

# View schema
.schema work_history_entry
```

## Key Commands Reference

| Command | Description |
|---------|-------------|
| `wails dev` | Development mode with hot reload |
| `wails build` | Production build |
| `wails doctor` | Check development dependencies |
| `go test ./...` | Run all tests |
| `go test -race ./...` | Run tests with race detection |
| `golangci-lint run` | Lint Go code |
| `gofmt -w .` | Format Go code |
