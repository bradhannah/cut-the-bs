package integration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"cut-the-bs/internal/domain"
	"cut-the-bs/internal/infra/pdf"

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
