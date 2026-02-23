# Data Model: Template Builder

**Phase 1 output** | **Branch**: `002-template-builder` | **Date**: 2026-02-21

## New Domain Entities

### DocumentTemplate

Replaces the existing `ResumeTemplate` struct (which was a simple name/description holder for hardcoded templates). A `DocumentTemplate` is a persistable, configurable template definition.

```go
// DocumentTemplate is a named, typed document layout configuration
// that defines how a resume or cover letter is rendered to PDF.
// Built-in templates are seeded during migration and cannot be
// deleted by users.
type DocumentTemplate struct {
    ID           int64   `json:"id"`
    Name         string  `json:"name"`
    Description  string  `json:"description"`
    TemplateType string  `json:"template_type"` // "resume" or "cover_letter"
    IsBuiltin    bool    `json:"is_builtin"`
    MarginTop    float64 `json:"margin_top"`    // points (default: 54.0 = 0.75in)
    MarginBottom float64 `json:"margin_bottom"` // points (default: 54.0 = 0.75in)
    MarginLeft   float64 `json:"margin_left"`   // points (default: 72.0 = 1.0in)
    MarginRight  float64 `json:"margin_right"`  // points (default: 72.0 = 1.0in)
    CreatedAt    string  `json:"created_at"`
    UpdatedAt    string  `json:"updated_at"`
}
```

**Fields**:
- `ID`: Auto-incrementing primary key (SQLite rowid)
- `Name`: User-provided template name. Must be non-empty. Unique constraint not enforced (users may name templates however they wish).
- `Description`: Optional description text.
- `TemplateType`: Discriminator. Valid values: `"resume"`, `"cover_letter"`. Determines which element types are valid for this template.
- `IsBuiltin`: True for the 4 seeded templates (Professional, Modern, Formal, Casual). Prevents deletion and rename.
- `MarginTop/Bottom/Left/Right`: Page margins in points. Per-template configurable (FR-033a). Defaults match current constants.
- `CreatedAt/UpdatedAt`: ISO 8601 UTC timestamps (standard pattern).

### DocumentTemplateInput

```go
// DocumentTemplateInput is the input type for creating or updating
// a document template.
type DocumentTemplateInput struct {
    Name         string  `json:"name"`
    Description  string  `json:"description"`
    TemplateType string  `json:"template_type"`
    MarginTop    float64 `json:"margin_top"`
    MarginBottom float64 `json:"margin_bottom"`
    MarginLeft   float64 `json:"margin_left"`
    MarginRight  float64 `json:"margin_right"`
}
```

### TemplateElement

```go
// TemplateElement is a single block within a template layout. Each
// element has a type that determines its rendering behavior and a
// JSON config blob containing type-specific properties. Elements
// may be top-level or children of a loop container (single-level
// nesting only; no nested loops per spec clarification).
type TemplateElement struct {
    ID          int64  `json:"id"`
    TemplateID  int64  `json:"template_id"`
    ParentID    *int64 `json:"parent_id"`    // nil for top-level, set for loop children
    ElementType string `json:"element_type"` // see Element Type Constants
    Config      string `json:"config"`       // JSON blob of type-specific properties
    SortOrder   int    `json:"sort_order"`
    CreatedAt   string `json:"created_at"`
    UpdatedAt   string `json:"updated_at"`
}
```

**Fields**:
- `ID`: Auto-incrementing primary key.
- `TemplateID`: FK to `document_template.id` (ON DELETE CASCADE).
- `ParentID`: Nullable FK to `template_element.id` (self-referential). `nil` for top-level elements, set for sub-elements within a loop container. Enforces single-level nesting: only elements whose own `ParentID` is nil can serve as parents (validated in Go, not SQL).
- `ElementType`: Discriminator for rendering dispatch. See constants below.
- `Config`: JSON string containing type-specific properties. Parsed and validated at the Go service layer. Default: `"{}"`.
- `SortOrder`: Integer position within its parent scope (top-level or within a specific loop container).

### TemplateElementInput

```go
// TemplateElementInput is the input type for creating or updating
// a template element.
type TemplateElementInput struct {
    ParentID    *int64 `json:"parent_id"`
    ElementType string `json:"element_type"`
    Config      string `json:"config"`
}
```

### TemplateDetail

```go
// TemplateDetail is a template with all its elements included,
// organized as a tree (top-level elements with children for loops).
type TemplateDetail struct {
    DocumentTemplate
    Elements []TemplateElement `json:"elements"` // all elements, flat list ordered by sort_order
}
```

---

## Element Type Constants

```go
// Template type constants.
const (
    TemplateTypeResume      = "resume"
    TemplateTypeCoverLetter = "cover_letter"
)

// Resume element type constants.
const (
    ElementProfileHeader    = "profile_header"
    ElementRoleDescriptors  = "role_descriptors"
    ElementProfSummary      = "professional_summary"
    ElementWorkHistoryLoop  = "work_history_loop"
    ElementWorkTitle        = "work_title"
    ElementWorkEmployer     = "work_employer"
    ElementWorkDates        = "work_dates"
    ElementWorkSummary      = "work_summary"
    ElementWorkBullets      = "work_bullets"
    ElementWorkOutcomes     = "work_outcomes"
    ElementSkills           = "skills"
    ElementEducationLoop    = "education_loop"
    ElementEduCredential    = "edu_credential"
    ElementEduInstitution   = "edu_institution"
    ElementEduDate          = "edu_date"
    ElementCertsLoop        = "certifications_loop"
    ElementCertName         = "cert_name"
    ElementCertDetail       = "cert_detail"
    ElementCoreExpertise    = "core_expertise"
    ElementSectionHeading   = "section_heading"
    ElementHorizontalRule   = "horizontal_rule"
    ElementSpacer           = "spacer"
    ElementStaticText       = "static_text"
)

// Cover letter element type constants.
const (
    ElementBodyText         = "body_text"
    ElementDate             = "date"
    ElementGreeting         = "greeting"
    ElementClosing          = "closing"
    ElementRecipientAddress = "recipient_address"
)

// Loop container element types (can have children).
var LoopElementTypes = []string{
    ElementWorkHistoryLoop,
    ElementEducationLoop,
    ElementCertsLoop,
}
```

### Valid Children per Loop Type

| Loop Container | Valid Child Element Types |
|----------------|-------------------------|
| `work_history_loop` | `work_title`, `work_employer`, `work_dates`, `work_summary`, `work_bullets`, `work_outcomes`, `section_heading`, `horizontal_rule`, `spacer`, `static_text` |
| `education_loop` | `edu_credential`, `edu_institution`, `edu_date`, `section_heading`, `horizontal_rule`, `spacer`, `static_text` |
| `certifications_loop` | `cert_name`, `cert_detail`, `section_heading`, `horizontal_rule`, `spacer`, `static_text` |

---

## Element Config Schemas

Each element type's `Config` JSON blob follows a specific schema. Common properties appear across multiple types.

### Common Formatting Properties (embedded in type-specific configs)

```json
{
    "font_size": 10.0,
    "font_style": "regular",       // "regular" | "bold" | "italic" | "bold_italic"
    "alignment": "left",           // "left" | "center" | "right"
    "space_before": 0.0,           // points
    "space_after": 0.0             // points
}
```

### Type-Specific Config Schemas

**`profile_header`**:
```json
{
    "name_font_size": 18.0,
    "detail_font_size": 10.0,
    "alignment": "center",
    "link_separator": " | ",
    "show_links": true,
    "show_links_inline": false,
    "space_after": 6.0
}
```

**`role_descriptors`**:
```json
{
    "font_size": 10.0,
    "font_style": "regular",
    "alignment": "center",
    "separator": " | ",
    "space_after": 14.0
}
```

**`professional_summary`**:
```json
{
    "font_size": 10.0,
    "bullet_char": "\u2022",
    "space_before": 0.0,
    "space_after": 0.0
}
```

**`work_history_loop`**:
```json
{
    "entry_gap": 4.0,
    "space_before": 0.0,
    "space_after": 0.0
}
```

**`work_title`**:
```json
{
    "font_size": 10.0,
    "font_style": "bold",
    "include_employer": true,
    "employer_separator": " \u2014 ",
    "employer_font_style": "italic",
    "space_after": 13.0
}
```

**`work_dates`**:
```json
{
    "font_size": 9.0,
    "alignment": "right"
}
```

**`work_bullets`** / **`work_outcomes`**:
```json
{
    "font_size": 10.0,
    "font_style": "regular",
    "bullet_char": "\u2022",
    "indent": 12.0,
    "bullet_sym_width": 10.0,
    "outcomes_label": "Outcomes:",
    "outcomes_gap": 2.0
}
```

**`skills`**:
```json
{
    "font_size": 10.0,
    "group_by_category": true,
    "include_legacy": true,
    "legacy_suffix": " (Legacy)",
    "category_font_style": "bold",
    "skill_separator": ", "
}
```

**`education_loop`** / **`certifications_loop`**:
```json
{
    "entry_gap": 0.0,
    "space_before": 0.0,
    "space_after": 0.0
}
```

**`section_heading`**:
```json
{
    "text": "SECTION TITLE",
    "font_size": 12.0,
    "font_style": "bold",
    "uppercase": true,
    "underline": true,
    "underline_weight": 0.5,
    "space_before": 10.0,
    "space_after": 4.0
}
```

**`horizontal_rule`**:
```json
{
    "weight": 0.5,
    "space_before": 0.0,
    "space_after": 6.0
}
```

**`spacer`**:
```json
{
    "height": 10.0
}
```

**`static_text`**:
```json
{
    "text": "",
    "font_size": 10.0,
    "font_style": "regular",
    "alignment": "left",
    "space_before": 0.0,
    "space_after": 0.0
}
```

**`body_text`** (cover letter):
```json
{
    "text": "Body text with {{company_name}} and {{prompt: Why this role?}} placeholders",
    "font_size": 10.0,
    "alignment": "left",
    "space_before": 0.0,
    "space_after": 13.0
}
```

**`date`** (cover letter):
```json
{
    "format": "January 2, 2006",
    "font_size": 10.0,
    "alignment": "left",
    "space_after": 13.0
}
```

**`greeting`** (cover letter):
```json
{
    "text": "Dear Hiring Manager,",
    "font_size": 10.0,
    "space_after": 13.0
}
```

**`closing`** (cover letter):
```json
{
    "text": "Sincerely,",
    "font_size": 10.0,
    "show_name": true,
    "space_before": 26.0,
    "space_after": 0.0
}
```

**`recipient_address`** (cover letter):
```json
{
    "fields": ["{{recipient_name}}", "{{company_name}}", "{{company_address}}"],
    "font_size": 10.0,
    "space_before": 0.0,
    "space_after": 13.0
}
```

---

## Database Schema (Migration v7)

```sql
CREATE TABLE document_template (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    template_type TEXT NOT NULL CHECK (template_type IN ('resume', 'cover_letter')),
    is_builtin INTEGER NOT NULL DEFAULT 0 CHECK (is_builtin IN (0, 1)),
    margin_top REAL NOT NULL DEFAULT 54.0,
    margin_bottom REAL NOT NULL DEFAULT 54.0,
    margin_left REAL NOT NULL DEFAULT 72.0,
    margin_right REAL NOT NULL DEFAULT 72.0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE template_element (
    id INTEGER PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES document_template(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES template_element(id) ON DELETE CASCADE,
    element_type TEXT NOT NULL,
    config TEXT NOT NULL DEFAULT '{}',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX idx_template_element_template ON template_element(template_id);
CREATE INDEX idx_template_element_parent ON template_element(parent_id);
```

### Built-in Template Seeds (in v7 migration)

The migration inserts 4 built-in templates and their elements:

1. **Professional** (resume): Centered header, pipe separators, underlined section headings, standard bullet chars
2. **Modern** (resume): Left-aligned header, middle-dot separators, no underlines, 22pt name, italic descriptors
3. **Formal** (cover letter): Profile header, date, recipient address block, greeting, structured body paragraphs, formal closing
4. **Casual** (cover letter): Profile header, greeting, conversational body, informal closing

Element configurations for Professional and Modern are derived from the exact formatting values documented in `research.md` to ensure pixel-perfect reproduction of current hardcoded output.

---

## Relationship to Existing Entities

### ExportData Extension

```go
type ExportData struct {
    // ... existing fields ...
    Templates []TemplateDetail `json:"templates"` // NEW
}
```

### RenderResumeRequest Evolution

The `RenderResumeRequest.TemplateID` field (currently `string`) will change to reference a `DocumentTemplate` by ID. The renderer will accept a `TemplateDetail` containing the full element tree rather than looking up a function by string key.

```go
type RenderResumeRequest struct {
    Template           DocumentTemplate   // was: TemplateID string
    Elements           []TemplateElement  // ordered element tree
    OutputDir          string
    Profile            UserProfile
    Links              []ProfileLink
    Summaries          []ProfessionalSummary
    MasterSummaryID    *int64
    WorkHistory        []WorkHistoryEntry
    Skills             []Skill
    SkillCategoryNames map[int64]string
    Academics          []AcademicCredential
    Certs              []Certification
    Descriptors        []RoleDescriptor
    CoreExpertise      []CoreExpertise
}
```

### ResumeExport Migration

The existing `resume_export.template_id TEXT` column references string IDs ("professional", "modern"). Migration v7 will:
1. Add a `template_ref_id INTEGER` column (nullable, FK to `document_template.id`)
2. Populate `template_ref_id` by matching existing `template_id` text values to seeded built-in template IDs
3. Keep `template_id TEXT` for backward compatibility (FR-050)

---

## Validation Rules

### Template Validation
- `Name` must be non-empty (max 100 chars)
- `TemplateType` must be `"resume"` or `"cover_letter"`
- `MarginTop/Bottom/Left/Right` must be >= 0 and <= 288 (4 inches)
- Built-in templates cannot be deleted or renamed

### Element Validation
- `ElementType` must be a recognized constant
- Element type must be compatible with the template's `TemplateType`
- If `ParentID` is set, the parent must exist and must be a loop container type
- If `ParentID` is set, the parent's own `ParentID` must be nil (single-level nesting)
- Element type must be a valid child of the parent's loop type (see table above)
- `Config` must be valid JSON; type-specific validation per schema above

### Store Interface Methods

```go
// --- Document Templates ---

ListDocumentTemplates(ctx context.Context) ([]DocumentTemplate, error)
GetDocumentTemplate(ctx context.Context, id int64) (TemplateDetail, error)
CreateDocumentTemplate(ctx context.Context, input DocumentTemplateInput) (DocumentTemplate, error)
UpdateDocumentTemplate(ctx context.Context, id int64, input DocumentTemplateInput) (DocumentTemplate, error)
DeleteDocumentTemplate(ctx context.Context, id int64) error
DuplicateDocumentTemplate(ctx context.Context, id int64, newName string) (DocumentTemplate, error)

// --- Template Elements ---

CreateTemplateElement(ctx context.Context, templateID int64, input TemplateElementInput) (TemplateElement, error)
UpdateTemplateElement(ctx context.Context, id int64, input TemplateElementInput) (TemplateElement, error)
DeleteTemplateElement(ctx context.Context, id int64) error
ReorderTemplateElements(ctx context.Context, templateID int64, parentID *int64, orderedIDs []int64) error
```
