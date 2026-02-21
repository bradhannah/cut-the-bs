<script lang="ts">
  import { onMount } from "svelte";
  import {
    listSummaries,
    createSummary,
    updateSummary,
    deleteSummary,
    addToast,
    type ProfessionalSummary,
    type SummaryInput,
  } from "../services/api";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";

  let summaries: ProfessionalSummary[] = [];
  let loading = true;

  // Form state
  let showForm = false;
  let editingSummary: ProfessionalSummary | null = null;
  let formLabel = "";
  let formBodyText = "";

  onMount(async () => {
    await loadSummaries();
  });

  async function loadSummaries(): Promise<void> {
    loading = true;
    try {
      summaries = (await listSummaries()) || [];
    } finally {
      loading = false;
    }
  }

  function openAddForm(): void {
    editingSummary = null;
    formLabel = "";
    formBodyText = "";
    showForm = true;
  }

  function openEditForm(summary: ProfessionalSummary): void {
    editingSummary = summary;
    formLabel = summary.label;
    formBodyText = summary.body_text;
    showForm = true;
  }

  function cancelForm(): void {
    showForm = false;
    editingSummary = null;
  }

  async function handleSubmit(): Promise<void> {
    const input: SummaryInput = {
      label: formLabel.trim(),
      body_text: formBodyText.trim(),
    };

    if (!input.label) {
      addToast("error", "Label is required");
      return;
    }
    if (!input.body_text) {
      addToast("error", "Summary text is required");
      return;
    }

    try {
      if (editingSummary) {
        await updateSummary(editingSummary.id, input);
      } else {
        await createSummary(input);
      }
      showForm = false;
      editingSummary = null;
      await loadSummaries();
    } catch {
      // Toast already shown
    }
  }

  async function handleDelete(id: number): Promise<void> {
    try {
      await deleteSummary(id);
      await loadSummaries();
    } catch {
      // Toast already shown
    }
  }
</script>

<div class="summaries-page">
  <div class="page-header">
    <h2>Professional Summaries</h2>
    <button class="btn btn-primary" on:click={openAddForm}>
      + Add Summary
    </button>
  </div>
  <p class="page-description">
    Create multiple summary variants for different job targets. Select which one
    to include when building a resume.
  </p>

  {#if showForm}
    <div class="entry-form">
      <h3 class="form-title">
        {editingSummary ? "Edit Summary" : "New Summary Variant"}
      </h3>
      <div class="form-field">
        <label class="form-label" for="summary-label">Label</label>
        <input
          id="summary-label"
          type="text"
          class="form-input"
          bind:value={formLabel}
          placeholder="e.g. Full-Stack Developer, Backend Focus"
        />
      </div>
      <div class="form-field" style="margin-top: 12px;">
        <label class="form-label" for="summary-body">Summary Text</label>
        <textarea
          id="summary-body"
          class="form-input form-textarea"
          bind:value={formBodyText}
          placeholder="A results-driven software engineer with 8+ years of experience..."
          rows="6"
        />
      </div>
      <div class="form-actions">
        <button class="btn btn-primary" on:click={handleSubmit}>
          {editingSummary ? "Update" : "Create"}
        </button>
        <button class="btn btn-cancel" on:click={cancelForm}>Cancel</button>
      </div>
    </div>
  {/if}

  {#if loading}
    <LoadingSpinner />
  {:else if summaries.length === 0}
    <div class="empty-state">
      <p>No summary variants yet.</p>
      <p class="empty-hint">
        Create different versions of your professional summary for different
        types of positions.
      </p>
    </div>
  {:else}
    <div class="summary-list">
      {#each summaries as summary (summary.id)}
        <div class="summary-card">
          <div class="summary-header">
            <span class="summary-label">{summary.label}</span>
            <div class="summary-actions">
              <button
                class="btn btn-small btn-ghost"
                on:click={() => openEditForm(summary)}
              >
                Edit
              </button>
              <button
                class="btn btn-small btn-danger"
                on:click={() => handleDelete(summary.id)}
              >
                Delete
              </button>
            </div>
          </div>
          <p class="summary-body">{summary.body_text}</p>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .summaries-page {
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

  .empty-state {
    text-align: center;
    padding: 48px 0;
    color: #5a6a7a;
  }

  .empty-hint {
    font-size: 0.85rem;
    margin-top: 8px;
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

  .form-field {
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
    font-family: inherit;
    resize: vertical;
    line-height: 1.5;
  }

  .form-actions {
    display: flex;
    gap: 8px;
    margin-top: 16px;
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

  .btn-small {
    padding: 4px 10px;
    font-size: 0.8rem;
  }

  .btn-ghost {
    background-color: transparent;
    border-color: transparent;
    color: #7a8a9a;
  }

  .btn-ghost:hover {
    background-color: #2a3a4a;
    color: #c0d0e0;
  }

  .btn-danger {
    background-color: transparent;
    border-color: transparent;
    color: #c05050;
  }

  .btn-danger:hover {
    background-color: #3a2020;
    color: #e06060;
  }

  /* --- Summary Cards --- */
  .summary-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .summary-card {
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    padding: 16px;
  }

  .summary-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 10px;
  }

  .summary-label {
    font-size: 0.95rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .summary-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  .summary-body {
    margin: 0;
    font-size: 0.85rem;
    color: #a0b0c0;
    line-height: 1.6;
    white-space: pre-wrap;
  }
</style>
