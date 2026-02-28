<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import {
    listDocumentTemplates,
    createDocumentTemplate,
    deleteDocumentTemplate,
    duplicateDocumentTemplate,
    previewTemplate,
    openFile,
    addToast,
    type DocumentTemplate,
    type DocumentTemplateInput,
  } from "../services/api";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";
  import { formatTimestamp } from "../services/dateFormat";

  let templates: DocumentTemplate[] = [];
  let loading = true;

  let showCreateDialog = false;
  let newName = "";
  let newType = "resume";
  let newDescription = "";
  let creating = false;

  let showDuplicateDialog = false;
  let duplicateName = "";
  let duplicatingTemplate: DocumentTemplate | null = null;
  let duplicating = false;

  let showDeleteDialog = false;
  let deletingTemplate: DocumentTemplate | null = null;
  let previewingTemplateID: number | null = null;

  $: resumeTemplates = templates.filter((t) => t.template_type === "resume");
  $: coverLetterTemplates = templates.filter(
    (t) => t.template_type === "cover_letter"
  );
  $: resumeCount = resumeTemplates.length;
  $: coverLetterCount = coverLetterTemplates.length;

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

  function openTemplateBuilder(tmpl: DocumentTemplate): void {
    if (tmpl.is_builtin) {
      push(`/templates/${tmpl.id}/builder?mode=view`);
      return;
    }
    push(`/templates/${tmpl.id}/builder`);
  }

  async function handlePreview(tmpl: DocumentTemplate): Promise<void> {
    if (previewingTemplateID !== null) {
      return;
    }

    previewingTemplateID = tmpl.id;
    try {
      const pdfPath = await previewTemplate(tmpl.id);
      await openFile(pdfPath);
    } catch (e: any) {
      addToast("error", e?.message || "Preview failed");
    } finally {
      previewingTemplateID = null;
    }
  }

  function openCreateDialog(): void {
    newName = "";
    newType = "resume";
    newDescription = "";
    showCreateDialog = true;
  }

  function closeCreateDialog(): void {
    showCreateDialog = false;
  }

  async function handleCreate(): Promise<void> {
    const trimmedName = newName.trim();
    if (!trimmedName) {
      addToast("error", "Template name is required");
      return;
    }

    creating = true;
    try {
      const input: DocumentTemplateInput = {
        name: trimmedName,
        description: newDescription.trim(),
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

  function openDuplicateDialog(tmpl: DocumentTemplate): void {
    duplicatingTemplate = tmpl;
    duplicateName = `${tmpl.name} (Copy)`;
    showDuplicateDialog = true;
  }

  function closeDuplicateDialog(): void {
    showDuplicateDialog = false;
    duplicatingTemplate = null;
    duplicateName = "";
  }

  async function handleDuplicate(): Promise<void> {
    if (!duplicatingTemplate || !duplicateName.trim()) {
      addToast("error", "Template name is required");
      return;
    }
    duplicating = true;
    try {
      const dup = await duplicateDocumentTemplate(
        duplicatingTemplate.id,
        duplicateName.trim()
      );
      templates = [...templates, dup];
      closeDuplicateDialog();
    } catch (e: any) {
      addToast("error", e?.message || "Failed to duplicate template");
    } finally {
      duplicating = false;
    }
  }

  function confirmDelete(tmpl: DocumentTemplate): void {
    if (tmpl.is_builtin) {
      addToast("error", "Built-in templates cannot be deleted");
      return;
    }
    deletingTemplate = tmpl;
    showDeleteDialog = true;
  }

  function closeDeleteDialog(): void {
    showDeleteDialog = false;
    deletingTemplate = null;
  }

  async function handleDelete(): Promise<void> {
    if (!deletingTemplate) return;
    try {
      await deleteDocumentTemplate(deletingTemplate.id);
      templates = templates.filter((t) => t.id !== deletingTemplate!.id);
      closeDeleteDialog();
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
    if (showDuplicateDialog && e.key === "Escape") {
      closeDuplicateDialog();
    }
    if (showDuplicateDialog && e.key === "Enter") {
      handleDuplicate();
    }
    if (showDeleteDialog && e.key === "Escape") {
      closeDeleteDialog();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="template-list-page">
  <div class="page-header">
    <div>
      <h2>Templates</h2>
      <p class="page-subtitle">
        Built-ins are view-only. Duplicate built-ins to create editable copies.
        Rename and edit descriptions from the template editor header.
      </p>
    </div>
    <button class="btn btn-primary" on:click={openCreateDialog}>+ New Template</button>
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
    <div class="counts-row">
      <span class="count-pill">Resume: {resumeCount}</span>
      <span class="count-pill">Cover Letter: {coverLetterCount}</span>
      <span class="count-pill">Total: {templates.length}</span>
    </div>

    <section class="template-section">
      <div class="section-header">
        <h3>Resume Templates</h3>
        <span class="section-count">{resumeCount}</span>
      </div>
      {#if resumeTemplates.length === 0}
        <p class="section-empty">No resume templates yet.</p>
      {:else}
        <div class="template-table-wrap">
          <table class="template-table">
            <colgroup>
              <col class="col-name" />
              <col class="col-description" />
              <col class="col-updated" />
              <col class="col-access" />
              <col class="col-actions" />
            </colgroup>
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Updated</th>
                <th>Access</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {#each resumeTemplates as tmpl (tmpl.id)}
                <tr class:builtin-row={tmpl.is_builtin}>
                  <td><div class="tmpl-name">{tmpl.name}</div></td>
                  <td><span class="tmpl-description">{tmpl.description || "-"}</span></td>
                  <td><span class="tmpl-date">{formatTimestamp(tmpl.updated_at)}</span></td>
                  <td>
                    {#if tmpl.is_builtin}
                      <span class="access-badge builtin">Built-in (Read-only)</span>
                    {:else}
                      <span class="access-badge custom">Editable</span>
                    {/if}
                  </td>
                  <td>
                    <div class="row-actions">
                      <button
                        class="btn btn-small btn-ghost"
                        on:click={() => handlePreview(tmpl)}
                        disabled={previewingTemplateID !== null}
                      >
                        {previewingTemplateID === tmpl.id ? "Previewing..." : "Preview"}
                      </button>
                      <button class="btn btn-small btn-primary" on:click={() => openTemplateBuilder(tmpl)}>
                        {tmpl.is_builtin ? "View" : "Edit"}
                      </button>
                      <button class="btn btn-small btn-ghost" on:click={() => openDuplicateDialog(tmpl)}>
                        Duplicate
                      </button>
                      {#if !tmpl.is_builtin}
                        <button class="btn btn-small btn-danger" on:click={() => confirmDelete(tmpl)}>
                          Delete
                        </button>
                      {/if}
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </section>

    <section class="template-section">
      <div class="section-header">
        <h3>Cover Letter Templates</h3>
        <span class="section-count">{coverLetterCount}</span>
      </div>
      {#if coverLetterTemplates.length === 0}
        <p class="section-empty">No cover letter templates yet.</p>
      {:else}
        <div class="template-table-wrap">
          <table class="template-table">
            <colgroup>
              <col class="col-name" />
              <col class="col-description" />
              <col class="col-updated" />
              <col class="col-access" />
              <col class="col-actions" />
            </colgroup>
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Updated</th>
                <th>Access</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {#each coverLetterTemplates as tmpl (tmpl.id)}
                <tr class:builtin-row={tmpl.is_builtin}>
                  <td><div class="tmpl-name">{tmpl.name}</div></td>
                  <td><span class="tmpl-description">{tmpl.description || "-"}</span></td>
                  <td><span class="tmpl-date">{formatTimestamp(tmpl.updated_at)}</span></td>
                  <td>
                    {#if tmpl.is_builtin}
                      <span class="access-badge builtin">Built-in (Read-only)</span>
                    {:else}
                      <span class="access-badge custom">Editable</span>
                    {/if}
                  </td>
                  <td>
                    <div class="row-actions">
                      <button
                        class="btn btn-small btn-ghost"
                        on:click={() => handlePreview(tmpl)}
                        disabled={previewingTemplateID !== null}
                      >
                        {previewingTemplateID === tmpl.id ? "Previewing..." : "Preview"}
                      </button>
                      <button class="btn btn-small btn-primary" on:click={() => openTemplateBuilder(tmpl)}>
                        {tmpl.is_builtin ? "View" : "Edit"}
                      </button>
                      <button class="btn btn-small btn-ghost" on:click={() => openDuplicateDialog(tmpl)}>
                        Duplicate
                      </button>
                      {#if !tmpl.is_builtin}
                        <button class="btn btn-small btn-danger" on:click={() => confirmDelete(tmpl)}>
                          Delete
                        </button>
                      {/if}
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </section>
  {/if}
</div>

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
          <label class="form-label" for="tmpl-desc">Description</label>
          <textarea
            id="tmpl-desc"
            class="form-input"
            rows="3"
            bind:value={newDescription}
            placeholder="Optional template description"
            maxlength="300"
          ></textarea>
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
        <button class="btn btn-cancel" on:click={closeCreateDialog}>Cancel</button>
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

{#if showDuplicateDialog}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="dialog-overlay" on:click|self={closeDuplicateDialog}>
    <div class="dialog">
      <h3 class="dialog-title">Duplicate Template</h3>
      <div class="dialog-body">
        <div class="form-field">
          <label class="form-label" for="dup-name">Name</label>
          <input
            id="dup-name"
            class="form-input"
            type="text"
            bind:value={duplicateName}
            placeholder="Template name"
            maxlength="100"
          />
        </div>
      </div>
      <div class="dialog-actions">
        <button class="btn btn-cancel" on:click={closeDuplicateDialog}>Cancel</button>
        <button
          class="btn btn-primary"
          on:click={handleDuplicate}
          disabled={duplicating || !duplicateName.trim()}
        >
          {duplicating ? "Duplicating..." : "Duplicate"}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showDeleteDialog && deletingTemplate}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="dialog-overlay" on:click|self={closeDeleteDialog}>
    <div class="dialog">
      <h3 class="dialog-title">Delete Template</h3>
      <div class="dialog-body">
        <p class="confirm-text">
          Delete <strong>{deletingTemplate.name}</strong>? This also deletes all
          template elements and cannot be undone.
        </p>
      </div>
      <div class="dialog-actions">
        <button class="btn btn-cancel" on:click={closeDeleteDialog}>Cancel</button>
        <button class="btn btn-danger" on:click={handleDelete}>Delete</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .template-list-page {
    max-width: 1100px;
  }

  .page-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 16px;
    gap: 14px;
  }

  .page-header h2 {
    margin: 0;
    font-size: 1.32rem;
    color: #e0e0e0;
  }

  .page-subtitle {
    margin: 6px 0 0;
    color: #7a8a9a;
    font-size: 0.85rem;
    line-height: 1.45;
    max-width: 700px;
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

  .counts-row {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    margin-bottom: 14px;
  }

  .template-section {
    margin-bottom: 18px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .section-header h3 {
    margin: 0;
    color: #c9d9ec;
    font-size: 0.95rem;
    letter-spacing: 0.02em;
  }

  .section-count {
    font-size: 0.72rem;
    color: #92a6bb;
    border: 1px solid #2f4459;
    border-radius: 999px;
    padding: 3px 8px;
    background-color: #1a2534;
  }

  .section-empty {
    margin: 0;
    color: #77899d;
    font-size: 0.82rem;
    padding: 10px 0 4px;
  }

  .count-pill {
    font-size: 0.74rem;
    color: #9cb0c5;
    border: 1px solid #2a3a4a;
    border-radius: 999px;
    padding: 3px 10px;
    background: #1a2534;
  }

  .template-table-wrap {
    border: 1px solid #2a3a4a;
    border-radius: 8px;
    overflow: auto;
    background-color: #1e2d3d;
  }

  .template-table {
    width: 100%;
    min-width: 1020px;
    border-collapse: collapse;
    table-layout: fixed;
  }

  .template-table .col-name {
    width: 18%;
  }

  .template-table .col-description {
    width: 31%;
  }

  .template-table .col-updated {
    width: 13%;
  }

  .template-table .col-access {
    width: 14%;
  }

  .template-table .col-actions {
    width: 24%;
  }

  .template-table th {
    text-align: left;
    color: #8ea2b8;
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 700;
    padding: 10px 12px;
    border-bottom: 1px solid #2a3a4a;
    background-color: #1a2534;
    white-space: nowrap;
  }

  .template-table td {
    padding: 10px 12px;
    border-bottom: 1px solid #25384b;
    vertical-align: middle;
  }

  .template-table tbody tr:last-child td {
    border-bottom: none;
  }

  .template-table tbody tr:hover {
    background-color: #1b2a3b;
  }

  .template-table tr.builtin-row {
    background-color: rgba(58, 74, 42, 0.1);
  }

  .tmpl-name {
    color: #dce5f0;
    font-size: 0.89rem;
    font-weight: 600;
  }

  .tmpl-description {
    color: #9bb0c4;
    font-size: 0.79rem;
    line-height: 1.3;
  }

  .tmpl-date {
    color: #8196ab;
    font-size: 0.76rem;
  }

  .access-badge {
    font-size: 0.7rem;
    border-radius: 999px;
    padding: 3px 9px;
    border: 1px solid;
    white-space: nowrap;
  }

  .access-badge.builtin {
    color: #c5dd9e;
    border-color: #4d6436;
    background-color: rgba(77, 100, 54, 0.18);
  }

  .access-badge.custom {
    color: #8dc5ff;
    border-color: #345e8c;
    background-color: rgba(52, 94, 140, 0.18);
  }

  .row-actions {
    display: flex;
    gap: 7px;
    flex-wrap: nowrap;
    align-items: center;
    white-space: nowrap;
  }

  .row-actions .btn {
    flex: 0 0 auto;
  }

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

  .btn:disabled {
    opacity: 0.55;
    cursor: not-allowed;
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
    padding: 4px 8px;
    font-size: 0.75rem;
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

  .dialog-overlay {
    position: fixed;
    inset: 0;
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
    width: 420px;
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
    min-height: 42px;
    padding: 8px 10px;
    background-color: #1b2636;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    color: #e0e0e0;
    font-size: 0.9rem;
    outline: none;
    transition: border-color 0.15s;
    font-family: inherit;
  }

  .form-input:focus {
    border-color: #4a8af4;
  }

  textarea.form-input {
    min-height: 90px;
    resize: vertical;
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }

  .confirm-text {
    color: #c0d0e0;
    font-size: 0.9rem;
    line-height: 1.5;
    margin: 0;
  }
</style>
