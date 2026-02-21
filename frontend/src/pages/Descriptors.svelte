<script lang="ts">
  import { onMount } from "svelte";
  import {
    listDescriptors,
    createDescriptor,
    updateDescriptor,
    deleteDescriptor,
    reorderDescriptors,
    checkDescriptorLensReferences,
    addToast,
    type RoleDescriptor,
  } from "../services/api";
  import DragHandle from "../components/DragHandle.svelte";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";

  let descriptors: RoleDescriptor[] = [];
  let loading = true;

  // Form state
  let showForm = false;
  let formTitle = "";

  // Inline edit state
  let inlineEditId: number | null = null;
  let inlineEditTitle = "";

  // Delete confirmation state
  let deleteConfirmDesc: RoleDescriptor | null = null;
  let lensReferences: string[] = [];

  onMount(async () => {
    await loadDescriptors();
  });

  async function loadDescriptors(): Promise<void> {
    loading = true;
    try {
      descriptors = (await listDescriptors()) || [];
    } finally {
      loading = false;
    }
  }

  function openAddForm(): void {
    formTitle = "";
    showForm = true;
  }

  function cancelForm(): void {
    showForm = false;
  }

  async function handleSubmit(): Promise<void> {
    const title = formTitle.trim();
    if (!title) {
      addToast("error", "Title is required");
      return;
    }

    try {
      await createDescriptor(title);
      showForm = false;
      await loadDescriptors();
    } catch {
      // Toast already shown
    }
  }

  function startInlineEdit(desc: RoleDescriptor): void {
    inlineEditId = desc.id;
    inlineEditTitle = desc.title;
  }

  function cancelInlineEdit(): void {
    inlineEditId = null;
    inlineEditTitle = "";
  }

  async function handleInlineEditSave(): Promise<void> {
    if (inlineEditId === null) return;
    const title = inlineEditTitle.trim();
    if (!title) {
      addToast("error", "Title is required");
      return;
    }

    try {
      await updateDescriptor(inlineEditId, title);
      inlineEditId = null;
      inlineEditTitle = "";
      await loadDescriptors();
    } catch {
      // Toast already shown
    }
  }

  async function confirmDelete(desc: RoleDescriptor): Promise<void> {
    try {
      lensReferences = (await checkDescriptorLensReferences(desc.id)) || [];
      deleteConfirmDesc = desc;
    } catch {
      // Toast already shown
    }
  }

  async function handleDelete(): Promise<void> {
    if (!deleteConfirmDesc) return;
    try {
      await deleteDescriptor(deleteConfirmDesc.id);
      deleteConfirmDesc = null;
      lensReferences = [];
      await loadDescriptors();
    } catch {
      // Toast already shown
    }
  }

  function cancelDelete(): void {
    deleteConfirmDesc = null;
    lensReferences = [];
  }

  async function handleReorder(orderedIDs: number[]): Promise<void> {
    try {
      await reorderDescriptors(orderedIDs);
      await loadDescriptors();
    } catch {
      // Toast already shown
    }
  }

  function getDescriptor(id: number): RoleDescriptor {
    return descriptors.find((d) => d.id === id) as RoleDescriptor;
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key === "Enter") {
      handleInlineEditSave();
    } else if (event.key === "Escape") {
      cancelInlineEdit();
    }
  }

  $: sortedDescriptors = [...descriptors].sort(
    (a, b) => a.sort_order - b.sort_order
  );
</script>

<div class="descriptors-page">
  <div class="page-header">
    <h2>Role Descriptors</h2>
    <button class="btn btn-primary" on:click={openAddForm}>
      + Add Descriptor
    </button>
  </div>
  <p class="page-description">
    Define role descriptor titles for your resume. These are short labels like
    "Full-Stack Developer" or "Technical Lead" used to describe your role in a
    lens.
  </p>

  {#if showForm}
    <div class="entry-form">
      <h3 class="form-title">New Role Descriptor</h3>
      <div class="form-row">
        <div class="form-field">
          <label class="form-label" for="desc-title">Title</label>
          <input
            id="desc-title"
            type="text"
            class="form-input"
            bind:value={formTitle}
            placeholder="e.g. Senior Software Engineer"
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
  {#if deleteConfirmDesc}
    <div class="confirm-dialog">
      <p>
        Delete descriptor <strong>{deleteConfirmDesc.title}</strong>?
      </p>
      {#if lensReferences.length > 0}
        <p class="warn-text">
          This descriptor is referenced by {lensReferences.length} lens{lensReferences.length
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
  {:else if sortedDescriptors.length === 0}
    <div class="empty-state">
      <p>No role descriptors yet.</p>
      <p class="empty-hint">
        Add titles that describe your professional roles.
      </p>
    </div>
  {:else}
    <div class="descriptors-list">
      <DragHandle
        items={sortedDescriptors}
        on:reorder={(e) => handleReorder(e.detail.orderedIDs)}
        let:item
      >
        {@const desc = getDescriptor(item.id)}
        <div class="descriptor-card">
          {#if inlineEditId === desc.id}
            <input
              type="text"
              class="form-input inline-edit"
              bind:value={inlineEditTitle}
              on:keydown={handleKeydown}
            />
            <div class="descriptor-actions">
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
            <span class="descriptor-title">{desc.title}</span>
            <div class="descriptor-actions">
              <button
                class="btn btn-small btn-ghost"
                on:click={() => startInlineEdit(desc)}
              >
                Edit
              </button>
              <button
                class="btn btn-small btn-danger"
                on:click={() => confirmDelete(desc)}
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
  .descriptors-page {
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

  /* --- Descriptor Cards --- */
  .descriptors-list {
    display: flex;
    flex-direction: column;
  }

  .descriptor-card {
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

  .descriptor-title {
    font-size: 0.9rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .descriptor-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }
</style>
