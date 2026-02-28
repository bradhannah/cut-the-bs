<script lang="ts">
  // ElementBlock.svelte — T031
  // Displays a single template element in the canvas. Shows type label,
  // icon, drag handle for reordering, delete button, and click-to-select.
  // For loop containers, renders a LoopContainer sub-component.
  import { createEventDispatcher } from "svelte";
  import type { TemplateElement } from "../../services/api";
  import { elementLabels, elementIcons, loopElementTypes } from "./elementTypes";
  import LoopContainer from "./LoopContainer.svelte";

  export let element: TemplateElement;
  export let children: TemplateElement[] = [];
  export let selected: boolean = false;
  export let readOnly: boolean = false;

  const dispatch = createEventDispatcher<{
    select: { id: number };
    delete: { id: number };
  }>();

  $: label = elementLabels[element.element_type] || element.element_type;
  $: icon = elementIcons[element.element_type] || "?";
  $: isLoop = loopElementTypes.has(element.element_type);

  // Extract a preview from the config (e.g., section heading text).
  $: configPreview = getConfigPreview(element);

  function getConfigPreview(el: TemplateElement): string {
    try {
      const cfg = JSON.parse(el.config);
      if (el.element_type === "section_heading" && cfg.text) return cfg.text;
      if (el.element_type === "static_text" && cfg.text) {
        return cfg.text.length > 40 ? cfg.text.substring(0, 40) + "..." : cfg.text;
      }
      if (el.element_type === "spacer" && cfg.height) return `${cfg.height}pt`;
      if (el.element_type === "horizontal_rule" && cfg.weight) return `${cfg.weight}pt`;
      return "";
    } catch {
      return "";
    }
  }

  function handleClick(e: MouseEvent): void {
    e.stopPropagation();
    dispatch("select", { id: element.id });
  }

  function handleMouseDown(e: MouseEvent): void {
    if (e.button !== 0) return;

    const target = e.target;
    if (target instanceof Element) {
      if (target.closest(".delete-btn")) return;
      if (target.closest(".loop-container")) return;
    }

    e.stopPropagation();
    dispatch("select", { id: element.id });
  }

  function handleDelete(e: MouseEvent): void {
    if (readOnly) return;
    e.stopPropagation();
    dispatch("delete", { id: element.id });
  }

  function handleKeydown(e: KeyboardEvent): void {
    if (readOnly) return;
    if (e.key === "Delete" || e.key === "Backspace") {
      if (selected) {
        dispatch("delete", { id: element.id });
      }
    }
  }
</script>

<div
  class="element-block"
  class:selected
  class:is-loop={isLoop}
  data-template-element-id={element.id}
  on:mousedown={handleMouseDown}
  on:click={handleClick}
  on:keydown={handleKeydown}
  tabindex="0"
  role="button"
  aria-label="Template element: {label}"
>
  <div class="element-header">
    <span class="drag-handle" class:readonly={readOnly} title="Drag to reorder">&#x2630;</span>
    <span class="element-icon">{icon}</span>
    <span class="element-label">{label}</span>
    {#if configPreview}
      <span class="config-preview">{configPreview}</span>
    {/if}
    {#if !readOnly}
      <button
        class="delete-btn"
        title="Delete element"
        on:click={handleDelete}
        aria-label="Delete {label}"
      >
        &times;
      </button>
    {/if}
  </div>

  {#if isLoop}
    <LoopContainer
      parentElement={element}
      {children}
      {readOnly}
      on:select
      on:delete
    />
  {/if}
</div>

<style>
  .element-block {
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    background-color: #1b2636;
    transition: border-color 0.12s, box-shadow 0.12s;
    outline: none;
  }

  .element-block:hover {
    border-color: #3a4a5a;
  }

  .element-block.selected {
    border-color: #4a8af4;
    box-shadow: 0 0 0 1px #4a8af4;
  }

  .element-block.is-loop {
    border-left: 3px solid #7a6af4;
  }

  .element-block.is-loop.selected {
    border-color: #4a8af4;
    border-left-color: #7a6af4;
  }

  .element-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
  }

  .drag-handle {
    cursor: grab;
    color: #5a6a7a;
    font-size: 0.9rem;
    user-select: none;
    flex-shrink: 0;
  }

  .drag-handle:active {
    cursor: grabbing;
  }

  .drag-handle.readonly {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .element-icon {
    font-size: 0.85rem;
    flex-shrink: 0;
  }

  .element-label {
    font-size: 0.82rem;
    color: #e0e0e0;
    font-weight: 500;
    white-space: nowrap;
  }

  .config-preview {
    font-size: 0.75rem;
    color: #7a8a9a;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
  }

  .delete-btn {
    margin-left: auto;
    background: none;
    border: none;
    color: #5a6a7a;
    font-size: 1.1rem;
    cursor: pointer;
    padding: 0 4px;
    line-height: 1;
    border-radius: 3px;
    transition: color 0.12s, background-color 0.12s;
    flex-shrink: 0;
  }

  .delete-btn:hover {
    color: #ff6b6b;
    background-color: rgba(255, 107, 107, 0.1);
  }
</style>
