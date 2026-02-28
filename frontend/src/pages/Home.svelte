<script lang="ts">
  import { onMount } from "svelte";
  import {
    addToast,
    listApplications,
    openFile,
    prepareApplicationUploadFolder,
    type JobApplication,
  } from "../services/api";
  import { formatDate } from "../services/dateFormat";

  const inactiveStatuses = new Set([
    "Offer Accepted",
    "Offer Declined",
    "Employer Rejected",
    "User Withdrawn",
    "Ghosted",
  ]);

  let loading = true;
  let recentApplications: JobApplication[] = [];
  let openingFolderID: number | null = null;

  onMount(async () => {
    await loadRecentApplications();
  });

  async function loadRecentApplications(): Promise<void> {
    loading = true;
    try {
      const all = (await listApplications()) || [];
      recentApplications = all
        .filter((app) => !inactiveStatuses.has(app.status))
        .sort((a, b) => Date.parse(b.updated_at) - Date.parse(a.updated_at))
        .slice(0, 6);
    } catch {
      // Toast already shown by API layer
      recentApplications = [];
    } finally {
      loading = false;
    }
  }

  async function openUploadFolder(app: JobApplication): Promise<void> {
    if (!app.resume_export_id && !app.cover_letter_latest_export_id) {
      addToast("info", "No generated documents linked for this application");
      return;
    }

    openingFolderID = app.id;
    try {
      const folderPath = await prepareApplicationUploadFolder(app.id);
      await openFile(folderPath);
    } catch {
      // Toast already shown
    } finally {
      if (openingFolderID === app.id) {
        openingFolderID = null;
      }
    }
  }
</script>

<div class="home-page">
  <section class="hero-card">
    <p class="eyebrow">Welcome</p>
    <h2>Cut the Bullsh*t</h2>
    <p class="hero-description">
      Build clean, targeted resumes and cover letters fast. Keep your source data in one place,
      generate for each role, and open upload-ready files without digging through folders. Use
      Application Helper for fast field-by-field copy when filling job portals.
    </p>

    <div class="hero-actions">
      <a class="action-link action-primary" href="#/applications">Open Applications</a>
      <a class="action-link" href="#/helper">Open Application Helper</a>
      <a class="action-link" href="#/templates">Open Templates</a>
      <a class="action-link" href="#/export">Open Export</a>
      <a class="action-link" href="#/work-history">Update Profile Data</a>
    </div>
  </section>

  <section class="recent-card">
    <div class="section-header">
      <h3>Recent Active Applications</h3>
      <a class="view-all" href="#/applications">View all</a>
    </div>

    {#if loading}
      <p class="hint">Loading recent applications...</p>
    {:else if recentApplications.length === 0}
      <p class="hint">
        No active applications yet. Start in <a href="#/applications">Applications</a>.
      </p>
    {:else}
      <div class="recent-list">
        {#each recentApplications as app (app.id)}
          <div class="recent-item">
            <div class="recent-main">
              <p class="recent-company">{app.company_name}</p>
              <p class="recent-role">{app.position_title}</p>
              <p class="recent-meta">
                {app.status} · Updated {formatDate(app.updated_at.slice(0, 10), "day")}
              </p>
            </div>
            <div class="recent-actions">
              <button
                class="mini-btn"
                on:click={() => openUploadFolder(app)}
                disabled={openingFolderID === app.id}
              >
                {openingFolderID === app.id ? "Opening..." : "Open Folder"}
              </button>
              <a class="mini-link" href={`#/applications?app=${app.id}`}>Open App</a>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </section>
</div>

<style>
  .home-page {
    max-width: 980px;
    display: grid;
    gap: 18px;
  }

  .hero-card,
  .recent-card {
    background: linear-gradient(160deg, #1c2b3b 0%, #182535 100%);
    border: 1px solid #2e4155;
    border-radius: 12px;
    padding: 18px;
  }

  .eyebrow {
    margin: 0;
    color: #7a8a9a;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    font-size: 0.7rem;
    font-weight: 700;
  }

  h2 {
    margin: 6px 0 10px;
    color: #e7eef7;
    font-size: 1.65rem;
    line-height: 1.2;
  }

  .hero-description {
    margin: 0;
    color: #98acbf;
    line-height: 1.5;
    max-width: 760px;
  }

  .hero-actions {
    margin-top: 16px;
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .action-link {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 38px;
    padding: 0 12px;
    border: 1px solid #38516a;
    border-radius: 7px;
    color: #b4c6d8;
    text-decoration: none;
    font-size: 0.82rem;
    font-weight: 600;
    transition: background-color 0.15s, border-color 0.15s, color 0.15s;
  }

  .action-link:hover {
    background: #24374b;
    color: #eaf2fb;
    border-color: #49698a;
  }

  .action-primary {
    background: #2b4d72;
    border-color: #416b98;
    color: #edf4fc;
  }

  .action-primary:hover {
    background: #355d88;
    border-color: #5280b2;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 10px;
  }

  h3 {
    margin: 0;
    color: #d7e3ef;
    font-size: 1rem;
  }

  .view-all {
    color: #8eb4de;
    text-decoration: none;
    font-size: 0.8rem;
  }

  .view-all:hover {
    color: #c3dbf4;
    text-decoration: underline;
  }

  .hint {
    margin: 8px 0 0;
    color: #7f93a7;
    font-size: 0.86rem;
  }

  .hint a {
    color: #8eb4de;
  }

  .recent-list {
    display: grid;
    gap: 10px;
  }

  .recent-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
    background: #172333;
    border: 1px solid #2a3d51;
    border-radius: 8px;
    padding: 10px 12px;
  }

  .recent-main {
    min-width: 0;
  }

  .recent-company {
    margin: 0;
    color: #dbe8f5;
    font-size: 0.9rem;
    font-weight: 700;
  }

  .recent-role {
    margin: 2px 0 0;
    color: #9fb3c7;
    font-size: 0.82rem;
  }

  .recent-meta {
    margin: 4px 0 0;
    color: #6f8498;
    font-size: 0.76rem;
  }

  .recent-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .mini-btn,
  .mini-link {
    min-height: 32px;
    border-radius: 6px;
    font-size: 0.76rem;
    font-weight: 600;
  }

  .mini-btn {
    border: 1px solid #3a556f;
    background: transparent;
    color: #b4c6d8;
    cursor: pointer;
    padding: 0 10px;
  }

  .mini-btn:hover:not(:disabled) {
    background: #24374b;
    color: #eaf2fb;
  }

  .mini-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .mini-link {
    display: inline-flex;
    align-items: center;
    text-decoration: none;
    border: 1px solid #416b98;
    background: #2b4d72;
    color: #edf4fc;
    padding: 0 10px;
  }

  .mini-link:hover {
    background: #355d88;
  }

  @media (max-width: 760px) {
    .recent-item {
      flex-direction: column;
      align-items: flex-start;
    }

    .recent-actions {
      width: 100%;
      justify-content: flex-start;
    }
  }
</style>
