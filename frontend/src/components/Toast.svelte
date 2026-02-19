<script lang="ts">
  import { onDestroy } from "svelte";
  import {
    subscribeToasts,
    removeToast,
    type ToastMessage,
  } from "../services/api";

  let toasts: ToastMessage[] = [];

  const unsubscribe = subscribeToasts((t) => {
    toasts = t;
  });

  onDestroy(unsubscribe);
</script>

{#if toasts.length > 0}
  <div class="toast-container">
    {#each toasts as toast (toast.id)}
      <div class="toast toast-{toast.level}">
        <span class="toast-message">{toast.message}</span>
        <button class="toast-close" on:click={() => removeToast(toast.id)}>
          x
        </button>
      </div>
    {/each}
  </div>
{/if}

<style>
  .toast-container {
    position: fixed;
    bottom: 16px;
    right: 16px;
    z-index: 1000;
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: 400px;
  }

  .toast {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    border-radius: 6px;
    font-size: 0.9rem;
    color: #fff;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
    animation: slide-in 0.2s ease-out;
  }

  .toast-info {
    background-color: #2a5080;
  }

  .toast-success {
    background-color: #2a6040;
  }

  .toast-error {
    background-color: #802a2a;
  }

  .toast-message {
    flex: 1;
    margin-right: 12px;
  }

  .toast-close {
    background: none;
    border: none;
    color: rgba(255, 255, 255, 0.7);
    cursor: pointer;
    font-size: 0.85rem;
    padding: 0 4px;
    line-height: 1;
  }

  .toast-close:hover {
    color: #fff;
  }

  @keyframes slide-in {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
