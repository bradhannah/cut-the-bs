package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDescriptorStore implements DescriptorStore for testing.
type mockDescriptorStore struct {
	descriptors []domain.RoleDescriptor
	descriptor  domain.RoleDescriptor
	err         error

	// call tracking
	createCalls  []string
	updateCalls  []updateDescriptorCall
	deleteCalls  []int64
	reorderCalls [][]int64
}

type updateDescriptorCall struct {
	ID    int64
	Title string
}

func (m *mockDescriptorStore) ListDescriptors(_ context.Context) ([]domain.RoleDescriptor, error) {
	return m.descriptors, m.err
}

func (m *mockDescriptorStore) CreateDescriptor(_ context.Context, title string) (domain.RoleDescriptor, error) {
	m.createCalls = append(m.createCalls, title)
	if m.err != nil {
		return domain.RoleDescriptor{}, m.err
	}
	return m.descriptor, nil
}

func (m *mockDescriptorStore) UpdateDescriptor(_ context.Context, id int64, title string) (domain.RoleDescriptor, error) {
	m.updateCalls = append(m.updateCalls, updateDescriptorCall{ID: id, Title: title})
	if m.err != nil {
		return domain.RoleDescriptor{}, m.err
	}
	return m.descriptor, nil
}

func (m *mockDescriptorStore) DeleteDescriptor(_ context.Context, id int64) error {
	m.deleteCalls = append(m.deleteCalls, id)
	return m.err
}

func (m *mockDescriptorStore) ReorderDescriptors(_ context.Context, orderedIDs []int64) error {
	m.reorderCalls = append(m.reorderCalls, orderedIDs)
	return m.err
}

// =================================================================
// ListDescriptors
// =================================================================

func TestDescriptorService_ListDescriptors_Success(t *testing.T) {
	store := &mockDescriptorStore{
		descriptors: []domain.RoleDescriptor{
			{ID: 1, Title: "Senior Software Engineer", SortOrder: 1},
			{ID: 2, Title: "Tech Lead", SortOrder: 2},
		},
	}
	svc := NewDescriptorService(store)

	descs, err := svc.ListDescriptors(context.Background())
	require.NoError(t, err)
	require.Len(t, descs, 2)
	assert.Equal(t, "Senior Software Engineer", descs[0].Title)
}

func TestDescriptorService_ListDescriptors_Empty(t *testing.T) {
	store := &mockDescriptorStore{descriptors: []domain.RoleDescriptor{}}
	svc := NewDescriptorService(store)

	descs, err := svc.ListDescriptors(context.Background())
	require.NoError(t, err)
	assert.Empty(t, descs)
}

func TestDescriptorService_ListDescriptors_StoreError(t *testing.T) {
	store := &mockDescriptorStore{err: fmt.Errorf("db error")}
	svc := NewDescriptorService(store)

	_, err := svc.ListDescriptors(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// =================================================================
// CreateDescriptor — validation
// =================================================================

func TestDescriptorService_CreateDescriptor_EmptyTitle(t *testing.T) {
	store := &mockDescriptorStore{}
	svc := NewDescriptorService(store)

	_, err := svc.CreateDescriptor(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title")
	assert.Empty(t, store.createCalls)
}

func TestDescriptorService_CreateDescriptor_WhitespaceTitle(t *testing.T) {
	store := &mockDescriptorStore{}
	svc := NewDescriptorService(store)

	_, err := svc.CreateDescriptor(context.Background(), "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title")
	assert.Empty(t, store.createCalls)
}

// =================================================================
// CreateDescriptor — happy path
// =================================================================

func TestDescriptorService_CreateDescriptor_Success(t *testing.T) {
	store := &mockDescriptorStore{
		descriptor: domain.RoleDescriptor{
			ID: 1, Title: "Senior Software Engineer", SortOrder: 1,
		},
	}
	svc := NewDescriptorService(store)

	desc, err := svc.CreateDescriptor(context.Background(), "Senior Software Engineer")
	require.NoError(t, err)
	assert.Equal(t, "Senior Software Engineer", desc.Title)
	require.Len(t, store.createCalls, 1)
	assert.Equal(t, "Senior Software Engineer", store.createCalls[0])
}

func TestDescriptorService_CreateDescriptor_StoreError(t *testing.T) {
	store := &mockDescriptorStore{err: fmt.Errorf("duplicate title")}
	svc := NewDescriptorService(store)

	_, err := svc.CreateDescriptor(context.Background(), "Senior Software Engineer")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate title")
}

// =================================================================
// UpdateDescriptor — validation
// =================================================================

func TestDescriptorService_UpdateDescriptor_EmptyTitle(t *testing.T) {
	store := &mockDescriptorStore{}
	svc := NewDescriptorService(store)

	_, err := svc.UpdateDescriptor(context.Background(), 1, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title")
	assert.Empty(t, store.updateCalls)
}

func TestDescriptorService_UpdateDescriptor_WhitespaceTitle(t *testing.T) {
	store := &mockDescriptorStore{}
	svc := NewDescriptorService(store)

	_, err := svc.UpdateDescriptor(context.Background(), 1, "  \t  ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title")
	assert.Empty(t, store.updateCalls)
}

// =================================================================
// UpdateDescriptor — happy path
// =================================================================

func TestDescriptorService_UpdateDescriptor_Success(t *testing.T) {
	store := &mockDescriptorStore{
		descriptor: domain.RoleDescriptor{
			ID: 1, Title: "Staff Engineer", SortOrder: 1,
		},
	}
	svc := NewDescriptorService(store)

	desc, err := svc.UpdateDescriptor(context.Background(), 1, "Staff Engineer")
	require.NoError(t, err)
	assert.Equal(t, "Staff Engineer", desc.Title)
	require.Len(t, store.updateCalls, 1)
	assert.Equal(t, int64(1), store.updateCalls[0].ID)
	assert.Equal(t, "Staff Engineer", store.updateCalls[0].Title)
}

func TestDescriptorService_UpdateDescriptor_StoreError(t *testing.T) {
	store := &mockDescriptorStore{err: fmt.Errorf("not found")}
	svc := NewDescriptorService(store)

	_, err := svc.UpdateDescriptor(context.Background(), 999, "Staff Engineer")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// DeleteDescriptor
// =================================================================

func TestDescriptorService_DeleteDescriptor_Success(t *testing.T) {
	store := &mockDescriptorStore{}
	svc := NewDescriptorService(store)

	err := svc.DeleteDescriptor(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, store.deleteCalls, 1)
	assert.Equal(t, int64(5), store.deleteCalls[0])
}

func TestDescriptorService_DeleteDescriptor_StoreError(t *testing.T) {
	store := &mockDescriptorStore{err: fmt.Errorf("not found")}
	svc := NewDescriptorService(store)

	err := svc.DeleteDescriptor(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// ReorderDescriptors
// =================================================================

func TestDescriptorService_ReorderDescriptors_Success(t *testing.T) {
	store := &mockDescriptorStore{}
	svc := NewDescriptorService(store)

	err := svc.ReorderDescriptors(context.Background(), []int64{3, 1, 2})
	require.NoError(t, err)
	require.Len(t, store.reorderCalls, 1)
	assert.Equal(t, []int64{3, 1, 2}, store.reorderCalls[0])
}

func TestDescriptorService_ReorderDescriptors_Empty(t *testing.T) {
	store := &mockDescriptorStore{}
	svc := NewDescriptorService(store)

	err := svc.ReorderDescriptors(context.Background(), []int64{})
	require.NoError(t, err)
}
