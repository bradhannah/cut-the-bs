<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { splitBulletText } from "../services/api";

  export let workHistoryId: number;

  const dispatch = createEventDispatcher<{
    confirm: { workHistoryId: number; lines: string[] };
    cancel: void;
  }>();

  let rawText = "";
  let previewLines: string[] = [];
  let loading = false;
  let previewed = false;

  async function handlePreview(): Promise<void> {
    if (rawText.trim().length === 0) return;
    loading = true;
    try {
      previewLines = await splitBulletText(rawText);
      previewed = true;
    } finally {
      loading = false;
    }
  }

  function handleConfirm(): void {
    dispatch("confirm", { workHistoryId, lines: previewLines });
  }

  function handleCancel(): void {
    dispatch("cancel");
  }

  function removePreviewLine(index: number): void {
    previewLines = previewLines.filter((_, i) => i !== index);
    if (previewLines.length === 0) {
      previewed = false;
    }
  }

  function handleOverlayClick(event: MouseEvent): void {
    if (event.target === event.currentTarget) {
      handleCancel();
    }
  }

  function handleOverlayKeydown(event: KeyboardEvent): void {
    if (event.key === "Escape") {
      handleCancel();
    }
  }
</script>

<div
  class="dialog-overlay"
  on:click={handleOverlayClick}
  on:keydown={handleOverlayKeydown}
  role="dialog"
  aria-modal="true"
  aria-label="Paste bullets"
  tabindex="-1"
>
  <div class="dialog">
    <h3 class="dialog-title">Paste Multiple Bullets</h3>
    <p class="dialog-description">
      Paste a block of text below. Each line will become a separate bullet.
    </p>

    {#if !previewed}
      <textarea
        class="paste-area"
        bind:value={rawText}
        rows="8"
        placeholder="Paste your text here...&#10;Each line becomes a bullet point."
      />
      <div class="dialog-actions">
        <button
          class="btn btn-primary"
          on:click={handlePreview}
          disabled={loading || rawText.trim().length === 0}
        >
          {loading ? "Processing..." : "Preview Split"}
        </button>
        <button class="btn btn-cancel" on:click={handleCancel}>Cancel</button>
      </div>
    {:else}
      <div class="preview-section">
        <p class="preview-label">
          {previewLines.length} bullet{previewLines.length !== 1 ? "s" : ""} will
          be created:
        </p>
        <ul class="preview-list">
          {#each previewLines as line, i (i)}
            <li class="preview-item">
              <span class="preview-marker">-</span>
              <span class="preview-text">{line}</span>
              <button
                class="btn-icon-remove"
                on:click={() => removePreviewLine(i)}
                title="Remove this line"
              >
                &#10005;
              </button>
            </li>
          {/each}
        </ul>
      </div>
      <div class="dialog-actions">
        <button
          class="btn btn-primary"
          on:click={handleConfirm}
          disabled={previewLines.length === 0}
        >
          Create {previewLines.length} Bullet{previewLines.length !== 1
            ? "s"
            : ""}
        </button>
        <button
          class="btn btn-secondary"
          on:click={() => {
            previewed = false;
          }}
        >
          Back
        </button>
        <button class="btn btn-cancel" on:click={handleCancel}>Cancel</button>
      </div>
    {/if}
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

  .paste-area {
    width: 100%;
    background-color: #1a2332;
    color: #e0e0e0;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 10px 12px;
    font-size: 0.9rem;
    font-family: inherit;
    resize: vertical;
    margin-bottom: 16px;
  }

  .paste-area:focus {
    outline: none;
    border-color: #4a8af4;
    box-shadow: 0 0 0 2px rgba(74, 138, 244, 0.15);
  }

  .preview-section {
    margin-bottom: 16px;
  }

  .preview-label {
    font-size: 0.85rem;
    color: #7a8a9a;
    margin: 0 0 8px;
  }

  .preview-list {
    list-style: none;
    margin: 0;
    padding: 0;
    max-height: 300px;
    overflow-y: auto;
  }

  .preview-item {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 6px 8px;
    border-bottom: 1px solid #1a2332;
    font-size: 0.9rem;
  }

  .preview-item:last-child {
    border-bottom: none;
  }

  .preview-marker {
    color: #4a8af4;
    font-weight: 700;
    flex-shrink: 0;
  }

  .preview-text {
    flex: 1;
    color: #c0d0e0;
  }

  .btn-icon-remove {
    background: none;
    border: none;
    color: #5a6a7a;
    cursor: pointer;
    padding: 0 4px;
    font-size: 0.75rem;
    flex-shrink: 0;
  }

  .btn-icon-remove:hover {
    color: #ff6b6b;
  }

  .dialog-actions {
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

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background-color: #2a5090;
    border-color: #3a60a0;
    color: #e0e0e0;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: #3a60a0;
  }

  .btn-secondary {
    background-color: #2a3a4a;
    border-color: #3a4a5a;
    color: #c0d0e0;
  }

  .btn-secondary:hover {
    background-color: #3a4a5a;
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
