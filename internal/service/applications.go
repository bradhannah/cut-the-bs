package service

import (
	"context"
	"fmt"

	"cut-the-bs/internal/domain"
)

// ApplicationStore defines the persistence operations required by
// ApplicationService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type ApplicationStore interface {
	ListApplications(ctx context.Context) ([]domain.JobApplication, error)
	SearchApplications(ctx context.Context, query string) ([]domain.JobApplication, error)
	GetApplication(ctx context.Context, id int64) (domain.JobApplication, error)
	CreateApplication(ctx context.Context, input domain.ApplicationInput) (domain.JobApplication, error)
	UpdateApplication(ctx context.Context, id int64, input domain.ApplicationInput) (domain.JobApplication, error)
	UpdateApplicationStatus(ctx context.Context, id int64, newStatus string) (domain.JobApplication, error)
	UpdateApplicationFit(ctx context.Context, id int64, fitIndicator string) (domain.JobApplication, error)
	DeleteApplication(ctx context.Context, id int64) error
	GetApplicationHistory(ctx context.Context, applicationID int64) ([]domain.StatusChange, error)
	CreateStatusChange(ctx context.Context, change domain.StatusChange) (domain.StatusChange, error)
	GetApplicationPromptValues(ctx context.Context, applicationID int64, templateID int64) (map[string]string, error)
	SaveApplicationPromptValues(ctx context.Context, applicationID int64, templateID int64, values map[string]string) error
}

// ApplicationService provides business-logic operations for job
// applications. It validates inputs before delegating to the store.
type ApplicationService struct {
	store ApplicationStore
}

// NewApplicationService creates an ApplicationService backed by the
// given store.
func NewApplicationService(store ApplicationStore) *ApplicationService {
	return &ApplicationService{store: store}
}

// ListApplications returns all job applications.
func (s *ApplicationService) ListApplications(ctx context.Context) ([]domain.JobApplication, error) {
	return s.store.ListApplications(ctx)
}

// SearchApplications searches by company name or position title.
func (s *ApplicationService) SearchApplications(ctx context.Context, query string) ([]domain.JobApplication, error) {
	if query == "" {
		return s.store.ListApplications(ctx)
	}
	return s.store.SearchApplications(ctx, query)
}

// GetApplication returns a single job application by ID.
func (s *ApplicationService) GetApplication(ctx context.Context, id int64) (domain.JobApplication, error) {
	if id <= 0 {
		return domain.JobApplication{}, fmt.Errorf("invalid application id: %d", id)
	}
	return s.store.GetApplication(ctx, id)
}

// CreateApplication validates the input and creates a new job
// application.
func (s *ApplicationService) CreateApplication(ctx context.Context, input domain.ApplicationInput) (domain.JobApplication, error) {
	if err := domain.ValidateApplicationInput(input); err != nil {
		return domain.JobApplication{}, err
	}
	return s.store.CreateApplication(ctx, input)
}

// UpdateApplication validates the input and updates an existing
// application.
func (s *ApplicationService) UpdateApplication(ctx context.Context, id int64, input domain.ApplicationInput) (domain.JobApplication, error) {
	if err := domain.ValidateApplicationInput(input); err != nil {
		return domain.JobApplication{}, err
	}
	return s.store.UpdateApplication(ctx, id, input)
}

// UpdateApplicationStatus validates the status and transitions
// an application.
func (s *ApplicationService) UpdateApplicationStatus(ctx context.Context, id int64, newStatus string) (domain.JobApplication, error) {
	if !domain.ValidStatus(newStatus) {
		return domain.JobApplication{}, fmt.Errorf("invalid status: %s", newStatus)
	}
	return s.store.UpdateApplicationStatus(ctx, id, newStatus)
}

// UpdateApplicationFit validates the fit indicator and updates
// an application.
func (s *ApplicationService) UpdateApplicationFit(ctx context.Context, id int64, fitIndicator string) (domain.JobApplication, error) {
	if !domain.ValidFitIndicator(fitIndicator) {
		return domain.JobApplication{}, fmt.Errorf("invalid fit indicator: %s", fitIndicator)
	}
	return s.store.UpdateApplicationFit(ctx, id, fitIndicator)
}

// DeleteApplication deletes a job application by ID.
func (s *ApplicationService) DeleteApplication(ctx context.Context, id int64) error {
	return s.store.DeleteApplication(ctx, id)
}

// GetApplicationHistory returns the status change history for an
// application.
func (s *ApplicationService) GetApplicationHistory(ctx context.Context, id int64) ([]domain.StatusChange, error) {
	return s.store.GetApplicationHistory(ctx, id)
}

// GetApplicationStatuses returns the fixed list of valid statuses.
func (s *ApplicationService) GetApplicationStatuses() []string {
	return domain.AllStatuses
}

// GetFitIndicators returns the fixed list of valid fit indicator
// values.
func (s *ApplicationService) GetFitIndicators() []string {
	return domain.AllFitIndicators
}

// GetApplicationPromptValues returns saved cover-letter prompt values
// for an application+template pair.
func (s *ApplicationService) GetApplicationPromptValues(
	ctx context.Context,
	applicationID int64,
	templateID int64,
) (map[string]string, error) {
	if applicationID <= 0 {
		return nil, fmt.Errorf("invalid application id: %d", applicationID)
	}
	if templateID <= 0 {
		return nil, fmt.Errorf("invalid template id: %d", templateID)
	}
	return s.store.GetApplicationPromptValues(ctx, applicationID, templateID)
}

// SaveApplicationPromptValues replaces saved cover-letter prompt values
// for an application+template pair.
func (s *ApplicationService) SaveApplicationPromptValues(
	ctx context.Context,
	applicationID int64,
	templateID int64,
	values map[string]string,
) error {
	if applicationID <= 0 {
		return fmt.Errorf("invalid application id: %d", applicationID)
	}
	if templateID <= 0 {
		return fmt.Errorf("invalid template id: %d", templateID)
	}
	return s.store.SaveApplicationPromptValues(ctx, applicationID, templateID, values)
}
