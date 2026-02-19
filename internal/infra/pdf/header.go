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
}

// DefaultHeaderConfig returns a sensible default header configuration
// for Letter-size pages with 1-inch margins.
func DefaultHeaderConfig() HeaderConfig {
	return HeaderConfig{
		NameFontSize:   18,
		DetailFontSize: 10,
		LinkSeparator:  " | ",
		MarginLeft:     72,  // 1 inch
		PageWidth:      468, // Letter width (612) minus 2x 1-inch margins
	}
}

// RenderProfileHeader renders the profile header at the current
// cursor position of the given GoPdf instance. It renders:
//  1. Full name (bold, large)
//  2. Contact line: email, phone, location (separated by " | ")
//  3. Profile links line (separated by " | ")
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

	// 2. Contact details line
	contactParts := buildContactLine(profile)
	if len(contactParts) > 0 {
		if err := pdf.SetFont("LiberationSans-Regular", "", int(cfg.DetailFontSize)); err != nil {
			return currentY, fmt.Errorf("header: set detail font: %w", err)
		}

		contactLine := strings.Join(contactParts, cfg.LinkSeparator)
		contactWidth, err := pdf.MeasureTextWidth(contactLine)
		if err != nil {
			return currentY, fmt.Errorf("header: measure contact: %w", err)
		}

		contactX := cfg.MarginLeft + (cfg.PageWidth-contactWidth)/2
		pdf.SetX(contactX)
		pdf.SetY(currentY)

		if err := pdf.Cell(nil, contactLine); err != nil {
			return currentY, fmt.Errorf("header: render contact: %w", err)
		}

		currentY += cfg.DetailFontSize + 3
	}

	// 3. Profile links line
	if len(links) > 0 {
		if err := pdf.SetFont("LiberationSans-Regular", "", int(cfg.DetailFontSize)); err != nil {
			return currentY, fmt.Errorf("header: set link font: %w", err)
		}

		linkTexts := make([]string, len(links))
		for i, link := range links {
			linkTexts[i] = link.URL
		}
		linkLine := strings.Join(linkTexts, cfg.LinkSeparator)

		linkWidth, err := pdf.MeasureTextWidth(linkLine)
		if err != nil {
			return currentY, fmt.Errorf("header: measure links: %w", err)
		}

		linkX := cfg.MarginLeft + (cfg.PageWidth-linkWidth)/2
		pdf.SetX(linkX)
		pdf.SetY(currentY)

		if err := pdf.Cell(nil, linkLine); err != nil {
			return currentY, fmt.Errorf("header: render links: %w", err)
		}

		currentY += cfg.DetailFontSize + 3
	}

	// Add spacing after header.
	currentY += 6

	return currentY, nil
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
