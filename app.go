package main

import (
	"context"
	"fmt"
	"log/slog"

	"cut-the-bs/internal/config"
	"cut-the-bs/internal/domain"
	"cut-the-bs/internal/infra"
	"cut-the-bs/internal/infra/sqlite"
	"cut-the-bs/internal/service"
)

// App is the main application struct that serves as the Wails
// binding target. All public methods on App are exposed to the
// Svelte frontend as JavaScript Promise-returning functions.
type App struct {
	ctx    context.Context
	cfg    config.Config
	store  *sqlite.Store
	logger *slog.Logger

	// Services
	workHistorySvc *service.WorkHistoryService
	profileSvc     *service.ProfileService
}

// NewApp creates a new App instance. Service dependencies are
// initialized during the startup lifecycle hook, not here.
func NewApp() *App {
	logger := infra.NewLogger(slog.LevelInfo, nil)
	return &App{
		logger: logger,
	}
}

// startup is the Wails OnStartup lifecycle hook. It loads
// configuration, ensures the data directory exists, opens the
// SQLite database, and runs migrations.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	log := infra.ComponentLogger(a.logger, "startup")

	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config, using defaults",
			slog.String("error", err.Error()))
		cfg = config.DefaultConfig()
	}
	a.cfg = cfg

	// Ensure data directory exists.
	if err := a.cfg.EnsureDataDir(); err != nil {
		log.Error("failed to create data directory",
			slog.String("error", err.Error()))
		return
	}

	// Open database.
	dbPath, err := a.cfg.DBPath()
	if err != nil {
		log.Error("failed to resolve database path",
			slog.String("error", err.Error()))
		return
	}

	dbLogger := infra.ComponentLogger(a.logger, "sqlite")
	store, err := sqlite.NewStore(dbPath, dbLogger)
	if err != nil {
		log.Error("failed to open database",
			slog.String("error", err.Error()))
		return
	}
	a.store = store

	// Run migrations.
	if err := sqlite.Migrate(store); err != nil {
		log.Error("failed to run migrations",
			slog.String("error", err.Error()))
		return
	}

	// Initialize services.
	a.workHistorySvc = service.NewWorkHistoryService(store)
	a.profileSvc = service.NewProfileService(store)

	log.Info("application started",
		slog.String("db_path", dbPath))
}

// beforeClose is the Wails OnBeforeClose lifecycle hook. It
// checkpoints the WAL to ensure data durability before the window
// closes. Returning false allows the close to proceed.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	if a.store != nil {
		log := infra.ComponentLogger(a.logger, "lifecycle")
		if err := a.store.Checkpoint(); err != nil {
			log.Warn("WAL checkpoint failed before close",
				slog.String("error", err.Error()))
		}
	}
	return false
}

// shutdown is the Wails OnShutdown lifecycle hook. It closes the
// database connection and releases all resources.
func (a *App) shutdown(ctx context.Context) {
	log := infra.ComponentLogger(a.logger, "lifecycle")

	if a.store != nil {
		if err := a.store.Close(); err != nil {
			log.Error("failed to close database",
				slog.String("error", err.Error()))
		}
	}

	log.Info("application shut down")
}

// Greet returns a greeting for the given name.
// This is a placeholder binding method for initial setup verification.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// =================================================================
// Work History Bindings
// =================================================================

// ListWorkHistory returns all work history entries ordered by
// sort_order, each with its achievement bullets.
func (a *App) ListWorkHistory() ([]domain.WorkHistoryEntry, error) {
	return a.workHistorySvc.ListWorkHistory(a.ctx)
}

// GetWorkHistory returns a single work history entry by ID with
// its bullets included.
func (a *App) GetWorkHistory(id int64) (domain.WorkHistoryEntry, error) {
	return a.workHistorySvc.GetWorkHistory(a.ctx, id)
}

// CreateWorkHistory creates a new work history entry.
// Returns the created entry with generated ID.
func (a *App) CreateWorkHistory(entry domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	return a.workHistorySvc.CreateWorkHistory(a.ctx, entry)
}

// UpdateWorkHistory updates an existing work history entry.
func (a *App) UpdateWorkHistory(id int64, entry domain.WorkHistoryInput) (domain.WorkHistoryEntry, error) {
	return a.workHistorySvc.UpdateWorkHistory(a.ctx, id, entry)
}

// DeleteWorkHistory deletes a work history entry and its bullets.
func (a *App) DeleteWorkHistory(id int64) error {
	return a.workHistorySvc.DeleteWorkHistory(a.ctx, id)
}

// ReorderWorkHistory updates the sort_order of all entries.
// Accepts a slice of IDs in the desired order.
func (a *App) ReorderWorkHistory(orderedIDs []int64) error {
	return a.workHistorySvc.ReorderWorkHistory(a.ctx, orderedIDs)
}

// CreateBullet adds an achievement bullet to a work history entry.
func (a *App) CreateBullet(workHistoryID int64, text string) (domain.AchievementBullet, error) {
	return a.workHistorySvc.CreateBullet(a.ctx, workHistoryID, text)
}

// UpdateBullet updates an achievement bullet's text.
func (a *App) UpdateBullet(id int64, text string) (domain.AchievementBullet, error) {
	return a.workHistorySvc.UpdateBullet(a.ctx, id, text)
}

// DeleteBullet deletes an achievement bullet.
func (a *App) DeleteBullet(id int64) error {
	return a.workHistorySvc.DeleteBullet(a.ctx, id)
}

// ReorderBullets updates the sort_order of bullets within an entry.
func (a *App) ReorderBullets(workHistoryID int64, orderedIDs []int64) error {
	return a.workHistorySvc.ReorderBullets(a.ctx, workHistoryID, orderedIDs)
}

// SplitBulletText accepts a block of text and returns individual
// lines for preview before creating bullets.
func (a *App) SplitBulletText(text string) []string {
	return a.workHistorySvc.SplitBulletText(text)
}

// =================================================================
// Profile Bindings
// =================================================================

// GetProfile returns the user's profile. Creates a default empty
// profile if none exists.
func (a *App) GetProfile() (domain.UserProfile, error) {
	return a.profileSvc.GetProfile(a.ctx)
}

// UpdateProfile updates the user's profile fields.
// Returns the updated profile.
func (a *App) UpdateProfile(profile domain.UserProfile) (domain.UserProfile, error) {
	return a.profileSvc.UpdateProfile(a.ctx, profile)
}

// =================================================================
// Profile Link Bindings
// =================================================================

// ListProfileLinks returns all profile links ordered by sort_order.
func (a *App) ListProfileLinks() ([]domain.ProfileLink, error) {
	return a.profileSvc.ListProfileLinks(a.ctx)
}

// CreateProfileLink creates a new profile link.
func (a *App) CreateProfileLink(input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	return a.profileSvc.CreateProfileLink(a.ctx, input)
}

// UpdateProfileLink updates an existing profile link.
func (a *App) UpdateProfileLink(id int64, input domain.ProfileLinkInput) (domain.ProfileLink, error) {
	return a.profileSvc.UpdateProfileLink(a.ctx, id, input)
}

// DeleteProfileLink deletes a profile link.
func (a *App) DeleteProfileLink(id int64) error {
	return a.profileSvc.DeleteProfileLink(a.ctx, id)
}

// ReorderProfileLinks updates the sort_order of all profile links.
func (a *App) ReorderProfileLinks(orderedIDs []int64) error {
	return a.profileSvc.ReorderProfileLinks(a.ctx, orderedIDs)
}
