<script lang="ts">
  import { location } from "svelte-spa-router";

  // Derive page name from current route
  $: pageName = $location
    .replace("/", "")
    .replace(/-/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
</script>

<div class="placeholder">
  <h2>{pageName || "Home"}</h2>
  <p>This section is not yet implemented.</p>
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
