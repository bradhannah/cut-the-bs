package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Job Applications ---

// ListApplications returns all job applications ordered by
// date_applied descending (most recent first).
func (s *Store) ListApplications(ctx context.Context) ([]domain.JobApplication, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_name, position_title,
		        COALESCE(job_posting_url, ''), COALESCE(application_url, ''), COALESCE(research_url, ''),
		        date_applied,
		        status, COALESCE(fit_indicator, ''),
		        resume_export_id, cover_letter_template_id, cover_letter_latest_export_id,
		        COALESCE(notes, ''),
		        created_at, updated_at
		 FROM job_application
		 ORDER BY date_applied DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("job_application: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	apps := make([]domain.JobApplication, 0)
	for rows.Next() {
		app, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, rows.Err()
}

// SearchApplications searches by company name or position title
// (case-insensitive LIKE).
func (s *Store) SearchApplications(ctx context.Context, query string) ([]domain.JobApplication, error) {
	like := "%" + query + "%"
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_name, position_title,
		        COALESCE(job_posting_url, ''), COALESCE(application_url, ''), COALESCE(research_url, ''),
		        date_applied,
		        status, COALESCE(fit_indicator, ''),
		        resume_export_id, cover_letter_template_id, cover_letter_latest_export_id,
		        COALESCE(notes, ''),
		        created_at, updated_at
		 FROM job_application
		 WHERE company_name LIKE ? OR position_title LIKE ?
		 ORDER BY date_applied DESC, id DESC`,
		like, like,
	)
	if err != nil {
		return nil, fmt.Errorf("job_application: search: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	apps := make([]domain.JobApplication, 0)
	for rows.Next() {
		app, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, rows.Err()
}

// GetApplication returns a single job application by ID.
func (s *Store) GetApplication(ctx context.Context, id int64) (domain.JobApplication, error) {
	return s.getApplication(ctx, id)
}

// getApplication is the internal helper for fetching a single
// application by ID.
func (s *Store) getApplication(ctx context.Context, id int64) (domain.JobApplication, error) {
	var app domain.JobApplication
	var fitIndicator sql.NullString
	var notes sql.NullString
	var resumeExportID sql.NullInt64
	var coverLetterTemplateID sql.NullInt64
	var coverLetterLatestExportID sql.NullInt64

	err := s.db.QueryRowContext(ctx,
		`SELECT id, company_name, position_title,
		        COALESCE(job_posting_url, ''), COALESCE(application_url, ''), COALESCE(research_url, ''),
		        date_applied,
		        status, fit_indicator,
		        resume_export_id, cover_letter_template_id, cover_letter_latest_export_id,
		        notes,
		        created_at, updated_at
		 FROM job_application WHERE id = ?`, id,
	).Scan(
		&app.ID, &app.CompanyName, &app.PositionTitle,
		&app.JobPostingURL, &app.ApplicationURL, &app.ResearchURL,
		&app.DateApplied,
		&app.Status, &fitIndicator,
		&resumeExportID, &coverLetterTemplateID, &coverLetterLatestExportID,
		&notes,
		&app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		return domain.JobApplication{}, fmt.Errorf("job_application: get %d: %w", id, err)
	}
	if fitIndicator.Valid {
		app.FitIndicator = fitIndicator.String
	}
	if notes.Valid {
		app.Notes = notes.String
	}
	if resumeExportID.Valid {
		v := resumeExportID.Int64
		app.ResumeExportID = &v
	}
	if coverLetterTemplateID.Valid {
		v := coverLetterTemplateID.Int64
		app.CoverLetterTemplateID = &v
	}
	if coverLetterLatestExportID.Valid {
		v := coverLetterLatestExportID.Int64
		app.CoverLetterLatestExportID = &v
	}

	return app, nil
}

// CreateApplication creates a new job application with initial
// status "Applied".
func (s *Store) CreateApplication(ctx context.Context, input domain.ApplicationInput) (domain.JobApplication, error) {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO job_application
		 (company_name, position_title, job_posting_url, application_url, research_url,
		  date_applied, status,
		  fit_indicator, resume_export_id, cover_letter_template_id, cover_letter_latest_export_id, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.CompanyName, input.PositionTitle,
		input.JobPostingURL, input.ApplicationURL, input.ResearchURL,
		input.DateApplied,
		domain.StatusApplied,
		nullableString(input.FitIndicator),
		nullableInt64Ptr(input.ResumeExportID),
		nullableInt64Ptr(input.CoverLetterTemplateID),
		nullableInt64Ptr(input.CoverLetterLatestExportID),
		nullableString(input.Notes),
	)
	if err != nil {
		return domain.JobApplication{}, fmt.Errorf("job_application: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.JobApplication{}, fmt.Errorf("job_application: last insert id: %w", err)
	}

	return s.getApplication(ctx, id)
}

// UpdateApplication updates application fields (not status or fit).
func (s *Store) UpdateApplication(ctx context.Context, id int64, input domain.ApplicationInput) (domain.JobApplication, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE job_application
		 SET company_name = ?, position_title = ?,
		     job_posting_url = ?, application_url = ?, research_url = ?,
		     date_applied = ?,
		     fit_indicator = ?, resume_export_id = ?,
		     cover_letter_template_id = ?,
		     cover_letter_latest_export_id = ?, notes = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.CompanyName, input.PositionTitle,
		input.JobPostingURL, input.ApplicationURL, input.ResearchURL,
		input.DateApplied,
		nullableString(input.FitIndicator),
		nullableInt64Ptr(input.ResumeExportID),
		nullableInt64Ptr(input.CoverLetterTemplateID),
		nullableInt64Ptr(input.CoverLetterLatestExportID),
		nullableString(input.Notes),
		id,
	)
	if err != nil {
		return domain.JobApplication{}, fmt.Errorf("job_application: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.JobApplication{}, fmt.Errorf("job_application: not found: %d", id)
	}

	return s.getApplication(ctx, id)
}

// UpdateApplicationStatus changes the status of an application and
// records the transition in status_history.
func (s *Store) UpdateApplicationStatus(ctx context.Context, id int64, newStatus string) (domain.JobApplication, error) {
	// Get current status.
	current, err := s.getApplication(ctx, id)
	if err != nil {
		return domain.JobApplication{}, fmt.Errorf("job_application: not found: %d", id)
	}

	oldStatus := current.Status

	// Update status.
	_, err = s.db.ExecContext(ctx,
		`UPDATE job_application
		 SET status = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		newStatus, id,
	)
	if err != nil {
		return domain.JobApplication{}, fmt.Errorf("job_application: update status: %w", err)
	}

	// Record transition.
	_, err = s.CreateStatusChange(ctx, domain.StatusChange{
		ApplicationID: id,
		FromStatus:    oldStatus,
		ToStatus:      newStatus,
	})
	if err != nil {
		return domain.JobApplication{}, fmt.Errorf("job_application: record history: %w", err)
	}

	return s.getApplication(ctx, id)
}

// UpdateApplicationFit updates the fit indicator of an application.
func (s *Store) UpdateApplicationFit(ctx context.Context, id int64, fitIndicator string) (domain.JobApplication, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE job_application
		 SET fit_indicator = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		nullableString(fitIndicator), id,
	)
	if err != nil {
		return domain.JobApplication{}, fmt.Errorf("job_application: update fit: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.JobApplication{}, fmt.Errorf("job_application: not found: %d", id)
	}

	return s.getApplication(ctx, id)
}

// DeleteApplication deletes a job application and its status history
// (CASCADE).
func (s *Store) DeleteApplication(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM job_application WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("job_application: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("job_application: not found: %d", id)
	}

	return nil
}

// --- Status History ---

// GetApplicationHistory returns the status change history for an
// application, ordered chronologically.
func (s *Store) GetApplicationHistory(ctx context.Context, applicationID int64) ([]domain.StatusChange, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, application_id, COALESCE(from_status, ''),
		        to_status, changed_at
		 FROM status_history
		 WHERE application_id = ?
		 ORDER BY changed_at ASC, id ASC`,
		applicationID,
	)
	if err != nil {
		return nil, fmt.Errorf("status_history: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	changes := make([]domain.StatusChange, 0)
	for rows.Next() {
		var sc domain.StatusChange
		if err := rows.Scan(
			&sc.ID, &sc.ApplicationID, &sc.FromStatus,
			&sc.ToStatus, &sc.ChangedAt,
		); err != nil {
			return nil, fmt.Errorf("status_history: scan: %w", err)
		}
		changes = append(changes, sc)
	}

	return changes, rows.Err()
}

// CreateStatusChange records a status transition.
func (s *Store) CreateStatusChange(ctx context.Context, change domain.StatusChange) (domain.StatusChange, error) {
	var fromStatus interface{}
	if change.FromStatus != "" {
		fromStatus = change.FromStatus
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO status_history
		 (application_id, from_status, to_status)
		 VALUES (?, ?, ?)`,
		change.ApplicationID, fromStatus, change.ToStatus,
	)
	if err != nil {
		return domain.StatusChange{}, fmt.Errorf("status_history: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.StatusChange{}, fmt.Errorf("status_history: last insert id: %w", err)
	}

	var sc domain.StatusChange
	err = s.db.QueryRowContext(ctx,
		`SELECT id, application_id, COALESCE(from_status, ''),
		        to_status, changed_at
		 FROM status_history WHERE id = ?`, id,
	).Scan(&sc.ID, &sc.ApplicationID, &sc.FromStatus,
		&sc.ToStatus, &sc.ChangedAt)
	if err != nil {
		return domain.StatusChange{}, fmt.Errorf("status_history: get: %w", err)
	}

	return sc, nil
}

// --- helpers ---

// scanApplication scans a row from a job_application query.
type applicationScanner interface {
	Scan(dest ...interface{}) error
}

func scanApplication(row applicationScanner) (domain.JobApplication, error) {
	var app domain.JobApplication
	var resumeExportID sql.NullInt64
	var coverLetterTemplateID sql.NullInt64
	var coverLetterLatestExportID sql.NullInt64

	if err := row.Scan(
		&app.ID, &app.CompanyName, &app.PositionTitle,
		&app.JobPostingURL, &app.ApplicationURL, &app.ResearchURL,
		&app.DateApplied,
		&app.Status, &app.FitIndicator,
		&resumeExportID, &coverLetterTemplateID, &coverLetterLatestExportID,
		&app.Notes,
		&app.CreatedAt, &app.UpdatedAt,
	); err != nil {
		return domain.JobApplication{}, fmt.Errorf("job_application: scan: %w", err)
	}

	if resumeExportID.Valid {
		v := resumeExportID.Int64
		app.ResumeExportID = &v
	}
	if coverLetterTemplateID.Valid {
		v := coverLetterTemplateID.Int64
		app.CoverLetterTemplateID = &v
	}
	if coverLetterLatestExportID.Valid {
		v := coverLetterLatestExportID.Int64
		app.CoverLetterLatestExportID = &v
	}

	return app, nil
}

// nullableString returns nil for empty strings, otherwise the value.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nullableInt64Ptr returns nil for nil pointers, otherwise the value.
func nullableInt64Ptr(p *int64) interface{} {
	if p == nil {
		return nil
	}
	return *p
}
