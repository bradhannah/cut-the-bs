<script lang="ts">
  import { onMount, onDestroy } from "svelte";

  const STORAGE_KEY = "cut-the-bs-zoom";
  const MIN_ZOOM = 50;
  const MAX_ZOOM = 200;
  const STEP = 10;

  let zoom = 100;

  function loadZoom(): number {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const val = parseInt(stored, 10);
        if (!isNaN(val) && val >= MIN_ZOOM && val <= MAX_ZOOM) {
          return val;
        }
      }
    } catch {
      // localStorage unavailable
    }
    return 100;
  }

  function saveZoom(val: number) {
    try {
      localStorage.setItem(STORAGE_KEY, String(val));
    } catch {
      // localStorage unavailable
    }
  }

  function applyZoom(val: number) {
    const content = document.querySelector(".content");
    if (content instanceof HTMLElement) {
      // Use CSS zoom instead of transform:scale() so the browser
      // re-renders at the target resolution — keeps fonts sharp.
      (content.style as any).zoom = `${val / 100}`;
    }
  }

  function setZoom(val: number) {
    zoom = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, val));
    saveZoom(zoom);
    applyZoom(zoom);
  }

  function zoomIn() {
    setZoom(zoom + STEP);
  }

  function zoomOut() {
    setZoom(zoom - STEP);
  }

  function resetZoom() {
    setZoom(100);
  }

  function handleKeydown(e: KeyboardEvent) {
    const mod = e.metaKey || e.ctrlKey;
    if (!mod) return;

    if (e.key === "=" || e.key === "+") {
      e.preventDefault();
      zoomIn();
    } else if (e.key === "-") {
      e.preventDefault();
      zoomOut();
    } else if (e.key === "0") {
      e.preventDefault();
      resetZoom();
    }
  }

  onMount(() => {
    zoom = loadZoom();
    applyZoom(zoom);
    window.addEventListener("keydown", handleKeydown);
  });

  onDestroy(() => {
    window.removeEventListener("keydown", handleKeydown);
  });
</script>

<div class="zoom-widget">
  <button
    class="zoom-btn"
    on:click={zoomOut}
    disabled={zoom <= MIN_ZOOM}
    title="Zoom out (Cmd/Ctrl -)"
  >
    -
  </button>
  <button class="zoom-level" on:click={resetZoom} title="Reset zoom (Cmd/Ctrl 0)">
    {zoom}%
  </button>
  <button
    class="zoom-btn"
    on:click={zoomIn}
    disabled={zoom >= MAX_ZOOM}
    title="Zoom in (Cmd/Ctrl +)"
  >
    +
  </button>
</div>

<style>
  .zoom-widget {
    width: 100%;
    display: grid;
    grid-template-columns: 34px 1fr 34px;
    align-items: center;
    gap: 0;
    background-color: #111a26;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    overflow: hidden;
  }

  .zoom-btn {
    background: none;
    border: none;
    color: #7a8a9a;
    cursor: pointer;
    font-size: 0.92rem;
    width: 34px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    line-height: 1;
  }

  .zoom-btn:hover:not(:disabled) {
    background-color: #2a3a4a;
    color: #e0e0e0;
  }

  .zoom-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .zoom-level {
    background: none;
    border: none;
    color: #a0b0c0;
    cursor: pointer;
    font-size: 0.74rem;
    padding: 0 8px;
    min-width: 48px;
    min-height: 30px;
    text-align: center;
    border-left: 1px solid #2a3a4a;
    border-right: 1px solid #2a3a4a;
    font-weight: 700;
  }

  .zoom-level:hover {
    background-color: #2a3a4a;
    color: #e0e0e0;
  }
</style>
