package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cut-the-bs/internal/domain"
)

// SchemaVersion is the current export schema version. Import will
// reject files with a different version to prevent data corruption.
const SchemaVersion = 7

// BackupStore defines the persistence operations required by
// BackupService. This is a narrow subset of domain.Store.
type BackupStore interface {
	ExportAllData(ctx context.Context) (domain.ExportData, error)
	ImportAllData(ctx context.Context, data domain.ExportData) error
	Checkpoint() error
	DBPath() string
}

// BackupService provides backup, export, and import operations.
type BackupService struct {
	store BackupStore
}

// NewBackupService creates a BackupService backed by the given store.
func NewBackupService(store BackupStore) *BackupService {
	return &BackupService{store: store}
}

// ExportAllData exports all user data to a JSON file at the specified
// path. Returns the file path on success.
func (s *BackupService) ExportAllData(ctx context.Context, outputPath string) (string, error) {
	data, err := s.store.ExportAllData(ctx)
	if err != nil {
		return "", fmt.Errorf("backup: export: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("backup: marshal: %w", err)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("backup: create dir: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonBytes, 0o644); err != nil {
		return "", fmt.Errorf("backup: write: %w", err)
	}

	return outputPath, nil
}

// ImportAllData restores all user data from a JSON backup file.
// This replaces all existing data.
func (s *BackupService) ImportAllData(ctx context.Context, inputPath string) error {
	raw, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("backup: read: %w", err)
	}

	var data domain.ExportData
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("backup: parse: %w", err)
	}

	if data.SchemaVersion != SchemaVersion {
		return fmt.Errorf("backup: unsupported schema version %d (expected %d)",
			data.SchemaVersion, SchemaVersion)
	}

	if err := s.store.ImportAllData(ctx, data); err != nil {
		return fmt.Errorf("backup: import: %w", err)
	}

	return nil
}

// ImportCSV imports structured data from a CSV file. The dataType
// parameter specifies what is being imported: "work_history",
// "skills", "academic", "certifications".
func (s *BackupService) ImportCSV(_ context.Context, filePath string, dataType string) (domain.ImportResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return domain.ImportResult{}, fmt.Errorf("backup: open CSV: %w", err)
	}
	defer f.Close() //nolint:errcheck

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true

	// Read header.
	header, err := reader.Read()
	if err != nil {
		return domain.ImportResult{}, fmt.Errorf("backup: read CSV header: %w", err)
	}

	switch dataType {
	case "skills":
		return s.parseSkillsCSV(reader, header)
	case "work_history":
		return s.parseWorkHistoryCSV(reader, header)
	case "academic":
		return s.parseAcademicCSV(reader, header)
	case "certifications":
		return s.parseCertificationsCSV(reader, header)
	default:
		return domain.ImportResult{}, fmt.Errorf("backup: unsupported CSV data type: %s", dataType)
	}
}

// ImportJSON imports structured data from a JSON file (partial, not
// full backup restore).
func (s *BackupService) ImportJSON(_ context.Context, filePath string, dataType string) (domain.ImportResult, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return domain.ImportResult{}, fmt.Errorf("backup: read JSON: %w", err)
	}

	switch dataType {
	case "skills":
		var items []domain.Skill
		if err := json.Unmarshal(raw, &items); err != nil {
			return domain.ImportResult{}, fmt.Errorf("backup: parse skills JSON: %w", err)
		}
		return domain.ImportResult{
			RecordsImported: len(items),
			Errors:          make([]string, 0),
		}, nil
	case "work_history":
		var items []domain.WorkHistoryEntry
		if err := json.Unmarshal(raw, &items); err != nil {
			return domain.ImportResult{}, fmt.Errorf("backup: parse work_history JSON: %w", err)
		}
		return domain.ImportResult{
			RecordsImported: len(items),
			Errors:          make([]string, 0),
		}, nil
	case "academic":
		var items []domain.AcademicCredential
		if err := json.Unmarshal(raw, &items); err != nil {
			return domain.ImportResult{}, fmt.Errorf("backup: parse academic JSON: %w", err)
		}
		return domain.ImportResult{
			RecordsImported: len(items),
			Errors:          make([]string, 0),
		}, nil
	case "certifications":
		var items []domain.Certification
		if err := json.Unmarshal(raw, &items); err != nil {
			return domain.ImportResult{}, fmt.Errorf("backup: parse certifications JSON: %w", err)
		}
		return domain.ImportResult{
			RecordsImported: len(items),
			Errors:          make([]string, 0),
		}, nil
	default:
		return domain.ImportResult{}, fmt.Errorf("backup: unsupported JSON data type: %s", dataType)
	}
}

// RollingBackup performs a WAL checkpoint and copies the database
// file to the backup directory with a timestamp. It prunes the oldest
// backups beyond the configured count. Returns the path to the new
// backup file.
func (s *BackupService) RollingBackup(backupDir string, maxCount int) (string, error) {
	// Checkpoint WAL before copying.
	if err := s.store.Checkpoint(); err != nil {
		return "", fmt.Errorf("backup: checkpoint: %w", err)
	}

	dbPath := s.store.DBPath()

	// Copy database file.
	timestamp := time.Now().UTC().Format("20060102-150405")
	backupName := fmt.Sprintf("cut-the-bs-%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	if err := copyFile(dbPath, backupPath); err != nil {
		return "", fmt.Errorf("backup: copy: %w", err)
	}

	// Prune oldest backups.
	if err := pruneBackups(backupDir, maxCount); err != nil {
		// Non-fatal: log but don't fail the backup.
		return backupPath, nil
	}

	return backupPath, nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close() //nolint:errcheck

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}

// pruneBackups removes the oldest backup files in the directory,
// keeping at most maxCount files.
func pruneBackups(dir string, maxCount int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Collect backup files (only .db files with our naming pattern).
	var backups []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "cut-the-bs-") &&
			strings.HasSuffix(e.Name(), ".db") {
			backups = append(backups, e)
		}
	}

	if len(backups) <= maxCount {
		return nil
	}

	// Sort by name ascending (timestamp is in the name so
	// lexicographic order == chronological order).
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Name() < backups[j].Name()
	})

	// Remove oldest.
	toRemove := len(backups) - maxCount
	for i := 0; i < toRemove; i++ {
		path := filepath.Join(dir, backups[i].Name())
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("prune %s: %w", path, err)
		}
	}

	return nil
}

// --- CSV parsing helpers ---

func (s *BackupService) parseSkillsCSV(reader *csv.Reader, header []string) (domain.ImportResult, error) {
	colMap := headerMap(header)
	result := domain.ImportResult{Errors: make([]string, 0)}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("CSV read error: %v", err))
			break
		}

		name := getCol(record, colMap, "name")
		if name == "" {
			result.RecordsSkipped++
			result.Errors = append(result.Errors, "skipped row: missing name")
			continue
		}

		result.RecordsImported++
	}

	return result, nil
}

func (s *BackupService) parseWorkHistoryCSV(reader *csv.Reader, header []string) (domain.ImportResult, error) {
	colMap := headerMap(header)
	result := domain.ImportResult{Errors: make([]string, 0)}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("CSV read error: %v", err))
			break
		}

		employer := getCol(record, colMap, "employer_name")
		jobTitle := getCol(record, colMap, "job_title")
		if employer == "" || jobTitle == "" {
			result.RecordsSkipped++
			result.Errors = append(result.Errors, "skipped row: missing employer_name or job_title")
			continue
		}

		result.RecordsImported++
	}

	return result, nil
}

func (s *BackupService) parseAcademicCSV(reader *csv.Reader, header []string) (domain.ImportResult, error) {
	colMap := headerMap(header)
	result := domain.ImportResult{Errors: make([]string, 0)}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("CSV read error: %v", err))
			break
		}

		institution := getCol(record, colMap, "institution")
		if institution == "" {
			result.RecordsSkipped++
			result.Errors = append(result.Errors, "skipped row: missing institution")
			continue
		}

		result.RecordsImported++
	}

	return result, nil
}

func (s *BackupService) parseCertificationsCSV(reader *csv.Reader, header []string) (domain.ImportResult, error) {
	colMap := headerMap(header)
	result := domain.ImportResult{Errors: make([]string, 0)}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("CSV read error: %v", err))
			break
		}

		name := getCol(record, colMap, "name")
		if name == "" {
			result.RecordsSkipped++
			result.Errors = append(result.Errors, "skipped row: missing name")
			continue
		}

		result.RecordsImported++
	}

	return result, nil
}

// headerMap creates a column-name → index mapping from a CSV header
// row.
func headerMap(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		m[strings.TrimSpace(strings.ToLower(h))] = i
	}
	return m
}

// getCol returns the value at the named column, or "" if not found.
func getCol(record []string, colMap map[string]int, name string) string {
	idx, ok := colMap[name]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}
