<script lang="ts">
  // Canvas.svelte — T030
  // Main drop zone that receives elements from the Palette, displays
  // ordered ElementBlock components, and supports reorder via drag.
  import { dndzone } from "svelte-dnd-action";
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
    createTemplateElement,
    deleteTemplateElement,
    reorderTemplateElements,
    addToast,
    type TemplateElement,
  } from "../../services/api";
  import {
    defaultConfigs,
    loopElementTypes,
    isElementTypeValid,
    validLoopChildren,
    elementLabels,
  } from "./elementTypes";
  import { nativeDraggedElementType } from "./nativeDragState";
  import ElementBlock from "./ElementBlock.svelte";

  // DnD items need an `id`; we build a wrapper that the DnD zone works with.
  interface CanvasDndItem {
    id: number | string;
    element_type: string;
    isNew?: boolean; // true for items dragged from palette (not yet persisted)
    [key: string]: any;
  }

  let dndItems: CanvasDndItem[] = [];
  let canvasZoneEl: HTMLDivElement | null = null;
  let dropIndicatorIndex: number | null = null;
  let reorderInFlight = false;

  // Rebuild dndItems from the store.  Uses == null so that parent_id values
  // of both null (top-level elements) and undefined (Wails may omit the key
  // for nil *int64) are treated as top-level.
  function rebuildDndItems(): CanvasDndItem[] {
    return ($canvasElements || [])
      // eslint-disable-next-line eqeqeq
      .filter((el) => el.parent_id == null || el.parent_id === 0)
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

  function normalizeDndItems(items: CanvasDndItem[]): CanvasDndItem[] {
    return items.map((item) => ({
      ...item,
      id: coerceDndID(item.id),
    }));
  }

  function orderedIDsForReorder(items: CanvasDndItem[]): number[] | null {
    const ids: number[] = [];
    for (const item of items) {
      if (typeof item.id !== "number") {
        return null;
      }
      ids.push(item.id);
    }
    return ids;
  }

  // Keep dndItems in sync with the store except during async reorder finalize.
  // NOTE: $canvasElements must be referenced DIRECTLY here (not inside a called
  // function) so Svelte 3's static dependency tracker sees it and re-runs this
  // block whenever the store changes (e.g. after add / delete).
  $: if (!reorderInFlight) {
    dndItems = ($canvasElements || [])
      // eslint-disable-next-line eqeqeq
      .filter((el) => el.parent_id == null || el.parent_id === 0)
      .sort((a, b) => a.sort_order - b.sort_order)
      .map((el) => ({ id: el.id, element_type: el.element_type }));
  }

  // Get children for a given parent element.
  function childrenOf(parentId: number): TemplateElement[] {
    return ($canvasElements || [])
      .filter((el) => el.parent_id === parentId)
      .sort((a, b) => a.sort_order - b.sort_order);
  }

  // Get the full element from the store by id.
  function getElement(id: number): TemplateElement | undefined {
    return ($canvasElements || []).find((el) => el.id === id);
  }

  function handleCanvasClick(): void {
    selectedElementId.set(null);
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

  function loopParentForChildType(childType: string): string | null {
    const loopTypes = Object.keys(validLoopChildren);
    for (const loopType of loopTypes) {
      if (validLoopChildren[loopType]?.has(childType)) {
        return loopType;
      }
    }
    return null;
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
    if ($builderReadOnly) return;
    if (!canvasZoneEl || !isTemplateElementDrag(e)) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = "copy";
    dropIndicatorIndex = getInsertIndex(canvasZoneEl, e.clientY);
  }

  function handleNativeDragLeave(e: DragEvent): void {
    if ($builderReadOnly) return;
    const next = e.relatedTarget;
    if (canvasZoneEl && next instanceof Node && canvasZoneEl.contains(next)) {
      return;
    }
    dropIndicatorIndex = null;
  }

  async function handleNativeDrop(e: DragEvent): Promise<void> {
    if ($builderReadOnly) return;
    const elementType = getDraggedElementType(e);
    if (!elementType || !canvasZoneEl) return;

    e.preventDefault();

    const templateId = $currentTemplate?.id;
    if (!templateId) {
      dropIndicatorIndex = null;
      nativeDraggedElementType.set(null);
      return;
    }

    const templateType = $currentTemplate?.template_type || "resume";
    if (!isElementTypeValid(elementType, templateType)) {
      const parentLoopType = loopParentForChildType(elementType);
      if (templateType === "resume" && parentLoopType) {
        addToast(
          "error",
          `${elementLabels[elementType] || elementType} can only be added inside ${elementLabels[parentLoopType] || parentLoopType}`
        );
        dropIndicatorIndex = null;
        nativeDraggedElementType.set(null);
        return;
      }
      addToast(
        "error",
        `"${elementType}" is not compatible with ${templateType} templates`
      );
      dropIndicatorIndex = null;
      nativeDraggedElementType.set(null);
      return;
    }

    const insertIndex =
      dropIndicatorIndex ?? getInsertIndex(canvasZoneEl, e.clientY);
    const config = defaultConfigs[elementType] || "{}";

    try {
      markSaving();
      const created = await createTemplateElement(templateId, {
        parent_id: null,
        element_type: elementType,
        config,
      });

      canvasElements.update((els) => [...els, created]);

      const orderedIds = rebuildDndItems()
        .map((item) => item.id as number)
        .filter((id) => id !== created.id);
      orderedIds.splice(insertIndex, 0, created.id);

      await reorderTemplateElements(templateId, null, orderedIds);

      canvasElements.update((els) =>
        els.map((el) => {
          const idx = orderedIds.indexOf(el.id);
          // eslint-disable-next-line eqeqeq
          if (idx !== -1 && (el.parent_id == null || el.parent_id === 0)) {
            return { ...el, sort_order: idx };
          }
          return el;
        })
      );
      markSaved();
    } catch (err: any) {
      console.error("[Canvas] createTemplateElement error:", err);
      markSaveError();
      addToast("error", err?.message || "Failed to add element");
    } finally {
      dropIndicatorIndex = null;
      nativeDraggedElementType.set(null);
    }
  }

  async function handleConsider(
    e: CustomEvent<{ items: CanvasDndItem[]; info: { source: string; trigger: string } }>
  ): Promise<void> {
    if ($builderReadOnly) return;
    dndItems = normalizeDndItems(e.detail.items);
  }

  async function handleFinalize(
    e: CustomEvent<{ items: CanvasDndItem[]; info: { source: string; trigger: string } }>
  ): Promise<void> {
    if ($builderReadOnly) return;
    const items = normalizeDndItems(e.detail.items);
    dndItems = items;

    const orderedIds = orderedIDsForReorder(items);
    if (!orderedIds) {
      dndItems = rebuildDndItems();
      return;
    }

    const templateId = $currentTemplate?.id;

    if (!templateId) {
      dndItems = rebuildDndItems();
      return;
    }

    // svelte-dnd-action requires that `items` is updated synchronously when
    // `finalize` fires — before any awaited work — otherwise the library will
    // revert its visual state and the dropped element disappears from the canvas.
    dndItems = items;
    reorderInFlight = true;

    try {
      markSaving();
      await reorderTemplateElements(templateId, null, orderedIds);

      canvasElements.update((els) =>
        els.map((el) => {
          const idx = orderedIds.indexOf(el.id);
          // eslint-disable-next-line eqeqeq
          if (idx !== -1 && (el.parent_id == null || el.parent_id === 0)) {
            return { ...el, sort_order: idx };
          }
          return el;
        })
      );
      markSaved();
    } catch (err: any) {
      console.error("[Canvas] reorder error:", err);
      markSaveError();
      addToast("error", err?.message || "Failed to reorder elements");
      dndItems = rebuildDndItems();
    } finally {
      reorderInFlight = false;
    }
  }

  async function handleDeleteElement(e: CustomEvent<{ id: number }>): Promise<void> {
    if ($builderReadOnly) return;
    const elementId = e.detail.id;
    try {
      markSaving();
      await deleteTemplateElement(elementId);
      canvasElements.update((els) =>
        els.filter((el) => el.id !== elementId && el.parent_id !== elementId)
      );
      if ($selectedElementId === elementId) {
        selectedElementId.set(null);
      }
      markSaved();
    } catch (err: any) {
      markSaveError();
      addToast("error", err?.message || "Failed to delete element");
    } finally {
      dropIndicatorIndex = null;
    }
  }

  function handleSelectElement(e: CustomEvent<{ id: number }>): void {
    selectedElementId.set(e.detail.id);
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div class="canvas" on:click|self={handleCanvasClick}>
  {#if dndItems.length === 0}
    <div class="canvas-empty-overlay">
      <p>Drag elements from the palette to start building your template.</p>
    </div>
  {/if}
  <div
    class="canvas-elements"
    class:readonly={$builderReadOnly}
    bind:this={canvasZoneEl}
    use:dndzone={{
      items: dndItems,
      type: "template-element",
      dropTargetStyle: { outline: "2px dashed #4a8af4", outlineOffset: "-2px" },
      dropFromOthersDisabled: true,
      dragDisabled: $builderReadOnly,
      centreDraggedOnCursor: false,
      useCursorForDetection: true,
    }}
    on:consider={handleConsider}
    on:finalize={handleFinalize}
    on:dragover={handleNativeDragOver}
    on:drop={handleNativeDrop}
    on:dragleave={handleNativeDragLeave}
  >
    {#each dndItems as item, i (item.id)}
      {#if dropIndicatorIndex === i}
        <div class="drop-indicator" data-drop-indicator></div>
      {/if}
      {#if typeof item.id === "number" && getElement(item.id)}
        <ElementBlock
          element={getElement(item.id)}
          children={loopElementTypes.has(getElement(item.id).element_type) ? childrenOf(item.id) : []}
          selected={$selectedElementId === item.id}
          readOnly={$builderReadOnly}
          on:select={handleSelectElement}
          on:delete={handleDeleteElement}
        />
      {:else}
        <!-- Placeholder for items being dragged in from palette -->
        <div class="canvas-placeholder-block">
          <span class="placeholder-label">{item.element_type || item.id}</span>
        </div>
      {/if}
    {/each}
    {#if dropIndicatorIndex === dndItems.length}
      <div class="drop-indicator" data-drop-indicator></div>
    {/if}
  </div>
</div>

<style>
  .canvas {
    height: 100%;
    padding: 16px;
    position: relative;
    display: flex;
    flex-direction: column;
  }

  .canvas-empty-overlay {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none;
    z-index: 1;
    border: 2px dashed #2a3a4a;
    border-radius: 8px;
    margin: 16px;
  }

  .canvas-empty-overlay p {
    color: #7a8a9a;
    font-size: 0.9rem;
    text-align: center;
    max-width: 280px;
  }

  .canvas-elements {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-height: 220px;
    height: 100%;
    overflow-y: auto;
  }

  .canvas-elements.readonly {
    opacity: 0.95;
  }

  .canvas-placeholder-block {
    padding: 10px 14px;
    border: 2px dashed #4a8af4;
    border-radius: 4px;
    background-color: rgba(74, 138, 244, 0.08);
    color: #4a8af4;
    font-size: 0.82rem;
  }

  .placeholder-label {
    text-transform: capitalize;
  }

  .drop-indicator {
    height: 2px;
    background: #4a8af4;
    border-radius: 1px;
    margin: 1px 0;
    pointer-events: none;
  }
</style>
