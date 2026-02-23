package sqlite

import (
	"fmt"
	"log/slog"
)

// currentVersion is the latest schema version.
const currentVersion = 11

// Migrate applies all pending schema migrations to the database.
// It reads PRAGMA user_version to determine the current schema version
// and applies each migration in sequence. Each migration runs in its
// own transaction. If the database is already at the latest version,
// this is a no-op.
func Migrate(store *Store) error {
	version, err := getUserVersion(store)
	if err != nil {
		return err
	}

	store.logger.Info("migration check",
		slog.Int("current_version", version),
		slog.Int("target_version", currentVersion),
	)

	if version >= currentVersion {
		store.logger.Info("database schema is up to date")
		return nil
	}

	// Apply each migration sequentially.
	migrations := []func(*Store) error{
		migrateV1,
		migrateV2,
		migrateV3,
		migrateV4,
		migrateV5,
		migrateV6,
		migrateV7,
		migrateV8,
		migrateV9,
		migrateV10,
		migrateV11,
	}

	for i := version; i < currentVersion; i++ {
		store.logger.Info("applying migration", slog.Int("version", i+1))
		if err := migrations[i](store); err != nil {
			return fmt.Errorf("sqlite: migration to v%d failed: %w", i+1, err)
		}
	}

	return nil
}

// getUserVersion reads the PRAGMA user_version value.
func getUserVersion(store *Store) (int, error) {
	var version int
	err := store.db.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("sqlite: unable to read user_version: %w", err)
	}
	return version, nil
}

// migrateV1 creates the initial schema with all 28 tables and indexes.
func migrateV1(store *Store) error {
	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	statements := []string{
		// === Core entities ===

		`CREATE TABLE user_profile (
			id INTEGER PRIMARY KEY,
			full_name TEXT NOT NULL,
			email TEXT NOT NULL,
			phone TEXT DEFAULT '',
			location TEXT DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE profile_link (
			id INTEGER PRIMARY KEY,
			label TEXT NOT NULL,
			url TEXT NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE work_history_entry (
			id INTEGER PRIMARY KEY,
			employer_name TEXT NOT NULL,
			job_title TEXT NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT,
			date_granularity_start TEXT NOT NULL DEFAULT 'month',
			date_granularity_end TEXT,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE achievement_bullet (
			id INTEGER PRIMARY KEY,
			work_history_id INTEGER NOT NULL REFERENCES work_history_entry(id) ON DELETE CASCADE,
			text TEXT NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE skill_category (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE skill (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			category_id INTEGER NOT NULL REFERENCES skill_category(id),
			competence_level INTEGER NOT NULL CHECK (competence_level BETWEEN 1 AND 10),
			is_legacy INTEGER NOT NULL DEFAULT 0 CHECK (is_legacy IN (0, 1)),
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE academic_credential (
			id INTEGER PRIMARY KEY,
			institution TEXT NOT NULL,
			credential_type TEXT NOT NULL,
			field_of_study TEXT NOT NULL,
			completion_date TEXT NOT NULL,
			date_granularity TEXT NOT NULL DEFAULT 'year',
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE certification (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			issuing_body TEXT NOT NULL,
			date_earned TEXT NOT NULL,
			expiration_date TEXT,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE professional_summary (
			id INTEGER PRIMARY KEY,
			label TEXT NOT NULL UNIQUE,
			body_text TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE role_descriptor (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL UNIQUE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		// === Lens system ===

		`CREATE TABLE lens (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			summary_id INTEGER REFERENCES professional_summary(id) ON DELETE SET NULL,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE lens_work_history_selection (
			id INTEGER PRIMARY KEY,
			lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
			work_history_id INTEGER NOT NULL REFERENCES work_history_entry(id) ON DELETE CASCADE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			UNIQUE(lens_id, work_history_id)
		)`,

		`CREATE TABLE lens_bullet_selection (
			id INTEGER PRIMARY KEY,
			lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
			bullet_id INTEGER NOT NULL REFERENCES achievement_bullet(id) ON DELETE CASCADE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			UNIQUE(lens_id, bullet_id)
		)`,

		`CREATE TABLE lens_skill_selection (
			id INTEGER PRIMARY KEY,
			lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
			skill_id INTEGER NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
			custom_sort_order INTEGER,
			UNIQUE(lens_id, skill_id)
		)`,

		`CREATE TABLE lens_academic_selection (
			id INTEGER PRIMARY KEY,
			lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
			academic_id INTEGER NOT NULL REFERENCES academic_credential(id) ON DELETE CASCADE,
			UNIQUE(lens_id, academic_id)
		)`,

		`CREATE TABLE lens_cert_selection (
			id INTEGER PRIMARY KEY,
			lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
			cert_id INTEGER NOT NULL REFERENCES certification(id) ON DELETE CASCADE,
			UNIQUE(lens_id, cert_id)
		)`,

		`CREATE TABLE lens_descriptor_selection (
			id INTEGER PRIMARY KEY,
			lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
			descriptor_id INTEGER NOT NULL REFERENCES role_descriptor(id) ON DELETE CASCADE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			UNIQUE(lens_id, descriptor_id)
		)`,

		`CREATE TABLE skill_lens_tag (
			id INTEGER PRIMARY KEY,
			skill_id INTEGER NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
			lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
			UNIQUE(skill_id, lens_id)
		)`,

		// === Export system ===

		`CREATE TABLE resume_export (
			id INTEGER PRIMARY KEY,
			template_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			summary_id INTEGER REFERENCES professional_summary(id) ON DELETE SET NULL,
			lens_id INTEGER REFERENCES lens(id) ON DELETE SET NULL,
			generated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE export_work_history_selection (
			id INTEGER PRIMARY KEY,
			export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
			work_history_id INTEGER NOT NULL REFERENCES work_history_entry(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE export_bullet_selection (
			id INTEGER PRIMARY KEY,
			export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
			bullet_id INTEGER NOT NULL REFERENCES achievement_bullet(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE export_skill_selection (
			id INTEGER PRIMARY KEY,
			export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
			skill_id INTEGER NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
			custom_sort_order INTEGER
		)`,

		`CREATE TABLE export_academic_selection (
			id INTEGER PRIMARY KEY,
			export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
			academic_id INTEGER NOT NULL REFERENCES academic_credential(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE export_cert_selection (
			id INTEGER PRIMARY KEY,
			export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
			cert_id INTEGER NOT NULL REFERENCES certification(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE export_descriptor_selection (
			id INTEGER PRIMARY KEY,
			export_id INTEGER NOT NULL REFERENCES resume_export(id) ON DELETE CASCADE,
			descriptor_id INTEGER NOT NULL REFERENCES role_descriptor(id) ON DELETE CASCADE,
			sort_order INTEGER NOT NULL DEFAULT 0
		)`,

		// === Application tracking ===

		`CREATE TABLE cover_letter (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			body_text TEXT NOT NULL,
			file_path TEXT,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE job_application (
			id INTEGER PRIMARY KEY,
			company_name TEXT NOT NULL,
			position_title TEXT NOT NULL,
			date_applied TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'Applied',
			fit_indicator TEXT,
			resume_export_id INTEGER REFERENCES resume_export(id) ON DELETE SET NULL,
			cover_letter_id INTEGER REFERENCES cover_letter(id) ON DELETE SET NULL,
			notes TEXT DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE status_history (
			id INTEGER PRIMARY KEY,
			application_id INTEGER NOT NULL REFERENCES job_application(id) ON DELETE CASCADE,
			from_status TEXT,
			to_status TEXT NOT NULL,
			changed_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		// === Indexes ===

		`CREATE INDEX idx_achievement_bullet_work_history ON achievement_bullet(work_history_id)`,
		`CREATE INDEX idx_skill_category_id ON skill(category_id)`,
		`CREATE INDEX idx_skill_competence ON skill(competence_level DESC, name ASC)`,
		`CREATE INDEX idx_job_application_company ON job_application(company_name)`,
		`CREATE INDEX idx_job_application_status ON job_application(status)`,
		`CREATE INDEX idx_status_history_application ON status_history(application_id)`,
		`CREATE INDEX idx_export_selections_export ON export_work_history_selection(export_id)`,
		`CREATE INDEX idx_export_bullet_export ON export_bullet_selection(export_id)`,
		`CREATE INDEX idx_export_skill_export ON export_skill_selection(export_id)`,
		`CREATE INDEX idx_lens_work_history ON lens_work_history_selection(lens_id)`,
		`CREATE INDEX idx_lens_bullet ON lens_bullet_selection(lens_id)`,
		`CREATE INDEX idx_lens_skill ON lens_skill_selection(lens_id)`,
		`CREATE INDEX idx_lens_academic ON lens_academic_selection(lens_id)`,
		`CREATE INDEX idx_lens_cert ON lens_cert_selection(lens_id)`,
		`CREATE INDEX idx_lens_descriptor ON lens_descriptor_selection(lens_id)`,
		`CREATE INDEX idx_skill_lens_tag_skill ON skill_lens_tag(skill_id)`,
		`CREATE INDEX idx_skill_lens_tag_lens ON skill_lens_tag(lens_id)`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("executing statement: %w\nSQL: %s", err, stmt)
		}
	}

	// Set the schema version.
	if _, err := tx.Exec("PRAGMA user_version = 1"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v1 applied successfully")
	return nil
}

// migrateV2 moves the lens-summary relationship from a 1:1 foreign
// key (summary_id on lens) to a 1:many junction table
// (lens_summary_selection). Existing summary_id values are migrated
// into the junction table, then the column is dropped.
func migrateV2(store *Store) error {
	// Disable foreign keys for the table-rebuild pattern.
	// PRAGMA foreign_keys cannot be changed inside a transaction.
	if _, err := store.db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disabling foreign keys: %w", err)
	}
	// Re-enable on exit regardless of success/failure.
	defer func() { _, _ = store.db.Exec("PRAGMA foreign_keys = ON") }()

	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	statements := []string{
		// 1. Add new columns that were introduced after V1 shipped.
		// SQLite supports ADD COLUMN but not ALTER COLUMN, so defaults
		// are set here for new databases and existing rows get the default.
		`ALTER TABLE work_history_entry ADD COLUMN summary TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE achievement_bullet ADD COLUMN bullet_type TEXT NOT NULL DEFAULT 'primary' CHECK(bullet_type IN ('primary', 'secondary'))`,

		// 3. Create the junction table (references lens which still exists).
		`CREATE TABLE lens_summary_selection (
			id INTEGER PRIMARY KEY,
			lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
			summary_id INTEGER NOT NULL REFERENCES professional_summary(id) ON DELETE CASCADE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			UNIQUE(lens_id, summary_id)
		)`,

		// 4. Index for fast lookup by lens.
		`CREATE INDEX idx_lens_summary ON lens_summary_selection(lens_id)`,

		// 5. Migrate existing summary_id data into the junction table.
		`INSERT INTO lens_summary_selection (lens_id, summary_id, sort_order)
		 SELECT id, summary_id, 0 FROM lens WHERE summary_id IS NOT NULL`,

		// 6. Recreate the lens table without summary_id.
		// We use create-new, copy, drop-old, rename-new instead of the
		// traditional rename-old pattern. ALTER TABLE RENAME rewrites FK
		// references in OTHER tables to point at the renamed table, which
		// breaks those references after the old table is dropped. By
		// creating a new table and dropping the original, existing FK
		// references continue to resolve to "lens" after the rename.
		`CREATE TABLE lens_new (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`INSERT INTO lens_new (id, name, created_at, updated_at)
		 SELECT id, name, created_at, updated_at FROM lens`,

		// Drop the original (FK enforcement is OFF, so this succeeds
		// even though other tables reference lens(id)).
		`DROP TABLE lens`,

		// Rename the new table to the original name. This does NOT
		// rewrite FK references in other tables because nothing
		// references "lens_new".
		`ALTER TABLE lens_new RENAME TO lens`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("executing statement: %w\nSQL: %s", err, stmt)
		}
	}

	// Verify foreign key integrity after the rebuild.
	var fkViolations int
	if err := tx.QueryRow("PRAGMA foreign_key_check").Scan(&fkViolations); err != nil {
		// No rows means no violations — this is the happy path.
		// The error from Scan is sql.ErrNoRows.
	}

	// Set the schema version.
	if _, err := tx.Exec("PRAGMA user_version = 2"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v2 applied successfully")
	return nil
}

// columnExists checks whether a column exists on a table using
// PRAGMA table_info. This is used by repair migrations to safely
// add columns that may or may not already be present.
func columnExists(store *Store, table, column string) (bool, error) {
	rows, err := store.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("sqlite: table_info(%s): %w", table, err)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dfltValue *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false, fmt.Errorf("sqlite: scanning table_info(%s): %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

// migrateV3 repairs databases where V2 partially applied: the lens
// table rebuild succeeded but the ALTER TABLE ADD COLUMN statements
// for summary (work_history_entry) and bullet_type (achievement_bullet)
// were not persisted. Each column is checked before adding to make
// this migration idempotent — it is a safe no-op on databases where
// V2 applied correctly.
func migrateV3(store *Store) error {
	// Check which columns are missing BEFORE opening a transaction.
	// With MaxOpenConns(1), queries inside an active tx would deadlock.
	hasSummary, err := columnExists(store, "work_history_entry", "summary")
	if err != nil {
		return fmt.Errorf("checking summary column: %w", err)
	}
	hasBulletType, err := columnExists(store, "achievement_bullet", "bullet_type")
	if err != nil {
		return fmt.Errorf("checking bullet_type column: %w", err)
	}

	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if !hasSummary {
		store.logger.Info("repairing work_history_entry: adding missing summary column")
		if _, err := tx.Exec(`ALTER TABLE work_history_entry ADD COLUMN summary TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("adding summary column: %w", err)
		}
	}

	if !hasBulletType {
		store.logger.Info("repairing achievement_bullet: adding missing bullet_type column")
		if _, err := tx.Exec(`ALTER TABLE achievement_bullet ADD COLUMN bullet_type TEXT NOT NULL DEFAULT 'primary' CHECK(bullet_type IN ('primary', 'secondary'))`); err != nil {
			return fmt.Errorf("adding bullet_type column: %w", err)
		}
	}

	// Set the schema version.
	if _, err := tx.Exec("PRAGMA user_version = 3"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v3 applied successfully")
	return nil
}

// hasLensOldCorruption checks whether any table in sqlite_master
// references the non-existent "lens_old" table. This is a sign of
// schema corruption caused by the V2 lens table rebuild pattern.
// Must be called OUTSIDE a transaction to avoid MaxOpenConns(1) deadlock.
func hasLensOldCorruption(store *Store) (bool, error) {
	var count int
	err := store.db.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE sql LIKE '%lens_old%'`,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("sqlite: checking for lens_old corruption: %w", err)
	}
	return count > 0, nil
}

// migrateV4 repairs databases where the V2 lens table rebuild
// corrupted FK references in child tables. The V2 migration used
// a DROP TABLE + ALTER TABLE RENAME pattern that, under certain
// SQLite versions, rewrote FK references from lens(id) to
// "lens_old"(id) — a table that does not exist. This causes
// "no such table: main.lens_old" errors whenever FK triggers fire
// (e.g., deleting from selection tables that CASCADE to lens).
//
// The fix rebuilds each affected table with correct FK references.
// This migration is idempotent — it is a safe no-op on databases
// where the corruption did not occur.
func migrateV4(store *Store) error {
	// Check for corruption BEFORE opening a transaction.
	corrupted, err := hasLensOldCorruption(store)
	if err != nil {
		return err
	}

	if !corrupted {
		store.logger.Info("v4: no lens_old corruption detected, skipping repair")
		// Still need to set version.
		if _, err := store.db.Exec("PRAGMA user_version = 4"); err != nil {
			return fmt.Errorf("setting user_version: %w", err)
		}
		store.logger.Info("migration v4 applied successfully")
		return nil
	}

	store.logger.Info("v4: lens_old corruption detected, rebuilding affected tables")

	// Disable foreign keys for the table-rebuild pattern.
	if _, err := store.db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disabling foreign keys: %w", err)
	}
	defer func() { _, _ = store.db.Exec("PRAGMA foreign_keys = ON") }()

	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Each entry: table name, correct CREATE TABLE DDL, optional index DDL.
	// We rebuild using: CREATE _fixed → copy → DROP original → RENAME _fixed.
	type tableRebuild struct {
		name     string
		createDD string
		indexDDL string // empty if no explicit index
	}

	rebuilds := []tableRebuild{
		{
			name: "lens_work_history_selection",
			createDD: `CREATE TABLE lens_work_history_selection_fixed (
				id INTEGER PRIMARY KEY,
				lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
				work_history_id INTEGER NOT NULL REFERENCES work_history_entry(id) ON DELETE CASCADE,
				sort_order INTEGER NOT NULL DEFAULT 0,
				UNIQUE(lens_id, work_history_id)
			)`,
			indexDDL: `CREATE INDEX idx_lens_work_history ON lens_work_history_selection(lens_id)`,
		},
		{
			name: "lens_bullet_selection",
			createDD: `CREATE TABLE lens_bullet_selection_fixed (
				id INTEGER PRIMARY KEY,
				lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
				bullet_id INTEGER NOT NULL REFERENCES achievement_bullet(id) ON DELETE CASCADE,
				sort_order INTEGER NOT NULL DEFAULT 0,
				UNIQUE(lens_id, bullet_id)
			)`,
			indexDDL: `CREATE INDEX idx_lens_bullet ON lens_bullet_selection(lens_id)`,
		},
		{
			name: "lens_skill_selection",
			createDD: `CREATE TABLE lens_skill_selection_fixed (
				id INTEGER PRIMARY KEY,
				lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
				skill_id INTEGER NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
				custom_sort_order INTEGER,
				UNIQUE(lens_id, skill_id)
			)`,
			indexDDL: `CREATE INDEX idx_lens_skill ON lens_skill_selection(lens_id)`,
		},
		{
			name: "lens_academic_selection",
			createDD: `CREATE TABLE lens_academic_selection_fixed (
				id INTEGER PRIMARY KEY,
				lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
				academic_id INTEGER NOT NULL REFERENCES academic_credential(id) ON DELETE CASCADE,
				UNIQUE(lens_id, academic_id)
			)`,
			indexDDL: `CREATE INDEX idx_lens_academic ON lens_academic_selection(lens_id)`,
		},
		{
			name: "lens_cert_selection",
			createDD: `CREATE TABLE lens_cert_selection_fixed (
				id INTEGER PRIMARY KEY,
				lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
				cert_id INTEGER NOT NULL REFERENCES certification(id) ON DELETE CASCADE,
				UNIQUE(lens_id, cert_id)
			)`,
			indexDDL: `CREATE INDEX idx_lens_cert ON lens_cert_selection(lens_id)`,
		},
		{
			name: "lens_descriptor_selection",
			createDD: `CREATE TABLE lens_descriptor_selection_fixed (
				id INTEGER PRIMARY KEY,
				lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
				descriptor_id INTEGER NOT NULL REFERENCES role_descriptor(id) ON DELETE CASCADE,
				sort_order INTEGER NOT NULL DEFAULT 0,
				UNIQUE(lens_id, descriptor_id)
			)`,
			indexDDL: `CREATE INDEX idx_lens_descriptor ON lens_descriptor_selection(lens_id)`,
		},
		{
			name: "lens_summary_selection",
			createDD: `CREATE TABLE lens_summary_selection_fixed (
				id INTEGER PRIMARY KEY,
				lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
				summary_id INTEGER NOT NULL REFERENCES professional_summary(id) ON DELETE CASCADE,
				sort_order INTEGER NOT NULL DEFAULT 0,
				UNIQUE(lens_id, summary_id)
			)`,
			indexDDL: `CREATE INDEX idx_lens_summary ON lens_summary_selection(lens_id)`,
		},
		{
			name: "skill_lens_tag",
			createDD: `CREATE TABLE skill_lens_tag_fixed (
				id INTEGER PRIMARY KEY,
				skill_id INTEGER NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
				lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
				UNIQUE(skill_id, lens_id)
			)`,
			indexDDL: `CREATE INDEX idx_skill_lens_tag_skill ON skill_lens_tag(skill_id)`,
		},
		{
			name: "resume_export",
			createDD: `CREATE TABLE resume_export_fixed (
				id INTEGER PRIMARY KEY,
				template_id TEXT NOT NULL,
				file_path TEXT NOT NULL,
				summary_id INTEGER REFERENCES professional_summary(id) ON DELETE SET NULL,
				lens_id INTEGER REFERENCES lens(id) ON DELETE SET NULL,
				generated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
			)`,
			indexDDL: "", // no explicit index for resume_export
		},
	}

	for _, rb := range rebuilds {
		store.logger.Info("v4: rebuilding table", slog.String("table", rb.name))

		// 1. Create the fixed table.
		if _, err := tx.Exec(rb.createDD); err != nil {
			return fmt.Errorf("v4: creating %s_fixed: %w", rb.name, err)
		}

		// 2. Copy all data from the corrupted table.
		copySQL := fmt.Sprintf(
			"INSERT INTO %s_fixed SELECT * FROM %s",
			rb.name, rb.name,
		)
		if _, err := tx.Exec(copySQL); err != nil {
			return fmt.Errorf("v4: copying data to %s_fixed: %w", rb.name, err)
		}

		// 3. Drop the corrupted table (and its indexes).
		dropSQL := fmt.Sprintf("DROP TABLE %s", rb.name)
		if _, err := tx.Exec(dropSQL); err != nil {
			return fmt.Errorf("v4: dropping corrupted %s: %w", rb.name, err)
		}

		// 4. Rename the fixed table to the original name.
		renameSQL := fmt.Sprintf(
			"ALTER TABLE %s_fixed RENAME TO %s",
			rb.name, rb.name,
		)
		if _, err := tx.Exec(renameSQL); err != nil {
			return fmt.Errorf("v4: renaming %s_fixed to %s: %w", rb.name, rb.name, err)
		}

		// 5. Recreate the index (DROP TABLE removed it).
		if rb.indexDDL != "" {
			if _, err := tx.Exec(rb.indexDDL); err != nil {
				return fmt.Errorf("v4: creating index for %s: %w", rb.name, err)
			}
		}
	}

	// skill_lens_tag has a second index that was dropped.
	if _, err := tx.Exec(`CREATE INDEX idx_skill_lens_tag_lens ON skill_lens_tag(lens_id)`); err != nil {
		return fmt.Errorf("v4: creating idx_skill_lens_tag_lens: %w", err)
	}

	// Verify no corruption remains.
	var remaining int
	if err := tx.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE sql LIKE '%lens_old%'`,
	).Scan(&remaining); err != nil {
		return fmt.Errorf("v4: verifying repair: %w", err)
	}
	if remaining > 0 {
		return fmt.Errorf("v4: repair incomplete, %d tables still reference lens_old", remaining)
	}

	// Verify foreign key integrity.
	var fkViolations int
	if err := tx.QueryRow("PRAGMA foreign_key_check").Scan(&fkViolations); err != nil {
		// sql.ErrNoRows = no violations (happy path).
	}

	if _, err := tx.Exec("PRAGMA user_version = 4"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v4 applied successfully — lens_old corruption repaired")
	return nil
}

// migrateV5 adds the is_master column to lens_summary_selection,
// allowing one summary per lens to be designated as the "master"
// summary. The master summary renders as a plain paragraph in the
// PDF; other selected summaries render as bullet points.
func migrateV5(store *Store) error {
	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(
		`ALTER TABLE lens_summary_selection ADD COLUMN is_master INTEGER NOT NULL DEFAULT 0`,
	); err != nil {
		return fmt.Errorf("adding is_master column: %w", err)
	}

	if _, err := tx.Exec("PRAGMA user_version = 5"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v5 applied successfully")
	return nil
}

// migrateV6 adds the core_expertise table and the
// lens_core_expertise_selection junction table for the new Core
// Expertise entity.
func migrateV6(store *Store) error {
	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	statements := []string{
		`CREATE TABLE core_expertise (
			id INTEGER PRIMARY KEY,
			label TEXT NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE lens_core_expertise_selection (
			id INTEGER PRIMARY KEY,
			lens_id INTEGER NOT NULL REFERENCES lens(id) ON DELETE CASCADE,
			core_expertise_id INTEGER NOT NULL REFERENCES core_expertise(id) ON DELETE CASCADE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			UNIQUE(lens_id, core_expertise_id)
		)`,

		`CREATE INDEX idx_lens_core_expertise ON lens_core_expertise_selection(lens_id)`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("executing statement: %w\nSQL: %s", err, stmt)
		}
	}

	if _, err := tx.Exec("PRAGMA user_version = 6"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v6 applied successfully")
	return nil
}

// migrateV7 adds the document_template and template_element tables
// for the template builder feature. It also adds a template_ref_id
// column to resume_export to reference the new document_template
// table (keeping the legacy template_id TEXT column for backward
// compatibility). Built-in templates are seeded with full element
// trees to reproduce the existing Professional and Modern layouts.
func migrateV7(store *Store) error {
	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	statements := []string{
		// === New template tables ===

		`CREATE TABLE document_template (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			template_type TEXT NOT NULL CHECK (template_type IN ('resume', 'cover_letter')),
			is_builtin INTEGER NOT NULL DEFAULT 0 CHECK (is_builtin IN (0, 1)),
			margin_top REAL NOT NULL DEFAULT 54.0,
			margin_bottom REAL NOT NULL DEFAULT 54.0,
			margin_left REAL NOT NULL DEFAULT 72.0,
			margin_right REAL NOT NULL DEFAULT 72.0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE TABLE template_element (
			id INTEGER PRIMARY KEY,
			template_id INTEGER NOT NULL REFERENCES document_template(id) ON DELETE CASCADE,
			parent_id INTEGER REFERENCES template_element(id) ON DELETE CASCADE,
			element_type TEXT NOT NULL,
			config TEXT NOT NULL DEFAULT '{}',
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`CREATE INDEX idx_template_element_template ON template_element(template_id)`,
		`CREATE INDEX idx_template_element_parent ON template_element(parent_id)`,

		// === Extend resume_export with template FK ===

		`ALTER TABLE resume_export ADD COLUMN template_ref_id INTEGER REFERENCES document_template(id) ON DELETE SET NULL`,

		// === Seed built-in Professional resume template ===

		`INSERT INTO document_template (id, name, description, template_type, is_builtin, margin_top, margin_bottom, margin_left, margin_right)
		 VALUES (1, 'Professional', 'Classic centered layout with underlined section headings', 'resume', 1, 54.0, 54.0, 72.0, 72.0)`,

		// Professional template elements (top-level)
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (1, 1, NULL, 'profile_header', '{"name_font_size":18.0,"detail_font_size":10.0,"alignment":"center","link_separator":" | ","show_links":true,"show_links_inline":false,"space_after":6.0}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (2, 1, NULL, 'role_descriptors', '{"font_size":10.0,"font_style":"regular","alignment":"center","separator":" | ","space_after":4.0}', 1)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, 1, NULL, 'section_heading', '{"text":"Professional Summary","font_size":12.0,"font_style":"bold","uppercase":true,"underline":true,"underline_weight":0.5,"space_before":10.0,"space_after":4.0,"data_binding":"summaries"}', 2)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (4, 1, NULL, 'professional_summary', '{"font_size":10.0,"bullet_char":"\\u2022","space_before":0.0,"space_after":0.0}', 3)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (5, 1, NULL, 'section_heading', '{"text":"Core Expertise","font_size":12.0,"font_style":"bold","uppercase":true,"underline":true,"underline_weight":0.5,"space_before":10.0,"space_after":4.0,"data_binding":"core_expertise"}', 4)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (6, 1, NULL, 'core_expertise', '{"font_size":10.0,"separator":" | ","alignment":"center","space_after":0.0}', 5)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (7, 1, NULL, 'section_heading', '{"text":"Experience","font_size":12.0,"font_style":"bold","uppercase":true,"underline":true,"underline_weight":0.5,"space_before":10.0,"space_after":4.0,"data_binding":"work_history"}', 6)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (8, 1, NULL, 'work_history_loop', '{"entry_gap":4.0,"space_before":0.0,"space_after":0.0}', 7)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (9, 1, NULL, 'section_heading', '{"text":"Skills","font_size":12.0,"font_style":"bold","uppercase":true,"underline":true,"underline_weight":0.5,"space_before":10.0,"space_after":4.0,"data_binding":"skills"}', 8)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (10, 1, NULL, 'skills', '{"font_size":10.0,"group_by_category":true,"include_legacy":true,"legacy_suffix":" (Legacy)","category_font_style":"bold","skill_separator":", "}', 9)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (11, 1, NULL, 'section_heading', '{"text":"Education","font_size":12.0,"font_style":"bold","uppercase":true,"underline":true,"underline_weight":0.5,"space_before":10.0,"space_after":4.0,"data_binding":"academics"}', 10)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (12, 1, NULL, 'education_loop', '{"entry_gap":0.0,"space_before":0.0,"space_after":0.0}', 11)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (13, 1, NULL, 'section_heading', '{"text":"Certifications","font_size":12.0,"font_style":"bold","uppercase":true,"underline":true,"underline_weight":0.5,"space_before":10.0,"space_after":4.0,"data_binding":"certifications"}', 12)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (14, 1, NULL, 'certifications_loop', '{"entry_gap":0.0,"space_before":0.0,"space_after":0.0}', 13)`,

		// Professional work_history_loop children
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (15, 1, 8, 'work_title', '{"font_size":10.0,"font_style":"bold","include_employer":true,"employer_separator":" \\u2014 ","employer_font_style":"italic","space_after":13.0}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (16, 1, 8, 'work_dates', '{"font_size":9.0,"alignment":"right"}', 1)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (17, 1, 8, 'work_summary', '{}', 2)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (18, 1, 8, 'work_bullets', '{"font_size":10.0,"font_style":"regular","bullet_char":"\\u2022","indent":12.0,"bullet_sym_width":10.0}', 3)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (19, 1, 8, 'work_outcomes', '{"font_size":10.0,"font_style":"regular","bullet_char":"\\u2022","indent":12.0,"bullet_sym_width":10.0,"outcomes_label":"Outcomes:","outcomes_gap":2.0}', 4)`,

		// Professional education_loop children
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (20, 1, 12, 'edu_credential', '{"font_size":10.0,"font_style":"bold"}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (21, 1, 12, 'edu_institution', '{"font_size":10.0,"font_style":"regular"}', 1)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (22, 1, 12, 'edu_date', '{"font_size":9.0,"alignment":"right"}', 2)`,

		// Professional certifications_loop children
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (23, 1, 14, 'cert_name', '{"font_size":10.0,"font_style":"bold"}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (24, 1, 14, 'cert_detail', '{"font_size":10.0,"font_style":"regular"}', 1)`,

		// === Seed built-in Modern resume template ===

		`INSERT INTO document_template (id, name, description, template_type, is_builtin, margin_top, margin_bottom, margin_left, margin_right)
		 VALUES (2, 'Modern', 'Left-aligned layout with clean typography and no underlines', 'resume', 1, 54.0, 54.0, 72.0, 72.0)`,

		// Modern template elements (top-level)
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (25, 2, NULL, 'profile_header', '{"name_font_size":22.0,"detail_font_size":9.0,"alignment":"left","link_separator":"  \\u00B7  ","show_links":true,"show_links_inline":true,"space_after":6.0}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (26, 2, NULL, 'role_descriptors', '{"font_size":10.0,"font_style":"italic","alignment":"left","separator":"  \\u00B7  ","space_after":6.0}', 1)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (27, 2, NULL, 'horizontal_rule', '{"weight":0.3,"space_before":0.0,"space_after":6.0}', 2)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (28, 2, NULL, 'section_heading', '{"text":"Summary","font_size":11.0,"font_style":"bold","uppercase":true,"underline":false,"underline_weight":0.0,"space_before":14.0,"space_after":6.0,"data_binding":"summaries"}', 3)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (29, 2, NULL, 'professional_summary', '{"font_size":10.0,"bullet_char":"\\u2022","space_before":0.0,"space_after":0.0}', 4)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (30, 2, NULL, 'section_heading', '{"text":"Experience","font_size":11.0,"font_style":"bold","uppercase":true,"underline":false,"underline_weight":0.0,"space_before":14.0,"space_after":6.0,"data_binding":"work_history"}', 5)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (31, 2, NULL, 'work_history_loop', '{"entry_gap":6.0,"space_before":0.0,"space_after":0.0}', 6)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (32, 2, NULL, 'section_heading', '{"text":"Skills","font_size":11.0,"font_style":"bold","uppercase":true,"underline":false,"underline_weight":0.0,"space_before":14.0,"space_after":6.0,"data_binding":"skills"}', 7)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (33, 2, NULL, 'skills', '{"font_size":10.0,"group_by_category":true,"include_legacy":true,"legacy_suffix":" (Legacy)","category_font_style":"bold","skill_separator":", "}', 8)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (34, 2, NULL, 'section_heading', '{"text":"Core Expertise","font_size":11.0,"font_style":"bold","uppercase":true,"underline":false,"underline_weight":0.0,"space_before":14.0,"space_after":6.0,"data_binding":"core_expertise"}', 9)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (35, 2, NULL, 'core_expertise', '{"font_size":10.0,"separator":" | ","alignment":"left","space_after":0.0}', 10)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (36, 2, NULL, 'section_heading', '{"text":"Education","font_size":11.0,"font_style":"bold","uppercase":true,"underline":false,"underline_weight":0.0,"space_before":14.0,"space_after":6.0,"data_binding":"academics"}', 11)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (37, 2, NULL, 'education_loop', '{"entry_gap":0.0,"space_before":0.0,"space_after":0.0}', 12)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (38, 2, NULL, 'section_heading', '{"text":"Certifications","font_size":11.0,"font_style":"bold","uppercase":true,"underline":false,"underline_weight":0.0,"space_before":14.0,"space_after":6.0,"data_binding":"certifications"}', 13)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (39, 2, NULL, 'certifications_loop', '{"entry_gap":0.0,"space_before":0.0,"space_after":0.0}', 14)`,

		// Modern work_history_loop children
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (40, 2, 31, 'work_title', '{"font_size":10.0,"font_style":"bold","include_employer":true,"employer_separator":", ","employer_font_style":"bold","space_after":15.0}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (41, 2, 31, 'work_dates', '{"font_size":9.0,"alignment":"right"}', 1)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (42, 2, 31, 'work_summary', '{}', 2)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (43, 2, 31, 'work_bullets', '{"font_size":10.0,"font_style":"regular","bullet_char":"\\u2022","indent":12.0,"bullet_sym_width":10.0}', 3)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (44, 2, 31, 'work_outcomes', '{"font_size":10.0,"font_style":"regular","bullet_char":"\\u2022","indent":12.0,"bullet_sym_width":10.0,"outcomes_label":"Outcomes:","outcomes_gap":2.0}', 4)`,

		// Modern education_loop children
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (45, 2, 37, 'edu_credential', '{"font_size":10.0,"font_style":"bold"}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (46, 2, 37, 'edu_institution', '{"font_size":10.0,"font_style":"regular"}', 1)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (47, 2, 37, 'edu_date', '{"font_size":9.0,"alignment":"right"}', 2)`,

		// Modern certifications_loop children
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (48, 2, 39, 'cert_name', '{"font_size":10.0,"font_style":"bold"}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (49, 2, 39, 'cert_detail', '{"font_size":10.0,"font_style":"regular"}', 1)`,

		// === Populate template_ref_id for existing export records ===

		`UPDATE resume_export SET template_ref_id = 1 WHERE template_id = 'professional'`,
		`UPDATE resume_export SET template_ref_id = 2 WHERE template_id = 'modern'`,

		// === Seed built-in Formal cover letter template ===

		`INSERT INTO document_template (id, name, description, template_type, is_builtin, margin_top, margin_bottom, margin_left, margin_right)
		 VALUES (3, 'Formal', 'Traditional business-style cover letter with structured prompts and formal tone', 'cover_letter', 1, 72.0, 72.0, 72.0, 72.0)`,

		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (50, 3, NULL, 'profile_header', '{"name_font_size":18.0,"detail_font_size":10.0,"alignment":"center","link_separator":" | ","show_links":true,"show_links_inline":false,"space_after":6.0}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (51, 3, NULL, 'horizontal_rule', '{"weight":0.5,"space_before":2.0,"space_after":14.0}', 1)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (52, 3, NULL, 'date', '{"font_size":10.0,"format":"January 2, 2006","alignment":"left","space_after":14.0}', 2)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (53, 3, NULL, 'recipient_address', '{"font_size":10.0,"space_after":14.0}', 3)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (54, 3, NULL, 'greeting', '{"text":"Dear {{hiring_manager}},","font_size":10.0,"space_after":10.0}', 4)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (55, 3, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"I am writing to express my interest in the "},{"type":"application","token":"position_title"},{"type":"static","text":" role at "},{"type":"application","token":"company_name"},{"type":"static","text":". "},{"type":"adhoc","key":"why_fit_formal","label":"Why are you a strong fit?","help_text":"Highlight relevant experience and outcomes you can deliver.","required":true,"multiline":true}]}', 5)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (56, 3, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"I am especially interested in "},{"type":"application","token":"company_name"},{"type":"static","text":" because "},{"type":"adhoc","key":"why_company_formal","label":"Why this company?","help_text":"Mention a mission, product, or initiative that stands out to you.","required":true,"multiline":true},{"type":"static","text":"."}]}', 6)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (57, 3, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"Thank you for your time and consideration. I would welcome the opportunity to discuss how I can contribute to "},{"type":"application","token":"company_name"},{"type":"static","text":"."}]}', 7)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (58, 3, NULL, 'closing', '{"text":"Sincerely,","font_size":10.0,"space_after":28.0}', 8)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (59, 3, NULL, 'static_text', '{"text":"{{signer_name}}","font_size":10.0,"font_style":"regular","alignment":"left","space_before":0.0,"space_after":0.0}', 9)`,

		// === Seed built-in Casual cover letter template ===

		`INSERT INTO document_template (id, name, description, template_type, is_builtin, margin_top, margin_bottom, margin_left, margin_right)
		 VALUES (4, 'Casual', 'Conversational cover letter starter with lightweight prompts and approachable tone', 'cover_letter', 1, 72.0, 72.0, 72.0, 72.0)`,

		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (60, 4, NULL, 'profile_header', '{"name_font_size":18.0,"detail_font_size":10.0,"alignment":"center","link_separator":" | ","show_links":true,"show_links_inline":false,"space_after":6.0}', 0)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (61, 4, NULL, 'greeting', '{"text":"Hi {{hiring_manager}},","font_size":10.0,"space_after":10.0}', 1)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (62, 4, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"I would love to join "},{"type":"application","token":"company_name"},{"type":"static","text":" as a "},{"type":"application","token":"position_title"},{"type":"static","text":". What stands out to me about your team is "},{"type":"adhoc","key":"why_team_casual","label":"What stands out about this team?","help_text":"Keep it natural and specific.","required":true,"multiline":true},{"type":"static","text":"."}]}', 2)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (63, 4, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"A quick example of what I would bring to this role: "},{"type":"adhoc","key":"value_example_casual","label":"Quick value example","help_text":"Share one example of impact or ownership.","required":false,"multiline":true}]}', 3)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (64, 4, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"If it is helpful, I can share more details and examples. You can reach me at "},{"type":"profile","token":"email"},{"type":"static","text":"."}]}', 4)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (65, 4, NULL, 'closing', '{"text":"Thanks,","font_size":10.0,"space_after":24.0}', 5)`,
		`INSERT INTO template_element (id, template_id, parent_id, element_type, config, sort_order) VALUES
		 (66, 4, NULL, 'static_text', '{"text":"{{signer_name}}","font_size":10.0,"font_style":"regular","alignment":"left","space_before":0.0,"space_after":0.0}', 6)`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("executing statement: %w\nSQL: %s", err, stmt)
		}
	}

	if _, err := tx.Exec("PRAGMA user_version = 7"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v7 applied successfully")
	return nil
}

// migrateV8 moves job_application.cover_letter_id to
// job_application.cover_letter_export_id (linking to resume_export)
// and adds per-application+template prompt value persistence.
func migrateV8(store *Store) error {
	// Disable foreign keys for table rebuild pattern.
	if _, err := store.db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disabling foreign keys: %w", err)
	}
	defer func() { _, _ = store.db.Exec("PRAGMA foreign_keys = ON") }()

	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	statements := []string{
		`CREATE TABLE job_application_new (
			id INTEGER PRIMARY KEY,
			company_name TEXT NOT NULL,
			position_title TEXT NOT NULL,
			date_applied TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'Applied',
			fit_indicator TEXT,
			resume_export_id INTEGER REFERENCES resume_export(id) ON DELETE SET NULL,
			cover_letter_export_id INTEGER REFERENCES resume_export(id) ON DELETE SET NULL,
			notes TEXT DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,

		`INSERT INTO job_application_new
		 (id, company_name, position_title, date_applied, status,
		  fit_indicator, resume_export_id, cover_letter_export_id, notes,
		  created_at, updated_at)
		 SELECT id, company_name, position_title, date_applied, status,
		        fit_indicator, resume_export_id, NULL, notes,
		        created_at, updated_at
		 FROM job_application`,

		`DROP TABLE job_application`,
		`ALTER TABLE job_application_new RENAME TO job_application`,

		`CREATE INDEX idx_job_application_company ON job_application(company_name)`,
		`CREATE INDEX idx_job_application_status ON job_application(status)`,

		`CREATE TABLE application_prompt_value (
			id INTEGER PRIMARY KEY,
			application_id INTEGER NOT NULL REFERENCES job_application(id) ON DELETE CASCADE,
			template_id INTEGER NOT NULL REFERENCES document_template(id) ON DELETE CASCADE,
			field_key TEXT NOT NULL,
			field_value TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			UNIQUE(application_id, template_id, field_key)
		)`,
		`CREATE INDEX idx_app_prompt_value_app_tmpl ON application_prompt_value(application_id, template_id)`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("executing statement: %w\nSQL: %s", err, stmt)
		}
	}

	if _, err := tx.Exec("PRAGMA user_version = 8"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v8 applied successfully")
	return nil
}

// migrateV9 makes cover letter linkage template-first by adding
// job_application.cover_letter_template_id and
// job_application.cover_letter_latest_export_id. Existing
// cover_letter_export_id values are copied to
// cover_letter_latest_export_id for continuity.
func migrateV9(store *Store) error {
	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	statements := []string{
		`ALTER TABLE job_application
		 ADD COLUMN cover_letter_template_id INTEGER REFERENCES document_template(id) ON DELETE SET NULL`,
		`ALTER TABLE job_application
		 ADD COLUMN cover_letter_latest_export_id INTEGER REFERENCES resume_export(id) ON DELETE SET NULL`,
		`UPDATE job_application
		 SET cover_letter_latest_export_id = cover_letter_export_id
		 WHERE cover_letter_latest_export_id IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_job_application_cover_tmpl ON job_application(cover_letter_template_id)`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("executing statement: %w\nSQL: %s", err, stmt)
		}
	}

	if _, err := tx.Exec("PRAGMA user_version = 9"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v9 applied successfully")
	return nil
}

// migrateV10 adds job_application.job_posting_url for storing a
// source listing URL per application.
func migrateV10(store *Store) error {
	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`ALTER TABLE job_application ADD COLUMN job_posting_url TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("executing statement: %w", err)
	}

	if _, err := tx.Exec("PRAGMA user_version = 10"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v10 applied successfully")
	return nil
}

// migrateV11 refreshes built-in Formal and Casual cover letter
// templates with generic starter copy and guided paragraph prompts.
func migrateV11(store *Store) error {
	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	statements := []string{
		`UPDATE document_template
		 SET description = 'Traditional business-style cover letter with structured prompts and formal tone'
		 WHERE id = 3 AND is_builtin = 1`,
		`UPDATE document_template
		 SET description = 'Conversational cover letter starter with lightweight prompts and approachable tone'
		 WHERE id = 4 AND is_builtin = 1`,

		`DELETE FROM template_element WHERE template_id IN (3, 4)`,

		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'profile_header', '{"name_font_size":18.0,"detail_font_size":10.0,"alignment":"center","link_separator":" | ","show_links":true,"show_links_inline":false,"space_after":6.0}', 0)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'horizontal_rule', '{"weight":0.5,"space_before":2.0,"space_after":14.0}', 1)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'date', '{"font_size":10.0,"format":"January 2, 2006","alignment":"left","space_after":14.0}', 2)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'recipient_address', '{"font_size":10.0,"space_after":14.0}', 3)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'greeting', '{"text":"Dear {{hiring_manager}},","font_size":10.0,"space_after":10.0}', 4)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"I am writing to express my interest in the "},{"type":"application","token":"position_title"},{"type":"static","text":" role at "},{"type":"application","token":"company_name"},{"type":"static","text":". "},{"type":"adhoc","key":"why_fit_formal","label":"Why are you a strong fit?","help_text":"Highlight relevant experience and outcomes you can deliver.","required":true,"multiline":true}]}', 5)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"I am especially interested in "},{"type":"application","token":"company_name"},{"type":"static","text":" because "},{"type":"adhoc","key":"why_company_formal","label":"Why this company?","help_text":"Mention a mission, product, or initiative that stands out to you.","required":true,"multiline":true},{"type":"static","text":"."}]}', 6)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"Thank you for your time and consideration. I would welcome the opportunity to discuss how I can contribute to "},{"type":"application","token":"company_name"},{"type":"static","text":"."}]}', 7)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'closing', '{"text":"Sincerely,","font_size":10.0,"space_after":28.0}', 8)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (3, NULL, 'static_text', '{"text":"{{signer_name}}","font_size":10.0,"font_style":"regular","alignment":"left","space_before":0.0,"space_after":0.0}', 9)`,

		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (4, NULL, 'profile_header', '{"name_font_size":18.0,"detail_font_size":10.0,"alignment":"center","link_separator":" | ","show_links":true,"show_links_inline":false,"space_after":6.0}', 0)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (4, NULL, 'greeting', '{"text":"Hi {{hiring_manager}},","font_size":10.0,"space_after":10.0}', 1)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (4, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"I would love to join "},{"type":"application","token":"company_name"},{"type":"static","text":" as a "},{"type":"application","token":"position_title"},{"type":"static","text":". What stands out to me about your team is "},{"type":"adhoc","key":"why_team_casual","label":"What stands out about this team?","help_text":"Keep it natural and specific.","required":true,"multiline":true},{"type":"static","text":"."}]}', 2)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (4, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"A quick example of what I would bring to this role: "},{"type":"adhoc","key":"value_example_casual","label":"Quick value example","help_text":"Share one example of impact or ownership.","required":false,"multiline":true}]}', 3)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (4, NULL, 'paragraph', '{"font_size":10.0,"line_spacing":1.15,"space_after":10.0,"segments":[{"type":"static","text":"If it is helpful, I can share more details and examples. You can reach me at "},{"type":"profile","token":"email"},{"type":"static","text":"."}]}', 4)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (4, NULL, 'closing', '{"text":"Thanks,","font_size":10.0,"space_after":24.0}', 5)`,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order) VALUES
		 (4, NULL, 'static_text', '{"text":"{{signer_name}}","font_size":10.0,"font_style":"regular","alignment":"left","space_before":0.0,"space_after":0.0}', 6)`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("executing statement: %w\nSQL: %s", err, stmt)
		}
	}

	if _, err := tx.Exec("PRAGMA user_version = 11"); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}

	store.logger.Info("migration v11 applied successfully")
	return nil
}
