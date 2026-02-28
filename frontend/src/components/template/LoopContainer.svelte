<script lang="ts">
  // LoopContainer.svelte — T032
  // Renders a nested drop zone inside a loop ElementBlock. Shows
  // child elements and accepts valid child types from the palette.
  import { createEventDispatcher } from "svelte";
  import { dndzone } from "svelte-dnd-action";
  import type { TemplateElement } from "../../services/api";
  import {
    createTemplateElement,
    reorderTemplateElements,
    addToast,
  } from "../../services/api";
  import {
    canvasElements,
    selectedElementId,
    currentTemplate,
    builderReadOnly,
    markSaving,
    markSaved,
    markSaveError,
  } from "../../stores/templateBuilder";
  import {
    elementLabels,
    elementIcons,
    defaultConfigs,
    validLoopChildren,
    loopIterationLabels,
  } from "./elementTypes";
  import { nativeDraggedElementType } from "./nativeDragState";

  export let parentElement: TemplateElement;
  export let children: TemplateElement[] = [];
  export let readOnly: boolean = false;

  const dispatch = createEventDispatcher<{
    select: { id: number };
    delete: { id: number };
  }>();

  interface ChildDndItem {
    id: number | string;
    element_type: string;
    [key: string]: any;
  }

  let dndItems: ChildDndItem[] = [];
  let dropIndicatorIndex: number | null = null;
  let reorderInFlight = false;

  // Rebuild dndItems from the children prop.
  function rebuildDndItems(): ChildDndItem[] {
    return [...(children || [])]
      .sort((a, b) => a.sort_order - b.sort_order)
      .map((el) => ({ id: el.id, element_type: el.element_type }));
  }

  function coerceDndID(id: number | string): number | string {
    if (typeof id === "number") return id;
    const trimmed = id.trim();
    if (trimmed !== "" && /^-?\d+$/.test(trimmed)) {
      const parsed = Number(trimmed);
      if (Number.isFinite(parsed)) {
        return parsed;
      }
    }
    return id;
  }

  function normalizeDndItems(items: ChildDndItem[]): ChildDndItem[] {
    return items.map((item) => ({
      ...item,
      id: coerceDndID(item.id),
    }));
  }

  function orderedIDsForReorder(items: ChildDndItem[]): number[] | null {
    const ids: number[] = [];
    for (const item of items) {
      if (typeof item.id !== "number") {
        return null;
      }
      ids.push(item.id);
    }
    return ids;
  }

  // NOTE: `children` must be referenced DIRECTLY here so Svelte 3's static
  // dependency tracker sees it and re-runs this block when the prop changes
  // (e.g. after a child is deleted and Canvas passes a new children array).
  $: if (!reorderInFlight) {
    dndItems = [...(children || [])]
      .sort((a, b) => a.sort_order - b.sort_order)
      .map((el) => ({ id: el.id, element_type: el.element_type }));
  }

  $: iterationLabel = loopIterationLabels[parentElement.element_type] || "";
  $: validChildren = validLoopChildren[parentElement.element_type] || new Set<string>();

  // Track whether an invalid element is being dragged over the zone.
  let invalidDragOver = false;

  function handleChildClick(e: MouseEvent, childId: number): void {
    e.stopPropagation();
    dispatch("select", { id: childId });
  }

  function handleChildMouseDown(e: MouseEvent, childId: number): void {
    if (e.button !== 0) return;
    e.stopPropagation();
    dispatch("select", { id: childId });
  }

  function handleChildDelete(e: MouseEvent, childId: number): void {
    e.stopPropagation();
    dispatch("delete", { id: childId });
  }

  function getDraggedElementType(e: DragEvent): string {
    const dt = e.dataTransfer;
    if (!dt) return $nativeDraggedElementType || "";

    const transferred = (
      dt.getData("application/x-template-element") ||
      dt.getData("text/plain") ||
      ""
    ).trim();
    return transferred || $nativeDraggedElementType || "";
  }

  function isTemplateElementDrag(e: DragEvent): boolean {
    if ($nativeDraggedElementType) return true;
    const dt = e.dataTransfer;
    if (!dt) return false;
    const types = Array.from(dt.types || []);
    return (
      types.includes("application/x-template-element") ||
      types.includes("text/plain")
    );
  }

  function getInsertIndex(container: HTMLElement, clientY: number): number {
    const children = Array.from(container.children).filter(
      (el) => !el.hasAttribute("data-drop-indicator")
    );
    if (children.length === 0) return 0;
    for (let i = 0; i < children.length; i++) {
      const rect = children[i].getBoundingClientRect();
      const mid = rect.top + rect.height / 2;
      if (clientY < mid) return i;
    }
    return children.length;
  }

  function handleNativeDragOver(e: DragEvent): void {
    if (readOnly || $builderReadOnly) return;
    if (!isTemplateElementDrag(e)) return;

    const elementType = getDraggedElementType(e);

    e.stopPropagation();
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = "copy";

    invalidDragOver = !validChildren.has(elementType);
    if (invalidDragOver) {
      dropIndicatorIndex = null;
      return;
    }

    if (dndItems.length === 0) {
      dropIndicatorIndex = 0;
      return;
    }

    const container = e.currentTarget as HTMLElement | null;
    if (!container) return;
    dropIndicatorIndex = getInsertIndex(container, e.clientY);
  }

  function handleNativeDragLeave(e: DragEvent): void {
    if (readOnly || $builderReadOnly) return;
    e.stopPropagation();
    const container = e.currentTarget as HTMLElement | null;
    const next = e.relatedTarget;
    if (container && next instanceof Node && container.contains(next)) {
      return;
    }
    invalidDragOver = false;
    dropIndicatorIndex = null;
  }

  async function handleNativeDrop(e: DragEvent): Promise<void> {
    if (readOnly || $builderReadOnly) return;
    const elementType = getDraggedElementType(e);
    if (!elementType) return;

    e.stopPropagation();
    e.preventDefault();

    const templateId = $currentTemplate?.id;
    if (!templateId) {
      invalidDragOver = false;
      dropIndicatorIndex = null;
      nativeDraggedElementType.set(null);
      return;
    }

    if (!validChildren.has(elementType)) {
      addToast(
        "error",
        `${elementLabels[elementType] || elementType} cannot be added to ${elementLabels[parentElement.element_type]}`
      );
      invalidDragOver = false;
      dropIndicatorIndex = null;
      nativeDraggedElementType.set(null);
      return;
    }

    const container = e.currentTarget as HTMLElement | null;
    const insertIndex =
      dndItems.length === 0
        ? 0
        : dropIndicatorIndex ??
          (container ? getInsertIndex(container, e.clientY) : dndItems.length);
    const config = defaultConfigs[elementType] || "{}";

    try {
      markSaving();
      const created = await createTemplateElement(templateId, {
        parent_id: parentElement.id,
        element_type: elementType,
        config,
      });

      canvasElements.update((els) => [...els, created]);

      const orderedIds = rebuildDndItems()
        .map((item) => item.id as number)
        .filter((id) => id !== created.id);
      orderedIds.splice(insertIndex, 0, created.id);

      await reorderTemplateElements(templateId, parentElement.id, orderedIds);

      canvasElements.update((els) =>
        els.map((el) => {
          const idx = orderedIds.indexOf(el.id);
          if (idx !== -1 && el.parent_id === parentElement.id) {
            return { ...el, sort_order: idx };
          }
          return el;
        })
      );
      markSaved();
    } catch (err: any) {
      console.error("[LoopContainer] createTemplateElement error:", err);
      markSaveError();
      addToast("error", err?.message || "Failed to add child element");
    } finally {
      invalidDragOver = false;
      dropIndicatorIndex = null;
      nativeDraggedElementType.set(null);
    }
  }

  async function handleConsider(
    e: CustomEvent<{ items: ChildDndItem[]; info?: { source?: string; trigger?: string } }>
  ): Promise<void> {
    if (readOnly || $builderReadOnly) return;
    const items = normalizeDndItems(e.detail.items);

    // Detect if an invalid element type is being dragged over.
    const newItem = items.find((item) => typeof item.id === "string");
    if (newItem) {
      const elementType = newItem.element_type || (newItem.id as string);
      invalidDragOver = !validChildren.has(elementType);
    } else {
      invalidDragOver = false;
    }
    dndItems = items;
  }

  async function handleFinalize(
    e: CustomEvent<{ items: ChildDndItem[]; info?: { source?: string; trigger?: string } }>
  ): Promise<void> {
    if (readOnly || $builderReadOnly) return;
    const items = normalizeDndItems(e.detail.items);
    dndItems = items;

    const orderedIds = orderedIDsForReorder(items);
    if (!orderedIds) {
      invalidDragOver = false;
      dndItems = rebuildDndItems();
      return;
    }

    invalidDragOver = false;
    const templateId = $currentTemplate?.id;

    if (!templateId) {
      dndItems = rebuildDndItems();
      return;
    }

    // svelte-dnd-action requires that `items` is updated synchronously when
    // `finalize` fires — before any awaited work — otherwise the library will
    // revert its visual state and the dropped element disappears from the loop.
    reorderInFlight = true;

    try {
      markSaving();
      await reorderTemplateElements(templateId, parentElement.id, orderedIds);

      canvasElements.update((els) =>
        els.map((el) => {
          const idx = orderedIds.indexOf(el.id);
          if (idx !== -1 && el.parent_id === parentElement.id) {
            return { ...el, sort_order: idx };
          }
          return el;
        })
      );
      markSaved();
    } catch (err: any) {
      console.error("[LoopContainer] reorder error:", err);
      markSaveError();
      addToast("error", err?.message || "Failed to reorder child elements");
      dndItems = rebuildDndItems();
    } finally {
      reorderInFlight = false;
    }
  }
</script>

<div class="loop-container">
  <div class="loop-label">{iterationLabel}</div>

  {#if invalidDragOver}
    <div class="invalid-drop-hint">Not a valid child element for this loop</div>
  {/if}

  {#if dndItems.length === 0}
    <div
      class="loop-drop-empty"
      class:invalid-drop={invalidDragOver}
      use:dndzone={{
        items: dndItems,
        type: "template-element",
        dropTargetStyle: { outline: invalidDragOver ? "2px dashed #ff6b6b" : "2px dashed #7a6af4", outlineOffset: "-2px" },
        dropFromOthersDisabled: true,
        dragDisabled: readOnly || $builderReadOnly,
        centreDraggedOnCursor: false,
        useCursorForDetection: true,
      }}
      on:consider={handleConsider}
      on:finalize={handleFinalize}
      on:dragover={handleNativeDragOver}
      on:drop={handleNativeDrop}
      on:dragleave={handleNativeDragLeave}
    >
      {#if dropIndicatorIndex === 0}
        <div class="drop-indicator" data-drop-indicator></div>
      {/if}
      <p class="loop-empty-hint">Drop child elements here</p>
    </div>
  {:else}
    <div
      class="loop-children"
      class:invalid-drop={invalidDragOver}
      use:dndzone={{
        items: dndItems,
        type: "template-element",
        dropTargetStyle: { outline: invalidDragOver ? "2px dashed #ff6b6b" : "2px dashed #7a6af4", outlineOffset: "-2px" },
        dropFromOthersDisabled: true,
        dragDisabled: readOnly || $builderReadOnly,
        centreDraggedOnCursor: false,
        useCursorForDetection: true,
      }}
      on:consider={handleConsider}
      on:finalize={handleFinalize}
      on:dragover={handleNativeDragOver}
      on:drop={handleNativeDrop}
      on:dragleave={handleNativeDragLeave}
    >
      {#each dndItems as child, i (child.id)}
        {#if dropIndicatorIndex === i}
          <div class="drop-indicator" data-drop-indicator></div>
        {/if}
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <div
          class="child-element"
          class:selected={$selectedElementId === child.id}
          data-template-element-id={Number(child.id)}
          on:mousedown={(e) => handleChildMouseDown(e, Number(child.id))}
          on:click={(e) => handleChildClick(e, Number(child.id))}
          role="button"
          tabindex="0"
        >
          <span class="drag-handle">&#x2630;</span>
          <span class="child-icon">{elementIcons[child.element_type] || "?"}</span>
          <span class="child-label">{elementLabels[child.element_type] || child.element_type}</span>
          {#if !(readOnly || $builderReadOnly)}
            <button
              class="delete-btn"
              title="Delete child element"
              on:click={(e) => handleChildDelete(e, Number(child.id))}
            >
              &times;
            </button>
          {/if}
        </div>
      {/each}
      {#if dropIndicatorIndex === dndItems.length}
        <div class="drop-indicator" data-drop-indicator></div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .loop-container {
    padding: 4px 10px 10px;
    border-top: 1px solid #2a3a4a;
  }

  .loop-label {
    font-size: 0.7rem;
    color: #7a6af4;
    padding: 4px 0 6px;
    font-style: italic;
  }

  .invalid-drop-hint {
    font-size: 0.7rem;
    color: #ff6b6b;
    padding: 2px 0 4px;
    font-style: italic;
  }

  .loop-drop-empty {
    min-height: 50px;
    border: 2px dashed #3a3a5a;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: border-color 0.15s;
  }

  .loop-drop-empty.invalid-drop {
    border-color: #ff6b6b;
    background-color: rgba(255, 107, 107, 0.05);
  }

  .loop-empty-hint {
    color: #5a6a7a;
    font-size: 0.78rem;
    margin: 0;
  }

  .loop-children {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-height: 30px;
    transition: border-color 0.15s;
  }

  .loop-children.invalid-drop {
    outline: 2px dashed #ff6b6b;
    outline-offset: -2px;
    background-color: rgba(255, 107, 107, 0.05);
    border-radius: 4px;
  }

  .child-element {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    border: 1px solid #2a3a4a;
    border-radius: 3px;
    background-color: #1e2d3d;
    cursor: pointer;
    transition: border-color 0.12s;
    outline: none;
  }

  .child-element:hover {
    border-color: #3a4a5a;
  }

  .child-element.selected {
    border-color: #4a8af4;
    box-shadow: 0 0 0 1px #4a8af4;
  }

  .drag-handle {
    cursor: grab;
    color: #5a6a7a;
    font-size: 0.8rem;
    user-select: none;
    flex-shrink: 0;
  }

  .child-icon {
    font-size: 0.75rem;
    flex-shrink: 0;
  }

  .child-label {
    font-size: 0.78rem;
    color: #c0d0e0;
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .delete-btn {
    margin-left: auto;
    background: none;
    border: none;
    color: #5a6a7a;
    font-size: 1rem;
    cursor: pointer;
    padding: 0 3px;
    line-height: 1;
    border-radius: 3px;
    transition: color 0.12s;
    flex-shrink: 0;
  }

  .delete-btn:hover {
    color: #ff6b6b;
  }

  .drop-indicator {
    height: 2px;
    background: #4a8af4;
    border-radius: 1px;
    margin: 1px 0;
    pointer-events: none;
  }
</style>
