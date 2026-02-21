package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Cover Letters ---

// ListCoverLetters returns all cover letters ordered by ID.
func (s *Store) ListCoverLetters(ctx context.Context) ([]domain.CoverLetter, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, body_text, COALESCE(file_path, ''),
		        created_at, updated_at
		 FROM cover_letter
		 ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("cover_letter: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	letters := make([]domain.CoverLetter, 0)
	for rows.Next() {
		var cl domain.CoverLetter
		if err := rows.Scan(
			&cl.ID, &cl.Title, &cl.BodyText, &cl.FilePath,
			&cl.CreatedAt, &cl.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cover_letter: scan: %w", err)
		}
		letters = append(letters, cl)
	}

	return letters, rows.Err()
}

// GetCoverLetter returns a single cover letter by ID.
func (s *Store) GetCoverLetter(ctx context.Context, id int64) (domain.CoverLetter, error) {
	var cl domain.CoverLetter
	var filePath sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, body_text, file_path,
		        created_at, updated_at
		 FROM cover_letter WHERE id = ?`, id,
	).Scan(
		&cl.ID, &cl.Title, &cl.BodyText, &filePath,
		&cl.CreatedAt, &cl.UpdatedAt,
	)
	if err != nil {
		return domain.CoverLetter{}, fmt.Errorf("cover_letter: get %d: %w", id, err)
	}
	if filePath.Valid {
		cl.FilePath = filePath.String
	}
	return cl, nil
}

// CreateCoverLetter creates a new cover letter.
func (s *Store) CreateCoverLetter(ctx context.Context, input domain.CoverLetterInput) (domain.CoverLetter, error) {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO cover_letter (title, body_text) VALUES (?, ?)`,
		input.Title, input.BodyText,
	)
	if err != nil {
		return domain.CoverLetter{}, fmt.Errorf("cover_letter: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.CoverLetter{}, fmt.Errorf("cover_letter: last insert id: %w", err)
	}

	return s.GetCoverLetter(ctx, id)
}

// UpdateCoverLetter updates a cover letter's title and body text.
func (s *Store) UpdateCoverLetter(ctx context.Context, id int64, input domain.CoverLetterInput) (domain.CoverLetter, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE cover_letter
		 SET title = ?, body_text = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.Title, input.BodyText, id,
	)
	if err != nil {
		return domain.CoverLetter{}, fmt.Errorf("cover_letter: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.CoverLetter{}, fmt.Errorf("cover_letter: not found: %d", id)
	}

	return s.GetCoverLetter(ctx, id)
}

// UpdateCoverLetterFilePath sets the file_path after PDF export.
func (s *Store) UpdateCoverLetterFilePath(ctx context.Context, id int64, filePath string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE cover_letter
		 SET file_path = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		filePath, id,
	)
	if err != nil {
		return fmt.Errorf("cover_letter: update file_path: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cover_letter: not found: %d", id)
	}

	return nil
}

// DeleteCoverLetter deletes a cover letter by ID.
func (s *Store) DeleteCoverLetter(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM cover_letter WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("cover_letter: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cover_letter: not found: %d", id)
	}

	return nil
}
