package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCoverLetter(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cl, err := store.CreateCoverLetter(ctx, domain.CoverLetterInput{
		Title:    "Software Engineer at Acme",
		BodyText: "Dear Hiring Manager,\n\nI am writing to express...",
	})
	require.NoError(t, err)
	assert.NotZero(t, cl.ID)
	assert.Equal(t, "Software Engineer at Acme", cl.Title)
	assert.Equal(t, "Dear Hiring Manager,\n\nI am writing to express...", cl.BodyText)
	assert.Empty(t, cl.FilePath, "file_path should be empty initially")
	assert.NotEmpty(t, cl.CreatedAt)
	assert.NotEmpty(t, cl.UpdatedAt)
}

func TestListCoverLetters(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateCoverLetter(ctx, domain.CoverLetterInput{
		Title: "Letter A", BodyText: "Body A",
	})
	_, _ = store.CreateCoverLetter(ctx, domain.CoverLetterInput{
		Title: "Letter B", BodyText: "Body B",
	})

	letters, err := store.ListCoverLetters(ctx)
	require.NoError(t, err)
	require.Len(t, letters, 2)
	assert.Equal(t, "Letter A", letters[0].Title)
	assert.Equal(t, "Letter B", letters[1].Title)
}

func TestListCoverLetters_EmptyReturnsEmptySlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	letters, err := store.ListCoverLetters(ctx)
	require.NoError(t, err)
	assert.NotNil(t, letters)
	assert.Len(t, letters, 0)
}

func TestGetCoverLetter(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, _ := store.CreateCoverLetter(ctx, domain.CoverLetterInput{
		Title: "My Letter", BodyText: "Body text here",
	})

	fetched, err := store.GetCoverLetter(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "My Letter", fetched.Title)
	assert.Equal(t, "Body text here", fetched.BodyText)
}

func TestGetCoverLetter_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.GetCoverLetter(ctx, 999)
	require.Error(t, err)
}

func TestUpdateCoverLetter(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, _ := store.CreateCoverLetter(ctx, domain.CoverLetterInput{
		Title: "Original", BodyText: "Old body",
	})

	updated, err := store.UpdateCoverLetter(ctx, created.ID, domain.CoverLetterInput{
		Title:    "Updated Title",
		BodyText: "New body text",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, "New body text", updated.BodyText)
}

func TestUpdateCoverLetter_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateCoverLetter(ctx, 999, domain.CoverLetterInput{
		Title: "X", BodyText: "Y",
	})
	require.Error(t, err)
}

func TestUpdateCoverLetter_FilePath(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, _ := store.CreateCoverLetter(ctx, domain.CoverLetterInput{
		Title: "Letter", BodyText: "Body",
	})

	// Set file path after PDF export.
	err := store.UpdateCoverLetterFilePath(ctx, created.ID, "/exports/cover-letter.pdf")
	require.NoError(t, err)

	fetched, err := store.GetCoverLetter(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "/exports/cover-letter.pdf", fetched.FilePath)
}

func TestDeleteCoverLetter(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, _ := store.CreateCoverLetter(ctx, domain.CoverLetterInput{
		Title: "To Delete", BodyText: "Body",
	})

	err := store.DeleteCoverLetter(ctx, created.ID)
	require.NoError(t, err)

	letters, err := store.ListCoverLetters(ctx)
	require.NoError(t, err)
	assert.Len(t, letters, 0)
}

func TestDeleteCoverLetter_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteCoverLetter(ctx, 999)
	require.Error(t, err)
}
