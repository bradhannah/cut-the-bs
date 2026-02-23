package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"cut-the-bs/internal/config"
	"cut-the-bs/internal/domain"
	"cut-the-bs/internal/infra"
	"cut-the-bs/internal/infra/pdf"
	"cut-the-bs/internal/infra/sqlite"
	"cut-the-bs/internal/service"
)

// App is the main application struct that serves as the Wails
// binding target. All public methods on App are exposed to the
// Svelte frontend as JavaScript Promise-returning functions.
type App struct {
	ctx    context.Context
	cfg    config.Config
	store  *sqlite.Store
	logger *slog.Logger

	// Services
	workHistorySvc   *service.WorkHistoryService
	profileSvc       *service.ProfileService
	skillsSvc        *service.SkillsService
	academicSvc      *service.AcademicService
	summarySvc       *service.SummaryService
	descriptorSvc    *service.DescriptorService
	coreExpertiseSvc *service.CoreExpertiseService
	resumeSvc        *service.ResumeService
	coverLetterSvc   *service.CoverLetterService
	applicationSvc   *service.ApplicationService
	lensSvc          *service.LensService
	backupSvc        *service.BackupService
	templateSvc      *service.TemplateService
}

// NewApp creates a new App instance. Service dependencies are
// initialized during the startup lifecycle hook, not here.
func NewApp() *App {
	logger := infra.NewLogger(slog.LevelInfo, nil)
	return &App{
		logger: logger,
	}
}

// startup is the Wails OnStartup lifecycle hook. It loads
// configuration, ensures the data directory exists, opens the
// SQLite database, and runs migrations.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	log := infra.ComponentLogger(a.logger, "startup")

	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config, using defaults",
			slog.String("error", err.Error()))
		cfg = config.DefaultConfig()
	}
	a.cfg = cfg

	// Ensure data directory exists.
	if err := a.cfg.EnsureDataDir(); err != nil {
		log.Error("failed to create data directory",
			slog.String("error", err.Error()))
		return
	}

	// Open database.
	dbPath, err := a.cfg.DBPath()
	if err != nil {
		log.Error("failed to resolve database path",
			slog.String("error", err.Error()))
		return
	}

	dbLogger := infra.ComponentLogger(a.logger, "sqlite")
	store, err := sqlite.NewStore(dbPath, dbLogger)
	if err != nil {
		log.Error("failed to open database",
			slog.String("error", err.Error()))
		return
	}
	a.store = store

	// Run migrations.
	if err := sqlite.Migrate(store); err != nil {
		log.Error("failed to run migrations",
			slog.String("error", err.Error()))
		return
	}

	// Initialize services.
	a.workHistorySvc = service.NewWorkHistoryService(store)
	a.profileSvc = service.NewProfileService(store)
	a.skillsSvc = service.NewSkillsService(store)
	a.academicSvc = service.NewAcademicService(store)
	a.summarySvc = service.NewSummaryService(store)
	a.descriptorSvc = service.NewDescriptorService(store)
	a.coreExpertiseSvc = service.NewCoreExpertiseService(store)

	// PDF renderer and export service.
	renderer := pdf.NewRenderer()
	exportDir, err := a.cfg.ExportDir()
	if err != nil {
		log.Error("failed to resolve export directory",
			slog.String("error", err.Error()))
		return
	}
	a.resumeSvc = service.NewResumeService(store, renderer, exportDir)
	a.coverLetterSvc = service.NewCoverLetterService(store, renderer, exportDir)
	a.applicationSvc = service.NewApplicationService(store)
	a.lensSvc = service.NewLensService(store)
	a.backupSvc = service.NewBackupService(store)
	a.templateSvc = service.NewTemplateService(store)

	log.Info("application started",
		slog.String("db_path", dbPath))
}

// beforeClose is the Wails OnBeforeClose lifecycle hook. It
// checkpoints the WAL to ensure data durability before the window
// closes. Returning false allows the close to proceed.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	if a.store != nil {
		log := infra.ComponentLogger(a.logger, "lifecycle")
		if err := a.store.Checkpoint(); err != nil {
			log.Warn("WAL checkpoint failed before close",
				slog.String("error", err.Error()))
		}
	}
	return false
}

// shutdown is the Wails OnShutdown lifecycle hook. It closes the
// database connection and releases all resources.
func (a *App) shutdown(ctx context.Context) {
	log := infra.ComponentLogger(a.logger, "lifecycle")

	if a.store != nil {
		if err := a.store.Close(); err != nil {
			log.Error("failed to close database",
				slog.String("error", err.Error()))
		}
	}

	log.Info("application shut down")
}

// emitAutosave emits an autosave:complete event to notify the
// frontend that a data write has completed successfully.
func (a *App) emitAutosave() {
	wailsruntime.EventsEmit(a.ctx, "autosave:complete", map[string]string{
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// emitBackupComplete emits a backup:complete event after a
// successful rolling backup or export operation.
func (a *App) emitBackupComplete(path string) {
	wailsruntime.EventsEmit(a.ctx, "backup:complete", map[string]string{
		"timestamp": time.Now().Format(time.RFC3339),
		"path":      path,
	})
}

// emitBackupError emits a backup:error event when a backup
// operation fails.
func (a *App) emitBackupError(err error) {
	wailsruntime.EventsEmit(a.ctx, "backup:error", map[string]string{
		"error": err.Error(),
	})
}

// Greet returns a greeting for the given name.
// This is a placeholder binding method for initial setup verification.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// =================================================================
// Work History Bindings
// =================================================================

// ListWorkHistory returns all work history entries ordered by
// sort_order, each with its achievement bullets.
func (a *App) ListWorkHistory() ([]domain.WorkHistoryEntry, error) {
	return a.workHistorySvc.ListWorkHistory(a.ctx)
}

// GetWorkHistory returns a single work history entry by ID with
// its bullets included.
func (a *App) GetWorkHistory(id int64) (domain.WorkHistoryEntry, error) {
	return a.workHistorySvc.GetWorkHistory(a.ctx, id)
}

// CreateWorkHistory creates a new work history entry.
// Returns the created entry with generated ID.
func (a *App) CreateWorkHistory(entry domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	result, err := a.workHistorySvc.CreateWorkHistory(a.ctx, entry)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateWorkHistory updates an existing work history entry.
func (a *App) UpdateWorkHistory(id int64, entry domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	result, err := a.workHistorySvc.UpdateWorkHistory(a.ctx, id, entry)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteWorkHistory deletes a work history entry and its bullets.
func (a *App) DeleteWorkHistory(id int64) error {
	err := a.workHistorySvc.DeleteWorkHistory(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ReorderWorkHistory updates the sort_order of all entries.
// Accepts a slice of IDs in the desired order.
func (a *App) ReorderWorkHistory(orderedIDs []int64) error {
	err := a.workHistorySvc.ReorderWorkHistory(a.ctx, orderedIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// CreateBullet adds an achievement bullet to a work history entry.
// bulletType should be "primary" or "secondary" (defaults to "primary").
func (a *App) CreateBullet(workHistoryID int64, text string, bulletType string) (domain.AchievementBullet, error) {
	result, err := a.workHistorySvc.CreateBullet(a.ctx, workHistoryID, text, bulletType)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateBullet updates an achievement bullet's text.
func (a *App) UpdateBullet(id int64, text string) (domain.AchievementBullet, error) {
	result, err := a.workHistorySvc.UpdateBullet(a.ctx, id, text)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteBullet deletes an achievement bullet.
func (a *App) DeleteBullet(id int64) error {
	err := a.workHistorySvc.DeleteBullet(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ReorderBullets updates the sort_order of bullets within an entry.
func (a *App) ReorderBullets(workHistoryID int64, orderedIDs []int64) error {
	err := a.workHistorySvc.ReorderBullets(a.ctx, workHistoryID, orderedIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// SplitBulletText accepts a block of text and returns individual
// lines for preview before creating bullets.
func (a *App) SplitBulletText(text string) []string {
	return a.workHistorySvc.SplitBulletText(text)
}

// =================================================================
// Profile Bindings
// =================================================================

// GetProfile returns the user's profile. Creates a default empty
// profile if none exists.
func (a *App) GetProfile() (domain.UserProfile, error) {
	return a.profileSvc.GetProfile(a.ctx)
}

// UpdateProfile updates the user's profile fields.
// Returns the updated profile.
func (a *App) UpdateProfile(profile domain.UserProfile) (domain.UserProfile, error) {
	result, err := a.profileSvc.UpdateProfile(a.ctx, profile)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// =================================================================
// Profile Link Bindings
// =================================================================

// ListProfileLinks returns all profile links ordered by sort_order.
func (a *App) ListProfileLinks() ([]domain.ProfileLink, error) {
	return a.profileSvc.ListProfileLinks(a.ctx)
}

// CreateProfileLink creates a new profile link.
func (a *App) CreateProfileLink(input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	result, err := a.profileSvc.CreateProfileLink(a.ctx, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateProfileLink updates an existing profile link.
func (a *App) UpdateProfileLink(id int64, input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	result, err := a.profileSvc.UpdateProfileLink(a.ctx, id, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteProfileLink deletes a profile link.
func (a *App) DeleteProfileLink(id int64) error {
	err := a.profileSvc.DeleteProfileLink(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ReorderProfileLinks updates the sort_order of all profile links.
func (a *App) ReorderProfileLinks(orderedIDs []int64) error {
	err := a.profileSvc.ReorderProfileLinks(a.ctx, orderedIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// =================================================================
// Skills Bindings
// =================================================================

// ListSkills returns all skills sorted by competence level (desc),
// then alphabetically.
func (a *App) ListSkills() ([]domain.Skill, error) {
	return a.skillsSvc.ListSkills(a.ctx)
}

// ListSkillsByCategory returns skills grouped by category, with
// categories ordered by their sort_order.
func (a *App) ListSkillsByCategory() ([]domain.SkillCategoryWithSkills, error) {
	return a.skillsSvc.ListSkillsByCategory(a.ctx)
}

// CreateSkill creates a new skill.
func (a *App) CreateSkill(skill domain.SkillInput) (domain.Skill, error) {
	result, err := a.skillsSvc.CreateSkill(a.ctx, skill)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateSkill updates an existing skill.
func (a *App) UpdateSkill(id int64, skill domain.SkillInput) (domain.Skill, error) {
	result, err := a.skillsSvc.UpdateSkill(a.ctx, id, skill)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteSkill deletes a skill.
func (a *App) DeleteSkill(id int64) error {
	err := a.skillsSvc.DeleteSkill(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// CheckSkillLensReferences returns the names of lenses that
// reference a given skill, for use in delete confirmation dialogs.
func (a *App) CheckSkillLensReferences(id int64) ([]string, error) {
	return a.skillsSvc.CheckSkillLensReferences(a.ctx, id)
}

// SplitSkillsText accepts a comma-separated string and returns
// individual skill names for preview before creating.
func (a *App) SplitSkillsText(text string) []string {
	return a.skillsSvc.SplitSkillsText(text)
}

// GetCompetenceLevels returns the fixed competence scale with
// descriptive criteria for each level.
func (a *App) GetCompetenceLevels() []domain.CompetenceLevel {
	return a.skillsSvc.GetCompetenceLevels()
}

// =================================================================
// Skill Category Bindings
// =================================================================

// ListSkillCategories returns all skill categories ordered by
// sort_order.
func (a *App) ListSkillCategories() ([]domain.SkillCategory, error) {
	return a.skillsSvc.ListSkillCategories(a.ctx)
}

// CreateSkillCategory creates a new skill category.
func (a *App) CreateSkillCategory(name string) (domain.SkillCategory, error) {
	result, err := a.skillsSvc.CreateSkillCategory(a.ctx, name)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// RenameSkillCategory updates a category's name.
func (a *App) RenameSkillCategory(id int64, name string) (domain.SkillCategory, error) {
	result, err := a.skillsSvc.RenameSkillCategory(a.ctx, id, name)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteSkillCategory deletes a skill category. Fails if any skills
// still reference it.
func (a *App) DeleteSkillCategory(id int64) error {
	err := a.skillsSvc.DeleteSkillCategory(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ReorderSkillCategories updates the sort_order of all categories.
func (a *App) ReorderSkillCategories(orderedIDs []int64) error {
	err := a.skillsSvc.ReorderSkillCategories(a.ctx, orderedIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// =================================================================
// Academic Credential Bindings
// =================================================================

// ListAcademicCredentials returns all academic records ordered by
// sort_order.
func (a *App) ListAcademicCredentials() ([]domain.AcademicCredential, error) {
	return a.academicSvc.ListAcademicCredentials(a.ctx)
}

// CreateAcademicCredential creates a new academic record.
func (a *App) CreateAcademicCredential(cred domain.AcademicInput) (domain.AcademicCredential, error) {
	result, err := a.academicSvc.CreateAcademicCredential(a.ctx, cred)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateAcademicCredential updates an academic record.
func (a *App) UpdateAcademicCredential(id int64, cred domain.AcademicInput) (domain.AcademicCredential, error) {
	result, err := a.academicSvc.UpdateAcademicCredential(a.ctx, id, cred)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteAcademicCredential deletes an academic record.
func (a *App) DeleteAcademicCredential(id int64) error {
	err := a.academicSvc.DeleteAcademicCredential(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ReorderAcademicCredentials updates sort_order for all academic
// credentials.
func (a *App) ReorderAcademicCredentials(orderedIDs []int64) error {
	err := a.academicSvc.ReorderAcademicCredentials(a.ctx, orderedIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// =================================================================
// Certification Bindings
// =================================================================

// ListCertifications returns all certifications with computed
// active/inactive status.
func (a *App) ListCertifications() ([]domain.Certification, error) {
	return a.academicSvc.ListCertifications(a.ctx)
}

// CreateCertification creates a new certification.
func (a *App) CreateCertification(cert domain.CertificationInput) (domain.Certification, error) {
	result, err := a.academicSvc.CreateCertification(a.ctx, cert)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateCertification updates a certification.
func (a *App) UpdateCertification(id int64, cert domain.CertificationInput) (domain.Certification, error) {
	result, err := a.academicSvc.UpdateCertification(a.ctx, id, cert)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteCertification deletes a certification.
func (a *App) DeleteCertification(id int64) error {
	err := a.academicSvc.DeleteCertification(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ReorderCertifications updates sort_order for all certifications.
func (a *App) ReorderCertifications(orderedIDs []int64) error {
	err := a.academicSvc.ReorderCertifications(a.ctx, orderedIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// =================================================================
// Professional Summary Bindings
// =================================================================

// ListSummaries returns all professional summary variants.
func (a *App) ListSummaries() ([]domain.ProfessionalSummary, error) {
	return a.summarySvc.ListSummaries(a.ctx)
}

// CreateSummary creates a new summary variant.
func (a *App) CreateSummary(summary domain.SummaryInput) (domain.ProfessionalSummary, error) {
	result, err := a.summarySvc.CreateSummary(a.ctx, summary)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateSummary updates an existing summary.
func (a *App) UpdateSummary(id int64, summary domain.SummaryInput) (domain.ProfessionalSummary, error) {
	result, err := a.summarySvc.UpdateSummary(a.ctx, id, summary)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteSummary deletes a summary variant.
func (a *App) DeleteSummary(id int64) error {
	err := a.summarySvc.DeleteSummary(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// =================================================================
// Role Descriptor Bindings
// =================================================================

// ListDescriptors returns all role descriptors ordered by sort_order.
func (a *App) ListDescriptors() ([]domain.RoleDescriptor, error) {
	return a.descriptorSvc.ListDescriptors(a.ctx)
}

// CreateDescriptor creates a new role descriptor.
func (a *App) CreateDescriptor(title string) (domain.RoleDescriptor, error) {
	result, err := a.descriptorSvc.CreateDescriptor(a.ctx, title)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateDescriptor updates a descriptor's title.
func (a *App) UpdateDescriptor(id int64, title string) (domain.RoleDescriptor, error) {
	result, err := a.descriptorSvc.UpdateDescriptor(a.ctx, id, title)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteDescriptor deletes a role descriptor.
func (a *App) DeleteDescriptor(id int64) error {
	err := a.descriptorSvc.DeleteDescriptor(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ReorderDescriptors updates sort_order for all descriptors.
func (a *App) ReorderDescriptors(orderedIDs []int64) error {
	err := a.descriptorSvc.ReorderDescriptors(a.ctx, orderedIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// =================================================================
// Core Expertise Bindings
// =================================================================

// ListCoreExpertise returns all core expertise items ordered by
// sort_order.
func (a *App) ListCoreExpertise() ([]domain.CoreExpertise, error) {
	return a.coreExpertiseSvc.ListCoreExpertise(a.ctx)
}

// CreateCoreExpertise creates a new core expertise item.
func (a *App) CreateCoreExpertise(label string) (domain.CoreExpertise, error) {
	result, err := a.coreExpertiseSvc.CreateCoreExpertise(a.ctx, label)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateCoreExpertise updates a core expertise item's label.
func (a *App) UpdateCoreExpertise(id int64, label string) (domain.CoreExpertise, error) {
	result, err := a.coreExpertiseSvc.UpdateCoreExpertise(a.ctx, id, label)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteCoreExpertise deletes a core expertise item.
func (a *App) DeleteCoreExpertise(id int64) error {
	err := a.coreExpertiseSvc.DeleteCoreExpertise(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ReorderCoreExpertise updates sort_order for all core expertise
// items.
func (a *App) ReorderCoreExpertise(orderedIDs []int64) error {
	err := a.coreExpertiseSvc.ReorderCoreExpertise(a.ctx, orderedIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// CheckCoreExpertiseLensReferences returns the names of lenses that
// reference a given core expertise item, for delete confirmation.
func (a *App) CheckCoreExpertiseLensReferences(id int64) ([]string, error) {
	return a.lensSvc.CheckCoreExpertiseLensReferences(a.ctx, id)
}

// SetLensCoreExpertise replaces the core expertise selections for
// a lens.
func (a *App) SetLensCoreExpertise(lensID int64, selections []domain.LensCoreExpertiseItem) error {
	err := a.lensSvc.SetLensCoreExpertise(a.ctx, lensID, selections)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// SplitCoreExpertiseText accepts a block of text and returns
// individual core expertise labels by splitting on pipe, comma,
// or newline delimiters.
func (a *App) SplitCoreExpertiseText(text string) []string {
	return a.coreExpertiseSvc.SplitCoreExpertiseText(text)
}

// =================================================================
// Resume Export Bindings
// =================================================================

// ListTemplates returns the available built-in resume templates.
func (a *App) ListTemplates() []domain.ResumeTemplate {
	return a.resumeSvc.ListTemplates()
}

// PreviewExport generates a resume PDF with the given selections
// without creating an export record. Returns the file path.
func (a *App) PreviewExport(req domain.ExportRequest) (string, error) {
	return a.resumeSvc.PreviewExport(a.ctx, req)
}

// CreateExport generates a PDF resume, saves it, and creates an
// export record with a snapshot of the selections used.
func (a *App) CreateExport(req domain.ExportRequest) (domain.ResumeExport, error) {
	result, err := a.resumeSvc.CreateExport(a.ctx, req)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// OverwriteExport regenerates and overwrites an existing export in-place.
func (a *App) OverwriteExport(exportID int64, req domain.ExportRequest) (domain.ResumeExport, error) {
	result, err := a.resumeSvc.OverwriteExport(a.ctx, exportID, req)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// ListExports returns all previous export records.
func (a *App) ListExports() ([]domain.ResumeExport, error) {
	return a.resumeSvc.ListExports(a.ctx)
}

// OpenExportFile opens the PDF file for an export in the system's
// default viewer.
func (a *App) OpenExportFile(exportID int64) error {
	export, err := a.resumeSvc.GetExport(a.ctx, exportID)
	if err != nil {
		return fmt.Errorf("get export: %w", err)
	}

	absPath, err := filepath.Abs(export.FilePath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", absPath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", absPath)
	default: // linux and others
		cmd = exec.Command("xdg-open", absPath)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	return nil
}

// =================================================================
// Cover Letter Bindings
// =================================================================

// ListCoverLetters returns all cover letters.
func (a *App) ListCoverLetters() ([]domain.CoverLetter, error) {
	return a.coverLetterSvc.ListCoverLetters(a.ctx)
}

// CreateCoverLetter creates a new cover letter.
func (a *App) CreateCoverLetter(input domain.CoverLetterInput) (domain.CoverLetter, error) {
	result, err := a.coverLetterSvc.CreateCoverLetter(a.ctx, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateCoverLetter updates a cover letter's content.
func (a *App) UpdateCoverLetter(id int64, input domain.CoverLetterInput) (domain.CoverLetter, error) {
	result, err := a.coverLetterSvc.UpdateCoverLetter(a.ctx, id, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteCoverLetter deletes a cover letter.
func (a *App) DeleteCoverLetter(id int64) error {
	err := a.coverLetterSvc.DeleteCoverLetter(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ExportCoverLetter generates a PDF of the cover letter.
// Returns the file path of the generated PDF.
func (a *App) ExportCoverLetter(id int64) (string, error) {
	cl, err := a.coverLetterSvc.ExportCoverLetter(a.ctx, id)
	if err != nil {
		return "", err
	}
	return cl.FilePath, nil
}

// =================================================================
// Job Application Bindings
// =================================================================

// ListApplications returns all job applications with current status.
func (a *App) ListApplications() ([]domain.JobApplication, error) {
	return a.applicationSvc.ListApplications(a.ctx)
}

// SearchApplications searches by company name or position title.
func (a *App) SearchApplications(query string) ([]domain.JobApplication, error) {
	return a.applicationSvc.SearchApplications(a.ctx, query)
}

// CreateApplication creates a new job application record.
func (a *App) CreateApplication(input domain.ApplicationInput) (domain.JobApplication, error) {
	result, err := a.applicationSvc.CreateApplication(a.ctx, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateApplication updates application fields (not status).
func (a *App) UpdateApplication(id int64, input domain.ApplicationInput) (domain.JobApplication, error) {
	result, err := a.applicationSvc.UpdateApplication(a.ctx, id, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateApplicationStatus changes the status and records history.
func (a *App) UpdateApplicationStatus(id int64, newStatus string) (domain.JobApplication, error) {
	result, err := a.applicationSvc.UpdateApplicationStatus(a.ctx, id, newStatus)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateApplicationFit updates the fit indicator.
func (a *App) UpdateApplicationFit(id int64, fitIndicator string) (domain.JobApplication, error) {
	result, err := a.applicationSvc.UpdateApplicationFit(a.ctx, id, fitIndicator)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// GetApplicationHistory returns the status change history for an
// application.
func (a *App) GetApplicationHistory(id int64) ([]domain.StatusChange, error) {
	return a.applicationSvc.GetApplicationHistory(a.ctx, id)
}

// DeleteApplication deletes a job application and its history.
func (a *App) DeleteApplication(id int64) error {
	err := a.applicationSvc.DeleteApplication(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// GetApplicationStatuses returns the fixed list of valid statuses.
func (a *App) GetApplicationStatuses() []string {
	return a.applicationSvc.GetApplicationStatuses()
}

// GetFitIndicators returns the fixed list of fit indicator values.
func (a *App) GetFitIndicators() []string {
	return a.applicationSvc.GetFitIndicators()
}

// GetApplicationPromptValues returns saved cover-letter prompt values
// for an application+template pair.
func (a *App) GetApplicationPromptValues(applicationID int64, templateID int64) (map[string]string, error) {
	return a.applicationSvc.GetApplicationPromptValues(a.ctx, applicationID, templateID)
}

// SaveApplicationPromptValues replaces saved cover-letter prompt values
// for an application+template pair.
func (a *App) SaveApplicationPromptValues(applicationID int64, templateID int64, values map[string]string) error {
	err := a.applicationSvc.SaveApplicationPromptValues(a.ctx, applicationID, templateID, values)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// =================================================================
// Lens Bindings
// =================================================================

// ListLenses returns all lenses.
func (a *App) ListLenses() ([]domain.Lens, error) {
	return a.lensSvc.ListLenses(a.ctx)
}

// GetLens returns a single lens with all its content selections.
func (a *App) GetLens(id int64) (domain.LensDetail, error) {
	return a.lensSvc.GetLens(a.ctx, id)
}

// CreateLens creates a new lens. Returns the created lens.
func (a *App) CreateLens(input domain.LensInput) (domain.Lens, error) {
	result, err := a.lensSvc.CreateLens(a.ctx, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateLens updates a lens's name and summary selection.
func (a *App) UpdateLens(id int64, input domain.LensInput) (domain.Lens, error) {
	result, err := a.lensSvc.UpdateLens(a.ctx, id, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteLens deletes a lens and all its selections.
func (a *App) DeleteLens(id int64) error {
	err := a.lensSvc.DeleteLens(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// SetLensWorkHistory replaces the work history selections for a lens.
func (a *App) SetLensWorkHistory(lensID int64, selections []domain.LensWorkHistoryItem) error {
	err := a.lensSvc.SetLensWorkHistory(a.ctx, lensID, selections)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// SetLensSummaries replaces the summary selections for a lens.
func (a *App) SetLensSummaries(lensID int64, selections []domain.LensSummaryItem) error {
	err := a.lensSvc.SetLensSummaries(a.ctx, lensID, selections)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// SetLensBullets replaces the bullet selections for a lens.
func (a *App) SetLensBullets(lensID int64, selections []domain.LensBulletItem) error {
	err := a.lensSvc.SetLensBullets(a.ctx, lensID, selections)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// SetLensSkills replaces the skill selections for a lens.
func (a *App) SetLensSkills(lensID int64, selections []domain.LensSkillItem) error {
	err := a.lensSvc.SetLensSkills(a.ctx, lensID, selections)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// SetLensAcademics replaces the academic selections for a lens.
func (a *App) SetLensAcademics(lensID int64, academicIDs []int64) error {
	err := a.lensSvc.SetLensAcademics(a.ctx, lensID, academicIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// SetLensCerts replaces the certification selections for a lens.
func (a *App) SetLensCerts(lensID int64, certIDs []int64) error {
	err := a.lensSvc.SetLensCerts(a.ctx, lensID, certIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// SetLensDescriptors replaces the descriptor selections for a lens.
func (a *App) SetLensDescriptors(lensID int64, selections []domain.LensDescriptorItem) error {
	err := a.lensSvc.SetLensDescriptors(a.ctx, lensID, selections)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// GetLensExportSelections returns the full content selections for a
// lens, formatted as an ExportRequest that can pre-fill the export
// dialog.
func (a *App) GetLensExportSelections(lensID int64) (domain.ExportRequest, error) {
	return a.lensSvc.GetLensExportSelections(a.ctx, lensID)
}

// =================================================================
// Lens Reference Check Bindings
// =================================================================

// CheckWorkHistoryLensReferences returns the names of lenses that
// reference a given work history entry, for delete confirmation.
func (a *App) CheckWorkHistoryLensReferences(id int64) ([]string, error) {
	return a.lensSvc.CheckWorkHistoryLensReferences(a.ctx, id)
}

// CheckBulletLensReferences returns the names of lenses that
// reference a given bullet, for delete confirmation.
func (a *App) CheckBulletLensReferences(id int64) ([]string, error) {
	return a.lensSvc.CheckBulletLensReferences(a.ctx, id)
}

// CheckAcademicLensReferences returns the names of lenses that
// reference a given academic credential, for delete confirmation.
func (a *App) CheckAcademicLensReferences(id int64) ([]string, error) {
	return a.lensSvc.CheckAcademicLensReferences(a.ctx, id)
}

// CheckCertLensReferences returns the names of lenses that
// reference a given certification, for delete confirmation.
func (a *App) CheckCertLensReferences(id int64) ([]string, error) {
	return a.lensSvc.CheckCertLensReferences(a.ctx, id)
}

// CheckDescriptorLensReferences returns the names of lenses that
// reference a given role descriptor, for delete confirmation.
func (a *App) CheckDescriptorLensReferences(id int64) ([]string, error) {
	return a.lensSvc.CheckDescriptorLensReferences(a.ctx, id)
}

// CheckSummaryLensReferences returns the names of lenses that
// reference a given professional summary, for delete confirmation.
func (a *App) CheckSummaryLensReferences(id int64) ([]string, error) {
	return a.lensSvc.CheckSummaryLensReferences(a.ctx, id)
}

// =================================================================
// Skill Lens Tag Bindings
// =================================================================

// GetSkillLensTags returns all lens IDs tagged for a skill.
func (a *App) GetSkillLensTags(skillID int64) ([]int64, error) {
	return a.lensSvc.GetSkillLensTags(a.ctx, skillID)
}

// SetSkillLensTags replaces all lens tags for a skill.
func (a *App) SetSkillLensTags(skillID int64, lensIDs []int64) error {
	err := a.lensSvc.SetSkillLensTags(a.ctx, skillID, lensIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ListSkillsWithLensTags returns all skills with their lens tag
// associations included.
func (a *App) ListSkillsWithLensTags() ([]domain.SkillWithTags, error) {
	return a.lensSvc.ListSkillsWithLensTags(a.ctx)
}

// =================================================================
// Document Template Bindings
// =================================================================

// ListDocumentTemplates returns all document templates ordered by
// is_builtin DESC, name ASC. Built-in templates appear first.
func (a *App) ListDocumentTemplates() ([]domain.DocumentTemplate, error) {
	return a.templateSvc.ListDocumentTemplates(a.ctx)
}

// GetDocumentTemplate returns a template with all its elements.
func (a *App) GetDocumentTemplate(id int64) (domain.TemplateDetail, error) {
	return a.templateSvc.GetDocumentTemplate(a.ctx, id)
}

// CreateDocumentTemplate creates a new user document template.
func (a *App) CreateDocumentTemplate(input domain.DocumentTemplateInput) (domain.DocumentTemplate, error) {
	result, err := a.templateSvc.CreateDocumentTemplate(a.ctx, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateDocumentTemplate updates a user template's metadata.
func (a *App) UpdateDocumentTemplate(id int64, input domain.DocumentTemplateInput) (domain.DocumentTemplate, error) {
	result, err := a.templateSvc.UpdateDocumentTemplate(a.ctx, id, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteDocumentTemplate deletes a user template. Built-in
// templates cannot be deleted.
func (a *App) DeleteDocumentTemplate(id int64) error {
	err := a.templateSvc.DeleteDocumentTemplate(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// DuplicateDocumentTemplate creates a copy of a template with all
// its elements. The copy is always user-created (is_builtin=false).
func (a *App) DuplicateDocumentTemplate(id int64, newName string) (domain.DocumentTemplate, error) {
	result, err := a.templateSvc.DuplicateDocumentTemplate(a.ctx, id, newName)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// CreateTemplateElement adds a new element to a template, appended
// to the end of the sort order within its parent scope.
func (a *App) CreateTemplateElement(templateID int64, input domain.TemplateElementInput) (domain.TemplateElement, error) {
	result, err := a.templateSvc.CreateTemplateElement(a.ctx, templateID, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// UpdateTemplateElement updates an element's type and config.
func (a *App) UpdateTemplateElement(id int64, input domain.TemplateElementInput) (domain.TemplateElement, error) {
	result, err := a.templateSvc.UpdateTemplateElement(a.ctx, id, input)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// DeleteTemplateElement removes an element from a template.
func (a *App) DeleteTemplateElement(id int64) error {
	err := a.templateSvc.DeleteTemplateElement(a.ctx, id)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ReorderTemplateElements updates sort_order for elements within
// a specific parent scope (top-level if parentID is nil, or within
// a specific loop container).
func (a *App) ReorderTemplateElements(templateID int64, parentID *int64, orderedIDs []int64) error {
	err := a.templateSvc.ReorderTemplateElements(a.ctx, templateID, parentID, orderedIDs)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// PreviewTemplate generates a preview PDF for a template using all
// available user data (resume templates) or placeholder substitutions
// (cover letter templates). Returns the file path of the generated PDF.
func (a *App) PreviewTemplate(templateID int64) (string, error) {
	return a.resumeSvc.PreviewTemplate(a.ctx, templateID)
}

// OpenFile opens a file at the given path in the system's default
// application (e.g., PDF viewer for .pdf files).
func (a *App) OpenFile(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", absPath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", absPath)
	default:
		cmd = exec.Command("xdg-open", absPath)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	return nil
}

// ParseTemplateVariables scans a cover letter template for variable
// placeholders ({{var}}) and guided prompts ({{prompt: text}}).
func (a *App) ParseTemplateVariables(templateID int64) (domain.TemplateVariables, error) {
	detail, err := a.templateSvc.GetDocumentTemplate(a.ctx, templateID)
	if err != nil {
		return domain.TemplateVariables{}, err
	}
	return a.templateSvc.ParseTemplateVariables(detail), nil
}

// ExportTemplate serializes a template and all its elements to a
// standalone JSON file at the given path.
func (a *App) ExportTemplate(templateID int64, outputPath string) error {
	return a.templateSvc.ExportTemplate(a.ctx, templateID, outputPath)
}

// ImportTemplate reads a template JSON file and creates a new
// user template with fresh IDs. Returns the created template.
func (a *App) ImportTemplate(inputPath string) (domain.DocumentTemplate, error) {
	return a.templateSvc.ImportTemplate(a.ctx, inputPath)
}

// =================================================================
// Data Management Bindings
// =================================================================

// ExportAllData exports all user data to a JSON file at the specified
// path. Returns the file path.
func (a *App) ExportAllData(outputPath string) (string, error) {
	result, err := a.backupSvc.ExportAllData(a.ctx, outputPath)
	if err != nil {
		a.emitBackupError(err)
		return result, err
	}
	a.emitBackupComplete(result)
	return result, nil
}

// ImportAllData restores all user data from a JSON backup file.
// This replaces all existing data.
func (a *App) ImportAllData(inputPath string) error {
	err := a.backupSvc.ImportAllData(a.ctx, inputPath)
	if err == nil {
		a.emitAutosave()
	}
	return err
}

// ImportCSV imports structured data from a CSV file. The dataType
// parameter specifies what is being imported: "work_history",
// "skills", "academic", "certifications".
func (a *App) ImportCSV(filePath string, dataType string) (domain.ImportResult, error) {
	result, err := a.backupSvc.ImportCSV(a.ctx, filePath, dataType)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// ImportJSON imports structured data from a JSON file (partial, not
// full backup restore).
func (a *App) ImportJSON(filePath string, dataType string) (domain.ImportResult, error) {
	result, err := a.backupSvc.ImportJSON(a.ctx, filePath, dataType)
	if err == nil {
		a.emitAutosave()
	}
	return result, err
}

// GetDataDirectory returns the current data directory path.
func (a *App) GetDataDirectory() string {
	dir, err := a.cfg.ResolveDataDir()
	if err != nil {
		return ""
	}
	return dir
}

// SetDataDirectory changes the data directory. Requires app restart.
func (a *App) SetDataDirectory(path string) error {
	a.cfg.DataDirectory = path
	return a.cfg.Save()
}

// GetBackupSettings returns current backup configuration.
func (a *App) GetBackupSettings() domain.BackupSettings {
	return domain.BackupSettings{
		RollingBackupCount: a.cfg.Backup.RollingBackupCount,
	}
}

// UpdateBackupSettings updates backup configuration.
func (a *App) UpdateBackupSettings(settings domain.BackupSettings) error {
	if settings.RollingBackupCount <= 0 {
		settings.RollingBackupCount = config.DefaultRollingBackupCount
	}
	a.cfg.Backup.RollingBackupCount = settings.RollingBackupCount
	return a.cfg.Save()
}

// OpenDataDirectory opens the data directory in the system file
// manager.
func (a *App) OpenDataDirectory() error {
	dir, err := a.cfg.ResolveDataDir()
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", dir)
	case "linux":
		cmd = exec.Command("xdg-open", dir)
	default:
		cmd = exec.Command("explorer", dir)
	}

	return cmd.Start()
}
