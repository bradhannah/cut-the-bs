<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { WorkHistoryEntry } from "../services/api";
  import BulletList from "./BulletList.svelte";

  export let entry: WorkHistoryEntry;
  export let expanded = false;

  const dispatch = createEventDispatcher<{
    edit: { entry: WorkHistoryEntry };
    delete: { id: number };
    bulletCreate: { workHistoryId: number; text: string };
    bulletUpdate: { id: number; text: string };
    bulletDelete: { id: number };
    bulletReorder: { workHistoryId: number; orderedIDs: number[] };
    bulletPaste: { workHistoryId: number };
  }>();

  function toggle(): void {
    expanded = !expanded;
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      toggle();
    }
  }

  function formatDate(dateStr: string, granularity: string): string {
    if (!dateStr) return "Present";
    if (granularity === "year") return dateStr.substring(0, 4);
    if (granularity === "day") {
      // YYYY-MM-DD → render as-is
      return dateStr;
    }
    // month: YYYY-MM → render as "Mon YYYY"
    const parts = dateStr.split("-");
    if (parts.length >= 2) {
      const monthNames = [
        "Jan",
        "Feb",
        "Mar",
        "Apr",
        "May",
        "Jun",
        "Jul",
        "Aug",
        "Sep",
        "Oct",
        "Nov",
        "Dec",
      ];
      const monthIdx = parseInt(parts[1], 10) - 1;
      if (monthIdx >= 0 && monthIdx < 12) {
        return `${monthNames[monthIdx]} ${parts[0]}`;
      }
    }
    return dateStr;
  }

  function handleEdit(): void {
    dispatch("edit", { entry });
  }

  function handleDelete(): void {
    dispatch("delete", { id: entry.id });
  }

  function onBulletCreate(
    event: CustomEvent<{ workHistoryId: number; text: string }>
  ): void {
    dispatch("bulletCreate", event.detail);
  }

  function onBulletUpdate(
    event: CustomEvent<{ id: number; text: string }>
  ): void {
    dispatch("bulletUpdate", event.detail);
  }

  function onBulletDelete(event: CustomEvent<{ id: number }>): void {
    dispatch("bulletDelete", event.detail);
  }

  function onBulletReorder(
    event: CustomEvent<{ workHistoryId: number; orderedIDs: number[] }>
  ): void {
    dispatch("bulletReorder", event.detail);
  }

  function onBulletPaste(event: CustomEvent<{ workHistoryId: number }>): void {
    dispatch("bulletPaste", event.detail);
  }

  $: startFormatted = formatDate(
    entry.start_date,
    entry.date_granularity_start
  );
  $: endFormatted = formatDate(entry.end_date, entry.date_granularity_end);
  $: dateRange = `${startFormatted} - ${endFormatted}`;
  $: bulletCount = entry.bullets ? entry.bullets.length : 0;
</script>

<div class="card" class:expanded>
  <div
    class="card-header"
    on:click={toggle}
    on:keydown={handleKeydown}
    role="button"
    tabindex="0"
    aria-expanded={expanded}
  >
    <div class="card-expand-icon">
      <span class="chevron">{expanded ? "\u25BC" : "\u25B6"}</span>
    </div>
    <div class="card-info">
      <div class="card-title">
        <strong>{entry.job_title}</strong>
        <span class="card-employer">at {entry.employer_name}</span>
      </div>
      <div class="card-meta">
        <span class="card-dates">{dateRange}</span>
        <span class="card-bullet-count">
          {bulletCount} bullet{bulletCount !== 1 ? "s" : ""}
        </span>
      </div>
    </div>
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <div class="card-actions" on:click|stopPropagation={() => {}}>
      <button class="btn-icon" on:click={handleEdit} title="Edit entry">
        &#9998;
      </button>
      <button
        class="btn-icon btn-icon-danger"
        on:click={handleDelete}
        title="Delete entry"
      >
        &#10005;
      </button>
    </div>
  </div>

  {#if expanded}
    <div class="card-body">
      <BulletList
        bullets={entry.bullets || []}
        workHistoryId={entry.id}
        on:create={onBulletCreate}
        on:update={onBulletUpdate}
        on:delete={onBulletDelete}
        on:reorder={onBulletReorder}
        on:paste={onBulletPaste}
      />
    </div>
  {/if}
</div>

<style>
  .card {
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    margin-bottom: 8px;
    transition: border-color 0.15s;
  }

  .card:hover {
    border-color: #3a4a5a;
  }

  .card.expanded {
    border-color: #4a8af4;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    cursor: pointer;
    user-select: none;
  }

  .card-header:focus {
    outline: none;
    box-shadow: inset 0 0 0 2px rgba(74, 138, 244, 0.3);
    border-radius: 6px;
  }

  .card-expand-icon {
    flex-shrink: 0;
    width: 16px;
    text-align: center;
  }

  .chevron {
    color: #5a6a7a;
    font-size: 0.7rem;
    transition: color 0.15s;
  }

  .card.expanded .chevron {
    color: #4a8af4;
  }

  .card-info {
    flex: 1;
    min-width: 0;
  }

  .card-title {
    font-size: 0.95rem;
    color: #e0e0e0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .card-employer {
    color: #7a8a9a;
    font-weight: 400;
  }

  .card-meta {
    display: flex;
    gap: 16px;
    margin-top: 2px;
    font-size: 0.8rem;
    color: #5a6a7a;
  }

  .card-dates {
    color: #7a8a9a;
  }

  .card-bullet-count {
    color: #5a6a7a;
  }

  .card-actions {
    display: flex;
    gap: 2px;
    flex-shrink: 0;
    opacity: 0;
    transition: opacity 0.15s;
  }

  .card-header:hover .card-actions {
    opacity: 1;
  }

  .btn-icon {
    background: none;
    border: none;
    color: #7a8a9a;
    cursor: pointer;
    padding: 4px 8px;
    font-size: 0.85rem;
    border-radius: 3px;
  }

  .btn-icon:hover {
    background-color: #2a3a4a;
    color: #e0e0e0;
  }

  .btn-icon-danger:hover {
    background-color: #5a2020;
    color: #ff6b6b;
  }

  .card-body {
    padding: 0 14px 14px 40px;
    border-top: 1px solid #2a3a4a;
  }
</style>
