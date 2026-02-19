package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Profile Links ---

// ListProfileLinks returns all profile links ordered by sort_order.
func (s *Store) ListProfileLinks(ctx context.Context) ([]domain.ProfileLink, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, label, url, sort_order, created_at, updated_at
		 FROM profile_link
		 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("profile_link: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	links := make([]domain.ProfileLink, 0)
	for rows.Next() {
		var link domain.ProfileLink
		if err := rows.Scan(
			&link.ID, &link.Label, &link.URL,
			&link.SortOrder, &link.CreatedAt, &link.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("profile_link: scan: %w", err)
		}
		links = append(links, link)
	}

	return links, rows.Err()
}

// CreateProfileLink creates a new profile link. The sort_order is
// automatically set to one past the current maximum.
func (s *Store) CreateProfileLink(ctx context.Context, input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_order) FROM profile_link",
	).Scan(&maxOrder)
	if err != nil {
		return domain.ProfileLink{}, fmt.Errorf("profile_link: get max sort_order: %w", err)
	}

	nextOrder := 1
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO profile_link (label, url, sort_order)
		 VALUES (?, ?, ?)`,
		input.Label, input.URL, nextOrder,
	)
	if err != nil {
		return domain.ProfileLink{}, fmt.Errorf("profile_link: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.ProfileLink{}, fmt.Errorf("profile_link: last insert id: %w", err)
	}

	return s.getProfileLink(ctx, id)
}

// UpdateProfileLink updates an existing profile link's label and URL.
func (s *Store) UpdateProfileLink(ctx context.Context, id int64, input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE profile_link
		 SET label = ?, url = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.Label, input.URL, id,
	)
	if err != nil {
		return domain.ProfileLink{}, fmt.Errorf("profile_link: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.ProfileLink{}, fmt.Errorf("profile_link: not found: %d", id)
	}

	return s.getProfileLink(ctx, id)
}

// DeleteProfileLink deletes a profile link by ID.
func (s *Store) DeleteProfileLink(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM profile_link WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("profile_link: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("profile_link: not found: %d", id)
	}

	return nil
}

// ReorderProfileLinks updates the sort_order of all profile links.
// The orderedIDs slice specifies the desired order (index 0 gets
// sort_order 1, etc.).
func (s *Store) ReorderProfileLinks(ctx context.Context, orderedIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("profile_link: begin reorder tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE profile_link SET sort_order = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ?",
	)
	if err != nil {
		return fmt.Errorf("profile_link: prepare reorder: %w", err)
	}
	defer stmt.Close() //nolint:errcheck

	for i, id := range orderedIDs {
		if _, err := stmt.ExecContext(ctx, i+1, id); err != nil {
			return fmt.Errorf("profile_link: reorder id %d: %w", id, err)
		}
	}

	return tx.Commit()
}

// getProfileLink fetches a single profile link by ID.
func (s *Store) getProfileLink(ctx context.Context, id int64) (domain.ProfileLink, error) {
	var link domain.ProfileLink
	err := s.db.QueryRowContext(ctx,
		`SELECT id, label, url, sort_order, created_at, updated_at
		 FROM profile_link WHERE id = ?`, id,
	).Scan(
		&link.ID, &link.Label, &link.URL,
		&link.SortOrder, &link.CreatedAt, &link.UpdatedAt,
	)
	if err != nil {
		return domain.ProfileLink{}, fmt.Errorf("profile_link: get %d: %w", id, err)
	}
	return link, nil
}
