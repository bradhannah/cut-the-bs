<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { WorkHistoryEntry, WorkHistoryInput } from "../services/api";
  import { formatDateRange } from "../services/dateFormat";
  import BulletList from "./BulletList.svelte";
  import DateInput from "./DateInput.svelte";

  export let entry: WorkHistoryEntry;
  export let expanded = false;

  const dispatch = createEventDispatcher<{
    toggle: { id: number };
    save: { id: number; input: WorkHistoryInput };
    delete: { id: number };
    bulletCreate: { workHistoryId: number; text: string; bulletType: string };
    bulletUpdate: { id: number; text: string };
    bulletDelete: { id: number };
    bulletReorder: { workHistoryId: number; orderedIDs: number[] };
    bulletPaste: { workHistoryId: number; bulletType: string };
  }>();

  // Inline edit form state
  let editing = false;
  let formEmployerName = "";
  let formJobTitle = "";
  let formSummary = "";
  let formStartDate = "";
  let formStartGranularity = "month";
  let formEndDate = "";
  let formEndGranularity = "month";
  let formEndIsPresent = false;

  function toggle(): void {
    if (editing) return; // don't collapse while editing
    dispatch("toggle", { id: entry.id });
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      toggle();
    }
  }

  function startEdit(): void {
    formEmployerName = entry.employer_name;
    formJobTitle = entry.job_title;
    formSummary = entry.summary || "";
    formStartDate = entry.start_date;
    formStartGranularity = entry.date_granularity_start || "month";
    formEndDate = entry.end_date;
    formEndGranularity = entry.date_granularity_end || "month";
    formEndIsPresent = !entry.end_date;
    editing = true;
    if (!expanded) {
      dispatch("toggle", { id: entry.id });
    }
  }

  function closeEdit(): void {
    editing = false;
  }

  function cancelEdit(): void {
    editing = false;
  }

  function saveEdit(): void {
    const input: WorkHistoryInput = {
      employer_name: formEmployerName.trim(),
      job_title: formJobTitle.trim(),
      summary: formSummary.trim(),
      start_date: formStartDate,
      end_date: formEndIsPresent ? "" : formEndDate,
      date_granularity_start: formStartGranularity,
      date_granularity_end: formEndGranularity,
    };
    dispatch("save", { id: entry.id, input });
    editing = false;
  }

  function handleDelete(): void {
    dispatch("delete", { id: entry.id });
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

  function onBulletCreate(
    event: CustomEvent<{ workHistoryId: number; text: string; bulletType: string }>
  ): void {
    dispatch("bulletCreate", event.detail);
  }

  function onBulletUpdate(
    event: CustomEvent<{ id: number; text: string }>
  ): void {
    dispatch("bulletUpdate", event.detail);
  }

  function onBulletDelete(event: CustomEvent<{ id: number }>): void {
    dispatch("bulletDelete", event.detail);
  }

  function onBulletReorder(
    event: CustomEvent<{ workHistoryId: number; orderedIDs: number[] }>
  ): void {
    dispatch("bulletReorder", event.detail);
  }

  function onBulletPaste(event: CustomEvent<{ workHistoryId: number; bulletType: string }>): void {
    dispatch("bulletPaste", event.detail);
  }

  $: dateRange = formatDateRange(entry.start_date, entry.date_granularity_start, entry.end_date, entry.date_granularity_end);
  $: bulletCount = entry.bullets ? entry.bullets.length : 0;
</script>

<div class="card" class:expanded>
  <div
    class="card-header"
    on:click={toggle}
    on:keydown={handleKeydown}
    role="button"
    tabindex="0"
    aria-expanded={expanded}
  >
    <div class="card-expand-icon">
      <span class="chevron">{expanded ? "\u25BC" : "\u25B6"}</span>
    </div>
    <div class="card-info">
      <div class="card-title">
        <strong>{entry.job_title}</strong>
        <span class="card-employer">at {entry.employer_name}</span>
      </div>
      <div class="card-meta">
        <span class="card-dates">{dateRange}</span>
        <span class="card-bullet-count">
          {bulletCount} bullet{bulletCount !== 1 ? "s" : ""}
        </span>
      </div>
    </div>
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <div class="card-actions" on:click|stopPropagation={() => {}}>
      <button
        class="btn-text"
        on:click={editing ? closeEdit : startEdit}
        title={editing ? "Close editor" : "Edit entry"}
      >
        [{editing ? "Close" : "Edit"}]
      </button>
      <button
        class="btn-icon btn-icon-danger"
        on:click={handleDelete}
        title="Delete entry"
      >
        &#10005;
      </button>
    </div>
  </div>

  {#if expanded}
    <div class="card-body">
      {#if editing}
        <!-- ======== INLINE EDIT FORM ======== -->
        <div class="inline-edit-form">
          <div class="form-row">
            <div class="form-field">
              <label class="form-label" for="edit-employer-{entry.id}">Employer</label>
              <input
                id="edit-employer-{entry.id}"
                type="text"
                class="form-input"
                bind:value={formEmployerName}
                placeholder="Company name"
              />
            </div>
            <div class="form-field">
              <label class="form-label" for="edit-title-{entry.id}">Job Title</label>
              <input
                id="edit-title-{entry.id}"
                type="text"
                class="form-input"
                bind:value={formJobTitle}
                placeholder="Your role"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-field">
              <label class="form-label" for="edit-summary-{entry.id}">
                Summary <span class="optional-hint">(optional)</span>
              </label>
              <textarea
                id="edit-summary-{entry.id}"
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
            <button class="btn btn-primary" on:click={saveEdit}>Save</button>
            <button class="btn btn-cancel" on:click={cancelEdit}>Cancel</button>
          </div>
        </div>

        <!-- Bullet list below the edit form -->
        <BulletList
          bullets={entry.bullets || []}
          workHistoryId={entry.id}
          on:create={onBulletCreate}
          on:update={onBulletUpdate}
          on:delete={onBulletDelete}
          on:reorder={onBulletReorder}
          on:paste={onBulletPaste}
        />
      {:else}
        <!-- ======== DEFAULT COLLAPSED-EXPANDED VIEW ======== -->
        {#if entry.summary}
          <p class="entry-summary">{entry.summary}</p>
        {/if}
        <BulletList
          bullets={entry.bullets || []}
          workHistoryId={entry.id}
          on:create={onBulletCreate}
          on:update={onBulletUpdate}
          on:delete={onBulletDelete}
          on:reorder={onBulletReorder}
          on:paste={onBulletPaste}
        />
      {/if}
    </div>
  {/if}
</div>

<style>
  .card {
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    margin-bottom: 8px;
    transition: border-color 0.15s;
  }

  .card:hover {
    border-color: #3a4a5a;
  }

  .card.expanded {
    border-color: #4a8af4;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    cursor: pointer;
    user-select: none;
  }

  .card-header:focus {
    outline: none;
    box-shadow: inset 0 0 0 2px rgba(74, 138, 244, 0.3);
    border-radius: 6px;
  }

  .card-expand-icon {
    flex-shrink: 0;
    width: 16px;
    text-align: center;
  }

  .chevron {
    color: #5a6a7a;
    font-size: 0.7rem;
    transition: color 0.15s;
  }

  .card.expanded .chevron {
    color: #4a8af4;
  }

  .card-info {
    flex: 1;
    min-width: 0;
  }

  .card-title {
    font-size: 0.95rem;
    color: #e0e0e0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .card-employer {
    color: #7a8a9a;
    font-weight: 400;
  }

  .card-meta {
    display: flex;
    gap: 16px;
    margin-top: 2px;
    font-size: 0.8rem;
    color: #5a6a7a;
  }

  .card-dates {
    color: #7a8a9a;
  }

  .card-bullet-count {
    color: #5a6a7a;
  }

  .card-actions {
    display: flex;
    gap: 2px;
    flex-shrink: 0;
    align-items: center;
  }

  .btn-text {
    background: none;
    border: none;
    color: #7a8a9a;
    cursor: pointer;
    padding: 4px 8px;
    font-size: 0.82rem;
    border-radius: 3px;
    font-family: inherit;
  }

  .btn-text:hover {
    background-color: #2a3a4a;
    color: #e0e0e0;
  }

  .btn-icon {
    background: none;
    border: none;
    color: #7a8a9a;
    cursor: pointer;
    padding: 4px 8px;
    font-size: 0.85rem;
    border-radius: 3px;
  }

  .btn-icon:hover {
    background-color: #2a3a4a;
    color: #e0e0e0;
  }

  .btn-icon-danger:hover {
    background-color: #5a2020;
    color: #ff6b6b;
  }

  .card-body {
    padding: 0 14px 14px 40px;
    border-top: 1px solid #2a3a4a;
  }

  .entry-summary {
    margin: 12px 0 8px;
    padding: 8px 12px;
    font-size: 0.88rem;
    font-style: italic;
    color: #b0c0d0;
    line-height: 1.5;
    background-color: #1a2536;
    border-left: 3px solid #3a5a7a;
    border-radius: 0 4px 4px 0;
  }

  /* ======== Inline edit form ======== */

  .inline-edit-form {
    padding: 12px 0;
    border-bottom: 1px solid #2a3a4a;
    margin-bottom: 8px;
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
