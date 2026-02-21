<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { EventsOn } from "../../wailsjs/runtime/runtime";

  // Autosave indicator state
  let autosaveVisible = false;
  let autosaveTimer: ReturnType<typeof setTimeout> | null = null;

  // Backup notification state
  let backupMessage = "";
  let backupLevel: "success" | "error" | "" = "";
  let backupVisible = false;
  let backupTimer: ReturnType<typeof setTimeout> | null = null;

  // Cleanup functions returned by EventsOn
  let cleanups: (() => void)[] = [];

  function showAutosave() {
    autosaveVisible = true;
    if (autosaveTimer) clearTimeout(autosaveTimer);
    autosaveTimer = setTimeout(() => {
      autosaveVisible = false;
      autosaveTimer = null;
    }, 2000);
  }

  function showBackup(message: string, level: "success" | "error") {
    backupMessage = message;
    backupLevel = level;
    backupVisible = true;
    if (backupTimer) clearTimeout(backupTimer);
    backupTimer = setTimeout(() => {
      backupVisible = false;
      backupTimer = null;
    }, 4000);
  }

  onMount(() => {
    cleanups.push(
      EventsOn("autosave:complete", () => {
        showAutosave();
      })
    );

    cleanups.push(
      EventsOn("backup:complete", () => {
        showBackup("Backup created", "success");
      })
    );

    cleanups.push(
      EventsOn("backup:error", (data: { error: string }) => {
        showBackup("Backup failed: " + data.error, "error");
      })
    );
  });

  onDestroy(() => {
    for (const cleanup of cleanups) {
      cleanup();
    }
    if (autosaveTimer) clearTimeout(autosaveTimer);
    if (backupTimer) clearTimeout(backupTimer);
  });
</script>

<div class="status-bar">
  <div class="status-left">
    {#if autosaveVisible}
      <span class="autosave-indicator">Saved</span>
    {/if}
  </div>
  <div class="status-right">
    {#if backupVisible}
      <span class="backup-indicator backup-{backupLevel}">
        {backupMessage}
      </span>
    {/if}
  </div>
</div>

<style>
  .status-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 24px;
    padding: 0 12px;
    background-color: #151e2a;
    border-top: 1px solid #2a3a4a;
    font-size: 0.75rem;
    color: #5a6a7a;
    flex-shrink: 0;
  }

  .status-left,
  .status-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .autosave-indicator {
    color: #4a8a4a;
    animation: fade-in 0.15s ease-out;
  }

  .backup-indicator {
    animation: fade-in 0.15s ease-out;
  }

  .backup-success {
    color: #4a8a4a;
  }

  .backup-error {
    color: #c05050;
  }

  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
</style>
