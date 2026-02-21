# Data Model: Resume Manager

**Date**: 2026-02-19 | **Branch**: `001-resume-manager`

## Entity Relationship Overview

```
UserProfile (1)
    │
    ├── ProfileLink (many, ordered)  ← GitHub, LinkedIn, portfolio, etc.
    │
    ├── WorkHistoryEntry (many)
    │       └── AchievementBullet (many, ordered)
    │
    ├── SkillCategory (many, ordered) ← user-defined skill groupings
    │       └── Skill (many)
    │
    ├── AcademicCredential (many)
    │
    ├── Certification (many)
    │
    ├── ProfessionalSummary (many)
    │
    ├── RoleDescriptor (many)
    │
    ├── Lens (many)                  ← reusable content presets by job type
    │       ├── LensWorkHistorySelection (which entries + bullet order)
    │       ├── LensBulletSelection (which bullets per entry)
    │       ├── LensSkillSelection (which skills)
    │       ├── LensAcademicSelection
    │       ├── LensCertSelection
    │       ├── LensDescriptorSelection
    │       └── linked ProfessionalSummary
    │
    ├── SkillLensTag (many-to-many: Skill ↔ Lens)
    │
    ├── ResumeExport (many)
    │       ├── ExportSelection (snapshot of what was included)
    │       └── generated PDF file
    │
    ├── CoverLetter (many)
    │
    └── JobApplication (many)
            ├── StatusHistory (many, append-only)
            ├── linked ResumeExport (snapshot)
            └── linked CoverLetter (optional)
```

---

## Entities

### UserProfile

Single record per installation. Contact information included on all
resume exports.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | Always 1 (single user) |
| full_name | TEXT | NOT NULL | |
| email | TEXT | NOT NULL | |
| phone | TEXT | | Optional |
| location | TEXT | | City/region |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**: Email format validation. At minimum, full_name and
email are required. Links (LinkedIn, portfolio, GitHub, etc.) are
stored in the ProfileLink table (one-to-many).

**References**: FR-028, FR-029, FR-043

---

### ProfileLink

A labelled URL associated with the user's profile. Replaces the
previous fixed `linkedin_url`/`website_url` fields with a flexible
list of links. Common types: LinkedIn, GitHub, Portfolio, Personal
Website, etc.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| label | TEXT | NOT NULL | Display name, e.g., "GitHub" |
| url | TEXT | NOT NULL | Full URL |
| sort_order | INTEGER | NOT NULL | Position in the list |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**: label and url must be non-empty. url must be a valid
URL format. sort_order must be unique across all profile links.

**References**: FR-043

---

### WorkHistoryEntry

A single employment record containing an ordered list of achievement
bullets.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| employer_name | TEXT | NOT NULL | |
| job_title | TEXT | NOT NULL | |
| start_date | TEXT | NOT NULL | Flexible: "2023", "2023-03", or "2023-03-15" |
| end_date | TEXT | | NULL = "present" |
| date_granularity_start | TEXT | NOT NULL | "year", "month", or "day" |
| date_granularity_end | TEXT | | "year", "month", or "day" |
| sort_order | INTEGER | NOT NULL | Position in the list |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**:
- employer_name and job_title are required (non-empty)
- If end_date is provided, it must not be before start_date (compared
  at the coarsest granularity of the two dates)
- sort_order must be unique across all entries

**Date Granularity**: Dates are stored as text in the format that
matches their granularity. The `date_granularity_*` fields record
what precision the user entered so the display and PDF export can
format appropriately.

**References**: FR-001, FR-002, FR-003, FR-005

---

### AchievementBullet

A single accomplishment or responsibility line item. Belongs to one
WorkHistoryEntry.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| work_history_id | INTEGER | FK → WorkHistoryEntry, NOT NULL | |
| text | TEXT | NOT NULL | Plain text, system handles formatting |
| sort_order | INTEGER | NOT NULL | Position within parent entry |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**: text must be non-empty. sort_order must be unique
within the parent WorkHistoryEntry.

**References**: FR-001, FR-004, FR-034

---

### SkillCategory

A named grouping for skills with user-controlled display ordering.
Categories are user-defined (not a fixed list). Skills reference
their category by FK; renaming a category automatically applies to
all associated skills.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| name | TEXT | NOT NULL, UNIQUE | e.g., "AWS Technologies", "Technologies (Architectural)" |
| sort_order | INTEGER | NOT NULL | User-controlled display order on resume |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**: name must be non-empty and unique. New categories
appear at the end of the sort order by default.

**References**: FR-006, FR-049

---

### Skill

A single technical or professional capability with competence level
and relevancy.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| name | TEXT | NOT NULL | |
| category_id | INTEGER | FK → SkillCategory, NOT NULL | |
| competence_level | INTEGER | NOT NULL, 1-10 | Higher = more competent |
| is_legacy | INTEGER | NOT NULL, DEFAULT 0 | 0 = current, 1 = legacy |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**: name must be non-empty. competence_level must be within
the defined scale range (1-10). category_id must reference a valid
SkillCategory.

**Sort Rules**: Primary sort by competence_level descending. Secondary
sort alphabetical by name within the same level. Per-export overrides
stored separately (see ExportSkillOrder).

**PDF Rendering**: On the exported resume, skills are rendered as
comma-separated names under each category header (e.g., "AWS
Technologies: EC2, Lambda, S3, DynamoDB"). Competence level is used
only for sort ordering and is not displayed on the resume.

**Competence Scale**: 10 predefined levels stored as application
constants (not in the database). Each level has a numeric value and
descriptive criteria. The database stores only the numeric level.

**References**: FR-006, FR-007, FR-008, FR-030, FR-031, FR-035, FR-049

---

### AcademicCredential

A degree, diploma, or other academic award.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| institution | TEXT | NOT NULL | University/college name |
| credential_type | TEXT | NOT NULL | "degree", "diploma", etc. |
| field_of_study | TEXT | NOT NULL | |
| completion_date | TEXT | NOT NULL | Flexible granularity like WorkHistoryEntry |
| date_granularity | TEXT | NOT NULL | "year", "month", or "day" |
| sort_order | INTEGER | NOT NULL | |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**: institution, credential_type, and field_of_study must
be non-empty.

**References**: FR-009

---

### Certification

A professional certification with optional expiration tracking.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| name | TEXT | NOT NULL | Certification name |
| issuing_body | TEXT | NOT NULL | |
| date_earned | TEXT | NOT NULL | ISO 8601 date |
| expiration_date | TEXT | | NULL = no expiration |
| sort_order | INTEGER | NOT NULL | |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Derived Field**: `is_active` is computed at read time, not stored.
If `expiration_date` is NULL, the certification is active. If
`expiration_date` is in the past, it is inactive. This ensures status
updates automatically without background processes (FR-011).

**Validation**: name and issuing_body must be non-empty.

**References**: FR-010, FR-011

---

### ProfessionalSummary

A reusable summary block with a label for identification.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| label | TEXT | NOT NULL, UNIQUE | e.g., "Technical Leader" |
| body_text | TEXT | NOT NULL | Plain text |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**: label must be non-empty and unique. body_text must be
non-empty.

**References**: FR-012, FR-013

---

### RoleDescriptor

A short tag describing a professional role, selectable per export.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| title | TEXT | NOT NULL, UNIQUE | e.g., "Software Engineer" |
| sort_order | INTEGER | NOT NULL | |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**: title must be non-empty and unique.

**References**: FR-019, FR-020, FR-021

---

### Lens

A named, reusable content selection preset tied to a job type or role
variant (e.g., "Sales Engineer", "Solutions Architect"). A lens
pre-configures which work history entries are included, which bullets
within each entry (with per-lens ordering), which skills, which
academic credentials, which certifications, which role descriptors,
and which professional summary to use.

At export time the user picks a lens + template. The lens pre-fills
all selections but the user can override (tweak individual checkboxes)
before generating. One-off overrides do NOT modify the saved lens.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| name | TEXT | NOT NULL, UNIQUE | e.g., "Sales Engineer" |
| summary_id | INTEGER | FK → ProfessionalSummary | NULL if no summary selected |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Validation**: name must be non-empty and unique.

**References**: FR-044, FR-045, FR-046

---

### LensWorkHistorySelection

Records which work history entries are included in a lens.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| lens_id | INTEGER | FK → Lens, NOT NULL | |
| work_history_id | INTEGER | FK → WorkHistoryEntry, NOT NULL | |
| sort_order | INTEGER | NOT NULL | Per-lens ordering of entries |

**Constraint**: UNIQUE(lens_id, work_history_id)

---

### LensBulletSelection

Records which specific bullets within included work history entries
are visible in a lens, with per-lens ordering.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| lens_id | INTEGER | FK → Lens, NOT NULL | |
| bullet_id | INTEGER | FK → AchievementBullet, NOT NULL | |
| sort_order | INTEGER | NOT NULL | Per-lens bullet order within parent entry |

**Constraint**: UNIQUE(lens_id, bullet_id)

---

### LensSkillSelection

Records which skills are included in a lens.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| lens_id | INTEGER | FK → Lens, NOT NULL | |
| skill_id | INTEGER | FK → Skill, NOT NULL | |
| custom_sort_order | INTEGER | | NULL = use default competence sort |

**Constraint**: UNIQUE(lens_id, skill_id)

---

### LensAcademicSelection

Records which academic credentials are included in a lens.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| lens_id | INTEGER | FK → Lens, NOT NULL | |
| academic_id | INTEGER | FK → AcademicCredential, NOT NULL | |

**Constraint**: UNIQUE(lens_id, academic_id)

---

### LensCertSelection

Records which certifications are included in a lens.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| lens_id | INTEGER | FK → Lens, NOT NULL | |
| cert_id | INTEGER | FK → Certification, NOT NULL | |

**Constraint**: UNIQUE(lens_id, cert_id)

---

### LensDescriptorSelection

Records which role descriptors are included in a lens, with per-lens
ordering.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| lens_id | INTEGER | FK → Lens, NOT NULL | |
| descriptor_id | INTEGER | FK → RoleDescriptor, NOT NULL | |
| sort_order | INTEGER | NOT NULL | Per-lens descriptor order |

**Constraint**: UNIQUE(lens_id, descriptor_id)

---

### SkillLensTag

Many-to-many junction between Skill and Lens. Tags a skill as
belonging to one or more lenses. When a lens is applied at export
time, only skills tagged for that lens (or tagged "all" via having
no lens restrictions) are included.

A skill with NO SkillLensTag rows is treated as "untagged" and will
NOT be auto-included by any lens. A skill can be explicitly added to
a lens via LensSkillSelection regardless of tags.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| skill_id | INTEGER | FK → Skill, NOT NULL | |
| lens_id | INTEGER | FK → Lens, NOT NULL | |

**Constraint**: UNIQUE(skill_id, lens_id)

**References**: FR-047

---

### ResumeExport

A generated PDF artifact with a snapshot of what data was selected.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| template_id | TEXT | NOT NULL | Built-in template identifier |
| file_path | TEXT | NOT NULL | Path to generated PDF file |
| summary_id | INTEGER | FK → ProfessionalSummary | NULL if no summary selected |
| lens_id | INTEGER | FK → Lens | NULL if no lens was used |
| generated_at | TEXT | NOT NULL | ISO 8601 |

**Snapshot**: The export record stores the selections (which entries,
bullets, skills were included) but does NOT duplicate the source data.
For job application snapshots, the PDF file itself is the snapshot —
it is never overwritten or regenerated. If a lens was used, the
lens_id records which lens pre-filled the selections, but the actual
selections (which may have been overridden) are in the Export*Selection
tables.

**References**: FR-014, FR-015, FR-016, FR-017, FR-018

---

### ExportWorkHistorySelection

Records which work history entries and bullets were included in a
specific export.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| export_id | INTEGER | FK → ResumeExport, NOT NULL | |
| work_history_id | INTEGER | FK → WorkHistoryEntry, NOT NULL | |

---

### ExportBulletSelection

Records which specific bullets were included.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| export_id | INTEGER | FK → ResumeExport, NOT NULL | |
| bullet_id | INTEGER | FK → AchievementBullet, NOT NULL | |

---

### ExportSkillSelection

Records which skills were included and any per-export sort override.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| export_id | INTEGER | FK → ResumeExport, NOT NULL | |
| skill_id | INTEGER | FK → Skill, NOT NULL | |
| custom_sort_order | INTEGER | | NULL = use default competence sort |

---

### ExportAcademicSelection

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| export_id | INTEGER | FK → ResumeExport, NOT NULL | |
| academic_id | INTEGER | FK → AcademicCredential, NOT NULL | |

---

### ExportCertSelection

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| export_id | INTEGER | FK → ResumeExport, NOT NULL | |
| cert_id | INTEGER | FK → Certification, NOT NULL | |

---

### ExportDescriptorSelection

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| export_id | INTEGER | FK → ResumeExport, NOT NULL | |
| descriptor_id | INTEGER | FK → RoleDescriptor, NOT NULL | |
| sort_order | INTEGER | NOT NULL | Order in the descriptor bar |

---

### CoverLetter

A cover letter document associated with a job application.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| title | TEXT | NOT NULL | Descriptive label |
| body_text | TEXT | NOT NULL | Plain text content |
| file_path | TEXT | | Path to exported PDF, NULL if not exported |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**References**: FR-026, FR-027

---

### JobApplication

A record of a job submission with status tracking.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| company_name | TEXT | NOT NULL | |
| position_title | TEXT | NOT NULL | |
| date_applied | TEXT | NOT NULL | ISO 8601 date |
| status | TEXT | NOT NULL, DEFAULT 'Applied' | From fixed set |
| fit_indicator | TEXT | | From fixed set, NULL if not assessed |
| resume_export_id | INTEGER | FK → ResumeExport | Snapshot reference |
| cover_letter_id | INTEGER | FK → CoverLetter | Optional |
| notes | TEXT | | Free-form notes |
| created_at | TEXT | NOT NULL | ISO 8601 |
| updated_at | TEXT | NOT NULL | ISO 8601 |

**Status Values** (fixed set, any transition permitted):
Applied, Acknowledged, Screening, Phone Screen, Interview Scheduled,
Interview Completed, Technical Assessment, Final Round, Offer Received,
Offer Accepted, Offer Declined, Employer Rejected, User Withdrawn,
Ghosted, On Hold

**Fit Indicator Values** (fixed set):
Unlikely, Stretch Fit, Possible Fit, Strong Fit, Perfect Fit

**Validation**: company_name and position_title must be non-empty.
status must be from the fixed set. fit_indicator must be from its
fixed set if provided.

**References**: FR-022, FR-023, FR-024, FR-025, FR-032

---

### StatusHistory

Append-only log of status changes for a job application.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | INTEGER | PK, auto | |
| application_id | INTEGER | FK → JobApplication, NOT NULL | |
| from_status | TEXT | | NULL for initial status |
| to_status | TEXT | NOT NULL | |
| changed_at | TEXT | NOT NULL | ISO 8601 timestamp |

**References**: FR-023

---

## SQLite Schema (Initial Migration v1)

```sql
-- Migration v1: Initial schema
PRAGMA user_version = 1;

CREATE TABLE user_profile (
    id INTEGER PRIMARY KEY,
    full_name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT DEFAULT '',
    location TEXT DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE profile_link (
    id INTEGER PRIMARY KEY,
    label TEXT NOT NULL,
    url TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE work_history_entry (
    id INTEGER PRIMARY KEY,
    employer_name TEXT NOT NULL,
    job_title TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT,
    date_granularity_start TEXT NOT NULL DEFAULT 'month',
    date_granularity_end TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE achievement_bullet (
    id INTEGER PRIMARY KEY,
    work_history_id INTEGER NOT NULL REFERENCES work_history_entry(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE skill_category (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE skill (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    category_id INTEGER NOT NULL REFERENCES skill_category(id),
    competence_level INTEGER NOT NULL CHECK (competence_level BETWEEN 1 AND 10),
    is_legacy INTEGER NOT NULL DEFAULT 0 CHECK (is_legacy IN (0, 1)),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE academic_credential (
    id INTEGER PRIMARY KEY,
    institution TEXT NOT NULL,
    credential_type TEXT NOT NULL,
    field_of_study TEXT NOT NULL,
    completion_date TEXT NOT NULL,
    date_granularity TEXT NOT NULL DEFAULT 'year',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE certification (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    issuing_body TEXT NOT NULL,
    date_earned TEXT NOT NULL,
    expiration_date TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE professional_summary (
    id INTEGER PRIMARY KEY,
    label TEXT NOT NULL UNIQUE,
    body_text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE role_descriptor (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE lens (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    summary_id INTEGER REFERENCES professional_summary(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE lens_work_history_selection (
    id INTEGER PRIMARY KEY,
    lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
    work_history_id INTEGER NOT NULL REFERENCES work_history_entry(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(lens_id, work_history_id)
);

CREATE TABLE lens_bullet_selection (
    id INTEGER PRIMARY KEY,
    lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
    bullet_id INTEGER NOT NULL REFERENCES achievement_bullet(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(lens_id, bullet_id)
);

CREATE TABLE lens_skill_selection (
    id INTEGER PRIMARY KEY,
    lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
    skill_id INTEGER NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
    custom_sort_order INTEGER,
    UNIQUE(lens_id, skill_id)
);

CREATE TABLE lens_academic_selection (
    id INTEGER PRIMARY KEY,
    lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
    academic_id INTEGER NOT NULL REFERENCES academic_credential(id) ON DELETE CASCADE,
    UNIQUE(lens_id, academic_id)
);

CREATE TABLE lens_cert_selection (
    id INTEGER PRIMARY KEY,
    lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
    cert_id INTEGER NOT NULL REFERENCES certification(id) ON DELETE CASCADE,
    UNIQUE(lens_id, cert_id)
);

CREATE TABLE lens_descriptor_selection (
    id INTEGER PRIMARY KEY,
    lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
    descriptor_id INTEGER NOT NULL REFERENCES role_descriptor(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(lens_id, descriptor_id)
);

CREATE TABLE skill_lens_tag (
    id INTEGER PRIMARY KEY,
    skill_id INTEGER NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
    lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
    UNIQUE(skill_id, lens_id)
);

CREATE TABLE resume_export (
    id INTEGER PRIMARY KEY,
    template_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    summary_id INTEGER REFERENCES professional_summary(id) ON DELETE SET NULL,
    lens_id INTEGER REFERENCES lens(id) ON DELETE SET NULL,
    generated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE export_work_history_selection (
    id INTEGER PRIMARY KEY,
    export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
    work_history_id INTEGER NOT NULL REFERENCES work_history_entry(id) ON DELETE CASCADE
);

CREATE TABLE export_bullet_selection (
    id INTEGER PRIMARY KEY,
    export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
    bullet_id INTEGER NOT NULL REFERENCES achievement_bullet(id) ON DELETE CASCADE
);

CREATE TABLE export_skill_selection (
    id INTEGER PRIMARY KEY,
    export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
    skill_id INTEGER NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
    custom_sort_order INTEGER
);

CREATE TABLE export_academic_selection (
    id INTEGER PRIMARY KEY,
    export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
    academic_id INTEGER NOT NULL REFERENCES academic_credential(id) ON DELETE CASCADE
);

CREATE TABLE export_cert_selection (
    id INTEGER PRIMARY KEY,
    export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
    cert_id INTEGER NOT NULL REFERENCES certification(id) ON DELETE CASCADE
);

CREATE TABLE export_descriptor_selection (
    id INTEGER PRIMARY KEY,
    export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
    descriptor_id INTEGER NOT NULL REFERENCES role_descriptor(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE cover_letter (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    body_text TEXT NOT NULL,
    file_path TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE job_application (
    id INTEGER PRIMARY KEY,
    company_name TEXT NOT NULL,
    position_title TEXT NOT NULL,
    date_applied TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Applied',
    fit_indicator TEXT,
    resume_export_id INTEGER REFERENCES resume_export(id) ON DELETE SET NULL,
    cover_letter_id INTEGER REFERENCES cover_letter(id) ON DELETE SET NULL,
    notes TEXT DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE status_history (
    id INTEGER PRIMARY KEY,
    application_id INTEGER NOT NULL REFERENCES job_application(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status TEXT NOT NULL,
    changed_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- Indexes for common queries
CREATE INDEX idx_achievement_bullet_work_history ON achievement_bullet(work_history_id);
CREATE INDEX idx_skill_category_id ON skill(category_id);
CREATE INDEX idx_skill_competence ON skill(competence_level DESC, name ASC);
CREATE INDEX idx_job_application_company ON job_application(company_name);
CREATE INDEX idx_job_application_status ON job_application(status);
CREATE INDEX idx_status_history_application ON status_history(application_id);
CREATE INDEX idx_export_selections_export ON export_work_history_selection(export_id);
CREATE INDEX idx_export_bullet_export ON export_bullet_selection(export_id);
CREATE INDEX idx_export_skill_export ON export_skill_selection(export_id);
CREATE INDEX idx_lens_work_history ON lens_work_history_selection(lens_id);
CREATE INDEX idx_lens_bullet ON lens_bullet_selection(lens_id);
CREATE INDEX idx_lens_skill ON lens_skill_selection(lens_id);
CREATE INDEX idx_lens_academic ON lens_academic_selection(lens_id);
CREATE INDEX idx_lens_cert ON lens_cert_selection(lens_id);
CREATE INDEX idx_lens_descriptor ON lens_descriptor_selection(lens_id);
CREATE INDEX idx_skill_lens_tag_skill ON skill_lens_tag(skill_id);
CREATE INDEX idx_skill_lens_tag_lens ON skill_lens_tag(lens_id);
```

---

## State Transitions

### Job Application Status

All transitions between any two statuses are permitted (no forced
ordering). The status_history table records each transition with a
timestamp.

```
Applied ←→ Acknowledged ←→ Screening ←→ Phone Screen
  ↕              ↕              ↕             ↕
Interview Scheduled ←→ Interview Completed ←→ Technical Assessment
  ↕                          ↕                       ↕
Final Round ←→ Offer Received ←→ Offer Accepted
                     ↕                    ↕
              Offer Declined      Employer Rejected
                     ↕                    ↕
              User Withdrawn          Ghosted
                     ↕
                  On Hold

(Any → Any transitions allowed)
```

### Certification Status

Derived (not stored). Computed at read time:
- `expiration_date IS NULL` → Active
- `expiration_date >= today` → Active
- `expiration_date < today` → Inactive/Expired
