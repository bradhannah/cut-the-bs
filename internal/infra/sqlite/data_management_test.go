package sqlite

import (
	"context"
	"encoding/json"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportAllData_IncludesUserTemplates(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create a user template with nested elements.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "My Custom Resume",
		Description:  "A user-created resume template",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    50.0,
		MarginBottom: 50.0,
		MarginLeft:   60.0,
		MarginRight:  60.0,
	})
	require.NoError(t, err)

	// Add a top-level section heading.
	heading, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config:      `{"text":"Experience","uppercase":true,"font_size":11}`,
	})
	require.NoError(t, err)

	// Add a work history loop.
	loop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config:      `{"entry_gap":8}`,
	})
	require.NoError(t, err)

	// Add children inside the loop.
	child1, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &loop.ID,
		ElementType: domain.ElementWorkTitle,
		Config:      `{"font_size":10,"font_style":"bold"}`,
	})
	require.NoError(t, err)

	child2, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &loop.ID,
		ElementType: domain.ElementWorkBullets,
		Config:      `{"font_size":9,"bullet_char":"•"}`,
	})
	require.NoError(t, err)

	// Export all data.
	data, err := store.ExportAllData(ctx)
	require.NoError(t, err)

	// Templates field should contain only user-created templates, not built-in.
	require.Len(t, data.Templates, 1, "should export exactly 1 user template")

	exported := data.Templates[0]
	assert.Equal(t, tmpl.ID, exported.ID)
	assert.Equal(t, "My Custom Resume", exported.Name)
	assert.Equal(t, "A user-created resume template", exported.Description)
	assert.Equal(t, domain.TemplateTypeResume, exported.TemplateType)
	assert.False(t, exported.IsBuiltin)
	assert.Equal(t, 50.0, exported.MarginTop)
	assert.Equal(t, 60.0, exported.MarginLeft)

	// Should have 4 elements: heading, loop, 2 children.
	require.Len(t, exported.Elements, 4)

	// Verify element types are present.
	typeCount := make(map[string]int)
	for _, el := range exported.Elements {
		typeCount[el.ElementType]++
	}
	assert.Equal(t, 1, typeCount[domain.ElementSectionHeading])
	assert.Equal(t, 1, typeCount[domain.ElementWorkHistoryLoop])
	assert.Equal(t, 1, typeCount[domain.ElementWorkTitle])
	assert.Equal(t, 1, typeCount[domain.ElementWorkBullets])

	// Verify parent relationships are preserved.
	for _, el := range exported.Elements {
		if el.ID == heading.ID {
			assert.Nil(t, el.ParentID, "heading should be top-level")
		}
		if el.ID == loop.ID {
			assert.Nil(t, el.ParentID, "loop should be top-level")
		}
		if el.ID == child1.ID {
			require.NotNil(t, el.ParentID, "work_title should have a parent")
			assert.Equal(t, loop.ID, *el.ParentID)
		}
		if el.ID == child2.ID {
			require.NotNil(t, el.ParentID, "work_bullets should have a parent")
			assert.Equal(t, loop.ID, *el.ParentID)
		}
	}
}

func TestExportAllData_ExcludesBuiltinTemplates(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// No user templates created — only built-ins exist.
	data, err := store.ExportAllData(ctx)
	require.NoError(t, err)

	assert.Empty(t, data.Templates, "should not export built-in templates")
}

func TestImportAllData_RestoresUserTemplates(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create a user template with nested elements.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Imported Resume",
		Description:  "Will survive import",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    48.0,
		MarginBottom: 48.0,
		MarginLeft:   65.0,
		MarginRight:  65.0,
	})
	require.NoError(t, err)

	loop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config:      `{"entry_gap":6}`,
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &loop.ID,
		ElementType: domain.ElementWorkTitle,
		Config:      `{"font_size":10}`,
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &loop.ID,
		ElementType: domain.ElementWorkDates,
		Config:      `{"font_size":9}`,
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config:      `{"text":"Skills"}`,
	})
	require.NoError(t, err)

	// Export the data.
	data, err := store.ExportAllData(ctx)
	require.NoError(t, err)
	require.Len(t, data.Templates, 1)

	// Round-trip through JSON to simulate real backup/restore.
	jsonBytes, err := json.Marshal(data)
	require.NoError(t, err)

	var restored domain.ExportData
	require.NoError(t, json.Unmarshal(jsonBytes, &restored))

	// Import into a fresh database.
	store2 := testStore(t)
	require.NoError(t, Migrate(store2))

	err = store2.ImportAllData(ctx, restored)
	require.NoError(t, err)

	// Verify user template was restored.
	templates, err := store2.ListDocumentTemplates(ctx)
	require.NoError(t, err)

	// Should have 4 built-in + 1 user-created.
	userTemplates := make([]domain.DocumentTemplate, 0)
	for _, t := range templates {
		if !t.IsBuiltin {
			userTemplates = append(userTemplates, t)
		}
	}
	require.Len(t, userTemplates, 1)
	assert.Equal(t, "Imported Resume", userTemplates[0].Name)
	assert.Equal(t, "Will survive import", userTemplates[0].Description)
	assert.Equal(t, domain.TemplateTypeResume, userTemplates[0].TemplateType)
	assert.Equal(t, 48.0, userTemplates[0].MarginTop)
	assert.Equal(t, 65.0, userTemplates[0].MarginLeft)

	// Verify elements were restored.
	detail, err := store2.GetDocumentTemplate(ctx, userTemplates[0].ID)
	require.NoError(t, err)
	require.Len(t, detail.Elements, 4, "should have 4 elements: loop + 2 children + heading")

	// Verify element types.
	typeCount := make(map[string]int)
	for _, el := range detail.Elements {
		typeCount[el.ElementType]++
	}
	assert.Equal(t, 1, typeCount[domain.ElementWorkHistoryLoop])
	assert.Equal(t, 1, typeCount[domain.ElementWorkTitle])
	assert.Equal(t, 1, typeCount[domain.ElementWorkDates])
	assert.Equal(t, 1, typeCount[domain.ElementSectionHeading])

	// Verify parent-child relationships survived the round-trip.
	var loopEl *domain.TemplateElement
	for i, el := range detail.Elements {
		if el.ElementType == domain.ElementWorkHistoryLoop {
			loopEl = &detail.Elements[i]
			break
		}
	}
	require.NotNil(t, loopEl, "should have a work_history_loop element")

	childCount := 0
	for _, el := range detail.Elements {
		if el.ParentID != nil && *el.ParentID == loopEl.ID {
			childCount++
		}
	}
	assert.Equal(t, 2, childCount, "loop should have 2 children after import")

	// Verify built-in templates still exist.
	builtinCount := 0
	for _, t := range templates {
		if t.IsBuiltin {
			builtinCount++
		}
	}
	assert.Equal(t, 4, builtinCount, "built-in templates should survive import")
}

func TestImportAllData_OverwritesExistingUserTemplates(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Create a user template that should be wiped during import.
	_, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Old Template",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Import data with a different user template.
	importData := domain.ExportData{
		SchemaVersion: 7,
		ExportedAt:    "2026-01-01T00:00:00Z",
		Profile: domain.UserProfile{
			ID:       1,
			FullName: "Test User",
		},
		Templates: []domain.TemplateDetail{
			{
				DocumentTemplate: domain.DocumentTemplate{
					ID:           100,
					Name:         "New Template",
					TemplateType: domain.TemplateTypeCoverLetter,
					MarginTop:    36.0,
					MarginBottom: 36.0,
					MarginLeft:   54.0,
					MarginRight:  54.0,
					CreatedAt:    "2026-01-01T00:00:00Z",
					UpdatedAt:    "2026-01-01T00:00:00Z",
				},
				Elements: []domain.TemplateElement{
					{
						ID:          200,
						TemplateID:  100,
						ElementType: domain.ElementGreeting,
						Config:      `{"text":"Dear {{hiring_manager}},"}`,
						SortOrder:   0,
						CreatedAt:   "2026-01-01T00:00:00Z",
						UpdatedAt:   "2026-01-01T00:00:00Z",
					},
				},
			},
		},
	}

	err = store.ImportAllData(ctx, importData)
	require.NoError(t, err)

	// Verify "Old Template" is gone and "New Template" exists.
	templates, err := store.ListDocumentTemplates(ctx)
	require.NoError(t, err)

	userTemplates := make([]domain.DocumentTemplate, 0)
	for _, tmpl := range templates {
		if !tmpl.IsBuiltin {
			userTemplates = append(userTemplates, tmpl)
		}
	}
	require.Len(t, userTemplates, 1)
	assert.Equal(t, "New Template", userTemplates[0].Name)
	assert.Equal(t, domain.TemplateTypeCoverLetter, userTemplates[0].TemplateType)

	// Verify its element was imported.
	detail, err := store.GetDocumentTemplate(ctx, userTemplates[0].ID)
	require.NoError(t, err)
	require.Len(t, detail.Elements, 1)
	assert.Equal(t, domain.ElementGreeting, detail.Elements[0].ElementType)
	assert.Contains(t, detail.Elements[0].Config, "hiring_manager")
}

func TestImportAllData_EmptyTemplates(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Import data with no user templates.
	importData := domain.ExportData{
		SchemaVersion: 7,
		ExportedAt:    "2026-01-01T00:00:00Z",
		Profile: domain.UserProfile{
			ID:       1,
			FullName: "Test User",
		},
		Templates: []domain.TemplateDetail{},
	}

	err := store.ImportAllData(ctx, importData)
	require.NoError(t, err)

	// Only built-in templates should remain.
	templates, err := store.ListDocumentTemplates(ctx)
	require.NoError(t, err)

	for _, tmpl := range templates {
		assert.True(t, tmpl.IsBuiltin, "only built-in templates should exist after import with no user templates")
	}
	assert.Equal(t, 4, len(templates), "should have exactly 4 built-in templates")
}
