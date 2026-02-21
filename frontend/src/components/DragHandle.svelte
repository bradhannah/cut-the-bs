<script lang="ts">
  import { createEventDispatcher } from "svelte";

  // The parent provides an ordered array of IDs representing the current order.
  // When a drag-drop reorder completes, we dispatch "reorder" with the new
  // ordered array of IDs.
  export let items: { id: number }[] = [];

  const dispatch = createEventDispatcher<{
    reorder: { orderedIDs: number[] };
  }>();

  let dragIndex: number | null = null;
  let dragOverIndex: number | null = null;

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

    const reordered = [...items];
    const [moved] = reordered.splice(dragIndex, 1);
    reordered.splice(index, 0, moved);

    dragIndex = null;
    dragOverIndex = null;

    dispatch("reorder", {
      orderedIDs: reordered.map((item) => item.id),
    });
  }

  function handleDragEnd(): void {
    dragIndex = null;
    dragOverIndex = null;
  }
</script>

<!--
  Usage: Wrap each list item in a <DragHandle> slot.
  DragHandle renders one draggable row per item.

  Example:
    <DragHandle items={entries} on:reorder={handleReorder} let:item let:index>
      <MyCard entry={item} />
    </DragHandle>
-->
{#each items as item, index (item.id)}
  <div
    class="drag-row"
    class:dragging={dragIndex === index}
    class:drag-over={dragOverIndex === index && dragIndex !== index}
    draggable="true"
    on:dragstart={(e) => handleDragStart(index, e)}
    on:dragover={(e) => handleDragOver(index, e)}
    on:dragleave={handleDragLeave}
    on:drop={(e) => handleDrop(index, e)}
    on:dragend={handleDragEnd}
    role="listitem"
  >
    <div class="drag-handle" title="Drag to reorder">
      <span class="drag-icon">&#9776;</span>
    </div>
    <div class="drag-content">
      <slot {item} {index} />
    </div>
  </div>
{/each}

<style>
  .drag-row {
    display: flex;
    align-items: stretch;
    transition:
      opacity 0.15s,
      border-color 0.15s;
  }

  .drag-row.dragging {
    opacity: 0.4;
  }

  .drag-row.drag-over {
    border-top: 2px solid #4a8af4;
  }

  .drag-handle {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    flex-shrink: 0;
    cursor: grab;
    color: #4a5a6a;
    user-select: none;
    transition: color 0.15s;
  }

  .drag-handle:hover {
    color: #8a9aaa;
  }

  .drag-handle:active {
    cursor: grabbing;
  }

  .drag-icon {
    font-size: 0.85rem;
  }

  .drag-content {
    flex: 1;
    min-width: 0;
  }
</style>
