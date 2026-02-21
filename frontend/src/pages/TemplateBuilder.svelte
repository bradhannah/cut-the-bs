<script lang="ts">
  // Stub — Phase 4 (T028) will implement the full three-panel builder.
  export let params: { id?: string } = {};

  $: templateId = params.id ? parseInt(params.id, 10) : null;
</script>

<div class="placeholder">
  <h2>Template Builder</h2>
  {#if templateId}
    <p>Editing template #{templateId} — builder UI coming soon.</p>
  {:else}
    <p>No template selected.</p>
  {/if}
</div>

<style>
  .placeholder {
    text-align: center;
    padding: 60px 20px;
  }

  .placeholder h2 {
    margin: 0 0 12px;
    font-size: 1.5rem;
    color: #e0e0e0;
  }

  .placeholder p {
    color: #7a8a9a;
    font-size: 0.95rem;
  }
</style>
