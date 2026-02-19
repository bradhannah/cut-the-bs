package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =================================================================
// Skill Categories
// =================================================================

func TestCreateSkillCategory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)
	assert.NotZero(t, cat.ID)
	assert.Equal(t, "Languages", cat.Name)
	assert.Equal(t, 1, cat.SortOrder)
	assert.NotEmpty(t, cat.CreatedAt)
}

func TestCreateSkillCategory_AutoIncrementsSortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	first, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)
	assert.Equal(t, 1, first.SortOrder)

	second, err := store.CreateSkillCategory(ctx, "Frameworks")
	require.NoError(t, err)
	assert.Equal(t, 2, second.SortOrder)
}

func TestCreateSkillCategory_DuplicateNameFails(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	_, err = store.CreateSkillCategory(ctx, "Languages")
	require.Error(t, err)
}

func TestListSkillCategories_OrderedBySortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)
	_, err = store.CreateSkillCategory(ctx, "Frameworks")
	require.NoError(t, err)
	_, err = store.CreateSkillCategory(ctx, "Cloud")
	require.NoError(t, err)

	cats, err := store.ListSkillCategories(ctx)
	require.NoError(t, err)
	require.Len(t, cats, 3)
	assert.Equal(t, "Languages", cats[0].Name)
	assert.Equal(t, "Frameworks", cats[1].Name)
	assert.Equal(t, "Cloud", cats[2].Name)
}

func TestListSkillCategories_EmptyReturnsEmptySlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cats, err := store.ListSkillCategories(ctx)
	require.NoError(t, err)
	assert.NotNil(t, cats)
	assert.Len(t, cats, 0)
}

func TestRenameSkillCategory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	renamed, err := store.RenameSkillCategory(ctx, cat.ID, "Programming Languages")
	require.NoError(t, err)
	assert.Equal(t, cat.ID, renamed.ID)
	assert.Equal(t, "Programming Languages", renamed.Name)
	assert.Equal(t, cat.SortOrder, renamed.SortOrder)
}

func TestRenameSkillCategory_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.RenameSkillCategory(ctx, 999, "New Name")
	require.Error(t, err)
}

func TestDeleteSkillCategory_EmptyCategory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	err = store.DeleteSkillCategory(ctx, cat.ID)
	require.NoError(t, err)

	cats, err := store.ListSkillCategories(ctx)
	require.NoError(t, err)
	assert.Len(t, cats, 0)
}

func TestDeleteSkillCategory_WithSkillsFails(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	_, err = store.CreateSkill(ctx, domain.SkillInput{
		Name:            "Go",
		CategoryID:      cat.ID,
		CompetenceLevel: 7,
	})
	require.NoError(t, err)

	err = store.DeleteSkillCategory(ctx, cat.ID)
	require.Error(t, err, "should fail when skills reference the category")
}

func TestDeleteSkillCategory_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteSkillCategory(ctx, 999)
	require.Error(t, err)
}

func TestReorderSkillCategories(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	a, _ := store.CreateSkillCategory(ctx, "A")
	b, _ := store.CreateSkillCategory(ctx, "B")
	c, _ := store.CreateSkillCategory(ctx, "C")

	// Reverse: C, B, A
	err := store.ReorderSkillCategories(ctx, []int64{c.ID, b.ID, a.ID})
	require.NoError(t, err)

	cats, err := store.ListSkillCategories(ctx)
	require.NoError(t, err)
	require.Len(t, cats, 3)
	assert.Equal(t, "C", cats[0].Name)
	assert.Equal(t, "B", cats[1].Name)
	assert.Equal(t, "A", cats[2].Name)
	assert.Equal(t, 1, cats[0].SortOrder)
	assert.Equal(t, 2, cats[1].SortOrder)
	assert.Equal(t, 3, cats[2].SortOrder)
}

// =================================================================
// Skills
// =================================================================

func TestCreateSkill(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	skill, err := store.CreateSkill(ctx, domain.SkillInput{
		Name:            "Go",
		CategoryID:      cat.ID,
		CompetenceLevel: 8,
		IsLegacy:        false,
	})
	require.NoError(t, err)
	assert.NotZero(t, skill.ID)
	assert.Equal(t, "Go", skill.Name)
	assert.Equal(t, cat.ID, skill.CategoryID)
	assert.Equal(t, 8, skill.CompetenceLevel)
	assert.False(t, skill.IsLegacy)
	assert.NotEmpty(t, skill.CreatedAt)
}

func TestCreateSkill_InvalidCategoryFails(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateSkill(ctx, domain.SkillInput{
		Name:            "Go",
		CategoryID:      999,
		CompetenceLevel: 5,
	})
	require.Error(t, err, "FK constraint should reject invalid category_id")
}

func TestCreateSkill_LegacyFlag(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Legacy")

	skill, err := store.CreateSkill(ctx, domain.SkillInput{
		Name:            "jQuery",
		CategoryID:      cat.ID,
		CompetenceLevel: 6,
		IsLegacy:        true,
	})
	require.NoError(t, err)
	assert.True(t, skill.IsLegacy)
}

func TestUpdateSkill(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Languages")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name:            "Go",
		CategoryID:      cat.ID,
		CompetenceLevel: 5,
	})

	updated, err := store.UpdateSkill(ctx, skill.ID, domain.SkillInput{
		Name:            "Golang",
		CategoryID:      cat.ID,
		CompetenceLevel: 8,
		IsLegacy:        true,
	})
	require.NoError(t, err)
	assert.Equal(t, skill.ID, updated.ID)
	assert.Equal(t, "Golang", updated.Name)
	assert.Equal(t, 8, updated.CompetenceLevel)
	assert.True(t, updated.IsLegacy)
}

func TestUpdateSkill_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateSkill(ctx, 999, domain.SkillInput{
		Name:            "Go",
		CategoryID:      1,
		CompetenceLevel: 5,
	})
	require.Error(t, err)
}

func TestDeleteSkill(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Languages")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name:            "Go",
		CategoryID:      cat.ID,
		CompetenceLevel: 5,
	})

	err := store.DeleteSkill(ctx, skill.ID)
	require.NoError(t, err)

	skills, err := store.ListSkills(ctx)
	require.NoError(t, err)
	assert.Len(t, skills, 0)
}

func TestDeleteSkill_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteSkill(ctx, 999)
	require.Error(t, err)
}

func TestListSkills_SortedByCompetenceDescNameAsc(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Languages")

	// Create skills with varying competence levels and names.
	_, _ = store.CreateSkill(ctx, domain.SkillInput{Name: "Python", CategoryID: cat.ID, CompetenceLevel: 5})
	_, _ = store.CreateSkill(ctx, domain.SkillInput{Name: "Go", CategoryID: cat.ID, CompetenceLevel: 8})
	_, _ = store.CreateSkill(ctx, domain.SkillInput{Name: "Rust", CategoryID: cat.ID, CompetenceLevel: 8})
	_, _ = store.CreateSkill(ctx, domain.SkillInput{Name: "Java", CategoryID: cat.ID, CompetenceLevel: 5})

	skills, err := store.ListSkills(ctx)
	require.NoError(t, err)
	require.Len(t, skills, 4)
	// Level 8: Go before Rust (alpha). Level 5: Java before Python (alpha).
	assert.Equal(t, "Go", skills[0].Name)
	assert.Equal(t, "Rust", skills[1].Name)
	assert.Equal(t, "Java", skills[2].Name)
	assert.Equal(t, "Python", skills[3].Name)
}

func TestListSkills_EmptyReturnsEmptySlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	skills, err := store.ListSkills(ctx)
	require.NoError(t, err)
	assert.NotNil(t, skills)
	assert.Len(t, skills, 0)
}

func TestListSkillsByCategory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lang, _ := store.CreateSkillCategory(ctx, "Languages")
	cloud, _ := store.CreateSkillCategory(ctx, "Cloud")

	_, _ = store.CreateSkill(ctx, domain.SkillInput{Name: "Go", CategoryID: lang.ID, CompetenceLevel: 8})
	_, _ = store.CreateSkill(ctx, domain.SkillInput{Name: "Python", CategoryID: lang.ID, CompetenceLevel: 6})
	_, _ = store.CreateSkill(ctx, domain.SkillInput{Name: "AWS", CategoryID: cloud.ID, CompetenceLevel: 7})

	result, err := store.ListSkillsByCategory(ctx)
	require.NoError(t, err)
	require.Len(t, result, 2)

	// Categories ordered by sort_order.
	assert.Equal(t, "Languages", result[0].Category.Name)
	assert.Equal(t, "Cloud", result[1].Category.Name)

	// Skills within category sorted by competence desc, name asc.
	require.Len(t, result[0].Skills, 2)
	assert.Equal(t, "Go", result[0].Skills[0].Name)
	assert.Equal(t, "Python", result[0].Skills[1].Name)

	require.Len(t, result[1].Skills, 1)
	assert.Equal(t, "AWS", result[1].Skills[0].Name)
}

func TestListSkillsByCategory_EmptyCategory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateSkillCategory(ctx, "Empty")

	result, err := store.ListSkillsByCategory(ctx)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Empty", result[0].Category.Name)
	assert.NotNil(t, result[0].Skills)
	assert.Len(t, result[0].Skills, 0)
}

func TestListSkillsByCategory_NoCategories(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	result, err := store.ListSkillsByCategory(ctx)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}
