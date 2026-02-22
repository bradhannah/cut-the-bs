package domain

import "context"

// Store defines the persistence interface for all cut-the-bs
// entities. Per Constitution Principle III, this interface is defined
// in the domain layer and implemented by the infrastructure layer
// (internal/infra/sqlite).
//
// Methods are grouped by entity for readability. All methods accept
// a context.Context for cancellation support.
type Store interface {
	// --- User Profile ---

	// GetProfile returns the user's profile. If no profile exists,
	// a default empty profile is created and returned.
	GetProfile(ctx context.Context) (UserProfile, error)

	// UpdateProfile updates the user's profile fields and returns
	// the updated profile.
	UpdateProfile(ctx context.Context, profile UserProfile) (UserProfile, error)

	// --- Profile Links ---

	// ListProfileLinks returns all profile links ordered by
	// sort_order.
	ListProfileLinks(ctx context.Context) ([]ProfileLink, error)

	// CreateProfileLink creates a new profile link and returns
	// the created record with its generated ID.
	CreateProfileLink(ctx context.Context, input ProfileLinkInput) (ProfileLink, error)

	// UpdateProfileLink updates an existing profile link.
	UpdateProfileLink(ctx context.Context, id int64, input ProfileLinkInput) (ProfileLink, error)

	// DeleteProfileLink deletes a profile link by ID.
	DeleteProfileLink(ctx context.Context, id int64) error

	// ReorderProfileLinks updates the sort_order of all profile
	// links based on the provided ordered slice of IDs.
	ReorderProfileLinks(ctx context.Context, orderedIDs []int64) error

	// --- Work History ---

	// ListWorkHistory returns all work history entries ordered by
	// sort_order, each with its achievement bullets included.
	ListWorkHistory(ctx context.Context) ([]WorkHistoryEntry, error)

	// GetWorkHistory returns a single work history entry by ID
	// with its bullets included.
	GetWorkHistory(ctx context.Context, id int64) (WorkHistoryEntry, error)

	// CreateWorkHistory creates a new work history entry and
	// returns the created record with its generated ID.
	CreateWorkHistory(ctx context.Context, input WorkHistoryInput) (WorkHistoryEntry, error)

	// UpdateWorkHistory updates an existing work history entry.
	UpdateWorkHistory(ctx context.Context, id int64, input WorkHistoryInput) (WorkHistoryEntry, error)

	// DeleteWorkHistory deletes a work history entry and its
	// associated bullets (CASCADE).
	DeleteWorkHistory(ctx context.Context, id int64) error

	// ReorderWorkHistory updates the sort_order of all work
	// history entries based on the provided ordered slice of IDs.
	ReorderWorkHistory(ctx context.Context, orderedIDs []int64) error

	// --- Achievement Bullets ---

	// CreateBullet adds an achievement bullet to a work history
	// entry. bulletType must be BulletTypePrimary or BulletTypeSecondary.
	CreateBullet(ctx context.Context, workHistoryID int64, text string, bulletType string) (AchievementBullet, error)

	// UpdateBullet updates an achievement bullet's text.
	UpdateBullet(ctx context.Context, id int64, text string) (AchievementBullet, error)

	// DeleteBullet deletes an achievement bullet by ID.
	DeleteBullet(ctx context.Context, id int64) error

	// ReorderBullets updates the sort_order of bullets within a
	// single work history entry.
	ReorderBullets(ctx context.Context, workHistoryID int64, orderedIDs []int64) error

	// --- Skill Categories ---

	// ListSkillCategories returns all skill categories ordered by
	// sort_order.
	ListSkillCategories(ctx context.Context) ([]SkillCategory, error)

	// CreateSkillCategory creates a new skill category appended
	// to the end of the sort order.
	CreateSkillCategory(ctx context.Context, name string) (SkillCategory, error)

	// RenameSkillCategory updates a category's name.
	RenameSkillCategory(ctx context.Context, id int64, name string) (SkillCategory, error)

	// DeleteSkillCategory deletes a skill category. Fails if any
	// skills still reference it.
	DeleteSkillCategory(ctx context.Context, id int64) error

	// ReorderSkillCategories updates the sort_order of all
	// categories based on the provided ordered slice of IDs.
	ReorderSkillCategories(ctx context.Context, orderedIDs []int64) error

	// --- Skills ---

	// ListSkills returns all skills sorted by competence level
	// (descending), then alphabetically.
	ListSkills(ctx context.Context) ([]Skill, error)

	// ListSkillsByCategory returns skills grouped by category,
	// with categories ordered by their sort_order.
	ListSkillsByCategory(ctx context.Context) ([]SkillCategoryWithSkills, error)

	// CreateSkill creates a new skill.
	CreateSkill(ctx context.Context, input SkillInput) (Skill, error)

	// UpdateSkill updates an existing skill.
	UpdateSkill(ctx context.Context, id int64, input SkillInput) (Skill, error)

	// DeleteSkill deletes a skill by ID.
	DeleteSkill(ctx context.Context, id int64) error

	// CheckSkillLensReferences returns the names of lenses that
	// reference a given skill (for delete confirmation per FR-050).
	CheckSkillLensReferences(ctx context.Context, id int64) ([]string, error)

	// --- Skill Lens Tags ---

	// GetSkillLensTags returns all lens IDs tagged for a skill.
	GetSkillLensTags(ctx context.Context, skillID int64) ([]int64, error)

	// SetSkillLensTags replaces all lens tags for a skill.
	SetSkillLensTags(ctx context.Context, skillID int64, lensIDs []int64) error

	// ListSkillsWithLensTags returns all skills with their lens
	// tag associations included.
	ListSkillsWithLensTags(ctx context.Context) ([]SkillWithTags, error)

	// --- Academic Credentials ---

	// ListAcademicCredentials returns all academic records ordered
	// by sort_order.
	ListAcademicCredentials(ctx context.Context) ([]AcademicCredential, error)

	// CreateAcademicCredential creates a new academic record.
	CreateAcademicCredential(ctx context.Context, input AcademicInput) (AcademicCredential, error)

	// UpdateAcademicCredential updates an academic record.
	UpdateAcademicCredential(ctx context.Context, id int64, input AcademicInput) (AcademicCredential, error)

	// DeleteAcademicCredential deletes an academic record by ID.
	DeleteAcademicCredential(ctx context.Context, id int64) error

	// ReorderAcademicCredentials updates sort_order for all
	// academic credentials.
	ReorderAcademicCredentials(ctx context.Context, orderedIDs []int64) error

	// --- Certifications ---

	// ListCertifications returns all certifications with computed
	// active/inactive status.
	ListCertifications(ctx context.Context) ([]Certification, error)

	// CreateCertification creates a new certification.
	CreateCertification(ctx context.Context, input CertificationInput) (Certification, error)

	// UpdateCertification updates a certification.
	UpdateCertification(ctx context.Context, id int64, input CertificationInput) (Certification, error)

	// DeleteCertification deletes a certification by ID.
	DeleteCertification(ctx context.Context, id int64) error

	// ReorderCertifications updates sort_order for all
	// certifications.
	ReorderCertifications(ctx context.Context, orderedIDs []int64) error

	// --- Professional Summaries ---

	// ListSummaries returns all professional summary variants.
	ListSummaries(ctx context.Context) ([]ProfessionalSummary, error)

	// GetSummary returns a single summary by ID.
	GetSummary(ctx context.Context, id int64) (ProfessionalSummary, error)

	// CreateSummary creates a new summary variant.
	CreateSummary(ctx context.Context, input SummaryInput) (ProfessionalSummary, error)

	// UpdateSummary updates an existing summary.
	UpdateSummary(ctx context.Context, id int64, input SummaryInput) (ProfessionalSummary, error)

	// DeleteSummary deletes a summary variant by ID.
	DeleteSummary(ctx context.Context, id int64) error

	// --- Role Descriptors ---

	// ListDescriptors returns all role descriptors ordered by
	// sort_order.
	ListDescriptors(ctx context.Context) ([]RoleDescriptor, error)

	// CreateDescriptor creates a new role descriptor.
	CreateDescriptor(ctx context.Context, title string) (RoleDescriptor, error)

	// UpdateDescriptor updates a descriptor's title.
	UpdateDescriptor(ctx context.Context, id int64, title string) (RoleDescriptor, error)

	// DeleteDescriptor deletes a role descriptor by ID.
	DeleteDescriptor(ctx context.Context, id int64) error

	// ReorderDescriptors updates sort_order for all descriptors.
	ReorderDescriptors(ctx context.Context, orderedIDs []int64) error

	// --- Core Expertise ---

	// ListCoreExpertise returns all core expertise items ordered
	// by sort_order.
	ListCoreExpertise(ctx context.Context) ([]CoreExpertise, error)

	// CreateCoreExpertise creates a new core expertise item.
	CreateCoreExpertise(ctx context.Context, label string) (CoreExpertise, error)

	// UpdateCoreExpertise updates a core expertise item's label.
	UpdateCoreExpertise(ctx context.Context, id int64, label string) (CoreExpertise, error)

	// DeleteCoreExpertise deletes a core expertise item by ID.
	DeleteCoreExpertise(ctx context.Context, id int64) error

	// ReorderCoreExpertise updates sort_order for all core
	// expertise items.
	ReorderCoreExpertise(ctx context.Context, orderedIDs []int64) error

	// --- Lenses ---

	// ListLenses returns all lenses.
	ListLenses(ctx context.Context) ([]Lens, error)

	// GetLens returns a single lens with all its content
	// selections.
	GetLens(ctx context.Context, id int64) (LensDetail, error)

	// CreateLens creates a new lens.
	CreateLens(ctx context.Context, input LensInput) (Lens, error)

	// UpdateLens updates a lens's name.
	UpdateLens(ctx context.Context, id int64, input LensInput) (Lens, error)

	// DeleteLens deletes a lens and all its selections.
	DeleteLens(ctx context.Context, id int64) error

	// SetLensSummaries replaces the summary selections for a lens.
	SetLensSummaries(ctx context.Context, lensID int64, selections []LensSummaryItem) error

	// SetLensWorkHistory replaces the work history selections for
	// a lens.
	SetLensWorkHistory(ctx context.Context, lensID int64, selections []LensWorkHistoryItem) error

	// SetLensBullets replaces the bullet selections for a lens.
	SetLensBullets(ctx context.Context, lensID int64, selections []LensBulletItem) error

	// SetLensSkills replaces the skill selections for a lens.
	SetLensSkills(ctx context.Context, lensID int64, selections []LensSkillItem) error

	// SetLensAcademics replaces the academic selections for a
	// lens.
	SetLensAcademics(ctx context.Context, lensID int64, academicIDs []int64) error

	// SetLensCerts replaces the certification selections for a
	// lens.
	SetLensCerts(ctx context.Context, lensID int64, certIDs []int64) error

	// SetLensDescriptors replaces the descriptor selections for
	// a lens.
	SetLensDescriptors(ctx context.Context, lensID int64, selections []LensDescriptorItem) error

	// SetLensCoreExpertise replaces the core expertise selections
	// for a lens.
	SetLensCoreExpertise(ctx context.Context, lensID int64, selections []LensCoreExpertiseItem) error

	// --- Lens Reference Checks ---

	// CheckSummaryLensReferences returns the names of lenses that
	// reference a given professional summary.
	CheckSummaryLensReferences(ctx context.Context, id int64) ([]string, error)

	// CheckWorkHistoryLensReferences returns the names of lenses
	// that reference a given work history entry.
	CheckWorkHistoryLensReferences(ctx context.Context, id int64) ([]string, error)

	// CheckBulletLensReferences returns the names of lenses that
	// reference a given bullet.
	CheckBulletLensReferences(ctx context.Context, id int64) ([]string, error)

	// CheckAcademicLensReferences returns the names of lenses
	// that reference a given academic credential.
	CheckAcademicLensReferences(ctx context.Context, id int64) ([]string, error)

	// CheckCertLensReferences returns the names of lenses that
	// reference a given certification.
	CheckCertLensReferences(ctx context.Context, id int64) ([]string, error)

	// CheckDescriptorLensReferences returns the names of lenses
	// that reference a given role descriptor.
	CheckDescriptorLensReferences(ctx context.Context, id int64) ([]string, error)

	// CheckCoreExpertiseLensReferences returns the names of lenses
	// that reference a given core expertise item.
	CheckCoreExpertiseLensReferences(ctx context.Context, id int64) ([]string, error)

	// --- Resume Exports ---

	// ListExports returns all previous export records.
	ListExports(ctx context.Context) ([]ResumeExport, error)

	// CreateExport creates an export record.
	CreateExport(ctx context.Context, export ResumeExport) (ResumeExport, error)

	// GetExport returns a single export record by ID.
	GetExport(ctx context.Context, id int64) (ResumeExport, error)

	// --- Cover Letters ---

	// ListCoverLetters returns all cover letters.
	ListCoverLetters(ctx context.Context) ([]CoverLetter, error)

	// GetCoverLetter returns a single cover letter by ID.
	GetCoverLetter(ctx context.Context, id int64) (CoverLetter, error)

	// CreateCoverLetter creates a new cover letter.
	CreateCoverLetter(ctx context.Context, input CoverLetterInput) (CoverLetter, error)

	// UpdateCoverLetter updates a cover letter's content.
	UpdateCoverLetter(ctx context.Context, id int64, input CoverLetterInput) (CoverLetter, error)

	// DeleteCoverLetter deletes a cover letter by ID.
	DeleteCoverLetter(ctx context.Context, id int64) error

	// --- Job Applications ---

	// ListApplications returns all job applications.
	ListApplications(ctx context.Context) ([]JobApplication, error)

	// SearchApplications searches by company name or position
	// title.
	SearchApplications(ctx context.Context, query string) ([]JobApplication, error)

	// GetApplication returns a single job application by ID.
	GetApplication(ctx context.Context, id int64) (JobApplication, error)

	// CreateApplication creates a new job application record.
	CreateApplication(ctx context.Context, input ApplicationInput) (JobApplication, error)

	// UpdateApplication updates application fields (not status).
	UpdateApplication(ctx context.Context, id int64, input ApplicationInput) (JobApplication, error)

	// UpdateApplicationStatus changes the status of an
	// application.
	UpdateApplicationStatus(ctx context.Context, id int64, newStatus string) (JobApplication, error)

	// UpdateApplicationFit updates the fit indicator of an
	// application.
	UpdateApplicationFit(ctx context.Context, id int64, fitIndicator string) (JobApplication, error)

	// DeleteApplication deletes a job application and its status
	// history.
	DeleteApplication(ctx context.Context, id int64) error

	// --- Status History ---

	// GetApplicationHistory returns the status change history for
	// an application.
	GetApplicationHistory(ctx context.Context, applicationID int64) ([]StatusChange, error)

	// CreateStatusChange records a status transition.
	CreateStatusChange(ctx context.Context, change StatusChange) (StatusChange, error)

	// --- Document Templates ---

	// ListDocumentTemplates returns all document templates ordered
	// by is_builtin DESC, name ASC. Built-in templates appear first.
	ListDocumentTemplates(ctx context.Context) ([]DocumentTemplate, error)

	// GetDocumentTemplate returns a template with all its elements.
	GetDocumentTemplate(ctx context.Context, id int64) (TemplateDetail, error)

	// CreateDocumentTemplate creates a new user document template.
	CreateDocumentTemplate(ctx context.Context, input DocumentTemplateInput) (DocumentTemplate, error)

	// UpdateDocumentTemplate updates a user template's metadata.
	// Built-in templates cannot be updated.
	UpdateDocumentTemplate(ctx context.Context, id int64, input DocumentTemplateInput) (DocumentTemplate, error)

	// DeleteDocumentTemplate deletes a user template. Built-in
	// templates cannot be deleted. Elements are cascade deleted.
	DeleteDocumentTemplate(ctx context.Context, id int64) error

	// DuplicateDocumentTemplate creates a copy of a template with
	// all its elements. The copy is always user-created.
	DuplicateDocumentTemplate(ctx context.Context, id int64, newName string) (DocumentTemplate, error)

	// --- Template Elements ---

	// CreateTemplateElement adds a new element to a template,
	// appended to the end of the sort order within its parent scope.
	CreateTemplateElement(ctx context.Context, templateID int64, input TemplateElementInput) (TemplateElement, error)

	// UpdateTemplateElement updates an element's type and config.
	UpdateTemplateElement(ctx context.Context, id int64, input TemplateElementInput) (TemplateElement, error)

	// DeleteTemplateElement removes an element from a template.
	DeleteTemplateElement(ctx context.Context, id int64) error

	// ReorderTemplateElements updates sort_order for elements
	// within a specific parent scope (top-level if parentID is nil,
	// or within a specific loop container).
	ReorderTemplateElements(ctx context.Context, templateID int64, parentID *int64, orderedIDs []int64) error

	// --- Data Management ---

	// ExportAllData returns all data in the database as a single
	// ExportData envelope for JSON serialization.
	ExportAllData(ctx context.Context) (ExportData, error)

	// ImportAllData replaces all data with the contents of the
	// provided ExportData. This truncates all tables and inserts
	// fresh data inside a single transaction.
	ImportAllData(ctx context.Context, data ExportData) error

	// DBPath returns the path to the underlying database file.
	DBPath() string

	// Close closes the store and releases resources.
	Close() error

	// Checkpoint performs a WAL checkpoint without closing.
	Checkpoint() error
}

// PDFRenderer defines the interface for generating PDF documents.
// Per Constitution Principle III, this interface is defined in the
// domain layer and implemented by the infrastructure layer
// (internal/infra/pdf).
type PDFRenderer interface {
	// RenderResume generates a PDF resume from the given profile,
	// content selections, and template. Returns the file path of
	// the generated PDF.
	RenderResume(ctx context.Context, req RenderResumeRequest) (string, error)

	// RenderCoverLetter generates a PDF cover letter. Returns the
	// file path of the generated PDF.
	RenderCoverLetter(ctx context.Context, req RenderCoverLetterRequest) (string, error)
}

// RenderResumeRequest holds all data needed to render a resume PDF.
type RenderResumeRequest struct {
	Template           *TemplateDetail
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

// RenderCoverLetterRequest holds all data needed to render a cover
// letter PDF.
type RenderCoverLetterRequest struct {
	OutputDir string
	Profile   UserProfile
	Links     []ProfileLink
	Letter    CoverLetter
}
