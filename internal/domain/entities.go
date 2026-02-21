package domain

// UserProfile represents the user's personal contact information.
// A single record exists per installation.
type UserProfile struct {
	ID        int64  `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Location  string `json:"location"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ProfileLink is a labelled URL associated with the user's profile
// (e.g., LinkedIn, GitHub, portfolio).
type ProfileLink struct {
	ID        int64  `json:"id"`
	Label     string `json:"label"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ProfileLinkInput is the input type for creating or updating a
// profile link.
type ProfileLinkInput struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// WorkHistoryEntry is a single employment record containing an
// ordered list of achievement bullets.
type WorkHistoryEntry struct {
	ID                   int64               `json:"id"`
	EmployerName         string              `json:"employer_name"`
	JobTitle             string              `json:"job_title"`
	Summary              string              `json:"summary"`
	StartDate            string              `json:"start_date"`
	EndDate              string              `json:"end_date"`
	DateGranularityStart string              `json:"date_granularity_start"`
	DateGranularityEnd   string              `json:"date_granularity_end"`
	SortOrder            int                 `json:"sort_order"`
	Bullets              []AchievementBullet `json:"bullets"`
	CreatedAt            string              `json:"created_at"`
	UpdatedAt            string              `json:"updated_at"`
}

// WorkHistoryInput is the input type for creating or updating a
// work history entry.
type WorkHistoryInput struct {
	EmployerName         string `json:"employer_name"`
	JobTitle             string `json:"job_title"`
	Summary              string `json:"summary"`
	StartDate            string `json:"start_date"`
	EndDate              string `json:"end_date"`
	DateGranularityStart string `json:"date_granularity_start"`
	DateGranularityEnd   string `json:"date_granularity_end"`
}

// BulletTypePrimary represents standard achievement/responsibility bullets.
const BulletTypePrimary = "primary"

// BulletTypeSecondary represents high-level outcome bullets.
const BulletTypeSecondary = "secondary"

// AchievementBullet is a single accomplishment or responsibility
// line item belonging to one WorkHistoryEntry.
type AchievementBullet struct {
	ID            int64  `json:"id"`
	WorkHistoryID int64  `json:"work_history_id"`
	Text          string `json:"text"`
	BulletType    string `json:"bullet_type"`
	SortOrder     int    `json:"sort_order"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// SkillCategory is a named grouping for skills with user-controlled
// display ordering.
type SkillCategory struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SkillCategoryWithSkills is a skill category with its associated
// skills included.
type SkillCategoryWithSkills struct {
	Category SkillCategory `json:"category"`
	Skills   []Skill       `json:"skills"`
}

// Skill is a single technical or professional capability with
// competence level and relevancy indicator.
type Skill struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	CategoryID      int64  `json:"category_id"`
	CompetenceLevel int    `json:"competence_level"`
	IsLegacy        bool   `json:"is_legacy"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// SkillInput is the input type for creating or updating a skill.
type SkillInput struct {
	Name            string `json:"name"`
	CategoryID      int64  `json:"category_id"`
	CompetenceLevel int    `json:"competence_level"`
	IsLegacy        bool   `json:"is_legacy"`
}

// SkillWithTags is a skill with its lens tag associations included.
type SkillWithTags struct {
	Skill
	LensIDs []int64 `json:"lens_ids"`
}

// AcademicCredential is a degree, diploma, or other academic award.
type AcademicCredential struct {
	ID              int64  `json:"id"`
	Institution     string `json:"institution"`
	CredentialType  string `json:"credential_type"`
	FieldOfStudy    string `json:"field_of_study"`
	CompletionDate  string `json:"completion_date"`
	DateGranularity string `json:"date_granularity"`
	SortOrder       int    `json:"sort_order"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// AcademicInput is the input type for creating or updating an
// academic credential.
type AcademicInput struct {
	Institution     string `json:"institution"`
	CredentialType  string `json:"credential_type"`
	FieldOfStudy    string `json:"field_of_study"`
	CompletionDate  string `json:"completion_date"`
	DateGranularity string `json:"date_granularity"`
}

// Certification is a professional certification with optional
// expiration tracking.
type Certification struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	IssuingBody    string `json:"issuing_body"`
	DateEarned     string `json:"date_earned"`
	ExpirationDate string `json:"expiration_date"`
	IsActive       bool   `json:"is_active"` // computed at read time
	SortOrder      int    `json:"sort_order"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// CertificationInput is the input type for creating or updating a
// certification.
type CertificationInput struct {
	Name           string `json:"name"`
	IssuingBody    string `json:"issuing_body"`
	DateEarned     string `json:"date_earned"`
	ExpirationDate string `json:"expiration_date"`
}

// ProfessionalSummary is a reusable summary block with a label for
// identification.
type ProfessionalSummary struct {
	ID        int64  `json:"id"`
	Label     string `json:"label"`
	BodyText  string `json:"body_text"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SummaryInput is the input type for creating or updating a
// professional summary.
type SummaryInput struct {
	Label    string `json:"label"`
	BodyText string `json:"body_text"`
}

// RoleDescriptor is a short tag describing a professional role,
// selectable per export.
type RoleDescriptor struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CoreExpertise is a keyword or short phrase representing a core
// area of expertise, rendered as pipe-separated tags in the PDF.
type CoreExpertise struct {
	ID        int64  `json:"id"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Lens is a named, reusable content selection preset tied to a job
// type or role variant.
type Lens struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// LensInput is the input type for creating or updating a lens.
type LensInput struct {
	Name string `json:"name"`
}

// LensDetail is a lens with all its content selections included.
type LensDetail struct {
	Lens
	Summaries     []LensSummaryItem       `json:"summaries"`
	WorkHistory   []LensWorkHistoryItem   `json:"work_history"`
	Bullets       []LensBulletItem        `json:"bullets"`
	Skills        []LensSkillItem         `json:"skills"`
	AcademicIDs   []int64                 `json:"academic_ids"`
	CertIDs       []int64                 `json:"cert_ids"`
	Descriptors   []LensDescriptorItem    `json:"descriptors"`
	CoreExpertise []LensCoreExpertiseItem `json:"core_expertise"`
}

// LensSummaryItem records a summary's inclusion and position in a
// lens.
type LensSummaryItem struct {
	SummaryID int64 `json:"summary_id"`
	SortOrder int   `json:"sort_order"`
	IsMaster  bool  `json:"is_master"`
}

// LensWorkHistoryItem records a work history entry's inclusion and
// position in a lens.
type LensWorkHistoryItem struct {
	WorkHistoryID int64 `json:"work_history_id"`
	SortOrder     int   `json:"sort_order"`
}

// LensBulletItem records a bullet's inclusion and position in a lens.
type LensBulletItem struct {
	BulletID  int64 `json:"bullet_id"`
	SortOrder int   `json:"sort_order"`
}

// LensSkillItem records a skill's inclusion in a lens with optional
// custom sort order.
type LensSkillItem struct {
	SkillID         int64 `json:"skill_id"`
	CustomSortOrder *int  `json:"custom_sort_order"`
}

// LensDescriptorItem records a descriptor's inclusion and position
// in a lens.
type LensDescriptorItem struct {
	DescriptorID int64 `json:"descriptor_id"`
	SortOrder    int   `json:"sort_order"`
}

// LensCoreExpertiseItem records a core expertise item's inclusion
// and position in a lens.
type LensCoreExpertiseItem struct {
	CoreExpertiseID int64 `json:"core_expertise_id"`
	SortOrder       int   `json:"sort_order"`
}

// ResumeExport is a generated PDF artifact with a snapshot of
// what data was selected.
type ResumeExport struct {
	ID          int64  `json:"id"`
	TemplateID  string `json:"template_id"`
	FilePath    string `json:"file_path"`
	SummaryID   *int64 `json:"summary_id"`
	LensID      *int64 `json:"lens_id"`
	GeneratedAt string `json:"generated_at"`
}

// ResumeTemplate describes a built-in resume layout.
type ResumeTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PreviewURL  string `json:"preview_url"`
}

// ExportRequest holds the selections for generating a resume export.
type ExportRequest struct {
	TemplateID         string        `json:"template_id"`
	LensID             *int64        `json:"lens_id"`
	SummaryIDs         []int64       `json:"summary_ids"`
	MasterSummaryID    *int64        `json:"master_summary_id"`
	WorkHistoryIDs     []int64       `json:"work_history_ids"`
	BulletIDs          []int64       `json:"bullet_ids"`
	SkillIDs           []int64       `json:"skill_ids"`
	SkillSortOverrides map[int64]int `json:"skill_sort_overrides"`
	AcademicIDs        []int64       `json:"academic_ids"`
	CertificationIDs   []int64       `json:"certification_ids"`
	DescriptorIDs      []int64       `json:"descriptor_ids"`
	CoreExpertiseIDs   []int64       `json:"core_expertise_ids"`
}

// CoverLetter is a cover letter document.
type CoverLetter struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	BodyText  string `json:"body_text"`
	FilePath  string `json:"file_path"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CoverLetterInput is the input type for creating or updating a
// cover letter.
type CoverLetterInput struct {
	Title    string `json:"title"`
	BodyText string `json:"body_text"`
}

// JobApplication is a record of a job submission with status tracking.
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
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ApplicationInput is the input type for creating or updating a
// job application.
type ApplicationInput struct {
	CompanyName    string `json:"company_name"`
	PositionTitle  string `json:"position_title"`
	DateApplied    string `json:"date_applied"`
	FitIndicator   string `json:"fit_indicator"`
	ResumeExportID *int64 `json:"resume_export_id"`
	CoverLetterID  *int64 `json:"cover_letter_id"`
	Notes          string `json:"notes"`
}

// StatusChange records a single status transition for a job
// application.
type StatusChange struct {
	ID            int64  `json:"id"`
	ApplicationID int64  `json:"application_id"`
	FromStatus    string `json:"from_status"`
	ToStatus      string `json:"to_status"`
	ChangedAt     string `json:"changed_at"`
}

// ExportData is the JSON envelope for a full data export. It
// contains every entity in the system, allowing complete
// backup/restore of user data.
type ExportData struct {
	SchemaVersion int                   `json:"schema_version"`
	ExportedAt    string                `json:"exported_at"`
	Profile       UserProfile           `json:"profile"`
	ProfileLinks  []ProfileLink         `json:"profile_links"`
	WorkHistory   []WorkHistoryEntry    `json:"work_history"`
	Categories    []SkillCategory       `json:"skill_categories"`
	Skills        []Skill               `json:"skills"`
	Academics     []AcademicCredential  `json:"academics"`
	Certs         []Certification       `json:"certifications"`
	Summaries     []ProfessionalSummary `json:"summaries"`
	Descriptors   []RoleDescriptor      `json:"descriptors"`
	CoreExpertise []CoreExpertise       `json:"core_expertise"`
	Lenses        []LensDetail          `json:"lenses"`
	Exports       []ResumeExport        `json:"exports"`
	CoverLetters  []CoverLetter         `json:"cover_letters"`
	Applications  []JobApplication      `json:"applications"`
	StatusChanges []StatusChange        `json:"status_changes"`
	SkillLensTags []SkillLensTag        `json:"skill_lens_tags"`
}

// SkillLensTag records the association between a skill and a lens
// for import/export purposes.
type SkillLensTag struct {
	SkillID int64 `json:"skill_id"`
	LensID  int64 `json:"lens_id"`
}

// ImportResult holds the results of a data import operation.
type ImportResult struct {
	RecordsImported int      `json:"records_imported"`
	RecordsSkipped  int      `json:"records_skipped"`
	Errors          []string `json:"errors"`
}

// BackupSettings holds user-configurable backup settings. This is
// the domain-facing type (mirrors config.BackupConfig for the
// binding layer).
type BackupSettings struct {
	RollingBackupCount int `json:"rolling_backup_count"`
}
