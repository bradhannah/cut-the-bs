<script lang="ts">
  import { onMount } from "svelte";
  import {
    getProfile,
    updateProfile,
    listProfileLinks,
    createProfileLink,
    updateProfileLink,
    deleteProfileLink,
    reorderProfileLinks,
    getDataDirectory,
    setDataDirectory,
    browseForDataDirectory,
    getBackupSettings,
    updateBackupSettings,
    exportAllData,
    importAllData,
    openDataDirectory,
    addToast,
    type UserProfile,
    type ProfileLink,
    type ProfileLinkInput,
    type BackupSettings,
  } from "../services/api";
  import DragHandle from "../components/DragHandle.svelte";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";

  type SettingsTab = "profile" | "data";
  let activeTab: SettingsTab = "profile";

  // --- Profile State ---
  let profile: UserProfile = {
    id: 0,
    full_name: "",
    email: "",
    phone: "",
    location: "",
  };
  let profileLoading = true;
  let profileDirty = false;

  // --- Profile Links State ---
  let links: ProfileLink[] = [];
  let linksLoading = true;
  let showLinkForm = false;
  let editingLink: ProfileLink | null = null;
  let linkLabel = "";
  let linkUrl = "";

  onMount(async () => {
    await Promise.all([loadProfile(), loadLinks(), loadDataManagement()]);
  });

  async function loadProfile(): Promise<void> {
    profileLoading = true;
    try {
      profile = await getProfile();
    } finally {
      profileLoading = false;
      profileDirty = false;
    }
  }

  async function loadLinks(): Promise<void> {
    linksLoading = true;
    try {
      links = (await listProfileLinks()) || [];
    } finally {
      linksLoading = false;
    }
  }

  function markDirty(): void {
    profileDirty = true;
  }

  async function handleSaveProfile(): Promise<void> {
    if (!profile.full_name.trim()) {
      addToast("error", "Full name is required");
      return;
    }
    if (!profile.email.trim()) {
      addToast("error", "Email is required");
      return;
    }
    try {
      profile = await updateProfile(profile);
      profileDirty = false;
    } catch {
      // Toast already shown
    }
  }

  // --- Link CRUD ---

  function openAddLinkForm(): void {
    editingLink = null;
    linkLabel = "";
    linkUrl = "";
    showLinkForm = true;
  }

  function openEditLinkForm(link: ProfileLink): void {
    editingLink = link;
    linkLabel = link.label;
    linkUrl = link.url;
    showLinkForm = true;
  }

  function cancelLinkForm(): void {
    showLinkForm = false;
    editingLink = null;
  }

  async function handleLinkSubmit(): Promise<void> {
    const input: ProfileLinkInput = {
      label: linkLabel.trim(),
      url: linkUrl.trim(),
    };

    if (!input.label) {
      addToast("error", "Label is required");
      return;
    }
    if (!input.url) {
      addToast("error", "URL is required");
      return;
    }

    try {
      if (editingLink) {
        await updateProfileLink(editingLink.id, input);
      } else {
        await createProfileLink(input);
      }
      showLinkForm = false;
      editingLink = null;
      await loadLinks();
    } catch {
      // Toast already shown
    }
  }

  async function handleDeleteLink(id: number): Promise<void> {
    try {
      await deleteProfileLink(id);
      await loadLinks();
    } catch {
      // Toast already shown
    }
  }

  async function handleLinkReorder(orderedIDs: number[]): Promise<void> {
    try {
      await reorderProfileLinks(orderedIDs);
      await loadLinks();
    } catch {
      // Toast already shown
    }
  }

  function getLink(id: number): ProfileLink {
    return links.find((l) => l.id === id) as ProfileLink;
  }

  // --- Data Management State ---
  let dataDirectory = "";
  let backupSettings: BackupSettings = { rolling_backup_count: 5 };
  let backupCountInput = 5;
  let dataLoading = true;
  let exporting = false;
  let importing = false;
  let dataDirectoryInput = "";
  let updatingDataDirectory = false;

  async function loadDataManagement(): Promise<void> {
    dataLoading = true;
    try {
      dataDirectory = await getDataDirectory();
      dataDirectoryInput = dataDirectory;
      backupSettings = await getBackupSettings();
      backupCountInput = backupSettings.rolling_backup_count;
    } finally {
      dataLoading = false;
    }
  }

  async function handleExport(): Promise<void> {
    exporting = true;
    try {
      const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
      const outputPath = `${dataDirectory}/exports/cut-the-bs-export-${timestamp}.json`;
      await exportAllData(outputPath);
    } catch {
      // Toast already shown
    } finally {
      exporting = false;
    }
  }

  async function handleImportFromPath(): Promise<void> {
    const path = prompt("Enter the full path to your backup JSON file:");
    if (!path) return;

    importing = true;
    try {
      await importAllData(path.trim());
      // Reload all data after import.
      await Promise.all([loadProfile(), loadLinks(), loadDataManagement()]);
    } catch {
      // Toast already shown
    } finally {
      importing = false;
    }
  }

  async function handleUpdateBackupCount(): Promise<void> {
    if (backupCountInput < 1) {
      addToast("error", "Backup count must be at least 1");
      return;
    }
    try {
      await updateBackupSettings({
        rolling_backup_count: backupCountInput,
      });
      backupSettings.rolling_backup_count = backupCountInput;
    } catch {
      // Toast already shown
    }
  }

  async function handleOpenDataDir(): Promise<void> {
    try {
      await openDataDirectory();
    } catch {
      // Toast already shown
    }
  }

  async function handleSaveDataDirectory(): Promise<void> {
    if (updatingDataDirectory) {
      return;
    }

    updatingDataDirectory = true;
    try {
      await setDataDirectory(dataDirectoryInput.trim());
      await loadDataManagement();
    } catch {
      // Toast already shown
    } finally {
      updatingDataDirectory = false;
    }
  }

  async function handleBrowseDataDirectory(): Promise<void> {
    if (updatingDataDirectory) {
      return;
    }

    try {
      const selectedPath = await browseForDataDirectory();
      if (selectedPath.trim()) {
        dataDirectoryInput = selectedPath;
      }
    } catch {
      // Toast already shown
    }
  }

  async function handleResetDataDirectory(): Promise<void> {
    if (updatingDataDirectory) {
      return;
    }

    updatingDataDirectory = true;
    try {
      await setDataDirectory("");
      await loadDataManagement();
    } catch {
      // Toast already shown
    } finally {
      updatingDataDirectory = false;
    }
  }

  $: dataDirectoryInputDirty = dataDirectoryInput.trim() !== dataDirectory;

  $: sortedLinks = [...links].sort((a, b) => a.sort_order - b.sort_order);
</script>

<div class="settings-page">
  <h2>Settings</h2>
  <p class="page-description">
    Manage your profile details separately from workspace data and backups.
  </p>

  <div class="settings-layout">
    <div class="settings-tabs" role="tablist" aria-label="Settings sections">
      <button
        class="tab-btn"
        role="tab"
        class:active={activeTab === "profile"}
        aria-selected={activeTab === "profile"}
        on:click={() => (activeTab = "profile")}
      >
        <span class="tab-title">Profile</span>
        <span class="tab-meta">Identity and resume links</span>
      </button>

      <button
        class="tab-btn"
        role="tab"
        class:active={activeTab === "data"}
        aria-selected={activeTab === "data"}
        on:click={() => (activeTab = "data")}
      >
        <span class="tab-title">Data and Backup</span>
        <span class="tab-meta">Storage, import/export, retention</span>
      </button>
    </div>

    <div class="settings-content">
      {#if activeTab === "profile"}
        <section class="section">
          <h3 class="section-title">Profile</h3>

          {#if profileLoading}
            <LoadingSpinner message="Loading profile..." />
          {:else}
            <div class="form-row">
              <div class="form-field">
                <label class="form-label" for="full-name">Full Name *</label>
                <input
                  id="full-name"
                  type="text"
                  class="form-input"
                  bind:value={profile.full_name}
                  on:input={markDirty}
                  placeholder="Jane Doe"
                />
              </div>
              <div class="form-field">
                <label class="form-label" for="email">Email *</label>
                <input
                  id="email"
                  type="email"
                  class="form-input"
                  bind:value={profile.email}
                  on:input={markDirty}
                  placeholder="jane@example.com"
                />
              </div>
            </div>

            <div class="form-row">
              <div class="form-field">
                <label class="form-label" for="phone">Phone</label>
                <input
                  id="phone"
                  type="tel"
                  class="form-input"
                  bind:value={profile.phone}
                  on:input={markDirty}
                  placeholder="555-1234"
                />
              </div>
              <div class="form-field">
                <label class="form-label" for="location">Location</label>
                <input
                  id="location"
                  type="text"
                  class="form-input"
                  bind:value={profile.location}
                  on:input={markDirty}
                  placeholder="New York, NY"
                />
              </div>
            </div>

            <div class="form-actions">
              <button
                class="btn btn-primary"
                on:click={handleSaveProfile}
                disabled={!profileDirty}
              >
                Save Profile
              </button>
            </div>
          {/if}
        </section>

        <section class="section">
          <div class="section-header">
            <h3 class="section-title">Profile Links</h3>
            <button class="btn btn-small" on:click={openAddLinkForm}>
              + Add Link
            </button>
          </div>
          <p class="section-description">
            Links displayed on your resume header (LinkedIn, GitHub, portfolio, etc.).
          </p>

          {#if showLinkForm}
            <div class="link-form">
              <h4 class="form-subtitle">
                {editingLink ? "Edit Link" : "New Profile Link"}
              </h4>
              <div class="form-row">
                <div class="form-field">
                  <label class="form-label" for="link-label">Label</label>
                  <input
                    id="link-label"
                    type="text"
                    class="form-input"
                    bind:value={linkLabel}
                    placeholder="LinkedIn"
                  />
                </div>
                <div class="form-field">
                  <label class="form-label" for="link-url">URL</label>
                  <input
                    id="link-url"
                    type="url"
                    class="form-input"
                    bind:value={linkUrl}
                    placeholder="https://linkedin.com/in/yourname"
                  />
                </div>
              </div>
              <div class="form-actions">
                <button class="btn btn-primary" on:click={handleLinkSubmit}>
                  {editingLink ? "Update" : "Create"}
                </button>
                <button class="btn btn-cancel" on:click={cancelLinkForm}>
                  Cancel
                </button>
              </div>
            </div>
          {/if}

          {#if linksLoading}
            <LoadingSpinner message="Loading links..." />
          {:else if sortedLinks.length === 0}
            <div class="empty-state">
              <p>No profile links yet.</p>
              <p class="empty-hint">
                Add links like LinkedIn, GitHub, or your portfolio.
              </p>
            </div>
          {:else}
            <div class="links-list">
              <DragHandle
                items={sortedLinks}
                on:reorder={(e) => handleLinkReorder(e.detail.orderedIDs)}
                let:item
              >
                {@const link = getLink(item.id)}
                <div class="link-card">
                  <div class="link-info">
                    <span class="link-label">{link.label}</span>
                    <span class="link-url">{link.url}</span>
                  </div>
                  <div class="link-actions">
                    <button
                      class="btn btn-small btn-ghost"
                      on:click={() => openEditLinkForm(link)}
                    >
                      Edit
                    </button>
                    <button
                      class="btn btn-small btn-danger"
                      on:click={() => handleDeleteLink(link.id)}
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </DragHandle>
            </div>
          {/if}
        </section>
      {:else}
        <section class="section">
          <h3 class="section-title">Data Management</h3>
          <p class="section-description">
            Export, import, and backup your resume data.
          </p>

          {#if dataLoading}
            <LoadingSpinner message="Loading data settings..." />
          {:else}
            <div class="data-dir-card">
              <div class="data-dir-info">
                <label class="data-dir-label" for="data-directory">
                  Data Directory
                </label>
              </div>
              <div class="data-dir-edit-row">
                <input
                  id="data-directory"
                  type="text"
                  class="form-input data-dir-input"
                  bind:value={dataDirectoryInput}
                  placeholder="/path/to/your-data-directory"
                />
                <button
                  class="btn btn-small"
                  on:click={handleBrowseDataDirectory}
                  disabled={updatingDataDirectory}
                >
                  Browse...
                </button>
              </div>
              <div class="data-dir-actions-row">
                <button
                  class="btn btn-small"
                  on:click={handleSaveDataDirectory}
                  disabled={!dataDirectoryInputDirty || updatingDataDirectory}
                >
                  {updatingDataDirectory ? "Saving..." : "Save"}
                </button>
                <button
                  class="btn btn-small btn-ghost"
                  on:click={handleResetDataDirectory}
                  disabled={updatingDataDirectory}
                >
                  Reset
                </button>
                <button
                  class="btn btn-small"
                  on:click={handleOpenDataDir}
                  disabled={updatingDataDirectory}
                >
                  Open
                </button>
              </div>
              <div class="data-dir-current-row">
                <span class="data-dir-current-label">Current active directory</span>
                <span class="data-dir-path">{dataDirectory}</span>
              </div>
              <p class="data-dir-note">
                Use a custom folder for alternate datasets (for example, screenshot
                sample data). Restart the app after saving.
              </p>
            </div>

            <div class="data-actions">
              <div class="data-action-group">
                <h4 class="data-action-title">Export</h4>
                <p class="data-action-desc">
                  Export all data to a JSON file for backup or transfer.
                </p>
                <button
                  class="btn btn-primary"
                  on:click={handleExport}
                  disabled={exporting}
                >
                  {exporting ? "Exporting..." : "Export All Data"}
                </button>
              </div>

              <div class="data-action-group">
                <h4 class="data-action-title">Import</h4>
                <p class="data-action-desc">
                  Restore all data from a JSON backup file. This replaces existing
                  data.
                </p>
                <button
                  class="btn btn-primary"
                  on:click={handleImportFromPath}
                  disabled={importing}
                >
                  {importing ? "Importing..." : "Import All Data"}
                </button>
              </div>
            </div>

            <div class="backup-settings">
              <h4 class="data-action-title">Rolling Backups</h4>
              <p class="data-action-desc">
                Automatic database backups are retained up to the configured count.
              </p>
              <div class="backup-count-row">
                <label class="form-label" for="backup-count">
                  Max Backup Count
                </label>
                <input
                  id="backup-count"
                  type="number"
                  class="form-input backup-count-input"
                  min="1"
                  max="50"
                  bind:value={backupCountInput}
                />
                <button
                  class="btn btn-primary btn-small"
                  on:click={handleUpdateBackupCount}
                  disabled={backupCountInput === backupSettings.rolling_backup_count}
                >
                  Save
                </button>
              </div>
            </div>
          {/if}
        </section>
      {/if}
    </div>
  </div>
</div>

<style>
  .settings-page {
    max-width: 1080px;
  }

  .settings-page h2 {
    margin: 0 0 4px;
    font-size: 1.5rem;
    color: #e0e0e0;
  }

  .page-description {
    color: #7a8a9a;
    font-size: 0.95rem;
    margin: 0 0 24px;
  }

  .settings-layout {
    display: grid;
    grid-template-columns: 220px minmax(0, 1fr);
    gap: 18px;
    align-items: start;
  }

  .settings-tabs {
    display: flex;
    flex-direction: column;
    gap: 8px;
    position: sticky;
    top: 0;
  }

  .tab-btn {
    display: flex;
    flex-direction: column;
    gap: 3px;
    text-align: left;
    border: 1px solid #2f4257;
    background: #1a2838;
    border-radius: 8px;
    color: #9ab0c6;
    padding: 10px 12px;
    cursor: pointer;
    transition: border-color 0.15s, background-color 0.15s, color 0.15s;
  }

  .tab-btn:hover {
    border-color: #476584;
    background: #203245;
    color: #d4e4f4;
  }

  .tab-btn.active {
    border-color: #4a8af4;
    background: #223a56;
    color: #edf4ff;
  }

  .tab-title {
    font-size: 0.86rem;
    font-weight: 700;
  }

  .tab-meta {
    font-size: 0.75rem;
    color: #7f95aa;
  }

  .tab-btn.active .tab-meta {
    color: #b8d0e8;
  }

  .settings-content {
    min-width: 0;
  }

  /* --- Sections --- */
  .section {
    margin-bottom: 32px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 4px;
  }

  .section-title {
    margin: 0 0 12px;
    font-size: 1.1rem;
    color: #c0d0e0;
    font-weight: 600;
  }

  .section-header .section-title {
    margin-bottom: 0;
  }

  .section-description {
    color: #5a6a7a;
    font-size: 0.85rem;
    margin: 0 0 16px;
  }

  .empty-state {
    text-align: center;
    padding: 24px 0;
    color: #5a6a7a;
  }

  .empty-hint {
    font-size: 0.85rem;
    margin-top: 8px;
  }

  /* --- Forms --- */
  .form-row {
    display: flex;
    gap: 16px;
    margin-bottom: 16px;
  }

  .form-field {
    flex: 1;
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

  .form-actions {
    display: flex;
    gap: 8px;
  }

  .form-subtitle {
    margin: 0 0 12px;
    font-size: 0.95rem;
    color: #e0e0e0;
  }

  .link-form {
    background-color: #1e2d3d;
    border: 1px solid #3a4a5a;
    border-radius: 6px;
    padding: 16px;
    margin-bottom: 16px;
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

  .btn-primary:hover:not(:disabled) {
    background-color: #3a60a0;
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

  /* --- Link Cards --- */
  .links-list {
    display: flex;
    flex-direction: column;
  }

  .link-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px;
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    margin-bottom: 6px;
  }

  .link-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .link-label {
    font-size: 0.9rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .link-url {
    font-size: 0.8rem;
    color: #5a6a7a;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .link-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  /* --- Data Management --- */
  .data-dir-card {
    padding: 10px 12px;
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    margin-bottom: 16px;
  }

  .data-dir-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .data-dir-label {
    font-size: 0.8rem;
    color: #7a8a9a;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .data-dir-path {
    font-size: 0.85rem;
    color: #c0d0e0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: monospace;
  }

  .data-dir-edit-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
  }

  .data-dir-actions-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
  }

  .data-dir-current-row {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-top: 10px;
  }

  .data-dir-current-label {
    font-size: 0.75rem;
    color: #7a8a9a;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .data-dir-input {
    flex: 1;
    min-width: 0;
    font-family: monospace;
  }

  .data-dir-note {
    margin: 8px 0 0;
    font-size: 0.8rem;
    color: #5a6a7a;
  }

  .data-actions {
    display: flex;
    gap: 16px;
    margin-bottom: 16px;
  }

  .data-action-group {
    flex: 1;
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    padding: 16px;
  }

  .data-action-title {
    margin: 0 0 4px;
    font-size: 0.95rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .data-action-desc {
    color: #5a6a7a;
    font-size: 0.85rem;
    margin: 0 0 12px;
  }

  .backup-settings {
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    padding: 16px;
  }

  .backup-count-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .backup-count-input {
    width: 70px;
  }

  @media (max-width: 920px) {
    .settings-layout {
      grid-template-columns: 1fr;
    }

    .settings-tabs {
      position: static;
      flex-direction: row;
      flex-wrap: wrap;
    }

    .tab-btn {
      flex: 1 1 220px;
    }
  }
</style>
