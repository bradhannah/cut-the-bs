<script lang="ts">
  // Palette.svelte — T029
  // Draggable element types organized by category. Uses svelte-dnd-action
  // with copy behavior so palette items are never consumed.
  import { dndzone, SHADOW_ITEM_MARKER_PROPERTY_NAME } from "svelte-dnd-action";
  import { currentTemplate, canvasElements } from "../../stores/templateBuilder";
  import { elementLabels, elementIcons, repeatableElementTypes } from "./elementTypes";

  // Element type metadata for display in the palette.
  interface PaletteItem {
    id: string; // element_type used as id for DnD
    element_type: string;
    label: string;
    icon: string;
    isDndShadowItem?: boolean;
  }

  interface PaletteCategory {
    name: string;
    items: PaletteItem[];
  }

  function makeItem(elementType: string): PaletteItem {
    return {
      id: elementType,
      element_type: elementType,
      label: elementLabels[elementType] || elementType,
      icon: elementIcons[elementType] || "?",
    };
  }

  // Original (immutable) category definitions — used to restore items after drag.
  const resumeCategoryDefs: PaletteCategory[] = [
    {
      name: "Data",
      items: [
        makeItem("profile_header"),
        makeItem("role_descriptors"),
        makeItem("professional_summary"),
        makeItem("skills"),
        makeItem("core_expertise"),
      ],
    },
    {
      name: "Containers",
      items: [
        makeItem("work_history_loop"),
        makeItem("education_loop"),
        makeItem("certifications_loop"),
      ],
    },
    {
      name: "Formatting",
      items: [
        makeItem("section_heading"),
        makeItem("horizontal_rule"),
        makeItem("spacer"),
        makeItem("static_text"),
      ],
    },
  ];

  const coverLetterCategoryDefs: PaletteCategory[] = [
    {
      name: "Data",
      items: [
        makeItem("profile_header"),
        makeItem("recipient_address"),
        makeItem("date"),
        makeItem("greeting"),
        makeItem("body_text"),
        makeItem("closing"),
      ],
    },
    {
      name: "Formatting",
      items: [
        makeItem("section_heading"),
        makeItem("horizontal_rule"),
        makeItem("spacer"),
        makeItem("static_text"),
      ],
    },
  ];

  // Deep-clone category defs so svelte-dnd-action can mutate items arrays.
  function cloneCategories(defs: PaletteCategory[]): PaletteCategory[] {
    return defs.map((cat) => ({ ...cat, items: cat.items.map((it) => ({ ...it })) }));
  }

  // Mutable categories array that svelte-dnd-action will update via events.
  let categories: PaletteCategory[] = [];

  // Re-derive from template type changes.
  $: {
    const defs =
      $currentTemplate?.template_type === "cover_letter"
        ? coverLetterCategoryDefs
        : resumeCategoryDefs;
    categories = cloneCategories(defs);
  }

  // Compute the set of element types already present on the canvas.
  $: usedElementTypes = new Set(
    $canvasElements.map((el) => el.element_type)
  );

  // Check if an element type should appear subdued (already on canvas and not repeatable).
  function isUsed(elementType: string): boolean {
    return !repeatableElementTypes.has(elementType) && usedElementTypes.has(elementType);
  }

  function handleDndConsider(
    catIdx: number,
    e: CustomEvent<{ items: PaletteItem[] }>
  ): void {
    // svelte-dnd-action requires the items array to be updated during consider.
    categories[catIdx].items = e.detail.items;
    categories = categories; // trigger Svelte reactivity
  }

  function handleDndFinalize(
    catIdx: number,
    _e: CustomEvent<{ items: PaletteItem[] }>
  ): void {
    // Restore original items for this category (copy-from-source pattern).
    const defs =
      $currentTemplate?.template_type === "cover_letter"
        ? coverLetterCategoryDefs
        : resumeCategoryDefs;
    categories[catIdx].items = defs[catIdx].items.map((it) => ({ ...it }));
    categories = categories; // trigger Svelte reactivity
  }

  function transformDraggedElement(el: HTMLElement): void {
    el.classList.add("palette-dragging");
  }
</script>

<div class="palette">
  <div class="palette-header">
    <h3>Elements</h3>
  </div>

  {#each categories as category, catIdx (category.name)}
    <div class="palette-category">
      <div class="category-label">{category.name}</div>
      <div
        class="category-items"
        use:dndzone={{
          items: category.items,
          type: "template-element",
          dragDisabled: false,
          dropFromOthersDisabled: true,
          morphDisabled: true,
          centreDraggedOnCursor: true,
          transformDraggedElement,
        }}
        on:consider={(e) => handleDndConsider(catIdx, e)}
        on:finalize={(e) => handleDndFinalize(catIdx, e)}
      >
        {#each category.items as item (item.id)}
          <div
            class="palette-item"
            class:is-shadow={item[SHADOW_ITEM_MARKER_PROPERTY_NAME]}
            class:used={isUsed(item.element_type)}
          >
            <span class="item-icon">{item.icon}</span>
            <span class="item-label">{item.label}</span>
            {#if isUsed(item.element_type)}
              <span class="used-badge" title="Already on canvas">&#10003;</span>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/each}
</div>

<style>
  .palette {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .palette-header {
    padding: 12px 16px;
    border-bottom: 1px solid #2a3a4a;
  }

  .palette-header h3 {
    margin: 0;
    font-size: 0.95rem;
    color: #e0e0e0;
  }

  .palette-category {
    padding: 8px 0;
  }

  .category-label {
    padding: 4px 16px 6px;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #7a8a9a;
    font-weight: 600;
  }

  .category-items {
    min-height: 30px;
  }

  .palette-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 16px;
    cursor: grab;
    color: #c0d0e0;
    font-size: 0.82rem;
    transition: background-color 0.12s;
    user-select: none;
  }

  .palette-item:hover {
    background-color: #223344;
  }

  .palette-item:active {
    cursor: grabbing;
  }

  .palette-item.is-shadow {
    opacity: 0.4;
  }

  .palette-item.used {
    opacity: 0.4;
    cursor: grab;
  }

  .palette-item.used:hover {
    opacity: 0.55;
  }

  .used-badge {
    margin-left: auto;
    font-size: 0.65rem;
    color: #5a7a5a;
    flex-shrink: 0;
  }

  .item-icon {
    font-size: 0.9rem;
    width: 20px;
    text-align: center;
    flex-shrink: 0;
  }

  .item-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Dragging style applied via transformDraggedElement */
  :global(.palette-dragging) {
    background-color: #2a4060 !important;
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    opacity: 0.9;
  }
</style>
