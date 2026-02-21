package service

import (
	"context"
	"regexp"
	"strings"

	"cut-the-bs/internal/domain"
)

// CoreExpertiseStore defines the persistence operations required by
// CoreExpertiseService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type CoreExpertiseStore interface {
	ListCoreExpertise(ctx context.Context) ([]domain.CoreExpertise, error)
	CreateCoreExpertise(ctx context.Context, label string) (domain.CoreExpertise, error)
	UpdateCoreExpertise(ctx context.Context, id int64, label string) (domain.CoreExpertise, error)
	DeleteCoreExpertise(ctx context.Context, id int64) error
	ReorderCoreExpertise(ctx context.Context, orderedIDs []int64) error
}

// CoreExpertiseService provides business-logic operations for core
// expertise items. It validates inputs before delegating to the store.
type CoreExpertiseService struct {
	store CoreExpertiseStore
}

// NewCoreExpertiseService creates a CoreExpertiseService backed by
// the given store.
func NewCoreExpertiseService(store CoreExpertiseStore) *CoreExpertiseService {
	return &CoreExpertiseService{store: store}
}

// ListCoreExpertise returns all core expertise items ordered by
// sort_order.
func (s *CoreExpertiseService) ListCoreExpertise(ctx context.Context) ([]domain.CoreExpertise, error) {
	return s.store.ListCoreExpertise(ctx)
}

// CreateCoreExpertise validates the label and creates a new core
// expertise item.
func (s *CoreExpertiseService) CreateCoreExpertise(ctx context.Context, label string) (domain.CoreExpertise, error) {
	if err := domain.ValidateRequired(label, "label"); err != nil {
		return domain.CoreExpertise{}, err
	}
	return s.store.CreateCoreExpertise(ctx, label)
}

// UpdateCoreExpertise validates the label and updates an existing
// core expertise item.
func (s *CoreExpertiseService) UpdateCoreExpertise(ctx context.Context, id int64, label string) (domain.CoreExpertise, error) {
	if err := domain.ValidateRequired(label, "label"); err != nil {
		return domain.CoreExpertise{}, err
	}
	return s.store.UpdateCoreExpertise(ctx, id, label)
}

// DeleteCoreExpertise deletes a core expertise item by ID.
func (s *CoreExpertiseService) DeleteCoreExpertise(ctx context.Context, id int64) error {
	return s.store.DeleteCoreExpertise(ctx, id)
}

// ReorderCoreExpertise updates sort_order for all core expertise
// items based on the provided ordered slice of IDs.
func (s *CoreExpertiseService) ReorderCoreExpertise(ctx context.Context, orderedIDs []int64) error {
	return s.store.ReorderCoreExpertise(ctx, orderedIDs)
}

// splitPattern matches pipe, comma, or newline delimiters for
// splitting pasted core expertise text.
var splitPattern = regexp.MustCompile(`[|,\n]`)

// SplitCoreExpertiseText accepts a block of text and returns
// individual core expertise labels by splitting on pipe (|), comma,
// or newline delimiters. Whitespace is trimmed and blank entries are
// filtered out. This is a preview operation — no persistence occurs.
func (s *CoreExpertiseService) SplitCoreExpertiseText(text string) []string {
	parts := splitPattern.Split(text, -1)
	result := make([]string, 0)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
