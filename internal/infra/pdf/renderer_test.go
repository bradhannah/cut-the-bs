package pdf

import (
	"context"
	"encoding/json"
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

// renderTemplated renders a resume using the template-driven
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

// TestRenderer_EmptyLoop_ProducesNoOutput (T036) verifies that when
// no data items exist for a loop type, the loop and its preceding
// section heading (via DataBinding) are both omitted from the output.
func TestRenderer_EmptyLoop_ProducesNoOutput(t *testing.T) {
	// Build a template with a heading+loop for work history,
	// but supply no work history data.
	tmpl := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           999,
			Name:         "Empty Loop Test",
			TemplateType: domain.TemplateTypeResume,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		},
		Elements: []domain.TemplateElement{
			{
				ID:          1,
				TemplateID:  999,
				ElementType: domain.ElementProfileHeader,
				Config:      mustJSONTest(ProfileHeaderConfig{NameFontSize: 18.0, DetailFontSize: 10.0, Alignment: "center", SpaceAfter: 6.0}),
				SortOrder:   0,
			},
			{
				ID:          2,
				TemplateID:  999,
				ElementType: domain.ElementSectionHeading,
				Config:      mustJSONTest(SectionHeadingConfig{Text: "Experience", FontSize: 12.0, FontStyle: "bold", Uppercase: true, Underline: true, UnderlineWeight: 0.5, SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "work_history"}),
				SortOrder:   1,
			},
			{
				ID:          3,
				TemplateID:  999,
				ElementType: domain.ElementWorkHistoryLoop,
				Config:      mustJSONTest(WorkHistoryLoopConfig{EntryGap: 4.0}),
				SortOrder:   2,
			},
			// work_title child of the loop.
			{
				ID:          4,
				TemplateID:  999,
				ParentID:    int64Ptr(3),
				ElementType: domain.ElementWorkTitle,
				Config:      mustJSONTest(WorkTitleConfig{FontSize: 10.0, FontStyle: "bold", IncludeEmployer: true, EmployerSeparator: " — ", EmployerFontStyle: "italic", SpaceAfter: 13.0}),
				SortOrder:   0,
			},
			// work_bullets child of the loop.
			{
				ID:          5,
				TemplateID:  999,
				ParentID:    int64Ptr(3),
				ElementType: domain.ElementWorkBullets,
				Config:      mustJSONTest(WorkBulletsConfig{FontSize: 10.0, FontStyle: "regular", BulletChar: "•", Indent: 12.0, BulletSymWidth: 10.0}),
				SortOrder:   1,
			},
			// Summary section that HAS data — should still render.
			{
				ID:          6,
				TemplateID:  999,
				ElementType: domain.ElementSectionHeading,
				Config:      mustJSONTest(SectionHeadingConfig{Text: "Summary", FontSize: 12.0, FontStyle: "bold", Uppercase: true, Underline: true, UnderlineWeight: 0.5, SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "summaries"}),
				SortOrder:   3,
			},
			{
				ID:          7,
				TemplateID:  999,
				ElementType: domain.ElementProfSummary,
				Config:      mustJSONTest(ProfSummaryConfig{FontSize: 10.0, BulletChar: "•"}),
				SortOrder:   4,
			},
		},
	}

	req := domain.RenderResumeRequest{
		Template: &tmpl,
		Profile: domain.UserProfile{
			FullName: "Empty Loop User",
			Email:    "empty@loop.com",
		},
		// No work history data — the loop + heading should be omitted.
		WorkHistory: nil,
		// Summary data exists — heading + content should render.
		Summaries: []domain.ProfessionalSummary{
			{ID: 1, Label: "General", BodyText: "A versatile engineer."},
		},
		MasterSummaryID: int64Ptr(1),
	}

	// Render with template-driven pipeline.
	pdfBytes := renderTemplated(t, tmpl, req)
	require.NotEmpty(t, pdfBytes, "PDF should be non-empty")

	// Write to temp file and extract text.
	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "empty_loop.pdf")
	require.NoError(t, os.WriteFile(pdfPath, pdfBytes, 0o644))

	// Read PDF text. The extractPDFText from integration tests isn't
	// available here, so we just verify the PDF was generated and check size.
	// The key assertion is that the PDF is non-empty (rendering didn't crash)
	// and is smaller than a version with work data, confirming the loop
	// was skipped. We also do a comparative test.

	// Render a comparison with work history data.
	reqWithWork := req
	reqWithWork.WorkHistory = []domain.WorkHistoryEntry{
		{
			ID: 1, EmployerName: "SomeCorp", JobTitle: "Developer",
			StartDate: "2020-01", DateGranularityStart: "month",
			Bullets: []domain.AchievementBullet{
				{ID: 1, WorkHistoryID: 1, Text: "Built amazing things", BulletType: domain.BulletTypePrimary, SortOrder: 0},
			},
		},
	}
	pdfWithWork := renderTemplated(t, tmpl, reqWithWork)

	// The empty-loop PDF should be smaller — the work history section
	// (heading + loop entries) is omitted entirely.
	assert.Less(t, len(pdfBytes), len(pdfWithWork),
		"PDF with no work history should be smaller than PDF with work data (loop omitted)")
}

// TestRenderer_EducationLoop_EntryGap (T038) verifies that a non-zero
// entry_gap in EducationLoopConfig produces a larger PDF than entry_gap=0,
// confirming the spacing is applied between entries.
func TestRenderer_EducationLoop_EntryGap(t *testing.T) {
	buildTemplate := func(entryGap float64) domain.TemplateDetail {
		return domain.TemplateDetail{
			DocumentTemplate: domain.DocumentTemplate{
				ID:           999,
				Name:         "Edu EntryGap Test",
				TemplateType: domain.TemplateTypeResume,
				MarginTop:    54.0,
				MarginBottom: 54.0,
				MarginLeft:   72.0,
				MarginRight:  72.0,
			},
			Elements: []domain.TemplateElement{
				{
					ID:          1,
					TemplateID:  999,
					ElementType: domain.ElementProfileHeader,
					Config:      mustJSONTest(ProfileHeaderConfig{NameFontSize: 18.0, DetailFontSize: 10.0, Alignment: "center", SpaceAfter: 6.0}),
					SortOrder:   0,
				},
				{
					ID:          2,
					TemplateID:  999,
					ElementType: domain.ElementSectionHeading,
					Config:      mustJSONTest(SectionHeadingConfig{Text: "Education", FontSize: 12.0, FontStyle: "bold", Uppercase: true, Underline: true, UnderlineWeight: 0.5, SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "academics"}),
					SortOrder:   1,
				},
				{
					ID:          3,
					TemplateID:  999,
					ElementType: domain.ElementEducationLoop,
					Config:      mustJSONTest(EducationLoopConfig{EntryGap: entryGap}),
					SortOrder:   2,
				},
				// education_loop children
				{
					ID:          4,
					TemplateID:  999,
					ParentID:    int64Ptr(3),
					ElementType: domain.ElementEduCredential,
					Config:      mustJSONTest(EduCredentialConfig{FontSize: 10.0, FontStyle: "bold"}),
					SortOrder:   0,
				},
				{
					ID:          5,
					TemplateID:  999,
					ParentID:    int64Ptr(3),
					ElementType: domain.ElementEduInstitution,
					Config:      mustJSONTest(EduInstitutionConfig{FontSize: 10.0, FontStyle: "regular"}),
					SortOrder:   1,
				},
				{
					ID:          6,
					TemplateID:  999,
					ParentID:    int64Ptr(3),
					ElementType: domain.ElementEduDate,
					Config:      mustJSONTest(EduDateConfig{FontSize: 9.0, Alignment: "right"}),
					SortOrder:   2,
				},
			},
		}
	}

	req := domain.RenderResumeRequest{
		Profile: domain.UserProfile{FullName: "Test User", Email: "test@edu.com"},
		Academics: []domain.AcademicCredential{
			{ID: 1, CredentialType: "BSc", FieldOfStudy: "Computer Science", Institution: "MIT", CompletionDate: "2018-05", DateGranularity: "month"},
			{ID: 2, CredentialType: "MSc", FieldOfStudy: "AI", Institution: "Stanford", CompletionDate: "2020-05", DateGranularity: "month"},
		},
	}

	tmplNoGap := buildTemplate(0)
	tmplWithGap := buildTemplate(20.0) // 20pt gap between entries

	pdfNoGap := renderTemplated(t, tmplNoGap, req)
	pdfWithGap := renderTemplated(t, tmplWithGap, req)

	require.NotEmpty(t, pdfNoGap)
	require.NotEmpty(t, pdfWithGap)

	// With a 20pt entry_gap, the PDF content stream should be larger
	// (the Y positions differ, producing different position commands).
	assert.NotEqual(t, pdfNoGap, pdfWithGap,
		"entry_gap=20 should produce different PDF output than entry_gap=0")
}

// TestRenderer_CertsLoop_EntryGap (T038) verifies that a non-zero
// entry_gap in CertsLoopConfig produces a different PDF than entry_gap=0.
func TestRenderer_CertsLoop_EntryGap(t *testing.T) {
	buildTemplate := func(entryGap float64) domain.TemplateDetail {
		return domain.TemplateDetail{
			DocumentTemplate: domain.DocumentTemplate{
				ID:           999,
				Name:         "Certs EntryGap Test",
				TemplateType: domain.TemplateTypeResume,
				MarginTop:    54.0,
				MarginBottom: 54.0,
				MarginLeft:   72.0,
				MarginRight:  72.0,
			},
			Elements: []domain.TemplateElement{
				{
					ID:          1,
					TemplateID:  999,
					ElementType: domain.ElementProfileHeader,
					Config:      mustJSONTest(ProfileHeaderConfig{NameFontSize: 18.0, DetailFontSize: 10.0, Alignment: "center", SpaceAfter: 6.0}),
					SortOrder:   0,
				},
				{
					ID:          2,
					TemplateID:  999,
					ElementType: domain.ElementSectionHeading,
					Config:      mustJSONTest(SectionHeadingConfig{Text: "Certifications", FontSize: 12.0, FontStyle: "bold", Uppercase: true, Underline: true, UnderlineWeight: 0.5, SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "certifications"}),
					SortOrder:   1,
				},
				{
					ID:          3,
					TemplateID:  999,
					ElementType: domain.ElementCertsLoop,
					Config:      mustJSONTest(CertsLoopConfig{EntryGap: entryGap}),
					SortOrder:   2,
				},
				// certifications_loop children
				{
					ID:          4,
					TemplateID:  999,
					ParentID:    int64Ptr(3),
					ElementType: domain.ElementCertName,
					Config:      mustJSONTest(CertNameConfig{FontSize: 10.0, FontStyle: "bold"}),
					SortOrder:   0,
				},
				{
					ID:          5,
					TemplateID:  999,
					ParentID:    int64Ptr(3),
					ElementType: domain.ElementCertDetail,
					Config:      mustJSONTest(CertDetailConfig{FontSize: 9.0, FontStyle: "regular"}),
					SortOrder:   1,
				},
			},
		}
	}

	req := domain.RenderResumeRequest{
		Profile: domain.UserProfile{FullName: "Test User", Email: "test@cert.com"},
		Certs: []domain.Certification{
			{ID: 1, Name: "AWS Solutions Architect", IssuingBody: "Amazon", DateEarned: "2021-03"},
			{ID: 2, Name: "CKA", IssuingBody: "CNCF", DateEarned: "2022-06"},
		},
	}

	tmplNoGap := buildTemplate(0)
	tmplWithGap := buildTemplate(15.0) // 15pt gap between entries

	pdfNoGap := renderTemplated(t, tmplNoGap, req)
	pdfWithGap := renderTemplated(t, tmplWithGap, req)

	require.NotEmpty(t, pdfNoGap)
	require.NotEmpty(t, pdfWithGap)

	assert.NotEqual(t, pdfNoGap, pdfWithGap,
		"entry_gap=15 should produce different PDF output than entry_gap=0")
}

// TestRenderer_WorkSummary_BackcompatEmptyConfig ensures legacy
// templates with an empty work_summary config still render entry
// summary text using fallback defaults.
func TestRenderer_WorkSummary_BackcompatEmptyConfig(t *testing.T) {
	buildTemplate := func(includeWorkSummary bool, workSummaryConfig string) domain.TemplateDetail {
		elements := []domain.TemplateElement{
			{
				ID:          1,
				TemplateID:  999,
				ElementType: domain.ElementProfileHeader,
				Config:      mustJSONTest(ProfileHeaderConfig{NameFontSize: 18.0, DetailFontSize: 10.0, Alignment: "center", SpaceAfter: 6.0}),
				SortOrder:   0,
			},
			{
				ID:          2,
				TemplateID:  999,
				ElementType: domain.ElementSectionHeading,
				Config:      mustJSONTest(SectionHeadingConfig{Text: "Experience", FontSize: 12.0, FontStyle: "bold", Uppercase: true, Underline: true, UnderlineWeight: 0.5, SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "work_history"}),
				SortOrder:   1,
			},
			{
				ID:          3,
				TemplateID:  999,
				ElementType: domain.ElementWorkHistoryLoop,
				Config:      mustJSONTest(WorkHistoryLoopConfig{EntryGap: 4.0}),
				SortOrder:   2,
			},
			{
				ID:          4,
				TemplateID:  999,
				ParentID:    int64Ptr(3),
				ElementType: domain.ElementWorkTitle,
				Config:      mustJSONTest(WorkTitleConfig{FontSize: 10.0, FontStyle: "bold", IncludeEmployer: true, EmployerSeparator: " — ", EmployerFontStyle: "italic", SpaceAfter: 13.0}),
				SortOrder:   0,
			},
			{
				ID:          5,
				TemplateID:  999,
				ParentID:    int64Ptr(3),
				ElementType: domain.ElementWorkDates,
				Config:      mustJSONTest(WorkDatesConfig{FontSize: 9.0, Alignment: "right"}),
				SortOrder:   1,
			},
		}

		if includeWorkSummary {
			elements = append(elements, domain.TemplateElement{
				ID:          6,
				TemplateID:  999,
				ParentID:    int64Ptr(3),
				ElementType: domain.ElementWorkSummary,
				Config:      workSummaryConfig,
				SortOrder:   2,
			})
		}

		return domain.TemplateDetail{
			DocumentTemplate: domain.DocumentTemplate{
				ID:           999,
				Name:         "Work Summary Backcompat Test",
				TemplateType: domain.TemplateTypeResume,
				MarginTop:    54.0,
				MarginBottom: 54.0,
				MarginLeft:   72.0,
				MarginRight:  72.0,
			},
			Elements: elements,
		}
	}

	req := domain.RenderResumeRequest{
		Profile: domain.UserProfile{FullName: "Test User", Email: "test@work.com"},
		WorkHistory: []domain.WorkHistoryEntry{
			{
				ID:                   1,
				EmployerName:         "Acme Corp",
				JobTitle:             "Engineer",
				Summary:              "Lead summary line that should render even when work_summary config is empty.",
				StartDate:            "2021-01",
				DateGranularityStart: "month",
			},
		},
	}

	tmplWithLegacyConfig := buildTemplate(true, `{}`)
	tmplWithoutSummary := buildTemplate(false, ``)

	pdfWithLegacyConfig := renderTemplated(t, tmplWithLegacyConfig, req)
	pdfWithoutSummary := renderTemplated(t, tmplWithoutSummary, req)

	require.NotEmpty(t, pdfWithLegacyConfig)
	require.NotEmpty(t, pdfWithoutSummary)
	assert.NotEqual(t, pdfWithoutSummary, pdfWithLegacyConfig,
		"legacy empty work_summary config should still render summary text")
	assert.Greater(t, len(pdfWithLegacyConfig), len(pdfWithoutSummary),
		"PDF should be larger when summary text is rendered")
}

// =================================================================
// Cover Letter Template-Driven Tests (T048)
// =================================================================

// coverLetterTemplateTestData returns a cover letter request with a
// template attached, used for testing the template-driven cover letter
// rendering path.
func coverLetterTemplateTestData(outputDir string) domain.RenderCoverLetterRequest {
	return domain.RenderCoverLetterRequest{
		Template:  coverLetterTestTemplate(),
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
			BodyText: "I am writing to express my interest in the Software Engineer position. With over 10 years of experience, I am confident I can contribute to your team.",
		},
		SubstitutionMap: map[string]string{
			"company_name": "Acme Corp",
			"position":     "Software Engineer",
			"prompt:Why are you interested in this role?": "I love building scalable systems.",
		},
	}
}

// coverLetterTestTemplate builds a minimal cover letter template for tests.
func coverLetterTestTemplate() *domain.TemplateDetail {
	tmpl := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           100,
			Name:         "Test Cover Letter",
			TemplateType: domain.TemplateTypeCoverLetter,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		},
		Elements: []domain.TemplateElement{
			{ID: 1, TemplateID: 100, ElementType: domain.ElementProfileHeader, SortOrder: 0,
				Config: mustJSONTest(ProfileHeaderConfig{
					NameFontSize: 18.0, DetailFontSize: 10.0,
					Alignment: "center", LinkSeparator: " | ",
					ShowLinks: true, SpaceAfter: 6.0,
				})},
			{ID: 2, TemplateID: 100, ElementType: domain.ElementDate, SortOrder: 1,
				Config: mustJSONTest(DateConfig{
					FontSize: 10.0, Format: "January 2, 2006",
					Alignment: "left", SpaceAfter: 12.0,
				})},
			{ID: 3, TemplateID: 100, ElementType: domain.ElementRecipientAddress, SortOrder: 2,
				Config: mustJSONTest(RecipientAddressConfig{
					FontSize: 10.0, SpaceAfter: 12.0,
				})},
			{ID: 4, TemplateID: 100, ElementType: domain.ElementGreeting, SortOrder: 3,
				Config: mustJSONTest(GreetingConfig{
					Text: "Dear Hiring Manager,", FontSize: 10.0,
					SpaceAfter: 10.0,
				})},
			{ID: 5, TemplateID: 100, ElementType: domain.ElementBodyText, SortOrder: 4,
				Config: mustJSONTest(BodyTextConfig{
					FontSize: 10.0, LineSpacing: 3.0, SpaceAfter: 10.0,
				})},
			{ID: 6, TemplateID: 100, ElementType: domain.ElementClosing, SortOrder: 5,
				Config: mustJSONTest(ClosingConfig{
					Text: "Sincerely,", FontSize: 10.0,
					SpaceAfter: 24.0,
				})},
			{ID: 7, TemplateID: 100, ElementType: domain.ElementStaticText, SortOrder: 6,
				Config: mustJSONTest(StaticTextConfig{
					Text: "Jane Smith", FontSize: 10.0,
					FontStyle: "bold", SpaceAfter: 0.0,
				})},
		},
	}
	return &tmpl
}

func TestRenderer_RenderCoverLetter_TemplateDriven(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := coverLetterTemplateTestData(dir)
	path, err := r.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err, "template-driven cover letter should not error")
	require.NotEmpty(t, path)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
	assert.Equal(t, ".pdf", filepath.Ext(path))
}

func TestRenderer_RenderCoverLetter_TemplateDriven_NoLinks(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	req := coverLetterTemplateTestData(dir)
	req.Links = nil

	path, err := r.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderCoverLetter_WithVariableSubstitution(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	tmpl := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           101,
			Name:         "Variable CL",
			TemplateType: domain.TemplateTypeCoverLetter,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		},
		Elements: []domain.TemplateElement{
			{ID: 1, TemplateID: 101, ElementType: domain.ElementProfileHeader, SortOrder: 0,
				Config: mustJSONTest(ProfileHeaderConfig{
					NameFontSize: 18.0, DetailFontSize: 10.0,
					Alignment: "center", SpaceAfter: 6.0,
				})},
			{ID: 2, TemplateID: 101, ElementType: domain.ElementGreeting, SortOrder: 1,
				Config: mustJSONTest(GreetingConfig{
					Text: "Dear {{hiring_manager}},", FontSize: 10.0,
					SpaceAfter: 10.0,
				})},
			{ID: 3, TemplateID: 101, ElementType: domain.ElementBodyText, SortOrder: 2,
				Config: mustJSONTest(BodyTextConfig{
					FontSize: 10.0, LineSpacing: 3.0, SpaceAfter: 10.0,
				})},
			{ID: 4, TemplateID: 101, ElementType: domain.ElementClosing, SortOrder: 3,
				Config: mustJSONTest(ClosingConfig{
					Text: "Sincerely,", FontSize: 10.0,
					SpaceAfter: 24.0,
				})},
		},
	}

	req := domain.RenderCoverLetterRequest{
		Template:  &tmpl,
		OutputDir: dir,
		Profile: domain.UserProfile{
			FullName: "Test User",
			Email:    "test@example.com",
		},
		Letter: domain.CoverLetter{
			ID:       1,
			Title:    "Test CL",
			BodyText: "I am excited about the {{position}} role at {{company_name}}.",
		},
		SubstitutionMap: map[string]string{
			"hiring_manager": "Ms. Johnson",
			"position":       "Software Engineer",
			"company_name":   "Acme Corp",
		},
	}

	path, err := r.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err, "should render cover letter with variable substitution")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderCoverLetter_FormalTemplate(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	tmpl := FormalCoverLetterTemplate()
	req := domain.RenderCoverLetterRequest{
		Template:  &tmpl,
		OutputDir: dir,
		Profile: domain.UserProfile{
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
			Title:    "SE at Acme",
			BodyText: "I am writing to express my interest in the Software Engineer position at Acme Corp.",
		},
		SubstitutionMap: map[string]string{
			"hiring_manager":    "Ms. Johnson",
			"signer_name":       "Jane Smith",
			"recipient_address": "Acme Corp\n123 Main St\nNew York, NY 10001",
		},
	}

	path, err := r.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err, "Formal cover letter template should render")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderCoverLetter_CasualTemplate(t *testing.T) {
	dir := t.TempDir()
	r := NewRenderer()

	tmpl := CasualCoverLetterTemplate()
	req := domain.RenderCoverLetterRequest{
		Template:  &tmpl,
		OutputDir: dir,
		Profile: domain.UserProfile{
			FullName: "Jane Smith",
			Email:    "jane@example.com",
		},
		Letter: domain.CoverLetter{
			ID:       1,
			Title:    "Quick note",
			BodyText: "Just wanted to reach out about the open position. I think I'd be a great fit!",
		},
		SubstitutionMap: map[string]string{
			"hiring_manager": "Team",
			"signer_name":    "Jane",
		},
	}

	path, err := r.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err, "Casual cover letter template should render")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRenderer_RenderCoverLetter_TemplateDriven_AllElements(t *testing.T) {
	// Test with all cover letter element types present.
	dir := t.TempDir()
	r := NewRenderer()

	req := coverLetterTemplateTestData(dir)
	path, err := r.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err)

	// The template has 7 elements. Generate a hardcoded version for
	// comparison to verify template-driven produces a valid PDF.
	hardcodedReq := coverLetterTestData(filepath.Join(dir, "hardcoded"))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "hardcoded"), 0o755))
	pathHardcoded, err := r.RenderCoverLetter(context.Background(), hardcodedReq)
	require.NoError(t, err)

	info1, err := os.Stat(path)
	require.NoError(t, err)
	info2, err := os.Stat(pathHardcoded)
	require.NoError(t, err)

	// Both should be non-empty PDFs. They won't be identical since
	// template-driven includes more elements (date, greeting, closing, etc.).
	assert.Greater(t, info1.Size(), int64(0))
	assert.Greater(t, info2.Size(), int64(0))
}

func TestBuildParagraphText_PreservesStaticSegmentEdges(t *testing.T) {
	cfg := ParagraphConfig{
		Segments: []ParagraphSegmentConfig{
			{Type: "static", Text: "  "},
			{Type: "application", Token: "company_name"},
			{Type: "static", Text: "  "},
		},
	}

	got := buildParagraphText(cfg, map[string]string{"company_name": "Acme"})
	assert.Equal(t, "  Acme  ", got)
}

func TestBuildParagraphText_PreservesVariableSegmentEdges(t *testing.T) {
	cfg := ParagraphConfig{
		Segments: []ParagraphSegmentConfig{
			{Type: "static", Text: "Start:"},
			{Type: "adhoc", Key: "why_fit"},
			{Type: "static", Text: ":End"},
		},
	}

	got := buildParagraphText(cfg, map[string]string{"why_fit": "  spaced answer  "})
	assert.Equal(t, "Start:  spaced answer  :End", got)
}

func TestSplitLineTokensPreserveSpaces(t *testing.T) {
	got := splitLineTokensPreserveSpaces("  hello  world ")
	assert.Equal(t, []string{"  ", "hello", "  ", "world", " "}, got)
}

func TestWrapLinePreserveWhitespace_DropsLeadingSpaceOnContinuation(t *testing.T) {
	measure := func(text string) (float64, error) {
		return float64(len([]rune(text))), nil
	}

	got, err := wrapLinePreserveWhitespace("you can note it here", 7, measure)
	require.NoError(t, err)
	assert.Equal(t, []string{"you can", "note it", "here"}, got)
}

func TestWrapLinePreserveWhitespace_PreservesExplicitEdges(t *testing.T) {
	measure := func(text string) (float64, error) {
		return float64(len([]rune(text))), nil
	}

	got, err := wrapLinePreserveWhitespace("  hello  world  ", 40, measure)
	require.NoError(t, err)
	assert.Equal(t, []string{"  hello  world  "}, got)
}

func TestWrapLinePreserveWhitespace_WrapKeepsLineEdges(t *testing.T) {
	measure := func(text string) (float64, error) {
		return float64(len([]rune(text))), nil
	}

	got, err := wrapLinePreserveWhitespace("  hello  world  ", 9, measure)
	require.NoError(t, err)
	assert.Equal(t, []string{"  hello", "world  "}, got)
}

// mustJSONTest is a test helper that marshals v to JSON string.
func mustJSONTest(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("mustJSONTest: " + err.Error())
	}
	return string(b)
}

func TestRenderer_RenderResume_UnknownElementTypeSkipped(t *testing.T) {
	// A template with an unknown element type should render without
	// error — the unknown element is silently skipped.
	dir := t.TempDir()
	r := NewRenderer()

	tmpl := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           999,
			Name:         "Template With Unknown Element",
			TemplateType: domain.TemplateTypeResume,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		},
		Elements: []domain.TemplateElement{
			{ID: 1, TemplateID: 999, ElementType: domain.ElementProfileHeader, SortOrder: 0,
				Config: mustJSONTest(ProfileHeaderConfig{
					NameFontSize: 18.0, DetailFontSize: 10.0,
					Alignment: "center", SpaceAfter: 6.0,
				})},
			{ID: 2, TemplateID: 999, ElementType: "totally_unknown_widget", SortOrder: 1,
				Config: `{}`},
			{ID: 3, TemplateID: 999, ElementType: domain.ElementSpacer, SortOrder: 2,
				Config: mustJSONTest(SpacerConfig{Height: 10.0})},
		},
	}

	req := domain.RenderResumeRequest{
		Template:  &tmpl,
		OutputDir: dir,
		Profile: domain.UserProfile{
			FullName: "Test User",
			Email:    "test@example.com",
		},
	}

	path, err := r.RenderResume(context.Background(), req)
	require.NoError(t, err, "unknown element type should be skipped, not cause an error")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}
