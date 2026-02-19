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

func TestCreateWorkHistory_AutoIncrementsSortOrder(t *testing.T) {
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

	second, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "Second Co",
		JobTitle:             "Dev",
		StartDate:            "2021",
		DateGranularityStart: "year",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, second.SortOrder)
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

	_, err = store.CreateBullet(ctx, entry.ID, "Led project X")
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Improved perf by 50%")
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

	_, err = store.CreateBullet(ctx, entry.ID, "Bullet 1")
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Bullet 2")
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
	assert.Equal(t, "First", entries[0].EmployerName)
	assert.Equal(t, "Second", entries[1].EmployerName)
	assert.Equal(t, "Third", entries[2].EmployerName)
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

	_, err = store.CreateBullet(ctx, entry.ID, "Bullet A")
	require.NoError(t, err)
	_, err = store.CreateBullet(ctx, entry.ID, "Bullet B")
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

	bullet, err := store.CreateBullet(ctx, entry.ID, "Led a team of 5 engineers")
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

	b1, err := store.CreateBullet(ctx, entry.ID, "First")
	require.NoError(t, err)
	assert.Equal(t, 1, b1.SortOrder)

	b2, err := store.CreateBullet(ctx, entry.ID, "Second")
	require.NoError(t, err)
	assert.Equal(t, 2, b2.SortOrder)

	b3, err := store.CreateBullet(ctx, entry.ID, "Third")
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

	b1, err := store.CreateBullet(ctx, e1.ID, "E1 bullet")
	require.NoError(t, err)
	assert.Equal(t, 1, b1.SortOrder)

	// Second entry starts its own sort_order sequence at 1.
	b2, err := store.CreateBullet(ctx, e2.ID, "E2 bullet")
	require.NoError(t, err)
	assert.Equal(t, 1, b2.SortOrder)
}

func TestCreateBullet_InvalidWorkHistoryID(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.CreateBullet(ctx, 999, "orphan bullet")
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

	bullet, err := store.CreateBullet(ctx, entry.ID, "Old text")
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

	bullet, err := store.CreateBullet(ctx, entry.ID, "To be deleted")
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

	b1, err := store.CreateBullet(ctx, entry.ID, "Alpha")
	require.NoError(t, err)
	b2, err := store.CreateBullet(ctx, entry.ID, "Beta")
	require.NoError(t, err)
	b3, err := store.CreateBullet(ctx, entry.ID, "Gamma")
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
