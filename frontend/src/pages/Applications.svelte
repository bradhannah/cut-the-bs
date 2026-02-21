<script lang="ts">
  import { onMount } from "svelte";
  import {
    listApplications,
    searchApplications,
    createApplication,
    updateApplication,
    updateApplicationStatus,
    updateApplicationFit,
    deleteApplication,
    getApplicationHistory,
    getApplicationStatuses,
    getFitIndicators,
    listCoverLetters,
    listExports,
    addToast,
    type JobApplication,
    type ApplicationInput,
    type StatusChange,
    type CoverLetter,
    type ResumeExport,
  } from "../services/api";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";
  import { formatDate, formatTimestamp } from "../services/dateFormat";

  let applications: JobApplication[] = [];
  let loading = true;
  let searchQuery = "";
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;

  // Reference data
  let statuses: string[] = [];
  let fitIndicators: string[] = [];
  let coverLetters: CoverLetter[] = [];
  let exports: ResumeExport[] = [];

  // Form state
  let showForm = false;
  let editingApp: JobApplication | null = null;
  let formCompanyName = "";
  let formPositionTitle = "";
  let formDateApplied = "";
  let formFitIndicator = "";
  let formResumeExportId: number | null = null;
  let formCoverLetterId: number | null = null;
  let formNotes = "";

  // Status change state
  let changingStatusFor: number | null = null;
  let newStatusValue = "";

  // History state
  let historyFor: number | null = null;
  let historyEntries: StatusChange[] = [];

  onMount(async () => {
    await Promise.all([loadApplications(), loadReferenceData()]);
  });

  async function loadReferenceData(): Promise<void> {
    try {
      [statuses, fitIndicators, coverLetters, exports] = await Promise.all([
        getApplicationStatuses(),
        getFitIndicators(),
        listCoverLetters(),
        listExports(),
      ]);
      statuses = statuses || [];
      fitIndicators = fitIndicators || [];
      coverLetters = coverLetters || [];
      exports = exports || [];
    } catch {
      // Toast already shown
    }
  }

  async function loadApplications(): Promise<void> {
    loading = true;
    try {
      if (searchQuery.trim()) {
        applications = (await searchApplications(searchQuery.trim())) || [];
      } else {
        applications = (await listApplications()) || [];
      }
    } finally {
      loading = false;
    }
  }

  function handleSearchInput(): void {
    if (searchTimeout) clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      loadApplications();
    }, 300);
  }

  function openAddForm(): void {
    editingApp = null;
    formCompanyName = "";
    formPositionTitle = "";
    formDateApplied = new Date().toISOString().slice(0, 10);
    formFitIndicator = "";
    formResumeExportId = null;
    formCoverLetterId = null;
    formNotes = "";
    showForm = true;
  }

  function openEditForm(app: JobApplication): void {
    editingApp = app;
    formCompanyName = app.company_name;
    formPositionTitle = app.position_title;
    formDateApplied = app.date_applied;
    formFitIndicator = app.fit_indicator;
    formResumeExportId = app.resume_export_id;
    formCoverLetterId = app.cover_letter_id;
    formNotes = app.notes;
    showForm = true;
  }

  function cancelForm(): void {
    showForm = false;
    editingApp = null;
  }

  async function handleSubmit(): Promise<void> {
    const input: ApplicationInput = {
      company_name: formCompanyName.trim(),
      position_title: formPositionTitle.trim(),
      date_applied: formDateApplied,
      fit_indicator: formFitIndicator,
      resume_export_id: formResumeExportId,
      cover_letter_id: formCoverLetterId,
      notes: formNotes.trim(),
    };

    if (!input.company_name) {
      addToast("error", "Company name is required");
      return;
    }
    if (!input.position_title) {
      addToast("error", "Position title is required");
      return;
    }
    if (!input.date_applied) {
      addToast("error", "Date applied is required");
      return;
    }

    try {
      if (editingApp) {
        await updateApplication(editingApp.id, input);
      } else {
        await createApplication(input);
      }
      showForm = false;
      editingApp = null;
      await loadApplications();
    } catch {
      // Toast already shown
    }
  }

  async function handleDelete(id: number): Promise<void> {
    try {
      await deleteApplication(id);
      if (historyFor === id) historyFor = null;
      await loadApplications();
    } catch {
      // Toast already shown
    }
  }

  function openStatusChange(app: JobApplication): void {
    changingStatusFor = app.id;
    newStatusValue = app.status;
  }

  async function handleStatusChange(id: number): Promise<void> {
    try {
      await updateApplicationStatus(id, newStatusValue);
      changingStatusFor = null;
      await loadApplications();
    } catch {
      // Toast already shown
    }
  }

  async function handleFitChange(id: number, fit: string): Promise<void> {
    try {
      await updateApplicationFit(id, fit);
      await loadApplications();
    } catch {
      // Toast already shown
    }
  }

  async function toggleHistory(id: number): Promise<void> {
    if (historyFor === id) {
      historyFor = null;
      return;
    }
    try {
      historyEntries = (await getApplicationHistory(id)) || [];
      historyFor = id;
    } catch {
      // Toast already shown
    }
  }

  function statusColor(status: string): string {
    const colors: Record<string, string> = {
      Applied: "#4a8af4",
      Acknowledged: "#5a9af4",
      Screening: "#6ab0f4",
      "Phone Screen": "#4ac0a0",
      "Interview Scheduled": "#50c878",
      "Interview Completed": "#60d888",
      "Technical Assessment": "#f0a040",
      "Final Round": "#e0c040",
      "Offer Received": "#40d060",
      "Offer Accepted": "#30c050",
      "Offer Declined": "#c0a040",
      "Employer Rejected": "#c05050",
      "User Withdrawn": "#a06060",
      Ghosted: "#706070",
      "On Hold": "#8080a0",
    };
    return colors[status] || "#7a8a9a";
  }

  function fitColor(fit: string): string {
    const colors: Record<string, string> = {
      Unlikely: "#c05050",
      "Stretch Fit": "#c0a040",
      "Possible Fit": "#e0c040",
      "Strong Fit": "#50c878",
      "Perfect Fit": "#30c050",
    };
    return colors[fit] || "#7a8a9a";
  }

  function getCoverLetterTitle(id: number | null): string {
    if (!id) return "";
    const cl = coverLetters.find((c) => c.id === id);
    return cl ? cl.title : `#${id}`;
  }

  function getExportLabel(id: number | null): string {
    if (!id) return "";
    const ex = exports.find((e) => e.id === id);
    return ex
      ? `${ex.template_id} (${formatDate(ex.generated_at.slice(0, 10), "day")})`
      : `#${id}`;
  }
</script>

<div class="applications-page">
  <div class="page-header">
    <h2>Job Applications</h2>
    <button class="btn btn-primary" on:click={openAddForm}>
      + Add Application
    </button>
  </div>
  <p class="page-description">
    Track your job applications, update statuses, and monitor your pipeline.
  </p>

  <div class="search-bar">
    <input
      type="text"
      class="form-input search-input"
      placeholder="Search by company or position..."
      bind:value={searchQuery}
      on:input={handleSearchInput}
    />
  </div>

  {#if showForm}
    <div class="entry-form">
      <h3 class="form-title">
        {editingApp ? "Edit Application" : "New Application"}
      </h3>
      <div class="form-grid">
        <div class="form-field">
          <label class="form-label" for="app-company">Company</label>
          <input
            id="app-company"
            type="text"
            class="form-input"
            bind:value={formCompanyName}
            placeholder="e.g. Acme Corp"
          />
        </div>
        <div class="form-field">
          <label class="form-label" for="app-position">Position</label>
          <input
            id="app-position"
            type="text"
            class="form-input"
            bind:value={formPositionTitle}
            placeholder="e.g. Senior Developer"
          />
        </div>
        <div class="form-field">
          <label class="form-label" for="app-date">Date Applied</label>
          <input
            id="app-date"
            type="date"
            class="form-input"
            bind:value={formDateApplied}
          />
        </div>
        <div class="form-field">
          <label class="form-label" for="app-fit">Fit Indicator</label>
          <select id="app-fit" class="form-input" bind:value={formFitIndicator}>
            <option value="">-- None --</option>
            {#each fitIndicators as fi (fi)}
              <option value={fi}>{fi}</option>
            {/each}
          </select>
        </div>
        <div class="form-field">
          <label class="form-label" for="app-resume">Resume Export</label>
          <select
            id="app-resume"
            class="form-input"
            bind:value={formResumeExportId}
          >
            <option value={null}>-- None --</option>
            {#each exports as ex (ex.id)}
              <option value={ex.id}>
                {ex.template_id} ({formatDate(ex.generated_at.slice(0, 10), "day")})
              </option>
            {/each}
          </select>
        </div>
        <div class="form-field">
          <label class="form-label" for="app-cl">Cover Letter</label>
          <select id="app-cl" class="form-input" bind:value={formCoverLetterId}>
            <option value={null}>-- None --</option>
            {#each coverLetters as cl (cl.id)}
              <option value={cl.id}>{cl.title}</option>
            {/each}
          </select>
        </div>
      </div>
      <div class="form-field" style="margin-top: 12px;">
        <label class="form-label" for="app-notes">Notes</label>
        <textarea
          id="app-notes"
          class="form-input form-textarea"
          bind:value={formNotes}
          placeholder="Any notes about this application..."
          rows="3"
        />
      </div>
      <div class="form-actions">
        <button class="btn btn-primary" on:click={handleSubmit}>
          {editingApp ? "Update" : "Create"}
        </button>
        <button class="btn btn-cancel" on:click={cancelForm}>Cancel</button>
      </div>
    </div>
  {/if}

  {#if loading}
    <LoadingSpinner />
  {:else if applications.length === 0}
    <div class="empty-state">
      <p>
        {searchQuery.trim()
          ? "No applications match your search."
          : "No job applications yet."}
      </p>
      <p class="empty-hint">
        Track where you've applied, link resumes and cover letters, and monitor
        status changes.
      </p>
    </div>
  {:else}
    <div class="app-list">
      {#each applications as app (app.id)}
        <div class="app-card">
          <div class="app-header">
            <div class="app-info">
              <span class="app-company">{app.company_name}</span>
              <span class="app-position">{app.position_title}</span>
            </div>
            <div class="app-meta">
              <span class="app-date">{formatDate(app.date_applied, "day")}</span>
              {#if app.fit_indicator}
                <span
                  class="fit-badge"
                  style="color: {fitColor(
                    app.fit_indicator
                  )}; border-color: {fitColor(app.fit_indicator)}40;"
                >
                  {app.fit_indicator}
                </span>
              {/if}
            </div>
          </div>

          <div class="app-status-row">
            {#if changingStatusFor === app.id}
              <div class="status-change-form">
                <select
                  class="form-input status-select"
                  bind:value={newStatusValue}
                >
                  {#each statuses as s (s)}
                    <option value={s}>{s}</option>
                  {/each}
                </select>
                <button
                  class="btn btn-small btn-primary"
                  on:click={() => handleStatusChange(app.id)}
                >
                  Save
                </button>
                <button
                  class="btn btn-small btn-cancel"
                  on:click={() => (changingStatusFor = null)}
                >
                  Cancel
                </button>
              </div>
            {:else}
              <button
                class="status-badge"
                style="color: {statusColor(
                  app.status
                )}; border-color: {statusColor(app.status)}40;"
                on:click={() => openStatusChange(app)}
                title="Click to change status"
              >
                {app.status}
              </button>
            {/if}

            <div class="app-actions">
              <button
                class="btn btn-small btn-ghost"
                on:click={() => toggleHistory(app.id)}
              >
                {historyFor === app.id ? "Hide History" : "History"}
              </button>
              <button
                class="btn btn-small btn-ghost"
                on:click={() => openEditForm(app)}
              >
                Edit
              </button>
              <button
                class="btn btn-small btn-danger"
                on:click={() => handleDelete(app.id)}
              >
                Delete
              </button>
            </div>
          </div>

          {#if app.resume_export_id || app.cover_letter_id}
            <div class="app-links">
              {#if app.resume_export_id}
                <span class="link-tag"
                  >Resume: {getExportLabel(app.resume_export_id)}</span
                >
              {/if}
              {#if app.cover_letter_id}
                <span class="link-tag"
                  >CL: {getCoverLetterTitle(app.cover_letter_id)}</span
                >
              {/if}
            </div>
          {/if}

          {#if app.notes}
            <p class="app-notes">{app.notes}</p>
          {/if}

          {#if app.fit_indicator && changingStatusFor !== app.id}
            <div class="fit-row">
              <span class="fit-label">Fit:</span>
              {#each fitIndicators as fi (fi)}
                <button
                  class="fit-option"
                  class:fit-active={app.fit_indicator === fi}
                  style={app.fit_indicator === fi
                    ? `color: ${fitColor(fi)}; border-color: ${fitColor(fi)};`
                    : ""}
                  on:click={() => handleFitChange(app.id, fi)}
                >
                  {fi}
                </button>
              {/each}
            </div>
          {/if}

          {#if historyFor === app.id}
            <div class="history-section">
              <h4 class="history-title">Status History</h4>
              {#if historyEntries.length === 0}
                <p class="history-empty">No status changes recorded yet.</p>
              {:else}
                <div class="timeline">
                  {#each historyEntries as entry (entry.id)}
                    <div class="timeline-entry">
                      <span
                        class="timeline-dot"
                        style="background-color: {statusColor(
                          entry.to_status
                        )};"
                      ></span>
                      <div class="timeline-content">
                        <span class="timeline-status">
                          {entry.from_status} &rarr; {entry.to_status}
                        </span>
                        <span class="timeline-date">
                          {formatTimestamp(entry.changed_at)}
                        </span>
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .applications-page {
    max-width: 900px;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .page-header h2 {
    margin: 0;
    font-size: 1.5rem;
    color: #e0e0e0;
  }

  .page-description {
    color: #7a8a9a;
    font-size: 0.95rem;
    margin: 0 0 16px;
  }

  .search-bar {
    margin-bottom: 24px;
  }

  .search-input {
    width: 100%;
    box-sizing: border-box;
  }

  .empty-state {
    text-align: center;
    padding: 48px 0;
    color: #5a6a7a;
  }

  .empty-hint {
    font-size: 0.85rem;
    margin-top: 8px;
  }

  /* --- Form --- */
  .entry-form {
    background-color: #1e2d3d;
    border: 1px solid #3a4a5a;
    border-radius: 6px;
    padding: 20px;
    margin-bottom: 24px;
  }

  .form-title {
    margin: 0 0 16px;
    font-size: 1rem;
    color: #e0e0e0;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .form-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .form-label {
    font-size: 0.8rem;
    color: #7a8a9a;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .form-input {
    background-color: #1a2332;
    color: #e0e0e0;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 8px 10px;
    font-size: 0.9rem;
  }

  .form-input:focus {
    outline: none;
    border-color: #4a8af4;
    box-shadow: 0 0 0 2px rgba(74, 138, 244, 0.15);
  }

  .form-textarea {
    font-family: inherit;
    resize: vertical;
    line-height: 1.5;
  }

  .form-actions {
    display: flex;
    gap: 8px;
    margin-top: 16px;
  }

  /* --- Buttons --- */
  .btn {
    border: 1px solid #3a4a5a;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.85rem;
    padding: 8px 16px;
    transition:
      background-color 0.15s,
      border-color 0.15s;
  }

  .btn-primary {
    background-color: #2a5090;
    border-color: #3a60a0;
    color: #e0e0e0;
  }

  .btn-primary:hover {
    background-color: #3a60a0;
  }

  .btn-cancel {
    background-color: transparent;
    border-color: #3a4a5a;
    color: #7a8a9a;
  }

  .btn-cancel:hover {
    background-color: #2a3a4a;
    color: #c0d0e0;
  }

  .btn-small {
    padding: 4px 10px;
    font-size: 0.8rem;
  }

  .btn-ghost {
    background-color: transparent;
    border-color: transparent;
    color: #7a8a9a;
  }

  .btn-ghost:hover {
    background-color: #2a3a4a;
    color: #c0d0e0;
  }

  .btn-danger {
    background-color: transparent;
    border-color: transparent;
    color: #c05050;
  }

  .btn-danger:hover {
    background-color: #3a2020;
    color: #e06060;
  }

  /* --- Application Cards --- */
  .app-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .app-card {
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    padding: 16px;
  }

  .app-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .app-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .app-company {
    font-size: 1rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .app-position {
    font-size: 0.85rem;
    color: #a0b0c0;
  }

  .app-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .app-date {
    font-size: 0.8rem;
    color: #5a6a7a;
  }

  .fit-badge {
    font-size: 0.7rem;
    font-weight: 600;
    border: 1px solid;
    border-radius: 3px;
    padding: 1px 6px;
  }

  .app-status-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .status-badge {
    font-size: 0.8rem;
    font-weight: 600;
    border: 1px solid;
    border-radius: 4px;
    padding: 3px 10px;
    background: transparent;
    cursor: pointer;
    transition: opacity 0.15s;
  }

  .status-badge:hover {
    opacity: 0.8;
  }

  .status-change-form {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-select {
    min-width: 180px;
  }

  .app-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  .app-links {
    display: flex;
    gap: 8px;
    margin-bottom: 8px;
    flex-wrap: wrap;
  }

  .link-tag {
    font-size: 0.75rem;
    color: #7a8a9a;
    background-color: #1a2332;
    border: 1px solid #2a3a4a;
    border-radius: 3px;
    padding: 2px 8px;
  }

  .app-notes {
    margin: 0 0 8px;
    font-size: 0.8rem;
    color: #7a8a9a;
    line-height: 1.5;
    white-space: pre-wrap;
  }

  .fit-row {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-top: 8px;
    flex-wrap: wrap;
  }

  .fit-label {
    font-size: 0.75rem;
    color: #5a6a7a;
    font-weight: 600;
    margin-right: 4px;
  }

  .fit-option {
    font-size: 0.7rem;
    padding: 2px 6px;
    border: 1px solid #2a3a4a;
    border-radius: 3px;
    background: transparent;
    color: #5a6a7a;
    cursor: pointer;
    transition:
      color 0.15s,
      border-color 0.15s;
  }

  .fit-option:hover {
    color: #a0b0c0;
    border-color: #3a4a5a;
  }

  .fit-active {
    font-weight: 600;
  }

  /* --- History Timeline --- */
  .history-section {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #2a3a4a;
  }

  .history-title {
    margin: 0 0 8px;
    font-size: 0.85rem;
    color: #7a8a9a;
    font-weight: 600;
  }

  .history-empty {
    margin: 0;
    font-size: 0.8rem;
    color: #5a6a7a;
    font-style: italic;
  }

  .timeline {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-left: 12px;
  }

  .timeline-entry {
    display: flex;
    align-items: center;
    gap: 10px;
    position: relative;
  }

  .timeline-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .timeline-content {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .timeline-status {
    font-size: 0.8rem;
    color: #a0b0c0;
  }

  .timeline-date {
    font-size: 0.75rem;
    color: #5a6a7a;
  }
</style>
