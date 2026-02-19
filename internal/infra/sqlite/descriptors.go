package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Role Descriptors ---

// ListDescriptors returns all role descriptors ordered by
// sort_order.
func (s *Store) ListDescriptors(ctx context.Context) ([]domain.RoleDescriptor, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, sort_order, created_at, updated_at
		 FROM role_descriptor
		 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("role_descriptor: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	descs := make([]domain.RoleDescriptor, 0)
	for rows.Next() {
		var desc domain.RoleDescriptor
		if err := rows.Scan(
			&desc.ID, &desc.Title, &desc.SortOrder,
			&desc.CreatedAt, &desc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("role_descriptor: scan: %w", err)
		}
		descs = append(descs, desc)
	}

	return descs, rows.Err()
}

// CreateDescriptor creates a new role descriptor.
func (s *Store) CreateDescriptor(ctx context.Context, title string) (domain.RoleDescriptor, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_order) FROM role_descriptor",
	).Scan(&maxOrder)
	if err != nil {
		return domain.RoleDescriptor{}, fmt.Errorf("role_descriptor: get max sort_order: %w", err)
	}

	nextOrder := 1
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO role_descriptor (title, sort_order) VALUES (?, ?)`,
		title, nextOrder,
	)
	if err != nil {
		return domain.RoleDescriptor{}, fmt.Errorf("role_descriptor: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.RoleDescriptor{}, fmt.Errorf("role_descriptor: last insert id: %w", err)
	}

	return s.getDescriptor(ctx, id)
}

// UpdateDescriptor updates a descriptor's title.
func (s *Store) UpdateDescriptor(ctx context.Context, id int64, title string) (domain.RoleDescriptor, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE role_descriptor
		 SET title = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		title, id,
	)
	if err != nil {
		return domain.RoleDescriptor{}, fmt.Errorf("role_descriptor: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.RoleDescriptor{}, fmt.Errorf("role_descriptor: not found: %d", id)
	}

	return s.getDescriptor(ctx, id)
}

// DeleteDescriptor deletes a role descriptor by ID.
func (s *Store) DeleteDescriptor(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM role_descriptor WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("role_descriptor: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("role_descriptor: not found: %d", id)
	}

	return nil
}

// ReorderDescriptors updates sort_order for all descriptors.
func (s *Store) ReorderDescriptors(ctx context.Context, orderedIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("role_descriptor: begin reorder tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE role_descriptor SET sort_order = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ?",
	)
	if err != nil {
		return fmt.Errorf("role_descriptor: prepare reorder: %w", err)
	}
	defer stmt.Close() //nolint:errcheck

	for i, id := range orderedIDs {
		if _, err := stmt.ExecContext(ctx, i+1, id); err != nil {
			return fmt.Errorf("role_descriptor: reorder id %d: %w", id, err)
		}
	}

	return tx.Commit()
}

// getDescriptor fetches a single descriptor by ID.
func (s *Store) getDescriptor(ctx context.Context, id int64) (domain.RoleDescriptor, error) {
	var desc domain.RoleDescriptor
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, sort_order, created_at, updated_at
		 FROM role_descriptor WHERE id = ?`, id,
	).Scan(
		&desc.ID, &desc.Title, &desc.SortOrder,
		&desc.CreatedAt, &desc.UpdatedAt,
	)
	if err != nil {
		return domain.RoleDescriptor{}, fmt.Errorf("role_descriptor: get %d: %w", id, err)
	}
	return desc, nil
}
