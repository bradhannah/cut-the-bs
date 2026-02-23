package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =================================================================
// Job Application CRUD
// =================================================================

func TestCreateApplication(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	app, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "Software Engineer",
		JobPostingURL: "https://example.com/jobs/123",
		DateApplied:   "2025-01-15",
		FitIndicator:  domain.FitStrong,
		Notes:         "Referred by Jane",
	})
	require.NoError(t, err)
	assert.NotZero(t, app.ID)
	assert.Equal(t, "Acme Corp", app.CompanyName)
	assert.Equal(t, "Software Engineer", app.PositionTitle)
	assert.Equal(t, "https://example.com/jobs/123", app.JobPostingURL)
	assert.Equal(t, "2025-01-15", app.DateApplied)
	assert.Equal(t, domain.StatusApplied, app.Status)
	assert.Equal(t, domain.FitStrong, app.FitIndicator)
	assert.Equal(t, "Referred by Jane", app.Notes)
	assert.Nil(t, app.ResumeExportID)
	assert.Nil(t, app.CoverLetterTemplateID)
	assert.Nil(t, app.CoverLetterLatestExportID)
	assert.NotEmpty(t, app.CreatedAt)
}

func TestListApplications(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	_, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)

	_, err = s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "TechCo",
		PositionTitle: "Backend Dev",
		DateApplied:   "2025-02-01",
	})
	require.NoError(t, err)

	apps, err := s.ListApplications(ctx)
	require.NoError(t, err)
	require.Len(t, apps, 2)
	// Should be ordered by date_applied DESC.
	assert.Equal(t, "TechCo", apps[0].CompanyName)
	assert.Equal(t, "Acme Corp", apps[1].CompanyName)
}

func TestListApplications_EmptyReturnsEmptySlice(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	apps, err := s.ListApplications(ctx)
	require.NoError(t, err)
	require.NotNil(t, apps)
	assert.Empty(t, apps)
}

func TestGetApplication(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	created, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
		FitIndicator:  domain.FitPossible,
	})
	require.NoError(t, err)

	got, err := s.GetApplication(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Acme Corp", got.CompanyName)
	assert.Equal(t, domain.FitPossible, got.FitIndicator)
}

func TestGetApplication_NotFound(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	_, err := s.GetApplication(ctx, 9999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "9999")
}

func TestUpdateApplication(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	created, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)

	updated, err := s.UpdateApplication(ctx, created.ID, domain.ApplicationInput{
		CompanyName:   "Acme Corp Inc",
		PositionTitle: "Senior SWE",
		JobPostingURL: "https://example.com/jobs/456",
		DateApplied:   "2025-01-20",
		FitIndicator:  domain.FitPerfect,
		Notes:         "Updated notes",
	})
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp Inc", updated.CompanyName)
	assert.Equal(t, "Senior SWE", updated.PositionTitle)
	assert.Equal(t, "https://example.com/jobs/456", updated.JobPostingURL)
	assert.Equal(t, "2025-01-20", updated.DateApplied)
	assert.Equal(t, domain.FitPerfect, updated.FitIndicator)
	assert.Equal(t, "Updated notes", updated.Notes)
}

func TestUpdateApplication_NotFound(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	_, err := s.UpdateApplication(ctx, 9999, domain.ApplicationInput{
		CompanyName:   "Acme",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteApplication(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	created, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)

	err = s.DeleteApplication(ctx, created.ID)
	require.NoError(t, err)

	_, err = s.GetApplication(ctx, created.ID)
	require.Error(t, err)
}

func TestDeleteApplication_NotFound(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	err := s.DeleteApplication(ctx, 9999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// Status Transitions
// =================================================================

func TestUpdateApplicationStatus(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	created, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.StatusApplied, created.Status)

	updated, err := s.UpdateApplicationStatus(ctx, created.ID, domain.StatusPhoneScreen)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPhoneScreen, updated.Status)
}

func TestUpdateApplicationStatus_RecordsHistory(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	created, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)

	_, err = s.UpdateApplicationStatus(ctx, created.ID, domain.StatusScreening)
	require.NoError(t, err)

	_, err = s.UpdateApplicationStatus(ctx, created.ID, domain.StatusInterviewScheduled)
	require.NoError(t, err)

	history, err := s.GetApplicationHistory(ctx, created.ID)
	require.NoError(t, err)
	require.Len(t, history, 2)

	assert.Equal(t, domain.StatusApplied, history[0].FromStatus)
	assert.Equal(t, domain.StatusScreening, history[0].ToStatus)
	assert.Equal(t, domain.StatusScreening, history[1].FromStatus)
	assert.Equal(t, domain.StatusInterviewScheduled, history[1].ToStatus)
}

func TestUpdateApplicationStatus_NotFound(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	_, err := s.UpdateApplicationStatus(ctx, 9999, domain.StatusScreening)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateApplicationFit(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	created, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)

	updated, err := s.UpdateApplicationFit(ctx, created.ID, domain.FitPerfect)
	require.NoError(t, err)
	assert.Equal(t, domain.FitPerfect, updated.FitIndicator)
}

func TestUpdateApplicationFit_NotFound(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	_, err := s.UpdateApplicationFit(ctx, 9999, domain.FitStrong)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// Search
// =================================================================

func TestSearchApplications(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	_, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "Software Engineer",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)

	_, err = s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "TechCo",
		PositionTitle: "Backend Developer",
		DateApplied:   "2025-02-01",
	})
	require.NoError(t, err)

	// Search by company.
	results, err := s.SearchApplications(ctx, "acme")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Acme Corp", results[0].CompanyName)

	// Search by position.
	results, err = s.SearchApplications(ctx, "backend")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "TechCo", results[0].CompanyName)

	// Search with no matches.
	results, err = s.SearchApplications(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, results)
}

// =================================================================
// Status History
// =================================================================

func TestGetApplicationHistory_Empty(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	created, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)

	history, err := s.GetApplicationHistory(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, history)
	assert.Empty(t, history)
}

func TestCreateStatusChange(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	created, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)

	change, err := s.CreateStatusChange(ctx, domain.StatusChange{
		ApplicationID: created.ID,
		FromStatus:    domain.StatusApplied,
		ToStatus:      domain.StatusScreening,
	})
	require.NoError(t, err)
	assert.NotZero(t, change.ID)
	assert.Equal(t, created.ID, change.ApplicationID)
	assert.Equal(t, domain.StatusApplied, change.FromStatus)
	assert.Equal(t, domain.StatusScreening, change.ToStatus)
	assert.NotEmpty(t, change.ChangedAt)
}

func TestDeleteApplication_CascadesHistory(t *testing.T) {
	s := testStore(t)
	require.NoError(t, Migrate(s))
	ctx := context.Background()

	created, err := s.CreateApplication(ctx, domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)

	_, err = s.UpdateApplicationStatus(ctx, created.ID, domain.StatusScreening)
	require.NoError(t, err)

	err = s.DeleteApplication(ctx, created.ID)
	require.NoError(t, err)

	// History should be gone too (CASCADE).
	history, err := s.GetApplicationHistory(ctx, created.ID)
	require.NoError(t, err)
	assert.Empty(t, history)
}
