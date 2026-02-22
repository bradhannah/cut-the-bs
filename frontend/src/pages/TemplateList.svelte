<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import {
    listDocumentTemplates,
    createDocumentTemplate,
    deleteDocumentTemplate,
    duplicateDocumentTemplate,
    addToast,
    type DocumentTemplate,
    type DocumentTemplateInput,
  } from "../services/api";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";
  import { formatTimestamp } from "../services/dateFormat";

  let templates: DocumentTemplate[] = [];
  let loading = true;

  // --- Create dialog ---
  let showCreateDialog = false;
  let newName = "";
  let newType = "resume";
  let creating = false;

  onMount(async () => {
    await loadTemplates();
  });

  async function loadTemplates(): Promise<void> {
    loading = true;
    try {
      templates = await listDocumentTemplates();
    } catch (e: any) {
      addToast("error", e?.message || "Failed to load templates");
    } finally {
      loading = false;
    }
  }

  function openCreateDialog(): void {
    newName = "";
    newType = "resume";
    showCreateDialog = true;
  }

  function closeCreateDialog(): void {
    showCreateDialog = false;
  }

  async function handleCreate(): Promise<void> {
    if (!newName.trim()) {
      addToast("error", "Template name is required");
      return;
    }
    creating = true;
    try {
      const input: DocumentTemplateInput = {
        name: newName.trim(),
        description: "",
        template_type: newType,
        margin_top: 54,
        margin_bottom: 54,
        margin_left: 54,
        margin_right: 54,
      };
      const created = await createDocumentTemplate(input);
      showCreateDialog = false;
      push(`/templates/${created.id}/builder`);
    } catch (e: any) {
      addToast("error", e?.message || "Failed to create template");
    } finally {
      creating = false;
    }
  }

  async function handleDuplicate(tmpl: DocumentTemplate): Promise<void> {
    const name = `${tmpl.name} (Copy)`;
    try {
      const dup = await duplicateDocumentTemplate(tmpl.id, name);
      templates = [...templates, dup];
    } catch (e: any) {
      addToast("error", e?.message || "Failed to duplicate template");
    }
  }

  async function handleDelete(tmpl: DocumentTemplate): Promise<void> {
    if (tmpl.is_builtin) {
      addToast("error", "Built-in templates cannot be deleted");
      return;
    }
    try {
      await deleteDocumentTemplate(tmpl.id);
      templates = templates.filter((t) => t.id !== tmpl.id);
    } catch (e: any) {
      addToast("error", e?.message || "Failed to delete template");
    }
  }

  function handleKeydown(e: KeyboardEvent): void {
    if (showCreateDialog && e.key === "Escape") {
      closeCreateDialog();
    }
    if (showCreateDialog && e.key === "Enter") {
      handleCreate();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="template-list-page">
  <div class="page-header">
    <h2>Templates</h2>
    <button class="btn btn-primary" on:click={openCreateDialog}>
      + New Template
    </button>
  </div>

  {#if loading}
    <div class="loading-container">
      <LoadingSpinner />
    </div>
  {:else if templates.length === 0}
    <div class="empty-state">
      <p>No templates yet. Create one to get started.</p>
    </div>
  {:else}
    <div class="template-grid">
      {#each templates as tmpl (tmpl.id)}
        <div class="template-card" class:builtin={tmpl.is_builtin}>
          <div class="card-header">
            <h3 class="card-title">
              <a href={"#/templates/" + tmpl.id + "/builder"}>{tmpl.name}</a>
            </h3>
            <div class="card-badges">
              <span class="type-badge">{tmpl.template_type}</span>
              {#if tmpl.is_builtin}
                <span class="builtin-badge">Built-in</span>
              {/if}
            </div>
          </div>
          {#if tmpl.description}
            <p class="card-description">{tmpl.description}</p>
          {/if}
          <div class="card-meta">
            <span class="meta-date">Updated {formatTimestamp(tmpl.updated_at)}</span>
          </div>
          <div class="card-actions">
            <a
              href={"#/templates/" + tmpl.id + "/builder"}
              class="btn btn-small btn-primary"
            >
              Edit
            </a>
            <button
              class="btn btn-small btn-ghost"
              on:click={() => handleDuplicate(tmpl)}
            >
              Duplicate
            </button>
            {#if !tmpl.is_builtin}
              <button
                class="btn btn-small btn-danger"
                on:click={() => handleDelete(tmpl)}
              >
                Delete
              </button>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Create template dialog -->
{#if showCreateDialog}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="dialog-overlay" on:click|self={closeCreateDialog}>
    <div class="dialog">
      <h3 class="dialog-title">New Template</h3>
      <div class="dialog-body">
        <div class="form-field">
          <label class="form-label" for="tmpl-name">Name</label>
          <input
            id="tmpl-name"
            class="form-input"
            type="text"
            bind:value={newName}
            placeholder="My Resume Template"
            maxlength="100"
          />
        </div>
        <div class="form-field">
          <label class="form-label" for="tmpl-type">Type</label>
          <select id="tmpl-type" class="form-input" bind:value={newType}>
            <option value="resume">Resume</option>
            <option value="cover_letter">Cover Letter</option>
          </select>
        </div>
      </div>
      <div class="dialog-actions">
        <button class="btn btn-cancel" on:click={closeCreateDialog}>
          Cancel
        </button>
        <button
          class="btn btn-primary"
          on:click={handleCreate}
          disabled={creating || !newName.trim()}
        >
          {creating ? "Creating..." : "Create"}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .template-list-page {
    max-width: 900px;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
  }

  .page-header h2 {
    margin: 0;
    font-size: 1.3rem;
    color: #e0e0e0;
  }

  .loading-container {
    display: flex;
    justify-content: center;
    padding: 60px 0;
  }

  .empty-state {
    text-align: center;
    padding: 60px 20px;
    color: #7a8a9a;
  }

  /* --- Template grid --- */
  .template-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }

  .template-card {
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    transition: border-color 0.15s;
  }

  .template-card:hover {
    border-color: #3a4a5a;
  }

  .template-card.builtin {
    border-left: 3px solid #3a4a2a;
  }

  .card-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px;
  }

  .card-title {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
  }

  .card-title a {
    color: #e0e0e0;
    text-decoration: none;
  }

  .card-title a:hover {
    color: #4a8af4;
  }

  .card-badges {
    display: flex;
    gap: 6px;
    flex-shrink: 0;
  }

  .type-badge {
    font-size: 0.7rem;
    padding: 2px 6px;
    border-radius: 3px;
    background-color: #2a3a4a;
    color: #c0d0e0;
    text-transform: capitalize;
  }

  .builtin-badge {
    font-size: 0.7rem;
    padding: 2px 6px;
    border-radius: 3px;
    background-color: #3a4a2a;
    color: #b0d080;
  }

  .card-description {
    font-size: 0.85rem;
    color: #7a8a9a;
    margin: 0;
    line-height: 1.4;
  }

  .card-meta {
    font-size: 0.75rem;
    color: #5a6a7a;
  }

  .card-actions {
    display: flex;
    gap: 8px;
    margin-top: 4px;
  }

  /* --- Button styles --- */
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    font-size: 0.85rem;
    cursor: pointer;
    text-decoration: none;
    transition: background-color 0.15s, color 0.15s;
  }

  .btn-primary {
    background-color: #4a8af4;
    color: #ffffff;
  }

  .btn-primary:hover {
    background-color: #3a7ae4;
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-small {
    padding: 4px 10px;
    font-size: 0.78rem;
  }

  .btn-ghost {
    background: none;
    color: #a0b0c0;
    border: 1px solid #2a3a4a;
  }

  .btn-ghost:hover {
    background-color: #223344;
    color: #e0e0e0;
  }

  .btn-danger {
    background: none;
    color: #ff6b6b;
    border: 1px solid #4a2a2a;
  }

  .btn-danger:hover {
    background-color: #3a2020;
  }

  .btn-cancel {
    background: none;
    color: #a0b0c0;
  }

  .btn-cancel:hover {
    color: #e0e0e0;
  }

  /* --- Dialog --- */
  .dialog-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .dialog {
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 8px;
    padding: 24px;
    width: 400px;
    max-width: 90vw;
  }

  .dialog-title {
    margin: 0 0 16px;
    font-size: 1.1rem;
    color: #e0e0e0;
  }

  .dialog-body {
    display: flex;
    flex-direction: column;
    gap: 14px;
    margin-bottom: 20px;
  }

  .form-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .form-label {
    font-size: 0.8rem;
    color: #a0b0c0;
  }

  .form-input {
    padding: 8px 10px;
    background-color: #1b2636;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    color: #e0e0e0;
    font-size: 0.9rem;
    outline: none;
    transition: border-color 0.15s;
  }

  .form-input:focus {
    border-color: #4a8af4;
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }
</style>
