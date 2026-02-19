package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDescriptor(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	desc, err := store.CreateDescriptor(ctx, "Software Engineer")
	require.NoError(t, err)
	assert.NotZero(t, desc.ID)
	assert.Equal(t, "Software Engineer", desc.Title)
	assert.Equal(t, 1, desc.SortOrder)
	assert.NotEmpty(t, desc.CreatedAt)
}

func TestCreateDescriptor_AutoIncrementsSortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	first, _ := store.CreateDescriptor(ctx, "Engineer")
	second, _ := store.CreateDescriptor(ctx, "Architect")
	assert.Equal(t, 1, first.SortOrder)
	assert.Equal(t, 2, second.SortOrder)
}

func TestCreateDescriptor_DuplicateTitleFails(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateDescriptor(ctx, "Engineer")
	require.NoError(t, err)

	_, err = store.CreateDescriptor(ctx, "Engineer")
	require.Error(t, err, "duplicate title should fail UNIQUE constraint")
}

func TestListDescriptors_OrderedBySortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateDescriptor(ctx, "Engineer")
	_, _ = store.CreateDescriptor(ctx, "Architect")
	_, _ = store.CreateDescriptor(ctx, "Leader")

	descs, err := store.ListDescriptors(ctx)
	require.NoError(t, err)
	require.Len(t, descs, 3)
	assert.Equal(t, "Engineer", descs[0].Title)
	assert.Equal(t, "Architect", descs[1].Title)
	assert.Equal(t, "Leader", descs[2].Title)
}

func TestListDescriptors_EmptyReturnsEmptySlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	descs, err := store.ListDescriptors(ctx)
	require.NoError(t, err)
	assert.NotNil(t, descs)
	assert.Len(t, descs, 0)
}

func TestUpdateDescriptor(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	desc, _ := store.CreateDescriptor(ctx, "Engineer")

	updated, err := store.UpdateDescriptor(ctx, desc.ID, "Senior Engineer")
	require.NoError(t, err)
	assert.Equal(t, desc.ID, updated.ID)
	assert.Equal(t, "Senior Engineer", updated.Title)
	assert.Equal(t, desc.SortOrder, updated.SortOrder)
}

func TestUpdateDescriptor_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateDescriptor(ctx, 999, "New Title")
	require.Error(t, err)
}

func TestUpdateDescriptor_DuplicateTitleFails(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateDescriptor(ctx, "Engineer")
	b, _ := store.CreateDescriptor(ctx, "Architect")

	_, err := store.UpdateDescriptor(ctx, b.ID, "Engineer")
	require.Error(t, err, "updating to duplicate title should fail")
}

func TestDeleteDescriptor(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	desc, _ := store.CreateDescriptor(ctx, "Engineer")

	err := store.DeleteDescriptor(ctx, desc.ID)
	require.NoError(t, err)

	descs, err := store.ListDescriptors(ctx)
	require.NoError(t, err)
	assert.Len(t, descs, 0)
}

func TestDeleteDescriptor_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteDescriptor(ctx, 999)
	require.Error(t, err)
}

func TestReorderDescriptors(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	a, _ := store.CreateDescriptor(ctx, "A")
	b, _ := store.CreateDescriptor(ctx, "B")
	c, _ := store.CreateDescriptor(ctx, "C")

	err := store.ReorderDescriptors(ctx, []int64{c.ID, b.ID, a.ID})
	require.NoError(t, err)

	descs, err := store.ListDescriptors(ctx)
	require.NoError(t, err)
	require.Len(t, descs, 3)
	assert.Equal(t, "C", descs[0].Title)
	assert.Equal(t, "B", descs[1].Title)
	assert.Equal(t, "A", descs[2].Title)
	assert.Equal(t, 1, descs[0].SortOrder)
	assert.Equal(t, 2, descs[1].SortOrder)
	assert.Equal(t, 3, descs[2].SortOrder)
}
