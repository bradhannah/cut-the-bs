package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSummaryStore implements SummaryStore for testing.
type mockSummaryStore struct {
	summaries []domain.ProfessionalSummary
	summary   domain.ProfessionalSummary
	err       error

	// call tracking
	createCalls []domain.SummaryInput
	updateCalls []updateSummaryCall
	deleteCalls []int64
	getCalls    []int64
}

type updateSummaryCall struct {
	ID    int64
	Input domain.SummaryInput
}

func (m *mockSummaryStore) ListSummaries(_ context.Context) ([]domain.ProfessionalSummary, error) {
	return m.summaries, m.err
}

func (m *mockSummaryStore) GetSummary(_ context.Context, id int64) (domain.ProfessionalSummary, error) {
	m.getCalls = append(m.getCalls, id)
	if m.err != nil {
		return domain.ProfessionalSummary{}, m.err
	}
	return m.summary, nil
}

func (m *mockSummaryStore) CreateSummary(_ context.Context, input domain.SummaryInput) (domain.ProfessionalSummary, error) {
	m.createCalls = append(m.createCalls, input)
	if m.err != nil {
		return domain.ProfessionalSummary{}, m.err
	}
	return m.summary, nil
}

func (m *mockSummaryStore) UpdateSummary(_ context.Context, id int64, input domain.SummaryInput) (domain.ProfessionalSummary, error) {
	m.updateCalls = append(m.updateCalls, updateSummaryCall{ID: id, Input: input})
	if m.err != nil {
		return domain.ProfessionalSummary{}, m.err
	}
	return m.summary, nil
}

func (m *mockSummaryStore) DeleteSummary(_ context.Context, id int64) error {
	m.deleteCalls = append(m.deleteCalls, id)
	return m.err
}

// =================================================================
// ListSummaries
// =================================================================

func TestSummaryService_ListSummaries_Success(t *testing.T) {
	store := &mockSummaryStore{
		summaries: []domain.ProfessionalSummary{
			{ID: 1, Label: "General", BodyText: "Experienced developer..."},
			{ID: 2, Label: "Frontend", BodyText: "Frontend specialist..."},
		},
	}
	svc := NewSummaryService(store)

	summaries, err := svc.ListSummaries(context.Background())
	require.NoError(t, err)
	require.Len(t, summaries, 2)
	assert.Equal(t, "General", summaries[0].Label)
}

func TestSummaryService_ListSummaries_Empty(t *testing.T) {
	store := &mockSummaryStore{summaries: []domain.ProfessionalSummary{}}
	svc := NewSummaryService(store)

	summaries, err := svc.ListSummaries(context.Background())
	require.NoError(t, err)
	assert.Empty(t, summaries)
}

func TestSummaryService_ListSummaries_StoreError(t *testing.T) {
	store := &mockSummaryStore{err: fmt.Errorf("db error")}
	svc := NewSummaryService(store)

	_, err := svc.ListSummaries(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// =================================================================
// GetSummary
// =================================================================

func TestSummaryService_GetSummary_Success(t *testing.T) {
	store := &mockSummaryStore{
		summary: domain.ProfessionalSummary{
			ID: 1, Label: "General", BodyText: "Experienced developer...",
		},
	}
	svc := NewSummaryService(store)

	s, err := svc.GetSummary(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "General", s.Label)
	require.Len(t, store.getCalls, 1)
	assert.Equal(t, int64(1), store.getCalls[0])
}

func TestSummaryService_GetSummary_StoreError(t *testing.T) {
	store := &mockSummaryStore{err: fmt.Errorf("not found")}
	svc := NewSummaryService(store)

	_, err := svc.GetSummary(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// CreateSummary — validation
// =================================================================

func TestSummaryService_CreateSummary_EmptyLabel(t *testing.T) {
	store := &mockSummaryStore{}
	svc := NewSummaryService(store)

	_, err := svc.CreateSummary(context.Background(), domain.SummaryInput{
		Label:    "",
		BodyText: "Some text",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
	assert.Empty(t, store.createCalls)
}

func TestSummaryService_CreateSummary_WhitespaceLabel(t *testing.T) {
	store := &mockSummaryStore{}
	svc := NewSummaryService(store)

	_, err := svc.CreateSummary(context.Background(), domain.SummaryInput{
		Label:    "   ",
		BodyText: "Some text",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
	assert.Empty(t, store.createCalls)
}

func TestSummaryService_CreateSummary_EmptyBodyText(t *testing.T) {
	store := &mockSummaryStore{}
	svc := NewSummaryService(store)

	_, err := svc.CreateSummary(context.Background(), domain.SummaryInput{
		Label:    "General",
		BodyText: "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "body text")
	assert.Empty(t, store.createCalls)
}

func TestSummaryService_CreateSummary_WhitespaceBodyText(t *testing.T) {
	store := &mockSummaryStore{}
	svc := NewSummaryService(store)

	_, err := svc.CreateSummary(context.Background(), domain.SummaryInput{
		Label:    "General",
		BodyText: "   ",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "body text")
	assert.Empty(t, store.createCalls)
}

// =================================================================
// CreateSummary — happy path
// =================================================================

func TestSummaryService_CreateSummary_Success(t *testing.T) {
	store := &mockSummaryStore{
		summary: domain.ProfessionalSummary{
			ID: 1, Label: "General", BodyText: "Experienced developer...",
		},
	}
	svc := NewSummaryService(store)

	s, err := svc.CreateSummary(context.Background(), domain.SummaryInput{
		Label:    "General",
		BodyText: "Experienced developer...",
	})
	require.NoError(t, err)
	assert.Equal(t, "General", s.Label)
	assert.Equal(t, "Experienced developer...", s.BodyText)
	require.Len(t, store.createCalls, 1)
}

func TestSummaryService_CreateSummary_StoreError(t *testing.T) {
	store := &mockSummaryStore{err: fmt.Errorf("duplicate label")}
	svc := NewSummaryService(store)

	_, err := svc.CreateSummary(context.Background(), domain.SummaryInput{
		Label:    "General",
		BodyText: "Some text",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate label")
}

// =================================================================
// UpdateSummary — validation
// =================================================================

func TestSummaryService_UpdateSummary_EmptyLabel(t *testing.T) {
	store := &mockSummaryStore{}
	svc := NewSummaryService(store)

	_, err := svc.UpdateSummary(context.Background(), 1, domain.SummaryInput{
		Label:    "",
		BodyText: "Some text",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
	assert.Empty(t, store.updateCalls)
}

func TestSummaryService_UpdateSummary_EmptyBodyText(t *testing.T) {
	store := &mockSummaryStore{}
	svc := NewSummaryService(store)

	_, err := svc.UpdateSummary(context.Background(), 1, domain.SummaryInput{
		Label:    "General",
		BodyText: "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "body text")
	assert.Empty(t, store.updateCalls)
}

// =================================================================
// UpdateSummary — happy path
// =================================================================

func TestSummaryService_UpdateSummary_Success(t *testing.T) {
	store := &mockSummaryStore{
		summary: domain.ProfessionalSummary{
			ID: 1, Label: "Updated", BodyText: "New text",
		},
	}
	svc := NewSummaryService(store)

	s, err := svc.UpdateSummary(context.Background(), 1, domain.SummaryInput{
		Label:    "Updated",
		BodyText: "New text",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", s.Label)
	require.Len(t, store.updateCalls, 1)
	assert.Equal(t, int64(1), store.updateCalls[0].ID)
}

func TestSummaryService_UpdateSummary_StoreError(t *testing.T) {
	store := &mockSummaryStore{err: fmt.Errorf("not found")}
	svc := NewSummaryService(store)

	_, err := svc.UpdateSummary(context.Background(), 999, domain.SummaryInput{
		Label:    "Updated",
		BodyText: "New text",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// DeleteSummary
// =================================================================

func TestSummaryService_DeleteSummary_Success(t *testing.T) {
	store := &mockSummaryStore{}
	svc := NewSummaryService(store)

	err := svc.DeleteSummary(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, store.deleteCalls, 1)
	assert.Equal(t, int64(5), store.deleteCalls[0])
}

func TestSummaryService_DeleteSummary_StoreError(t *testing.T) {
	store := &mockSummaryStore{err: fmt.Errorf("not found")}
	svc := NewSummaryService(store)

	err := svc.DeleteSummary(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
