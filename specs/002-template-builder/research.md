# Research: Template Builder

**Phase 0 output** | **Branch**: `002-template-builder` | **Date**: 2026-02-21

## 1. PDF Rendering Architecture

### Decision: Extend the existing gopdf rendering pipeline with a template-driven dispatch layer

**Rationale**: The current renderer uses a `map[string]templateFunc` pattern where each template is a Go function. The data-driven approach replaces these functions with a single generic renderer that reads a `TemplateDetail` (template + ordered elements) and dispatches to per-element-type render functions. This preserves the existing gopdf infrastructure, font embedding, and all shared utilities (text wrapping, page breaks, bullet rendering).

**Alternatives considered**:
- Whole-renderer replacement (e.g., switch to a Go HTML-to-PDF library): Rejected because gopdf already produces ATS-compatible output, and the entire test infrastructure validates this.
- Template DSL (e.g., Go templates or custom markup language): Rejected because it requires users to write text, not drag-and-drop.

### Current Renderer Details

- **Entry point**: `Renderer.RenderResume(ctx, RenderResumeRequest) (string, error)` in `internal/infra/pdf/renderer.go`
- **Template lookup**: `r.templates[req.TemplateID]` returns a `templateFunc func(pdf *gopdf.GoPdf, req RenderResumeRequest) error`
- **Page constants**: US Letter 612x792pt, margins 72pt L/R, 54pt T/B, usable width 468pt
- **Fonts**: Liberation Sans (Regular, Bold, Italic, BoldItalic) embedded via `//go:embed`; BoldItalic is embedded but never used
- **Cover letters**: Separate `RenderCoverLetter(ctx, RenderCoverLetterRequest)` using `DefaultHeaderConfig()` + `renderWrappedText`

### Professional Template Formatting Reference

| Element | Font | Size | Style | Alignment | Separator | Spacing |
|---------|------|------|-------|-----------|-----------|---------|
| Name | Bold | 18 | Bold | Center | -- | +22pt after |
| Contact | Regular | 10 | Regular | Center | ` \| ` | +13pt after |
| Links | Regular | 10 | Regular | Center | ` \| ` | +13pt after |
| Post-header | -- | -- | -- | -- | -- | +6pt |
| Descriptors | Regular | 10 | Regular | Center | ` \| ` | +14pt after |
| Section heading | Bold | 12 | Bold UPPER | Left | -- | +10pt before, +14pt after heading, +4pt after underline |
| Heading underline | -- | -- | 0.5pt rule | -- | -- | -- |
| Summary (master) | Regular | 10 | Regular | Left | -- | per line +13pt |
| Summary (other) | Regular | 10 | Regular | Left | bullet | per line +13pt |
| Work title | Bold | 10 | Bold | Left | -- | +4pt gap between entries |
| Work employer | Italic | 10 | Italic | Same line | ` \u2014 ` | -- |
| Work dates | Regular | 9 | Regular | Right | -- | -- |
| Work summary | Italic | 10 | Italic | Left | -- | per line +13pt |
| Bullets (primary) | Regular | 10 | Regular | Left@94pt | `\u2022` at 84pt | per line +13pt |
| Outcomes label | Bold | 10 | Bold | Left@84pt | -- | +2pt before, +13pt after |
| Bullets (secondary) | Italic | 10 | Italic | Left@94pt | `\u2022` | per line +13pt |
| Skill category | Bold | 10 | Bold | Left | -- | -- |
| Skill names | Regular | 10 | Regular | Hanging wrap | `, ` | per line +13pt |
| Core expertise | Regular | 10 | Regular | Left | ` \| ` | per line +13pt |
| Academic credential | Bold | 10 | Bold | Left | -- | -- |
| Academic date | Regular | 9 | Regular | Right | -- | -- |
| Academic institution | Regular | 10 | Regular | Left | -- | +15pt after |
| Cert name | Bold | 10 | Bold | Left | -- | -- |
| Cert detail | Regular | 9 | Regular | Same line | ` \u2014 ` | +14pt after |

### Modern Template Differences from Professional

| Feature | Professional | Modern |
|---------|-------------|--------|
| Name size | 18pt | 22pt |
| Name alignment | Center | Left |
| Contact/links | Separate lines, center, `\|` | Single line, left, `\u00B7` (middle dot) |
| Descriptors | Regular, center, `\|` | Italic, left, `\u00B7` |
| Post-header | 6pt gap | 0.3pt horizontal rule + 6pt |
| Section heading size | 12pt | 11pt |
| Section gap before | 10pt | 14pt |
| Section underline | Yes (0.5pt) | No |
| Work title format | `Title \u2014 Employer` (bold+italic) | `Title, Employer` (all bold) |
| Work entry gap | 4pt | 6pt |
| Work title spacing | +13pt | +15pt |

### Function Reuse Matrix

| Function | Professional | Modern |
|----------|-------------|--------|
| `RenderProfileHeader` (shared) | YES | NO (own `renderModernHeader`) |
| `renderDescriptorBar` | YES | NO (own `renderModernDescriptors`) |
| `renderSectionHeading` | YES | NO (own `renderModernSectionHeading`) |
| `renderWorkEntry` | YES | NO (own `renderModernWorkEntry`) |
| `renderBulletPoint` | YES | YES |
| `renderSkillsSection` | YES | YES |
| `renderCoreExpertiseSection` | YES | YES |
| `renderAcademics` | YES | YES |
| `renderCertifications` | YES | YES |
| `renderWrappedText` / `renderWrappedTextHanging` | YES | YES |
| `checkPageBreak` | YES | YES |

---

## 2. Frontend Drag-and-Drop Approach

### Decision: Use `svelte-dnd-action` library

**Rationale**: The template builder requires three drag behaviors: (1) palette-to-canvas drag, (2) canvas reorder, and (3) nested container drops (loop elements containing sub-elements). `svelte-dnd-action` handles all three as first-class features with native Svelte 3 integration, TypeScript support, and built-in animations.

**Alternatives considered**:
- Native HTML5 DnD (current codebase approach in `DragHandle.svelte`): Rejected for nested containers -- the bubbling behavior of `dragenter`/`dragleave` across nested drop zones requires manual hit-testing that amounts to reimplementing what svelte-dnd-action provides. Works for flat lists but not container hierarchies.
- SortableJS with svelte-sortablejs wrapper: Rejected because SortableJS mutates the DOM directly, conflicting with Svelte's virtual DOM diffing and causing desync bugs.
- @formkit/drag-and-drop: Rejected due to less mature Svelte 3 integration and limited nested container support.

### Implementation Pattern

- **Two-phase events**: `consider` (preview during drag) and `finalize` (commit on drop) map naturally to showing placeholders during drag.
- **Cross-container**: Palette items create new elements on drop; canvas items reorder on drop. Use `dropFromOthersDisabled` to control which zones accept cross-drops.
- **Nested zones**: Loop containers render an inner `dndzone` that accepts sub-elements. The library handles depth resolution automatically.
- **New dependency**: `bun add svelte-dnd-action` (~15KB minified, no transitive deps).

---

## 3. Frontend State Management

### Decision: Use Svelte writable stores scoped to the template builder feature

**Rationale**: The template builder has fundamentally different state requirements than the existing CRUD pages. Multiple components (canvas, properties panel, palette) need shared access to the same state, and nested containers create deep component trees where prop drilling becomes impractical. A scoped store in `frontend/src/stores/templateBuilder.ts` provides clean shared state without the complexity of a global state management library.

**Alternatives considered**:
- Component-local state with prop drilling (current codebase pattern): Rejected because the three-panel layout with nested containers requires too many event dispatchers and prop chains. The existing CRUD pages each manage independent data; the template builder has multiple panels operating on the same data structure.
- Svelte context API: Rejected because context is read-only in Svelte 3 (no reactivity without wrapping a store in it, which reduces to using stores).

### Store Design

```typescript
// frontend/src/stores/templateBuilder.ts
export const canvasElements = writable<TemplateElement[]>([]);
export const selectedElementId = writable<string | null>(null);
export const selectedElement = derived(
  [canvasElements, selectedElementId],
  ([$elements, $id]) => $id ? findElementById($elements, $id) : null
);
```

---

## 4. Three-Panel Layout

### Decision: Use Flexbox (consistent with existing codebase)

**Rationale**: The existing codebase uses Flexbox exclusively with zero CSS Grid usage. While CSS Grid is technically optimal for three-column layouts, Flexbox handles this pattern well and maintains consistency.

**Layout**: Palette (240px fixed) | Canvas (flex: 1) | Properties (300px fixed, conditional)

---

## 5. Template Storage Schema

### Decision: SQLite migration v7 with separate `resume_template` and `template_element` tables; element config as JSON blob

**Rationale**: Follows established patterns exactly (parent table + child table with FK, sort_order, ON DELETE CASCADE). Element properties are stored as a JSON blob because element types have polymorphic configurations that would require many nullable columns or additional child tables if normalized. The JSON is parsed/validated at the Go layer.

**Alternatives considered**:
- Single table with JSON blob for entire template layout: Rejected because it prevents SQL-level queries on individual elements and doesn't follow the established parent-child pattern.
- Fully normalized element properties (one table per element type): Rejected as over-engineering -- violates Constitution Principle I (Simplicity First). The config blob is validated in Go, and the primary access pattern is "load all elements for a template" which doesn't need SQL-level property access.
- EAV (entity-attribute-value) for element properties: Rejected for the same complexity reasons.

### Built-in Template Seeding

**Decision**: Seed built-in templates in the v7 migration with `is_builtin = 1`

**Rationale**: The migration already runs in a transaction, seeding is trivial, and it makes backup/restore clean (built-in templates are just rows). The `is_builtin` flag prevents deletion (enforced at Go layer).

### Template-Lens Relationship

Templates and lenses are orthogonal: a lens selects content (which bullets, skills); a template selects layout (which sections, in what order, with what styling). At export time, the user selects both independently. The existing `resume_export.template_id TEXT` column will need migration to reference the new `resume_template.id INTEGER`.

### Schema Version Compatibility

The `ExportData.SchemaVersion` must increment to 7. The `ImportAllData` validator must accept version 7 data. Template tables must be added to the truncation list and import helpers.

---

## 6. Element Type Catalog

Based on the spec's functional requirements (FR-015 through FR-027, FR-034 through FR-039a) and the current rendering analysis:

### Resume Element Types

| Element Type | Config Properties | Source Data |
|-------------|-------------------|-------------|
| `profile_header` | `name_font_size`, `detail_font_size`, `alignment` (left/center), `link_separator`, `show_links` | `UserProfile`, `ProfileLink[]` |
| `role_descriptors` | `separator`, `font_style` (regular/italic) | `RoleDescriptor[]` |
| `professional_summary` | `show_master_as_paragraph`, `bullet_char` | `ProfessionalSummary[]` + `MasterSummaryID` |
| `work_history_loop` | `entry_gap`, `title_format` (dash/comma), `show_summary`, `show_outcomes`, `bullet_char`, `outcomes_label` | `WorkHistoryEntry[]` with `AchievementBullet[]` |
| `skills` | `group_by_category`, `include_legacy`, `legacy_suffix` | `Skill[]` + `SkillCategory[]` |
| `education_loop` | `show_date`, `show_institution` | `AcademicCredential[]` |
| `certifications_loop` | `show_date`, `show_expiration` | `Certification[]` |
| `core_expertise` | `separator` | `CoreExpertise[]` |
| `section_heading` | `text`, `font_size`, `bold`, `uppercase`, `underline`, `underline_weight` | Static |
| `horizontal_rule` | `weight` | Static |
| `spacer` | `height` | Static |
| `static_text` | `text`, `font_size`, `font_style`, `alignment` | Static |

### Cover Letter Element Types

| Element Type | Config Properties | Source Data |
|-------------|-------------------|-------------|
| `body_text` | `text` (with `{{variable}}` and `{{prompt:...}}` placeholders), `font_size`, `alignment` | Template variables + user prompts |
| `date` | `format`, `font_size`, `alignment` | Current date or specified date |
| `greeting` | `text`, `font_size`, `variable_substitution` | Template variables |
| `closing` | `text`, `font_size`, `show_name` | `UserProfile.FullName` |
| `recipient_address` | `fields` (name, company, address), `font_size` | Template variables |

### Common Formatting Properties (all text elements)

| Property | Type | Default |
|----------|------|---------|
| `font_size` | float64 | varies by element |
| `font_style` | enum: regular, bold, italic, bold_italic | regular |
| `alignment` | enum: left, center, right | left |
| `space_before` | float64 (points) | 0 |
| `space_after` | float64 (points) | varies |

---

## 7. Window Size

### Decision: Increase global window size (Option A, per user's choice)

The three-panel layout (240 + canvas + 300 = ~540px fixed + flexible canvas) requires more horizontal space. Current window is 1200x800. Recommendation: **1440x900** minimum, set in `main.go` (not `wails.json`).
