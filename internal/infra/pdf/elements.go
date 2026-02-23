package pdf

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"cut-the-bs/internal/domain"
	"cut-the-bs/internal/service"

	"github.com/signintech/gopdf"
)

// renderContext holds shared state passed through the element rendering
// pipeline. It carries the resume data, page geometry derived from the
// template margins, and the current Y position.
type renderContext struct {
	pdf  *gopdf.GoPdf
	req  domain.RenderResumeRequest
	tmpl domain.TemplateDetail

	// Cover letter fields (populated when rendering a cover letter).
	coverLetterReq *domain.RenderCoverLetterRequest

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

// newCoverLetterRenderContext builds a renderContext for cover letter
// rendering from a template and cover letter request.
func newCoverLetterRenderContext(
	pdf *gopdf.GoPdf,
	req domain.RenderCoverLetterRequest,
	tmpl domain.TemplateDetail,
) *renderContext {
	rc := &renderContext{
		pdf:  pdf,
		tmpl: tmpl,
		// Populate resume request fields from cover letter request
		// so shared elements (profile_header, static_text, etc.) work.
		req: domain.RenderResumeRequest{
			Profile: req.Profile,
			Links:   req.Links,
		},
		coverLetterReq: &req,
		marginLeft:     tmpl.MarginLeft,
		marginRight:    tmpl.MarginRight,
		marginTop:      tmpl.MarginTop,
		marginBottom:   tmpl.MarginBottom,
		usableWidth:    letterWidth - tmpl.MarginLeft - tmpl.MarginRight,
		y:              tmpl.MarginTop,
		childMap:       make(map[int64][]domain.TemplateElement),
	}

	// Build child map for loop containers (cover letters don't use
	// loops, but keep the mechanism for consistency).
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

	// Cover letter elements (T049)
	case domain.ElementBodyText:
		return renderBodyTextElement(rc, el)
	case domain.ElementParagraph:
		return renderParagraphElement(rc, el)
	case domain.ElementDate:
		return renderDateElement(rc, el)
	case domain.ElementGreeting:
		return renderGreetingElement(rc, el)
	case domain.ElementClosing:
		return renderClosingElement(rc, el)
	case domain.ElementRecipientAddress:
		return renderRecipientAddressElement(rc, el)

	default:
		slog.Warn("skipping unknown element type",
			"element_type", el.ElementType,
			"element_id", el.ID,
			"template", rc.tmpl.Name,
		)
		return nil
	}
}

// =========================================================
// T020: Formatting element renderers
// =========================================================

// renderSectionHeadingElement renders a section heading with optional
// underline rule. Uses config values (FontSize, Uppercase, Underline,
// SpaceBefore, SpaceAfter, UnderlineWeight) to allow full customisation.
//
// When DataBinding is set, the heading is skipped entirely if the
// bound data source is empty — matching legacy built-in behavior.
func renderSectionHeadingElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg SectionHeadingConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse section_heading config: %w", err)
	}

	// Skip heading if bound data is empty.
	if cfg.DataBinding != "" && !hasData(rc, cfg.DataBinding) {
		return nil
	}

	fontSize := cfg.FontSize
	if fontSize == 0 {
		fontSize = fontSizeSection
	}

	if err := checkPageBreak(rc.pdf, &rc.y, fontSize+cfg.SpaceBefore+cfg.SpaceAfter+4, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}

	rc.y += cfg.SpaceBefore

	fontName := fontNameFromStyle(cfg.FontStyle)
	if err := setFont(rc.pdf, fontName, fontSize); err != nil {
		return err
	}

	text := cfg.Text
	if cfg.Uppercase {
		text = strings.ToUpper(text)
	}

	rc.pdf.SetX(rc.marginLeft)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, text); err != nil {
		return fmt.Errorf("render section heading: %w", err)
	}

	if cfg.Underline {
		// Advance past text, add baseline-to-underline gap (2pt),
		// draw the rule, then add SpaceAfter below the rule.
		rc.y += fontSize + 2
		underlineWeight := cfg.UnderlineWeight
		if underlineWeight == 0 {
			underlineWeight = 0.5
		}
		rc.pdf.SetLineWidth(underlineWeight)
		rc.pdf.Line(rc.marginLeft, rc.y, rc.marginLeft+rc.usableWidth, rc.y)
		rc.y += cfg.SpaceAfter
	} else {
		// No underline: advance past text + SpaceAfter directly.
		rc.y += fontSize + cfg.SpaceAfter
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

// renderStaticTextElement renders a block of static text. When
// rendering within a cover letter context, variable placeholders
// (e.g., {{signer_name}}) are substituted before rendering.
func renderStaticTextElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg StaticTextConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse static_text config: %w", err)
	}

	text := cfg.Text

	// Apply variable substitutions for cover letter context.
	if rc.coverLetterReq != nil && len(rc.coverLetterReq.SubstitutionMap) > 0 {
		text = service.ApplySubstitutions(text, rc.coverLetterReq.SubstitutionMap)
	}

	rc.y += cfg.SpaceBefore

	fontName := fontNameFromStyle(cfg.FontStyle)
	if err := setFont(rc.pdf, fontName, cfg.FontSize); err != nil {
		return err
	}

	var x float64
	switch cfg.Alignment {
	case "right":
		w, err := rc.pdf.MeasureTextWidth(text)
		if err != nil {
			return fmt.Errorf("measure static text width: %w", err)
		}
		x = rc.marginLeft + rc.usableWidth - w
		if x < rc.marginLeft {
			x = rc.marginLeft
		}
	case "center":
		w, err := rc.pdf.MeasureTextWidth(text)
		if err != nil {
			return fmt.Errorf("measure static text width: %w", err)
		}
		x = rc.marginLeft + (rc.usableWidth-w)/2
		if x < rc.marginLeft {
			x = rc.marginLeft
		}
	default:
		x = rc.marginLeft
	}

	var err error
	rc.y, err = renderWrappedText(
		rc.pdf,
		text,
		x,
		rc.y,
		rc.usableWidth,
		cfg.FontSize,
		rc.marginBottom,
		rc.marginTop,
	)
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

// renderProfileHeaderCentered renders the centered profile header via
// the shared header helper.
func renderProfileHeaderCentered(rc *renderContext, cfg ProfileHeaderConfig) error {
	headerCfg := HeaderConfig{
		NameFontSize:    cfg.NameFontSize,
		DetailFontSize:  cfg.DetailFontSize,
		LinkSeparator:   cfg.LinkSeparator,
		MarginLeft:      rc.marginLeft,
		PageWidth:       rc.usableWidth,
		ShowLinks:       cfg.ShowLinks,
		ShowLinksInline: cfg.ShowLinksInline,
		SpaceAfter:      cfg.SpaceAfter,
	}

	rc.pdf.SetY(rc.y)
	y, err := RenderProfileHeader(rc.pdf, rc.req.Profile, rc.req.Links, headerCfg)
	if err != nil {
		return err
	}

	rc.y = y
	return nil
}

// renderProfileHeaderLeft renders a left-aligned profile header.
func renderProfileHeaderLeft(rc *renderContext, cfg ProfileHeaderConfig) error {
	if rc.req.Profile.FullName != "" {
		if err := setFont(rc.pdf, "LiberationSans-Bold", cfg.NameFontSize); err != nil {
			return err
		}
		rc.pdf.SetX(rc.marginLeft)
		rc.pdf.SetY(rc.y)
		if err := rc.pdf.Cell(nil, rc.req.Profile.FullName); err != nil {
			return err
		}
		rc.y += cfg.NameFontSize + 6
	}

	if err := setFont(rc.pdf, "LiberationSans-Regular", cfg.DetailFontSize); err != nil {
		return err
	}

	renderLeftDetailLine := func(line string) error {
		if strings.TrimSpace(line) == "" {
			return nil
		}
		y, err := renderWrappedText(
			rc.pdf,
			line,
			rc.marginLeft,
			rc.y,
			rc.usableWidth,
			cfg.DetailFontSize,
			rc.marginBottom,
			rc.marginTop,
		)
		if err != nil {
			return err
		}
		rc.y = y + 3
		return nil
	}

	contactParts := buildContactLine(rc.req.Profile)
	if cfg.ShowLinks && cfg.ShowLinksInline {
		for _, link := range rc.req.Links {
			contactParts = append(contactParts, link.URL)
		}
	}
	if len(contactParts) > 0 {
		if err := renderLeftDetailLine(strings.Join(contactParts, cfg.LinkSeparator)); err != nil {
			return err
		}
	}

	if cfg.ShowLinks && !cfg.ShowLinksInline && len(rc.req.Links) > 0 {
		linkParts := make([]string, 0, len(rc.req.Links))
		for _, link := range rc.req.Links {
			linkParts = append(linkParts, link.URL)
		}
		if err := renderLeftDetailLine(strings.Join(linkParts, cfg.LinkSeparator)); err != nil {
			return err
		}
	}

	rc.y += cfg.SpaceAfter
	return nil
}

// renderRoleDescriptorsElement renders role descriptors using configured
// font, separator, and alignment.
func renderRoleDescriptorsElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.Descriptors) == 0 {
		return nil
	}

	var cfg RoleDescriptorsConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse role_descriptors config: %w", err)
	}

	if err := setFont(rc.pdf, fontNameFromStyle(cfg.FontStyle), cfg.FontSize); err != nil {
		return err
	}

	titles := make([]string, 0, len(rc.req.Descriptors))
	for _, d := range rc.req.Descriptors {
		titles = append(titles, d.Title)
	}
	line := strings.Join(titles, cfg.Separator)

	x := rc.marginLeft
	if cfg.Alignment == "center" {
		w, err := rc.pdf.MeasureTextWidth(line)
		if err != nil {
			return fmt.Errorf("measure descriptor width: %w", err)
		}
		x = rc.marginLeft + (rc.usableWidth-w)/2
		if x < rc.marginLeft {
			x = rc.marginLeft
		}
	}

	rc.pdf.SetX(x)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, line); err != nil {
		return fmt.Errorf("render descriptors: %w", err)
	}

	rc.y += cfg.FontSize + cfg.SpaceAfter

	return nil
}

// renderProfSummaryElement renders the professional summary section.
// Master and non-master summaries can be independently enabled,
// and non-master summaries can render with or without bullet markers.
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

	showMaster := boolOrDefault(cfg.ShowMaster, true)
	showBulletSummaries := boolOrDefault(cfg.ShowBulletSummaries, true)
	enableBullets := boolOrDefault(cfg.EnableBullets, true)

	bulletChar := cfg.BulletChar
	if bulletChar == "" {
		bulletChar = "•"
	}
	bulletChar = decodeEscapedSymbol(bulletChar)

	started := false
	ensureStarted := func() {
		if !started {
			rc.y += cfg.SpaceBefore
			started = true
		}
	}

	if showMaster {
		for _, sum := range rc.req.Summaries {
			if sum.BodyText == "" {
				continue
			}
			isMaster := rc.req.MasterSummaryID != nil && sum.ID == *rc.req.MasterSummaryID
			if !isMaster {
				continue
			}

			ensureStarted()
			var err error
			rc.y, err = renderWrappedText(
				rc.pdf,
				sum.BodyText,
				rc.marginLeft,
				rc.y,
				rc.usableWidth,
				cfg.FontSize,
				rc.marginBottom,
				rc.marginTop,
			)
			if err != nil {
				return fmt.Errorf("summary text: %w", err)
			}
		}
	}

	if showBulletSummaries {
		for _, sum := range rc.req.Summaries {
			if sum.BodyText == "" {
				continue
			}
			isMaster := rc.req.MasterSummaryID != nil && sum.ID == *rc.req.MasterSummaryID
			if isMaster {
				continue
			}

			ensureStarted()
			if enableBullets {
				if err := renderConfigBullet(rc, sum.BodyText, cfg.FontSize, bulletChar, 12.0, 10.0); err != nil {
					return fmt.Errorf("summary bullet: %w", err)
				}
				continue
			}

			var err error
			rc.y, err = renderWrappedText(
				rc.pdf,
				sum.BodyText,
				rc.marginLeft,
				rc.y,
				rc.usableWidth,
				cfg.FontSize,
				rc.marginBottom,
				rc.marginTop,
			)
			if err != nil {
				return fmt.Errorf("summary paragraph: %w", err)
			}
		}
	}

	if started {
		rc.y += cfg.SpaceAfter
	}

	return nil
}

// renderSkillsElement renders skills using configurable grouping,
// filtering, label style, and separator settings.
func renderSkillsElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.Skills) == 0 {
		return nil
	}

	var cfg SkillsConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse skills config: %w", err)
	}

	type catGroup struct {
		categoryID int64
		skills     []domain.Skill
	}

	filtered := make([]domain.Skill, 0, len(rc.req.Skills))
	for _, skill := range rc.req.Skills {
		if !cfg.IncludeLegacy && skill.IsLegacy {
			continue
		}
		filtered = append(filtered, skill)
	}
	if len(filtered) == 0 {
		return nil
	}

	if err := setFont(rc.pdf, "LiberationSans-Regular", cfg.FontSize); err != nil {
		return err
	}

	buildSkillName := func(skill domain.Skill) string {
		if skill.IsLegacy {
			return skill.Name + cfg.LegacySuffix
		}
		return skill.Name
	}

	if !cfg.GroupByCategory {
		names := make([]string, 0, len(filtered))
		for _, skill := range filtered {
			names = append(names, buildSkillName(skill))
		}
		var err error
		rc.y, err = renderWrappedText(
			rc.pdf,
			strings.Join(names, cfg.SkillSeparator),
			rc.marginLeft,
			rc.y,
			rc.usableWidth,
			cfg.FontSize,
			rc.marginBottom,
			rc.marginTop,
		)
		if err != nil {
			return fmt.Errorf("skills: %w", err)
		}
		return nil
	}

	grouped := make(map[int64]*catGroup)
	order := make([]int64, 0)
	for _, skill := range filtered {
		g, ok := grouped[skill.CategoryID]
		if !ok {
			g = &catGroup{categoryID: skill.CategoryID}
			grouped[skill.CategoryID] = g
			order = append(order, skill.CategoryID)
		}
		g.skills = append(g.skills, skill)
	}

	lineHeight := cfg.FontSize + lineSpacing
	for _, catID := range order {
		group := grouped[catID]
		if len(group.skills) == 0 {
			continue
		}

		if err := checkPageBreak(rc.pdf, &rc.y, lineHeight, rc.marginBottom, rc.marginTop); err != nil {
			return err
		}

		labelWidth := 0.0
		catName := rc.req.SkillCategoryNames[catID]
		if catName != "" {
			if err := setFont(rc.pdf, fontNameFromStyle(cfg.CategoryFontStyle), cfg.FontSize); err != nil {
				return err
			}
			label := catName + ": "
			rc.pdf.SetX(rc.marginLeft)
			rc.pdf.SetY(rc.y)
			if err := rc.pdf.Cell(nil, label); err != nil {
				return err
			}
			w, err := rc.pdf.MeasureTextWidth(label)
			if err != nil {
				return err
			}
			labelWidth = w
		}

		names := make([]string, 0, len(group.skills))
		for _, skill := range group.skills {
			names = append(names, buildSkillName(skill))
		}

		if err := setFont(rc.pdf, "LiberationSans-Regular", cfg.FontSize); err != nil {
			return err
		}

		var err error
		rc.y, err = renderWrappedTextHanging(
			rc.pdf,
			strings.Join(names, cfg.SkillSeparator),
			rc.marginLeft,
			labelWidth,
			rc.y,
			rc.usableWidth,
			cfg.FontSize,
			rc.marginBottom,
			rc.marginTop,
		)
		if err != nil {
			return fmt.Errorf("skills category %d: %w", catID, err)
		}
	}

	return nil
}

// renderCoreExpertiseElement renders core expertise items as a
// separator-joined list using configured alignment and spacing.
func renderCoreExpertiseElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.CoreExpertise) == 0 {
		return nil
	}

	var cfg CoreExpertiseConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse core_expertise config: %w", err)
	}

	labels := make([]string, 0, len(rc.req.CoreExpertise))
	for _, item := range rc.req.CoreExpertise {
		labels = append(labels, item.Label)
	}
	line := strings.Join(labels, cfg.Separator)

	if err := setFont(rc.pdf, "LiberationSans-Regular", cfg.FontSize); err != nil {
		return err
	}

	if cfg.Alignment == "center" {
		w, err := rc.pdf.MeasureTextWidth(line)
		if err == nil && w <= rc.usableWidth {
			x := rc.marginLeft + (rc.usableWidth-w)/2
			if x < rc.marginLeft {
				x = rc.marginLeft
			}
			rc.pdf.SetX(x)
			rc.pdf.SetY(rc.y)
			if err := rc.pdf.Cell(nil, line); err != nil {
				return fmt.Errorf("render core expertise: %w", err)
			}
			rc.y += cfg.FontSize + lineSpacing + cfg.SpaceAfter
			return nil
		}
	}

	var err error
	rc.y, err = renderWrappedText(
		rc.pdf,
		line,
		rc.marginLeft,
		rc.y,
		rc.usableWidth,
		cfg.FontSize,
		rc.marginBottom,
		rc.marginTop,
	)
	if err != nil {
		return fmt.Errorf("render core expertise: %w", err)
	}
	rc.y += cfg.SpaceAfter

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
	rc.y += cfg.SpaceBefore

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
	rc.y += cfg.SpaceAfter

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
		slog.Warn("skipping unknown work child element type",
			"element_type", el.ElementType,
			"element_id", el.ID,
			"template", rc.tmpl.Name,
		)
		return nil
	}
}

// renderWorkTitleElement renders the work entry title line, including
// the employer (if configured) and the date aligned per config.
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
	rowLayout := cfg.TitleRowLayout
	if rowLayout == "" {
		rowLayout = "inline_with_dates"
	}
	stackDatesBelow := rowLayout == "stack_dates_below"

	// Find the dates config from sibling elements.
	datesCfg := findWorkDatesConfig(rc, el)
	requiresExtraDateLine := stackDatesBelow && datesCfg != nil

	requiredHeight := lineHeight * 2
	if requiresExtraDateLine {
		requiredHeight = lineHeight * 3
	}

	if err := checkPageBreak(rc.pdf, &rc.y, requiredHeight, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}

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
	} else if cfg.IncludeEmployer {
		// Employer included with regular (or other) font style: bold title + regular employer.
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

		// Render separator + employer in regular font.
		if err := setFont(rc.pdf, "LiberationSans-Regular", cfg.FontSize); err != nil {
			return err
		}

		empText := cfg.EmployerSeparator + entry.EmployerName
		rc.pdf.SetX(rc.marginLeft + titleW)
		rc.pdf.SetY(rc.y)
		if err := rc.pdf.Cell(nil, empText); err != nil {
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

	// Date range rendered either inline with title or on the next line.
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

		dateY := rc.y
		if stackDatesBelow {
			dateY += lineHeight
		}

		rc.pdf.SetX(workDatesX(rc, datesCfg.Alignment, dateW))
		rc.pdf.SetY(dateY)
		if err := rc.pdf.Cell(nil, dateStr); err != nil {
			return err
		}
	}

	if requiresExtraDateLine {
		rc.y += lineHeight
	}

	// Advance Y by SpaceAfter.
	// Professional: SpaceAfter=13 → matches lineHeight (10+3)
	// Modern: SpaceAfter=15 → matches lineHeight+2 (10+3+2)
	rc.y += cfg.SpaceAfter

	return nil
}

func workDatesX(rc *renderContext, alignment string, dateW float64) float64 {
	switch alignment {
	case "left":
		return rc.marginLeft
	case "center":
		return rc.marginLeft + (rc.usableWidth-dateW)/2
	default:
		return rc.marginLeft + rc.usableWidth - dateW
	}
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

// renderWorkSummaryElement renders the optional entry summary using
// configurable font size and style.
func renderWorkSummaryElement(
	rc *renderContext,
	el domain.TemplateElement,
	entry domain.WorkHistoryEntry,
) error {
	if strings.TrimSpace(entry.Summary) == "" {
		return nil
	}

	var cfg WorkSummaryConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse work_summary config: %w", err)
	}

	fontSize := cfg.FontSize
	if fontSize <= 0 {
		fontSize = 10.0
	}

	fontStyle := strings.TrimSpace(cfg.FontStyle)
	if fontStyle == "" {
		fontStyle = "italic"
	}

	if err := setFont(rc.pdf, fontNameFromStyle(fontStyle), fontSize); err != nil {
		return err
	}

	var err error
	rc.y, err = renderWrappedText(
		rc.pdf,
		entry.Summary,
		rc.marginLeft,
		rc.y,
		rc.usableWidth,
		fontSize,
		rc.marginBottom,
		rc.marginTop,
	)
	if err != nil {
		return fmt.Errorf("entry summary: %w", err)
	}

	return nil
}

// renderWorkBulletsElement renders primary bullets for a work entry.
func renderWorkBulletsElement(
	rc *renderContext,
	el domain.TemplateElement,
	entry domain.WorkHistoryEntry,
) error {
	var cfg WorkBulletsConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse work_bullets config: %w", err)
	}

	if err := setFont(rc.pdf, fontNameFromStyle(cfg.FontStyle), cfg.FontSize); err != nil {
		return err
	}

	bulletChar := cfg.BulletChar
	if bulletChar == "" {
		bulletChar = "•"
	}
	bulletChar = decodeEscapedSymbol(bulletChar)

	for _, bullet := range entry.Bullets {
		if bullet.BulletType == domain.BulletTypeSecondary {
			continue
		}
		if err := renderConfigBullet(rc, bullet.Text, cfg.FontSize, bulletChar, cfg.Indent, cfg.BulletSymWidth); err != nil {
			return err
		}
	}

	return nil
}

// renderWorkOutcomesElement renders secondary (outcome) bullets with
// an "Outcomes:" label.
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
	if err := checkPageBreak(rc.pdf, &rc.y, cfg.FontSize+lineSpacing, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}
	rc.pdf.SetX(rc.marginLeft + cfg.Indent)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, cfg.OutcomesLabel); err != nil {
		return err
	}
	rc.y += cfg.FontSize + lineSpacing

	if err := setFont(rc.pdf, fontNameFromStyle(cfg.FontStyle), cfg.FontSize); err != nil {
		return err
	}

	bulletChar := cfg.BulletChar
	if bulletChar == "" {
		bulletChar = "•"
	}
	bulletChar = decodeEscapedSymbol(bulletChar)

	for _, bullet := range entry.Bullets {
		if bullet.BulletType != domain.BulletTypeSecondary {
			continue
		}
		if err := renderConfigBullet(rc, bullet.Text, cfg.FontSize, bulletChar, cfg.Indent, cfg.BulletSymWidth); err != nil {
			return err
		}
	}

	return nil
}

// renderEducationLoopElement iterates over academic credentials and
// dispatches configured child elements per entry.
func renderEducationLoopElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.Academics) == 0 {
		return nil
	}

	var cfg EducationLoopConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse education_loop config: %w", err)
	}

	children := rc.childMap[el.ID]
	rc.y += cfg.SpaceBefore

	for i, ac := range rc.req.Academics {
		if i > 0 {
			rc.y += cfg.EntryGap
		}
		for _, child := range children {
			if err := dispatchEduChild(rc, child, ac); err != nil {
				return fmt.Errorf("education entry %q, child %d (%s): %w",
					ac.FieldOfStudy, child.ID, child.ElementType, err)
			}
		}
	}
	rc.y += cfg.SpaceAfter

	return nil
}

func dispatchEduChild(
	rc *renderContext,
	el domain.TemplateElement,
	ac domain.AcademicCredential,
) error {
	switch el.ElementType {
	case domain.ElementEduCredential:
		return renderEduCredentialElement(rc, el, ac)
	case domain.ElementEduInstitution:
		return renderEduInstitutionElement(rc, el, ac)
	case domain.ElementEduDate:
		return nil
	case domain.ElementSectionHeading:
		return renderSectionHeadingElement(rc, el)
	case domain.ElementHorizontalRule:
		return renderHorizontalRuleElement(rc, el)
	case domain.ElementSpacer:
		return renderSpacerElement(rc, el)
	case domain.ElementStaticText:
		return renderStaticTextElement(rc, el)
	default:
		slog.Warn("skipping unknown education child element type",
			"element_type", el.ElementType,
			"element_id", el.ID,
			"template", rc.tmpl.Name,
		)
		return nil
	}
}

func renderEduCredentialElement(rc *renderContext, el domain.TemplateElement, ac domain.AcademicCredential) error {
	var cfg EduCredentialConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse edu_credential config: %w", err)
	}

	line := ac.CredentialType
	if ac.FieldOfStudy != "" {
		if line != "" {
			line += ", " + ac.FieldOfStudy
		} else {
			line = ac.FieldOfStudy
		}
	}

	lineHeight := cfg.FontSize + lineSpacing
	if err := checkPageBreak(rc.pdf, &rc.y, lineHeight, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}

	if err := setFont(rc.pdf, fontNameFromStyle(cfg.FontStyle), cfg.FontSize); err != nil {
		return err
	}

	rc.pdf.SetX(rc.marginLeft)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, line); err != nil {
		return err
	}

	if dateCfg := findEduDateConfig(rc, el); dateCfg != nil {
		dateStr := formatSingleDate(ac.CompletionDate, ac.DateGranularity)
		if dateStr != "" {
			if err := setFont(rc.pdf, "LiberationSans-Regular", dateCfg.FontSize); err != nil {
				return err
			}
			w, err := rc.pdf.MeasureTextWidth(dateStr)
			if err != nil {
				return err
			}
			rc.pdf.SetX(workDatesX(rc, dateCfg.Alignment, w))
			rc.pdf.SetY(rc.y)
			if err := rc.pdf.Cell(nil, dateStr); err != nil {
				return err
			}
		}
	}

	rc.y += lineHeight
	return nil
}

func renderEduInstitutionElement(rc *renderContext, el domain.TemplateElement, ac domain.AcademicCredential) error {
	if ac.Institution == "" {
		return nil
	}

	var cfg EduInstitutionConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse edu_institution config: %w", err)
	}

	lineHeight := cfg.FontSize + lineSpacing
	if err := checkPageBreak(rc.pdf, &rc.y, lineHeight, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}

	if err := setFont(rc.pdf, fontNameFromStyle(cfg.FontStyle), cfg.FontSize); err != nil {
		return err
	}

	rc.pdf.SetX(rc.marginLeft)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, ac.Institution); err != nil {
		return err
	}

	rc.y += lineHeight
	return nil
}

func findEduDateConfig(rc *renderContext, credEl domain.TemplateElement) *EduDateConfig {
	if credEl.ParentID == nil {
		return nil
	}

	siblings := rc.childMap[*credEl.ParentID]
	for _, sibling := range siblings {
		if sibling.ElementType != domain.ElementEduDate {
			continue
		}
		var cfg EduDateConfig
		if err := json.Unmarshal([]byte(sibling.Config), &cfg); err != nil {
			return nil
		}
		return &cfg
	}

	return nil
}

// renderCertsLoopElement iterates over certifications and dispatches
// configured child elements per entry.
func renderCertsLoopElement(rc *renderContext, el domain.TemplateElement) error {
	if len(rc.req.Certs) == 0 {
		return nil
	}

	var cfg CertsLoopConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse certifications_loop config: %w", err)
	}

	children := rc.childMap[el.ID]
	rc.y += cfg.SpaceBefore

	for i, cert := range rc.req.Certs {
		if i > 0 {
			rc.y += cfg.EntryGap
		}
		for _, child := range children {
			if err := dispatchCertsChild(rc, child, cert); err != nil {
				return fmt.Errorf("certification entry %q, child %d (%s): %w",
					cert.Name, child.ID, child.ElementType, err)
			}
		}
	}
	rc.y += cfg.SpaceAfter

	return nil
}

func dispatchCertsChild(
	rc *renderContext,
	el domain.TemplateElement,
	cert domain.Certification,
) error {
	switch el.ElementType {
	case domain.ElementCertName:
		return renderCertNameElement(rc, el, cert)
	case domain.ElementCertDetail:
		return nil
	case domain.ElementSectionHeading:
		return renderSectionHeadingElement(rc, el)
	case domain.ElementHorizontalRule:
		return renderHorizontalRuleElement(rc, el)
	case domain.ElementSpacer:
		return renderSpacerElement(rc, el)
	case domain.ElementStaticText:
		return renderStaticTextElement(rc, el)
	default:
		slog.Warn("skipping unknown certification child element type",
			"element_type", el.ElementType,
			"element_id", el.ID,
			"template", rc.tmpl.Name,
		)
		return nil
	}
}

func renderCertNameElement(rc *renderContext, el domain.TemplateElement, cert domain.Certification) error {
	if cert.Name == "" {
		return nil
	}

	var cfg CertNameConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse cert_name config: %w", err)
	}

	lineHeight := cfg.FontSize + lineSpacing
	if err := checkPageBreak(rc.pdf, &rc.y, lineHeight, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}

	if err := setFont(rc.pdf, fontNameFromStyle(cfg.FontStyle), cfg.FontSize); err != nil {
		return err
	}

	rc.pdf.SetX(rc.marginLeft)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, cert.Name); err != nil {
		return err
	}

	nameW, err := rc.pdf.MeasureTextWidth(cert.Name)
	if err != nil {
		return err
	}

	if detailCfg := findCertDetailConfig(rc, el); detailCfg != nil {
		detail := ""
		if cert.IssuingBody != "" {
			detail += " — " + cert.IssuingBody
		}
		if cert.DateEarned != "" {
			detail += ", " + formatSingleDate(cert.DateEarned, "month")
		}
		if cert.ExpirationDate != "" {
			detail += " – Exp. " + formatSingleDate(cert.ExpirationDate, "month")
		}

		if detail != "" {
			if err := setFont(rc.pdf, fontNameFromStyle(detailCfg.FontStyle), detailCfg.FontSize); err != nil {
				return err
			}
			rc.pdf.SetX(rc.marginLeft + nameW)
			rc.pdf.SetY(rc.y)
			if err := rc.pdf.Cell(nil, detail); err != nil {
				return err
			}
		}
	}

	rc.y += lineHeight
	return nil
}

func findCertDetailConfig(rc *renderContext, nameEl domain.TemplateElement) *CertDetailConfig {
	if nameEl.ParentID == nil {
		return nil
	}

	siblings := rc.childMap[*nameEl.ParentID]
	for _, sibling := range siblings {
		if sibling.ElementType != domain.ElementCertDetail {
			continue
		}
		var cfg CertDetailConfig
		if err := json.Unmarshal([]byte(sibling.Config), &cfg); err != nil {
			return nil
		}
		return &cfg
	}

	return nil
}

// =========================================================
// T049: Cover letter element renderers
// =========================================================

// renderBodyTextElement renders the cover letter body text. When
// SubstitutionMap is present, variable placeholders in the body
// text are replaced before rendering.
func renderBodyTextElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg BodyTextConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse body_text config: %w", err)
	}

	if rc.coverLetterReq == nil {
		return nil // no cover letter data
	}

	bodyText := rc.coverLetterReq.Letter.BodyText
	if bodyText == "" {
		return nil
	}

	// Apply variable substitutions.
	if len(rc.coverLetterReq.SubstitutionMap) > 0 {
		bodyText = service.ApplySubstitutions(bodyText, rc.coverLetterReq.SubstitutionMap)
	}

	fontSize := cfg.FontSize
	if fontSize == 0 {
		fontSize = fontSizeBody
	}

	if err := setFont(rc.pdf, "LiberationSans-Regular", fontSize); err != nil {
		return err
	}

	var err error
	rc.y, err = renderWrappedText(
		rc.pdf,
		bodyText,
		rc.marginLeft,
		rc.y,
		rc.usableWidth,
		fontSize,
		rc.marginBottom,
		rc.marginTop,
	)
	if err != nil {
		return fmt.Errorf("render body text: %w", err)
	}

	rc.y += cfg.SpaceAfter
	return nil
}

// renderParagraphElement renders a cover letter paragraph assembled from
// ordered paragraph segments.
func renderParagraphElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg ParagraphConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse paragraph config: %w", err)
	}

	if rc.coverLetterReq == nil {
		return nil
	}

	paragraphText := buildParagraphText(cfg, rc.coverLetterReq.SubstitutionMap)
	if paragraphText == "" {
		return nil
	}

	fontSize := cfg.FontSize
	if fontSize == 0 {
		fontSize = fontSizeBody
	}

	if err := setFont(rc.pdf, "LiberationSans-Regular", fontSize); err != nil {
		return err
	}

	var err error
	rc.y, err = renderWrappedText(
		rc.pdf,
		paragraphText,
		rc.marginLeft,
		rc.y,
		rc.usableWidth,
		fontSize,
		rc.marginBottom,
		rc.marginTop,
	)
	if err != nil {
		return fmt.Errorf("render paragraph: %w", err)
	}

	rc.y += cfg.SpaceAfter
	return nil
}

func buildParagraphText(cfg ParagraphConfig, subs map[string]string) string {
	if len(cfg.Segments) == 0 {
		return ""
	}

	parts := make([]string, 0, len(cfg.Segments))
	for _, segment := range cfg.Segments {
		part := resolveParagraphSegment(segment, subs)
		part = normalizeParagraphSegmentText(part, segment.Type == "static")
		if part == "" {
			continue
		}
		parts = append(parts, part)
	}

	return strings.TrimSpace(strings.Join(parts, ""))
}

func resolveParagraphSegment(segment ParagraphSegmentConfig, subs map[string]string) string {
	switch segment.Type {
	case "static":
		return service.ApplySubstitutions(segment.Text, subs)
	case "profile", "application":
		return substitutionValue(subs, segment.Token)
	case "adhoc":
		if value := substitutionValue(subs, segment.Key); value != "" {
			return value
		}
		if value := substitutionValue(subs, "prompt:"+strings.TrimSpace(segment.Key)); value != "" {
			return value
		}
		if value := substitutionValue(subs, "prompt:"+strings.TrimSpace(segment.Label)); value != "" {
			return value
		}
		return ""
	default:
		return service.ApplySubstitutions(segment.Text, subs)
	}
}

func substitutionValue(subs map[string]string, key string) string {
	key = strings.TrimSpace(key)
	if key == "" || subs == nil {
		return ""
	}
	return strings.TrimSpace(subs[key])
}

func normalizeParagraphSegmentText(text string, preserveEdges bool) string {
	if text == "" {
		return ""
	}
	replacer := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ")
	cleaned := replacer.Replace(text)
	if preserveEdges {
		return cleaned
	}
	return strings.TrimSpace(cleaned)
}

// renderDateElement renders the current date on the cover letter.
func renderDateElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg DateConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse date config: %w", err)
	}

	fontSize := cfg.FontSize
	if fontSize == 0 {
		fontSize = fontSizeBody
	}

	format := cfg.Format
	if format == "" {
		format = "January 2, 2006"
	}

	dateStr := time.Now().Format(format)

	if err := setFont(rc.pdf, "LiberationSans-Regular", fontSize); err != nil {
		return err
	}

	if err := checkPageBreak(rc.pdf, &rc.y, fontSize+lineSpacing, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}

	switch cfg.Alignment {
	case "right":
		w, err := rc.pdf.MeasureTextWidth(dateStr)
		if err != nil {
			return fmt.Errorf("measure date width: %w", err)
		}
		rc.pdf.SetX(rc.marginLeft + rc.usableWidth - w)
	case "center":
		w, err := rc.pdf.MeasureTextWidth(dateStr)
		if err != nil {
			return fmt.Errorf("measure date width: %w", err)
		}
		rc.pdf.SetX(rc.marginLeft + (rc.usableWidth-w)/2)
	default: // "left"
		rc.pdf.SetX(rc.marginLeft)
	}

	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, dateStr); err != nil {
		return fmt.Errorf("render date: %w", err)
	}

	rc.y += fontSize + lineSpacing + cfg.SpaceAfter
	return nil
}

// renderGreetingElement renders the cover letter greeting/salutation.
// Supports variable substitution in the text.
func renderGreetingElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg GreetingConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse greeting config: %w", err)
	}

	text := cfg.Text
	if text == "" {
		return nil
	}

	// Apply variable substitutions.
	if rc.coverLetterReq != nil && len(rc.coverLetterReq.SubstitutionMap) > 0 {
		text = service.ApplySubstitutions(text, rc.coverLetterReq.SubstitutionMap)
	}

	fontSize := cfg.FontSize
	if fontSize == 0 {
		fontSize = fontSizeBody
	}

	if err := setFont(rc.pdf, "LiberationSans-Regular", fontSize); err != nil {
		return err
	}

	if err := checkPageBreak(rc.pdf, &rc.y, fontSize+lineSpacing, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}

	rc.pdf.SetX(rc.marginLeft)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, text); err != nil {
		return fmt.Errorf("render greeting: %w", err)
	}

	rc.y += fontSize + lineSpacing + cfg.SpaceAfter
	return nil
}

// renderClosingElement renders the cover letter closing/sign-off.
// Supports variable substitution in the text.
func renderClosingElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg ClosingConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse closing config: %w", err)
	}

	text := cfg.Text
	if text == "" {
		return nil
	}

	// Apply variable substitutions.
	if rc.coverLetterReq != nil && len(rc.coverLetterReq.SubstitutionMap) > 0 {
		text = service.ApplySubstitutions(text, rc.coverLetterReq.SubstitutionMap)
	}

	fontSize := cfg.FontSize
	if fontSize == 0 {
		fontSize = fontSizeBody
	}

	if err := setFont(rc.pdf, "LiberationSans-Regular", fontSize); err != nil {
		return err
	}

	if err := checkPageBreak(rc.pdf, &rc.y, fontSize+lineSpacing, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}

	rc.pdf.SetX(rc.marginLeft)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, text); err != nil {
		return fmt.Errorf("render closing: %w", err)
	}

	rc.y += fontSize + lineSpacing + cfg.SpaceAfter
	return nil
}

// renderRecipientAddressElement renders recipient address lines on
// the cover letter. The address text is pulled from the SubstitutionMap
// using the key "recipient_address".
func renderRecipientAddressElement(rc *renderContext, el domain.TemplateElement) error {
	var cfg RecipientAddressConfig
	if err := json.Unmarshal([]byte(el.Config), &cfg); err != nil {
		return fmt.Errorf("parse recipient_address config: %w", err)
	}

	if rc.coverLetterReq == nil {
		return nil
	}

	address := ""
	if rc.coverLetterReq.SubstitutionMap != nil {
		address = rc.coverLetterReq.SubstitutionMap["recipient_address"]
	}
	if address == "" {
		return nil // no address provided, skip
	}

	fontSize := cfg.FontSize
	if fontSize == 0 {
		fontSize = fontSizeBody
	}

	if err := setFont(rc.pdf, "LiberationSans-Regular", fontSize); err != nil {
		return err
	}

	var err error
	rc.y, err = renderWrappedText(
		rc.pdf,
		address,
		rc.marginLeft,
		rc.y,
		rc.usableWidth,
		fontSize,
		rc.marginBottom,
		rc.marginTop,
	)
	if err != nil {
		return fmt.Errorf("render recipient address: %w", err)
	}

	rc.y += cfg.SpaceAfter
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

func renderConfigBullet(
	rc *renderContext,
	text string,
	fontSize float64,
	bulletChar string,
	indent float64,
	bulletSymWidth float64,
) error {
	lineHeight := fontSize + lineSpacing

	if err := checkPageBreak(rc.pdf, &rc.y, lineHeight, rc.marginBottom, rc.marginTop); err != nil {
		return err
	}

	rc.pdf.SetX(rc.marginLeft + indent)
	rc.pdf.SetY(rc.y)
	if err := rc.pdf.Cell(nil, bulletChar); err != nil {
		return err
	}

	textX := rc.marginLeft + indent + bulletSymWidth
	textWidth := rc.usableWidth - indent - bulletSymWidth

	var err error
	rc.y, err = renderWrappedText(
		rc.pdf,
		text,
		textX,
		rc.y,
		textWidth,
		fontSize,
		rc.marginBottom,
		rc.marginTop,
	)
	if err != nil {
		return err
	}

	return nil
}

func decodeEscapedSymbol(symbol string) string {
	if symbol == "" || !strings.Contains(symbol, "\\") {
		return symbol
	}

	quoted := "\"" + strings.ReplaceAll(symbol, "\"", "\\\"") + "\""
	decoded, err := strconv.Unquote(quoted)
	if err != nil {
		return symbol
	}

	return decoded
}

func boolOrDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

// buildContactLine is re-exported from header.go. It assembles
// non-empty contact fields into a slice for joining with a separator.
// (Already defined in header.go, used here via package scope.)
