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
    addToast,
    type WorkHistoryEntry,
    type WorkHistoryInput,
  } from "../services/api";
  import WorkHistoryCard from "../components/WorkHistoryCard.svelte";
  import DateInput from "../components/DateInput.svelte";
  import BulletPasteDialog from "../components/BulletPasteDialog.svelte";
  import DragHandle from "../components/DragHandle.svelte";

  let entries: WorkHistoryEntry[] = [];
  let loading = true;

  // Form state
  let showForm = false;
  let editingEntry: WorkHistoryEntry | null = null;
  let formEmployerName = "";
  let formJobTitle = "";
  let formStartDate = "";
  let formStartGranularity = "month";
  let formEndDate = "";
  let formEndGranularity = "month";
  let formEndIsPresent = false;

  // Paste dialog state
  let showPasteDialog = false;
  let pasteWorkHistoryId = 0;

  onMount(async () => {
    await loadEntries();
  });

  async function loadEntries(): Promise<void> {
    loading = true;
    try {
      entries = await listWorkHistory();
    } finally {
      loading = false;
    }
  }

  // --- Entry CRUD ---

  function openAddForm(): void {
    editingEntry = null;
    formEmployerName = "";
    formJobTitle = "";
    formStartDate = "";
    formStartGranularity = "month";
    formEndDate = "";
    formEndGranularity = "month";
    formEndIsPresent = false;
    showForm = true;
  }

  function openEditForm(entry: WorkHistoryEntry): void {
    editingEntry = entry;
    formEmployerName = entry.employer_name;
    formJobTitle = entry.job_title;
    formStartDate = entry.start_date;
    formStartGranularity = entry.date_granularity_start || "month";
    formEndDate = entry.end_date;
    formEndGranularity = entry.date_granularity_end || "month";
    formEndIsPresent = !entry.end_date;
    showForm = true;
  }

  function cancelForm(): void {
    showForm = false;
    editingEntry = null;
  }

  async function handleSubmit(): Promise<void> {
    const input: WorkHistoryInput = {
      employer_name: formEmployerName.trim(),
      job_title: formJobTitle.trim(),
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
      if (editingEntry) {
        await updateWorkHistory(editingEntry.id, input);
      } else {
        await createWorkHistory(input);
      }
      showForm = false;
      editingEntry = null;
      await loadEntries();
    } catch {
      // Toast already shown by api.ts
    }
  }

  async function handleDeleteEntry(id: number): Promise<void> {
    try {
      await deleteWorkHistory(id);
      await loadEntries();
    } catch {
      // Toast already shown
    }
  }

  // --- Bullet CRUD ---

  async function handleBulletCreate(
    workHistoryId: number,
    text: string
  ): Promise<void> {
    try {
      await createBullet(workHistoryId, text);
      addToast("success", "Bullet added");
      await loadEntries();
    } catch {
      // Toast already shown
    }
  }

  async function handleBulletUpdate(id: number, text: string): Promise<void> {
    try {
      await updateBullet(id, text);
      addToast("success", "Bullet updated");
      await loadEntries();
    } catch {
      // Toast already shown
    }
  }

  async function handleBulletDelete(id: number): Promise<void> {
    try {
      await deleteBullet(id);
      addToast("success", "Bullet deleted");
      await loadEntries();
    } catch {
      // Toast already shown
    }
  }

  // --- Entry Reorder ---

  async function handleEntryReorder(orderedIDs: number[]): Promise<void> {
    try {
      await reorderWorkHistory(orderedIDs);
      await loadEntries();
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
      await loadEntries();
    } catch {
      // Toast already shown
    }
  }

  // --- Paste Dialog ---

  function openPasteDialog(workHistoryId: number): void {
    pasteWorkHistoryId = workHistoryId;
    showPasteDialog = true;
  }

  function closePasteDialog(): void {
    showPasteDialog = false;
    pasteWorkHistoryId = 0;
  }

  async function handlePasteConfirm(
    workHistoryId: number,
    lines: string[]
  ): Promise<void> {
    showPasteDialog = false;
    try {
      for (const line of lines) {
        await createBullet(workHistoryId, line);
      }
      addToast(
        "success",
        `Added ${lines.length} bullet${lines.length !== 1 ? "s" : ""}`
      );
      await loadEntries();
    } catch {
      // Toast already shown — partial creation may have occurred
      await loadEntries();
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
    <button class="btn btn-primary" on:click={openAddForm}>
      + Add Entry
    </button>
  </div>
  <p class="page-description">
    Manage your employment history and achievement bullets.
  </p>

  {#if showForm}
    <div class="entry-form">
      <h3 class="form-title">
        {editingEntry ? "Edit Entry" : "New Work History Entry"}
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
          {editingEntry ? "Update" : "Create"}
        </button>
        <button class="btn btn-cancel" on:click={cancelForm}>Cancel</button>
      </div>
    </div>
  {/if}

  {#if loading}
    <p class="loading-message">Loading...</p>
  {:else if sortedEntries.length === 0}
    <div class="empty-state">
      <p>No work history entries yet.</p>
      <p class="empty-hint">Click "+ Add Entry" to add your first position.</p>
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
          on:edit={(e) => openEditForm(e.detail.entry)}
          on:delete={(e) => handleDeleteEntry(e.detail.id)}
          on:bulletCreate={(e) =>
            handleBulletCreate(e.detail.workHistoryId, e.detail.text)}
          on:bulletUpdate={(e) =>
            handleBulletUpdate(e.detail.id, e.detail.text)}
          on:bulletDelete={(e) => handleBulletDelete(e.detail.id)}
          on:bulletReorder={(e) =>
            handleBulletReorder(e.detail.workHistoryId, e.detail.orderedIDs)}
          on:bulletPaste={(e) => openPasteDialog(e.detail.workHistoryId)}
        />
      </DragHandle>
    </div>
  {/if}

  {#if showPasteDialog}
    <BulletPasteDialog
      workHistoryId={pasteWorkHistoryId}
      on:confirm={(e) =>
        handlePasteConfirm(e.detail.workHistoryId, e.detail.lines)}
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

  .page-description {
    color: #7a8a9a;
    font-size: 0.95rem;
    margin: 0 0 24px;
  }

  .loading-message {
    color: #5a6a7a;
    font-style: italic;
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
</style>
