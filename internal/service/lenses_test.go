package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLensStore implements LensStore for testing.
type mockLensStore struct {
	lenses    []domain.Lens
	lens      domain.Lens
	detail    domain.LensDetail
	err       error
	lensIDs   []int64
	skillTags []domain.SkillWithTags

	// call tracking
	createCalls              []domain.LensInput
	updateCalls              []updateLensCall
	deleteCalls              []int64
	getCalls                 []int64
	setWHCalls               []setLensWHCall
	setSummaryCalls          []setLensSummaryCall
	setBulletCalls           []setLensBulletCall
	setSkillCalls            []setLensSkillCall
	setAcademicCalls         []setLensAcademicCall
	setCertCalls             []setLensCertCall
	setDescriptorCall        []setLensDescriptorCall
	getTagCalls              []int64
	setTagCalls              []setSkillTagCall
	checkLensNames           []string
	checkSummaryRefsCalls    []int64
	checkWHRefsCalls         []int64
	checkBulletRefsCalls     []int64
	checkAcademicRefsCalls   []int64
	checkCertRefsCalls       []int64
	checkDescriptorRefsCalls []int64
}

type setLensSummaryCall struct {
	LensID     int64
	Selections []domain.LensSummaryItem
}

type updateLensCall struct {
	ID    int64
	Input domain.LensInput
}

type setLensWHCall struct {
	LensID     int64
	Selections []domain.LensWorkHistoryItem
}

type setLensBulletCall struct {
	LensID     int64
	Selections []domain.LensBulletItem
}

type setLensSkillCall struct {
	LensID     int64
	Selections []domain.LensSkillItem
}

type setLensAcademicCall struct {
	LensID      int64
	AcademicIDs []int64
}

type setLensCertCall struct {
	LensID  int64
	CertIDs []int64
}

type setLensDescriptorCall struct {
	LensID     int64
	Selections []domain.LensDescriptorItem
}

type setSkillTagCall struct {
	SkillID int64
	LensIDs []int64
}

func (m *mockLensStore) ListLenses(_ context.Context) ([]domain.Lens, error) {
	return m.lenses, m.err
}

func (m *mockLensStore) GetLens(_ context.Context, id int64) (domain.LensDetail, error) {
	m.getCalls = append(m.getCalls, id)
	if m.err != nil {
		return domain.LensDetail{}, m.err
	}
	return m.detail, nil
}

func (m *mockLensStore) CreateLens(_ context.Context, input domain.LensInput) (domain.Lens, error) {
	m.createCalls = append(m.createCalls, input)
	if m.err != nil {
		return domain.Lens{}, m.err
	}
	return m.lens, nil
}

func (m *mockLensStore) UpdateLens(_ context.Context, id int64, input domain.LensInput) (domain.Lens, error) {
	m.updateCalls = append(m.updateCalls, updateLensCall{ID: id, Input: input})
	if m.err != nil {
		return domain.Lens{}, m.err
	}
	return m.lens, nil
}

func (m *mockLensStore) DeleteLens(_ context.Context, id int64) error {
	m.deleteCalls = append(m.deleteCalls, id)
	return m.err
}

func (m *mockLensStore) SetLensWorkHistory(_ context.Context, lensID int64, selections []domain.LensWorkHistoryItem) error {
	m.setWHCalls = append(m.setWHCalls, setLensWHCall{LensID: lensID, Selections: selections})
	return m.err
}

func (m *mockLensStore) SetLensBullets(_ context.Context, lensID int64, selections []domain.LensBulletItem) error {
	m.setBulletCalls = append(m.setBulletCalls, setLensBulletCall{LensID: lensID, Selections: selections})
	return m.err
}

func (m *mockLensStore) SetLensSkills(_ context.Context, lensID int64, selections []domain.LensSkillItem) error {
	m.setSkillCalls = append(m.setSkillCalls, setLensSkillCall{LensID: lensID, Selections: selections})
	return m.err
}

func (m *mockLensStore) SetLensAcademics(_ context.Context, lensID int64, academicIDs []int64) error {
	m.setAcademicCalls = append(m.setAcademicCalls, setLensAcademicCall{LensID: lensID, AcademicIDs: academicIDs})
	return m.err
}

func (m *mockLensStore) SetLensCerts(_ context.Context, lensID int64, certIDs []int64) error {
	m.setCertCalls = append(m.setCertCalls, setLensCertCall{LensID: lensID, CertIDs: certIDs})
	return m.err
}

func (m *mockLensStore) SetLensDescriptors(_ context.Context, lensID int64, selections []domain.LensDescriptorItem) error {
	m.setDescriptorCall = append(m.setDescriptorCall, setLensDescriptorCall{LensID: lensID, Selections: selections})
	return m.err
}

func (m *mockLensStore) GetSkillLensTags(_ context.Context, skillID int64) ([]int64, error) {
	m.getTagCalls = append(m.getTagCalls, skillID)
	if m.err != nil {
		return nil, m.err
	}
	return m.lensIDs, nil
}

func (m *mockLensStore) SetSkillLensTags(_ context.Context, skillID int64, lensIDs []int64) error {
	m.setTagCalls = append(m.setTagCalls, setSkillTagCall{SkillID: skillID, LensIDs: lensIDs})
	return m.err
}

func (m *mockLensStore) ListSkillsWithLensTags(_ context.Context) ([]domain.SkillWithTags, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.skillTags, nil
}

func (m *mockLensStore) CheckWorkHistoryLensReferences(_ context.Context, id int64) ([]string, error) {
	m.checkWHRefsCalls = append(m.checkWHRefsCalls, id)
	if m.err != nil {
		return nil, m.err
	}
	return m.checkLensNames, nil
}

func (m *mockLensStore) CheckBulletLensReferences(_ context.Context, id int64) ([]string, error) {
	m.checkBulletRefsCalls = append(m.checkBulletRefsCalls, id)
	if m.err != nil {
		return nil, m.err
	}
	return m.checkLensNames, nil
}

func (m *mockLensStore) CheckAcademicLensReferences(_ context.Context, id int64) ([]string, error) {
	m.checkAcademicRefsCalls = append(m.checkAcademicRefsCalls, id)
	if m.err != nil {
		return nil, m.err
	}
	return m.checkLensNames, nil
}

func (m *mockLensStore) CheckCertLensReferences(_ context.Context, id int64) ([]string, error) {
	m.checkCertRefsCalls = append(m.checkCertRefsCalls, id)
	if m.err != nil {
		return nil, m.err
	}
	return m.checkLensNames, nil
}

func (m *mockLensStore) CheckDescriptorLensReferences(_ context.Context, id int64) ([]string, error) {
	m.checkDescriptorRefsCalls = append(m.checkDescriptorRefsCalls, id)
	if m.err != nil {
		return nil, m.err
	}
	return m.checkLensNames, nil
}

func (m *mockLensStore) CheckSummaryLensReferences(_ context.Context, id int64) ([]string, error) {
	m.checkSummaryRefsCalls = append(m.checkSummaryRefsCalls, id)
	if m.err != nil {
		return nil, m.err
	}
	return m.checkLensNames, nil
}

func (m *mockLensStore) SetLensSummaries(_ context.Context, lensID int64, selections []domain.LensSummaryItem) error {
	m.setSummaryCalls = append(m.setSummaryCalls, setLensSummaryCall{LensID: lensID, Selections: selections})
	return m.err
}

// =================================================================
// ListLenses
// =================================================================

func TestLensService_ListLenses_Success(t *testing.T) {
	store := &mockLensStore{
		lenses: []domain.Lens{
			{ID: 1, Name: "Backend"},
			{ID: 2, Name: "Frontend"},
		},
	}
	svc := NewLensService(store)

	lenses, err := svc.ListLenses(context.Background())
	require.NoError(t, err)
	require.Len(t, lenses, 2)
	assert.Equal(t, "Backend", lenses[0].Name)
}

func TestLensService_ListLenses_Empty(t *testing.T) {
	store := &mockLensStore{lenses: []domain.Lens{}}
	svc := NewLensService(store)

	lenses, err := svc.ListLenses(context.Background())
	require.NoError(t, err)
	assert.Empty(t, lenses)
}

func TestLensService_ListLenses_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	_, err := svc.ListLenses(context.Background())
	require.Error(t, err)
}

// =================================================================
// GetLens
// =================================================================

func TestLensService_GetLens_Success(t *testing.T) {
	store := &mockLensStore{
		detail: domain.LensDetail{
			Lens:        domain.Lens{ID: 1, Name: "Backend"},
			WorkHistory: []domain.LensWorkHistoryItem{},
			Bullets:     []domain.LensBulletItem{},
			Skills:      []domain.LensSkillItem{},
			AcademicIDs: []int64{},
			CertIDs:     []int64{},
			Descriptors: []domain.LensDescriptorItem{},
		},
	}
	svc := NewLensService(store)

	detail, err := svc.GetLens(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Backend", detail.Name)
	require.Len(t, store.getCalls, 1)
}

func TestLensService_GetLens_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("not found")}
	svc := NewLensService(store)

	_, err := svc.GetLens(context.Background(), 999)
	require.Error(t, err)
}

// =================================================================
// CreateLens — validation
// =================================================================

func TestLensService_CreateLens_EmptyName(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	_, err := svc.CreateLens(context.Background(), domain.LensInput{Name: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lens name")
	assert.Empty(t, store.createCalls)
}

func TestLensService_CreateLens_WhitespaceName(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	_, err := svc.CreateLens(context.Background(), domain.LensInput{Name: "   "})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lens name")
	assert.Empty(t, store.createCalls)
}

func TestLensService_CreateLens_Success(t *testing.T) {
	store := &mockLensStore{
		lens: domain.Lens{ID: 1, Name: "Backend Engineer"},
	}
	svc := NewLensService(store)

	lens, err := svc.CreateLens(context.Background(), domain.LensInput{Name: "Backend Engineer"})
	require.NoError(t, err)
	assert.Equal(t, "Backend Engineer", lens.Name)
	require.Len(t, store.createCalls, 1)
}

func TestLensService_CreateLens_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("duplicate")}
	svc := NewLensService(store)

	_, err := svc.CreateLens(context.Background(), domain.LensInput{Name: "Backend"})
	require.Error(t, err)
}

// =================================================================
// UpdateLens — validation
// =================================================================

func TestLensService_UpdateLens_EmptyName(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	_, err := svc.UpdateLens(context.Background(), 1, domain.LensInput{Name: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lens name")
	assert.Empty(t, store.updateCalls)
}

func TestLensService_UpdateLens_Success(t *testing.T) {
	store := &mockLensStore{
		lens: domain.Lens{ID: 1, Name: "Backend Engineer"},
	}
	svc := NewLensService(store)

	lens, err := svc.UpdateLens(context.Background(), 1, domain.LensInput{Name: "Backend Engineer"})
	require.NoError(t, err)
	assert.Equal(t, "Backend Engineer", lens.Name)
	require.Len(t, store.updateCalls, 1)
}

func TestLensService_UpdateLens_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("not found")}
	svc := NewLensService(store)

	_, err := svc.UpdateLens(context.Background(), 999, domain.LensInput{Name: "X"})
	require.Error(t, err)
}

// =================================================================
// DeleteLens
// =================================================================

func TestLensService_DeleteLens_Success(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	err := svc.DeleteLens(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, store.deleteCalls, 1)
	assert.Equal(t, int64(5), store.deleteCalls[0])
}

func TestLensService_DeleteLens_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("not found")}
	svc := NewLensService(store)

	err := svc.DeleteLens(context.Background(), 999)
	require.Error(t, err)
}

// =================================================================
// Selection Setters — delegate to store
// =================================================================

func TestLensService_SetLensWorkHistory(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	selections := []domain.LensWorkHistoryItem{
		{WorkHistoryID: 1, SortOrder: 0},
		{WorkHistoryID: 2, SortOrder: 1},
	}
	err := svc.SetLensWorkHistory(context.Background(), 10, selections)
	require.NoError(t, err)
	require.Len(t, store.setWHCalls, 1)
	assert.Equal(t, int64(10), store.setWHCalls[0].LensID)
	assert.Len(t, store.setWHCalls[0].Selections, 2)
}

func TestLensService_SetLensBullets(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	err := svc.SetLensBullets(context.Background(), 10, []domain.LensBulletItem{
		{BulletID: 1, SortOrder: 0},
	})
	require.NoError(t, err)
	require.Len(t, store.setBulletCalls, 1)
}

func TestLensService_SetLensSkills(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	err := svc.SetLensSkills(context.Background(), 10, []domain.LensSkillItem{
		{SkillID: 1},
	})
	require.NoError(t, err)
	require.Len(t, store.setSkillCalls, 1)
}

func TestLensService_SetLensAcademics(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	err := svc.SetLensAcademics(context.Background(), 10, []int64{1, 2})
	require.NoError(t, err)
	require.Len(t, store.setAcademicCalls, 1)
	assert.Equal(t, []int64{1, 2}, store.setAcademicCalls[0].AcademicIDs)
}

func TestLensService_SetLensCerts(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	err := svc.SetLensCerts(context.Background(), 10, []int64{3, 4})
	require.NoError(t, err)
	require.Len(t, store.setCertCalls, 1)
}

func TestLensService_SetLensDescriptors(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	err := svc.SetLensDescriptors(context.Background(), 10, []domain.LensDescriptorItem{
		{DescriptorID: 1, SortOrder: 0},
	})
	require.NoError(t, err)
	require.Len(t, store.setDescriptorCall, 1)
}

func TestLensService_SetLensWorkHistory_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	err := svc.SetLensWorkHistory(context.Background(), 10, nil)
	require.Error(t, err)
}

// =================================================================
// Skill Lens Tags
// =================================================================

func TestLensService_GetSkillLensTags_Success(t *testing.T) {
	store := &mockLensStore{lensIDs: []int64{1, 2}}
	svc := NewLensService(store)

	tags, err := svc.GetSkillLensTags(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2}, tags)
	require.Len(t, store.getTagCalls, 1)
}

func TestLensService_SetSkillLensTags_Success(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	err := svc.SetSkillLensTags(context.Background(), 5, []int64{1, 2})
	require.NoError(t, err)
	require.Len(t, store.setTagCalls, 1)
	assert.Equal(t, int64(5), store.setTagCalls[0].SkillID)
}

func TestLensService_ListSkillsWithLensTags_Success(t *testing.T) {
	store := &mockLensStore{
		skillTags: []domain.SkillWithTags{
			{Skill: domain.Skill{ID: 1, Name: "Go"}, LensIDs: []int64{1}},
		},
	}
	svc := NewLensService(store)

	skills, err := svc.ListSkillsWithLensTags(context.Background())
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "Go", skills[0].Name)
}

// =================================================================
// GetLensExportSelections
// =================================================================

func TestLensService_GetLensExportSelections_Success(t *testing.T) {
	customSort := 3
	store := &mockLensStore{
		detail: domain.LensDetail{
			Lens: domain.Lens{ID: 1, Name: "Backend"},
			Summaries: []domain.LensSummaryItem{
				{SummaryID: 42, SortOrder: 0},
			},
			WorkHistory: []domain.LensWorkHistoryItem{
				{WorkHistoryID: 10, SortOrder: 0},
				{WorkHistoryID: 20, SortOrder: 1},
			},
			Bullets: []domain.LensBulletItem{
				{BulletID: 100, SortOrder: 0},
			},
			Skills: []domain.LensSkillItem{
				{SkillID: 5, CustomSortOrder: &customSort},
				{SkillID: 6, CustomSortOrder: nil},
			},
			AcademicIDs: []int64{7, 8},
			CertIDs:     []int64{9},
			Descriptors: []domain.LensDescriptorItem{
				{DescriptorID: 11, SortOrder: 0},
			},
		},
	}
	svc := NewLensService(store)

	req, err := svc.GetLensExportSelections(context.Background(), 1)
	require.NoError(t, err)

	// LensID should be set
	require.NotNil(t, req.LensID)
	assert.Equal(t, int64(1), *req.LensID)

	// SummaryIDs should be set
	assert.Equal(t, []int64{42}, req.SummaryIDs)

	// Work history IDs
	assert.Equal(t, []int64{10, 20}, req.WorkHistoryIDs)

	// Bullet IDs
	assert.Equal(t, []int64{100}, req.BulletIDs)

	// Skill IDs
	assert.Equal(t, []int64{5, 6}, req.SkillIDs)

	// Skill sort overrides — only skill 5 has custom sort
	require.Len(t, req.SkillSortOverrides, 1)
	assert.Equal(t, 3, req.SkillSortOverrides[5])

	// Academic IDs
	assert.Equal(t, []int64{7, 8}, req.AcademicIDs)

	// Certification IDs
	assert.Equal(t, []int64{9}, req.CertificationIDs)

	// Descriptor IDs
	assert.Equal(t, []int64{11}, req.DescriptorIDs)

	// TemplateID should be empty (user chooses)
	assert.Empty(t, req.TemplateID)
}

func TestLensService_GetLensExportSelections_NoSummary(t *testing.T) {
	store := &mockLensStore{
		detail: domain.LensDetail{
			Lens:        domain.Lens{ID: 1, Name: "Backend"},
			Summaries:   []domain.LensSummaryItem{},
			WorkHistory: []domain.LensWorkHistoryItem{},
			Bullets:     []domain.LensBulletItem{},
			Skills:      []domain.LensSkillItem{},
			AcademicIDs: []int64{},
			CertIDs:     []int64{},
			Descriptors: []domain.LensDescriptorItem{},
		},
	}
	svc := NewLensService(store)

	req, err := svc.GetLensExportSelections(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, req.SummaryIDs)
}

func TestLensService_GetLensExportSelections_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("not found")}
	svc := NewLensService(store)

	_, err := svc.GetLensExportSelections(context.Background(), 999)
	require.Error(t, err)
}

// =================================================================
// SetLensSummaries
// =================================================================

func TestLensService_SetLensSummaries_Success(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	selections := []domain.LensSummaryItem{
		{SummaryID: 1, SortOrder: 0},
	}
	err := svc.SetLensSummaries(context.Background(), 10, selections)
	require.NoError(t, err)
	require.Len(t, store.setSummaryCalls, 1)
	assert.Equal(t, int64(10), store.setSummaryCalls[0].LensID)
	assert.Equal(t, selections, store.setSummaryCalls[0].Selections)
}

func TestLensService_SetLensSummaries_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	err := svc.SetLensSummaries(context.Background(), 10, []domain.LensSummaryItem{
		{SummaryID: 1, SortOrder: 0},
	})
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestLensService_SetLensSummaries_EmptySlice(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	err := svc.SetLensSummaries(context.Background(), 10, []domain.LensSummaryItem{})
	require.NoError(t, err)
	require.Len(t, store.setSummaryCalls, 1)
	assert.Equal(t, int64(10), store.setSummaryCalls[0].LensID)
	assert.Empty(t, store.setSummaryCalls[0].Selections)
}

func TestLensService_SetLensSummaries_MultipleSummaries(t *testing.T) {
	store := &mockLensStore{}
	svc := NewLensService(store)

	selections := []domain.LensSummaryItem{
		{SummaryID: 1, SortOrder: 0},
		{SummaryID: 2, SortOrder: 1},
		{SummaryID: 3, SortOrder: 2},
	}
	err := svc.SetLensSummaries(context.Background(), 7, selections)
	require.NoError(t, err)
	require.Len(t, store.setSummaryCalls, 1)
	assert.Equal(t, int64(7), store.setSummaryCalls[0].LensID)
	require.Len(t, store.setSummaryCalls[0].Selections, 3)
	assert.Equal(t, int64(1), store.setSummaryCalls[0].Selections[0].SummaryID)
	assert.Equal(t, int64(2), store.setSummaryCalls[0].Selections[1].SummaryID)
	assert.Equal(t, int64(3), store.setSummaryCalls[0].Selections[2].SummaryID)
}

// =================================================================
// CheckSummaryLensReferences
// =================================================================

func TestLensService_CheckSummaryLensReferences_Success(t *testing.T) {
	store := &mockLensStore{
		checkLensNames: []string{"Backend", "Frontend"},
	}
	svc := NewLensService(store)

	names, err := svc.CheckSummaryLensReferences(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, []string{"Backend", "Frontend"}, names)
	require.Len(t, store.checkSummaryRefsCalls, 1)
	assert.Equal(t, int64(42), store.checkSummaryRefsCalls[0])
}

func TestLensService_CheckSummaryLensReferences_NoReferences(t *testing.T) {
	store := &mockLensStore{
		checkLensNames: []string{},
	}
	svc := NewLensService(store)

	names, err := svc.CheckSummaryLensReferences(context.Background(), 42)
	require.NoError(t, err)
	assert.Empty(t, names)
	require.Len(t, store.checkSummaryRefsCalls, 1)
}

func TestLensService_CheckSummaryLensReferences_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	_, err := svc.CheckSummaryLensReferences(context.Background(), 42)
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	require.Len(t, store.checkSummaryRefsCalls, 1)
}

// =================================================================
// GetLensExportSelections — summary-specific tests
// =================================================================

func TestLensService_GetLensExportSelections_MultipleSummaries(t *testing.T) {
	store := &mockLensStore{
		detail: domain.LensDetail{
			Lens: domain.Lens{ID: 5, Name: "Multi-Summary"},
			Summaries: []domain.LensSummaryItem{
				{SummaryID: 10, SortOrder: 0},
				{SummaryID: 20, SortOrder: 1},
				{SummaryID: 30, SortOrder: 2},
			},
			WorkHistory: []domain.LensWorkHistoryItem{},
			Bullets:     []domain.LensBulletItem{},
			Skills:      []domain.LensSkillItem{},
			AcademicIDs: []int64{},
			CertIDs:     []int64{},
			Descriptors: []domain.LensDescriptorItem{},
		},
	}
	svc := NewLensService(store)

	req, err := svc.GetLensExportSelections(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, []int64{10, 20, 30}, req.SummaryIDs)
}

func TestLensService_GetLensExportSelections_SummaryOrder(t *testing.T) {
	store := &mockLensStore{
		detail: domain.LensDetail{
			Lens: domain.Lens{ID: 5, Name: "Ordered"},
			Summaries: []domain.LensSummaryItem{
				{SummaryID: 99, SortOrder: 0},
				{SummaryID: 11, SortOrder: 1},
				{SummaryID: 55, SortOrder: 2},
			},
			WorkHistory: []domain.LensWorkHistoryItem{},
			Bullets:     []domain.LensBulletItem{},
			Skills:      []domain.LensSkillItem{},
			AcademicIDs: []int64{},
			CertIDs:     []int64{},
			Descriptors: []domain.LensDescriptorItem{},
		},
	}
	svc := NewLensService(store)

	req, err := svc.GetLensExportSelections(context.Background(), 5)
	require.NoError(t, err)
	// IDs should preserve the order from the detail.Summaries slice
	assert.Equal(t, []int64{99, 11, 55}, req.SummaryIDs)
}

// =================================================================
// Selection Setters — store error propagation
// =================================================================

func TestLensService_SetLensBullets_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	err := svc.SetLensBullets(context.Background(), 10, []domain.LensBulletItem{
		{BulletID: 1, SortOrder: 0},
	})
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestLensService_SetLensSkills_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	err := svc.SetLensSkills(context.Background(), 10, []domain.LensSkillItem{
		{SkillID: 1},
	})
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestLensService_SetLensAcademics_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	err := svc.SetLensAcademics(context.Background(), 10, []int64{1, 2})
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestLensService_SetLensCerts_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	err := svc.SetLensCerts(context.Background(), 10, []int64{3, 4})
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestLensService_SetLensDescriptors_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	err := svc.SetLensDescriptors(context.Background(), 10, []domain.LensDescriptorItem{
		{DescriptorID: 1, SortOrder: 0},
	})
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

// =================================================================
// Check*LensReferences — success + error tests
// =================================================================

func TestLensService_CheckWorkHistoryLensReferences_Success(t *testing.T) {
	store := &mockLensStore{
		checkLensNames: []string{"Backend", "Full-Stack"},
	}
	svc := NewLensService(store)

	names, err := svc.CheckWorkHistoryLensReferences(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, []string{"Backend", "Full-Stack"}, names)
	require.Len(t, store.checkWHRefsCalls, 1)
	assert.Equal(t, int64(7), store.checkWHRefsCalls[0])
}

func TestLensService_CheckWorkHistoryLensReferences_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	_, err := svc.CheckWorkHistoryLensReferences(context.Background(), 7)
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	require.Len(t, store.checkWHRefsCalls, 1)
}

func TestLensService_CheckBulletLensReferences_Success(t *testing.T) {
	store := &mockLensStore{
		checkLensNames: []string{"DevOps"},
	}
	svc := NewLensService(store)

	names, err := svc.CheckBulletLensReferences(context.Background(), 15)
	require.NoError(t, err)
	assert.Equal(t, []string{"DevOps"}, names)
	require.Len(t, store.checkBulletRefsCalls, 1)
	assert.Equal(t, int64(15), store.checkBulletRefsCalls[0])
}

func TestLensService_CheckAcademicLensReferences_Success(t *testing.T) {
	store := &mockLensStore{
		checkLensNames: []string{"Academic Focus"},
	}
	svc := NewLensService(store)

	names, err := svc.CheckAcademicLensReferences(context.Background(), 22)
	require.NoError(t, err)
	assert.Equal(t, []string{"Academic Focus"}, names)
	require.Len(t, store.checkAcademicRefsCalls, 1)
	assert.Equal(t, int64(22), store.checkAcademicRefsCalls[0])
}

func TestLensService_CheckCertLensReferences_Success(t *testing.T) {
	store := &mockLensStore{
		checkLensNames: []string{"Certified Pro"},
	}
	svc := NewLensService(store)

	names, err := svc.CheckCertLensReferences(context.Background(), 33)
	require.NoError(t, err)
	assert.Equal(t, []string{"Certified Pro"}, names)
	require.Len(t, store.checkCertRefsCalls, 1)
	assert.Equal(t, int64(33), store.checkCertRefsCalls[0])
}

func TestLensService_CheckDescriptorLensReferences_Success(t *testing.T) {
	store := &mockLensStore{
		checkLensNames: []string{"Descriptor Lens"},
	}
	svc := NewLensService(store)

	names, err := svc.CheckDescriptorLensReferences(context.Background(), 44)
	require.NoError(t, err)
	assert.Equal(t, []string{"Descriptor Lens"}, names)
	require.Len(t, store.checkDescriptorRefsCalls, 1)
	assert.Equal(t, int64(44), store.checkDescriptorRefsCalls[0])
}

// =================================================================
// Skill Lens Tags — store error tests
// =================================================================

func TestLensService_GetSkillLensTags_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	_, err := svc.GetSkillLensTags(context.Background(), 5)
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	require.Len(t, store.getTagCalls, 1)
}

func TestLensService_SetSkillLensTags_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	err := svc.SetSkillLensTags(context.Background(), 5, []int64{1, 2})
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	require.Len(t, store.setTagCalls, 1)
}

func TestLensService_ListSkillsWithLensTags_StoreError(t *testing.T) {
	store := &mockLensStore{err: fmt.Errorf("db error")}
	svc := NewLensService(store)

	_, err := svc.ListSkillsWithLensTags(context.Background())
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}
