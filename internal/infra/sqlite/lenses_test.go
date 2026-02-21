package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =================================================================
// Lens CRUD
// =================================================================

func TestCreateLens(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, err := store.CreateLens(ctx, domain.LensInput{
		Name: "Backend Engineer",
	})
	require.NoError(t, err)
	assert.NotZero(t, lens.ID)
	assert.Equal(t, "Backend Engineer", lens.Name)
	assert.NotEmpty(t, lens.CreatedAt)
}

func TestCreateLens_WithSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label: "Backend", BodyText: "Experienced backend dev...",
	})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{
		Name: "Backend Engineer",
	})
	require.NoError(t, err)

	// Associate summary via junction table
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: summary.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 1)
	assert.Equal(t, summary.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, 0, detail.Summaries[0].SortOrder)
}

func TestCreateLens_DuplicateNameFails(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	require.NoError(t, err)

	_, err = store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	require.Error(t, err, "duplicate name should fail UNIQUE constraint")
}

func TestListLenses(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	_, _ = store.CreateLens(ctx, domain.LensInput{Name: "Frontend"})

	lenses, err := store.ListLenses(ctx)
	require.NoError(t, err)
	require.Len(t, lenses, 2)
	assert.Equal(t, "Backend", lenses[0].Name)
	assert.Equal(t, "Frontend", lenses[1].Name)
}

func TestListLenses_EmptyReturnsEmptySlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lenses, err := store.ListLenses(ctx)
	require.NoError(t, err)
	assert.NotNil(t, lenses)
	assert.Len(t, lenses, 0)
}

func TestGetLens_BasicFields(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, err := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, detail.ID)
	assert.Equal(t, "Backend", detail.Name)
	// Empty selections should be non-nil empty slices
	assert.NotNil(t, detail.Summaries)
	assert.Len(t, detail.Summaries, 0)
	assert.NotNil(t, detail.WorkHistory)
	assert.Len(t, detail.WorkHistory, 0)
	assert.NotNil(t, detail.Bullets)
	assert.Len(t, detail.Bullets, 0)
	assert.NotNil(t, detail.Skills)
	assert.Len(t, detail.Skills, 0)
	assert.NotNil(t, detail.AcademicIDs)
	assert.Len(t, detail.AcademicIDs, 0)
	assert.NotNil(t, detail.CertIDs)
	assert.Len(t, detail.CertIDs, 0)
	assert.NotNil(t, detail.Descriptors)
	assert.Len(t, detail.Descriptors, 0)
}

func TestGetLens_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.GetLens(ctx, 999)
	require.Error(t, err)
}

func TestUpdateLens(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})

	updated, err := store.UpdateLens(ctx, created.ID, domain.LensInput{
		Name: "Backend Engineer",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Backend Engineer", updated.Name)
}

func TestSetLensSummaries_ClearAll(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	summary, _ := store.CreateSummary(ctx, domain.SummaryInput{
		Label: "Summary", BodyText: "Text...",
	})
	lens, _ := store.CreateLens(ctx, domain.LensInput{
		Name: "Backend",
	})

	// Set a summary
	err := store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: summary.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Clear all summaries
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	assert.Len(t, detail.Summaries, 0)
}

func TestUpdateLens_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateLens(ctx, 999, domain.LensInput{Name: "X"})
	require.Error(t, err)
}

func TestUpdateLens_DuplicateNameFails(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	b, _ := store.CreateLens(ctx, domain.LensInput{Name: "Frontend"})

	_, err := store.UpdateLens(ctx, b.ID, domain.LensInput{Name: "Backend"})
	require.Error(t, err, "updating to duplicate name should fail")
}

func TestDeleteLens(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})

	err := store.DeleteLens(ctx, created.ID)
	require.NoError(t, err)

	lenses, err := store.ListLenses(ctx)
	require.NoError(t, err)
	assert.Len(t, lenses, 0)
}

func TestDeleteLens_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteLens(ctx, 999)
	require.Error(t, err)
}

func TestDeleteSummary_CascadesFromLens(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Deleting a summary should CASCADE-remove it from lens_summary_selection
	summary, _ := store.CreateSummary(ctx, domain.SummaryInput{
		Label: "Summary", BodyText: "Text",
	})
	lens, _ := store.CreateLens(ctx, domain.LensInput{
		Name: "Backend",
	})

	err := store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: summary.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	err = store.DeleteSummary(ctx, summary.ID)
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	assert.Len(t, detail.Summaries, 0, "summary selection should be removed after summary deleted")
}

// =================================================================
// Selection Setters — Work History
// =================================================================

func TestSetLensWorkHistory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	wh1, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})
	wh2, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Beta", JobTitle: "Lead",
		StartDate: "2022-01", DateGranularityStart: "month",
	})

	err := store.SetLensWorkHistory(ctx, lens.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh1.ID, SortOrder: 0},
		{WorkHistoryID: wh2.ID, SortOrder: 1},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.WorkHistory, 2)
	assert.Equal(t, wh1.ID, detail.WorkHistory[0].WorkHistoryID)
	assert.Equal(t, 0, detail.WorkHistory[0].SortOrder)
	assert.Equal(t, wh2.ID, detail.WorkHistory[1].WorkHistoryID)
	assert.Equal(t, 1, detail.WorkHistory[1].SortOrder)
}

func TestSetLensWorkHistory_Replace(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	wh1, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})
	wh2, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Beta", JobTitle: "Lead",
		StartDate: "2022-01", DateGranularityStart: "month",
	})

	// Set initial
	_ = store.SetLensWorkHistory(ctx, lens.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh1.ID, SortOrder: 0},
	})

	// Replace with different set
	err := store.SetLensWorkHistory(ctx, lens.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh2.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	detail, _ := store.GetLens(ctx, lens.ID)
	require.Len(t, detail.WorkHistory, 1)
	assert.Equal(t, wh2.ID, detail.WorkHistory[0].WorkHistoryID)
}

func TestSetLensWorkHistory_ClearAll(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	wh1, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})

	_ = store.SetLensWorkHistory(ctx, lens.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh1.ID, SortOrder: 0},
	})

	// Clear
	err := store.SetLensWorkHistory(ctx, lens.ID, []domain.LensWorkHistoryItem{})
	require.NoError(t, err)

	detail, _ := store.GetLens(ctx, lens.ID)
	assert.Len(t, detail.WorkHistory, 0)
}

// =================================================================
// Selection Setters — Bullets
// =================================================================

func TestSetLensBullets(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	wh, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})
	b1, _ := store.CreateBullet(ctx, wh.ID, "Built microservices", domain.BulletTypePrimary)
	b2, _ := store.CreateBullet(ctx, wh.ID, "Led team of 5", domain.BulletTypePrimary)

	err := store.SetLensBullets(ctx, lens.ID, []domain.LensBulletItem{
		{BulletID: b1.ID, SortOrder: 0},
		{BulletID: b2.ID, SortOrder: 1},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Bullets, 2)
	assert.Equal(t, b1.ID, detail.Bullets[0].BulletID)
	assert.Equal(t, b2.ID, detail.Bullets[1].BulletID)
}

// =================================================================
// Selection Setters — Skills
// =================================================================

func TestSetLensSkills(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	cat, _ := store.CreateSkillCategory(ctx, "Languages")
	sk1, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 8,
	})
	sk2, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Python", CategoryID: cat.ID, CompetenceLevel: 7,
	})

	customSort := 5
	err := store.SetLensSkills(ctx, lens.ID, []domain.LensSkillItem{
		{SkillID: sk1.ID, CustomSortOrder: nil},
		{SkillID: sk2.ID, CustomSortOrder: &customSort},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Skills, 2)
	assert.Equal(t, sk1.ID, detail.Skills[0].SkillID)
	assert.Nil(t, detail.Skills[0].CustomSortOrder)
	assert.Equal(t, sk2.ID, detail.Skills[1].SkillID)
	require.NotNil(t, detail.Skills[1].CustomSortOrder)
	assert.Equal(t, 5, *detail.Skills[1].CustomSortOrder)
}

// =================================================================
// Selection Setters — Academics
// =================================================================

func TestSetLensAcademics(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	ac1, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "MIT", CredentialType: "BSc",
		FieldOfStudy: "CS", CompletionDate: "2020",
	})
	ac2, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "Stanford", CredentialType: "MSc",
		FieldOfStudy: "AI", CompletionDate: "2022",
	})

	err := store.SetLensAcademics(ctx, lens.ID, []int64{ac1.ID, ac2.ID})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.AcademicIDs, 2)
	assert.Contains(t, detail.AcademicIDs, ac1.ID)
	assert.Contains(t, detail.AcademicIDs, ac2.ID)
}

// =================================================================
// Selection Setters — Certifications
// =================================================================

func TestSetLensCerts(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	cert1, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "AWS SA", IssuingBody: "AWS", DateEarned: "2023-01-01",
	})
	cert2, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "CKA", IssuingBody: "CNCF", DateEarned: "2023-06-01",
	})

	err := store.SetLensCerts(ctx, lens.ID, []int64{cert1.ID, cert2.ID})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.CertIDs, 2)
	assert.Contains(t, detail.CertIDs, cert1.ID)
	assert.Contains(t, detail.CertIDs, cert2.ID)
}

// =================================================================
// Selection Setters — Descriptors
// =================================================================

func TestSetLensDescriptors(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	d1, _ := store.CreateDescriptor(ctx, "Software Engineer")
	d2, _ := store.CreateDescriptor(ctx, "Tech Lead")

	err := store.SetLensDescriptors(ctx, lens.ID, []domain.LensDescriptorItem{
		{DescriptorID: d1.ID, SortOrder: 0},
		{DescriptorID: d2.ID, SortOrder: 1},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Descriptors, 2)
	assert.Equal(t, d1.ID, detail.Descriptors[0].DescriptorID)
	assert.Equal(t, 0, detail.Descriptors[0].SortOrder)
	assert.Equal(t, d2.ID, detail.Descriptors[1].DescriptorID)
	assert.Equal(t, 1, detail.Descriptors[1].SortOrder)
}

// =================================================================
// GetLens with all selections populated
// =================================================================

func TestGetLens_FullDetail(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create test data
	summary, _ := store.CreateSummary(ctx, domain.SummaryInput{
		Label: "Backend", BodyText: "Text...",
	})
	lens, _ := store.CreateLens(ctx, domain.LensInput{
		Name: "Backend",
	})

	// Associate summary via junction table
	_ = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: summary.ID, SortOrder: 0},
	})

	wh, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})
	bullet, _ := store.CreateBullet(ctx, wh.ID, "Built APIs", domain.BulletTypePrimary)

	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})

	ac, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "MIT", CredentialType: "BSc",
		FieldOfStudy: "CS", CompletionDate: "2018",
	})
	cert, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "CKA", IssuingBody: "CNCF", DateEarned: "2023-01-01",
	})
	desc, _ := store.CreateDescriptor(ctx, "Backend Engineer")

	// Set all selections
	_ = store.SetLensWorkHistory(ctx, lens.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh.ID, SortOrder: 0},
	})
	_ = store.SetLensBullets(ctx, lens.ID, []domain.LensBulletItem{
		{BulletID: bullet.ID, SortOrder: 0},
	})
	customSort := 3
	_ = store.SetLensSkills(ctx, lens.ID, []domain.LensSkillItem{
		{SkillID: skill.ID, CustomSortOrder: &customSort},
	})
	_ = store.SetLensAcademics(ctx, lens.ID, []int64{ac.ID})
	_ = store.SetLensCerts(ctx, lens.ID, []int64{cert.ID})
	_ = store.SetLensDescriptors(ctx, lens.ID, []domain.LensDescriptorItem{
		{DescriptorID: desc.ID, SortOrder: 0},
	})

	// Get full detail
	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	assert.Equal(t, "Backend", detail.Name)
	require.Len(t, detail.Summaries, 1)
	assert.Equal(t, summary.ID, detail.Summaries[0].SummaryID)

	require.Len(t, detail.WorkHistory, 1)
	assert.Equal(t, wh.ID, detail.WorkHistory[0].WorkHistoryID)

	require.Len(t, detail.Bullets, 1)
	assert.Equal(t, bullet.ID, detail.Bullets[0].BulletID)

	require.Len(t, detail.Skills, 1)
	assert.Equal(t, skill.ID, detail.Skills[0].SkillID)
	require.NotNil(t, detail.Skills[0].CustomSortOrder)
	assert.Equal(t, 3, *detail.Skills[0].CustomSortOrder)

	require.Len(t, detail.AcademicIDs, 1)
	assert.Equal(t, ac.ID, detail.AcademicIDs[0])

	require.Len(t, detail.CertIDs, 1)
	assert.Equal(t, cert.ID, detail.CertIDs[0])

	require.Len(t, detail.Descriptors, 1)
	assert.Equal(t, desc.ID, detail.Descriptors[0].DescriptorID)
}

// =================================================================
// Cascade Deletes
// =================================================================

func TestDeleteLens_CascadesSelections(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	wh, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})

	_ = store.SetLensWorkHistory(ctx, lens.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh.ID, SortOrder: 0},
	})

	// Delete lens — selections should cascade
	err := store.DeleteLens(ctx, lens.ID)
	require.NoError(t, err)

	// Work history entry itself should still exist
	_, err = store.GetWorkHistory(ctx, wh.ID)
	require.NoError(t, err, "work history should survive lens deletion")
}

func TestDeleteContentItem_CascadesFromLens(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})

	_ = store.SetLensSkills(ctx, lens.ID, []domain.LensSkillItem{
		{SkillID: skill.ID},
	})

	// Delete the skill — should cascade-remove from lens selections
	err := store.DeleteSkill(ctx, skill.ID)
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	assert.Len(t, detail.Skills, 0, "skill selection should be removed after skill deleted")
}

// =================================================================
// Skill Lens Tags
// =================================================================

func TestGetSkillLensTags_Empty(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})

	tags, err := store.GetSkillLensTags(ctx, skill.ID)
	require.NoError(t, err)
	assert.NotNil(t, tags)
	assert.Len(t, tags, 0)
}

func TestSetSkillLensTags(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	lens1, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	lens2, _ := store.CreateLens(ctx, domain.LensInput{Name: "DevOps"})

	err := store.SetSkillLensTags(ctx, skill.ID, []int64{lens1.ID, lens2.ID})
	require.NoError(t, err)

	tags, err := store.GetSkillLensTags(ctx, skill.ID)
	require.NoError(t, err)
	require.Len(t, tags, 2)
	assert.Contains(t, tags, lens1.ID)
	assert.Contains(t, tags, lens2.ID)
}

func TestSetSkillLensTags_Replace(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	lens1, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	lens2, _ := store.CreateLens(ctx, domain.LensInput{Name: "DevOps"})

	_ = store.SetSkillLensTags(ctx, skill.ID, []int64{lens1.ID})

	// Replace with different set
	err := store.SetSkillLensTags(ctx, skill.ID, []int64{lens2.ID})
	require.NoError(t, err)

	tags, err := store.GetSkillLensTags(ctx, skill.ID)
	require.NoError(t, err)
	require.Len(t, tags, 1)
	assert.Equal(t, lens2.ID, tags[0])
}

func TestSetSkillLensTags_ClearAll(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	lens1, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})

	_ = store.SetSkillLensTags(ctx, skill.ID, []int64{lens1.ID})

	err := store.SetSkillLensTags(ctx, skill.ID, []int64{})
	require.NoError(t, err)

	tags, err := store.GetSkillLensTags(ctx, skill.ID)
	require.NoError(t, err)
	assert.Len(t, tags, 0)
}

func TestListSkillsWithLensTags(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	sk1, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	sk2, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Python", CategoryID: cat.ID, CompetenceLevel: 7,
	})
	lens1, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	lens2, _ := store.CreateLens(ctx, domain.LensInput{Name: "DevOps"})

	_ = store.SetSkillLensTags(ctx, sk1.ID, []int64{lens1.ID, lens2.ID})
	_ = store.SetSkillLensTags(ctx, sk2.ID, []int64{lens1.ID})

	skills, err := store.ListSkillsWithLensTags(ctx)
	require.NoError(t, err)
	require.Len(t, skills, 2)

	// Skills sorted by competence desc then name asc: Go(9) then Python(7)
	assert.Equal(t, "Go", skills[0].Name)
	require.Len(t, skills[0].LensIDs, 2)

	assert.Equal(t, "Python", skills[1].Name)
	require.Len(t, skills[1].LensIDs, 1)
	assert.Equal(t, lens1.ID, skills[1].LensIDs[0])
}

func TestListSkillsWithLensTags_NoTags(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	_, _ = store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})

	skills, err := store.ListSkillsWithLensTags(ctx)
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "Go", skills[0].Name)
	assert.NotNil(t, skills[0].LensIDs)
	assert.Len(t, skills[0].LensIDs, 0)
}

func TestDeleteLens_CascadesSkillLensTags(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})

	_ = store.SetSkillLensTags(ctx, skill.ID, []int64{lens.ID})

	// Delete lens — tag should cascade
	err := store.DeleteLens(ctx, lens.ID)
	require.NoError(t, err)

	tags, err := store.GetSkillLensTags(ctx, skill.ID)
	require.NoError(t, err)
	assert.Len(t, tags, 0, "skill lens tags should be removed after lens deleted")
}

// =================================================================
// Check*LensReferences
// =================================================================

func TestCheckWorkHistoryLensReferences_NoReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	wh, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})

	names, err := store.CheckWorkHistoryLensReferences(ctx, wh.ID)
	require.NoError(t, err)
	assert.Len(t, names, 0)
}

func TestCheckWorkHistoryLensReferences_WithReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	wh, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})
	lens1, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	lens2, _ := store.CreateLens(ctx, domain.LensInput{Name: "Frontend"})

	_ = store.SetLensWorkHistory(ctx, lens1.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh.ID, SortOrder: 0},
	})
	_ = store.SetLensWorkHistory(ctx, lens2.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh.ID, SortOrder: 0},
	})

	names, err := store.CheckWorkHistoryLensReferences(ctx, wh.ID)
	require.NoError(t, err)
	require.Len(t, names, 2)
	assert.Contains(t, names, "Backend")
	assert.Contains(t, names, "Frontend")
}

func TestCheckBulletLensReferences_NoReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	wh, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})
	bullet, _ := store.CreateBullet(ctx, wh.ID, "Built APIs", domain.BulletTypePrimary)

	names, err := store.CheckBulletLensReferences(ctx, bullet.ID)
	require.NoError(t, err)
	assert.Len(t, names, 0)
}

func TestCheckBulletLensReferences_WithReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	wh, _ := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})
	bullet, _ := store.CreateBullet(ctx, wh.ID, "Built APIs", domain.BulletTypePrimary)
	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})

	_ = store.SetLensBullets(ctx, lens.ID, []domain.LensBulletItem{
		{BulletID: bullet.ID, SortOrder: 0},
	})

	names, err := store.CheckBulletLensReferences(ctx, bullet.ID)
	require.NoError(t, err)
	require.Len(t, names, 1)
	assert.Equal(t, "Backend", names[0])
}

func TestCheckAcademicLensReferences_NoReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	ac, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "MIT", CredentialType: "BSc",
		FieldOfStudy: "CS", CompletionDate: "2020",
	})

	names, err := store.CheckAcademicLensReferences(ctx, ac.ID)
	require.NoError(t, err)
	assert.Len(t, names, 0)
}

func TestCheckAcademicLensReferences_WithReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	ac, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "MIT", CredentialType: "BSc",
		FieldOfStudy: "CS", CompletionDate: "2020",
	})
	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	_ = store.SetLensAcademics(ctx, lens.ID, []int64{ac.ID})

	names, err := store.CheckAcademicLensReferences(ctx, ac.ID)
	require.NoError(t, err)
	require.Len(t, names, 1)
	assert.Equal(t, "Backend", names[0])
}

func TestCheckCertLensReferences_NoReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cert, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "CKA", IssuingBody: "CNCF", DateEarned: "2023-01-01",
	})

	names, err := store.CheckCertLensReferences(ctx, cert.ID)
	require.NoError(t, err)
	assert.Len(t, names, 0)
}

func TestCheckCertLensReferences_WithReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cert, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "CKA", IssuingBody: "CNCF", DateEarned: "2023-01-01",
	})
	lens1, _ := store.CreateLens(ctx, domain.LensInput{Name: "DevOps"})
	lens2, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	_ = store.SetLensCerts(ctx, lens1.ID, []int64{cert.ID})
	_ = store.SetLensCerts(ctx, lens2.ID, []int64{cert.ID})

	names, err := store.CheckCertLensReferences(ctx, cert.ID)
	require.NoError(t, err)
	require.Len(t, names, 2)
	assert.Contains(t, names, "DevOps")
	assert.Contains(t, names, "Backend")
}

func TestCheckDescriptorLensReferences_NoReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	desc, _ := store.CreateDescriptor(ctx, "Software Engineer")

	names, err := store.CheckDescriptorLensReferences(ctx, desc.ID)
	require.NoError(t, err)
	assert.Len(t, names, 0)
}

func TestCheckDescriptorLensReferences_WithReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	desc, _ := store.CreateDescriptor(ctx, "Software Engineer")
	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})
	_ = store.SetLensDescriptors(ctx, lens.ID, []domain.LensDescriptorItem{
		{DescriptorID: desc.ID, SortOrder: 0},
	})

	names, err := store.CheckDescriptorLensReferences(ctx, desc.ID)
	require.NoError(t, err)
	require.Len(t, names, 1)
	assert.Equal(t, "Backend", names[0])
}

func TestDeleteSkill_CascadesSkillLensTags(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cat, _ := store.CreateSkillCategory(ctx, "Lang")
	skill, _ := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	lens, _ := store.CreateLens(ctx, domain.LensInput{Name: "Backend"})

	_ = store.SetSkillLensTags(ctx, skill.ID, []int64{lens.ID})

	// Delete skill — tag should cascade
	err := store.DeleteSkill(ctx, skill.ID)
	require.NoError(t, err)

	// Verify via ListSkillsWithLensTags (skill is gone entirely)
	skills, err := store.ListSkillsWithLensTags(ctx)
	require.NoError(t, err)
	assert.Len(t, skills, 0)
}

// =================================================================
// SetLensSummaries Tests
// =================================================================

func TestSetLensSummaries_MultipleSummaries(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "A", BodyText: "Text A"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "B", BodyText: "Text B"})
	require.NoError(t, err)
	s3, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "C", BodyText: "Text C"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "Multi"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s1.ID, SortOrder: 0},
		{SummaryID: s2.ID, SortOrder: 1},
		{SummaryID: s3.ID, SortOrder: 2},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 3)
	assert.Equal(t, s1.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, 0, detail.Summaries[0].SortOrder)
	assert.Equal(t, s2.ID, detail.Summaries[1].SummaryID)
	assert.Equal(t, 1, detail.Summaries[1].SortOrder)
	assert.Equal(t, s3.ID, detail.Summaries[2].SummaryID)
	assert.Equal(t, 2, detail.Summaries[2].SortOrder)
}

func TestSetLensSummaries_Replace(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	sA, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "A", BodyText: "Text A"})
	require.NoError(t, err)
	sB, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "B", BodyText: "Text B"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "Replace"})
	require.NoError(t, err)

	// Set summary A
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: sA.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Replace with summary B
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: sB.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 1)
	assert.Equal(t, sB.ID, detail.Summaries[0].SummaryID)
}

func TestSetLensSummaries_SortOrderPreserved(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "X", BodyText: "X"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Y", BodyText: "Y"})
	require.NoError(t, err)
	s3, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Z", BodyText: "Z"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "SortTest"})
	require.NoError(t, err)

	// Insert with sort_orders 2, 0, 1 (not in order)
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s1.ID, SortOrder: 2},
		{SummaryID: s2.ID, SortOrder: 0},
		{SummaryID: s3.ID, SortOrder: 1},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 3)
	// Should be returned sorted by sort_order ASC: 0, 1, 2
	assert.Equal(t, 0, detail.Summaries[0].SortOrder)
	assert.Equal(t, s2.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, 1, detail.Summaries[1].SortOrder)
	assert.Equal(t, s3.ID, detail.Summaries[1].SummaryID)
	assert.Equal(t, 2, detail.Summaries[2].SortOrder)
	assert.Equal(t, s1.ID, detail.Summaries[2].SummaryID)
}

func TestSetLensSummaries_SameSummaryDifferentLenses(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Shared", BodyText: "Shared text"})
	require.NoError(t, err)

	lens1, err := store.CreateLens(ctx, domain.LensInput{Name: "Lens1"})
	require.NoError(t, err)
	lens2, err := store.CreateLens(ctx, domain.LensInput{Name: "Lens2"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens1.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens2.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	detail1, err := store.GetLens(ctx, lens1.ID)
	require.NoError(t, err)
	require.Len(t, detail1.Summaries, 1)
	assert.Equal(t, s.ID, detail1.Summaries[0].SummaryID)

	detail2, err := store.GetLens(ctx, lens2.ID)
	require.NoError(t, err)
	require.Len(t, detail2.Summaries, 1)
	assert.Equal(t, s.ID, detail2.Summaries[0].SummaryID)
}

func TestSetLensSummaries_EmptyToEmpty(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "Empty"})
	require.NoError(t, err)

	// Clear when already empty
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	assert.Len(t, detail.Summaries, 0)
}

func TestSetLensSummaries_ReplaceWithDifferentSet(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	sA, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "A", BodyText: "A"})
	require.NoError(t, err)
	sB, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "B", BodyText: "B"})
	require.NoError(t, err)
	sC, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "C", BodyText: "C"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "ReplaceSet"})
	require.NoError(t, err)

	// Set {A, B}
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: sA.ID, SortOrder: 0},
		{SummaryID: sB.ID, SortOrder: 1},
	})
	require.NoError(t, err)

	// Replace with {B, C}
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: sB.ID, SortOrder: 0},
		{SummaryID: sC.ID, SortOrder: 1},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 2)
	assert.Equal(t, sB.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, sC.ID, detail.Summaries[1].SummaryID)
}

func TestSetLensSummaries_SameDataIdempotent(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Idem", BodyText: "Idem"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "Idempotent"})
	require.NoError(t, err)

	items := []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	}

	// Set same data twice
	err = store.SetLensSummaries(ctx, lens.ID, items)
	require.NoError(t, err)
	err = store.SetLensSummaries(ctx, lens.ID, items)
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 1)
	assert.Equal(t, s.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, 0, detail.Summaries[0].SortOrder)
}

func TestSetLensSummaries_DoesNotAffectOtherSelections(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "Isolated"})
	require.NoError(t, err)

	wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Corp", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})
	require.NoError(t, err)

	cat, err := store.CreateSkillCategory(ctx, "Lang")
	require.NoError(t, err)
	sk, err := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	require.NoError(t, err)

	// Set work history and skills
	err = store.SetLensWorkHistory(ctx, lens.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh.ID, SortOrder: 0},
	})
	require.NoError(t, err)
	err = store.SetLensSkills(ctx, lens.ID, []domain.LensSkillItem{
		{SkillID: sk.ID},
	})
	require.NoError(t, err)

	// Now set summaries
	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Sum", BodyText: "T"})
	require.NoError(t, err)
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Verify work history and skills are unchanged
	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 1)
	require.Len(t, detail.WorkHistory, 1)
	assert.Equal(t, wh.ID, detail.WorkHistory[0].WorkHistoryID)
	require.Len(t, detail.Skills, 1)
	assert.Equal(t, sk.ID, detail.Skills[0].SkillID)
}

func TestSetLensSummaries_DoesNotAffectOtherLens(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "S1", BodyText: "T1"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "S2", BodyText: "T2"})
	require.NoError(t, err)

	lensA, err := store.CreateLens(ctx, domain.LensInput{Name: "LensA"})
	require.NoError(t, err)
	lensB, err := store.CreateLens(ctx, domain.LensInput{Name: "LensB"})
	require.NoError(t, err)

	// Set summary on lens B
	err = store.SetLensSummaries(ctx, lensB.ID, []domain.LensSummaryItem{
		{SummaryID: s2.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Set summary on lens A — should not affect lens B
	err = store.SetLensSummaries(ctx, lensA.ID, []domain.LensSummaryItem{
		{SummaryID: s1.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	detailB, err := store.GetLens(ctx, lensB.ID)
	require.NoError(t, err)
	require.Len(t, detailB.Summaries, 1)
	assert.Equal(t, s2.ID, detailB.Summaries[0].SummaryID)
}

func TestSetLensSummaries_NilSlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Nil", BodyText: "Nil"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "NilSlice"})
	require.NoError(t, err)

	// Set a summary first
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Pass nil slice — should clear all
	err = store.SetLensSummaries(ctx, lens.ID, nil)
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	assert.Len(t, detail.Summaries, 0)
}

// =================================================================
// CheckSummaryLensReferences Tests
// =================================================================

func TestCheckSummaryLensReferences_NoReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Alone", BodyText: "Alone"})
	require.NoError(t, err)

	names, err := store.CheckSummaryLensReferences(ctx, s.ID)
	require.NoError(t, err)
	assert.Len(t, names, 0)
}

func TestCheckSummaryLensReferences_WithOneReference(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "One", BodyText: "One"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "OnlyLens"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	names, err := store.CheckSummaryLensReferences(ctx, s.ID)
	require.NoError(t, err)
	require.Len(t, names, 1)
	assert.Equal(t, "OnlyLens", names[0])
}

func TestCheckSummaryLensReferences_WithMultipleReferences(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Multi", BodyText: "Multi"})
	require.NoError(t, err)

	lens1, err := store.CreateLens(ctx, domain.LensInput{Name: "Alpha"})
	require.NoError(t, err)
	lens2, err := store.CreateLens(ctx, domain.LensInput{Name: "Beta"})
	require.NoError(t, err)
	lens3, err := store.CreateLens(ctx, domain.LensInput{Name: "Gamma"})
	require.NoError(t, err)

	for _, l := range []domain.Lens{lens1, lens2, lens3} {
		err = store.SetLensSummaries(ctx, l.ID, []domain.LensSummaryItem{
			{SummaryID: s.ID, SortOrder: 0},
		})
		require.NoError(t, err)
	}

	names, err := store.CheckSummaryLensReferences(ctx, s.ID)
	require.NoError(t, err)
	require.Len(t, names, 3)
	assert.Contains(t, names, "Alpha")
	assert.Contains(t, names, "Beta")
	assert.Contains(t, names, "Gamma")
}

func TestCheckSummaryLensReferences_ReturnsAlphabeticalOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Order", BodyText: "Order"})
	require.NoError(t, err)

	// Create lenses in non-alphabetical order
	lensC, err := store.CreateLens(ctx, domain.LensInput{Name: "Charlie"})
	require.NoError(t, err)
	lensA, err := store.CreateLens(ctx, domain.LensInput{Name: "Alpha"})
	require.NoError(t, err)
	lensB, err := store.CreateLens(ctx, domain.LensInput{Name: "Bravo"})
	require.NoError(t, err)

	for _, l := range []domain.Lens{lensC, lensA, lensB} {
		err = store.SetLensSummaries(ctx, l.ID, []domain.LensSummaryItem{
			{SummaryID: s.ID, SortOrder: 0},
		})
		require.NoError(t, err)
	}

	names, err := store.CheckSummaryLensReferences(ctx, s.ID)
	require.NoError(t, err)
	require.Len(t, names, 3)
	assert.Equal(t, "Alpha", names[0])
	assert.Equal(t, "Bravo", names[1])
	assert.Equal(t, "Charlie", names[2])
}

func TestCheckSummaryLensReferences_NonExistentSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	names, err := store.CheckSummaryLensReferences(ctx, 999)
	require.NoError(t, err)
	assert.Len(t, names, 0)
}

// =================================================================
// GetLens Summary Integration Tests
// =================================================================

func TestGetLens_SummariesSortedBySortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "First", BodyText: "1"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Second", BodyText: "2"})
	require.NoError(t, err)
	s3, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Third", BodyText: "3"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "Sorted"})
	require.NoError(t, err)

	// Set with reversed sort orders
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s3.ID, SortOrder: 0},
		{SummaryID: s1.ID, SortOrder: 2},
		{SummaryID: s2.ID, SortOrder: 1},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 3)
	assert.Equal(t, 0, detail.Summaries[0].SortOrder)
	assert.Equal(t, s3.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, 1, detail.Summaries[1].SortOrder)
	assert.Equal(t, s2.ID, detail.Summaries[1].SummaryID)
	assert.Equal(t, 2, detail.Summaries[2].SortOrder)
	assert.Equal(t, s1.ID, detail.Summaries[2].SummaryID)
}

func TestGetLens_MultipleSummariesWithGaps(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Gap0", BodyText: "0"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Gap5", BodyText: "5"})
	require.NoError(t, err)
	s3, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Gap10", BodyText: "10"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "Gaps"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s1.ID, SortOrder: 0},
		{SummaryID: s2.ID, SortOrder: 5},
		{SummaryID: s3.ID, SortOrder: 10},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 3)
	assert.Equal(t, 0, detail.Summaries[0].SortOrder)
	assert.Equal(t, s1.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, 5, detail.Summaries[1].SortOrder)
	assert.Equal(t, s2.ID, detail.Summaries[1].SummaryID)
	assert.Equal(t, 10, detail.Summaries[2].SortOrder)
	assert.Equal(t, s3.ID, detail.Summaries[2].SummaryID)
}

// =================================================================
// CASCADE Behavior Tests
// =================================================================

func TestDeleteSummary_CascadesFromMultipleLenses(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Shared", BodyText: "Shared"})
	require.NoError(t, err)

	lens1, err := store.CreateLens(ctx, domain.LensInput{Name: "CascL1"})
	require.NoError(t, err)
	lens2, err := store.CreateLens(ctx, domain.LensInput{Name: "CascL2"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens1.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)
	err = store.SetLensSummaries(ctx, lens2.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Delete the summary
	err = store.DeleteSummary(ctx, s.ID)
	require.NoError(t, err)

	// Both lenses should have 0 summaries
	detail1, err := store.GetLens(ctx, lens1.ID)
	require.NoError(t, err)
	assert.Len(t, detail1.Summaries, 0)

	detail2, err := store.GetLens(ctx, lens2.ID)
	require.NoError(t, err)
	assert.Len(t, detail2.Summaries, 0)
}

func TestDeleteLens_CascadesSummarySelections(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Survive", BodyText: "Survive"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "ToDel"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Delete the lens
	err = store.DeleteLens(ctx, lens.ID)
	require.NoError(t, err)

	// Summary should still exist independently
	summaries, err := store.ListSummaries(ctx)
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	assert.Equal(t, s.ID, summaries[0].ID)
}

func TestDeleteLens_WithAllSelectionTypes_CascadesAll(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create content entities
	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Sum", BodyText: "Sum"})
	require.NoError(t, err)

	wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Corp", JobTitle: "Dev",
		StartDate: "2020-01", DateGranularityStart: "month",
	})
	require.NoError(t, err)

	cat, err := store.CreateSkillCategory(ctx, "Lang")
	require.NoError(t, err)
	sk, err := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "FullDelete"})
	require.NoError(t, err)

	// Set all selections
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)
	err = store.SetLensWorkHistory(ctx, lens.ID, []domain.LensWorkHistoryItem{
		{WorkHistoryID: wh.ID, SortOrder: 0},
	})
	require.NoError(t, err)
	err = store.SetLensSkills(ctx, lens.ID, []domain.LensSkillItem{
		{SkillID: sk.ID},
	})
	require.NoError(t, err)

	// Delete the lens
	err = store.DeleteLens(ctx, lens.ID)
	require.NoError(t, err)

	// Verify all content entities still exist
	summaries, err := store.ListSummaries(ctx)
	require.NoError(t, err)
	assert.Len(t, summaries, 1)

	_, err = store.GetWorkHistory(ctx, wh.ID)
	require.NoError(t, err)

	skills, err := store.ListSkills(ctx)
	require.NoError(t, err)
	assert.Len(t, skills, 1)
}

// =================================================================
// Cross-Function Interaction Tests
// =================================================================

func TestSetLensSummaries_ThenSetAgain_FullReplace(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Old1", BodyText: "Old1"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Old2", BodyText: "Old2"})
	require.NoError(t, err)
	s3, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Old3", BodyText: "Old3"})
	require.NoError(t, err)
	s4, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "New1", BodyText: "New1"})
	require.NoError(t, err)
	s5, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "New2", BodyText: "New2"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "FullReplace"})
	require.NoError(t, err)

	// Set 3 summaries
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s1.ID, SortOrder: 0},
		{SummaryID: s2.ID, SortOrder: 1},
		{SummaryID: s3.ID, SortOrder: 2},
	})
	require.NoError(t, err)

	// Replace with 2 different summaries
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s4.ID, SortOrder: 0},
		{SummaryID: s5.ID, SortOrder: 1},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 2)
	assert.Equal(t, s4.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, s5.ID, detail.Summaries[1].SummaryID)
}

func TestMultipleLenses_IndependentSummaries(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "ForA", BodyText: "ForA"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "ForB", BodyText: "ForB"})
	require.NoError(t, err)

	lensA, err := store.CreateLens(ctx, domain.LensInput{Name: "IndepA"})
	require.NoError(t, err)
	lensB, err := store.CreateLens(ctx, domain.LensInput{Name: "IndepB"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lensA.ID, []domain.LensSummaryItem{
		{SummaryID: s1.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lensB.ID, []domain.LensSummaryItem{
		{SummaryID: s2.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	detailA, err := store.GetLens(ctx, lensA.ID)
	require.NoError(t, err)
	require.Len(t, detailA.Summaries, 1)
	assert.Equal(t, s1.ID, detailA.Summaries[0].SummaryID)

	detailB, err := store.GetLens(ctx, lensB.ID)
	require.NoError(t, err)
	require.Len(t, detailB.Summaries, 1)
	assert.Equal(t, s2.ID, detailB.Summaries[0].SummaryID)
}

func TestLensSummary_RoundtripThroughExportImport(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Export", BodyText: "ExportText"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "ExportLens"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 3},
	})
	require.NoError(t, err)

	// Export
	data, err := store.ExportAllData(ctx)
	require.NoError(t, err)
	require.Len(t, data.Lenses, 1)
	require.Len(t, data.Lenses[0].Summaries, 1)
	assert.Equal(t, s.ID, data.Lenses[0].Summaries[0].SummaryID)
	assert.Equal(t, 3, data.Lenses[0].Summaries[0].SortOrder)

	// Import into a fresh store
	store2 := testStore(t)
	require.NoError(t, Migrate(store2))

	err = store2.ImportAllData(ctx, data)
	require.NoError(t, err)

	// Verify round-trip
	lenses, err := store2.ListLenses(ctx)
	require.NoError(t, err)
	require.Len(t, lenses, 1)

	detail, err := store2.GetLens(ctx, lenses[0].ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 1)
	assert.Equal(t, 3, detail.Summaries[0].SortOrder)
}

// =================================================================
// Additional Edge Cases
// =================================================================

func TestSetLensSummaries_LargeSortOrders(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "L100", BodyText: "100"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "L200", BodyText: "200"})
	require.NoError(t, err)
	s3, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "L300", BodyText: "300"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "LargeSort"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s1.ID, SortOrder: 100},
		{SummaryID: s2.ID, SortOrder: 200},
		{SummaryID: s3.ID, SortOrder: 300},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 3)
	assert.Equal(t, 100, detail.Summaries[0].SortOrder)
	assert.Equal(t, 200, detail.Summaries[1].SortOrder)
	assert.Equal(t, 300, detail.Summaries[2].SortOrder)
}

func TestSetLensSummaries_SingleSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Solo", BodyText: "Solo"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "Single"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 1)
	assert.Equal(t, s.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, 0, detail.Summaries[0].SortOrder)
}

func TestGetLens_EmptySummariesNonNil(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "EmptyNonNil"})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	assert.NotNil(t, detail.Summaries, "empty summaries slice should not be nil")
	assert.Len(t, detail.Summaries, 0)
}

func TestCheckSummaryLensReferences_AfterLensDeletion(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "RefDel", BodyText: "RefDel"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "ToDelete"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Verify reference exists
	names, err := store.CheckSummaryLensReferences(ctx, s.ID)
	require.NoError(t, err)
	require.Len(t, names, 1)

	// Delete the lens
	err = store.DeleteLens(ctx, lens.ID)
	require.NoError(t, err)

	// References should now be empty
	names, err = store.CheckSummaryLensReferences(ctx, s.ID)
	require.NoError(t, err)
	assert.Len(t, names, 0)
}

func TestCheckSummaryLensReferences_AfterSummaryCleared(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Clear", BodyText: "Clear"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "ClearLens"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Verify reference exists
	names, err := store.CheckSummaryLensReferences(ctx, s.ID)
	require.NoError(t, err)
	require.Len(t, names, 1)

	// Clear summaries from lens
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{})
	require.NoError(t, err)

	// References should now be empty
	names, err = store.CheckSummaryLensReferences(ctx, s.ID)
	require.NoError(t, err)
	assert.Len(t, names, 0)
}

func TestSetLensSummaries_ClearThenSetNew(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Old", BodyText: "Old"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "New", BodyText: "New"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "ClearSet"})
	require.NoError(t, err)

	// Set old summary
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s1.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Clear all
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{})
	require.NoError(t, err)

	// Set new summary
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s2.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 1)
	assert.Equal(t, s2.ID, detail.Summaries[0].SummaryID)
}

func TestDeleteSummary_DoesNotDeleteLens(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "Del", BodyText: "Del"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "Survives"})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Delete the summary
	err = store.DeleteSummary(ctx, s.ID)
	require.NoError(t, err)

	// Lens should still exist
	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	assert.Equal(t, "Survives", detail.Name)
	assert.Len(t, detail.Summaries, 0)

	// Lens should still appear in listing
	lenses, err := store.ListLenses(ctx)
	require.NoError(t, err)
	require.Len(t, lenses, 1)
	assert.Equal(t, lens.ID, lenses[0].ID)
}

func TestSetLensSummaries_OrderMatchesInput(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	s1, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "M1", BodyText: "M1"})
	require.NoError(t, err)
	s2, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "M2", BodyText: "M2"})
	require.NoError(t, err)
	s3, err := store.CreateSummary(ctx, domain.SummaryInput{Label: "M3", BodyText: "M3"})
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{Name: "InputOrder"})
	require.NoError(t, err)

	// Insert in reverse ID order but explicit sort_order
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: s3.ID, SortOrder: 0},
		{SummaryID: s2.ID, SortOrder: 1},
		{SummaryID: s1.ID, SortOrder: 2},
	})
	require.NoError(t, err)

	detail, err := store.GetLens(ctx, lens.ID)
	require.NoError(t, err)
	require.Len(t, detail.Summaries, 3)
	// Returned order should match sort_order, not insertion order or ID order
	assert.Equal(t, s3.ID, detail.Summaries[0].SummaryID)
	assert.Equal(t, 0, detail.Summaries[0].SortOrder)
	assert.Equal(t, s2.ID, detail.Summaries[1].SummaryID)
	assert.Equal(t, 1, detail.Summaries[1].SortOrder)
	assert.Equal(t, s1.ID, detail.Summaries[2].SummaryID)
	assert.Equal(t, 2, detail.Summaries[2].SortOrder)
}
