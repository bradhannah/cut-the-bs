package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCoverLetterStore implements CoverLetterStore for testing.
type mockCoverLetterStore struct {
	letters []domain.CoverLetter
	letter  domain.CoverLetter
	profile domain.UserProfile
	links   []domain.ProfileLink
	err     error

	// call tracking
	createCalls   []domain.CoverLetterInput
	updateCalls   []updateCoverLetterCall
	deleteCalls   []int64
	getCalls      []int64
	fpUpdateCalls []filePathUpdateCall
}

type updateCoverLetterCall struct {
	ID    int64
	Input domain.CoverLetterInput
}

type filePathUpdateCall struct {
	ID       int64
	FilePath string
}

func (m *mockCoverLetterStore) ListCoverLetters(_ context.Context) ([]domain.CoverLetter, error) {
	return m.letters, m.err
}

func (m *mockCoverLetterStore) GetCoverLetter(_ context.Context, id int64) (domain.CoverLetter, error) {
	m.getCalls = append(m.getCalls, id)
	if m.err != nil {
		return domain.CoverLetter{}, m.err
	}
	return m.letter, nil
}

func (m *mockCoverLetterStore) CreateCoverLetter(_ context.Context, input domain.CoverLetterInput) (domain.CoverLetter, error) {
	m.createCalls = append(m.createCalls, input)
	if m.err != nil {
		return domain.CoverLetter{}, m.err
	}
	return m.letter, nil
}

func (m *mockCoverLetterStore) UpdateCoverLetter(_ context.Context, id int64, input domain.CoverLetterInput) (domain.CoverLetter, error) {
	m.updateCalls = append(m.updateCalls, updateCoverLetterCall{ID: id, Input: input})
	if m.err != nil {
		return domain.CoverLetter{}, m.err
	}
	return m.letter, nil
}

func (m *mockCoverLetterStore) DeleteCoverLetter(_ context.Context, id int64) error {
	m.deleteCalls = append(m.deleteCalls, id)
	return m.err
}

func (m *mockCoverLetterStore) UpdateCoverLetterFilePath(_ context.Context, id int64, fp string) error {
	m.fpUpdateCalls = append(m.fpUpdateCalls, filePathUpdateCall{ID: id, FilePath: fp})
	return m.err
}

func (m *mockCoverLetterStore) GetProfile(_ context.Context) (domain.UserProfile, error) {
	return m.profile, m.err
}

func (m *mockCoverLetterStore) ListProfileLinks(_ context.Context) ([]domain.ProfileLink, error) {
	return m.links, m.err
}

// mockCoverLetterRenderer implements CoverLetterRenderer for testing.
type mockCoverLetterRenderer struct {
	filePath string
	err      error
	calls    []domain.RenderCoverLetterRequest
}

func (m *mockCoverLetterRenderer) RenderCoverLetter(_ context.Context, req domain.RenderCoverLetterRequest) (string, error) {
	m.calls = append(m.calls, req)
	return m.filePath, m.err
}

// =================================================================
// ListCoverLetters
// =================================================================

func TestCoverLetterService_ListCoverLetters_Success(t *testing.T) {
	store := &mockCoverLetterStore{
		letters: []domain.CoverLetter{
			{ID: 1, Title: "Acme Corp", BodyText: "Dear hiring manager..."},
			{ID: 2, Title: "TechCo", BodyText: "I am writing to apply..."},
		},
	}
	svc := NewCoverLetterService(store, nil, "")

	letters, err := svc.ListCoverLetters(context.Background())
	require.NoError(t, err)
	require.Len(t, letters, 2)
	assert.Equal(t, "Acme Corp", letters[0].Title)
}

func TestCoverLetterService_ListCoverLetters_Empty(t *testing.T) {
	store := &mockCoverLetterStore{letters: []domain.CoverLetter{}}
	svc := NewCoverLetterService(store, nil, "")

	letters, err := svc.ListCoverLetters(context.Background())
	require.NoError(t, err)
	assert.Empty(t, letters)
}

func TestCoverLetterService_ListCoverLetters_StoreError(t *testing.T) {
	store := &mockCoverLetterStore{err: fmt.Errorf("db error")}
	svc := NewCoverLetterService(store, nil, "")

	_, err := svc.ListCoverLetters(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// =================================================================
// CreateCoverLetter — validation
// =================================================================

func TestCoverLetterService_CreateCoverLetter_EmptyTitle(t *testing.T) {
	store := &mockCoverLetterStore{}
	svc := NewCoverLetterService(store, nil, "")

	_, err := svc.CreateCoverLetter(context.Background(), domain.CoverLetterInput{
		Title:    "",
		BodyText: "Some text",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title")
	assert.Empty(t, store.createCalls)
}

func TestCoverLetterService_CreateCoverLetter_WhitespaceTitle(t *testing.T) {
	store := &mockCoverLetterStore{}
	svc := NewCoverLetterService(store, nil, "")

	_, err := svc.CreateCoverLetter(context.Background(), domain.CoverLetterInput{
		Title:    "   ",
		BodyText: "Some text",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title")
	assert.Empty(t, store.createCalls)
}

func TestCoverLetterService_CreateCoverLetter_EmptyBodyText(t *testing.T) {
	store := &mockCoverLetterStore{}
	svc := NewCoverLetterService(store, nil, "")

	_, err := svc.CreateCoverLetter(context.Background(), domain.CoverLetterInput{
		Title:    "My Cover Letter",
		BodyText: "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "body text")
	assert.Empty(t, store.createCalls)
}

// =================================================================
// CreateCoverLetter — happy path
// =================================================================

func TestCoverLetterService_CreateCoverLetter_Success(t *testing.T) {
	store := &mockCoverLetterStore{
		letter: domain.CoverLetter{
			ID: 1, Title: "Acme Corp", BodyText: "Dear hiring manager...",
		},
	}
	svc := NewCoverLetterService(store, nil, "")

	cl, err := svc.CreateCoverLetter(context.Background(), domain.CoverLetterInput{
		Title:    "Acme Corp",
		BodyText: "Dear hiring manager...",
	})
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", cl.Title)
	require.Len(t, store.createCalls, 1)
}

func TestCoverLetterService_CreateCoverLetter_StoreError(t *testing.T) {
	store := &mockCoverLetterStore{err: fmt.Errorf("insert error")}
	svc := NewCoverLetterService(store, nil, "")

	_, err := svc.CreateCoverLetter(context.Background(), domain.CoverLetterInput{
		Title:    "Acme Corp",
		BodyText: "Dear hiring manager...",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert error")
}

// =================================================================
// UpdateCoverLetter
// =================================================================

func TestCoverLetterService_UpdateCoverLetter_EmptyTitle(t *testing.T) {
	store := &mockCoverLetterStore{}
	svc := NewCoverLetterService(store, nil, "")

	_, err := svc.UpdateCoverLetter(context.Background(), 1, domain.CoverLetterInput{
		Title:    "",
		BodyText: "Some text",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title")
	assert.Empty(t, store.updateCalls)
}

func TestCoverLetterService_UpdateCoverLetter_Success(t *testing.T) {
	store := &mockCoverLetterStore{
		letter: domain.CoverLetter{
			ID: 1, Title: "Updated", BodyText: "New text",
		},
	}
	svc := NewCoverLetterService(store, nil, "")

	cl, err := svc.UpdateCoverLetter(context.Background(), 1, domain.CoverLetterInput{
		Title:    "Updated",
		BodyText: "New text",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", cl.Title)
	require.Len(t, store.updateCalls, 1)
	assert.Equal(t, int64(1), store.updateCalls[0].ID)
}

func TestCoverLetterService_UpdateCoverLetter_StoreError(t *testing.T) {
	store := &mockCoverLetterStore{err: fmt.Errorf("not found")}
	svc := NewCoverLetterService(store, nil, "")

	_, err := svc.UpdateCoverLetter(context.Background(), 999, domain.CoverLetterInput{
		Title:    "Updated",
		BodyText: "New text",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// DeleteCoverLetter
// =================================================================

func TestCoverLetterService_DeleteCoverLetter_Success(t *testing.T) {
	store := &mockCoverLetterStore{}
	svc := NewCoverLetterService(store, nil, "")

	err := svc.DeleteCoverLetter(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, store.deleteCalls, 1)
	assert.Equal(t, int64(5), store.deleteCalls[0])
}

func TestCoverLetterService_DeleteCoverLetter_StoreError(t *testing.T) {
	store := &mockCoverLetterStore{err: fmt.Errorf("not found")}
	svc := NewCoverLetterService(store, nil, "")

	err := svc.DeleteCoverLetter(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =================================================================
// ExportCoverLetter
// =================================================================

func TestCoverLetterService_ExportCoverLetter_Success(t *testing.T) {
	store := &mockCoverLetterStore{
		letter: domain.CoverLetter{
			ID: 1, Title: "Acme Corp", BodyText: "Dear hiring manager...",
		},
		profile: domain.UserProfile{ID: 1, FullName: "Jane", Email: "jane@example.com"},
		links:   []domain.ProfileLink{{ID: 1, Label: "LinkedIn", URL: "https://linkedin.com"}},
	}
	renderer := &mockCoverLetterRenderer{filePath: "/tmp/cover-letter-123.pdf"}
	svc := NewCoverLetterService(store, renderer, "/tmp")

	cl, err := svc.ExportCoverLetter(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", cl.Title)

	// Renderer should have been called.
	require.Len(t, renderer.calls, 1)
	assert.Equal(t, "/tmp", renderer.calls[0].OutputDir)
	assert.Equal(t, "Jane", renderer.calls[0].Profile.FullName)

	// File path should have been stored.
	require.Len(t, store.fpUpdateCalls, 1)
	assert.Equal(t, int64(1), store.fpUpdateCalls[0].ID)
	assert.Equal(t, "/tmp/cover-letter-123.pdf", store.fpUpdateCalls[0].FilePath)
}

func TestCoverLetterService_ExportCoverLetter_NotFound(t *testing.T) {
	store := &mockCoverLetterStore{err: fmt.Errorf("not found")}
	svc := NewCoverLetterService(store, nil, "/tmp")

	_, err := svc.ExportCoverLetter(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCoverLetterService_ExportCoverLetter_RenderError(t *testing.T) {
	store := &mockCoverLetterStore{
		letter:  domain.CoverLetter{ID: 1, Title: "Acme", BodyText: "text"},
		profile: domain.UserProfile{ID: 1, FullName: "Jane", Email: "j@e.com"},
	}
	renderer := &mockCoverLetterRenderer{err: fmt.Errorf("render failed")}
	svc := NewCoverLetterService(store, renderer, "/tmp")

	_, err := svc.ExportCoverLetter(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "render")
}
