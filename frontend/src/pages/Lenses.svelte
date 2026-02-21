<script lang="ts">
  import { onMount } from "svelte";
  import {
    listLenses,
    getLens,
    createLens,
    updateLens,
    deleteLens,
    setLensWorkHistory,
    setLensBullets,
    setLensSkills,
    setLensAcademics,
    setLensCerts,
    setLensDescriptors,
    setLensSummaries,
    setLensCoreExpertise,
    listWorkHistory,
    listSkillsByCategory,
    listAcademicCredentials,
    listCertifications,
    listDescriptors,
    listSummaries,
    listCoreExpertise,
    addToast,
    type Lens,
    type LensInput,
    type LensDetail,
    type LensWorkHistoryItem,
    type LensBulletItem,
    type LensSkillItem,
    type LensDescriptorItem,
    type LensCoreExpertiseItem,
    type LensSummaryItem,
    type WorkHistoryEntry,
    type SkillCategoryWithSkills,
    type AcademicCredential,
    type Certification,
    type RoleDescriptor,
    type ProfessionalSummary,
    type CoreExpertise,
  } from "../services/api";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";

  // --- Data ---
  let lenses: Lens[] = [];
  let loading = true;

  // Content data for selection panels
  let workHistory: WorkHistoryEntry[] = [];
  let skillCategories: SkillCategoryWithSkills[] = [];
  let academics: AcademicCredential[] = [];
  let certs: Certification[] = [];
  let descriptors: RoleDescriptor[] = [];
  let summaries: ProfessionalSummary[] = [];
  let coreExpertise: CoreExpertise[] = [];

  // Lens form state
  let showLensForm = false;
  let editingLens: Lens | null = null;
  let formName = "";

  // Currently selected lens for detail viewing/editing
  let activeLensId: number | null = null;
  let activeLensDetail: LensDetail | null = null;
  let loadingDetail = false;
  let editMode = false;

  // Selection state (mirrors the lens detail for editing)
  let selectedWorkHistoryIds: Set<number> = new Set();
  let selectedBulletIds: Set<number> = new Set();
  let selectedSkillIds: Set<number> = new Set();
  let selectedAcademicIds: Set<number> = new Set();
  let selectedCertIds: Set<number> = new Set();
  let selectedDescriptorIds: Set<number> = new Set();
  let selectedSummaryIds: Set<number> = new Set();
  let selectedCoreExpertiseIds: Set<number> = new Set();
  let masterSummaryId: number | null = null;

  // Delete confirmation
  let deleteConfirmLens: Lens | null = null;

  // Skills category panel state
  let skillCategoriesExpanded = true;

  // Unified save state
  let saving = false;

  // Saved-state snapshots for dirty tracking (set when lens is loaded)
  let savedWorkHistoryIds: Set<number> = new Set();
  let savedBulletIds: Set<number> = new Set();
  let savedSkillIds: Set<number> = new Set();
  let savedAcademicIds: Set<number> = new Set();
  let savedCertIds: Set<number> = new Set();
  let savedDescriptorIds: Set<number> = new Set();
  let savedSummaryIds: Set<number> = new Set();
  let savedCoreExpertiseIds: Set<number> = new Set();
  let savedMasterSummaryId: number | null = null;

  // Dirty state detection
  function setsEqual(a: Set<number>, b: Set<number>): boolean {
    if (a.size !== b.size) return false;
    for (const v of a) {
      if (!b.has(v)) return false;
    }
    return true;
  }

  $: isDirty =
    activeLensId != null &&
    (!setsEqual(selectedWorkHistoryIds, savedWorkHistoryIds) ||
      !setsEqual(selectedBulletIds, savedBulletIds) ||
      !setsEqual(selectedSkillIds, savedSkillIds) ||
      !setsEqual(selectedAcademicIds, savedAcademicIds) ||
      !setsEqual(selectedCertIds, savedCertIds) ||
      !setsEqual(selectedDescriptorIds, savedDescriptorIds) ||
      !setsEqual(selectedSummaryIds, savedSummaryIds) ||
      !setsEqual(selectedCoreExpertiseIds, savedCoreExpertiseIds) ||
      masterSummaryId !== savedMasterSummaryId);

  onMount(async () => {
    await loadAllData();
  });

  async function loadAllData(): Promise<void> {
    loading = true;
    try {
      const results = await Promise.all([
        listLenses(),
        listWorkHistory(),
        listSkillsByCategory(),
        listAcademicCredentials(),
        listCertifications(),
        listDescriptors(),
        listSummaries(),
        listCoreExpertise(),
      ]);

      lenses = results[0] || [];
      workHistory = results[1] || [];
      skillCategories = results[2] || [];
      academics = results[3] || [];
      certs = results[4] || [];
      descriptors = results[5] || [];
      summaries = results[6] || [];
      coreExpertise = results[7] || [];
    } finally {
      loading = false;
    }
  }

  // --- Lens CRUD ---

  function openAddLens(): void {
    editingLens = null;
    formName = "";
    showLensForm = true;
  }

  function openEditLens(lens: Lens): void {
    editingLens = lens;
    formName = lens.name;
    showLensForm = true;
  }

  function cancelLensForm(): void {
    showLensForm = false;
    editingLens = null;
  }

  async function handleLensSubmit(): Promise<void> {
    const input: LensInput = {
      name: formName.trim(),
    };

    if (!input.name) {
      addToast("error", "Lens name is required");
      return;
    }

    try {
      if (editingLens) {
        await updateLens(editingLens.id, input);
      } else {
        const created = await createLens(input);
        activeLensId = created.id;
      }
      showLensForm = false;
      editingLens = null;
      lenses = await listLenses();

      if (activeLensId) {
        await loadLensDetail(activeLensId);
      }
    } catch {
      // Toast already shown
    }
  }

  function confirmDeleteLens(lens: Lens): void {
    deleteConfirmLens = lens;
  }

  async function handleDeleteLens(): Promise<void> {
    if (!deleteConfirmLens) return;
    try {
      await deleteLens(deleteConfirmLens.id);
      if (activeLensId === deleteConfirmLens.id) {
        activeLensId = null;
        activeLensDetail = null;
        editMode = false;
      }
      deleteConfirmLens = null;
      lenses = await listLenses();
    } catch {
      // Toast already shown
    }
  }

  function cancelDelete(): void {
    deleteConfirmLens = null;
  }

  // --- Lens Detail Loading ---

  async function selectLens(id: number): Promise<void> {
    editMode = false;
    await loadLensDetail(id);
  }

  function goBackToList(): void {
    if (editMode && isDirty) {
      discardChanges();
    }
    activeLensId = null;
    activeLensDetail = null;
    editMode = false;
  }

  function enterEditMode(): void {
    editMode = true;
  }

  async function loadLensDetail(id: number): Promise<void> {
    loadingDetail = true;
    activeLensId = id;
    try {
      activeLensDetail = await getLens(id);

      // Populate selection sets from detail
      selectedSummaryIds = new Set(
        (activeLensDetail.summaries || []).map((s) => s.summary_id)
      );
      // Find the master summary (if any).
      const masterItem = (activeLensDetail.summaries || []).find((s) => s.is_master);
      masterSummaryId = masterItem ? masterItem.summary_id : null;
      selectedWorkHistoryIds = new Set(
        (activeLensDetail.work_history || []).map((w) => w.work_history_id)
      );
      selectedBulletIds = new Set(
        (activeLensDetail.bullets || []).map((b) => b.bullet_id)
      );
      selectedSkillIds = new Set(
        (activeLensDetail.skills || []).map((s) => s.skill_id)
      );
      selectedAcademicIds = new Set(activeLensDetail.academic_ids || []);
      selectedCertIds = new Set(activeLensDetail.cert_ids || []);
      selectedDescriptorIds = new Set(
        (activeLensDetail.descriptors || []).map((d) => d.descriptor_id)
      );
      selectedCoreExpertiseIds = new Set(
        (activeLensDetail.core_expertise || []).map((ce) => ce.core_expertise_id)
      );

      // Snapshot the saved state for dirty tracking.
      savedSummaryIds = new Set(selectedSummaryIds);
      savedMasterSummaryId = masterSummaryId;
      savedWorkHistoryIds = new Set(selectedWorkHistoryIds);
      savedBulletIds = new Set(selectedBulletIds);
      savedSkillIds = new Set(selectedSkillIds);
      savedAcademicIds = new Set(selectedAcademicIds);
      savedCertIds = new Set(selectedCertIds);
      savedDescriptorIds = new Set(selectedDescriptorIds);
      savedCoreExpertiseIds = new Set(selectedCoreExpertiseIds);
    } catch {
      activeLensId = null;
      activeLensDetail = null;
    } finally {
      loadingDetail = false;
    }
  }

  // --- Selection Toggles & Save ---

  function toggleWorkHistory(id: number): void {
    const next = new Set(selectedWorkHistoryIds);
    if (next.has(id)) {
      next.delete(id);
      // Also remove bullets for this entry
      const entry = workHistory.find((e) => e.id === id);
      if (entry) {
        for (const b of entry.bullets || []) {
          selectedBulletIds.delete(b.id);
        }
        selectedBulletIds = new Set(selectedBulletIds);
      }
    } else {
      next.add(id);
      // Select all bullets for this entry by default
      const entry = workHistory.find((e) => e.id === id);
      if (entry) {
        for (const b of entry.bullets || []) {
          selectedBulletIds.add(b.id);
        }
        selectedBulletIds = new Set(selectedBulletIds);
      }
    }
    selectedWorkHistoryIds = next;
  }

  function toggleBullet(bulletId: number): void {
    const next = new Set(selectedBulletIds);
    if (next.has(bulletId)) {
      next.delete(bulletId);
    } else {
      next.add(bulletId);
    }
    selectedBulletIds = next;
  }

  function toggleSkill(id: number): void {
    const next = new Set(selectedSkillIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedSkillIds = next;
  }

  function toggleAcademic(id: number): void {
    const next = new Set(selectedAcademicIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedAcademicIds = next;
  }

  function toggleCert(id: number): void {
    const next = new Set(selectedCertIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedCertIds = next;
  }

  function toggleDescriptor(id: number): void {
    const next = new Set(selectedDescriptorIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedDescriptorIds = next;
  }

  function toggleCoreExpertise(id: number): void {
    const next = new Set(selectedCoreExpertiseIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedCoreExpertiseIds = next;
  }

  function toggleSummary(id: number): void {
    const next = new Set(selectedSummaryIds);
    if (next.has(id)) {
      next.delete(id);
      // Clear master if deselected.
      if (masterSummaryId === id) {
        masterSummaryId = null;
      }
    } else {
      next.add(id);
    }
    selectedSummaryIds = next;
  }

  // --- Save All & Discard ---

  async function saveAllSelections(): Promise<void> {
    if (!activeLensId) return;
    saving = true;
    try {
      const whSelections: LensWorkHistoryItem[] = [
        ...selectedWorkHistoryIds,
      ].map((id, i) => ({ work_history_id: id, sort_order: i }));

      const bulletSelections: LensBulletItem[] = [...selectedBulletIds].map(
        (id, i) => ({ bullet_id: id, sort_order: i })
      );

      const skillSelections: LensSkillItem[] = [...selectedSkillIds].map(
        (id) => ({ skill_id: id, custom_sort_order: null })
      );

      const descSelections: LensDescriptorItem[] = [
        ...selectedDescriptorIds,
      ].map((id, i) => ({ descriptor_id: id, sort_order: i }));

      const ceSelections: LensCoreExpertiseItem[] = [
        ...selectedCoreExpertiseIds,
      ].map((id, i) => ({ core_expertise_id: id, sort_order: i }));

      const summarySelections: LensSummaryItem[] = [...selectedSummaryIds].map(
        (id, i) => ({ summary_id: id, sort_order: i, is_master: id === masterSummaryId })
      );

      await Promise.all([
        setLensWorkHistory(activeLensId, whSelections),
        setLensBullets(activeLensId, bulletSelections),
        setLensSkills(activeLensId, skillSelections),
        setLensAcademics(activeLensId, [...selectedAcademicIds]),
        setLensCerts(activeLensId, [...selectedCertIds]),
        setLensDescriptors(activeLensId, descSelections),
        setLensCoreExpertise(activeLensId, ceSelections),
        setLensSummaries(activeLensId, summarySelections),
      ]);

      // Update snapshots to reflect saved state.
      savedWorkHistoryIds = new Set(selectedWorkHistoryIds);
      savedBulletIds = new Set(selectedBulletIds);
      savedSkillIds = new Set(selectedSkillIds);
      savedAcademicIds = new Set(selectedAcademicIds);
      savedCertIds = new Set(selectedCertIds);
      savedDescriptorIds = new Set(selectedDescriptorIds);
      savedCoreExpertiseIds = new Set(selectedCoreExpertiseIds);
      savedSummaryIds = new Set(selectedSummaryIds);
      savedMasterSummaryId = masterSummaryId;

      addToast("success", "Lens selections saved");
      editMode = false;
    } catch {
      // Toast already shown by API layer
    } finally {
      saving = false;
    }
  }

  function discardChanges(): void {
    selectedWorkHistoryIds = new Set(savedWorkHistoryIds);
    selectedBulletIds = new Set(savedBulletIds);
    selectedSkillIds = new Set(savedSkillIds);
    selectedAcademicIds = new Set(savedAcademicIds);
    selectedCertIds = new Set(savedCertIds);
    selectedDescriptorIds = new Set(savedDescriptorIds);
    selectedCoreExpertiseIds = new Set(savedCoreExpertiseIds);
    selectedSummaryIds = new Set(savedSummaryIds);
    masterSummaryId = savedMasterSummaryId;
    editMode = false;
  }

  // --- Select All / Deselect All ---

  function selectAllWorkHistory(): void {
    const next = new Set<number>();
    const bulletNext = new Set<number>();
    for (const e of workHistory) {
      next.add(e.id);
      for (const b of e.bullets || []) {
        bulletNext.add(b.id);
      }
    }
    selectedWorkHistoryIds = next;
    selectedBulletIds = bulletNext;
  }

  function deselectAllWorkHistory(): void {
    selectedWorkHistoryIds = new Set();
    selectedBulletIds = new Set();
  }

  function selectAllSkills(): void {
    const next = new Set<number>();
    for (const cat of skillCategories) {
      for (const s of cat.skills) {
        next.add(s.id);
      }
    }
    selectedSkillIds = next;
  }

  function deselectAllSkills(): void {
    selectedSkillIds = new Set();
  }

  function selectAllAcademics(): void {
    selectedAcademicIds = new Set(academics.map((a) => a.id));
  }

  function deselectAllAcademics(): void {
    selectedAcademicIds = new Set();
  }

  function selectAllCerts(): void {
    selectedCertIds = new Set(certs.map((c) => c.id));
  }

  function deselectAllCerts(): void {
    selectedCertIds = new Set();
  }

  function selectAllDescriptors(): void {
    selectedDescriptorIds = new Set(descriptors.map((d) => d.id));
  }

  function deselectAllDescriptors(): void {
    selectedDescriptorIds = new Set();
  }

  function selectAllCoreExpertise(): void {
    selectedCoreExpertiseIds = new Set(coreExpertise.map((ce) => ce.id));
  }

  function deselectAllCoreExpertise(): void {
    selectedCoreExpertiseIds = new Set();
  }

  function selectAllSummaries(): void {
    selectedSummaryIds = new Set(summaries.map((s) => s.id));
  }

  function deselectAllSummaries(): void {
    selectedSummaryIds = new Set();
    masterSummaryId = null;
  }

  function toggleAllSkillsInCategory(
    skills: { id: number }[],
    allSelected: boolean
  ): void {
    const next = new Set(selectedSkillIds);
    for (const s of skills) {
      if (allSelected) {
        next.delete(s.id);
      } else {
        next.add(s.id);
      }
    }
    selectedSkillIds = next;
  }
</script>

<div class="lenses-page">
  {#if !activeLensId}
    <!-- ======== STATE 1: Lens List ======== -->
    <div class="page-header">
      <h2>Lenses</h2>
      <button class="btn btn-primary" on:click={openAddLens}>+ New Lens</button>
    </div>
    <p class="page-description">
      Lenses are named content selections for different job types. Select a lens
      to view its content.
    </p>

    <!-- Lens Form (create/rename) -->
    {#if showLensForm}
      <div class="entry-form">
        <h3 class="form-title">
          {editingLens ? "Rename Lens" : "New Lens"}
        </h3>
        <div class="form-row">
          <div class="form-field">
            <label class="form-label" for="lens-name">Name</label>
            <input
              id="lens-name"
              type="text"
              class="form-input"
              bind:value={formName}
              placeholder="e.g. Backend Engineer"
            />
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" on:click={handleLensSubmit}>
            {editingLens ? "Update" : "Create"}
          </button>
          <button class="btn btn-cancel" on:click={cancelLensForm}>
            Cancel
          </button>
        </div>
      </div>
    {/if}

    <!-- Delete Confirmation -->
    {#if deleteConfirmLens}
      <div class="confirm-dialog">
        <p>
          Delete lens <strong>{deleteConfirmLens.name}</strong>? This will remove
          all content selections for this lens.
        </p>
        <div class="form-actions">
          <button class="btn btn-danger-solid" on:click={handleDeleteLens}>
            Delete
          </button>
          <button class="btn btn-cancel" on:click={cancelDelete}>Cancel</button>
        </div>
      </div>
    {/if}

    {#if loading}
      <LoadingSpinner />
    {:else if lenses.length === 0}
      <div class="empty-state">
        <p>No lenses yet.</p>
        <p class="empty-hint">
          Create a lens to define content selections for a job type.
        </p>
      </div>
    {:else}
      <div class="lens-list">
        {#each lenses as lens (lens.id)}
          <div
            class="lens-card"
            role="button"
            tabindex="0"
            on:click={() => selectLens(lens.id)}
            on:keypress={(e) => {
              if (e.key === "Enter") selectLens(lens.id);
            }}
          >
            <div class="lens-info">
              <span class="lens-name">{lens.name}</span>
            </div>
            <div
              class="lens-actions"
              on:click|stopPropagation={() => {}}
              on:keypress|stopPropagation={() => {}}
              role="group"
            >
              <button
                class="btn btn-small btn-ghost"
                on:click|stopPropagation={() => openEditLens(lens)}
              >
                Rename
              </button>
              <button
                class="btn btn-small btn-danger"
                on:click|stopPropagation={() => confirmDeleteLens(lens)}
              >
                Delete
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {:else}
    <!-- ======== STATE 2/3: Single Lens Detail ======== -->

    <!-- Delete Confirmation (in detail view) -->
    {#if deleteConfirmLens}
      <div class="confirm-dialog">
        <p>
          Delete lens <strong>{deleteConfirmLens.name}</strong>? This will remove
          all content selections for this lens.
        </p>
        <div class="form-actions">
          <button class="btn btn-danger-solid" on:click={handleDeleteLens}>
            Delete
          </button>
          <button class="btn btn-cancel" on:click={cancelDelete}>Cancel</button>
        </div>
      </div>
    {/if}

    {#if loadingDetail}
      <LoadingSpinner message="Loading lens details..." />
    {:else}
      {@const activeLens = lenses.find((l) => l.id === activeLensId)}
      <div class="detail-panel">
        <!-- Detail Header -->
        <div class="detail-header">
          <button class="btn-back" on:click={goBackToList}>
            &#8592; Back to Lenses
          </button>
          <div class="detail-header-row">
            <h2 class="detail-lens-name">{activeLens?.name}</h2>
            <div class="detail-header-actions">
              <!-- Rename Form (inline) -->
              {#if showLensForm && editingLens}
                <div class="inline-rename">
                  <input
                    type="text"
                    class="form-input form-input-inline"
                    bind:value={formName}
                    placeholder="Lens name"
                  />
                  <button class="btn btn-small btn-primary" on:click={handleLensSubmit}>
                    Save
                  </button>
                  <button class="btn btn-small btn-cancel" on:click={cancelLensForm}>
                    Cancel
                  </button>
                </div>
              {:else}
                <button
                  class="btn btn-small btn-ghost"
                  on:click={() => { if (activeLens) openEditLens(activeLens); }}
                >
                  Rename
                </button>
              {/if}
              {#if !editMode}
                <button
                  class="btn btn-small btn-edit"
                  on:click={enterEditMode}
                >
                  Edit Selections
                </button>
              {/if}
              <button
                class="btn btn-small btn-danger"
                on:click={() => { if (activeLens) confirmDeleteLens(activeLens); }}
              >
                Delete
              </button>
            </div>
          </div>
        </div>

        {#if editMode}
          <!-- ======== EDIT MODE: Checkboxes ======== -->
          <div class="content-selections">
            <!-- Descriptors -->
            {#if descriptors.length > 0}
              <section class="selection-section">
                <div class="section-header">
                  <h4 class="section-title">Role Descriptors</h4>
                  <div class="section-header-actions">
                    <button class="btn btn-tiny" on:click={selectAllDescriptors}>All</button>
                    <button class="btn btn-tiny" on:click={deselectAllDescriptors}>None</button>
                  </div>
                </div>
                <div class="check-list">
                  {#each descriptors as desc (desc.id)}
                    <label class="checkbox-item">
                      <input
                        type="checkbox"
                        checked={selectedDescriptorIds.has(desc.id)}
                        on:change={() => toggleDescriptor(desc.id)}
                      />
                      <span class="check-label">{desc.title}</span>
                    </label>
                  {/each}
                </div>
              </section>
            {/if}

            <!-- Core Expertise -->
            {#if coreExpertise.length > 0}
              <section class="selection-section">
                <div class="section-header">
                  <h4 class="section-title">Core Expertise</h4>
                  <div class="section-header-actions">
                    <button class="btn btn-tiny" on:click={selectAllCoreExpertise}>All</button>
                    <button class="btn btn-tiny" on:click={deselectAllCoreExpertise}>None</button>
                  </div>
                </div>
                <div class="check-list">
                  {#each coreExpertise as ce (ce.id)}
                    <label class="checkbox-item">
                      <input
                        type="checkbox"
                        checked={selectedCoreExpertiseIds.has(ce.id)}
                        on:change={() => toggleCoreExpertise(ce.id)}
                      />
                      <span class="check-label">{ce.label}</span>
                    </label>
                  {/each}
                </div>
              </section>
            {/if}

            <!-- Summaries -->
            {#if summaries.length > 0}
              <section class="selection-section">
                <div class="section-header">
                  <h4 class="section-title">Professional Summaries</h4>
                  <div class="section-header-actions">
                    <button class="btn btn-tiny" on:click={selectAllSummaries}>All</button>
                    <button class="btn btn-tiny" on:click={deselectAllSummaries}>None</button>
                  </div>
                </div>
                {#if selectedSummaryIds.size > 0 && masterSummaryId === null}
                  <div class="master-warning">
                    One summary should be the lead paragraph. Click a "Bullet" pill to promote it.
                  </div>
                {/if}
                <div class="check-list">
                  {#each summaries as summary (summary.id)}
                    <label class="checkbox-item">
                      <input
                        type="checkbox"
                        checked={selectedSummaryIds.has(summary.id)}
                        on:change={() => toggleSummary(summary.id)}
                      />
                      <span class="check-label">{summary.label}</span>
                      {#if selectedSummaryIds.has(summary.id)}
                        <button
                          class="summary-pill"
                          class:master={masterSummaryId === summary.id}
                          title={masterSummaryId === summary.id ? "Master summary (renders as lead paragraph)" : "Click to set as master summary"}
                          on:click|stopPropagation={() => {
                            masterSummaryId = summary.id;
                          }}
                        >
                          {masterSummaryId === summary.id ? "Master" : "Bullet"}
                        </button>
                      {/if}
                    </label>
                  {/each}
                </div>
              </section>
            {/if}

            <!-- Work History -->
            {#if workHistory.length > 0}
              <section class="selection-section">
                <div class="section-header">
                  <h4 class="section-title">Work History</h4>
                  <div class="section-header-actions">
                    <button class="btn btn-tiny" on:click={selectAllWorkHistory}>All</button>
                    <button class="btn btn-tiny" on:click={deselectAllWorkHistory}>None</button>
                  </div>
                </div>
                <div class="check-list">
                  {#each workHistory as entry (entry.id)}
                    {@const primaryBullets = (entry.bullets || []).filter((b) => b.bullet_type === "primary")}
                    {@const secondaryBullets = (entry.bullets || []).filter((b) => b.bullet_type === "secondary")}
                    <div class="work-history-item">
                      <label class="checkbox-item">
                        <input
                          type="checkbox"
                          checked={selectedWorkHistoryIds.has(entry.id)}
                          on:change={() => toggleWorkHistory(entry.id)}
                        />
                        <span class="check-label">
                          <strong>{entry.job_title}</strong> at {entry.employer_name}
                        </span>
                      </label>
                      {#if selectedWorkHistoryIds.has(entry.id)}
                        {#if primaryBullets.length > 0}
                          <div class="bullet-group">
                            <span class="bullet-group-label">Bullets</span>
                            <div class="bullet-check-list">
                              {#each primaryBullets as bullet (bullet.id)}
                                <label class="checkbox-item bullet-item">
                                  <input
                                    type="checkbox"
                                    checked={selectedBulletIds.has(bullet.id)}
                                    on:change={() => toggleBullet(bullet.id)}
                                  />
                                  <span class="check-label check-label-small">
                                    {bullet.text}
                                  </span>
                                </label>
                              {/each}
                            </div>
                          </div>
                        {/if}
                        {#if secondaryBullets.length > 0}
                          <div class="bullet-group">
                            <span class="bullet-group-label outcome-label">Outcomes</span>
                            <div class="bullet-check-list">
                              {#each secondaryBullets as bullet (bullet.id)}
                                <label class="checkbox-item bullet-item">
                                  <input
                                    type="checkbox"
                                    checked={selectedBulletIds.has(bullet.id)}
                                    on:change={() => toggleBullet(bullet.id)}
                                  />
                                  <span class="check-label check-label-small outcome-text">
                                    {bullet.text}
                                  </span>
                                </label>
                              {/each}
                            </div>
                          </div>
                        {/if}
                      {/if}
                    </div>
                  {/each}
                </div>
              </section>
            {/if}

            <!-- Skills by Category -->
            {#if skillCategories.length > 0}
              <section class="selection-section">
                <div class="section-header">
                  <h4 class="section-title">Skills</h4>
                  <div class="section-header-actions">
                    <button class="btn btn-tiny" on:click={selectAllSkills}>All</button>
                    <button class="btn btn-tiny" on:click={deselectAllSkills}>None</button>
                  </div>
                </div>
                <div class="skills-layout">
                  <div class="skills-chips-area">
                    {#each skillCategories as cat (cat.category.id)}
                      {#each cat.skills as skill (skill.id)}
                        <button
                          class="skill-chip"
                          class:selected={selectedSkillIds.has(skill.id)}
                          class:legacy={skill.is_legacy}
                          on:click={() => toggleSkill(skill.id)}
                        >
                          {skill.name}
                          {#if skill.is_legacy}<span class="legacy-badge">L</span>{/if}
                        </button>
                      {/each}
                    {/each}
                  </div>
                  <div class="skills-categories-pane" class:collapsed={!skillCategoriesExpanded}>
                    <button
                      class="categories-toggle"
                      on:click={() => (skillCategoriesExpanded = !skillCategoriesExpanded)}
                      title={skillCategoriesExpanded ? "Collapse categories" : "Expand categories"}
                    >
                      {skillCategoriesExpanded ? "Categories \u25BC" : "\u25B6 Cat."}
                    </button>
                    {#if skillCategoriesExpanded}
                      <div class="categories-list">
                        {#each skillCategories as cat (cat.category.id)}
                          {@const allSelected = cat.skills.every((s) =>
                            selectedSkillIds.has(s.id)
                          )}
                          <label class="checkbox-item category-toggle">
                            <input
                              type="checkbox"
                              checked={allSelected}
                              on:change={() =>
                                toggleAllSkillsInCategory(cat.skills, allSelected)}
                            />
                            <span class="check-label category-label">
                              {cat.category.name}
                            </span>
                          </label>
                        {/each}
                      </div>
                    {/if}
                  </div>
                </div>
              </section>
            {/if}

            <!-- Education -->
            {#if academics.length > 0}
              <section class="selection-section">
                <div class="section-header">
                  <h4 class="section-title">Education</h4>
                  <div class="section-header-actions">
                    <button class="btn btn-tiny" on:click={selectAllAcademics}>All</button>
                    <button class="btn btn-tiny" on:click={deselectAllAcademics}>None</button>
                  </div>
                </div>
                <div class="check-list">
                  {#each academics as acad (acad.id)}
                    <label class="checkbox-item">
                      <input
                        type="checkbox"
                        checked={selectedAcademicIds.has(acad.id)}
                        on:change={() => toggleAcademic(acad.id)}
                      />
                      <span class="check-label">
                        {acad.credential_type ? `${acad.credential_type} in ` : ""}{acad.field_of_study},
                        {acad.institution}
                      </span>
                    </label>
                  {/each}
                </div>
              </section>
            {/if}

            <!-- Certifications -->
            {#if certs.length > 0}
              <section class="selection-section">
                <div class="section-header">
                  <h4 class="section-title">Certifications</h4>
                  <div class="section-header-actions">
                    <button class="btn btn-tiny" on:click={selectAllCerts}>All</button>
                    <button class="btn btn-tiny" on:click={deselectAllCerts}>None</button>
                  </div>
                </div>
                <div class="check-list">
                  {#each certs as cert (cert.id)}
                    <label class="checkbox-item">
                      <input
                        type="checkbox"
                        checked={selectedCertIds.has(cert.id)}
                        on:change={() => toggleCert(cert.id)}
                      />
                      <span class="check-label">
                        {cert.name}
                        <span class="cert-issuer">({cert.issuing_body})</span>
                        {#if !cert.is_active}
                          <span class="expired-badge">Expired</span>
                        {/if}
                      </span>
                    </label>
                  {/each}
                </div>
              </section>
            {/if}
          </div>

          <!-- Sticky Save/Discard Bar -->
          <div class="sticky-bar" class:dirty={isDirty}>
            <div class="sticky-bar-content">
              {#if isDirty}
                <span class="dirty-indicator">Unsaved changes</span>
              {/if}
              <div class="sticky-bar-actions">
                <button
                  class="btn btn-cancel"
                  on:click={discardChanges}
                  disabled={saving}
                >
                  {isDirty ? "Discard" : "Cancel"}
                </button>
                <button
                  class="btn btn-primary"
                  on:click={saveAllSelections}
                  disabled={!isDirty || saving}
                >
                  {saving ? "Saving..." : "Save All"}
                </button>
              </div>
            </div>
          </div>

        {:else}
          <!-- ======== READ-ONLY VIEW ======== -->
          <div class="readonly-content">
            <!-- Descriptors -->
            {#if descriptors.length > 0}
              <section class="readonly-section">
                <h4 class="section-title">Role Descriptors</h4>
                {#if selectedDescriptorIds.size > 0}
                  <ul class="readonly-list">
                    {#each descriptors.filter((d) => selectedDescriptorIds.has(d.id)) as desc (desc.id)}
                      <li>{desc.title}</li>
                    {/each}
                  </ul>
                {:else}
                  <p class="none-selected">None selected</p>
                {/if}
              </section>
            {/if}

            <!-- Core Expertise -->
            {#if coreExpertise.length > 0}
              <section class="readonly-section">
                <h4 class="section-title">Core Expertise</h4>
                {#if selectedCoreExpertiseIds.size > 0}
                  <ul class="readonly-list">
                    {#each coreExpertise.filter((ce) => selectedCoreExpertiseIds.has(ce.id)) as ce (ce.id)}
                      <li>{ce.label}</li>
                    {/each}
                  </ul>
                {:else}
                  <p class="none-selected">None selected</p>
                {/if}
              </section>
            {/if}

            <!-- Summaries -->
            {#if summaries.length > 0}
              <section class="readonly-section">
                <h4 class="section-title">Professional Summaries</h4>
                {#if selectedSummaryIds.size > 0}
                  <ul class="readonly-list">
                    {#each summaries.filter((s) => selectedSummaryIds.has(s.id)) as summary (summary.id)}
                      <li>
                        {summary.label}
                        <span class="summary-pill-readonly" class:master={masterSummaryId === summary.id}>
                          {masterSummaryId === summary.id ? "Master" : "Bullet"}
                        </span>
                      </li>
                    {/each}
                  </ul>
                {:else}
                  <p class="none-selected">None selected</p>
                {/if}
              </section>
            {/if}

            <!-- Work History -->
            {#if workHistory.length > 0}
              <section class="readonly-section">
                <h4 class="section-title">Work History</h4>
                {#if selectedWorkHistoryIds.size > 0}
                  <div class="readonly-work-history">
                    {#each workHistory.filter((e) => selectedWorkHistoryIds.has(e.id)) as entry (entry.id)}
                      {@const selectedPrimary = (entry.bullets || []).filter((b) => selectedBulletIds.has(b.id) && b.bullet_type === "primary")}
                      {@const selectedSecondary = (entry.bullets || []).filter((b) => selectedBulletIds.has(b.id) && b.bullet_type === "secondary")}
                      <div class="readonly-wh-entry">
                        <div class="readonly-wh-header">
                          <strong>{entry.job_title}</strong> at {entry.employer_name}
                        </div>
                        {#if selectedPrimary.length > 0}
                          <ul class="readonly-bullet-list">
                            {#each selectedPrimary as bullet (bullet.id)}
                              <li>{bullet.text}</li>
                            {/each}
                          </ul>
                        {/if}
                        {#if selectedSecondary.length > 0}
                          <ul class="readonly-bullet-list readonly-outcome-list">
                            {#each selectedSecondary as bullet (bullet.id)}
                              <li>{bullet.text}</li>
                            {/each}
                          </ul>
                        {/if}
                      </div>
                    {/each}
                  </div>
                {:else}
                  <p class="none-selected">None selected</p>
                {/if}
              </section>
            {/if}

            <!-- Skills -->
            {#if skillCategories.length > 0}
              <section class="readonly-section">
                <h4 class="section-title">Skills</h4>
                {#if selectedSkillIds.size > 0}
                  <div class="readonly-skills">
                    {#each skillCategories as cat (cat.category.id)}
                      {@const selectedInCat = cat.skills.filter((s) => selectedSkillIds.has(s.id))}
                      {#if selectedInCat.length > 0}
                        <div class="readonly-skill-cat">
                          <span class="readonly-cat-name">{cat.category.name}:</span>
                          <span class="readonly-skill-names">
                            {selectedInCat.map((s) => s.name).join(", ")}
                          </span>
                        </div>
                      {/if}
                    {/each}
                  </div>
                {:else}
                  <p class="none-selected">None selected</p>
                {/if}
              </section>
            {/if}

            <!-- Education -->
            {#if academics.length > 0}
              <section class="readonly-section">
                <h4 class="section-title">Education</h4>
                {#if selectedAcademicIds.size > 0}
                  <ul class="readonly-list">
                    {#each academics.filter((a) => selectedAcademicIds.has(a.id)) as acad (acad.id)}
                      <li>
                        {acad.credential_type ? `${acad.credential_type} in ` : ""}{acad.field_of_study},
                        {acad.institution}
                      </li>
                    {/each}
                  </ul>
                {:else}
                  <p class="none-selected">None selected</p>
                {/if}
              </section>
            {/if}

            <!-- Certifications -->
            {#if certs.length > 0}
              <section class="readonly-section">
                <h4 class="section-title">Certifications</h4>
                {#if selectedCertIds.size > 0}
                  <ul class="readonly-list">
                    {#each certs.filter((c) => selectedCertIds.has(c.id)) as cert (cert.id)}
                      <li>
                        {cert.name}
                        <span class="cert-issuer">({cert.issuing_body})</span>
                        {#if !cert.is_active}
                          <span class="expired-badge">Expired</span>
                        {/if}
                      </li>
                    {/each}
                  </ul>
                {:else}
                  <p class="none-selected">None selected</p>
                {/if}
              </section>
            {/if}
          </div>
        {/if}
      </div>
    {/if}
  {/if}
</div>

<style>
  .lenses-page {
    max-width: 900px;
    padding-bottom: 24px;
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
    margin: 0 0 24px;
  }

  .empty-state {
    text-align: center;
    padding: 48px 0;
    color: #5a6a7a;
  }

  .empty-hint {
    font-size: 0.85rem;
    color: #5a6a7a;
    margin-top: 8px;
  }

  /* --- Lens List --- */
  .lens-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .lens-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background-color: #1e2d3d;
    border: 2px solid #2a3a4a;
    border-radius: 6px;
    cursor: pointer;
    transition:
      border-color 0.15s,
      background-color 0.15s;
  }

  .lens-card:hover {
    border-color: #4a8af4;
    background-color: #223344;
  }

  .lens-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .lens-name {
    font-size: 0.95rem;
    font-weight: 600;
    color: #e0e0e0;
  }

  .lens-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  /* --- Detail Panel --- */
  .detail-panel {
    background-color: #1e2d3d;
    border: 2px solid #2a3a4a;
    border-radius: 6px;
    padding: 20px;
  }

  .detail-header {
    margin-bottom: 20px;
  }

  .btn-back {
    background: none;
    border: none;
    color: #4a8af4;
    font-size: 0.85rem;
    cursor: pointer;
    padding: 0;
    margin-bottom: 12px;
  }

  .btn-back:hover {
    color: #6aa0f8;
    text-decoration: underline;
  }

  .detail-header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .detail-lens-name {
    margin: 0;
    font-size: 1.3rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .detail-header-actions {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  .inline-rename {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .form-input-inline {
    width: 180px;
    padding: 4px 8px;
    font-size: 0.85rem;
  }

  /* --- Content Selections (edit mode) --- */
  .content-selections {
    margin-top: 0;
  }

  /* --- Read-Only Content --- */
  .readonly-content {
    margin-top: 0;
  }

  .readonly-section {
    border-top: 1px solid #2a3a4a;
    padding: 12px 0;
  }

  .readonly-section:first-child {
    border-top: none;
    padding-top: 0;
  }

  .readonly-list {
    margin: 6px 0 0 0;
    padding-left: 20px;
    list-style-type: disc;
  }

  .readonly-list li {
    font-size: 0.85rem;
    color: #c0d0e0;
    line-height: 1.5;
    padding: 1px 0;
  }

  .readonly-work-history {
    margin-top: 6px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .readonly-wh-entry {
    padding-left: 4px;
  }

  .readonly-wh-header {
    font-size: 0.85rem;
    color: #c0d0e0;
    line-height: 1.4;
  }

  .readonly-bullet-list {
    margin: 4px 0 0 0;
    padding-left: 24px;
    list-style-type: disc;
  }

  .readonly-bullet-list li {
    font-size: 0.8rem;
    color: #a0b0c0;
    line-height: 1.4;
    padding: 1px 0;
  }

  .readonly-skills {
    margin-top: 6px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .readonly-skill-cat {
    font-size: 0.85rem;
    color: #c0d0e0;
    line-height: 1.4;
  }

  .readonly-cat-name {
    font-weight: 600;
    color: #7a8a9a;
    text-transform: uppercase;
    font-size: 0.8rem;
    letter-spacing: 0.03em;
  }

  .readonly-skill-names {
    color: #c0d0e0;
  }

  .none-selected {
    color: #5a6a7a;
    font-size: 0.85rem;
    font-style: italic;
    margin: 6px 0 0;
  }

  /* --- Sections (edit mode) --- */
  .selection-section {
    background-color: #1a2636;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .selection-section:first-child {
    margin-top: 0;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .section-header-actions {
    display: flex;
    gap: 4px;
  }

  .section-title {
    margin: 0;
    font-size: 0.95rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  /* --- Forms --- */
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
    margin-top: 8px;
  }

  /* --- Checkboxes --- */
  .check-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .checkbox-item {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    cursor: pointer;
    padding: 4px 0;
  }

  .checkbox-item input[type="checkbox"] {
    margin-top: 2px;
    accent-color: #4a8af4;
    flex-shrink: 0;
  }

  .check-label {
    font-size: 0.85rem;
    color: #c0d0e0;
    line-height: 1.4;
  }

  .check-label-small {
    font-size: 0.8rem;
    color: #a0b0c0;
  }

  /* --- Work History Bullets --- */
  .work-history-item {
    border-bottom: 1px solid #2a3a4a;
    padding-bottom: 8px;
    margin-bottom: 4px;
  }

  .work-history-item:last-child {
    border-bottom: none;
    padding-bottom: 0;
    margin-bottom: 0;
  }

  .bullet-check-list {
    margin-left: 28px;
    margin-top: 4px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .bullet-item {
    padding: 2px 0;
  }

  /* --- Skill Chips --- */
  .category-toggle {
    margin-bottom: 2px;
  }

  .category-label {
    font-weight: 600;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    color: #7a8a9a;
  }

  .skill-chip {
    padding: 4px 10px;
    font-size: 0.8rem;
    border: 1px solid #2a3a4a;
    border-radius: 14px;
    background-color: #1a2332;
    color: #a0b0c0;
    cursor: pointer;
    transition:
      background-color 0.15s,
      border-color 0.15s;
  }

  .skill-chip:hover {
    border-color: #4a8af4;
  }

  .skill-chip.selected {
    background-color: #1e3555;
    border-color: #4a8af4;
    color: #e0e0e0;
  }

  .skill-chip.legacy {
    font-style: italic;
  }

  .legacy-badge {
    font-size: 0.65rem;
    color: #7a8a9a;
    margin-left: 2px;
    vertical-align: super;
  }

  .cert-issuer {
    color: #7a8a9a;
    font-size: 0.8rem;
  }

  .expired-badge {
    font-size: 0.7rem;
    color: #c05050;
    margin-left: 4px;
    font-weight: 600;
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

  .btn-cancel:hover:not(:disabled) {
    background-color: #2a3a4a;
    color: #c0d0e0;
  }

  .btn-cancel:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

  .btn-edit {
    background-color: transparent;
    border-color: transparent;
    color: #50b060;
    font-weight: 600;
  }

  .btn-edit:hover {
    background-color: #1a3020;
    color: #70d080;
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

  .btn-tiny {
    padding: 2px 8px;
    font-size: 0.75rem;
    background-color: transparent;
    border-color: #3a4a5a;
    color: #7a8a9a;
  }

  .btn-tiny:hover {
    background-color: #2a3a4a;
    color: #c0d0e0;
  }

  /* --- Sticky Save/Discard Bar --- */
  .sticky-bar {
    position: sticky;
    bottom: 0;
    background-color: #1a2636;
    border-top: 2px solid #2a3a4a;
    padding: 12px 16px;
    margin: 0 -20px -20px;
    border-radius: 0 0 6px 6px;
    transition: border-color 0.2s;
  }

  .sticky-bar.dirty {
    border-top-color: #4a8af4;
  }

  .sticky-bar-content {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 16px;
  }

  .dirty-indicator {
    font-size: 0.85rem;
    color: #4a8af4;
    font-weight: 500;
  }

  .sticky-bar-actions {
    display: flex;
    gap: 8px;
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

  /* Master summary controls */
  .summary-pill {
    display: inline-block;
    padding: 1px 10px;
    font-size: 0.7rem;
    font-weight: 600;
    border-radius: 10px;
    border: 1px solid #3a4a5a;
    background-color: #1a2332;
    color: #7a8a9a;
    cursor: pointer;
    margin-left: auto;
    line-height: 1.4;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    transition: background-color 0.15s, border-color 0.15s, color 0.15s;
  }

  .summary-pill:hover {
    border-color: #e0a060;
    color: #e0a060;
  }

  .summary-pill.master {
    background-color: #3d2e1a;
    border-color: #e0a060;
    color: #e0a060;
  }

  .summary-pill-readonly {
    display: inline-block;
    padding: 0 8px;
    font-size: 0.65rem;
    font-weight: 600;
    border-radius: 10px;
    border: 1px solid #3a4a5a;
    background-color: #1a2332;
    color: #7a8a9a;
    margin-left: 6px;
    line-height: 1.6;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    vertical-align: middle;
  }

  .summary-pill-readonly.master {
    background-color: #3d2e1a;
    border-color: #e0a060;
    color: #e0a060;
  }

  .master-warning {
    background-color: #3d2e1a;
    border: 1px solid #705020;
    border-radius: 4px;
    padding: 6px 12px;
    font-size: 0.8rem;
    color: #e0a060;
    margin-bottom: 8px;
  }

  /* --- Bullet/Outcome grouping --- */
  .bullet-group {
    margin-left: 28px;
    margin-top: 6px;
  }

  .bullet-group-label {
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: #4a8af4;
    display: block;
    margin-bottom: 2px;
  }

  .outcome-label {
    color: #a070d0;
  }

  .outcome-text {
    font-style: italic;
    color: #b090d0;
  }

  .readonly-outcome-list {
    list-style-type: "\25B7  ";
  }

  .readonly-outcome-list li {
    font-style: italic;
    color: #b090d0;
  }

  /* --- Skills layout with collapsible categories --- */
  .skills-layout {
    display: flex;
    gap: 12px;
  }

  .skills-chips-area {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    align-content: flex-start;
  }

  .skills-categories-pane {
    width: 180px;
    flex-shrink: 0;
    border-left: 1px solid #2a3a4a;
    padding-left: 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .skills-categories-pane.collapsed {
    width: auto;
    border-left: none;
    padding-left: 0;
  }

  .categories-toggle {
    background: none;
    border: 1px solid #3a4a5a;
    border-radius: 4px;
    color: #7a8a9a;
    font-size: 0.75rem;
    padding: 2px 8px;
    cursor: pointer;
    text-align: left;
    white-space: nowrap;
    transition: background-color 0.15s, color 0.15s;
  }

  .categories-toggle:hover {
    background-color: #2a3a4a;
    color: #c0d0e0;
  }

  .categories-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-top: 4px;
  }
</style>
