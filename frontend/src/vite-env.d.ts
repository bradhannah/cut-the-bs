/// <reference types="svelte" />
/// <reference types="vite/client" />

// Extend HTML attributes to support svelte-dnd-action custom events.
// See: https://github.com/isaacHagworthy/svelte-dnd-action#typescript
declare namespace svelte.JSX {
  interface HTMLAttributes<T> {
    onconsider?: (event: CustomEvent<any>) => void;
    onfinalize?: (event: CustomEvent<any>) => void;
  }
}

declare namespace svelteHTML {
  interface HTMLAttributes<T> {
    "on:consider"?: (event: CustomEvent<any>) => void;
    "on:finalize"?: (event: CustomEvent<any>) => void;
  }
}

