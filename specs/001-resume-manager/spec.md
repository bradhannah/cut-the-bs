# Feature Specification: Resume Manager

**Feature Branch**: `001-resume-manager`
**Created**: 2026-02-19
**Status**: Draft
**Input**: User description: "Desktop resume management application with work history tracking, configurable PDF export, and job application tracking"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Work History Management (Priority: P1)

A job seeker opens cut-the-bs and enters their complete work history.
For each position they record the employer name, job title, date range
(flexible granularity: year-only, month+year, or full date), and a list
of individual achievement bullet points. They can add, edit, reorder,
and delete both positions and individual bullets at any time. Each
bullet is a discrete, selectable unit so the user can later include or
exclude specific achievements when tailoring a resume for a particular
job.

**Why this priority**: Without work history data in the system, no
resume can be generated. This is the foundational data entry capability
that every other feature depends on.

**Independent Test**: Can be fully tested by launching the application,
creating several work history entries with varying date granularities
and multiple bullets each, then verifying the entries persist after
closing and reopening the app.

**Acceptance Scenarios**:

1. **Given** the user has no work history entries, **When** they create
   a new position with employer, title, date range (month+year), and
   three achievement bullets, **Then** the entry appears in their work
   history list and persists after restarting the application.
2. **Given** the user has an existing position with five bullets,
   **When** they delete the third bullet and edit the first bullet's
   text, **Then** the changes are saved and reflected immediately.
3. **Given** the user has multiple positions, **When** they reorder
   positions by dragging or using move controls, **Then** the new order
   is preserved.
4. **Given** the user enters a position with only a year for the start
   date, **When** they save, **Then** the system accepts the entry
   without requiring month or day values.

---

### User Story 2 - PDF Resume Export (Priority: P2)

The user selects a built-in resume template and chooses which work
history entries and bullets to include. The system generates a PDF
document using the selected template. The PDF is ATS-compatible:
text is selectable, searchable, free of mid-word spaces, and has no
unexpected line breaks. The user downloads or saves the PDF locally.

**Why this priority**: PDF export is the primary output of the
application. Without it the user has no deliverable to send to
employers, making the work history data entry (US1) valuable only
as a data store rather than a productivity tool.

**Independent Test**: Can be tested by selecting a template, choosing
a subset of work history entries, generating a PDF, and verifying the
output opens correctly, text is selectable and copy-pasteable without
artifacts, and the layout matches the chosen template.

**Acceptance Scenarios**:

1. **Given** the user has work history entries and selects a resume
   template, **When** they choose specific positions and bullets to
   include and trigger export, **Then** a PDF file is generated
   containing only the selected content in the template's layout.
2. **Given** a generated PDF, **When** the user copies text from it
   and pastes into a plain text editor, **Then** the pasted text has no
   mid-word spaces, no broken words across lines, and no garbled
   characters.
3. **Given** the user has ten positions but selects only three,
   **When** the PDF is generated, **Then** only the three selected
   positions appear in the output.
4. **Given** multiple built-in templates are available, **When** the
   user switches between templates with the same data selection,
   **Then** each template produces a visually distinct but correctly
   formatted PDF.

---

### User Story 3 - Skills & Competence Tracking (Priority: P3)

The user maintains a comprehensive list of technical and professional
skills, each with an assigned competence level (e.g., expert,
proficient, familiar). Skills are organized into user-defined
categories (SkillCategory entities) that reflect how the user thinks
about their skill set (e.g., "AWS Technologies," "Azure Technologies,"
"Technologies (Architectural)," "Technologies (Hands On),"
"Programming Languages"). Categories are named entities with their own
identity — the user creates, renames, reorders, and deletes them as
needed. Skills reference their category by FK, so renaming a category
automatically applies to all skills in that group. When generating a
resume, the skills
section is auto-populated and sorted by competence level, allowing
the user to quickly present their strongest skills first. The user
can override the auto-sort order for a specific resume if needed.

**Why this priority**: Skills lists are a standard resume section that
employers and ATS systems scan for keyword matches. Tracking competence
separately from work history allows the system to auto-generate this
section, saving significant manual effort per application.

**Independent Test**: Can be tested by adding skills across multiple
categories with varying competence levels, verifying they sort
correctly by competence, and confirming the sorted list appears in a
generated PDF.

**Acceptance Scenarios**:

1. **Given** the user has no skills recorded, **When** they add a skill
   with a name, category, and competence level, **Then** the skill
   appears in their skills list under the correct category.
2. **Given** the user has 20 skills with varying competence levels,
   **When** they view the skills list, **Then** skills are displayed
   sorted by competence level (highest first) within each category.
3. **Given** the user generates a resume with the skills section
   enabled, **When** the PDF is produced, **Then** the skills section
   reflects the competence-based sort order.
4. **Given** the user has skills sorted by competence, **When** they
   manually reorder skills for a specific resume export, **Then** that
   custom order applies only to that export and does not change the
   master skills list.

---

### User Story 4 - Academic History & Certifications (Priority: P4)

The user records their academic credentials: universities/colleges with
degrees or diplomas, graduation dates, and fields of study. They also
track professional certifications, including the issuing body,
certification name, date earned, expiration date (if applicable), and
active/inactive status. Expired or inactive certifications are visually
distinguished from active ones.

**Why this priority**: Academic history and certifications are standard
resume sections. Tracking expiration dates for certifications adds
ongoing value beyond a single resume generation, alerting the user to
renewals.

**Independent Test**: Can be tested by adding academic entries and
certifications with various date ranges and statuses, verifying they
display correctly and appear in a generated PDF with the appropriate
active/inactive distinction.

**Acceptance Scenarios**:

1. **Given** the user has no academic entries, **When** they add a
   university with degree type, field of study, and graduation year,
   **Then** the entry appears in their academic history.
2. **Given** the user adds a certification with a start date and
   expiration date in the past, **When** they view their certifications,
   **Then** it is marked as inactive/expired.
3. **Given** the user adds a certification with no expiration date,
   **When** they view it, **Then** it is displayed as active with no
   expiration shown.
4. **Given** the user has academic entries and certifications, **When**
   they generate a resume, **Then** these sections appear in the PDF
   with dates formatted consistently.

---

### User Story 5 - Professional Summary (Priority: P5)

The user writes one or more professional summary paragraphs that
describe their qualitative strengths, leadership qualities, and career
narrative. These summaries are high-impact, concise statements designed
for the top of a resume. The user can maintain multiple summary
variants (e.g., one emphasizing technical leadership, another
emphasizing hands-on development) and select which to use per resume
export.

**Why this priority**: The professional summary is often the first
(and sometimes only) section a recruiter reads. Supporting multiple
variants enables the "enter once, use many" goal for different job
types.

**Independent Test**: Can be tested by creating multiple summary
variants, selecting one for a resume export, and verifying only the
selected summary appears at the top of the generated PDF.

**Acceptance Scenarios**:

1. **Given** the user has no summaries, **When** they create a summary
   with a descriptive label (e.g., "Technical Leader") and body text,
   **Then** the summary is saved and available for selection during
   export.
2. **Given** the user has three summary variants, **When** they select
   one during resume export, **Then** only that summary appears in the
   professional summary section of the PDF.
3. **Given** the user edits an existing summary, **When** they save,
   **Then** the changes are reflected in any subsequent export that
   selects that summary.

---

### User Story 6 - Job Application Tracking (Priority: P6)

After generating a resume and optional cover letter, the user records a
job application: the company name, position title, date applied, which
resume version was used, and which cover letter (if any) was attached.
The user can later update the application status (e.g., applied,
interview scheduled, offer received, rejected). The application
history provides a searchable log of every application with links to
the exact resume and cover letter versions that were sent.

**Why this priority**: Application tracking closes the loop on the
resume workflow. It provides historical context ("what did I send
to Company X?") and helps the user avoid re-applying or losing track
of active applications. However, it depends on resume export (US2)
being functional first.

**Independent Test**: Can be tested by creating a job application
record linked to an exported resume, updating the status through
several stages, and verifying the full history is searchable and the
linked resume is retrievable.

**Acceptance Scenarios**:

1. **Given** the user has exported a resume, **When** they create a job
   application record with company, position, date, and the exported
   resume reference, **Then** the application appears in their
   application history.
2. **Given** the user has an existing application with status "Applied,"
   **When** they update the status to "Interview Scheduled," **Then**
   the status change is saved with a timestamp.
3. **Given** the user has 50 application records, **When** they search
   by company name, **Then** matching applications are returned with
   their associated resume and cover letter references.
4. **Given** the user views an application record, **When** they click
   the linked resume, **Then** the exact PDF that was sent is displayed
   or downloadable.

---

### User Story 7 - Role Descriptor Tags (Priority: P7)

Some resume templates include a tagline or descriptor bar near the top
(e.g., "Software Engineer | Cloud Architect | Technical Project
Manager"). The user maintains a list of role descriptors and selects
which ones to include per resume export. The selected descriptors
appear separated by a visual divider in the chosen template layout.

**Why this priority**: Role descriptors help position the candidate at
a glance. This is a lower priority because it is a presentation
enhancement that adds polish but is not essential for a functional
resume.

**Independent Test**: Can be tested by creating several role
descriptors, selecting a subset for export, and verifying they appear
in the correct template position separated by dividers.

**Acceptance Scenarios**:

1. **Given** the user has no role descriptors, **When** they add three
   descriptors (e.g., "Software Engineer," "Cloud Architect,"
   "Technical Leader"), **Then** all three appear in their descriptor
   list.
2. **Given** the user selects two of three descriptors for a resume
   export, **When** the PDF is generated, **Then** only the two
   selected descriptors appear, separated by the template's divider
   style (e.g., vertical bar).
3. **Given** the user selects zero descriptors, **When** the PDF is
   generated, **Then** the descriptor section is omitted entirely
   from the resume layout.

---

### Edge Cases

- What happens when the user enters a position with an end date before
  the start date? The system MUST reject the entry with a clear error
  message indicating the date conflict.
- What happens when the user attempts to export a resume with no work
  history entries selected? The system MUST warn the user that no
  content is selected and prevent generation of an empty resume.
- What happens when a certification's expiration date passes while the
  application is running? The status MUST update to inactive on the
  next view or refresh without requiring manual intervention.
- How does the system handle very long achievement bullets that exceed
  the PDF template's available width? Text MUST wrap within the
  template's content area without overflowing, truncating, or
  introducing mid-word breaks.
- What happens when the user deletes a work history entry that is
  referenced by an existing job application record? The application
  record MUST retain a snapshot of the resume as it was at the time
  of application; the deletion MUST NOT retroactively alter historical
  application records.
- What happens when two skills have the same competence level? They
  MUST be sorted alphabetically within that competence tier.
- What happens when the user deletes a skill, work history entry,
  certification, or other content that is referenced by one or more
  lenses? The system MUST display a confirmation dialog listing the
  affected lenses before proceeding. On confirmation, the item is
  deleted and removed from all lens selections automatically.

## Clarifications

### Session 2026-02-19

- Q: How is the user's contact information (name, email, phone, etc.) handled? → A: Dedicated user profile with contact fields (name, email, phone, location, LinkedIn, website). Entered once; appears on all resume exports automatically.
- Q: What competence level scale is used for skills? → A: Fixed scale with 10 predefined levels, each with descriptive criteria (e.g., "Could teach a class on it," "Expert among peers," "Used in a learning project," "Never used"). Each level includes guiding hints to help users self-assess accurately and maintain consistent ordering. Skills also carry an optional relevancy indicator to flag whether a skill is current or legacy (e.g., SunOS may be expert-level but irrelevant for non-legacy roles).
- Q: What are the job application status values and transition rules? → A: Fixed set with any-direction transitions allowed. Statuses distinguish who ended the process: Applied, Acknowledged, Screening, Phone Screen, Interview Scheduled, Interview Completed, Technical Assessment, Final Round, Offer Received, Offer Accepted, Offer Declined (user chose not to accept), Employer Rejected, User Withdrawn, Ghosted (no response after reasonable period), On Hold. Each application also carries a self-assessed fit indicator: Unlikely, Stretch Fit, Possible Fit, Strong Fit, Perfect Fit.
- Q: Can users import existing resume data from external sources? → A: Yes. System supports structured data import via JSON/CSV for bulk-loading work history, skills, etc. Additionally, smart plain-text input helpers are included: an input box that accepts a pasted block of text and splits it into individual achievement bullets (e.g., by line breaks), and comma-separated skill entry that auto-splits into individual skill records. All user input is plain text; the system handles all formatting. No parsing of Word/PDF/LinkedIn documents.
- Q: How is data backed up and protected against loss? → A: Internal data is stored in SQLite. Users can export a full JSON backup for portability and machine migration, and import from a previous JSON backup to restore. The user can configure the data directory location, enabling placement on a local cloud drive (e.g., iCloud, Dropbox, OneDrive) for automatic sync. The system autosaves changes and maintains a configurable number of rolling backup copies to protect against corruption or accidental edits.
- Q: How are skill categories ordered on the exported PDF resume? → A: User-controlled order. Categories are managed as SkillCategory entities (see Q9 below) with a user-defined sort_order. The user can drag-reorder categories on the skills management screen to control their display order on the resume PDF. New categories appear at the end of the order by default. *(Clarified by Q9: categories are proper entities with FK, not free-text.)*
- Q: How are skills visually rendered on the exported PDF resume? → A: Comma-separated skill names listed under each category header (e.g., "AWS Technologies: EC2, Lambda, S3, DynamoDB"). Skills are sorted by competence level (highest first) within each category, but the competence level itself is not displayed on the resume. Competence is an internal sorting mechanism only.
- Q: When the user deletes content referenced by one or more lenses, what should happen? → A: Warn the user with a list of affected lenses, then proceed if confirmed. A confirmation dialog lists the lenses that reference the item (e.g., "This skill is included in lenses: Solutions Architect, Sales Engineer. Delete anyway?"). On confirmation, the item is deleted and silently removed from all lens selections (CASCADE).
- Q: Can the user rename a skill category and have it apply to all skills in that category? → A: Yes. Skill categories are a proper entity (SkillCategory) with their own ID. Skills reference the category by FK. Renaming a category updates one record and all skills in that category reflect the change automatically. The SkillCategory entity also carries a sort_order for user-controlled display ordering on the resume.
- Q: What determines the PDF layout for cover letters? → A: Profile header (name, contact info, links) at the top, then the cover letter body text rendered below in a standard letter format. Initially a single built-in cover letter layout. Multiple cover letter templates (analogous to resume templates) are a future enhancement.

## Requirements *(mandatory)*

### Functional Requirements

**User Profile**

- **FR-028**: System MUST allow the user to create and maintain a
  personal profile containing: full name, email address, phone number,
  and location (city/region). The user may also manage an ordered list
  of labelled profile links (e.g., LinkedIn, GitHub, portfolio,
  personal website) — each with a display label and URL.
- **FR-029**: System MUST automatically include the user's profile
  contact information and selected profile links on every generated
  resume, positioned according to the selected template's layout.

**Work History**

- **FR-001**: System MUST allow users to create, read, update, and
  delete work history entries consisting of: employer name, job title,
  start date, end date (or "present"), and an ordered list of
  achievement bullets.
- **FR-002**: System MUST accept dates at three granularity levels:
  year only (e.g., 2023), month and year (e.g., March 2023), or full
  date (e.g., March 15, 2023).
- **FR-003**: System MUST validate that end dates are not before start
  dates.
- **FR-004**: System MUST allow individual achievement bullets to be
  independently selectable for inclusion or exclusion during resume
  export.
- **FR-005**: System MUST persist all work history data locally so it
  survives application restarts.

**Skills**

- **FR-006**: System MUST allow users to create, read, update, and
  delete skills with: name, category (referencing a SkillCategory
  entity), competence level, and optional relevancy indicator (current
  or legacy). Categories are user-defined (e.g., "AWS Technologies,"
  "Technologies (Architectural)," "Technologies (Hands On)"). The
  system MUST NOT constrain categories to a fixed list.
- **FR-049**: System MUST allow users to manage skill categories as
  named entities. Users can create, rename, reorder, and delete
  categories. Renaming a category MUST automatically apply to all
  skills in that category (via FK relationship). Users can control the
  display order of categories; new categories MUST appear at the end
  by default. The PDF resume MUST render skill categories in the
  user-defined order.
- **FR-007**: System MUST auto-sort skills by competence level
  (highest first), with alphabetical secondary sort for ties.
  On the exported PDF, skills MUST be rendered as comma-separated
  names under each category header (e.g., "AWS Technologies: EC2,
  Lambda, S3, DynamoDB"). Competence level MUST NOT be displayed
  on the resume — it is used only for internal sort ordering.
- **FR-008**: System MUST allow users to override sort order for a
  specific resume export without changing the master list.
- **FR-030**: System MUST provide a fixed competence scale of 10
  predefined levels, ordered from highest to lowest. Each level MUST
  include descriptive criteria (hints) that help the user self-assess
  accurately (e.g., "Could teach a class on this topic," "Considered
  an expert among peers," "Used in a production environment," "Used in
  a learning project," "Awareness only, never used").
- **FR-031**: System MUST allow users to mark individual skills with a
  relevancy indicator distinguishing current/in-demand skills from
  legacy skills (e.g., SunOS: expert-level but legacy). Resume exports
  MUST be able to filter or visually distinguish skills by relevancy.

**Academic History & Certifications**

- **FR-009**: System MUST allow users to record academic credentials
  with: institution name, credential type (degree, diploma, etc.),
  field of study, and completion date.
- **FR-010**: System MUST allow users to record certifications with:
  certification name, issuing body, date earned, expiration date
  (optional), and active/inactive status.
- **FR-011**: System MUST automatically mark certifications as inactive
  when their expiration date has passed.

**Professional Summary**

- **FR-012**: System MUST allow users to create and maintain multiple
  professional summary variants, each with a descriptive label and
  body text.
- **FR-013**: System MUST allow users to select one summary variant per
  resume export.

**Resume Export**

- **FR-014**: System MUST generate PDF documents from user-selected
  data and a chosen template.
- **FR-015**: Generated PDFs MUST be ATS-compatible: text MUST be
  selectable, searchable, free of mid-word spaces, and free of
  unexpected line breaks.
- **FR-016**: System MUST provide at least two distinct built-in resume
  templates at launch.
- **FR-017**: System MUST allow users to select which work history
  entries, bullets, skills, academic entries, certifications, summary
  variant, and role descriptors to include in each export.
- **FR-018**: System MUST render dates consistently within a single
  resume export according to the template's date format style.

**Role Descriptors**

- **FR-019**: System MUST allow users to create and manage a list of
  role descriptor tags (e.g., "Software Engineer," "Cloud Architect").
- **FR-020**: System MUST allow users to select a subset of descriptors
  per resume export, displayed with template-defined dividers.
- **FR-021**: System MUST omit the descriptor section entirely when no
  descriptors are selected.

**Job Application Tracking**

- **FR-022**: System MUST allow users to record job applications with:
  company name, position title, date applied, linked resume export,
  and optional linked cover letter.
- **FR-023**: System MUST allow users to update application status
  using a fixed set of values: Applied, Acknowledged, Screening,
  Phone Screen, Interview Scheduled, Interview Completed, Technical
  Assessment, Final Round, Offer Received, Offer Accepted, Offer
  Declined (user chose not to accept), Employer Rejected, User
  Withdrawn, Ghosted (no response after reasonable period), On Hold.
  Any transition between statuses MUST be permitted (no forced
  ordering). Each status change MUST be recorded with a timestamp.
- **FR-032**: System MUST allow users to assign a self-assessed fit
  indicator to each job application from a fixed set: Unlikely,
  Stretch Fit, Possible Fit, Strong Fit, Perfect Fit. The fit
  indicator MUST be editable at any time during the application
  lifecycle.
- **FR-024**: System MUST retain a snapshot of the resume and cover
  letter as they were at time of application, independent of later
  edits to the source data.
- **FR-025**: System MUST support searching application history by
  company name and position title.

**Cover Letters**

- **FR-026**: System MUST allow users to create, edit, and store cover
  letter documents associated with specific job applications.
- **FR-027**: System MUST support exporting cover letters as PDF
  documents with the same ATS-compatibility standards as resumes. The
  cover letter PDF MUST include the user's profile header (name,
  contact info, profile links) at the top, followed by the body text.
  Initially a single built-in cover letter layout is provided.

**Data Import & Smart Input**

- **FR-033**: System MUST support importing work history, skills,
  academic credentials, and certifications from structured data files
  (JSON and CSV formats).
- **FR-034**: System MUST provide a plain-text input helper for
  achievement bullets that accepts a pasted block of text and splits
  it into individual bullet entries (e.g., splitting on line breaks).
  The user MUST be able to review and edit the split results before
  confirming.
- **FR-035**: System MUST provide a comma-separated input mode for
  skills that auto-splits a pasted string into individual skill
  records. The user MUST be able to assign category and competence
  level to each resulting skill before saving.
- **FR-036**: All user input MUST be accepted as plain text. The system
  MUST handle all formatting, styling, and presentation. Users MUST
  NOT be required to enter markup, HTML, or rich text.

**Data Storage & Backup**

- **FR-037**: System MUST store all user data internally in a SQLite
  database.
- **FR-038**: System MUST allow the user to configure the data
  directory location (where the SQLite database and backups reside),
  enabling placement on a local cloud drive (e.g., iCloud, Dropbox,
  OneDrive) for automatic sync.
- **FR-039**: System MUST support full data export to a JSON file
  containing all user data (profile, work history, skills, academic
  credentials, certifications, summaries, role descriptors, cover
  letters, and job application records) for backup and machine
  migration.
- **FR-040**: System MUST support restoring all user data from a
  previously exported JSON backup file.
- **FR-041**: System MUST autosave all changes without requiring
  manual save actions from the user.
- **FR-042**: System MUST maintain a configurable number of rolling
  backup copies of the database to protect against corruption or
  accidental edits. The user MUST be able to set the number of backup
  copies to retain.

**Profile Links**

- **FR-043**: System MUST allow the user to manage an ordered list of
  profile links, each with a display label (e.g., "GitHub,"
  "LinkedIn," "Portfolio") and a URL. Links can be added, edited,
  reordered, and deleted independently. Resume templates MUST render
  whichever links the user has configured.

**Lenses (Job-Type Variants)**

- **FR-044**: System MUST allow users to create, edit, and delete
  named lenses (reusable content selection presets). Each lens has a
  unique name identifying the job type or role variant (e.g., "Sales
  Engineer," "Solutions Architect").
- **FR-045**: A lens MUST pre-configure the full set of content
  selections: which work history entries are included, which bullets
  within each entry (with per-lens ordering), which skills, which
  academic credentials, which certifications, which role descriptors,
  and which professional summary. A lens controls all content
  sections — it is a full lens, not a partial filter.
- **FR-046**: At export time, the user MUST be able to select a lens
  to pre-fill all content selections. The user MUST be able to
  override any individual selection (add or remove items) before
  generating the PDF. One-off overrides MUST NOT modify the saved
  lens configuration.

**Skill Lens Tagging**

- **FR-047**: System MUST allow users to tag individual skills with
  one or more lenses they belong to. When exporting with a lens, only
  skills tagged for that lens are auto-included. Skills with no lens
  tags are not auto-included by any lens but can be manually added
  during export override.
- **FR-050**: When deleting any content item (skill, work history
  entry, bullet, academic credential, certification, role descriptor)
  that is referenced by one or more lenses, the system MUST display a
  confirmation dialog listing the affected lenses. On confirmation,
  the item MUST be deleted and automatically removed from all lens
  selections.

**Zoom Widget**

- **FR-048**: The application UI MUST include a zoom control widget
  in the bottom-left corner, defaulting to 100% zoom. The widget MUST
  provide +/- buttons and support Cmd/Ctrl +/- keyboard shortcuts for
  adjusting the UI zoom level.

### Key Entities

- **User Profile**: The user's personal contact information. Attributes:
  full name, email address, phone number, location. One profile per
  installation. Included on all resume exports. Links (LinkedIn,
  GitHub, portfolio, etc.) are managed separately as Profile Links.
- **Profile Link**: A labelled URL associated with the user's profile
  (e.g., "GitHub" → https://github.com/user). Attributes: display
  label, URL, sort order. Multiple links per profile, rendered by
  resume templates.
- **Work History Entry**: A single employment record. Attributes:
  employer name, job title, start date, end date, employment status
  (current or past). Contains an ordered list of Achievement Bullets.
- **Achievement Bullet**: A single accomplishment or responsibility
  line item. Belongs to one Work History Entry. Can be independently
  selected or excluded per export.
- **Skill**: A single technical or professional capability. Attributes:
   name, category (FK to SkillCategory), competence level (from fixed
   10 level scale with descriptive criteria per level), relevancy
  indicator (current or legacy). Sortable by competence. A skill may
  be high-competence but low-relevancy (e.g., expert in a deprecated
  technology). Skills can be tagged with one or more lenses for
  automatic inclusion during lens-based export.
- **Skill Category**: A named grouping for skills. Attributes: name
  (unique), sort order. User-defined (e.g., "AWS Technologies,"
  "Technologies (Architectural)"). Skills reference their category by
  ID; renaming a category updates all associated skills automatically.
- **Academic Credential**: A degree, diploma, or other academic award.
  Attributes: institution, credential type, field of study, completion
  date.
- **Certification**: A professional certification. Attributes: name,
  issuing body, date earned, expiration date (optional), status
  (active/inactive). Status derived from expiration date when present.
- **Professional Summary**: A reusable summary block. Attributes:
  label (for identification), body text. Multiple variants per user.
- **Role Descriptor**: A short tag describing a professional role.
  Attributes: title text. Selectable per export.
- **Lens**: A named, reusable content selection preset tied to a job
  type or role variant (e.g., "Sales Engineer," "Solutions Architect").
  Attributes: name, linked professional summary, and selections for
  work history entries (with bullet-level control and per-lens
  ordering), skills, academic credentials, certifications, and role
  descriptors. A lens pre-fills export selections but allows one-off
  overrides that do not modify the saved lens.
- **Resume Template**: A built-in layout definition that controls the
  visual presentation of resume data. Not user-editable in the initial
  version.
- **Resume Export**: A generated PDF artifact. Attributes: template
  used, data selections (which entries, bullets, skills, etc.),
  generation date, file reference.
- **Cover Letter**: A document associated with a job application.
  Attributes: body text, associated job application, export file
  reference.
- **Job Application**: A record of a submission to an employer.
  Attributes: company name, position title, date applied, status
  (from fixed set: Applied, Acknowledged, Screening, Phone Screen,
  Interview Scheduled, Interview Completed, Technical Assessment,
  Final Round, Offer Received, Offer Accepted, Offer Declined,
  Employer Rejected, User Withdrawn, Ghosted, On Hold), status change
  history with timestamps, self-assessed fit indicator (Unlikely,
  Stretch Fit, Possible Fit, Strong Fit, Perfect Fit), linked resume
  export, linked cover letter (optional).

## Assumptions

- The application runs locally as a desktop app. No user account
  system, authentication, or multi-user support is required.
- All data is stored locally in a SQLite database within a
  user-configurable directory. No dedicated cloud sync service is
  built into the application, but users can place the data directory
  on a cloud-synced folder for indirect sync.
- Cover letters are free-form text documents with a single built-in
  layout (profile header + body text). Multiple cover letter templates
  (analogous to resume templates) are planned as a future enhancement.
- AI-assisted features (language optimization, metric suggestions,
  voice matching) are explicitly out of scope for this feature and will
  be specified separately.
- "ATS-compatible" means the PDF text layer accurately represents the
  visual content with no artifacts when parsed by standard text
  extraction tools.
- Resume templates are built-in and not user-customizable in this
  version. Users select from provided templates but cannot modify
  layouts.
- The application supports a single user per installation. There is no
  concept of switching between user profiles.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can enter a complete work history entry (employer,
  title, dates, three bullets) in under 3 minutes on first use.
- **SC-002**: Users can generate a tailored PDF resume from existing
  data (selecting entries, skills, template) in under 2 minutes.
- **SC-003**: 100% of generated PDFs pass ATS text extraction
  validation: copied text contains no mid-word spaces, no garbled
  characters, and no broken words across lines.
- **SC-004**: Users can find a specific past job application by company
  name within 10 seconds when 50+ applications are recorded.
- **SC-005**: All user data entered in a session persists correctly
  after closing and reopening the application with zero data loss.
- **SC-006**: Users can switch between resume templates and regenerate
  a PDF in under 30 seconds.
- **SC-007**: The skills section of an exported resume correctly
  reflects the competence-based sort order in 100% of exports.
