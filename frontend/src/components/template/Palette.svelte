<script lang="ts">
  // Palette.svelte — T029
  // Draggable element types organized by category. Uses svelte-dnd-action
  // with copy behavior so palette items are never consumed.
  import { dndzone, SHADOW_ITEM_MARKER_PROPERTY_NAME } from "svelte-dnd-action";
  import { currentTemplate } from "../../stores/templateBuilder";
  import { elementLabels, elementIcons } from "./elementTypes";

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

  const resumeCategories: PaletteCategory[] = [
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

  const coverLetterCategories: PaletteCategory[] = [
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

  $: categories =
    $currentTemplate?.template_type === "cover_letter"
      ? coverLetterCategories
      : resumeCategories;

  function handleDndConsider(
    _catIdx: number,
    _e: CustomEvent<{ items: PaletteItem[] }>
  ): void {
    // Palette is the source — we don't mutate it on consider.
  }

  function handleDndFinalize(
    _catIdx: number,
    _e: CustomEvent<{ items: PaletteItem[] }>
  ): void {
    // Palette is read-only; finalize is a no-op.
  }

  function transformDraggedElement(el: HTMLElement): void {
    el.classList.add("palette-dragging");
  }
</script>

<div class="palette">
  <div class="palette-header">
    <h3>Elements</h3>
  </div>

  {#each categories as category (category.name)}
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
        on:consider={(e) => handleDndConsider(0, e)}
        on:finalize={(e) => handleDndFinalize(0, e)}
      >
        {#each category.items as item (item.id)}
          <div
            class="palette-item"
            class:is-shadow={item[SHADOW_ITEM_MARKER_PROPERTY_NAME]}
          >
            <span class="item-icon">{item.icon}</span>
            <span class="item-label">{item.label}</span>
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
