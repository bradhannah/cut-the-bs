package pdf

import (
	"fmt"
	"strings"

	"cut-the-bs/internal/domain"

	"github.com/signintech/gopdf"
)

// Modern template constants — differs from professional via larger
// name, more whitespace, and different section styling.
const (
	modernNameSize    = 22.0
	modernSectionSize = 11.0
	modernBodySize    = 10.0
	modernSmallSize   = 9.0
	modernSectionGap  = 14.0
)

// renderModern renders a contemporary resume layout with more
// generous spacing and a distinct visual hierarchy from the
// professional template. Same section order, different typography.
func renderModern(
	pdf *gopdf.GoPdf,
	req domain.RenderResumeRequest,
) error {
	y := marginTop

	// 1. Profile header — custom larger name, left-aligned.
	var err error
	y, err = renderModernHeader(pdf, req.Profile, req.Links, y)
	if err != nil {
		return fmt.Errorf("modern header: %w", err)
	}

	// 2. Descriptor bar — left-aligned, dot-separated.
	if len(req.Descriptors) > 0 {
		y, err = renderModernDescriptors(pdf, req.Descriptors, y)
		if err != nil {
			return fmt.Errorf("modern descriptors: %w", err)
		}
	}

	// Thin horizontal rule after header block.
	pdf.SetLineWidth(0.3)
	pdf.Line(marginLeft, y, marginLeft+usableWidth, y)
	y += 6

	// 3. Summary.
	if len(req.Summaries) > 0 {
		y, err = renderModernSectionHeading(pdf, "Summary", y)
		if err != nil {
			return err
		}

		if err := setFont(pdf, "LiberationSans-Regular", modernBodySize); err != nil {
			return err
		}

		// Render master summary as a plain paragraph, others as bullets.
		for _, sum := range req.Summaries {
			if sum.BodyText == "" {
				continue
			}
			isMaster := req.MasterSummaryID != nil && sum.ID == *req.MasterSummaryID
			if isMaster {
				y, err = renderWrappedText(pdf, sum.BodyText, marginLeft, y, usableWidth, modernBodySize)
				if err != nil {
					return fmt.Errorf("modern summary: %w", err)
				}
			}
		}
		for _, sum := range req.Summaries {
			if sum.BodyText == "" {
				continue
			}
			isMaster := req.MasterSummaryID != nil && sum.ID == *req.MasterSummaryID
			if !isMaster {
				y, err = renderBulletPoint(pdf, sum.BodyText, y)
				if err != nil {
					return fmt.Errorf("modern summary bullet: %w", err)
				}
			}
		}
	}

	// 4. Experience.
	if len(req.WorkHistory) > 0 {
		y, err = renderModernSectionHeading(pdf, "Experience", y)
		if err != nil {
			return err
		}

		for i, entry := range req.WorkHistory {
			if i > 0 {
				y += 6
			}
			y, err = renderModernWorkEntry(pdf, entry, y)
			if err != nil {
				return fmt.Errorf("modern work entry: %w", err)
			}
		}
	}

	// 5. Skills.
	if len(req.Skills) > 0 {
		y, err = renderModernSectionHeading(pdf, "Skills", y)
		if err != nil {
			return err
		}
		y, err = renderSkillsSection(pdf, req.Skills, req.SkillCategoryNames, y, false)
		if err != nil {
			return fmt.Errorf("modern skills: %w", err)
		}
	}

	// 5.5. Core Expertise.
	if len(req.CoreExpertise) > 0 {
		y, err = renderModernSectionHeading(pdf, "Core Expertise", y)
		if err != nil {
			return err
		}
		y, err = renderCoreExpertiseSection(pdf, req.CoreExpertise, y)
		if err != nil {
			return fmt.Errorf("modern core expertise: %w", err)
		}
	}

	// 6. Education.
	if len(req.Academics) > 0 {
		y, err = renderModernSectionHeading(pdf, "Education", y)
		if err != nil {
			return err
		}
		y, err = renderAcademics(pdf, req.Academics, y)
		if err != nil {
			return fmt.Errorf("modern education: %w", err)
		}
	}

	// 7. Certifications.
	if len(req.Certs) > 0 {
		y, err = renderModernSectionHeading(pdf, "Certifications", y)
		if err != nil {
			return err
		}
		y, err = renderCertifications(pdf, req.Certs, y)
		if err != nil {
			return fmt.Errorf("modern certifications: %w", err)
		}
	}

	_ = y
	return nil
}

// renderModernHeader renders a left-aligned header with large name
// and contact details on the same line below.
func renderModernHeader(
	pdf *gopdf.GoPdf,
	profile domain.UserProfile,
	links []domain.ProfileLink,
	y float64,
) (float64, error) {
	// Name — large, bold, left-aligned.
	if profile.FullName != "" {
		if err := setFont(pdf, "LiberationSans-Bold", modernNameSize); err != nil {
			return y, err
		}
		pdf.SetX(marginLeft)
		pdf.SetY(y)
		if err := pdf.Cell(nil, profile.FullName); err != nil {
			return y, err
		}
		y += modernNameSize + 6
	}

	// Contact line — left-aligned, smaller.
	if err := setFont(pdf, "LiberationSans-Regular", modernSmallSize); err != nil {
		return y, err
	}

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
	for _, link := range links {
		parts = append(parts, link.URL)
	}

	if len(parts) > 0 {
		line := strings.Join(parts, "  \u00B7  ") // middle dot separator
		pdf.SetX(marginLeft)
		pdf.SetY(y)
		if err := pdf.Cell(nil, line); err != nil {
			return y, err
		}
		y += modernSmallSize + 6
	}

	return y, nil
}

// renderModernDescriptors renders descriptors left-aligned with
// dot separators.
func renderModernDescriptors(
	pdf *gopdf.GoPdf,
	descriptors []domain.RoleDescriptor,
	y float64,
) (float64, error) {
	if err := setFont(pdf, "LiberationSans-Italic", modernBodySize); err != nil {
		return y, err
	}

	titles := make([]string, len(descriptors))
	for i, d := range descriptors {
		titles[i] = d.Title
	}
	line := strings.Join(titles, "  \u00B7  ")

	pdf.SetX(marginLeft)
	pdf.SetY(y)
	if err := pdf.Cell(nil, line); err != nil {
		return y, err
	}

	y += modernBodySize + 6
	return y, nil
}

// renderModernSectionHeading renders a section heading with
// uppercase text and spacing (no underline — modern style).
func renderModernSectionHeading(
	pdf *gopdf.GoPdf,
	title string,
	y float64,
) (float64, error) {
	if err := checkPageBreak(pdf, &y, modernSectionSize+modernSectionGap+4); err != nil {
		return y, err
	}

	y += modernSectionGap

	if err := setFont(pdf, "LiberationSans-Bold", modernSectionSize); err != nil {
		return y, err
	}

	pdf.SetX(marginLeft)
	pdf.SetY(y)
	if err := pdf.Cell(nil, strings.ToUpper(title)); err != nil {
		return y, fmt.Errorf("render modern heading: %w", err)
	}

	y += modernSectionSize + 6
	return y, nil
}

// renderModernWorkEntry renders a work entry with slightly
// different layout — employer and title on the same bold line.
func renderModernWorkEntry(
	pdf *gopdf.GoPdf,
	entry domain.WorkHistoryEntry,
	y float64,
) (float64, error) {
	lineHeight := modernBodySize + lineSpacing

	if err := checkPageBreak(pdf, &y, lineHeight*2); err != nil {
		return y, err
	}

	// Employer + Title combined (bold).
	if err := setFont(pdf, "LiberationSans-Bold", modernBodySize); err != nil {
		return y, err
	}

	titleLine := entry.JobTitle + ", " + entry.EmployerName

	pdf.SetX(marginLeft)
	pdf.SetY(y)
	if err := pdf.Cell(nil, titleLine); err != nil {
		return y, err
	}

	// Date range right-aligned.
	dateStr := formatDateRange(
		entry.StartDate, entry.EndDate,
		entry.DateGranularityStart, entry.DateGranularityEnd,
	)

	if err := setFont(pdf, "LiberationSans-Regular", modernSmallSize); err != nil {
		return y, err
	}

	dateW, err := pdf.MeasureTextWidth(dateStr)
	if err != nil {
		return y, err
	}

	pdf.SetX(marginLeft + usableWidth - dateW)
	pdf.SetY(y)
	if err := pdf.Cell(nil, dateStr); err != nil {
		return y, err
	}

	y += lineHeight + 2

	// Entry summary (optional, italic paragraph above bullets).
	if entry.Summary != "" {
		if err := setFont(pdf, "LiberationSans-Italic", modernBodySize); err != nil {
			return y, err
		}
		var summaryErr error
		y, summaryErr = renderWrappedText(pdf, entry.Summary, marginLeft, y, usableWidth, modernBodySize)
		if summaryErr != nil {
			return y, fmt.Errorf("modern entry summary: %w", summaryErr)
		}
	}

	// Primary bullets.
	if err := setFont(pdf, "LiberationSans-Regular", modernBodySize); err != nil {
		return y, err
	}

	for _, bullet := range entry.Bullets {
		if bullet.BulletType == domain.BulletTypeSecondary {
			continue
		}
		y, err = renderBulletPoint(pdf, bullet.Text, y)
		if err != nil {
			return y, err
		}
	}

	// Secondary (outcome) bullets — rendered with an "Outcomes:" label
	// and italic text to visually distinguish from primary bullets.
	hasSecondary := false
	for _, bullet := range entry.Bullets {
		if bullet.BulletType == domain.BulletTypeSecondary {
			hasSecondary = true
			break
		}
	}

	if hasSecondary {
		y += 2 // small gap before outcomes block

		// Render "Outcomes:" label in bold.
		if err := setFont(pdf, "LiberationSans-Bold", modernBodySize); err != nil {
			return y, err
		}
		if err := checkPageBreak(pdf, &y, modernBodySize+lineSpacing); err != nil {
			return y, err
		}
		pdf.SetX(marginLeft + bulletIndent)
		pdf.SetY(y)
		if err := pdf.Cell(nil, "Outcomes:"); err != nil {
			return y, err
		}
		y += modernBodySize + lineSpacing

		if err := setFont(pdf, "LiberationSans-Italic", modernBodySize); err != nil {
			return y, err
		}

		for _, bullet := range entry.Bullets {
			if bullet.BulletType != domain.BulletTypeSecondary {
				continue
			}
			y, err = renderBulletPoint(pdf, bullet.Text, y)
			if err != nil {
				return y, err
			}
		}
	}

	return y, nil
}
