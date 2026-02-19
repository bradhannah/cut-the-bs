<script lang="ts">
  import { onMount } from "svelte";
  import {
    listTemplates,
    listWorkHistory,
    listSkillsByCategory,
    listAcademicCredentials,
    listCertifications,
    listDescriptors,
    listSummaries,
    listExports,
    createExport,
    openExportFile,
    addToast,
    type ResumeTemplate,
    type WorkHistoryEntry,
    type SkillCategoryWithSkills,
    type AcademicCredential,
    type Certification,
    type RoleDescriptor,
    type ProfessionalSummary,
    type ResumeExport,
    type ExportRequest,
  } from "../services/api";

  // --- Data ---
  let templates: ResumeTemplate[] = [];
  let workHistory: WorkHistoryEntry[] = [];
  let skillCategories: SkillCategoryWithSkills[] = [];
  let academics: AcademicCredential[] = [];
  let certs: Certification[] = [];
  let descriptors: RoleDescriptor[] = [];
  let summaries: ProfessionalSummary[] = [];
  let exports: ResumeExport[] = [];

  let loading = true;
  let generating = false;

  // --- Selections ---
  let selectedTemplate = "";
  let selectedSummaryId: number | null = null;
  let selectedWorkHistoryIds: Set<number> = new Set();
  let selectedBulletIds: Set<number> = new Set();
  let selectedSkillIds: Set<number> = new Set();
  let selectedAcademicIds: Set<number> = new Set();
  let selectedCertIds: Set<number> = new Set();
  let selectedDescriptorIds: Set<number> = new Set();

  onMount(async () => {
    await loadAllData();
  });

  async function loadAllData(): Promise<void> {
    loading = true;
    try {
      const results = await Promise.all([
        listTemplates(),
        listWorkHistory(),
        listSkillsByCategory(),
        listAcademicCredentials(),
        listCertifications(),
        listDescriptors(),
        listSummaries(),
        listExports(),
      ]);

      templates = results[0];
      workHistory = results[1];
      skillCategories = results[2];
      academics = results[3];
      certs = results[4];
      descriptors = results[5];
      summaries = results[6];
      exports = results[7];

      // Default to first template.
      if (templates.length > 0 && !selectedTemplate) {
        selectedTemplate = templates[0].id;
      }
    } finally {
      loading = false;
    }
  }

  // --- Toggle helpers ---
  function toggleWorkHistory(id: number): void {
    const next = new Set(selectedWorkHistoryIds);
    if (next.has(id)) {
      next.delete(id);
      // Also remove bullets for this entry.
      const entry = workHistory.find((e) => e.id === id);
      if (entry) {
        for (const b of entry.bullets || []) {
          selectedBulletIds.delete(b.id);
        }
        selectedBulletIds = new Set(selectedBulletIds);
      }
    } else {
      next.add(id);
      // Select all bullets for this entry by default.
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

  function selectAllCerts(): void {
    selectedCertIds = new Set(certs.map((c) => c.id));
  }

  function selectAllDescriptors(): void {
    selectedDescriptorIds = new Set(descriptors.map((d) => d.id));
  }

  // --- Computed ---
  $: hasContent =
    selectedWorkHistoryIds.size > 0 ||
    selectedSkillIds.size > 0 ||
    selectedAcademicIds.size > 0 ||
    selectedCertIds.size > 0 ||
    selectedDescriptorIds.size > 0;

  $: canGenerate = selectedTemplate && hasContent && !generating;

  // --- Generate ---
  async function handleGenerate(): Promise<void> {
    if (!canGenerate) {
      addToast("error", "Select a template and at least one content item");
      return;
    }

    generating = true;
    try {
      const req: ExportRequest = {
        template_id: selectedTemplate,
        summary_id: selectedSummaryId,
        work_history_ids: [...selectedWorkHistoryIds],
        bullet_ids: [...selectedBulletIds],
        skill_ids: [...selectedSkillIds],
        skill_sort_overrides: {},
        academic_ids: [...selectedAcademicIds],
        certification_ids: [...selectedCertIds],
        descriptor_ids: [...selectedDescriptorIds],
      };
      await createExport(req);
      exports = await listExports();
    } catch {
      // Toast already shown
    } finally {
      generating = false;
    }
  }

  async function handleOpenExport(exportId: number): Promise<void> {
    try {
      await openExportFile(exportId);
    } catch {
      // Toast already shown
    }
  }

  function formatDate(dateStr: string): string {
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return dateStr;
    }
  }

  function getTemplateName(templateId: string): string {
    const t = templates.find((t) => t.id === templateId);
    return t ? t.name : templateId;
  }
</script>

<div class="export-page">
  <div class="page-header">
    <h2>Export Resume</h2>
    <button
      class="btn btn-primary btn-generate"
      on:click={handleGenerate}
      disabled={!canGenerate}
    >
      {generating ? "Generating..." : "Generate PDF"}
    </button>
  </div>
  <p class="page-description">
    Select a template and the content to include in your resume, then generate a
    PDF.
  </p>

  {#if loading}
    <p class="loading-message">Loading...</p>
  {:else}
    <div class="export-layout">
      <!-- Left: Selections -->
      <div class="selections-panel">
        <!-- Template Selection -->
        <section class="selection-section">
          <h3 class="section-title">Template</h3>
          <div class="template-grid">
            {#each templates as template (template.id)}
              <button
                class="template-card"
                class:selected={selectedTemplate === template.id}
                on:click={() => (selectedTemplate = template.id)}
              >
                <span class="template-name">{template.name}</span>
                <span class="template-desc">{template.description}</span>
              </button>
            {/each}
          </div>
        </section>

        <!-- Summary Selection -->
        {#if summaries.length > 0}
          <section class="selection-section">
            <h3 class="section-title">Professional Summary</h3>
            <div class="summary-select">
              <label class="checkbox-item">
                <input
                  type="radio"
                  name="summary"
                  value={null}
                  bind:group={selectedSummaryId}
                />
                <span class="check-label">None</span>
              </label>
              {#each summaries as summary (summary.id)}
                <label class="checkbox-item">
                  <input
                    type="radio"
                    name="summary"
                    value={summary.id}
                    bind:group={selectedSummaryId}
                  />
                  <span class="check-label">{summary.label}</span>
                </label>
              {/each}
            </div>
          </section>
        {/if}

        <!-- Descriptors -->
        {#if descriptors.length > 0}
          <section class="selection-section">
            <div class="section-header">
              <h3 class="section-title">Role Descriptors</h3>
              <button
                class="btn btn-tiny"
                on:click={selectAllDescriptors}
              >
                All
              </button>
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

        <!-- Work History -->
        {#if workHistory.length > 0}
          <section class="selection-section">
            <div class="section-header">
              <h3 class="section-title">Work History</h3>
              <div class="section-header-actions">
                <button
                  class="btn btn-tiny"
                  on:click={selectAllWorkHistory}
                >
                  All
                </button>
                <button
                  class="btn btn-tiny"
                  on:click={deselectAllWorkHistory}
                >
                  None
                </button>
              </div>
            </div>
            <div class="check-list">
              {#each workHistory as entry (entry.id)}
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
                  {#if selectedWorkHistoryIds.has(entry.id) && (entry.bullets || []).length > 0}
                    <div class="bullet-check-list">
                      {#each entry.bullets || [] as bullet (bullet.id)}
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
              <h3 class="section-title">Skills</h3>
              <div class="section-header-actions">
                <button
                  class="btn btn-tiny"
                  on:click={selectAllSkills}
                >
                  All
                </button>
                <button
                  class="btn btn-tiny"
                  on:click={deselectAllSkills}
                >
                  None
                </button>
              </div>
            </div>
            {#each skillCategories as cat (cat.category.id)}
              {@const allSelected = cat.skills.every((s) =>
                selectedSkillIds.has(s.id)
              )}
              <div class="skill-category-group">
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
                <div class="skill-chips">
                  {#each cat.skills as skill (skill.id)}
                    <button
                      class="skill-chip"
                      class:selected={selectedSkillIds.has(skill.id)}
                      class:legacy={skill.is_legacy}
                      on:click={() => toggleSkill(skill.id)}
                    >
                      {skill.name}
                      {#if skill.is_legacy}<span class="legacy-badge">L</span
                        >{/if}
                    </button>
                  {/each}
                </div>
              </div>
            {/each}
          </section>
        {/if}

        <!-- Education -->
        {#if academics.length > 0}
          <section class="selection-section">
            <div class="section-header">
              <h3 class="section-title">Education</h3>
              <button class="btn btn-tiny" on:click={selectAllAcademics}>
                All
              </button>
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
                    {acad.credential_type} in {acad.field_of_study},
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
              <h3 class="section-title">Certifications</h3>
              <button class="btn btn-tiny" on:click={selectAllCerts}>
                All
              </button>
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

      <!-- Right: Export History -->
      <div class="history-panel">
        <h3 class="section-title">Export History</h3>
        {#if exports.length === 0}
          <p class="empty-hint">No exports yet.</p>
        {:else}
          <div class="export-list">
            {#each exports as exp (exp.id)}
              <button
                class="export-item"
                on:click={() => handleOpenExport(exp.id)}
                title="Click to open PDF"
              >
                <span class="export-template">
                  {getTemplateName(exp.template_id)}
                </span>
                <span class="export-date">
                  {formatDate(exp.generated_at)}
                </span>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .export-page {
    max-width: 1100px;
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

  .loading-message {
    color: #5a6a7a;
    font-style: italic;
  }

  .export-layout {
    display: flex;
    gap: 24px;
  }

  .selections-panel {
    flex: 1;
    min-width: 0;
  }

  .history-panel {
    width: 280px;
    flex-shrink: 0;
  }

  /* --- Sections --- */
  .selection-section {
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    padding: 16px;
    margin-bottom: 16px;
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
    margin: 0 0 8px;
    font-size: 0.95rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .section-header .section-title {
    margin-bottom: 0;
  }

  /* --- Template Grid --- */
  .template-grid {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .template-card {
    flex: 1;
    min-width: 160px;
    padding: 14px;
    background-color: #1a2332;
    border: 2px solid #2a3a4a;
    border-radius: 6px;
    cursor: pointer;
    text-align: left;
    display: flex;
    flex-direction: column;
    gap: 6px;
    transition:
      border-color 0.15s,
      background-color 0.15s;
    color: #a0b0c0;
  }

  .template-card:hover {
    border-color: #4a8af4;
    background-color: #223344;
  }

  .template-card.selected {
    border-color: #4a8af4;
    background-color: #1e3555;
  }

  .template-name {
    font-weight: 600;
    font-size: 0.95rem;
    color: #e0e0e0;
  }

  .template-desc {
    font-size: 0.8rem;
    color: #7a8a9a;
  }

  /* --- Checkboxes --- */
  .check-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .summary-select {
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

  .checkbox-item input[type="checkbox"],
  .checkbox-item input[type="radio"] {
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
  .skill-category-group {
    margin-bottom: 10px;
  }

  .skill-category-group:last-child {
    margin-bottom: 0;
  }

  .category-toggle {
    margin-bottom: 6px;
  }

  .category-label {
    font-weight: 600;
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    color: #7a8a9a;
  }

  .skill-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-left: 24px;
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

  .btn-generate {
    padding: 10px 24px;
    font-size: 0.95rem;
    font-weight: 600;
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

  /* --- Export History --- */
  .empty-hint {
    color: #5a6a7a;
    font-size: 0.85rem;
    font-style: italic;
  }

  .export-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .export-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 10px 12px;
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    cursor: pointer;
    text-align: left;
    transition:
      background-color 0.15s,
      border-color 0.15s;
    color: #c0d0e0;
  }

  .export-item:hover {
    background-color: #223344;
    border-color: #4a8af4;
  }

  .export-template {
    font-weight: 600;
    font-size: 0.85rem;
    color: #e0e0e0;
  }

  .export-date {
    font-size: 0.75rem;
    color: #7a8a9a;
  }
</style>
