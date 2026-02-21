# Fix: Missing `summary` column on `work_history_entry`

## Root Cause

The user's database has `PRAGMA user_version = 2` but is **missing V2 columns**:
- `work_history_entry` lacks `summary`
- `achievement_bullet` lacks `bullet_type`

The lens-related V2 changes (create `lens_summary_selection`, rebuild `lens` table) **were** applied. The V2 migration partially applied: the ALTER TABLE ADD COLUMN statements didn't take effect, but `user_version` was set to 2. Since `Migrate()` checks `user_version >= currentVersion (2)` and skips, the columns are never added, causing the `ListWorkHistory` query to fail with `no such column: summary`.

## Plan

### 1. Write regression test reproducing the bug

File: `internal/infra/sqlite/migrations_test.go`

Add a helper `simulatePartialV2(t, store)` that applies only the lens migration parts of V2 (junction table + lens rebuild) without the ALTER TABLE ADD COLUMN statements, then sets `user_version = 2`.

Add `TestBug_MissingSummaryColumn_V2PartialApply`:
- Run migrateV1
- Insert a work_history_entry using V1 schema
- Call `simulatePartialV2` to create the broken state
- Verify `user_version = 2`
- Call `ListWorkHistory` and assert it errors with "summary"

### 2. Create `migrateV3` in `migrations.go`

Add a `columnExists(db, table, column)` helper using `PRAGMA table_info(table)`.

Add `migrateV3(store)`:
- Begin transaction
- Check if `summary` column exists on `work_history_entry`; if not, `ALTER TABLE work_history_entry ADD COLUMN summary TEXT NOT NULL DEFAULT ''`
- Check if `bullet_type` column exists on `achievement_bullet`; if not, `ALTER TABLE achievement_bullet ADD COLUMN bullet_type TEXT NOT NULL DEFAULT 'primary' CHECK(bullet_type IN ('primary', 'secondary'))`
- Set `PRAGMA user_version = 3`
- Commit

### 3. Update migration infrastructure

In `migrations.go`:
- Change `currentVersion` from `2` to `3`
- Add `migrateV3` to the `migrations` slice

### 4. Write V3 migration tests

Add to `migrations_test.go`:

**`TestV3Fix_RepairsMissingSummaryColumn`**:
- Set up broken state (V1 + simulatePartialV2)
- Insert data before repair
- Call `Migrate()` (which now includes V3)
- Assert `ListWorkHistory` works, summary defaults to `""`, bullet_type defaults to `"primary"`
- Assert full CRUD with summary and bullet_type works

**`TestV3_NoOpOnHealthyDatabase`**:
- Run full V1 + V2 (columns already exist)
- Insert data using V2 columns
- Run `migrateV3`
- Assert version is 3, existing data is intact

### 5. Run full test suite

Run `go test ./...` to verify all existing tests still pass with the new `currentVersion = 3`.

## Files Modified

- `internal/infra/sqlite/migrations.go` — add `columnExists`, `migrateV3`, bump `currentVersion`
- `internal/infra/sqlite/migrations_test.go` — add `simulatePartialV2`, 3 new test functions
