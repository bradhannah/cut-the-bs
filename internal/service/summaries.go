package service

import (
	"context"

	"cut-the-bs/internal/domain"
)

// SummaryStore defines the persistence operations required by
// SummaryService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type SummaryStore interface {
	ListSummaries(ctx context.Context) ([]domain.ProfessionalSummary, error)
	GetSummary(ctx context.Context, id int64) (domain.ProfessionalSummary, error)
	CreateSummary(ctx context.Context, input domain.SummaryInput) (domain.ProfessionalSummary, error)
	UpdateSummary(ctx context.Context, id int64, input domain.SummaryInput) (domain.ProfessionalSummary, error)
	DeleteSummary(ctx context.Context, id int64) error
}

// SummaryService provides business-logic operations for
// professional summary variants. It validates inputs before
// delegating to the store.
type SummaryService struct {
	store SummaryStore
}

// NewSummaryService creates a SummaryService backed by the given
// store.
func NewSummaryService(store SummaryStore) *SummaryService {
	return &SummaryService{store: store}
}

// ListSummaries returns all professional summary variants.
func (s *SummaryService) ListSummaries(ctx context.Context) ([]domain.ProfessionalSummary, error) {
	return s.store.ListSummaries(ctx)
}

// GetSummary returns a single summary by ID.
func (s *SummaryService) GetSummary(ctx context.Context, id int64) (domain.ProfessionalSummary, error) {
	return s.store.GetSummary(ctx, id)
}

// CreateSummary validates the input and creates a new summary
// variant.
func (s *SummaryService) CreateSummary(ctx context.Context, input domain.SummaryInput) (domain.ProfessionalSummary, error) {
	if err := domain.ValidateSummaryInput(input); err != nil {
		return domain.ProfessionalSummary{}, err
	}
	return s.store.CreateSummary(ctx, input)
}

// UpdateSummary validates the input and updates an existing summary.
func (s *SummaryService) UpdateSummary(ctx context.Context, id int64, input domain.SummaryInput) (domain.ProfessionalSummary, error) {
	if err := domain.ValidateSummaryInput(input); err != nil {
		return domain.ProfessionalSummary{}, err
	}
	return s.store.UpdateSummary(ctx, id, input)
}

// DeleteSummary deletes a summary variant by ID.
func (s *SummaryService) DeleteSummary(ctx context.Context, id int64) error {
	return s.store.DeleteSummary(ctx, id)
}
