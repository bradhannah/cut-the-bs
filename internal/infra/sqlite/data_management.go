package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cut-the-bs/internal/domain"
)

// ExportAllData queries every table and assembles a complete
// ExportData envelope. This is used for full JSON backup.
func (s *Store) ExportAllData(ctx context.Context) (domain.ExportData, error) {
	data := domain.ExportData{
		SchemaVersion: currentVersion,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	var err error

	// Profile
	data.Profile, err = s.GetProfile(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: profile: %w", err)
	}

	// Profile links
	data.ProfileLinks, err = s.ListProfileLinks(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: profile_links: %w", err)
	}

	// Work history (includes bullets)
	data.WorkHistory, err = s.ListWorkHistory(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: work_history: %w", err)
	}

	// Skill categories
	data.Categories, err = s.ListSkillCategories(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: skill_categories: %w", err)
	}

	// Skills
	data.Skills, err = s.ListSkills(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: skills: %w", err)
	}

	// Academics
	data.Academics, err = s.ListAcademicCredentials(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: academics: %w", err)
	}

	// Certifications
	data.Certs, err = s.ListCertifications(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: certifications: %w", err)
	}

	// Summaries
	data.Summaries, err = s.ListSummaries(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: summaries: %w", err)
	}

	// Descriptors
	data.Descriptors, err = s.ListDescriptors(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: descriptors: %w", err)
	}

	// Core expertise
	data.CoreExpertise, err = s.ListCoreExpertise(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: core_expertise: %w", err)
	}

	// Lenses (with full detail)
	data.Lenses, err = s.exportLenses(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: lenses: %w", err)
	}

	// Exports
	data.Exports, err = s.ListExports(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: exports: %w", err)
	}

	// Cover letters
	data.CoverLetters, err = s.ListCoverLetters(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: cover_letters: %w", err)
	}

	// Applications
	data.Applications, err = s.ListApplications(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: applications: %w", err)
	}

	// Status changes (all across all applications)
	data.StatusChanges, err = s.exportAllStatusChanges(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: status_changes: %w", err)
	}

	// Skill lens tags
	data.SkillLensTags, err = s.exportSkillLensTags(ctx)
	if err != nil {
		return domain.ExportData{}, fmt.Errorf("export: skill_lens_tags: %w", err)
	}

	return data, nil
}

// exportLenses returns all lenses with their full detail (selections).
func (s *Store) exportLenses(ctx context.Context) ([]domain.LensDetail, error) {
	lenses, err := s.ListLenses(ctx)
	if err != nil {
		return nil, err
	}

	details := make([]domain.LensDetail, 0, len(lenses))
	for _, l := range lenses {
		detail, err := s.GetLens(ctx, l.ID)
		if err != nil {
			return nil, fmt.Errorf("lens %d: %w", l.ID, err)
		}
		details = append(details, detail)
	}

	return details, nil
}

// exportAllStatusChanges returns all status changes across all
// applications.
func (s *Store) exportAllStatusChanges(ctx context.Context) ([]domain.StatusChange, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, application_id, from_status, to_status, changed_at
		 FROM status_history
		 ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("status_history: export: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	changes := make([]domain.StatusChange, 0)
	for rows.Next() {
		var sc domain.StatusChange
		var fromStatus sql.NullString
		if err := rows.Scan(
			&sc.ID, &sc.ApplicationID, &fromStatus,
			&sc.ToStatus, &sc.ChangedAt,
		); err != nil {
			return nil, fmt.Errorf("status_history: scan: %w", err)
		}
		if fromStatus.Valid {
			sc.FromStatus = fromStatus.String
		}
		changes = append(changes, sc)
	}

	return changes, rows.Err()
}

// exportSkillLensTags returns all skill-lens tag associations.
func (s *Store) exportSkillLensTags(ctx context.Context) ([]domain.SkillLensTag, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT skill_id, lens_id FROM skill_lens_tag ORDER BY skill_id, lens_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("skill_lens_tag: export: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	tags := make([]domain.SkillLensTag, 0)
	for rows.Next() {
		var t domain.SkillLensTag
		if err := rows.Scan(&t.SkillID, &t.LensID); err != nil {
			return nil, fmt.Errorf("skill_lens_tag: scan: %w", err)
		}
		tags = append(tags, t)
	}

	return tags, rows.Err()
}

// ImportAllData replaces all data with the contents of the provided
// ExportData. All tables are truncated and data is inserted within a
// single transaction. Foreign key checks are temporarily disabled to
// allow insertion in any order.
func (s *Store) ImportAllData(ctx context.Context, data domain.ExportData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("import: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Temporarily disable foreign keys for bulk import.
	if _, err := tx.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("import: disable FK: %w", err)
	}

	// Truncate all tables in reverse dependency order.
	tables := []string{
		"status_history",
		"job_application",
		"cover_letter",
		"export_descriptor_selection",
		"export_cert_selection",
		"export_academic_selection",
		"export_skill_selection",
		"export_bullet_selection",
		"export_work_history_selection",
		"resume_export",
		"skill_lens_tag",
		"lens_core_expertise_selection",
		"lens_descriptor_selection",
		"lens_summary_selection",
		"lens_cert_selection",
		"lens_academic_selection",
		"lens_skill_selection",
		"lens_bullet_selection",
		"lens_work_history_selection",
		"lens",
		"core_expertise",
		"role_descriptor",
		"professional_summary",
		"certification",
		"academic_credential",
		"skill",
		"skill_category",
		"achievement_bullet",
		"work_history_entry",
		"profile_link",
		"user_profile",
	}

	for _, table := range tables {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("import: truncate %s: %w", table, err)
		}
	}

	// Insert profile
	if err := s.importProfile(ctx, tx, data.Profile); err != nil {
		return err
	}

	// Insert profile links
	for _, pl := range data.ProfileLinks {
		if err := s.importProfileLink(ctx, tx, pl); err != nil {
			return err
		}
	}

	// Insert work history entries and bullets
	for _, wh := range data.WorkHistory {
		if err := s.importWorkHistory(ctx, tx, wh); err != nil {
			return err
		}
	}

	// Insert skill categories
	for _, cat := range data.Categories {
		if err := s.importSkillCategory(ctx, tx, cat); err != nil {
			return err
		}
	}

	// Insert skills
	for _, sk := range data.Skills {
		if err := s.importSkill(ctx, tx, sk); err != nil {
			return err
		}
	}

	// Insert academics
	for _, ac := range data.Academics {
		if err := s.importAcademic(ctx, tx, ac); err != nil {
			return err
		}
	}

	// Insert certifications
	for _, cert := range data.Certs {
		if err := s.importCertification(ctx, tx, cert); err != nil {
			return err
		}
	}

	// Insert summaries
	for _, sum := range data.Summaries {
		if err := s.importSummary(ctx, tx, sum); err != nil {
			return err
		}
	}

	// Insert descriptors
	for _, desc := range data.Descriptors {
		if err := s.importDescriptor(ctx, tx, desc); err != nil {
			return err
		}
	}

	// Insert core expertise
	for _, ce := range data.CoreExpertise {
		if err := s.importCoreExpertise(ctx, tx, ce); err != nil {
			return err
		}
	}

	// Insert lenses and their selections
	for _, lens := range data.Lenses {
		if err := s.importLens(ctx, tx, lens); err != nil {
			return err
		}
	}

	// Insert skill lens tags
	for _, tag := range data.SkillLensTags {
		if err := s.importSkillLensTag(ctx, tx, tag); err != nil {
			return err
		}
	}

	// Insert exports
	for _, exp := range data.Exports {
		if err := s.importExport(ctx, tx, exp); err != nil {
			return err
		}
	}

	// Insert cover letters
	for _, cl := range data.CoverLetters {
		if err := s.importCoverLetter(ctx, tx, cl); err != nil {
			return err
		}
	}

	// Insert applications
	for _, app := range data.Applications {
		if err := s.importApplication(ctx, tx, app); err != nil {
			return err
		}
	}

	// Insert status changes
	for _, sc := range data.StatusChanges {
		if err := s.importStatusChange(ctx, tx, sc); err != nil {
			return err
		}
	}

	// Re-enable foreign keys.
	if _, err := tx.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("import: re-enable FK: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("import: commit: %w", err)
	}

	return nil
}

// --- Import helper methods ---

func (s *Store) importProfile(ctx context.Context, tx *sql.Tx, p domain.UserProfile) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_profile (id, full_name, email, phone, location, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.FullName, p.Email, p.Phone, p.Location, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: user_profile: %w", err)
	}
	return nil
}

func (s *Store) importProfileLink(ctx context.Context, tx *sql.Tx, pl domain.ProfileLink) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO profile_link (id, label, url, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		pl.ID, pl.Label, pl.URL, pl.SortOrder, pl.CreatedAt, pl.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: profile_link: %w", err)
	}
	return nil
}

func (s *Store) importWorkHistory(ctx context.Context, tx *sql.Tx, wh domain.WorkHistoryEntry) error {
	var endDate interface{}
	if wh.EndDate != "" {
		endDate = wh.EndDate
	}
	var granEnd interface{}
	if wh.DateGranularityEnd != "" {
		granEnd = wh.DateGranularityEnd
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO work_history_entry
		 (id, employer_name, job_title, summary, start_date, end_date,
		  date_granularity_start, date_granularity_end, sort_order,
		  created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		wh.ID, wh.EmployerName, wh.JobTitle, wh.Summary, wh.StartDate, endDate,
		wh.DateGranularityStart, granEnd, wh.SortOrder,
		wh.CreatedAt, wh.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: work_history_entry %d: %w", wh.ID, err)
	}

	for _, b := range wh.Bullets {
		var bulletType interface{}
		if b.BulletType != "" {
			bulletType = b.BulletType
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO achievement_bullet
			 (id, work_history_id, text, bullet_type, sort_order, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			b.ID, b.WorkHistoryID, b.Text, bulletType, b.SortOrder, b.CreatedAt, b.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("import: achievement_bullet %d: %w", b.ID, err)
		}
	}

	return nil
}

func (s *Store) importSkillCategory(ctx context.Context, tx *sql.Tx, cat domain.SkillCategory) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO skill_category (id, name, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		cat.ID, cat.Name, cat.SortOrder, cat.CreatedAt, cat.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: skill_category %d: %w", cat.ID, err)
	}
	return nil
}

func (s *Store) importSkill(ctx context.Context, tx *sql.Tx, sk domain.Skill) error {
	isLegacy := 0
	if sk.IsLegacy {
		isLegacy = 1
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO skill (id, name, category_id, competence_level, is_legacy, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sk.ID, sk.Name, sk.CategoryID, sk.CompetenceLevel, isLegacy,
		sk.CreatedAt, sk.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: skill %d: %w", sk.ID, err)
	}
	return nil
}

func (s *Store) importAcademic(ctx context.Context, tx *sql.Tx, ac domain.AcademicCredential) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO academic_credential
		 (id, institution, credential_type, field_of_study,
		  completion_date, date_granularity, sort_order,
		  created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ac.ID, ac.Institution, ac.CredentialType, ac.FieldOfStudy,
		ac.CompletionDate, ac.DateGranularity, ac.SortOrder,
		ac.CreatedAt, ac.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: academic_credential %d: %w", ac.ID, err)
	}
	return nil
}

func (s *Store) importCertification(ctx context.Context, tx *sql.Tx, cert domain.Certification) error {
	var exp interface{}
	if cert.ExpirationDate != "" {
		exp = cert.ExpirationDate
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO certification
		 (id, name, issuing_body, date_earned, expiration_date,
		  sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		cert.ID, cert.Name, cert.IssuingBody, cert.DateEarned, exp,
		cert.SortOrder, cert.CreatedAt, cert.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: certification %d: %w", cert.ID, err)
	}
	return nil
}

func (s *Store) importSummary(ctx context.Context, tx *sql.Tx, sum domain.ProfessionalSummary) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO professional_summary (id, label, body_text, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		sum.ID, sum.Label, sum.BodyText, sum.CreatedAt, sum.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: professional_summary %d: %w", sum.ID, err)
	}
	return nil
}

func (s *Store) importDescriptor(ctx context.Context, tx *sql.Tx, desc domain.RoleDescriptor) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO role_descriptor (id, title, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		desc.ID, desc.Title, desc.SortOrder, desc.CreatedAt, desc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: role_descriptor %d: %w", desc.ID, err)
	}
	return nil
}

func (s *Store) importCoreExpertise(ctx context.Context, tx *sql.Tx, ce domain.CoreExpertise) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO core_expertise (id, label, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		ce.ID, ce.Label, ce.SortOrder, ce.CreatedAt, ce.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: core_expertise %d: %w", ce.ID, err)
	}
	return nil
}

func (s *Store) importLens(ctx context.Context, tx *sql.Tx, lens domain.LensDetail) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO lens (id, name, created_at, updated_at)
		 VALUES (?, ?, ?, ?)`,
		lens.ID, lens.Name, lens.CreatedAt, lens.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: lens %d: %w", lens.ID, err)
	}

	// Summary selections
	for _, sum := range lens.Summaries {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO lens_summary_selection (lens_id, summary_id, sort_order)
			 VALUES (?, ?, ?)`,
			lens.ID, sum.SummaryID, sum.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("import: lens_summary %d: %w", lens.ID, err)
		}
	}

	// Work history selections
	for _, wh := range lens.WorkHistory {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO lens_work_history_selection (lens_id, work_history_id, sort_order)
			 VALUES (?, ?, ?)`,
			lens.ID, wh.WorkHistoryID, wh.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("import: lens_work_history %d: %w", lens.ID, err)
		}
	}

	// Bullet selections
	for _, b := range lens.Bullets {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO lens_bullet_selection (lens_id, bullet_id, sort_order)
			 VALUES (?, ?, ?)`,
			lens.ID, b.BulletID, b.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("import: lens_bullet %d: %w", lens.ID, err)
		}
	}

	// Skill selections
	for _, sk := range lens.Skills {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO lens_skill_selection (lens_id, skill_id, custom_sort_order)
			 VALUES (?, ?, ?)`,
			lens.ID, sk.SkillID, sk.CustomSortOrder,
		)
		if err != nil {
			return fmt.Errorf("import: lens_skill %d: %w", lens.ID, err)
		}
	}

	// Academic selections
	for _, aid := range lens.AcademicIDs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO lens_academic_selection (lens_id, academic_id) VALUES (?, ?)`,
			lens.ID, aid,
		)
		if err != nil {
			return fmt.Errorf("import: lens_academic %d: %w", lens.ID, err)
		}
	}

	// Cert selections
	for _, cid := range lens.CertIDs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO lens_cert_selection (lens_id, cert_id) VALUES (?, ?)`,
			lens.ID, cid,
		)
		if err != nil {
			return fmt.Errorf("import: lens_cert %d: %w", lens.ID, err)
		}
	}

	// Descriptor selections
	for _, d := range lens.Descriptors {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO lens_descriptor_selection (lens_id, descriptor_id, sort_order)
			 VALUES (?, ?, ?)`,
			lens.ID, d.DescriptorID, d.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("import: lens_descriptor %d: %w", lens.ID, err)
		}
	}

	// Core expertise selections
	for _, ce := range lens.CoreExpertise {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO lens_core_expertise_selection (lens_id, core_expertise_id, sort_order)
			 VALUES (?, ?, ?)`,
			lens.ID, ce.CoreExpertiseID, ce.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("import: lens_core_expertise %d: %w", lens.ID, err)
		}
	}

	return nil
}

func (s *Store) importSkillLensTag(ctx context.Context, tx *sql.Tx, tag domain.SkillLensTag) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO skill_lens_tag (skill_id, lens_id) VALUES (?, ?)`,
		tag.SkillID, tag.LensID,
	)
	if err != nil {
		return fmt.Errorf("import: skill_lens_tag: %w", err)
	}
	return nil
}

func (s *Store) importExport(ctx context.Context, tx *sql.Tx, exp domain.ResumeExport) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO resume_export (id, template_id, file_path, summary_id, lens_id, generated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		exp.ID, exp.TemplateID, exp.FilePath, exp.SummaryID, exp.LensID, exp.GeneratedAt,
	)
	if err != nil {
		return fmt.Errorf("import: resume_export %d: %w", exp.ID, err)
	}
	return nil
}

func (s *Store) importCoverLetter(ctx context.Context, tx *sql.Tx, cl domain.CoverLetter) error {
	var filePath interface{}
	if cl.FilePath != "" {
		filePath = cl.FilePath
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO cover_letter (id, title, body_text, file_path, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		cl.ID, cl.Title, cl.BodyText, filePath, cl.CreatedAt, cl.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: cover_letter %d: %w", cl.ID, err)
	}
	return nil
}

func (s *Store) importApplication(ctx context.Context, tx *sql.Tx, app domain.JobApplication) error {
	var fitIndicator interface{}
	if app.FitIndicator != "" {
		fitIndicator = app.FitIndicator
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO job_application
		 (id, company_name, position_title, date_applied, status,
		  fit_indicator, resume_export_id, cover_letter_id, notes,
		  created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		app.ID, app.CompanyName, app.PositionTitle, app.DateApplied,
		app.Status, fitIndicator, app.ResumeExportID, app.CoverLetterID,
		app.Notes, app.CreatedAt, app.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: job_application %d: %w", app.ID, err)
	}
	return nil
}

func (s *Store) importStatusChange(ctx context.Context, tx *sql.Tx, sc domain.StatusChange) error {
	var fromStatus interface{}
	if sc.FromStatus != "" {
		fromStatus = sc.FromStatus
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO status_history (id, application_id, from_status, to_status, changed_at)
		 VALUES (?, ?, ?, ?, ?)`,
		sc.ID, sc.ApplicationID, fromStatus, sc.ToStatus, sc.ChangedAt,
	)
	if err != nil {
		return fmt.Errorf("import: status_history %d: %w", sc.ID, err)
	}
	return nil
}
