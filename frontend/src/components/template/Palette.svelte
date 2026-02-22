<script lang="ts">
  // Palette.svelte — T029
  // Draggable element types organized by category.
  import { currentTemplate, canvasElements } from "../../stores/templateBuilder";
  import {
    elementLabels,
    elementIcons,
    repeatableElementTypes,
    loopElementTypes,
    loopSpecificChildren,
    loopCategoryNames,
  } from "./elementTypes";
  import { nativeDraggedElementType } from "./nativeDragState";

  // Element type metadata for display in the palette.
  interface PaletteItem {
    id: string; // element_type used as id for DnD
    element_type: string;
    label: string;
    icon: string;
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

  // Category definitions for the palette.
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

  let categories: PaletteCategory[] = [];

  // NOTE: $canvasElements must be referenced DIRECTLY here (not inside a called
  // function) so Svelte 3's static dependency tracker sees it and re-runs this
  // block whenever the store changes (e.g. after a loop container is added or
  // deleted, making its child-element category appear or disappear).
  $: {
    const baseCategories =
      $currentTemplate?.template_type === "cover_letter"
        ? coverLetterCategoryDefs
        : resumeCategoryDefs;

    if ($currentTemplate?.template_type !== "resume") {
      categories = baseCategories;
    } else {
      const activeLoopTypes = new Set(
        $canvasElements
          .map((el) => el.element_type)
          .filter((elementType) => loopElementTypes.has(elementType))
      );

      const loopOrder = ["work_history_loop", "education_loop", "certifications_loop"];
      const dynamicLoopCategories: PaletteCategory[] = [];

      for (const loopType of loopOrder) {
        if (!activeLoopTypes.has(loopType)) continue;
        const children = loopSpecificChildren[loopType] || [];
        if (children.length === 0) continue;
        dynamicLoopCategories.push({
          name: loopCategoryNames[loopType] || elementLabels[loopType] || loopType,
          items: children.map((elementType) => makeItem(elementType)),
        });
      }

      categories = [
        baseCategories[0],
        baseCategories[1],
        ...dynamicLoopCategories,
        baseCategories[2],
      ].filter(Boolean) as PaletteCategory[];
    }
  }

  // Compute the set of element types already present on the canvas.
  $: usedElementTypes = new Set(
    $canvasElements.map((el) => el.element_type)
  );

  // Check if an element type should appear subdued (already on canvas and not repeatable).
  function isUsed(elementType: string): boolean {
    return !repeatableElementTypes.has(elementType) && usedElementTypes.has(elementType);
  }

  function handleDragStart(e: DragEvent, item: PaletteItem): void {
    if (!e.dataTransfer) return;
    nativeDraggedElementType.set(item.element_type);
    e.dataTransfer.setData("application/x-template-element", item.element_type);
    e.dataTransfer.setData("text/plain", item.element_type);
    e.dataTransfer.effectAllowed = "copy";
    (e.currentTarget as HTMLElement | null)?.classList.add("palette-dragging");
  }

  function handleDragEnd(e: DragEvent): void {
    nativeDraggedElementType.set(null);
    (e.currentTarget as HTMLElement | null)?.classList.remove("palette-dragging");
  }
</script>

<div class="palette">
  <div class="palette-header">
    <h3>Elements</h3>
  </div>

  {#each categories as category (category.name)}
    <div class="palette-category">
      <div class="category-label">{category.name}</div>
      <div class="category-items">
        {#each category.items as item (item.id)}
          <div
            class="palette-item"
            class:used={isUsed(item.element_type)}
            draggable="true"
            on:dragstart={(e) => handleDragStart(e, item)}
            on:dragend={handleDragEnd}
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

  /* Dragging style applied in handleDragStart */
  :global(.palette-dragging) {
    background-color: #2a4060 !important;
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    opacity: 0.9;
  }
</style>
