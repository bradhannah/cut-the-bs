package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Lenses ---

// ListLenses returns all lenses ordered by name.
func (s *Store) ListLenses(ctx context.Context) ([]domain.Lens, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, created_at, updated_at
		 FROM lens
		 ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("lens: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	lenses := make([]domain.Lens, 0)
	for rows.Next() {
		var l domain.Lens
		if err := rows.Scan(
			&l.ID, &l.Name,
			&l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("lens: scan: %w", err)
		}
		lenses = append(lenses, l)
	}

	return lenses, rows.Err()
}

// GetLens returns a single lens with all its content selections.
func (s *Store) GetLens(ctx context.Context, id int64) (domain.LensDetail, error) {
	// 1. Fetch the lens record itself.
	var detail domain.LensDetail
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at, updated_at
		 FROM lens WHERE id = ?`, id,
	).Scan(
		&detail.ID, &detail.Name,
		&detail.CreatedAt, &detail.UpdatedAt,
	)
	if err != nil {
		return domain.LensDetail{}, fmt.Errorf("lens: get %d: %w", id, err)
	}

	// 2. Fetch summary selections.
	detail.Summaries, err = s.getLensSummaries(ctx, id)
	if err != nil {
		return domain.LensDetail{}, err
	}

	// 3. Fetch work history selections.
	detail.WorkHistory, err = s.getLensWorkHistory(ctx, id)
	if err != nil {
		return domain.LensDetail{}, err
	}

	// 4. Fetch bullet selections.
	detail.Bullets, err = s.getLensBullets(ctx, id)
	if err != nil {
		return domain.LensDetail{}, err
	}

	// 5. Fetch skill selections.
	detail.Skills, err = s.getLensSkills(ctx, id)
	if err != nil {
		return domain.LensDetail{}, err
	}

	// 6. Fetch academic selections.
	detail.AcademicIDs, err = s.getLensAcademics(ctx, id)
	if err != nil {
		return domain.LensDetail{}, err
	}

	// 7. Fetch cert selections.
	detail.CertIDs, err = s.getLensCerts(ctx, id)
	if err != nil {
		return domain.LensDetail{}, err
	}

	// 8. Fetch descriptor selections.
	detail.Descriptors, err = s.getLensDescriptors(ctx, id)
	if err != nil {
		return domain.LensDetail{}, err
	}

	// 9. Fetch core expertise selections.
	detail.CoreExpertise, err = s.getLensCoreExpertise(ctx, id)
	if err != nil {
		return domain.LensDetail{}, err
	}

	return detail, nil
}

// CreateLens creates a new lens.
func (s *Store) CreateLens(ctx context.Context, input domain.LensInput) (domain.Lens, error) {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO lens (name) VALUES (?)`,
		input.Name,
	)
	if err != nil {
		return domain.Lens{}, fmt.Errorf("lens: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Lens{}, fmt.Errorf("lens: last insert id: %w", err)
	}

	return s.getLens(ctx, id)
}

// UpdateLens updates a lens's name.
func (s *Store) UpdateLens(ctx context.Context, id int64, input domain.LensInput) (domain.Lens, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE lens
		 SET name = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.Name, id,
	)
	if err != nil {
		return domain.Lens{}, fmt.Errorf("lens: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.Lens{}, fmt.Errorf("lens: not found: %d", id)
	}

	return s.getLens(ctx, id)
}

// DeleteLens deletes a lens and all its selections (CASCADE).
func (s *Store) DeleteLens(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM lens WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("lens: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("lens: not found: %d", id)
	}

	return nil
}

// --- Selection Setters ---
// Each setter deletes all existing rows for the lens, then inserts
// the new set within a transaction.

// SetLensSummaries replaces the summary selections for a lens.
func (s *Store) SetLensSummaries(ctx context.Context, lensID int64, selections []domain.LensSummaryItem) error {
	return s.setLensSelections(ctx, lensID,
		"DELETE FROM lens_summary_selection WHERE lens_id = ?",
		"INSERT INTO lens_summary_selection (lens_id, summary_id, sort_order, is_master) VALUES (?, ?, ?, ?)",
		func(i int) []interface{} {
			var isMaster int
			if selections[i].IsMaster {
				isMaster = 1
			}
			return []interface{}{lensID, selections[i].SummaryID, selections[i].SortOrder, isMaster}
		},
		len(selections),
	)
}

// SetLensWorkHistory replaces the work history selections for a lens.
func (s *Store) SetLensWorkHistory(ctx context.Context, lensID int64, selections []domain.LensWorkHistoryItem) error {
	return s.setLensSelections(ctx, lensID,
		"DELETE FROM lens_work_history_selection WHERE lens_id = ?",
		"INSERT INTO lens_work_history_selection (lens_id, work_history_id, sort_order) VALUES (?, ?, ?)",
		func(i int) []interface{} {
			return []interface{}{lensID, selections[i].WorkHistoryID, selections[i].SortOrder}
		},
		len(selections),
	)
}

// SetLensBullets replaces the bullet selections for a lens.
func (s *Store) SetLensBullets(ctx context.Context, lensID int64, selections []domain.LensBulletItem) error {
	return s.setLensSelections(ctx, lensID,
		"DELETE FROM lens_bullet_selection WHERE lens_id = ?",
		"INSERT INTO lens_bullet_selection (lens_id, bullet_id, sort_order) VALUES (?, ?, ?)",
		func(i int) []interface{} {
			return []interface{}{lensID, selections[i].BulletID, selections[i].SortOrder}
		},
		len(selections),
	)
}

// SetLensSkills replaces the skill selections for a lens.
func (s *Store) SetLensSkills(ctx context.Context, lensID int64, selections []domain.LensSkillItem) error {
	return s.setLensSelections(ctx, lensID,
		"DELETE FROM lens_skill_selection WHERE lens_id = ?",
		"INSERT INTO lens_skill_selection (lens_id, skill_id, custom_sort_order) VALUES (?, ?, ?)",
		func(i int) []interface{} {
			var customSort interface{}
			if selections[i].CustomSortOrder != nil {
				customSort = *selections[i].CustomSortOrder
			}
			return []interface{}{lensID, selections[i].SkillID, customSort}
		},
		len(selections),
	)
}

// SetLensAcademics replaces the academic selections for a lens.
func (s *Store) SetLensAcademics(ctx context.Context, lensID int64, academicIDs []int64) error {
	return s.setLensSelections(ctx, lensID,
		"DELETE FROM lens_academic_selection WHERE lens_id = ?",
		"INSERT INTO lens_academic_selection (lens_id, academic_id) VALUES (?, ?)",
		func(i int) []interface{} {
			return []interface{}{lensID, academicIDs[i]}
		},
		len(academicIDs),
	)
}

// SetLensCerts replaces the certification selections for a lens.
func (s *Store) SetLensCerts(ctx context.Context, lensID int64, certIDs []int64) error {
	return s.setLensSelections(ctx, lensID,
		"DELETE FROM lens_cert_selection WHERE lens_id = ?",
		"INSERT INTO lens_cert_selection (lens_id, cert_id) VALUES (?, ?)",
		func(i int) []interface{} {
			return []interface{}{lensID, certIDs[i]}
		},
		len(certIDs),
	)
}

// SetLensDescriptors replaces the descriptor selections for a lens.
func (s *Store) SetLensDescriptors(ctx context.Context, lensID int64, selections []domain.LensDescriptorItem) error {
	return s.setLensSelections(ctx, lensID,
		"DELETE FROM lens_descriptor_selection WHERE lens_id = ?",
		"INSERT INTO lens_descriptor_selection (lens_id, descriptor_id, sort_order) VALUES (?, ?, ?)",
		func(i int) []interface{} {
			return []interface{}{lensID, selections[i].DescriptorID, selections[i].SortOrder}
		},
		len(selections),
	)
}

// SetLensCoreExpertise replaces the core expertise selections for a lens.
func (s *Store) SetLensCoreExpertise(ctx context.Context, lensID int64, selections []domain.LensCoreExpertiseItem) error {
	return s.setLensSelections(ctx, lensID,
		"DELETE FROM lens_core_expertise_selection WHERE lens_id = ?",
		"INSERT INTO lens_core_expertise_selection (lens_id, core_expertise_id, sort_order) VALUES (?, ?, ?)",
		func(i int) []interface{} {
			return []interface{}{lensID, selections[i].CoreExpertiseID, selections[i].SortOrder}
		},
		len(selections),
	)
}

// --- Lens Reference Checks ---
// These methods return the names of lenses that reference a given
// content item, for use in delete confirmation dialogs per FR-050.

// CheckWorkHistoryLensReferences returns lens names referencing a
// work history entry.
func (s *Store) CheckWorkHistoryLensReferences(ctx context.Context, id int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.name FROM lens l
		 JOIN lens_work_history_selection lw ON l.id = lw.lens_id
		 WHERE lw.work_history_id = ?
		 ORDER BY l.name ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("work_history: check lens refs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	return scanLensNames(rows)
}

// CheckBulletLensReferences returns lens names referencing a bullet.
func (s *Store) CheckBulletLensReferences(ctx context.Context, id int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.name FROM lens l
		 JOIN lens_bullet_selection lb ON l.id = lb.lens_id
		 WHERE lb.bullet_id = ?
		 ORDER BY l.name ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("bullet: check lens refs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	return scanLensNames(rows)
}

// CheckAcademicLensReferences returns lens names referencing an
// academic credential.
func (s *Store) CheckAcademicLensReferences(ctx context.Context, id int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.name FROM lens l
		 JOIN lens_academic_selection la ON l.id = la.lens_id
		 WHERE la.academic_id = ?
		 ORDER BY l.name ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("academic: check lens refs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	return scanLensNames(rows)
}

// CheckCertLensReferences returns lens names referencing a
// certification.
func (s *Store) CheckCertLensReferences(ctx context.Context, id int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.name FROM lens l
		 JOIN lens_cert_selection lc ON l.id = lc.lens_id
		 WHERE lc.cert_id = ?
		 ORDER BY l.name ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("cert: check lens refs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	return scanLensNames(rows)
}

// CheckDescriptorLensReferences returns lens names referencing a
// role descriptor.
func (s *Store) CheckDescriptorLensReferences(ctx context.Context, id int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.name FROM lens l
		 JOIN lens_descriptor_selection ld ON l.id = ld.lens_id
		 WHERE ld.descriptor_id = ?
		 ORDER BY l.name ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("descriptor: check lens refs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	return scanLensNames(rows)
}

// CheckCoreExpertiseLensReferences returns lens names referencing a
// core expertise item.
func (s *Store) CheckCoreExpertiseLensReferences(ctx context.Context, id int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.name FROM lens l
		 JOIN lens_core_expertise_selection lce ON l.id = lce.lens_id
		 WHERE lce.core_expertise_id = ?
		 ORDER BY l.name ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("core_expertise: check lens refs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	return scanLensNames(rows)
}

// CheckSummaryLensReferences returns lens names referencing a
// professional summary.
func (s *Store) CheckSummaryLensReferences(ctx context.Context, id int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.name FROM lens l
		 JOIN lens_summary_selection ls ON l.id = ls.lens_id
		 WHERE ls.summary_id = ?
		 ORDER BY l.name ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("summary: check lens refs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	return scanLensNames(rows)
}

// scanLensNames scans a rows cursor of lens names into a string
// slice.
func scanLensNames(rows *sql.Rows) ([]string, error) {
	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan lens name: %w", err)
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// --- Private helpers ---

// getLens fetches a single lens by ID (without selections).
func (s *Store) getLens(ctx context.Context, id int64) (domain.Lens, error) {
	var l domain.Lens
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at, updated_at
		 FROM lens WHERE id = ?`, id,
	).Scan(
		&l.ID, &l.Name,
		&l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return domain.Lens{}, fmt.Errorf("lens: get %d: %w", id, err)
	}
	return l, nil
}

// getLensSummaries fetches summary selections for a lens.
func (s *Store) getLensSummaries(ctx context.Context, lensID int64) ([]domain.LensSummaryItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT summary_id, sort_order, is_master
		 FROM lens_summary_selection
		 WHERE lens_id = ?
		 ORDER BY sort_order ASC`,
		lensID,
	)
	if err != nil {
		return nil, fmt.Errorf("lens: get summaries: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	items := make([]domain.LensSummaryItem, 0)
	for rows.Next() {
		var item domain.LensSummaryItem
		var isMaster int
		if err := rows.Scan(&item.SummaryID, &item.SortOrder, &isMaster); err != nil {
			return nil, fmt.Errorf("lens: scan summary: %w", err)
		}
		item.IsMaster = isMaster != 0
		items = append(items, item)
	}
	return items, rows.Err()
}

// setLensSelections is a generic helper that replaces all selection
// rows for a lens within a transaction. It first deletes all existing
// rows using deleteSQL, then inserts count new rows using insertSQL
// with args provided by the argsFn callback.
func (s *Store) setLensSelections(
	ctx context.Context,
	lensID int64,
	deleteSQL, insertSQL string,
	argsFn func(i int) []interface{},
	count int,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("lens: set selections: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, deleteSQL, lensID); err != nil {
		return fmt.Errorf("lens: delete selections: %w", err)
	}

	for i := 0; i < count; i++ {
		if _, err := tx.ExecContext(ctx, insertSQL, argsFn(i)...); err != nil {
			return fmt.Errorf("lens: insert selection %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// getLensWorkHistory fetches work history selections for a lens.
func (s *Store) getLensWorkHistory(ctx context.Context, lensID int64) ([]domain.LensWorkHistoryItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT work_history_id, sort_order
		 FROM lens_work_history_selection
		 WHERE lens_id = ?
		 ORDER BY sort_order ASC`,
		lensID,
	)
	if err != nil {
		return nil, fmt.Errorf("lens: get work history: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	items := make([]domain.LensWorkHistoryItem, 0)
	for rows.Next() {
		var item domain.LensWorkHistoryItem
		if err := rows.Scan(&item.WorkHistoryID, &item.SortOrder); err != nil {
			return nil, fmt.Errorf("lens: scan work history: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// getLensBullets fetches bullet selections for a lens.
func (s *Store) getLensBullets(ctx context.Context, lensID int64) ([]domain.LensBulletItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT bullet_id, sort_order
		 FROM lens_bullet_selection
		 WHERE lens_id = ?
		 ORDER BY sort_order ASC`,
		lensID,
	)
	if err != nil {
		return nil, fmt.Errorf("lens: get bullets: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	items := make([]domain.LensBulletItem, 0)
	for rows.Next() {
		var item domain.LensBulletItem
		if err := rows.Scan(&item.BulletID, &item.SortOrder); err != nil {
			return nil, fmt.Errorf("lens: scan bullet: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// getLensSkills fetches skill selections for a lens.
func (s *Store) getLensSkills(ctx context.Context, lensID int64) ([]domain.LensSkillItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT skill_id, custom_sort_order
		 FROM lens_skill_selection
		 WHERE lens_id = ?
		 ORDER BY skill_id ASC`,
		lensID,
	)
	if err != nil {
		return nil, fmt.Errorf("lens: get skills: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	items := make([]domain.LensSkillItem, 0)
	for rows.Next() {
		var item domain.LensSkillItem
		var customSort sql.NullInt64
		if err := rows.Scan(&item.SkillID, &customSort); err != nil {
			return nil, fmt.Errorf("lens: scan skill: %w", err)
		}
		if customSort.Valid {
			cs := int(customSort.Int64)
			item.CustomSortOrder = &cs
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// getLensAcademics fetches academic selections for a lens.
func (s *Store) getLensAcademics(ctx context.Context, lensID int64) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT academic_id
		 FROM lens_academic_selection
		 WHERE lens_id = ?
		 ORDER BY academic_id ASC`,
		lensID,
	)
	if err != nil {
		return nil, fmt.Errorf("lens: get academics: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("lens: scan academic: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// getLensCerts fetches cert selections for a lens.
func (s *Store) getLensCerts(ctx context.Context, lensID int64) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT cert_id
		 FROM lens_cert_selection
		 WHERE lens_id = ?
		 ORDER BY cert_id ASC`,
		lensID,
	)
	if err != nil {
		return nil, fmt.Errorf("lens: get certs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("lens: scan cert: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// getLensDescriptors fetches descriptor selections for a lens.
func (s *Store) getLensDescriptors(ctx context.Context, lensID int64) ([]domain.LensDescriptorItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT descriptor_id, sort_order
		 FROM lens_descriptor_selection
		 WHERE lens_id = ?
		 ORDER BY sort_order ASC`,
		lensID,
	)
	if err != nil {
		return nil, fmt.Errorf("lens: get descriptors: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	items := make([]domain.LensDescriptorItem, 0)
	for rows.Next() {
		var item domain.LensDescriptorItem
		if err := rows.Scan(&item.DescriptorID, &item.SortOrder); err != nil {
			return nil, fmt.Errorf("lens: scan descriptor: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// getLensCoreExpertise fetches core expertise selections for a lens.
func (s *Store) getLensCoreExpertise(ctx context.Context, lensID int64) ([]domain.LensCoreExpertiseItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT core_expertise_id, sort_order
		 FROM lens_core_expertise_selection
		 WHERE lens_id = ?
		 ORDER BY sort_order ASC`,
		lensID,
	)
	if err != nil {
		return nil, fmt.Errorf("lens: get core expertise: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	items := make([]domain.LensCoreExpertiseItem, 0)
	for rows.Next() {
		var item domain.LensCoreExpertiseItem
		if err := rows.Scan(&item.CoreExpertiseID, &item.SortOrder); err != nil {
			return nil, fmt.Errorf("lens: scan core expertise: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
