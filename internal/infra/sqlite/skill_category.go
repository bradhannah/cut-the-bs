package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Skill Categories ---

// ListSkillCategories returns all skill categories ordered by
// sort_order.
func (s *Store) ListSkillCategories(ctx context.Context) ([]domain.SkillCategory, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, sort_order, created_at, updated_at
		 FROM skill_category
		 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("skill_category: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	cats := make([]domain.SkillCategory, 0)
	for rows.Next() {
		var cat domain.SkillCategory
		if err := rows.Scan(
			&cat.ID, &cat.Name, &cat.SortOrder,
			&cat.CreatedAt, &cat.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("skill_category: scan: %w", err)
		}
		cats = append(cats, cat)
	}

	return cats, rows.Err()
}

// CreateSkillCategory creates a new skill category appended to
// the end of the sort order.
func (s *Store) CreateSkillCategory(ctx context.Context, name string) (domain.SkillCategory, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_order) FROM skill_category",
	).Scan(&maxOrder)
	if err != nil {
		return domain.SkillCategory{}, fmt.Errorf("skill_category: get max sort_order: %w", err)
	}

	nextOrder := 1
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO skill_category (name, sort_order) VALUES (?, ?)`,
		name, nextOrder,
	)
	if err != nil {
		return domain.SkillCategory{}, fmt.Errorf("skill_category: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.SkillCategory{}, fmt.Errorf("skill_category: last insert id: %w", err)
	}

	return s.getSkillCategory(ctx, id)
}

// RenameSkillCategory updates a category's name.
func (s *Store) RenameSkillCategory(ctx context.Context, id int64, name string) (domain.SkillCategory, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE skill_category
		 SET name = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		name, id,
	)
	if err != nil {
		return domain.SkillCategory{}, fmt.Errorf("skill_category: rename: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.SkillCategory{}, fmt.Errorf("skill_category: not found: %d", id)
	}

	return s.getSkillCategory(ctx, id)
}

// DeleteSkillCategory deletes a skill category. Fails if skills
// still reference it (FK constraint).
func (s *Store) DeleteSkillCategory(ctx context.Context, id int64) error {
	// Check if any skills reference this category.
	var count int
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM skill WHERE category_id = ?", id,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("skill_category: check references: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("skill_category: cannot delete category with %d skill(s) still assigned", count)
	}

	result, err := s.db.ExecContext(ctx,
		"DELETE FROM skill_category WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("skill_category: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("skill_category: not found: %d", id)
	}

	return nil
}

// ReorderSkillCategories updates the sort_order of all categories.
func (s *Store) ReorderSkillCategories(ctx context.Context, orderedIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("skill_category: begin reorder tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE skill_category SET sort_order = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ?",
	)
	if err != nil {
		return fmt.Errorf("skill_category: prepare reorder: %w", err)
	}
	defer stmt.Close() //nolint:errcheck

	for i, id := range orderedIDs {
		if _, err := stmt.ExecContext(ctx, i+1, id); err != nil {
			return fmt.Errorf("skill_category: reorder id %d: %w", id, err)
		}
	}

	return tx.Commit()
}

// getSkillCategory fetches a single category by ID.
func (s *Store) getSkillCategory(ctx context.Context, id int64) (domain.SkillCategory, error) {
	var cat domain.SkillCategory
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, sort_order, created_at, updated_at
		 FROM skill_category WHERE id = ?`, id,
	).Scan(
		&cat.ID, &cat.Name, &cat.SortOrder,
		&cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		return domain.SkillCategory{}, fmt.Errorf("skill_category: get %d: %w", id, err)
	}
	return cat, nil
}
