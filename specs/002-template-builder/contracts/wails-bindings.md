# API Contracts: Template Builder

**Phase 1 output** | **Branch**: `002-template-builder` | **Date**: 2026-02-21

This project uses **Wails v2 bindings** (not REST/HTTP). All API methods are public methods on the `App` struct in `app.go`, automatically exposed to the Svelte frontend as `window.go.main.App.MethodName()` returning `Promise<T>`. The frontend calls them via the existing `call<T>(method, ...args)` wrapper in `frontend/src/services/api.ts`.

---

## Template Management Bindings

### ListDocumentTemplates

Returns all templates (both built-in and user-created).

```go
// ListDocumentTemplates returns all document templates ordered by
// is_builtin DESC, name ASC. Built-in templates appear first.
func (a *App) ListDocumentTemplates() ([]domain.DocumentTemplate, error)
```

**Frontend call**: `call<DocumentTemplate[]>('ListDocumentTemplates')`
**Returns**: Array of `DocumentTemplate` objects (without elements).

---

### GetDocumentTemplate

Returns a single template with all its elements (full detail).

```go
// GetDocumentTemplate returns a template with all its elements.
func (a *App) GetDocumentTemplate(id int64) (domain.TemplateDetail, error)
```

**Frontend call**: `call<TemplateDetail>('GetDocumentTemplate', id)`
**Returns**: `TemplateDetail` (template + flat element list ordered by sort_order).
**Error**: Template not found.

---

### CreateDocumentTemplate

Creates a new user template.

```go
// CreateDocumentTemplate creates a new user document template.
func (a *App) CreateDocumentTemplate(input domain.DocumentTemplateInput) (domain.DocumentTemplate, error)
```

**Frontend call**: `call<DocumentTemplate>('CreateDocumentTemplate', input)`
**Input**:
```typescript
interface DocumentTemplateInput {
    name: string;           // required, non-empty
    description: string;
    template_type: string;  // "resume" | "cover_letter"
    margin_top: number;     // points, default 54.0
    margin_bottom: number;  // points, default 54.0
    margin_left: number;    // points, default 72.0
    margin_right: number;   // points, default 72.0
}
```
**Returns**: Created template with generated ID and timestamps.
**Error**: Validation failure (empty name, invalid type, invalid margins).

---

### UpdateDocumentTemplate

Updates a user template's metadata (name, description, margins). Cannot update built-in templates.

```go
// UpdateDocumentTemplate updates a user template's metadata.
func (a *App) UpdateDocumentTemplate(id int64, input domain.DocumentTemplateInput) (domain.DocumentTemplate, error)
```

**Frontend call**: `call<DocumentTemplate>('UpdateDocumentTemplate', id, input)`
**Error**: Template not found, template is built-in, validation failure.

---

### DeleteDocumentTemplate

Deletes a user-created template and all its elements (CASCADE).

```go
// DeleteDocumentTemplate deletes a user template. Built-in
// templates cannot be deleted.
func (a *App) DeleteDocumentTemplate(id int64) error
```

**Frontend call**: `call<void>('DeleteDocumentTemplate', id)`
**Error**: Template not found, template is built-in.

---

### DuplicateDocumentTemplate

Creates a copy of any template (built-in or user-created) with a new name.

```go
// DuplicateDocumentTemplate creates a copy of a template with all
// its elements. The copy is always user-created (is_builtin=false).
func (a *App) DuplicateDocumentTemplate(id int64, newName string) (domain.DocumentTemplate, error)
```

**Frontend call**: `call<DocumentTemplate>('DuplicateDocumentTemplate', id, newName)`
**Returns**: The newly created template (without elements; call `GetDocumentTemplate` for full detail).
**Error**: Source template not found, empty name.

---

## Template Element Bindings

### CreateTemplateElement

Adds a new element to a template.

```go
// CreateTemplateElement adds a new element to a template, appended
// to the end of the sort order within its parent scope.
func (a *App) CreateTemplateElement(templateID int64, input domain.TemplateElementInput) (domain.TemplateElement, error)
```

**Frontend call**: `call<TemplateElement>('CreateTemplateElement', templateID, input)`
**Input**:
```typescript
interface TemplateElementInput {
    parent_id: number | null;  // null for top-level, set for loop children
    element_type: string;       // element type constant
    config: string;             // JSON blob of type-specific properties
}
```
**Returns**: Created element with generated ID, sort_order, timestamps.
**Error**: Template not found, invalid element type, invalid parent, nesting violation.

---

### UpdateTemplateElement

Updates an element's config (and optionally its parent/type).

```go
// UpdateTemplateElement updates an element's type and config.
func (a *App) UpdateTemplateElement(id int64, input domain.TemplateElementInput) (domain.TemplateElement, error)
```

**Frontend call**: `call<TemplateElement>('UpdateTemplateElement', id, input)`
**Error**: Element not found, validation failure.

---

### DeleteTemplateElement

Deletes an element (and its children if it's a loop container, via CASCADE).

```go
// DeleteTemplateElement removes an element from a template.
func (a *App) DeleteTemplateElement(id int64) error
```

**Frontend call**: `call<void>('DeleteTemplateElement', id)`
**Error**: Element not found.

---

### ReorderTemplateElements

Sets the sort order for elements within a parent scope.

```go
// ReorderTemplateElements updates sort_order for elements within
// a specific parent scope (top-level if parentID is nil, or within
// a specific loop container).
func (a *App) ReorderTemplateElements(templateID int64, parentID *int64, orderedIDs []int64) error
```

**Frontend call**: `call<void>('ReorderTemplateElements', templateID, parentID, orderedIDs)`
**Error**: Template not found, ID mismatch.

---

## Template Export/Import Bindings

### ExportTemplate

Exports a single template as a standalone JSON file for sharing.

```go
// ExportTemplate exports a single template (with all elements) to
// a JSON file at the specified path.
func (a *App) ExportTemplate(id int64, outputPath string) error
```

**Frontend call**: `call<void>('ExportTemplate', id, outputPath)`
**Error**: Template not found, file write error.

---

### ImportTemplate

Imports a template from a standalone JSON file.

```go
// ImportTemplate imports a template from a JSON file. The imported
// template is always user-created (is_builtin=false) with a new ID.
func (a *App) ImportTemplate(inputPath string) (domain.DocumentTemplate, error)
```

**Frontend call**: `call<DocumentTemplate>('ImportTemplate', inputPath)`
**Returns**: The newly created template.
**Error**: File read error, invalid format, validation failure.

---

## PDF Rendering Bindings (Modified)

### ExportResume (modified)

The existing `ExportResume` binding changes to accept a template ID (int64) instead of a template string slug.

```go
// ExportResume generates a resume PDF using the specified template
// and content selections.
func (a *App) ExportResume(req domain.ExportRequest) (domain.ResumeExport, error)
```

**Modified `ExportRequest`**:
```go
type ExportRequest struct {
    TemplateID         int64         `json:"template_id"`  // was: string
    // ... rest unchanged ...
}
```

---

### PreviewTemplate

Generates a preview PDF using the current template configuration and user data.

```go
// PreviewTemplate generates a preview PDF for the given template
// using the user's actual profile data and most recent lens
// selections (or placeholder data if none exist).
func (a *App) PreviewTemplate(templateID int64) (string, error)
```

**Frontend call**: `call<string>('PreviewTemplate', templateID)`
**Returns**: File path of the generated preview PDF.
**Error**: Template not found, rendering failure.

---

## Frontend TypeScript Types

These types are generated by Wails from the Go structs but are documented here for contract clarity.

```typescript
// frontend/wailsjs/go/models.ts (auto-generated)

interface DocumentTemplate {
    id: number;
    name: string;
    description: string;
    template_type: string;
    is_builtin: boolean;
    margin_top: number;
    margin_bottom: number;
    margin_left: number;
    margin_right: number;
    created_at: string;
    updated_at: string;
}

interface DocumentTemplateInput {
    name: string;
    description: string;
    template_type: string;
    margin_top: number;
    margin_bottom: number;
    margin_left: number;
    margin_right: number;
}

interface TemplateElement {
    id: number;
    template_id: number;
    parent_id: number | null;
    element_type: string;
    config: string;     // JSON blob, parsed client-side
    sort_order: number;
    created_at: string;
    updated_at: string;
}

interface TemplateElementInput {
    parent_id: number | null;
    element_type: string;
    config: string;
}

interface TemplateDetail {
    // extends DocumentTemplate fields
    id: number;
    name: string;
    description: string;
    template_type: string;
    is_builtin: boolean;
    margin_top: number;
    margin_bottom: number;
    margin_left: number;
    margin_right: number;
    created_at: string;
    updated_at: string;
    elements: TemplateElement[];
}
```

---

## Frontend API Layer Additions

New functions to add to `frontend/src/services/api.ts`:

```typescript
// Template Management
export const listDocumentTemplates = () =>
    call<DocumentTemplate[]>('ListDocumentTemplates');

export const getDocumentTemplate = (id: number) =>
    call<TemplateDetail>('GetDocumentTemplate', id);

export const createDocumentTemplate = (input: DocumentTemplateInput) =>
    call<DocumentTemplate>('CreateDocumentTemplate', input);

export const updateDocumentTemplate = (id: number, input: DocumentTemplateInput) =>
    call<DocumentTemplate>('UpdateDocumentTemplate', id, input);

export const deleteDocumentTemplate = (id: number) =>
    call<void>('DeleteDocumentTemplate', id);

export const duplicateDocumentTemplate = (id: number, newName: string) =>
    call<DocumentTemplate>('DuplicateDocumentTemplate', id, newName);

// Template Elements
export const createTemplateElement = (templateID: number, input: TemplateElementInput) =>
    call<TemplateElement>('CreateTemplateElement', templateID, input);

export const updateTemplateElement = (id: number, input: TemplateElementInput) =>
    call<TemplateElement>('UpdateTemplateElement', id, input);

export const deleteTemplateElement = (id: number) =>
    call<void>('DeleteTemplateElement', id);

export const reorderTemplateElements = (templateID: number, parentID: number | null, orderedIDs: number[]) =>
    call<void>('ReorderTemplateElements', templateID, parentID, orderedIDs);

// Template Import/Export
export const exportTemplate = (id: number, outputPath: string) =>
    call<void>('ExportTemplate', id, outputPath);

export const importTemplate = (inputPath: string) =>
    call<DocumentTemplate>('ImportTemplate', inputPath);

// Preview
export const previewTemplate = (templateID: number) =>
    call<string>('PreviewTemplate', templateID);
```
