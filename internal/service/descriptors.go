package service

import (
	"context"

	"cut-the-bs/internal/domain"
)

// DescriptorStore defines the persistence operations required by
// DescriptorService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type DescriptorStore interface {
	ListDescriptors(ctx context.Context) ([]domain.RoleDescriptor, error)
	CreateDescriptor(ctx context.Context, title string) (domain.RoleDescriptor, error)
	UpdateDescriptor(ctx context.Context, id int64, title string) (domain.RoleDescriptor, error)
	DeleteDescriptor(ctx context.Context, id int64) error
	ReorderDescriptors(ctx context.Context, orderedIDs []int64) error
}

// DescriptorService provides business-logic operations for role
// descriptors. It validates inputs before delegating to the store.
type DescriptorService struct {
	store DescriptorStore
}

// NewDescriptorService creates a DescriptorService backed by the
// given store.
func NewDescriptorService(store DescriptorStore) *DescriptorService {
	return &DescriptorService{store: store}
}

// ListDescriptors returns all role descriptors ordered by
// sort_order.
func (s *DescriptorService) ListDescriptors(ctx context.Context) ([]domain.RoleDescriptor, error) {
	return s.store.ListDescriptors(ctx)
}

// CreateDescriptor validates the title and creates a new role
// descriptor.
func (s *DescriptorService) CreateDescriptor(ctx context.Context, title string) (domain.RoleDescriptor, error) {
	if err := domain.ValidateRequired(title, "title"); err != nil {
		return domain.RoleDescriptor{}, err
	}
	return s.store.CreateDescriptor(ctx, title)
}

// UpdateDescriptor validates the title and updates an existing
// descriptor.
func (s *DescriptorService) UpdateDescriptor(ctx context.Context, id int64, title string) (domain.RoleDescriptor, error) {
	if err := domain.ValidateRequired(title, "title"); err != nil {
		return domain.RoleDescriptor{}, err
	}
	return s.store.UpdateDescriptor(ctx, id, title)
}

// DeleteDescriptor deletes a role descriptor by ID.
func (s *DescriptorService) DeleteDescriptor(ctx context.Context, id int64) error {
	return s.store.DeleteDescriptor(ctx, id)
}

// ReorderDescriptors updates sort_order for all descriptors based
// on the provided ordered slice of IDs.
func (s *DescriptorService) ReorderDescriptors(ctx context.Context, orderedIDs []int64) error {
	return s.store.ReorderDescriptors(ctx, orderedIDs)
}
