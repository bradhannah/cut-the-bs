<script lang="ts">
  // Canvas.svelte — T030
  // Main drop zone that receives elements from the Palette, displays
  // ordered ElementBlock components, and supports reorder via drag.
  import { dndzone } from "svelte-dnd-action";
  import {
    canvasElements,
    selectedElementId,
    currentTemplate,
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
  import { defaultConfigs, loopElementTypes, isElementTypeValid } from "./elementTypes";
  import ElementBlock from "./ElementBlock.svelte";

  // DnD items need an `id`; we build a wrapper that the DnD zone works with.
  interface CanvasDndItem {
    id: number | string;
    element_type: string;
    isNew?: boolean; // true for items dragged from palette (not yet persisted)
    [key: string]: any;
  }

  let dndItems: CanvasDndItem[] = [];

  // Keep dndItems in sync with the store — top-level elements only.
  $: {
    dndItems = ($canvasElements || [])
      .filter((el) => el.parent_id === null || el.parent_id === 0)
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

  async function handleConsider(
    e: CustomEvent<{ items: CanvasDndItem[]; info: { source: string; trigger: string } }>
  ): Promise<void> {
    dndItems = e.detail.items;
  }

  async function handleFinalize(
    e: CustomEvent<{ items: CanvasDndItem[]; info: { source: string; trigger: string } }>
  ): Promise<void> {
    const items = e.detail.items;
    const templateId = $currentTemplate?.id;
    if (!templateId) return;

    // Check if there's a new item (dragged from palette — has a string id).
    const newItem = items.find((item) => typeof item.id === "string");

    if (newItem) {
      const elementType = newItem.element_type || (newItem.id as string);

      // Validate element type compatibility with the template type.
      const templateType = $currentTemplate?.template_type || "resume";
      if (!isElementTypeValid(elementType, templateType)) {
        addToast(
          "error",
          `"${elementType}" is not compatible with ${templateType} templates`
        );
        // Rebuild dndItems from the store (removes the rejected item).
        dndItems = ($canvasElements || [])
          .filter((el) => el.parent_id === null || el.parent_id === 0)
          .sort((a, b) => a.sort_order - b.sort_order)
          .map((el) => ({ id: el.id, element_type: el.element_type }));
        return;
      }

      const config = defaultConfigs[elementType] || "{}";

      try {
        markSaving();
        const created = await createTemplateElement(templateId, {
          parent_id: null,
          element_type: elementType,
          config: config,
        });

        // Add the new element to the store.
        canvasElements.update((els) => [...els, created]);

        // Now reorder to match the drop position.
        const orderedIds = items.map((item) =>
          typeof item.id === "string" ? created.id : (item.id as number)
        );

        await reorderTemplateElements(templateId, null, orderedIds);

        // Update sort_order in the store.
        canvasElements.update((els) =>
          els.map((el) => {
            const idx = orderedIds.indexOf(el.id);
            if (idx !== -1 && (el.parent_id === null || el.parent_id === 0)) {
              return { ...el, sort_order: idx };
            }
            return el;
          })
        );
        markSaved();
      } catch (err: any) {
        markSaveError();
        addToast("error", err?.message || "Failed to add element");
      }
    } else {
      // Pure reorder of existing items.
      const orderedIds = items.map((item) => item.id as number);

      try {
        markSaving();
        await reorderTemplateElements(templateId, null, orderedIds);

        canvasElements.update((els) =>
          els.map((el) => {
            const idx = orderedIds.indexOf(el.id);
            if (idx !== -1 && (el.parent_id === null || el.parent_id === 0)) {
              return { ...el, sort_order: idx };
            }
            return el;
          })
        );
        markSaved();
      } catch (err: any) {
        markSaveError();
        addToast("error", err?.message || "Failed to reorder elements");
      }
    }
  }

  async function handleDeleteElement(e: CustomEvent<{ id: number }>): Promise<void> {
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
    use:dndzone={{
      items: dndItems,
      type: "template-element",
      dropTargetStyle: { outline: "2px dashed #4a8af4", outlineOffset: "-2px" },
      centreDraggedOnCursor: false,
    }}
    on:consider={handleConsider}
    on:finalize={handleFinalize}
  >
    {#each dndItems as item (item.id)}
      {#if typeof item.id === "number" && getElement(item.id)}
        <ElementBlock
          element={getElement(item.id)}
          children={loopElementTypes.has(getElement(item.id).element_type) ? childrenOf(item.id) : []}
          selected={$selectedElementId === item.id}
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
  </div>
</div>

<style>
  .canvas {
    height: 100%;
    padding: 16px;
    overflow-y: auto;
    position: relative;
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
    min-height: 100%;
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
</style>
