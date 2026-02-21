package pdf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fullTestData returns a complete RenderResumeRequest with all
// sections populated, for testing the full rendering path.
func fullTestData(outputDir string) domain.RenderResumeRequest {
	return domain.RenderResumeRequest{
		TemplateID: "professional",
		OutputDir:  outputDir,
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
	return domain.RenderResumeRequest{
		TemplateID: "professional",
		OutputDir:  outputDir,
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
	req.TemplateID = "modern"

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "RenderResume should not error with modern template")

	info, err := os.Stat(path)
	require.NoError(t, err, "generated PDF file should exist")
	assert.Greater(t, info.Size(), int64(0), "PDF file should be non-empty")
}

func TestRenderer_RenderResume_UnknownTemplate(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := fullTestData(dir)
	req.TemplateID = "nonexistent"

	_, err := r.RenderResume(context.Background(), req)
	assert.Error(t, err, "should reject unknown template ID")
	assert.Contains(t, err.Error(), "nonexistent")
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
	req1.TemplateID = "professional"

	// Use a separate subdir to avoid filename collision.
	dir2 := filepath.Join(dir, "modern")
	require.NoError(t, os.MkdirAll(dir2, 0o755))

	req2 := fullTestData(dir2)
	req2.TemplateID = "modern"

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
