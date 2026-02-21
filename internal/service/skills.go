package service

import (
	"context"
	"strings"

	"cut-the-bs/internal/domain"
)

// SkillsStore defines the persistence operations required by
// SkillsService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type SkillsStore interface {
	ListSkills(ctx context.Context) ([]domain.Skill, error)
	ListSkillsByCategory(ctx context.Context) ([]domain.SkillCategoryWithSkills, error)
	CreateSkill(ctx context.Context, input domain.SkillInput) (domain.Skill, error)
	UpdateSkill(ctx context.Context, id int64, input domain.SkillInput) (domain.Skill, error)
	DeleteSkill(ctx context.Context, id int64) error
	CheckSkillLensReferences(ctx context.Context, id int64) ([]string, error)
	ListSkillCategories(ctx context.Context) ([]domain.SkillCategory, error)
	CreateSkillCategory(ctx context.Context, name string) (domain.SkillCategory, error)
	RenameSkillCategory(ctx context.Context, id int64, name string) (domain.SkillCategory, error)
	DeleteSkillCategory(ctx context.Context, id int64) error
	ReorderSkillCategories(ctx context.Context, orderedIDs []int64) error
}

// SkillsService provides business-logic operations for skills and
// skill categories. It validates inputs before delegating to the
// store.
type SkillsService struct {
	store SkillsStore
}

// NewSkillsService creates a SkillsService backed by the given
// store.
func NewSkillsService(store SkillsStore) *SkillsService {
	return &SkillsService{store: store}
}

// --- Skills ---

// ListSkills returns all skills sorted by competence level
// (descending), then alphabetically.
func (s *SkillsService) ListSkills(ctx context.Context) ([]domain.Skill, error) {
	return s.store.ListSkills(ctx)
}

// ListSkillsByCategory returns skills grouped by category, with
// categories ordered by their sort_order.
func (s *SkillsService) ListSkillsByCategory(ctx context.Context) ([]domain.SkillCategoryWithSkills, error) {
	return s.store.ListSkillsByCategory(ctx)
}

// CreateSkill validates the input and creates a new skill.
func (s *SkillsService) CreateSkill(ctx context.Context, input domain.SkillInput) (domain.Skill, error) {
	if err := domain.ValidateSkillInput(input); err != nil {
		return domain.Skill{}, err
	}
	return s.store.CreateSkill(ctx, input)
}

// UpdateSkill validates the input and updates an existing skill.
func (s *SkillsService) UpdateSkill(ctx context.Context, id int64, input domain.SkillInput) (domain.Skill, error) {
	if err := domain.ValidateSkillInput(input); err != nil {
		return domain.Skill{}, err
	}
	return s.store.UpdateSkill(ctx, id, input)
}

// DeleteSkill deletes a skill by ID.
func (s *SkillsService) DeleteSkill(ctx context.Context, id int64) error {
	return s.store.DeleteSkill(ctx, id)
}

// CheckSkillLensReferences returns the names of lenses that
// reference a given skill, for use in delete confirmation dialogs.
func (s *SkillsService) CheckSkillLensReferences(ctx context.Context, id int64) ([]string, error) {
	return s.store.CheckSkillLensReferences(ctx, id)
}

// SplitSkillsText accepts a comma-separated string and returns
// individual skill names, trimming whitespace and filtering out
// blank entries. This is a preview operation — no persistence
// occurs.
func (s *SkillsService) SplitSkillsText(text string) []string {
	parts := strings.Split(text, ",")
	result := make([]string, 0)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetCompetenceLevels returns the fixed competence scale with
// descriptive criteria for each level.
func (s *SkillsService) GetCompetenceLevels() []domain.CompetenceLevel {
	return domain.CompetenceLevels
}

// --- Skill Categories ---

// ListSkillCategories returns all skill categories ordered by
// sort_order.
func (s *SkillsService) ListSkillCategories(ctx context.Context) ([]domain.SkillCategory, error) {
	return s.store.ListSkillCategories(ctx)
}

// CreateSkillCategory validates the name and creates a new skill
// category.
func (s *SkillsService) CreateSkillCategory(ctx context.Context, name string) (domain.SkillCategory, error) {
	if err := domain.ValidateRequired(name, "category name"); err != nil {
		return domain.SkillCategory{}, err
	}
	return s.store.CreateSkillCategory(ctx, name)
}

// RenameSkillCategory validates the name and updates a category's
// name.
func (s *SkillsService) RenameSkillCategory(ctx context.Context, id int64, name string) (domain.SkillCategory, error) {
	if err := domain.ValidateRequired(name, "category name"); err != nil {
		return domain.SkillCategory{}, err
	}
	return s.store.RenameSkillCategory(ctx, id, name)
}

// DeleteSkillCategory deletes a skill category by ID. Fails if any
// skills still reference it.
func (s *SkillsService) DeleteSkillCategory(ctx context.Context, id int64) error {
	return s.store.DeleteSkillCategory(ctx, id)
}

// ReorderSkillCategories updates the sort_order of all categories
// based on the provided ordered slice of IDs.
func (s *SkillsService) ReorderSkillCategories(ctx context.Context, orderedIDs []int64) error {
	return s.store.ReorderSkillCategories(ctx, orderedIDs)
}
