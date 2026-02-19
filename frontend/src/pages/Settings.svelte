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
    addToast,
    type UserProfile,
    type ProfileLink,
    type ProfileLinkInput,
  } from "../services/api";
  import DragHandle from "../components/DragHandle.svelte";

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
    await Promise.all([loadProfile(), loadLinks()]);
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
      links = await listProfileLinks();
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

  $: sortedLinks = [...links].sort((a, b) => a.sort_order - b.sort_order);
</script>

<div class="settings-page">
  <h2>Settings</h2>
  <p class="page-description">
    Your profile information for resumes and cover letters.
  </p>

  <!-- Profile Section -->
  <section class="section">
    <h3 class="section-title">Profile</h3>

    {#if profileLoading}
      <p class="loading-message">Loading profile...</p>
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

  <!-- Profile Links Section -->
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
      <p class="loading-message">Loading links...</p>
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
</div>

<style>
  .settings-page {
    max-width: 800px;
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

  .loading-message {
    color: #5a6a7a;
    font-style: italic;
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
</style>
