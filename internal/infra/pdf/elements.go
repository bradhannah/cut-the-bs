package pdf

import (
	"encoding/json"
	"fmt"

	"cut-the-bs/internal/domain"

	"github.com/signintech/gopdf"
)

// renderContext holds shared state passed through the element rendering
// pipeline. It carries the resume data, page geometry derived from the
// template margins, and the current Y position.
type renderContext struct {
	pdf  *gopdf.GoPdf
	req  domain.RenderResumeRequest
	tmpl domain.TemplateDetail

	// Page geometry derived from template margins.
	marginLeft   float64
	marginRight  float64
	marginTop    float64
	marginBottom float64
	usableWidth  float64

	// Current Y cursor.
	y float64

	// childMap maps parent element ID to ordered child elements for
	// quick lookup during loop rendering.
	childMap map[int64][]domain.TemplateElement
}

// newRenderContext builds a renderContext from a template and request.
func newRenderContext(
	pdf *gopdf.GoPdf,
	req domain.RenderResumeRequest,
	tmpl domain.TemplateDetail,
) *renderContext {
	rc := &renderContext{
		pdf:          pdf,
		req:          req,
		tmpl:         tmpl,
		marginLeft:   tmpl.MarginLeft,
		marginRight:  tmpl.MarginRight,
		marginTop:    tmpl.MarginTop,
		marginBottom: tmpl.MarginBottom,
		usableWidth:  letterWidth - tmpl.MarginLeft - tmpl.MarginRight,
		y:            tmpl.MarginTop,
		childMap:     make(map[int64][]domain.TemplateElement),
	}

	// Build child map for loop containers.
	for _, el := range tmpl.Elements {
		if el.ParentID != nil {
			rc.childMap[*el.ParentID] = append(rc.childMap[*el.ParentID], el)
		}
	}

	return rc
}

// topLevelElements returns elements with no parent, sorted by sort_order.
// The elements are already sorted by sort_order from the store.
func topLevelElements(tmpl domain.TemplateDetail) []domain.TemplateElement {
	var top []domain.TemplateElement
	for _, el := range tmpl.Elements {
		if el.ParentID == nil {
			top = append(top, el)
		}
	}
	return top
}

// renderElements iterates top-level elements and dispatches each to
// the appropriate renderer. This is the main entry point for the
// template-driven pipeline.
func renderElements(rc *renderContext) error {
	for _, el := range topLevelElements(rc.tmpl) {
		if err := dispatchElement(rc, el); err != nil {
			return fmt.Errorf("element %d (%s): %w", el.ID, el.ElementType, err)
		}
	}
	return nil
}

// dispatchElement routes an element to its type-specific renderer.
func dispatchElement(rc *renderContext, el domain.TemplateElement) error {
	switch el.ElementType {
	// Formatting elements (T020)
	case domain.ElementSectionHeading:
		return renderSectionHeadingElement(rc, el)
	case domain.ElementHorizontalRule:
		return renderHorizontalRuleElement(rc, el)
	case domain.ElementSpacer:
		return renderSpacerElement(rc, el)
	case domain.ElementStaticText:
		return renderStaticTextElement(rc, el)

	// Data-bound elements (T021)
	case domain.ElementProfileHeader:
		return renderProfileHeaderElement(rc, el)
	case domain.ElementRoleDescriptors:
		return renderRoleDescriptorsElement(rc, el)
	case domain.ElementProfSummary:
		return renderProfSummaryElement(rc, el)
	case domain.ElementSkills:
		return renderSkillsElement(rc, el)
	case domain.ElementCoreExpertise:
		return renderCoreExpertiseElement(rc, el)

	// Loop containers (T022)
	case domain.ElementWorkHistoryLoop:
		return renderWorkHistoryLoopElement(rc, el)
	case domain.ElementEducationLoop:
		return renderEducationLoopElement(rc, el)
	case domain.ElementCertsLoop:
		return renderCertsLoopElement(rc, el)

	default:
		return fmt.Errorf("unknown element type: %q", el.ElementType)
	}
}

// =========================================================
// T020: Formatting element renderers
// =========================================================

// renderSectionHeadingElement renders a section heading with optional
// underline rule. Delegates to the shared renderSectionHeading (for
// professional/underlined) and renderModernSectionHeading (for modern/
// non-underlined) functions to guarantee byte-identical output.
//
// When DataBinding is set, the heading is skipped entirely if the
// bound data source is empty — matching the hardcoded templates.
func renderSectionHeadingElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg SectionHeadingConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse section_heading config: %w", err)
	}

	// Skip heading if bound data is empty.
	if cfg.DataBinding != "" && !hasData(rc, cfg.DataBinding) {
		return nil
	}

	if cfg.Underline {
		// Professional-style: uses shared renderSectionHeading which adds
		// sectionGap before, fontSize+2 after text, underline rule, +4 after.
		y, err := renderSectionHeading(rc.pdf, cfg.Text, rc.y, true)
		if err != nil {
			return err
		}
		rc.y = y
	} else {
		// Modern-style: uses shared renderModernSectionHeading which adds
		// modernSectionGap before, fontSize+6 after text, no underline.
		y, err := renderModernSectionHeading(rc.pdf, cfg.Text, rc.y)
		if err != nil {
			return err
		}
		rc.y = y
	}

	return nil
}

// hasData checks whether the named data source contains any data.
func hasData(rc *renderContext, binding string) bool {
	switch binding {
	case "summaries":
		return len(rc.req.Summaries) > 0
	case "work_history":
		return len(rc.req.WorkHistory) > 0
	case "skills":
		return len(rc.req.Skills) > 0
	case "core_expertise":
		return len(rc.req.CoreExpertise) > 0
	case "academics":
		return len(rc.req.Academics) > 0
	case "certifications":
		return len(rc.req.Certs) > 0
	default:
		return true // unknown binding → always render
	}
}

// renderHorizontalRuleElement renders a horizontal line across the
// usable page width.
func renderHorizontalRuleElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg HorizontalRuleConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse horizontal_rule config: %w", err)
	}

	rc.y += cfg.SpaceBefore

	rc.pdf.SetLineWidth(cfg.Weight)
	rc.pdf.Line(rc.marginLeft, rc.y, rc.marginLeft+rc.usableWidth, rc.y)

	rc.y += cfg.SpaceAfter

	return nil
}

// renderSpacerElement adds vertical space.
func renderSpacerElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg SpacerConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse spacer config: %w", err)
	}

	rc.y += cfg.Height
	return nil
}

// renderStaticTextElement renders a block of static text.
func renderStaticTextElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg StaticTextConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse static_text config: %w", err)
	}

	rc.y += cfg.SpaceBefore

	fontName := fontNameFromStyle(cfg.FontStyle)
	if err := setFont(rc.pdf, fontName, cfg.FontSize); err != nil {
		return err
	}

	var err error
	rc.y, err = renderWrappedText(rc.pdf, cfg.Text, rc.marginLeft, rc.y, rc.usableWidth, cfg.FontSize)
	if err != nil {
		return fmt.Errorf("render static text: %w", err)
	}

	rc.y += cfg.SpaceAfter
	return nil
}

// =========================================================
// T021: Data-bound element renderers
// =========================================================

// renderProfileHeaderElement renders the profile header (name, contact,
// links). Supports both centered (professional) and left-aligned (modern)
// layouts via config.
func renderProfileHeaderElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg ProfileHeaderConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse profile_header config: %w", err)
	}

	if cfg.Alignment == "left" {
		return renderProfileHeaderLeft(rc, cfg)
	}
	return renderProfileHeaderCentered(rc, cfg)
}

// renderProfileHeaderCentered delegates to the shared RenderProfileHeader
// from header.go to guarantee byte-identical output with the hardcoded
// professional template. RenderProfileHeader uses pdf.GetY() as the
// starting position, which matches the hardcoded pipeline where AddPage()
// sets Y to the gopdf default (10).
func renderProfileHeaderCentered(rc *renderContext, cfg ProfileHeaderConfig) error {
	headerCfg := HeaderConfig{
		NameFontSize:   cfg.NameFontSize,
		DetailFontSize: cfg.DetailFontSize,
		LinkSeparator:  cfg.LinkSeparator,
		MarginLeft:     rc.marginLeft,
		PageWidth:      rc.usableWidth,
	}

	// Do NOT set pdf.SetY — let RenderProfileHeader use pdf.GetY()
	// which matches the hardcoded renderProfessional behavior.
	y, err := RenderProfileHeader(rc.pdf, rc.req.Profile, rc.req.Links, headerCfg)
	if err != nil {
		return err
	}

	rc.y = y
	return nil
}

// renderProfileHeaderLeft reproduces the Modern header — left-aligned
// name, contact+links on one line with dot separator. Delegates to
// the shared renderModernHeader.
func renderProfileHeaderLeft(rc *renderContext, cfg ProfileHeaderConfig) error {
	y, err := renderModernHeader(rc.pdf, rc.req.Profile, rc.req.Links, rc.y)
	if err != nil {
		return err
	}
	rc.y = y
	return nil
}

// renderRoleDescriptorsElement renders the role descriptor bar.
// Delegates to the shared renderDescriptorBar (professional, centered)
// or renderModernDescriptors (modern, left-aligned).
func renderRoleDescriptorsElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.Descriptors) == 0 {
		return nil
	}

	var cfg RoleDescriptorsConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse role_descriptors config: %w", err)
	}

	if cfg.Alignment == "center" {
		// Professional style: centered, pipe-separated, regular font.
		y, err := renderDescriptorBar(rc.pdf, rc.req.Descriptors, rc.y)
		if err != nil {
			return err
		}
		rc.y = y
	} else {
		// Modern style: left-aligned, dot-separated, italic font.
		y, err := renderModernDescriptors(rc.pdf, rc.req.Descriptors, rc.y)
		if err != nil {
			return err
		}
		rc.y = y
	}

	return nil
}

// renderProfSummaryElement renders the professional summary section.
// Master summary is rendered as a paragraph, others as bullet points.
// Delegates to the shared renderWrappedText and renderBulletPoint.
func renderProfSummaryElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.Summaries) == 0 {
		return nil
	}

	var cfg ProfSummaryConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse professional_summary config: %w", err)
	}

	if err := setFont(rc.pdf, "LiberationSans-Regular", cfg.FontSize); err != nil {
		return err
	}

	// Render master summary as a plain paragraph.
	for _, sum := range rc.req.Summaries {
		if sum.BodyText == "" {
			continue
		}
		isMaster := rc.req.MasterSummaryID != nil && sum.ID == *rc.req.MasterSummaryID
		if isMaster {
			var err error
			rc.y, err = renderWrappedText(rc.pdf, sum.BodyText, rc.marginLeft, rc.y, rc.usableWidth, cfg.FontSize)
			if err != nil {
				return fmt.Errorf("summary text: %w", err)
			}
		}
	}

	// Render non-master summaries as bullet points using the shared function.
	for _, sum := range rc.req.Summaries {
		if sum.BodyText == "" {
			continue
		}
		isMaster := rc.req.MasterSummaryID != nil && sum.ID == *rc.req.MasterSummaryID
		if !isMaster {
			var err error
			rc.y, err = renderBulletPoint(rc.pdf, sum.BodyText, rc.y)
			if err != nil {
				return fmt.Errorf("summary bullet: %w", err)
			}
		}
	}

	return nil
}

// renderSkillsElement renders the skills section with category grouping.
// Delegates to the shared renderSkillsSection.
func renderSkillsElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.Skills) == 0 {
		return nil
	}

	var cfg SkillsConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse skills config: %w", err)
	}

	filterLegacy := !cfg.IncludeLegacy
	var err error
	rc.y, err = renderSkillsSection(rc.pdf, rc.req.Skills, rc.req.SkillCategoryNames, rc.y, filterLegacy)
	if err != nil {
		return fmt.Errorf("skills: %w", err)
	}

	return nil
}

// renderCoreExpertiseElement renders core expertise items as a
// separator-joined inline list. Delegates to the shared
// renderCoreExpertiseSection.
func renderCoreExpertiseElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.CoreExpertise) == 0 {
		return nil
	}

	var cfg CoreExpertiseConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse core_expertise config: %w", err)
	}

	var err error
	rc.y, err = renderCoreExpertiseSection(rc.pdf, rc.req.CoreExpertise, rc.y)
	if err != nil {
		return fmt.Errorf("render core expertise: %w", err)
	}

	return nil
}

// =========================================================
// T022: Loop container renderers
// =========================================================

// renderWorkHistoryLoopElement iterates over work history entries and
// dispatches child elements for each entry.
func renderWorkHistoryLoopElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.WorkHistory) == 0 {
		return nil
	}

	var cfg WorkHistoryLoopConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse work_history_loop config: %w", err)
	}

	children := rc.childMap[el.ID]

	for i, entry := range rc.req.WorkHistory {
		if i > 0 {
			rc.y += cfg.EntryGap
		}
		for _, child := range children {
			if err := dispatchWorkChild(rc, child, entry); err != nil {
				return fmt.Errorf("work entry %q, child %d (%s): %w",
					entry.EmployerName, child.ID, child.ElementType, err)
			}
		}
	}

	return nil
}

// dispatchWorkChild routes a work history loop child element to its
// renderer, passing the current work entry.
func dispatchWorkChild(
	rc *renderContext,
	el domain.TemplateElement,
	entry domain.WorkHistoryEntry,
) error {
	switch el.ElementType {
	case domain.ElementWorkTitle:
		return renderWorkTitleElement(rc, el, entry)
	case domain.ElementWorkDates:
		// Work dates are rendered inline with the title (right-aligned on
		// the same line), so they're handled inside renderWorkTitleElement.
		return nil
	case domain.ElementWorkSummary:
		return renderWorkSummaryElement(rc, el, entry)
	case domain.ElementWorkBullets:
		return renderWorkBulletsElement(rc, el, entry)
	case domain.ElementWorkOutcomes:
		return renderWorkOutcomesElement(rc, el, entry)
	case domain.ElementSectionHeading:
		return renderSectionHeadingElement(rc, el)
	case domain.ElementHorizontalRule:
		return renderHorizontalRuleElement(rc, el)
	case domain.ElementSpacer:
		return renderSpacerElement(rc, el)
	case domain.ElementStaticText:
		return renderStaticTextElement(rc, el)
	default:
		return fmt.Errorf("unknown work child element type: %q", el.ElementType)
	}
}

// renderWorkTitleElement renders the work entry title line, including
// the employer (if configured) and the right-aligned date. This
// delegates to the shared renderWorkEntry (professional) or
// renderModernWorkEntry (modern) pattern.
func renderWorkTitleElement(
	rc *renderContext,
	el domain.TemplateElement,
	entry domain.WorkHistoryEntry,
) error {
	var cfg WorkTitleConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse work_title config: %w", err)
	}

	lineHeight := cfg.FontSize + lineSpacing

	if err := checkPageBreak(rc.pdf, &rc.y, lineHeight*2); err != nil {
		return err
	}

	// Find the dates config from sibling elements.
	datesCfg := findWorkDatesConfig(rc, el)

	if cfg.IncludeEmployer && cfg.EmployerFontStyle == "italic" {
		// Professional style: bold title + italic " — EmployerName".
		if err := setFont(rc.pdf, "LiberationSans-Bold", cfg.FontSize); err != nil {
			return err
		}

		rc.pdf.SetX(rc.marginLeft)
		rc.pdf.SetY(rc.y)
		if err := rc.pdf.Cell(nil, entry.JobTitle); err != nil {
			return err
		}

		titleW, err := rc.pdf.MeasureTextWidth(entry.JobTitle)
		if err != nil {
			return err
		}

		// Render separator + employer in italic.
		if err := setFont(rc.pdf, "LiberationSans-Italic", cfg.FontSize); err != nil {
			return err
		}

		empText := cfg.EmployerSeparator + entry.EmployerName
		rc.pdf.SetX(rc.marginLeft + titleW)
		rc.pdf.SetY(rc.y)
		if err := rc.pdf.Cell(nil, empText); err != nil {
			return err
		}
	} else if cfg.IncludeEmployer && cfg.EmployerFontStyle == "bold" {
		// Modern style: "JobTitle, EmployerName" all bold.
		if err := setFont(rc.pdf, "LiberationSans-Bold", cfg.FontSize); err != nil {
			return err
		}

		titleLine := entry.JobTitle + cfg.EmployerSeparator + entry.EmployerName

		rc.pdf.SetX(rc.marginLeft)
		rc.pdf.SetY(rc.y)
		if err := rc.pdf.Cell(nil, titleLine); err != nil {
			return err
		}
	} else {
		// Title only.
		fontName := fontNameFromStyle(cfg.FontStyle)
		if err := setFont(rc.pdf, fontName, cfg.FontSize); err != nil {
			return err
		}

		rc.pdf.SetX(rc.marginLeft)
		rc.pdf.SetY(rc.y)
		if err := rc.pdf.Cell(nil, entry.JobTitle); err != nil {
			return err
		}
	}

	// Date range right-aligned on the same line.
	if datesCfg != nil {
		dateStr := formatDateRange(
			entry.StartDate, entry.EndDate,
			entry.DateGranularityStart, entry.DateGranularityEnd,
		)

		if err := setFont(rc.pdf, "LiberationSans-Regular", datesCfg.FontSize); err != nil {
			return err
		}

		dateW, err := rc.pdf.MeasureTextWidth(dateStr)
		if err != nil {
			return err
		}

		rc.pdf.SetX(rc.marginLeft + rc.usableWidth - dateW)
		rc.pdf.SetY(rc.y)
		if err := rc.pdf.Cell(nil, dateStr); err != nil {
			return err
		}
	}

	// Advance Y by SpaceAfter.
	// Professional: SpaceAfter=13 → matches lineHeight (10+3)
	// Modern: SpaceAfter=15 → matches lineHeight+2 (10+3+2)
	rc.y += cfg.SpaceAfter

	return nil
}

// findWorkDatesConfig looks for a work_dates sibling element within
// the same parent and returns its parsed config.
func findWorkDatesConfig(rc *renderContext, titleEl domain.TemplateElement) *WorkDatesConfig {
	if titleEl.ParentID == nil {
		return nil
	}
	siblings := rc.childMap[*titleEl.ParentID]
	for _, sib := range siblings {
		if sib.ElementType == domain.ElementWorkDates {
			var cfg WorkDatesConfig
			if err := json.Unmarshal([]byte(sib.Config), &cfg); err != nil {
				return nil
			}
			return &cfg
		}
	}
	return nil
}

// renderWorkSummaryElement renders the optional entry summary as
// italic wrapped text.
func renderWorkSummaryElement(
	rc *renderContext,
	_ domain.TemplateElement,
	entry domain.WorkHistoryEntry,
) error {
	if entry.Summary == "" {
		return nil
	}

	if err := setFont(rc.pdf, "LiberationSans-Italic", fontSizeBody); err != nil {
		return err
	}

	var err error
	rc.y, err = renderWrappedText(rc.pdf, entry.Summary, rc.marginLeft, rc.y, rc.usableWidth, fontSizeBody)
	if err != nil {
		return fmt.Errorf("entry summary: %w", err)
	}

	return nil
}

// renderWorkBulletsElement renders primary bullets for a work entry.
// Uses the shared renderBulletPoint to guarantee byte-identical output.
func renderWorkBulletsElement(
	rc *renderContext,
	el domain.TemplateElement,
	entry domain.WorkHistoryEntry,
) error {
	var cfg WorkBulletsConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse work_bullets config: %w", err)
	}

	// Set font before the bullet loop, matching the hardcoded behavior
	// which calls setFont("Regular", fontSizeBody) once before the loop.
	if err := setFont(rc.pdf, "LiberationSans-Regular", cfg.FontSize); err != nil {
		return err
	}

	for _, bullet := range entry.Bullets {
		if bullet.BulletType == domain.BulletTypeSecondary {
			continue
		}
		var err error
		rc.y, err = renderBulletPoint(rc.pdf, bullet.Text, rc.y)
		if err != nil {
			return err
		}
	}

	return nil
}

// renderWorkOutcomesElement renders secondary (outcome) bullets with
// an "Outcomes:" label. Uses the shared renderBulletPoint for the
// individual bullets to guarantee byte-identical output.
func renderWorkOutcomesElement(
	rc *renderContext,
	el domain.TemplateElement,
	entry domain.WorkHistoryEntry,
) error {
	// Check if there are any secondary bullets.
	hasSecondary := false
	for _, bullet := range entry.Bullets {
		if bullet.BulletType == domain.BulletTypeSecondary {
			hasSecondary = true
			break
		}
	}
	if !hasSecondary {
		return nil
	}

	var cfg WorkOutcomesConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse work_outcomes config: %w", err)
	}

	// Gap before outcomes block.
	rc.y += cfg.OutcomesGap

	// Render "Outcomes:" label in bold.
	if err := setFont(rc.pdf, "LiberationSans-Bold", cfg.FontSize); err != nil {
		return err
	}
	if err := checkPageBreak(rc.pdf, &rc.y, cfg.FontSize+lineSpacing); err != nil {
		return err
	}
	rc.pdf.SetX(rc.marginLeft + cfg.Indent)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, cfg.OutcomesLabel); err != nil {
		return err
	}
	rc.y += cfg.FontSize + lineSpacing

	// Render outcome bullets in italic using the shared renderBulletPoint.
	if err := setFont(rc.pdf, "LiberationSans-Italic", cfg.FontSize); err != nil {
		return err
	}

	for _, bullet := range entry.Bullets {
		if bullet.BulletType != domain.BulletTypeSecondary {
			continue
		}
		var err error
		rc.y, err = renderBulletPoint(rc.pdf, bullet.Text, rc.y)
		if err != nil {
			return err
		}
	}

	return nil
}

// renderEducationLoopElement iterates over academic credentials.
// Delegates to the shared renderAcademics.
func renderEducationLoopElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.Academics) == 0 {
		return nil
	}

	var err error
	rc.y, err = renderAcademics(rc.pdf, rc.req.Academics, rc.y)
	if err != nil {
		return fmt.Errorf("education: %w", err)
	}

	return nil
}

// renderCertsLoopElement iterates over certifications.
// Delegates to the shared renderCertifications.
func renderCertsLoopElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.Certs) == 0 {
		return nil
	}

	var err error
	rc.y, err = renderCertifications(rc.pdf, rc.req.Certs, rc.y)
	if err != nil {
		return fmt.Errorf("certifications: %w", err)
	}

	return nil
}

// =========================================================
// Shared helpers
// =========================================================

// fontNameFromStyle maps a config font style string to the
// Liberation Sans font name.
func fontNameFromStyle(style string) string {
	switch style {
	case "bold":
		return "LiberationSans-Bold"
	case "italic":
		return "LiberationSans-Italic"
	case "bold_italic":
		return "LiberationSans-BoldItalic"
	default:
		return "LiberationSans-Regular"
	}
}

// buildContactLine is re-exported from header.go. It assembles
// non-empty contact fields into a slice for joining with a separator.
// (Already defined in header.go, used here via package scope.)
