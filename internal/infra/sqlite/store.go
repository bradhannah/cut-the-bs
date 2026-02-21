package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite" // Pure-Go SQLite driver
)

// Store is the SQLite-backed persistence layer for all cut-the-bs
// entities. It wraps a single database/sql.DB connection configured
// with WAL mode, foreign keys, and a busy timeout.
type Store struct {
	db     *sql.DB
	logger *slog.Logger
	dbPath string
}

// NewStore opens or creates a SQLite database at the given path and
// configures it with the required PRAGMAs:
//   - journal_mode = WAL (write-ahead logging for concurrent reads)
//   - busy_timeout = 5000 (5 second busy timeout)
//   - foreign_keys = ON (enforce FK constraints)
//
// The caller must call Close() when done.
func NewStore(dbPath string, logger *slog.Logger) (*Store, error) {
	if logger == nil {
		logger = slog.Default()
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("sqlite: unable to open database at %s: %w", dbPath, err)
	}

	// SQLite with modernc.org/sqlite works best with a single connection
	// to avoid locking issues.
	db.SetMaxOpenConns(1)

	// Apply required PRAGMAs.
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = ON",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("sqlite: failed to set %s: %w", pragma, err)
		}
	}

	logger.Info("database opened", slog.String("path", dbPath))

	return &Store{
		db:     db,
		logger: logger,
		dbPath: dbPath,
	}, nil
}

// DBPath returns the path to the underlying database file.
func (s *Store) DBPath() string {
	return s.dbPath
}

// DB returns the underlying *sql.DB. This is primarily for use by
// migration and query code within the sqlite package.
func (s *Store) DB() *sql.DB {
	return s.db
}

// Close closes the database connection. It checkpoints the WAL
// before closing to ensure all data is written to the main database
// file.
func (s *Store) Close() error {
	// Checkpoint the WAL before closing.
	if _, err := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		s.logger.Warn("WAL checkpoint failed during close", slog.String("error", err.Error()))
	}

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("sqlite: unable to close database: %w", err)
	}

	s.logger.Info("database closed")
	return nil
}

// Checkpoint performs a WAL checkpoint without closing the database.
// This is called during the OnBeforeClose lifecycle hook to ensure
// data durability.
func (s *Store) Checkpoint() error {
	_, err := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	if err != nil {
		return fmt.Errorf("sqlite: WAL checkpoint failed: %w", err)
	}
	s.logger.Debug("WAL checkpoint completed")
	return nil
}
