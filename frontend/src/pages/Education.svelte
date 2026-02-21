<script lang="ts">
  import { onMount } from "svelte";
  import {
    listAcademicCredentials,
    createAcademicCredential,
    updateAcademicCredential,
    deleteAcademicCredential,
    reorderAcademicCredentials,
    listCertifications,
    createCertification,
    updateCertification,
    deleteCertification,
    reorderCertifications,
    checkAcademicLensReferences,
    checkCertLensReferences,
    addToast,
    type AcademicCredential,
    type AcademicInput,
    type Certification,
    type CertificationInput,
  } from "../services/api";
  import DragHandle from "../components/DragHandle.svelte";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";
  import { formatDate } from "../services/dateFormat";

  // --- Academic State ---
  let academics: AcademicCredential[] = [];
  let academicLoading = true;
  let showAcademicForm = false;
  let editingAcademic: AcademicCredential | null = null;
  let acadInstitution = "";
  let acadCredentialType = "";
  let acadFieldOfStudy = "";
  let acadCompletionDate = "";
  let acadDateGranularity = "month";

  // --- Certification State ---
  let certifications: Certification[] = [];
  let certLoading = true;
  let showCertForm = false;
  let editingCert: Certification | null = null;
  let certName = "";
  let certIssuingBody = "";
  let certDateEarned = "";
  let certExpirationDate = "";
  let certNoExpiration = true;

  // Delete confirmation state
  let deleteConfirmAcademic: AcademicCredential | null = null;
  let deleteConfirmCert: Certification | null = null;
  let lensReferences: string[] = [];

  onMount(async () => {
    await Promise.all([loadAcademics(), loadCertifications()]);
  });

  async function loadAcademics(): Promise<void> {
    academicLoading = true;
    try {
      academics = (await listAcademicCredentials()) || [];
    } finally {
      academicLoading = false;
    }
  }

  async function loadCertifications(): Promise<void> {
    certLoading = true;
    try {
      certifications = (await listCertifications()) || [];
    } finally {
      certLoading = false;
    }
  }

  // --- Academic CRUD ---

  function openAddAcademic(): void {
    editingAcademic = null;
    acadInstitution = "";
    acadCredentialType = "";
    acadFieldOfStudy = "";
    acadCompletionDate = "";
    acadDateGranularity = "month";
    showAcademicForm = true;
  }

  function openEditAcademic(cred: AcademicCredential): void {
    editingAcademic = cred;
    acadInstitution = cred.institution;
    acadCredentialType = cred.credential_type;
    acadFieldOfStudy = cred.field_of_study;
    acadCompletionDate = cred.completion_date;
    acadDateGranularity = cred.date_granularity || "month";
    showAcademicForm = true;
  }

  function cancelAcademicForm(): void {
    showAcademicForm = false;
    editingAcademic = null;
  }

  async function handleAcademicSubmit(): Promise<void> {
    const input: AcademicInput = {
      institution: acadInstitution.trim(),
      credential_type: acadCredentialType.trim(),
      field_of_study: acadFieldOfStudy.trim(),
      completion_date: acadCompletionDate,
      date_granularity: acadDateGranularity,
    };

    if (!input.institution || !input.field_of_study) {
      addToast(
        "error",
        "Institution and field of study are required"
      );
      return;
    }
    if (!input.completion_date) {
      addToast("error", "Completion date is required");
      return;
    }

    try {
      if (editingAcademic) {
        await updateAcademicCredential(editingAcademic.id, input);
      } else {
        await createAcademicCredential(input);
      }
      showAcademicForm = false;
      editingAcademic = null;
      await loadAcademics();
    } catch {
      // Toast already shown
    }
  }

  async function confirmDeleteAcademic(cred: AcademicCredential): Promise<void> {
    try {
      lensReferences = (await checkAcademicLensReferences(cred.id)) || [];
      deleteConfirmAcademic = cred;
    } catch {
      // Toast already shown
    }
  }

  async function handleDeleteAcademic(): Promise<void> {
    if (!deleteConfirmAcademic) return;
    try {
      await deleteAcademicCredential(deleteConfirmAcademic.id);
      deleteConfirmAcademic = null;
      lensReferences = [];
      await loadAcademics();
    } catch {
      // Toast already shown
    }
  }

  function cancelDeleteAcademic(): void {
    deleteConfirmAcademic = null;
    lensReferences = [];
  }

  async function handleAcademicReorder(orderedIDs: number[]): Promise<void> {
    try {
      await reorderAcademicCredentials(orderedIDs);
      await loadAcademics();
    } catch {
      // Toast already shown
    }
  }

  // --- Certification CRUD ---

  function openAddCert(): void {
    editingCert = null;
    certName = "";
    certIssuingBody = "";
    certDateEarned = "";
    certExpirationDate = "";
    certNoExpiration = true;
    showCertForm = true;
  }

  function openEditCert(cert: Certification): void {
    editingCert = cert;
    certName = cert.name;
    certIssuingBody = cert.issuing_body;
    certDateEarned = cert.date_earned;
    certExpirationDate = cert.expiration_date;
    certNoExpiration = !cert.expiration_date;
    showCertForm = true;
  }

  function cancelCertForm(): void {
    showCertForm = false;
    editingCert = null;
  }

  async function handleCertSubmit(): Promise<void> {
    const input: CertificationInput = {
      name: certName.trim(),
      issuing_body: certIssuingBody.trim(),
      date_earned: certDateEarned,
      expiration_date: certNoExpiration ? "" : certExpirationDate,
    };

    if (!input.name || !input.issuing_body) {
      addToast("error", "Name and issuing body are required");
      return;
    }
    if (!input.date_earned) {
      addToast("error", "Date earned is required");
      return;
    }

    try {
      if (editingCert) {
        await updateCertification(editingCert.id, input);
      } else {
        await createCertification(input);
      }
      showCertForm = false;
      editingCert = null;
      await loadCertifications();
    } catch {
      // Toast already shown
    }
  }

  async function confirmDeleteCert(cert: Certification): Promise<void> {
    try {
      lensReferences = (await checkCertLensReferences(cert.id)) || [];
      deleteConfirmCert = cert;
    } catch {
      // Toast already shown
    }
  }

  async function handleDeleteCert(): Promise<void> {
    if (!deleteConfirmCert) return;
    try {
      await deleteCertification(deleteConfirmCert.id);
      deleteConfirmCert = null;
      lensReferences = [];
      await loadCertifications();
    } catch {
      // Toast already shown
    }
  }

  function cancelDeleteCert(): void {
    deleteConfirmCert = null;
    lensReferences = [];
  }

  async function handleCertReorder(orderedIDs: number[]): Promise<void> {
    try {
      await reorderCertifications(orderedIDs);
      await loadCertifications();
    } catch {
      // Toast already shown
    }
  }

  // Helpers
  function getAcademic(id: number): AcademicCredential {
    return academics.find((a) => a.id === id) as AcademicCredential;
  }

  function getCert(id: number): Certification {
    return certifications.find((c) => c.id === id) as Certification;
  }

  $: sortedAcademics = [...academics].sort(
    (a, b) => a.sort_order - b.sort_order
  );
  $: sortedCerts = [...certifications].sort(
    (a, b) => a.sort_order - b.sort_order
  );
</script>

<div class="education-page">
  <h2>Education</h2>
  <p class="page-description">
    Manage your academic credentials and professional certifications.
  </p>

  <!-- Academic Credentials Section -->
  <section class="section">
    <div class="section-header">
      <h3 class="section-title">Academic Credentials</h3>
      <button class="btn btn-primary" on:click={openAddAcademic}>
        + Add Credential
      </button>
    </div>

    {#if showAcademicForm}
      <div class="entry-form">
        <h4 class="form-subtitle">
          {editingAcademic ? "Edit Credential" : "New Academic Credential"}
        </h4>
        <div class="form-row">
          <div class="form-field">
            <label class="form-label" for="acad-institution">Institution</label>
            <input
              id="acad-institution"
              type="text"
              class="form-input"
              bind:value={acadInstitution}
              placeholder="University of Example"
            />
          </div>
          <div class="form-field">
            <label class="form-label" for="acad-type">Credential Type</label>
            <input
              id="acad-type"
              type="text"
              class="form-input"
              bind:value={acadCredentialType}
              placeholder="B.S., M.S., Ph.D., etc."
            />
          </div>
        </div>
        <div class="form-row">
          <div class="form-field">
            <label class="form-label" for="acad-field">Field of Study</label>
            <input
              id="acad-field"
              type="text"
              class="form-input"
              bind:value={acadFieldOfStudy}
              placeholder="Computer Science"
            />
          </div>
          <div class="form-field">
            <label class="form-label" for="acad-date">Completion Date</label>
            <input
              id="acad-date"
              type="date"
              class="form-input"
              bind:value={acadCompletionDate}
            />
          </div>
        </div>
        <div class="form-row">
          <div class="form-field">
            <label class="form-label" for="acad-granularity">Date Display</label
            >
            <select
              id="acad-granularity"
              class="form-input"
              bind:value={acadDateGranularity}
            >
              <option value="year">Year only</option>
              <option value="month">Month + Year</option>
              <option value="day">Full date</option>
            </select>
          </div>
          <div class="form-field" />
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" on:click={handleAcademicSubmit}>
            {editingAcademic ? "Update" : "Create"}
          </button>
          <button class="btn btn-cancel" on:click={cancelAcademicForm}>
            Cancel
          </button>
        </div>
      </div>
    {/if}

    <!-- Academic Delete Confirmation -->
    {#if deleteConfirmAcademic}
      <div class="confirm-dialog">
        <p>
          Delete <strong>{deleteConfirmAcademic.credential_type ? `${deleteConfirmAcademic.credential_type} in ` : ""}{deleteConfirmAcademic.field_of_study}</strong> from {deleteConfirmAcademic.institution}?
        </p>
        {#if lensReferences.length > 0}
          <p class="warn-text">
            This credential is referenced by {lensReferences.length} lens{lensReferences.length !== 1 ? "es" : ""}:
            {lensReferences.join(", ")}
          </p>
        {/if}
        <div class="form-actions">
          <button class="btn btn-danger-solid" on:click={handleDeleteAcademic}>
            Delete
          </button>
          <button class="btn btn-cancel" on:click={cancelDeleteAcademic}>Cancel</button>
        </div>
      </div>
    {/if}

    {#if academicLoading}
      <LoadingSpinner />
    {:else if sortedAcademics.length === 0}
      <div class="empty-state">
        <p>No academic credentials yet.</p>
        <p class="empty-hint">
          Add your degrees, diplomas, or other academic achievements.
        </p>
      </div>
    {:else}
      <div class="items-list">
        <DragHandle
          items={sortedAcademics}
          on:reorder={(e) => handleAcademicReorder(e.detail.orderedIDs)}
          let:item
        >
          {@const cred = getAcademic(item.id)}
          <div class="item-card">
            <div class="item-info">
              <span class="item-primary">
                {cred.credential_type ? `${cred.credential_type} in ` : ""}{cred.field_of_study}
              </span>
              <span class="item-secondary">
                {cred.institution}
                {#if cred.completion_date}
                  &middot; {formatDate(
                    cred.completion_date,
                    cred.date_granularity
                  )}
                {/if}
              </span>
            </div>
            <div class="item-actions">
              <button
                class="btn btn-small btn-ghost"
                on:click={() => openEditAcademic(cred)}
              >
                Edit
              </button>
              <button
                class="btn btn-small btn-danger"
                on:click={() => confirmDeleteAcademic(cred)}
              >
                Delete
              </button>
            </div>
          </div>
        </DragHandle>
      </div>
    {/if}
  </section>

  <!-- Certifications Section -->
  <section class="section">
    <div class="section-header">
      <h3 class="section-title">Certifications</h3>
      <button class="btn btn-primary" on:click={openAddCert}>
        + Add Certification
      </button>
    </div>

    {#if showCertForm}
      <div class="entry-form">
        <h4 class="form-subtitle">
          {editingCert ? "Edit Certification" : "New Certification"}
        </h4>
        <div class="form-row">
          <div class="form-field">
            <label class="form-label" for="cert-name">Name</label>
            <input
              id="cert-name"
              type="text"
              class="form-input"
              bind:value={certName}
              placeholder="AWS Solutions Architect"
            />
          </div>
          <div class="form-field">
            <label class="form-label" for="cert-issuer">Issuing Body</label>
            <input
              id="cert-issuer"
              type="text"
              class="form-input"
              bind:value={certIssuingBody}
              placeholder="Amazon Web Services"
            />
          </div>
        </div>
        <div class="form-row">
          <div class="form-field">
            <label class="form-label" for="cert-earned">Date Earned</label>
            <input
              id="cert-earned"
              type="date"
              class="form-input"
              bind:value={certDateEarned}
            />
          </div>
          <div class="form-field">
            <label class="form-label">
              <input type="checkbox" bind:checked={certNoExpiration} />
              No expiration date
            </label>
            {#if !certNoExpiration}
              <input
                type="date"
                class="form-input"
                bind:value={certExpirationDate}
              />
            {/if}
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" on:click={handleCertSubmit}>
            {editingCert ? "Update" : "Create"}
          </button>
          <button class="btn btn-cancel" on:click={cancelCertForm}>
            Cancel
          </button>
        </div>
      </div>
    {/if}

    <!-- Cert Delete Confirmation -->
    {#if deleteConfirmCert}
      <div class="confirm-dialog">
        <p>
          Delete certification <strong>{deleteConfirmCert.name}</strong>?
        </p>
        {#if lensReferences.length > 0}
          <p class="warn-text">
            This certification is referenced by {lensReferences.length} lens{lensReferences.length !== 1 ? "es" : ""}:
            {lensReferences.join(", ")}
          </p>
        {/if}
        <div class="form-actions">
          <button class="btn btn-danger-solid" on:click={handleDeleteCert}>
            Delete
          </button>
          <button class="btn btn-cancel" on:click={cancelDeleteCert}>Cancel</button>
        </div>
      </div>
    {/if}

    {#if certLoading}
      <LoadingSpinner />
    {:else if sortedCerts.length === 0}
      <div class="empty-state">
        <p>No certifications yet.</p>
        <p class="empty-hint">
          Add your professional certifications and licenses.
        </p>
      </div>
    {:else}
      <div class="items-list">
        <DragHandle
          items={sortedCerts}
          on:reorder={(e) => handleCertReorder(e.detail.orderedIDs)}
          let:item
        >
          {@const cert = getCert(item.id)}
          <div class="item-card">
            <div class="item-info">
              <div class="item-primary-row">
                <span class="item-primary">{cert.name}</span>
                {#if cert.is_active}
                  <span class="status-badge active">Active</span>
                {:else}
                  <span class="status-badge expired">Expired</span>
                {/if}
              </div>
              <span class="item-secondary">
                {cert.issuing_body}
                &middot; Earned {formatDate(cert.date_earned, "day")}
                {#if cert.expiration_date}
                  &middot; Expires {formatDate(cert.expiration_date, "day")}
                {/if}
              </span>
            </div>
            <div class="item-actions">
              <button
                class="btn btn-small btn-ghost"
                on:click={() => openEditCert(cert)}
              >
                Edit
              </button>
              <button
                class="btn btn-small btn-danger"
                on:click={() => confirmDeleteCert(cert)}
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
  .education-page {
    max-width: 800px;
  }

  .education-page h2 {
    margin: 0 0 4px;
    font-size: 1.5rem;
    color: #e0e0e0;
  }

  .page-description {
    color: #7a8a9a;
    font-size: 0.95rem;
    margin: 0 0 24px;
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

  /* --- Sections --- */
  .section {
    margin-bottom: 32px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .section-title {
    margin: 0;
    font-size: 1.1rem;
    color: #c0d0e0;
    font-weight: 600;
  }

  /* --- Forms --- */
  .entry-form {
    background-color: #1e2d3d;
    border: 1px solid #3a4a5a;
    border-radius: 6px;
    padding: 20px;
    margin-bottom: 16px;
  }

  .form-subtitle {
    margin: 0 0 16px;
    font-size: 1rem;
    color: #e0e0e0;
  }

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

  .form-field label:has(input[type="checkbox"]) {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    text-transform: none;
    font-weight: 400;
    font-size: 0.85rem;
    color: #7a8a9a;
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

  .btn-danger-solid {
    background-color: #802020;
    border-color: #a03030;
    color: #e0e0e0;
  }

  .btn-danger-solid:hover {
    background-color: #a03030;
  }

  /* --- Confirm Dialog --- */
  .confirm-dialog {
    background-color: #2a1a1a;
    border: 1px solid #5a3030;
    border-radius: 6px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .confirm-dialog p {
    margin: 0 0 8px;
    color: #e0e0e0;
    font-size: 0.9rem;
  }

  .warn-text {
    color: #e0a060 !important;
    font-size: 0.85rem !important;
  }

  /* --- Item List --- */
  .items-list {
    display: flex;
    flex-direction: column;
  }

  .item-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px;
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    margin-bottom: 6px;
  }

  .item-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .item-primary-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .item-primary {
    font-size: 0.9rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .item-secondary {
    font-size: 0.8rem;
    color: #5a6a7a;
  }

  .item-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  /* --- Status Badges --- */
  .status-badge {
    font-size: 0.7rem;
    padding: 2px 8px;
    border-radius: 8px;
    font-weight: 600;
  }

  .status-badge.active {
    background-color: #1a3a2a;
    color: #60c080;
  }

  .status-badge.expired {
    background-color: #3a2020;
    color: #c06060;
  }
</style>
