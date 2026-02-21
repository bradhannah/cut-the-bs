package service

import (
	"context"

	"cut-the-bs/internal/domain"
)

// LensStore defines the persistence operations required by
// LensService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type LensStore interface {
	ListLenses(ctx context.Context) ([]domain.Lens, error)
	GetLens(ctx context.Context, id int64) (domain.LensDetail, error)
	CreateLens(ctx context.Context, input domain.LensInput) (domain.Lens, error)
	UpdateLens(ctx context.Context, id int64, input domain.LensInput) (domain.Lens, error)
	DeleteLens(ctx context.Context, id int64) error

	SetLensWorkHistory(ctx context.Context, lensID int64, selections []domain.LensWorkHistoryItem) error
	SetLensSummaries(ctx context.Context, lensID int64, selections []domain.LensSummaryItem) error
	SetLensBullets(ctx context.Context, lensID int64, selections []domain.LensBulletItem) error
	SetLensSkills(ctx context.Context, lensID int64, selections []domain.LensSkillItem) error
	SetLensAcademics(ctx context.Context, lensID int64, academicIDs []int64) error
	SetLensCerts(ctx context.Context, lensID int64, certIDs []int64) error
	SetLensDescriptors(ctx context.Context, lensID int64, selections []domain.LensDescriptorItem) error
	SetLensCoreExpertise(ctx context.Context, lensID int64, selections []domain.LensCoreExpertiseItem) error

	CheckWorkHistoryLensReferences(ctx context.Context, id int64) ([]string, error)
	CheckBulletLensReferences(ctx context.Context, id int64) ([]string, error)
	CheckAcademicLensReferences(ctx context.Context, id int64) ([]string, error)
	CheckCertLensReferences(ctx context.Context, id int64) ([]string, error)
	CheckDescriptorLensReferences(ctx context.Context, id int64) ([]string, error)
	CheckSummaryLensReferences(ctx context.Context, id int64) ([]string, error)
	CheckCoreExpertiseLensReferences(ctx context.Context, id int64) ([]string, error)

	GetSkillLensTags(ctx context.Context, skillID int64) ([]int64, error)
	SetSkillLensTags(ctx context.Context, skillID int64, lensIDs []int64) error
	ListSkillsWithLensTags(ctx context.Context) ([]domain.SkillWithTags, error)
}

// LensService provides business-logic operations for lenses and
// skill lens tags. It validates inputs before delegating to the
// store.
type LensService struct {
	store LensStore
}

// NewLensService creates a LensService backed by the given store.
func NewLensService(store LensStore) *LensService {
	return &LensService{store: store}
}

// ListLenses returns all lenses.
func (s *LensService) ListLenses(ctx context.Context) ([]domain.Lens, error) {
	return s.store.ListLenses(ctx)
}

// GetLens returns a single lens with all its content selections.
func (s *LensService) GetLens(ctx context.Context, id int64) (domain.LensDetail, error) {
	return s.store.GetLens(ctx, id)
}

// CreateLens validates the input and creates a new lens.
func (s *LensService) CreateLens(ctx context.Context, input domain.LensInput) (domain.Lens, error) {
	if err := domain.ValidateLensInput(input); err != nil {
		return domain.Lens{}, err
	}
	return s.store.CreateLens(ctx, input)
}

// UpdateLens validates the input and updates an existing lens.
func (s *LensService) UpdateLens(ctx context.Context, id int64, input domain.LensInput) (domain.Lens, error) {
	if err := domain.ValidateLensInput(input); err != nil {
		return domain.Lens{}, err
	}
	return s.store.UpdateLens(ctx, id, input)
}

// DeleteLens deletes a lens and all its selections.
func (s *LensService) DeleteLens(ctx context.Context, id int64) error {
	return s.store.DeleteLens(ctx, id)
}

// --- Selection setters — delegate to store ---

// SetLensWorkHistory replaces the work history selections for a lens.
func (s *LensService) SetLensWorkHistory(ctx context.Context, lensID int64, selections []domain.LensWorkHistoryItem) error {
	return s.store.SetLensWorkHistory(ctx, lensID, selections)
}

// SetLensSummaries replaces the summary selections for a lens.
func (s *LensService) SetLensSummaries(ctx context.Context, lensID int64, selections []domain.LensSummaryItem) error {
	return s.store.SetLensSummaries(ctx, lensID, selections)
}

// SetLensBullets replaces the bullet selections for a lens.
func (s *LensService) SetLensBullets(ctx context.Context, lensID int64, selections []domain.LensBulletItem) error {
	return s.store.SetLensBullets(ctx, lensID, selections)
}

// SetLensSkills replaces the skill selections for a lens.
func (s *LensService) SetLensSkills(ctx context.Context, lensID int64, selections []domain.LensSkillItem) error {
	return s.store.SetLensSkills(ctx, lensID, selections)
}

// SetLensAcademics replaces the academic selections for a lens.
func (s *LensService) SetLensAcademics(ctx context.Context, lensID int64, academicIDs []int64) error {
	return s.store.SetLensAcademics(ctx, lensID, academicIDs)
}

// SetLensCerts replaces the certification selections for a lens.
func (s *LensService) SetLensCerts(ctx context.Context, lensID int64, certIDs []int64) error {
	return s.store.SetLensCerts(ctx, lensID, certIDs)
}

// SetLensDescriptors replaces the descriptor selections for a lens.
func (s *LensService) SetLensDescriptors(ctx context.Context, lensID int64, selections []domain.LensDescriptorItem) error {
	return s.store.SetLensDescriptors(ctx, lensID, selections)
}

// SetLensCoreExpertise replaces the core expertise selections for a lens.
func (s *LensService) SetLensCoreExpertise(ctx context.Context, lensID int64, selections []domain.LensCoreExpertiseItem) error {
	return s.store.SetLensCoreExpertise(ctx, lensID, selections)
}

// --- Skill Lens Tags ---

// GetSkillLensTags returns all lens IDs tagged for a skill.
func (s *LensService) GetSkillLensTags(ctx context.Context, skillID int64) ([]int64, error) {
	return s.store.GetSkillLensTags(ctx, skillID)
}

// SetSkillLensTags replaces all lens tags for a skill.
func (s *LensService) SetSkillLensTags(ctx context.Context, skillID int64, lensIDs []int64) error {
	return s.store.SetSkillLensTags(ctx, skillID, lensIDs)
}

// ListSkillsWithLensTags returns all skills with their lens tag
// associations included.
func (s *LensService) ListSkillsWithLensTags(ctx context.Context) ([]domain.SkillWithTags, error) {
	return s.store.ListSkillsWithLensTags(ctx)
}

// --- Lens Reference Checks ---

// CheckWorkHistoryLensReferences returns lens names referencing a
// work history entry (for delete confirmation per FR-050).
func (s *LensService) CheckWorkHistoryLensReferences(ctx context.Context, id int64) ([]string, error) {
	return s.store.CheckWorkHistoryLensReferences(ctx, id)
}

// CheckBulletLensReferences returns lens names referencing a bullet.
func (s *LensService) CheckBulletLensReferences(ctx context.Context, id int64) ([]string, error) {
	return s.store.CheckBulletLensReferences(ctx, id)
}

// CheckAcademicLensReferences returns lens names referencing an
// academic credential.
func (s *LensService) CheckAcademicLensReferences(ctx context.Context, id int64) ([]string, error) {
	return s.store.CheckAcademicLensReferences(ctx, id)
}

// CheckCertLensReferences returns lens names referencing a cert.
func (s *LensService) CheckCertLensReferences(ctx context.Context, id int64) ([]string, error) {
	return s.store.CheckCertLensReferences(ctx, id)
}

// CheckDescriptorLensReferences returns lens names referencing a
// descriptor.
func (s *LensService) CheckDescriptorLensReferences(ctx context.Context, id int64) ([]string, error) {
	return s.store.CheckDescriptorLensReferences(ctx, id)
}

// CheckSummaryLensReferences returns lens names referencing a
// professional summary.
func (s *LensService) CheckSummaryLensReferences(ctx context.Context, id int64) ([]string, error) {
	return s.store.CheckSummaryLensReferences(ctx, id)
}

// CheckCoreExpertiseLensReferences returns lens names referencing a
// core expertise item.
func (s *LensService) CheckCoreExpertiseLensReferences(ctx context.Context, id int64) ([]string, error) {
	return s.store.CheckCoreExpertiseLensReferences(ctx, id)
}

// --- Export Selection Conversion ---

// GetLensExportSelections fetches a lens detail and converts it
// into an ExportRequest that can pre-fill the export dialog. The
// TemplateID is left empty for the user to choose.
func (s *LensService) GetLensExportSelections(ctx context.Context, lensID int64) (domain.ExportRequest, error) {
	detail, err := s.store.GetLens(ctx, lensID)
	if err != nil {
		return domain.ExportRequest{}, err
	}

	lid := detail.ID
	req := domain.ExportRequest{
		LensID: &lid,
	}

	// Summary IDs + master
	req.SummaryIDs = make([]int64, len(detail.Summaries))
	for i, sum := range detail.Summaries {
		req.SummaryIDs[i] = sum.SummaryID
		if sum.IsMaster {
			id := sum.SummaryID
			req.MasterSummaryID = &id
		}
	}

	// Work history IDs
	req.WorkHistoryIDs = make([]int64, len(detail.WorkHistory))
	for i, wh := range detail.WorkHistory {
		req.WorkHistoryIDs[i] = wh.WorkHistoryID
	}

	// Bullet IDs
	req.BulletIDs = make([]int64, len(detail.Bullets))
	for i, b := range detail.Bullets {
		req.BulletIDs[i] = b.BulletID
	}

	// Skill IDs + sort overrides
	req.SkillIDs = make([]int64, len(detail.Skills))
	req.SkillSortOverrides = make(map[int64]int)
	for i, sk := range detail.Skills {
		req.SkillIDs[i] = sk.SkillID
		if sk.CustomSortOrder != nil {
			req.SkillSortOverrides[sk.SkillID] = *sk.CustomSortOrder
		}
	}

	// Academic IDs (copy to avoid aliasing)
	req.AcademicIDs = make([]int64, len(detail.AcademicIDs))
	copy(req.AcademicIDs, detail.AcademicIDs)

	// Certification IDs
	req.CertificationIDs = make([]int64, len(detail.CertIDs))
	copy(req.CertificationIDs, detail.CertIDs)

	// Descriptor IDs
	req.DescriptorIDs = make([]int64, len(detail.Descriptors))
	for i, d := range detail.Descriptors {
		req.DescriptorIDs[i] = d.DescriptorID
	}

	// Core Expertise IDs
	req.CoreExpertiseIDs = make([]int64, len(detail.CoreExpertise))
	for i, ce := range detail.CoreExpertise {
		req.CoreExpertiseIDs[i] = ce.CoreExpertiseID
	}

	return req, nil
}
