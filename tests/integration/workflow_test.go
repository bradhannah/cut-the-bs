package integration

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"cut-the-bs/internal/domain"
	"cut-the-bs/internal/infra/pdf"
	"cut-the-bs/internal/infra/sqlite"
	"cut-the-bs/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testStore creates a fresh SQLite store with migrations applied,
// registered for cleanup when the test ends.
func testStore(t *testing.T) *sqlite.Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	store, err := sqlite.NewStore(dbPath, logger)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, sqlite.Migrate(store))
	return store
}

func TestWorkflow_CreateProfile(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// GetProfile auto-creates the row if it doesn't exist.
	_, err := store.GetProfile(ctx)
	require.NoError(t, err)

	profile, err := store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Jane Smith",
		Email:    "jane@example.com",
		Phone:    "555-0100",
		Location: "New York, NY",
	})
	require.NoError(t, err)
	assert.Equal(t, "Jane Smith", profile.FullName)
	assert.Equal(t, "jane@example.com", profile.Email)
	assert.NotZero(t, profile.ID)
}

func TestWorkflow_AddProfileLinks(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	link, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "https://linkedin.com/in/janesmith",
	})
	require.NoError(t, err)
	assert.Equal(t, "LinkedIn", link.Label)
	assert.NotZero(t, link.ID)

	links, err := store.ListProfileLinks(ctx)
	require.NoError(t, err)
	assert.Len(t, links, 1)
}

func TestWorkflow_AddWorkHistory(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Senior Software Engineer",
		StartDate:            "2020-01",
		EndDate:              "",
		DateGranularityStart: "month",
		DateGranularityEnd:   "",
	})
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", wh.EmployerName)

	bullet, err := store.CreateBullet(ctx, wh.ID, "Led migration to microservices architecture", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, wh.ID, bullet.WorkHistoryID)
	assert.NotZero(t, bullet.ID)
}

func TestWorkflow_AddSkills(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	cat, err := store.CreateSkillCategory(ctx, "Programming Languages")
	require.NoError(t, err)
	assert.NotZero(t, cat.ID)

	skill, err := store.CreateSkill(ctx, domain.SkillInput{
		Name:            "Go",
		CategoryID:      cat.ID,
		CompetenceLevel: 9,
		IsLegacy:        false,
	})
	require.NoError(t, err)
	assert.Equal(t, "Go", skill.Name)
	assert.Equal(t, 9, skill.CompetenceLevel)
}

func TestWorkflow_CreateLens(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Create prerequisite data.
	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "General",
		BodyText: "Experienced engineer.",
	})
	require.NoError(t, err)

	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	skill, err := store.CreateSkill(ctx, domain.SkillInput{
		Name:            "Go",
		CategoryID:      cat.ID,
		CompetenceLevel: 9,
	})
	require.NoError(t, err)

	wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		StartDate:            "2020-01",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)

	// Create lens.
	lens, err := store.CreateLens(ctx, domain.LensInput{
		Name: "Backend Focus",
	})
	require.NoError(t, err)
	assert.Equal(t, "Backend Focus", lens.Name)

	// Associate summary with the lens via junction table.
	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: summary.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	// Associate skills and work history with the lens.
	err = store.SetLensSkills(ctx, lens.ID, []domain.LensSkillItem{
		{SkillID: skill.ID},
	})
	require.NoError(t, err)

	err = store.SetSkillLensTags(ctx, skill.ID, []int64{lens.ID})
	require.NoError(t, err)

	tags, err := store.GetSkillLensTags(ctx, skill.ID)
	require.NoError(t, err)
	assert.Contains(t, tags, lens.ID)

	_ = wh // wh created for completeness; lens-work association tested separately
}

func TestWorkflow_GenerateResume(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Seed all data.
	_, err := store.GetProfile(ctx) // ensure row exists
	require.NoError(t, err)
	profile, err := store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Jane Smith",
		Email:    "jane@example.com",
		Phone:    "555-0100",
		Location: "New York, NY",
	})
	require.NoError(t, err)

	link, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "GitHub",
		URL:   "https://github.com/janesmith",
	})
	require.NoError(t, err)

	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "General",
		BodyText: "Experienced software engineer specializing in distributed systems.",
	})
	require.NoError(t, err)

	wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Senior Software Engineer",
		StartDate:            "2020-01",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, wh.ID, "Led migration to microservices architecture", domain.BulletTypePrimary)
	require.NoError(t, err)

	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	goSkill, err := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	require.NoError(t, err)

	academic, err := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution:     "MIT",
		CredentialType:  "Bachelor of Science",
		FieldOfStudy:    "Computer Science",
		CompletionDate:  "2017-05",
		DateGranularity: "month",
	})
	require.NoError(t, err)

	cert, err := store.CreateCertification(ctx, domain.CertificationInput{
		Name:        "AWS Solutions Architect",
		IssuingBody: "Amazon Web Services",
		DateEarned:  "2022-03",
	})
	require.NoError(t, err)

	desc, err := store.CreateDescriptor(ctx, "Software Engineer")
	require.NoError(t, err)

	// Fetch work history with bullets for the render request.
	workList, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)

	// Build render request from stored data.
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	profTmpl2 := pdf.ProfessionalTemplate()
	pdfPath, err := renderer.RenderResume(ctx, domain.RenderResumeRequest{
		Template:    &profTmpl2,
		OutputDir:   outputDir,
		Profile:     profile,
		Links:       []domain.ProfileLink{link},
		Summaries:   []domain.ProfessionalSummary{summary},
		WorkHistory: workList,
		Skills:      []domain.Skill{goSkill},
		Academics:   []domain.AcademicCredential{academic},
		Certs:       []domain.Certification{cert},
		Descriptors: []domain.RoleDescriptor{desc},
	})
	require.NoError(t, err)

	// Verify PDF was created.
	info, err := os.Stat(pdfPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	// Record the export.
	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   pdfPath,
		SummaryID:  &summary.ID,
	})
	require.NoError(t, err)
	assert.NotZero(t, export.ID)
}

func TestWorkflow_CreateApplication(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Create an export record (simulate having generated a PDF).
	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
	})
	require.NoError(t, err)

	coverTemplateID := int64(3) // built-in Formal template

	// Create application linked to resume export and cover letter template.
	app, err := store.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:               "Acme Corp",
		PositionTitle:             "Senior Software Engineer",
		DateApplied:               "2025-01-15",
		FitIndicator:              "strong",
		ResumeExportID:            &export.ID,
		CoverLetterTemplateID:     &coverTemplateID,
		CoverLetterLatestExportID: &export.ID,
		Notes:                     "Referred by John",
	})
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", app.CompanyName)
	assert.Equal(t, domain.StatusApplied, app.Status)
	assert.Equal(t, &export.ID, app.ResumeExportID)
	assert.Equal(t, &coverTemplateID, app.CoverLetterTemplateID)
	assert.Equal(t, &export.ID, app.CoverLetterLatestExportID)
}

func TestWorkflow_JSONExportImportRoundtrip(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Seed comprehensive data.
	_, err := store.GetProfile(ctx) // ensure row exists
	require.NoError(t, err)
	_, err = store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Jane Smith",
		Email:    "jane@example.com",
		Phone:    "555-0100",
		Location: "New York, NY",
	})
	require.NoError(t, err)

	_, err = store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "GitHub",
		URL:   "https://github.com/janesmith",
	})
	require.NoError(t, err)

	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "General",
		BodyText: "Experienced engineer.",
	})
	require.NoError(t, err)

	wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		StartDate:            "2020-01",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, wh.ID, "Built awesome stuff", domain.BulletTypePrimary)
	require.NoError(t, err)

	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	_, err = store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	require.NoError(t, err)

	_, err = store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution:     "MIT",
		CredentialType:  "BS",
		FieldOfStudy:    "CS",
		CompletionDate:  "2017-05",
		DateGranularity: "month",
	})
	require.NoError(t, err)

	_, err = store.CreateCertification(ctx, domain.CertificationInput{
		Name:        "AWS SA",
		IssuingBody: "AWS",
		DateEarned:  "2022-03",
	})
	require.NoError(t, err)

	_, err = store.CreateDescriptor(ctx, "Software Engineer")
	require.NoError(t, err)

	lens, err := store.CreateLens(ctx, domain.LensInput{
		Name: "Backend",
	})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: summary.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	_, err = store.CreateCoverLetter(ctx, domain.CoverLetterInput{
		Title:    "Cover Letter",
		BodyText: "Dear Hiring Manager...",
	})
	require.NoError(t, err)

	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   "/tmp/resume.pdf",
		LensID:     &lens.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:    "Acme Corp",
		PositionTitle:  "Engineer",
		DateApplied:    "2025-01-15",
		FitIndicator:   "strong",
		ResumeExportID: &export.ID,
		Notes:          "Great opportunity",
	})
	require.NoError(t, err)

	// Export all data via the BackupService.
	backupSvc := service.NewBackupService(store)
	exportDir := t.TempDir()
	exportPath := filepath.Join(exportDir, "export.json")

	outputPath, err := backupSvc.ExportAllData(ctx, exportPath)
	require.NoError(t, err)

	// Read and verify the JSON file.
	jsonData, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	var exported domain.ExportData
	require.NoError(t, json.Unmarshal(jsonData, &exported))

	assert.Equal(t, 11, exported.SchemaVersion)
	assert.NotEmpty(t, exported.ExportedAt)
	assert.Equal(t, "Jane Smith", exported.Profile.FullName)
	assert.Len(t, exported.ProfileLinks, 1)
	assert.Len(t, exported.WorkHistory, 1)
	assert.Len(t, exported.WorkHistory[0].Bullets, 1)
	assert.Len(t, exported.Categories, 1)
	assert.Len(t, exported.Skills, 1)
	assert.Len(t, exported.Academics, 1)
	assert.Len(t, exported.Certs, 1)
	assert.Len(t, exported.Summaries, 1)
	assert.Len(t, exported.Descriptors, 1)
	assert.Len(t, exported.Lenses, 1)
	assert.Len(t, exported.Exports, 1)
	assert.Len(t, exported.CoverLetters, 1)
	assert.Len(t, exported.Applications, 1)

	// Import into a fresh store and verify roundtrip.
	store2 := testStore(t)

	err = backupSvc.ImportAllData(ctx, outputPath)
	require.NoError(t, err)

	// The import went into the original store (backupSvc references it).
	// For a true roundtrip, import into store2.
	backupSvc2 := service.NewBackupService(store2)
	err = backupSvc2.ImportAllData(ctx, outputPath)
	require.NoError(t, err)

	// Verify data in store2 matches.
	reimported, err := store2.ExportAllData(ctx)
	require.NoError(t, err)

	assert.Equal(t, exported.Profile.FullName, reimported.Profile.FullName)
	assert.Equal(t, exported.Profile.Email, reimported.Profile.Email)
	assert.Equal(t, exported.Profile.Phone, reimported.Profile.Phone)
	assert.Equal(t, exported.Profile.Location, reimported.Profile.Location)
	assert.Len(t, reimported.ProfileLinks, len(exported.ProfileLinks))
	assert.Len(t, reimported.WorkHistory, len(exported.WorkHistory))
	assert.Len(t, reimported.WorkHistory[0].Bullets, len(exported.WorkHistory[0].Bullets))
	assert.Len(t, reimported.Categories, len(exported.Categories))
	assert.Len(t, reimported.Skills, len(exported.Skills))
	assert.Len(t, reimported.Academics, len(exported.Academics))
	assert.Len(t, reimported.Certs, len(exported.Certs))
	assert.Len(t, reimported.Summaries, len(exported.Summaries))
	assert.Len(t, reimported.Descriptors, len(exported.Descriptors))
	assert.Len(t, reimported.Lenses, len(exported.Lenses))
	assert.Len(t, reimported.Exports, len(exported.Exports))
	assert.Len(t, reimported.CoverLetters, len(exported.CoverLetters))
	assert.Len(t, reimported.Applications, len(exported.Applications))

	// Verify specific field values survived the roundtrip.
	assert.Equal(t, exported.WorkHistory[0].EmployerName, reimported.WorkHistory[0].EmployerName)
	assert.Equal(t, exported.WorkHistory[0].Bullets[0].Text, reimported.WorkHistory[0].Bullets[0].Text)
	assert.Equal(t, exported.Skills[0].Name, reimported.Skills[0].Name)
	assert.Equal(t, exported.Skills[0].CompetenceLevel, reimported.Skills[0].CompetenceLevel)
	assert.Equal(t, exported.Academics[0].Institution, reimported.Academics[0].Institution)
	assert.Equal(t, exported.Certs[0].Name, reimported.Certs[0].Name)
	assert.Equal(t, exported.Summaries[0].BodyText, reimported.Summaries[0].BodyText)
	assert.Equal(t, exported.Descriptors[0].Title, reimported.Descriptors[0].Title)
	assert.Equal(t, exported.Lenses[0].Name, reimported.Lenses[0].Name)
	assert.Equal(t, exported.CoverLetters[0].Title, reimported.CoverLetters[0].Title)
	assert.Equal(t, exported.Applications[0].CompanyName, reimported.Applications[0].CompanyName)
	assert.Equal(t, exported.Applications[0].Status, reimported.Applications[0].Status)
}

func TestWorkflow_FullEndToEnd(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	renderer := pdf.NewRenderer()

	// Step 1: Create profile.
	_, err := store.GetProfile(ctx) // ensure row exists
	require.NoError(t, err)
	profile, err := store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Alice Johnson",
		Email:    "alice@example.com",
		Phone:    "555-0200",
		Location: "San Francisco, CA",
	})
	require.NoError(t, err)

	link, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "Portfolio",
		URL:   "https://alice.dev",
	})
	require.NoError(t, err)

	// Step 2: Add work history with bullets.
	wh1, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "BigTech Inc",
		JobTitle:             "Staff Engineer",
		StartDate:            "2021-03",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)

	b1, err := store.CreateBullet(ctx, wh1.ID, "Architected event-driven platform serving 10M daily events", domain.BulletTypePrimary)
	require.NoError(t, err)
	b2, err := store.CreateBullet(ctx, wh1.ID, "Reduced P99 latency from 500ms to 50ms", domain.BulletTypePrimary)
	require.NoError(t, err)

	wh2, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "StartupCo",
		JobTitle:             "Software Engineer",
		StartDate:            "2018-06",
		EndDate:              "2021-02",
		DateGranularityStart: "month",
		DateGranularityEnd:   "month",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, wh2.ID, "Built real-time analytics dashboard", domain.BulletTypePrimary)
	require.NoError(t, err)

	// Step 3: Add skills.
	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)

	goSkill, err := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 10,
	})
	require.NoError(t, err)

	pySkill, err := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Python", CategoryID: cat.ID, CompetenceLevel: 8,
	})
	require.NoError(t, err)

	// Step 4: Add education + certifications.
	academic, err := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution:     "Stanford University",
		CredentialType:  "Master of Science",
		FieldOfStudy:    "Computer Science",
		CompletionDate:  "2018-06",
		DateGranularity: "month",
	})
	require.NoError(t, err)

	cert, err := store.CreateCertification(ctx, domain.CertificationInput{
		Name:        "GCP Professional Cloud Architect",
		IssuingBody: "Google Cloud",
		DateEarned:  "2023-01",
	})
	require.NoError(t, err)

	// Step 5: Add summary and descriptors.
	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "Backend Focus",
		BodyText: "Staff-level engineer specializing in distributed systems and platform engineering.",
	})
	require.NoError(t, err)

	desc, err := store.CreateDescriptor(ctx, "Staff Software Engineer")
	require.NoError(t, err)

	// Step 6: Create a lens.
	lens, err := store.CreateLens(ctx, domain.LensInput{
		Name: "Platform Engineering",
	})
	require.NoError(t, err)

	err = store.SetLensSummaries(ctx, lens.ID, []domain.LensSummaryItem{
		{SummaryID: summary.ID, SortOrder: 0},
	})
	require.NoError(t, err)

	err = store.SetLensSkills(ctx, lens.ID, []domain.LensSkillItem{
		{SkillID: goSkill.ID},
		{SkillID: pySkill.ID},
	})
	require.NoError(t, err)

	// Step 7: Generate resume with all data.
	workList, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)

	outputDir := t.TempDir()
	profTmpl := pdf.ProfessionalTemplate()
	pdfPath, err := renderer.RenderResume(ctx, domain.RenderResumeRequest{
		Template:    &profTmpl,
		OutputDir:   outputDir,
		Profile:     profile,
		Links:       []domain.ProfileLink{link},
		Summaries:   []domain.ProfessionalSummary{summary},
		WorkHistory: workList,
		Skills:      []domain.Skill{goSkill},
		Academics:   []domain.AcademicCredential{academic},
		Certs:       []domain.Certification{cert},
		Descriptors: []domain.RoleDescriptor{desc},
	})
	require.NoError(t, err)

	info, err := os.Stat(pdfPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	// Record the export.
	export, err := store.CreateExport(ctx, domain.ResumeExport{
		TemplateID: "professional",
		FilePath:   pdfPath,
		SummaryID:  &summary.ID,
		LensID:     &lens.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateCoverLetter(ctx, domain.CoverLetterInput{
		Title:    "BigTech Cover Letter",
		BodyText: "Dear Hiring Manager, I am excited to apply...",
	})
	require.NoError(t, err)

	coverTemplateID := int64(3) // built-in Formal template

	app, err := store.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:               "BigTech Inc",
		PositionTitle:             "Staff Platform Engineer",
		DateApplied:               "2025-06-15",
		FitIndicator:              "strong",
		ResumeExportID:            &export.ID,
		CoverLetterTemplateID:     &coverTemplateID,
		CoverLetterLatestExportID: &export.ID,
		Notes:                     "Internal referral from Dave",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.StatusApplied, app.Status)

	// Step 9: Verify JSON export/import roundtrip.
	backupSvc := service.NewBackupService(store)
	exportPath := filepath.Join(t.TempDir(), "full-export.json")

	_, err = backupSvc.ExportAllData(ctx, exportPath)
	require.NoError(t, err)

	store2 := testStore(t)
	backupSvc2 := service.NewBackupService(store2)
	err = backupSvc2.ImportAllData(ctx, exportPath)
	require.NoError(t, err)

	// Verify everything roundtripped.
	reimported, err := store2.ExportAllData(ctx)
	require.NoError(t, err)

	assert.Equal(t, "Alice Johnson", reimported.Profile.FullName)
	assert.Len(t, reimported.ProfileLinks, 1)
	assert.Len(t, reimported.WorkHistory, 2)
	assert.Len(t, reimported.Skills, 2)
	assert.Len(t, reimported.Academics, 1)
	assert.Len(t, reimported.Certs, 1)
	assert.Len(t, reimported.Summaries, 1)
	assert.Len(t, reimported.Descriptors, 1)
	assert.Len(t, reimported.Lenses, 1)
	assert.Len(t, reimported.Exports, 1)
	assert.Len(t, reimported.CoverLetters, 1)
	assert.Len(t, reimported.Applications, 1)
	assert.Equal(t, domain.StatusApplied, reimported.Applications[0].Status)

	// Verify bullet counts survived.
	totalBullets := 0
	for _, w := range reimported.WorkHistory {
		totalBullets += len(w.Bullets)
	}
	assert.Equal(t, 3, totalBullets, "all 3 bullets should survive roundtrip")

	_ = b1
	_ = b2
}
