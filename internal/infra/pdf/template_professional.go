package pdf

import (
	"fmt"
	"strings"

	"cut-the-bs/internal/domain"

	"github.com/signintech/gopdf"
)

// renderProfessional renders a clean single-column ATS-optimized
// resume. Layout order:
//  1. Profile header (name, contact, links)
//  2. Role descriptor bar (pipe-separated)
//  3. Professional summary
//  4. Work history with bullets
//  5. Skills (comma-separated under category headers)
//  6. Education
//  7. Certifications
func renderProfessional(
	pdf *gopdf.GoPdf,
	req domain.RenderResumeRequest,
) error {
	// 1. Profile header.
	cfg := DefaultHeaderConfig()
	y, err := RenderProfileHeader(pdf, req.Profile, req.Links, cfg)
	if err != nil {
		return fmt.Errorf("profile header: %w", err)
	}

	// 2. Role descriptor bar.
	if len(req.Descriptors) > 0 {
		y, err = renderDescriptorBar(pdf, req.Descriptors, y)
		if err != nil {
			return fmt.Errorf("descriptor bar: %w", err)
		}
	}

	// 3. Professional summary.
	if len(req.Summaries) > 0 {
		y, err = renderSectionHeading(pdf, "Professional Summary", y, true)
		if err != nil {
			return err
		}

		if err := setFont(pdf, "LiberationSans-Regular", fontSizeBody); err != nil {
			return err
		}

		// Render master summary as a plain paragraph, others as bullets.
		for _, sum := range req.Summaries {
			if sum.BodyText == "" {
				continue
			}
			isMaster := req.MasterSummaryID != nil && sum.ID == *req.MasterSummaryID
			if isMaster {
				y, err = renderWrappedText(pdf, sum.BodyText, marginLeft, y, usableWidth, fontSizeBody)
				if err != nil {
					return fmt.Errorf("summary text: %w", err)
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
					return fmt.Errorf("summary bullet: %w", err)
				}
			}
		}
	}

	// 4. Work history.
	if len(req.WorkHistory) > 0 {
		y, err = renderSectionHeading(pdf, "Experience", y, true)
		if err != nil {
			return err
		}

		for i, entry := range req.WorkHistory {
			if i > 0 {
				y += 4
			}
			y, err = renderWorkEntry(pdf, entry, y)
			if err != nil {
				return fmt.Errorf("work entry %q: %w", entry.EmployerName, err)
			}
		}
	}

	// 5. Skills.
	if len(req.Skills) > 0 {
		y, err = renderSectionHeading(pdf, "Skills", y, true)
		if err != nil {
			return err
		}
		y, err = renderSkillsSection(pdf, req.Skills, req.SkillCategoryNames, y, false)
		if err != nil {
			return fmt.Errorf("skills: %w", err)
		}
	}

	// 5.5. Core Expertise.
	if len(req.CoreExpertise) > 0 {
		y, err = renderSectionHeading(pdf, "Core Expertise", y, true)
		if err != nil {
			return err
		}
		y, err = renderCoreExpertiseSection(pdf, req.CoreExpertise, y)
		if err != nil {
			return fmt.Errorf("core expertise: %w", err)
		}
	}

	// 6. Education.
	if len(req.Academics) > 0 {
		y, err = renderSectionHeading(pdf, "Education", y, true)
		if err != nil {
			return err
		}
		y, err = renderAcademics(pdf, req.Academics, y)
		if err != nil {
			return fmt.Errorf("education: %w", err)
		}
	}

	// 7. Certifications.
	if len(req.Certs) > 0 {
		y, err = renderSectionHeading(pdf, "Certifications", y, true)
		if err != nil {
			return err
		}
		y, err = renderCertifications(pdf, req.Certs, y)
		if err != nil {
			return fmt.Errorf("certifications: %w", err)
		}
	}

	_ = y
	return nil
}

// renderDescriptorBar renders role descriptors as a centered
// pipe-separated line below the header.
func renderDescriptorBar(
	pdf *gopdf.GoPdf,
	descriptors []domain.RoleDescriptor,
	y float64,
) (float64, error) {
	if err := setFont(pdf, "LiberationSans-Regular", fontSizeDescBar); err != nil {
		return y, err
	}

	titles := make([]string, len(descriptors))
	for i, d := range descriptors {
		titles[i] = d.Title
	}
	line := strings.Join(titles, " | ")

	width, err := pdf.MeasureTextWidth(line)
	if err != nil {
		return y, fmt.Errorf("measure descriptor bar: %w", err)
	}

	x := marginLeft + (usableWidth-width)/2
	if x < marginLeft {
		x = marginLeft
	}

	pdf.SetX(x)
	pdf.SetY(y)
	if err := pdf.Cell(nil, line); err != nil {
		return y, fmt.Errorf("render descriptor bar: %w", err)
	}

	y += fontSizeDescBar + 4
	return y, nil
}

// renderWorkEntry renders a single work history entry with
// employer, title, dates, and bullets.
func renderWorkEntry(
	pdf *gopdf.GoPdf,
	entry domain.WorkHistoryEntry,
	y float64,
) (float64, error) {
	lineHeight := fontSizeBody + lineSpacing

	// Title line: Bold job title — italic employer on one line.
	if err := checkPageBreak(pdf, &y, lineHeight*2); err != nil {
		return y, err
	}

	// Render bold job title.
	if err := setFont(pdf, "LiberationSans-Bold", fontSizeBody); err != nil {
		return y, err
	}

	pdf.SetX(marginLeft)
	pdf.SetY(y)
	if err := pdf.Cell(nil, entry.JobTitle); err != nil {
		return y, err
	}

	// Measure title width while still in bold font.
	titleW, err := pdf.MeasureTextWidth(entry.JobTitle)
	if err != nil {
		return y, err
	}

	// Render em-dash + employer in italic on same line.
	if err := setFont(pdf, "LiberationSans-Italic", fontSizeBody); err != nil {
		return y, err
	}

	emDash := " \u2014 " + entry.EmployerName
	pdf.SetX(marginLeft + titleW)
	pdf.SetY(y)
	if err := pdf.Cell(nil, emDash); err != nil {
		return y, err
	}

	// Date range right-aligned on the same line.
	dateStr := formatDateRange(
		entry.StartDate, entry.EndDate,
		entry.DateGranularityStart, entry.DateGranularityEnd,
	)

	if err := setFont(pdf, "LiberationSans-Regular", fontSizeSmall); err != nil {
		return y, err
	}

	dateWidth, err := pdf.MeasureTextWidth(dateStr)
	if err != nil {
		return y, err
	}

	pdf.SetX(marginLeft + usableWidth - dateWidth)
	pdf.SetY(y)
	if err := pdf.Cell(nil, dateStr); err != nil {
		return y, err
	}

	y += lineHeight

	// Entry summary (optional, italic paragraph above bullets).
	if entry.Summary != "" {
		if err := setFont(pdf, "LiberationSans-Italic", fontSizeBody); err != nil {
			return y, err
		}
		y, err = renderWrappedText(pdf, entry.Summary, marginLeft, y, usableWidth, fontSizeBody)
		if err != nil {
			return y, fmt.Errorf("entry summary: %w", err)
		}
	}

	// Primary bullets.
	if err := setFont(pdf, "LiberationSans-Regular", fontSizeBody); err != nil {
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
		if err := setFont(pdf, "LiberationSans-Bold", fontSizeBody); err != nil {
			return y, err
		}
		if err := checkPageBreak(pdf, &y, fontSizeBody+lineSpacing); err != nil {
			return y, err
		}
		pdf.SetX(marginLeft + bulletIndent)
		pdf.SetY(y)
		if err := pdf.Cell(nil, "Outcomes:"); err != nil {
			return y, err
		}
		y += fontSizeBody + lineSpacing

		if err := setFont(pdf, "LiberationSans-Italic", fontSizeBody); err != nil {
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

// renderBulletPoint renders a single bullet point with a bullet
// character and wrapped text.
func renderBulletPoint(
	pdf *gopdf.GoPdf,
	text string,
	y float64,
) (float64, error) {
	lineHeight := fontSizeBody + lineSpacing

	if err := checkPageBreak(pdf, &y, lineHeight); err != nil {
		return y, err
	}

	// Bullet character.
	pdf.SetX(marginLeft + bulletIndent)
	pdf.SetY(y)
	if err := pdf.Cell(nil, "\u2022"); err != nil { //nolint:gosmopolitan
		return y, err
	}

	// Wrapped text after bullet.
	textX := marginLeft + bulletIndent + bulletSymWidth
	textWidth := usableWidth - bulletIndent - bulletSymWidth

	y, err := renderWrappedText(pdf, text, textX, y, textWidth, fontSizeBody)
	if err != nil {
		return y, err
	}

	return y, nil
}

// renderSkillsSection renders skills grouped by category as
// comma-separated lists with bold category labels.
// Legacy skills get "(Legacy)" suffix.
func renderSkillsSection(
	pdf *gopdf.GoPdf,
	skills []domain.Skill,
	categoryNames map[int64]string,
	y float64,
	filterLegacy bool,
) (float64, error) {
	// Group skills by category ID for rendering.
	type catGroup struct {
		categoryID int64
		skills     []domain.Skill
	}

	grouped := make(map[int64]*catGroup)
	var order []int64

	for _, s := range skills {
		if filterLegacy && s.IsLegacy {
			continue
		}

		g, ok := grouped[s.CategoryID]
		if !ok {
			g = &catGroup{categoryID: s.CategoryID}
			grouped[s.CategoryID] = g
			order = append(order, s.CategoryID)
		}
		g.skills = append(g.skills, s)
	}

	lineHeight := fontSizeBody + lineSpacing

	for _, catID := range order {
		g := grouped[catID]
		if len(g.skills) == 0 {
			continue
		}

		if err := checkPageBreak(pdf, &y, lineHeight); err != nil {
			return y, err
		}

		// Render bold category label if available.
		catName := categoryNames[catID]
		labelWidth := 0.0
		if catName != "" {
			if err := setFont(pdf, "LiberationSans-Bold", fontSizeBody); err != nil {
				return y, err
			}
			label := catName + ": "
			pdf.SetX(marginLeft)
			pdf.SetY(y)
			if err := pdf.Cell(nil, label); err != nil {
				return y, err
			}
			w, err := pdf.MeasureTextWidth(label)
			if err != nil {
				return y, err
			}
			labelWidth = w
		}

		// Build skill names list.
		names := make([]string, len(g.skills))
		for i, s := range g.skills {
			if s.IsLegacy {
				names[i] = s.Name + " (Legacy)"
			} else {
				names[i] = s.Name
			}
		}
		line := strings.Join(names, ", ")

		// Render as wrapped text after the label, but continuation lines
		// return to the left margin at full usable width (paragraph style).
		if err := setFont(pdf, "LiberationSans-Regular", fontSizeBody); err != nil {
			return y, err
		}

		var wErr error
		y, wErr = renderWrappedTextHanging(pdf, line, marginLeft, labelWidth, y, usableWidth, fontSizeBody)
		if wErr != nil {
			return y, wErr
		}
	}

	return y, nil
}

// renderAcademics renders the education section.
func renderAcademics(
	pdf *gopdf.GoPdf,
	academics []domain.AcademicCredential,
	y float64,
) (float64, error) {
	lineHeight := fontSizeBody + lineSpacing

	for _, ac := range academics {
		if err := checkPageBreak(pdf, &y, lineHeight*2); err != nil {
			return y, err
		}

		// Credential type + field of study (bold).
		if err := setFont(pdf, "LiberationSans-Bold", fontSizeBody); err != nil {
			return y, err
		}

		credLine := ac.FieldOfStudy
		if ac.CredentialType != "" {
			credLine = ac.CredentialType + " in " + credLine
		}
		pdf.SetX(marginLeft)
		pdf.SetY(y)
		if err := pdf.Cell(nil, credLine); err != nil {
			return y, err
		}

		// Date right-aligned.
		dateStr := formatSingleDate(ac.CompletionDate, ac.DateGranularity)
		if dateStr != "" {
			if err := setFont(pdf, "LiberationSans-Regular", fontSizeSmall); err != nil {
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
		}

		y += lineHeight

		// Institution (regular).
		if err := setFont(pdf, "LiberationSans-Regular", fontSizeBody); err != nil {
			return y, err
		}
		pdf.SetX(marginLeft)
		pdf.SetY(y)
		if err := pdf.Cell(nil, ac.Institution); err != nil {
			return y, err
		}

		y += lineHeight + 2
	}

	return y, nil
}

// renderCertifications renders the certifications section.
func renderCertifications(
	pdf *gopdf.GoPdf,
	certs []domain.Certification,
	y float64,
) (float64, error) {
	lineHeight := fontSizeBody + lineSpacing

	for _, cert := range certs {
		if err := checkPageBreak(pdf, &y, lineHeight); err != nil {
			return y, err
		}

		// Cert name (bold) — issuing body.
		if err := setFont(pdf, "LiberationSans-Bold", fontSizeBody); err != nil {
			return y, err
		}
		pdf.SetX(marginLeft)
		pdf.SetY(y)
		if err := pdf.Cell(nil, cert.Name); err != nil {
			return y, err
		}

		// Measure cert name width while still in bold font.
		w, err := pdf.MeasureTextWidth(cert.Name)
		if err != nil {
			return y, err
		}

		// Issuing body + dates (regular, smaller).
		if err := setFont(pdf, "LiberationSans-Regular", fontSizeSmall); err != nil {
			return y, err
		}

		detail := " — " + cert.IssuingBody
		if cert.DateEarned != "" {
			detail += ", " + formatSingleDate(cert.DateEarned, "month")
		}
		if cert.ExpirationDate != "" {
			detail += " – Exp. " + formatSingleDate(cert.ExpirationDate, "month")
		}

		pdf.SetX(marginLeft + w)
		pdf.SetY(y)
		if err := pdf.Cell(nil, detail); err != nil {
			return y, err
		}

		y += lineHeight + 1
	}

	return y, nil
}

// renderCoreExpertiseSection renders core expertise items as
// pipe-separated keywords on one line (wrapping if needed).
func renderCoreExpertiseSection(
	pdf *gopdf.GoPdf,
	items []domain.CoreExpertise,
	y float64,
) (float64, error) {
	labels := make([]string, len(items))
	for i, ce := range items {
		labels[i] = ce.Label
	}
	line := strings.Join(labels, " | ")

	if err := setFont(pdf, "LiberationSans-Regular", fontSizeBody); err != nil {
		return y, err
	}

	y, err := renderWrappedText(pdf, line, marginLeft, y, usableWidth, fontSizeBody)
	if err != nil {
		return y, fmt.Errorf("render core expertise: %w", err)
	}

	return y, nil
}
