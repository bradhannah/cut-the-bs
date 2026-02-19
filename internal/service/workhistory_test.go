package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStore implements WorkHistoryStore for testing. All methods
// return preconfigured values and track calls for assertions.
type mockStore struct {
	// --- configurable returns ---
	workHistories  []domain.WorkHistoryEntry
	workHistory    domain.WorkHistoryEntry
	bullet         domain.AchievementBullet
	err            error
	deleteErr      error
	reorderErr     error
	createBulletFn func(ctx context.Context, whID int64, text string) (domain.AchievementBullet, error)

	// --- call tracking ---
	createCalls       []domain.WorkHistoryInput
	updateCalls       []updateCall
	deleteCalls       []int64
	reorderCalls      [][]int64
	createBulletCalls []createBulletCall
	updateBulletCalls []updateBulletCall
	deleteBulletCalls []int64
	reorderBulletCall *reorderBulletCall
}

type updateCall struct {
	ID    int64
	Input domain.WorkHistoryInput
}

type createBulletCall struct {
	WorkHistoryID int64
	Text          string
}

type updateBulletCall struct {
	ID   int64
	Text string
}

type reorderBulletCall struct {
	WorkHistoryID int64
	OrderedIDs    []int64
}

func (m *mockStore) ListWorkHistory(_ context.Context) ([]domain.WorkHistoryEntry, error) {
	return m.workHistories, m.err
}

func (m *mockStore) GetWorkHistory(_ context.Context, id int64) (domain.WorkHistoryEntry, error) {
	if m.err != nil {
		return domain.WorkHistoryEntry{}, m.err
	}
	return m.workHistory, nil
}

func (m *mockStore) CreateWorkHistory(_ context.Context, input domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	m.createCalls = append(m.createCalls, input)
	if m.err != nil {
		return domain.WorkHistoryEntry{}, m.err
	}
	return m.workHistory, nil
}

func (m *mockStore) UpdateWorkHistory(_ context.Context, id int64, input domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	m.updateCalls = append(m.updateCalls, updateCall{ID: id, Input: input})
	if m.err != nil {
		return domain.WorkHistoryEntry{}, m.err
	}
	return m.workHistory, nil
}

func (m *mockStore) DeleteWorkHistory(_ context.Context, id int64) error {
	m.deleteCalls = append(m.deleteCalls, id)
	return m.deleteErr
}

func (m *mockStore) ReorderWorkHistory(_ context.Context, orderedIDs []int64) error {
	m.reorderCalls = append(m.reorderCalls, orderedIDs)
	return m.reorderErr
}

func (m *mockStore) CreateBullet(ctx context.Context, workHistoryID int64, text string) (domain.AchievementBullet, error) {
	m.createBulletCalls = append(m.createBulletCalls, createBulletCall{WorkHistoryID: workHistoryID, Text: text})
	if m.createBulletFn != nil {
		return m.createBulletFn(ctx, workHistoryID, text)
	}
	if m.err != nil {
		return domain.AchievementBullet{}, m.err
	}
	return m.bullet, nil
}

func (m *mockStore) UpdateBullet(_ context.Context, id int64, text string) (domain.AchievementBullet, error) {
	m.updateBulletCalls = append(m.updateBulletCalls, updateBulletCall{ID: id, Text: text})
	if m.err != nil {
		return domain.AchievementBullet{}, m.err
	}
	return m.bullet, nil
}

func (m *mockStore) DeleteBullet(_ context.Context, id int64) error {
	m.deleteBulletCalls = append(m.deleteBulletCalls, id)
	return m.deleteErr
}

func (m *mockStore) ReorderBullets(_ context.Context, workHistoryID int64, orderedIDs []int64) error {
	m.reorderBulletCall = &reorderBulletCall{WorkHistoryID: workHistoryID, OrderedIDs: orderedIDs}
	return m.reorderErr
}

// --- Helper ---

func validInput() domain.WorkHistoryInput {
	return domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Software Engineer",
		StartDate:            "2020-01",
		EndDate:              "2023-06",
		DateGranularityStart: "month",
		DateGranularityEnd:   "month",
	}
}

// =================================================================
// Validation error tests
// =================================================================

func TestCreateWorkHistory_EmptyEmployerName(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	input := validInput()
	input.EmployerName = ""

	_, err := svc.CreateWorkHistory(context.Background(), input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "employer name")
	// Must NOT delegate to store when validation fails.
	assert.Empty(t, store.createCalls)
}

func TestCreateWorkHistory_EmptyJobTitle(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	input := validInput()
	input.JobTitle = ""

	_, err := svc.CreateWorkHistory(context.Background(), input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job title")
	assert.Empty(t, store.createCalls)
}

func TestCreateWorkHistory_EmptyStartDate(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	input := validInput()
	input.StartDate = ""

	_, err := svc.CreateWorkHistory(context.Background(), input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start date")
	assert.Empty(t, store.createCalls)
}

func TestCreateWorkHistory_InvalidGranularity(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	input := validInput()
	input.DateGranularityStart = "century"

	_, err := svc.CreateWorkHistory(context.Background(), input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "granularity")
	assert.Empty(t, store.createCalls)
}

func TestCreateWorkHistory_EndBeforeStart(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	input := domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Dev",
		StartDate:            "2023-06",
		EndDate:              "2020-01",
		DateGranularityStart: "month",
		DateGranularityEnd:   "month",
	}

	_, err := svc.CreateWorkHistory(context.Background(), input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "end date")
	assert.Empty(t, store.createCalls)
}

func TestUpdateWorkHistory_EndBeforeStart(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	input := domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Dev",
		StartDate:            "2023",
		EndDate:              "2020",
		DateGranularityStart: "year",
		DateGranularityEnd:   "year",
	}

	_, err := svc.UpdateWorkHistory(context.Background(), 1, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "end date")
	assert.Empty(t, store.updateCalls)
}

func TestUpdateWorkHistory_EmptyEmployerName(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	input := validInput()
	input.EmployerName = "   " // whitespace-only

	_, err := svc.UpdateWorkHistory(context.Background(), 1, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "employer name")
	assert.Empty(t, store.updateCalls)
}

func TestCreateWorkHistory_EndDateWithoutGranularity(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	input := validInput()
	input.EndDate = "2023-06"
	input.DateGranularityEnd = ""

	_, err := svc.CreateWorkHistory(context.Background(), input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "granularity")
	assert.Empty(t, store.createCalls)
}

// =================================================================
// Happy path tests — Create
// =================================================================

func TestCreateWorkHistory_Success(t *testing.T) {
	store := &mockStore{
		workHistory: domain.WorkHistoryEntry{
			ID:                   1,
			EmployerName:         "Acme Corp",
			JobTitle:             "Software Engineer",
			StartDate:            "2020-01",
			EndDate:              "2023-06",
			DateGranularityStart: "month",
			DateGranularityEnd:   "month",
			SortOrder:            1,
		},
	}
	svc := NewWorkHistoryService(store)

	entry, err := svc.CreateWorkHistory(context.Background(), validInput())
	require.NoError(t, err)
	assert.Equal(t, int64(1), entry.ID)
	assert.Equal(t, "Acme Corp", entry.EmployerName)
	require.Len(t, store.createCalls, 1)
	assert.Equal(t, "Acme Corp", store.createCalls[0].EmployerName)
}

func TestCreateWorkHistory_PresentJob(t *testing.T) {
	store := &mockStore{
		workHistory: domain.WorkHistoryEntry{
			ID:                   1,
			EmployerName:         "Current Co",
			JobTitle:             "Lead",
			StartDate:            "2023-01",
			DateGranularityStart: "month",
			SortOrder:            1,
		},
	}
	svc := NewWorkHistoryService(store)

	input := domain.WorkHistoryInput{
		EmployerName:         "Current Co",
		JobTitle:             "Lead",
		StartDate:            "2023-01",
		EndDate:              "",
		DateGranularityStart: "month",
		DateGranularityEnd:   "",
	}

	entry, err := svc.CreateWorkHistory(context.Background(), input)
	require.NoError(t, err)
	assert.Empty(t, entry.EndDate)
}

func TestCreateWorkHistory_StoreError(t *testing.T) {
	store := &mockStore{err: fmt.Errorf("db connection lost")}
	svc := NewWorkHistoryService(store)

	_, err := svc.CreateWorkHistory(context.Background(), validInput())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db connection lost")
}

// =================================================================
// Happy path tests — Update
// =================================================================

func TestUpdateWorkHistory_Success(t *testing.T) {
	store := &mockStore{
		workHistory: domain.WorkHistoryEntry{
			ID:                   42,
			EmployerName:         "New Name",
			JobTitle:             "New Title",
			StartDate:            "2019-06",
			EndDate:              "2023-12",
			DateGranularityStart: "month",
			DateGranularityEnd:   "month",
			SortOrder:            1,
		},
	}
	svc := NewWorkHistoryService(store)

	input := domain.WorkHistoryInput{
		EmployerName:         "New Name",
		JobTitle:             "New Title",
		StartDate:            "2019-06",
		EndDate:              "2023-12",
		DateGranularityStart: "month",
		DateGranularityEnd:   "month",
	}

	entry, err := svc.UpdateWorkHistory(context.Background(), 42, input)
	require.NoError(t, err)
	assert.Equal(t, int64(42), entry.ID)
	assert.Equal(t, "New Name", entry.EmployerName)
	require.Len(t, store.updateCalls, 1)
	assert.Equal(t, int64(42), store.updateCalls[0].ID)
}

// =================================================================
// Happy path tests — Delete
// =================================================================

func TestDeleteWorkHistory_Success(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	err := svc.DeleteWorkHistory(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, store.deleteCalls, 1)
	assert.Equal(t, int64(7), store.deleteCalls[0])
}

func TestDeleteWorkHistory_StoreError(t *testing.T) {
	store := &mockStore{deleteErr: fmt.Errorf("foreign key violation")}
	svc := NewWorkHistoryService(store)

	err := svc.DeleteWorkHistory(context.Background(), 7)
	require.Error(t, err)
}

// =================================================================
// Happy path tests — List
// =================================================================

func TestListWorkHistory_Success(t *testing.T) {
	store := &mockStore{
		workHistories: []domain.WorkHistoryEntry{
			{ID: 1, EmployerName: "First", SortOrder: 1},
			{ID: 2, EmployerName: "Second", SortOrder: 2},
		},
	}
	svc := NewWorkHistoryService(store)

	entries, err := svc.ListWorkHistory(context.Background())
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "First", entries[0].EmployerName)
}

func TestListWorkHistory_Empty(t *testing.T) {
	store := &mockStore{workHistories: []domain.WorkHistoryEntry{}}
	svc := NewWorkHistoryService(store)

	entries, err := svc.ListWorkHistory(context.Background())
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// =================================================================
// Happy path tests — Reorder
// =================================================================

func TestReorderWorkHistory_Success(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	err := svc.ReorderWorkHistory(context.Background(), []int64{3, 1, 2})
	require.NoError(t, err)
	require.Len(t, store.reorderCalls, 1)
	assert.Equal(t, []int64{3, 1, 2}, store.reorderCalls[0])
}

func TestReorderWorkHistory_EmptySlice(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	err := svc.ReorderWorkHistory(context.Background(), []int64{})
	require.NoError(t, err)
}

// =================================================================
// Bullet CRUD — happy paths
// =================================================================

func TestCreateBullet_Success(t *testing.T) {
	store := &mockStore{
		bullet: domain.AchievementBullet{
			ID:            1,
			WorkHistoryID: 10,
			Text:          "Led project X",
			SortOrder:     1,
		},
	}
	svc := NewWorkHistoryService(store)

	bullet, err := svc.CreateBullet(context.Background(), 10, "Led project X")
	require.NoError(t, err)
	assert.Equal(t, "Led project X", bullet.Text)
	require.Len(t, store.createBulletCalls, 1)
	assert.Equal(t, int64(10), store.createBulletCalls[0].WorkHistoryID)
}

func TestCreateBullet_EmptyText(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	_, err := svc.CreateBullet(context.Background(), 10, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "text")
	assert.Empty(t, store.createBulletCalls)
}

func TestCreateBullet_WhitespaceOnlyText(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	_, err := svc.CreateBullet(context.Background(), 10, "   \n\t  ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "text")
	assert.Empty(t, store.createBulletCalls)
}

func TestUpdateBullet_Success(t *testing.T) {
	store := &mockStore{
		bullet: domain.AchievementBullet{
			ID:            5,
			WorkHistoryID: 10,
			Text:          "Updated text",
			SortOrder:     1,
		},
	}
	svc := NewWorkHistoryService(store)

	bullet, err := svc.UpdateBullet(context.Background(), 5, "Updated text")
	require.NoError(t, err)
	assert.Equal(t, "Updated text", bullet.Text)
	require.Len(t, store.updateBulletCalls, 1)
	assert.Equal(t, int64(5), store.updateBulletCalls[0].ID)
}

func TestUpdateBullet_EmptyText(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	_, err := svc.UpdateBullet(context.Background(), 5, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "text")
	assert.Empty(t, store.updateBulletCalls)
}

func TestDeleteBullet_Success(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	err := svc.DeleteBullet(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, store.deleteBulletCalls, 1)
	assert.Equal(t, int64(5), store.deleteBulletCalls[0])
}

func TestReorderBullets_Success(t *testing.T) {
	store := &mockStore{}
	svc := NewWorkHistoryService(store)

	err := svc.ReorderBullets(context.Background(), 10, []int64{3, 1, 2})
	require.NoError(t, err)
	require.NotNil(t, store.reorderBulletCall)
	assert.Equal(t, int64(10), store.reorderBulletCall.WorkHistoryID)
	assert.Equal(t, []int64{3, 1, 2}, store.reorderBulletCall.OrderedIDs)
}

// =================================================================
// SplitBulletText tests
// =================================================================

func TestSplitBulletText_MultipleLines(t *testing.T) {
	svc := NewWorkHistoryService(nil) // no store needed

	result := svc.SplitBulletText("Led project X\nImproved perf by 50%\nMentored 3 juniors")
	require.Len(t, result, 3)
	assert.Equal(t, "Led project X", result[0])
	assert.Equal(t, "Improved perf by 50%", result[1])
	assert.Equal(t, "Mentored 3 juniors", result[2])
}

func TestSplitBulletText_EmptyInput(t *testing.T) {
	svc := NewWorkHistoryService(nil)

	result := svc.SplitBulletText("")
	assert.Empty(t, result)
}

func TestSplitBulletText_TrailingNewlines(t *testing.T) {
	svc := NewWorkHistoryService(nil)

	result := svc.SplitBulletText("Led project X\n\n")
	require.Len(t, result, 1)
	assert.Equal(t, "Led project X", result[0])
}

func TestSplitBulletText_BlankLinesFiltered(t *testing.T) {
	svc := NewWorkHistoryService(nil)

	result := svc.SplitBulletText("Line 1\n\n\nLine 2\n  \nLine 3")
	require.Len(t, result, 3)
	assert.Equal(t, "Line 1", result[0])
	assert.Equal(t, "Line 2", result[1])
	assert.Equal(t, "Line 3", result[2])
}

func TestSplitBulletText_TrimWhitespace(t *testing.T) {
	svc := NewWorkHistoryService(nil)

	result := svc.SplitBulletText("  Led project X  \n  Improved perf  ")
	require.Len(t, result, 2)
	assert.Equal(t, "Led project X", result[0])
	assert.Equal(t, "Improved perf", result[1])
}

func TestSplitBulletText_CarriageReturns(t *testing.T) {
	svc := NewWorkHistoryService(nil)

	result := svc.SplitBulletText("Line 1\r\nLine 2\r\nLine 3")
	require.Len(t, result, 3)
	assert.Equal(t, "Line 1", result[0])
	assert.Equal(t, "Line 2", result[1])
	assert.Equal(t, "Line 3", result[2])
}

func TestSplitBulletText_SingleLine(t *testing.T) {
	svc := NewWorkHistoryService(nil)

	result := svc.SplitBulletText("Just one bullet")
	require.Len(t, result, 1)
	assert.Equal(t, "Just one bullet", result[0])
}

// =================================================================
// GetWorkHistory — delegation
// =================================================================

func TestGetWorkHistory_Success(t *testing.T) {
	store := &mockStore{
		workHistory: domain.WorkHistoryEntry{
			ID:           1,
			EmployerName: "Acme Corp",
		},
	}
	svc := NewWorkHistoryService(store)

	entry, err := svc.GetWorkHistory(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", entry.EmployerName)
}

func TestGetWorkHistory_NotFound(t *testing.T) {
	store := &mockStore{err: fmt.Errorf("work_history: not found: id=999")}
	svc := NewWorkHistoryService(store)

	_, err := svc.GetWorkHistory(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
