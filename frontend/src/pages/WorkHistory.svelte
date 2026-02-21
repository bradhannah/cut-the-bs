<script lang="ts">
  import { onMount } from "svelte";
  import {
    listWorkHistory,
    createWorkHistory,
    updateWorkHistory,
    deleteWorkHistory,
    reorderWorkHistory,
    reorderBullets,
    createBullet,
    updateBullet,
    deleteBullet,
    checkWorkHistoryLensReferences,
    addToast,
    type WorkHistoryEntry,
    type WorkHistoryInput,
  } from "../services/api";
  import WorkHistoryCard from "../components/WorkHistoryCard.svelte";
  import DateInput from "../components/DateInput.svelte";
  import BulletPasteDialog from "../components/BulletPasteDialog.svelte";
  import DragHandle from "../components/DragHandle.svelte";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";

  let entries: WorkHistoryEntry[] = [];
  let loading = true;

  // Track which cards are expanded (survives data refreshes)
  let expandedIds = new Set<number>();

  function toggleExpanded(id: number): void {
    if (expandedIds.has(id)) {
      expandedIds.delete(id);
    } else {
      expandedIds.add(id);
    }
    expandedIds = expandedIds; // trigger reactivity
  }

  // Form state
  let showForm = false;
  let formEmployerName = "";
  let formJobTitle = "";
  let formSummary = "";
  let formStartDate = "";
  let formStartGranularity = "month";
  let formEndDate = "";
  let formEndGranularity = "month";
  let formEndIsPresent = false;

  // Paste dialog state
  let showPasteDialog = false;
  let pasteWorkHistoryId = 0;
  let pasteBulletType = "primary";

  // Delete confirmation state
  let deleteConfirmEntry: WorkHistoryEntry | null = null;
  let lensReferences: string[] = [];

  onMount(async () => {
    await loadEntries();
  });

  async function loadEntries(): Promise<void> {
    loading = true;
    try {
      entries = (await listWorkHistory()) || [];
    } finally {
      loading = false;
    }
  }

  /** Reload entries without toggling loading (preserves card instances). */
  async function refreshEntries(): Promise<void> {
    try {
      entries = (await listWorkHistory()) || [];
    } catch {
      // Toast already shown by api.ts
    }
  }

  // --- Entry CRUD ---

  function openAddForm(): void {
    formEmployerName = "";
    formJobTitle = "";
    formSummary = "";
    formStartDate = "";
    formStartGranularity = "month";
    formEndDate = "";
    formEndGranularity = "month";
    formEndIsPresent = false;
    showForm = true;
  }

  async function handleEntrySave(id: number, input: WorkHistoryInput): Promise<void> {
    try {
      await updateWorkHistory(id, input);
      await refreshEntries();
    } catch {
      // Toast already shown by api.ts
    }
  }

  function cancelForm(): void {
    showForm = false;
  }

  async function handleSubmit(): Promise<void> {
    const input: WorkHistoryInput = {
      employer_name: formEmployerName.trim(),
      job_title: formJobTitle.trim(),
      summary: formSummary.trim(),
      start_date: formStartDate,
      end_date: formEndIsPresent ? "" : formEndDate,
      date_granularity_start: formStartGranularity,
      date_granularity_end: formEndGranularity,
    };

    if (!input.employer_name || !input.job_title) {
      addToast("error", "Employer name and job title are required");
      return;
    }

    if (!input.start_date) {
      addToast("error", "Start date is required");
      return;
    }

    try {
      await createWorkHistory(input);
      showForm = false;
      await refreshEntries();
    } catch {
      // Toast already shown by api.ts
    }
  }

  async function confirmDeleteEntry(entry: WorkHistoryEntry): Promise<void> {
    try {
      lensReferences = (await checkWorkHistoryLensReferences(entry.id)) || [];
      deleteConfirmEntry = entry;
    } catch {
      // Toast already shown
    }
  }

  async function handleDeleteEntry(): Promise<void> {
    if (!deleteConfirmEntry) return;
    try {
      await deleteWorkHistory(deleteConfirmEntry.id);
      deleteConfirmEntry = null;
      lensReferences = [];
      await refreshEntries();
    } catch {
      // Toast already shown
    }
  }

  function cancelDelete(): void {
    deleteConfirmEntry = null;
    lensReferences = [];
  }

  // --- Bullet CRUD ---

  async function handleBulletCreate(
    workHistoryId: number,
    text: string,
    bulletType: string = "primary"
  ): Promise<void> {
    try {
      await createBullet(workHistoryId, text, bulletType);
      addToast("success", "Bullet added");
      await refreshEntries();
    } catch {
      // Toast already shown
    }
  }

  async function handleBulletUpdate(id: number, text: string): Promise<void> {
    try {
      await updateBullet(id, text);
      addToast("success", "Bullet updated");
      await refreshEntries();
    } catch {
      // Toast already shown
    }
  }

  async function handleBulletDelete(id: number): Promise<void> {
    try {
      await deleteBullet(id);
      addToast("success", "Bullet deleted");
      await refreshEntries();
    } catch {
      // Toast already shown
    }
  }

  // --- Entry Reorder ---

  async function handleEntryReorder(orderedIDs: number[]): Promise<void> {
    try {
      await reorderWorkHistory(orderedIDs);
      await refreshEntries();
    } catch {
      // Toast already shown
    }
  }

  // --- Bullet Reorder ---

  async function handleBulletReorder(
    workHistoryId: number,
    orderedIDs: number[]
  ): Promise<void> {
    try {
      await reorderBullets(workHistoryId, orderedIDs);
      await refreshEntries();
    } catch {
      // Toast already shown
    }
  }

  // --- Paste Dialog ---

  function openPasteDialog(workHistoryId: number, bulletType: string = "primary"): void {
    pasteWorkHistoryId = workHistoryId;
    pasteBulletType = bulletType;
    showPasteDialog = true;
  }

  function closePasteDialog(): void {
    showPasteDialog = false;
    pasteWorkHistoryId = 0;
    pasteBulletType = "primary";
  }

  async function handlePasteConfirm(
    workHistoryId: number,
    lines: string[],
    bulletType: string = "primary"
  ): Promise<void> {
    showPasteDialog = false;
    try {
      for (const line of lines) {
        await createBullet(workHistoryId, line, bulletType);
      }
      addToast(
        "success",
        `Added ${lines.length} bullet${lines.length !== 1 ? "s" : ""}`
      );
      await refreshEntries();
    } catch {
      // Toast already shown — partial creation may have occurred
      await refreshEntries();
    }
  }

  function handleStartDateChange(
    event: CustomEvent<{ value: string; granularity: string }>
  ): void {
    formStartDate = event.detail.value;
    formStartGranularity = event.detail.granularity;
  }

  function handleEndDateChange(
    event: CustomEvent<{ value: string; granularity: string }>
  ): void {
    formEndDate = event.detail.value;
    formEndGranularity = event.detail.granularity;
  }

  $: sortedEntries = [...entries].sort((a, b) => a.sort_order - b.sort_order);

  // Helper to look up a full entry by ID (for DragHandle slot typing)
  function getEntry(id: number): WorkHistoryEntry {
    return entries.find((e) => e.id === id) as WorkHistoryEntry;
  }
</script>

<div class="work-history-page">
  <div class="page-header">
    <h2>Work History</h2>
    <div class="page-header-actions">
      <button class="btn btn-primary" on:click={openAddForm}>
        + Add Entry
      </button>
    </div>
  </div>
  <p class="page-description">
    Manage your employment history and achievement bullets.
  </p>

  {#if showForm}
    <div class="entry-form">
      <h3 class="form-title">
        New Work History Entry
      </h3>

      <div class="form-row">
        <div class="form-field">
          <label class="form-label" for="employer-name">Employer</label>
          <input
            id="employer-name"
            type="text"
            class="form-input"
            bind:value={formEmployerName}
            placeholder="Company name"
          />
        </div>
        <div class="form-field">
          <label class="form-label" for="job-title">Job Title</label>
          <input
            id="job-title"
            type="text"
            class="form-input"
            bind:value={formJobTitle}
            placeholder="Your role"
          />
        </div>
      </div>

      <div class="form-row">
        <div class="form-field">
          <label class="form-label" for="summary">Summary <span class="optional-hint">(optional)</span></label>
          <textarea
            id="summary"
            class="form-input form-textarea"
            bind:value={formSummary}
            placeholder="Brief overview of your role and impact at this company"
            rows="3"
          ></textarea>
        </div>
      </div>

      <div class="form-row">
        <DateInput
          label="Start Date"
          value={formStartDate}
          granularity={formStartGranularity}
          on:change={handleStartDateChange}
        />
        <DateInput
          label="End Date"
          value={formEndDate}
          granularity={formEndGranularity}
          allowPresent={true}
          isPresent={formEndIsPresent}
          on:change={handleEndDateChange}
        />
      </div>

      <div class="form-actions">
        <button class="btn btn-primary" on:click={handleSubmit}>
          Create
        </button>
        <button class="btn btn-cancel" on:click={cancelForm}>Cancel</button>
      </div>
    </div>
  {/if}

  <!-- Delete Confirmation -->
  {#if deleteConfirmEntry}
    <div class="confirm-dialog">
      <p>
        Delete <strong>{deleteConfirmEntry.employer_name} — {deleteConfirmEntry.job_title}</strong>?
      </p>
      {#if lensReferences.length > 0}
        <p class="warn-text">
          This entry is referenced by {lensReferences.length} lens{lensReferences.length !== 1 ? "es" : ""}:
          {lensReferences.join(", ")}
        </p>
      {/if}
      <div class="form-actions">
        <button class="btn btn-danger-solid" on:click={handleDeleteEntry}>
          Delete
        </button>
        <button class="btn btn-cancel" on:click={cancelDelete}>Cancel</button>
      </div>
    </div>
  {/if}

  {#if loading}
    <LoadingSpinner />
  {:else if sortedEntries.length === 0}
    <div class="empty-state">
      <p>No work history entries yet.</p>
      <p class="empty-hint">
        Click "+ Add Entry" to add your first position.
      </p>
    </div>
  {:else}
    <div class="entries-list">
      <DragHandle
        items={sortedEntries}
        on:reorder={(e) => handleEntryReorder(e.detail.orderedIDs)}
        let:item
      >
        <WorkHistoryCard
          entry={getEntry(item.id)}
          expanded={expandedIds.has(item.id)}
          on:toggle={(e) => toggleExpanded(e.detail.id)}
          on:save={(e) => handleEntrySave(e.detail.id, e.detail.input)}
          on:delete={(e) => confirmDeleteEntry(getEntry(e.detail.id))}
          on:bulletCreate={(e) =>
            handleBulletCreate(e.detail.workHistoryId, e.detail.text, e.detail.bulletType)}
          on:bulletUpdate={(e) =>
            handleBulletUpdate(e.detail.id, e.detail.text)}
          on:bulletDelete={(e) => handleBulletDelete(e.detail.id)}
          on:bulletReorder={(e) =>
            handleBulletReorder(e.detail.workHistoryId, e.detail.orderedIDs)}
          on:bulletPaste={(e) => openPasteDialog(e.detail.workHistoryId, e.detail.bulletType)}
        />
      </DragHandle>
    </div>
  {/if}

  {#if showPasteDialog}
    <BulletPasteDialog
      workHistoryId={pasteWorkHistoryId}
      bulletType={pasteBulletType}
      on:confirm={(e) =>
        handlePasteConfirm(e.detail.workHistoryId, e.detail.lines, e.detail.bulletType)}
      on:cancel={closePasteDialog}
    />
  {/if}
</div>

<style>
  .work-history-page {
    max-width: 800px;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .page-header h2 {
    margin: 0;
    font-size: 1.5rem;
    color: #e0e0e0;
  }

  .page-header-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .page-description {
    color: #7a8a9a;
    font-size: 0.95rem;
    margin: 0 0 24px;
  }

  .empty-state {
    text-align: center;
    padding: 48px 0;
    color: #5a6a7a;
  }

  .empty-hint {
    font-size: 0.85rem;
    margin-top: 8px;
  }

  .entries-list {
    display: flex;
    flex-direction: column;
  }

  /* --- Form --- */
  .entry-form {
    background-color: #1e2d3d;
    border: 1px solid #3a4a5a;
    border-radius: 6px;
    padding: 20px;
    margin-bottom: 24px;
  }

  .form-title {
    margin: 0 0 16px;
    font-size: 1rem;
    color: #e0e0e0;
  }

  .form-row {
    display: flex;
    gap: 16px;
    margin-bottom: 16px;
  }

  .form-field {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .form-label {
    font-size: 0.8rem;
    color: #7a8a9a;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .form-input {
    background-color: #1a2332;
    color: #e0e0e0;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 8px 10px;
    font-size: 0.9rem;
  }

  .form-input:focus {
    outline: none;
    border-color: #4a8af4;
    box-shadow: 0 0 0 2px rgba(74, 138, 244, 0.15);
  }

  .form-textarea {
    resize: vertical;
    font-family: inherit;
    line-height: 1.4;
    min-height: 60px;
    max-height: 300px;
  }

  .optional-hint {
    font-weight: 400;
    text-transform: none;
    color: #5a6a7a;
    font-size: 0.75rem;
  }

  .form-actions {
    display: flex;
    gap: 8px;
  }

  /* --- Buttons --- */
  .btn {
    border: 1px solid #3a4a5a;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.85rem;
    padding: 8px 16px;
    transition:
      background-color 0.15s,
      border-color 0.15s;
  }

  .btn-primary {
    background-color: #2a5090;
    border-color: #3a60a0;
    color: #e0e0e0;
  }

  .btn-primary:hover {
    background-color: #3a60a0;
  }

  .btn-cancel {
    background-color: transparent;
    border-color: #3a4a5a;
    color: #7a8a9a;
  }

  .btn-cancel:hover {
    background-color: #2a3a4a;
    color: #c0d0e0;
  }

  .btn-danger-solid {
    background-color: #802020;
    border-color: #a03030;
    color: #e0e0e0;
  }

  .btn-danger-solid:hover {
    background-color: #a03030;
  }

  /* --- Confirm Dialog --- */
  .confirm-dialog {
    background-color: #2a1a1a;
    border: 1px solid #5a3030;
    border-radius: 6px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .confirm-dialog p {
    margin: 0 0 8px;
    color: #e0e0e0;
    font-size: 0.9rem;
  }

  .warn-text {
    color: #e0a060 !important;
    font-size: 0.85rem !important;
  }
</style>
