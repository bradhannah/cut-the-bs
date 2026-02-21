<script lang="ts">
  export let message = "Loading...";
</script>

<div class="loading-state">
  <div class="spinner" />
  <p class="loading-text">{message}</p>
</div>

<style>
  .loading-state {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 16px 0;
  }

  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid #2a3a4a;
    border-top-color: #4a8af4;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
    flex-shrink: 0;
  }

  .loading-text {
    color: #5a6a7a;
    font-style: italic;
    font-size: 0.9rem;
    margin: 0;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
