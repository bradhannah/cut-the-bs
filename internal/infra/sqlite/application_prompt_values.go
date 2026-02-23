package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// GetApplicationPromptValues returns saved prompt/variable values for a
// specific application + template pair.
func (s *Store) GetApplicationPromptValues(
	ctx context.Context,
	applicationID int64,
	templateID int64,
) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT field_key, field_value
		 FROM application_prompt_value
		 WHERE application_id = ? AND template_id = ?
		 ORDER BY id ASC`,
		applicationID,
		templateID,
	)
	if err != nil {
		return nil, fmt.Errorf("application_prompt_value: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	values := make(map[string]string)
	for rows.Next() {
		var key string
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("application_prompt_value: scan: %w", err)
		}
		values[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("application_prompt_value: rows: %w", err)
	}

	return values, nil
}

// SaveApplicationPromptValues replaces saved prompt/variable values for a
// specific application + template pair.
func (s *Store) SaveApplicationPromptValues(
	ctx context.Context,
	applicationID int64,
	templateID int64,
	values map[string]string,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("application_prompt_value: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM application_prompt_value WHERE application_id = ? AND template_id = ?`,
		applicationID,
		templateID,
	); err != nil {
		return fmt.Errorf("application_prompt_value: delete existing: %w", err)
	}

	for key, value := range values {
		if key == "" || value == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO application_prompt_value
			 (application_id, template_id, field_key, field_value)
			 VALUES (?, ?, ?, ?)`,
			applicationID,
			templateID,
			key,
			value,
		); err != nil {
			return fmt.Errorf("application_prompt_value: insert %q: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("application_prompt_value: commit: %w", err)
	}

	return nil
}

// exportApplicationPromptValues returns all saved application prompt values.
func (s *Store) exportApplicationPromptValues(
	ctx context.Context,
) ([]domain.ApplicationPromptValue, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, application_id, template_id, field_key, field_value, created_at, updated_at
		 FROM application_prompt_value
		 ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("application_prompt_value: export: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	items := make([]domain.ApplicationPromptValue, 0)
	for rows.Next() {
		var item domain.ApplicationPromptValue
		if err := rows.Scan(
			&item.ID,
			&item.ApplicationID,
			&item.TemplateID,
			&item.FieldKey,
			&item.FieldValue,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("application_prompt_value: scan export: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("application_prompt_value: export rows: %w", err)
	}

	return items, nil
}

func (s *Store) importApplicationPromptValue(
	ctx context.Context,
	tx *sql.Tx,
	item domain.ApplicationPromptValue,
) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO application_prompt_value
		 (id, application_id, template_id, field_key, field_value, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.ID,
		item.ApplicationID,
		item.TemplateID,
		item.FieldKey,
		item.FieldValue,
		item.CreatedAt,
		item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("import: application_prompt_value %d: %w", item.ID, err)
	}
	return nil
}
