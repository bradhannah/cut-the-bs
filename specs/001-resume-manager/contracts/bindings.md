# Wails Binding API Contracts

**Date**: 2026-02-19 | **Branch**: `001-resume-manager`

This document defines the Go binding methods exposed to the Svelte
frontend via Wails v2. Each public method on a bound struct becomes
a JavaScript Promise-returning function.

## Conventions

- All methods return `(result, error)` — errors reject the JS Promise
- All entity IDs are `int64`
- All timestamps are ISO 8601 strings
- Input validation occurs in the service layer, not the binding layer
- Binding methods are thin wrappers that delegate to service layer

---

## UserProfile Bindings

```go
// GetProfile returns the user's profile. Creates a default empty
// profile if none exists.
func (a *App) GetProfile() (UserProfile, error)

// UpdateProfile updates the user's profile fields.
// Returns the updated profile.
func (a *App) UpdateProfile(profile UserProfile) (UserProfile, error)
```

### Types

```go
type UserProfile struct {
    ID       int64  `json:"id"`
    FullName string `json:"full_name"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    Location string `json:"location"`
}
```

**FR Coverage**: FR-028, FR-029

---

## ProfileLink Bindings

```go
// ListProfileLinks returns all profile links ordered by sort_order.
func (a *App) ListProfileLinks() ([]ProfileLink, error)

// CreateProfileLink creates a new profile link.
func (a *App) CreateProfileLink(input ProfileLinkInput) (ProfileLink, error)

// UpdateProfileLink updates an existing profile link.
func (a *App) UpdateProfileLink(id int64, input ProfileLinkInput) (ProfileLink, error)

// DeleteProfileLink deletes a profile link.
func (a *App) DeleteProfileLink(id int64) error

// ReorderProfileLinks updates the sort_order of all profile links.
func (a *App) ReorderProfileLinks(orderedIDs []int64) error
```

### Types

```go
type ProfileLinkInput struct {
    Label string `json:"label"`
    URL   string `json:"url"`
}

type ProfileLink struct {
    ID        int64  `json:"id"`
    Label     string `json:"label"`
    URL       string `json:"url"`
    SortOrder int    `json:"sort_order"`
}
```

**FR Coverage**: FR-043

---

## WorkHistory Bindings

```go
// ListWorkHistory returns all work history entries ordered by
// sort_order, each with its achievement bullets.
func (a *App) ListWorkHistory() ([]WorkHistoryEntry, error)

// CreateWorkHistory creates a new work history entry.
// Returns the created entry with generated ID.
func (a *App) CreateWorkHistory(entry WorkHistoryInput) (WorkHistoryEntry, error)

// UpdateWorkHistory updates an existing work history entry.
func (a *App) UpdateWorkHistory(id int64, entry WorkHistoryInput) (WorkHistoryEntry, error)

// DeleteWorkHistory deletes a work history entry and its bullets.
func (a *App) DeleteWorkHistory(id int64) error

// ReorderWorkHistory updates the sort_order of all entries.
// Accepts a slice of IDs in the desired order.
func (a *App) ReorderWorkHistory(orderedIDs []int64) error

// CreateBullet adds an achievement bullet to a work history entry.
func (a *App) CreateBullet(workHistoryID int64, text string) (AchievementBullet, error)

// UpdateBullet updates an achievement bullet's text.
func (a *App) UpdateBullet(id int64, text string) (AchievementBullet, error)

// DeleteBullet deletes an achievement bullet.
func (a *App) DeleteBullet(id int64) error

// ReorderBullets updates the sort_order of bullets within an entry.
func (a *App) ReorderBullets(workHistoryID int64, orderedIDs []int64) error

// SplitBulletText accepts a block of text and returns individual
// lines for preview before creating bullets.
func (a *App) SplitBulletText(text string) []string
```

### Types

```go
type WorkHistoryInput struct {
    EmployerName         string `json:"employer_name"`
    JobTitle             string `json:"job_title"`
    StartDate            string `json:"start_date"`
    EndDate              string `json:"end_date"`       // empty = "present"
    DateGranularityStart string `json:"date_granularity_start"` // "year"|"month"|"day"
    DateGranularityEnd   string `json:"date_granularity_end"`
}

type WorkHistoryEntry struct {
    ID                   int64              `json:"id"`
    EmployerName         string             `json:"employer_name"`
    JobTitle             string             `json:"job_title"`
    StartDate            string             `json:"start_date"`
    EndDate              string             `json:"end_date"`
    DateGranularityStart string             `json:"date_granularity_start"`
    DateGranularityEnd   string             `json:"date_granularity_end"`
    SortOrder            int                `json:"sort_order"`
    Bullets              []AchievementBullet `json:"bullets"`
}

type AchievementBullet struct {
    ID            int64  `json:"id"`
    WorkHistoryID int64  `json:"work_history_id"`
    Text          string `json:"text"`
    SortOrder     int    `json:"sort_order"`
}
```

**FR Coverage**: FR-001, FR-002, FR-003, FR-004, FR-005, FR-034

---

## Skills Bindings

```go
// ListSkills returns all skills sorted by competence level (desc),
// then alphabetically.
func (a *App) ListSkills() ([]Skill, error)

// ListSkillsByCategory returns skills grouped by category, with
// categories ordered by their sort_order.
func (a *App) ListSkillsByCategory() ([]SkillCategoryWithSkills, error)

// CreateSkill creates a new skill.
func (a *App) CreateSkill(skill SkillInput) (Skill, error)

// UpdateSkill updates an existing skill.
func (a *App) UpdateSkill(id int64, skill SkillInput) (Skill, error)

// DeleteSkill deletes a skill. If the skill is referenced by any
// lenses, warns the user with affected lens names before proceeding
// (FR-050).
func (a *App) DeleteSkill(id int64) error

// CheckSkillLensReferences returns the names of lenses that
// reference a given skill, for use in delete confirmation dialogs.
func (a *App) CheckSkillLensReferences(id int64) ([]string, error)

// SplitSkillsText accepts a comma-separated string and returns
// individual skill names for preview before creating.
func (a *App) SplitSkillsText(text string) []string

// GetCompetenceLevels returns the fixed competence scale with
// descriptive criteria for each level.
func (a *App) GetCompetenceLevels() []CompetenceLevel
```

### Types

```go
type SkillInput struct {
    Name            string `json:"name"`
    CategoryID      int64  `json:"category_id"`
    CompetenceLevel int    `json:"competence_level"`
    IsLegacy        bool   `json:"is_legacy"`
}

type Skill struct {
    ID              int64  `json:"id"`
    Name            string `json:"name"`
    CategoryID      int64  `json:"category_id"`
    CompetenceLevel int    `json:"competence_level"`
    IsLegacy        bool   `json:"is_legacy"`
}

type SkillCategoryWithSkills struct {
    Category SkillCategory `json:"category"`
    Skills   []Skill       `json:"skills"`
}

type CompetenceLevel struct {
    Level       int    `json:"level"`
    Label       string `json:"label"`
    Description string `json:"description"`
}
```

**FR Coverage**: FR-006, FR-007, FR-008, FR-030, FR-031, FR-035, FR-049, FR-050

---

## Skill Category Bindings

```go
// ListSkillCategories returns all skill categories ordered by
// sort_order.
func (a *App) ListSkillCategories() ([]SkillCategory, error)

// CreateSkillCategory creates a new skill category. It is
// automatically appended to the end of the sort order.
func (a *App) CreateSkillCategory(name string) (SkillCategory, error)

// RenameSkillCategory updates a category's name. All skills
// referencing this category reflect the change automatically via FK.
func (a *App) RenameSkillCategory(id int64, name string) (SkillCategory, error)

// DeleteSkillCategory deletes a skill category. Fails if any skills
// still reference it — user must reassign or delete skills first.
func (a *App) DeleteSkillCategory(id int64) error

// ReorderSkillCategories updates the sort_order of all categories.
// Accepts a slice of IDs in the desired display order.
func (a *App) ReorderSkillCategories(orderedIDs []int64) error
```

### Types

```go
type SkillCategory struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    SortOrder int    `json:"sort_order"`
}
```

**FR Coverage**: FR-006, FR-049

---

## Academic & Certification Bindings

```go
// ListAcademicCredentials returns all academic records ordered by
// sort_order.
func (a *App) ListAcademicCredentials() ([]AcademicCredential, error)

// CreateAcademicCredential creates a new academic record.
func (a *App) CreateAcademicCredential(cred AcademicInput) (AcademicCredential, error)

// UpdateAcademicCredential updates an academic record.
func (a *App) UpdateAcademicCredential(id int64, cred AcademicInput) (AcademicCredential, error)

// DeleteAcademicCredential deletes an academic record.
func (a *App) DeleteAcademicCredential(id int64) error

// ListCertifications returns all certifications with computed
// active/inactive status.
func (a *App) ListCertifications() ([]Certification, error)

// CreateCertification creates a new certification.
func (a *App) CreateCertification(cert CertificationInput) (Certification, error)

// UpdateCertification updates a certification.
func (a *App) UpdateCertification(id int64, cert CertificationInput) (Certification, error)

// DeleteCertification deletes a certification.
func (a *App) DeleteCertification(id int64) error
```

### Types

```go
type AcademicInput struct {
    Institution    string `json:"institution"`
    CredentialType string `json:"credential_type"`
    FieldOfStudy   string `json:"field_of_study"`
    CompletionDate string `json:"completion_date"`
    DateGranularity string `json:"date_granularity"`
}

type AcademicCredential struct {
    ID              int64  `json:"id"`
    Institution     string `json:"institution"`
    CredentialType  string `json:"credential_type"`
    FieldOfStudy    string `json:"field_of_study"`
    CompletionDate  string `json:"completion_date"`
    DateGranularity string `json:"date_granularity"`
    SortOrder       int    `json:"sort_order"`
}

type CertificationInput struct {
    Name           string `json:"name"`
    IssuingBody    string `json:"issuing_body"`
    DateEarned     string `json:"date_earned"`
    ExpirationDate string `json:"expiration_date"` // empty = no expiration
}

type Certification struct {
    ID             int64  `json:"id"`
    Name           string `json:"name"`
    IssuingBody    string `json:"issuing_body"`
    DateEarned     string `json:"date_earned"`
    ExpirationDate string `json:"expiration_date"`
    IsActive       bool   `json:"is_active"` // computed
    SortOrder      int    `json:"sort_order"`
}
```

**FR Coverage**: FR-009, FR-010, FR-011

---

## Professional Summary Bindings

```go
// ListSummaries returns all professional summary variants.
func (a *App) ListSummaries() ([]ProfessionalSummary, error)

// CreateSummary creates a new summary variant.
func (a *App) CreateSummary(summary SummaryInput) (ProfessionalSummary, error)

// UpdateSummary updates an existing summary.
func (a *App) UpdateSummary(id int64, summary SummaryInput) (ProfessionalSummary, error)

// DeleteSummary deletes a summary variant.
func (a *App) DeleteSummary(id int64) error
```

### Types

```go
type SummaryInput struct {
    Label    string `json:"label"`
    BodyText string `json:"body_text"`
}

type ProfessionalSummary struct {
    ID       int64  `json:"id"`
    Label    string `json:"label"`
    BodyText string `json:"body_text"`
}
```

**FR Coverage**: FR-012, FR-013

---

## Role Descriptor Bindings

```go
// ListDescriptors returns all role descriptors ordered by sort_order.
func (a *App) ListDescriptors() ([]RoleDescriptor, error)

// CreateDescriptor creates a new role descriptor.
func (a *App) CreateDescriptor(title string) (RoleDescriptor, error)

// UpdateDescriptor updates a descriptor's title.
func (a *App) UpdateDescriptor(id int64, title string) (RoleDescriptor, error)

// DeleteDescriptor deletes a role descriptor.
func (a *App) DeleteDescriptor(id int64) error

// ReorderDescriptors updates sort_order for all descriptors.
func (a *App) ReorderDescriptors(orderedIDs []int64) error
```

### Types

```go
type RoleDescriptor struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    SortOrder int    `json:"sort_order"`
}
```

**FR Coverage**: FR-019, FR-020, FR-021

---

## Lens Bindings

```go
// ListLenses returns all lenses.
func (a *App) ListLenses() ([]Lens, error)

// GetLens returns a single lens with all its content selections.
func (a *App) GetLens(id int64) (LensDetail, error)

// CreateLens creates a new lens. Returns the created lens.
func (a *App) CreateLens(input LensInput) (Lens, error)

// UpdateLens updates a lens's name and summary selection.
func (a *App) UpdateLens(id int64, input LensInput) (Lens, error)

// DeleteLens deletes a lens and all its selections.
func (a *App) DeleteLens(id int64) error

// SetLensWorkHistory replaces the work history selections for a lens.
func (a *App) SetLensWorkHistory(lensID int64, selections []LensWorkHistoryItem) error

// SetLensBullets replaces the bullet selections for a lens.
func (a *App) SetLensBullets(lensID int64, selections []LensBulletItem) error

// SetLensSkills replaces the skill selections for a lens.
func (a *App) SetLensSkills(lensID int64, selections []LensSkillItem) error

// SetLensAcademics replaces the academic selections for a lens.
func (a *App) SetLensAcademics(lensID int64, academicIDs []int64) error

// SetLensCerts replaces the certification selections for a lens.
func (a *App) SetLensCerts(lensID int64, certIDs []int64) error

// SetLensDescriptors replaces the descriptor selections for a lens.
func (a *App) SetLensDescriptors(lensID int64, selections []LensDescriptorItem) error

// GetLensExportSelections returns the full content selections for a
// lens, formatted as an ExportRequest that can be used to pre-fill
// the export dialog. The user can then override before generating.
func (a *App) GetLensExportSelections(lensID int64) (ExportRequest, error)
```

### Types

```go
type LensInput struct {
    Name      string `json:"name"`
    SummaryID *int64 `json:"summary_id"`
}

type Lens struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    SummaryID *int64 `json:"summary_id"`
}

type LensDetail struct {
    Lens
    WorkHistory []LensWorkHistoryItem `json:"work_history"`
    Bullets     []LensBulletItem      `json:"bullets"`
    Skills      []LensSkillItem       `json:"skills"`
    AcademicIDs []int64               `json:"academic_ids"`
    CertIDs     []int64               `json:"cert_ids"`
    Descriptors []LensDescriptorItem  `json:"descriptors"`
}

type LensWorkHistoryItem struct {
    WorkHistoryID int64 `json:"work_history_id"`
    SortOrder     int   `json:"sort_order"`
}

type LensBulletItem struct {
    BulletID  int64 `json:"bullet_id"`
    SortOrder int   `json:"sort_order"`
}

type LensSkillItem struct {
    SkillID         int64 `json:"skill_id"`
    CustomSortOrder *int  `json:"custom_sort_order"` // nil = default
}

type LensDescriptorItem struct {
    DescriptorID int64 `json:"descriptor_id"`
    SortOrder    int   `json:"sort_order"`
}
```

**FR Coverage**: FR-044, FR-045, FR-046

---

## Skill Lens Tag Bindings

```go
// GetSkillLensTags returns all lens tags for a skill.
func (a *App) GetSkillLensTags(skillID int64) ([]int64, error)

// SetSkillLensTags replaces all lens tags for a skill.
// Pass an empty slice to remove all tags (skill won't be
// auto-included by any lens).
func (a *App) SetSkillLensTags(skillID int64, lensIDs []int64) error

// ListSkillsWithLensTags returns all skills with their lens tag
// associations included.
func (a *App) ListSkillsWithLensTags() ([]SkillWithTags, error)
```

### Types

```go
type SkillWithTags struct {
    Skill
    LensIDs []int64 `json:"lens_ids"`
}
```

**FR Coverage**: FR-047

---

## Resume Export Bindings

```go
// ListTemplates returns available built-in resume templates.
func (a *App) ListTemplates() []ResumeTemplate

// PreviewExport generates a resume with the given selections and
// returns the file path without creating an export record.
// Used for template preview before committing.
func (a *App) PreviewExport(req ExportRequest) (string, error)

// CreateExport generates a PDF resume, saves it, and creates an
// export record. Returns the export record with file path.
func (a *App) CreateExport(req ExportRequest) (ResumeExport, error)

// ListExports returns all previous exports.
func (a *App) ListExports() ([]ResumeExport, error)

// OpenExportFile opens the PDF file in the system default viewer.
func (a *App) OpenExportFile(exportID int64) error
```

### Types

```go
type ResumeTemplate struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    PreviewURL  string `json:"preview_url"` // thumbnail path
}

type ExportRequest struct {
    TemplateID      string  `json:"template_id"`
    LensID          *int64  `json:"lens_id"`          // nil = no lens used
    SummaryID       *int64  `json:"summary_id"`       // nil = no summary
    WorkHistoryIDs  []int64 `json:"work_history_ids"`
    BulletIDs       []int64 `json:"bullet_ids"`
    SkillIDs        []int64 `json:"skill_ids"`
    SkillSortOverrides map[int64]int `json:"skill_sort_overrides"` // skillID → custom order
    AcademicIDs     []int64 `json:"academic_ids"`
    CertificationIDs []int64 `json:"certification_ids"`
    DescriptorIDs   []int64 `json:"descriptor_ids"`
}

type ResumeExport struct {
    ID          int64  `json:"id"`
    TemplateID  string `json:"template_id"`
    FilePath    string `json:"file_path"`
    SummaryID   *int64 `json:"summary_id"`
    LensID      *int64 `json:"lens_id"`
    GeneratedAt string `json:"generated_at"`
}
```

**FR Coverage**: FR-014, FR-015, FR-016, FR-017, FR-018

---

## Cover Letter Bindings

```go
// ListCoverLetters returns all cover letters.
func (a *App) ListCoverLetters() ([]CoverLetter, error)

// CreateCoverLetter creates a new cover letter.
func (a *App) CreateCoverLetter(input CoverLetterInput) (CoverLetter, error)

// UpdateCoverLetter updates a cover letter's content.
func (a *App) UpdateCoverLetter(id int64, input CoverLetterInput) (CoverLetter, error)

// DeleteCoverLetter deletes a cover letter.
func (a *App) DeleteCoverLetter(id int64) error

// ExportCoverLetter generates a PDF of the cover letter.
func (a *App) ExportCoverLetter(id int64) (string, error)
```

### Types

```go
type CoverLetterInput struct {
    Title    string `json:"title"`
    BodyText string `json:"body_text"`
}

type CoverLetter struct {
    ID       int64  `json:"id"`
    Title    string `json:"title"`
    BodyText string `json:"body_text"`
    FilePath string `json:"file_path"`
}
```

**FR Coverage**: FR-026, FR-027

---

## Job Application Bindings

```go
// ListApplications returns all job applications with current status.
func (a *App) ListApplications() ([]JobApplication, error)

// SearchApplications searches by company name or position title.
func (a *App) SearchApplications(query string) ([]JobApplication, error)

// CreateApplication creates a new job application record.
func (a *App) CreateApplication(input ApplicationInput) (JobApplication, error)

// UpdateApplication updates application fields (not status).
func (a *App) UpdateApplication(id int64, input ApplicationInput) (JobApplication, error)

// UpdateApplicationStatus changes the status and records history.
func (a *App) UpdateApplicationStatus(id int64, newStatus string) (JobApplication, error)

// UpdateApplicationFit updates the fit indicator.
func (a *App) UpdateApplicationFit(id int64, fitIndicator string) (JobApplication, error)

// GetApplicationHistory returns the status change history for an
// application.
func (a *App) GetApplicationHistory(id int64) ([]StatusChange, error)

// DeleteApplication deletes a job application and its history.
func (a *App) DeleteApplication(id int64) error

// GetApplicationStatuses returns the fixed list of valid statuses.
func (a *App) GetApplicationStatuses() []string

// GetFitIndicators returns the fixed list of fit indicator values.
func (a *App) GetFitIndicators() []string
```

### Types

```go
type ApplicationInput struct {
    CompanyName    string `json:"company_name"`
    PositionTitle  string `json:"position_title"`
    DateApplied    string `json:"date_applied"`
    FitIndicator   string `json:"fit_indicator"`
    ResumeExportID *int64 `json:"resume_export_id"`
    CoverLetterID  *int64 `json:"cover_letter_id"`
    Notes          string `json:"notes"`
}

type JobApplication struct {
    ID             int64  `json:"id"`
    CompanyName    string `json:"company_name"`
    PositionTitle  string `json:"position_title"`
    DateApplied    string `json:"date_applied"`
    Status         string `json:"status"`
    FitIndicator   string `json:"fit_indicator"`
    ResumeExportID *int64 `json:"resume_export_id"`
    CoverLetterID  *int64 `json:"cover_letter_id"`
    Notes          string `json:"notes"`
}

type StatusChange struct {
    ID            int64  `json:"id"`
    FromStatus    string `json:"from_status"`
    ToStatus      string `json:"to_status"`
    ChangedAt     string `json:"changed_at"`
}
```

**FR Coverage**: FR-022, FR-023, FR-024, FR-025, FR-032

---

## Data Management Bindings

```go
// ExportAllData exports all user data to a JSON file at the specified
// path. Returns the file path.
func (a *App) ExportAllData(outputPath string) (string, error)

// ImportAllData restores all user data from a JSON backup file.
// This replaces all existing data.
func (a *App) ImportAllData(inputPath string) error

// ImportCSV imports structured data from a CSV file. The dataType
// parameter specifies what is being imported: "work_history",
// "skills", "academic", "certifications".
func (a *App) ImportCSV(filePath string, dataType string) (ImportResult, error)

// ImportJSON imports structured data from a JSON file (partial, not
// full backup restore).
func (a *App) ImportJSON(filePath string, dataType string) (ImportResult, error)

// GetDataDirectory returns the current data directory path.
func (a *App) GetDataDirectory() string

// SetDataDirectory changes the data directory. Requires app restart.
func (a *App) SetDataDirectory(path string) error

// GetBackupSettings returns current backup configuration.
func (a *App) GetBackupSettings() BackupSettings

// UpdateBackupSettings updates backup configuration.
func (a *App) UpdateBackupSettings(settings BackupSettings) error
```

### Types

```go
type ImportResult struct {
    RecordsImported int      `json:"records_imported"`
    RecordsSkipped  int      `json:"records_skipped"`
    Errors          []string `json:"errors"`
}

type BackupSettings struct {
    RollingBackupCount int `json:"rolling_backup_count"`
}
```

**FR Coverage**: FR-033, FR-036, FR-037, FR-038, FR-039, FR-040, FR-041, FR-042

---

## Events (Go → Frontend)

These events are emitted from the Go backend to notify the frontend
of background operations:

| Event Name | Payload | Description |
|------------|---------|-------------|
| `autosave:complete` | `{timestamp: string}` | Autosave completed |
| `backup:complete` | `{timestamp: string, path: string}` | Rolling backup created |
| `backup:error` | `{error: string}` | Backup failed |
| `import:progress` | `{current: int, total: int}` | Import progress |

---

## Frontend-Only Concerns (No Backend Bindings)

### Zoom Widget

The UI includes a zoom control widget in the bottom-left corner of
the application window. Default zoom is 100%. The widget provides +/-
buttons and supports Cmd/Ctrl +/- keyboard shortcuts. This is a pure
frontend concern — zoom is applied via CSS transform on the root
container. No backend bindings or entity changes are needed.

**FR Coverage**: FR-048
