package sqlite

import (
	"context"
	"fmt"

	"cut-the-bs/internal/domain"
)

// --- Skill Lens Tags ---

// GetSkillLensTags returns all lens IDs tagged for a skill.
func (s *Store) GetSkillLensTags(ctx context.Context, skillID int64) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT lens_id
		 FROM skill_lens_tag
		 WHERE skill_id = ?
		 ORDER BY lens_id ASC`,
		skillID,
	)
	if err != nil {
		return nil, fmt.Errorf("skill_lens_tag: get: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("skill_lens_tag: scan: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// SetSkillLensTags replaces all lens tags for a skill.
func (s *Store) SetSkillLensTags(ctx context.Context, skillID int64, lensIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("skill_lens_tag: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM skill_lens_tag WHERE skill_id = ?", skillID,
	); err != nil {
		return fmt.Errorf("skill_lens_tag: delete: %w", err)
	}

	if len(lensIDs) > 0 {
		stmt, err := tx.PrepareContext(ctx,
			"INSERT INTO skill_lens_tag (skill_id, lens_id) VALUES (?, ?)",
		)
		if err != nil {
			return fmt.Errorf("skill_lens_tag: prepare: %w", err)
		}
		defer stmt.Close() //nolint:errcheck

		for _, lensID := range lensIDs {
			if _, err := stmt.ExecContext(ctx, skillID, lensID); err != nil {
				return fmt.Errorf("skill_lens_tag: insert: %w", err)
			}
		}
	}

	return tx.Commit()
}

// ListSkillsWithLensTags returns all skills with their lens tag
// associations included.
func (s *Store) ListSkillsWithLensTags(ctx context.Context) ([]domain.SkillWithTags, error) {
	// First, fetch all skills.
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, category_id, competence_level, is_legacy,
		        created_at, updated_at
		 FROM skill
		 ORDER BY competence_level DESC, name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("skill_lens_tag: list skills: %w", err)
	}

	type skillRow struct {
		skill  domain.Skill
		legacy int
	}
	var skillRows []skillRow
	for rows.Next() {
		var sr skillRow
		if err := rows.Scan(
			&sr.skill.ID, &sr.skill.Name, &sr.skill.CategoryID,
			&sr.skill.CompetenceLevel, &sr.legacy,
			&sr.skill.CreatedAt, &sr.skill.UpdatedAt,
		); err != nil {
			rows.Close() //nolint:errcheck
			return nil, fmt.Errorf("skill_lens_tag: scan skill: %w", err)
		}
		sr.skill.IsLegacy = sr.legacy == 1
		skillRows = append(skillRows, sr)
	}
	rows.Close() //nolint:errcheck
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("skill_lens_tag: rows err: %w", err)
	}

	if len(skillRows) == 0 {
		return make([]domain.SkillWithTags, 0), nil
	}

	// Fetch all tags in one query, then group by skill_id.
	tagRows, err := s.db.QueryContext(ctx,
		`SELECT skill_id, lens_id
		 FROM skill_lens_tag
		 ORDER BY skill_id, lens_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("skill_lens_tag: list tags: %w", err)
	}

	tagsBySkill := make(map[int64][]int64)
	for tagRows.Next() {
		var skillID, lensID int64
		if err := tagRows.Scan(&skillID, &lensID); err != nil {
			tagRows.Close() //nolint:errcheck
			return nil, fmt.Errorf("skill_lens_tag: scan tag: %w", err)
		}
		tagsBySkill[skillID] = append(tagsBySkill[skillID], lensID)
	}
	tagRows.Close() //nolint:errcheck
	if err := tagRows.Err(); err != nil {
		return nil, fmt.Errorf("skill_lens_tag: tag rows err: %w", err)
	}

	// Assemble result.
	result := make([]domain.SkillWithTags, 0, len(skillRows))
	for _, sr := range skillRows {
		swt := domain.SkillWithTags{
			Skill:   sr.skill,
			LensIDs: tagsBySkill[sr.skill.ID],
		}
		if swt.LensIDs == nil {
			swt.LensIDs = make([]int64, 0)
		}
		result = append(result, swt)
	}

	return result, nil
}
