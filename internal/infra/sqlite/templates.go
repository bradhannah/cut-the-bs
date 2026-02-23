package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// ListDocumentTemplates returns all document templates ordered by
// is_builtin DESC, name ASC. Built-in templates appear first.
func (s *Store) ListDocumentTemplates(ctx context.Context) ([]domain.DocumentTemplate, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description, template_type, is_builtin,
		        margin_top, margin_bottom, margin_left, margin_right,
		        created_at, updated_at
		 FROM document_template
		 ORDER BY is_builtin DESC, name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("template: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	templates := make([]domain.DocumentTemplate, 0)
	for rows.Next() {
		var t domain.DocumentTemplate
		var isBuiltin int
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.TemplateType, &isBuiltin,
			&t.MarginTop, &t.MarginBottom, &t.MarginLeft, &t.MarginRight,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("template: list scan: %w", err)
		}
		t.IsBuiltin = isBuiltin == 1
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("template: list rows: %w", err)
	}
	return templates, nil
}

// GetDocumentTemplate returns a template with all its elements.
func (s *Store) GetDocumentTemplate(ctx context.Context, id int64) (domain.TemplateDetail, error) {
	var detail domain.TemplateDetail
	var isBuiltin int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, template_type, is_builtin,
		        margin_top, margin_bottom, margin_left, margin_right,
		        created_at, updated_at
		 FROM document_template WHERE id = ?`, id,
	).Scan(
		&detail.ID, &detail.Name, &detail.Description, &detail.TemplateType, &isBuiltin,
		&detail.MarginTop, &detail.MarginBottom, &detail.MarginLeft, &detail.MarginRight,
		&detail.CreatedAt, &detail.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return detail, fmt.Errorf("template: not found: id=%d", id)
	}
	if err != nil {
		return detail, fmt.Errorf("template: get: %w", err)
	}
	detail.IsBuiltin = isBuiltin == 1

	// Fetch elements in a separate query (MaxOpenConns=1 pattern).
	elements, err := s.listTemplateElements(ctx, id)
	if err != nil {
		return detail, err
	}
	detail.Elements = elements

	return detail, nil
}

// listTemplateElements returns all elements for a template ordered
// by parent_id NULLS FIRST, then sort_order.
func (s *Store) listTemplateElements(ctx context.Context, templateID int64) ([]domain.TemplateElement, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, template_id, parent_id, element_type, config, sort_order,
		        created_at, updated_at
		 FROM template_element
		 WHERE template_id = ?
		 ORDER BY parent_id IS NOT NULL, parent_id, sort_order`,
		templateID,
	)
	if err != nil {
		return nil, fmt.Errorf("template_element: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	elements := make([]domain.TemplateElement, 0)
	for rows.Next() {
		var e domain.TemplateElement
		var parentID sql.NullInt64
		if err := rows.Scan(
			&e.ID, &e.TemplateID, &parentID, &e.ElementType, &e.Config, &e.SortOrder,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("template_element: list scan: %w", err)
		}
		if parentID.Valid {
			e.ParentID = &parentID.Int64
		}
		elements = append(elements, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("template_element: list rows: %w", err)
	}
	return elements, nil
}

// CreateDocumentTemplate creates a new user document template.
func (s *Store) CreateDocumentTemplate(ctx context.Context, input domain.DocumentTemplateInput) (domain.DocumentTemplate, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO document_template (name, description, template_type, margin_top, margin_bottom, margin_left, margin_right)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.Description, input.TemplateType,
		input.MarginTop, input.MarginBottom, input.MarginLeft, input.MarginRight,
	)
	if err != nil {
		return domain.DocumentTemplate{}, fmt.Errorf("template: insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.DocumentTemplate{}, fmt.Errorf("template: last insert id: %w", err)
	}
	return s.getDocumentTemplate(ctx, id)
}

// getDocumentTemplate is a private helper to re-fetch a template by
// ID (without elements).
func (s *Store) getDocumentTemplate(ctx context.Context, id int64) (domain.DocumentTemplate, error) {
	var t domain.DocumentTemplate
	var isBuiltin int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, template_type, is_builtin,
		        margin_top, margin_bottom, margin_left, margin_right,
		        created_at, updated_at
		 FROM document_template WHERE id = ?`, id,
	).Scan(
		&t.ID, &t.Name, &t.Description, &t.TemplateType, &isBuiltin,
		&t.MarginTop, &t.MarginBottom, &t.MarginLeft, &t.MarginRight,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return t, fmt.Errorf("template: not found: id=%d", id)
	}
	if err != nil {
		return t, fmt.Errorf("template: get: %w", err)
	}
	t.IsBuiltin = isBuiltin == 1
	return t, nil
}

// UpdateDocumentTemplate updates a user template's metadata.
// Built-in templates cannot be updated.
func (s *Store) UpdateDocumentTemplate(ctx context.Context, id int64, input domain.DocumentTemplateInput) (domain.DocumentTemplate, error) {
	// Check if template is built-in before updating.
	existing, err := s.getDocumentTemplate(ctx, id)
	if err != nil {
		return domain.DocumentTemplate{}, err
	}
	if existing.IsBuiltin {
		return domain.DocumentTemplate{}, fmt.Errorf("template: cannot update built-in template: id=%d", id)
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE document_template
		 SET name = ?, description = ?, template_type = ?,
		     margin_top = ?, margin_bottom = ?, margin_left = ?, margin_right = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.Name, input.Description, input.TemplateType,
		input.MarginTop, input.MarginBottom, input.MarginLeft, input.MarginRight,
		id,
	)
	if err != nil {
		return domain.DocumentTemplate{}, fmt.Errorf("template: update: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.DocumentTemplate{}, fmt.Errorf("template: update rows affected: %w", err)
	}
	if affected == 0 {
		return domain.DocumentTemplate{}, fmt.Errorf("template: not found: id=%d", id)
	}
	return s.getDocumentTemplate(ctx, id)
}

// DeleteDocumentTemplate deletes a user template. Built-in
// templates cannot be deleted. Elements are cascade deleted.
func (s *Store) DeleteDocumentTemplate(ctx context.Context, id int64) error {
	// Check if template is built-in before deleting.
	existing, err := s.getDocumentTemplate(ctx, id)
	if err != nil {
		return err
	}
	if existing.IsBuiltin {
		return fmt.Errorf("template: cannot delete built-in template: id=%d", id)
	}

	result, err := s.db.ExecContext(ctx, "DELETE FROM document_template WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("template: delete: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("template: delete rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("template: not found: id=%d", id)
	}
	return nil
}

// DuplicateDocumentTemplate creates a copy of a template with all
// its elements. The copy is always user-created (is_builtin=false).
func (s *Store) DuplicateDocumentTemplate(ctx context.Context, id int64, newName string) (domain.DocumentTemplate, error) {
	// Fetch the source template with all elements.
	source, err := s.GetDocumentTemplate(ctx, id)
	if err != nil {
		return domain.DocumentTemplate{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DocumentTemplate{}, fmt.Errorf("template: begin duplicate tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Insert the new template (always user-created).
	res, err := tx.ExecContext(ctx,
		`INSERT INTO document_template (name, description, template_type, is_builtin, margin_top, margin_bottom, margin_left, margin_right)
		 VALUES (?, ?, ?, 0, ?, ?, ?, ?)`,
		newName, source.Description, source.TemplateType,
		source.MarginTop, source.MarginBottom, source.MarginLeft, source.MarginRight,
	)
	if err != nil {
		return domain.DocumentTemplate{}, fmt.Errorf("template: duplicate insert: %w", err)
	}
	newTemplateID, err := res.LastInsertId()
	if err != nil {
		return domain.DocumentTemplate{}, fmt.Errorf("template: duplicate last insert id: %w", err)
	}

	// Build a map of old element ID → new element ID for parent
	// references. Insert top-level elements first, then children.
	idMap := make(map[int64]int64)

	// First pass: top-level elements (parent_id IS NULL).
	for _, e := range source.Elements {
		if e.ParentID != nil {
			continue
		}
		res, err := tx.ExecContext(ctx,
			`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order)
			 VALUES (?, NULL, ?, ?, ?)`,
			newTemplateID, e.ElementType, e.Config, e.SortOrder,
		)
		if err != nil {
			return domain.DocumentTemplate{}, fmt.Errorf("template: duplicate element: %w", err)
		}
		newID, err := res.LastInsertId()
		if err != nil {
			return domain.DocumentTemplate{}, fmt.Errorf("template: duplicate element id: %w", err)
		}
		idMap[e.ID] = newID
	}

	// Second pass: child elements (parent_id IS NOT NULL).
	for _, e := range source.Elements {
		if e.ParentID == nil {
			continue
		}
		newParentID, ok := idMap[*e.ParentID]
		if !ok {
			return domain.DocumentTemplate{}, fmt.Errorf("template: duplicate child element: parent id %d not found in id map", *e.ParentID)
		}
		res, err := tx.ExecContext(ctx,
			`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order)
			 VALUES (?, ?, ?, ?, ?)`,
			newTemplateID, newParentID, e.ElementType, e.Config, e.SortOrder,
		)
		if err != nil {
			return domain.DocumentTemplate{}, fmt.Errorf("template: duplicate child element: %w", err)
		}
		newID, err := res.LastInsertId()
		if err != nil {
			return domain.DocumentTemplate{}, fmt.Errorf("template: duplicate child element id: %w", err)
		}
		idMap[e.ID] = newID
	}

	if err := tx.Commit(); err != nil {
		return domain.DocumentTemplate{}, fmt.Errorf("template: duplicate commit: %w", err)
	}

	return s.getDocumentTemplate(ctx, newTemplateID)
}

// CreateTemplateElement adds a new element to a template, appended
// to the end of the sort order within its parent scope.
func (s *Store) CreateTemplateElement(ctx context.Context, templateID int64, input domain.TemplateElementInput) (domain.TemplateElement, error) {
	// Verify template exists.
	_, err := s.getDocumentTemplate(ctx, templateID)
	if err != nil {
		return domain.TemplateElement{}, err
	}

	// Determine next sort_order within the parent scope.
	var maxOrder sql.NullInt64
	if input.ParentID != nil {
		err = s.db.QueryRowContext(ctx,
			`SELECT MAX(sort_order) FROM template_element WHERE template_id = ? AND parent_id = ?`,
			templateID, *input.ParentID,
		).Scan(&maxOrder)
	} else {
		err = s.db.QueryRowContext(ctx,
			`SELECT MAX(sort_order) FROM template_element WHERE template_id = ? AND parent_id IS NULL`,
			templateID,
		).Scan(&maxOrder)
	}
	if err != nil {
		return domain.TemplateElement{}, fmt.Errorf("template_element: get max sort_order: %w", err)
	}
	nextOrder := 0
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	config := input.Config
	if config == "" {
		config = "{}"
	}

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO template_element (template_id, parent_id, element_type, config, sort_order)
		 VALUES (?, ?, ?, ?, ?)`,
		templateID, input.ParentID, input.ElementType, config, nextOrder,
	)
	if err != nil {
		return domain.TemplateElement{}, fmt.Errorf("template_element: insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.TemplateElement{}, fmt.Errorf("template_element: last insert id: %w", err)
	}
	return s.getTemplateElement(ctx, id)
}

// getTemplateElement is a private helper to re-fetch an element by ID.
func (s *Store) getTemplateElement(ctx context.Context, id int64) (domain.TemplateElement, error) {
	var e domain.TemplateElement
	var parentID sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, template_id, parent_id, element_type, config, sort_order,
		        created_at, updated_at
		 FROM template_element WHERE id = ?`, id,
	).Scan(
		&e.ID, &e.TemplateID, &parentID, &e.ElementType, &e.Config, &e.SortOrder,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return e, fmt.Errorf("template_element: not found: id=%d", id)
	}
	if err != nil {
		return e, fmt.Errorf("template_element: get: %w", err)
	}
	if parentID.Valid {
		e.ParentID = &parentID.Int64
	}
	return e, nil
}

// UpdateTemplateElement updates an element's type and config.
func (s *Store) UpdateTemplateElement(ctx context.Context, id int64, input domain.TemplateElementInput) (domain.TemplateElement, error) {
	config := input.Config
	if config == "" {
		config = "{}"
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE template_element
		 SET parent_id = ?, element_type = ?, config = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.ParentID, input.ElementType, config, id,
	)
	if err != nil {
		return domain.TemplateElement{}, fmt.Errorf("template_element: update: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.TemplateElement{}, fmt.Errorf("template_element: update rows affected: %w", err)
	}
	if affected == 0 {
		return domain.TemplateElement{}, fmt.Errorf("template_element: not found: id=%d", id)
	}
	return s.getTemplateElement(ctx, id)
}

// DeleteTemplateElement removes an element from a template.
func (s *Store) DeleteTemplateElement(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM template_element WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("template_element: delete: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("template_element: delete rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("template_element: not found: id=%d", id)
	}
	return nil
}

// ReorderTemplateElements updates sort_order for elements within a
// specific parent scope (top-level if parentID is nil, or within a
// specific loop container).
func (s *Store) ReorderTemplateElements(ctx context.Context, templateID int64, parentID *int64, orderedIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("template_element: begin reorder tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var stmt *sql.Stmt
	if parentID != nil {
		stmt, err = tx.PrepareContext(ctx,
			`UPDATE template_element
			 SET sort_order = ?,
			     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
			 WHERE id = ? AND template_id = ? AND parent_id = ?`,
		)
	} else {
		stmt, err = tx.PrepareContext(ctx,
			`UPDATE template_element
			 SET sort_order = ?,
			     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
			 WHERE id = ? AND template_id = ? AND parent_id IS NULL`,
		)
	}
	if err != nil {
		return fmt.Errorf("template_element: prepare reorder: %w", err)
	}
	defer stmt.Close() //nolint:errcheck

	for i, id := range orderedIDs {
		var result sql.Result
		if parentID != nil {
			result, err = stmt.ExecContext(ctx, i, id, templateID, *parentID)
		} else {
			result, err = stmt.ExecContext(ctx, i, id, templateID)
		}
		if err != nil {
			return fmt.Errorf("template_element: reorder id=%d: %w", id, err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("template_element: reorder rows affected id=%d: %w", id, err)
		}
		if affected == 0 {
			return fmt.Errorf("template_element: reorder: element id=%d not found in scope", id)
		}
	}

	return tx.Commit()
}
