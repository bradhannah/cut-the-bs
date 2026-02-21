package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDocumentTemplates_ReturnsBuiltinFirst(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	templates, err := store.ListDocumentTemplates(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(templates), 2)

	// Built-in templates should come first.
	assert.True(t, templates[0].IsBuiltin)
	assert.True(t, templates[1].IsBuiltin)
}

func TestCreateDocumentTemplate(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	input := domain.DocumentTemplateInput{
		Name:         "My Template",
		Description:  "A test template",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	}

	tmpl, err := store.CreateDocumentTemplate(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, "My Template", tmpl.Name)
	assert.Equal(t, "A test template", tmpl.Description)
	assert.Equal(t, domain.TemplateTypeResume, tmpl.TemplateType)
	assert.False(t, tmpl.IsBuiltin)
	assert.Equal(t, 54.0, tmpl.MarginTop)
	assert.Equal(t, 72.0, tmpl.MarginLeft)
	assert.NotEmpty(t, tmpl.CreatedAt)
	assert.NotEmpty(t, tmpl.UpdatedAt)
}

func TestGetDocumentTemplate_WithElements(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Get the seeded Professional template.
	detail, err := store.GetDocumentTemplate(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Professional", detail.Name)
	assert.True(t, detail.IsBuiltin)
	assert.Greater(t, len(detail.Elements), 0)

	// Verify elements are ordered (top-level first).
	hasTopLevel := false
	hasChild := false
	for _, e := range detail.Elements {
		if e.ParentID == nil {
			hasTopLevel = true
		} else {
			hasChild = true
		}
	}
	assert.True(t, hasTopLevel)
	assert.True(t, hasChild)
}

func TestGetDocumentTemplate_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.GetDocumentTemplate(ctx, 9999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateDocumentTemplate(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create a user template.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Original",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Update it.
	updated, err := store.UpdateDocumentTemplate(ctx, tmpl.ID, domain.DocumentTemplateInput{
		Name:         "Updated",
		Description:  "New description",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    36.0,
		MarginBottom: 36.0,
		MarginLeft:   54.0,
		MarginRight:  54.0,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, "New description", updated.Description)
	assert.Equal(t, 36.0, updated.MarginTop)
}

func TestUpdateDocumentTemplate_BuiltinRejected(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateDocumentTemplate(ctx, 1, domain.DocumentTemplateInput{
		Name:         "Renamed",
		TemplateType: domain.TemplateTypeResume,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot update built-in")
}

func TestDeleteDocumentTemplate(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create and delete a user template.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "ToDelete",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	err = store.DeleteDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)

	// Should be gone.
	_, err = store.GetDocumentTemplate(ctx, tmpl.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteDocumentTemplate_BuiltinRejected(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteDocumentTemplate(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete built-in")
}

func TestDeleteDocumentTemplate_CascadesElements(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create template with elements.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "WithElements",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config:      `{"text":"TEST"}`,
	})
	require.NoError(t, err)

	// Delete the template.
	err = store.DeleteDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)

	// Elements should be gone.
	var count int
	err = store.DB().QueryRow(
		"SELECT count(*) FROM template_element WHERE template_id = ?", tmpl.ID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDuplicateDocumentTemplate(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Duplicate the built-in Professional template.
	dup, err := store.DuplicateDocumentTemplate(ctx, 1, "Professional Copy")
	require.NoError(t, err)
	assert.Equal(t, "Professional Copy", dup.Name)
	assert.False(t, dup.IsBuiltin, "duplicate should not be built-in")

	// Get the duplicate with elements.
	dupDetail, err := store.GetDocumentTemplate(ctx, dup.ID)
	require.NoError(t, err)

	// Get the source with elements.
	srcDetail, err := store.GetDocumentTemplate(ctx, 1)
	require.NoError(t, err)

	// Element count should match.
	assert.Equal(t, len(srcDetail.Elements), len(dupDetail.Elements),
		"duplicated template should have same number of elements")

	// Verify element types match.
	srcTypes := make(map[string]int)
	dupTypes := make(map[string]int)
	for _, e := range srcDetail.Elements {
		srcTypes[e.ElementType]++
	}
	for _, e := range dupDetail.Elements {
		dupTypes[e.ElementType]++
	}
	assert.Equal(t, srcTypes, dupTypes, "element type distribution should match")

	// Verify IDs are different (new elements, not shared).
	srcIDs := make(map[int64]bool)
	for _, e := range srcDetail.Elements {
		srcIDs[e.ID] = true
	}
	for _, e := range dupDetail.Elements {
		assert.False(t, srcIDs[e.ID], "duplicate elements should have new IDs")
	}
}

func TestDuplicateDocumentTemplate_PreservesParentRelationships(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Duplicate Professional (has work_history_loop with children).
	dup, err := store.DuplicateDocumentTemplate(ctx, 1, "Copy")
	require.NoError(t, err)

	dupDetail, err := store.GetDocumentTemplate(ctx, dup.ID)
	require.NoError(t, err)

	// Find the work_history_loop element.
	var loopID int64
	for _, e := range dupDetail.Elements {
		if e.ElementType == domain.ElementWorkHistoryLoop && e.ParentID == nil {
			loopID = e.ID
			break
		}
	}
	require.NotZero(t, loopID, "should have a work_history_loop element")

	// Verify children reference the new loop's ID (not the source's).
	childCount := 0
	for _, e := range dupDetail.Elements {
		if e.ParentID != nil && *e.ParentID == loopID {
			childCount++
		}
	}
	assert.Greater(t, childCount, 0, "loop should have children in duplicated template")
}

func TestCreateTemplateElement(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create a user template.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Add a section heading.
	elem, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config:      `{"text":"TEST HEADING"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, tmpl.ID, elem.TemplateID)
	assert.Nil(t, elem.ParentID)
	assert.Equal(t, domain.ElementSectionHeading, elem.ElementType)
	assert.Equal(t, `{"text":"TEST HEADING"}`, elem.Config)
	assert.Equal(t, 0, elem.SortOrder)

	// Add another — sort_order should increment.
	elem2, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      `{"height":10}`,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, elem2.SortOrder)
}

func TestCreateTemplateElement_WithParent(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Create a loop container.
	loop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config:      `{"entry_gap":4}`,
	})
	require.NoError(t, err)

	// Create a child inside the loop.
	child, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &loop.ID,
		ElementType: domain.ElementWorkTitle,
		Config:      `{"font_size":10}`,
	})
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, loop.ID, *child.ParentID)
	assert.Equal(t, 0, child.SortOrder)
}

func TestCreateTemplateElement_EmptyConfig(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Empty config should default to "{}".
	elem, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
	})
	require.NoError(t, err)
	assert.Equal(t, "{}", elem.Config)
}

func TestUpdateTemplateElement(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	elem, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config:      `{"text":"OLD"}`,
	})
	require.NoError(t, err)

	updated, err := store.UpdateTemplateElement(ctx, elem.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config:      `{"text":"NEW"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, `{"text":"NEW"}`, updated.Config)
}

func TestUpdateTemplateElement_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateTemplateElement(ctx, 99999, domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      "{}",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteTemplateElement(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	elem, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      "{}",
	})
	require.NoError(t, err)

	err = store.DeleteTemplateElement(ctx, elem.ID)
	require.NoError(t, err)

	// Should be gone.
	_, err = store.getTemplateElement(ctx, elem.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteTemplateElement_CascadesChildren(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	loop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config:      "{}",
	})
	require.NoError(t, err)

	child, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &loop.ID,
		ElementType: domain.ElementWorkTitle,
		Config:      "{}",
	})
	require.NoError(t, err)

	// Delete the loop — child should cascade.
	err = store.DeleteTemplateElement(ctx, loop.ID)
	require.NoError(t, err)

	_, err = store.getTemplateElement(ctx, child.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestReorderTemplateElements_TopLevel(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	e1, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config:      `{"text":"First"}`,
	})
	require.NoError(t, err)

	e2, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSpacer,
		Config:      `{"height":10}`,
	})
	require.NoError(t, err)

	e3, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementHorizontalRule,
		Config:      `{"weight":0.5}`,
	})
	require.NoError(t, err)

	// Reorder: reverse the order.
	err = store.ReorderTemplateElements(ctx, tmpl.ID, nil, []int64{e3.ID, e2.ID, e1.ID})
	require.NoError(t, err)

	// Verify new order.
	detail, err := store.GetDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)

	topLevel := make([]domain.TemplateElement, 0)
	for _, e := range detail.Elements {
		if e.ParentID == nil {
			topLevel = append(topLevel, e)
		}
	}
	require.Len(t, topLevel, 3)
	assert.Equal(t, e3.ID, topLevel[0].ID)
	assert.Equal(t, e2.ID, topLevel[1].ID)
	assert.Equal(t, e1.ID, topLevel[2].ID)
}

func TestReorderTemplateElements_WithinLoop(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	loop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config:      "{}",
	})
	require.NoError(t, err)

	c1, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &loop.ID,
		ElementType: domain.ElementWorkTitle,
		Config:      "{}",
	})
	require.NoError(t, err)

	c2, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &loop.ID,
		ElementType: domain.ElementWorkDates,
		Config:      "{}",
	})
	require.NoError(t, err)

	// Reorder children within the loop.
	err = store.ReorderTemplateElements(ctx, tmpl.ID, &loop.ID, []int64{c2.ID, c1.ID})
	require.NoError(t, err)

	// Verify.
	detail, err := store.GetDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)

	children := make([]domain.TemplateElement, 0)
	for _, e := range detail.Elements {
		if e.ParentID != nil && *e.ParentID == loop.ID {
			children = append(children, e)
		}
	}
	require.Len(t, children, 2)
	assert.Equal(t, c2.ID, children[0].ID)
	assert.Equal(t, c1.ID, children[1].ID)
}

func TestReorderTemplateElements_InvalidID(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Reorder with a non-existent element ID should fail.
	err = store.ReorderTemplateElements(ctx, tmpl.ID, nil, []int64{99999})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in scope")
}

func TestListDocumentTemplates_IncludesUserCreated(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create a user template.
	_, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Custom Template",
		TemplateType: domain.TemplateTypeCoverLetter,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	templates, err := store.ListDocumentTemplates(ctx)
	require.NoError(t, err)

	// Should have 2 built-in + 1 user-created.
	assert.Len(t, templates, 3)

	// First two should be built-in, last should be user-created.
	assert.True(t, templates[0].IsBuiltin)
	assert.True(t, templates[1].IsBuiltin)
	assert.False(t, templates[2].IsBuiltin)
	assert.Equal(t, "Custom Template", templates[2].Name)
}
