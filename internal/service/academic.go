package service

import (
	"context"

	"cut-the-bs/internal/domain"
)

// AcademicStore defines the persistence operations required by
// AcademicService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type AcademicStore interface {
	ListAcademicCredentials(ctx context.Context) ([]domain.AcademicCredential, error)
	CreateAcademicCredential(ctx context.Context, input domain.AcademicInput) (domain.AcademicCredential, error)
	UpdateAcademicCredential(ctx context.Context, id int64, input domain.AcademicInput) (domain.AcademicCredential, error)
	DeleteAcademicCredential(ctx context.Context, id int64) error
	ReorderAcademicCredentials(ctx context.Context, orderedIDs []int64) error
	ListCertifications(ctx context.Context) ([]domain.Certification, error)
	CreateCertification(ctx context.Context, input domain.CertificationInput) (domain.Certification, error)
	UpdateCertification(ctx context.Context, id int64, input domain.CertificationInput) (domain.Certification, error)
	DeleteCertification(ctx context.Context, id int64) error
	ReorderCertifications(ctx context.Context, orderedIDs []int64) error
}

// AcademicService provides business-logic operations for academic
// credentials and certifications. It validates inputs before
// delegating to the store.
type AcademicService struct {
	store AcademicStore
}

// NewAcademicService creates an AcademicService backed by the given
// store.
func NewAcademicService(store AcademicStore) *AcademicService {
	return &AcademicService{store: store}
}

// --- Academic Credentials ---

// ListAcademicCredentials returns all academic records ordered by
// sort_order.
func (s *AcademicService) ListAcademicCredentials(ctx context.Context) ([]domain.AcademicCredential, error) {
	return s.store.ListAcademicCredentials(ctx)
}

// CreateAcademicCredential validates the input and creates a new
// academic record.
func (s *AcademicService) CreateAcademicCredential(ctx context.Context, input domain.AcademicInput) (domain.AcademicCredential, error) {
	if err := domain.ValidateAcademicInput(input); err != nil {
		return domain.AcademicCredential{}, err
	}
	return s.store.CreateAcademicCredential(ctx, input)
}

// UpdateAcademicCredential validates the input and updates an
// existing academic record.
func (s *AcademicService) UpdateAcademicCredential(ctx context.Context, id int64, input domain.AcademicInput) (domain.AcademicCredential, error) {
	if err := domain.ValidateAcademicInput(input); err != nil {
		return domain.AcademicCredential{}, err
	}
	return s.store.UpdateAcademicCredential(ctx, id, input)
}

// DeleteAcademicCredential deletes an academic record by ID.
func (s *AcademicService) DeleteAcademicCredential(ctx context.Context, id int64) error {
	return s.store.DeleteAcademicCredential(ctx, id)
}

// ReorderAcademicCredentials updates sort_order for all academic
// credentials based on the provided ordered slice of IDs.
func (s *AcademicService) ReorderAcademicCredentials(ctx context.Context, orderedIDs []int64) error {
	return s.store.ReorderAcademicCredentials(ctx, orderedIDs)
}

// --- Certifications ---

// ListCertifications returns all certifications with computed
// active/inactive status.
func (s *AcademicService) ListCertifications(ctx context.Context) ([]domain.Certification, error) {
	return s.store.ListCertifications(ctx)
}

// CreateCertification validates the input and creates a new
// certification.
func (s *AcademicService) CreateCertification(ctx context.Context, input domain.CertificationInput) (domain.Certification, error) {
	if err := domain.ValidateCertificationInput(input); err != nil {
		return domain.Certification{}, err
	}
	return s.store.CreateCertification(ctx, input)
}

// UpdateCertification validates the input and updates an existing
// certification.
func (s *AcademicService) UpdateCertification(ctx context.Context, id int64, input domain.CertificationInput) (domain.Certification, error) {
	if err := domain.ValidateCertificationInput(input); err != nil {
		return domain.Certification{}, err
	}
	return s.store.UpdateCertification(ctx, id, input)
}

// DeleteCertification deletes a certification by ID.
func (s *AcademicService) DeleteCertification(ctx context.Context, id int64) error {
	return s.store.DeleteCertification(ctx, id)
}

// ReorderCertifications updates sort_order for all certifications
// based on the provided ordered slice of IDs.
func (s *AcademicService) ReorderCertifications(ctx context.Context, orderedIDs []int64) error {
	return s.store.ReorderCertifications(ctx, orderedIDs)
}
