package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exportTestData creates the seed data needed for export selection
// tests and returns a struct with IDs for all created entities.
type exportSeedData struct {
	workHistoryIDs []int64
	bulletIDs      []int64
	skillIDs       []int64
	academicIDs    []int64
	certIDs        []int64
	descriptorIDs  []int64
	summaryID      int64
}

func seedExportTestData(t *testing.T, store *Store, ctx context.Context) exportSeedData {
	t.Helper()
	var data exportSeedData

	// Work history.
	wh1, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme Corp",
		JobTitle:     "Engineer",
		StartDate:    "2020-01",
	})
	require.NoError(t, err)
	data.workHistoryIDs = append(data.workHistoryIDs, wh1.ID)

	wh2, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "TechCo",
		JobTitle:     "Developer",
		StartDate:    "2018-01",
	})
	require.NoError(t, err)
	data.workHistoryIDs = append(data.workHistoryIDs, wh2.ID)

	// Bullets.
	b1, err := store.CreateBullet(ctx, wh1.ID, "Built something great")
	require.NoError(t, err)
	data.bulletIDs = append(data.bulletIDs, b1.ID)

	b2, err := store.CreateBullet(ctx, wh1.ID, "Led a team")
	require.NoError(t, err)
	data.bulletIDs = append(data.bulletIDs, b2.ID)

	// Skill category + skills.
	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	s1, err := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	require.NoError(t, err)
	data.skillIDs = append(data.skillIDs, s1.ID)

	s2, err := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Python", CategoryID: cat.ID, CompetenceLevel: 8,
	})
	require.NoError(t, err)
	data.skillIDs = append(data.skillIDs, s2.ID)

	// Academic.
	ac, err := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution:    "MIT",
		CredentialType: "BS",
		FieldOfStudy:   "CS",
		CompletionDate: "2017-05",
	})
	require.NoError(t, err)
	data.academicIDs = append(data.academicIDs, ac.ID)

	// Certification.
	cert, err := store.CreateCertification(ctx, domain.CertificationInput{
		Name:        "AWS SA",
		IssuingBody: "Amazon",
		DateEarned:  "2022-01",
	})
	require.NoError(t, err)
	data.certIDs = append(data.certIDs, cert.ID)

	// Descriptor.
	desc, err := store.CreateDescriptor(ctx, "Software Engineer")
	require.NoError(t, err)
	data.descriptorIDs = append(data.descriptorIDs, desc.ID)

	desc2, err := store.CreateDescriptor(ctx, "Technical Lead")
	require.NoError(t, err)
	data.descriptorIDs = append(data.descriptorIDs, desc2.ID)

	// Summary.
	sum, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "General",
		BodyText: "Experienced engineer.",
	})
	require.NoError(t, err)
	data.summaryID = sum.ID

	return data
}

func TestCreateExport_Basic(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	export := domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	}

	created, err := store.CreateExport(ctx, export)
	require.NoError(t, err)
	assert.Greater(t, created.ID, int64(0))
	assert.Equal(t, "professional", created.TemplateID)
	assert.Equal(t, "/tmp/resume.pdf", created.FilePath)
	assert.NotEmpty(t, created.GeneratedAt)
}

func TestCreateExport_WithSummaryAndLens(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	seed := seedExportTestData(t, store, ctx)

	export := domain.ResumeExport{
		TemplateID: "modern",
		FilePath:   "/tmp/modern.pdf",
		SummaryID:  &seed.summaryID,
		// LensID left nil — no lens table seed data needed.
	}

	created, err := store.CreateExport(ctx, export)
	require.NoError(t, err)
	require.NotNil(t, created.SummaryID)
	assert.Equal(t, seed.summaryID, *created.SummaryID)
	assert.Nil(t, created.LensID)
}

func TestGetExport(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	export := domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	}

	created, err := store.CreateExport(ctx, export)
	require.NoError(t, err)

	got, err := store.GetExport(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.TemplateID, got.TemplateID)
	assert.Equal(t, created.FilePath, got.FilePath)
}

func TestGetExport_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.GetExport(ctx, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListExports_Empty(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	exports, err := store.ListExports(ctx)
	require.NoError(t, err)
	assert.Empty(t, exports)
}

func TestListExports_Multiple(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := store.CreateExport(ctx, domain.ResumeExport{
			TemplateID: "professional",
			FilePath:   "/tmp/resume.pdf",
		})
		require.NoError(t, err)
	}

	exports, err := store.ListExports(ctx)
	require.NoError(t, err)
	assert.Len(t, exports, 3)
}

func TestListExports_OrderedByIDDesc(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	e1, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/first.pdf",
	})
	require.NoError(t, err)

	e2, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "modern",
		FilePath:   "/tmp/second.pdf",
	})
	require.NoError(t, err)

	exports, err := store.ListExports(ctx)
	require.NoError(t, err)
	require.Len(t, exports, 2)

	// Most recent first (by ID since timestamps may collide in tests).
	assert.Equal(t, e2.ID, exports[0].ID)
	assert.Equal(t, e1.ID, exports[1].ID)
}

func TestCreateExportSelections_WorkHistory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	seed := seedExportTestData(t, store, ctx)

	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	})
	require.NoError(t, err)

	err = store.CreateExportSelections(ctx, export.ID, domain.ExportRequest{
		WorkHistoryIDs: seed.workHistoryIDs,
	})
	require.NoError(t, err)

	ids, err := store.GetExportWorkHistoryIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Equal(t, seed.workHistoryIDs, ids)
}

func TestCreateExportSelections_Bullets(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	seed := seedExportTestData(t, store, ctx)

	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	})
	require.NoError(t, err)

	err = store.CreateExportSelections(ctx, export.ID, domain.ExportRequest{
		BulletIDs: seed.bulletIDs,
	})
	require.NoError(t, err)

	ids, err := store.GetExportBulletIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Equal(t, seed.bulletIDs, ids)
}

func TestCreateExportSelections_Skills(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	seed := seedExportTestData(t, store, ctx)

	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	})
	require.NoError(t, err)

	err = store.CreateExportSelections(ctx, export.ID, domain.ExportRequest{
		SkillIDs:           seed.skillIDs,
		SkillSortOverrides: map[int64]int{seed.skillIDs[0]: 1, seed.skillIDs[1]: 0},
	})
	require.NoError(t, err)

	ids, err := store.GetExportSkillIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Equal(t, seed.skillIDs, ids)
}

func TestCreateExportSelections_Academics(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	seed := seedExportTestData(t, store, ctx)

	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	})
	require.NoError(t, err)

	err = store.CreateExportSelections(ctx, export.ID, domain.ExportRequest{
		AcademicIDs: seed.academicIDs,
	})
	require.NoError(t, err)

	ids, err := store.GetExportAcademicIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Equal(t, seed.academicIDs, ids)
}

func TestCreateExportSelections_Certifications(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	seed := seedExportTestData(t, store, ctx)

	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	})
	require.NoError(t, err)

	err = store.CreateExportSelections(ctx, export.ID, domain.ExportRequest{
		CertificationIDs: seed.certIDs,
	})
	require.NoError(t, err)

	ids, err := store.GetExportCertIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Equal(t, seed.certIDs, ids)
}

func TestCreateExportSelections_Descriptors(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	seed := seedExportTestData(t, store, ctx)

	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	})
	require.NoError(t, err)

	err = store.CreateExportSelections(ctx, export.ID, domain.ExportRequest{
		DescriptorIDs: seed.descriptorIDs,
	})
	require.NoError(t, err)

	ids, err := store.GetExportDescriptorIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Equal(t, seed.descriptorIDs, ids)
}

func TestCreateExportSelections_AllTypes(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	seed := seedExportTestData(t, store, ctx)

	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	})
	require.NoError(t, err)

	req := domain.ExportRequest{
		WorkHistoryIDs:   seed.workHistoryIDs,
		BulletIDs:        seed.bulletIDs,
		SkillIDs:         seed.skillIDs,
		AcademicIDs:      seed.academicIDs,
		CertificationIDs: seed.certIDs,
		DescriptorIDs:    seed.descriptorIDs,
	}

	err = store.CreateExportSelections(ctx, export.ID, req)
	require.NoError(t, err)

	whIDs, err := store.GetExportWorkHistoryIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Len(t, whIDs, 2)

	bIDs, err := store.GetExportBulletIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Len(t, bIDs, 2)

	sIDs, err := store.GetExportSkillIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Len(t, sIDs, 2)

	aIDs, err := store.GetExportAcademicIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Len(t, aIDs, 1)

	cIDs, err := store.GetExportCertIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Len(t, cIDs, 1)

	dIDs, err := store.GetExportDescriptorIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Len(t, dIDs, 2)
}

func TestCreateExportSelections_Empty(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	})
	require.NoError(t, err)

	// Creating selections with empty slices should succeed.
	err = store.CreateExportSelections(ctx, export.ID, domain.ExportRequest{})
	require.NoError(t, err)

	whIDs, err := store.GetExportWorkHistoryIDs(ctx, export.ID)
	require.NoError(t, err)
	assert.Empty(t, whIDs)
}
