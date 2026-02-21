<!--
  Sync Impact Report
  ==================
  Version change: N/A (initial) → 1.0.0
  Modified principles: N/A (first population)
  Added sections:
    - 7 Core Principles:
      I.   Simplicity First
      II.  Test-First Development
      III. Clean Architecture
      IV.  Configurability
      V.   Observability
      VI.  Code Quality
      VII. Versioning & Breaking Changes
    - Technology Constraints & Quality Standards
    - Development Workflow & Release Process
    - Governance
  Removed sections: None
  Templates requiring updates:
    - .specify/templates/plan-template.md ✅ no changes needed (generic)
    - .specify/templates/spec-template.md ✅ no changes needed (generic)
    - .specify/templates/tasks-template.md ✅ no changes needed (generic)
    - .specify/templates/checklist-template.md ✅ no changes needed (generic)
    - .specify/templates/agent-file-template.md ✅ no changes needed (generic)
    - .opencode/command/speckit.constitution.md ⚠ stale path reference
      on line 49: `.specify/templates/commands/*.md` should be
      `.opencode/command/*.md` (does not block constitution correctness)
  Follow-up TODOs: None
-->

# cut-the-bs Constitution

## Core Principles

### I. Simplicity First

Every feature and abstraction MUST justify its existence against
immediate, demonstrable need. Speculative generality is prohibited.

- New code MUST solve a current requirement, not a hypothetical future
  one. If a capability is not needed now, it MUST NOT be built now.
- When two approaches satisfy the same requirement, the simpler one
  MUST be chosen unless a measurable, documented trade-off justifies
  the more complex option.
- Abstractions MUST be introduced only when duplication or complexity
  has been observed in practice, never preemptively.
- YAGNI violations MUST be flagged during code review and either
  removed or justified with a linked specification requirement.

**Rationale**: cut-the-bs is a desktop application whose primary value
is straightforward data entry and PDF export. Over-engineering the
architecture early delays shipping usable software and creates
maintenance burden disproportionate to user benefit.

### II. Test-First Development

All production code MUST be preceded by a failing test that defines the
expected behavior. The Red-Green-Refactor cycle is non-negotiable.

- Tests MUST be written and confirmed to fail before any corresponding
  implementation code is written.
- Each test MUST verify one logical behavior. Tests that assert
  multiple unrelated outcomes MUST be split.
- Test names MUST describe the scenario and expected result
  (e.g., `TestExportPDF_MissingWorkHistory_ReturnsError`).
- Refactoring steps MUST NOT change observable behavior; the existing
  test suite MUST continue to pass without modification.
- Test coverage MUST NOT be used as a sole quality metric. Coverage
  without meaningful assertions is prohibited.

**Rationale**: TDD ensures that every feature is verifiable from the
moment it exists, prevents regressions as the resume data model grows,
and produces a living specification of system behavior.

### III. Clean Architecture

The system MUST maintain strict separation of concerns with explicit
dependency boundaries. Inner layers MUST NOT depend on outer layers.

- The domain layer (entities, value objects, business rules) MUST have
  zero dependencies on infrastructure, UI, or framework code.
- Infrastructure concerns (PDF generation, file I/O, persistence) MUST
  be accessed through interfaces defined in the domain layer.
- The Wails frontend MUST communicate with the Go backend exclusively
  through a defined binding API. Direct data store access from the
  frontend is prohibited.
- Each architectural layer MUST be independently testable without
  requiring the layers above or below it.
- Dependency direction MUST always point inward: UI -> Application ->
  Domain. Violations MUST be flagged and corrected.

**Rationale**: Resume data, PDF rendering, and UI presentation are
distinct concerns that will evolve at different rates. Clean boundaries
allow swapping PDF engines, adding new export formats, or replacing
the UI framework without rewriting business logic.

### IV. Configurability

All literal values that represent user-facing data, visual styling, or
behavioral parameters MUST be abstracted into named constants, theme
definitions, or configuration structures.

- Magic numbers (font sizes, margins, spacing, limits) MUST be defined
  as named constants or configuration fields, never inline literals.
- Colors MUST be defined in a centralized theme or palette structure.
  Hex/RGB literals in layout or rendering code are prohibited.
- Resume section labels, default text, and format strings MUST be
  externalized so they can be localized or customized without code
  changes.
- Template layouts (PDF resume designs) MUST be data-driven. Adding a
  new resume template MUST NOT require changes to rendering logic.
- Competence levels, date format preferences, and sort orders MUST be
  user-configurable and persisted as part of the user's profile data.

**Rationale**: The core value proposition of cut-the-bs is "enter once,
use many." Users MUST be able to tailor output to different job
applications without modifying source code. Hard-coded values directly
undermine this goal.

### V. Observability

The application MUST produce structured, actionable diagnostic output
that enables rapid identification of problems without requiring a
debugger.

- All errors MUST be wrapped with context describing the operation that
  failed and the relevant input state (e.g., which export template,
  which work history entry).
- Structured logging MUST be used for backend operations. Log entries
  MUST include at minimum: timestamp, severity, component, and message.
- PDF generation failures MUST produce logs that identify the exact
  section and data element that caused the failure.
- Silent swallowing of errors is prohibited. Every error MUST be either
  handled with a recovery path or propagated to the caller.
- User-facing error messages MUST be clear and actionable, distinct
  from internal diagnostic logs.

**Rationale**: Resume export involves complex data transformation
(structured data to styled PDF). When output is wrong, users need to
know which data element or template section caused the issue, not just
that "something failed."

### VI. Code Quality

All code MUST conform to enforced formatting, linting, and structural
standards. Consistency is mandatory, not aspirational.

- Go code MUST pass `gofmt` and `go vet` with zero warnings. A linter
  configuration (e.g., `golangci-lint`) MUST be maintained and enforced.
- Frontend code MUST pass the configured linter and formatter with zero
  warnings.
- Functions exceeding 50 lines SHOULD be reviewed for decomposition.
  Functions exceeding 80 lines MUST be split unless a documented
  justification is provided.
- Exported Go types and functions MUST have doc comments describing
  their purpose and any non-obvious behavior.
- Dead code, commented-out code, and TODO comments without linked
  issues MUST be removed before merge.

**Rationale**: A single-developer project benefits disproportionately
from strict code quality enforcement because there is no second pair of
eyes to catch drift. Automated tooling substitutes for human review
consistency.

### VII. Versioning & Breaking Changes

All versioned artifacts MUST follow semantic versioning. Breaking
changes MUST be deliberate, documented, and accompanied by a migration
path.

- The application MUST use MAJOR.MINOR.PATCH versioning. MAJOR bumps
  indicate backward-incompatible changes to user data formats or
  exported file structures. MINOR bumps indicate new features. PATCH
  bumps indicate bug fixes.
- Changes to the persisted data schema (work history, certifications,
  skills, application tracking) MUST include a migration strategy that
  preserves existing user data without manual intervention.
- Resume template format changes MUST be versioned independently from
  the application version so that users can pin a template version for
  consistency across job applications.
- The Wails binding API between frontend and backend MUST treat any
  removal or signature change of an existing binding as a breaking
  change requiring a MAJOR version bump.
- Every release MUST include a changelog entry that clearly identifies
  breaking changes, new features, and fixes. Users MUST be able to
  determine upgrade impact from the changelog alone.

**Rationale**: Users will accumulate work history, application records,
and customized templates over time. Unmanaged breaking changes risk
data loss or silent corruption of resume output. Strict versioning
protects the user's investment in their data.

## Technology Constraints & Quality Standards

### Technology Stack

- **Language**: Go (latest stable release)
- **Desktop Framework**: Wails (Go backend + web-based frontend)
- **Target Platform**: macOS (primary); Windows and Linux (future)
- **PDF Generation**: ATS-compatible PDF output; the chosen library
  MUST produce text that is selectable, searchable, and free of
  mid-word spaces or unexpected line breaks
- **Data Persistence**: Local storage (file-based or embedded
  database); no external server dependencies for core functionality
- **Frontend**: Web technologies (HTML/CSS/JS or TypeScript) rendered
  within the Wails webview

### Quality Gates

- All tests MUST pass before a PR is merged.
- Linter and formatter checks MUST pass with zero warnings.
- New features MUST include tests covering primary success and error
  paths.
- PDF output for any new or modified template MUST be manually
  verified for ATS compatibility (no mid-word spaces, correct text
  extraction order).
- No new dependencies MUST be introduced without documenting the
  rationale and verifying the license is MIT-compatible.

### Data Integrity

- Work history, academic records, certifications, and skills data MUST
  be validated on input. Invalid dates, empty required fields, and
  duplicate entries MUST be rejected with clear error messages.
- All user data MUST be stored locally. No data MUST be transmitted to
  external services without explicit user consent (relevant for future
  AI integration).
- Export operations MUST NOT modify the source data. The data store
  MUST remain unchanged after any PDF or template rendering operation.

## Development Workflow & Release Process

### Branching Strategy

- Feature work MUST occur on feature branches following the naming
  convention `NNN-short-name` (e.g., `001-work-history-entry`).
- The `main` branch MUST always be in a buildable, testable state.
- Feature branches MUST be rebased or merged cleanly; force-pushes to
  `main` are prohibited.

### Code Review

- All changes MUST be submitted as pull requests with a clear
  description of what changed and why.
- PRs MUST reference the related specification or issue.
- Review MUST verify compliance with this constitution's principles
  before approval.

### Release Process

- Releases MUST follow semantic versioning (MAJOR.MINOR.PATCH).
- Each release MUST include a changelog entry summarizing user-visible
  changes.
- Release artifacts MUST be built from a tagged commit on `main`.
- macOS builds MUST be tested on the target platform before release.

### Continuous Integration

- CI MUST run on every push and PR: build, lint, test.
- CI failures MUST block merge.
- CI configuration MUST be version-controlled alongside the source.

## Governance

This constitution is the authoritative source of project principles and
development standards. It supersedes any conflicting guidance in other
documents, comments, or verbal agreements.

### Amendment Procedure

1. Proposed amendments MUST be documented in a PR with a clear
   rationale for the change.
2. Amendments MUST include a version bump following semantic versioning:
   - **MAJOR**: Removal or redefinition of a principle that changes
     existing compliance expectations.
   - **MINOR**: Addition of a new principle or material expansion of
     existing guidance.
   - **PATCH**: Clarifications, wording improvements, or typo fixes
     that do not alter compliance expectations.
3. Amendments MUST include a migration plan if existing code would be
   non-compliant under the new rules.
4. The Sync Impact Report (HTML comment at top of this file) MUST be
   updated with every amendment.

### Compliance Review

- All PRs and code reviews MUST verify compliance with the principles
  defined in this constitution.
- Complexity that violates Principle I (Simplicity First) MUST be
  justified in the PR description and tracked in the plan's Complexity
  Tracking table.
- Constitution violations discovered in existing code MUST be filed as
  issues and addressed within the current or next development cycle.

### Guidance Files

- Runtime development guidance (build commands, environment setup,
  agent-specific instructions) belongs in dedicated guidance files,
  not in this constitution.
- This constitution defines *what* and *why*; guidance files define
  *how*.

**Version**: 1.0.0 | **Ratified**: 2026-02-19 | **Last Amended**: 2026-02-19
