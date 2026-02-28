package pdf

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"cut-the-bs/internal/domain"

	"github.com/signintech/gopdf"
)

// Page dimensions in points (1 inch = 72 points).
const (
	letterWidth  = 612.0
	letterHeight = 792.0
	marginLeft   = 72.0 // 1 inch
	marginRight  = 72.0
	marginTop    = 54.0 // 0.75 inch
	marginBottom = 54.0
	usableWidth  = letterWidth - marginLeft - marginRight // 468
)

// Font sizes.
const (
	fontSizeSection = 12.0
	fontSizeBody    = 10.0
)

// Spacing.
const (
	lineSpacing = 3.0 // extra space between lines
)

// Renderer implements domain.PDFRenderer using signintech/gopdf.
type Renderer struct{}

// NewRenderer creates a new PDF renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// ListTemplates returns the available built-in templates.
func (r *Renderer) ListTemplates() []domain.ResumeTemplate {
	return []domain.ResumeTemplate{
		{
			ID:          "professional",
			Name:        "Professional",
			Description: "Clean single-column layout with traditional formatting. ATS-optimized with clear section headings.",
			PreviewURL:  "",
		},
		{
			ID:          "modern",
			Name:        "Modern",
			Description: "Contemporary layout with subtle visual hierarchy and modern typography spacing.",
			PreviewURL:  "",
		},
	}
}

// RenderResume generates a PDF resume from the given data using the
// template embedded in req.Template. The template provides page
// margins and an ordered list of elements dispatched through the
// element rendering pipeline. Returns the file path of the generated PDF.
func (r *Renderer) RenderResume(
	_ context.Context,
	req domain.RenderResumeRequest,
) (string, error) {
	if req.OutputDir == "" {
		return "", fmt.Errorf("output directory is required")
	}

	if req.Template == nil {
		return "", fmt.Errorf("template is required")
	}

	tmpl := *req.Template

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeLetter,
	})

	// Register fonts.
	if err := registerFonts(pdf); err != nil {
		return "", fmt.Errorf("register fonts: %w", err)
	}

	pdf.AddPage()

	rc := newRenderContext(pdf, req, tmpl)
	if err := renderElements(rc); err != nil {
		return "", fmt.Errorf("render template %q: %w", tmpl.Name, err)
	}

	// Generate filename with timestamp.
	ts := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("resume-%s-%s.pdf", strings.ToLower(tmpl.Name), ts)
	outPath := filepath.Join(req.OutputDir, filename)

	if err := pdf.WritePdf(outPath); err != nil {
		return "", fmt.Errorf("write PDF: %w", err)
	}

	return outPath, nil
}

// RenderCoverLetter generates a PDF cover letter. When req.Template
// is provided, the template-driven element pipeline is used. When
// Template is nil, falls back to the hardcoded rendering path for
// backward compatibility.
func (r *Renderer) RenderCoverLetter(
	_ context.Context,
	req domain.RenderCoverLetterRequest,
) (string, error) {
	if req.OutputDir == "" {
		return "", fmt.Errorf("output directory is required")
	}

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeLetter,
	})

	if err := registerFonts(pdf); err != nil {
		return "", fmt.Errorf("register fonts: %w", err)
	}

	pdf.AddPage()

	if req.Template != nil {
		// Template-driven path.
		rc := newCoverLetterRenderContext(pdf, req, *req.Template)
		if err := renderElements(rc); err != nil {
			return "", fmt.Errorf("render cover letter template %q: %w", req.Template.Name, err)
		}
	} else {
		// Hardcoded fallback path.
		pdf.SetY(marginTop)

		cfg := DefaultHeaderConfig()
		y, err := RenderProfileHeader(pdf, req.Profile, req.Links, cfg)
		if err != nil {
			return "", fmt.Errorf("render header: %w", err)
		}

		if err := setFont(pdf, "LiberationSans-Regular", fontSizeBody); err != nil {
			return "", err
		}

		y += 6
		y, err = renderWrappedText(pdf, req.Letter.BodyText, marginLeft, y, usableWidth, fontSizeBody, marginBottom, marginTop)
		if err != nil {
			return "", fmt.Errorf("render body: %w", err)
		}
		_ = y
	}

	ts := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("cover-letter-%s.pdf", ts)
	outPath := filepath.Join(req.OutputDir, filename)

	if err := pdf.WritePdf(outPath); err != nil {
		return "", fmt.Errorf("write PDF: %w", err)
	}

	return outPath, nil
}

// registerFonts loads all embedded Liberation Sans font variants.
func registerFonts(pdf *gopdf.GoPdf) error {
	fonts := []struct {
		name string
		data []byte
	}{
		{"LiberationSans-Regular", FontRegular},
		{"LiberationSans-Bold", FontBold},
		{"LiberationSans-Italic", FontItalic},
		{"LiberationSans-BoldItalic", FontBoldItalic},
	}

	for _, f := range fonts {
		if err := pdf.AddTTFFontData(f.name, f.data); err != nil {
			return fmt.Errorf("load font %s: %w", f.name, err)
		}
	}
	return nil
}

// setFont is a helper to set the current font.
func setFont(pdf *gopdf.GoPdf, name string, size float64) error {
	return pdf.SetFont(name, "", int(size))
}

// renderWrappedText renders multi-line text within maxWidth,
// wrapping at word boundaries. Returns the Y position after
// the last line.
func renderWrappedText(
	pdf *gopdf.GoPdf,
	text string,
	x, y, maxWidth, fontSize float64,
	mBottom, mTop float64,
) (float64, error) {
	lineHeight := fontSize + lineSpacing

	// Split into paragraphs on newlines.
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			y += lineHeight
			continue
		}

		words := strings.Fields(para)
		currentLine := ""

		for _, word := range words {
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			width, err := pdf.MeasureTextWidth(testLine)
			if err != nil {
				return y, fmt.Errorf("measure text: %w", err)
			}

			if width > maxWidth && currentLine != "" {
				// Emit current line.
				if err := checkPageBreak(pdf, &y, lineHeight, mBottom, mTop); err != nil {
					return y, err
				}
				pdf.SetX(x)
				pdf.SetY(y)
				if err := pdf.Cell(nil, currentLine); err != nil {
					return y, fmt.Errorf("render text: %w", err)
				}
				y += lineHeight
				currentLine = word
			} else {
				currentLine = testLine
			}
		}

		// Emit remaining text.
		if currentLine != "" {
			if err := checkPageBreak(pdf, &y, lineHeight, mBottom, mTop); err != nil {
				return y, err
			}
			pdf.SetX(x)
			pdf.SetY(y)
			if err := pdf.Cell(nil, currentLine); err != nil {
				return y, fmt.Errorf("render text: %w", err)
			}
			y += lineHeight
		}
	}

	return y, nil
}

// renderWrappedTextPreserveWhitespace renders text like renderWrappedText,
// but preserves leading/trailing spaces for each line instead of trimming
// them away.
func renderWrappedTextPreserveWhitespace(
	pdf *gopdf.GoPdf,
	text string,
	x, y, maxWidth, fontSize float64,
	mBottom, mTop float64,
) (float64, error) {
	lineHeight := fontSize + lineSpacing

	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.ReplaceAll(normalized, "\t", " ")

	lines := strings.Split(normalized, "\n")

	for _, line := range lines {
		if line == "" {
			y += lineHeight
			continue
		}

		tokens := splitLineTokensPreserveSpaces(line)
		currentLine := ""

		for _, token := range tokens {
			testLine := currentLine + token

			width, err := pdf.MeasureTextWidth(testLine)
			if err != nil {
				return y, fmt.Errorf("measure text: %w", err)
			}

			if width > maxWidth && currentLine != "" {
				if err := checkPageBreak(pdf, &y, lineHeight, mBottom, mTop); err != nil {
					return y, err
				}
				pdf.SetX(x)
				pdf.SetY(y)
				if err := pdf.Cell(nil, currentLine); err != nil {
					return y, fmt.Errorf("render text: %w", err)
				}
				y += lineHeight
				currentLine = token
			} else {
				currentLine = testLine
			}
		}

		if currentLine != "" {
			if err := checkPageBreak(pdf, &y, lineHeight, mBottom, mTop); err != nil {
				return y, err
			}
			pdf.SetX(x)
			pdf.SetY(y)
			if err := pdf.Cell(nil, currentLine); err != nil {
				return y, fmt.Errorf("render text: %w", err)
			}
			y += lineHeight
		}
	}

	return y, nil
}

func splitLineTokensPreserveSpaces(line string) []string {
	if line == "" {
		return nil
	}

	tokens := make([]string, 0)
	lineRunes := []rune(line)
	start := 0
	inSpace := unicode.IsSpace(lineRunes[0])

	for i := 1; i < len(lineRunes); i++ {
		isSpace := unicode.IsSpace(lineRunes[i])
		if isSpace == inSpace {
			continue
		}
		tokens = append(tokens, string(lineRunes[start:i]))
		start = i
		inSpace = isSpace
	}

	tokens = append(tokens, string(lineRunes[start:]))
	return tokens
}

// renderWrappedTextHanging renders text starting at x+firstLineIndent for
// the first line, then wraps continuation lines back to x at the full
// maxWidth (paragraph / "hanging first-line" style). This is used for
// labelled entries like "Category: skill1, skill2, skill3" where only
// the first line is offset by the label width.
func renderWrappedTextHanging(
	pdf *gopdf.GoPdf,
	text string,
	x, firstLineIndent, y, maxWidth, fontSize float64,
	mBottom, mTop float64,
) (float64, error) {
	lineHeight := fontSize + lineSpacing

	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return y, nil
	}

	// First line uses the reduced width after the label.
	curX := x + firstLineIndent
	curWidth := maxWidth - firstLineIndent
	currentLine := ""
	firstLine := true

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		width, err := pdf.MeasureTextWidth(testLine)
		if err != nil {
			return y, fmt.Errorf("measure text: %w", err)
		}

		if width > curWidth && currentLine != "" {
			// Emit the current line.
			if err := checkPageBreak(pdf, &y, lineHeight, mBottom, mTop); err != nil {
				return y, err
			}
			pdf.SetX(curX)
			pdf.SetY(y)
			if err := pdf.Cell(nil, currentLine); err != nil {
				return y, fmt.Errorf("render text: %w", err)
			}
			y += lineHeight
			currentLine = word

			// After the first line is emitted, switch to full-width left margin.
			if firstLine {
				firstLine = false
				curX = x
				curWidth = maxWidth
			}
		} else {
			currentLine = testLine
		}
	}

	// Emit any remaining text.
	if currentLine != "" {
		if err := checkPageBreak(pdf, &y, lineHeight, mBottom, mTop); err != nil {
			return y, err
		}
		pdf.SetX(curX)
		pdf.SetY(y)
		if err := pdf.Cell(nil, currentLine); err != nil {
			return y, fmt.Errorf("render text: %w", err)
		}
		y += lineHeight
	}

	return y, nil
}

// checkPageBreak adds a new page if the current Y position would
// exceed the printable area.
func checkPageBreak(pdf *gopdf.GoPdf, y *float64, needed, mBottom, mTop float64) error {
	if *y+needed > letterHeight-mBottom {
		pdf.AddPage()
		*y = mTop
	}
	return nil
}

// formatDateRange formats start/end dates for display on a resume.
func formatDateRange(start, end, granStart, granEnd string) string {
	s := formatSingleDate(start, granStart)
	if end == "" {
		return s + " - Present"
	}
	e := formatSingleDate(end, granEnd)
	return s + " - " + e
}

// formatSingleDate formats a date string based on granularity.
func formatSingleDate(date, granularity string) string {
	if date == "" {
		return ""
	}

	switch granularity {
	case "year":
		// YYYY
		if len(date) >= 4 {
			return date[:4]
		}
		return date
	case "month":
		// YYYY-MM → Mon YYYY
		t, err := time.Parse("2006-01", date)
		if err == nil {
			return t.Format("Jan 2006")
		}
		// Try full date.
		t, err = time.Parse("2006-01-02", date)
		if err == nil {
			return t.Format("Jan 2006")
		}
		return date
	case "day":
		t, err := time.Parse("2006-01-02", date)
		if err == nil {
			return t.Format("Jan 2, 2006")
		}
		return date
	default:
		// Default: try month.
		t, err := time.Parse("2006-01", date)
		if err == nil {
			return t.Format("Jan 2006")
		}
		t, err = time.Parse("2006-01-02", date)
		if err == nil {
			return t.Format("Jan 2006")
		}
		return date
	}
}
