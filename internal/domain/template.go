package domain

// Template type constants.
const (
	TemplateTypeResume      = "resume"
	TemplateTypeCoverLetter = "cover_letter"
)

// Resume element type constants.
const (
	ElementProfileHeader   = "profile_header"
	ElementRoleDescriptors = "role_descriptors"
	ElementProfSummary     = "professional_summary"
	ElementWorkHistoryLoop = "work_history_loop"
	ElementWorkTitle       = "work_title"
	ElementWorkEmployer    = "work_employer"
	ElementWorkDates       = "work_dates"
	ElementWorkSummary     = "work_summary"
	ElementWorkBullets     = "work_bullets"
	ElementWorkOutcomes    = "work_outcomes"
	ElementSkills          = "skills"
	ElementEducationLoop   = "education_loop"
	ElementEduCredential   = "edu_credential"
	ElementEduInstitution  = "edu_institution"
	ElementEduDate         = "edu_date"
	ElementCertsLoop       = "certifications_loop"
	ElementCertName        = "cert_name"
	ElementCertDetail      = "cert_detail"
	ElementCoreExpertise   = "core_expertise"
	ElementSectionHeading  = "section_heading"
	ElementHorizontalRule  = "horizontal_rule"
	ElementSpacer          = "spacer"
	ElementStaticText      = "static_text"
)

// Cover letter element type constants.
const (
	ElementBodyText         = "body_text"
	ElementDate             = "date"
	ElementGreeting         = "greeting"
	ElementClosing          = "closing"
	ElementRecipientAddress = "recipient_address"
)

// LoopElementTypes lists all element types that can contain children.
var LoopElementTypes = []string{
	ElementWorkHistoryLoop,
	ElementEducationLoop,
	ElementCertsLoop,
}

// ValidLoopChildren maps each loop container type to its valid child
// element types.
var ValidLoopChildren = map[string][]string{
	ElementWorkHistoryLoop: {
		ElementWorkTitle, ElementWorkEmployer, ElementWorkDates,
		ElementWorkSummary, ElementWorkBullets, ElementWorkOutcomes,
		ElementSectionHeading, ElementHorizontalRule, ElementSpacer, ElementStaticText,
	},
	ElementEducationLoop: {
		ElementEduCredential, ElementEduInstitution, ElementEduDate,
		ElementSectionHeading, ElementHorizontalRule, ElementSpacer, ElementStaticText,
	},
	ElementCertsLoop: {
		ElementCertName, ElementCertDetail,
		ElementSectionHeading, ElementHorizontalRule, ElementSpacer, ElementStaticText,
	},
}

// resumeElementTypes lists all element types valid for resume templates.
var resumeElementTypes = map[string]bool{
	ElementProfileHeader:   true,
	ElementRoleDescriptors: true,
	ElementProfSummary:     true,
	ElementWorkHistoryLoop: true,
	ElementWorkTitle:       true,
	ElementWorkEmployer:    true,
	ElementWorkDates:       true,
	ElementWorkSummary:     true,
	ElementWorkBullets:     true,
	ElementWorkOutcomes:    true,
	ElementSkills:          true,
	ElementEducationLoop:   true,
	ElementEduCredential:   true,
	ElementEduInstitution:  true,
	ElementEduDate:         true,
	ElementCertsLoop:       true,
	ElementCertName:        true,
	ElementCertDetail:      true,
	ElementCoreExpertise:   true,
	ElementSectionHeading:  true,
	ElementHorizontalRule:  true,
	ElementSpacer:          true,
	ElementStaticText:      true,
}

// coverLetterElementTypes lists all element types valid for cover letter templates.
var coverLetterElementTypes = map[string]bool{
	ElementProfileHeader:    true,
	ElementBodyText:         true,
	ElementDate:             true,
	ElementGreeting:         true,
	ElementClosing:          true,
	ElementRecipientAddress: true,
	ElementSectionHeading:   true,
	ElementHorizontalRule:   true,
	ElementSpacer:           true,
	ElementStaticText:       true,
}

// allElementTypes is the union of resume and cover letter element types.
var allElementTypes map[string]bool

func init() {
	allElementTypes = make(map[string]bool)
	for k := range resumeElementTypes {
		allElementTypes[k] = true
	}
	for k := range coverLetterElementTypes {
		allElementTypes[k] = true
	}
}

// IsValidElementType returns true if the element type is recognized.
func IsValidElementType(elementType string) bool {
	return allElementTypes[elementType]
}

// IsElementTypeCompatible returns true if the element type is valid
// for the given template type.
func IsElementTypeCompatible(templateType, elementType string) bool {
	switch templateType {
	case TemplateTypeResume:
		return resumeElementTypes[elementType]
	case TemplateTypeCoverLetter:
		return coverLetterElementTypes[elementType]
	default:
		return false
	}
}

// IsLoopElementType returns true if the element type is a loop
// container that can have children.
func IsLoopElementType(elementType string) bool {
	for _, t := range LoopElementTypes {
		if t == elementType {
			return true
		}
	}
	return false
}

// IsValidLoopChild returns true if childType is a valid child of
// parentType.
func IsValidLoopChild(parentType, childType string) bool {
	children, ok := ValidLoopChildren[parentType]
	if !ok {
		return false
	}
	for _, c := range children {
		if c == childType {
			return true
		}
	}
	return false
}

// IsValidTemplateType returns true if the template type is
// recognized ("resume" or "cover_letter").
func IsValidTemplateType(templateType string) bool {
	return templateType == TemplateTypeResume || templateType == TemplateTypeCoverLetter
}

// DocumentTemplate is a named, typed document layout configuration
// that defines how a resume or cover letter is rendered to PDF.
// Built-in templates are seeded during migration and cannot be
// deleted by users.
type DocumentTemplate struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	TemplateType string  `json:"template_type"`
	IsBuiltin    bool    `json:"is_builtin"`
	MarginTop    float64 `json:"margin_top"`
	MarginBottom float64 `json:"margin_bottom"`
	MarginLeft   float64 `json:"margin_left"`
	MarginRight  float64 `json:"margin_right"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// DocumentTemplateInput is the input type for creating or updating
// a document template.
type DocumentTemplateInput struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	TemplateType string  `json:"template_type"`
	MarginTop    float64 `json:"margin_top"`
	MarginBottom float64 `json:"margin_bottom"`
	MarginLeft   float64 `json:"margin_left"`
	MarginRight  float64 `json:"margin_right"`
}

// TemplateElement is a single block within a template layout. Each
// element has a type that determines its rendering behavior and a
// JSON config blob containing type-specific properties. Elements
// may be top-level or children of a loop container (single-level
// nesting only; no nested loops per spec clarification).
type TemplateElement struct {
	ID          int64  `json:"id"`
	TemplateID  int64  `json:"template_id"`
	ParentID    *int64 `json:"parent_id"`
	ElementType string `json:"element_type"`
	Config      string `json:"config"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TemplateElementInput is the input type for creating or updating
// a template element.
type TemplateElementInput struct {
	ParentID    *int64 `json:"parent_id"`
	ElementType string `json:"element_type"`
	Config      string `json:"config"`
}

// TemplateDetail is a template with all its elements included,
// organized as a flat list ordered by sort_order.
type TemplateDetail struct {
	DocumentTemplate
	Elements []TemplateElement `json:"elements"`
}
