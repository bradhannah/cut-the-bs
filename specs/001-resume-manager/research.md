# Research: Resume Manager

**Date**: 2026-02-19 | **Branch**: `001-resume-manager`

## Decision Summary

| Area | Decision | Key Alternative Rejected |
|------|----------|-------------------------|
| Desktop Framework | Wails v2.11.0 | Wails v3 (alpha, not stable) |
| PDF Generation | signintech/gopdf | go-pdf/fpdf (archived Mar 2025) |
| SQLite Driver | modernc.org/sqlite (pure Go) | mattn/go-sqlite3 (requires CGO) |
| Logging | log/slog (stdlib) | zerolog, zap (unnecessary deps) |
| Frontend Framework | Svelte + TypeScript | React (heavier for this scope) |
| Frontend Routing | svelte-spa-router | Conditional rendering (unwieldy at 12+ pages) |
| Frontend Linting | ESLint + Prettier + eslint-plugin-svelte | Biome (less Svelte support) |
| Schema Migrations | PRAGMA user_version | golang-migrate (extra dep) |
| Testing | go test + testify | — |

---

## 1. PDF Generation: signintech/gopdf

### Decision
Use `github.com/signintech/gopdf` for all PDF output (resumes and cover
letters).

### Rationale
- Actively maintained (2.9k stars, 84 releases, latest activity 2025)
- Pure Go, no CGO dependency
- TrueType font subsetting with `/ToUnicode` CMap — critical for ATS
  text extraction
- Built-in word-aware text wrapping (`SplitTextWithWordWrap`,
  `MultiCellWithOption` with `BreakModeIndicatorSensitive`)
- Table layout support (though manual positioning preferred for ATS)
- External hyperlink support (annotation-based, ATS-neutral)
- Multi-page support with header/footer callbacks

### ATS Compatibility Notes
- Each `Cell()` call emits a separate `BT...ET` block with absolute
  positioning. Text is extracted in the order it is written to the
  content stream.
- No PDF/UA tagged structure support — ATS relies on spatial layout
  for reading order. Mitigated by writing text in strict top-to-bottom,
  left-to-right order.
- Font subsetting embeds only used glyphs with proper Unicode mappings.
  No mid-word space artifacts when using well-designed fonts.
- Two-column layouts risk ATS misreading. Use single-column for body
  content; columns only for compact sections (contact info, skills
  grid).

### Key API Patterns
```go
// Font embedding (automatic subsetting)
pdf.AddTTFFont("sans", "fonts/LiberationSans-Regular.ttf")
pdf.AddTTFFontWithOption("sans", "fonts/LiberationSans-Bold.ttf",
    gopdf.TtfOption{Style: gopdf.Bold, UseKerning: true})

// Word-aware wrapping
lines, _ := pdf.SplitTextWithWordWrap(text, maxWidth)

// Multi-cell with word break
pdf.MultiCellWithOption(&gopdf.Rect{W: w, H: h}, text, gopdf.CellOption{
    Align: gopdf.Left | gopdf.Top,
    BreakOption: &gopdf.BreakOption{
        Mode:           gopdf.BreakModeIndicatorSensitive,
        BreakIndicator: ' ',
    },
})

// Hyperlinks (overlay annotation on rendered text)
pdf.AddExternalLink(url, x, y, width, height)
```

### Alternatives Considered
- **go-pdf/fpdf**: Archived March 2025. No longer maintained.
- **pdfcpu**: PDF manipulation library, not a generation library.
- **HTML-to-PDF via headless browser**: Heavy dependency, slow
  generation, harder to control ATS output quality. Rejected per
  Principle I (Simplicity First).

---

## 2. SQLite: modernc.org/sqlite

### Decision
Use `modernc.org/sqlite` as the database/sql driver for all data
persistence.

### Rationale
- Pure Go — no CGO, trivial cross-compilation for macOS/Windows/Linux
- Registered as `"sqlite"` driver, standard `database/sql` interface
- Competitive or superior performance to mattn/go-sqlite3 for
  concurrent reads and small-to-medium queries (114 vs 85 benchmark
  points)
- Full SQLite 3.51.2 feature support including backup API
- DSN supports inline pragma configuration

### Configuration
```go
dsn := fmt.Sprintf("file:%s?"+
    "_pragma=journal_mode(WAL)&"+
    "_pragma=busy_timeout(5000)&"+
    "_pragma=synchronous(NORMAL)&"+
    "_pragma=foreign_keys(1)",
    dbPath)
db, err := sql.Open("sqlite", dsn)
```

Key pragmas:
- `journal_mode(WAL)`: Readers don't block writers, persistent across
  restarts
- `busy_timeout(5000)`: Retry on SQLITE_BUSY instead of immediate fail
- `synchronous(NORMAL)`: Safe with WAL, skips fsync on commits
- `foreign_keys(1)`: Enforce referential integrity

### Migration Strategy
Use SQLite's built-in `PRAGMA user_version` for schema versioning:
```go
var version int
db.QueryRow("PRAGMA user_version").Scan(&version)
if version < 1 { /* create tables, set user_version = 1 */ }
if version < 2 { /* alter tables, set user_version = 2 */ }
```
Zero dependencies, atomic, idiomatic SQLite. Aligns with Principle I.

### Backup Strategy
- **Rolling backups**: Checkpoint WAL (`PRAGMA wal_checkpoint(TRUNCATE)`)
  then copy the `.db` file. Single-process desktop app makes this safe.
- **JSON export**: Query all tables, marshal to JSON. Independent of
  SQLite file format — portable across versions and platforms.
- **Autosave**: All writes go through transactions; SQLite handles
  durability. No explicit save action needed.

### Connection Pool
```go
db.SetMaxOpenConns(1)    // Single writer, simplest for desktop app
db.SetConnMaxLifetime(0) // Keep connections alive
db.SetConnMaxIdleTime(0)
```

### Alternatives Considered
- **mattn/go-sqlite3**: Requires CGO, complicates cross-compilation.
  Faster for bulk inserts and large BLOB scanning, but desktop app
  workload is dominated by small queries where modernc is faster.

---

## 3. Desktop Framework: Wails v2.11.0

### Decision
Use Wails v2.11.0 (latest stable) for the desktop application shell.

### Rationale
- Stable release (v3 is alpha only, v3.0.0-alpha.72)
- Go backend with web frontend in native webview (webkit on macOS)
- Public Go methods on bound structs are auto-exposed as JavaScript
  Promises with auto-generated TypeScript models
- Embedded assets via Go `embed.FS` — single binary distribution
- Built-in dev workflow with hot reload (`wails dev`)

### Binding Architecture
```go
// Go: public methods on bound structs become JS functions
type App struct { ctx context.Context }
func (a *App) GetWorkHistory() ([]WorkHistoryEntry, error) { ... }

// Generated TypeScript:
// import { GetWorkHistory } from "../wailsjs/go/main/App"
// GetWorkHistory().then(entries => ...)
```

- Structs with `json` tags auto-generate TypeScript classes in
  `wailsjs/go/models.ts`
- Errors returned from Go reject the JS Promise
- Events bridge Go↔JS: `runtime.EventsEmit(ctx, "event", data)`

### Frontend: Svelte + TypeScript
Svelte chosen over React for this project:
- Lighter bundle, less boilerplate for a form-heavy app
- Built-in reactivity without hooks/state management library
- First-class Wails template support
- Sufficient component ecosystem for data entry forms and lists

### Static Asset Embedding
Fonts and PDF template data embedded via Go's `embed` package:
```go
//go:embed assets/fonts/*
var fonts embed.FS
```
Available to Go code directly for PDF generation. No external files
needed at runtime.

### macOS Configuration
```go
wails.Run(&options.App{
    Title:    "Cut the BS",
    Width:    1200, Height: 800,
    MinWidth: 800, MinHeight: 600,
    Mac: &mac.Options{
        TitleBar:   mac.TitleBarDefault(),
        Appearance: mac.DefaultAppearance,
    },
})
```

### Lifecycle Hooks
- `OnStartup`: Initialize services, open database
- `OnDomReady`: Frontend loaded, safe to emit events
- `OnBeforeClose`: Checkpoint database, ensure backups
- `OnShutdown`: Close database connection

---

## 4. Logging: log/slog

### Decision
Use Go standard library `log/slog` for all structured logging.

### Rationale
- Standard library since Go 1.21 — no external dependency
- Structured key-value logging with JSON and text handlers
- Satisfies Constitution Principle V (Observability): timestamp,
  severity, component, message
- Sufficient for a desktop application's diagnostic needs

### Usage Pattern
```go
logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

slog.Info("PDF generated",
    "template", templateName,
    "entries", len(selectedEntries),
    "duration_ms", elapsed.Milliseconds())

slog.Error("PDF generation failed",
    "template", templateName,
    "section", "work_history",
    "entry_id", entryID,
    "error", err)
```

### Alternatives Considered
- **zerolog**: Zero-allocation, but unnecessary for a desktop app with
  low log volume.
- **zap**: Large ecosystem, but adds dependency for no benefit here.

---

## 5. Frontend Framework: Svelte + TypeScript

### Decision
Use Svelte with TypeScript for the Wails frontend.

### Rationale
- Lightweight compiled output (no virtual DOM runtime)
- Built-in reactivity eliminates need for state management library
- Form-heavy data entry is Svelte's strength (two-way binding)
- First-class Wails v2 template (`wails init -n app -t svelte-ts`)
- TypeScript for type safety with auto-generated Wails models

### Alternatives Considered
- **React**: Heavier runtime, requires state management library (Redux,
  Zustand). Better ecosystem for complex UIs, but overkill for this
  app's scope (~10-15 screens of forms and lists).
- **Vue**: Viable alternative, but Svelte's compile-time approach
  produces smaller bundles and the reactivity model is simpler.
- **Vanilla TS**: Too much boilerplate for a multi-screen app.

---

## 6. Schema Migration: PRAGMA user_version

### Decision
Use SQLite's built-in `PRAGMA user_version` for schema versioning and
migrations.

### Rationale
- Zero external dependencies
- Atomic — version is part of the database file header
- Simple sequential migration pattern (if version < N, apply migration)
- Idiomatic SQLite practice for embedded databases
- Aligns with Principle I (Simplicity First)

### Alternatives Considered
- **golang-migrate/migrate**: Full migration framework with SQL files.
  Overkill for a single-user desktop app with a single database.
- **pressly/goose**: Similar to golang-migrate. Adds dependency for a
  simple use case.

---

## Unresolved Items

None. All NEEDS CLARIFICATION items were resolved during the spec
clarification phase.

---

## 7. Frontend Routing: svelte-spa-router

### Decision
Use `svelte-spa-router` for frontend page navigation within the Wails
webview.

### Rationale
- Lightweight hash-based routing, well-suited for Wails desktop apps
  (no server-side routing, no URL rewriting needed)
- Simple API: `<Router>` component with route definitions
- Supports named routes, route parameters, and navigation guards
- Most popular Svelte routing library for SPA contexts
- No SSR complexity — purely client-side, matching the Wails model

### Usage Pattern
```svelte
<script>
  import Router from 'svelte-spa-router'
  const routes = {
    '/': WorkHistory,
    '/skills': Skills,
    '/education': Education,
    '/summaries': Summaries,
    '/descriptors': Descriptors,
    '/export': Export,
    '/cover-letters': CoverLetters,
    '/applications': Applications,
    '/lenses': Lenses,
    '/settings': Settings,
  }
</script>
<Router {routes} />
```

### Alternatives Considered
- **Conditional rendering ({#if})**: Simplest, but becomes unwieldy
  with 12+ pages. No browser history support, no deep linking.
- **svelte-routing**: Similar capability but less maintained and
  designed for server-rendered contexts.

---

## 8. Frontend Linting: ESLint + Prettier

### Decision
Use ESLint with `eslint-plugin-svelte` and Prettier for frontend code
quality enforcement.

### Rationale
- Most popular and best-supported linting setup for Svelte+TypeScript
- `eslint-plugin-svelte` provides Svelte-specific rules (a11y,
  component structure, reactive statement validation)
- Prettier handles formatting consistency (single source of truth for
  style)
- Aligns with Constitution Principle VI (Code Quality): zero warnings
  enforced

### Configuration
- ESLint: `@typescript-eslint/parser` + `eslint-plugin-svelte` +
  `eslint-config-prettier`
- Prettier: `prettier-plugin-svelte` for `.svelte` file formatting
- Scripts: `bun run lint` and `bun run format` in `frontend/package.json`

### Alternatives Considered
- **Biome**: Faster all-in-one linter/formatter, but less mature Svelte
  support. May be revisited when Svelte support reaches parity.
