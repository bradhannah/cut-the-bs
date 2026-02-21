package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Core Expertise ---

// ListCoreExpertise returns all core expertise items ordered by
// sort_order.
func (s *Store) ListCoreExpertise(ctx context.Context) ([]domain.CoreExpertise, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, label, sort_order, created_at, updated_at
		 FROM core_expertise
		 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("core_expertise: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	items := make([]domain.CoreExpertise, 0)
	for rows.Next() {
		var item domain.CoreExpertise
		if err := rows.Scan(
			&item.ID, &item.Label, &item.SortOrder,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("core_expertise: scan: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// CreateCoreExpertise creates a new core expertise item.
func (s *Store) CreateCoreExpertise(ctx context.Context, label string) (domain.CoreExpertise, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_order) FROM core_expertise",
	).Scan(&maxOrder)
	if err != nil {
		return domain.CoreExpertise{}, fmt.Errorf("core_expertise: get max sort_order: %w", err)
	}

	nextOrder := 1
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO core_expertise (label, sort_order) VALUES (?, ?)`,
		label, nextOrder,
	)
	if err != nil {
		return domain.CoreExpertise{}, fmt.Errorf("core_expertise: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.CoreExpertise{}, fmt.Errorf("core_expertise: last insert id: %w", err)
	}

	return s.getCoreExpertise(ctx, id)
}

// UpdateCoreExpertise updates a core expertise item's label.
func (s *Store) UpdateCoreExpertise(ctx context.Context, id int64, label string) (domain.CoreExpertise, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE core_expertise
		 SET label = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		label, id,
	)
	if err != nil {
		return domain.CoreExpertise{}, fmt.Errorf("core_expertise: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.CoreExpertise{}, fmt.Errorf("core_expertise: not found: %d", id)
	}

	return s.getCoreExpertise(ctx, id)
}

// DeleteCoreExpertise deletes a core expertise item by ID.
func (s *Store) DeleteCoreExpertise(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM core_expertise WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("core_expertise: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("core_expertise: not found: %d", id)
	}

	return nil
}

// ReorderCoreExpertise updates sort_order for all core expertise
// items.
func (s *Store) ReorderCoreExpertise(ctx context.Context, orderedIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("core_expertise: begin reorder tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE core_expertise SET sort_order = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ?",
	)
	if err != nil {
		return fmt.Errorf("core_expertise: prepare reorder: %w", err)
	}
	defer stmt.Close() //nolint:errcheck

	for i, id := range orderedIDs {
		if _, err := stmt.ExecContext(ctx, i+1, id); err != nil {
			return fmt.Errorf("core_expertise: reorder id %d: %w", id, err)
		}
	}

	return tx.Commit()
}

// getCoreExpertise fetches a single core expertise item by ID.
func (s *Store) getCoreExpertise(ctx context.Context, id int64) (domain.CoreExpertise, error) {
	var item domain.CoreExpertise
	err := s.db.QueryRowContext(ctx,
		`SELECT id, label, sort_order, created_at, updated_at
		 FROM core_expertise WHERE id = ?`, id,
	).Scan(
		&item.ID, &item.Label, &item.SortOrder,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return domain.CoreExpertise{}, fmt.Errorf("core_expertise: get %d: %w", id, err)
	}
	return item, nil
}
