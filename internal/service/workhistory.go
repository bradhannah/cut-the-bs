package service

import (
	"context"
	"strings"

	"cut-the-bs/internal/domain"
)

// WorkHistoryStore defines the persistence operations required by
// WorkHistoryService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type WorkHistoryStore interface {
	ListWorkHistory(ctx context.Context) ([]domain.WorkHistoryEntry, error)
	GetWorkHistory(ctx context.Context, id int64) (domain.WorkHistoryEntry, error)
	CreateWorkHistory(ctx context.Context, input domain.WorkHistoryInput) (domain.WorkHistoryEntry, error)
	UpdateWorkHistory(ctx context.Context, id int64, input domain.WorkHistoryInput) (domain.WorkHistoryEntry, error)
	DeleteWorkHistory(ctx context.Context, id int64) error
	ReorderWorkHistory(ctx context.Context, orderedIDs []int64) error
	CreateBullet(ctx context.Context, workHistoryID int64, text string) (domain.AchievementBullet, error)
	UpdateBullet(ctx context.Context, id int64, text string) (domain.AchievementBullet, error)
	DeleteBullet(ctx context.Context, id int64) error
	ReorderBullets(ctx context.Context, workHistoryID int64, orderedIDs []int64) error
}

// WorkHistoryService provides business-logic operations for work
// history entries and their achievement bullets. It validates inputs
// before delegating to the store and implements autosave semantics
// (every write is immediate — no batching or deferred saves).
type WorkHistoryService struct {
	store WorkHistoryStore
}

// NewWorkHistoryService creates a WorkHistoryService backed by the
// given store. The store must not be nil for methods that perform
// persistence; SplitBulletText is a pure function that works without
// a store.
func NewWorkHistoryService(store WorkHistoryStore) *WorkHistoryService {
	return &WorkHistoryService{store: store}
}

// --- Work History Entries ---

// CreateWorkHistory validates the input and creates a new work
// history entry.
func (s *WorkHistoryService) CreateWorkHistory(ctx context.Context, input domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	if err := domain.ValidateWorkHistoryInput(input); err != nil {
		return domain.WorkHistoryEntry{}, err
	}
	return s.store.CreateWorkHistory(ctx, input)
}

// GetWorkHistory returns a single work history entry by ID with
// its bullets included.
func (s *WorkHistoryService) GetWorkHistory(ctx context.Context, id int64) (domain.WorkHistoryEntry, error) {
	return s.store.GetWorkHistory(ctx, id)
}

// UpdateWorkHistory validates the input and updates an existing
// work history entry.
func (s *WorkHistoryService) UpdateWorkHistory(ctx context.Context, id int64, input domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	if err := domain.ValidateWorkHistoryInput(input); err != nil {
		return domain.WorkHistoryEntry{}, err
	}
	return s.store.UpdateWorkHistory(ctx, id, input)
}

// DeleteWorkHistory deletes a work history entry by ID. Associated
// bullets are deleted via CASCADE in the schema.
func (s *WorkHistoryService) DeleteWorkHistory(ctx context.Context, id int64) error {
	return s.store.DeleteWorkHistory(ctx, id)
}

// ListWorkHistory returns all work history entries ordered by
// sort_order, each with its bullets included.
func (s *WorkHistoryService) ListWorkHistory(ctx context.Context) ([]domain.WorkHistoryEntry, error) {
	return s.store.ListWorkHistory(ctx)
}

// ReorderWorkHistory updates the sort_order of all entries based
// on the provided ordered slice of IDs.
func (s *WorkHistoryService) ReorderWorkHistory(ctx context.Context, orderedIDs []int64) error {
	return s.store.ReorderWorkHistory(ctx, orderedIDs)
}

// --- Achievement Bullets ---

// CreateBullet validates that text is non-empty and creates a new
// achievement bullet for the given work history entry.
func (s *WorkHistoryService) CreateBullet(ctx context.Context, workHistoryID int64, text string) (domain.AchievementBullet, error) {
	if err := domain.ValidateRequired(text, "bullet text"); err != nil {
		return domain.AchievementBullet{}, err
	}
	return s.store.CreateBullet(ctx, workHistoryID, text)
}

// UpdateBullet validates that text is non-empty and updates an
// existing achievement bullet.
func (s *WorkHistoryService) UpdateBullet(ctx context.Context, id int64, text string) (domain.AchievementBullet, error) {
	if err := domain.ValidateRequired(text, "bullet text"); err != nil {
		return domain.AchievementBullet{}, err
	}
	return s.store.UpdateBullet(ctx, id, text)
}

// DeleteBullet deletes an achievement bullet by ID.
func (s *WorkHistoryService) DeleteBullet(ctx context.Context, id int64) error {
	return s.store.DeleteBullet(ctx, id)
}

// ReorderBullets updates the sort_order of bullets within a work
// history entry.
func (s *WorkHistoryService) ReorderBullets(ctx context.Context, workHistoryID int64, orderedIDs []int64) error {
	return s.store.ReorderBullets(ctx, workHistoryID, orderedIDs)
}

// --- Bullet Text Splitting ---

// SplitBulletText accepts a block of text and splits it into
// individual lines, trimming whitespace and filtering out blank
// lines. This is a preview operation — no persistence occurs.
// Handles both \n and \r\n line endings.
func (s *WorkHistoryService) SplitBulletText(text string) []string {
	// Normalize \r\n to \n.
	text = strings.ReplaceAll(text, "\r\n", "\n")

	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
