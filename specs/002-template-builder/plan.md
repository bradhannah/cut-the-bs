# Implementation Plan: Template Builder

**Branch**: `002-template-builder` | **Date**: 2026-02-21 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-template-builder/spec.md`

## Summary

Replace the current hardcoded PDF resume templates with a data-driven template builder. Users visually compose resume and cover letter templates by dragging elements onto a canvas, configuring formatting properties, and defining loop containers for repeating data. Two built-in resume templates ("Professional", "Modern") and two built-in cover letter templates ("Formal", "Casual") ship as default configurations that reproduce current output exactly. Templates are persisted in SQLite and rendered via the existing gopdf pipeline.

## Technical Context

**Language/Version**: Go 1.24.1 (go.mod specifies 1.24.1)
**Primary Dependencies**: Wails v2.11.0 (desktop framework), signintech/gopdf (PDF generation), modernc.org/sqlite (pure-Go SQLite driver), Svelte 3 + TypeScript + Vite 3 (frontend), Bun (frontend package manager/runtime)
**Storage**: Local SQLite database (WAL mode, single connection, pure-Go driver — no CGO)
**Testing**: `go test -race ./...` with real SQLite databases in `t.TempDir()`; 27 test files across domain/infra/service/integration layers; CI on macos-latest via GitHub Actions; integration tests include ATS PDF text extraction validation and performance benchmarks
**Target Platform**: macOS (primary); Windows and Linux (future)
**Project Type**: Desktop app (Go backend + web frontend via Wails webview)
**Performance Goals**: PDF generation < 5 seconds for typical resume (5 work entries, 20 skills, 3 education); DB operations < 100ms; Preview PDF within 5 seconds (SC-006)
**Constraints**: Offline-capable (all local storage); no CGO; ATS-compatible PDF output (no mid-word spaces, correct text extraction order); existing Liberation Sans font family only
**Scale/Scope**: Single-user desktop app; ~31 existing DB tables; 12 frontend pages; 2 existing PDF templates to reproduce exactly

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Simplicity First | PASS | Feature replaces hardcoded templates with data-driven system — justified by spec FR-003 (reproduce existing) + FR-001 (user-created). Single-column vertical layout keeps builder simple. No speculative abstractions. |
| II. Test-First Development | PASS (gate) | TDD required for all new code. Test strategy: unit tests for template entity validation, service-layer tests for template CRUD + rendering, integration tests for PDF output fidelity (compare built-in template output to current hardcoded output), ATS tests for user-created templates. |
| III. Clean Architecture | PASS | Template entity + interfaces in domain layer; template storage in infra/sqlite; template rendering in infra/pdf; template service in service layer; Wails bindings in app.go. No cross-layer violations. |
| IV. Configurability | PASS | This feature IS the configurability principle applied to templates. Replaces hardcoded layouts with user-configurable data-driven definitions. Per-template margins, font sizes, spacing, bullet chars, separator chars. |
| V. Observability | PASS (gate) | Rendering errors must identify which element and data field caused failure. Template validation errors must be user-facing and actionable. |
| VI. Code Quality | PASS (gate) | Must pass gofmt, go vet, golangci-lint. Functions < 80 lines. Exported types documented. |
| VII. Versioning & Breaking Changes | PASS | New DB migration (v7) with migration path. Template format independently versioned per constitution requirement. Existing export records preserved (FR-050). Wails binding API: additive only (new methods for template operations). |

**Gate Result**: All principles PASS. Two principles (II, V) noted as gates requiring active enforcement during implementation but no blockers.

**Post-Design Re-evaluation** (after Phase 0 research + Phase 1 design artifacts):

| Principle | Status | Post-Design Notes |
|-----------|--------|-------------------|
| I. Simplicity First | PASS | Design avoids speculative abstractions: single-level nesting only (validated in Go), JSON config blob instead of over-normalized columns, `svelte-dnd-action` instead of custom DnD. Svelte stores are a justified exception to component-local pattern (shared state across 3 panels + nested containers). |
| II. Test-First Development | PASS (gate) | No change. data-model.md defines validation rules testable with table-driven tests. 14 Wails bindings in contracts provide clear API surface for integration tests. |
| III. Clean Architecture | PASS | data-model.md defines domain entities/interfaces cleanly. Wails bindings contract keeps frontend-backend boundary explicit. Template rendering stays in infra/pdf. No cross-layer violations introduced. |
| IV. Configurability | PASS | 23+ element types with per-element JSON config. Per-template margins, font sizes, spacing, separators. Built-in templates seeded via migration, editable by user. |
| V. Observability | PASS (gate) | No change. Element-level error identification requirement preserved in design. |
| VI. Code Quality | PASS (gate) | No change. Element rendering dispatch in `elements.go` keeps functions focused. |
| VII. Versioning & Breaking Changes | PASS | Migration v7 schema designed. `ExportRequest.TemplateID` changes from `string` to `int64` (breaking change to Wails binding) — mitigated by being on feature branch and additive in practice (old templates will be seeded as built-in). Template format version field included in `document_template` table. |

**Post-Design Gate Result**: All principles continue to PASS. No new concerns from detailed design.

## Project Structure

### Documentation (this feature)

```text
specs/002-template-builder/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Backend (Go)
internal/
├── domain/
│   ├── entities.go          # + Template, TemplateElement, TemplateVariable types
│   ├── interfaces.go        # + TemplateStore interface methods
│   └── constants.go         # + element type, style, alignment enums
├── service/
│   ├── template.go          # NEW: template CRUD, duplication, validation
│   ├── template_test.go     # NEW: template service tests
│   ├── resume.go            # MODIFY: use template definition for rendering
│   └── export.go            # MODIFY: template selection, portability
├── infra/
│   ├── sqlite/
│   │   ├── migrations.go    # MODIFY: add v7 migration for template tables
│   │   ├── template.go      # NEW: TemplateStore implementation
│   │   ├── template_test.go # NEW: template store tests
│   │   └── export.go        # MODIFY: include templates in backup
│   └── pdf/
│       ├── renderer.go      # MODIFY: template-driven rendering pipeline
│       ├── elements.go      # NEW: element rendering dispatch
│       └── builtin.go       # NEW: built-in template configurations (replaces template_professional.go, template_modern.go)
app.go                       # MODIFY: add template Wails bindings
main.go                      # MODIFY: increase window size

# Frontend (Svelte + TypeScript)
frontend/src/
├── pages/
│   ├── TemplateBuilder.svelte  # NEW: main builder page (palette + canvas + properties)
│   ├── TemplateList.svelte     # NEW: template management list
│   └── Export.svelte           # MODIFY: template selection from DB
├── components/
│   ├── template/
│   │   ├── Canvas.svelte       # NEW: drag-and-drop canvas
│   │   ├── Palette.svelte      # NEW: element palette
│   │   ├── Properties.svelte   # NEW: element properties panel
│   │   ├── ElementBlock.svelte # NEW: canvas element representation
│   │   └── LoopContainer.svelte# NEW: loop container with sub-elements
│   └── coverletter/
│       └── PromptDialog.svelte # NEW: guided prompt input dialog
├── services/
│   └── api.ts               # MODIFY: add template API calls

# Tests
tests/integration/
├── template_test.go         # NEW: end-to-end template builder workflows
└── ats_test.go              # MODIFY: add user-created template ATS tests
```

**Structure Decision**: Follows the existing project structure (Go backend with `internal/` layers + Svelte frontend). No new top-level directories needed. Template code integrates into each existing layer following clean architecture. The `template_professional.go` and `template_modern.go` files will be replaced by data-driven built-in configurations in `builtin.go`.

## Complexity Tracking

> No constitution violations to justify. All principles pass.
