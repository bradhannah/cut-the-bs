package pdf

import (
	"encoding/json"

	"cut-the-bs/internal/domain"
)

// =========================================================
// Element config structs — parsed from TemplateElement.Config
// =========================================================

// ProfileHeaderConfig controls the profile header rendering.
type ProfileHeaderConfig struct {
	NameFontSize    float64 `json:"name_font_size"`
	DetailFontSize  float64 `json:"detail_font_size"`
	Alignment       string  `json:"alignment"` // "center" | "left"
	LinkSeparator   string  `json:"link_separator"`
	ShowLinks       bool    `json:"show_links"`
	ShowLinksInline bool    `json:"show_links_inline"` // true = links on contact line
	SpaceAfter      float64 `json:"space_after"`
}

// RoleDescriptorsConfig controls the descriptor bar rendering.
type RoleDescriptorsConfig struct {
	FontSize   float64 `json:"font_size"`
	FontStyle  string  `json:"font_style"` // "regular" | "italic"
	Alignment  string  `json:"alignment"`  // "center" | "left"
	Separator  string  `json:"separator"`
	SpaceAfter float64 `json:"space_after"`
}

// ProfSummaryConfig controls professional summary rendering.
type ProfSummaryConfig struct {
	FontSize            float64 `json:"font_size"`
	BulletChar          string  `json:"bullet_char"`
	ShowMaster          *bool   `json:"show_master,omitempty"`
	ShowBulletSummaries *bool   `json:"show_bullet_summaries,omitempty"`
	EnableBullets       *bool   `json:"enable_bullets,omitempty"`
	SpaceBefore         float64 `json:"space_before"`
	SpaceAfter          float64 `json:"space_after"`
}

// SectionHeadingConfig controls section heading rendering.
type SectionHeadingConfig struct {
	Text            string  `json:"text"`
	FontSize        float64 `json:"font_size"`
	FontStyle       string  `json:"font_style"` // "bold"
	Uppercase       bool    `json:"uppercase"`
	Underline       bool    `json:"underline"`
	UnderlineWeight float64 `json:"underline_weight"`
	SpaceBefore     float64 `json:"space_before"`
	SpaceAfter      float64 `json:"space_after"`
	// DataBinding ties this heading to a data source. When the bound
	// data is empty, the heading is skipped entirely — matching
	// historical built-in behavior where section headings only render
	// when corresponding data exists. Valid values: "summaries",
	// "work_history", "skills", "core_expertise", "academics",
	// "certifications", or "" (always render).
	DataBinding string `json:"data_binding,omitempty"`
}

// HorizontalRuleConfig controls horizontal rule rendering.
type HorizontalRuleConfig struct {
	Weight      float64 `json:"weight"`
	SpaceBefore float64 `json:"space_before"`
	SpaceAfter  float64 `json:"space_after"`
}

// SpacerConfig controls spacer rendering.
type SpacerConfig struct {
	Height float64 `json:"height"`
}

// StaticTextConfig controls static text rendering.
type StaticTextConfig struct {
	Text        string  `json:"text"`
	FontSize    float64 `json:"font_size"`
	FontStyle   string  `json:"font_style"` // "regular" | "bold" | "italic" | "bold_italic"
	Alignment   string  `json:"alignment"`  // "left" | "center" | "right"
	SpaceBefore float64 `json:"space_before"`
	SpaceAfter  float64 `json:"space_after"`
}

// WorkHistoryLoopConfig controls the work history loop container.
type WorkHistoryLoopConfig struct {
	EntryGap    float64 `json:"entry_gap"`
	SpaceBefore float64 `json:"space_before"`
	SpaceAfter  float64 `json:"space_after"`
}

// WorkTitleConfig controls work entry title rendering.
type WorkTitleConfig struct {
	FontSize          float64 `json:"font_size"`
	FontStyle         string  `json:"font_style"` // "bold"
	TitleRowLayout    string  `json:"title_row_layout"`
	IncludeEmployer   bool    `json:"include_employer"`
	EmployerSeparator string  `json:"employer_separator"`
	EmployerFontStyle string  `json:"employer_font_style"` // "italic" | "bold"
	SpaceAfter        float64 `json:"space_after"`
}

// WorkDatesConfig controls work entry date rendering.
type WorkDatesConfig struct {
	FontSize  float64 `json:"font_size"`
	Alignment string  `json:"alignment"` // "right"
}

// WorkBulletsConfig controls work bullet rendering.
type WorkBulletsConfig struct {
	FontSize       float64 `json:"font_size"`
	FontStyle      string  `json:"font_style"` // "regular"
	BulletChar     string  `json:"bullet_char"`
	Indent         float64 `json:"indent"`
	BulletSymWidth float64 `json:"bullet_sym_width"`
}

// WorkOutcomesConfig controls work outcome bullet rendering.
type WorkOutcomesConfig struct {
	FontSize       float64 `json:"font_size"`
	FontStyle      string  `json:"font_style"` // "regular"
	BulletChar     string  `json:"bullet_char"`
	Indent         float64 `json:"indent"`
	BulletSymWidth float64 `json:"bullet_sym_width"`
	OutcomesLabel  string  `json:"outcomes_label"`
	OutcomesGap    float64 `json:"outcomes_gap"`
}

// SkillsConfig controls the skills section rendering.
type SkillsConfig struct {
	FontSize          float64 `json:"font_size"`
	GroupByCategory   bool    `json:"group_by_category"`
	IncludeLegacy     bool    `json:"include_legacy"`
	LegacySuffix      string  `json:"legacy_suffix"`
	CategoryFontStyle string  `json:"category_font_style"` // "bold"
	SkillSeparator    string  `json:"skill_separator"`
}

// CoreExpertiseConfig controls the core expertise rendering.
type CoreExpertiseConfig struct {
	FontSize   float64 `json:"font_size"`
	Separator  string  `json:"separator"`
	Alignment  string  `json:"alignment"` // "center" | "left"
	SpaceAfter float64 `json:"space_after"`
}

// EducationLoopConfig controls the education loop container.
type EducationLoopConfig struct {
	EntryGap    float64 `json:"entry_gap"`
	SpaceBefore float64 `json:"space_before"`
	SpaceAfter  float64 `json:"space_after"`
}

// CertsLoopConfig controls the certifications loop container.
type CertsLoopConfig struct {
	EntryGap    float64 `json:"entry_gap"`
	SpaceBefore float64 `json:"space_before"`
	SpaceAfter  float64 `json:"space_after"`
}

// EduCredentialConfig controls education credential rendering.
type EduCredentialConfig struct {
	FontSize  float64 `json:"font_size"`
	FontStyle string  `json:"font_style"` // "bold"
}

// EduInstitutionConfig controls education institution rendering.
type EduInstitutionConfig struct {
	FontSize  float64 `json:"font_size"`
	FontStyle string  `json:"font_style"` // "regular"
}

// EduDateConfig controls education date rendering.
type EduDateConfig struct {
	FontSize  float64 `json:"font_size"`
	Alignment string  `json:"alignment"` // "right"
}

// CertNameConfig controls certification name rendering.
type CertNameConfig struct {
	FontSize  float64 `json:"font_size"`
	FontStyle string  `json:"font_style"` // "bold"
}

// CertDetailConfig controls certification detail rendering.
type CertDetailConfig struct {
	FontSize  float64 `json:"font_size"`
	FontStyle string  `json:"font_style"` // "regular"
}

// WorkSummaryConfig controls work summary rendering.
type WorkSummaryConfig struct {
	FontSize  float64 `json:"font_size"`
	FontStyle string  `json:"font_style"` // "italic" | "regular"
}

// =========================================================
// Cover letter element config structs
// =========================================================

// BodyTextConfig controls cover letter body text rendering.
type BodyTextConfig struct {
	FontSize    float64 `json:"font_size"`
	LineSpacing float64 `json:"line_spacing"`
	SpaceAfter  float64 `json:"space_after"`
}

// ParagraphSegmentConfig defines one ordered sub-component inside a
// cover letter paragraph.
// Type values:
// - "static": use Text
// - "profile": use Token (profile token)
// - "application": use Token (application token)
// - "adhoc": prompt for Key/Label and substitute answer value
type ParagraphSegmentConfig struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Token     string `json:"token,omitempty"`
	Key       string `json:"key,omitempty"`
	Label     string `json:"label,omitempty"`
	HelpText  string `json:"help_text,omitempty"`
	Required  bool   `json:"required,omitempty"`
	Multiline bool   `json:"multiline,omitempty"`
}

// ParagraphConfig controls cover letter paragraph rendering.
type ParagraphConfig struct {
	FontSize    float64                  `json:"font_size"`
	LineSpacing float64                  `json:"line_spacing"`
	SpaceAfter  float64                  `json:"space_after"`
	Segments    []ParagraphSegmentConfig `json:"segments"`
}

// DateConfig controls cover letter date rendering.
type DateConfig struct {
	FontSize   float64 `json:"font_size"`
	Format     string  `json:"format"`    // Go time format string e.g. "January 2, 2006"
	Alignment  string  `json:"alignment"` // "left", "center", "right"
	SpaceAfter float64 `json:"space_after"`
}

// GreetingConfig controls cover letter greeting rendering.
type GreetingConfig struct {
	Text       string  `json:"text"` // e.g. "Dear Hiring Manager,"
	FontSize   float64 `json:"font_size"`
	SpaceAfter float64 `json:"space_after"`
}

// ClosingConfig controls cover letter closing rendering.
type ClosingConfig struct {
	Text       string  `json:"text"` // e.g. "Sincerely,"
	FontSize   float64 `json:"font_size"`
	SpaceAfter float64 `json:"space_after"`
}

// RecipientAddressConfig controls cover letter recipient address rendering.
type RecipientAddressConfig struct {
	FontSize   float64 `json:"font_size"`
	SpaceAfter float64 `json:"space_after"`
}

// =========================================================
// Helper to serialize config structs to JSON
// =========================================================

func mustMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("builtin: marshal config: " + err.Error())
	}
	return string(b)
}

// =========================================================
// Built-in Professional template
// =========================================================

// ProfessionalTemplate returns the built-in Professional resume
// template as a TemplateDetail. The element configs preserve the
// proven Professional visual defaults.
func ProfessionalTemplate() domain.TemplateDetail {
	return domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           1,
			Name:         "Professional",
			Description:  "Classic centered layout with underlined section headings",
			TemplateType: domain.TemplateTypeResume,
			IsBuiltin:    true,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		},
		Elements: []domain.TemplateElement{
			// Top-level elements (sort_order 0–13)
			{ID: 1, TemplateID: 1, ParentID: nil, ElementType: domain.ElementProfileHeader, SortOrder: 0,
				Config: mustMarshal(ProfileHeaderConfig{
					NameFontSize: 18.0, DetailFontSize: 10.0,
					Alignment: "center", LinkSeparator: " | ",
					ShowLinks: true, ShowLinksInline: false,
					SpaceAfter: 6.0,
				})},
			{ID: 2, TemplateID: 1, ParentID: nil, ElementType: domain.ElementRoleDescriptors, SortOrder: 1,
				Config: mustMarshal(RoleDescriptorsConfig{
					FontSize: 10.0, FontStyle: "regular",
					Alignment: "center", Separator: " | ",
					SpaceAfter: 4.0,
				})},
			{ID: 3, TemplateID: 1, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 2,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Professional Summary", FontSize: 12.0,
					FontStyle: "bold", Uppercase: true,
					Underline: true, UnderlineWeight: 0.5,
					SpaceBefore: 10.0, SpaceAfter: 4.0,
					DataBinding: "summaries",
				})},
			{ID: 4, TemplateID: 1, ParentID: nil, ElementType: domain.ElementProfSummary, SortOrder: 3,
				Config: mustMarshal(ProfSummaryConfig{
					FontSize: 10.0, BulletChar: "\u2022",
					ShowMaster: boolPtr(true), ShowBulletSummaries: boolPtr(true),
					EnableBullets: boolPtr(true),
					SpaceBefore:   0.0, SpaceAfter: 0.0,
				})},
			{ID: 7, TemplateID: 1, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 4,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Experience", FontSize: 12.0,
					FontStyle: "bold", Uppercase: true,
					Underline: true, UnderlineWeight: 0.5,
					SpaceBefore: 10.0, SpaceAfter: 4.0,
					DataBinding: "work_history",
				})},
			{ID: 8, TemplateID: 1, ParentID: nil, ElementType: domain.ElementWorkHistoryLoop, SortOrder: 5,
				Config: mustMarshal(WorkHistoryLoopConfig{
					EntryGap: 4.0, SpaceBefore: 0.0, SpaceAfter: 0.0,
				})},
			{ID: 9, TemplateID: 1, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 6,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Skills", FontSize: 12.0,
					FontStyle: "bold", Uppercase: true,
					Underline: true, UnderlineWeight: 0.5,
					SpaceBefore: 10.0, SpaceAfter: 4.0,
					DataBinding: "skills",
				})},
			{ID: 10, TemplateID: 1, ParentID: nil, ElementType: domain.ElementSkills, SortOrder: 7,
				Config: mustMarshal(SkillsConfig{
					FontSize: 10.0, GroupByCategory: true,
					IncludeLegacy: true, LegacySuffix: " (Legacy)",
					CategoryFontStyle: "bold", SkillSeparator: ", ",
				})},
			{ID: 5, TemplateID: 1, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 8,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Core Expertise", FontSize: 12.0,
					FontStyle: "bold", Uppercase: true,
					Underline: true, UnderlineWeight: 0.5,
					SpaceBefore: 10.0, SpaceAfter: 4.0,
					DataBinding: "core_expertise",
				})},
			{ID: 6, TemplateID: 1, ParentID: nil, ElementType: domain.ElementCoreExpertise, SortOrder: 9,
				Config: mustMarshal(CoreExpertiseConfig{
					FontSize: 10.0, Separator: " | ",
					Alignment: "center", SpaceAfter: 0.0,
				})},
			{ID: 11, TemplateID: 1, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 10,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Education", FontSize: 12.0,
					FontStyle: "bold", Uppercase: true,
					Underline: true, UnderlineWeight: 0.5,
					SpaceBefore: 10.0, SpaceAfter: 4.0,
					DataBinding: "academics",
				})},
			{ID: 12, TemplateID: 1, ParentID: nil, ElementType: domain.ElementEducationLoop, SortOrder: 11,
				Config: mustMarshal(EducationLoopConfig{
					EntryGap: 2.0, SpaceBefore: 0.0, SpaceAfter: 0.0,
				})},
			{ID: 13, TemplateID: 1, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 12,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Certifications", FontSize: 12.0,
					FontStyle: "bold", Uppercase: true,
					Underline: true, UnderlineWeight: 0.5,
					SpaceBefore: 10.0, SpaceAfter: 4.0,
					DataBinding: "certifications",
				})},
			{ID: 14, TemplateID: 1, ParentID: nil, ElementType: domain.ElementCertsLoop, SortOrder: 13,
				Config: mustMarshal(CertsLoopConfig{
					EntryGap: 1.0, SpaceBefore: 0.0, SpaceAfter: 0.0,
				})},

			// work_history_loop children (parent_id = 8)
			{ID: 15, TemplateID: 1, ParentID: int64Ptr(8), ElementType: domain.ElementWorkTitle, SortOrder: 0,
				Config: mustMarshal(WorkTitleConfig{
					FontSize: 10.0, FontStyle: "bold",
					IncludeEmployer: true, EmployerSeparator: " \u2014 ",
					EmployerFontStyle: "italic", SpaceAfter: 13.0,
				})},
			{ID: 16, TemplateID: 1, ParentID: int64Ptr(8), ElementType: domain.ElementWorkDates, SortOrder: 1,
				Config: mustMarshal(WorkDatesConfig{
					FontSize: 9.0, Alignment: "right",
				})},
			{ID: 17, TemplateID: 1, ParentID: int64Ptr(8), ElementType: domain.ElementWorkSummary, SortOrder: 2,
				Config: mustMarshal(WorkSummaryConfig{FontSize: 10.0, FontStyle: "italic"})},
			{ID: 18, TemplateID: 1, ParentID: int64Ptr(8), ElementType: domain.ElementWorkBullets, SortOrder: 3,
				Config: mustMarshal(WorkBulletsConfig{
					FontSize: 10.0, FontStyle: "regular",
					BulletChar: "\u2022", Indent: 12.0,
					BulletSymWidth: 10.0,
				})},
			{ID: 19, TemplateID: 1, ParentID: int64Ptr(8), ElementType: domain.ElementWorkOutcomes, SortOrder: 4,
				Config: mustMarshal(WorkOutcomesConfig{
					FontSize: 10.0, FontStyle: "regular",
					BulletChar: "\u2022", Indent: 12.0,
					BulletSymWidth: 10.0, OutcomesLabel: "Outcomes:",
					OutcomesGap: 2.0,
				})},

			// education_loop children (parent_id = 12)
			{ID: 20, TemplateID: 1, ParentID: int64Ptr(12), ElementType: domain.ElementEduCredential, SortOrder: 0,
				Config: mustMarshal(EduCredentialConfig{
					FontSize: 10.0, FontStyle: "bold",
				})},
			{ID: 21, TemplateID: 1, ParentID: int64Ptr(12), ElementType: domain.ElementEduInstitution, SortOrder: 1,
				Config: mustMarshal(EduInstitutionConfig{
					FontSize: 10.0, FontStyle: "regular",
				})},
			{ID: 22, TemplateID: 1, ParentID: int64Ptr(12), ElementType: domain.ElementEduDate, SortOrder: 2,
				Config: mustMarshal(EduDateConfig{
					FontSize: 9.0, Alignment: "right",
				})},

			// certifications_loop children (parent_id = 14)
			{ID: 23, TemplateID: 1, ParentID: int64Ptr(14), ElementType: domain.ElementCertName, SortOrder: 0,
				Config: mustMarshal(CertNameConfig{
					FontSize: 10.0, FontStyle: "bold",
				})},
			{ID: 24, TemplateID: 1, ParentID: int64Ptr(14), ElementType: domain.ElementCertDetail, SortOrder: 1,
				Config: mustMarshal(CertDetailConfig{
					FontSize: 9.0, FontStyle: "regular",
				})},
		},
	}
}

// =========================================================
// Built-in Modern template
// =========================================================

// ModernTemplate returns the built-in Modern resume template as a
// TemplateDetail. The element configs preserve the proven Modern
// visual defaults.
func ModernTemplate() domain.TemplateDetail {
	return domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           2,
			Name:         "Modern",
			Description:  "Left-aligned layout with clean typography and no underlines",
			TemplateType: domain.TemplateTypeResume,
			IsBuiltin:    true,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		},
		Elements: []domain.TemplateElement{
			// Top-level elements
			{ID: 25, TemplateID: 2, ParentID: nil, ElementType: domain.ElementProfileHeader, SortOrder: 0,
				Config: mustMarshal(ProfileHeaderConfig{
					NameFontSize: 22.0, DetailFontSize: 9.0,
					Alignment: "left", LinkSeparator: "  \u00B7  ",
					ShowLinks: true, ShowLinksInline: true,
					SpaceAfter: 6.0,
				})},
			{ID: 26, TemplateID: 2, ParentID: nil, ElementType: domain.ElementRoleDescriptors, SortOrder: 1,
				Config: mustMarshal(RoleDescriptorsConfig{
					FontSize: 10.0, FontStyle: "italic",
					Alignment: "left", Separator: "  \u00B7  ",
					SpaceAfter: 6.0,
				})},
			{ID: 27, TemplateID: 2, ParentID: nil, ElementType: domain.ElementHorizontalRule, SortOrder: 2,
				Config: mustMarshal(HorizontalRuleConfig{
					Weight: 0.3, SpaceBefore: 0.0, SpaceAfter: 6.0,
				})},
			{ID: 28, TemplateID: 2, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 3,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Summary", FontSize: 11.0,
					FontStyle: "bold", Uppercase: true,
					Underline: false, UnderlineWeight: 0.0,
					SpaceBefore: 14.0, SpaceAfter: 6.0,
					DataBinding: "summaries",
				})},
			{ID: 29, TemplateID: 2, ParentID: nil, ElementType: domain.ElementProfSummary, SortOrder: 4,
				Config: mustMarshal(ProfSummaryConfig{
					FontSize: 10.0, BulletChar: "\u2022",
					ShowMaster: boolPtr(true), ShowBulletSummaries: boolPtr(true),
					EnableBullets: boolPtr(true),
					SpaceBefore:   0.0, SpaceAfter: 0.0,
				})},
			{ID: 30, TemplateID: 2, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 5,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Experience", FontSize: 11.0,
					FontStyle: "bold", Uppercase: true,
					Underline: false, UnderlineWeight: 0.0,
					SpaceBefore: 14.0, SpaceAfter: 6.0,
					DataBinding: "work_history",
				})},
			{ID: 31, TemplateID: 2, ParentID: nil, ElementType: domain.ElementWorkHistoryLoop, SortOrder: 6,
				Config: mustMarshal(WorkHistoryLoopConfig{
					EntryGap: 6.0, SpaceBefore: 0.0, SpaceAfter: 0.0,
				})},
			{ID: 32, TemplateID: 2, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 7,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Skills", FontSize: 11.0,
					FontStyle: "bold", Uppercase: true,
					Underline: false, UnderlineWeight: 0.0,
					SpaceBefore: 14.0, SpaceAfter: 6.0,
					DataBinding: "skills",
				})},
			{ID: 33, TemplateID: 2, ParentID: nil, ElementType: domain.ElementSkills, SortOrder: 8,
				Config: mustMarshal(SkillsConfig{
					FontSize: 10.0, GroupByCategory: true,
					IncludeLegacy: true, LegacySuffix: " (Legacy)",
					CategoryFontStyle: "bold", SkillSeparator: ", ",
				})},
			{ID: 34, TemplateID: 2, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 9,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Core Expertise", FontSize: 11.0,
					FontStyle: "bold", Uppercase: true,
					Underline: false, UnderlineWeight: 0.0,
					SpaceBefore: 14.0, SpaceAfter: 6.0,
					DataBinding: "core_expertise",
				})},
			{ID: 35, TemplateID: 2, ParentID: nil, ElementType: domain.ElementCoreExpertise, SortOrder: 10,
				Config: mustMarshal(CoreExpertiseConfig{
					FontSize: 10.0, Separator: " | ",
					Alignment: "left", SpaceAfter: 0.0,
				})},
			{ID: 36, TemplateID: 2, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 11,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Education", FontSize: 11.0,
					FontStyle: "bold", Uppercase: true,
					Underline: false, UnderlineWeight: 0.0,
					SpaceBefore: 14.0, SpaceAfter: 6.0,
					DataBinding: "academics",
				})},
			{ID: 37, TemplateID: 2, ParentID: nil, ElementType: domain.ElementEducationLoop, SortOrder: 12,
				Config: mustMarshal(EducationLoopConfig{
					EntryGap: 2.0, SpaceBefore: 0.0, SpaceAfter: 0.0,
				})},
			{ID: 38, TemplateID: 2, ParentID: nil, ElementType: domain.ElementSectionHeading, SortOrder: 13,
				Config: mustMarshal(SectionHeadingConfig{
					Text: "Certifications", FontSize: 11.0,
					FontStyle: "bold", Uppercase: true,
					Underline: false, UnderlineWeight: 0.0,
					SpaceBefore: 14.0, SpaceAfter: 6.0,
					DataBinding: "certifications",
				})},
			{ID: 39, TemplateID: 2, ParentID: nil, ElementType: domain.ElementCertsLoop, SortOrder: 14,
				Config: mustMarshal(CertsLoopConfig{
					EntryGap: 1.0, SpaceBefore: 0.0, SpaceAfter: 0.0,
				})},

			// work_history_loop children (parent_id = 31)
			{ID: 40, TemplateID: 2, ParentID: int64Ptr(31), ElementType: domain.ElementWorkTitle, SortOrder: 0,
				Config: mustMarshal(WorkTitleConfig{
					FontSize: 10.0, FontStyle: "bold",
					IncludeEmployer: true, EmployerSeparator: ", ",
					EmployerFontStyle: "bold", SpaceAfter: 15.0,
				})},
			{ID: 41, TemplateID: 2, ParentID: int64Ptr(31), ElementType: domain.ElementWorkDates, SortOrder: 1,
				Config: mustMarshal(WorkDatesConfig{
					FontSize: 9.0, Alignment: "right",
				})},
			{ID: 42, TemplateID: 2, ParentID: int64Ptr(31), ElementType: domain.ElementWorkSummary, SortOrder: 2,
				Config: mustMarshal(WorkSummaryConfig{FontSize: 10.0, FontStyle: "italic"})},
			{ID: 43, TemplateID: 2, ParentID: int64Ptr(31), ElementType: domain.ElementWorkBullets, SortOrder: 3,
				Config: mustMarshal(WorkBulletsConfig{
					FontSize: 10.0, FontStyle: "regular",
					BulletChar: "\u2022", Indent: 12.0,
					BulletSymWidth: 10.0,
				})},
			{ID: 44, TemplateID: 2, ParentID: int64Ptr(31), ElementType: domain.ElementWorkOutcomes, SortOrder: 4,
				Config: mustMarshal(WorkOutcomesConfig{
					FontSize: 10.0, FontStyle: "regular",
					BulletChar: "\u2022", Indent: 12.0,
					BulletSymWidth: 10.0, OutcomesLabel: "Outcomes:",
					OutcomesGap: 2.0,
				})},

			// education_loop children (parent_id = 37)
			{ID: 45, TemplateID: 2, ParentID: int64Ptr(37), ElementType: domain.ElementEduCredential, SortOrder: 0,
				Config: mustMarshal(EduCredentialConfig{
					FontSize: 10.0, FontStyle: "bold",
				})},
			{ID: 46, TemplateID: 2, ParentID: int64Ptr(37), ElementType: domain.ElementEduInstitution, SortOrder: 1,
				Config: mustMarshal(EduInstitutionConfig{
					FontSize: 10.0, FontStyle: "regular",
				})},
			{ID: 47, TemplateID: 2, ParentID: int64Ptr(37), ElementType: domain.ElementEduDate, SortOrder: 2,
				Config: mustMarshal(EduDateConfig{
					FontSize: 9.0, Alignment: "right",
				})},

			// certifications_loop children (parent_id = 39)
			{ID: 48, TemplateID: 2, ParentID: int64Ptr(39), ElementType: domain.ElementCertName, SortOrder: 0,
				Config: mustMarshal(CertNameConfig{
					FontSize: 10.0, FontStyle: "bold",
				})},
			{ID: 49, TemplateID: 2, ParentID: int64Ptr(39), ElementType: domain.ElementCertDetail, SortOrder: 1,
				Config: mustMarshal(CertDetailConfig{
					FontSize: 9.0, FontStyle: "regular",
				})},
		},
	}
}

// int64Ptr is a helper to create an *int64 from a literal.
func int64Ptr(v int64) *int64 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

// =========================================================
// Built-in Formal Cover Letter template
// =========================================================

// FormalCoverLetterTemplate returns the built-in Formal cover letter
// template as a TemplateDetail. Standard business letter format with
// date, recipient address, greeting, body text, closing, and
// signature name.
func FormalCoverLetterTemplate() domain.TemplateDetail {
	return domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           3,
			Name:         "Formal",
			Description:  "Traditional business-style cover letter with structured prompts and formal tone",
			TemplateType: domain.TemplateTypeCoverLetter,
			IsBuiltin:    true,
			MarginTop:    72.0, // 1 inch
			MarginBottom: 72.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		},
		Elements: []domain.TemplateElement{
			{ID: 50, TemplateID: 3, ParentID: nil, ElementType: domain.ElementProfileHeader, SortOrder: 0,
				Config: mustMarshal(ProfileHeaderConfig{
					NameFontSize: 18.0, DetailFontSize: 10.0,
					Alignment: "center", LinkSeparator: " | ",
					ShowLinks: true, ShowLinksInline: false,
					SpaceAfter: 6.0,
				})},
			{ID: 51, TemplateID: 3, ParentID: nil, ElementType: domain.ElementHorizontalRule, SortOrder: 1,
				Config: mustMarshal(HorizontalRuleConfig{
					Weight: 0.5, SpaceBefore: 2.0, SpaceAfter: 14.0,
				})},
			{ID: 52, TemplateID: 3, ParentID: nil, ElementType: domain.ElementDate, SortOrder: 2,
				Config: mustMarshal(DateConfig{
					FontSize: 10.0, Format: "January 2, 2006",
					Alignment: "left", SpaceAfter: 14.0,
				})},
			{ID: 53, TemplateID: 3, ParentID: nil, ElementType: domain.ElementRecipientAddress, SortOrder: 3,
				Config: mustMarshal(RecipientAddressConfig{
					FontSize: 10.0, SpaceAfter: 14.0,
				})},
			{ID: 54, TemplateID: 3, ParentID: nil, ElementType: domain.ElementGreeting, SortOrder: 4,
				Config: mustMarshal(GreetingConfig{
					Text: "Dear {{hiring_manager}},", FontSize: 10.0,
					SpaceAfter: 10.0,
				})},
			{ID: 55, TemplateID: 3, ParentID: nil, ElementType: domain.ElementParagraph, SortOrder: 5,
				Config: mustMarshal(ParagraphConfig{
					FontSize: 10.0, LineSpacing: 1.15, SpaceAfter: 10.0,
					Segments: []ParagraphSegmentConfig{
						{Type: "static", Text: "I am writing to express my interest in the "},
						{Type: "application", Token: "position_title"},
						{Type: "static", Text: " role at "},
						{Type: "application", Token: "company_name"},
						{Type: "static", Text: ". "},
						{Type: "adhoc", Key: "why_fit_formal", Label: "Why are you a strong fit?", HelpText: "Highlight relevant experience and outcomes you can deliver.", Required: true, Multiline: true},
					},
				})},
			{ID: 56, TemplateID: 3, ParentID: nil, ElementType: domain.ElementParagraph, SortOrder: 6,
				Config: mustMarshal(ParagraphConfig{
					FontSize: 10.0, LineSpacing: 1.15, SpaceAfter: 10.0,
					Segments: []ParagraphSegmentConfig{
						{Type: "static", Text: "I am especially interested in "},
						{Type: "application", Token: "company_name"},
						{Type: "static", Text: " because "},
						{Type: "adhoc", Key: "why_company_formal", Label: "Why this company?", HelpText: "Mention a mission, product, or initiative that stands out to you.", Required: true, Multiline: true},
						{Type: "static", Text: "."},
					},
				})},
			{ID: 57, TemplateID: 3, ParentID: nil, ElementType: domain.ElementParagraph, SortOrder: 7,
				Config: mustMarshal(ParagraphConfig{
					FontSize: 10.0, LineSpacing: 1.15, SpaceAfter: 10.0,
					Segments: []ParagraphSegmentConfig{
						{Type: "static", Text: "Thank you for your time and consideration. I would welcome the opportunity to discuss how I can contribute to "},
						{Type: "application", Token: "company_name"},
						{Type: "static", Text: "."},
					},
				})},
			{ID: 58, TemplateID: 3, ParentID: nil, ElementType: domain.ElementClosing, SortOrder: 8,
				Config: mustMarshal(ClosingConfig{
					Text: "Sincerely,", FontSize: 10.0,
					SpaceAfter: 28.0,
				})},
			{ID: 59, TemplateID: 3, ParentID: nil, ElementType: domain.ElementStaticText, SortOrder: 9,
				Config: mustMarshal(StaticTextConfig{
					Text: "{{signer_name}}", FontSize: 10.0,
					FontStyle: "regular", Alignment: "left",
					SpaceBefore: 0.0, SpaceAfter: 0.0,
				})},
		},
	}
}

// =========================================================
// Built-in Casual Cover Letter template
// =========================================================

// CasualCoverLetterTemplate returns the built-in Casual cover letter
// template as a TemplateDetail. A relaxed format with no date or
// recipient address — just a greeting, body, and sign-off.
func CasualCoverLetterTemplate() domain.TemplateDetail {
	return domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           4,
			Name:         "Casual",
			Description:  "Conversational cover letter starter with lightweight prompts and approachable tone",
			TemplateType: domain.TemplateTypeCoverLetter,
			IsBuiltin:    true,
			MarginTop:    72.0,
			MarginBottom: 72.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		},
		Elements: []domain.TemplateElement{
			{ID: 60, TemplateID: 4, ParentID: nil, ElementType: domain.ElementProfileHeader, SortOrder: 0,
				Config: mustMarshal(ProfileHeaderConfig{
					NameFontSize: 18.0, DetailFontSize: 10.0,
					Alignment: "center", LinkSeparator: " | ",
					ShowLinks: true, ShowLinksInline: false,
					SpaceAfter: 6.0,
				})},
			{ID: 61, TemplateID: 4, ParentID: nil, ElementType: domain.ElementGreeting, SortOrder: 1,
				Config: mustMarshal(GreetingConfig{
					Text: "Hi {{hiring_manager}},", FontSize: 10.0,
					SpaceAfter: 10.0,
				})},
			{ID: 62, TemplateID: 4, ParentID: nil, ElementType: domain.ElementParagraph, SortOrder: 2,
				Config: mustMarshal(ParagraphConfig{
					FontSize: 10.0, LineSpacing: 1.15, SpaceAfter: 10.0,
					Segments: []ParagraphSegmentConfig{
						{Type: "static", Text: "I would love to join "},
						{Type: "application", Token: "company_name"},
						{Type: "static", Text: " as a "},
						{Type: "application", Token: "position_title"},
						{Type: "static", Text: ". What stands out to me about your team is "},
						{Type: "adhoc", Key: "why_team_casual", Label: "What stands out about this team?", HelpText: "Keep it natural and specific.", Required: true, Multiline: true},
						{Type: "static", Text: "."},
					},
				})},
			{ID: 63, TemplateID: 4, ParentID: nil, ElementType: domain.ElementParagraph, SortOrder: 3,
				Config: mustMarshal(ParagraphConfig{
					FontSize: 10.0, LineSpacing: 1.15, SpaceAfter: 10.0,
					Segments: []ParagraphSegmentConfig{
						{Type: "static", Text: "A quick example of what I would bring to this role: "},
						{Type: "adhoc", Key: "value_example_casual", Label: "Quick value example", HelpText: "Share one example of impact or ownership.", Required: false, Multiline: true},
					},
				})},
			{ID: 64, TemplateID: 4, ParentID: nil, ElementType: domain.ElementParagraph, SortOrder: 4,
				Config: mustMarshal(ParagraphConfig{
					FontSize: 10.0, LineSpacing: 1.15, SpaceAfter: 10.0,
					Segments: []ParagraphSegmentConfig{
						{Type: "static", Text: "If it is helpful, I can share more details and examples. You can reach me at "},
						{Type: "profile", Token: "email"},
						{Type: "static", Text: "."},
					},
				})},
			{ID: 65, TemplateID: 4, ParentID: nil, ElementType: domain.ElementClosing, SortOrder: 5,
				Config: mustMarshal(ClosingConfig{
					Text: "Thanks,", FontSize: 10.0,
					SpaceAfter: 24.0,
				})},
			{ID: 66, TemplateID: 4, ParentID: nil, ElementType: domain.ElementStaticText, SortOrder: 6,
				Config: mustMarshal(StaticTextConfig{
					Text: "{{signer_name}}", FontSize: 10.0,
					FontStyle: "regular", Alignment: "left",
					SpaceBefore: 0.0, SpaceAfter: 0.0,
				})},
		},
	}
}
