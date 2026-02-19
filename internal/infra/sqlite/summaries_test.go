package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "Technical Leader",
		BodyText: "Experienced technical leader with 10+ years...",
	})
	require.NoError(t, err)
	assert.NotZero(t, summary.ID)
	assert.Equal(t, "Technical Leader", summary.Label)
	assert.Equal(t, "Experienced technical leader with 10+ years...", summary.BodyText)
	assert.NotEmpty(t, summary.CreatedAt)
}

func TestCreateSummary_DuplicateLabelFails(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "Leader",
		BodyText: "First version",
	})
	require.NoError(t, err)

	_, err = store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "Leader",
		BodyText: "Second version",
	})
	require.Error(t, err, "duplicate label should fail UNIQUE constraint")
}

func TestListSummaries(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateSummary(ctx, domain.SummaryInput{
		Label: "A", BodyText: "Text A",
	})
	_, _ = store.CreateSummary(ctx, domain.SummaryInput{
		Label: "B", BodyText: "Text B",
	})

	summaries, err := store.ListSummaries(ctx)
	require.NoError(t, err)
	require.Len(t, summaries, 2)
	assert.Equal(t, "A", summaries[0].Label)
	assert.Equal(t, "B", summaries[1].Label)
}

func TestListSummaries_EmptyReturnsEmptySlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	summaries, err := store.ListSummaries(ctx)
	require.NoError(t, err)
	assert.NotNil(t, summaries)
	assert.Len(t, summaries, 0)
}

func TestGetSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, _ := store.CreateSummary(ctx, domain.SummaryInput{
		Label: "Leader", BodyText: "Leading teams...",
	})

	fetched, err := store.GetSummary(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "Leader", fetched.Label)
	assert.Equal(t, "Leading teams...", fetched.BodyText)
}

func TestGetSummary_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.GetSummary(ctx, 999)
	require.Error(t, err)
}

func TestUpdateSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, _ := store.CreateSummary(ctx, domain.SummaryInput{
		Label: "Leader", BodyText: "Old text",
	})

	updated, err := store.UpdateSummary(ctx, created.ID, domain.SummaryInput{
		Label:    "Technical Leader",
		BodyText: "New text with more detail",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Technical Leader", updated.Label)
	assert.Equal(t, "New text with more detail", updated.BodyText)
}

func TestUpdateSummary_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateSummary(ctx, 999, domain.SummaryInput{
		Label: "X", BodyText: "Y",
	})
	require.Error(t, err)
}

func TestUpdateSummary_DuplicateLabelFails(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateSummary(ctx, domain.SummaryInput{
		Label: "Leader", BodyText: "Text A",
	})
	b, _ := store.CreateSummary(ctx, domain.SummaryInput{
		Label: "Builder", BodyText: "Text B",
	})

	_, err := store.UpdateSummary(ctx, b.ID, domain.SummaryInput{
		Label:    "Leader",
		BodyText: "Trying to steal label",
	})
	require.Error(t, err, "updating to duplicate label should fail")
}

func TestDeleteSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, _ := store.CreateSummary(ctx, domain.SummaryInput{
		Label: "Leader", BodyText: "Text",
	})

	err := store.DeleteSummary(ctx, created.ID)
	require.NoError(t, err)

	summaries, err := store.ListSummaries(ctx)
	require.NoError(t, err)
	assert.Len(t, summaries, 0)
}

func TestDeleteSummary_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteSummary(ctx, 999)
	require.Error(t, err)
}
