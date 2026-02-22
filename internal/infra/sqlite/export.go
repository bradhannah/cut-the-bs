package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// CreateExport creates an export record and returns it with the
// generated ID and timestamp.
func (s *Store) CreateExport(
	ctx context.Context,
	export domain.ResumeExport,
) (domain.ResumeExport, error) {
	query := `
		INSERT INTO resume_export (template_id, template_ref_id, file_path, summary_id, lens_id)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		export.TemplateID,
		export.TemplateRefID,
		export.FilePath,
		export.SummaryID,
		export.LensID,
	)
	if err != nil {
		return domain.ResumeExport{}, fmt.Errorf("create export: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.ResumeExport{}, fmt.Errorf("create export: last insert id: %w", err)
	}

	return s.GetExport(ctx, id)
}

// GetExport returns a single export record by ID.
func (s *Store) GetExport(
	ctx context.Context,
	id int64,
) (domain.ResumeExport, error) {
	query := `
		SELECT id, template_id, template_ref_id, file_path, summary_id, lens_id, generated_at
		FROM resume_export
		WHERE id = ?
	`

	var e domain.ResumeExport
	var summaryID sql.NullInt64
	var lensID sql.NullInt64
	var templateRefID sql.NullInt64

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID,
		&e.TemplateID,
		&templateRefID,
		&e.FilePath,
		&summaryID,
		&lensID,
		&e.GeneratedAt,
	)
	if err == sql.ErrNoRows {
		return domain.ResumeExport{}, fmt.Errorf("export %d not found", id)
	}
	if err != nil {
		return domain.ResumeExport{}, fmt.Errorf("get export: %w", err)
	}

	if summaryID.Valid {
		e.SummaryID = &summaryID.Int64
	}
	if lensID.Valid {
		e.LensID = &lensID.Int64
	}
	if templateRefID.Valid {
		e.TemplateRefID = &templateRefID.Int64
	}

	return e, nil
}

// ListExports returns all export records ordered by generated_at
// descending (most recent first).
func (s *Store) ListExports(
	ctx context.Context,
) ([]domain.ResumeExport, error) {
	query := `
		SELECT id, template_id, template_ref_id, file_path, summary_id, lens_id, generated_at
		FROM resume_export
		ORDER BY id DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list exports: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	exports := make([]domain.ResumeExport, 0)
	for rows.Next() {
		var e domain.ResumeExport
		var summaryID sql.NullInt64
		var lensID sql.NullInt64
		var templateRefID sql.NullInt64

		if err := rows.Scan(
			&e.ID,
			&e.TemplateID,
			&templateRefID,
			&e.FilePath,
			&summaryID,
			&lensID,
			&e.GeneratedAt,
		); err != nil {
			return nil, fmt.Errorf("list exports: scan: %w", err)
		}

		if summaryID.Valid {
			e.SummaryID = &summaryID.Int64
		}
		if lensID.Valid {
			e.LensID = &lensID.Int64
		}
		if templateRefID.Valid {
			e.TemplateRefID = &templateRefID.Int64
		}

		exports = append(exports, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list exports: rows: %w", err)
	}

	return exports, nil
}

// CreateExportSelections saves the content selection snapshot for
// an export. This records which work history entries, bullets,
// skills, academics, certifications, and descriptors were included.
func (s *Store) CreateExportSelections(
	ctx context.Context,
	exportID int64,
	req domain.ExportRequest,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("create export selections: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Work history entries.
	for _, whID := range req.WorkHistoryIDs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO export_work_history_selection (export_id, work_history_id) VALUES (?, ?)`,
			exportID, whID,
		)
		if err != nil {
			return fmt.Errorf("insert work history selection: %w", err)
		}
	}

	// Bullets.
	for _, bID := range req.BulletIDs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO export_bullet_selection (export_id, bullet_id) VALUES (?, ?)`,
			exportID, bID,
		)
		if err != nil {
			return fmt.Errorf("insert bullet selection: %w", err)
		}
	}

	// Skills (with optional custom sort order).
	for _, sID := range req.SkillIDs {
		var customSort *int
		if order, ok := req.SkillSortOverrides[sID]; ok {
			customSort = &order
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO export_skill_selection (export_id, skill_id, custom_sort_order) VALUES (?, ?, ?)`,
			exportID, sID, customSort,
		)
		if err != nil {
			return fmt.Errorf("insert skill selection: %w", err)
		}
	}

	// Academics.
	for _, aID := range req.AcademicIDs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO export_academic_selection (export_id, academic_id) VALUES (?, ?)`,
			exportID, aID,
		)
		if err != nil {
			return fmt.Errorf("insert academic selection: %w", err)
		}
	}

	// Certifications.
	for _, cID := range req.CertificationIDs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO export_cert_selection (export_id, cert_id) VALUES (?, ?)`,
			exportID, cID,
		)
		if err != nil {
			return fmt.Errorf("insert cert selection: %w", err)
		}
	}

	// Descriptors.
	for i, dID := range req.DescriptorIDs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO export_descriptor_selection (export_id, descriptor_id, sort_order) VALUES (?, ?, ?)`,
			exportID, dID, i,
		)
		if err != nil {
			return fmt.Errorf("insert descriptor selection: %w", err)
		}
	}

	return tx.Commit()
}

// GetExportWorkHistoryIDs returns the work history entry IDs
// selected for an export.
func (s *Store) GetExportWorkHistoryIDs(
	ctx context.Context,
	exportID int64,
) ([]int64, error) {
	return s.getExportIDs(ctx,
		`SELECT work_history_id FROM export_work_history_selection WHERE export_id = ?`,
		exportID,
	)
}

// GetExportBulletIDs returns the bullet IDs selected for an export.
func (s *Store) GetExportBulletIDs(
	ctx context.Context,
	exportID int64,
) ([]int64, error) {
	return s.getExportIDs(ctx,
		`SELECT bullet_id FROM export_bullet_selection WHERE export_id = ?`,
		exportID,
	)
}

// GetExportSkillIDs returns the skill IDs selected for an export.
func (s *Store) GetExportSkillIDs(
	ctx context.Context,
	exportID int64,
) ([]int64, error) {
	return s.getExportIDs(ctx,
		`SELECT skill_id FROM export_skill_selection WHERE export_id = ?`,
		exportID,
	)
}

// GetExportAcademicIDs returns the academic credential IDs selected
// for an export.
func (s *Store) GetExportAcademicIDs(
	ctx context.Context,
	exportID int64,
) ([]int64, error) {
	return s.getExportIDs(ctx,
		`SELECT academic_id FROM export_academic_selection WHERE export_id = ?`,
		exportID,
	)
}

// GetExportCertIDs returns the certification IDs selected for an
// export.
func (s *Store) GetExportCertIDs(
	ctx context.Context,
	exportID int64,
) ([]int64, error) {
	return s.getExportIDs(ctx,
		`SELECT cert_id FROM export_cert_selection WHERE export_id = ?`,
		exportID,
	)
}

// GetExportDescriptorIDs returns the descriptor IDs selected for
// an export, ordered by sort_order.
func (s *Store) GetExportDescriptorIDs(
	ctx context.Context,
	exportID int64,
) ([]int64, error) {
	return s.getExportIDs(ctx,
		`SELECT descriptor_id FROM export_descriptor_selection WHERE export_id = ? ORDER BY sort_order`,
		exportID,
	)
}

// getExportIDs is a helper to scan a single-column int64 query.
func (s *Store) getExportIDs(
	ctx context.Context,
	query string,
	exportID int64,
) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx, query, exportID)
	if err != nil {
		return nil, fmt.Errorf("get export ids: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("get export ids: scan: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get export ids: rows: %w", err)
	}

	return ids, nil
}
