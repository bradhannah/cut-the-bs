package sqlite

import (
	"fmt"
	"log/slog"
)

// currentVersion is the latest schema version.
const currentVersion = 1

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
