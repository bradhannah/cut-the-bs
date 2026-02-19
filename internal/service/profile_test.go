package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProfileStore implements ProfileStore for testing.
type mockProfileStore struct {
	profile domain.UserProfile
	link    domain.ProfileLink
	links   []domain.ProfileLink
	err     error

	// call tracking
	updateProfileCalls []domain.UserProfile
	createLinkCalls    []domain.ProfileLinkInput
	updateLinkCalls    []updateLinkCall
	deleteLinkCalls    []int64
	reorderLinkCalls   [][]int64
}

type updateLinkCall struct {
	ID    int64
	Input domain.ProfileLinkInput
}

func (m *mockProfileStore) GetProfile(_ context.Context) (domain.UserProfile, error) {
	return m.profile, m.err
}

func (m *mockProfileStore) UpdateProfile(_ context.Context, profile domain.UserProfile) (domain.UserProfile, error) {
	m.updateProfileCalls = append(m.updateProfileCalls, profile)
	if m.err != nil {
		return domain.UserProfile{}, m.err
	}
	return m.profile, nil
}

func (m *mockProfileStore) ListProfileLinks(_ context.Context) ([]domain.ProfileLink, error) {
	return m.links, m.err
}

func (m *mockProfileStore) CreateProfileLink(_ context.Context, input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	m.createLinkCalls = append(m.createLinkCalls, input)
	if m.err != nil {
		return domain.ProfileLink{}, m.err
	}
	return m.link, nil
}

func (m *mockProfileStore) UpdateProfileLink(_ context.Context, id int64, input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	m.updateLinkCalls = append(m.updateLinkCalls, updateLinkCall{ID: id, Input: input})
	if m.err != nil {
		return domain.ProfileLink{}, m.err
	}
	return m.link, nil
}

func (m *mockProfileStore) DeleteProfileLink(_ context.Context, id int64) error {
	m.deleteLinkCalls = append(m.deleteLinkCalls, id)
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockProfileStore) ReorderProfileLinks(_ context.Context, orderedIDs []int64) error {
	m.reorderLinkCalls = append(m.reorderLinkCalls, orderedIDs)
	if m.err != nil {
		return m.err
	}
	return nil
}

// =================================================================
// GetProfile
// =================================================================

func TestProfileService_GetProfile_Success(t *testing.T) {
	store := &mockProfileStore{
		profile: domain.UserProfile{
			ID:       1,
			FullName: "Jane Doe",
			Email:    "jane@example.com",
		},
	}
	svc := NewProfileService(store)

	profile, err := svc.GetProfile(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), profile.ID)
	assert.Equal(t, "Jane Doe", profile.FullName)
}

func TestProfileService_GetProfile_StoreError(t *testing.T) {
	store := &mockProfileStore{err: fmt.Errorf("db error")}
	svc := NewProfileService(store)

	_, err := svc.GetProfile(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// =================================================================
// UpdateProfile — validation
// =================================================================

func TestProfileService_UpdateProfile_EmptyFullName(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.UpdateProfile(context.Background(), domain.UserProfile{
		ID:       1,
		FullName: "",
		Email:    "jane@example.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "full name")
	assert.Empty(t, store.updateProfileCalls)
}

func TestProfileService_UpdateProfile_WhitespaceFullName(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.UpdateProfile(context.Background(), domain.UserProfile{
		ID:       1,
		FullName: "   ",
		Email:    "jane@example.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "full name")
	assert.Empty(t, store.updateProfileCalls)
}

func TestProfileService_UpdateProfile_EmptyEmail(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.UpdateProfile(context.Background(), domain.UserProfile{
		ID:       1,
		FullName: "Jane Doe",
		Email:    "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")
	assert.Empty(t, store.updateProfileCalls)
}

func TestProfileService_UpdateProfile_InvalidEmail(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.UpdateProfile(context.Background(), domain.UserProfile{
		ID:       1,
		FullName: "Jane Doe",
		Email:    "not-an-email",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")
	assert.Empty(t, store.updateProfileCalls)
}

func TestProfileService_UpdateProfile_EmailMissingDomain(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.UpdateProfile(context.Background(), domain.UserProfile{
		ID:       1,
		FullName: "Jane Doe",
		Email:    "jane@",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")
	assert.Empty(t, store.updateProfileCalls)
}

// =================================================================
// UpdateProfile — happy path
// =================================================================

func TestProfileService_UpdateProfile_Success(t *testing.T) {
	store := &mockProfileStore{
		profile: domain.UserProfile{
			ID:       1,
			FullName: "Jane Doe",
			Email:    "jane@example.com",
			Phone:    "555-1234",
			Location: "NYC",
		},
	}
	svc := NewProfileService(store)

	profile, err := svc.UpdateProfile(context.Background(), domain.UserProfile{
		ID:       1,
		FullName: "Jane Doe",
		Email:    "jane@example.com",
		Phone:    "555-1234",
		Location: "NYC",
	})
	require.NoError(t, err)
	assert.Equal(t, "Jane Doe", profile.FullName)
	require.Len(t, store.updateProfileCalls, 1)
	assert.Equal(t, "Jane Doe", store.updateProfileCalls[0].FullName)
}

func TestProfileService_UpdateProfile_OptionalFieldsEmpty(t *testing.T) {
	store := &mockProfileStore{
		profile: domain.UserProfile{
			ID:       1,
			FullName: "Jane",
			Email:    "jane@test.com",
		},
	}
	svc := NewProfileService(store)

	profile, err := svc.UpdateProfile(context.Background(), domain.UserProfile{
		ID:       1,
		FullName: "Jane",
		Email:    "jane@test.com",
		Phone:    "",
		Location: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "Jane", profile.FullName)
}

func TestProfileService_UpdateProfile_StoreError(t *testing.T) {
	store := &mockProfileStore{err: fmt.Errorf("connection lost")}
	svc := NewProfileService(store)

	_, err := svc.UpdateProfile(context.Background(), domain.UserProfile{
		ID:       1,
		FullName: "Jane Doe",
		Email:    "jane@example.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection lost")
}

// =================================================================
// CreateProfileLink — validation
// =================================================================

func TestProfileService_CreateProfileLink_EmptyLabel(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.CreateProfileLink(context.Background(), domain.ProfileLinkInput{
		Label: "",
		URL:   "https://linkedin.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
	assert.Empty(t, store.createLinkCalls)
}

func TestProfileService_CreateProfileLink_EmptyURL(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.CreateProfileLink(context.Background(), domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL")
	assert.Empty(t, store.createLinkCalls)
}

func TestProfileService_CreateProfileLink_InvalidURL(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.CreateProfileLink(context.Background(), domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "not-a-url",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL")
	assert.Empty(t, store.createLinkCalls)
}

func TestProfileService_CreateProfileLink_FTPSchemeRejected(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.CreateProfileLink(context.Background(), domain.ProfileLinkInput{
		Label: "Files",
		URL:   "ftp://files.example.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http")
	assert.Empty(t, store.createLinkCalls)
}

// =================================================================
// CreateProfileLink — happy path
// =================================================================

func TestProfileService_CreateProfileLink_Success(t *testing.T) {
	store := &mockProfileStore{
		link: domain.ProfileLink{
			ID:        1,
			Label:     "LinkedIn",
			URL:       "https://linkedin.com/in/jane",
			SortOrder: 1,
		},
	}
	svc := NewProfileService(store)

	link, err := svc.CreateProfileLink(context.Background(), domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "https://linkedin.com/in/jane",
	})
	require.NoError(t, err)
	assert.Equal(t, "LinkedIn", link.Label)
	require.Len(t, store.createLinkCalls, 1)
}

func TestProfileService_CreateProfileLink_StoreError(t *testing.T) {
	store := &mockProfileStore{err: fmt.Errorf("db full")}
	svc := NewProfileService(store)

	_, err := svc.CreateProfileLink(context.Background(), domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "https://linkedin.com/in/jane",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db full")
}

// =================================================================
// UpdateProfileLink — validation
// =================================================================

func TestProfileService_UpdateProfileLink_EmptyLabel(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.UpdateProfileLink(context.Background(), 1, domain.ProfileLinkInput{
		Label: "",
		URL:   "https://linkedin.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
	assert.Empty(t, store.updateLinkCalls)
}

func TestProfileService_UpdateProfileLink_InvalidURL(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	_, err := svc.UpdateProfileLink(context.Background(), 1, domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "ftp://files.example.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http")
	assert.Empty(t, store.updateLinkCalls)
}

// =================================================================
// UpdateProfileLink — happy path
// =================================================================

func TestProfileService_UpdateProfileLink_Success(t *testing.T) {
	store := &mockProfileStore{
		link: domain.ProfileLink{
			ID:        1,
			Label:     "LinkedIn Profile",
			URL:       "https://linkedin.com/in/new",
			SortOrder: 1,
		},
	}
	svc := NewProfileService(store)

	link, err := svc.UpdateProfileLink(context.Background(), 1, domain.ProfileLinkInput{
		Label: "LinkedIn Profile",
		URL:   "https://linkedin.com/in/new",
	})
	require.NoError(t, err)
	assert.Equal(t, "LinkedIn Profile", link.Label)
	require.Len(t, store.updateLinkCalls, 1)
	assert.Equal(t, int64(1), store.updateLinkCalls[0].ID)
}

// =================================================================
// ListProfileLinks
// =================================================================

func TestProfileService_ListProfileLinks_Success(t *testing.T) {
	store := &mockProfileStore{
		links: []domain.ProfileLink{
			{ID: 1, Label: "LinkedIn", SortOrder: 1},
			{ID: 2, Label: "GitHub", SortOrder: 2},
		},
	}
	svc := NewProfileService(store)

	links, err := svc.ListProfileLinks(context.Background())
	require.NoError(t, err)
	require.Len(t, links, 2)
	assert.Equal(t, "LinkedIn", links[0].Label)
}

func TestProfileService_ListProfileLinks_Empty(t *testing.T) {
	store := &mockProfileStore{links: []domain.ProfileLink{}}
	svc := NewProfileService(store)

	links, err := svc.ListProfileLinks(context.Background())
	require.NoError(t, err)
	assert.Empty(t, links)
}

// =================================================================
// DeleteProfileLink
// =================================================================

func TestProfileService_DeleteProfileLink_Success(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	err := svc.DeleteProfileLink(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, store.deleteLinkCalls, 1)
	assert.Equal(t, int64(5), store.deleteLinkCalls[0])
}

func TestProfileService_DeleteProfileLink_StoreError(t *testing.T) {
	store := &mockProfileStore{err: fmt.Errorf("not found")}
	svc := NewProfileService(store)

	err := svc.DeleteProfileLink(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// ReorderProfileLinks
// =================================================================

func TestProfileService_ReorderProfileLinks_Success(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	err := svc.ReorderProfileLinks(context.Background(), []int64{3, 1, 2})
	require.NoError(t, err)
	require.Len(t, store.reorderLinkCalls, 1)
	assert.Equal(t, []int64{3, 1, 2}, store.reorderLinkCalls[0])
}

func TestProfileService_ReorderProfileLinks_EmptySlice(t *testing.T) {
	store := &mockProfileStore{}
	svc := NewProfileService(store)

	err := svc.ReorderProfileLinks(context.Background(), []int64{})
	require.NoError(t, err)
}
