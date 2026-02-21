package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSkillsStore implements SkillsStore for testing.
type mockSkillsStore struct {
	skills     []domain.Skill
	skill      domain.Skill
	categories []domain.SkillCategory
	category   domain.SkillCategory
	grouped    []domain.SkillCategoryWithSkills
	lensNames  []string
	err        error

	// call tracking
	createSkillCalls       []domain.SkillInput
	updateSkillCalls       []updateSkillCall
	deleteSkillCalls       []int64
	checkLensRefCalls      []int64
	createCategoryCalls    []string
	renameCategoryCalls    []renameCategoryCall
	deleteCategoryCalls    []int64
	reorderCategoriesCalls [][]int64
}

type updateSkillCall struct {
	ID    int64
	Input domain.SkillInput
}

type renameCategoryCall struct {
	ID   int64
	Name string
}

func (m *mockSkillsStore) ListSkills(_ context.Context) ([]domain.Skill, error) {
	return m.skills, m.err
}

func (m *mockSkillsStore) ListSkillsByCategory(_ context.Context) ([]domain.SkillCategoryWithSkills, error) {
	return m.grouped, m.err
}

func (m *mockSkillsStore) CreateSkill(_ context.Context, input domain.SkillInput) (domain.Skill, error) {
	m.createSkillCalls = append(m.createSkillCalls, input)
	if m.err != nil {
		return domain.Skill{}, m.err
	}
	return m.skill, nil
}

func (m *mockSkillsStore) UpdateSkill(_ context.Context, id int64, input domain.SkillInput) (domain.Skill, error) {
	m.updateSkillCalls = append(m.updateSkillCalls, updateSkillCall{ID: id, Input: input})
	if m.err != nil {
		return domain.Skill{}, m.err
	}
	return m.skill, nil
}

func (m *mockSkillsStore) DeleteSkill(_ context.Context, id int64) error {
	m.deleteSkillCalls = append(m.deleteSkillCalls, id)
	return m.err
}

func (m *mockSkillsStore) CheckSkillLensReferences(_ context.Context, id int64) ([]string, error) {
	m.checkLensRefCalls = append(m.checkLensRefCalls, id)
	if m.err != nil {
		return nil, m.err
	}
	return m.lensNames, nil
}

func (m *mockSkillsStore) ListSkillCategories(_ context.Context) ([]domain.SkillCategory, error) {
	return m.categories, m.err
}

func (m *mockSkillsStore) CreateSkillCategory(_ context.Context, name string) (domain.SkillCategory, error) {
	m.createCategoryCalls = append(m.createCategoryCalls, name)
	if m.err != nil {
		return domain.SkillCategory{}, m.err
	}
	return m.category, nil
}

func (m *mockSkillsStore) RenameSkillCategory(_ context.Context, id int64, name string) (domain.SkillCategory, error) {
	m.renameCategoryCalls = append(m.renameCategoryCalls, renameCategoryCall{ID: id, Name: name})
	if m.err != nil {
		return domain.SkillCategory{}, m.err
	}
	return m.category, nil
}

func (m *mockSkillsStore) DeleteSkillCategory(_ context.Context, id int64) error {
	m.deleteCategoryCalls = append(m.deleteCategoryCalls, id)
	return m.err
}

func (m *mockSkillsStore) ReorderSkillCategories(_ context.Context, orderedIDs []int64) error {
	m.reorderCategoriesCalls = append(m.reorderCategoriesCalls, orderedIDs)
	return m.err
}

// =================================================================
// ListSkills
// =================================================================

func TestSkillsService_ListSkills_Success(t *testing.T) {
	store := &mockSkillsStore{
		skills: []domain.Skill{
			{ID: 1, Name: "Go", CompetenceLevel: 8},
			{ID: 2, Name: "Python", CompetenceLevel: 6},
		},
	}
	svc := NewSkillsService(store)

	skills, err := svc.ListSkills(context.Background())
	require.NoError(t, err)
	require.Len(t, skills, 2)
	assert.Equal(t, "Go", skills[0].Name)
}

func TestSkillsService_ListSkills_Empty(t *testing.T) {
	store := &mockSkillsStore{skills: []domain.Skill{}}
	svc := NewSkillsService(store)

	skills, err := svc.ListSkills(context.Background())
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestSkillsService_ListSkills_StoreError(t *testing.T) {
	store := &mockSkillsStore{err: fmt.Errorf("db error")}
	svc := NewSkillsService(store)

	_, err := svc.ListSkills(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// =================================================================
// ListSkillsByCategory
// =================================================================

func TestSkillsService_ListSkillsByCategory_Success(t *testing.T) {
	store := &mockSkillsStore{
		grouped: []domain.SkillCategoryWithSkills{
			{
				Category: domain.SkillCategory{ID: 1, Name: "Languages"},
				Skills:   []domain.Skill{{ID: 1, Name: "Go"}},
			},
		},
	}
	svc := NewSkillsService(store)

	result, err := svc.ListSkillsByCategory(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Languages", result[0].Category.Name)
}

// =================================================================
// CreateSkill — validation
// =================================================================

func TestSkillsService_CreateSkill_EmptyName(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	_, err := svc.CreateSkill(context.Background(), domain.SkillInput{
		Name:            "",
		CategoryID:      1,
		CompetenceLevel: 5,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skill name")
	assert.Empty(t, store.createSkillCalls)
}

func TestSkillsService_CreateSkill_ZeroCategoryID(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	_, err := svc.CreateSkill(context.Background(), domain.SkillInput{
		Name:            "Go",
		CategoryID:      0,
		CompetenceLevel: 5,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "category")
	assert.Empty(t, store.createSkillCalls)
}

func TestSkillsService_CreateSkill_InvalidCompetenceLevel(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	_, err := svc.CreateSkill(context.Background(), domain.SkillInput{
		Name:            "Go",
		CategoryID:      1,
		CompetenceLevel: 11,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "competence level")
	assert.Empty(t, store.createSkillCalls)
}

func TestSkillsService_CreateSkill_CompetenceLevelZero(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	_, err := svc.CreateSkill(context.Background(), domain.SkillInput{
		Name:            "Go",
		CategoryID:      1,
		CompetenceLevel: 0,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "competence level")
	assert.Empty(t, store.createSkillCalls)
}

// =================================================================
// CreateSkill — happy path
// =================================================================

func TestSkillsService_CreateSkill_Success(t *testing.T) {
	store := &mockSkillsStore{
		skill: domain.Skill{
			ID: 1, Name: "Go", CategoryID: 1, CompetenceLevel: 8,
		},
	}
	svc := NewSkillsService(store)

	skill, err := svc.CreateSkill(context.Background(), domain.SkillInput{
		Name:            "Go",
		CategoryID:      1,
		CompetenceLevel: 8,
	})
	require.NoError(t, err)
	assert.Equal(t, "Go", skill.Name)
	assert.Equal(t, 8, skill.CompetenceLevel)
	require.Len(t, store.createSkillCalls, 1)
}

func TestSkillsService_CreateSkill_WithLegacy(t *testing.T) {
	store := &mockSkillsStore{
		skill: domain.Skill{
			ID: 1, Name: "jQuery", CategoryID: 1, CompetenceLevel: 4, IsLegacy: true,
		},
	}
	svc := NewSkillsService(store)

	skill, err := svc.CreateSkill(context.Background(), domain.SkillInput{
		Name:            "jQuery",
		CategoryID:      1,
		CompetenceLevel: 4,
		IsLegacy:        true,
	})
	require.NoError(t, err)
	assert.True(t, skill.IsLegacy)
}

func TestSkillsService_CreateSkill_StoreError(t *testing.T) {
	store := &mockSkillsStore{err: fmt.Errorf("duplicate")}
	svc := NewSkillsService(store)

	_, err := svc.CreateSkill(context.Background(), domain.SkillInput{
		Name:            "Go",
		CategoryID:      1,
		CompetenceLevel: 8,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

// =================================================================
// UpdateSkill — validation
// =================================================================

func TestSkillsService_UpdateSkill_EmptyName(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	_, err := svc.UpdateSkill(context.Background(), 1, domain.SkillInput{
		Name:            "",
		CategoryID:      1,
		CompetenceLevel: 5,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skill name")
	assert.Empty(t, store.updateSkillCalls)
}

// =================================================================
// UpdateSkill — happy path
// =================================================================

func TestSkillsService_UpdateSkill_Success(t *testing.T) {
	store := &mockSkillsStore{
		skill: domain.Skill{
			ID: 1, Name: "Golang", CategoryID: 1, CompetenceLevel: 9,
		},
	}
	svc := NewSkillsService(store)

	skill, err := svc.UpdateSkill(context.Background(), 1, domain.SkillInput{
		Name:            "Golang",
		CategoryID:      1,
		CompetenceLevel: 9,
	})
	require.NoError(t, err)
	assert.Equal(t, "Golang", skill.Name)
	require.Len(t, store.updateSkillCalls, 1)
	assert.Equal(t, int64(1), store.updateSkillCalls[0].ID)
}

// =================================================================
// DeleteSkill
// =================================================================

func TestSkillsService_DeleteSkill_Success(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	err := svc.DeleteSkill(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, store.deleteSkillCalls, 1)
	assert.Equal(t, int64(5), store.deleteSkillCalls[0])
}

func TestSkillsService_DeleteSkill_StoreError(t *testing.T) {
	store := &mockSkillsStore{err: fmt.Errorf("not found")}
	svc := NewSkillsService(store)

	err := svc.DeleteSkill(context.Background(), 999)
	require.Error(t, err)
}

// =================================================================
// CheckSkillLensReferences
// =================================================================

func TestSkillsService_CheckSkillLensReferences_Success(t *testing.T) {
	store := &mockSkillsStore{
		lensNames: []string{"Frontend", "Backend"},
	}
	svc := NewSkillsService(store)

	names, err := svc.CheckSkillLensReferences(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, names, 2)
	assert.Equal(t, "Frontend", names[0])
}

func TestSkillsService_CheckSkillLensReferences_None(t *testing.T) {
	store := &mockSkillsStore{lensNames: []string{}}
	svc := NewSkillsService(store)

	names, err := svc.CheckSkillLensReferences(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, names)
}

// =================================================================
// SplitSkillsText
// =================================================================

func TestSkillsService_SplitSkillsText_CommaSeparated(t *testing.T) {
	svc := NewSkillsService(nil)

	result := svc.SplitSkillsText("Go, Python, TypeScript")
	require.Len(t, result, 3)
	assert.Equal(t, "Go", result[0])
	assert.Equal(t, "Python", result[1])
	assert.Equal(t, "TypeScript", result[2])
}

func TestSkillsService_SplitSkillsText_TrimsWhitespace(t *testing.T) {
	svc := NewSkillsService(nil)

	result := svc.SplitSkillsText("  Go ,  Python  ,TypeScript  ")
	require.Len(t, result, 3)
	assert.Equal(t, "Go", result[0])
	assert.Equal(t, "Python", result[1])
	assert.Equal(t, "TypeScript", result[2])
}

func TestSkillsService_SplitSkillsText_FiltersEmpty(t *testing.T) {
	svc := NewSkillsService(nil)

	result := svc.SplitSkillsText("Go,, ,Python")
	require.Len(t, result, 2)
	assert.Equal(t, "Go", result[0])
	assert.Equal(t, "Python", result[1])
}

func TestSkillsService_SplitSkillsText_EmptyInput(t *testing.T) {
	svc := NewSkillsService(nil)

	result := svc.SplitSkillsText("")
	assert.Empty(t, result)
}

func TestSkillsService_SplitSkillsText_SingleSkill(t *testing.T) {
	svc := NewSkillsService(nil)

	result := svc.SplitSkillsText("Go")
	require.Len(t, result, 1)
	assert.Equal(t, "Go", result[0])
}

// =================================================================
// GetCompetenceLevels
// =================================================================

func TestSkillsService_GetCompetenceLevels(t *testing.T) {
	svc := NewSkillsService(nil)

	levels := svc.GetCompetenceLevels()
	require.Len(t, levels, 10)
	assert.Equal(t, 10, levels[0].Level)
	assert.Equal(t, "Visionary", levels[0].Label)
	assert.Equal(t, 1, levels[9].Level)
	assert.Equal(t, "Awareness", levels[9].Label)
}

// =================================================================
// Skill Category operations
// =================================================================

func TestSkillsService_ListSkillCategories_Success(t *testing.T) {
	store := &mockSkillsStore{
		categories: []domain.SkillCategory{
			{ID: 1, Name: "Languages", SortOrder: 1},
			{ID: 2, Name: "Frameworks", SortOrder: 2},
		},
	}
	svc := NewSkillsService(store)

	cats, err := svc.ListSkillCategories(context.Background())
	require.NoError(t, err)
	require.Len(t, cats, 2)
	assert.Equal(t, "Languages", cats[0].Name)
}

func TestSkillsService_CreateSkillCategory_EmptyName(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	_, err := svc.CreateSkillCategory(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "category name")
	assert.Empty(t, store.createCategoryCalls)
}

func TestSkillsService_CreateSkillCategory_WhitespaceName(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	_, err := svc.CreateSkillCategory(context.Background(), "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "category name")
	assert.Empty(t, store.createCategoryCalls)
}

func TestSkillsService_CreateSkillCategory_Success(t *testing.T) {
	store := &mockSkillsStore{
		category: domain.SkillCategory{ID: 1, Name: "Languages", SortOrder: 1},
	}
	svc := NewSkillsService(store)

	cat, err := svc.CreateSkillCategory(context.Background(), "Languages")
	require.NoError(t, err)
	assert.Equal(t, "Languages", cat.Name)
	require.Len(t, store.createCategoryCalls, 1)
}

func TestSkillsService_CreateSkillCategory_StoreError(t *testing.T) {
	store := &mockSkillsStore{err: fmt.Errorf("duplicate")}
	svc := NewSkillsService(store)

	_, err := svc.CreateSkillCategory(context.Background(), "Languages")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestSkillsService_RenameSkillCategory_EmptyName(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	_, err := svc.RenameSkillCategory(context.Background(), 1, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "category name")
	assert.Empty(t, store.renameCategoryCalls)
}

func TestSkillsService_RenameSkillCategory_Success(t *testing.T) {
	store := &mockSkillsStore{
		category: domain.SkillCategory{ID: 1, Name: "Programming Languages", SortOrder: 1},
	}
	svc := NewSkillsService(store)

	cat, err := svc.RenameSkillCategory(context.Background(), 1, "Programming Languages")
	require.NoError(t, err)
	assert.Equal(t, "Programming Languages", cat.Name)
	require.Len(t, store.renameCategoryCalls, 1)
	assert.Equal(t, int64(1), store.renameCategoryCalls[0].ID)
}

func TestSkillsService_DeleteSkillCategory_Success(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	err := svc.DeleteSkillCategory(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, store.deleteCategoryCalls, 1)
	assert.Equal(t, int64(1), store.deleteCategoryCalls[0])
}

func TestSkillsService_DeleteSkillCategory_StoreError(t *testing.T) {
	store := &mockSkillsStore{err: fmt.Errorf("has skills")}
	svc := NewSkillsService(store)

	err := svc.DeleteSkillCategory(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "has skills")
}

func TestSkillsService_ReorderSkillCategories_Success(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	err := svc.ReorderSkillCategories(context.Background(), []int64{3, 1, 2})
	require.NoError(t, err)
	require.Len(t, store.reorderCategoriesCalls, 1)
	assert.Equal(t, []int64{3, 1, 2}, store.reorderCategoriesCalls[0])
}

func TestSkillsService_ReorderSkillCategories_Empty(t *testing.T) {
	store := &mockSkillsStore{}
	svc := NewSkillsService(store)

	err := svc.ReorderSkillCategories(context.Background(), []int64{})
	require.NoError(t, err)
}
