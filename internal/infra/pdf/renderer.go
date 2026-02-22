package pdf

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

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
	fontSizeName    = 18.0
	fontSizeSection = 12.0
	fontSizeBody    = 10.0
	fontSizeSmall   = 9.0
	fontSizeDescBar = 10.0
)

// Spacing.
const (
	lineSpacing    = 3.0  // extra space between lines
	sectionGap     = 10.0 // space before a section heading
	bulletIndent   = 12.0 // left indent for bullet text
	bulletSymWidth = 10.0 // width of the bullet character
)

// templateFunc is a function that renders a complete resume onto
// a prepared GoPdf instance. It returns the final Y position.
type templateFunc func(
	pdf *gopdf.GoPdf,
	req domain.RenderResumeRequest,
) error

// Renderer implements domain.PDFRenderer using signintech/gopdf.
type Renderer struct {
	templates map[string]templateFunc
}

// NewRenderer creates a new PDF renderer with built-in templates.
func NewRenderer() *Renderer {
	r := &Renderer{
		templates: make(map[string]templateFunc),
	}
	r.templates["professional"] = renderProfessional
	r.templates["modern"] = renderModern
	return r
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

// renderResumeToBytes is an internal helper that renders a resume
// template to a byte slice without writing to disk. Used by the
// hardcoded template path for backward compatibility in tests.
func (r *Renderer) renderResumeToBytes(
	tmplName string,
	req domain.RenderResumeRequest,
) ([]byte, error) {
	tmplFn, ok := r.templates[tmplName]
	if !ok {
		return nil, fmt.Errorf("unknown template: %q", tmplName)
	}

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeLetter,
	})

	if err := registerFonts(pdf); err != nil {
		return nil, fmt.Errorf("register fonts: %w", err)
	}

	pdf.AddPage()

	if err := tmplFn(pdf, req); err != nil {
		return nil, fmt.Errorf("render template %q: %w", tmplName, err)
	}

	return pdf.GetBytesPdfReturnErr()
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
		y, err = renderWrappedText(pdf, req.Letter.BodyText, marginLeft, y, usableWidth, fontSizeBody)
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
				if err := checkPageBreak(pdf, &y, lineHeight); err != nil {
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
			if err := checkPageBreak(pdf, &y, lineHeight); err != nil {
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

// renderWrappedTextHanging renders text starting at x+firstLineIndent for
// the first line, then wraps continuation lines back to x at the full
// maxWidth (paragraph / "hanging first-line" style). This is used for
// labelled entries like "Category: skill1, skill2, skill3" where only
// the first line is offset by the label width.
func renderWrappedTextHanging(
	pdf *gopdf.GoPdf,
	text string,
	x, firstLineIndent, y, maxWidth, fontSize float64,
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
			if err := checkPageBreak(pdf, &y, lineHeight); err != nil {
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
		if err := checkPageBreak(pdf, &y, lineHeight); err != nil {
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
func checkPageBreak(pdf *gopdf.GoPdf, y *float64, needed float64) error {
	if *y+needed > letterHeight-marginBottom {
		pdf.AddPage()
		*y = marginTop
	}
	return nil
}

// renderSectionHeading renders a bold section heading with an
// underline rule. Returns the Y position after the heading.
func renderSectionHeading(
	pdf *gopdf.GoPdf,
	title string,
	y float64,
	underline bool,
) (float64, error) {
	if err := checkPageBreak(pdf, &y, fontSizeSection+sectionGap+4); err != nil {
		return y, err
	}

	y += sectionGap

	if err := setFont(pdf, "LiberationSans-Bold", fontSizeSection); err != nil {
		return y, err
	}

	pdf.SetX(marginLeft)
	pdf.SetY(y)
	if err := pdf.Cell(nil, strings.ToUpper(title)); err != nil {
		return y, fmt.Errorf("render section heading: %w", err)
	}

	y += fontSizeSection + 2

	if underline {
		pdf.SetLineWidth(0.5)
		pdf.Line(marginLeft, y, marginLeft+usableWidth, y)
		y += 4
	}

	return y, nil
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
