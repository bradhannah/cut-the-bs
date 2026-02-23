package pdf

import (
	"fmt"
	"strings"

	"cut-the-bs/internal/domain"

	"github.com/signintech/gopdf"
)

// HeaderConfig controls the visual layout of the profile header.
type HeaderConfig struct {
	// NameFontSize is the size of the name text (default: 18).
	NameFontSize float64
	// DetailFontSize is the size for contact details and links (default: 10).
	DetailFontSize float64
	// LinkSeparator is the string between profile links (default: " | ").
	LinkSeparator string
	// MarginLeft is the left margin in points.
	MarginLeft float64
	// PageWidth is the usable page width (between margins) in points.
	PageWidth float64
	// ShowLinks controls whether profile links are rendered (default: true).
	ShowLinks bool
	// ShowLinksInline appends links to the contact line when true,
	// rather than rendering them on a separate third line (default: false).
	ShowLinksInline bool
	// SpaceAfter is the extra space added below the header (default: 6).
	SpaceAfter float64
}

// DefaultHeaderConfig returns a sensible default header configuration
// for Letter-size pages with 1-inch margins.
func DefaultHeaderConfig() HeaderConfig {
	return HeaderConfig{
		NameFontSize:    18,
		DetailFontSize:  10,
		LinkSeparator:   " | ",
		MarginLeft:      72,  // 1 inch
		PageWidth:       468, // Letter width (612) minus 2x 1-inch margins
		ShowLinks:       true,
		ShowLinksInline: false,
		SpaceAfter:      6.0,
	}
}

// RenderProfileHeader renders the profile header at the current
// cursor position of the given GoPdf instance. It renders:
//  1. Full name (bold, large)
//  2. Contact line: email, phone, location (separated by LinkSeparator).
//     When ShowLinksInline is true, profile link URLs are appended to
//     this line rather than rendered on a separate third line.
//  3. Profile links line (when ShowLinks is true and ShowLinksInline is false)
//
// Returns the Y position after the header (for the next section to
// continue from). The caller is responsible for setting up fonts
// before calling this function.
//
// Required font names that must be registered with the GoPdf instance:
//   - "LiberationSans-Bold" for the name
//   - "LiberationSans-Regular" for contact details and links
func RenderProfileHeader(
	pdf *gopdf.GoPdf,
	profile domain.UserProfile,
	links []domain.ProfileLink,
	cfg HeaderConfig,
) (float64, error) {
	startY := pdf.GetY()
	currentY := startY

	// 1. Name (bold, large, centered)
	if profile.FullName != "" {
		if err := pdf.SetFont("LiberationSans-Bold", "", int(cfg.NameFontSize)); err != nil {
			return currentY, fmt.Errorf("header: set name font: %w", err)
		}

		nameWidth, err := pdf.MeasureTextWidth(profile.FullName)
		if err != nil {
			return currentY, fmt.Errorf("header: measure name: %w", err)
		}

		// Center the name within the usable page width.
		nameX := cfg.MarginLeft + (cfg.PageWidth-nameWidth)/2
		pdf.SetX(nameX)
		pdf.SetY(currentY)

		if err := pdf.Cell(nil, profile.FullName); err != nil {
			return currentY, fmt.Errorf("header: render name: %w", err)
		}

		currentY += cfg.NameFontSize + 4
	}

	// 2. Contact details line (optionally with links inline)
	contactParts := buildContactLine(profile)
	if cfg.ShowLinks && cfg.ShowLinksInline && len(links) > 0 {
		for _, link := range links {
			contactParts = append(contactParts, link.URL)
		}
	}
	if len(contactParts) > 0 {
		if err := pdf.SetFont("LiberationSans-Regular", "", int(cfg.DetailFontSize)); err != nil {
			return currentY, fmt.Errorf("header: set detail font: %w", err)
		}

		contactLine := strings.Join(contactParts, cfg.LinkSeparator)
		y, err := renderCenteredWrappedHeaderText(
			pdf,
			contactLine,
			cfg.MarginLeft,
			cfg.PageWidth,
			currentY,
			cfg.DetailFontSize,
		)
		if err != nil {
			return currentY, fmt.Errorf("header: render contact: %w", err)
		}
		currentY = y
	}

	// 3. Profile links line (only when ShowLinks is true and not inline)
	if cfg.ShowLinks && !cfg.ShowLinksInline && len(links) > 0 {
		if err := pdf.SetFont("LiberationSans-Regular", "", int(cfg.DetailFontSize)); err != nil {
			return currentY, fmt.Errorf("header: set link font: %w", err)
		}

		linkTexts := make([]string, len(links))
		for i, link := range links {
			linkTexts[i] = link.URL
		}
		linkLine := strings.Join(linkTexts, cfg.LinkSeparator)
		y, err := renderCenteredWrappedHeaderText(
			pdf,
			linkLine,
			cfg.MarginLeft,
			cfg.PageWidth,
			currentY,
			cfg.DetailFontSize,
		)
		if err != nil {
			return currentY, fmt.Errorf("header: render links: %w", err)
		}
		currentY = y
	}

	// Add spacing after header.
	currentY += cfg.SpaceAfter

	return currentY, nil
}

func renderCenteredWrappedHeaderText(
	pdf *gopdf.GoPdf,
	text string,
	x float64,
	maxWidth float64,
	y float64,
	fontSize float64,
) (float64, error) {
	lineHeight := fontSize + lineSpacing
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			y += lineHeight
			continue
		}

		words := strings.Fields(para)
		if len(words) == 0 {
			continue
		}

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
				lineWidth, err := pdf.MeasureTextWidth(currentLine)
				if err != nil {
					return y, fmt.Errorf("measure line: %w", err)
				}
				lineX := x + (maxWidth-lineWidth)/2
				if lineX < x {
					lineX = x
				}
				pdf.SetX(lineX)
				pdf.SetY(y)
				if err := pdf.Cell(nil, currentLine); err != nil {
					return y, fmt.Errorf("render line: %w", err)
				}
				y += lineHeight
				currentLine = word
				continue
			}

			currentLine = testLine
		}

		if currentLine != "" {
			lineWidth, err := pdf.MeasureTextWidth(currentLine)
			if err != nil {
				return y, fmt.Errorf("measure line: %w", err)
			}
			lineX := x + (maxWidth-lineWidth)/2
			if lineX < x {
				lineX = x
			}
			pdf.SetX(lineX)
			pdf.SetY(y)
			if err := pdf.Cell(nil, currentLine); err != nil {
				return y, fmt.Errorf("render line: %w", err)
			}
			y += lineHeight
		}
	}

	return y, nil
}

// buildContactLine assembles the non-empty contact fields into
// a slice for joining with a separator.
func buildContactLine(profile domain.UserProfile) []string {
	var parts []string
	if profile.Email != "" {
		parts = append(parts, profile.Email)
	}
	if profile.Phone != "" {
		parts = append(parts, profile.Phone)
	}
	if profile.Location != "" {
		parts = append(parts, profile.Location)
	}
	return parts
}
