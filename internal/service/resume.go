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
	ListWorkHistory(ctx context.Context) ([]domain.WorkHistoryEntry, error)
	ListSkills(ctx context.Context) ([]domain.Skill, error)
	ListAcademicCredentials(ctx context.Context) ([]domain.AcademicCredential, error)
	ListCertifications(ctx context.Context) ([]domain.Certification, error)
	ListDescriptors(ctx context.Context) ([]domain.RoleDescriptor, error)

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

// PreviewExport generates a PDF without creating an export record.
// Returns the file path of the generated PDF.
func (s *ResumeService) PreviewExport(
	ctx context.Context,
	req domain.ExportRequest,
) (string, error) {
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
// snapshots the content selections.
func (s *ResumeService) CreateExport(
	ctx context.Context,
	req domain.ExportRequest,
) (domain.ResumeExport, error) {
	if err := validateExportRequest(req); err != nil {
		return domain.ResumeExport{}, err
	}

	renderReq, err := s.assembleRenderRequest(ctx, req)
	if err != nil {
		return domain.ResumeExport{}, err
	}

	filePath, err := s.renderer.RenderResume(ctx, renderReq)
	if err != nil {
		return domain.ResumeExport{}, fmt.Errorf("render resume: %w", err)
	}

	export, err := s.store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: req.TemplateID,
		FilePath:   filePath,
		SummaryID:  req.SummaryID,
		LensID:     req.LensID,
	})
	if err != nil {
		return domain.ResumeExport{}, fmt.Errorf("create export record: %w", err)
	}

	if err := s.store.CreateExportSelections(ctx, export.ID, req); err != nil {
		return domain.ResumeExport{}, fmt.Errorf("save export selections: %w", err)
	}

	return export, nil
}

// validateExportRequest checks that the export request has at least
// one content item selected and a template ID.
func validateExportRequest(req domain.ExportRequest) error {
	if req.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}

	hasContent := len(req.WorkHistoryIDs) > 0 ||
		len(req.SkillIDs) > 0 ||
		len(req.AcademicIDs) > 0 ||
		len(req.CertificationIDs) > 0 ||
		len(req.DescriptorIDs) > 0
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

	// Fetch optional summary.
	var summary *domain.ProfessionalSummary
	if req.SummaryID != nil {
		s, err := s.store.GetSummary(ctx, *req.SummaryID)
		if err != nil {
			return domain.RenderResumeRequest{}, fmt.Errorf("get summary: %w", err)
		}
		summary = &s
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

	return domain.RenderResumeRequest{
		TemplateID:  req.TemplateID,
		OutputDir:   s.outputDir,
		Profile:     profile,
		Links:       links,
		Summary:     summary,
		WorkHistory: workHistory,
		Skills:      skills,
		Academics:   academics,
		Certs:       certs,
		Descriptors: descriptors,
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
		return nil, nil
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
			var filtered []domain.AchievementBullet
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
		return nil, nil
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
	var selected []domain.Skill
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

// filterAcademics fetches all academic credentials and keeps those
// in the selected IDs, preserving request order.
func (s *ResumeService) filterAcademics(
	ctx context.Context,
	academicIDs []int64,
) ([]domain.AcademicCredential, error) {
	if len(academicIDs) == 0 {
		return nil, nil
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
		return nil, nil
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
		return nil, nil
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
