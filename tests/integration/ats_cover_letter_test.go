package integration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"cut-the-bs/internal/atscheck"
	"cut-the-bs/internal/domain"
	"cut-the-bs/internal/infra/pdf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func coverLetterATSRequest(outputDir string, tmpl *domain.TemplateDetail) domain.RenderCoverLetterRequest {
	return domain.RenderCoverLetterRequest{
		Template:  tmpl,
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
			Title:    "Platform Role",
			BodyText: "Fallback body text for cover letter ATS tests.",
		},
		SubstitutionMap: map[string]string{
			"hiring_manager":       "Ms. Rivera",
			"signer_name":          "Jane Smith",
			"email":                "jane@example.com",
			"recipient_address":    "Acme Robotics\n123 Main St\nNew York, NY 10001",
			"position_title":       "Principal Platform Engineer",
			"company_name":         "Acme Robotics",
			"why_fit_formal":       "I have led platform modernization that cut deployment risk by 40 percent.",
			"why_company_formal":   "your mission to make robotics safer at scale aligns with how I build systems.",
			"why_team_casual":      "you pair bold product goals with practical engineering discipline",
			"value_example_casual": "I improved release reliability and reduced incidents by standardizing ownership",
			"impact_example":       "I scaled the deployment pipeline without increasing incident volume.",
		},
	}
}

func TestATS_CoverLetter_FormalTemplate(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()
	tmpl := pdf.FormalCoverLetterTemplate()
	req := coverLetterATSRequest(outputDir, &tmpl)

	path, err := renderer.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err)

	report, err := atscheck.AnalyzePDF(path, atscheck.Options{
		Required: []string{
			"Jane Smith",
			"jane@example.com",
			"Dear Ms. Rivera,",
			"I am writing to express my interest in the",
			"Principal Platform Engineer",
			"I am especially interested in Acme Robotics because",
			"Thank you for your time and consideration.",
			"Sincerely,",
		},
		Ordered: [][]string{{
			"Dear Ms. Rivera,",
			"I am writing to express my interest in the",
			"I am especially interested in Acme Robotics because",
			"Thank you for your time and consideration.",
			"Sincerely,",
		}},
	})
	require.NoError(t, err)
	assert.Empty(t, report.Errors, strings.Join(report.Errors, "\n"))
	assert.NotEmpty(t, strings.TrimSpace(report.Text))
}

func TestATS_CoverLetter_CasualTemplate_MultiParagraphOrder(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()
	tmpl := pdf.CasualCoverLetterTemplate()
	req := coverLetterATSRequest(outputDir, &tmpl)

	path, err := renderer.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err)

	report, err := atscheck.AnalyzePDF(path, atscheck.Options{
		Required: []string{
			"Hi Ms. Rivera,",
			"I would love to join Acme Robotics as a Principal Platform Engineer.",
			"A quick example of what I would bring to this role:",
			"If it is helpful, I can share more details and examples.",
			"Thanks,",
		},
		Ordered: [][]string{{
			"Hi Ms. Rivera,",
			"I would love to join Acme Robotics as a Principal Platform Engineer.",
			"A quick example of what I would bring to this role:",
			"If it is helpful, I can share more details and examples.",
			"Thanks,",
		}},
	})
	require.NoError(t, err)
	assert.Empty(t, report.Errors, strings.Join(report.Errors, "\n"))
	assert.NotEmpty(t, strings.TrimSpace(report.Text))
}

func TestATS_CoverLetter_CustomParagraphTemplate_AllParagraphsPresent(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	tmpl := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           7300,
			Name:         "ATS Paragraph Template",
			TemplateType: domain.TemplateTypeCoverLetter,
			MarginTop:    72.0,
			MarginBottom: 72.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		},
		Elements: []domain.TemplateElement{
			{ID: 1, TemplateID: 7300, ElementType: domain.ElementGreeting, SortOrder: 0,
				Config: mustJSONCoverLetterATS(pdf.GreetingConfig{
					Text: "Dear {{hiring_manager}},", FontSize: 10.0, SpaceAfter: 10.0,
				})},
			{ID: 2, TemplateID: 7300, ElementType: domain.ElementParagraph, SortOrder: 1,
				Config: mustJSONCoverLetterATS(pdf.ParagraphConfig{
					FontSize: 10.0, LineSpacing: 1.15, SpaceAfter: 10.0,
					Segments: []pdf.ParagraphSegmentConfig{
						{Type: "static", Text: "Paragraph ALPHA for "},
						{Type: "application", Token: "position_title"},
						{Type: "static", Text: "."},
					},
				})},
			{ID: 3, TemplateID: 7300, ElementType: domain.ElementParagraph, SortOrder: 2,
				Config: mustJSONCoverLetterATS(pdf.ParagraphConfig{
					FontSize: 10.0, LineSpacing: 1.15, SpaceAfter: 10.0,
					Segments: []pdf.ParagraphSegmentConfig{
						{Type: "static", Text: "Paragraph BETA impact: "},
						{Type: "adhoc", Key: "impact_example", Label: "Impact example", Required: true, Multiline: true},
					},
				})},
			{ID: 4, TemplateID: 7300, ElementType: domain.ElementParagraph, SortOrder: 3,
				Config: mustJSONCoverLetterATS(pdf.ParagraphConfig{
					FontSize: 10.0, LineSpacing: 1.15, SpaceAfter: 10.0,
					Segments: []pdf.ParagraphSegmentConfig{
						{Type: "static", Text: "Paragraph GAMMA contact: "},
						{Type: "profile", Token: "email"},
						{Type: "static", Text: "."},
					},
				})},
			{ID: 5, TemplateID: 7300, ElementType: domain.ElementClosing, SortOrder: 4,
				Config: mustJSONCoverLetterATS(pdf.ClosingConfig{
					Text: "Sincerely,", FontSize: 10.0, SpaceAfter: 20.0,
				})},
			{ID: 6, TemplateID: 7300, ElementType: domain.ElementStaticText, SortOrder: 5,
				Config: mustJSONCoverLetterATS(pdf.StaticTextConfig{
					Text: "{{signer_name}}", FontSize: 10.0, FontStyle: "regular", Alignment: "left",
				})},
		},
	}

	req := coverLetterATSRequest(outputDir, &tmpl)
	path, err := renderer.RenderCoverLetter(context.Background(), req)
	require.NoError(t, err)

	report, err := atscheck.AnalyzePDF(path, atscheck.Options{
		Required: []string{
			"Dear Ms. Rivera,",
			"Paragraph ALPHA for Principal Platform Engineer.",
			"Paragraph BETA impact:",
			"Paragraph GAMMA contact:",
			"jane@example.com",
			"Sincerely,",
		},
		Ordered: [][]string{{
			"Dear Ms. Rivera,",
			"Paragraph ALPHA",
			"Paragraph BETA",
			"Paragraph GAMMA",
			"Sincerely,",
		}},
	})
	require.NoError(t, err)
	assert.Empty(t, report.Errors, strings.Join(report.Errors, "\n"))
}

func TestATS_CoverLetter_StaticTextWrap_NoLeadingSpaceOnContinuation(t *testing.T) {
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	// Very narrow usable width to force wrap between "you can" and
	// "note it here" for regression coverage of continuation spacing.
	tmpl := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           7400,
			Name:         "ATS Static Wrap Spacing",
			TemplateType: domain.TemplateTypeCoverLetter,
			MarginTop:    72.0,
			MarginBottom: 72.0,
			MarginLeft:   280.0,
			MarginRight:  280.0,
		},
		Elements: []domain.TemplateElement{
			{ID: 1, TemplateID: 7400, ElementType: domain.ElementStaticText, SortOrder: 0,
				Config: mustJSONCoverLetterATS(pdf.StaticTextConfig{
					Text: "you can note it here", FontSize: 10.0,
					FontStyle: "regular", Alignment: "left",
				})},
		},
	}

	path, err := renderer.RenderCoverLetter(context.Background(), domain.RenderCoverLetterRequest{
		Template:  &tmpl,
		OutputDir: outputDir,
	})
	require.NoError(t, err)

	report, err := atscheck.AnalyzePDF(path, atscheck.Options{})
	require.NoError(t, err)
	assert.Empty(t, report.Errors, strings.Join(report.Errors, "\n"))

	normalized := strings.ReplaceAll(report.Text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	assert.Contains(t, normalized, "you can\n", "narrow template should wrap after 'you can'")
	assert.Contains(t, normalized, "\nnote it here", "continuation should start with 'note'")
	assert.NotContains(t, normalized, "\n note it here",
		"wrapped continuation should not gain a leading space")
}

func mustJSONCoverLetterATS(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("mustJSONCoverLetterATS: " + err.Error())
	}
	return string(b)
}
