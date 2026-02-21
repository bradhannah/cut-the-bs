<script lang="ts">
  import { onMount } from "svelte";
  import {
    listCoreExpertise,
    createCoreExpertise,
    updateCoreExpertise,
    deleteCoreExpertise,
    reorderCoreExpertise,
    checkCoreExpertiseLensReferences,
    splitCoreExpertiseText,
    addToast,
    type CoreExpertise,
  } from "../services/api";
  import DragHandle from "../components/DragHandle.svelte";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";

  let items: CoreExpertise[] = [];
  let loading = true;

  // Form state
  let showForm = false;
  let formLabel = "";

  // Paste import state
  let showPasteImport = false;
  let pasteText = "";
  let importing = false;

  // Inline edit state
  let inlineEditId: number | null = null;
  let inlineEditLabel = "";

  // Delete confirmation state
  let deleteConfirmItem: CoreExpertise | null = null;
  let lensReferences: string[] = [];

  onMount(async () => {
    await loadItems();
  });

  async function loadItems(): Promise<void> {
    loading = true;
    try {
      items = (await listCoreExpertise()) || [];
    } finally {
      loading = false;
    }
  }

  function openAddForm(): void {
    formLabel = "";
    showForm = true;
  }

  function cancelForm(): void {
    showForm = false;
  }

  async function handleSubmit(): Promise<void> {
    const label = formLabel.trim();
    if (!label) {
      addToast("error", "Label is required");
      return;
    }

    try {
      await createCoreExpertise(label);
      showForm = false;
      await loadItems();
    } catch {
      // Toast already shown
    }
  }

  // --- Paste Import ---

  function openPasteImport(): void {
    pasteText = "";
    showPasteImport = true;
  }

  function cancelPasteImport(): void {
    showPasteImport = false;
    pasteText = "";
  }

  async function handlePasteImport(): Promise<void> {
    const text = pasteText.trim();
    if (!text) {
      addToast("error", "Paste some text first");
      return;
    }

    importing = true;
    try {
      const labels = await splitCoreExpertiseText(text);
      if (labels.length === 0) {
        addToast("error", "No items found in pasted text");
        return;
      }

      for (const label of labels) {
        await createCoreExpertise(label);
      }

      addToast("success", `Imported ${labels.length} item${labels.length !== 1 ? "s" : ""}`);
      showPasteImport = false;
      pasteText = "";
      await loadItems();
    } catch {
      // Toast already shown
    } finally {
      importing = false;
    }
  }

  // --- Inline Edit ---

  function startInlineEdit(item: CoreExpertise): void {
    inlineEditId = item.id;
    inlineEditLabel = item.label;
  }

  function cancelInlineEdit(): void {
    inlineEditId = null;
    inlineEditLabel = "";
  }

  async function handleInlineEditSave(): Promise<void> {
    if (inlineEditId === null) return;
    const label = inlineEditLabel.trim();
    if (!label) {
      addToast("error", "Label is required");
      return;
    }

    try {
      await updateCoreExpertise(inlineEditId, label);
      inlineEditId = null;
      inlineEditLabel = "";
      await loadItems();
    } catch {
      // Toast already shown
    }
  }

  async function confirmDelete(item: CoreExpertise): Promise<void> {
    try {
      lensReferences = (await checkCoreExpertiseLensReferences(item.id)) || [];
      deleteConfirmItem = item;
    } catch {
      // Toast already shown
    }
  }

  async function handleDelete(): Promise<void> {
    if (!deleteConfirmItem) return;
    try {
      await deleteCoreExpertise(deleteConfirmItem.id);
      deleteConfirmItem = null;
      lensReferences = [];
      await loadItems();
    } catch {
      // Toast already shown
    }
  }

  function cancelDelete(): void {
    deleteConfirmItem = null;
    lensReferences = [];
  }

  async function handleReorder(orderedIDs: number[]): Promise<void> {
    try {
      await reorderCoreExpertise(orderedIDs);
      await loadItems();
    } catch {
      // Toast already shown
    }
  }

  function getItem(id: number): CoreExpertise {
    return items.find((d) => d.id === id) as CoreExpertise;
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key === "Enter") {
      handleInlineEditSave();
    } else if (event.key === "Escape") {
      cancelInlineEdit();
    }
  }

  $: sortedItems = [...items].sort((a, b) => a.sort_order - b.sort_order);
</script>

<div class="ce-page">
  <div class="page-header">
    <h2>Core Expertise</h2>
    <div class="page-header-actions">
      <button class="btn btn-secondary" on:click={openPasteImport}>
        Paste Import
      </button>
      <button class="btn btn-primary" on:click={openAddForm}>
        + Add Item
      </button>
    </div>
  </div>
  <p class="page-description">
    Define core expertise keywords for your resume. These are short phrases like
    "Cloud-Native Architecture" or "DevOps" rendered as pipe-separated text in
    the PDF export.
  </p>

  {#if showPasteImport}
    <div class="entry-form">
      <h3 class="form-title">Paste Import</h3>
      <p class="paste-hint">
        Paste keywords separated by pipes (|), commas, or newlines.
      </p>
      <div class="form-field">
        <textarea
          class="form-input paste-textarea"
          bind:value={pasteText}
          placeholder="Cloud-Native Architecture | DevOps | Microservices"
          rows="4"
        ></textarea>
      </div>
      <div class="form-actions">
        <button
          class="btn btn-primary"
          on:click={handlePasteImport}
          disabled={importing}
        >
          {importing ? "Importing..." : "Import"}
        </button>
        <button class="btn btn-cancel" on:click={cancelPasteImport}>
          Cancel
        </button>
      </div>
    </div>
  {/if}

  {#if showForm}
    <div class="entry-form">
      <h3 class="form-title">New Core Expertise</h3>
      <div class="form-row">
        <div class="form-field">
          <label class="form-label" for="ce-label">Label</label>
          <input
            id="ce-label"
            type="text"
            class="form-input"
            bind:value={formLabel}
            placeholder="e.g. Cloud-Native Architecture"
          />
        </div>
      </div>
      <div class="form-actions">
        <button class="btn btn-primary" on:click={handleSubmit}>Create</button>
        <button class="btn btn-cancel" on:click={cancelForm}>Cancel</button>
      </div>
    </div>
  {/if}

  <!-- Delete Confirmation -->
  {#if deleteConfirmItem}
    <div class="confirm-dialog">
      <p>
        Delete <strong>{deleteConfirmItem.label}</strong>?
      </p>
      {#if lensReferences.length > 0}
        <p class="warn-text">
          This item is referenced by {lensReferences.length} lens{lensReferences.length
            !== 1
            ? "es"
            : ""}:
          {lensReferences.join(", ")}
        </p>
      {/if}
      <div class="form-actions">
        <button class="btn btn-danger-solid" on:click={handleDelete}>
          Delete
        </button>
        <button class="btn btn-cancel" on:click={cancelDelete}>Cancel</button>
      </div>
    </div>
  {/if}

  {#if loading}
    <LoadingSpinner />
  {:else if sortedItems.length === 0}
    <div class="empty-state">
      <p>No core expertise items yet.</p>
      <p class="empty-hint">
        Add keywords that highlight your core areas of expertise.
      </p>
    </div>
  {:else}
    <div class="items-list">
      <DragHandle
        items={sortedItems}
        on:reorder={(e) => handleReorder(e.detail.orderedIDs)}
        let:item
      >
        {@const ce = getItem(item.id)}
        <div class="item-card">
          {#if inlineEditId === ce.id}
            <input
              type="text"
              class="form-input inline-edit"
              bind:value={inlineEditLabel}
              on:keydown={handleKeydown}
            />
            <div class="item-actions">
              <button
                class="btn btn-small btn-primary"
                on:click={handleInlineEditSave}
              >
                Save
              </button>
              <button
                class="btn btn-small btn-cancel"
                on:click={cancelInlineEdit}
              >
                Cancel
              </button>
            </div>
          {:else}
            <span class="item-label">{ce.label}</span>
            <div class="item-actions">
              <button
                class="btn btn-small btn-ghost"
                on:click={() => startInlineEdit(ce)}
              >
                Edit
              </button>
              <button
                class="btn btn-small btn-danger"
                on:click={() => confirmDelete(ce)}
              >
                Delete
              </button>
            </div>
          {/if}
        </div>
      </DragHandle>
    </div>
  {/if}
</div>

<style>
  .ce-page {
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

  .inline-edit {
    flex: 1;
  }

  .paste-hint {
    color: #7a8a9a;
    font-size: 0.85rem;
    margin: 0 0 12px;
  }

  .paste-textarea {
    width: 100%;
    resize: vertical;
    font-family: inherit;
  }

  .form-actions {
    display: flex;
    gap: 8px;
    margin-top: 12px;
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

  .btn-secondary {
    background-color: #2a3a4a;
    border-color: #3a4a5a;
    color: #c0d0e0;
  }

  .btn-secondary:hover {
    background-color: #3a4a5a;
    color: #e0e0e0;
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

  .btn-danger-solid {
    background-color: #802020;
    border-color: #a03030;
    color: #e0e0e0;
  }

  .btn-danger-solid:hover {
    background-color: #a03030;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

  /* --- Item Cards --- */
  .items-list {
    display: flex;
    flex-direction: column;
  }

  .item-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px;
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    margin-bottom: 6px;
    gap: 12px;
  }

  .item-label {
    font-size: 0.9rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .item-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }
</style>
