package sqlite

import (
	"context"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Professional Summaries ---

// ListSummaries returns all professional summary variants.
func (s *Store) ListSummaries(ctx context.Context) ([]domain.ProfessionalSummary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, label, body_text, created_at, updated_at
		 FROM professional_summary
		 ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("professional_summary: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	summaries := make([]domain.ProfessionalSummary, 0)
	for rows.Next() {
		var s domain.ProfessionalSummary
		if err := rows.Scan(
			&s.ID, &s.Label, &s.BodyText,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("professional_summary: scan: %w", err)
		}
		summaries = append(summaries, s)
	}

	return summaries, rows.Err()
}

// GetSummary returns a single summary by ID.
func (s *Store) GetSummary(ctx context.Context, id int64) (domain.ProfessionalSummary, error) {
	var summary domain.ProfessionalSummary
	err := s.db.QueryRowContext(ctx,
		`SELECT id, label, body_text, created_at, updated_at
		 FROM professional_summary WHERE id = ?`, id,
	).Scan(
		&summary.ID, &summary.Label, &summary.BodyText,
		&summary.CreatedAt, &summary.UpdatedAt,
	)
	if err != nil {
		return domain.ProfessionalSummary{}, fmt.Errorf("professional_summary: get %d: %w", id, err)
	}
	return summary, nil
}

// CreateSummary creates a new summary variant.
func (s *Store) CreateSummary(ctx context.Context, input domain.SummaryInput) (domain.ProfessionalSummary, error) {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO professional_summary (label, body_text) VALUES (?, ?)`,
		input.Label, input.BodyText,
	)
	if err != nil {
		return domain.ProfessionalSummary{}, fmt.Errorf("professional_summary: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.ProfessionalSummary{}, fmt.Errorf("professional_summary: last insert id: %w", err)
	}

	return s.GetSummary(ctx, id)
}

// UpdateSummary updates an existing summary.
func (s *Store) UpdateSummary(ctx context.Context, id int64, input domain.SummaryInput) (domain.ProfessionalSummary, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE professional_summary
		 SET label = ?, body_text = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.Label, input.BodyText, id,
	)
	if err != nil {
		return domain.ProfessionalSummary{}, fmt.Errorf("professional_summary: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.ProfessionalSummary{}, fmt.Errorf("professional_summary: not found: %d", id)
	}

	return s.GetSummary(ctx, id)
}

// DeleteSummary deletes a summary variant by ID.
func (s *Store) DeleteSummary(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM professional_summary WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("professional_summary: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("professional_summary: not found: %d", id)
	}

	return nil
}
