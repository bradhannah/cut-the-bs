package atscheck

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeText_ReportsMissingRequiredAndOrderErrors(t *testing.T) {
	text := "Header\nFirst section\nSecond section\n"
	report := AnalyzeText(text, Options{
		Required: []string{"Header", "Missing bit"},
		Ordered:  [][]string{{"First section", "Second section"}, {"Second section", "First section"}},
	})

	assert.False(t, report.Passed())
	assert.Equal(t, text, report.Text)

	errorBlob := strings.Join(report.Errors, "\n")
	assert.Contains(t, errorBlob, `missing required text: "Missing bit"`)
	assert.Contains(t, errorBlob, "ordered check 2 missing marker")
}

func TestAnalyzeText_ReportsPlaceholderAndSplitWordWarnings(t *testing.T) {
	text := "Dear {{hiring_manager}}, welcome to [company_name]. This has archi t ecture issues."
	report := AnalyzeText(text, Options{})

	assert.False(t, report.Passed())
	assert.NotEmpty(t, report.Errors)
	assert.NotEmpty(t, report.Warnings)

	errorBlob := strings.Join(report.Errors, "\n")
	assert.Contains(t, errorBlob, "unresolved template placeholders")
	assert.Contains(t, errorBlob, "unresolved prompt placeholders")
	assert.Contains(t, strings.Join(report.Warnings, "\n"), "split-word")
}

func TestAnalyzeText_PassesCleanContent(t *testing.T) {
	text := "Jane Smith\nExperience\nAcme Corp\n"
	report := AnalyzeText(text, Options{
		Required: []string{"Jane Smith", "Acme Corp"},
		Ordered:  [][]string{{"Jane Smith", "Experience", "Acme Corp"}},
	})

	assert.True(t, report.Passed())
	assert.Empty(t, report.Errors)
}
