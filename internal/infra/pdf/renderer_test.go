package pdf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/signintech/gopdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fullTestData returns a complete RenderResumeRequest with all
// sections populated, for testing the full rendering path.
func fullTestData(outputDir string) domain.RenderResumeRequest {
	tmpl := ProfessionalTemplate()
	return domain.RenderResumeRequest{
		Template:  &tmpl,
		OutputDir: outputDir,
		Profile: domain.UserProfile{
			ID:       1,
			FullName: "Jane Smith",
			Email:    "jane@example.com",
			Phone:    "555-0100",
			Location: "New York, NY",
		},
		Links: []domain.ProfileLink{
			{ID: 1, Label: "LinkedIn", URL: "https://linkedin.com/in/janesmith", SortOrder: 0},
			{ID: 2, Label: "GitHub", URL: "https://github.com/janesmith", SortOrder: 1},
		},
		Summaries: []domain.ProfessionalSummary{
			{
				ID:       1,
				Label:    "General",
				BodyText: "Experienced software engineer with 10 years of expertise in building scalable distributed systems and leading cross-functional teams.",
			},
		},
		WorkHistory: []domain.WorkHistoryEntry{
			{
				ID:                   1,
				EmployerName:         "Acme Corp",
				JobTitle:             "Senior Software Engineer",
				StartDate:            "2020-01",
				EndDate:              "",
				DateGranularityStart: "month",
				DateGranularityEnd:   "",
				SortOrder:            0,
				Bullets: []domain.AchievementBullet{
					{ID: 1, WorkHistoryID: 1, Text: "Led migration of monolithic application to microservices architecture, reducing deployment time by 75%", SortOrder: 0},
					{ID: 2, WorkHistoryID: 1, Text: "Mentored team of 5 junior engineers on best practices and code review standards", SortOrder: 1},
					{ID: 3, WorkHistoryID: 1, Text: "Designed and implemented real-time data pipeline processing 1M events per day", SortOrder: 2},
				},
			},
			{
				ID:                   2,
				EmployerName:         "TechStart Inc",
				JobTitle:             "Software Engineer",
				StartDate:            "2017-06",
				EndDate:              "2019-12",
				DateGranularityStart: "month",
				DateGranularityEnd:   "month",
				SortOrder:            1,
				Bullets: []domain.AchievementBullet{
					{ID: 4, WorkHistoryID: 2, Text: "Built RESTful API serving 50K daily active users with 99.9% uptime", SortOrder: 0},
					{ID: 5, WorkHistoryID: 2, Text: "Reduced infrastructure costs by 40% through containerization and auto-scaling", SortOrder: 1},
				},
			},
		},
		Skills: []domain.Skill{
			{ID: 1, Name: "Go", CategoryID: 1, CompetenceLevel: 9, IsLegacy: false},
			{ID: 2, Name: "Python", CategoryID: 1, CompetenceLevel: 8, IsLegacy: false},
			{ID: 3, Name: "PostgreSQL", CategoryID: 2, CompetenceLevel: 8, IsLegacy: false},
			{ID: 4, Name: "AWS", CategoryID: 3, CompetenceLevel: 7, IsLegacy: false},
			{ID: 5, Name: "SunOS", CategoryID: 3, CompetenceLevel: 6, IsLegacy: true},
		},
		Academics: []domain.AcademicCredential{
			{
				ID:              1,
				Institution:     "MIT",
				CredentialType:  "Bachelor of Science",
				FieldOfStudy:    "Computer Science",
				CompletionDate:  "2017-05",
				DateGranularity: "month",
				SortOrder:       0,
			},
		},
		Certs: []domain.Certification{
			{
				ID:             1,
				Name:           "AWS Solutions Architect",
				IssuingBody:    "Amazon Web Services",
				DateEarned:     "2022-03",
				ExpirationDate: "2025-03",
				IsActive:       true,
				SortOrder:      0,
			},
		},
		Descriptors: []domain.RoleDescriptor{
			{ID: 1, Title: "Software Engineer", SortOrder: 0},
			{ID: 2, Title: "Technical Lead", SortOrder: 1},
		},
	}
}

// minimalTestData returns a RenderResumeRequest with only the
// minimum required data: profile name and email, one work entry.
func minimalTestData(outputDir string) domain.RenderResumeRequest {
	tmpl := ProfessionalTemplate()
	return domain.RenderResumeRequest{
		Template:  &tmpl,
		OutputDir: outputDir,
		Profile: domain.UserProfile{
			ID:       1,
			FullName: "John Doe",
			Email:    "john@example.com",
		},
		WorkHistory: []domain.WorkHistoryEntry{
			{
				ID:                   1,
				EmployerName:         "Company A",
				JobTitle:             "Developer",
				StartDate:            "2023-01",
				EndDate:              "",
				DateGranularityStart: "month",
				SortOrder:            0,
				Bullets: []domain.AchievementBullet{
					{ID: 1, WorkHistoryID: 1, Text: "Developed core features", SortOrder: 0},
				},
			},
		},
	}
}

func TestRenderer_RenderResume_FullData(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	path, err := r.RenderResume(context.Background(), fullTestData(dir))
	require.NoError(t, err, "RenderResume should not error with full data")
	require.NotEmpty(t, path, "file path should not be empty")

	// File should exist and be non-empty.
	info, err := os.Stat(path)
	require.NoError(t, err, "generated PDF file should exist")
	assert.Greater(t, info.Size(), int64(0), "PDF file should be non-empty")

	// File should be in the output directory.
	assert.True(t, strings.HasPrefix(path, dir),
		"PDF should be written to the output directory")

	// File should have .pdf extension.
	assert.Equal(t, ".pdf", filepath.Ext(path))
}

func TestRenderer_RenderResume_MinimalData(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	path, err := r.RenderResume(context.Background(), minimalTestData(dir))
	require.NoError(t, err, "RenderResume should not error with minimal data")

	info, err := os.Stat(path)
	require.NoError(t, err, "generated PDF file should exist")
	assert.Greater(t, info.Size(), int64(0), "PDF file should be non-empty")
}

func TestRenderer_RenderResume_ModernTemplate(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	modernTmpl := ModernTemplate()
	req.Template = &modernTmpl

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "RenderResume should not error with modern template")

	info, err := os.Stat(path)
	require.NoError(t, err, "generated PDF file should exist")
	assert.Greater(t, info.Size(), int64(0), "PDF file should be non-empty")
}

func TestRenderer_RenderResume_NilTemplate(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	req.Template = nil

	_, err := r.RenderResume(context.Background(), req)
	assert.Error(t, err, "should reject nil template")
	assert.Contains(t, err.Error(), "template is required")
}

func TestRenderer_RenderResume_EmptyOutputDir(t *testing.T) {
	r := NewRenderer()

	req := fullTestData("")
	_, err := r.RenderResume(context.Background(), req)
	assert.Error(t, err, "should error when output dir is empty")
}

func TestRenderer_RenderResume_NoWorkHistory(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	req.WorkHistory = nil

	// Should still generate — the service layer decides whether
	// to reject empty content. The renderer renders what it gets.
	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "renderer should handle empty work history")

	info, err := os.Stat(path)
	require.NoError(t, err, "generated PDF file should exist")
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderResume_LegacySkillsIncluded(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	// Verify legacy skills are present in test data.
	hasLegacy := false
	for _, s := range req.Skills {
		if s.IsLegacy {
			hasLegacy = true
			break
		}
	}
	require.True(t, hasLegacy, "test data should include legacy skills")

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "should render with legacy skills present")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderResume_NoSummary(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	req.Summaries = nil

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "should render without summary")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderResume_NoDescriptors(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	req.Descriptors = nil

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "should render without descriptors")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderResume_NoCertifications(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	req.Certs = nil

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "should render without certifications")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderResume_NoAcademics(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	req.Academics = nil

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "should render without academics")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderResume_NoSkills(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	req.Skills = nil

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "should render without skills")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderResume_PDFStartsWithMagic(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	path, err := r.RenderResume(context.Background(), fullTestData(dir))
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.True(t, len(data) >= 5, "PDF too small")

	// PDF files always start with %PDF-
	assert.Equal(t, "%PDF-", string(data[:5]),
		"generated file should be valid PDF (starts with %%PDF-)")
}

func TestRenderer_RenderResume_BothTemplatesProduceDifferentFiles(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req1 := fullTestData(dir)
	profTmpl := ProfessionalTemplate()
	req1.Template = &profTmpl

	// Use a separate subdir to avoid filename collision.
	dir2 := filepath.Join(dir, "modern")
	require.NoError(t, os.MkdirAll(dir2, 0o755))

	req2 := fullTestData(dir2)
	modernTmpl := ModernTemplate()
	req2.Template = &modernTmpl

	path1, err := r.RenderResume(context.Background(), req1)
	require.NoError(t, err)

	path2, err := r.RenderResume(context.Background(), req2)
	require.NoError(t, err)

	data1, err := os.ReadFile(path1)
	require.NoError(t, err)

	data2, err := os.ReadFile(path2)
	require.NoError(t, err)

	// Templates should produce different output.
	assert.NotEqual(t, data1, data2,
		"professional and modern templates should produce different PDFs")
}

func TestRenderer_RenderResume_LongBulletWraps(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := minimalTestData(dir)
	// Very long bullet text that should wrap without error.
	req.WorkHistory[0].Bullets[0].Text = strings.Repeat("This is a long achievement bullet that should wrap properly within the PDF template content area without overflowing or truncating. ", 5)

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "should handle long wrapping bullets")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderResume_PresentEndDate(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := minimalTestData(dir)
	req.WorkHistory[0].EndDate = "" // empty = present

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

// TestRenderer_ImplementsPDFRenderer verifies that Renderer
// satisfies the domain.PDFRenderer interface.
func TestRenderer_ImplementsPDFRenderer(t *testing.T) {
	var _ domain.PDFRenderer = (*Renderer)(nil)
}

// =================================================================
// RenderCoverLetter Tests (T075-T076)
// =================================================================

func coverLetterTestData(outputDir string) domain.RenderCoverLetterRequest {
	return domain.RenderCoverLetterRequest{
		OutputDir: outputDir,
		Profile: domain.UserProfile{
			ID:       1,
			FullName: "Jane Smith",
			Email:    "jane@example.com",
			Phone:    "555-0100",
			Location: "New York, NY",
		},
		Links: []domain.ProfileLink{
			{ID: 1, Label: "LinkedIn", URL: "https://linkedin.com/in/janesmith", SortOrder: 0},
		},
		Letter: domain.CoverLetter{
			ID:       1,
			Title:    "Software Engineer at Acme",
			BodyText: "Dear Hiring Manager,\n\nI am writing to express my interest in the Software Engineer position at Acme Corp. With over 10 years of experience in building scalable distributed systems, I am confident that I can make a significant contribution to your team.\n\nI look forward to hearing from you.\n\nSincerely,\nJane Smith",
		},
	}
}

func TestRenderer_RenderCoverLetter_Success(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	path, err := r.RenderCoverLetter(context.Background(), coverLetterTestData(dir))
	require.NoError(t, err, "RenderCoverLetter should not error")
	require.NotEmpty(t, path, "file path should not be empty")

	info, err := os.Stat(path)
	require.NoError(t, err, "generated PDF file should exist")
	assert.Greater(t, info.Size(), int64(0), "PDF file should be non-empty")
	assert.True(t, strings.HasPrefix(path, dir), "PDF should be in output dir")
	assert.Equal(t, ".pdf", filepath.Ext(path))
}

func TestRenderer_RenderCoverLetter_EmptyOutputDir(t *testing.T) {
	r := NewRenderer()

	req := coverLetterTestData("")
	_, err := r.RenderCoverLetter(context.Background(), req)
	assert.Error(t, err, "should error when output dir is empty")
}

func TestRenderer_RenderCoverLetter_PDFMagic(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	path, err := r.RenderCoverLetter(context.Background(), coverLetterTestData(dir))
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.True(t, len(data) >= 5, "PDF too small")
	assert.Equal(t, "%PDF-", string(data[:5]),
		"generated file should be valid PDF")
}

func TestRenderer_RenderCoverLetter_LongBody(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := coverLetterTestData(dir)
	req.Letter.BodyText = strings.Repeat("This is a paragraph of text that should wrap properly and potentially span multiple pages when the cover letter body is very long. ", 30)

	path, err := r.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err, "should handle long wrapping body text")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderCoverLetter_NoLinks(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := coverLetterTestData(dir)
	req.Links = nil

	path, err := r.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err, "should render without profile links")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderCoverLetter_FilenameFormat(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	path, err := r.RenderCoverLetter(context.Background(), coverLetterTestData(dir))
	require.NoError(t, err)

	filename := filepath.Base(path)
	assert.True(t, strings.HasPrefix(filename, "cover-letter-"),
		"filename should start with cover-letter-")
	assert.True(t, strings.HasSuffix(filename, ".pdf"),
		"filename should end with .pdf")
}

// TestRenderer_ListTemplates verifies that the renderer provides
// at least two built-in templates per FR-016.
func TestRenderer_ListTemplates(t *testing.T) {
	r := NewRenderer()
	templates := r.ListTemplates()

	require.GreaterOrEqual(t, len(templates), 2,
		"must provide at least 2 templates (FR-016)")

	ids := make(map[string]bool)
	for _, tmpl := range templates {
		assert.NotEmpty(t, tmpl.ID, "template ID should not be empty")
		assert.NotEmpty(t, tmpl.Name, "template Name should not be empty")
		assert.NotEmpty(t, tmpl.Description, "template Description should not be empty")
		assert.False(t, ids[tmpl.ID], "template IDs should be unique")
		ids[tmpl.ID] = true
	}
}

// =================================================================
// Fidelity Tests (T016, T017)
// =================================================================

// renderHardcoded renders a resume using the old hardcoded template
// function and returns the raw PDF bytes. It mirrors the RenderResume
// method logic but uses GetBytesPdfReturnErr instead of WritePdf.
func renderHardcoded(t *testing.T, tmplID string, req domain.RenderResumeRequest) []byte {
	t.Helper()

	r := NewRenderer()
	tmplFn, ok := r.templates[tmplID]
	require.True(t, ok, "template %q must exist", tmplID)

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeLetter,
	})
	require.NoError(t, registerFonts(pdf))
	pdf.AddPage()

	require.NoError(t, tmplFn(pdf, req))

	data, err := pdf.GetBytesPdfReturnErr()
	require.NoError(t, err)
	return data
}

// renderTemplated renders a resume using the new template-driven
// element pipeline and returns the raw PDF bytes.
func renderTemplated(t *testing.T, tmpl domain.TemplateDetail, req domain.RenderResumeRequest) []byte {
	t.Helper()

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeLetter,
	})
	require.NoError(t, registerFonts(pdf))
	pdf.AddPage()

	rc := newRenderContext(pdf, req, tmpl)
	require.NoError(t, renderElements(rc))

	data, err := pdf.GetBytesPdfReturnErr()
	require.NoError(t, err)
	return data
}

// fullTestDataWithCoreExpertise extends fullTestData with core
// expertise items and skill category names, exercising all sections.
func fullTestDataWithCoreExpertise(outputDir string) domain.RenderResumeRequest {
	req := fullTestData(outputDir)
	req.MasterSummaryID = int64Ptr(1)
	req.CoreExpertise = []domain.CoreExpertise{
		{ID: 1, Label: "Distributed Systems", SortOrder: 0},
		{ID: 2, Label: "API Design", SortOrder: 1},
		{ID: 3, Label: "Cloud Architecture", SortOrder: 2},
	}
	req.SkillCategoryNames = map[int64]string{
		1: "Languages",
		2: "Databases",
		3: "Infrastructure",
	}
	return req
}

// TestRenderer_Fidelity_Professional (T016) verifies that the
// template-driven pipeline produces byte-identical output to the
// hardcoded renderProfessional function.
func TestRenderer_Fidelity_Professional(t *testing.T) {
	req := fullTestDataWithCoreExpertise("")
	tmpl := ProfessionalTemplate()

	hardcoded := renderHardcoded(t, "professional", req)
	templated := renderTemplated(t, tmpl, req)

	if !assert.Equal(t, hardcoded, templated,
		"template-driven Professional output must be byte-identical to hardcoded") {
		// Write both files for manual inspection on failure.
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "hardcoded.pdf"), hardcoded, 0o644)
		os.WriteFile(filepath.Join(dir, "templated.pdf"), templated, 0o644)
		t.Logf("PDFs written to %s for inspection", dir)
		t.Logf("hardcoded size: %d bytes, templated size: %d bytes",
			len(hardcoded), len(templated))
	}
}

// TestRenderer_Fidelity_Modern (T017) verifies that the template-driven
// pipeline produces byte-identical output to the hardcoded renderModern
// function.
func TestRenderer_Fidelity_Modern(t *testing.T) {
	req := fullTestDataWithCoreExpertise("")
	tmpl := ModernTemplate()

	hardcoded := renderHardcoded(t, "modern", req)
	templated := renderTemplated(t, tmpl, req)

	if !assert.Equal(t, hardcoded, templated,
		"template-driven Modern output must be byte-identical to hardcoded") {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "hardcoded.pdf"), hardcoded, 0o644)
		os.WriteFile(filepath.Join(dir, "templated.pdf"), templated, 0o644)
		t.Logf("PDFs written to %s for inspection", dir)
		t.Logf("hardcoded size: %d bytes, templated size: %d bytes",
			len(hardcoded), len(templated))
	}
}

// TestRenderer_Fidelity_Professional_MinimalData tests that
// minimal data also produces identical output between pipelines.
func TestRenderer_Fidelity_Professional_MinimalData(t *testing.T) {
	req := minimalTestData("")
	tmpl := ProfessionalTemplate()

	hardcoded := renderHardcoded(t, "professional", req)
	templated := renderTemplated(t, tmpl, req)

	if !assert.Equal(t, hardcoded, templated,
		"minimal-data Professional output must be byte-identical") {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "hardcoded.pdf"), hardcoded, 0o644)
		os.WriteFile(filepath.Join(dir, "templated.pdf"), templated, 0o644)
		t.Logf("PDFs written to %s for inspection", dir)
		t.Logf("hardcoded size: %d bytes, templated size: %d bytes",
			len(hardcoded), len(templated))
	}
}

// TestRenderer_Fidelity_Modern_MinimalData tests minimal data
// through the Modern template pipeline.
func TestRenderer_Fidelity_Modern_MinimalData(t *testing.T) {
	req := minimalTestData("")
	tmpl := ModernTemplate()

	hardcoded := renderHardcoded(t, "modern", req)
	templated := renderTemplated(t, tmpl, req)

	if !assert.Equal(t, hardcoded, templated,
		"minimal-data Modern output must be byte-identical") {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "hardcoded.pdf"), hardcoded, 0o644)
		os.WriteFile(filepath.Join(dir, "templated.pdf"), templated, 0o644)
		t.Logf("PDFs written to %s for inspection", dir)
		t.Logf("hardcoded size: %d bytes, templated size: %d bytes",
			len(hardcoded), len(templated))
	}
}

// TestRenderer_Fidelity_Professional_WithSecondaryBullets tests
// the outcomes block rendering with secondary bullets.
func TestRenderer_Fidelity_Professional_WithSecondaryBullets(t *testing.T) {
	req := fullTestDataWithCoreExpertise("")
	// Add secondary (outcome) bullets to the first work entry.
	req.WorkHistory[0].Bullets = append(req.WorkHistory[0].Bullets,
		domain.AchievementBullet{
			ID:            100,
			WorkHistoryID: 1,
			Text:          "Achieved 99.99% uptime SLA across all production services",
			BulletType:    domain.BulletTypeSecondary,
			SortOrder:     3,
		},
		domain.AchievementBullet{
			ID:            101,
			WorkHistoryID: 1,
			Text:          "Reduced mean time to recovery from 4 hours to 15 minutes",
			BulletType:    domain.BulletTypeSecondary,
			SortOrder:     4,
		},
	)
	tmpl := ProfessionalTemplate()

	hardcoded := renderHardcoded(t, "professional", req)
	templated := renderTemplated(t, tmpl, req)

	assert.Equal(t, hardcoded, templated,
		"Professional with secondary bullets must be byte-identical")
}

// TestRenderer_Fidelity_Professional_WithWorkSummary tests
// the work entry summary rendering.
func TestRenderer_Fidelity_Professional_WithWorkSummary(t *testing.T) {
	req := fullTestDataWithCoreExpertise("")
	req.WorkHistory[0].Summary = "Led the platform engineering team responsible for core infrastructure and developer tooling."
	tmpl := ProfessionalTemplate()

	hardcoded := renderHardcoded(t, "professional", req)
	templated := renderTemplated(t, tmpl, req)

	assert.Equal(t, hardcoded, templated,
		"Professional with work summary must be byte-identical")
}
