package integration

import (
	"context"
	"os"
	"strings"
	"testing"

	"cut-the-bs/internal/domain"
	"cut-the-bs/internal/infra/pdf"

	ledongpdf "github.com/ledongthuc/pdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// extractPDFText reads the generated PDF file and extracts all text
// content, returning it as a single string. This simulates what an
// ATS (Applicant Tracking System) would see when parsing the file.
func extractPDFText(t *testing.T, pdfPath string) string {
	t.Helper()

	f, r, err := ledongpdf.Open(pdfPath)
	require.NoError(t, err, "failed to open PDF for text extraction")
	defer func() { _ = f.Close() }()

	var buf strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			// Some pages may not have extractable text; skip
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}
	return buf.String()
}

// fullATSTestData returns a fully-populated RenderResumeRequest
// covering every section, for ATS validation testing.
func fullATSTestData(outputDir string) domain.RenderResumeRequest {
	summaryID := int64(1)
	profTmpl := pdf.ProfessionalTemplate()
	return domain.RenderResumeRequest{
		Template:  &profTmpl,
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
				ID:       summaryID,
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
					{ID: 1, WorkHistoryID: 1, Text: "Led migration of monolithic application to microservices architecture reducing deployment time by 75 percent", SortOrder: 0},
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
					{ID: 4, WorkHistoryID: 2, Text: "Built RESTful API serving 50K daily active users with 99.9 percent uptime", SortOrder: 0},
					{ID: 5, WorkHistoryID: 2, Text: "Reduced infrastructure costs by 40 percent through containerization and auto-scaling", SortOrder: 1},
				},
			},
		},
		Skills: []domain.Skill{
			{ID: 1, Name: "Go", CategoryID: 1, CompetenceLevel: 9, IsLegacy: false},
			{ID: 2, Name: "Python", CategoryID: 1, CompetenceLevel: 8, IsLegacy: false},
			{ID: 3, Name: "PostgreSQL", CategoryID: 2, CompetenceLevel: 8, IsLegacy: false},
			{ID: 4, Name: "AWS", CategoryID: 3, CompetenceLevel: 7, IsLegacy: false},
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

func TestATS_PDFIsValidFile(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	// Verify the file exists and has content.
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0), "PDF file should not be empty")

	// Verify PDF magic bytes.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, len(data) >= 5, "PDF too small")
	assert.Equal(t, "%PDF-", string(data[:5]), "file should start with PDF magic bytes")
}

func TestATS_TextExtractionSucceeds(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)
	assert.NotEmpty(t, text, "should be able to extract text from the PDF")
}

func TestATS_ProfileInfoPresent(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	// Profile fields should be present in extracted text.
	assert.Contains(t, text, "Jane Smith", "full name should appear in PDF text")
	assert.Contains(t, text, "jane@example.com", "email should appear in PDF text")
	assert.Contains(t, text, "555-0100", "phone should appear in PDF text")
	assert.Contains(t, text, "New York", "location should appear in PDF text")
}

func TestATS_WorkHistoryPresent(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	// Employer names and job titles.
	assert.Contains(t, text, "Acme Corp", "employer name should appear")
	assert.Contains(t, text, "Senior Software Engineer", "job title should appear")
	assert.Contains(t, text, "TechStart", "second employer should appear")

	// Bullet text — check key phrases from each bullet.
	assert.Contains(t, text, "microservices", "bullet content should appear")
	assert.Contains(t, text, "Mentored", "bullet content should appear")
	assert.Contains(t, text, "data pipeline", "bullet content should appear")
	assert.Contains(t, text, "RESTful", "bullet content should appear")
	assert.Contains(t, text, "containerization", "bullet content should appear")
}

func TestATS_SkillsPresent(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	// Each non-legacy skill should appear.
	for _, skill := range req.Skills {
		if !skill.IsLegacy {
			assert.Contains(t, text, skill.Name,
				"skill %q should appear in PDF text", skill.Name)
		}
	}
}

func TestATS_EducationPresent(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	assert.Contains(t, text, "MIT", "institution should appear")
	assert.Contains(t, text, "Computer Science", "field of study should appear")
}

func TestATS_CertificationsPresent(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	assert.Contains(t, text, "AWS Solutions Architect", "cert name should appear")
	assert.Contains(t, text, "Amazon Web Services", "cert issuer should appear")
}

func TestATS_SummaryPresent(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	// Check that the summary text appears (at least key phrases).
	assert.Contains(t, text, "Experienced software engineer", "summary should appear")
	assert.Contains(t, text, "distributed systems", "summary content should appear")
}

func TestATS_DescriptorsPresent(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	for _, d := range req.Descriptors {
		assert.Contains(t, text, d.Title,
			"descriptor %q should appear in PDF text", d.Title)
	}
}

func TestATS_NoMidWordSpaces(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	// Check that multi-character words from the input are not
	// broken by spurious spaces in the extracted text. We test a
	// selection of words that span different sections.
	wordsToCheck := []string{
		"Jane",
		"Smith",
		"example",
		"Engineer",
		"microservices",
		"architecture",
		"containerization",
		"PostgreSQL",
		"Computer",
		"Science",
		"Architect",
		"Experienced",
	}

	for _, word := range wordsToCheck {
		assert.Contains(t, text, word,
			"word %q should appear without mid-word spaces", word)
	}
}

func TestATS_ReadingOrder(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	// Verify that content appears in logical reading order:
	// Profile info -> Summary -> Work History -> Skills -> Education/Certs
	//
	// We check that earlier sections appear before later sections
	// in the extracted text stream.
	nameIdx := strings.Index(text, "Jane Smith")
	summaryIdx := strings.Index(text, "Experienced software engineer")
	workIdx := strings.Index(text, "Acme Corp")

	require.NotEqual(t, -1, nameIdx, "name should be found")
	require.NotEqual(t, -1, summaryIdx, "summary should be found")
	require.NotEqual(t, -1, workIdx, "work history should be found")

	assert.Less(t, nameIdx, summaryIdx,
		"name should appear before summary in reading order")
	assert.Less(t, summaryIdx, workIdx,
		"summary should appear before work history in reading order")
}

func TestATS_BothTemplatesProduceValidPDF(t *testing.T) {
	renderer := pdf.NewRenderer()

	templates := map[string]domain.TemplateDetail{
		"professional": pdf.ProfessionalTemplate(),
		"modern":       pdf.ModernTemplate(),
	}
	for name, tmpl := range templates {
		t.Run(name, func(t *testing.T) {
			outputDir := t.TempDir()
			req := fullATSTestData(outputDir)
			req.Template = &tmpl

			path, err := renderer.RenderResume(context.Background(), req)
			require.NoError(t, err)

			text := extractPDFText(t, path)
			assert.NotEmpty(t, text, "template %q should produce extractable text", name)
			assert.Contains(t, text, "Jane Smith",
				"template %q should contain profile name", name)
			assert.Contains(t, text, "Acme Corp",
				"template %q should contain work history", name)
		})
	}
}

func TestATS_LinksPresent(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	req := fullATSTestData(outputDir)
	path, err := renderer.RenderResume(context.Background(), req)
	require.NoError(t, err)

	text := extractPDFText(t, path)

	// Links may appear as labels or URLs depending on template.
	// At minimum, the label or URL should be present.
	assert.True(t,
		strings.Contains(text, "LinkedIn") || strings.Contains(text, "linkedin.com"),
		"LinkedIn link (label or URL) should appear in PDF text")
	assert.True(t,
		strings.Contains(text, "GitHub") || strings.Contains(text, "github.com"),
		"GitHub link (label or URL) should appear in PDF text")
}
