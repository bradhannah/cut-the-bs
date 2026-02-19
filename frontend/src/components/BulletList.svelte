<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { AchievementBullet } from "../services/api";

  export let bullets: AchievementBullet[] = [];
  export let workHistoryId: number;

  const dispatch = createEventDispatcher<{
    create: { workHistoryId: number; text: string };
    update: { id: number; text: string };
    delete: { id: number };
    reorder: { workHistoryId: number; orderedIDs: number[] };
    paste: { workHistoryId: number };
  }>();

  let editingId: number | null = null;
  let editText = "";
  let newBulletText = "";
  let addingNew = false;

  // Drag state for bullet reordering
  let dragIndex: number | null = null;
  let dragOverIndex: number | null = null;

  function startEdit(bullet: AchievementBullet): void {
    editingId = bullet.id;
    editText = bullet.text;
  }

  function cancelEdit(): void {
    editingId = null;
    editText = "";
  }

  function saveEdit(): void {
    if (editingId === null) return;
    const trimmed = editText.trim();
    if (trimmed.length === 0) return;
    dispatch("update", { id: editingId, text: trimmed });
    editingId = null;
    editText = "";
  }

  function handleEditKeydown(event: KeyboardEvent): void {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      saveEdit();
    } else if (event.key === "Escape") {
      cancelEdit();
    }
  }

  function startAdd(): void {
    addingNew = true;
    newBulletText = "";
  }

  function cancelAdd(): void {
    addingNew = false;
    newBulletText = "";
  }

  function saveNew(): void {
    const trimmed = newBulletText.trim();
    if (trimmed.length === 0) return;
    dispatch("create", { workHistoryId: workHistoryId, text: trimmed });
    newBulletText = "";
    addingNew = false;
  }

  function handleNewKeydown(event: KeyboardEvent): void {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      saveNew();
    } else if (event.key === "Escape") {
      cancelAdd();
    }
  }

  function handleDelete(id: number): void {
    dispatch("delete", { id });
  }

  function handlePaste(): void {
    dispatch("paste", { workHistoryId: workHistoryId });
  }

  // --- Drag-reorder for bullets ---

  function handleDragStart(index: number, event: DragEvent): void {
    dragIndex = index;
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = "move";
      event.dataTransfer.setData("text/plain", String(index));
    }
  }

  function handleDragOver(index: number, event: DragEvent): void {
    event.preventDefault();
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = "move";
    }
    dragOverIndex = index;
  }

  function handleDragLeave(): void {
    dragOverIndex = null;
  }

  function handleDrop(index: number, event: DragEvent): void {
    event.preventDefault();
    if (dragIndex === null || dragIndex === index) {
      dragIndex = null;
      dragOverIndex = null;
      return;
    }
    const reordered = [...sortedBullets];
    const [moved] = reordered.splice(dragIndex, 1);
    reordered.splice(index, 0, moved);
    dragIndex = null;
    dragOverIndex = null;
    dispatch("reorder", {
      workHistoryId: workHistoryId,
      orderedIDs: reordered.map((b) => b.id),
    });
  }

  function handleDragEnd(): void {
    dragIndex = null;
    dragOverIndex = null;
  }

  $: sortedBullets = [...bullets].sort((a, b) => a.sort_order - b.sort_order);
</script>

<div class="bullet-list">
  {#if sortedBullets.length === 0 && !addingNew}
    <p class="empty-message">No achievement bullets yet.</p>
  {/if}

  <ul class="bullets">
    {#each sortedBullets as bullet, index (bullet.id)}
      <li
        class="bullet-item"
        class:editing={editingId === bullet.id}
        class:dragging={dragIndex === index}
        class:drag-over={dragOverIndex === index && dragIndex !== index}
        draggable={editingId !== bullet.id}
        on:dragstart={(e) => handleDragStart(index, e)}
        on:dragover={(e) => handleDragOver(index, e)}
        on:dragleave={handleDragLeave}
        on:drop={(e) => handleDrop(index, e)}
        on:dragend={handleDragEnd}
      >
        {#if editingId === bullet.id}
          <div class="bullet-edit">
            <textarea
              class="bullet-textarea"
              bind:value={editText}
              on:keydown={handleEditKeydown}
              rows="2"
            />
            <div class="bullet-edit-actions">
              <button class="btn btn-small btn-save" on:click={saveEdit}>
                Save
              </button>
              <button class="btn btn-small btn-cancel" on:click={cancelEdit}>
                Cancel
              </button>
            </div>
          </div>
        {:else}
          <div class="bullet-content">
            <span class="bullet-drag-handle" title="Drag to reorder">
              &#9776;
            </span>
            <span class="bullet-marker">-</span>
            <span
              class="bullet-text"
              on:dblclick={() => startEdit(bullet)}
              on:keydown={(e) => {
                if (e.key === "Enter") startEdit(bullet);
              }}
              role="button"
              tabindex="0"
              title="Double-click to edit"
            >
              {bullet.text}
            </span>
            <div class="bullet-actions">
              <button
                class="btn-icon"
                on:click={() => startEdit(bullet)}
                title="Edit"
              >
                &#9998;
              </button>
              <button
                class="btn-icon btn-icon-danger"
                on:click={() => handleDelete(bullet.id)}
                title="Delete"
              >
                &#10005;
              </button>
            </div>
          </div>
        {/if}
      </li>
    {/each}
  </ul>

  {#if addingNew}
    <div class="bullet-add-form">
      <textarea
        class="bullet-textarea"
        bind:value={newBulletText}
        on:keydown={handleNewKeydown}
        rows="2"
        placeholder="Type a bullet point..."
      />
      <div class="bullet-edit-actions">
        <button class="btn btn-small btn-save" on:click={saveNew}> Add </button>
        <button class="btn btn-small btn-cancel" on:click={cancelAdd}>
          Cancel
        </button>
      </div>
    </div>
  {/if}

  <div class="bullet-toolbar">
    {#if !addingNew}
      <button class="btn btn-small" on:click={startAdd}>+ Add Bullet</button>
    {/if}
    <button class="btn btn-small" on:click={handlePaste}>
      Paste Multiple
    </button>
  </div>
</div>

<style>
  .bullet-list {
    padding: 0 0 0 4px;
  }

  .empty-message {
    color: #5a6a7a;
    font-size: 0.85rem;
    font-style: italic;
    margin: 4px 0 8px;
  }

  .bullets {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .bullet-item {
    padding: 4px 0;
    border-bottom: 1px solid #1a2332;
  }

  .bullet-item:last-child {
    border-bottom: none;
  }

  .bullet-item.dragging {
    opacity: 0.4;
  }

  .bullet-item.drag-over {
    border-top: 2px solid #4a8af4;
  }

  .bullet-content {
    display: flex;
    align-items: flex-start;
    gap: 8px;
  }

  .bullet-drag-handle {
    cursor: grab;
    color: #3a4a5a;
    font-size: 0.7rem;
    flex-shrink: 0;
    user-select: none;
    padding: 0 2px;
    margin-top: 2px;
    transition: color 0.15s;
  }

  .bullet-drag-handle:hover {
    color: #7a8a9a;
  }

  .bullet-drag-handle:active {
    cursor: grabbing;
  }

  .bullet-marker {
    color: #4a8af4;
    font-weight: 700;
    flex-shrink: 0;
    margin-top: 1px;
  }

  .bullet-text {
    flex: 1;
    color: #c0d0e0;
    font-size: 0.9rem;
    line-height: 1.4;
    cursor: pointer;
    padding: 2px 4px;
    border-radius: 3px;
  }

  .bullet-text:hover {
    background-color: #1a2332;
  }

  .bullet-actions {
    display: flex;
    gap: 2px;
    flex-shrink: 0;
    opacity: 0;
    transition: opacity 0.15s;
  }

  .bullet-content:hover .bullet-actions {
    opacity: 1;
  }

  .btn-icon {
    background: none;
    border: none;
    color: #7a8a9a;
    cursor: pointer;
    padding: 2px 6px;
    font-size: 0.8rem;
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

  .bullet-edit,
  .bullet-add-form {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0;
  }

  .bullet-textarea {
    background-color: #1a2332;
    color: #e0e0e0;
    border: 1px solid #4a8af4;
    border-radius: 4px;
    padding: 8px 10px;
    font-size: 0.9rem;
    font-family: inherit;
    resize: vertical;
    width: 100%;
  }

  .bullet-textarea:focus {
    outline: none;
    border-color: #6aa0ff;
    box-shadow: 0 0 0 2px rgba(74, 138, 244, 0.15);
  }

  .bullet-edit-actions {
    display: flex;
    gap: 6px;
  }

  .bullet-toolbar {
    display: flex;
    gap: 8px;
    margin-top: 8px;
  }

  .btn {
    background-color: #2a3a4a;
    color: #c0d0e0;
    border: 1px solid #3a4a5a;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.85rem;
    padding: 6px 12px;
    transition:
      background-color 0.15s,
      border-color 0.15s;
  }

  .btn:hover {
    background-color: #3a4a5a;
    border-color: #4a5a6a;
  }

  .btn-small {
    padding: 4px 10px;
    font-size: 0.8rem;
  }

  .btn-save {
    background-color: #2a5040;
    border-color: #3a6050;
  }

  .btn-save:hover {
    background-color: #3a6050;
  }

  .btn-cancel {
    background-color: #3a3a4a;
    border-color: #4a4a5a;
  }

  .btn-cancel:hover {
    background-color: #4a4a5a;
  }
</style>
