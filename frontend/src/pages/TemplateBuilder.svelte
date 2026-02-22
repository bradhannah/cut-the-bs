<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    getDocumentTemplate,
    previewTemplate,
    openFile,
    addToast,
  } from "../services/api";
  import {
    currentTemplate,
    resetBuilderStores,
    loadTemplateIntoStores,
    saveStatus,
  } from "../stores/templateBuilder";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";
  import Palette from "../components/template/Palette.svelte";
  import Canvas from "../components/template/Canvas.svelte";
  import Properties from "../components/template/Properties.svelte";

  export let params: { id?: string } = {};

  let loading = true;
  let error = "";
  let previewing = false;

  $: templateId = params.id ? parseInt(params.id, 10) : null;

  onMount(async () => {
    if (!templateId) {
      error = "No template ID provided.";
      loading = false;
      return;
    }
    await loadTemplate(templateId);
  });

  onDestroy(() => {
    resetBuilderStores();
  });

  async function loadTemplate(id: number): Promise<void> {
    loading = true;
    error = "";
    try {
      const detail = await getDocumentTemplate(id);
      loadTemplateIntoStores(detail);
    } catch (e: any) {
      error = e?.message || "Failed to load template.";
      addToast("error", error);
    } finally {
      loading = false;
    }
  }

  async function handlePreview(): Promise<void> {
    if (!templateId || previewing) return;
    previewing = true;
    try {
      const pdfPath = await previewTemplate(templateId);
      await openFile(pdfPath);
    } catch (e: any) {
      addToast("error", e?.message || "Preview failed.");
    } finally {
      previewing = false;
    }
  }
</script>

{#if loading}
  <div class="builder-loading">
    <LoadingSpinner />
  </div>
{:else if error}
  <div class="builder-error">
    <h2>Error</h2>
    <p>{error}</p>
    <a href="#/templates" class="btn btn-primary">Back to Templates</a>
  </div>
{:else if $currentTemplate}
  <div class="builder-header">
    <a href="#/templates" class="back-link">&larr; Templates</a>
    <h2 class="builder-title">{$currentTemplate.name}</h2>
    <span class="template-type-badge">{$currentTemplate.template_type}</span>
    {#if $currentTemplate.is_builtin}
      <span class="builtin-badge">Built-in</span>
    {/if}
    <span class="header-spacer"></span>
    <button
      class="btn btn-small btn-preview"
      on:click={handlePreview}
      disabled={previewing}
      title="Generate a preview PDF and open it"
    >
      {previewing ? "Generating..." : "Preview PDF"}
    </button>
    {#if $saveStatus === "saving"}
      <span class="save-indicator save-saving">Saving...</span>
    {:else if $saveStatus === "saved"}
      <span class="save-indicator save-saved">Saved</span>
    {:else if $saveStatus === "error"}
      <span class="save-indicator save-error">Save failed</span>
    {/if}
  </div>

  <div class="builder-panels">
    <div class="panel palette-panel">
      <Palette />
    </div>

    <div class="panel canvas-panel">
      <Canvas />
    </div>

    <div class="panel properties-panel">
      <Properties />
    </div>
  </div>
{/if}

<style>
  /* --- Loading / Error states --- */
  .builder-loading {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 400px;
  }

  .builder-error {
    text-align: center;
    padding: 60px 20px;
  }

  .builder-error h2 {
    margin: 0 0 12px;
    font-size: 1.5rem;
    color: #e0e0e0;
  }

  .builder-error p {
    color: #7a8a9a;
    font-size: 0.95rem;
    margin-bottom: 20px;
  }

  /* --- Header --- */
  .builder-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding-bottom: 16px;
    border-bottom: 1px solid #2a3a4a;
    margin-bottom: 16px;
  }

  .back-link {
    color: #4a8af4;
    text-decoration: none;
    font-size: 0.85rem;
    white-space: nowrap;
  }

  .back-link:hover {
    text-decoration: underline;
  }

  .builder-title {
    margin: 0;
    font-size: 1.2rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .template-type-badge {
    font-size: 0.75rem;
    padding: 2px 8px;
    border-radius: 4px;
    background-color: #2a3a4a;
    color: #c0d0e0;
    text-transform: capitalize;
  }

  .builtin-badge {
    font-size: 0.75rem;
    padding: 2px 8px;
    border-radius: 4px;
    background-color: #3a4a2a;
    color: #b0d080;
  }

  .header-spacer {
    flex: 1;
  }

  .btn-preview {
    background-color: #2a5a3a;
    color: #c0e0c0;
    font-size: 0.8rem;
    padding: 5px 14px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    white-space: nowrap;
    transition: background-color 0.15s;
  }

  .btn-preview:hover:not(:disabled) {
    background-color: #3a6a4a;
  }

  .btn-preview:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .save-indicator {
    font-size: 0.78rem;
    padding: 2px 10px;
    border-radius: 4px;
    white-space: nowrap;
    transition: opacity 0.2s;
  }

  .save-saving {
    color: #7a8a9a;
  }

  .save-saved {
    color: #6bc96b;
  }

  .save-error {
    color: #ff6b6b;
  }

  /* --- Three-panel layout --- */
  .builder-panels {
    display: flex;
    gap: 0;
    /* Fill remaining height: viewport minus sidebar-header area, builder-header, and status bar */
    height: calc(100vh - 120px);
    margin: 0 -24px -24px; /* bleed into .content padding */
  }

  .panel {
    display: flex;
    flex-direction: column;
    border-right: 1px solid #2a3a4a;
  }

  .panel:last-child {
    border-right: none;
  }

  /* Palette — fixed 240px left */
  .palette-panel {
    width: 240px;
    min-width: 240px;
    background-color: #1a2332;
    overflow-y: auto;
  }

  /* Canvas — flexible center */
  .canvas-panel {
    flex: 1;
    min-width: 0;
    background-color: #1e2d3d;
    overflow: hidden;
  }

  /* Properties — fixed 300px right */
  .properties-panel {
    width: 300px;
    min-width: 300px;
    background-color: #1a2332;
    border-left: 1px solid #2a3a4a;
    border-right: none;
    overflow-y: auto;
  }

  /* --- Button resets (used by error state) --- */
  :global(.btn) {
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

  :global(.btn-primary) {
    background-color: #4a8af4;
    color: #ffffff;
  }

  :global(.btn-primary:hover) {
    background-color: #3a7ae4;
  }
</style>
