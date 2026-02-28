package atscheck

import (
	"fmt"
	"regexp"
	"strings"

	ledongpdf "github.com/ledongthuc/pdf"
)

var (
	curlyPlaceholderPattern   = regexp.MustCompile(`\{\{[^{}]+\}\}`)
	bracketPlaceholderPattern = regexp.MustCompile(`\[[A-Za-z][A-Za-z0-9_:\-]*\]`)
	splitWordPattern          = regexp.MustCompile(`\b[A-Za-z]{2,}\s[A-Za-z]\s[A-Za-z]{2,}\b`)
)

// Options controls ATS checks for extracted PDF text.
type Options struct {
	Required []string
	Ordered  [][]string
}

// Report is the result of ATS compatibility checks.
type Report struct {
	PDFPath  string
	Text     string
	Errors   []string
	Warnings []string
}

// Passed reports whether the PDF passes all required checks.
func (r Report) Passed() bool {
	return len(r.Errors) == 0
}

// AnalyzePDF extracts plain text from a PDF and runs ATS checks.
func AnalyzePDF(pdfPath string, opts Options) (Report, error) {
	text, err := ExtractText(pdfPath)
	if err != nil {
		return Report{}, err
	}

	report := AnalyzeText(text, opts)
	report.PDFPath = pdfPath
	return report, nil
}

// ExtractText extracts plain text from all pages in a PDF file.
func ExtractText(pdfPath string) (string, error) {
	f, reader, err := ledongpdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("open PDF: %w", err)
	}
	defer func() { _ = f.Close() }()

	var buf strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, textErr := page.GetPlainText(nil)
		if textErr != nil {
			continue
		}

		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// AnalyzeText validates extracted PDF text using ATS-focused checks.
func AnalyzeText(text string, opts Options) Report {
	report := Report{Text: text}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		report.Errors = append(report.Errors, "no extractable text found in PDF")
		return report
	}

	for _, marker := range opts.Required {
		needle := strings.TrimSpace(marker)
		if needle == "" {
			continue
		}
		if !strings.Contains(text, needle) {
			report.Errors = append(report.Errors, fmt.Sprintf("missing required text: %q", needle))
		}
	}

	for i, sequence := range opts.Ordered {
		if len(sequence) < 2 {
			continue
		}

		searchStart := 0
		for _, marker := range sequence {
			needle := strings.TrimSpace(marker)
			if needle == "" {
				continue
			}

			idx := strings.Index(text[searchStart:], needle)
			if idx < 0 {
				report.Errors = append(report.Errors,
					fmt.Sprintf("ordered check %d missing marker after previous text: %q", i+1, needle),
				)
				break
			}

			searchStart += idx + len(needle)
		}
	}

	if unresolved := uniqueMatches(curlyPlaceholderPattern.FindAllString(text, -1)); len(unresolved) > 0 {
		report.Errors = append(report.Errors,
			fmt.Sprintf("unresolved template placeholders found: %s", strings.Join(unresolved, ", ")),
		)
	}

	if unresolved := uniqueMatches(bracketPlaceholderPattern.FindAllString(text, -1)); len(unresolved) > 0 {
		report.Errors = append(report.Errors,
			fmt.Sprintf("unresolved prompt placeholders found: %s", strings.Join(unresolved, ", ")),
		)
	}

	if suspicious := uniqueMatches(splitWordPattern.FindAllString(text, -1)); len(suspicious) > 0 {
		limit := len(suspicious)
		if limit > 6 {
			limit = 6
		}
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("possible split-word extraction artifacts: %s", strings.Join(suspicious[:limit], ", ")),
		)
	}

	return report
}

func uniqueMatches(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}

	return unique
}
