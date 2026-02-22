import { writable } from "svelte/store";

export const nativeDraggedElementType = writable<string | null>(null);
