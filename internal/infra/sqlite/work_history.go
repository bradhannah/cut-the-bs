package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Work History Entries ---

// CreateWorkHistory creates a new work history entry. The sort_order
// is automatically set to one past the current maximum.
func (s *Store) CreateWorkHistory(ctx context.Context, input domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_order) FROM work_history_entry",
	).Scan(&maxOrder)
	if err != nil {
		return domain.WorkHistoryEntry{}, fmt.Errorf("work_history: get max sort_order: %w", err)
	}

	nextOrder := 1
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO work_history_entry
			(employer_name, job_title, start_date, end_date,
			 date_granularity_start, date_granularity_end, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.EmployerName, input.JobTitle,
		input.StartDate, input.EndDate,
		input.DateGranularityStart, input.DateGranularityEnd,
		nextOrder,
	)
	if err != nil {
		return domain.WorkHistoryEntry{}, fmt.Errorf("work_history: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.WorkHistoryEntry{}, fmt.Errorf("work_history: last insert id: %w", err)
	}

	return s.GetWorkHistory(ctx, id)
}

// GetWorkHistory returns a single work history entry by ID with its
// bullets included, ordered by sort_order.
func (s *Store) GetWorkHistory(ctx context.Context, id int64) (domain.WorkHistoryEntry, error) {
	var entry domain.WorkHistoryEntry
	var endDate, granEnd sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT id, employer_name, job_title, start_date, end_date,
		        date_granularity_start, date_granularity_end,
		        sort_order, created_at, updated_at
		 FROM work_history_entry WHERE id = ?`, id,
	).Scan(
		&entry.ID, &entry.EmployerName, &entry.JobTitle,
		&entry.StartDate, &endDate,
		&entry.DateGranularityStart, &granEnd,
		&entry.SortOrder, &entry.CreatedAt, &entry.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.WorkHistoryEntry{}, fmt.Errorf("work_history: not found: id=%d", id)
		}
		return domain.WorkHistoryEntry{}, fmt.Errorf("work_history: get: %w", err)
	}

	entry.EndDate = endDate.String
	entry.DateGranularityEnd = granEnd.String

	bullets, err := s.listBulletsForEntry(ctx, id)
	if err != nil {
		return domain.WorkHistoryEntry{}, err
	}
	entry.Bullets = bullets

	return entry, nil
}

// UpdateWorkHistory updates an existing work history entry's fields.
// The sort_order is not changed.
func (s *Store) UpdateWorkHistory(ctx context.Context, id int64, input domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE work_history_entry
		 SET employer_name = ?, job_title = ?, start_date = ?,
		     end_date = ?, date_granularity_start = ?,
		     date_granularity_end = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.EmployerName, input.JobTitle,
		input.StartDate, input.EndDate,
		input.DateGranularityStart, input.DateGranularityEnd,
		id,
	)
	if err != nil {
		return domain.WorkHistoryEntry{}, fmt.Errorf("work_history: update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.WorkHistoryEntry{}, fmt.Errorf("work_history: rows affected: %w", err)
	}
	if rows == 0 {
		return domain.WorkHistoryEntry{}, fmt.Errorf("work_history: not found: id=%d", id)
	}

	return s.GetWorkHistory(ctx, id)
}

// DeleteWorkHistory deletes a work history entry. Associated bullets
// are deleted via ON DELETE CASCADE in the schema.
func (s *Store) DeleteWorkHistory(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM work_history_entry WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("work_history: delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("work_history: rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("work_history: not found: id=%d", id)
	}

	return nil
}

// ListWorkHistory returns all work history entries ordered by
// sort_order. Each entry includes its bullets.
//
// Note: entries are fully scanned before fetching bullets because
// SQLite with MaxOpenConns(1) deadlocks if a second query runs
// while rows from the first are still open.
func (s *Store) ListWorkHistory(ctx context.Context) ([]domain.WorkHistoryEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, employer_name, job_title, start_date, end_date,
		        date_granularity_start, date_granularity_end,
		        sort_order, created_at, updated_at
		 FROM work_history_entry ORDER BY sort_order`,
	)
	if err != nil {
		return nil, fmt.Errorf("work_history: list: %w", err)
	}
	defer rows.Close()

	var entries []domain.WorkHistoryEntry
	for rows.Next() {
		var entry domain.WorkHistoryEntry
		var endDate, granEnd sql.NullString

		if err := rows.Scan(
			&entry.ID, &entry.EmployerName, &entry.JobTitle,
			&entry.StartDate, &endDate,
			&entry.DateGranularityStart, &granEnd,
			&entry.SortOrder, &entry.CreatedAt, &entry.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("work_history: scan: %w", err)
		}

		entry.EndDate = endDate.String
		entry.DateGranularityEnd = granEnd.String

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("work_history: rows: %w", err)
	}

	if entries == nil {
		entries = []domain.WorkHistoryEntry{}
	}

	// Fetch bullets in a second pass now that the rows cursor is consumed.
	for i := range entries {
		bullets, err := s.listBulletsForEntry(ctx, entries[i].ID)
		if err != nil {
			return nil, err
		}
		entries[i].Bullets = bullets
	}

	return entries, nil
}

// ReorderWorkHistory sets the sort_order for each entry based on its
// position in the provided ordered slice of IDs.
func (s *Store) ReorderWorkHistory(ctx context.Context, orderedIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("work_history: begin reorder tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx,
		`UPDATE work_history_entry
		 SET sort_order = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
	)
	if err != nil {
		return fmt.Errorf("work_history: prepare reorder: %w", err)
	}
	defer stmt.Close()

	for i, id := range orderedIDs {
		if _, err := stmt.ExecContext(ctx, i+1, id); err != nil {
			return fmt.Errorf("work_history: reorder id=%d: %w", id, err)
		}
	}

	return tx.Commit()
}

// --- Achievement Bullets ---

// CreateBullet adds an achievement bullet to a work history entry.
// The sort_order is automatically set to one past the current maximum
// for that entry.
func (s *Store) CreateBullet(ctx context.Context, workHistoryID int64, text string) (domain.AchievementBullet, error) {
	// Verify the work history entry exists.
	var exists int
	err := s.db.QueryRowContext(ctx,
		"SELECT 1 FROM work_history_entry WHERE id = ?", workHistoryID,
	).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.AchievementBullet{}, fmt.Errorf("bullet: work history not found: id=%d", workHistoryID)
		}
		return domain.AchievementBullet{}, fmt.Errorf("bullet: check entry: %w", err)
	}

	var maxOrder sql.NullInt64
	err = s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_order) FROM achievement_bullet WHERE work_history_id = ?",
		workHistoryID,
	).Scan(&maxOrder)
	if err != nil {
		return domain.AchievementBullet{}, fmt.Errorf("bullet: get max sort_order: %w", err)
	}

	nextOrder := 1
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO achievement_bullet
			(work_history_id, text, sort_order)
		 VALUES (?, ?, ?)`,
		workHistoryID, text, nextOrder,
	)
	if err != nil {
		return domain.AchievementBullet{}, fmt.Errorf("bullet: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.AchievementBullet{}, fmt.Errorf("bullet: last insert id: %w", err)
	}

	return s.getBullet(ctx, id)
}

// UpdateBullet updates an achievement bullet's text.
func (s *Store) UpdateBullet(ctx context.Context, id int64, text string) (domain.AchievementBullet, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE achievement_bullet
		 SET text = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		text, id,
	)
	if err != nil {
		return domain.AchievementBullet{}, fmt.Errorf("bullet: update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.AchievementBullet{}, fmt.Errorf("bullet: rows affected: %w", err)
	}
	if rows == 0 {
		return domain.AchievementBullet{}, fmt.Errorf("bullet: not found: id=%d", id)
	}

	return s.getBullet(ctx, id)
}

// DeleteBullet deletes an achievement bullet by ID.
func (s *Store) DeleteBullet(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM achievement_bullet WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("bullet: delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("bullet: rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bullet: not found: id=%d", id)
	}

	return nil
}

// ReorderBullets sets the sort_order for each bullet within a work
// history entry based on position in the provided ordered slice.
func (s *Store) ReorderBullets(ctx context.Context, workHistoryID int64, orderedIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("bullet: begin reorder tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx,
		`UPDATE achievement_bullet
		 SET sort_order = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ? AND work_history_id = ?`,
	)
	if err != nil {
		return fmt.Errorf("bullet: prepare reorder: %w", err)
	}
	defer stmt.Close()

	for i, id := range orderedIDs {
		if _, err := stmt.ExecContext(ctx, i+1, id, workHistoryID); err != nil {
			return fmt.Errorf("bullet: reorder id=%d: %w", id, err)
		}
	}

	return tx.Commit()
}

// --- Internal helpers ---

// getBullet returns a single achievement bullet by ID.
func (s *Store) getBullet(ctx context.Context, id int64) (domain.AchievementBullet, error) {
	var b domain.AchievementBullet
	err := s.db.QueryRowContext(ctx,
		`SELECT id, work_history_id, text, sort_order, created_at, updated_at
		 FROM achievement_bullet WHERE id = ?`, id,
	).Scan(&b.ID, &b.WorkHistoryID, &b.Text, &b.SortOrder, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return domain.AchievementBullet{}, fmt.Errorf("bullet: get: %w", err)
	}
	return b, nil
}

// listBulletsForEntry returns all bullets for a work history entry,
// ordered by sort_order.
func (s *Store) listBulletsForEntry(ctx context.Context, workHistoryID int64) ([]domain.AchievementBullet, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, work_history_id, text, sort_order, created_at, updated_at
		 FROM achievement_bullet
		 WHERE work_history_id = ?
		 ORDER BY sort_order`, workHistoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("bullet: list for entry %d: %w", workHistoryID, err)
	}
	defer rows.Close()

	var bullets []domain.AchievementBullet
	for rows.Next() {
		var b domain.AchievementBullet
		if err := rows.Scan(&b.ID, &b.WorkHistoryID, &b.Text, &b.SortOrder, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("bullet: scan: %w", err)
		}
		bullets = append(bullets, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("bullet: rows: %w", err)
	}

	return bullets, nil
}
