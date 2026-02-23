# Tasks: Template Builder

**Input**: Design documents from `/specs/002-template-builder/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/wails-bindings.md, quickstart.md

**Tests**: Included — TDD required per project constitution (Principle II) and plan.md gate.

**Organization**: Tasks grouped by user story (P1 → P2 → P3). Within each story: tests first → implementation. Each story is independently testable after completion.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1–US7)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Install new frontend dependency and configure window size for the three-panel builder layout.

- [x] T001 Install svelte-dnd-action dependency via `bun add svelte-dnd-action` in frontend/
- [x] T002 [P] Increase application window size from 1200x800 to 1440x900 in main.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain types, database schema, store layer, service layer, Wails bindings, and frontend API wiring that ALL user stories depend on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T003 Create DocumentTemplate, DocumentTemplateInput, TemplateElement, TemplateElementInput, and TemplateDetail domain types; add all element type constants (28 types: 23 resume + 5 cover letter), TemplateTypeResume/TemplateTypeCoverLetter, LoopElementTypes slice, ValidLoopChildren map, and element-type-to-template-type compatibility helpers in internal/domain/template.go
- [x] T004 Add template Store interface methods (ListDocumentTemplates, GetDocumentTemplate, CreateDocumentTemplate, UpdateDocumentTemplate, DeleteDocumentTemplate, DuplicateDocumentTemplate, CreateTemplateElement, UpdateTemplateElement, DeleteTemplateElement, ReorderTemplateElements) to internal/domain/interfaces.go
- [x] T005 [P] Add `Templates []TemplateDetail` field to ExportData struct in internal/domain/entities.go
- [x] T006 [P] Implement migration v7: CREATE TABLE document_template and template_element with indexes (idx_template_element_template, idx_template_element_parent), FK constraints with ON DELETE CASCADE, and ADD COLUMN resume_export.template_ref_id in internal/infra/sqlite/migrations.go
- [x] T007 Write migration v7 test verifying table creation, FK cascade behavior, index existence, and template_ref_id column in internal/infra/sqlite/migrations_test.go
- [x] T008 Implement template store CRUD (all 10 Store interface methods): list with built-in-first ordering, get with elements, create, update (reject built-in), delete (reject built-in, cascade elements), duplicate (deep copy with elements), element CRUD with auto sort_order, reorder in internal/infra/sqlite/templates.go
- [x] T009 Write template store tests: create/read/update/delete templates, built-in protection, duplicate with deep element copy, element CRUD, reorder, cascade delete, and sort ordering in internal/infra/sqlite/templates_test.go
- [x] T010 Implement template service with validation (name non-empty ≤100 chars, valid template_type, margins 0–288pt, element type compatibility with template type, single-level nesting enforcement, valid loop children) and template-to-render assembly in internal/service/template.go
- [x] T011 Write template service tests: table-driven tests for all validation rules (empty name, invalid type, out-of-range margins, wrong element type for template type, nested loop rejection, invalid loop children) in internal/service/template_test.go
- [x] T012 Add 14 template Wails binding methods to App struct (ListDocumentTemplates, GetDocumentTemplate, CreateDocumentTemplate, UpdateDocumentTemplate, DeleteDocumentTemplate, DuplicateDocumentTemplate, CreateTemplateElement, UpdateTemplateElement, DeleteTemplateElement, ReorderTemplateElements, ExportTemplate, ImportTemplate, PreviewTemplate, and update ExportResume to accept int64 TemplateID) in app.go
- [x] T013 [P] Add 14 template API wrapper functions (listDocumentTemplates through previewTemplate) in frontend/src/services/api.ts
- [x] T014 [P] Add /templates and /templates/:id/builder routes in frontend/src/App.svelte
- [x] T015 [P] Create templateBuilder Svelte writable stores (canvasElements, selectedElementId, derived selectedElement, currentTemplate for metadata) in frontend/src/stores/templateBuilder.ts

**Checkpoint**: Foundation ready — domain types, storage, service, bindings, and frontend wiring are in place. User story implementation can now begin.

---

## Phase 3: User Story 2 — Reproduce Existing Templates (Priority: P1) 🎯 MVP

**Goal**: Ship built-in "Professional" and "Modern" resume templates that produce PDF output identical to the current hardcoded templates. Validates the template-driven rendering pipeline.

**Independent Test**: Export a resume using the built-in Professional template and compare PDF output to current hardcoded Professional template — must be visually identical.

### Tests for User Story 2

- [x] T016 [P] [US2] Write renderer fidelity test: render resume with Professional built-in template, extract text from PDF, compare section order/content/formatting to output from current hardcoded renderProfessional function in internal/infra/pdf/renderer_test.go
- [x] T017 [P] [US2] Write renderer fidelity test: render resume with Modern built-in template, compare to current hardcoded renderModern output in internal/infra/pdf/renderer_test.go

### Implementation for User Story 2

- [x] T018 [US2] Create built-in Professional and Modern template configurations as TemplateDetail data structures with exact formatting values from research.md tables (Professional: 18pt centered name, pipe separators, 12pt bold uppercase headings with 0.5pt underlines, 4pt entry gap; Modern: 22pt left name, middle-dot separators, 11pt headings without underlines, 6pt entry gap) in internal/infra/pdf/builtin.go
- [x] T019 [US2] Seed built-in Professional and Modern resume templates with full element trees in migration v7 (INSERT statements for document_template rows with is_builtin=1 and all template_element rows with config JSON matching builtin.go values) and populate resume_export.template_ref_id for existing export records by matching template_id text to seeded built-in IDs in internal/infra/sqlite/migrations.go
- [x] T020 [US2] Implement element rendering dispatch function: parse element Config JSON into type-specific struct, switch on ElementType, call per-element render function; implement formatting element renderers (section_heading with optional underline rule, horizontal_rule, spacer, static_text with wrapped text) in internal/infra/pdf/elements.go
- [x] T021 [US2] Implement data-bound element renderers: profile_header (name + contact + links with configurable alignment and separator), role_descriptors (with configurable separator and font style), professional_summary (master as paragraph + others as bullets), skills (grouped by category with configurable separator), core_expertise (inline with separator) in internal/infra/pdf/elements.go
- [x] T022 [US2] Implement loop container renderers: work_history_loop (iterate entries, dispatch sub-elements work_title/work_employer/work_dates/work_summary/work_bullets/work_outcomes per entry with configurable entry_gap), education_loop (edu_credential/edu_institution/edu_date), certifications_loop (cert_name/cert_detail) in internal/infra/pdf/elements.go
- [x] T023 [US2] Refactor RenderResume in renderer.go: replace `r.templates[req.TemplateID]` function-map lookup with template-driven pipeline that sets page margins from DocumentTemplate fields and iterates TemplateDetail.Elements calling the dispatch function from elements.go in internal/infra/pdf/renderer.go
- [x] T024 [US2] Update ExportRequest.TemplateID from string to int64 and update RenderResumeRequest to embed DocumentTemplate + []TemplateElement instead of TemplateID string in internal/domain/entities.go
- [x] T025 [US2] Update resume service to load DocumentTemplate by ID from store, build RenderResumeRequest with template data + elements, and pass to renderer in internal/service/resume.go
- [x] T026 [US2] Update Export.svelte to load templates from DB via listDocumentTemplates API, display in template selector dropdown grouped by type (built-in first), and pass int64 template ID in export request in frontend/src/pages/Export.svelte

**Checkpoint**: Built-in Professional and Modern templates produce identical PDF output to hardcoded versions. Existing export workflows function with template-driven renderer. Backward-compatible with existing export records.

---

## Phase 4: User Story 1 — Build a Resume Template from Scratch (Priority: P1)

**Goal**: Users can visually create a resume template by dragging elements from a palette onto a canvas, arrange them, save, and export a PDF using the template.

**Independent Test**: Create a new template in the builder, drag elements (header, heading, work history loop with sub-elements), save, then export a resume using it — verify PDF renders all elements in configured order.

### Tests for User Story 1

- [x] T027 [US1] Write integration test: create template via CreateDocumentTemplate, add elements via CreateTemplateElement, reorder via ReorderTemplateElements, export resume, verify PDF contains expected sections in correct order in tests/integration/template_test.go

### Implementation for User Story 1

- [x] T028 [US1] Create TemplateBuilder.svelte page with three-panel Flexbox layout (Palette 240px fixed left | Canvas flex:1 center | Properties 300px fixed right, shown only when element selected), load template via GetDocumentTemplate on mount, populate stores in frontend/src/pages/TemplateBuilder.svelte
- [x] T029 [P] [US1] Create Palette.svelte component: list of draggable element types organized by category (Formatting: section_heading, horizontal_rule, spacer, static_text; Data: profile_header, role_descriptors, professional_summary, skills, core_expertise; Containers: work_history_loop, education_loop, certifications_loop), using svelte-dnd-action with copy behavior in frontend/src/components/template/Palette.svelte
- [x] T030 [US1] Create Canvas.svelte component: svelte-dnd-action drop zone receiving elements from Palette, displaying ordered ElementBlock components, supporting reorder via drag, calling CreateTemplateElement on palette drop and ReorderTemplateElements on reorder in frontend/src/components/template/Canvas.svelte
- [x] T031 [US1] Create ElementBlock.svelte component: displays element type label and icon, drag handle for reordering, delete button calling DeleteTemplateElement, click-to-select updating selectedElementId store in frontend/src/components/template/ElementBlock.svelte
- [x] T032 [US1] Create LoopContainer.svelte component: renders as ElementBlock with nested svelte-dnd-action drop zone for sub-elements, iteration indicator label (e.g., "Repeats for each work history entry"), accepts only valid child element types per ValidLoopChildren map in frontend/src/components/template/LoopContainer.svelte
- [x] T033 [US1] Implement template create flow: "New Template" button/dialog prompting for name and type (resume/cover letter), call CreateDocumentTemplate API, navigate to /templates/:id/builder on success in frontend/src/pages/TemplateBuilder.svelte
- [x] T034 [US1] Implement template auto-save: on element add/delete/reorder, immediately persist via API calls; show save status indicator in builder header in frontend/src/pages/TemplateBuilder.svelte

**Checkpoint**: Users can create a resume template via the builder, add/reorder/delete elements on the canvas, and export a PDF using their custom template. Properties panel shows but is not yet interactive (US3).

---

## Phase 5: User Story 5 — Work History Looping (Priority: P2)

**Goal**: Loop containers iterate over repeating data (work history, education, certifications) at export time. Sub-elements within loops define per-entry layout. Empty loops are omitted.

**Independent Test**: Create template with work_history_loop containing work_title + work_bullets sub-elements, export with 3 work entries selected — verify all 3 render in PDF with the sub-element layout.

### Tests for User Story 5

- [x] T035 [US5] Write integration test: create template with work_history_loop + sub-elements, export with multiple work entries, verify all entries render with correct sub-element layout in tests/integration/template_test.go
- [x] T036 [P] [US5] Write renderer test: loop with zero data items produces no output (section heading + loop both omitted) in internal/infra/pdf/renderer_test.go

### Implementation for User Story 5

- [x] T037 [US5] Implement sub-element drag-and-drop validation in LoopContainer.svelte: filter palette items to show only valid children per loop type (ValidLoopChildren map), reject invalid element type drops with visual feedback in frontend/src/components/template/LoopContainer.svelte
- [x] T038 [US5] Implement configurable entry_gap spacing between loop iterations in work_history_loop, education_loop, and certifications_loop renderers in internal/infra/pdf/elements.go
- [x] T039 [US5] Implement empty loop handling: when no data items exist for a loop type at export time, omit the entire loop output and any immediately preceding section_heading element in internal/infra/pdf/elements.go

**Checkpoint**: Loop containers accept valid sub-elements via DnD, render correct layout for each data entry, handle configurable spacing between entries, and gracefully omit when no data exists.

---

## Phase 6: User Story 3 — Configure Element Formatting (Priority: P2)

**Goal**: Properties panel shows context-sensitive formatting options when an element is selected. Changes persist via UpdateTemplateElement and appear in exported PDFs.

**Independent Test**: Create template, add section_heading, change font size to 14pt and toggle underline off in properties panel, export — verify PDF reflects the formatting changes.

### Tests for User Story 3

- [x] T040 [US3] Write integration test: create template, add elements, update element configs via UpdateTemplateElement with modified formatting values, export resume, verify PDF reflects formatting changes in tests/integration/template_test.go

### Implementation for User Story 3

- [x] T041 [US3] Create Properties.svelte panel component shell: render when selectedElement is non-null, display element type header, show property editors based on element_type, show template-level properties when no element selected in frontend/src/components/template/Properties.svelte
- [x] T042 [US3] Implement property editors for formatting elements: section_heading (text input, font_size number, bold/uppercase/underline toggles, underline_weight, space_before/space_after), horizontal_rule (weight, spacing), spacer (height), static_text (text textarea, font_size, font_style select, alignment select, spacing) in frontend/src/components/template/Properties.svelte
- [x] T043 [US3] Implement property editors for data-bound elements: profile_header (name_font_size, detail_font_size, alignment select, link_separator, show_links toggle), role_descriptors (separator, font_style), professional_summary (bullet_char select), skills (group_by_category toggle, include_legacy toggle, legacy_suffix, skill_separator), core_expertise (separator) in frontend/src/components/template/Properties.svelte
- [x] T044 [US3] Implement property editors for loop containers and sub-elements: work_history_loop (entry_gap), work_title (font_size, font_style, include_employer toggle, employer_separator, employer_font_style), work_bullets/work_outcomes (bullet_char select, indent, outcomes_label), work_dates (font_size, alignment), education/certification sub-element editors in frontend/src/components/template/Properties.svelte
- [x] T045 [US3] Implement property save flow: on input change → debounce 300ms → parse current config JSON → update changed field → serialize to JSON → call updateTemplateElement API → update canvasElements store with returned element in frontend/src/components/template/Properties.svelte
- [x] T046 [US3] Implement template-level margin editor: when no element selected, show margin inputs (top, bottom, left, right in points with inch conversion display), call updateDocumentTemplate API on change in frontend/src/components/template/Properties.svelte

**Checkpoint**: All element types have functional property editors. Formatting changes persist and appear correctly in exported PDFs. Template margins are configurable.

---

## Phase 7: User Story 4 — Build a Cover Letter Template (Priority: P2)

**Goal**: Users can create cover letter templates with body text, variable substitution (`{{company_name}}`), and guided prompt placeholders (`{{prompt: text}}`). Two built-in cover letter templates (Formal, Casual) ship as defaults.

**Independent Test**: Create cover letter template with body_text containing `{{company_name}}` and `{{prompt: Why this role?}}`, generate for a job application — verify PDF contains the substituted company name and user's prompt response.

### Tests for User Story 4

- [x] T047 [US4] Write test for template variable parsing: extract `{{variable_name}}` and `{{prompt: descriptive text}}` placeholders from body_text config, return list of variables and prompts in internal/service/template_test.go
- [x] T048 [P] [US4] Write test for variable substitution: given body text with `{{company_name}}` and `{{position_title}}`, substitute from job application data map, verify output text in internal/infra/pdf/renderer_test.go

### Implementation for User Story 4

- [x] T049 [US4] Implement cover letter element renderers: body_text (render paragraphs with variable substitution applied), date (format current/specified date per Go time format string), greeting (text with optional variable substitution), closing (text + user full name from profile), recipient_address (multi-line fields with variable substitution) in internal/infra/pdf/elements.go
- [x] T050 [US4] Implement template variable parsing in service layer: scan cover letter element configs for `{{variable_name}}` and `{{prompt: text}}` patterns, return structured list of TemplateVariable and GuidedPrompt entries with their source locations in internal/service/template.go
- [x] T051 [US4] Implement variable substitution in renderer: accept substitution map (variable_name → value), replace all `{{variable_name}}` tokens in text before rendering, replace `{{prompt:...}}` tokens with user-provided responses in internal/infra/pdf/renderer.go
- [x] T052 [US4] Create PromptDialog.svelte component: given list of guided prompts from template, display each prompt text as a label with text area for user response, collect all responses into substitution map, return map on submit in frontend/src/components/coverletter/PromptDialog.svelte
- [x] T053 [US4] Create built-in Formal and Casual cover letter template configurations (Formal: profile header, date, recipient address block, greeting, structured body paragraphs with variable placeholders, formal closing; Casual: profile header, greeting, conversational body, informal closing) in internal/infra/pdf/builtin.go
- [x] T054 [US4] Seed built-in Formal and Casual cover letter templates with full element trees in migration v7 in internal/infra/sqlite/migrations.go
- [x] T055 [US4] Filter element palette by template_type: show only resume elements (profile_header, role_descriptors, professional_summary, loops, skills, core_expertise) for resume templates; show cover-letter elements (body_text, date, greeting, closing, recipient_address) plus shared formatting elements for cover letter templates; reject incompatible drops with visual indicator in frontend/src/components/template/Palette.svelte
- [x] T056 [US4] Implement cover letter export flow: on export with cover letter template, parse variables → if job application linked, auto-fill company_name/position_title → show PromptDialog for guided prompts → pass substitution map to renderer in frontend/src/pages/Export.svelte
- [x] T057 [US4] Implement manual variable input: when generating cover letter without linked job application, prompt user to enter company_name and position_title values before showing guided prompts in frontend/src/pages/Export.svelte

**Checkpoint**: Cover letter templates with variables and prompts produce correct PDF output. Built-in Formal and Casual templates available. Palette filters elements by template type.

---

## Phase 8: User Story 6 — Preview PDF Output (Priority: P3)

**Goal**: Users can click Preview in the builder to generate a PDF from the current template, providing rapid feedback without going through the full export workflow.

**Independent Test**: Build a template with several elements, click Preview — verify a PDF opens that reflects the current template layout with actual user data.

### Tests for User Story 6

- [x] T058 [US6] Write test for PreviewTemplate: invoke with template ID, verify PDF file is generated at returned path and contains expected content using user profile data in internal/service/template_test.go

### Implementation for User Story 6

- [x] T059 [US6] Implement PreviewTemplate in service layer: load template by ID, load user profile + profile links + most recent lens data (or generate placeholder data if none exist), assemble render request, generate PDF to temp file, return file path in internal/service/template.go
- [x] T060 [US6] Handle cover letter preview: when template is cover_letter type, substitute `{{company_name}}` with "[Company Name]", `{{position_title}}` with "[Position Title]", and `{{prompt:...}}` with "[Prompt response placeholder]" in internal/service/template.go
- [x] T061 [US6] Add Preview button to TemplateBuilder.svelte header bar: call previewTemplate API with current template ID, open returned PDF file path via Wails runtime BrowserOpenURL or file open, show loading state during generation in frontend/src/pages/TemplateBuilder.svelte

**Checkpoint**: Preview generates and opens a PDF within 5 seconds (SC-006). Cover letter previews show placeholder text for variables.

---

## Phase 9: User Story 7 — Manage Templates (Priority: P3)

**Goal**: Users can view all templates in a list, rename user-created templates, duplicate any template, and delete user-created templates. Built-in templates cannot be deleted or renamed.

**Independent Test**: Create several templates, rename one, duplicate another (including a built-in), delete a third — verify list updates correctly and built-in templates are protected.

### Tests for User Story 7

- [x] T062 [US7] Write integration test: create templates, rename, duplicate (user-created and built-in), delete, verify list state and built-in protection (cannot delete or rename built-in) in tests/integration/template_test.go

### Implementation for User Story 7

- [x] T063 [US7] Create TemplateList.svelte page: load all templates via listDocumentTemplates, display in table/list with columns for name, type badge (resume/cover letter), built-in indicator, and action buttons in frontend/src/pages/TemplateList.svelte
- [x] T064 [US7] Implement rename flow: inline edit or modal for user-created templates (disabled for built-in), call updateDocumentTemplate API with new name, refresh list on success in frontend/src/pages/TemplateList.svelte
- [x] T065 [US7] Implement duplicate flow: prompt for new name (default: "Copy of [original]"), call duplicateDocumentTemplate API, add new template to list on success in frontend/src/pages/TemplateList.svelte
- [x] T066 [US7] Implement delete flow: confirmation dialog for user-created templates, call deleteDocumentTemplate API, remove from list on success; hide delete button for built-in templates in frontend/src/pages/TemplateList.svelte
- [x] T067 [US7] Add navigation: "New Template" button opening create dialog, "Edit" button per template navigating to /templates/:id/builder, link from main navigation to /templates in frontend/src/pages/TemplateList.svelte

**Checkpoint**: Template list is fully functional with all CRUD operations and built-in protection enforced.

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Data portability (FR-051, FR-052), backup/restore integration, edge case handling, ATS validation, and final quality checks.

- [x] T068 Implement ExportTemplate in service: serialize TemplateDetail (template + all elements) to standalone JSON file at specified output path in internal/service/template.go
- [x] T069 Implement ImportTemplate in service: read JSON file, validate structure and element types, create as user-created template (is_builtin=false) with new IDs, return created template in internal/service/template.go
- [x] T070 [P] Write tests for template export/import round-trip: export template, import it back, verify identical structure and configs in internal/service/template_test.go
- [x] T071 Add templates to backup/restore: include TemplateDetail list in ExportAllData output and process templates in ImportAllData with proper ID remapping in internal/infra/sqlite/data_management.go
- [x] T072 [P] Write test for backup/restore including templates: export all data with templates, import into fresh DB, verify templates and elements restored correctly in internal/infra/sqlite/data_management_test.go
- [x] T073 Handle edge case: template references data sections with no data — ensure all element renderers silently omit output when their source data is empty (e.g., no certifications → certifications_loop produces nothing) in internal/infra/pdf/elements.go
- [x] T074 Handle edge case: unknown element types — skip with structured warning log identifying the unknown type and template name in internal/infra/pdf/elements.go
- [x] T075 Handle edge case: deleted template referenced by export record — ensure resume_export retains template name as snapshot text even after template deletion in internal/infra/sqlite/templates.go
- [x] T076 Handle edge case: cover letter variable references missing job application field — replace with empty string and return warning to user before PDF generation in internal/service/template.go
- [x] T077 Handle edge case: incompatible element type drop — in Canvas.svelte, validate element_type against template's template_type before accepting drop, show rejection indicator with explanation in frontend/src/components/template/Canvas.svelte
- [x] T078 Write ATS compatibility tests for user-created template PDF output: export resume with custom template, extract text, verify no mid-word spaces and correct text extraction order in tests/integration/ats_test.go
- [x] T079 Run full CI pipeline: `make ci` (go vet, golangci-lint, gofmt check, go test -race, svelte-check, frontend build, production build)

**Checkpoint**: Feature complete — all user stories functional, edge cases handled, data portability working, ATS compatibility validated, CI passing.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US2 (Phase 3)**: Depends on Foundational — renderer refactor is the core backend work
- **US1 (Phase 4)**: Depends on US2 — needs template-driven renderer to verify exported PDFs
- **US5 (Phase 5)**: Depends on US1 — needs builder UI for loop container DnD
- **US3 (Phase 6)**: Depends on US1 — needs builder UI for properties panel integration
- **US4 (Phase 7)**: Depends on US2 (renderer pipeline) + US3 (properties panel for config editing)
- **US6 (Phase 8)**: Depends on US2 (renderer) + US1 (builder UI for preview button)
- **US7 (Phase 9)**: Depends on Foundational only — can start after Phase 2 (list page is independent of builder)
- **Polish (Phase 10)**: Depends on all user stories being complete

### User Story Dependencies

- **US2 (P1)**: After Foundational → no other story dependencies
- **US1 (P1)**: After US2 → renderer must work for export verification
- **US5 (P2)**: After US1 → needs builder UI with loop containers
- **US3 (P2)**: After US1 → needs builder UI with element selection
- **US4 (P2)**: After US2 + US3 → needs renderer pipeline + properties panel
- **US6 (P3)**: After US1 + US2 → needs builder + renderer
- **US7 (P3)**: After Foundational → independent of builder (list page only)

### Within Each User Story

- Tests MUST be written first and confirmed to FAIL before implementation (Constitution Principle II)
- Domain/store changes before service changes
- Service changes before UI changes
- Core implementation before integration
- Story checkpoint validation before moving to next priority

### Parallel Opportunities

**Phase 2**: T005 + T006 can run in parallel (different files). T013 + T014 + T015 (frontend files) can run in parallel with each other and with backend tasks once interfaces are defined.

**Phase 3**: T016 + T017 (fidelity tests) can run in parallel. T018 built-in configs are in one file but Professional and Modern are independent functions.

**Phase 5 + Phase 6**: US5 and US3 can run in parallel after US1 — US5 modifies elements.go + LoopContainer.svelte; US3 creates Properties.svelte. No file conflicts.

**Phase 8 + Phase 9**: US6 and US7 can run in parallel — US6 adds preview to TemplateBuilder.svelte; US7 creates TemplateList.svelte. No file conflicts.

---

## Parallel Example: Phase 2 (Foundational)

```bash
# Batch 1: Domain types first (other tasks depend on these)
Task: T003 "Create domain types and constants in internal/domain/template.go"

# Batch 2: Interfaces + migration + ExportData (parallel — different files)
Task: T004 "Store interface methods in internal/domain/interfaces.go"
Task: T005 "ExportData.Templates field in internal/domain/entities.go"
Task: T006 "Migration v7 tables in internal/infra/sqlite/migrations.go"

# Batch 3: Frontend wiring (parallel — all different files, no backend deps)
Task: T013 "API functions in frontend/src/services/api.ts"
Task: T014 "Routes in frontend/src/App.svelte"
Task: T015 "Stores in frontend/src/stores/templateBuilder.ts"
```

## Parallel Example: User Story 2

```bash
# Batch 1: Fidelity tests first (TDD — must fail before implementation)
Task: T016 "Professional fidelity test in renderer_test.go"
Task: T017 "Modern fidelity test in renderer_test.go"

# Batch 2: Built-in configs (independent functions, same file)
Task: T018 "Professional + Modern configs in builtin.go"
```

---

## Implementation Strategy

### MVP First (US2 → US1)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks everything)
3. Complete Phase 3: US2 (renderer refactor + built-in templates)
4. **STOP AND VALIDATE**: Export with built-in Professional/Modern → output identical to current
5. Complete Phase 4: US1 (builder UI)
6. **STOP AND VALIDATE**: Create template in builder, export PDF → elements render correctly
7. This is the MVP — existing templates work, basic builder functional

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. US2 → Built-in templates work, no regression ✓ (MVP backend)
3. US1 → Builder works → Core feature delivered (MVP!)
4. US5 + US3 (parallel) → Loops + formatting → Full builder
5. US4 → Cover letters → Complete template system
6. US6 + US7 (parallel) → Preview + management → Quality of life
7. Polish → Edge cases + portability → Production ready

### Parallel Team Strategy

With multiple developers after Foundational is complete:
- Developer A: US2 (backend renderer) → US4 (cover letters)
- Developer B: US1 (builder UI) → US3 (properties) → US5 (loops)
- Developer C: US7 (template list) → US6 (preview) → Polish

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks in same phase
- [Story] label maps task to specific user story for traceability
- All element renderers in internal/infra/pdf/elements.go are listed as sequential tasks (same file, worked on progressively)
- Migration v7 in internal/infra/sqlite/migrations.go is modified by T006 (schema), T019 (resume template seeds), and T054 (cover letter seeds) — sequential modifications across phases
- Constitution Principle II (TDD): Test tasks precede implementation tasks in every user story phase
- Constitution Principle VII: Migration v7 preserves existing data; ExportRequest.TemplateID change is on feature branch
- Existing hardcoded template files (template_professional.go, template_modern.go) should be removed AFTER fidelity tests pass — not before
