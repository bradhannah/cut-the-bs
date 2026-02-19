package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- User Profile ---

func TestGetProfile_CreatesDefaultOnFirstGet(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	profile, err := store.GetProfile(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), profile.ID)
	assert.Empty(t, profile.FullName)
	assert.Empty(t, profile.Email)
	assert.Empty(t, profile.Phone)
	assert.Empty(t, profile.Location)
	assert.NotEmpty(t, profile.CreatedAt)
	assert.NotEmpty(t, profile.UpdatedAt)
}

func TestGetProfile_ReturnsSameOnSubsequentCalls(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	first, err := store.GetProfile(ctx)
	require.NoError(t, err)

	second, err := store.GetProfile(ctx)
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
}

func TestUpdateProfile(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	// Ensure profile exists.
	_, err := store.GetProfile(ctx)
	require.NoError(t, err)

	updated, err := store.UpdateProfile(ctx, domain.UserProfile{
		ID:       1,
		FullName: "Jane Doe",
		Email:    "jane@example.com",
		Phone:    "555-1234",
		Location: "New York, NY",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), updated.ID)
	assert.Equal(t, "Jane Doe", updated.FullName)
	assert.Equal(t, "jane@example.com", updated.Email)
	assert.Equal(t, "555-1234", updated.Phone)
	assert.Equal(t, "New York, NY", updated.Location)

	// Verify persistence.
	fetched, err := store.GetProfile(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Jane Doe", fetched.FullName)
	assert.Equal(t, "jane@example.com", fetched.Email)
}

func TestUpdateProfile_PartialFields(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.GetProfile(ctx)
	require.NoError(t, err)

	// Update only name and email, leave phone/location empty.
	updated, err := store.UpdateProfile(ctx, domain.UserProfile{
		ID:       1,
		FullName: "John",
		Email:    "john@test.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "John", updated.FullName)
	assert.Equal(t, "john@test.com", updated.Email)
	assert.Empty(t, updated.Phone)
	assert.Empty(t, updated.Location)
}

// --- Profile Links ---

func TestCreateProfileLink(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	link, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "https://linkedin.com/in/janedoe",
	})
	require.NoError(t, err)
	assert.NotZero(t, link.ID)
	assert.Equal(t, "LinkedIn", link.Label)
	assert.Equal(t, "https://linkedin.com/in/janedoe", link.URL)
	assert.Equal(t, 1, link.SortOrder)
	assert.NotEmpty(t, link.CreatedAt)
}

func TestCreateProfileLink_AutoIncrementsSortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	first, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "https://linkedin.com/in/test",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, first.SortOrder)

	second, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "GitHub",
		URL:   "https://github.com/test",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, second.SortOrder)
}

func TestListProfileLinks_OrderedBySortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "https://linkedin.com",
	})
	require.NoError(t, err)

	_, err = store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "GitHub",
		URL:   "https://github.com",
	})
	require.NoError(t, err)

	_, err = store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "Portfolio",
		URL:   "https://mysite.com",
	})
	require.NoError(t, err)

	links, err := store.ListProfileLinks(ctx)
	require.NoError(t, err)
	require.Len(t, links, 3)
	assert.Equal(t, "LinkedIn", links[0].Label)
	assert.Equal(t, "GitHub", links[1].Label)
	assert.Equal(t, "Portfolio", links[2].Label)
}

func TestListProfileLinks_EmptyReturnsEmptySlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	links, err := store.ListProfileLinks(ctx)
	require.NoError(t, err)
	assert.NotNil(t, links)
	assert.Len(t, links, 0)
}

func TestUpdateProfileLink(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "LinkedIn",
		URL:   "https://linkedin.com/in/old",
	})
	require.NoError(t, err)

	updated, err := store.UpdateProfileLink(ctx, created.ID, domain.ProfileLinkInput{
		Label: "LinkedIn Profile",
		URL:   "https://linkedin.com/in/new",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "LinkedIn Profile", updated.Label)
	assert.Equal(t, "https://linkedin.com/in/new", updated.URL)
	// sort_order should not change on update.
	assert.Equal(t, created.SortOrder, updated.SortOrder)
}

func TestDeleteProfileLink(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "GitHub",
		URL:   "https://github.com/test",
	})
	require.NoError(t, err)

	err = store.DeleteProfileLink(ctx, created.ID)
	require.NoError(t, err)

	links, err := store.ListProfileLinks(ctx)
	require.NoError(t, err)
	assert.Len(t, links, 0)
}

func TestDeleteProfileLink_NotFoundReturnsError(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteProfileLink(ctx, 999)
	require.Error(t, err)
}

func TestReorderProfileLinks(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	a, _ := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "A", URL: "https://a.com",
	})
	b, _ := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "B", URL: "https://b.com",
	})
	c, _ := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
		Label: "C", URL: "https://c.com",
	})

	// Reverse the order: C, B, A
	err := store.ReorderProfileLinks(ctx, []int64{c.ID, b.ID, a.ID})
	require.NoError(t, err)

	links, err := store.ListProfileLinks(ctx)
	require.NoError(t, err)
	require.Len(t, links, 3)
	assert.Equal(t, "C", links[0].Label)
	assert.Equal(t, "B", links[1].Label)
	assert.Equal(t, "A", links[2].Label)
	assert.Equal(t, 1, links[0].SortOrder)
	assert.Equal(t, 2, links[1].SortOrder)
	assert.Equal(t, 3, links[2].SortOrder)
}
