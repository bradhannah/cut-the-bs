# cut-the-bs Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-19

## Active Technologies
- Go 1.24.1 (go.mod specifies 1.24.1) + Wails v2.11.0 (desktop framework), signintech/gopdf (PDF generation), modernc.org/sqlite (pure-Go SQLite driver), Svelte 3 + TypeScript + Vite 3 (frontend), Bun (frontend package manager/runtime) (002-template-builder)
- Local SQLite database (WAL mode, single connection, pure-Go driver — no CGO) (002-template-builder)

- Go 1.24+ + Wails v2.11.0 (desktop framework), signintech/gopdf (PDF generation), modernc.org/sqlite (pure-Go SQLite driver), Bun (frontend JS runtime/package manager), Svelte + TypeScript (frontend UI) (001-resume-manager)

## Project Structure

```text
backend/
frontend/
tests/
```

## Commands

# Add commands for Go 1.26 (latest stable)

## Code Style

Go 1.26 (latest stable): Follow standard conventions

## Recent Changes
- 002-template-builder: Added Go 1.24.1 (go.mod specifies 1.24.1) + Wails v2.11.0 (desktop framework), signintech/gopdf (PDF generation), modernc.org/sqlite (pure-Go SQLite driver), Svelte 3 + TypeScript + Vite 3 (frontend), Bun (frontend package manager/runtime)

- 001-resume-manager: Added Go 1.26 (latest stable) + Wails v2.11.0 (desktop framework), signintech/gopdf (PDF generation), modernc.org/sqlite (pure-Go SQLite driver)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
