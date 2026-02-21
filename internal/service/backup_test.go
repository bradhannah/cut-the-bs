package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBackupStore implements BackupStore for testing.
type mockBackupStore struct {
	exportData domain.ExportData
	importData domain.ExportData
	err        error
	dbPath     string

	// call tracking
	exportCalls int
	importCalls int
}

func (m *mockBackupStore) ExportAllData(_ context.Context) (domain.ExportData, error) {
	m.exportCalls++
	return m.exportData, m.err
}

func (m *mockBackupStore) ImportAllData(_ context.Context, data domain.ExportData) error {
	m.importCalls++
	m.importData = data
	return m.err
}

func (m *mockBackupStore) Checkpoint() error {
	return nil
}

func (m *mockBackupStore) DBPath() string {
	return m.dbPath
}

// =================================================================
// ExportAllData
// =================================================================

func TestBackupService_ExportAllData_Success(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{
		exportData: domain.ExportData{
			SchemaVersion: SchemaVersion,
			ExportedAt:    "2026-01-01T00:00:00Z",
			Profile:       domain.UserProfile{ID: 1, FullName: "Test User"},
		},
	}
	svc := NewBackupService(store)

	outputPath := filepath.Join(dir, "export.json")
	result, err := svc.ExportAllData(context.Background(), outputPath)
	require.NoError(t, err)
	assert.Equal(t, outputPath, result)
	assert.Equal(t, 1, store.exportCalls)

	// Verify the file was written and is valid JSON.
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	var exported domain.ExportData
	require.NoError(t, json.Unmarshal(data, &exported))
	assert.Equal(t, SchemaVersion, exported.SchemaVersion)
	assert.Equal(t, "Test User", exported.Profile.FullName)
}

func TestBackupService_ExportAllData_StoreError(t *testing.T) {
	store := &mockBackupStore{
		err: fmt.Errorf("database locked"),
	}
	svc := NewBackupService(store)

	_, err := svc.ExportAllData(context.Background(), "/tmp/export.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database locked")
}

func TestBackupService_ExportAllData_InvalidPath(t *testing.T) {
	store := &mockBackupStore{
		exportData: domain.ExportData{SchemaVersion: SchemaVersion},
	}
	svc := NewBackupService(store)

	// Non-existent deeply nested directory.
	_, err := svc.ExportAllData(context.Background(),
		"/nonexistent/deep/path/export.json")
	require.Error(t, err)
}

// =================================================================
// ImportAllData
// =================================================================

func TestBackupService_ImportAllData_Success(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	// Write a valid export file.
	exportData := domain.ExportData{
		SchemaVersion: SchemaVersion,
		ExportedAt:    "2026-01-01T00:00:00Z",
		Profile:       domain.UserProfile{ID: 1, FullName: "Restored User"},
		Skills: []domain.Skill{
			{ID: 1, Name: "Go", CategoryID: 1, CompetenceLevel: 8},
		},
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	require.NoError(t, err)

	inputPath := filepath.Join(dir, "backup.json")
	require.NoError(t, os.WriteFile(inputPath, data, 0o644))

	err = svc.ImportAllData(context.Background(), inputPath)
	require.NoError(t, err)
	assert.Equal(t, 1, store.importCalls)
	assert.Equal(t, "Restored User", store.importData.Profile.FullName)
	assert.Len(t, store.importData.Skills, 1)
}

func TestBackupService_ImportAllData_FileNotFound(t *testing.T) {
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	err := svc.ImportAllData(context.Background(), "/nonexistent/file.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read")
}

func TestBackupService_ImportAllData_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	inputPath := filepath.Join(dir, "bad.json")
	require.NoError(t, os.WriteFile(inputPath, []byte("not json"), 0o644))

	err := svc.ImportAllData(context.Background(), inputPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestBackupService_ImportAllData_WrongSchemaVersion(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	exportData := domain.ExportData{
		SchemaVersion: 999,
		ExportedAt:    "2026-01-01T00:00:00Z",
	}
	data, _ := json.Marshal(exportData)
	inputPath := filepath.Join(dir, "bad_version.json")
	require.NoError(t, os.WriteFile(inputPath, data, 0o644))

	err := svc.ImportAllData(context.Background(), inputPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema version")
}

func TestBackupService_ImportAllData_StoreError(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{
		err: fmt.Errorf("constraint violation"),
	}
	svc := NewBackupService(store)

	exportData := domain.ExportData{
		SchemaVersion: SchemaVersion,
		ExportedAt:    "2026-01-01T00:00:00Z",
	}
	data, _ := json.Marshal(exportData)
	inputPath := filepath.Join(dir, "backup.json")
	require.NoError(t, os.WriteFile(inputPath, data, 0o644))

	err := svc.ImportAllData(context.Background(), inputPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "constraint violation")
}

// =================================================================
// ImportCSV
// =================================================================

func TestBackupService_ImportCSV_Skills(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	csvContent := "name,category_id,competence_level,is_legacy\nGo,1,8,false\nRust,1,5,false\n"
	csvPath := filepath.Join(dir, "skills.csv")
	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0o644))

	result, err := svc.ImportCSV(context.Background(), csvPath, "skills")
	require.NoError(t, err)
	assert.Equal(t, 2, result.RecordsImported)
	assert.Equal(t, 0, result.RecordsSkipped)
}

func TestBackupService_ImportCSV_UnsupportedType(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	csvPath := filepath.Join(dir, "test.csv")
	require.NoError(t, os.WriteFile(csvPath, []byte("header\nval\n"), 0o644))

	_, err := svc.ImportCSV(context.Background(), csvPath, "unsupported_type")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestBackupService_ImportCSV_FileNotFound(t *testing.T) {
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	_, err := svc.ImportCSV(context.Background(), "/nonexistent.csv", "skills")
	require.Error(t, err)
}

func TestBackupService_ImportCSV_WorkHistory(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	csvContent := "employer_name,job_title,start_date,end_date\nAcme Corp,Developer,2020-01,2023-06\n"
	csvPath := filepath.Join(dir, "work_history.csv")
	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0o644))

	result, err := svc.ImportCSV(context.Background(), csvPath, "work_history")
	require.NoError(t, err)
	assert.Equal(t, 1, result.RecordsImported)
}

// =================================================================
// ImportJSON (partial)
// =================================================================

func TestBackupService_ImportJSON_Skills(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	skills := []domain.Skill{
		{Name: "Go", CategoryID: 1, CompetenceLevel: 8},
		{Name: "Rust", CategoryID: 1, CompetenceLevel: 5},
	}
	data, _ := json.Marshal(skills)
	jsonPath := filepath.Join(dir, "skills.json")
	require.NoError(t, os.WriteFile(jsonPath, data, 0o644))

	result, err := svc.ImportJSON(context.Background(), jsonPath, "skills")
	require.NoError(t, err)
	assert.Equal(t, 2, result.RecordsImported)
}

func TestBackupService_ImportJSON_UnsupportedType(t *testing.T) {
	dir := t.TempDir()
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	jsonPath := filepath.Join(dir, "test.json")
	require.NoError(t, os.WriteFile(jsonPath, []byte("[]"), 0o644))

	_, err := svc.ImportJSON(context.Background(), jsonPath, "unsupported_type")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestBackupService_ImportJSON_FileNotFound(t *testing.T) {
	store := &mockBackupStore{}
	svc := NewBackupService(store)

	_, err := svc.ImportJSON(context.Background(), "/nonexistent.json", "skills")
	require.Error(t, err)
}

// =================================================================
// RollingBackup
// =================================================================

func TestBackupService_RollingBackup_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("fake db"), 0o644))

	backupDir := filepath.Join(dir, "backups")
	require.NoError(t, os.MkdirAll(backupDir, 0o755))

	store := &mockBackupStore{dbPath: dbPath}
	svc := NewBackupService(store)

	path, err := svc.RollingBackup(backupDir, 3)
	require.NoError(t, err)
	assert.NotEmpty(t, path)

	// Verify the backup file exists.
	_, err = os.Stat(path)
	require.NoError(t, err)

	// Verify contents match.
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "fake db", string(content))
}

func TestBackupService_RollingBackup_PrunesOldBackups(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("fake db"), 0o644))

	backupDir := filepath.Join(dir, "backups")
	require.NoError(t, os.MkdirAll(backupDir, 0o755))

	// Pre-create 4 older backup files with distinct timestamps.
	for i := 0; i < 4; i++ {
		name := fmt.Sprintf("cut-the-bs-20260101-00000%d.db", i)
		path := filepath.Join(backupDir, name)
		require.NoError(t, os.WriteFile(path, []byte("old"), 0o644))
	}

	store := &mockBackupStore{dbPath: dbPath}
	svc := NewBackupService(store)

	// Create one more backup with maxCount=3 — oldest 2 should be
	// pruned so only 3 remain.
	path, err := svc.RollingBackup(backupDir, 3)
	require.NoError(t, err)
	assert.NotEmpty(t, path)

	// Count remaining backup files.
	entries, err := os.ReadDir(backupDir)
	require.NoError(t, err)

	backupCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			backupCount++
		}
	}
	assert.Equal(t, 3, backupCount)
}

func TestBackupService_RollingBackup_DBNotFound(t *testing.T) {
	store := &mockBackupStore{dbPath: "/nonexistent/db.sqlite"}
	svc := NewBackupService(store)

	_, err := svc.RollingBackup(t.TempDir(), 3)
	require.Error(t, err)
}
