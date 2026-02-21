package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Academic Credentials ---

// ListAcademicCredentials returns all academic records ordered by
// sort_order.
func (s *Store) ListAcademicCredentials(ctx context.Context) ([]domain.AcademicCredential, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, institution, credential_type, field_of_study,
		        completion_date, date_granularity, sort_order,
		        created_at, updated_at
		 FROM academic_credential
		 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("academic_credential: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	creds := make([]domain.AcademicCredential, 0)
	for rows.Next() {
		var cred domain.AcademicCredential
		if err := rows.Scan(
			&cred.ID, &cred.Institution, &cred.CredentialType,
			&cred.FieldOfStudy, &cred.CompletionDate,
			&cred.DateGranularity, &cred.SortOrder,
			&cred.CreatedAt, &cred.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("academic_credential: scan: %w", err)
		}
		creds = append(creds, cred)
	}

	return creds, rows.Err()
}

// CreateAcademicCredential creates a new academic record.
func (s *Store) CreateAcademicCredential(ctx context.Context, input domain.AcademicInput) (domain.AcademicCredential, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_order) FROM academic_credential",
	).Scan(&maxOrder)
	if err != nil {
		return domain.AcademicCredential{}, fmt.Errorf("academic_credential: get max sort_order: %w", err)
	}

	nextOrder := 1
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	gran := input.DateGranularity
	if gran == "" {
		gran = "year"
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO academic_credential
		 (institution, credential_type, field_of_study, completion_date, date_granularity, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		input.Institution, input.CredentialType, input.FieldOfStudy,
		input.CompletionDate, gran, nextOrder,
	)
	if err != nil {
		return domain.AcademicCredential{}, fmt.Errorf("academic_credential: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.AcademicCredential{}, fmt.Errorf("academic_credential: last insert id: %w", err)
	}

	return s.getAcademicCredential(ctx, id)
}

// UpdateAcademicCredential updates an academic record.
func (s *Store) UpdateAcademicCredential(ctx context.Context, id int64, input domain.AcademicInput) (domain.AcademicCredential, error) {
	gran := input.DateGranularity
	if gran == "" {
		gran = "year"
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE academic_credential
		 SET institution = ?, credential_type = ?, field_of_study = ?,
		     completion_date = ?, date_granularity = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.Institution, input.CredentialType, input.FieldOfStudy,
		input.CompletionDate, gran, id,
	)
	if err != nil {
		return domain.AcademicCredential{}, fmt.Errorf("academic_credential: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.AcademicCredential{}, fmt.Errorf("academic_credential: not found: %d", id)
	}

	return s.getAcademicCredential(ctx, id)
}

// DeleteAcademicCredential deletes an academic record by ID.
func (s *Store) DeleteAcademicCredential(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM academic_credential WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("academic_credential: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("academic_credential: not found: %d", id)
	}

	return nil
}

// ReorderAcademicCredentials updates sort_order for all academic
// credentials.
func (s *Store) ReorderAcademicCredentials(ctx context.Context, orderedIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("academic_credential: begin reorder tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE academic_credential SET sort_order = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ?",
	)
	if err != nil {
		return fmt.Errorf("academic_credential: prepare reorder: %w", err)
	}
	defer stmt.Close() //nolint:errcheck

	for i, id := range orderedIDs {
		if _, err := stmt.ExecContext(ctx, i+1, id); err != nil {
			return fmt.Errorf("academic_credential: reorder id %d: %w", id, err)
		}
	}

	return tx.Commit()
}

// getAcademicCredential fetches a single academic credential by ID.
func (s *Store) getAcademicCredential(ctx context.Context, id int64) (domain.AcademicCredential, error) {
	var cred domain.AcademicCredential
	err := s.db.QueryRowContext(ctx,
		`SELECT id, institution, credential_type, field_of_study,
		        completion_date, date_granularity, sort_order,
		        created_at, updated_at
		 FROM academic_credential WHERE id = ?`, id,
	).Scan(
		&cred.ID, &cred.Institution, &cred.CredentialType,
		&cred.FieldOfStudy, &cred.CompletionDate,
		&cred.DateGranularity, &cred.SortOrder,
		&cred.CreatedAt, &cred.UpdatedAt,
	)
	if err != nil {
		return domain.AcademicCredential{}, fmt.Errorf("academic_credential: get %d: %w", id, err)
	}
	return cred, nil
}
