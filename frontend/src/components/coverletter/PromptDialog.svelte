<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { TemplateVariable, GuidedPrompt } from "../../services/api";

  export let variables: TemplateVariable[] = [];
  export let prompts: GuidedPrompt[] = [];
  /** Pre-filled values (e.g., from job application context). */
  export let prefilled: Record<string, string> = {};

  const dispatch = createEventDispatcher<{
    submit: { substitutions: Record<string, string> };
    cancel: void;
  }>();

  // Initialise value map from prefilled + blank for remaining.
  let values: Record<string, string> = {};
  $: {
    const map: Record<string, string> = {};
    for (const v of variables) {
      map[v.name] = prefilled[v.name] ?? "";
    }
    for (const p of prompts) {
      const key = "prompt:" + p.prompt_text;
      map[key] = prefilled[key] ?? "";
    }
    values = map;
  }

  function handleSubmit(): void {
    dispatch("submit", { substitutions: { ...values } });
  }

  function handleCancel(): void {
    dispatch("cancel");
  }

  function handleOverlayClick(event: MouseEvent): void {
    if (event.target === event.currentTarget) {
      handleCancel();
    }
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key === "Escape") {
      handleCancel();
    }
  }

  /** Pretty label for a variable name (e.g. company_name → Company Name). */
  function labelFor(name: string): string {
    return name
      .split("_")
      .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
      .join(" ");
  }

  $: hasContent = variables.length > 0 || prompts.length > 0;
</script>

<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<div
  class="dialog-overlay"
  on:click={handleOverlayClick}
  on:keydown={handleKeydown}
  role="dialog"
  aria-modal="true"
  aria-label="Cover letter variables"
  tabindex="-1"
>
  <div class="dialog">
    <h3 class="dialog-title">Cover Letter Variables</h3>
    <p class="dialog-description">
      Fill in the values below. They will be substituted into the cover letter
      template before generating the PDF.
    </p>

    {#if !hasContent}
      <p class="empty-state">
        This template has no variable placeholders or guided prompts.
      </p>
    {/if}

    {#if variables.length > 0}
      <div class="section">
        <h4 class="section-label">Variables</h4>
        {#each variables as v (v.name)}
          <div class="field">
            <label class="field-label" for="var-{v.name}">
              {labelFor(v.name)}
            </label>
            <input
              id="var-{v.name}"
              type="text"
              class="field-input"
              placeholder="Enter {labelFor(v.name).toLowerCase()}..."
              bind:value={values[v.name]}
            />
          </div>
        {/each}
      </div>
    {/if}

    {#if prompts.length > 0}
      <div class="section">
        <h4 class="section-label">Guided Prompts</h4>
        {#each prompts as p, i (p.prompt_text)}
          <div class="field">
            <label class="field-label" for="prompt-{i}">
              {p.prompt_text}
            </label>
            <textarea
              id="prompt-{i}"
              class="field-textarea"
              rows="4"
              placeholder="Write your response..."
              bind:value={values["prompt:" + p.prompt_text]}
            />
          </div>
        {/each}
      </div>
    {/if}

    <div class="dialog-actions">
      <button class="btn btn-cancel" on:click={handleCancel}>Cancel</button>
      <button class="btn btn-primary" on:click={handleSubmit}>
        Generate PDF
      </button>
    </div>
  </div>
</div>

<style>
  .dialog-overlay {
    position: fixed;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 900;
  }

  .dialog {
    background-color: #1e2d3d;
    border: 1px solid #3a4a5a;
    border-radius: 8px;
    padding: 24px;
    width: 90%;
    max-width: 560px;
    max-height: 80vh;
    overflow-y: auto;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  }

  .dialog-title {
    margin: 0 0 8px;
    font-size: 1.1rem;
    color: #e0e0e0;
  }

  .dialog-description {
    margin: 0 0 16px;
    font-size: 0.85rem;
    color: #7a8a9a;
  }

  .empty-state {
    margin: 12px 0 16px;
    font-size: 0.85rem;
    color: #7a8a9a;
    font-style: italic;
  }

  .section {
    margin-bottom: 16px;
  }

  .section-label {
    margin: 0 0 10px;
    font-size: 0.9rem;
    color: #c0d0e0;
    border-bottom: 1px solid #2a3a4a;
    padding-bottom: 6px;
  }

  .field {
    margin-bottom: 12px;
  }

  .field-label {
    display: block;
    margin-bottom: 4px;
    font-size: 0.82rem;
    color: #c0d0e0;
  }

  .field-input,
  .field-textarea {
    width: 100%;
    padding: 8px 10px;
    background-color: #1a2332;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    color: #e0e0e0;
    font-size: 0.85rem;
    font-family: inherit;
    box-sizing: border-box;
  }

  .field-input:focus,
  .field-textarea:focus {
    outline: none;
    border-color: #4a8af4;
  }

  .field-textarea {
    resize: vertical;
    min-height: 80px;
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 8px;
  }
</style>
