package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"cut-the-bs/internal/domain"
)

// GetProfile returns the user's profile. If no profile exists yet,
// a default empty profile is created with id=1.
func (s *Store) GetProfile(ctx context.Context) (domain.UserProfile, error) {
	var profile domain.UserProfile

	err := s.db.QueryRowContext(ctx,
		`SELECT id, full_name, email, phone, location,
		        created_at, updated_at
		 FROM user_profile WHERE id = 1`,
	).Scan(
		&profile.ID, &profile.FullName, &profile.Email,
		&profile.Phone, &profile.Location,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Auto-create a default empty profile.
		_, execErr := s.db.ExecContext(ctx,
			`INSERT INTO user_profile (id, full_name, email, phone, location)
			 VALUES (1, '', '', '', '')`,
		)
		if execErr != nil {
			return domain.UserProfile{}, fmt.Errorf("profile: auto-create: %w", execErr)
		}

		// Re-fetch to get timestamps.
		return s.GetProfile(ctx)
	}

	if err != nil {
		return domain.UserProfile{}, fmt.Errorf("profile: get: %w", err)
	}

	return profile, nil
}

// UpdateProfile updates the user's profile fields and returns the
// updated profile. The profile must already exist (call GetProfile
// first if unsure).
func (s *Store) UpdateProfile(ctx context.Context, profile domain.UserProfile) (domain.UserProfile, error) {
	_, err := s.db.ExecContext(ctx,
		`UPDATE user_profile
		 SET full_name = ?, email = ?, phone = ?, location = ?,
		     updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = 1`,
		profile.FullName, profile.Email, profile.Phone, profile.Location,
	)
	if err != nil {
		return domain.UserProfile{}, fmt.Errorf("profile: update: %w", err)
	}

	return s.GetProfile(ctx)
}
