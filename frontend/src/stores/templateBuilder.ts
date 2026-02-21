// Template builder Svelte stores — scoped state management for the
// three-panel builder UI. This is a justified exception to the project's
// "no Svelte stores" convention; the builder needs shared state across
// Palette, Canvas, and Properties panel components.

import { writable, derived } from "svelte/store";
import type {
  TemplateElement,
  TemplateDetail,
  DocumentTemplate,
} from "../services/api";

// --- Canvas Elements ---

/** Flat, ordered list of all elements in the current template. */
export const canvasElements = writable<TemplateElement[]>([]);

// --- Selection ---

/** ID of the currently selected element, or null if none. */
export const selectedElementId = writable<number | null>(null);

/** Derived: the full TemplateElement object for the current selection. */
export const selectedElement = derived(
  [canvasElements, selectedElementId],
  ([$elements, $id]) => {
    if ($id === null) return null;
    return $elements.find((el) => el.id === $id) ?? null;
  }
);

// --- Current Template Metadata ---

/** The template metadata (without elements) for the currently open template. */
export const currentTemplate = writable<DocumentTemplate | null>(null);

// --- Helpers ---

/** Reset all builder stores to their initial state. */
export function resetBuilderStores(): void {
  canvasElements.set([]);
  selectedElementId.set(null);
  currentTemplate.set(null);
}

/**
 * Populate builder stores from a loaded TemplateDetail.
 * Call this after GetDocumentTemplate returns.
 */
export function loadTemplateIntoStores(detail: TemplateDetail): void {
  const { elements, ...meta } = detail;
  currentTemplate.set(meta);
  canvasElements.set(elements);
  selectedElementId.set(null);
}
