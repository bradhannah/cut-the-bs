package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =========================================================
// Mock template store
// =========================================================

type mockTemplateStore struct {
	// Return values
	templates []domain.DocumentTemplate
	detail    domain.TemplateDetail
	template  domain.DocumentTemplate
	element   domain.TemplateElement
	err       error

	// Call tracking
	listCalls            int
	getCalls             []int64
	createCalls          []domain.DocumentTemplateInput
	updateCalls          []updateTemplateCall
	deleteCalls          []int64
	duplicateCalls       []duplicateCall
	createElementCalls   []createElementCall
	updateElementCalls   []updateElementCall
	deleteElementCalls   []int64
	reorderElementsCalls []reorderElementsCall
}

type updateTemplateCall struct {
	ID    int64
	Input domain.DocumentTemplateInput
}

type duplicateCall struct {
	ID      int64
	NewName string
}

type createElementCall struct {
	TemplateID int64
	Input      domain.TemplateElementInput
}

type updateElementCall struct {
	ID    int64
	Input domain.TemplateElementInput
}

type reorderElementsCall struct {
	TemplateID int64
	ParentID   *int64
	OrderedIDs []int64
}

func (m *mockTemplateStore) ListDocumentTemplates(_ context.Context) ([]domain.DocumentTemplate, error) {
	m.listCalls++
	return m.templates, m.err
}

func (m *mockTemplateStore) GetDocumentTemplate(_ context.Context, id int64) (domain.TemplateDetail, error) {
	m.getCalls = append(m.getCalls, id)
	if m.err != nil {
		return domain.TemplateDetail{}, m.err
	}
	return m.detail, nil
}

func (m *mockTemplateStore) CreateDocumentTemplate(_ context.Context, input domain.DocumentTemplateInput) (domain.DocumentTemplate, error) {
	m.createCalls = append(m.createCalls, input)
	if m.err != nil {
		return domain.DocumentTemplate{}, m.err
	}
	return m.template, nil
}

func (m *mockTemplateStore) UpdateDocumentTemplate(_ context.Context, id int64, input domain.DocumentTemplateInput) (domain.DocumentTemplate, error) {
	m.updateCalls = append(m.updateCalls, updateTemplateCall{ID: id, Input: input})
	if m.err != nil {
		return domain.DocumentTemplate{}, m.err
	}
	return m.template, nil
}

func (m *mockTemplateStore) DeleteDocumentTemplate(_ context.Context, id int64) error {
	m.deleteCalls = append(m.deleteCalls, id)
	return m.err
}

func (m *mockTemplateStore) DuplicateDocumentTemplate(_ context.Context, id int64, newName string) (domain.DocumentTemplate, error) {
	m.duplicateCalls = append(m.duplicateCalls, duplicateCall{ID: id, NewName: newName})
	if m.err != nil {
		return domain.DocumentTemplate{}, m.err
	}
	return m.template, nil
}

func (m *mockTemplateStore) CreateTemplateElement(_ context.Context, templateID int64, input domain.TemplateElementInput) (domain.TemplateElement, error) {
	m.createElementCalls = append(m.createElementCalls, createElementCall{TemplateID: templateID, Input: input})
	if m.err != nil {
		return domain.TemplateElement{}, m.err
	}
	return m.element, nil
}

func (m *mockTemplateStore) UpdateTemplateElement(_ context.Context, id int64, input domain.TemplateElementInput) (domain.TemplateElement, error) {
	m.updateElementCalls = append(m.updateElementCalls, updateElementCall{ID: id, Input: input})
	if m.err != nil {
		return domain.TemplateElement{}, m.err
	}
	return m.element, nil
}

func (m *mockTemplateStore) DeleteTemplateElement(_ context.Context, id int64) error {
	m.deleteElementCalls = append(m.deleteElementCalls, id)
	return m.err
}

func (m *mockTemplateStore) ReorderTemplateElements(_ context.Context, templateID int64, parentID *int64, orderedIDs []int64) error {
	m.reorderElementsCalls = append(m.reorderElementsCalls, reorderElementsCall{
		TemplateID: templateID,
		ParentID:   parentID,
		OrderedIDs: orderedIDs,
	})
	return m.err
}

// =========================================================
// Helpers
// =========================================================

func defaultTemplateStore() *mockTemplateStore {
	return &mockTemplateStore{
		templates: []domain.DocumentTemplate{
			{ID: 1, Name: "Professional", TemplateType: "resume", IsBuiltin: true},
			{ID: 2, Name: "My Template", TemplateType: "resume", IsBuiltin: false},
		},
		detail: domain.TemplateDetail{
			DocumentTemplate: domain.DocumentTemplate{
				ID: 1, Name: "Professional", TemplateType: "resume", IsBuiltin: true,
				MarginTop: 54, MarginBottom: 54, MarginLeft: 72, MarginRight: 72,
			},
			Elements: []domain.TemplateElement{
				{ID: 10, TemplateID: 1, ElementType: domain.ElementProfileHeader, SortOrder: 0},
				{ID: 11, TemplateID: 1, ElementType: domain.ElementWorkHistoryLoop, SortOrder: 1},
				{ID: 12, TemplateID: 1, ParentID: int64Ptr(11), ElementType: domain.ElementWorkTitle, SortOrder: 0},
			},
		},
		template: domain.DocumentTemplate{
			ID: 3, Name: "New Template", TemplateType: "resume",
			MarginTop: 54, MarginBottom: 54, MarginLeft: 72, MarginRight: 72,
		},
		element: domain.TemplateElement{
			ID: 20, TemplateID: 1, ElementType: domain.ElementSpacer, Config: `{"height":10}`, SortOrder: 2,
		},
	}
}

func validTemplateInput() domain.DocumentTemplateInput {
	return domain.DocumentTemplateInput{
		Name:         "My Resume",
		Description:  "A custom resume template",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54,
		MarginBottom: 54,
		MarginLeft:   72,
		MarginRight:  72,
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

// =========================================================
// ListDocumentTemplates
// =========================================================

func TestTemplateService_ListDocumentTemplates(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	templates, err := svc.ListDocumentTemplates(context.Background())
	require.NoError(t, err)
	require.Len(t, templates, 2)
	assert.Equal(t, "Professional", templates[0].Name)
	assert.Equal(t, 1, store.listCalls)
}

func TestTemplateService_ListDocumentTemplates_StoreError(t *testing.T) {
	store := defaultTemplateStore()
	store.err = fmt.Errorf("db error")
	svc := NewTemplateService(store)

	_, err := svc.ListDocumentTemplates(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// =========================================================
// GetDocumentTemplate
// =========================================================

func TestTemplateService_GetDocumentTemplate(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	detail, err := svc.GetDocumentTemplate(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Professional", detail.Name)
	require.Len(t, detail.Elements, 3)
	require.Len(t, store.getCalls, 1)
	assert.Equal(t, int64(1), store.getCalls[0])
}

// =========================================================
// CreateDocumentTemplate — validation
// =========================================================

func TestTemplateService_CreateDocumentTemplate_Valid(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	result, err := svc.CreateDocumentTemplate(context.Background(), validTemplateInput())
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.ID)
	require.Len(t, store.createCalls, 1)
}

func TestTemplateService_CreateDocumentTemplate_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*domain.DocumentTemplateInput)
		wantError string
	}{
		{
			name:      "empty name",
			modify:    func(i *domain.DocumentTemplateInput) { i.Name = "" },
			wantError: "template name is required",
		},
		{
			name:      "whitespace-only name",
			modify:    func(i *domain.DocumentTemplateInput) { i.Name = "   " },
			wantError: "template name is required",
		},
		{
			name:      "name too long",
			modify:    func(i *domain.DocumentTemplateInput) { i.Name = strings.Repeat("a", 101) },
			wantError: "100 characters or fewer",
		},
		{
			name:      "name exactly 100 chars is valid",
			modify:    func(i *domain.DocumentTemplateInput) { i.Name = strings.Repeat("a", 100) },
			wantError: "", // no error
		},
		{
			name:      "invalid template type",
			modify:    func(i *domain.DocumentTemplateInput) { i.TemplateType = "letter" },
			wantError: "invalid template type",
		},
		{
			name:      "empty template type",
			modify:    func(i *domain.DocumentTemplateInput) { i.TemplateType = "" },
			wantError: "invalid template type",
		},
		{
			name:      "margin_top negative",
			modify:    func(i *domain.DocumentTemplateInput) { i.MarginTop = -1 },
			wantError: "margin_top must be between 0 and 288",
		},
		{
			name:      "margin_top too large",
			modify:    func(i *domain.DocumentTemplateInput) { i.MarginTop = 289 },
			wantError: "margin_top must be between 0 and 288",
		},
		{
			name:      "margin_top at boundary 0 is valid",
			modify:    func(i *domain.DocumentTemplateInput) { i.MarginTop = 0 },
			wantError: "",
		},
		{
			name:      "margin_top at boundary 288 is valid",
			modify:    func(i *domain.DocumentTemplateInput) { i.MarginTop = 288 },
			wantError: "",
		},
		{
			name:      "margin_bottom negative",
			modify:    func(i *domain.DocumentTemplateInput) { i.MarginBottom = -0.5 },
			wantError: "margin_bottom must be between 0 and 288",
		},
		{
			name:      "margin_left too large",
			modify:    func(i *domain.DocumentTemplateInput) { i.MarginLeft = 500 },
			wantError: "margin_left must be between 0 and 288",
		},
		{
			name:      "margin_right negative",
			modify:    func(i *domain.DocumentTemplateInput) { i.MarginRight = -10 },
			wantError: "margin_right must be between 0 and 288",
		},
		{
			name:      "cover_letter type is valid",
			modify:    func(i *domain.DocumentTemplateInput) { i.TemplateType = domain.TemplateTypeCoverLetter },
			wantError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := defaultTemplateStore()
			svc := NewTemplateService(store)

			input := validTemplateInput()
			tc.modify(&input)

			_, err := svc.CreateDocumentTemplate(context.Background(), input)
			if tc.wantError == "" {
				require.NoError(t, err)
				require.Len(t, store.createCalls, 1)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError)
				assert.Empty(t, store.createCalls, "store should not be called on validation error")
			}
		})
	}
}

// =========================================================
// UpdateDocumentTemplate — validation
// =========================================================

func TestTemplateService_UpdateDocumentTemplate_Valid(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	result, err := svc.UpdateDocumentTemplate(context.Background(), 2, validTemplateInput())
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.ID) // returns mock template
	require.Len(t, store.updateCalls, 1)
	assert.Equal(t, int64(2), store.updateCalls[0].ID)
}

func TestTemplateService_UpdateDocumentTemplate_EmptyName(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	input := validTemplateInput()
	input.Name = ""

	_, err := svc.UpdateDocumentTemplate(context.Background(), 2, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template name is required")
	assert.Empty(t, store.updateCalls)
}

func TestTemplateService_UpdateDocumentTemplate_InvalidMargin(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	input := validTemplateInput()
	input.MarginTop = 300

	_, err := svc.UpdateDocumentTemplate(context.Background(), 2, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "margin_top")
	assert.Empty(t, store.updateCalls)
}

// =========================================================
// DeleteDocumentTemplate
// =========================================================

func TestTemplateService_DeleteDocumentTemplate(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	err := svc.DeleteDocumentTemplate(context.Background(), 2)
	require.NoError(t, err)
	require.Len(t, store.deleteCalls, 1)
	assert.Equal(t, int64(2), store.deleteCalls[0])
}

func TestTemplateService_DeleteDocumentTemplate_StoreError(t *testing.T) {
	store := defaultTemplateStore()
	store.err = fmt.Errorf("cannot delete built-in template")
	svc := NewTemplateService(store)

	err := svc.DeleteDocumentTemplate(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "built-in")
}

// =========================================================
// DuplicateDocumentTemplate — name validation
// =========================================================

func TestTemplateService_DuplicateDocumentTemplate_Valid(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	result, err := svc.DuplicateDocumentTemplate(context.Background(), 1, "Professional Copy")
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.ID)
	require.Len(t, store.duplicateCalls, 1)
	assert.Equal(t, int64(1), store.duplicateCalls[0].ID)
	assert.Equal(t, "Professional Copy", store.duplicateCalls[0].NewName)
}

func TestTemplateService_DuplicateDocumentTemplate_EmptyName(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	_, err := svc.DuplicateDocumentTemplate(context.Background(), 1, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template name is required")
	assert.Empty(t, store.duplicateCalls)
}

func TestTemplateService_DuplicateDocumentTemplate_WhitespaceName(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	_, err := svc.DuplicateDocumentTemplate(context.Background(), 1, "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template name is required")
	assert.Empty(t, store.duplicateCalls)
}

func TestTemplateService_DuplicateDocumentTemplate_NameTooLong(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	_, err := svc.DuplicateDocumentTemplate(context.Background(), 1, strings.Repeat("x", 101))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "100 characters or fewer")
	assert.Empty(t, store.duplicateCalls)
}

func TestTemplateService_DuplicateDocumentTemplate_TrimsWhitespace(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	_, err := svc.DuplicateDocumentTemplate(context.Background(), 1, "  My Copy  ")
	require.NoError(t, err)
	require.Len(t, store.duplicateCalls, 1)
	assert.Equal(t, "My Copy", store.duplicateCalls[0].NewName)
}

// =========================================================
// CreateTemplateElement — validation
// =========================================================

func TestTemplateService_CreateTemplateElement_ValidTopLevel(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	input := domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      `{"height": 10}`,
	}

	result, err := svc.CreateTemplateElement(context.Background(), 1, input)
	require.NoError(t, err)
	assert.Equal(t, int64(20), result.ID)
	require.Len(t, store.createElementCalls, 1)
	assert.Equal(t, int64(1), store.createElementCalls[0].TemplateID)
}

func TestTemplateService_CreateTemplateElement_ValidLoopChild(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	parentID := int64(11) // work_history_loop
	input := domain.TemplateElementInput{
		ParentID:    &parentID,
		ElementType: domain.ElementWorkTitle,
		Config:      `{}`,
	}

	_, err := svc.CreateTemplateElement(context.Background(), 1, input)
	require.NoError(t, err)
	require.Len(t, store.createElementCalls, 1)
}

func TestTemplateService_CreateTemplateElement_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     domain.TemplateElementInput
		wantError string
	}{
		{
			name: "invalid element type",
			input: domain.TemplateElementInput{
				ElementType: "nonexistent_type",
				Config:      `{}`,
			},
			wantError: "invalid element type",
		},
		{
			name: "incompatible element type for resume template",
			input: domain.TemplateElementInput{
				ElementType: domain.ElementBodyText, // cover letter only
				Config:      `{}`,
			},
			wantError: "not compatible with resume templates",
		},
		{
			name: "invalid JSON config",
			input: domain.TemplateElementInput{
				ElementType: domain.ElementSpacer,
				Config:      `{invalid json`,
			},
			wantError: "config must be valid JSON",
		},
		{
			name: "parent not found",
			input: domain.TemplateElementInput{
				ParentID:    int64Ptr(999),
				ElementType: domain.ElementWorkTitle,
				Config:      `{}`,
			},
			wantError: "parent element 999 not found",
		},
		{
			name: "parent is not a loop container",
			input: domain.TemplateElementInput{
				ParentID:    int64Ptr(10), // profile_header, not a loop
				ElementType: domain.ElementWorkTitle,
				Config:      `{}`,
			},
			wantError: "is not a loop container",
		},
		{
			name: "nested loop rejected (parent is already a child)",
			input: domain.TemplateElementInput{
				ParentID:    int64Ptr(12), // work_title has parent_id=11, is a child
				ElementType: domain.ElementSpacer,
				Config:      `{}`,
			},
			wantError: "is not a loop container",
		},
		{
			name: "invalid child type for loop",
			input: domain.TemplateElementInput{
				ParentID:    int64Ptr(11), // work_history_loop
				ElementType: domain.ElementEduCredential,
				Config:      `{}`,
			},
			wantError: "not a valid child of",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := defaultTemplateStore()
			svc := NewTemplateService(store)

			_, err := svc.CreateTemplateElement(context.Background(), 1, tc.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantError)
			assert.Empty(t, store.createElementCalls, "store should not be called on validation error")
		})
	}
}

func TestTemplateService_CreateTemplateElement_EmptyConfigOK(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	input := domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      "",
	}

	_, err := svc.CreateTemplateElement(context.Background(), 1, input)
	require.NoError(t, err)
	require.Len(t, store.createElementCalls, 1)
}

func TestTemplateService_CreateTemplateElement_StoreError(t *testing.T) {
	store := defaultTemplateStore()
	store.err = fmt.Errorf("template not found")
	svc := NewTemplateService(store)

	input := domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      `{}`,
	}

	_, err := svc.CreateTemplateElement(context.Background(), 999, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

// =========================================================
// CreateTemplateElement — cover letter compatibility
// =========================================================

func TestTemplateService_CreateTemplateElement_CoverLetterTypes(t *testing.T) {
	store := defaultTemplateStore()
	// Override detail to be a cover letter template.
	store.detail = domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID: 3, Name: "Formal", TemplateType: domain.TemplateTypeCoverLetter,
		},
		Elements: make([]domain.TemplateElement, 0),
	}
	svc := NewTemplateService(store)

	// body_text is valid for cover letter.
	input := domain.TemplateElementInput{
		ElementType: domain.ElementBodyText,
		Config:      `{}`,
	}
	_, err := svc.CreateTemplateElement(context.Background(), 3, input)
	require.NoError(t, err)
}

func TestTemplateService_CreateTemplateElement_CoverLetterRejectsResume(t *testing.T) {
	store := defaultTemplateStore()
	store.detail = domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID: 3, Name: "Formal", TemplateType: domain.TemplateTypeCoverLetter,
		},
		Elements: make([]domain.TemplateElement, 0),
	}
	svc := NewTemplateService(store)

	// work_history_loop is resume-only, should be rejected for cover letter.
	input := domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config:      `{}`,
	}
	_, err := svc.CreateTemplateElement(context.Background(), 3, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not compatible with cover_letter")
}

// =========================================================
// UpdateTemplateElement — validation
// =========================================================

func TestTemplateService_UpdateTemplateElement_Valid(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	input := domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      `{"height": 20}`,
	}

	result, err := svc.UpdateTemplateElement(context.Background(), 10, input)
	require.NoError(t, err)
	assert.Equal(t, int64(20), result.ID)
	require.Len(t, store.updateElementCalls, 1)
	assert.Equal(t, int64(10), store.updateElementCalls[0].ID)
}

func TestTemplateService_UpdateTemplateElement_InvalidType(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	input := domain.TemplateElementInput{
		ElementType: "invalid_type",
		Config:      `{}`,
	}

	_, err := svc.UpdateTemplateElement(context.Background(), 10, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid element type")
	assert.Empty(t, store.updateElementCalls)
}

func TestTemplateService_UpdateTemplateElement_InvalidJSON(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	input := domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      `not json`,
	}

	_, err := svc.UpdateTemplateElement(context.Background(), 10, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config must be valid JSON")
	assert.Empty(t, store.updateElementCalls)
}

func TestTemplateService_UpdateTemplateElement_EmptyConfigOK(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	input := domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      "",
	}

	_, err := svc.UpdateTemplateElement(context.Background(), 10, input)
	require.NoError(t, err)
	require.Len(t, store.updateElementCalls, 1)
}

// =========================================================
// DeleteTemplateElement
// =========================================================

func TestTemplateService_DeleteTemplateElement(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	err := svc.DeleteTemplateElement(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, store.deleteElementCalls, 1)
	assert.Equal(t, int64(10), store.deleteElementCalls[0])
}

// =========================================================
// ReorderTemplateElements
// =========================================================

func TestTemplateService_ReorderTemplateElements_TopLevel(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	err := svc.ReorderTemplateElements(context.Background(), 1, nil, []int64{11, 10})
	require.NoError(t, err)
	require.Len(t, store.reorderElementsCalls, 1)
	call := store.reorderElementsCalls[0]
	assert.Equal(t, int64(1), call.TemplateID)
	assert.Nil(t, call.ParentID)
	assert.Equal(t, []int64{11, 10}, call.OrderedIDs)
}

func TestTemplateService_ReorderTemplateElements_WithinLoop(t *testing.T) {
	store := defaultTemplateStore()
	svc := NewTemplateService(store)

	parentID := int64(11)
	err := svc.ReorderTemplateElements(context.Background(), 1, &parentID, []int64{12})
	require.NoError(t, err)
	require.Len(t, store.reorderElementsCalls, 1)
	assert.Equal(t, int64(11), *store.reorderElementsCalls[0].ParentID)
}

// =========================================================
// validateConfigJSON — edge cases
// =========================================================

func TestValidateConfigJSON(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{"empty string", "", false},
		{"empty object", "{}", false},
		{"valid object", `{"font_size": 12}`, false},
		{"valid array", `[1, 2, 3]`, false},
		{"valid nested", `{"a": {"b": 1}}`, false},
		{"invalid JSON", `{invalid`, true},
		{"truncated", `{"key":`, true},
		{"plain text", `hello`, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfigJSON(tc.config)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// =========================================================
// validateTemplateInput — comprehensive
// =========================================================

func TestValidateTemplateInput_AllMarginsChecked(t *testing.T) {
	// Verify each margin field is independently validated.
	margins := []struct {
		name   string
		modify func(*domain.DocumentTemplateInput)
	}{
		{"margin_top", func(i *domain.DocumentTemplateInput) { i.MarginTop = -1 }},
		{"margin_bottom", func(i *domain.DocumentTemplateInput) { i.MarginBottom = -1 }},
		{"margin_left", func(i *domain.DocumentTemplateInput) { i.MarginLeft = -1 }},
		{"margin_right", func(i *domain.DocumentTemplateInput) { i.MarginRight = -1 }},
	}

	for _, tc := range margins {
		t.Run(tc.name, func(t *testing.T) {
			input := validTemplateInput()
			tc.modify(&input)
			err := validateTemplateInput(input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.name)
		})
	}
}

// =========================================================
// Nesting validation — nested loop rejection
// =========================================================

func TestTemplateService_CreateTemplateElement_NestedLoopRejected(t *testing.T) {
	store := defaultTemplateStore()
	// Set up a detail where the work_history_loop (ID 11) is top-level,
	// and we try to add another loop as its child.
	svc := NewTemplateService(store)

	parentID := int64(11) // work_history_loop
	input := domain.TemplateElementInput{
		ParentID:    &parentID,
		ElementType: domain.ElementEducationLoop, // trying to nest a loop
		Config:      `{}`,
	}

	_, err := svc.CreateTemplateElement(context.Background(), 1, input)
	require.Error(t, err)
	// education_loop is not a valid child of work_history_loop
	assert.Contains(t, err.Error(), "not a valid child")
}

func TestTemplateService_CreateTemplateElement_DeepNestingRejected(t *testing.T) {
	store := defaultTemplateStore()
	// Element 12 (work_title) has parent_id=11, so it's already nested.
	// Trying to use it as a parent should fail because it's not a loop.
	svc := NewTemplateService(store)

	childOfChild := int64(12) // work_title, which is already a child
	input := domain.TemplateElementInput{
		ParentID:    &childOfChild,
		ElementType: domain.ElementSpacer,
		Config:      `{}`,
	}

	_, err := svc.CreateTemplateElement(context.Background(), 1, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a loop container")
}

// =========================================================
// T047 — Variable Parsing Tests
// =========================================================

func TestParseTemplateVariables_BasicVariables(t *testing.T) {
	detail := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           1,
			TemplateType: domain.TemplateTypeCoverLetter,
		},
		Elements: []domain.TemplateElement{
			{
				ID:          1,
				ElementType: domain.ElementGreeting,
				Config:      `{"text": "Dear {{company_name}} Team,", "font_size": 11, "space_after": 12}`,
			},
			{
				ID:          2,
				ElementType: domain.ElementBodyText,
				Config:      `{"font_size": 11, "line_spacing": 1.15, "space_after": 12}`,
			},
		},
	}

	svc := NewTemplateService(nil)
	result := svc.ParseTemplateVariables(detail)

	require.Len(t, result.Variables, 1)
	assert.Equal(t, "company_name", result.Variables[0].Name)
	assert.Equal(t, domain.ElementGreeting, result.Variables[0].Source)
	assert.Empty(t, result.Prompts)
}

func TestParseTemplateVariables_GuidedPrompts(t *testing.T) {
	detail := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           1,
			TemplateType: domain.TemplateTypeCoverLetter,
		},
		Elements: []domain.TemplateElement{
			{
				ID:          1,
				ElementType: domain.ElementBodyText,
				Config:      `{"text": "I am drawn to {{company_name}} because {{prompt: Why are you interested in this company?}}", "font_size": 11}`,
			},
		},
	}

	svc := NewTemplateService(nil)
	result := svc.ParseTemplateVariables(detail)

	require.Len(t, result.Variables, 1)
	assert.Equal(t, "company_name", result.Variables[0].Name)

	require.Len(t, result.Prompts, 1)
	assert.Equal(t, "Why are you interested in this company?", result.Prompts[0].PromptText)
	assert.Equal(t, domain.ElementBodyText, result.Prompts[0].Source)
}

func TestParseTemplateVariables_MultipleVariables(t *testing.T) {
	detail := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           1,
			TemplateType: domain.TemplateTypeCoverLetter,
		},
		Elements: []domain.TemplateElement{
			{
				ID:          1,
				ElementType: domain.ElementGreeting,
				Config:      `{"text": "Dear {{company_name}} Hiring Team,"}`,
			},
			{
				ID:          2,
				ElementType: domain.ElementBodyText,
				Config:      `{"text": "I am applying for {{position_title}} at {{company_name}}."}`,
			},
			{
				ID:          3,
				ElementType: domain.ElementBodyText,
				Config:      `{"text": "{{prompt: Describe your relevant experience}} {{prompt: Why this company?}}"}`,
			},
		},
	}

	svc := NewTemplateService(nil)
	result := svc.ParseTemplateVariables(detail)

	// company_name appears twice but should be deduplicated
	require.Len(t, result.Variables, 2)
	names := []string{result.Variables[0].Name, result.Variables[1].Name}
	assert.Contains(t, names, "company_name")
	assert.Contains(t, names, "position_title")

	require.Len(t, result.Prompts, 2)
	promptTexts := []string{result.Prompts[0].PromptText, result.Prompts[1].PromptText}
	assert.Contains(t, promptTexts, "Describe your relevant experience")
	assert.Contains(t, promptTexts, "Why this company?")
}

func TestParseTemplateVariables_NoVariables(t *testing.T) {
	detail := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           1,
			TemplateType: domain.TemplateTypeCoverLetter,
		},
		Elements: []domain.TemplateElement{
			{
				ID:          1,
				ElementType: domain.ElementBodyText,
				Config:      `{"text": "No variables here.", "font_size": 11}`,
			},
		},
	}

	svc := NewTemplateService(nil)
	result := svc.ParseTemplateVariables(detail)

	assert.Empty(t, result.Variables)
	assert.Empty(t, result.Prompts)
}

func TestParseTemplateVariables_InvalidConfig(t *testing.T) {
	detail := domain.TemplateDetail{
		DocumentTemplate: domain.DocumentTemplate{
			ID:           1,
			TemplateType: domain.TemplateTypeCoverLetter,
		},
		Elements: []domain.TemplateElement{
			{
				ID:          1,
				ElementType: domain.ElementBodyText,
				Config:      `invalid json`,
			},
		},
	}

	svc := NewTemplateService(nil)
	result := svc.ParseTemplateVariables(detail)

	// Should not panic; returns empty results
	assert.Empty(t, result.Variables)
	assert.Empty(t, result.Prompts)
}
