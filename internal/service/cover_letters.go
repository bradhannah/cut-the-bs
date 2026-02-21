package service

import (
	"context"
	"fmt"

	"cut-the-bs/internal/domain"
)

// CoverLetterStore defines the persistence operations required by
// CoverLetterService. This is a narrow subset of domain.Store,
// following the interface segregation principle.
type CoverLetterStore interface {
	ListCoverLetters(ctx context.Context) ([]domain.CoverLetter, error)
	GetCoverLetter(ctx context.Context, id int64) (domain.CoverLetter, error)
	CreateCoverLetter(ctx context.Context, input domain.CoverLetterInput) (domain.CoverLetter, error)
	UpdateCoverLetter(ctx context.Context, id int64, input domain.CoverLetterInput) (domain.CoverLetter, error)
	DeleteCoverLetter(ctx context.Context, id int64) error
	UpdateCoverLetterFilePath(ctx context.Context, id int64, filePath string) error

	// Profile data for PDF rendering.
	GetProfile(ctx context.Context) (domain.UserProfile, error)
	ListProfileLinks(ctx context.Context) ([]domain.ProfileLink, error)
}

// CoverLetterRenderer defines the PDF rendering capability needed
// by CoverLetterService.
type CoverLetterRenderer interface {
	RenderCoverLetter(ctx context.Context, req domain.RenderCoverLetterRequest) (string, error)
}

// CoverLetterService provides business-logic operations for cover
// letters. It validates inputs before delegating to the store, and
// handles PDF export via the renderer.
type CoverLetterService struct {
	store     CoverLetterStore
	renderer  CoverLetterRenderer
	outputDir string
}

// NewCoverLetterService creates a CoverLetterService backed by the
// given store and renderer. outputDir is where generated PDFs are
// saved.
func NewCoverLetterService(
	store CoverLetterStore,
	renderer CoverLetterRenderer,
	outputDir string,
) *CoverLetterService {
	return &CoverLetterService{
		store:     store,
		renderer:  renderer,
		outputDir: outputDir,
	}
}

// ListCoverLetters returns all cover letters.
func (s *CoverLetterService) ListCoverLetters(ctx context.Context) ([]domain.CoverLetter, error) {
	return s.store.ListCoverLetters(ctx)
}

// CreateCoverLetter validates the input and creates a new cover
// letter.
func (s *CoverLetterService) CreateCoverLetter(ctx context.Context, input domain.CoverLetterInput) (domain.CoverLetter, error) {
	if err := domain.ValidateCoverLetterInput(input); err != nil {
		return domain.CoverLetter{}, err
	}
	return s.store.CreateCoverLetter(ctx, input)
}

// UpdateCoverLetter validates the input and updates an existing
// cover letter.
func (s *CoverLetterService) UpdateCoverLetter(ctx context.Context, id int64, input domain.CoverLetterInput) (domain.CoverLetter, error) {
	if err := domain.ValidateCoverLetterInput(input); err != nil {
		return domain.CoverLetter{}, err
	}
	return s.store.UpdateCoverLetter(ctx, id, input)
}

// DeleteCoverLetter deletes a cover letter by ID.
func (s *CoverLetterService) DeleteCoverLetter(ctx context.Context, id int64) error {
	return s.store.DeleteCoverLetter(ctx, id)
}

// ExportCoverLetter generates a PDF for the given cover letter,
// stores the file path, and returns the updated letter.
func (s *CoverLetterService) ExportCoverLetter(ctx context.Context, id int64) (domain.CoverLetter, error) {
	letter, err := s.store.GetCoverLetter(ctx, id)
	if err != nil {
		return domain.CoverLetter{}, fmt.Errorf("get cover letter: %w", err)
	}

	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return domain.CoverLetter{}, fmt.Errorf("get profile: %w", err)
	}

	links, err := s.store.ListProfileLinks(ctx)
	if err != nil {
		return domain.CoverLetter{}, fmt.Errorf("list profile links: %w", err)
	}

	filePath, err := s.renderer.RenderCoverLetter(ctx, domain.RenderCoverLetterRequest{
		OutputDir: s.outputDir,
		Profile:   profile,
		Links:     links,
		Letter:    letter,
	})
	if err != nil {
		return domain.CoverLetter{}, fmt.Errorf("render cover letter: %w", err)
	}

	if err := s.store.UpdateCoverLetterFilePath(ctx, id, filePath); err != nil {
		return domain.CoverLetter{}, fmt.Errorf("update file path: %w", err)
	}

	letter.FilePath = filePath
	return letter, nil
}
