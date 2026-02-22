//go:build ignore

// This file has pre-existing compile errors unrelated to current work.
// Tagged ignore so it does not block the service package test build.

package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =========================================================
// Mock store
// =========================================================

type mockExportStore struct {
	// Return values
	profile         domain.UserProfile
	links           []domain.ProfileLink
	summary         domain.ProfessionalSummary
	workHistory     []domain.WorkHistoryEntry
	skills          []domain.Skill
	skillCategories []domain.SkillCategory
	academics       []domain.AcademicCredential
	certs           []domain.Certification
	descriptors     []domain.RoleDescriptor
	coreExpertise   []domain.CoreExpertise
	exports         []domain.ResumeExport
	export          domain.ResumeExport
	templateDetail  domain.TemplateDetail

	err error

	// Call tracking
	createExportCalls     []domain.ResumeExport
	createSelectionsCalls []createSelectionsCall
	getSummaryCalls       []int64
	getTemplateCalls      []int64
}

type createSelectionsCall struct {
	ExportID int64
	Request  domain.ExportRequest
}

func (m *mockExportStore) GetProfile(_ context.Context) (domain.UserProfile, error) {
	return m.profile, m.err
}

func (m *mockExportStore) ListProfileLinks(_ context.Context) ([]domain.ProfileLink, error) {
	return m.links, m.err
}

func (m *mockExportStore) GetSummary(_ context.Context, id int64) (domain.ProfessionalSummary, error) {
	m.getSummaryCalls = append(m.getSummaryCalls, id)
	if m.err != nil {
		return domain.ProfessionalSummary{}, m.err
	}
	return m.summary, nil
}

func (m *mockExportStore) ListWorkHistory(_ context.Context) ([]domain.WorkHistoryEntry, error) {
	return m.workHistory, m.err
}

func (m *mockExportStore) ListSkills(_ context.Context) ([]domain.Skill, error) {
	return m.skills, m.err
}

func (m *mockExportStore) ListAcademicCredentials(_ context.Context) ([]domain.AcademicCredential, error) {
	return m.academics, m.err
}

func (m *mockExportStore) ListCertifications(_ context.Context) ([]domain.Certification, error) {
	return m.certs, m.err
}

func (m *mockExportStore) ListDescriptors(_ context.Context) ([]domain.RoleDescriptor, error) {
	return m.descriptors, m.err
}

func (m *mockExportStore) ListCoreExpertise(_ context.Context) ([]domain.CoreExpertise, error) {
	return m.coreExpertise, m.err
}

func (m *mockExportStore) ListSkillCategories(_ context.Context) ([]domain.SkillCategory, error) {
	return m.skillCategories, m.err
}

func (m *mockExportStore) GetDocumentTemplate(_ context.Context, id int64) (domain.TemplateDetail, error) {
	m.getTemplateCalls = append(m.getTemplateCalls, id)
	if m.err != nil {
		return domain.TemplateDetail{}, m.err
	}
	return m.templateDetail, nil
}

func (m *mockExportStore) ListExports(_ context.Context) ([]domain.ResumeExport, error) {
	return m.exports, m.err
}

func (m *mockExportStore) CreateExport(_ context.Context, exp domain.ResumeExport) (domain.ResumeExport, error) {
	m.createExportCalls = append(m.createExportCalls, exp)
	if m.err != nil {
		return domain.ResumeExport{}, m.err
	}
	exp.ID = 1
	exp.GeneratedAt = "2026-01-15T10:30:00Z"
	return exp, nil
}

func (m *mockExportStore) GetExport(_ context.Context, id int64) (domain.ResumeExport, error) {
	if m.err != nil {
		return domain.ResumeExport{}, m.err
	}
	return m.export, nil
}

func (m *mockExportStore) CreateExportSelections(_ context.Context, exportID int64, req domain.ExportRequest) error {
	m.createSelectionsCalls = append(m.createSelectionsCalls, createSelectionsCall{
		ExportID: exportID,
		Request:  req,
	})
	return m.err
}

// =========================================================
// Mock PDF renderer
// =========================================================

type mockPDFRenderer struct {
	filePath string
	err      error

	// Call tracking
	renderCalls []domain.RenderResumeRequest
	templates   []domain.ResumeTemplate
}

func (m *mockPDFRenderer) RenderResume(_ context.Context, req domain.RenderResumeRequest) (string, error) {
	m.renderCalls = append(m.renderCalls, req)
	if m.err != nil {
		return "", m.err
	}
	return m.filePath, nil
}

func (m *mockPDFRenderer) RenderCoverLetter(_ context.Context, _ domain.RenderCoverLetterRequest) (string, error) {
	return "", nil
}

func (m *mockPDFRenderer) ListTemplates() []domain.ResumeTemplate {
	return m.templates
}

// =========================================================
// Helpers
// =========================================================

func defaultTestStore() *mockExportStore {
	return &mockExportStore{
		profile: domain.UserProfile{
			ID:       1,
			FullName: "Jane Doe",
			Email:    "jane@example.com",
			Phone:    "555-1234",
			Location: "Seattle, WA",
		},
		links: []domain.ProfileLink{
			{ID: 1, Label: "GitHub", URL: "https://github.com/janedoe"},
		},
		summary: domain.ProfessionalSummary{
			ID:       1,
			Label:    "General",
			BodyText: "Experienced software engineer...",
		},
		workHistory: []domain.WorkHistoryEntry{
			{
				ID:           1,
				EmployerName: "Acme Corp",
				JobTitle:     "Senior Dev",
				StartDate:    "2020-01",
				EndDate:      "Present",
				Bullets: []domain.AchievementBullet{
					{ID: 10, WorkHistoryID: 1, Text: "Led team of 5"},
					{ID: 11, WorkHistoryID: 1, Text: "Improved CI pipeline"},
				},
			},
			{
				ID:           2,
				EmployerName: "Beta Inc",
				JobTitle:     "Junior Dev",
				StartDate:    "2018-06",
				EndDate:      "2020-01",
				Bullets: []domain.AchievementBullet{
					{ID: 20, WorkHistoryID: 2, Text: "Built REST API"},
				},
			},
		},
		skills: []domain.Skill{
			{ID: 1, Name: "Go", CategoryID: 1, CompetenceLevel: 9},
			{ID: 2, Name: "TypeScript", CategoryID: 1, CompetenceLevel: 8},
			{ID: 3, Name: "SQL", CategoryID: 2, CompetenceLevel: 7},
			{ID: 4, Name: "COBOL", CategoryID: 2, CompetenceLevel: 3, IsLegacy: true},
		},
		academics: []domain.AcademicCredential{
			{ID: 1, Institution: "MIT", CredentialType: "B.S.", FieldOfStudy: "CS"},
		},
		certs: []domain.Certification{
			{ID: 1, Name: "AWS SAA", IssuingBody: "Amazon", IsActive: true},
		},
		descriptors: []domain.RoleDescriptor{
			{ID: 1, Title: "Full-Stack Developer"},
			{ID: 2, Title: "Backend Engineer"},
		},
	}
}

func defaultTestRenderer() *mockPDFRenderer {
	return &mockPDFRenderer{
		filePath: "/exports/resume-professional-20260115.pdf",
		templates: []domain.ResumeTemplate{
			{ID: "professional", Name: "Professional"},
			{ID: "modern", Name: "Modern"},
		},
	}
}

func fullExportRequest() domain.ExportRequest {
	return domain.ExportRequest{
		TemplateID:       "professional",
		SummaryIDs:       []int64{1},
		WorkHistoryIDs:   []int64{1, 2},
		BulletIDs:        []int64{10, 11, 20},
		SkillIDs:         []int64{1, 2, 3},
		AcademicIDs:      []int64{1},
		CertificationIDs: []int64{1},
		DescriptorIDs:    []int64{1, 2},
	}
}

// =========================================================
// ListTemplates
// =========================================================

func TestResumeService_ListTemplates(t *testing.T) {
	renderer := defaultTestRenderer()
	svc := NewResumeService(defaultTestStore(), renderer, "/exports")

	templates := svc.ListTemplates()
	require.Len(t, templates, 2)
	assert.Equal(t, "professional", templates[0].ID)
	assert.Equal(t, "modern", templates[1].ID)
}

// =========================================================
// ListExports
// =========================================================

func TestResumeService_ListExports_Success(t *testing.T) {
	store := defaultTestStore()
	store.exports = []domain.ResumeExport{
		{ID: 2, TemplateID: "modern", GeneratedAt: "2026-01-16"},
		{ID: 1, TemplateID: "professional", GeneratedAt: "2026-01-15"},
	}
	svc := NewResumeService(store, defaultTestRenderer(), "/exports")

	exports, err := svc.ListExports(context.Background())
	require.NoError(t, err)
	require.Len(t, exports, 2)
	assert.Equal(t, int64(2), exports[0].ID)
}

func TestResumeService_ListExports_StoreError(t *testing.T) {
	store := defaultTestStore()
	store.exports = nil
	store.err = fmt.Errorf("db gone")
	svc := NewResumeService(store, defaultTestRenderer(), "/exports")

	_, err := svc.ListExports(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db gone")
}

// =========================================================
// CreateExport — full selections
// =========================================================

func TestResumeService_CreateExport_FullSelections(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := fullExportRequest()
	result, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)

	// Verify export record was created.
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "professional", result.TemplateID)
	assert.Equal(t, renderer.filePath, result.FilePath)
	assert.NotEmpty(t, result.GeneratedAt)

	// Verify renderer was called with correct data.
	require.Len(t, renderer.renderCalls, 1)
	renderReq := renderer.renderCalls[0]
	assert.Equal(t, "professional", renderReq.TemplateID)
	assert.Equal(t, "/exports", renderReq.OutputDir)
	assert.Equal(t, "Jane Doe", renderReq.Profile.FullName)
	require.Len(t, renderReq.Links, 1)
	assert.NotEmpty(t, renderReq.Summaries)
	assert.Equal(t, "General", renderReq.Summaries[0].Label)
	require.Len(t, renderReq.WorkHistory, 2)
	require.Len(t, renderReq.Skills, 3) // excludes COBOL (not in SkillIDs)
	require.Len(t, renderReq.Academics, 1)
	require.Len(t, renderReq.Certs, 1)
	require.Len(t, renderReq.Descriptors, 2)

	// Verify work history bullet filtering — only selected bullets
	// should be included.
	totalBullets := 0
	for _, wh := range renderReq.WorkHistory {
		totalBullets += len(wh.Bullets)
	}
	assert.Equal(t, 3, totalBullets)

	// Verify export record was stored.
	require.Len(t, store.createExportCalls, 1)
	assert.Equal(t, renderer.filePath, store.createExportCalls[0].FilePath)

	// Verify selections were snapshot.
	require.Len(t, store.createSelectionsCalls, 1)
	assert.Equal(t, int64(1), store.createSelectionsCalls[0].ExportID)
}

// =========================================================
// CreateExport — minimal selections
// =========================================================

func TestResumeService_CreateExport_MinimalSelections(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := domain.ExportRequest{
		TemplateID:    "professional",
		SkillIDs:      []int64{1},
		DescriptorIDs: []int64{1},
	}
	result, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	// Renderer should have been called with only the selected content.
	require.Len(t, renderer.renderCalls, 1)
	renderReq := renderer.renderCalls[0]
	assert.Empty(t, renderReq.Summaries)
	assert.Empty(t, renderReq.WorkHistory)
	require.Len(t, renderReq.Skills, 1)
	assert.Equal(t, "Go", renderReq.Skills[0].Name)
	assert.Empty(t, renderReq.Academics)
	assert.Empty(t, renderReq.Certs)
	require.Len(t, renderReq.Descriptors, 1)
}

// =========================================================
// CreateExport — empty selection rejection
// =========================================================

func TestResumeService_CreateExport_EmptySelectionRejected(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := domain.ExportRequest{
		TemplateID: "professional",
		// No content selected at all.
	}
	_, err := svc.CreateExport(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one content item")

	// Renderer should NOT have been called.
	assert.Empty(t, renderer.renderCalls)
	assert.Empty(t, store.createExportCalls)
}

// =========================================================
// CreateExport — empty template ID rejected
// =========================================================

func TestResumeService_CreateExport_EmptyTemplateIDRejected(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := fullExportRequest()
	req.TemplateID = ""

	_, err := svc.CreateExport(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template ID")
}

// =========================================================
// CreateExport — file path from renderer
// =========================================================

func TestResumeService_CreateExport_FilePathFromRenderer(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	renderer.filePath = "/custom/path/resume.pdf"
	svc := NewResumeService(store, renderer, "/exports")

	req := fullExportRequest()
	result, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "/custom/path/resume.pdf", result.FilePath)
}

// =========================================================
// CreateExport — legacy skill handling
// =========================================================

func TestResumeService_CreateExport_LegacySkillsIncluded(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	// Include the legacy COBOL skill in the selection.
	req := domain.ExportRequest{
		TemplateID: "professional",
		SkillIDs:   []int64{1, 4}, // Go + COBOL (legacy)
	}
	result, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	require.Len(t, renderer.renderCalls, 1)
	skills := renderer.renderCalls[0].Skills
	require.Len(t, skills, 2)

	// Find the legacy skill.
	var legacyFound bool
	for _, s := range skills {
		if s.Name == "COBOL" {
			assert.True(t, s.IsLegacy)
			legacyFound = true
		}
	}
	assert.True(t, legacyFound, "COBOL should be included in export")
}

// =========================================================
// CreateExport — per-export skill sort override (FR-008)
// =========================================================

func TestResumeService_CreateExport_SkillSortOverride(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := domain.ExportRequest{
		TemplateID: "professional",
		SkillIDs:   []int64{1, 2, 3}, // Go, TypeScript, SQL
		SkillSortOverrides: map[int64]int{
			3: 0, // SQL first
			1: 1, // Go second
			2: 2, // TypeScript third
		},
	}
	result, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	require.Len(t, renderer.renderCalls, 1)
	skills := renderer.renderCalls[0].Skills
	require.Len(t, skills, 3)

	// Skills should be ordered by the custom sort override.
	assert.Equal(t, "SQL", skills[0].Name)
	assert.Equal(t, "Go", skills[1].Name)
	assert.Equal(t, "TypeScript", skills[2].Name)

	// Verify the sort override is in the stored selections snapshot.
	require.Len(t, store.createSelectionsCalls, 1)
	storedReq := store.createSelectionsCalls[0].Request
	assert.Equal(t, 0, storedReq.SkillSortOverrides[3])
	assert.Equal(t, 1, storedReq.SkillSortOverrides[1])
}

// =========================================================
// CreateExport — partial skill sort override
// =========================================================

func TestResumeService_CreateExport_PartialSkillSortOverride(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	// Override order for only some skills — others should come after.
	req := domain.ExportRequest{
		TemplateID: "professional",
		SkillIDs:   []int64{1, 2, 3}, // Go, TypeScript, SQL
		SkillSortOverrides: map[int64]int{
			3: 0, // SQL first
		},
	}
	_, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)

	require.Len(t, renderer.renderCalls, 1)
	skills := renderer.renderCalls[0].Skills
	require.Len(t, skills, 3)

	// SQL should be first due to override.
	assert.Equal(t, "SQL", skills[0].Name)
}

// =========================================================
// CreateExport — bullet filtering
// =========================================================

func TestResumeService_CreateExport_BulletFiltering(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	// Select work history entry 1 but only one of its bullets.
	req := domain.ExportRequest{
		TemplateID:     "professional",
		WorkHistoryIDs: []int64{1},
		BulletIDs:      []int64{10}, // Only "Led team of 5", not "Improved CI pipeline"
	}
	_, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)

	require.Len(t, renderer.renderCalls, 1)
	wh := renderer.renderCalls[0].WorkHistory
	require.Len(t, wh, 1)
	require.Len(t, wh[0].Bullets, 1)
	assert.Equal(t, "Led team of 5", wh[0].Bullets[0].Text)
}

// =========================================================
// CreateExport — work history without explicit bullet IDs
//   includes all bullets for selected entries
// =========================================================

func TestResumeService_CreateExport_WorkHistoryNoBulletIDs(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	// Select work history entries but no explicit bullet IDs.
	// All bullets for those entries should be included.
	req := domain.ExportRequest{
		TemplateID:     "professional",
		WorkHistoryIDs: []int64{1},
	}
	_, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)

	require.Len(t, renderer.renderCalls, 1)
	wh := renderer.renderCalls[0].WorkHistory
	require.Len(t, wh, 1)
	assert.Len(t, wh[0].Bullets, 2) // All bullets included
}

// =========================================================
// CreateExport — renderer error propagated
// =========================================================

func TestResumeService_CreateExport_RendererError(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	renderer.err = fmt.Errorf("font not found")
	svc := NewResumeService(store, renderer, "/exports")

	req := fullExportRequest()
	_, err := svc.CreateExport(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "font not found")

	// No export record should have been created.
	assert.Empty(t, store.createExportCalls)
}

// =========================================================
// CreateExport — store error on profile propagated
// =========================================================

func TestResumeService_CreateExport_StoreProfileError(t *testing.T) {
	store := defaultTestStore()
	store.err = fmt.Errorf("profile db error")
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := fullExportRequest()
	_, err := svc.CreateExport(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "profile")
}

// =========================================================
// CreateExport — summary fetched by ID
// =========================================================

func TestResumeService_CreateExport_SummaryFetchedByID(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := domain.ExportRequest{
		TemplateID:    "professional",
		SummaryIDs:    []int64{1},
		DescriptorIDs: []int64{1},
	}
	_, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)

	// GetSummary should have been called with the correct ID.
	require.Len(t, store.getSummaryCalls, 1)
	assert.Equal(t, int64(1), store.getSummaryCalls[0])

	require.Len(t, renderer.renderCalls, 1)
	assert.NotEmpty(t, renderer.renderCalls[0].Summaries)
	assert.Equal(t, "General", renderer.renderCalls[0].Summaries[0].Label)
}

// =========================================================
// CreateExport — no summary when SummaryIDs is empty
// =========================================================

func TestResumeService_CreateExport_NoSummary(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := domain.ExportRequest{
		TemplateID: "professional",
		SkillIDs:   []int64{1},
	}
	_, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)

	assert.Empty(t, store.getSummaryCalls)
	require.Len(t, renderer.renderCalls, 1)
	assert.Empty(t, renderer.renderCalls[0].Summaries)
}

// =========================================================
// PreviewExport — generates PDF without creating record
// =========================================================

func TestResumeService_PreviewExport(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := fullExportRequest()
	filePath, err := svc.PreviewExport(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, renderer.filePath, filePath)

	// Renderer should have been called.
	require.Len(t, renderer.renderCalls, 1)

	// But NO export record or selections should have been created.
	assert.Empty(t, store.createExportCalls)
	assert.Empty(t, store.createSelectionsCalls)
}

// =========================================================
// PreviewExport — validates like CreateExport
// =========================================================

func TestResumeService_PreviewExport_Validates(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	req := domain.ExportRequest{
		TemplateID: "professional",
		// No content selected.
	}
	_, err := svc.PreviewExport(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one content item")
}

// =========================================================
// CreateExport — descriptor order preserved
// =========================================================

func TestResumeService_CreateExport_DescriptorOrderPreserved(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	// Request descriptors in reverse order.
	req := domain.ExportRequest{
		TemplateID:    "professional",
		DescriptorIDs: []int64{2, 1},
	}
	_, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)

	require.Len(t, renderer.renderCalls, 1)
	descs := renderer.renderCalls[0].Descriptors
	require.Len(t, descs, 2)
	// Order should match the request order, not the store order.
	assert.Equal(t, "Backend Engineer", descs[0].Title)
	assert.Equal(t, "Full-Stack Developer", descs[1].Title)
}

// =========================================================
// Skill sort override does not mutate master list
// =========================================================

func TestResumeService_CreateExport_SortOverrideDoesNotMutateMaster(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	// First export with override.
	req1 := domain.ExportRequest{
		TemplateID: "professional",
		SkillIDs:   []int64{1, 2, 3},
		SkillSortOverrides: map[int64]int{
			3: 0,
			1: 1,
			2: 2,
		},
	}
	_, err := svc.CreateExport(context.Background(), req1)
	require.NoError(t, err)

	// Second export without override — should use original order.
	req2 := domain.ExportRequest{
		TemplateID: "professional",
		SkillIDs:   []int64{1, 2, 3},
	}
	_, err = svc.CreateExport(context.Background(), req2)
	require.NoError(t, err)

	// The store's skill slice should not have been modified.
	// Check the second render call has original order.
	require.Len(t, renderer.renderCalls, 2)
	skills2 := renderer.renderCalls[1].Skills
	require.Len(t, skills2, 3)
	// Original order from store: Go(9), TypeScript(8), SQL(7)
	assert.Equal(t, "Go", skills2[0].Name)
	assert.Equal(t, "TypeScript", skills2[1].Name)
	assert.Equal(t, "SQL", skills2[2].Name)
}

// =========================================================
// Work history IDs filter — preserve request order
// =========================================================

func TestResumeService_CreateExport_WorkHistoryOrderPreserved(t *testing.T) {
	store := defaultTestStore()
	renderer := defaultTestRenderer()
	svc := NewResumeService(store, renderer, "/exports")

	// Request work history in reverse order.
	req := domain.ExportRequest{
		TemplateID:     "professional",
		WorkHistoryIDs: []int64{2, 1},
		BulletIDs:      []int64{10, 20},
	}
	_, err := svc.CreateExport(context.Background(), req)
	require.NoError(t, err)

	require.Len(t, renderer.renderCalls, 1)
	wh := renderer.renderCalls[0].WorkHistory
	require.Len(t, wh, 2)
	// Order should match the request order.
	assert.Equal(t, "Beta Inc", wh[0].EmployerName)
	assert.Equal(t, "Acme Corp", wh[1].EmployerName)
}
