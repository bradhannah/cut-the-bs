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

  export let parentElement: TemplateElement;
  export let children: TemplateElement[] = [];

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

  $: {
    dndItems = (children || [])
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

  function handleChildDelete(e: MouseEvent, childId: number): void {
    e.stopPropagation();
    dispatch("delete", { id: childId });
  }

  async function handleConsider(
    e: CustomEvent<{ items: ChildDndItem[] }>
  ): Promise<void> {
    const items = e.detail.items;
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
    e: CustomEvent<{ items: ChildDndItem[] }>
  ): Promise<void> {
    invalidDragOver = false;
    const items = e.detail.items;
    const templateId = $currentTemplate?.id;
    if (!templateId) return;

    const newItem = items.find((item) => typeof item.id === "string");

    if (newItem) {
      const elementType = newItem.element_type || (newItem.id as string);

      // Validate the child type is allowed for this loop.
      if (!validChildren.has(elementType)) {
        addToast("error", `${elementLabels[elementType] || elementType} cannot be added to ${elementLabels[parentElement.element_type]}`);
        // Reset dndItems to current children.
        dndItems = (children || [])
          .sort((a, b) => a.sort_order - b.sort_order)
          .map((el) => ({ id: el.id, element_type: el.element_type }));
        return;
      }

      const config = defaultConfigs[elementType] || "{}";

      try {
        markSaving();
        const created = await createTemplateElement(templateId, {
          parent_id: parentElement.id,
          element_type: elementType,
          config: config,
        });

        canvasElements.update((els) => [...els, created]);

        const orderedIds = items.map((item) =>
          typeof item.id === "string" ? created.id : (item.id as number)
        );

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
        markSaveError();
        addToast("error", err?.message || "Failed to add child element");
      }
    } else {
      const orderedIds = items.map((item) => item.id as number);

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
        markSaveError();
        addToast("error", err?.message || "Failed to reorder child elements");
      }
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
        centreDraggedOnCursor: false,
      }}
      on:consider={handleConsider}
      on:finalize={handleFinalize}
    >
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
        centreDraggedOnCursor: false,
      }}
      on:consider={handleConsider}
      on:finalize={handleFinalize}
    >
      {#each dndItems as child (child.id)}
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <div
          class="child-element"
          class:selected={$selectedElementId === child.id}
          on:click={(e) => handleChildClick(e, Number(child.id))}
          role="button"
          tabindex="0"
        >
          <span class="drag-handle">&#x2630;</span>
          <span class="child-icon">{elementIcons[child.element_type] || "?"}</span>
          <span class="child-label">{elementLabels[child.element_type] || child.element_type}</span>
          <button
            class="delete-btn"
            title="Delete child element"
            on:click={(e) => handleChildDelete(e, Number(child.id))}
          >
            &times;
          </button>
        </div>
      {/each}
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
</style>
