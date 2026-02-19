package sqlite

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testStore creates a temporary in-memory Store for testing.
func testStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	store, err := NewStore(dbPath, logger)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

func TestMigrateV1_CreatesAllTables(t *testing.T) {
	store := testStore(t)
	err := Migrate(store)
	require.NoError(t, err)

	// All 28 tables that should exist after v1 migration.
	expectedTables := []string{
		"user_profile",
		"profile_link",
		"work_history_entry",
		"achievement_bullet",
		"skill_category",
		"skill",
		"academic_credential",
		"certification",
		"professional_summary",
		"role_descriptor",
		"lens",
		"lens_work_history_selection",
		"lens_bullet_selection",
		"lens_skill_selection",
		"lens_academic_selection",
		"lens_cert_selection",
		"lens_descriptor_selection",
		"skill_lens_tag",
		"resume_export",
		"export_work_history_selection",
		"export_bullet_selection",
		"export_skill_selection",
		"export_academic_selection",
		"export_cert_selection",
		"export_descriptor_selection",
		"cover_letter",
		"job_application",
		"status_history",
	}

	for _, table := range expectedTables {
		var count int
		err := store.DB().QueryRow(
			"SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "table %q should exist", table)
	}
}

func TestMigrateV1_SetsUserVersion(t *testing.T) {
	store := testStore(t)
	err := Migrate(store)
	require.NoError(t, err)

	var version int
	err = store.DB().QueryRow("PRAGMA user_version").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, 1, version)
}

func TestMigrateV1_IdempotentRerun(t *testing.T) {
	store := testStore(t)

	// Run migration twice — second run should be a no-op.
	err := Migrate(store)
	require.NoError(t, err)

	err = Migrate(store)
	require.NoError(t, err)

	// Verify version is still 1.
	var version int
	err = store.DB().QueryRow("PRAGMA user_version").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, 1, version)
}

func TestMigrateV1_ForeignKeyConstraints(t *testing.T) {
	store := testStore(t)
	err := Migrate(store)
	require.NoError(t, err)

	// Verify foreign keys are enforced: inserting an achievement_bullet
	// with a non-existent work_history_id should fail.
	_, err = store.DB().Exec(
		`INSERT INTO achievement_bullet (work_history_id, text, sort_order)
		 VALUES (9999, 'test bullet', 0)`,
	)
	assert.Error(t, err, "FK constraint should reject invalid work_history_id")

	// Verify cascade delete: create a work history entry, add a bullet,
	// then delete the entry — the bullet should be deleted too.
	res, err := store.DB().Exec(
		`INSERT INTO work_history_entry (employer_name, job_title, start_date, date_granularity_start, sort_order)
		 VALUES ('Acme', 'Dev', '2023', 'year', 0)`,
	)
	require.NoError(t, err)
	entryID, err := res.LastInsertId()
	require.NoError(t, err)

	_, err = store.DB().Exec(
		`INSERT INTO achievement_bullet (work_history_id, text, sort_order)
		 VALUES (?, 'did stuff', 0)`, entryID,
	)
	require.NoError(t, err)

	// Delete the parent entry.
	_, err = store.DB().Exec("DELETE FROM work_history_entry WHERE id = ?", entryID)
	require.NoError(t, err)

	// Bullet should be gone due to CASCADE.
	var bulletCount int
	err = store.DB().QueryRow(
		"SELECT count(*) FROM achievement_bullet WHERE work_history_id = ?",
		entryID,
	).Scan(&bulletCount)
	require.NoError(t, err)
	assert.Equal(t, 0, bulletCount, "bullets should be cascade-deleted with parent entry")
}

func TestMigrateV1_SkillCompetenceConstraint(t *testing.T) {
	store := testStore(t)
	err := Migrate(store)
	require.NoError(t, err)

	// Create a skill category first.
	_, err = store.DB().Exec(
		`INSERT INTO skill_category (name, sort_order) VALUES ('Test', 0)`,
	)
	require.NoError(t, err)

	// competence_level out of range [1,10] should fail.
	_, err = store.DB().Exec(
		`INSERT INTO skill (name, category_id, competence_level, is_legacy)
		 VALUES ('bad', 1, 0, 0)`,
	)
	assert.Error(t, err, "competence_level 0 should be rejected by CHECK constraint")

	_, err = store.DB().Exec(
		`INSERT INTO skill (name, category_id, competence_level, is_legacy)
		 VALUES ('bad', 1, 11, 0)`,
	)
	assert.Error(t, err, "competence_level 11 should be rejected by CHECK constraint")

	// Valid level should work.
	_, err = store.DB().Exec(
		`INSERT INTO skill (name, category_id, competence_level, is_legacy)
		 VALUES ('good', 1, 5, 0)`,
	)
	assert.NoError(t, err, "competence_level 5 should be accepted")
}

func TestMigrateV1_UniqueConstraints(t *testing.T) {
	store := testStore(t)
	err := Migrate(store)
	require.NoError(t, err)

	// skill_category name must be unique.
	_, err = store.DB().Exec(
		`INSERT INTO skill_category (name, sort_order) VALUES ('Languages', 0)`,
	)
	require.NoError(t, err)

	_, err = store.DB().Exec(
		`INSERT INTO skill_category (name, sort_order) VALUES ('Languages', 1)`,
	)
	assert.Error(t, err, "duplicate skill_category name should fail UNIQUE constraint")

	// professional_summary label must be unique.
	_, err = store.DB().Exec(
		`INSERT INTO professional_summary (label, body_text) VALUES ('Leader', 'text')`,
	)
	require.NoError(t, err)

	_, err = store.DB().Exec(
		`INSERT INTO professional_summary (label, body_text) VALUES ('Leader', 'other')`,
	)
	assert.Error(t, err, "duplicate professional_summary label should fail UNIQUE constraint")

	// role_descriptor title must be unique.
	_, err = store.DB().Exec(
		`INSERT INTO role_descriptor (title, sort_order) VALUES ('Engineer', 0)`,
	)
	require.NoError(t, err)

	_, err = store.DB().Exec(
		`INSERT INTO role_descriptor (title, sort_order) VALUES ('Engineer', 1)`,
	)
	assert.Error(t, err, "duplicate role_descriptor title should fail UNIQUE constraint")

	// lens name must be unique.
	_, err = store.DB().Exec(
		`INSERT INTO lens (name) VALUES ('Sales')`,
	)
	require.NoError(t, err)

	_, err = store.DB().Exec(
		`INSERT INTO lens (name) VALUES ('Sales')`,
	)
	assert.Error(t, err, "duplicate lens name should fail UNIQUE constraint")
}

func TestMigrateV1_IndexesExist(t *testing.T) {
	store := testStore(t)
	err := Migrate(store)
	require.NoError(t, err)

	expectedIndexes := []string{
		"idx_achievement_bullet_work_history",
		"idx_skill_category_id",
		"idx_skill_competence",
		"idx_job_application_company",
		"idx_job_application_status",
		"idx_status_history_application",
		"idx_export_selections_export",
		"idx_export_bullet_export",
		"idx_export_skill_export",
		"idx_lens_work_history",
		"idx_lens_bullet",
		"idx_lens_skill",
		"idx_lens_academic",
		"idx_lens_cert",
		"idx_lens_descriptor",
		"idx_skill_lens_tag_skill",
		"idx_skill_lens_tag_lens",
	}

	for _, idx := range expectedIndexes {
		var count int
		err := store.DB().QueryRow(
			"SELECT count(*) FROM sqlite_master WHERE type='index' AND name=?",
			idx,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "index %q should exist", idx)
	}
}
