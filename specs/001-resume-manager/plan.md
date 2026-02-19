# Implementation Plan: Resume Manager

**Branch**: `001-resume-manager` | **Date**: 2026-02-19 | **Spec**: [specs/001-resume-manager/spec.md](spec.md)
**Input**: Feature specification from `/specs/001-resume-manager/spec.md`

## Summary

Build a desktop resume management application using Go + Wails v2 that
provides a single source of truth for work history, skills, academic
credentials, certifications, professional summaries, and role descriptors.
The application stores all data locally in SQLite, generates ATS-compatible
PDF resumes from built-in templates with per-export content selection, and
tracks job applications with status history and resume snapshots. Data is
portable via JSON backup/restore and the data directory is user-configurable
for cloud drive placement.

## Technical Context

**Language/Version**: Go 1.26 (latest stable)
**Primary Dependencies**: Wails v2.11.0 (desktop framework), signintech/gopdf (PDF generation), modernc.org/sqlite (pure-Go SQLite driver)
**Storage**: SQLite via modernc.org/sqlite (pure-Go, no CGO, database/sql compatible)
**Testing**: go test (standard library), testify for assertions
**Target Platform**: macOS (primary); Windows and Linux (future)
**Project Type**: Desktop application (Wails: Go backend + web frontend)
**Performance Goals**: PDF generation <5s for a full resume; UI interactions <100ms perceived latency
**Constraints**: Offline-capable (no network required), single-user, local data only, <100MB memory typical usage
**Scale/Scope**: Single user, ~12-15 screens/views, ~23 entities (including junction tables), 50 functional requirements

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Status | Evidence |
|---|-----------|--------|----------|
| I | Simplicity First | PASS | No speculative abstractions. Direct SQLite access via database/sql. Single Go binary with embedded frontend. No repository pattern or ORM — plain SQL queries behind service interfaces. |
| II | Test-First Development | PASS | Plan calls for go test with testify assertions. Domain logic separated from Wails runtime for independent testability. PDF output validation tests planned. |
| III | Clean Architecture | PASS | Three layers: domain (entities, validation), service (business logic, interfaces), infrastructure (SQLite, PDF, Wails bindings). Domain has zero framework dependencies. Wails frontend communicates exclusively through binding API. |
| IV | Configurability | PASS | Competence levels, date formats, sort orders stored as data. Resume templates are data-driven. Colors/fonts in centralized theme config. Data directory location configurable. |
| V | Observability | PASS | log/slog (stdlib) for structured logging. Errors wrapped with context. PDF generation failures identify the specific section/element. No silent error swallowing. |
| VI | Code Quality | PASS | gofmt + go vet + golangci-lint enforced. Frontend: ESLint + Prettier with eslint-plugin-svelte. Exported types documented. 50/80 line function limits observed. |
| VII | Versioning & Breaking Changes | PASS | SemVer for application. SQLite schema migrations for data model changes. Template versions independent of app version. Wails binding changes treated as breaking. |

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
main.go                    # Wails entry point: wails.Run() configuration
app.go                     # App struct with lifecycle hooks, thin binding methods
wails.json                 # Wails project config (frontend build commands, app metadata)
go.mod / go.sum            # Go module definition

internal/
├── domain/                # Domain layer: entities, value objects, validation
│   ├── entities.go        # Core entity structs (WorkHistory, Skill, etc.)
│   ├── validation.go      # Business rule validation (date ranges, required fields)
│   └── interfaces.go      # Store/renderer interfaces (defined here per Constitution III, implemented in infra)
├── service/               # Service layer: business logic
│   ├── workhistory.go     # Work history CRUD + business logic
│   ├── skills.go          # Skills CRUD + competence sorting
│   ├── resume.go          # Resume assembly + export orchestration
│   ├── applications.go    # Job application tracking + status transitions
│   └── backup.go          # JSON export/import, rolling backups
├── infra/                 # Infrastructure layer: external concerns
│   ├── sqlite/            # SQLite persistence (implements domain interfaces)
│   │   ├── store.go       # Connection management
│   │   ├── migrations.go  # Schema migrations via PRAGMA user_version
│   │   ├── work_history.go    # Work history query implementations
│   │   ├── skills.go          # Skill query implementations
│   │   ├── skill_category.go  # Skill category query implementations
│   │   ├── academic.go        # Academic credential query implementations
│   │   ├── certifications.go  # Certification query implementations
│   │   ├── summaries.go       # Professional summary query implementations
│   │   ├── descriptors.go     # Role descriptor query implementations
│   │   ├── profile.go         # User profile query implementations
│   │   ├── profile_links.go   # Profile link query implementations
│   │   ├── lenses.go          # Lens query implementations
│   │   ├── skill_lens_tags.go # Skill lens tag query implementations
│   │   ├── export.go          # Resume export query implementations
│   │   ├── cover_letters.go   # Cover letter query implementations
│   │   └── applications.go    # Job application query implementations
│   └── pdf/               # PDF generation (implements renderer interface)
│       ├── renderer.go    # gopdf-based PDF rendering
│       └── templates/     # Data-driven template definitions
└── config/                # App configuration (data dir, backup count, etc.)
    └── config.go

build/
├── appicon.png
└── darwin/                # macOS build resources (Info.plist)

frontend/                  # Web frontend (Wails webview)
├── src/
│   ├── components/        # Reusable UI components
│   ├── pages/             # Page-level views (WorkHistory, Skills, Export, etc.)
│   ├── services/          # Frontend API calls to Wails bindings
│   └── stores/            # Frontend state management
├── package.json
└── index.html

tests/                     # Integration and contract tests
├── integration/           # Tests requiring SQLite + multiple services
└── fixtures/              # Test data (sample resumes, import files)
```

**Structure Decision**: Wails v2 desktop application with Go backend using
`internal/` package layout (domain → service → infra layering) and a web
frontend in `frontend/`. Unit tests co-located with packages (`_test.go`
files alongside source). Integration tests in a top-level `tests/` directory.

## Complexity Tracking

> No Constitution Check violations. All seven principles pass.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none) | — | — |

## Constitution Re-Check (Post-Design)

*Re-evaluated after Phase 1 design artifacts were produced.*

| # | Principle | Status | Post-Design Notes |
|---|-----------|--------|-------------------|
| I | Simplicity First | PASS | 28 tables total, all directly mapped to spec entities. No ORM, no repository pattern, no dependency injection framework. Export selection tables and lens selection tables are junction tables (simplest relational approach). Lens concept adds tables but no speculative abstractions. |
| II | Test-First Development | PASS | Domain layer has zero external dependencies — fully unit-testable. Service interfaces enable mock-based testing. PDF output testable via text extraction. |
| III | Clean Architecture | PASS | Dependency direction confirmed: domain ← service ← infra. Interfaces defined in domain layer (internal/domain/interfaces.go), implemented in infra. Binding contracts in contracts/bindings.md show thin App methods delegating to services. No domain-to-infra references. |
| IV | Configurability | PASS | Competence levels defined as application constants (data-driven). Data directory configurable. Backup count configurable. Template selection at export time. |
| V | Observability | PASS | Structured logging via log/slog throughout. PDF failures traceable to specific template section + data element per research.md findings. |
| VI | Code Quality | PASS | Go toolchain (gofmt, go vet, golangci-lint) + frontend linting (ESLint + Prettier with eslint-plugin-svelte). Binding methods are intentionally thin (delegation only). |
| VII | Versioning & Breaking Changes | PASS | Schema migration via PRAGMA user_version. Export and lens selection tables preserve data snapshots. Binding API changes documented as breaking. Lens bindings are new additions, not breaking changes. |

## Generated Artifacts

| Artifact | Path | Description |
|----------|------|-------------|
| Plan | `specs/001-resume-manager/plan.md` | This file |
| Research | `specs/001-resume-manager/research.md` | Technology decisions and rationale |
| Data Model | `specs/001-resume-manager/data-model.md` | Entity definitions + SQLite schema |
| Contracts | `specs/001-resume-manager/contracts/bindings.md` | Wails binding API definitions |
| Quickstart | `specs/001-resume-manager/quickstart.md` | Development setup guide |
| Agent Context | `AGENTS.md` | Updated by update-agent-context.sh |
