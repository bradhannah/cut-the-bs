package service

import (
	"context"

	"cut-the-bs/internal/domain"
)

// ProfileStore defines the persistence operations required by
// ProfileService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type ProfileStore interface {
	GetProfile(ctx context.Context) (domain.UserProfile, error)
	UpdateProfile(ctx context.Context, profile domain.UserProfile) (domain.UserProfile, error)
	ListProfileLinks(ctx context.Context) ([]domain.ProfileLink, error)
	CreateProfileLink(ctx context.Context, input domain.ProfileLinkInput) (domain.ProfileLink, error)
	UpdateProfileLink(ctx context.Context, id int64, input domain.ProfileLinkInput) (domain.ProfileLink, error)
	DeleteProfileLink(ctx context.Context, id int64) error
	ReorderProfileLinks(ctx context.Context, orderedIDs []int64) error
}

// ProfileService provides business-logic operations for the user's
// profile and profile links. It validates inputs before delegating
// to the store.
type ProfileService struct {
	store ProfileStore
}

// NewProfileService creates a ProfileService backed by the given
// store.
func NewProfileService(store ProfileStore) *ProfileService {
	return &ProfileService{store: store}
}

// --- User Profile ---

// GetProfile returns the user's profile. If no profile exists, the
// store auto-creates a default empty one.
func (s *ProfileService) GetProfile(ctx context.Context) (domain.UserProfile, error) {
	return s.store.GetProfile(ctx)
}

// UpdateProfile validates required fields (full_name, email) and
// delegates to the store. Phone and location are optional.
func (s *ProfileService) UpdateProfile(ctx context.Context, profile domain.UserProfile) (domain.UserProfile, error) {
	if err := domain.ValidateRequired(profile.FullName, "full name"); err != nil {
		return domain.UserProfile{}, err
	}
	if err := domain.ValidateEmail(profile.Email); err != nil {
		return domain.UserProfile{}, err
	}
	return s.store.UpdateProfile(ctx, profile)
}

// --- Profile Links ---

// ListProfileLinks returns all profile links ordered by sort_order.
func (s *ProfileService) ListProfileLinks(ctx context.Context) ([]domain.ProfileLink, error) {
	return s.store.ListProfileLinks(ctx)
}

// CreateProfileLink validates the input and creates a new profile
// link.
func (s *ProfileService) CreateProfileLink(ctx context.Context, input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	if err := domain.ValidateProfileLinkInput(input); err != nil {
		return domain.ProfileLink{}, err
	}
	return s.store.CreateProfileLink(ctx, input)
}

// UpdateProfileLink validates the input and updates an existing
// profile link.
func (s *ProfileService) UpdateProfileLink(ctx context.Context, id int64, input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	if err := domain.ValidateProfileLinkInput(input); err != nil {
		return domain.ProfileLink{}, err
	}
	return s.store.UpdateProfileLink(ctx, id, input)
}

// DeleteProfileLink deletes a profile link by ID.
func (s *ProfileService) DeleteProfileLink(ctx context.Context, id int64) error {
	return s.store.DeleteProfileLink(ctx, id)
}

// ReorderProfileLinks updates the sort_order of all profile links
// based on the provided ordered slice of IDs.
func (s *ProfileService) ReorderProfileLinks(ctx context.Context, orderedIDs []int64) error {
	return s.store.ReorderProfileLinks(ctx, orderedIDs)
}
