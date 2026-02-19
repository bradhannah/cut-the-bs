package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cut-the-bs/internal/domain"
)

// --- Certifications ---

// ListCertifications returns all certifications with computed
// active/inactive status, ordered by sort_order.
func (s *Store) ListCertifications(ctx context.Context) ([]domain.Certification, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, issuing_body, date_earned, expiration_date,
		        sort_order, created_at, updated_at
		 FROM certification
		 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("certification: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	certs := make([]domain.Certification, 0)
	for rows.Next() {
		var cert domain.Certification
		var expiration sql.NullString
		if err := rows.Scan(
			&cert.ID, &cert.Name, &cert.IssuingBody,
			&cert.DateEarned, &expiration,
			&cert.SortOrder, &cert.CreatedAt, &cert.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("certification: scan: %w", err)
		}
		if expiration.Valid {
			cert.ExpirationDate = expiration.String
		}
		cert.IsActive = computeIsActive(cert.ExpirationDate)
		certs = append(certs, cert)
	}

	return certs, rows.Err()
}

// CreateCertification creates a new certification.
func (s *Store) CreateCertification(ctx context.Context, input domain.CertificationInput) (domain.Certification, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_order) FROM certification",
	).Scan(&maxOrder)
	if err != nil {
		return domain.Certification{}, fmt.Errorf("certification: get max sort_order: %w", err)
	}

	nextOrder := 1
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	var expiration interface{}
	if input.ExpirationDate != "" {
		expiration = input.ExpirationDate
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO certification
		 (name, issuing_body, date_earned, expiration_date, sort_order)
		 VALUES (?, ?, ?, ?, ?)`,
		input.Name, input.IssuingBody, input.DateEarned, expiration, nextOrder,
	)
	if err != nil {
		return domain.Certification{}, fmt.Errorf("certification: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Certification{}, fmt.Errorf("certification: last insert id: %w", err)
	}

	return s.getCertification(ctx, id)
}

// UpdateCertification updates a certification.
func (s *Store) UpdateCertification(ctx context.Context, id int64, input domain.CertificationInput) (domain.Certification, error) {
	var expiration interface{}
	if input.ExpirationDate != "" {
		expiration = input.ExpirationDate
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE certification
		 SET name = ?, issuing_body = ?, date_earned = ?, expiration_date = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.Name, input.IssuingBody, input.DateEarned, expiration, id,
	)
	if err != nil {
		return domain.Certification{}, fmt.Errorf("certification: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.Certification{}, fmt.Errorf("certification: not found: %d", id)
	}

	return s.getCertification(ctx, id)
}

// DeleteCertification deletes a certification by ID.
func (s *Store) DeleteCertification(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM certification WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("certification: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("certification: not found: %d", id)
	}

	return nil
}

// ReorderCertifications updates sort_order for all certifications.
func (s *Store) ReorderCertifications(ctx context.Context, orderedIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("certification: begin reorder tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE certification SET sort_order = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ?",
	)
	if err != nil {
		return fmt.Errorf("certification: prepare reorder: %w", err)
	}
	defer stmt.Close() //nolint:errcheck

	for i, id := range orderedIDs {
		if _, err := stmt.ExecContext(ctx, i+1, id); err != nil {
			return fmt.Errorf("certification: reorder id %d: %w", id, err)
		}
	}

	return tx.Commit()
}

// getCertification fetches a single certification by ID with
// computed IsActive status.
func (s *Store) getCertification(ctx context.Context, id int64) (domain.Certification, error) {
	var cert domain.Certification
	var expiration sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, issuing_body, date_earned, expiration_date,
		        sort_order, created_at, updated_at
		 FROM certification WHERE id = ?`, id,
	).Scan(
		&cert.ID, &cert.Name, &cert.IssuingBody,
		&cert.DateEarned, &expiration,
		&cert.SortOrder, &cert.CreatedAt, &cert.UpdatedAt,
	)
	if err != nil {
		return domain.Certification{}, fmt.Errorf("certification: get %d: %w", id, err)
	}
	if expiration.Valid {
		cert.ExpirationDate = expiration.String
	}
	cert.IsActive = computeIsActive(cert.ExpirationDate)
	return cert, nil
}

// computeIsActive returns true if the certification is currently
// active. A cert is active if expiration_date is empty (no
// expiration) or if the expiration date is in the future.
func computeIsActive(expirationDate string) bool {
	if expirationDate == "" {
		return true
	}
	// Try parsing as YYYY-MM-DD first, then YYYY-MM, then YYYY.
	for _, layout := range []string{"2006-01-02", "2006-01", "2006"} {
		if t, err := time.Parse(layout, expirationDate); err == nil {
			return t.After(time.Now())
		}
	}
	// If we can't parse, assume active.
	return true
}
