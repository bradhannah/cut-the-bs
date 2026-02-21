# Quickstart: Template Builder Feature

**Phase 1 output** | **Branch**: `002-template-builder` | **Date**: 2026-02-21

## Prerequisites

```bash
# Verify you're on the correct branch
git branch --show-current  # should output: 002-template-builder

# Ensure all dependencies are installed
make setup
```

## Build & Run

```bash
# Start development mode (Go backend + Svelte frontend with hot reload)
make dev

# Build production macOS app
make build
```

## Testing

```bash
# Run all Go tests
make test

# Run with race detector (matches CI)
make test-race

# Run specific test file or package
go test -v ./internal/infra/sqlite/ -run TestTemplate
go test -v ./internal/service/ -run TestTemplate
go test -v ./tests/integration/ -run TestTemplate
```

## Quality Checks

```bash
# Full CI pipeline (mirrors GitHub Actions)
make ci

# Individual checks
make vet            # go vet
make lint           # golangci-lint + frontend ESLint
make fmt-check      # gofmt + Prettier check
make frontend-check # svelte-check TypeScript
```

## Frontend Development

```bash
# Install new dependencies
cd frontend && bun add svelte-dnd-action

# Frontend build/check
make frontend-build
make frontend-check
```

## Key File Locations

### New Files to Create

| File | Purpose |
|------|---------|
| `internal/domain/template.go` | DocumentTemplate, TemplateElement types and constants |
| `internal/service/template.go` | Template CRUD, validation, duplication logic |
| `internal/service/template_test.go` | Template service unit tests |
| `internal/infra/sqlite/templates.go` | Template store implementation |
| `internal/infra/sqlite/templates_test.go` | Template store tests |
| `internal/infra/pdf/elements.go` | Per-element-type rendering dispatch |
| `internal/infra/pdf/builtin.go` | Built-in template configurations as data |
| `frontend/src/pages/TemplateBuilder.svelte` | Builder page (palette + canvas + properties) |
| `frontend/src/pages/TemplateList.svelte` | Template management list |
| `frontend/src/components/template/*.svelte` | Canvas, Palette, Properties, ElementBlock, LoopContainer |
| `frontend/src/stores/templateBuilder.ts` | Svelte writable stores for builder state |
| `tests/integration/template_test.go` | End-to-end template workflows |

### Files to Modify

| File | Change |
|------|--------|
| `internal/domain/entities.go` | Add ExportData.Templates field |
| `internal/domain/interfaces.go` | Add template Store methods, update RenderResumeRequest |
| `internal/infra/sqlite/migrations.go` | Add migrateV7 (template tables + built-in seeds) |
| `internal/infra/sqlite/data_management.go` | Add template export/import |
| `internal/infra/pdf/renderer.go` | Template-driven rendering pipeline |
| `internal/service/resume.go` | Use template from DB instead of string lookup |
| `app.go` | Add template Wails bindings |
| `main.go` | Increase window size to 1440x900 |
| `frontend/src/services/api.ts` | Add template API functions |
| `frontend/src/App.svelte` | Add template routes |
| `frontend/src/pages/Export.svelte` | Template selection from DB |

## Architecture Overview

```
Frontend (Svelte)                    Wails Bindings (app.go)
  TemplateBuilder.svelte ──────────→  CreateDocumentTemplate()
  TemplateList.svelte    ──────────→  ListDocumentTemplates()
  Canvas + Properties    ──────────→  CreateTemplateElement()
  Export.svelte          ──────────→  ExportResume() (modified)

                              │
                              ▼
                    Service Layer (template.go)
                    - Validation
                    - Duplication logic
                    - Template-to-render assembly
                              │
                              ▼
              ┌───────────────┴───────────────┐
              ▼                               ▼
     Store (sqlite/templates.go)    Renderer (pdf/renderer.go)
     - CRUD for templates           - Template-driven dispatch
     - Element ordering              - Per-element render functions
     - Built-in seeds (v7)          - Built-in config reproduction
```

## Implementation Order (Recommended)

1. **Domain types** -- Add new entities, constants, and interface methods
2. **Migration v7** -- Create tables, seed built-in templates
3. **Store implementation** -- Template CRUD following existing patterns
4. **Store tests** -- Test all CRUD operations
5. **Template service** -- Validation, duplication, template assembly
6. **Service tests** -- Test validation rules, duplication
7. **Renderer refactor** -- Replace template func map with element dispatch
8. **Built-in template configs** -- Reproduce Professional/Modern exactly
9. **Renderer tests** -- PDF output fidelity (compare to existing)
10. **Wails bindings** -- Wire service to app.go
11. **Frontend: API layer** -- Add template functions to api.ts
12. **Frontend: Template list page** -- Basic CRUD UI
13. **Frontend: Builder page** -- Three-panel layout with drag-and-drop
14. **Frontend: Properties panel** -- Context-sensitive element editing
15. **Integration tests** -- End-to-end template workflows
16. **Cover letter templates** -- Cover letter element types + built-in templates
17. **Export/import** -- Template portability (standalone + backup)
18. **Window size** -- Increase to 1440x900
