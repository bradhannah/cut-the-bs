package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"cut-the-bs/internal/domain"
)

// TemplateStore defines the persistence operations required by
// TemplateService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type TemplateStore interface {
	ListDocumentTemplates(ctx context.Context) ([]domain.DocumentTemplate, error)
	GetDocumentTemplate(ctx context.Context, id int64) (domain.TemplateDetail, error)
	CreateDocumentTemplate(ctx context.Context, input domain.DocumentTemplateInput) (domain.DocumentTemplate, error)
	UpdateDocumentTemplate(ctx context.Context, id int64, input domain.DocumentTemplateInput) (domain.DocumentTemplate, error)
	DeleteDocumentTemplate(ctx context.Context, id int64) error
	DuplicateDocumentTemplate(ctx context.Context, id int64, newName string) (domain.DocumentTemplate, error)
	CreateTemplateElement(ctx context.Context, templateID int64, input domain.TemplateElementInput) (domain.TemplateElement, error)
	UpdateTemplateElement(ctx context.Context, id int64, input domain.TemplateElementInput) (domain.TemplateElement, error)
	DeleteTemplateElement(ctx context.Context, id int64) error
	ReorderTemplateElements(ctx context.Context, templateID int64, parentID *int64, orderedIDs []int64) error
}

// TemplateService provides business-logic operations for managing
// document templates and their elements.
type TemplateService struct {
	store TemplateStore
}

// NewTemplateService creates a TemplateService backed by the given
// store.
func NewTemplateService(store TemplateStore) *TemplateService {
	return &TemplateService{store: store}
}

// ListDocumentTemplates returns all templates (built-in first).
func (s *TemplateService) ListDocumentTemplates(
	ctx context.Context,
) ([]domain.DocumentTemplate, error) {
	return s.store.ListDocumentTemplates(ctx)
}

// GetDocumentTemplate returns a template with all its elements.
func (s *TemplateService) GetDocumentTemplate(
	ctx context.Context,
	id int64,
) (domain.TemplateDetail, error) {
	return s.store.GetDocumentTemplate(ctx, id)
}

// CreateDocumentTemplate validates the input and creates a new
// user template.
func (s *TemplateService) CreateDocumentTemplate(
	ctx context.Context,
	input domain.DocumentTemplateInput,
) (domain.DocumentTemplate, error) {
	if err := validateTemplateInput(input); err != nil {
		return domain.DocumentTemplate{}, err
	}
	return s.store.CreateDocumentTemplate(ctx, input)
}

// UpdateDocumentTemplate validates the input and updates a user
// template's metadata.
func (s *TemplateService) UpdateDocumentTemplate(
	ctx context.Context,
	id int64,
	input domain.DocumentTemplateInput,
) (domain.DocumentTemplate, error) {
	if err := validateTemplateInput(input); err != nil {
		return domain.DocumentTemplate{}, err
	}
	return s.store.UpdateDocumentTemplate(ctx, id, input)
}

// DeleteDocumentTemplate deletes a user template. Built-in
// templates cannot be deleted (enforced by the store).
func (s *TemplateService) DeleteDocumentTemplate(
	ctx context.Context,
	id int64,
) error {
	return s.store.DeleteDocumentTemplate(ctx, id)
}

// DuplicateDocumentTemplate validates the new name and creates a
// copy of a template with all its elements.
func (s *TemplateService) DuplicateDocumentTemplate(
	ctx context.Context,
	id int64,
	newName string,
) (domain.DocumentTemplate, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return domain.DocumentTemplate{}, fmt.Errorf("template name is required")
	}
	if len(newName) > 100 {
		return domain.DocumentTemplate{}, fmt.Errorf("template name must be 100 characters or fewer")
	}
	return s.store.DuplicateDocumentTemplate(ctx, id, newName)
}

// CreateTemplateElement validates the input and adds a new element
// to a template. It fetches the template to check type compatibility
// and validates nesting rules.
func (s *TemplateService) CreateTemplateElement(
	ctx context.Context,
	templateID int64,
	input domain.TemplateElementInput,
) (domain.TemplateElement, error) {
	detail, err := s.store.GetDocumentTemplate(ctx, templateID)
	if err != nil {
		return domain.TemplateElement{}, fmt.Errorf("get template: %w", err)
	}

	if err := validateElementInput(input, detail); err != nil {
		return domain.TemplateElement{}, err
	}

	return s.store.CreateTemplateElement(ctx, templateID, input)
}

// UpdateTemplateElement validates the input and updates an
// element's type and config. Validates element type and JSON
// config. Template-type compatibility is not checked here because
// the store interface does not expose a single-element lookup;
// the frontend enforces compatible types for the template.
func (s *TemplateService) UpdateTemplateElement(
	ctx context.Context,
	id int64,
	input domain.TemplateElementInput,
) (domain.TemplateElement, error) {
	if !domain.IsValidElementType(input.ElementType) {
		return domain.TemplateElement{}, fmt.Errorf("invalid element type: %q", input.ElementType)
	}

	if err := validateConfigJSON(input.Config); err != nil {
		return domain.TemplateElement{}, err
	}

	return s.store.UpdateTemplateElement(ctx, id, input)
}

// DeleteTemplateElement removes an element from a template.
func (s *TemplateService) DeleteTemplateElement(
	ctx context.Context,
	id int64,
) error {
	return s.store.DeleteTemplateElement(ctx, id)
}

// ReorderTemplateElements updates the sort order for elements
// within a parent scope.
func (s *TemplateService) ReorderTemplateElements(
	ctx context.Context,
	templateID int64,
	parentID *int64,
	orderedIDs []int64,
) error {
	return s.store.ReorderTemplateElements(ctx, templateID, parentID, orderedIDs)
}

// validateTemplateInput checks all template input validation rules.
func validateTemplateInput(input domain.DocumentTemplateInput) error {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return fmt.Errorf("template name is required")
	}
	if len(name) > 100 {
		return fmt.Errorf("template name must be 100 characters or fewer")
	}

	if !domain.IsValidTemplateType(input.TemplateType) {
		return fmt.Errorf("invalid template type: %q (must be %q or %q)",
			input.TemplateType, domain.TemplateTypeResume, domain.TemplateTypeCoverLetter)
	}

	if err := validateMargin("margin_top", input.MarginTop); err != nil {
		return err
	}
	if err := validateMargin("margin_bottom", input.MarginBottom); err != nil {
		return err
	}
	if err := validateMargin("margin_left", input.MarginLeft); err != nil {
		return err
	}
	if err := validateMargin("margin_right", input.MarginRight); err != nil {
		return err
	}

	return nil
}

// validateMargin checks that a margin value is within the valid
// range of 0 to 288 points (0 to 4 inches).
func validateMargin(field string, value float64) error {
	if value < 0 || value > 288 {
		return fmt.Errorf("%s must be between 0 and 288 points (got %.1f)", field, value)
	}
	return nil
}

// validateElementInput checks all element input validation rules
// against the parent template.
func validateElementInput(input domain.TemplateElementInput, detail domain.TemplateDetail) error {
	if !domain.IsValidElementType(input.ElementType) {
		return fmt.Errorf("invalid element type: %q", input.ElementType)
	}

	if !domain.IsElementTypeCompatible(detail.TemplateType, input.ElementType) {
		return fmt.Errorf("element type %q is not compatible with %s templates",
			input.ElementType, detail.TemplateType)
	}

	if err := validateConfigJSON(input.Config); err != nil {
		return err
	}

	// Validate nesting rules.
	if input.ParentID != nil {
		parentID := *input.ParentID

		// Find the parent element in the template's elements.
		var parent *domain.TemplateElement
		for i := range detail.Elements {
			if detail.Elements[i].ID == parentID {
				parent = &detail.Elements[i]
				break
			}
		}

		if parent == nil {
			return fmt.Errorf("parent element %d not found in template", parentID)
		}

		// Parent must be a loop container.
		if !domain.IsLoopElementType(parent.ElementType) {
			return fmt.Errorf("parent element %d (type %q) is not a loop container",
				parentID, parent.ElementType)
		}

		// Enforce single-level nesting: the parent's own ParentID
		// must be nil (it must be a top-level element).
		if parent.ParentID != nil {
			return fmt.Errorf("nested loops are not allowed: parent element %d is already a child element", parentID)
		}

		// The child element type must be valid for this loop type.
		if !domain.IsValidLoopChild(parent.ElementType, input.ElementType) {
			return fmt.Errorf("element type %q is not a valid child of %q",
				input.ElementType, parent.ElementType)
		}
	} else {
		// Top-level elements: loop child types should not be placed
		// at the top level (they only make sense inside their loop).
		// However, we don't enforce this as a hard rule — the spec
		// only enforces that loop children are valid within their
		// parent. Utility types (spacer, heading, etc.) are valid
		// both at top-level and within loops.
	}

	return nil
}

// validateConfigJSON checks that the config string is valid JSON.
// An empty string is treated as "{}".
func validateConfigJSON(config string) error {
	if config == "" || config == "{}" {
		return nil
	}
	if !json.Valid([]byte(config)) {
		return fmt.Errorf("config must be valid JSON")
	}
	return nil
}

// =========================================================
// Template Variable Parsing (T050)
// =========================================================

// variablePattern matches {{variable_name}} placeholders.
// Captures: group 1 = full content between braces.
var variablePattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// ParseTemplateVariables scans all elements in a cover letter template
// for {{variable_name}} and {{prompt: descriptive text}} patterns.
// Variables are deduplicated by name. Prompts are returned in order.
func (s *TemplateService) ParseTemplateVariables(detail domain.TemplateDetail) domain.TemplateVariables {
	result := domain.TemplateVariables{
		Variables: make([]domain.TemplateVariable, 0),
		Prompts:   make([]domain.GuidedPrompt, 0),
	}
	seenVars := make(map[string]bool)

	for _, el := range detail.Elements {
		text := extractTextFromConfig(el.Config)
		if text == "" {
			continue
		}

		matches := variablePattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			inner := strings.TrimSpace(match[1])
			if inner == "" {
				continue
			}

			if strings.HasPrefix(inner, "prompt:") || strings.HasPrefix(inner, "prompt: ") {
				promptText := strings.TrimSpace(strings.TrimPrefix(inner, "prompt:"))
				if promptText != "" {
					result.Prompts = append(result.Prompts, domain.GuidedPrompt{
						PromptText: promptText,
						Source:     el.ElementType,
					})
				}
			} else {
				if !seenVars[inner] {
					seenVars[inner] = true
					result.Variables = append(result.Variables, domain.TemplateVariable{
						Name:   inner,
						Source: el.ElementType,
					})
				}
			}
		}
	}

	return result
}

// ApplySubstitutions replaces all {{variable_name}} and {{prompt: text}}
// placeholders in the given text with values from the substitution map.
// Missing variables are replaced with empty strings.
func ApplySubstitutions(text string, subs map[string]string) string {
	if subs == nil || text == "" {
		return text
	}
	return variablePattern.ReplaceAllStringFunc(text, func(match string) string {
		inner := strings.TrimSpace(match[2 : len(match)-2])
		// Check for prompt: prefix
		if strings.HasPrefix(inner, "prompt:") || strings.HasPrefix(inner, "prompt: ") {
			promptText := strings.TrimSpace(strings.TrimPrefix(inner, "prompt:"))
			key := "prompt:" + promptText
			if val, ok := subs[key]; ok {
				return val
			}
			return "" // unresolved prompt
		}
		// Regular variable
		if val, ok := subs[inner]; ok {
			return val
		}
		return "" // unresolved variable
	})
}

// extractTextFromConfig attempts to extract the "text" field from a
// JSON config string. Returns empty string if parsing fails or no
// text field exists.
func extractTextFromConfig(config string) string {
	if config == "" || config == "{}" {
		return ""
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(config), &parsed); err != nil {
		return ""
	}
	if text, ok := parsed["text"]; ok {
		if s, ok := text.(string); ok {
			return s
		}
	}
	return ""
}
