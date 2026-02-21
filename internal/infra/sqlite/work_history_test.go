package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Work History CRUD ---

func TestCreateWorkHistory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	input := domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Software Engineer",
		StartDate:            "2020-01",
		EndDate:              "2023-06",
		DateGranularityStart: "month",
		DateGranularityEnd:   "month",
	}

	entry, err := store.CreateWorkHistory(ctx, input)
	require.NoError(t, err)
	assert.NotZero(t, entry.ID)
	assert.Equal(t, "Acme Corp", entry.EmployerName)
	assert.Equal(t, "Software Engineer", entry.JobTitle)
	assert.Equal(t, "2020-01", entry.StartDate)
	assert.Equal(t, "2023-06", entry.EndDate)
	assert.Equal(t, "month", entry.DateGranularityStart)
	assert.Equal(t, "month", entry.DateGranularityEnd)
	assert.Equal(t, 1, entry.SortOrder)
	assert.NotEmpty(t, entry.CreatedAt)
	assert.NotEmpty(t, entry.UpdatedAt)
	assert.Empty(t, entry.Bullets)
}

func TestCreateWorkHistory_NewEntrySortsToTop(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	first, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "First Co",
		JobTitle:             "Dev",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, first.SortOrder)

	// Second entry should be inserted at position 1, pushing the first to 2.
	second, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Second Co",
		JobTitle:             "Dev",
		StartDate:            "2021",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, second.SortOrder)

	// Verify the first entry was shifted to position 2.
	firstRefresh, err := store.GetWorkHistory(ctx, first.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, firstRefresh.SortOrder)

	// List should return second (newest) first, then first (oldest).
	entries, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "Second Co", entries[0].EmployerName)
	assert.Equal(t, "First Co", entries[1].EmployerName)
}

func TestCreateWorkHistory_PresentJob(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Current Co",
		JobTitle:             "Lead",
		StartDate:            "2023-01-15",
		EndDate:              "",
		DateGranularityStart: "day",
		DateGranularityEnd:   "",
	})
	require.NoError(t, err)
	assert.Empty(t, entry.EndDate)
	assert.Empty(t, entry.DateGranularityEnd)
}

func TestGetWorkHistory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Acme Corp", got.EmployerName)
	assert.Empty(t, got.Bullets)
}

func TestGetWorkHistory_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.GetWorkHistory(ctx, 999)
	require.Error(t, err)
}

func TestGetWorkHistory_IncludesBullets(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, entry.ID, "Led project X", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Improved perf by 50%", domain.BulletTypePrimary)
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	require.Len(t, got.Bullets, 2)
	assert.Equal(t, "Led project X", got.Bullets[0].Text)
	assert.Equal(t, "Improved perf by 50%", got.Bullets[1].Text)
}

func TestUpdateWorkHistory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Old Name",
		JobTitle:             "Old Title",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	updated, err := store.UpdateWorkHistory(ctx, created.ID, domain.WorkHistoryInput{
		EmployerName:         "New Name",
		JobTitle:             "New Title",
		StartDate:            "2019-06",
		EndDate:              "2023-12",
		DateGranularityStart: "month",
		DateGranularityEnd:   "month",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "New Name", updated.EmployerName)
	assert.Equal(t, "New Title", updated.JobTitle)
	assert.Equal(t, "2019-06", updated.StartDate)
	assert.Equal(t, "2023-12", updated.EndDate)
	assert.Equal(t, "month", updated.DateGranularityStart)
	assert.Equal(t, "month", updated.DateGranularityEnd)
	// sort_order should not change on update
	assert.Equal(t, created.SortOrder, updated.SortOrder)
}

func TestUpdateWorkHistory_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateWorkHistory(ctx, 999, domain.WorkHistoryInput{
		EmployerName:         "Name",
		JobTitle:             "Title",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.Error(t, err)
}

func TestDeleteWorkHistory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	err = store.DeleteWorkHistory(ctx, entry.ID)
	require.NoError(t, err)

	_, err = store.GetWorkHistory(ctx, entry.ID)
	require.Error(t, err)
}

func TestDeleteWorkHistory_CascadesBullets(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, entry.ID, "Bullet 1", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Bullet 2", domain.BulletTypePrimary)
	require.NoError(t, err)

	// Verify bullets exist before delete.
	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	require.Len(t, got.Bullets, 2)

	// Delete the entry — bullets should cascade.
	err = store.DeleteWorkHistory(ctx, entry.ID)
	require.NoError(t, err)

	// Verify the entry is gone.
	_, err = store.GetWorkHistory(ctx, entry.ID)
	require.Error(t, err)

	// Verify bullets are gone via a raw query (no GetBullet method).
	var count int
	err = store.db.QueryRow(
		"SELECT COUNT(*) FROM achievement_bullet WHERE work_history_id = ?",
		entry.ID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeleteWorkHistory_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteWorkHistory(ctx, 999)
	require.Error(t, err)
}

// --- List & Reorder ---

func TestListWorkHistory_OrderedBySortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "First", JobTitle: "Dev",
		StartDate: "2018", DateGranularityStart: "year",
	})
	require.NoError(t, err)
	_, err = store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Second", JobTitle: "Dev",
		StartDate: "2019", DateGranularityStart: "year",
	})
	require.NoError(t, err)
	_, err = store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Third", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	entries, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)
	require.Len(t, entries, 3)
	// Newest entries are inserted at the top (sort_order 1).
	assert.Equal(t, "Third", entries[0].EmployerName)
	assert.Equal(t, "Second", entries[1].EmployerName)
	assert.Equal(t, "First", entries[2].EmployerName)
}

func TestListWorkHistory_IncludesBullets(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, entry.ID, "Bullet A", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Bullet B", domain.BulletTypePrimary)
	require.NoError(t, err)

	entries, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Len(t, entries[0].Bullets, 2)
	assert.Equal(t, "Bullet A", entries[0].Bullets[0].Text)
	assert.Equal(t, "Bullet B", entries[0].Bullets[1].Text)
}

func TestListWorkHistory_Empty(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entries, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestReorderWorkHistory(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	e1, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "First", JobTitle: "Dev",
		StartDate: "2018", DateGranularityStart: "year",
	})
	require.NoError(t, err)
	e2, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Second", JobTitle: "Dev",
		StartDate: "2019", DateGranularityStart: "year",
	})
	require.NoError(t, err)
	e3, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Third", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	// Reverse the order: Third, Second, First
	err = store.ReorderWorkHistory(ctx, []int64{e3.ID, e2.ID, e1.ID})
	require.NoError(t, err)

	entries, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)
	require.Len(t, entries, 3)
	assert.Equal(t, "Third", entries[0].EmployerName)
	assert.Equal(t, 1, entries[0].SortOrder)
	assert.Equal(t, "Second", entries[1].EmployerName)
	assert.Equal(t, 2, entries[1].SortOrder)
	assert.Equal(t, "First", entries[2].EmployerName)
	assert.Equal(t, 3, entries[2].SortOrder)
}

// --- Date Granularity Storage ---

func TestWorkHistory_DateGranularityVariants(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	tests := []struct {
		name      string
		start     string
		end       string
		granStart string
		granEnd   string
	}{
		{"year only", "2020", "2023", "year", "year"},
		{"month only", "2020-03", "2023-06", "month", "month"},
		{"day only", "2020-03-15", "2023-06-30", "day", "day"},
		{"mixed year-month", "2020", "2023-06", "year", "month"},
		{"present job", "2023-01", "", "month", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
				EmployerName:         "Test Co",
				JobTitle:             "Dev",
				StartDate:            tc.start,
				EndDate:              tc.end,
				DateGranularityStart: tc.granStart,
				DateGranularityEnd:   tc.granEnd,
			})
			require.NoError(t, err)

			got, err := store.GetWorkHistory(ctx, entry.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.start, got.StartDate)
			assert.Equal(t, tc.end, got.EndDate)
			assert.Equal(t, tc.granStart, got.DateGranularityStart)
			assert.Equal(t, tc.granEnd, got.DateGranularityEnd)
		})
	}
}

// --- Bullet CRUD ---

func TestCreateBullet(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	bullet, err := store.CreateBullet(ctx, entry.ID, "Led a team of 5 engineers", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.NotZero(t, bullet.ID)
	assert.Equal(t, entry.ID, bullet.WorkHistoryID)
	assert.Equal(t, "Led a team of 5 engineers", bullet.Text)
	assert.Equal(t, 1, bullet.SortOrder)
	assert.NotEmpty(t, bullet.CreatedAt)
}

func TestCreateBullet_AutoIncrementsSortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	b1, err := store.CreateBullet(ctx, entry.ID, "First", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, 1, b1.SortOrder)

	b2, err := store.CreateBullet(ctx, entry.ID, "Second", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, 2, b2.SortOrder)

	b3, err := store.CreateBullet(ctx, entry.ID, "Third", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, 3, b3.SortOrder)
}

func TestCreateBullet_SortOrderPerEntry(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	e1, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)
	e2, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Beta", JobTitle: "Dev",
		StartDate: "2021", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	b1, err := store.CreateBullet(ctx, e1.ID, "E1 bullet", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, 1, b1.SortOrder)

	// Second entry starts its own sort_order sequence at 1.
	b2, err := store.CreateBullet(ctx, e2.ID, "E2 bullet", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, 1, b2.SortOrder)
}

func TestCreateBullet_InvalidWorkHistoryID(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateBullet(ctx, 999, "orphan bullet", domain.BulletTypePrimary)
	require.Error(t, err)
}

func TestUpdateBullet(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	bullet, err := store.CreateBullet(ctx, entry.ID, "Old text", domain.BulletTypePrimary)
	require.NoError(t, err)

	updated, err := store.UpdateBullet(ctx, bullet.ID, "New text")
	require.NoError(t, err)
	assert.Equal(t, bullet.ID, updated.ID)
	assert.Equal(t, "New text", updated.Text)
	assert.Equal(t, bullet.SortOrder, updated.SortOrder)
}

func TestUpdateBullet_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateBullet(ctx, 999, "text")
	require.Error(t, err)
}

func TestDeleteBullet(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	bullet, err := store.CreateBullet(ctx, entry.ID, "To be deleted", domain.BulletTypePrimary)
	require.NoError(t, err)

	err = store.DeleteBullet(ctx, bullet.ID)
	require.NoError(t, err)

	// Verify bullet is gone.
	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Bullets)
}

func TestDeleteBullet_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteBullet(ctx, 999)
	require.Error(t, err)
}

// --- Bullet Reorder ---

func TestReorderBullets(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	b1, err := store.CreateBullet(ctx, entry.ID, "Alpha", domain.BulletTypePrimary)
	require.NoError(t, err)
	b2, err := store.CreateBullet(ctx, entry.ID, "Beta", domain.BulletTypePrimary)
	require.NoError(t, err)
	b3, err := store.CreateBullet(ctx, entry.ID, "Gamma", domain.BulletTypePrimary)
	require.NoError(t, err)

	// Reorder: Gamma, Alpha, Beta
	err = store.ReorderBullets(ctx, entry.ID, []int64{b3.ID, b1.ID, b2.ID})
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	require.Len(t, got.Bullets, 3)
	assert.Equal(t, "Gamma", got.Bullets[0].Text)
	assert.Equal(t, 1, got.Bullets[0].SortOrder)
	assert.Equal(t, "Alpha", got.Bullets[1].Text)
	assert.Equal(t, 2, got.Bullets[1].SortOrder)
	assert.Equal(t, "Beta", got.Bullets[2].Text)
	assert.Equal(t, 3, got.Bullets[2].SortOrder)
}

// =================================================================
// Summary Field Tests
// =================================================================

func TestCreateWorkHistory_WithSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Senior Engineer",
		Summary:              "Led backend platform team responsible for core services.",
		StartDate:            "2020-01",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)
	assert.Equal(t, "Led backend platform team responsible for core services.", entry.Summary)
}

func TestCreateWorkHistory_EmptySummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		Summary:              "",
		StartDate:            "2020-01",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)
	assert.Empty(t, entry.Summary, "summary should default to empty string")
}

func TestGetWorkHistory_ReturnsSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		Summary:              "Responsible for API design and implementation.",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Responsible for API design and implementation.", got.Summary)
}

func TestUpdateWorkHistory_Summary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		Summary:              "Old summary",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	updated, err := store.UpdateWorkHistory(ctx, created.ID, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		Summary:              "New summary describing expanded responsibilities.",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)
	assert.Equal(t, "New summary describing expanded responsibilities.", updated.Summary)
}

func TestUpdateWorkHistory_ClearSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	created, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		Summary:              "Had a summary before",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.Summary)

	updated, err := store.UpdateWorkHistory(ctx, created.ID, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		Summary:              "",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)
	assert.Empty(t, updated.Summary, "summary should be clearable")
}

func TestListWorkHistory_IncludesSummary(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		Summary:   "Summary for Acme.",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)
	_, err = store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Beta", JobTitle: "Lead",
		Summary:   "",
		StartDate: "2021", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	entries, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	// Newest entry (Beta) is at the top; Acme has the summary.
	assert.Empty(t, entries[0].Summary)
	assert.Equal(t, "Summary for Acme.", entries[1].Summary)
}

// =================================================================
// Secondary Bullet (bullet_type) Tests
// =================================================================

func TestCreateBullet_DefaultBulletType(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	// Passing empty string should default to "primary"
	bullet, err := store.CreateBullet(ctx, entry.ID, "Default type bullet", "")
	require.NoError(t, err)
	assert.Equal(t, domain.BulletTypePrimary, bullet.BulletType)
}

func TestCreateBullet_PrimaryType(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	bullet, err := store.CreateBullet(ctx, entry.ID, "Primary bullet", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, domain.BulletTypePrimary, bullet.BulletType)
}

func TestCreateBullet_SecondaryType(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	bullet, err := store.CreateBullet(ctx, entry.ID, "Outcome bullet", domain.BulletTypeSecondary)
	require.NoError(t, err)
	assert.Equal(t, domain.BulletTypeSecondary, bullet.BulletType)
	assert.Equal(t, "Outcome bullet", bullet.Text)
	assert.Equal(t, 1, bullet.SortOrder)
}

func TestCreateBullet_InvalidBulletType(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	// The CHECK constraint should reject invalid bullet_type values.
	_, err = store.CreateBullet(ctx, entry.ID, "Bad type", "tertiary")
	require.Error(t, err, "invalid bullet_type should be rejected by CHECK constraint")
}

func TestCreateBullet_SortOrderPerBulletType(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	// Create primary bullets — they get sort_order 1, 2
	p1, err := store.CreateBullet(ctx, entry.ID, "Primary 1", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, 1, p1.SortOrder)

	p2, err := store.CreateBullet(ctx, entry.ID, "Primary 2", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, 2, p2.SortOrder)

	// Create secondary bullets — they start their own sequence at 1
	s1, err := store.CreateBullet(ctx, entry.ID, "Outcome 1", domain.BulletTypeSecondary)
	require.NoError(t, err)
	assert.Equal(t, 1, s1.SortOrder)

	s2, err := store.CreateBullet(ctx, entry.ID, "Outcome 2", domain.BulletTypeSecondary)
	require.NoError(t, err)
	assert.Equal(t, 2, s2.SortOrder)

	// Adding another primary should continue from 2
	p3, err := store.CreateBullet(ctx, entry.ID, "Primary 3", domain.BulletTypePrimary)
	require.NoError(t, err)
	assert.Equal(t, 3, p3.SortOrder)
}

func TestGetWorkHistory_BulletsOrderedByTypeAndSortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	// Create mixed bullet types
	_, err = store.CreateBullet(ctx, entry.ID, "Primary A", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Outcome X", domain.BulletTypeSecondary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Primary B", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Outcome Y", domain.BulletTypeSecondary)
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	require.Len(t, got.Bullets, 4)

	// ORDER BY bullet_type, sort_order means primary comes first
	assert.Equal(t, domain.BulletTypePrimary, got.Bullets[0].BulletType)
	assert.Equal(t, "Primary A", got.Bullets[0].Text)
	assert.Equal(t, domain.BulletTypePrimary, got.Bullets[1].BulletType)
	assert.Equal(t, "Primary B", got.Bullets[1].Text)
	assert.Equal(t, domain.BulletTypeSecondary, got.Bullets[2].BulletType)
	assert.Equal(t, "Outcome X", got.Bullets[2].Text)
	assert.Equal(t, domain.BulletTypeSecondary, got.Bullets[3].BulletType)
	assert.Equal(t, "Outcome Y", got.Bullets[3].Text)
}

func TestListWorkHistory_BulletsIncludeBulletType(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, entry.ID, "Primary bullet", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Secondary bullet", domain.BulletTypeSecondary)
	require.NoError(t, err)

	entries, err := store.ListWorkHistory(ctx)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Len(t, entries[0].Bullets, 2)

	// Primary comes first (alphabetically before "secondary")
	assert.Equal(t, domain.BulletTypePrimary, entries[0].Bullets[0].BulletType)
	assert.Equal(t, "Primary bullet", entries[0].Bullets[0].Text)
	assert.Equal(t, domain.BulletTypeSecondary, entries[0].Bullets[1].BulletType)
	assert.Equal(t, "Secondary bullet", entries[0].Bullets[1].Text)
}

func TestDeleteWorkHistory_CascadesBothBulletTypes(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, entry.ID, "Primary", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Secondary", domain.BulletTypeSecondary)
	require.NoError(t, err)

	err = store.DeleteWorkHistory(ctx, entry.ID)
	require.NoError(t, err)

	var count int
	err = store.db.QueryRow(
		"SELECT COUNT(*) FROM achievement_bullet WHERE work_history_id = ?",
		entry.ID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "all bullet types should cascade on delete")
}

func TestUpdateBullet_PreservesBulletType(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	bullet, err := store.CreateBullet(ctx, entry.ID, "Original outcome", domain.BulletTypeSecondary)
	require.NoError(t, err)
	assert.Equal(t, domain.BulletTypeSecondary, bullet.BulletType)

	// Updating text should not change bullet_type
	updated, err := store.UpdateBullet(ctx, bullet.ID, "Updated outcome")
	require.NoError(t, err)
	assert.Equal(t, "Updated outcome", updated.Text)
	assert.Equal(t, domain.BulletTypeSecondary, updated.BulletType, "bullet_type should be preserved on text update")
}

func TestReorderBullets_MixedTypes(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	p1, err := store.CreateBullet(ctx, entry.ID, "Primary A", domain.BulletTypePrimary)
	require.NoError(t, err)
	p2, err := store.CreateBullet(ctx, entry.ID, "Primary B", domain.BulletTypePrimary)
	require.NoError(t, err)
	s1, err := store.CreateBullet(ctx, entry.ID, "Outcome X", domain.BulletTypeSecondary)
	require.NoError(t, err)

	// Reorder all bullets (including mixed types)
	err = store.ReorderBullets(ctx, entry.ID, []int64{p2.ID, s1.ID, p1.ID})
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	require.Len(t, got.Bullets, 3)

	// After reorder, the ORDER BY is bullet_type, sort_order
	// p1 has sort_order 3 (primary), p2 has sort_order 1 (primary), s1 has sort_order 2 (secondary)
	// primary sorted by sort_order: p2(1), p1(3) then secondary: s1(2)
	assert.Equal(t, "Primary B", got.Bullets[0].Text)
	assert.Equal(t, 1, got.Bullets[0].SortOrder)
	assert.Equal(t, "Primary A", got.Bullets[1].Text)
	assert.Equal(t, 3, got.Bullets[1].SortOrder)
	assert.Equal(t, "Outcome X", got.Bullets[2].Text)
	assert.Equal(t, 2, got.Bullets[2].SortOrder)
}

func TestCreateBullet_SecondaryOnMultipleEntries(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	e1, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)
	e2, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Beta", JobTitle: "Lead",
		StartDate: "2021", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	// Each entry has its own secondary bullet sort_order sequence
	s1, err := store.CreateBullet(ctx, e1.ID, "E1 outcome", domain.BulletTypeSecondary)
	require.NoError(t, err)
	assert.Equal(t, 1, s1.SortOrder)

	s2, err := store.CreateBullet(ctx, e2.ID, "E2 outcome", domain.BulletTypeSecondary)
	require.NoError(t, err)
	assert.Equal(t, 1, s2.SortOrder, "secondary bullet sort_order should be per-entry")
}

func TestCreateWorkHistory_SummaryWithSpecialCharacters(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	summary := "Led team with 5+ engineers & 3 contractors. Managed $2M budget; achieved 99.9% uptime."
	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "VP Engineering",
		Summary:              summary,
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	assert.Equal(t, summary, got.Summary, "special characters in summary should roundtrip correctly")
}

func TestCreateWorkHistory_SummaryWithNewlines(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	summary := "First paragraph.\n\nSecond paragraph."
	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Engineer",
		Summary:              summary,
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	assert.Equal(t, summary, got.Summary, "newlines in summary should be preserved")
}

func TestDeleteBullet_SecondaryType(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	primary, err := store.CreateBullet(ctx, entry.ID, "Keep this", domain.BulletTypePrimary)
	require.NoError(t, err)
	secondary, err := store.CreateBullet(ctx, entry.ID, "Delete this", domain.BulletTypeSecondary)
	require.NoError(t, err)

	err = store.DeleteBullet(ctx, secondary.ID)
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	require.Len(t, got.Bullets, 1)
	assert.Equal(t, primary.ID, got.Bullets[0].ID)
	assert.Equal(t, domain.BulletTypePrimary, got.Bullets[0].BulletType)
}

func TestGetWorkHistory_OnlyPrimaryBullets(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, entry.ID, "P1", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "P2", domain.BulletTypePrimary)
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	require.Len(t, got.Bullets, 2)
	for _, b := range got.Bullets {
		assert.Equal(t, domain.BulletTypePrimary, b.BulletType)
	}
}

func TestGetWorkHistory_OnlySecondaryBullets(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, entry.ID, "S1", domain.BulletTypeSecondary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "S2", domain.BulletTypeSecondary)
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	require.Len(t, got.Bullets, 2)
	for _, b := range got.Bullets {
		assert.Equal(t, domain.BulletTypeSecondary, b.BulletType)
	}
}

func TestGetWorkHistory_NoBullets(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName: "Acme", JobTitle: "Dev",
		StartDate: "2020", DateGranularityStart: "year",
	})
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Bullets, "no bullets means empty slice from listBulletsForEntry")
	assert.NotNil(t, got.Bullets, "bullets should be empty slice, not nil (JSON serializes nil as null)")
}

func TestGetWorkHistory_SummaryAndBulletsTogether(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	entry, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Acme",
		JobTitle:             "Senior Dev",
		Summary:              "Technical lead for platform team.",
		StartDate:            "2020",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)

	_, err = store.CreateBullet(ctx, entry.ID, "Built APIs", domain.BulletTypePrimary)
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Increased revenue 20%", domain.BulletTypeSecondary)
	require.NoError(t, err)

	got, err := store.GetWorkHistory(ctx, entry.ID)
	require.NoError(t, err)
	assert.Equal(t, "Technical lead for platform team.", got.Summary)
	require.Len(t, got.Bullets, 2)
	assert.Equal(t, domain.BulletTypePrimary, got.Bullets[0].BulletType)
	assert.Equal(t, domain.BulletTypeSecondary, got.Bullets[1].BulletType)
}
