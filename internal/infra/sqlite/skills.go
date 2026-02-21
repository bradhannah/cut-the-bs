package sqlite

import (
	"context"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Skills ---

// ListSkills returns all skills sorted by competence level
// (descending), then alphabetically by name.
func (s *Store) ListSkills(ctx context.Context) ([]domain.Skill, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, category_id, competence_level, is_legacy, created_at, updated_at
		 FROM skill
		 ORDER BY competence_level DESC, name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("skill: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	skills := make([]domain.Skill, 0)
	for rows.Next() {
		var sk domain.Skill
		var legacy int
		if err := rows.Scan(
			&sk.ID, &sk.Name, &sk.CategoryID,
			&sk.CompetenceLevel, &legacy,
			&sk.CreatedAt, &sk.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("skill: scan: %w", err)
		}
		sk.IsLegacy = legacy == 1
		skills = append(skills, sk)
	}

	return skills, rows.Err()
}

// ListSkillsByCategory returns skills grouped by category, with
// categories ordered by their sort_order and skills sorted by
// competence descending then name ascending within each category.
func (s *Store) ListSkillsByCategory(ctx context.Context) ([]domain.SkillCategoryWithSkills, error) {
	// First, fetch all categories.
	cats, err := s.ListSkillCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("skill: list by category: categories: %w", err)
	}

	if len(cats) == 0 {
		return make([]domain.SkillCategoryWithSkills, 0), nil
	}

	// Then fetch all skills sorted by category sort_order, then
	// competence desc, then name asc.
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.id, s.name, s.category_id, s.competence_level, s.is_legacy,
		        s.created_at, s.updated_at
		 FROM skill s
		 JOIN skill_category c ON s.category_id = c.id
		 ORDER BY c.sort_order ASC, s.competence_level DESC, s.name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("skill: list by category: skills: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	// Build a map of category_id -> skills.
	skillsByCategory := make(map[int64][]domain.Skill)
	for rows.Next() {
		var sk domain.Skill
		var legacy int
		if err := rows.Scan(
			&sk.ID, &sk.Name, &sk.CategoryID,
			&sk.CompetenceLevel, &legacy,
			&sk.CreatedAt, &sk.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("skill: list by category: scan: %w", err)
		}
		sk.IsLegacy = legacy == 1
		skillsByCategory[sk.CategoryID] = append(skillsByCategory[sk.CategoryID], sk)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("skill: list by category: rows: %w", err)
	}

	// Assemble the result in category sort_order.
	result := make([]domain.SkillCategoryWithSkills, 0, len(cats))
	for _, cat := range cats {
		skills := skillsByCategory[cat.ID]
		if skills == nil {
			skills = make([]domain.Skill, 0)
		}
		result = append(result, domain.SkillCategoryWithSkills{
			Category: cat,
			Skills:   skills,
		})
	}

	return result, nil
}

// CreateSkill creates a new skill.
func (s *Store) CreateSkill(ctx context.Context, input domain.SkillInput) (domain.Skill, error) {
	legacy := 0
	if input.IsLegacy {
		legacy = 1
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO skill (name, category_id, competence_level, is_legacy)
		 VALUES (?, ?, ?, ?)`,
		input.Name, input.CategoryID, input.CompetenceLevel, legacy,
	)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("skill: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Skill{}, fmt.Errorf("skill: last insert id: %w", err)
	}

	return s.getSkill(ctx, id)
}

// UpdateSkill updates an existing skill.
func (s *Store) UpdateSkill(ctx context.Context, id int64, input domain.SkillInput) (domain.Skill, error) {
	legacy := 0
	if input.IsLegacy {
		legacy = 1
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE skill
		 SET name = ?, category_id = ?, competence_level = ?, is_legacy = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`,
		input.Name, input.CategoryID, input.CompetenceLevel, legacy, id,
	)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("skill: update: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.Skill{}, fmt.Errorf("skill: not found: %d", id)
	}

	return s.getSkill(ctx, id)
}

// DeleteSkill deletes a skill by ID.
func (s *Store) DeleteSkill(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM skill WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("skill: delete: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("skill: not found: %d", id)
	}

	return nil
}

// CheckSkillLensReferences returns the names of lenses that
// reference a given skill.
func (s *Store) CheckSkillLensReferences(ctx context.Context, id int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.name FROM lens l
		 JOIN lens_skill_selection ls ON l.id = ls.lens_id
		 WHERE ls.skill_id = ?
		 ORDER BY l.name ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("skill: check lens refs: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("skill: check lens refs: scan: %w", err)
		}
		names = append(names, name)
	}

	return names, rows.Err()
}

// getSkill fetches a single skill by ID.
func (s *Store) getSkill(ctx context.Context, id int64) (domain.Skill, error) {
	var sk domain.Skill
	var legacy int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, category_id, competence_level, is_legacy, created_at, updated_at
		 FROM skill WHERE id = ?`, id,
	).Scan(
		&sk.ID, &sk.Name, &sk.CategoryID,
		&sk.CompetenceLevel, &legacy,
		&sk.CreatedAt, &sk.UpdatedAt,
	)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("skill: get %d: %w", id, err)
	}
	sk.IsLegacy = legacy == 1
	return sk, nil
}
