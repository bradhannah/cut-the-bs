<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let label = "Date";
  export let value = "";
  export let granularity = "month";
  export let allowPresent = false;
  export let isPresent = false;

  // Generate a unique ID for associating label with the date field
  const fieldId = `date-${label.toLowerCase().replace(/\s+/g, "-")}-${Math.random().toString(36).substring(2, 7)}`;

  const dispatch = createEventDispatcher<{
    change: { value: string; granularity: string };
  }>();

  const granularities = [
    { value: "year", label: "Year" },
    { value: "month", label: "Month" },
    { value: "day", label: "Day" },
  ];

  function getInputType(g: string): string {
    if (g === "year") return "number";
    if (g === "day") return "date";
    return "month";
  }

  function getPlaceholder(g: string): string {
    if (g === "year") return "YYYY";
    if (g === "day") return "YYYY-MM-DD";
    return "YYYY-MM";
  }

  // Normalize value when granularity changes so the input stays valid.
  function handleGranularityChange(
    event: Event & { currentTarget: HTMLSelectElement }
  ): void {
    granularity = event.currentTarget.value;
    // Truncate or clear value to fit new granularity
    if (granularity === "year" && value.length > 4) {
      value = value.substring(0, 4);
    } else if (granularity === "month" && value.length > 7) {
      value = value.substring(0, 7);
    }
    dispatch("change", { value, granularity });
  }

  function handleValueChange(
    event: Event & { currentTarget: HTMLInputElement }
  ): void {
    value = event.currentTarget.value;
    dispatch("change", { value, granularity });
  }

  function handlePresentToggle(
    event: Event & { currentTarget: HTMLInputElement }
  ): void {
    isPresent = event.currentTarget.checked;
    if (isPresent) {
      value = "";
    }
    dispatch("change", { value: isPresent ? "" : value, granularity });
  }
</script>

<div class="date-input">
  <label class="date-label" for={fieldId}>{label}</label>
  <div class="date-controls">
    <select
      id={fieldId}
      class="granularity-select"
      value={granularity}
      on:change={handleGranularityChange}
      disabled={allowPresent && isPresent}
    >
      {#each granularities as g (g.value)}
        <option value={g.value}>{g.label}</option>
      {/each}
    </select>
    {#if allowPresent && isPresent}
      <span class="present-label">Present</span>
    {:else if granularity === "year"}
      <input
        type="number"
        class="date-field"
        placeholder={getPlaceholder(granularity)}
        {value}
        min="1900"
        max="2100"
        on:input={handleValueChange}
      />
    {:else}
      <input
        type={getInputType(granularity)}
        class="date-field"
        placeholder={getPlaceholder(granularity)}
        {value}
        on:input={handleValueChange}
      />
    {/if}
    {#if allowPresent}
      <label class="present-check">
        <input
          type="checkbox"
          checked={isPresent}
          on:change={handlePresentToggle}
        />
        Present
      </label>
    {/if}
  </div>
</div>

<style>
  .date-input {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .date-label {
    font-size: 0.8rem;
    color: #7a8a9a;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .date-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .granularity-select {
    background-color: #1a2332;
    color: #e0e0e0;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 6px 8px;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .granularity-select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .date-field {
    background-color: #1a2332;
    color: #e0e0e0;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 6px 10px;
    font-size: 0.9rem;
    flex: 1;
    min-width: 0;
  }

  .date-field:focus {
    outline: none;
    border-color: #4a8af4;
  }

  .present-label {
    color: #4a8af4;
    font-size: 0.9rem;
    font-weight: 600;
    flex: 1;
    padding: 6px 0;
  }

  .present-check {
    display: flex;
    align-items: center;
    gap: 4px;
    color: #a0b0c0;
    font-size: 0.8rem;
    cursor: pointer;
    white-space: nowrap;
  }

  .present-check input[type="checkbox"] {
    accent-color: #4a8af4;
  }
</style>
