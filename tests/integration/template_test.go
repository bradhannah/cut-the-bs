package integration

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"cut-the-bs/internal/domain"
	"cut-the-bs/internal/infra/pdf"
	"cut-the-bs/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemplate_UpdateElementConfig (T040) verifies that updating an
// element's config via UpdateTemplateElement changes the resulting
// PDF output. Creates a template with a section heading using
// uppercase=true, exports, then updates to uppercase=false, exports
// again, and verifies the heading text differs between the two PDFs.
func TestTemplate_UpdateElementConfig(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Create a template.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Config Update Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Add profile header.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfileHeader,
		Config: mustJSON(pdf.ProfileHeaderConfig{
			NameFontSize: 18.0, DetailFontSize: 10.0,
			Alignment: "center", SpaceAfter: 6.0,
		}),
	})
	require.NoError(t, err)

	// Add a section heading with uppercase=true.
	heading, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Professional Summary", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0,
			DataBinding: "summaries",
		}),
	})
	require.NoError(t, err)

	// Add summary element.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfSummary,
		Config: mustJSON(pdf.ProfSummaryConfig{
			FontSize: 10.0, BulletChar: "\u2022",
		}),
	})
	require.NoError(t, err)

	// Seed data.
	_, err = store.GetProfile(ctx)
	require.NoError(t, err)
	profile, err := store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Config Test User",
		Email:    "config@test.com",
	})
	require.NoError(t, err)

	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "General",
		BodyText: "A versatile engineer with broad experience.",
	})
	require.NoError(t, err)

	// ── First render: heading should be "PROFESSIONAL SUMMARY" (uppercase) ──
	detail1, err := store.GetDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)

	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	path1, err := renderer.RenderResume(ctx, domain.RenderResumeRequest{
		Template:  &detail1,
		OutputDir: outputDir,
		Profile:   profile,
		Summaries: []domain.ProfessionalSummary{summary},
	})
	require.NoError(t, err)

	text1 := extractPDFText(t, path1)
	assert.Contains(t, text1, "PROFESSIONAL SUMMARY",
		"first render should have uppercase heading")

	// ── Update the heading: turn off uppercase, change font_size, remove underline ──
	_, err = store.UpdateTemplateElement(ctx, heading.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Professional Summary", FontSize: 14.0, FontStyle: "bold",
			Uppercase: false, Underline: false, UnderlineWeight: 0.0,
			SpaceBefore: 10.0, SpaceAfter: 4.0,
			DataBinding: "summaries",
		}),
	})
	require.NoError(t, err)

	// Verify the update persisted.
	detail2, err := store.GetDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)

	// Find the heading element in the updated template.
	var updatedHeading *domain.TemplateElement
	for i := range detail2.Elements {
		if detail2.Elements[i].ID == heading.ID {
			updatedHeading = &detail2.Elements[i]
			break
		}
	}
	require.NotNil(t, updatedHeading, "heading element should still exist")

	// Parse and verify the updated config.
	var headingCfg pdf.SectionHeadingConfig
	require.NoError(t, json.Unmarshal([]byte(updatedHeading.Config), &headingCfg))
	assert.False(t, headingCfg.Uppercase, "uppercase should be false after update")
	assert.Equal(t, 14.0, headingCfg.FontSize, "font_size should be 14 after update")
	assert.False(t, headingCfg.Underline, "underline should be false after update")

	// ── Second render: heading should be "Professional Summary" (not uppercase) ──
	path2, err := renderer.RenderResume(ctx, domain.RenderResumeRequest{
		Template:  &detail2,
		OutputDir: outputDir,
		Profile:   profile,
		Summaries: []domain.ProfessionalSummary{summary},
	})
	require.NoError(t, err)

	text2 := extractPDFText(t, path2)
	assert.Contains(t, text2, "Professional Summary",
		"second render should have mixed-case heading")
	assert.NotContains(t, text2, "PROFESSIONAL SUMMARY",
		"second render should NOT have uppercase heading")

	// Both should still contain the summary content.
	assert.Contains(t, text1, "broad experience",
		"first render should have summary text")
	assert.Contains(t, text2, "broad experience",
		"second render should have summary text")
}

// TestTemplate_ManagementOperations (T062) verifies CRUD operations
// for template management: create, list, rename/update, duplicate,
// delete, and built-in protection.
func TestTemplate_ManagementOperations(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	t.Run("list includes seeded built-in templates", func(t *testing.T) {
		templates, err := store.ListDocumentTemplates(ctx)
		require.NoError(t, err)
		// Migration v7 seeds 4 built-in templates: Professional,
		// Modern (resume), Formal, Casual (cover letter).
		assert.GreaterOrEqual(t, len(templates), 4,
			"should have at least 4 built-in templates")

		builtinCount := 0
		for _, tmpl := range templates {
			if tmpl.IsBuiltin {
				builtinCount++
			}
		}
		assert.Equal(t, 4, builtinCount,
			"should have exactly 4 built-in templates")
	})

	t.Run("create user template", func(t *testing.T) {
		tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
			Name:         "My Custom Resume",
			Description:  "A custom template for testing",
			TemplateType: domain.TemplateTypeResume,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		})
		require.NoError(t, err)
		assert.NotZero(t, tmpl.ID)
		assert.Equal(t, "My Custom Resume", tmpl.Name)
		assert.False(t, tmpl.IsBuiltin,
			"user-created template should not be built-in")
	})

	t.Run("update user template name and description", func(t *testing.T) {
		tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
			Name:         "Before Rename",
			TemplateType: domain.TemplateTypeResume,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		})
		require.NoError(t, err)

		updated, err := store.UpdateDocumentTemplate(ctx, tmpl.ID, domain.DocumentTemplateInput{
			Name:         "After Rename",
			Description:  "Updated description",
			TemplateType: domain.TemplateTypeResume,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		})
		require.NoError(t, err)
		assert.Equal(t, "After Rename", updated.Name)
		assert.Equal(t, "Updated description", updated.Description)
	})

	t.Run("cannot update built-in template", func(t *testing.T) {
		templates, err := store.ListDocumentTemplates(ctx)
		require.NoError(t, err)

		var builtinID int64
		for _, tmpl := range templates {
			if tmpl.IsBuiltin {
				builtinID = tmpl.ID
				break
			}
		}
		require.NotZero(t, builtinID, "should find a built-in template")

		_, err = store.UpdateDocumentTemplate(ctx, builtinID, domain.DocumentTemplateInput{
			Name:         "Hacked Name",
			TemplateType: domain.TemplateTypeResume,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		})
		require.Error(t, err,
			"should not allow updating built-in template")
	})

	t.Run("cannot delete built-in template", func(t *testing.T) {
		templates, err := store.ListDocumentTemplates(ctx)
		require.NoError(t, err)

		var builtinID int64
		for _, tmpl := range templates {
			if tmpl.IsBuiltin {
				builtinID = tmpl.ID
				break
			}
		}
		require.NotZero(t, builtinID)

		err = store.DeleteDocumentTemplate(ctx, builtinID)
		require.Error(t, err,
			"should not allow deleting built-in template")
	})

	t.Run("delete user template", func(t *testing.T) {
		tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
			Name:         "To Be Deleted",
			TemplateType: domain.TemplateTypeResume,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		})
		require.NoError(t, err)

		// Add an element so we verify cascade delete.
		_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
			ElementType: domain.ElementProfileHeader,
			Config: mustJSON(pdf.ProfileHeaderConfig{
				NameFontSize: 18.0, DetailFontSize: 10.0,
				Alignment: "center", SpaceAfter: 6.0,
			}),
		})
		require.NoError(t, err)

		err = store.DeleteDocumentTemplate(ctx, tmpl.ID)
		require.NoError(t, err)

		// Verify template is gone.
		_, err = store.GetDocumentTemplate(ctx, tmpl.ID)
		require.Error(t, err, "should not find deleted template")
	})

	t.Run("duplicate user-created template", func(t *testing.T) {
		original, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
			Name:         "Original Template",
			Description:  "The original",
			TemplateType: domain.TemplateTypeResume,
			MarginTop:    54.0,
			MarginBottom: 54.0,
			MarginLeft:   72.0,
			MarginRight:  72.0,
		})
		require.NoError(t, err)

		// Add elements to verify they are duplicated.
		_, err = store.CreateTemplateElement(ctx, original.ID, domain.TemplateElementInput{
			ElementType: domain.ElementProfileHeader,
			Config: mustJSON(pdf.ProfileHeaderConfig{
				NameFontSize: 18.0, DetailFontSize: 10.0,
				Alignment: "center", SpaceAfter: 6.0,
			}),
		})
		require.NoError(t, err)

		_, err = store.CreateTemplateElement(ctx, original.ID, domain.TemplateElementInput{
			ElementType: domain.ElementSectionHeading,
			Config: mustJSON(pdf.SectionHeadingConfig{
				Text: "Summary", FontSize: 12.0, FontStyle: "bold",
				Uppercase: true, Underline: true, UnderlineWeight: 0.5,
				SpaceBefore: 10.0, SpaceAfter: 4.0,
			}),
		})
		require.NoError(t, err)

		copy, err := store.DuplicateDocumentTemplate(ctx, original.ID, "Copy of Original")
		require.NoError(t, err)
		assert.NotEqual(t, original.ID, copy.ID,
			"copy should have different ID")
		assert.Equal(t, "Copy of Original", copy.Name)
		assert.False(t, copy.IsBuiltin,
			"copy should never be built-in")

		// Verify elements were duplicated.
		copyDetail, err := store.GetDocumentTemplate(ctx, copy.ID)
		require.NoError(t, err)
		assert.Len(t, copyDetail.Elements, 2,
			"copy should have same number of elements")
	})

	t.Run("duplicate built-in template", func(t *testing.T) {
		templates, err := store.ListDocumentTemplates(ctx)
		require.NoError(t, err)

		var builtinID int64
		var builtinName string
		for _, tmpl := range templates {
			if tmpl.IsBuiltin && tmpl.TemplateType == domain.TemplateTypeResume {
				builtinID = tmpl.ID
				builtinName = tmpl.Name
				break
			}
		}
		require.NotZero(t, builtinID)

		copy, err := store.DuplicateDocumentTemplate(ctx, builtinID, builtinName+" Copy")
		require.NoError(t, err)
		assert.False(t, copy.IsBuiltin,
			"duplicate of built-in should not be built-in")
		assert.Equal(t, builtinName+" Copy", copy.Name)

		// Verify the copy has elements from the built-in template.
		copyDetail, err := store.GetDocumentTemplate(ctx, copy.ID)
		require.NoError(t, err)
		assert.Greater(t, len(copyDetail.Elements), 0,
			"copy of built-in should have elements")
	})

	t.Run("list reflects all operations", func(t *testing.T) {
		templates, err := store.ListDocumentTemplates(ctx)
		require.NoError(t, err)

		// Should have built-in + all user-created templates
		// (minus deleted ones).
		assert.Greater(t, len(templates), 4,
			"should have more than just built-ins after creating templates")

		// Built-in templates should appear first.
		foundFirstNonBuiltin := false
		for _, tmpl := range templates {
			if !tmpl.IsBuiltin {
				foundFirstNonBuiltin = true
			}
			if foundFirstNonBuiltin && tmpl.IsBuiltin {
				t.Error("built-in template appeared after non-built-in")
			}
		}
	})
}

// mustJSON marshals v to a JSON string, panicking on error.
// Used to build element Config values in tests.
func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("mustJSON: " + err.Error())
	}
	return string(b)
}

// TestTemplate_BuildFromScratchAndExport exercises the full
// round-trip for US1: create a custom template, add elements
// (including loop children), reorder top-level elements, then
// render a resume PDF and verify expected sections appear in the
// correct order.
func TestTemplate_BuildFromScratchAndExport(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// ──────────────────────────────────────────────────────
	// Step 1: Create a custom resume template.
	// ──────────────────────────────────────────────────────
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Custom Integration Test",
		Description:  "Template built from scratch for integration testing",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)
	assert.NotZero(t, tmpl.ID)
	assert.Equal(t, "Custom Integration Test", tmpl.Name)
	assert.False(t, tmpl.IsBuiltin, "user-created template should not be built-in")

	// ──────────────────────────────────────────────────────
	// Step 2: Add top-level elements in initial order.
	//   0: profile_header
	//   1: section_heading ("Summary")
	//   2: professional_summary
	//   3: section_heading ("Experience")
	//   4: work_history_loop
	//   5: section_heading ("Skills")
	//   6: skills
	//   7: section_heading ("Education")
	//   8: education_loop
	// ──────────────────────────────────────────────────────
	header, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfileHeader,
		Config: mustJSON(pdf.ProfileHeaderConfig{
			NameFontSize: 18.0, DetailFontSize: 10.0,
			Alignment: "center", LinkSeparator: " | ",
			ShowLinks: true, SpaceAfter: 6.0,
		}),
	})
	require.NoError(t, err)

	summaryHeading, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Summary", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0,
			DataBinding: "summaries",
		}),
	})
	require.NoError(t, err)

	summaryEl, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfSummary,
		Config: mustJSON(pdf.ProfSummaryConfig{
			FontSize: 10.0, BulletChar: "\u2022",
		}),
	})
	require.NoError(t, err)

	expHeading, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Experience", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0,
			DataBinding: "work_history",
		}),
	})
	require.NoError(t, err)

	workLoop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config: mustJSON(pdf.WorkHistoryLoopConfig{
			EntryGap: 4.0,
		}),
	})
	require.NoError(t, err)

	skillsHeading, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Skills", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0,
			DataBinding: "skills",
		}),
	})
	require.NoError(t, err)

	skillsEl, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSkills,
		Config: mustJSON(pdf.SkillsConfig{
			FontSize: 10.0, GroupByCategory: true,
			IncludeLegacy: false, CategoryFontStyle: "bold",
			SkillSeparator: ", ",
		}),
	})
	require.NoError(t, err)

	eduHeading, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Education", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0,
			DataBinding: "academics",
		}),
	})
	require.NoError(t, err)

	eduLoop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementEducationLoop,
		Config: mustJSON(pdf.EducationLoopConfig{
			EntryGap: 0.0,
		}),
	})
	require.NoError(t, err)

	// ──────────────────────────────────────────────────────
	// Step 3: Add children to loop containers.
	// ──────────────────────────────────────────────────────

	// Work history loop children.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &workLoop.ID,
		ElementType: domain.ElementWorkTitle,
		Config: mustJSON(pdf.WorkTitleConfig{
			FontSize: 10.0, FontStyle: "bold",
			IncludeEmployer: true, EmployerSeparator: " \u2014 ",
			EmployerFontStyle: "italic", SpaceAfter: 13.0,
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &workLoop.ID,
		ElementType: domain.ElementWorkDates,
		Config: mustJSON(pdf.WorkDatesConfig{
			FontSize: 9.0, Alignment: "right",
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &workLoop.ID,
		ElementType: domain.ElementWorkBullets,
		Config: mustJSON(pdf.WorkBulletsConfig{
			FontSize: 10.0, FontStyle: "regular",
			BulletChar: "\u2022", Indent: 12.0, BulletSymWidth: 10.0,
		}),
	})
	require.NoError(t, err)

	// Education loop children.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &eduLoop.ID,
		ElementType: domain.ElementEduCredential,
		Config: mustJSON(pdf.EduCredentialConfig{
			FontSize: 10.0, FontStyle: "bold",
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &eduLoop.ID,
		ElementType: domain.ElementEduInstitution,
		Config: mustJSON(pdf.EduInstitutionConfig{
			FontSize: 10.0, FontStyle: "regular",
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &eduLoop.ID,
		ElementType: domain.ElementEduDate,
		Config: mustJSON(pdf.EduDateConfig{
			FontSize: 9.0, Alignment: "right",
		}),
	})
	require.NoError(t, err)

	// ──────────────────────────────────────────────────────
	// Step 4: Reorder top-level elements.
	// Move Experience before Summary to verify reorder
	// affects PDF output order.
	//
	// New order:
	//   0: profile_header
	//   1: section_heading ("Experience")
	//   2: work_history_loop
	//   3: section_heading ("Summary")
	//   4: professional_summary
	//   5: section_heading ("Skills")
	//   6: skills
	//   7: section_heading ("Education")
	//   8: education_loop
	// ──────────────────────────────────────────────────────
	err = store.ReorderTemplateElements(ctx, tmpl.ID, nil, []int64{
		header.ID,
		expHeading.ID, workLoop.ID,
		summaryHeading.ID, summaryEl.ID,
		skillsHeading.ID, skillsEl.ID,
		eduHeading.ID, eduLoop.ID,
	})
	require.NoError(t, err)

	// ──────────────────────────────────────────────────────
	// Step 5: Load the completed template and verify structure.
	// ──────────────────────────────────────────────────────
	detail, err := store.GetDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, tmpl.ID, detail.ID)
	// 9 top-level + 3 work children + 3 edu children = 15 elements.
	assert.Len(t, detail.Elements, 15)

	// Verify reorder took effect for top-level elements.
	var topLevel []domain.TemplateElement
	for _, el := range detail.Elements {
		if el.ParentID == nil {
			topLevel = append(topLevel, el)
		}
	}
	require.Len(t, topLevel, 9)
	assert.Equal(t, domain.ElementProfileHeader, topLevel[0].ElementType)
	assert.Equal(t, domain.ElementSectionHeading, topLevel[1].ElementType) // Experience
	assert.Equal(t, domain.ElementWorkHistoryLoop, topLevel[2].ElementType)
	assert.Equal(t, domain.ElementSectionHeading, topLevel[3].ElementType) // Summary
	assert.Equal(t, domain.ElementProfSummary, topLevel[4].ElementType)
	assert.Equal(t, domain.ElementSectionHeading, topLevel[5].ElementType) // Skills
	assert.Equal(t, domain.ElementSkills, topLevel[6].ElementType)
	assert.Equal(t, domain.ElementSectionHeading, topLevel[7].ElementType) // Education
	assert.Equal(t, domain.ElementEducationLoop, topLevel[8].ElementType)

	// ──────────────────────────────────────────────────────
	// Step 6: Seed resume data.
	// ──────────────────────────────────────────────────────
	_, err = store.GetProfile(ctx) // ensure row exists
	require.NoError(t, err)
	profile, err := store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Test User",
		Email:    "test@example.com",
		Phone:    "555-1234",
		Location: "Test City, TS",
	})
	require.NoError(t, err)

	link, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "GitHub",
		URL:   "https://github.com/testuser",
	})
	require.NoError(t, err)

	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "General",
		BodyText: "Skilled software developer with expertise in backend systems and cloud infrastructure.",
	})
	require.NoError(t, err)

	wh1, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "TestCorp",
		JobTitle:             "Lead Developer",
		StartDate:            "2021-01",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh1.ID, "Implemented scalable microservices architecture", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh1.ID, "Reduced deployment time by 60 percent", domain.BulletTypePrimary)
	require.NoError(t, err)

	wh2, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "DevShop Inc",
		JobTitle:             "Software Engineer",
		StartDate:            "2018-06",
		EndDate:              "2020-12",
		DateGranularityStart: "month",
		DateGranularityEnd:   "month",
	})
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh2.ID, "Built REST API serving thousands of requests per second", domain.BulletTypePrimary)
	require.NoError(t, err)

	cat, err := store.CreateSkillCategory(ctx, "Programming Languages")
	require.NoError(t, err)

	goSkill, err := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	require.NoError(t, err)

	pySkill, err := store.CreateSkill(ctx, domain.SkillInput{
		Name: "Python", CategoryID: cat.ID, CompetenceLevel: 8,
	})
	require.NoError(t, err)

	academic, err := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution:     "State University",
		CredentialType:  "Bachelor of Science",
		FieldOfStudy:    "Computer Science",
		CompletionDate:  "2018-05",
		DateGranularity: "month",
	})
	require.NoError(t, err)

	// ──────────────────────────────────────────────────────
	// Step 7: Render PDF using the custom template.
	// ──────────────────────────────────────────────────────
	workList, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)

	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	pdfPath, err := renderer.RenderResume(ctx, domain.RenderResumeRequest{
		Template:    &detail,
		OutputDir:   outputDir,
		Profile:     profile,
		Links:       []domain.ProfileLink{link},
		Summaries:   []domain.ProfessionalSummary{summary},
		WorkHistory: workList,
		Skills:      []domain.Skill{goSkill, pySkill},
		SkillCategoryNames: map[int64]string{
			cat.ID: "Programming Languages",
		},
		Academics: []domain.AcademicCredential{academic},
	})
	require.NoError(t, err)

	// ──────────────────────────────────────────────────────
	// Step 8: Extract text and verify all expected content.
	// ──────────────────────────────────────────────────────
	text := extractPDFText(t, pdfPath)

	// Profile info.
	assert.Contains(t, text, "Test User", "profile name should appear")
	assert.Contains(t, text, "test@example.com", "email should appear")
	assert.Contains(t, text, "555-1234", "phone should appear")

	// Work history.
	assert.Contains(t, text, "TestCorp", "first employer should appear")
	assert.Contains(t, text, "Lead Developer", "first job title should appear")
	assert.Contains(t, text, "microservices", "bullet content should appear")
	assert.Contains(t, text, "DevShop", "second employer should appear")
	assert.Contains(t, text, "REST API", "second job bullet should appear")

	// Summary.
	assert.Contains(t, text, "backend systems", "summary content should appear")

	// Skills.
	assert.Contains(t, text, "Go", "skill should appear")
	assert.Contains(t, text, "Python", "skill should appear")

	// Education.
	assert.Contains(t, text, "State University", "institution should appear")
	assert.Contains(t, text, "Computer Science", "field of study should appear")

	// ──────────────────────────────────────────────────────
	// Step 9: Verify section reading order.
	// After reorder, Experience should appear before Summary.
	// ──────────────────────────────────────────────────────
	nameIdx := strings.Index(text, "Test User")
	expIdx := strings.Index(text, "EXPERIENCE")
	workIdx := strings.Index(text, "TestCorp")
	sumHeadIdx := strings.Index(text, "SUMMARY")
	sumTextIdx := strings.Index(text, "backend systems")
	skillsIdx := strings.Index(text, "SKILLS")
	eduIdx := strings.Index(text, "EDUCATION")

	require.NotEqual(t, -1, nameIdx, "name should be found")
	require.NotEqual(t, -1, expIdx, "EXPERIENCE heading should be found")
	require.NotEqual(t, -1, workIdx, "employer should be found")
	require.NotEqual(t, -1, sumHeadIdx, "SUMMARY heading should be found")
	require.NotEqual(t, -1, sumTextIdx, "summary text should be found")
	require.NotEqual(t, -1, skillsIdx, "SKILLS heading should be found")
	require.NotEqual(t, -1, eduIdx, "EDUCATION heading should be found")

	// Profile appears first.
	assert.Less(t, nameIdx, expIdx,
		"name should appear before EXPERIENCE")

	// Experience section appears before Summary section (due to reorder).
	assert.Less(t, expIdx, workIdx,
		"EXPERIENCE heading should appear before work content")
	assert.Less(t, workIdx, sumHeadIdx,
		"work content should appear before SUMMARY heading (reorder)")
	assert.Less(t, expIdx, sumHeadIdx,
		"EXPERIENCE should appear before SUMMARY after reorder")

	// Summary appears before Skills.
	assert.Less(t, sumHeadIdx, sumTextIdx,
		"SUMMARY heading should appear before summary text")
	assert.Less(t, sumTextIdx, skillsIdx,
		"summary text should appear before SKILLS heading")

	// Skills appears before Education.
	assert.Less(t, skillsIdx, eduIdx,
		"SKILLS should appear before EDUCATION")
}

// TestTemplate_CustomTemplateWithBuiltinComparison creates a custom
// template that mirrors the Professional layout and verifies both
// produce PDFs containing the same data sections.
func TestTemplate_CustomTemplateWithBuiltinComparison(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Create a custom template mirroring Professional layout.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Custom Professional Clone",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Add same elements as Professional template.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfileHeader,
		Config: mustJSON(pdf.ProfileHeaderConfig{
			NameFontSize: 18.0, DetailFontSize: 10.0,
			Alignment: "center", LinkSeparator: " | ",
			ShowLinks: true, ShowLinksInline: false, SpaceAfter: 6.0,
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementRoleDescriptors,
		Config: mustJSON(pdf.RoleDescriptorsConfig{
			FontSize: 10.0, FontStyle: "regular",
			Alignment: "center", Separator: " | ", SpaceAfter: 4.0,
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Professional Summary", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "summaries",
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfSummary,
		Config: mustJSON(pdf.ProfSummaryConfig{
			FontSize: 10.0, BulletChar: "\u2022",
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Experience", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "work_history",
		}),
	})
	require.NoError(t, err)

	workLoop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config: mustJSON(pdf.WorkHistoryLoopConfig{
			EntryGap: 4.0,
		}),
	})
	require.NoError(t, err)

	// Work loop children.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &workLoop.ID,
		ElementType: domain.ElementWorkTitle,
		Config: mustJSON(pdf.WorkTitleConfig{
			FontSize: 10.0, FontStyle: "bold",
			IncludeEmployer: true, EmployerSeparator: " \u2014 ",
			EmployerFontStyle: "italic", SpaceAfter: 13.0,
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &workLoop.ID,
		ElementType: domain.ElementWorkDates,
		Config: mustJSON(pdf.WorkDatesConfig{
			FontSize: 9.0, Alignment: "right",
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &workLoop.ID,
		ElementType: domain.ElementWorkBullets,
		Config: mustJSON(pdf.WorkBulletsConfig{
			FontSize: 10.0, FontStyle: "regular",
			BulletChar: "\u2022", Indent: 12.0, BulletSymWidth: 10.0,
		}),
	})
	require.NoError(t, err)

	// Load the template.
	detail, err := store.GetDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)

	// Seed test data.
	_, err = store.GetProfile(ctx)
	require.NoError(t, err)
	profile, err := store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Jane Doe",
		Email:    "jane@test.com",
		Phone:    "555-9999",
		Location: "Portland, OR",
	})
	require.NoError(t, err)

	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "General",
		BodyText: "Full-stack engineer with cloud expertise.",
	})
	require.NoError(t, err)

	desc, err := store.CreateDescriptor(ctx, "Senior Engineer")
	require.NoError(t, err)

	wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "CloudCo",
		JobTitle:             "Senior Engineer",
		StartDate:            "2020-01",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh.ID, "Deployed Kubernetes clusters in production", domain.BulletTypePrimary)
	require.NoError(t, err)

	workList, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)

	// Render with custom template.
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	customPath, err := renderer.RenderResume(ctx, domain.RenderResumeRequest{
		Template:    &detail,
		OutputDir:   outputDir,
		Profile:     profile,
		Summaries:   []domain.ProfessionalSummary{summary},
		WorkHistory: workList,
		Descriptors: []domain.RoleDescriptor{desc},
	})
	require.NoError(t, err)

	// Render with built-in Professional template using the same data.
	profTmpl := pdf.ProfessionalTemplate()
	builtinPath, err := renderer.RenderResume(ctx, domain.RenderResumeRequest{
		Template:    &profTmpl,
		OutputDir:   outputDir,
		Profile:     profile,
		Summaries:   []domain.ProfessionalSummary{summary},
		WorkHistory: workList,
		Descriptors: []domain.RoleDescriptor{desc},
	})
	require.NoError(t, err)

	// Both PDFs should contain the same key data.
	customText := extractPDFText(t, customPath)
	builtinText := extractPDFText(t, builtinPath)

	for _, expected := range []string{
		"Jane Doe",
		"jane@test.com",
		"CloudCo",
		"Senior Engineer",
		"Kubernetes",
		"cloud expertise",
	} {
		assert.Contains(t, customText, expected,
			"custom template PDF should contain %q", expected)
		assert.Contains(t, builtinText, expected,
			"built-in template PDF should contain %q", expected)
	}
}

// TestTemplate_EmptySectionsOmitted verifies that when resume data
// is missing for certain sections (e.g., no skills, no education),
// those sections and their headings are omitted from the PDF output.
func TestTemplate_EmptySectionsOmitted(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Create a template with headings for all sections.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Sparse Template",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Profile header.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfileHeader,
		Config: mustJSON(pdf.ProfileHeaderConfig{
			NameFontSize: 18.0, DetailFontSize: 10.0,
			Alignment: "center", SpaceAfter: 6.0,
		}),
	})
	require.NoError(t, err)

	// Summary section (will have data).
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Summary", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "summaries",
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfSummary,
		Config: mustJSON(pdf.ProfSummaryConfig{
			FontSize: 10.0, BulletChar: "\u2022",
		}),
	})
	require.NoError(t, err)

	// Skills section (will have NO data — heading should be omitted).
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Skills", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "skills",
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSkills,
		Config: mustJSON(pdf.SkillsConfig{
			FontSize: 10.0, GroupByCategory: true,
			CategoryFontStyle: "bold", SkillSeparator: ", ",
		}),
	})
	require.NoError(t, err)

	// Education section (will have NO data — heading should be omitted).
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Education", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0, DataBinding: "academics",
		}),
	})
	require.NoError(t, err)

	eduLoop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementEducationLoop,
		Config: mustJSON(pdf.EducationLoopConfig{
			EntryGap: 0.0,
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &eduLoop.ID,
		ElementType: domain.ElementEduCredential,
		Config: mustJSON(pdf.EduCredentialConfig{
			FontSize: 10.0, FontStyle: "bold",
		}),
	})
	require.NoError(t, err)

	// Load the template.
	detail, err := store.GetDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)

	// Seed only profile + summary (no skills, no education).
	_, err = store.GetProfile(ctx)
	require.NoError(t, err)
	profile, err := store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Sparse User",
		Email:    "sparse@test.com",
		Phone:    "555-0000",
		Location: "Nowhere, NA",
	})
	require.NoError(t, err)

	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "General",
		BodyText: "A focused professional with narrow expertise.",
	})
	require.NoError(t, err)

	// Render PDF with sparse data.
	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	pdfPath, err := renderer.RenderResume(ctx, domain.RenderResumeRequest{
		Template:  &detail,
		OutputDir: outputDir,
		Profile:   profile,
		Summaries: []domain.ProfessionalSummary{summary},
		// No skills, no academics, no work history.
	})
	require.NoError(t, err)

	text := extractPDFText(t, pdfPath)

	// Summary section SHOULD appear (it has data).
	assert.Contains(t, text, "SUMMARY", "SUMMARY heading should appear — data exists")
	assert.Contains(t, text, "narrow expertise", "summary text should appear")

	// Skills and Education sections should NOT appear (no data,
	// data_binding should cause heading to be skipped).
	assert.NotContains(t, text, "SKILLS",
		"SKILLS heading should be omitted — no skill data")
	assert.NotContains(t, text, "EDUCATION",
		"EDUCATION heading should be omitted — no academic data")
}

// TestTemplate_PreviewTemplate (T058) verifies that
// ResumeService.PreviewTemplate generates valid PDFs for both resume
// and cover letter templates using all available user data.
func TestTemplate_PreviewTemplate(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Seed user data.
	_, err := store.GetProfile(ctx)
	require.NoError(t, err)
	_, err = store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Preview User",
		Email:    "preview@test.com",
		Phone:    "555-7777",
		Location: "Seattle, WA",
	})
	require.NoError(t, err)

	_, err = store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "https://linkedin.com/in/preview",
	})
	require.NoError(t, err)

	_, err = store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "Preview Summary",
		BodyText: "Experienced engineer specializing in preview testing.",
	})
	require.NoError(t, err)

	wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "PreviewCo",
		JobTitle:             "Staff Engineer",
		StartDate:            "2022-06",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh.ID, "Led preview infrastructure team", domain.BulletTypePrimary)
	require.NoError(t, err)

	cat, err := store.CreateSkillCategory(ctx, "Languages")
	require.NoError(t, err)
	_, err = store.CreateSkill(ctx, domain.SkillInput{
		Name: "Go", CategoryID: cat.ID, CompetenceLevel: 9,
	})
	require.NoError(t, err)

	_, err = store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution:     "Preview University",
		CredentialType:  "BS",
		FieldOfStudy:    "CS",
		CompletionDate:  "2020-05",
		DateGranularity: "month",
	})
	require.NoError(t, err)

	_, err = store.CreateDescriptor(ctx, "Platform Architect")
	require.NoError(t, err)

	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()
	svc := service.NewResumeService(store, renderer, outputDir)

	t.Run("resume template preview", func(t *testing.T) {
		// Use the built-in Professional template (seeded by migration).
		templates, err := store.ListDocumentTemplates(ctx)
		require.NoError(t, err)

		var resumeTemplateID int64
		for _, tmpl := range templates {
			if tmpl.TemplateType == domain.TemplateTypeResume {
				resumeTemplateID = tmpl.ID
				break
			}
		}
		require.NotZero(t, resumeTemplateID, "should find a resume template")

		pdfPath, err := svc.PreviewTemplate(ctx, resumeTemplateID)
		require.NoError(t, err)

		// Verify file exists and is non-empty.
		info, err := os.Stat(pdfPath)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0), "PDF should not be empty")

		// Verify content includes seeded data.
		text := extractPDFText(t, pdfPath)
		assert.Contains(t, text, "Preview User", "profile name should appear")
		assert.Contains(t, text, "preview@test.com", "email should appear")
		assert.Contains(t, text, "PreviewCo", "employer should appear")
		assert.Contains(t, text, "Staff Engineer", "job title should appear")
		assert.Contains(t, text, "preview infrastructure", "bullet should appear")
		assert.Contains(t, text, "Go", "skill should appear")
		assert.Contains(t, text, "Preview University", "academic should appear")
		assert.Contains(t, text, "Platform Architect", "descriptor should appear")
	})

	t.Run("cover letter template preview with placeholders", func(t *testing.T) {
		// Use the built-in Formal cover letter template.
		templates, err := store.ListDocumentTemplates(ctx)
		require.NoError(t, err)

		var clTemplateID int64
		for _, tmpl := range templates {
			if tmpl.TemplateType == domain.TemplateTypeCoverLetter {
				clTemplateID = tmpl.ID
				break
			}
		}
		require.NotZero(t, clTemplateID, "should find a cover letter template")

		pdfPath, err := svc.PreviewTemplate(ctx, clTemplateID)
		require.NoError(t, err)

		// Verify file exists.
		info, err := os.Stat(pdfPath)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0), "PDF should not be empty")

		// Verify profile data appears.
		text := extractPDFText(t, pdfPath)
		assert.Contains(t, text, "Preview User", "profile name should appear in CL")
		assert.Contains(t, text, "preview@test.com", "email should appear in CL")

		// Verify placeholder substitutions (variables become [variable_name]).
		assert.Contains(t, text, "[hiring_manager]",
			"hiring_manager placeholder should appear")
		assert.Contains(t, text, "[signer_name]",
			"signer_name placeholder should appear")
	})

	t.Run("invalid template ID returns error", func(t *testing.T) {
		_, err := svc.PreviewTemplate(ctx, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "template ID is required")
	})

	t.Run("nonexistent template returns error", func(t *testing.T) {
		_, err := svc.PreviewTemplate(ctx, 99999)
		require.Error(t, err)
	})
}

// TestTemplate_WorkHistoryLoopMultipleEntries (T035) verifies that
// a work_history_loop iterates over multiple work entries and renders
// sub-elements (title, dates, bullets) in the correct per-entry layout.
func TestTemplate_WorkHistoryLoopMultipleEntries(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Create a minimal template focused on work history.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Loop Iteration Test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Profile header.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfileHeader,
		Config: mustJSON(pdf.ProfileHeaderConfig{
			NameFontSize: 18.0, DetailFontSize: 10.0,
			Alignment: "center", SpaceAfter: 6.0,
		}),
	})
	require.NoError(t, err)

	// Experience heading.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Experience", FontSize: 12.0, FontStyle: "bold",
			Uppercase: true, Underline: true, UnderlineWeight: 0.5,
			SpaceBefore: 10.0, SpaceAfter: 4.0,
			DataBinding: "work_history",
		}),
	})
	require.NoError(t, err)

	// Work history loop with entry_gap.
	workLoop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config: mustJSON(pdf.WorkHistoryLoopConfig{
			EntryGap: 6.0,
		}),
	})
	require.NoError(t, err)

	// Sub-elements: work_title, work_dates, work_bullets.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &workLoop.ID,
		ElementType: domain.ElementWorkTitle,
		Config: mustJSON(pdf.WorkTitleConfig{
			FontSize: 10.0, FontStyle: "bold",
			IncludeEmployer: true, EmployerSeparator: " \u2014 ",
			EmployerFontStyle: "italic", SpaceAfter: 13.0,
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &workLoop.ID,
		ElementType: domain.ElementWorkDates,
		Config: mustJSON(pdf.WorkDatesConfig{
			FontSize: 9.0, Alignment: "right",
		}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ParentID:    &workLoop.ID,
		ElementType: domain.ElementWorkBullets,
		Config: mustJSON(pdf.WorkBulletsConfig{
			FontSize: 10.0, FontStyle: "regular",
			BulletChar: "\u2022", Indent: 12.0, BulletSymWidth: 10.0,
		}),
	})
	require.NoError(t, err)

	// Load the template.
	detail, err := store.GetDocumentTemplate(ctx, tmpl.ID)
	require.NoError(t, err)

	// Seed profile.
	_, err = store.GetProfile(ctx)
	require.NoError(t, err)
	profile, err := store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Loop Test User",
		Email:    "loop@test.com",
	})
	require.NoError(t, err)

	// Seed THREE work history entries.
	wh1, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "AlphaCo",
		JobTitle:             "Lead Engineer",
		StartDate:            "2022-01",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh1.ID, "Built distributed event processing system", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh1.ID, "Led team of eight engineers", domain.BulletTypePrimary)
	require.NoError(t, err)

	wh2, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "BetaCorp",
		JobTitle:             "Senior Developer",
		StartDate:            "2019-06",
		EndDate:              "2021-12",
		DateGranularityStart: "month",
		DateGranularityEnd:   "month",
	})
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh2.ID, "Designed GraphQL API layer", domain.BulletTypePrimary)
	require.NoError(t, err)

	wh3, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "GammaTech",
		JobTitle:             "Software Engineer",
		StartDate:            "2017-03",
		EndDate:              "2019-05",
		DateGranularityStart: "month",
		DateGranularityEnd:   "month",
	})
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh3.ID, "Migrated legacy monolith to microservices", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, wh3.ID, "Improved test coverage from 40 to 90 percent", domain.BulletTypePrimary)
	require.NoError(t, err)

	// Render the PDF.
	workList, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)
	require.Len(t, workList, 3, "should have 3 work entries")

	renderer := pdf.NewRenderer()
	outputDir := t.TempDir()

	pdfPath, err := renderer.RenderResume(ctx, domain.RenderResumeRequest{
		Template:    &detail,
		OutputDir:   outputDir,
		Profile:     profile,
		WorkHistory: workList,
	})
	require.NoError(t, err)

	// Extract text and verify all three entries render.
	text := extractPDFText(t, pdfPath)

	// All three employers should appear.
	assert.Contains(t, text, "AlphaCo", "first employer should appear")
	assert.Contains(t, text, "BetaCorp", "second employer should appear")
	assert.Contains(t, text, "GammaTech", "third employer should appear")

	// All three job titles should appear.
	assert.Contains(t, text, "Lead Engineer", "first job title should appear")
	assert.Contains(t, text, "Senior Developer", "second job title should appear")
	assert.Contains(t, text, "Software Engineer", "third job title should appear")

	// Bullets from all entries should appear.
	assert.Contains(t, text, "distributed event", "first entry bullet should appear")
	assert.Contains(t, text, "eight engineers", "first entry second bullet should appear")
	assert.Contains(t, text, "GraphQL", "second entry bullet should appear")
	assert.Contains(t, text, "legacy monolith", "third entry bullet should appear")
	assert.Contains(t, text, "test coverage", "third entry second bullet should appear")

	// Verify order: entries appear by sort_order. CreateWorkHistory
	// inserts at sort_order 1 and shifts existing entries, so the
	// last-created entry (GammaTech) comes first, then BetaCorp,
	// then AlphaCo.
	gammaIdx := strings.Index(text, "GammaTech")
	betaIdx := strings.Index(text, "BetaCorp")
	alphaIdx := strings.Index(text, "AlphaCo")

	require.NotEqual(t, -1, gammaIdx, "GammaTech must be found")
	require.NotEqual(t, -1, betaIdx, "BetaCorp must be found")
	require.NotEqual(t, -1, alphaIdx, "AlphaCo must be found")

	assert.Less(t, gammaIdx, betaIdx,
		"GammaTech should appear before BetaCorp (sort_order)")
	assert.Less(t, betaIdx, alphaIdx,
		"BetaCorp should appear before AlphaCo")

	// Verify the EXPERIENCE heading appears before any work content.
	expIdx := strings.Index(text, "EXPERIENCE")
	require.NotEqual(t, -1, expIdx, "EXPERIENCE heading should be found")
	assert.Less(t, expIdx, gammaIdx,
		"EXPERIENCE heading should appear before first entry")
}

// TestTemplate_ExportImportRoundTrip (T070) verifies that exporting a
// template to JSON and importing it back produces an identical structure.
func TestTemplate_ExportImportRoundTrip(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	templateSvc := service.NewTemplateService(store)

	// Create a template with multiple elements including nested children.
	tmpl, err := store.CreateDocumentTemplate(ctx, domain.DocumentTemplateInput{
		Name:         "Round-Trip Test",
		Description:  "Export/import round-trip test",
		TemplateType: domain.TemplateTypeResume,
		MarginTop:    54.0,
		MarginBottom: 54.0,
		MarginLeft:   72.0,
		MarginRight:  72.0,
	})
	require.NoError(t, err)

	// Add a profile header.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementProfileHeader,
		Config: mustJSON(pdf.ProfileHeaderConfig{
			NameFontSize: 18.0, DetailFontSize: 10.0,
			Alignment: "center", SpaceAfter: 6.0,
		}),
	})
	require.NoError(t, err)

	// Add a section heading.
	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementSectionHeading,
		Config: mustJSON(pdf.SectionHeadingConfig{
			Text: "Experience", FontSize: 11.0, Uppercase: true,
			Underline: true, SpaceBefore: 10.0, SpaceAfter: 4.0,
			UnderlineWeight: 0.5, DataBinding: "work_history",
		}),
	})
	require.NoError(t, err)

	// Add a work history loop with child elements.
	loop, err := store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkHistoryLoop,
		Config:      mustJSON(pdf.WorkHistoryLoopConfig{EntryGap: 8.0}),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkTitle,
		Config:      mustJSON(pdf.WorkTitleConfig{FontSize: 11.0, FontStyle: "bold", SpaceAfter: 2.0}),
		ParentID:    &loop.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateTemplateElement(ctx, tmpl.ID, domain.TemplateElementInput{
		ElementType: domain.ElementWorkBullets,
		Config:      mustJSON(pdf.WorkBulletsConfig{FontSize: 10.0, Indent: 18.0, BulletChar: "\u2022"}),
		ParentID:    &loop.ID,
	})
	require.NoError(t, err)

	// Export to a temp file.
	tmpFile := t.TempDir() + "/exported-template.json"
	err = templateSvc.ExportTemplate(ctx, tmpl.ID, tmpFile)
	require.NoError(t, err)

	// Verify the file is valid JSON.
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var exported domain.TemplateDetail
	err = json.Unmarshal(data, &exported)
	require.NoError(t, err)
	assert.Equal(t, "Round-Trip Test", exported.Name)
	assert.Equal(t, "Export/import round-trip test", exported.Description)
	assert.Equal(t, domain.TemplateTypeResume, exported.TemplateType)
	assert.Len(t, exported.Elements, 5, "should have 5 elements (header + heading + loop + 2 children)")

	// Import from the file.
	imported, err := templateSvc.ImportTemplate(ctx, tmpFile)
	require.NoError(t, err)
	assert.NotEqual(t, tmpl.ID, imported.ID, "imported template should have a new ID")
	assert.Equal(t, "Round-Trip Test", imported.Name)
	assert.Equal(t, domain.TemplateTypeResume, imported.TemplateType)

	// Fetch the full detail of the imported template.
	importedDetail, err := store.GetDocumentTemplate(ctx, imported.ID)
	require.NoError(t, err)
	assert.Len(t, importedDetail.Elements, 5, "imported should have 5 elements")

	// Verify element types match in order.
	expectedTypes := []string{
		domain.ElementProfileHeader,
		domain.ElementSectionHeading,
		domain.ElementWorkHistoryLoop,
		domain.ElementWorkTitle,
		domain.ElementWorkBullets,
	}
	for i, el := range importedDetail.Elements {
		assert.Equal(t, expectedTypes[i], el.ElementType, "element %d type mismatch", i)
	}

	// Verify parent relationships are preserved.
	// The loop should be a top-level element.
	var importedLoop *domain.TemplateElement
	for i, el := range importedDetail.Elements {
		if el.ElementType == domain.ElementWorkHistoryLoop {
			importedLoop = &importedDetail.Elements[i]
			break
		}
	}
	require.NotNil(t, importedLoop, "should find work_history_loop")
	assert.Nil(t, importedLoop.ParentID, "loop should be top-level")

	// Children should reference the imported loop's ID.
	for _, el := range importedDetail.Elements {
		if el.ElementType == domain.ElementWorkTitle || el.ElementType == domain.ElementWorkBullets {
			require.NotNil(t, el.ParentID, "%s should have a parent", el.ElementType)
			assert.Equal(t, importedLoop.ID, *el.ParentID, "%s should be child of loop", el.ElementType)
		}
	}

	// Verify config data round-trips correctly.
	for _, el := range importedDetail.Elements {
		if el.ElementType == domain.ElementSectionHeading {
			var cfg pdf.SectionHeadingConfig
			err := json.Unmarshal([]byte(el.Config), &cfg)
			require.NoError(t, err)
			assert.Equal(t, "Experience", cfg.Text)
			assert.Equal(t, true, cfg.Uppercase)
			assert.Equal(t, 11.0, cfg.FontSize)
			break
		}
	}

	// Import should create a non-built-in template.
	assert.False(t, importedDetail.IsBuiltin, "imported template should not be built-in")
}

// TestTemplate_ExportImportBuiltin (T070) verifies that a built-in
// template can be exported and imported back as a user template.
func TestTemplate_ExportImportBuiltin(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	templateSvc := service.NewTemplateService(store)

	// Export the built-in Professional template (ID=1).
	tmpFile := t.TempDir() + "/builtin-export.json"
	err := templateSvc.ExportTemplate(ctx, 1, tmpFile)
	require.NoError(t, err)

	// Import it.
	imported, err := templateSvc.ImportTemplate(ctx, tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "Professional", imported.Name)
	assert.False(t, imported.IsBuiltin, "imported should be user template, not built-in")

	// Get the original and imported details.
	originalDetail, err := store.GetDocumentTemplate(ctx, 1)
	require.NoError(t, err)

	importedDetail, err := store.GetDocumentTemplate(ctx, imported.ID)
	require.NoError(t, err)

	// Should have the same number of elements.
	assert.Equal(t, len(originalDetail.Elements), len(importedDetail.Elements),
		"imported should have same element count as original")

	// Element types should match in order.
	for i := range originalDetail.Elements {
		assert.Equal(t, originalDetail.Elements[i].ElementType, importedDetail.Elements[i].ElementType,
			"element %d type mismatch", i)
	}
}

// TestTemplate_ImportInvalidFile (T070) verifies that importing invalid
// files returns appropriate errors.
func TestTemplate_ImportInvalidFile(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	templateSvc := service.NewTemplateService(store)

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := templateSvc.ImportTemplate(ctx, "/nonexistent/path.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "import template: read:")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpFile := t.TempDir() + "/bad.json"
		err := os.WriteFile(tmpFile, []byte("not json"), 0644)
		require.NoError(t, err)

		_, err = templateSvc.ImportTemplate(ctx, tmpFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "import template: parse:")
	})

	t.Run("missing name", func(t *testing.T) {
		tmpFile := t.TempDir() + "/no-name.json"
		data := `{"template_type": "resume", "elements": []}`
		err := os.WriteFile(tmpFile, []byte(data), 0644)
		require.NoError(t, err)

		_, err = templateSvc.ImportTemplate(ctx, tmpFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("invalid template type", func(t *testing.T) {
		tmpFile := t.TempDir() + "/bad-type.json"
		data := `{"name": "Test", "template_type": "unknown", "elements": []}`
		err := os.WriteFile(tmpFile, []byte(data), 0644)
		require.NoError(t, err)

		_, err = templateSvc.ImportTemplate(ctx, tmpFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid template type")
	})

	t.Run("invalid element type", func(t *testing.T) {
		tmpFile := t.TempDir() + "/bad-elem.json"
		data := `{"name": "Test", "template_type": "resume", "elements": [{"element_type": "bogus", "config": "{}"}]}`
		err := os.WriteFile(tmpFile, []byte(data), 0644)
		require.NoError(t, err)

		_, err = templateSvc.ImportTemplate(ctx, tmpFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid element type")
	})

	t.Run("incompatible element type", func(t *testing.T) {
		tmpFile := t.TempDir() + "/incompatible.json"
		// body_text is a cover letter element, not resume.
		data := `{"name": "Test", "template_type": "resume", "elements": [{"element_type": "body_text", "config": "{}"}]}`
		err := os.WriteFile(tmpFile, []byte(data), 0644)
		require.NoError(t, err)

		_, err = templateSvc.ImportTemplate(ctx, tmpFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not compatible")
	})
}
