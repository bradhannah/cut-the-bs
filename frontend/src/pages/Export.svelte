<script lang="ts">
  import { onMount } from "svelte";
  import {
    listDocumentTemplates,
    listWorkHistory,
    listSkillsByCategory,
    listAcademicCredentials,
    listCertifications,
    listDescriptors,
    listSummaries,
    listExports,
    listLenses,
    listCoreExpertise,
    getLensExportSelections,
    createExport,
    openExportFile,
    parseTemplateVariables,
    addToast,
    type DocumentTemplate,
    type WorkHistoryEntry,
    type SkillCategoryWithSkills,
    type AcademicCredential,
    type Certification,
    type RoleDescriptor,
    type ProfessionalSummary,
    type CoreExpertise,
    type ResumeExport,
    type ExportRequest,
    type Lens,
    type TemplateVariable,
    type GuidedPrompt,
  } from "../services/api";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";
  import PromptDialog from "../components/coverletter/PromptDialog.svelte";
  import { formatTimestamp } from "../services/dateFormat";

  // --- Data ---
  let docTemplates: DocumentTemplate[] = [];
  let workHistory: WorkHistoryEntry[] = [];
  let skillCategories: SkillCategoryWithSkills[] = [];
  let academics: AcademicCredential[] = [];
  let certs: Certification[] = [];
  let descriptors: RoleDescriptor[] = [];
  let coreExpertise: CoreExpertise[] = [];
  let summaries: ProfessionalSummary[] = [];
  let exports: ResumeExport[] = [];
  let lenses: Lens[] = [];

  let loading = true;
  let generating = false;
  let skillCategoriesExpanded = true;

  // --- Selections ---
  let selectedTemplateId: number | null = null;
  let selectedLensId: number | null = null;
  let selectedSummaryIds: Set<number> = new Set();
  let masterSummaryId: number | null = null;
  let selectedWorkHistoryIds: Set<number> = new Set();
  let selectedBulletIds: Set<number> = new Set();
  let selectedSkillIds: Set<number> = new Set();
  let selectedAcademicIds: Set<number> = new Set();
  let selectedCertIds: Set<number> = new Set();
  let selectedDescriptorIds: Set<number> = new Set();
  let selectedCoreExpertiseIds: Set<number> = new Set();

  // --- Cover letter prompt dialog ---
  let showPromptDialog = false;
  let promptVariables: TemplateVariable[] = [];
  let promptPrompts: GuidedPrompt[] = [];
  let promptPrefilled: Record<string, string> = {};

  // --- Derived ---
  $: resumeTemplates = docTemplates.filter(
    (t) => t.template_type === "resume"
  );
  $: coverLetterTemplates = docTemplates.filter(
    (t) => t.template_type === "cover_letter"
  );
  $: selectedTemplate = docTemplates.find(
    (t) => t.id === selectedTemplateId
  ) || null;
  $: isCoverLetter = selectedTemplate?.template_type === "cover_letter";

  onMount(async () => {
    await loadAllData();
  });

  async function loadAllData(): Promise<void> {
    loading = true;
    try {
      const results = await Promise.all([
        listDocumentTemplates(),
        listWorkHistory(),
        listSkillsByCategory(),
        listAcademicCredentials(),
        listCertifications(),
        listDescriptors(),
        listSummaries(),
        listExports(),
        listLenses(),
        listCoreExpertise(),
      ]);

      docTemplates = results[0] || [];
      workHistory = results[1] || [];
      skillCategories = results[2] || [];
      academics = results[3] || [];
      certs = results[4] || [];
      descriptors = results[5] || [];
      summaries = results[6] || [];
      exports = results[7] || [];
      lenses = results[8] || [];
      coreExpertise = results[9] || [];

      // Default to first resume template.
      if (docTemplates.length > 0 && selectedTemplateId === null) {
        const firstResume = docTemplates.find(
          (t) => t.template_type === "resume"
        );
        selectedTemplateId = firstResume ? firstResume.id : docTemplates[0].id;
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

  function toggleCoreExpertise(id: number): void {
    const next = new Set(selectedCoreExpertiseIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedCoreExpertiseIds = next;
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

  async function handleLensChange(): Promise<void> {
    if (!selectedLensId) {
      return;
    }
    try {
      const req = await getLensExportSelections(selectedLensId);
      // Pre-fill selections from lens
      selectedSummaryIds = new Set(req.summary_ids || []);
      masterSummaryId = req.master_summary_id ?? null;
      selectedWorkHistoryIds = new Set(req.work_history_ids || []);
      selectedBulletIds = new Set(req.bullet_ids || []);
      selectedSkillIds = new Set(req.skill_ids || []);
      selectedAcademicIds = new Set(req.academic_ids || []);
      selectedCertIds = new Set(req.certification_ids || []);
      selectedDescriptorIds = new Set(req.descriptor_ids || []);
      selectedCoreExpertiseIds = new Set(req.core_expertise_ids || []);
      addToast("info", "Selections loaded from lens");
    } catch {
      // Toast already shown
    }
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

  function selectAllCoreExpertise(): void {
    selectedCoreExpertiseIds = new Set(coreExpertise.map((ce) => ce.id));
  }

  // --- Computed ---
  $: hasContent =
    selectedWorkHistoryIds.size > 0 ||
    selectedSkillIds.size > 0 ||
    selectedAcademicIds.size > 0 ||
    selectedCertIds.size > 0 ||
    selectedDescriptorIds.size > 0 ||
    selectedCoreExpertiseIds.size > 0;

  $: canGenerate =
    selectedTemplateId !== null &&
    (isCoverLetter || hasContent) &&
    !generating;

  // --- Generate ---
  async function handleGenerate(): Promise<void> {
    if (!canGenerate || selectedTemplateId === null) {
      addToast("error", "Select a template and at least one content item");
      return;
    }

    // For cover letter templates, collect variables/prompts first.
    if (isCoverLetter) {
      try {
        const vars = await parseTemplateVariables(selectedTemplateId);
        if (
          (vars.variables && vars.variables.length > 0) ||
          (vars.prompts && vars.prompts.length > 0)
        ) {
          promptVariables = vars.variables || [];
          promptPrompts = vars.prompts || [];
          promptPrefilled = {};
          showPromptDialog = true;
          return; // Wait for dialog submit.
        }
        // No variables — proceed directly.
        await generateCoverLetter({});
      } catch {
        // Toast already shown.
      }
      return;
    }

    // Resume export.
    generating = true;
    try {
      const req: ExportRequest = {
        template_id: selectedTemplateId,
        lens_id: selectedLensId,
        summary_ids: [...selectedSummaryIds],
        master_summary_id: masterSummaryId,
        work_history_ids: [...selectedWorkHistoryIds],
        bullet_ids: [...selectedBulletIds],
        skill_ids: [...selectedSkillIds],
        skill_sort_overrides: {},
        academic_ids: [...selectedAcademicIds],
        certification_ids: [...selectedCertIds],
        descriptor_ids: [...selectedDescriptorIds],
        core_expertise_ids: [...selectedCoreExpertiseIds],
      };
      await createExport(req);
      exports = await listExports();
    } catch {
      // Toast already shown.
    } finally {
      generating = false;
    }
  }

  async function generateCoverLetter(
    substitutions: Record<string, string>
  ): Promise<void> {
    if (selectedTemplateId === null) return;
    generating = true;
    try {
      const req: ExportRequest = {
        template_id: selectedTemplateId,
        lens_id: selectedLensId,
        summary_ids: [],
        master_summary_id: null,
        work_history_ids: [],
        bullet_ids: [],
        skill_ids: [],
        skill_sort_overrides: {},
        academic_ids: [],
        certification_ids: [],
        descriptor_ids: [],
        core_expertise_ids: [],
        substitution_map:
          Object.keys(substitutions).length > 0 ? substitutions : undefined,
      };
      await createExport(req);
      exports = await listExports();
    } catch {
      // Toast already shown.
    } finally {
      generating = false;
    }
  }

  function handlePromptSubmit(
    e: CustomEvent<{ substitutions: Record<string, string> }>
  ): void {
    showPromptDialog = false;
    generateCoverLetter(e.detail.substitutions);
  }

  function handlePromptCancel(): void {
    showPromptDialog = false;
  }

  async function handleOpenExport(exportId: number): Promise<void> {
    try {
      await openExportFile(exportId);
    } catch {
      // Toast already shown
    }
  }

  function getTemplateName(templateId: string): string {
    // templateId in exports is the template name (string snapshot).
    return templateId;
  }
</script>

<div class="export-page">
  <div class="page-header">
    <h2>{isCoverLetter ? "Export Cover Letter" : "Export Resume"}</h2>
    <button
      class="btn btn-primary btn-generate"
      on:click={handleGenerate}
      disabled={!canGenerate}
    >
      {generating ? "Generating..." : "Generate PDF"}
    </button>
  </div>
  <p class="page-description">
    Select a template and the content to include, then generate a PDF.
  </p>

  {#if loading}
    <LoadingSpinner />
  {:else}
    <div class="export-layout">
      <!-- Left: Selections -->
      <div class="selections-panel">
        <!-- Template Selection -->
        <section class="selection-section">
          <h3 class="section-title">Template</h3>
          {#if resumeTemplates.length > 0}
            <div class="template-type-label">Resume</div>
            <div class="template-grid">
              {#each resumeTemplates as template (template.id)}
                <button
                  class="template-card"
                  class:selected={selectedTemplateId === template.id}
                  on:click={() => (selectedTemplateId = template.id)}
                >
                  <span class="template-name">{template.name}</span>
                  <span class="template-desc">{template.description}</span>
                  {#if template.is_builtin}
                    <span class="template-badge">Built-in</span>
                  {/if}
                </button>
              {/each}
            </div>
          {/if}
          {#if coverLetterTemplates.length > 0}
            <div class="template-type-label" style="margin-top: 12px">
              Cover Letter
            </div>
            <div class="template-grid">
              {#each coverLetterTemplates as template (template.id)}
                <button
                  class="template-card"
                  class:selected={selectedTemplateId === template.id}
                  on:click={() => (selectedTemplateId = template.id)}
                >
                  <span class="template-name">{template.name}</span>
                  <span class="template-desc">{template.description}</span>
                  {#if template.is_builtin}
                    <span class="template-badge">Built-in</span>
                  {/if}
                </button>
              {/each}
            </div>
          {/if}
        </section>

        <!-- Lens Pre-fill (resume only) -->
        {#if !isCoverLetter && lenses.length > 0}
          <section class="selection-section">
            <h3 class="section-title">Load from Lens</h3>
            <div class="lens-select-row">
              <select
                class="form-input lens-select"
                bind:value={selectedLensId}
              >
                <option value={null}>-- Select a lens --</option>
                {#each lenses as lens (lens.id)}
                  <option value={lens.id}>{lens.name}</option>
                {/each}
              </select>
              <button
                class="btn btn-small"
                on:click={handleLensChange}
                disabled={!selectedLensId}
              >
                Load
              </button>
            </div>
          </section>
        {/if}

        <!-- Resume content sections (hidden for cover letters) -->
        {#if !isCoverLetter}
        <!-- Summary Selection -->
        {#if summaries.length > 0}
          <section class="selection-section">
            <div class="section-header">
              <h3 class="section-title">Professional Summaries</h3>
              <div class="section-header-actions">
                <button
                  class="btn btn-tiny"
                  on:click={() =>
                    (selectedSummaryIds = new Set(summaries.map((s) => s.id)))}
                >
                  All
                </button>
                <button
                  class="btn btn-tiny"
                  on:click={() => {
                    selectedSummaryIds = new Set();
                    masterSummaryId = null;
                  }}
                >
                  None
                </button>
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
                    on:change={() => {
                      const next = new Set(selectedSummaryIds);
                      if (next.has(summary.id)) {
                        next.delete(summary.id);
                        if (masterSummaryId === summary.id) {
                          masterSummaryId = null;
                        }
                      } else {
                        next.add(summary.id);
                      }
                      selectedSummaryIds = next;
                    }}
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

        <!-- Descriptors -->
        {#if descriptors.length > 0}
          <section class="selection-section">
            <div class="section-header">
              <h3 class="section-title">Role Descriptors</h3>
              <button class="btn btn-tiny" on:click={selectAllDescriptors}>
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

        <!-- Core Expertise -->
        {#if coreExpertise.length > 0}
          <section class="selection-section">
            <div class="section-header">
              <h3 class="section-title">Core Expertise</h3>
              <button class="btn btn-tiny" on:click={selectAllCoreExpertise}>
                All
              </button>
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

        <!-- Work History -->
        {#if workHistory.length > 0}
          <section class="selection-section">
            <div class="section-header">
              <h3 class="section-title">Work History</h3>
              <div class="section-header-actions">
                <button class="btn btn-tiny" on:click={selectAllWorkHistory}>
                  All
                </button>
                <button class="btn btn-tiny" on:click={deselectAllWorkHistory}>
                  None
                </button>
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

        <!-- Skills -->
        {#if skillCategories.length > 0}
          <section class="selection-section">
            <div class="section-header">
              <h3 class="section-title">Skills</h3>
              <div class="section-header-actions">
                <button class="btn btn-tiny" on:click={selectAllSkills}>
                  All
                </button>
                <button class="btn btn-tiny" on:click={deselectAllSkills}>
                  None
                </button>
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
        {/if}
        <!-- End resume content sections -->

        {#if isCoverLetter}
          <section class="selection-section">
            <h3 class="section-title">Cover Letter</h3>
            <p class="cover-letter-hint">
              Click "Generate PDF" to fill in any variable placeholders and
              guided prompts, then generate the cover letter.
            </p>
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
                  {formatTimestamp(exp.generated_at)}
                </span>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

{#if showPromptDialog}
  <PromptDialog
    variables={promptVariables}
    prompts={promptPrompts}
    prefilled={promptPrefilled}
    on:submit={handlePromptSubmit}
    on:cancel={handlePromptCancel}
  />
{/if}

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

  /* --- Lens Select --- */
  .lens-select-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .lens-select {
    flex: 1;
    background-color: #1a2332;
    color: #e0e0e0;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 8px 10px;
    font-size: 0.9rem;
  }

  .lens-select:focus {
    outline: none;
    border-color: #4a8af4;
    box-shadow: 0 0 0 2px rgba(74, 138, 244, 0.15);
  }

  .form-input {
    background-color: #1a2332;
    color: #e0e0e0;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 8px 10px;
    font-size: 0.9rem;
  }

  /* --- Checkboxes --- */
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

  .check-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
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

  /* --- Master Summary Pills --- */
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

  .master-warning {
    background-color: #3d2e1a;
    border: 1px solid #705020;
    border-radius: 4px;
    padding: 6px 12px;
    font-size: 0.8rem;
    color: #e0a060;
    margin-bottom: 8px;
  }

  .template-type-label {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: #7a8a9a;
    margin-bottom: 6px;
  }

  .template-badge {
    font-size: 0.7rem;
    color: #4a8af4;
    font-weight: 600;
  }

  .cover-letter-hint {
    font-size: 0.85rem;
    color: #7a8a9a;
    line-height: 1.5;
    margin: 0;
  }
</style>
