package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockApplicationStore implements ApplicationStore for testing.
type mockApplicationStore struct {
	applications []domain.JobApplication
	application  domain.JobApplication
	history      []domain.StatusChange
	statusChange domain.StatusChange
	err          error
	promptValues map[string]string

	// call tracking
	createCalls       []domain.ApplicationInput
	updateCalls       []updateApplicationCall
	statusCalls       []statusUpdateCall
	fitCalls          []fitUpdateCall
	deleteCalls       []int64
	searchCalls       []string
	historyCalls      []int64
	statusChangeCalls []domain.StatusChange
	promptGetCalls    []promptValuesGetCall
	promptSaveCalls   []promptValuesSaveCall
}

type promptValuesGetCall struct {
	ApplicationID int64
	TemplateID    int64
}

type promptValuesSaveCall struct {
	ApplicationID int64
	TemplateID    int64
	Values        map[string]string
}

type updateApplicationCall struct {
	ID    int64
	Input domain.ApplicationInput
}

type statusUpdateCall struct {
	ID     int64
	Status string
}

type fitUpdateCall struct {
	ID  int64
	Fit string
}

func (m *mockApplicationStore) ListApplications(_ context.Context) ([]domain.JobApplication, error) {
	return m.applications, m.err
}

func (m *mockApplicationStore) SearchApplications(_ context.Context, query string) ([]domain.JobApplication, error) {
	m.searchCalls = append(m.searchCalls, query)
	return m.applications, m.err
}

func (m *mockApplicationStore) GetApplication(_ context.Context, id int64) (domain.JobApplication, error) {
	if m.err != nil {
		return domain.JobApplication{}, m.err
	}
	return m.application, nil
}

func (m *mockApplicationStore) CreateApplication(_ context.Context, input domain.ApplicationInput) (domain.JobApplication, error) {
	m.createCalls = append(m.createCalls, input)
	if m.err != nil {
		return domain.JobApplication{}, m.err
	}
	return m.application, nil
}

func (m *mockApplicationStore) UpdateApplication(_ context.Context, id int64, input domain.ApplicationInput) (domain.JobApplication, error) {
	m.updateCalls = append(m.updateCalls, updateApplicationCall{ID: id, Input: input})
	if m.err != nil {
		return domain.JobApplication{}, m.err
	}
	return m.application, nil
}

func (m *mockApplicationStore) UpdateApplicationStatus(_ context.Context, id int64, status string) (domain.JobApplication, error) {
	m.statusCalls = append(m.statusCalls, statusUpdateCall{ID: id, Status: status})
	if m.err != nil {
		return domain.JobApplication{}, m.err
	}
	return m.application, nil
}

func (m *mockApplicationStore) UpdateApplicationFit(_ context.Context, id int64, fit string) (domain.JobApplication, error) {
	m.fitCalls = append(m.fitCalls, fitUpdateCall{ID: id, Fit: fit})
	if m.err != nil {
		return domain.JobApplication{}, m.err
	}
	return m.application, nil
}

func (m *mockApplicationStore) DeleteApplication(_ context.Context, id int64) error {
	m.deleteCalls = append(m.deleteCalls, id)
	return m.err
}

func (m *mockApplicationStore) GetApplicationHistory(_ context.Context, appID int64) ([]domain.StatusChange, error) {
	m.historyCalls = append(m.historyCalls, appID)
	return m.history, m.err
}

func (m *mockApplicationStore) CreateStatusChange(_ context.Context, change domain.StatusChange) (domain.StatusChange, error) {
	m.statusChangeCalls = append(m.statusChangeCalls, change)
	return m.statusChange, m.err
}

func (m *mockApplicationStore) GetApplicationPromptValues(
	_ context.Context,
	applicationID int64,
	templateID int64,
) (map[string]string, error) {
	m.promptGetCalls = append(m.promptGetCalls, promptValuesGetCall{
		ApplicationID: applicationID,
		TemplateID:    templateID,
	})
	if m.err != nil {
		return nil, m.err
	}
	if m.promptValues == nil {
		return map[string]string{}, nil
	}
	result := make(map[string]string, len(m.promptValues))
	for k, v := range m.promptValues {
		result[k] = v
	}
	return result, nil
}

func (m *mockApplicationStore) SaveApplicationPromptValues(
	_ context.Context,
	applicationID int64,
	templateID int64,
	values map[string]string,
) error {
	copyValues := make(map[string]string, len(values))
	for k, v := range values {
		copyValues[k] = v
	}
	m.promptSaveCalls = append(m.promptSaveCalls, promptValuesSaveCall{
		ApplicationID: applicationID,
		TemplateID:    templateID,
		Values:        copyValues,
	})
	if m.err != nil {
		return m.err
	}
	m.promptValues = copyValues
	return nil
}

// =================================================================
// ListApplications
// =================================================================

func TestApplicationService_ListApplications_Success(t *testing.T) {
	store := &mockApplicationStore{
		applications: []domain.JobApplication{
			{ID: 1, CompanyName: "Acme Corp", PositionTitle: "SWE"},
			{ID: 2, CompanyName: "TechCo", PositionTitle: "Backend Dev"},
		},
	}
	svc := NewApplicationService(store)

	apps, err := svc.ListApplications(context.Background())
	require.NoError(t, err)
	require.Len(t, apps, 2)
	assert.Equal(t, "Acme Corp", apps[0].CompanyName)
}

func TestApplicationService_ListApplications_Empty(t *testing.T) {
	store := &mockApplicationStore{applications: []domain.JobApplication{}}
	svc := NewApplicationService(store)

	apps, err := svc.ListApplications(context.Background())
	require.NoError(t, err)
	assert.Empty(t, apps)
}

func TestApplicationService_ListApplications_StoreError(t *testing.T) {
	store := &mockApplicationStore{err: fmt.Errorf("db error")}
	svc := NewApplicationService(store)

	_, err := svc.ListApplications(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// =================================================================
// SearchApplications
// =================================================================

func TestApplicationService_SearchApplications_Success(t *testing.T) {
	store := &mockApplicationStore{
		applications: []domain.JobApplication{
			{ID: 1, CompanyName: "Acme Corp"},
		},
	}
	svc := NewApplicationService(store)

	apps, err := svc.SearchApplications(context.Background(), "acme")
	require.NoError(t, err)
	require.Len(t, apps, 1)
	require.Len(t, store.searchCalls, 1)
	assert.Equal(t, "acme", store.searchCalls[0])
}

func TestApplicationService_SearchApplications_EmptyQueryReturnsList(t *testing.T) {
	store := &mockApplicationStore{
		applications: []domain.JobApplication{
			{ID: 1}, {ID: 2},
		},
	}
	svc := NewApplicationService(store)

	apps, err := svc.SearchApplications(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, apps, 2)
	// Should call ListApplications, not SearchApplications.
	assert.Empty(t, store.searchCalls)
}

// =================================================================
// CreateApplication — validation
// =================================================================

func TestApplicationService_CreateApplication_EmptyCompanyName(t *testing.T) {
	store := &mockApplicationStore{}
	svc := NewApplicationService(store)

	_, err := svc.CreateApplication(context.Background(), domain.ApplicationInput{
		CompanyName:   "",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "company name")
	assert.Empty(t, store.createCalls)
}

func TestApplicationService_CreateApplication_EmptyPositionTitle(t *testing.T) {
	store := &mockApplicationStore{}
	svc := NewApplicationService(store)

	_, err := svc.CreateApplication(context.Background(), domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "",
		DateApplied:   "2025-01-15",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "position title")
	assert.Empty(t, store.createCalls)
}

func TestApplicationService_CreateApplication_EmptyDateApplied(t *testing.T) {
	store := &mockApplicationStore{}
	svc := NewApplicationService(store)

	_, err := svc.CreateApplication(context.Background(), domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "date applied")
	assert.Empty(t, store.createCalls)
}

func TestApplicationService_CreateApplication_InvalidFitIndicator(t *testing.T) {
	store := &mockApplicationStore{}
	svc := NewApplicationService(store)

	_, err := svc.CreateApplication(context.Background(), domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
		FitIndicator:  "Invalid Fit",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fit indicator")
	assert.Empty(t, store.createCalls)
}

// =================================================================
// CreateApplication — happy path
// =================================================================

func TestApplicationService_CreateApplication_Success(t *testing.T) {
	store := &mockApplicationStore{
		application: domain.JobApplication{
			ID: 1, CompanyName: "Acme Corp", PositionTitle: "SWE",
			DateApplied: "2025-01-15", Status: domain.StatusApplied,
		},
	}
	svc := NewApplicationService(store)

	app, err := svc.CreateApplication(context.Background(), domain.ApplicationInput{
		CompanyName:   "Acme Corp",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", app.CompanyName)
	assert.Equal(t, domain.StatusApplied, app.Status)
	require.Len(t, store.createCalls, 1)
}

func TestApplicationService_CreateApplication_WithOptionalFit(t *testing.T) {
	store := &mockApplicationStore{
		application: domain.JobApplication{
			ID: 1, CompanyName: "Acme", FitIndicator: domain.FitStrong,
		},
	}
	svc := NewApplicationService(store)

	_, err := svc.CreateApplication(context.Background(), domain.ApplicationInput{
		CompanyName:   "Acme",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
		FitIndicator:  domain.FitStrong,
	})
	require.NoError(t, err)
	require.Len(t, store.createCalls, 1)
	assert.Equal(t, domain.FitStrong, store.createCalls[0].FitIndicator)
}

func TestApplicationService_CreateApplication_EmptyFitIsValid(t *testing.T) {
	store := &mockApplicationStore{
		application: domain.JobApplication{ID: 1},
	}
	svc := NewApplicationService(store)

	_, err := svc.CreateApplication(context.Background(), domain.ApplicationInput{
		CompanyName:   "Acme",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
		FitIndicator:  "", // empty is valid — not assessed
	})
	require.NoError(t, err)
}

func TestApplicationService_CreateApplication_StoreError(t *testing.T) {
	store := &mockApplicationStore{err: fmt.Errorf("insert error")}
	svc := NewApplicationService(store)

	_, err := svc.CreateApplication(context.Background(), domain.ApplicationInput{
		CompanyName:   "Acme",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert error")
}

// =================================================================
// UpdateApplication
// =================================================================

func TestApplicationService_UpdateApplication_Success(t *testing.T) {
	store := &mockApplicationStore{
		application: domain.JobApplication{
			ID: 1, CompanyName: "Updated Corp",
		},
	}
	svc := NewApplicationService(store)

	app, err := svc.UpdateApplication(context.Background(), 1, domain.ApplicationInput{
		CompanyName:   "Updated Corp",
		PositionTitle: "Senior SWE",
		DateApplied:   "2025-02-01",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Corp", app.CompanyName)
	require.Len(t, store.updateCalls, 1)
	assert.Equal(t, int64(1), store.updateCalls[0].ID)
}

func TestApplicationService_UpdateApplication_ValidationError(t *testing.T) {
	store := &mockApplicationStore{}
	svc := NewApplicationService(store)

	_, err := svc.UpdateApplication(context.Background(), 1, domain.ApplicationInput{
		CompanyName:   "", // required
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "company name")
	assert.Empty(t, store.updateCalls)
}

func TestApplicationService_UpdateApplication_StoreError(t *testing.T) {
	store := &mockApplicationStore{err: fmt.Errorf("not found")}
	svc := NewApplicationService(store)

	_, err := svc.UpdateApplication(context.Background(), 999, domain.ApplicationInput{
		CompanyName:   "Acme",
		PositionTitle: "SWE",
		DateApplied:   "2025-01-15",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// UpdateApplicationStatus
// =================================================================

func TestApplicationService_UpdateApplicationStatus_Success(t *testing.T) {
	store := &mockApplicationStore{
		application: domain.JobApplication{
			ID: 1, Status: domain.StatusScreening,
		},
	}
	svc := NewApplicationService(store)

	app, err := svc.UpdateApplicationStatus(context.Background(), 1, domain.StatusScreening)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusScreening, app.Status)
	require.Len(t, store.statusCalls, 1)
	assert.Equal(t, int64(1), store.statusCalls[0].ID)
	assert.Equal(t, domain.StatusScreening, store.statusCalls[0].Status)
}

func TestApplicationService_UpdateApplicationStatus_InvalidStatus(t *testing.T) {
	store := &mockApplicationStore{}
	svc := NewApplicationService(store)

	_, err := svc.UpdateApplicationStatus(context.Background(), 1, "Not A Status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
	assert.Empty(t, store.statusCalls)
}

func TestApplicationService_UpdateApplicationStatus_StoreError(t *testing.T) {
	store := &mockApplicationStore{err: fmt.Errorf("not found")}
	svc := NewApplicationService(store)

	_, err := svc.UpdateApplicationStatus(context.Background(), 999, domain.StatusScreening)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// UpdateApplicationFit
// =================================================================

func TestApplicationService_UpdateApplicationFit_Success(t *testing.T) {
	store := &mockApplicationStore{
		application: domain.JobApplication{
			ID: 1, FitIndicator: domain.FitPerfect,
		},
	}
	svc := NewApplicationService(store)

	app, err := svc.UpdateApplicationFit(context.Background(), 1, domain.FitPerfect)
	require.NoError(t, err)
	assert.Equal(t, domain.FitPerfect, app.FitIndicator)
	require.Len(t, store.fitCalls, 1)
}

func TestApplicationService_UpdateApplicationFit_EmptyIsValid(t *testing.T) {
	store := &mockApplicationStore{
		application: domain.JobApplication{ID: 1},
	}
	svc := NewApplicationService(store)

	_, err := svc.UpdateApplicationFit(context.Background(), 1, "")
	require.NoError(t, err)
	require.Len(t, store.fitCalls, 1)
	assert.Equal(t, "", store.fitCalls[0].Fit)
}

func TestApplicationService_UpdateApplicationFit_InvalidFit(t *testing.T) {
	store := &mockApplicationStore{}
	svc := NewApplicationService(store)

	_, err := svc.UpdateApplicationFit(context.Background(), 1, "Super Fit")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fit indicator")
	assert.Empty(t, store.fitCalls)
}

func TestApplicationService_UpdateApplicationFit_StoreError(t *testing.T) {
	store := &mockApplicationStore{err: fmt.Errorf("not found")}
	svc := NewApplicationService(store)

	_, err := svc.UpdateApplicationFit(context.Background(), 999, domain.FitStrong)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// DeleteApplication
// =================================================================

func TestApplicationService_DeleteApplication_Success(t *testing.T) {
	store := &mockApplicationStore{}
	svc := NewApplicationService(store)

	err := svc.DeleteApplication(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, store.deleteCalls, 1)
	assert.Equal(t, int64(5), store.deleteCalls[0])
}

func TestApplicationService_DeleteApplication_StoreError(t *testing.T) {
	store := &mockApplicationStore{err: fmt.Errorf("not found")}
	svc := NewApplicationService(store)

	err := svc.DeleteApplication(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// GetApplicationHistory
// =================================================================

func TestApplicationService_GetApplicationHistory_Success(t *testing.T) {
	store := &mockApplicationStore{
		history: []domain.StatusChange{
			{ID: 1, FromStatus: domain.StatusApplied, ToStatus: domain.StatusScreening},
			{ID: 2, FromStatus: domain.StatusScreening, ToStatus: domain.StatusInterviewScheduled},
		},
	}
	svc := NewApplicationService(store)

	history, err := svc.GetApplicationHistory(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, history, 2)
	assert.Equal(t, domain.StatusApplied, history[0].FromStatus)
	assert.Equal(t, domain.StatusScreening, history[0].ToStatus)
}

func TestApplicationService_GetApplicationHistory_Empty(t *testing.T) {
	store := &mockApplicationStore{history: []domain.StatusChange{}}
	svc := NewApplicationService(store)

	history, err := svc.GetApplicationHistory(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, history)
}

func TestApplicationService_GetApplicationHistory_StoreError(t *testing.T) {
	store := &mockApplicationStore{err: fmt.Errorf("db error")}
	svc := NewApplicationService(store)

	_, err := svc.GetApplicationHistory(context.Background(), 1)
	require.Error(t, err)
}

// =================================================================
// GetApplicationStatuses / GetFitIndicators
// =================================================================

func TestApplicationService_GetApplicationStatuses(t *testing.T) {
	svc := NewApplicationService(nil)
	statuses := svc.GetApplicationStatuses()
	assert.Len(t, statuses, 15)
	assert.Equal(t, domain.StatusApplied, statuses[0])
	assert.Equal(t, domain.StatusOnHold, statuses[14])
}

func TestApplicationService_GetFitIndicators(t *testing.T) {
	svc := NewApplicationService(nil)
	fits := svc.GetFitIndicators()
	assert.Len(t, fits, 5)
	assert.Equal(t, domain.FitUnlikely, fits[0])
	assert.Equal(t, domain.FitPerfect, fits[4])
}
