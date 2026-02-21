# Tasks: Resume Manager

**Input**: Design documents from `/specs/001-resume-manager/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/bindings.md, quickstart.md

**Tests**: Tests are included as co-located `_test.go` files per the plan. The spec requires data persistence and PDF correctness validation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**TDD Compliance**: Per Constitution Principle II (Test-First Development), test tasks are listed BEFORE their corresponding implementation tasks in every phase. The Red-Green-Refactor cycle is non-negotiable: write a failing test, then implement the code to make it pass. Data definition tasks (entity structs, constants) precede their tests since tests reference those types.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go backend**: `main.go`, `app.go`, `internal/`
- **Svelte frontend**: `frontend/src/`
- **Tests**: Co-located `_test.go` files + `tests/` for integration

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Initialize the Wails project, install dependencies, create the base directory structure, and configure CI.

- [x] T001 Initialize Wails v2 project with Svelte+TS template: run `wails init -n cut-the-bs -t svelte-ts`, configure `wails.json` with app metadata (title: "Cut the BS", width: 1200, height: 800, minWidth: 800, minHeight: 600)
- [x] T002 Create Go module and install core dependencies: `go mod init`, add `modernc.org/sqlite`, `github.com/signintech/gopdf`, `github.com/stretchr/testify` to `go.mod`
- [x] T003 Create backend directory structure per plan.md: `internal/domain/`, `internal/service/`, `internal/infra/sqlite/`, `internal/infra/pdf/`, `internal/config/`, `tests/integration/`, `tests/fixtures/`
- [x] T004 [P] Configure Go linting: add `.golangci.yml` with standard rules (gofmt, govet, errcheck, staticcheck, unused)
- [x] T005 [P] Install frontend dependencies and linting: `cd frontend && bun install`, configure Svelte TypeScript settings, install and configure ESLint + Prettier with `eslint-plugin-svelte` and `prettier-plugin-svelte`, add `lint` and `format` scripts to `frontend/package.json`
- [x] T006 [P] Embed font assets for PDF generation: add `assets/fonts/` directory with LiberationSans TTF files (Regular, Bold, Italic, BoldItalic), create `internal/infra/pdf/embed.go` with `//go:embed` directives
- [x] T007 [P] Configure CI pipeline: create GitHub Actions workflow (`.github/workflows/ci.yml`) that runs build, lint (Go + frontend), and test on every push and PR per Constitution CI requirements

**Checkpoint**: Project compiles, `wails dev` launches an empty window, `go test ./...` passes (no tests yet), CI pipeline runs on push.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T008 Implement app configuration in `internal/config/config.go`: data directory resolution (default: `~/Library/Application Support/cut-the-bs/`), backup settings struct, load/save config from JSON file
- [x] T009 Implement SQLite connection manager in `internal/infra/sqlite/store.go`: open database with WAL mode, busy_timeout, foreign_keys pragmas per research.md, single-connection pool, `Close()` method
- [x] T010 Implement domain entity structs in `internal/domain/entities.go`: all entity types matching data-model.md and contracts/bindings.md (UserProfile, ProfileLink, WorkHistoryEntry, AchievementBullet, SkillCategory, Skill, AcademicCredential, Certification, ProfessionalSummary, RoleDescriptor, Lens, CoverLetter, JobApplication, StatusChange, ResumeExport, ImportResult, BackupSettings, CompetenceLevel)
- [x] T011 [P] Define competence level scale constants in `internal/domain/competence.go`: 10 levels with numeric value, label, and descriptive criteria per FR-030
- [x] T012 [P] Define application status and fit indicator constants in `internal/domain/constants.go`: all 15 status values, all 5 fit indicator values per FR-023/FR-032
- [x] T013 Write migration tests in `internal/infra/sqlite/migrations_test.go`: verify v1 migration creates all 28 tables, verify foreign key constraints work, verify PRAGMA user_version is set correctly, verify idempotent re-run (TDD: tests written FIRST, will fail until T014 implements migrations)
- [x] T014 Implement schema migration system in `internal/infra/sqlite/migrations.go`: read `PRAGMA user_version`, apply sequential migrations, wrap each migration in a transaction. Include v1 migration with full schema from data-model.md (all 28 CREATE TABLE + all CREATE INDEX statements)
- [x] T015 Write domain validation tests in `internal/domain/validation_test.go`: date validation, required fields, email format, competence range, status values (TDD: tests written FIRST, will fail until T016 implements validation)
- [x] T016 [P] Implement domain validation rules in `internal/domain/validation.go`: date range validation (end >= start at coarsest granularity), required field checks, email format, URL format, competence level range (1-10), status value validation, fit indicator validation
- [x] T017 Define service interfaces in `internal/domain/interfaces.go`: Store interface (all DB operations grouped by entity), PDFRenderer interface (RenderResume, RenderCoverLetter) — per Constitution Principle III, interfaces are defined in the domain layer
- [x] T018 Create Wails app scaffold in `app.go`: App struct with `ctx context.Context`, lifecycle hooks (OnStartup: init services + open DB, OnBeforeClose: checkpoint DB, OnShutdown: close DB), wire service dependencies
- [x] T019 Create `main.go` entry point: `wails.Run()` with App binding, window config from T001, embed frontend assets
- [x] T020 [P] Set up structured logging in `internal/infra/logger.go`: configure `log/slog` with JSON handler, log level from config, helper for component-scoped loggers

**Checkpoint**: App launches, creates SQLite database on first run, migrations apply, database file exists at configured location. `go test ./internal/...` passes.

---

## Phase 3: User Story 1 - Work History Management (Priority: P1) MVP

**Goal**: Users can CRUD work history entries with achievement bullets, reorder both, and data persists across app restarts.

**Independent Test**: Launch app, create work history entries with varying date granularities and bullets, close and reopen, verify all data persists.

### Implementation for User Story 1

- [x] T021 [US1] Write work history query tests in `internal/infra/sqlite/work_history_test.go`: CRUD operations, sort_order updates, bullet cascade on entry delete, date granularity storage, reorder logic (TDD: tests written FIRST, will fail until T022 implements queries)
- [x] T022 [US1] Implement work history SQLite queries in `internal/infra/sqlite/work_history.go`: CreateWorkHistory, GetWorkHistory, UpdateWorkHistory, DeleteWorkHistory, ListWorkHistory (ordered by sort_order), ReorderWorkHistory, CreateBullet, UpdateBullet, DeleteBullet, ReorderBullets (within entry)
- [x] T023 [US1] Write work history service tests in `internal/service/workhistory_test.go`: validation errors (end before start, empty employer), bullet text splitting, happy paths (TDD: tests written FIRST, will fail until T024 implements service)
- [x] T024 [US1] Implement work history service in `internal/service/workhistory.go`: validation (date range, required fields), bullet text splitting (SplitBulletText), delegate to store, autosave semantics (every write is immediate)
- [x] T025 [US1] Add work history Wails bindings to `app.go`: ListWorkHistory, CreateWorkHistory, UpdateWorkHistory, DeleteWorkHistory, ReorderWorkHistory, CreateBullet, UpdateBullet, DeleteBullet, ReorderBullets, SplitBulletText — all delegating to service layer
- [x] T026 [US1] Create frontend navigation shell in `frontend/src/App.svelte`: sidebar navigation with menu items for each section (Work History, Skills, Education, Certifications, Summaries, Role Descriptors, Lenses, Export, Cover Letters, Applications, Settings), svelte-spa-router setup, active state highlighting
- [x] T027 [US1] Create frontend API service layer in `frontend/src/services/api.ts`: typed wrapper functions around Wails-generated bindings, error handling with user-facing toast/notification component (`frontend/src/components/Toast.svelte`) — auto-dismiss after timeout, display validation errors and operation failures
- [x] T028 [US1] Create frontend work history page in `frontend/src/pages/WorkHistory.svelte`: list view with entries ordered by sort_order, each entry expandable to show bullets, add/edit/delete entry forms, add/edit/delete bullet inline
- [x] T029 [US1] Create frontend work history components in `frontend/src/components/`: WorkHistoryCard.svelte (entry display with expand/collapse), BulletList.svelte (ordered bullet list with inline edit), DateInput.svelte (granularity selector: year/month/day), BulletPasteDialog.svelte (paste text, preview split, confirm)
- [x] T030 [US1] Implement drag-reorder for entries and bullets in `frontend/src/components/`: DragHandle.svelte reusable component, integrate with ReorderWorkHistory and ReorderBullets bindings

**Checkpoint**: Work history CRUD fully functional. User can create entries with flexible dates, add/edit/delete bullets, reorder both, close and reopen app with all data intact.

---

## Phase 4: User Profile & Profile Links

**Goal**: Users manage their contact information and profile links. Profile data appears on all exports.

**IMPORTANT**: This phase MUST complete before Phase 5 (PDF Export) so that profile headers are available for resume and cover letter templates.

### Implementation

- [x] T031 Write profile query tests in `internal/infra/sqlite/profile_test.go`: auto-create on first get, update, profile link CRUD, reorder (TDD: tests written FIRST, will fail until T032-T033 implement queries)
- [x] T032 [P] Implement user profile SQLite queries in `internal/infra/sqlite/profile.go`: GetProfile (create default if none), UpdateProfile
- [x] T033 [P] Implement profile link SQLite queries in `internal/infra/sqlite/profile_links.go`: CRUD operations, list ordered by sort_order, reorder
- [x] T034 Write profile service tests in `internal/service/profile_test.go`: email format validation, URL format validation, profile link ordering (TDD: tests written FIRST, will fail until T035 implements service)
- [x] T035 Implement profile service in `internal/service/profile.go`: validation (email format, URL format), profile link ordering
- [x] T036 Add profile Wails bindings to `app.go`: GetProfile, UpdateProfile, ListProfileLinks, CreateProfileLink, UpdateProfileLink, DeleteProfileLink, ReorderProfileLinks
- [x] T037 Create frontend profile/settings page in `frontend/src/pages/Settings.svelte`: profile form (name, email, phone, location), profile links list with add/edit/delete and drag-reorder, data directory configuration
- [x] T038 Update PDF templates to include profile header: render name, contact info, and profile links at top of both resume templates and cover letter template

**Checkpoint**: Profile data persists and appears on all generated PDFs. Profile links render in user-defined order.

---

## Phase 5: User Story 2 - PDF Resume Export (Priority: P2)

**Goal**: Users select a template, choose content, and generate an ATS-compatible PDF resume.

**Independent Test**: Select template, choose subset of work history entries, generate PDF, verify text is selectable and copy-pasteable without ATS artifacts.

### Implementation for User Story 2

- [x] T039 [US2] Write PDF renderer tests in `internal/infra/pdf/renderer_test.go`: verify PDF generates without error for complete data, verify PDF generates for minimal data (single entry), verify text extraction from generated PDF has no mid-word spaces (ATS validation), verify each section renders correctly (TDD: tests written FIRST, will fail until T040-T042 implement renderer)
- [x] T040 [US2] Implement PDF renderer foundation in `internal/infra/pdf/renderer.go`: initialize gopdf, load embedded fonts (Regular, Bold, Italic), page setup (Letter/A4), margin configuration, helper methods for text positioning and line drawing
- [x] T041 [US2] Implement resume template 1 ("Professional") in `internal/infra/pdf/template_professional.go`: single-column layout, contact header (name, email, phone, location, links), role descriptor bar, professional summary section, work history section with date-aligned entries and bullets, skills section (comma-separated under category headers with legacy skill filtering/visual distinction per FR-031), education section, certifications section
- [x] T042 [US2] Implement resume template 2 ("Modern") in `internal/infra/pdf/template_modern.go`: visually distinct layout from template 1, same sections rendered with different typography, spacing, and visual hierarchy, legacy skill filtering/visual distinction per FR-031
- [x] T043 [US2] Write resume export query tests in `internal/infra/sqlite/export_test.go`: export record creation, selection snapshot CRUD, list exports (TDD: tests written FIRST, will fail until T044)
- [x] T044 [US2] Implement resume export SQLite queries in `internal/infra/sqlite/export.go`: CreateExport (with all ExportSelection junction records), ListExports, GetExport, export selection snapshot queries
- [x] T045 [US2] Write resume export service tests in `internal/service/resume_test.go`: export with full selections, export with minimal selections, empty selection rejection, file path generation, legacy skill handling, per-export skill sort override (custom order applied without changing master list per FR-008) (TDD: tests written FIRST, will fail until T046)
- [x] T046 [US2] Implement resume export service in `internal/service/resume.go`: assemble all selected data from store, validate at least one content item selected, apply per-export skill sort override if provided (FR-008: custom ordering supersedes default competence-based sort without modifying master skill records), call renderer, save PDF to file, create export record with snapshot, template listing
- [x] T047 [US2] Add resume export Wails bindings to `app.go`: ListTemplates, PreviewExport, CreateExport, ListExports, OpenExportFile
- [x] T048 [US2] Create frontend export page in `frontend/src/pages/Export.svelte`: template selector with preview thumbnails, content selection panel (checkboxes for work history entries, bullets, skills, education, certifications, summary picker, descriptor picker), generate button, export history list
- [x] T049 [US2] Create frontend export components in `frontend/src/components/`: TemplateCard.svelte (template preview), ContentSelector.svelte (tree-style checkbox for entries → bullets), SkillSelector.svelte (grouped by category with legacy indicator, drag-reorder for per-export sort override per FR-008 — custom order does not modify master list), ExportHistoryList.svelte

**Checkpoint**: User can select a template, choose content, generate a PDF. PDF opens in system viewer, text is selectable, no ATS artifacts. Legacy skills are filtered or visually distinguished per FR-031.

---

## Phase 6: User Story 3 - Skills & Competence Tracking (Priority: P3)

**Goal**: Users manage skills with categories, competence levels, and relevancy. Skills auto-sort by competence and render on PDF under category headers.

**Independent Test**: Add skills across multiple categories with varying competence levels, verify sort order, generate a PDF and confirm skills section renders correctly.

### Implementation for User Story 3

- [x] T050 [US3] Write skill and category query tests in `internal/infra/sqlite/skills_test.go`: CRUD operations, category FK enforcement, sort ordering, delete category with existing skills fails, reorder categories (TDD: tests written FIRST, will fail until T051-T052 implement queries)
- [x] T051 [P] [US3] Implement skill category SQLite queries in `internal/infra/sqlite/skill_category.go`: CreateSkillCategory, RenameSkillCategory, DeleteSkillCategory (fail if skills still reference it), ListSkillCategories (ordered by sort_order), ReorderSkillCategories
- [x] T052 [P] [US3] Implement skill SQLite queries in `internal/infra/sqlite/skills.go`: CreateSkill, UpdateSkill, DeleteSkill, ListSkills (sorted by competence desc, name asc), ListSkillsByCategory (grouped by category sort_order, skills sorted within)
- [x] T053 [US3] Write skills service tests in `internal/service/skills_test.go`: validation errors, auto-sort verification, category operations, split text (TDD: tests written FIRST, will fail until T054)
- [x] T054 [US3] Implement skills service in `internal/service/skills.go`: validation (competence range, category exists), skill text splitting (SplitSkillsText), competence level constants, ListSkillsByCategory returns SkillCategoryWithSkills
- [x] T055 [US3] Add skills and category Wails bindings to `app.go`: ListSkills, ListSkillsByCategory, CreateSkill, UpdateSkill, DeleteSkill, SplitSkillsText, GetCompetenceLevels, ListSkillCategories, CreateSkillCategory, RenameSkillCategory, DeleteSkillCategory, ReorderSkillCategories
- [x] T056 [US3] Create frontend skills page in `frontend/src/pages/Skills.svelte`: skills grouped by category with competence badges, add/edit/delete skill forms with category dropdown and competence selector, category management panel (create, rename, reorder, delete), comma-separated skill paste input
- [x] T057 [US3] Create frontend skills components in `frontend/src/components/`: SkillCard.svelte (name, category, competence badge, legacy indicator), CompetenceSelector.svelte (level picker with descriptive hints), CategoryManager.svelte (list with drag-reorder, rename inline, add/delete), SkillPasteDialog.svelte (paste CSV, assign category/competence to each)

**Checkpoint**: Skills CRUD works with categories. Categories can be created, renamed, reordered, deleted (when empty). PDF export skills section renders comma-separated names under category headers in user-defined order.

---

## Phase 7: User Story 4 - Academic History & Certifications (Priority: P4)

**Goal**: Users record academic credentials and certifications with auto-computed active/inactive status for certs. Both support user-controlled sort ordering.

**Independent Test**: Add academic entries and certifications with various dates, verify active/inactive status computes correctly, confirm they appear in generated PDF.

### Implementation for User Story 4

- [x] T058 [US4] Write academic and certification query tests in `internal/infra/sqlite/academic_test.go`: CRUD, sort_order, certification active/inactive derivation, reorder operations (TDD: tests written FIRST, will fail until T059-T060 implement queries)
- [x] T059 [P] [US4] Implement academic credential SQLite queries in `internal/infra/sqlite/academic.go`: CRUD operations, list ordered by sort_order, ReorderAcademicCredentials
- [x] T060 [P] [US4] Implement certification SQLite queries in `internal/infra/sqlite/certifications.go`: CRUD operations, list ordered by sort_order, compute `is_active` at read time from expiration_date, ReorderCertifications
- [x] T061 [US4] Write academic/certification service tests in `internal/service/academic_test.go`: validation errors (empty institution, empty cert name), date format, reorder operations (TDD: tests written FIRST, will fail until T062)
- [x] T062 [US4] Implement academic/certification service in `internal/service/academic.go`: validation (required fields, date format), delegation to store
- [x] T063 [US4] Add academic and certification Wails bindings to `app.go`: ListAcademicCredentials, CreateAcademicCredential, UpdateAcademicCredential, DeleteAcademicCredential, ReorderAcademicCredentials, ListCertifications, CreateCertification, UpdateCertification, DeleteCertification, ReorderCertifications
- [x] T064 [US4] Create frontend education page in `frontend/src/pages/Education.svelte`: academic credentials list with add/edit/delete and drag-reorder, certifications list with active/inactive visual distinction and drag-reorder, expiration date highlighting
- [x] T065 [US4] Create frontend education components in `frontend/src/components/`: AcademicCard.svelte, CertificationCard.svelte (with active/expired badge computed from expiration_date)

**Checkpoint**: Academic and certification CRUD works. Expired certs show as inactive. Both sections support reordering and appear in generated PDF.

---

## Phase 8: User Story 5 - Professional Summary (Priority: P5)

**Goal**: Users create multiple summary variants and select one per export.

**Independent Test**: Create multiple summary variants, select one for export, verify only that summary appears at the top of the generated PDF.

### Implementation for User Story 5

- [x] T066 [US5] Write summary query tests in `internal/infra/sqlite/summaries_test.go`: CRUD, unique label constraint (TDD: tests written FIRST, will fail until T067)
- [x] T067 [US5] Implement professional summary SQLite queries in `internal/infra/sqlite/summaries.go`: CRUD operations, list all
- [x] T068 [US5] Write summary service tests in `internal/service/summaries_test.go`: validation (non-empty label and body, unique label), update and delete paths (TDD: tests written FIRST, will fail until T069)
- [x] T069 [US5] Implement summary service in `internal/service/summaries.go`: validation (non-empty label and body, unique label)
- [x] T070 [US5] Add summary Wails bindings to `app.go`: ListSummaries, CreateSummary, UpdateSummary, DeleteSummary
- [x] T071 [US5] Create frontend summaries page in `frontend/src/pages/Summaries.svelte`: list of summary variants with label and preview, add/edit/delete forms with multi-line text input
- [x] T072 [US5] Update export page to include summary selection: add summary picker dropdown to `frontend/src/pages/Export.svelte`, update ExportRequest to include selected summary_id

**Checkpoint**: Summary CRUD works. Export page shows summary picker. Selected summary renders at top of PDF.

---

## Phase 9: User Story 6 - Job Application Tracking (Priority: P6)

**Goal**: Users record job applications linked to resume exports and cover letters, track status changes with history.

**Independent Test**: Create application linked to exported resume, update status through several stages, verify history is searchable and linked resume is retrievable.

### Implementation for User Story 6

- [x] T073 [US6] Write cover letter query tests in `internal/infra/sqlite/cover_letters_test.go`: CRUD operations, title constraints (TDD: tests written FIRST, will fail until T074)
- [x] T074 [US6] Implement cover letter SQLite queries in `internal/infra/sqlite/cover_letters.go`: CRUD operations, list all
- [x] T075 [US6] Write cover letter PDF rendering tests in `internal/infra/pdf/cover_letter_test.go`: verify PDF generates, verify profile header renders, verify body text renders, verify ATS text extraction (TDD: tests written FIRST, will fail until T076)
- [x] T076 [US6] Implement cover letter PDF rendering in `internal/infra/pdf/cover_letter.go`: profile header (name, contact, links) at top, body text below, single built-in layout per FR-027
- [x] T077 [US6] Write cover letter service tests in `internal/service/cover_letters_test.go`: validation (non-empty title and body), PDF export delegation (TDD: tests written FIRST, will fail until T078)
- [x] T078 [US6] Implement cover letter service in `internal/service/cover_letters.go`: validation, PDF export delegation
- [x] T079 [US6] Add cover letter Wails bindings to `app.go`: ListCoverLetters, CreateCoverLetter, UpdateCoverLetter, DeleteCoverLetter, ExportCoverLetter
- [x] T080 [US6] Write application query tests in `internal/infra/sqlite/applications_test.go`: CRUD, status transitions, history recording, search, FK references to resume_export and cover_letter (TDD: tests written FIRST, will fail until T081)
- [x] T081 [US6] Implement job application SQLite queries in `internal/infra/sqlite/applications.go`: CRUD operations, status update with StatusHistory insert, search by company/position, list with current status
- [x] T082 [US6] Write job application service tests in `internal/service/applications_test.go`: validation (required fields, valid status/fit values), status update with history, search, list with computed status (TDD: tests written FIRST, will fail until T083)
- [x] T083 [US6] Implement job application service in `internal/service/applications.go`: validation (required fields, valid status/fit values), status update with history, search, list with computed status
- [x] T084 [US6] Add job application Wails bindings to `app.go`: ListApplications, SearchApplications, CreateApplication, UpdateApplication, UpdateApplicationStatus, UpdateApplicationFit, GetApplicationHistory, DeleteApplication, GetApplicationStatuses, GetFitIndicators
- [x] T085 [US6] Create frontend cover letters page in `frontend/src/pages/CoverLetters.svelte`: list view, create/edit with multi-line text, export to PDF button
- [x] T086 [US6] Create frontend applications page in `frontend/src/pages/Applications.svelte`: application list with status badges and fit indicators, search bar, add/edit forms, status update dropdown with history timeline, linked resume/cover letter viewer
- [x] T087 [US6] Create frontend application components in `frontend/src/components/`: ApplicationCard.svelte (company, position, status badge, fit indicator), StatusTimeline.svelte (chronological status changes), FitIndicator.svelte (visual scale), ApplicationSearch.svelte

**Checkpoint**: Full application tracking workflow: create cover letter, export it, create application linked to resume + cover letter, update status, search by company. History timeline shows all status changes.

---

## Phase 10: User Story 7 - Role Descriptor Tags (Priority: P7)

**Goal**: Users manage role descriptor tags and select a subset per export, rendered with template dividers.

**Independent Test**: Create descriptors, select subset for export, verify they appear separated by dividers in PDF. Select zero descriptors, verify section is omitted.

### Implementation for User Story 7

- [x] T088 [US7] Write descriptor query tests in `internal/infra/sqlite/descriptors_test.go`: CRUD, unique title constraint, reorder (TDD: tests written FIRST, will fail until T089)
- [x] T089 [US7] Implement role descriptor SQLite queries in `internal/infra/sqlite/descriptors.go`: CRUD operations, list ordered by sort_order, reorder
- [x] T090 [US7] Write descriptor service tests in `internal/service/descriptors_test.go`: validation (non-empty, unique title) (TDD: tests written FIRST, will fail until T091)
- [x] T091 [US7] Implement descriptor service in `internal/service/descriptors.go`: validation (non-empty, unique title)
- [x] T092 [US7] Add descriptor Wails bindings to `app.go`: ListDescriptors, CreateDescriptor, UpdateDescriptor, DeleteDescriptor, ReorderDescriptors
- [x] T093 [US7] Create frontend descriptors page in `frontend/src/pages/Descriptors.svelte`: list with drag-reorder, add/edit/delete forms
- [x] T094 [US7] Update export page for descriptor selection: add descriptor checkboxes to `frontend/src/pages/Export.svelte`, omit section when none selected

**Checkpoint**: Descriptors CRUD works. Export page includes descriptor selection. PDF renders selected descriptors with dividers, omits section when none selected.

---

## Phase 11: Lenses (Job-Type Variants)

**Goal**: Users create reusable content selection presets (lenses) that pre-fill export selections. Lens deletion warnings for referenced content.

### Implementation

- [x] T095 Write lens query tests in `internal/infra/sqlite/lenses_test.go`: CRUD, selection set/get for all 7 types, cascade on lens delete, skill lens tag operations (TDD: tests written FIRST, will fail until T096-T097)
- [x] T096 Implement lens SQLite queries in `internal/infra/sqlite/lenses.go`: CRUD for lens, set/get for all 7 selection tables (LensWorkHistorySelection, LensBulletSelection, LensSkillSelection, LensAcademicSelection, LensCertSelection, LensDescriptorSelection, linked summary), GetLensDetail (full selections)
- [x] T097 Implement skill lens tag SQLite queries in `internal/infra/sqlite/skill_lens_tags.go`: GetSkillLensTags, SetSkillLensTags, ListSkillsWithLensTags
- [x] T098 Write lens service tests in `internal/service/lenses_test.go`: validation, export selection conversion, reference checking for delete warnings (TDD: tests written FIRST, will fail until T099)
- [x] T099 Implement lens service in `internal/service/lenses.go`: validation, GetLensExportSelections (convert lens selections to ExportRequest), CheckLensReferences (find lenses referencing a given content item for delete warning per FR-050)
- [x] T100 Add lens Wails bindings to `app.go`: ListLenses, GetLens, CreateLens, UpdateLens, DeleteLens, SetLensWorkHistory, SetLensBullets, SetLensSkills, SetLensAcademics, SetLensCerts, SetLensDescriptors, GetLensExportSelections, GetSkillLensTags, SetSkillLensTags, ListSkillsWithLensTags, CheckSkillLensReferences
- [x] T101 Create frontend lens management page in `frontend/src/pages/Lenses.svelte`: lens list, create/edit lens form, content selection panel per lens (work history with bullet-level control, skills, education, certifications, descriptors, summary picker), per-lens ordering for entries and bullets
- [x] T102 Update export page for lens integration: add lens selector dropdown to `frontend/src/pages/Export.svelte`, on lens select → pre-fill all content selections via GetLensExportSelections, allow override without modifying saved lens
- [x] T103 Update skill page for lens tagging: add lens tag checkboxes to skill edit form in `frontend/src/pages/Skills.svelte`, show which lenses each skill belongs to
- [x] T104 Implement lens deletion warnings across all content pages: when deleting any content item, call CheckLensReferences, show confirmation dialog listing affected lenses before proceeding

**Checkpoint**: Lenses CRUD works. Selecting a lens on export page pre-fills all selections. Overrides don't modify saved lens. Deleting lens-referenced content shows warning with affected lens names.

---

## Phase 12: Data Management (Import/Export/Backup)

**Goal**: Users can export all data to JSON, restore from backup, import from CSV/JSON, configure backups.

### Implementation

- [x] T105 Write backup service tests in `internal/service/backup_test.go`: full export/import roundtrip, partial import with errors, rolling backup file management (TDD: tests written FIRST, will fail until T106-T109)
- [x] T106 Implement JSON full export in `internal/service/backup.go`: query all tables, marshal to structured JSON file, include schema version
- [x] T107 Implement JSON full import in `internal/service/backup.go`: parse JSON, validate schema version, truncate all tables, insert all records in transaction
- [x] T108 Implement CSV/JSON partial import in `internal/service/backup.go`: parse CSV or JSON for specific data types (work_history, skills, academic, certifications), map to entities, insert, return ImportResult with counts and errors
- [x] T109 Implement rolling backup system in `internal/service/backup.go`: checkpoint WAL, copy .db file with timestamp, prune oldest copies beyond configured count, trigger after configurable interval or on shutdown
- [x] T110 Add data management Wails bindings to `app.go`: ExportAllData, ImportAllData, ImportCSV, ImportJSON, GetDataDirectory, SetDataDirectory, GetBackupSettings, UpdateBackupSettings
- [x] T111 Update frontend settings page for data management in `frontend/src/pages/Settings.svelte`: data directory path display with change button, JSON export/import buttons, CSV import section, backup count configuration, backup history display
- [x] T112 Add autosave event handling: emit `autosave:complete` event from Go after each write, display autosave indicator in frontend status bar
- [x] T113 Add backup event handling: emit `backup:complete` and `backup:error` events, display backup status notifications in frontend

**Checkpoint**: Full data export/import works. Rolling backups create timestamped copies. Autosave indicator shows in UI. Data directory is configurable.

---

## Phase 13: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories.

- [x] T114 Implement zoom widget in `frontend/src/components/ZoomWidget.svelte`: bottom-left corner, 100% default, +/- buttons, Cmd/Ctrl +/- keyboard shortcuts, apply via CSS transform on root container per FR-048
- [x] T115 Add zoom widget to `frontend/src/App.svelte`: persistent across all pages, save zoom preference to localStorage
- [x] T116 Add loading states across all frontend pages: skeleton loaders or spinners during data fetch and PDF generation
- [x] T117 Implement consistent date formatting across all frontend components: use DateInput.svelte consistently, format display dates based on granularity
- [x] T118 PDF ATS validation integration test in `tests/integration/ats_test.go`: generate PDF with all sections populated, extract text, verify no mid-word spaces, verify correct reading order, verify all selected content present
- [x] T119 Full end-to-end integration test in `tests/integration/workflow_test.go`: create profile, add work history, add skills, create lens, generate resume with lens, create application linked to export, verify all data roundtrips through JSON export/import
- [x] T120 Performance validation tests in `tests/integration/performance_test.go`: verify PDF generation completes within 5 seconds for a full resume, verify UI data operations complete within 100ms perceived latency threshold (per plan.md performance goals)
- [x] T121 Run `go vet ./...` and `golangci-lint run` and fix all warnings
- [x] T122 Run `wails build` and verify production binary launches correctly on macOS
- [x] T123 Validate quickstart.md instructions: follow quickstart.md step-by-step on a clean checkout, verify all commands work

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Phase 2 completion — MVP
- **Profile (Phase 4)**: Depends on Phase 2 completion — MUST complete before Phase 5
- **User Stories (Phases 5-10)**: All depend on Phase 2 completion
  - US2 (Export, Phase 5) depends on Phase 3 (needs content to render) AND Phase 4 (needs profile header)
  - US3-US5 (Skills, Education, Summaries) can be implemented independently after Phase 3
  - US6 (Applications, Phase 9) depends on Phase 5 (Export) being functional
  - US7 (Descriptors, Phase 10) is independent after Phase 3
- **Lenses (Phase 11)**: Depends on all content entity phases (3-10) being complete
- **Data Management (Phase 12)**: Depends on all content entities existing (Phases 3-10)
- **Polish (Phase 13)**: Depends on all prior phases

### User Story Dependencies

```
Phase 2 (Foundation)
    │
    ├── Phase 3: US1 Work History (P1) ─── MVP
    │       │
    │       ├── Phase 4: Profile ──── MUST precede US2
    │       │       │
    │       │       └── Phase 5: US2 PDF Export (P2) ── needs content + profile
    │       │               │
    │       │               └── Phase 9: US6 Applications (P6) ── needs exports
    │       │
    │       ├── Phase 6: US3 Skills (P3) ──── independent
    │       ├── Phase 7: US4 Education (P4) ── independent
    │       ├── Phase 8: US5 Summaries (P5) ── independent
    │       └── Phase 10: US7 Descriptors (P7) ── independent
    │
    ├── Phase 11: Lenses ── after all content entities
    ├── Phase 12: Data Management ── after all content entities
    └── Phase 13: Polish ── after everything
```

### Within Each User Story (TDD Order)

1. Write query tests FIRST (they will fail — Red)
2. Implement SQLite queries (tests pass — Green)
3. Write service tests NEXT (they will fail — Red)
4. Implement service layer (tests pass — Green)
5. Add Wails bindings (after service layer)
6. Build frontend pages and components (after bindings)

### Parallel Opportunities

- T004, T005, T006, T007 (linting, frontend deps, fonts, CI) are independent
- T011, T012 (competence levels, status constants) are independent
- T032, T033 (profile queries, profile link queries) are independent
- T051, T052 (skill category queries, skill queries) are independent
- T059, T060 (academic queries, cert queries) are independent
- All user story phases after US1 that don't depend on each other can run in parallel if multiple developers are working

---

## Parallel Example: User Story 3 (Skills)

```bash
# Write tests first (TDD):
Task: T050 "Write skill and category query tests" (will fail — no queries yet)

# These can then run in parallel (different files):
Task: T051 "Implement skill category SQLite queries in internal/infra/sqlite/skill_category.go"
Task: T052 "Implement skill SQLite queries in internal/infra/sqlite/skills.go"

# Then sequentially (TDD for service):
Task: T053 "Write skills service tests" (will fail — no service yet)
Task: T054 "Implement skills service" (tests pass)
Task: T055 "Add Wails bindings" (needs T054)
Task: T056 "Create frontend page" (needs T055)
Task: T057 "Create frontend components" (needs T055)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1 (Work History)
4. **STOP and VALIDATE**: Launch app, create entries, restart, verify persistence
5. Working data entry system — foundation for everything else

### Incremental Delivery

1. Setup + Foundation -> App launches with empty database
2. US1 (Work History) -> Data entry works, persists (MVP!)
3. Profile (Phase 4) -> Contact info for exports
4. US2 (PDF Export, Phase 5) -> First deliverable PDF (profile header available)
5. US3 (Skills) -> Skills section on PDF
6. US4 (Education) -> Education + certs on PDF
7. US5 (Summaries) -> Summary section on PDF
8. US7 (Descriptors) -> Descriptor bar on PDF
9. Lenses (Phase 11) -> Reusable export presets
10. US6 (Applications) -> Full tracking workflow
11. Data Management (Phase 12) -> Backup/restore
12. Polish (Phase 13) -> Zoom, ATS tests, perf validation, build validation

---

## Summary

| Metric | Count |
|--------|-------|
| Total tasks | 123 |
| Phase 1 (Setup) | 7 |
| Phase 2 (Foundation) | 13 |
| Phase 3 (US1 - Work History) | 10 |
| Phase 4 (Profile) | 8 |
| Phase 5 (US2 - PDF Export) | 11 |
| Phase 6 (US3 - Skills) | 8 |
| Phase 7 (US4 - Education) | 8 |
| Phase 8 (US5 - Summaries) | 7 |
| Phase 9 (US6 - Applications) | 15 |
| Phase 10 (US7 - Descriptors) | 7 |
| Phase 11 (Lenses) | 10 |
| Phase 12 (Data Management) | 9 |
| Phase 13 (Polish) | 10 |
| Parallel opportunities | 6 groups |
| MVP scope | Phases 1-3 (30 tasks) |

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All user input is plain text; formatting is handled by the system
- Constitution Principle I (Simplicity First) applies: no ORM, no DI framework, plain SQL behind service interfaces
- Constitution Principle II (Test-First Development): test tasks ALWAYS precede their implementation tasks
- Constitution Principle III (Clean Architecture): interfaces defined in domain layer, not service layer

## Remediation Log

Changes made to address analysis findings:

| ID | Severity | Issue | Resolution |
|----|----------|-------|------------|
| CRITICAL-1 | Critical | TDD violation: implementation before tests in all phases | Reordered all phases: test tasks now precede implementation tasks per Constitution Principle II |
| CRITICAL-2 | Critical | Table count: plan.md/tasks.md said 26, data-model.md has 28 | Updated to 28 in plan.md and tasks.md |
| HIGH-3 | High | Profile (Phase 10) after US2 (Phase 4) — PDF templates lack profile header | Moved Profile to Phase 4; US2 is now Phase 5. Dependency graph updated. |
| HIGH-4 | High | Duplicate toast: T029 (Phase 3) and T105 (Phase 13) | Consolidated into T027 (Phase 3 API service layer). Removed from Phase 13. |
| HIGH-7 | High | Missing service-layer test tasks for 6 services | Added T034 (profile), T061 (academic/cert), T068 (summary), T077 (cover letter), T082 (application), T090 (descriptor), T098 (lens) |
| HIGH-8 | High | FR-031 legacy skills — no PDF implementation task | Added legacy skill filtering/visual distinction to T041 and T042 (resume templates) |
| MEDIUM-12 | Medium | No CI setup task (constitution mandates CI) | Added T007 (GitHub Actions CI pipeline) |
| MEDIUM-14 | Medium | Academic/Cert sort_order with no reorder tasks | Added ReorderAcademicCredentials and ReorderCertifications to T059, T060, T063 |
| MEDIUM-15 | Medium | Performance requirements with no validation tasks | Added T120 (performance validation tests) |
| MEDIUM-16 | Medium | FR-008 (per-export skill sort override) with zero task coverage | Added FR-008 coverage to T045 (export service tests), T046 (export service impl), and T049 (SkillSelector.svelte drag-reorder) |
