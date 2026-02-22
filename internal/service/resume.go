package service

import (
	"context"
	"fmt"
	"sort"

	"cut-the-bs/internal/domain"
)

// ExportStore defines the persistence operations required by
// ResumeService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type ExportStore interface {
	// Profile
	GetProfile(ctx context.Context) (domain.UserProfile, error)
	ListProfileLinks(ctx context.Context) ([]domain.ProfileLink, error)

	// Content
	GetSummary(ctx context.Context, id int64) (domain.ProfessionalSummary, error)
	ListSummaries(ctx context.Context) ([]domain.ProfessionalSummary, error)
	ListWorkHistory(ctx context.Context) ([]domain.WorkHistoryEntry, error)
	ListSkills(ctx context.Context) ([]domain.Skill, error)
	ListSkillCategories(ctx context.Context) ([]domain.SkillCategory, error)
	ListAcademicCredentials(ctx context.Context) ([]domain.AcademicCredential, error)
	ListCertifications(ctx context.Context) ([]domain.Certification, error)
	ListDescriptors(ctx context.Context) ([]domain.RoleDescriptor, error)
	ListCoreExpertise(ctx context.Context) ([]domain.CoreExpertise, error)

	// Templates
	GetDocumentTemplate(ctx context.Context, id int64) (domain.TemplateDetail, error)

	// Export records
	ListExports(ctx context.Context) ([]domain.ResumeExport, error)
	CreateExport(ctx context.Context, export domain.ResumeExport) (domain.ResumeExport, error)
	GetExport(ctx context.Context, id int64) (domain.ResumeExport, error)
	CreateExportSelections(ctx context.Context, exportID int64, req domain.ExportRequest) error
}

// PDFRendererWithTemplates extends domain.PDFRenderer with template
// listing.
type PDFRendererWithTemplates interface {
	domain.PDFRenderer
	ListTemplates() []domain.ResumeTemplate
}

// ResumeService provides business-logic operations for generating
// and managing resume exports.
type ResumeService struct {
	store     ExportStore
	renderer  PDFRendererWithTemplates
	outputDir string
}

// NewResumeService creates a ResumeService backed by the given
// store and renderer. outputDir is where generated PDFs are saved.
func NewResumeService(
	store ExportStore,
	renderer PDFRendererWithTemplates,
	outputDir string,
) *ResumeService {
	return &ResumeService{
		store:     store,
		renderer:  renderer,
		outputDir: outputDir,
	}
}

// ListTemplates returns the available resume templates.
func (s *ResumeService) ListTemplates() []domain.ResumeTemplate {
	return s.renderer.ListTemplates()
}

// ListExports returns all previous export records.
func (s *ResumeService) ListExports(
	ctx context.Context,
) ([]domain.ResumeExport, error) {
	return s.store.ListExports(ctx)
}

// GetExport returns a single export record by ID.
func (s *ResumeService) GetExport(
	ctx context.Context,
	id int64,
) (domain.ResumeExport, error) {
	return s.store.GetExport(ctx, id)
}

// PreviewTemplate generates a preview PDF for a template using all
// available user data (no content selection filtering). For resume
// templates, all summaries, work history, skills, etc. are included.
// For cover letter templates, placeholder values are substituted for
// variables (e.g., "{{company_name}}" becomes "[Company Name]").
// Returns the file path of the generated PDF.
func (s *ResumeService) PreviewTemplate(
	ctx context.Context,
	templateID int64,
) (string, error) {
	if templateID == 0 {
		return "", fmt.Errorf("template ID is required")
	}

	tmpl, err := s.store.GetDocumentTemplate(ctx, templateID)
	if err != nil {
		return "", fmt.Errorf("get template: %w", err)
	}

	if tmpl.TemplateType == domain.TemplateTypeCoverLetter {
		return s.previewCoverLetterTemplate(ctx, tmpl)
	}

	return s.previewResumeTemplate(ctx, tmpl)
}

// previewResumeTemplate renders a resume template with all user data.
func (s *ResumeService) previewResumeTemplate(
	ctx context.Context,
	tmpl domain.TemplateDetail,
) (string, error) {
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return "", fmt.Errorf("get profile: %w", err)
	}

	links, err := s.store.ListProfileLinks(ctx)
	if err != nil {
		return "", fmt.Errorf("list profile links: %w", err)
	}

	summaries, err := s.store.ListSummaries(ctx)
	if err != nil {
		return "", fmt.Errorf("list summaries: %w", err)
	}

	workHistory, err := s.store.ListWorkHistory(ctx)
	if err != nil {
		return "", fmt.Errorf("list work history: %w", err)
	}

	skills, err := s.store.ListSkills(ctx)
	if err != nil {
		return "", fmt.Errorf("list skills: %w", err)
	}

	skillCatNames, err := s.loadSkillCategoryNames(ctx)
	if err != nil {
		return "", err
	}

	academics, err := s.store.ListAcademicCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("list academics: %w", err)
	}

	certs, err := s.store.ListCertifications(ctx)
	if err != nil {
		return "", fmt.Errorf("list certifications: %w", err)
	}

	descriptors, err := s.store.ListDescriptors(ctx)
	if err != nil {
		return "", fmt.Errorf("list descriptors: %w", err)
	}

	coreExpertise, err := s.store.ListCoreExpertise(ctx)
	if err != nil {
		return "", fmt.Errorf("list core expertise: %w", err)
	}

	renderReq := domain.RenderResumeRequest{
		Template:           &tmpl,
		OutputDir:          s.outputDir,
		Profile:            profile,
		Links:              links,
		Summaries:          summaries,
		WorkHistory:        workHistory,
		Skills:             skills,
		SkillCategoryNames: skillCatNames,
		Academics:          academics,
		Certs:              certs,
		Descriptors:        descriptors,
		CoreExpertise:      coreExpertise,
	}

	return s.renderer.RenderResume(ctx, renderReq)
}

// previewCoverLetterTemplate renders a cover letter template with
// profile data and placeholder substitutions for variables.
func (s *ResumeService) previewCoverLetterTemplate(
	ctx context.Context,
	tmpl domain.TemplateDetail,
) (string, error) {
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return "", fmt.Errorf("get profile: %w", err)
	}

	links, err := s.store.ListProfileLinks(ctx)
	if err != nil {
		return "", fmt.Errorf("list profile links: %w", err)
	}

	// Build placeholder substitution map for variables/prompts.
	ts := NewTemplateService(nil)
	vars := ts.ParseTemplateVariables(tmpl)

	subs := make(map[string]string, len(vars.Variables)+len(vars.Prompts))
	for _, v := range vars.Variables {
		subs[v.Name] = "[" + v.Name + "]"
	}
	for _, p := range vars.Prompts {
		subs["prompt:"+p.PromptText] = "[" + p.PromptText + "]"
	}

	clReq := domain.RenderCoverLetterRequest{
		Template:        &tmpl,
		OutputDir:       s.outputDir,
		Profile:         profile,
		Links:           links,
		SubstitutionMap: subs,
	}

	return s.renderer.RenderCoverLetter(ctx, clReq)
}

// PreviewExport generates a PDF without creating an export record.
// Returns the file path of the generated PDF. Supports both resume
// and cover letter templates.
func (s *ResumeService) PreviewExport(
	ctx context.Context,
	req domain.ExportRequest,
) (string, error) {
	if req.TemplateID == 0 {
		return "", fmt.Errorf("template ID is required")
	}

	// Load the template to detect its type.
	tmpl, err := s.store.GetDocumentTemplate(ctx, req.TemplateID)
	if err != nil {
		return "", fmt.Errorf("get template: %w", err)
	}

	if tmpl.TemplateType == domain.TemplateTypeCoverLetter {
		return s.createCoverLetterExport(ctx, req, tmpl)
	}

	if err := validateExportRequest(req); err != nil {
		return "", err
	}

	renderReq, err := s.assembleRenderRequest(ctx, req)
	if err != nil {
		return "", err
	}

	return s.renderer.RenderResume(ctx, renderReq)
}

// CreateExport generates a PDF, creates an export record, and
// snapshots the content selections. Detects the template type
// and routes to either RenderResume or RenderCoverLetter.
func (s *ResumeService) CreateExport(
	ctx context.Context,
	req domain.ExportRequest,
) (domain.ResumeExport, error) {
	if req.TemplateID == 0 {
		return domain.ResumeExport{}, fmt.Errorf("template ID is required")
	}

	// Load the template to detect its type.
	tmpl, err := s.store.GetDocumentTemplate(ctx, req.TemplateID)
	if err != nil {
		return domain.ResumeExport{}, fmt.Errorf("get template: %w", err)
	}

	var filePath string

	if tmpl.TemplateType == domain.TemplateTypeCoverLetter {
		filePath, err = s.createCoverLetterExport(ctx, req, tmpl)
	} else {
		if err := validateExportRequest(req); err != nil {
			return domain.ResumeExport{}, err
		}
		filePath, err = s.createResumeExport(ctx, req, tmpl)
	}
	if err != nil {
		return domain.ResumeExport{}, err
	}

	// For the historical export record, snapshot the first summary ID
	// (if any) for backward compatibility.
	var snapshotSummaryID *int64
	if len(req.SummaryIDs) > 0 {
		sid := req.SummaryIDs[0]
		snapshotSummaryID = &sid
	}

	templateRefID := req.TemplateID
	export, err := s.store.CreateExport(ctx, domain.ResumeExport{
		TemplateID:    tmpl.Name,
		TemplateRefID: &templateRefID,
		FilePath:      filePath,
		SummaryID:     snapshotSummaryID,
		LensID:        req.LensID,
	})
	if err != nil {
		return domain.ResumeExport{}, fmt.Errorf("create export record: %w", err)
	}

	if err := s.store.CreateExportSelections(ctx, export.ID, req); err != nil {
		return domain.ResumeExport{}, fmt.Errorf("save export selections: %w", err)
	}

	return export, nil
}

// createResumeExport assembles and renders a resume PDF.
func (s *ResumeService) createResumeExport(
	ctx context.Context,
	req domain.ExportRequest,
	tmpl domain.TemplateDetail,
) (string, error) {
	renderReq, err := s.assembleRenderRequest(ctx, req)
	if err != nil {
		return "", err
	}

	filePath, err := s.renderer.RenderResume(ctx, renderReq)
	if err != nil {
		return "", fmt.Errorf("render resume: %w", err)
	}

	return filePath, nil
}

// createCoverLetterExport assembles and renders a cover letter PDF.
func (s *ResumeService) createCoverLetterExport(
	ctx context.Context,
	req domain.ExportRequest,
	tmpl domain.TemplateDetail,
) (string, error) {
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return "", fmt.Errorf("get profile: %w", err)
	}

	links, err := s.store.ListProfileLinks(ctx)
	if err != nil {
		return "", fmt.Errorf("list profile links: %w", err)
	}

	clReq := domain.RenderCoverLetterRequest{
		Template:        &tmpl,
		OutputDir:       s.outputDir,
		Profile:         profile,
		Links:           links,
		SubstitutionMap: req.SubstitutionMap,
	}

	filePath, err := s.renderer.RenderCoverLetter(ctx, clReq)
	if err != nil {
		return "", fmt.Errorf("render cover letter: %w", err)
	}

	return filePath, nil
}

// validateExportRequest checks that the export request has at least
// one content item selected and a template ID.
func validateExportRequest(req domain.ExportRequest) error {
	if req.TemplateID == 0 {
		return fmt.Errorf("template ID is required")
	}

	hasContent := len(req.WorkHistoryIDs) > 0 ||
		len(req.SkillIDs) > 0 ||
		len(req.AcademicIDs) > 0 ||
		len(req.CertificationIDs) > 0 ||
		len(req.DescriptorIDs) > 0 ||
		len(req.CoreExpertiseIDs) > 0
	if !hasContent {
		return fmt.Errorf("at least one content item must be selected")
	}

	return nil
}

// assembleRenderRequest fetches all selected data from the store
// and builds a RenderResumeRequest.
func (s *ResumeService) assembleRenderRequest(
	ctx context.Context,
	req domain.ExportRequest,
) (domain.RenderResumeRequest, error) {
	// Fetch profile and links.
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return domain.RenderResumeRequest{}, fmt.Errorf("get profile: %w", err)
	}

	links, err := s.store.ListProfileLinks(ctx)
	if err != nil {
		return domain.RenderResumeRequest{}, fmt.Errorf("list profile links: %w", err)
	}

	// Fetch summaries.
	summaries := make([]domain.ProfessionalSummary, 0, len(req.SummaryIDs))
	for _, sumID := range req.SummaryIDs {
		sum, err := s.store.GetSummary(ctx, sumID)
		if err != nil {
			return domain.RenderResumeRequest{}, fmt.Errorf("get summary %d: %w", sumID, err)
		}
		summaries = append(summaries, sum)
	}

	// Fetch and filter work history.
	workHistory, err := s.filterWorkHistory(ctx, req.WorkHistoryIDs, req.BulletIDs)
	if err != nil {
		return domain.RenderResumeRequest{}, err
	}

	// Fetch and filter skills, applying sort overrides.
	skills, err := s.filterSkills(ctx, req.SkillIDs, req.SkillSortOverrides)
	if err != nil {
		return domain.RenderResumeRequest{}, err
	}

	// Build category name lookup for skill rendering.
	skillCatNames, err := s.loadSkillCategoryNames(ctx)
	if err != nil {
		return domain.RenderResumeRequest{}, err
	}

	// Fetch and filter academics.
	academics, err := s.filterAcademics(ctx, req.AcademicIDs)
	if err != nil {
		return domain.RenderResumeRequest{}, err
	}

	// Fetch and filter certifications.
	certs, err := s.filterCertifications(ctx, req.CertificationIDs)
	if err != nil {
		return domain.RenderResumeRequest{}, err
	}

	// Fetch and filter descriptors (preserving request order).
	descriptors, err := s.filterDescriptors(ctx, req.DescriptorIDs)
	if err != nil {
		return domain.RenderResumeRequest{}, err
	}

	// Fetch and filter core expertise (preserving request order).
	coreExpertise, err := s.filterCoreExpertise(ctx, req.CoreExpertiseIDs)
	if err != nil {
		return domain.RenderResumeRequest{}, err
	}

	// Load the document template for rendering.
	tmpl, err := s.store.GetDocumentTemplate(ctx, req.TemplateID)
	if err != nil {
		return domain.RenderResumeRequest{}, fmt.Errorf("get template: %w", err)
	}

	return domain.RenderResumeRequest{
		Template:           &tmpl,
		OutputDir:          s.outputDir,
		Profile:            profile,
		Links:              links,
		Summaries:          summaries,
		MasterSummaryID:    req.MasterSummaryID,
		WorkHistory:        workHistory,
		Skills:             skills,
		SkillCategoryNames: skillCatNames,
		Academics:          academics,
		Certs:              certs,
		Descriptors:        descriptors,
		CoreExpertise:      coreExpertise,
	}, nil
}

// filterWorkHistory fetches all work history entries, keeps those
// in the selected IDs (preserving request order), and filters
// bullets. If bulletIDs is empty, all bullets for selected entries
// are included.
func (s *ResumeService) filterWorkHistory(
	ctx context.Context,
	whIDs []int64,
	bulletIDs []int64,
) ([]domain.WorkHistoryEntry, error) {
	if len(whIDs) == 0 {
		return []domain.WorkHistoryEntry{}, nil
	}

	allEntries, err := s.store.ListWorkHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("list work history: %w", err)
	}

	// Build lookup map.
	entryByID := make(map[int64]domain.WorkHistoryEntry, len(allEntries))
	for _, e := range allEntries {
		entryByID[e.ID] = e
	}

	// Build bullet filter set (empty means include all).
	bulletSet := make(map[int64]bool, len(bulletIDs))
	for _, id := range bulletIDs {
		bulletSet[id] = true
	}

	// Filter entries in request order.
	result := make([]domain.WorkHistoryEntry, 0, len(whIDs))
	for _, id := range whIDs {
		entry, ok := entryByID[id]
		if !ok {
			continue
		}

		// Filter bullets if specific bullet IDs were provided.
		if len(bulletIDs) > 0 {
			filtered := make([]domain.AchievementBullet, 0)
			for _, b := range entry.Bullets {
				if bulletSet[b.ID] {
					filtered = append(filtered, b)
				}
			}
			entry.Bullets = filtered
		}

		result = append(result, entry)
	}

	return result, nil
}

// filterSkills fetches all skills, keeps those in the selected IDs,
// and applies optional sort overrides (FR-008). Skills without an
// override sort after those with overrides, preserving their
// original order.
func (s *ResumeService) filterSkills(
	ctx context.Context,
	skillIDs []int64,
	overrides map[int64]int,
) ([]domain.Skill, error) {
	if len(skillIDs) == 0 {
		return []domain.Skill{}, nil
	}

	allSkills, err := s.store.ListSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}

	// Build lookup map.
	skillByID := make(map[int64]domain.Skill, len(allSkills))
	for _, sk := range allSkills {
		skillByID[sk.ID] = sk
	}

	// Build selected set for quick lookup.
	selectedSet := make(map[int64]bool, len(skillIDs))
	for _, id := range skillIDs {
		selectedSet[id] = true
	}

	// Collect selected skills preserving original list order.
	selected := make([]domain.Skill, 0)
	for _, sk := range allSkills {
		if selectedSet[sk.ID] {
			selected = append(selected, sk)
		}
	}

	// Apply sort overrides if any.
	if len(overrides) > 0 {
		// maxOverride tracks the highest explicit override value
		// so un-overridden skills sort after all overridden ones.
		maxOverride := 0
		for _, v := range overrides {
			if v > maxOverride {
				maxOverride = v
			}
		}

		sort.SliceStable(selected, func(i, j int) bool {
			oi, okI := overrides[selected[i].ID]
			oj, okJ := overrides[selected[j].ID]
			if okI && okJ {
				return oi < oj
			}
			if okI {
				return true
			}
			if okJ {
				return false
			}
			return false // preserve original order
		})
	}

	return selected, nil
}

// loadSkillCategoryNames fetches all skill categories and returns
// a map from category ID to category name.
func (s *ResumeService) loadSkillCategoryNames(
	ctx context.Context,
) (map[int64]string, error) {
	cats, err := s.store.ListSkillCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list skill categories: %w", err)
	}

	names := make(map[int64]string, len(cats))
	for _, c := range cats {
		names[c.ID] = c.Name
	}
	return names, nil
}

// filterAcademics fetches all academic credentials and keeps those
// in the selected IDs, preserving request order.
func (s *ResumeService) filterAcademics(
	ctx context.Context,
	academicIDs []int64,
) ([]domain.AcademicCredential, error) {
	if len(academicIDs) == 0 {
		return []domain.AcademicCredential{}, nil
	}

	all, err := s.store.ListAcademicCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("list academics: %w", err)
	}

	byID := make(map[int64]domain.AcademicCredential, len(all))
	for _, a := range all {
		byID[a.ID] = a
	}

	result := make([]domain.AcademicCredential, 0, len(academicIDs))
	for _, id := range academicIDs {
		if a, ok := byID[id]; ok {
			result = append(result, a)
		}
	}

	return result, nil
}

// filterCertifications fetches all certifications and keeps those
// in the selected IDs, preserving request order.
func (s *ResumeService) filterCertifications(
	ctx context.Context,
	certIDs []int64,
) ([]domain.Certification, error) {
	if len(certIDs) == 0 {
		return []domain.Certification{}, nil
	}

	all, err := s.store.ListCertifications(ctx)
	if err != nil {
		return nil, fmt.Errorf("list certifications: %w", err)
	}

	byID := make(map[int64]domain.Certification, len(all))
	for _, c := range all {
		byID[c.ID] = c
	}

	result := make([]domain.Certification, 0, len(certIDs))
	for _, id := range certIDs {
		if c, ok := byID[id]; ok {
			result = append(result, c)
		}
	}

	return result, nil
}

// filterDescriptors fetches all descriptors and keeps those in the
// selected IDs, preserving request order.
func (s *ResumeService) filterDescriptors(
	ctx context.Context,
	descriptorIDs []int64,
) ([]domain.RoleDescriptor, error) {
	if len(descriptorIDs) == 0 {
		return []domain.RoleDescriptor{}, nil
	}

	all, err := s.store.ListDescriptors(ctx)
	if err != nil {
		return nil, fmt.Errorf("list descriptors: %w", err)
	}

	byID := make(map[int64]domain.RoleDescriptor, len(all))
	for _, d := range all {
		byID[d.ID] = d
	}

	result := make([]domain.RoleDescriptor, 0, len(descriptorIDs))
	for _, id := range descriptorIDs {
		if d, ok := byID[id]; ok {
			result = append(result, d)
		}
	}

	return result, nil
}

// filterCoreExpertise fetches all core expertise items and keeps
// those in the selected IDs, preserving request order.
func (s *ResumeService) filterCoreExpertise(
	ctx context.Context,
	coreExpertiseIDs []int64,
) ([]domain.CoreExpertise, error) {
	if len(coreExpertiseIDs) == 0 {
		return []domain.CoreExpertise{}, nil
	}

	all, err := s.store.ListCoreExpertise(ctx)
	if err != nil {
		return nil, fmt.Errorf("list core expertise: %w", err)
	}

	byID := make(map[int64]domain.CoreExpertise, len(all))
	for _, ce := range all {
		byID[ce.ID] = ce
	}

	result := make([]domain.CoreExpertise, 0, len(coreExpertiseIDs))
	for _, id := range coreExpertiseIDs {
		if ce, ok := byID[id]; ok {
			result = append(result, ce)
		}
	}

	return result, nil
}
