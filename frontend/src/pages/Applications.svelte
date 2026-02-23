<script lang="ts">
  import { onMount } from "svelte";
  import {
    listApplications,
    searchApplications,
    createApplication,
    updateApplication,
    updateApplicationStatus,
    deleteApplication,
    getApplicationHistory,
    getApplicationStatuses,
    getFitIndicators,
    listExports,
    listDocumentTemplates,
    listLenses,
    getProfile,
    getLensExportSelections,
    createExport,
    overwriteExport,
    openExportFile,
    parseTemplateVariables,
    getApplicationPromptValues,
    saveApplicationPromptValues,
    addToast,
    type JobApplication,
    type ApplicationInput,
    type StatusChange,
    type ResumeExport,
    type DocumentTemplate,
    type UserProfile,
    type GuidedPrompt,
    type TemplateVariable,
    type ExportRequest,
    type Lens,
  } from "../services/api";
  import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";
  import { formatDate, formatTimestamp } from "../services/dateFormat";

  type GenerationMode = "resume" | "cover_letter" | "both";
  type VersionMode = "new" | "overwrite_latest";
  type GenerationVersionChoice = {
    resume: VersionMode;
    cover: VersionMode;
  };
  type ListFilterMode = "all" | "active";
  type TimelineEntry = {
    id: string;
    from_status: string;
    to_status: string;
    changed_at: string;
    is_baseline?: boolean;
  };

  let applications: JobApplication[] = [];
  let visibleApplications: JobApplication[] = [];
  let loading = true;
  let searchQuery = "";
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;
  let listFilterMode: ListFilterMode = "all";

  let statuses: string[] = [];
  let fitIndicators: string[] = [];
  let exports: ResumeExport[] = [];
  let allTemplates: DocumentTemplate[] = [];
  let resumeTemplates: DocumentTemplate[] = [];
  let coverLetterTemplates: DocumentTemplate[] = [];
  let lenses: Lens[] = [];
  let profile: UserProfile | null = null;

  let openingExportID: number | null = null;

  let showEditor = false;
  let creating = false;
  let saving = false;
  let deleting = false;
  let editingApp: JobApplication | null = null;

  let formCompanyName = "";
  let formPositionTitle = "";
  let formJobPostingURL = "";
  let formDateApplied = "";
  let formStatus = "Applied";
  let formFitIndicator = "";
  let formResumeExportID: number | null = null;
  let formCoverLetterTemplateID: number | null = null;
  let formCoverLetterLatestExportID: number | null = null;
  let formNotes = "";

  let historyEntries: StatusChange[] = [];
  let timelineEntries: TimelineEntry[] = [];
  let historyLoading = false;

  let selectedResumeTemplateID: number | null = null;
  let selectedLensID: number | null = null;
  let promptVariables: TemplateVariable[] = [];
  let promptPrompts: GuidedPrompt[] = [];
  let promptValues: Record<string, string> = {};
  let loadingPromptFields = false;
  let generatingDocuments = false;

  const inactiveStatuses = new Set([
    "Offer Accepted",
    "Offer Declined",
    "Employer Rejected",
    "User Withdrawn",
    "Ghosted",
  ]);

  let showVersionChoiceModal = false;
  let versionChoiceNeedsResume = false;
  let versionChoiceNeedsCover = false;
  let versionChoiceResume: VersionMode = "overwrite_latest";
  let versionChoiceCover: VersionMode = "overwrite_latest";
  let versionChoiceResolver:
    | ((value: GenerationVersionChoice | null) => void)
    | null = null;

  const autoBoundVariableNames = new Set([
    "company_name",
    "position_title",
    "signer_name",
    "email",
  ]);
  const removedVariableNames = new Set(["phone"]);
  const autoBoundPromptIdentifiers = new Set([
    "companyname",
    "positiontitle",
    "signername",
    "email",
    "phone",
  ]);

  $: visibleApplications = applications.filter((app) =>
    listFilterMode === "active" ? !inactiveStatuses.has(app.status) : true
  );

  $: timelineEntries = buildTimelineEntries(editingApp, historyEntries);

  onMount(async () => {
    await Promise.all([loadApplications(), loadReferenceData()]);
  });

  async function loadReferenceData(): Promise<void> {
    try {
      const [
        loadedStatuses,
        loadedFitIndicators,
        loadedExports,
        templates,
        loadedLenses,
        loadedProfile,
      ] = await Promise.all([
        getApplicationStatuses(),
        getFitIndicators(),
        listExports(),
        listDocumentTemplates(),
        listLenses(),
        getProfile(),
      ]);

      statuses = loadedStatuses || [];
      fitIndicators = loadedFitIndicators || [];
      exports = loadedExports || [];
      lenses = loadedLenses || [];
      profile = loadedProfile || null;

      allTemplates = templates || [];
      resumeTemplates = allTemplates.filter((t) => t.template_type === "resume");
      coverLetterTemplates = allTemplates.filter(
        (t) => t.template_type === "cover_letter"
      );
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

  async function reloadPageData(): Promise<void> {
    await Promise.all([loadApplications(), loadReferenceData()]);
    if (!showEditor || !editingApp) {
      return;
    }

    const refreshed = applications.find((a) => a.id === editingApp?.id);
    if (refreshed) {
      editingApp = refreshed;
      applyAppToForm(refreshed);
    }
  }

  function handleSearchInput(): void {
    if (searchTimeout) clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      loadApplications();
    }, 300);
  }

  function resetGenerationState(): void {
    selectedResumeTemplateID = resumeTemplates[0]?.id ?? null;
    selectedLensID = lenses[0]?.id ?? null;
    promptVariables = [];
    promptPrompts = [];
    promptValues = {};
    loadingPromptFields = false;
    generatingDocuments = false;
  }

  function defaultStatus(): string {
    return statuses[0] || "Applied";
  }

  function openCreateEditor(): void {
    creating = true;
    editingApp = null;
    showEditor = true;

    formCompanyName = "";
    formPositionTitle = "";
    formJobPostingURL = "";
    formDateApplied = new Date().toISOString().slice(0, 10);
    formStatus = defaultStatus();
    formFitIndicator = "";
    formResumeExportID = null;
    formCoverLetterTemplateID = coverLetterTemplates[0]?.id ?? null;
    formCoverLetterLatestExportID = null;
    formNotes = "";

    historyEntries = [];
    resetGenerationState();
    void loadPromptFields();
  }

  function applyAppToForm(app: JobApplication): void {
    formCompanyName = app.company_name;
    formPositionTitle = app.position_title;
    formJobPostingURL = app.job_posting_url || "";
    formDateApplied = app.date_applied;
    formStatus = app.status;
    formFitIndicator = app.fit_indicator;
    formResumeExportID = app.resume_export_id;
    formCoverLetterTemplateID = app.cover_letter_template_id;
    formCoverLetterLatestExportID = app.cover_letter_latest_export_id;
    formNotes = app.notes;
  }

  function exportByID(exportID: number | null): ResumeExport | null {
    if (!exportID) return null;
    return exports.find((e) => e.id === exportID) || null;
  }

  function guessResumeTemplateFromExport(exportID: number | null): number | null {
    const ex = exportByID(exportID);
    if (!ex) return null;
    if (typeof ex.template_ref_id === "number") return ex.template_ref_id;

    const byName = resumeTemplates.find((t) =>
      t.name.toLowerCase().includes((ex.template_id || "").toLowerCase())
    );
    return byName?.id ?? null;
  }

  async function loadHistory(): Promise<void> {
    if (!editingApp) {
      historyEntries = [];
      return;
    }
    historyLoading = true;
    try {
      historyEntries = (await getApplicationHistory(editingApp.id)) || [];
    } catch {
      // Toast already shown
    } finally {
      historyLoading = false;
    }
  }

  async function openEditEditor(app: JobApplication): Promise<void> {
    creating = false;
    editingApp = app;
    applyAppToForm(app);
    showEditor = true;

    selectedResumeTemplateID =
      guessResumeTemplateFromExport(app.resume_export_id) ||
      resumeTemplates[0]?.id ||
      null;
    if (!formCoverLetterTemplateID) {
      formCoverLetterTemplateID = coverLetterTemplates[0]?.id ?? null;
    }
    selectedLensID = exportByID(app.resume_export_id)?.lens_id ?? lenses[0]?.id ?? null;

    await Promise.all([loadHistory(), loadPromptFields()]);
  }

  function closeEditor(): void {
    showEditor = false;
    creating = false;
    editingApp = null;
    historyEntries = [];
    saving = false;
    deleting = false;
    resetGenerationState();
  }

  function buildApplicationInput(
    overrides: Partial<ApplicationInput> = {}
  ): ApplicationInput {
    return {
      company_name: formCompanyName.trim(),
      position_title: formPositionTitle.trim(),
      job_posting_url: formJobPostingURL.trim(),
      date_applied: formDateApplied,
      fit_indicator: formFitIndicator,
      resume_export_id: formResumeExportID,
      cover_letter_template_id: formCoverLetterTemplateID,
      cover_letter_latest_export_id: formCoverLetterLatestExportID,
      notes: formNotes.trim(),
      ...overrides,
    };
  }

  async function saveEditor(): Promise<void> {
    if (!formCompanyName.trim()) {
      addToast("error", "Company name is required");
      return;
    }
    if (!formPositionTitle.trim()) {
      addToast("error", "Position title is required");
      return;
    }
    if (!formDateApplied) {
      addToast("error", "Date applied is required");
      return;
    }

    const effectiveCoverTemplateID =
      formCoverLetterTemplateID ?? coverLetterTemplates[0]?.id ?? null;
    if (effectiveCoverTemplateID !== formCoverLetterTemplateID) {
      formCoverLetterTemplateID = effectiveCoverTemplateID;
    }

    saving = true;
    try {
      let savedApp: JobApplication | null = null;

      if (creating) {
        let created = await createApplication(
          buildApplicationInput({
            cover_letter_template_id: effectiveCoverTemplateID,
          })
        );
        if (formStatus && formStatus !== created.status) {
          created = await updateApplicationStatus(created.id, formStatus);
        }
        savedApp = created;
        editingApp = created;
        creating = false;
      } else if (editingApp) {
        let updated = await updateApplication(
          editingApp.id,
          buildApplicationInput({
            cover_letter_template_id: effectiveCoverTemplateID,
          })
        );
        if (formStatus && formStatus !== editingApp.status) {
          updated = await updateApplicationStatus(updated.id, formStatus);
        }
        savedApp = updated;
        editingApp = updated;
      }

      if (savedApp && effectiveCoverTemplateID) {
        await saveApplicationPromptValues(
          savedApp.id,
          effectiveCoverTemplateID,
          buildPromptValuesToSave()
        );
      }

      await reloadPageData();
      await loadHistory();
      if (editingApp && formCoverLetterTemplateID) {
        await loadPromptFields();
      }
    } catch {
      // Toast already shown
    } finally {
      saving = false;
    }
  }

  async function deleteFromEditor(): Promise<void> {
    if (!editingApp) return;
    if (!window.confirm("Delete this application? This cannot be undone.")) {
      return;
    }

    deleting = true;
    try {
      await deleteApplication(editingApp.id);
      closeEditor();
      await reloadPageData();
    } catch {
      // Toast already shown
    } finally {
      deleting = false;
    }
  }

  async function openLinkedExport(
    exportID: number | null,
    documentType: "resume" | "cover letter"
  ): Promise<void> {
    if (!exportID) {
      addToast("info", `No ${documentType} generated yet`);
      return;
    }

    openingExportID = exportID;
    try {
      await openExportFile(exportID);
    } catch {
      // Toast already shown
    } finally {
      if (openingExportID === exportID) {
        openingExportID = null;
      }
    }
  }

  function openJobPostingURL(url: string): void {
    const trimmed = url.trim();
    if (!trimmed) {
      return;
    }
    BrowserOpenURL(trimmed);
  }

  function exportLabel(exportID: number | null): string {
    if (!exportID) return "Not generated";
    const ex = exportByID(exportID);
    if (!ex) return `Export #${exportID}`;

    let template = ex.template_id;
    if (typeof ex.template_ref_id === "number") {
      const tmpl = allTemplates.find((t) => t.id === ex.template_ref_id);
      if (tmpl) template = tmpl.name;
    }
    return `${template} (${formatDate(ex.generated_at.slice(0, 10), "day")})`;
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

  function buildTimelineEntries(
    app: JobApplication | null,
    history: StatusChange[]
  ): TimelineEntry[] {
    if (!app) {
      return [];
    }

    const baselineStatus =
      history.length > 0
        ? history[0].from_status || history[0].to_status
        : app.status;

    const items: TimelineEntry[] = [
      {
        id: `baseline-${app.id}`,
        from_status: "",
        to_status: baselineStatus,
        changed_at: app.created_at,
        is_baseline: true,
      },
    ];

    for (const h of history) {
      items.push({
        id: `change-${h.id}`,
        from_status: h.from_status,
        to_status: h.to_status,
        changed_at: h.changed_at,
      });
    }

    return items;
  }

  function promptKey(prompt: GuidedPrompt): string {
    if (prompt.key && prompt.key.trim()) return prompt.key;
    return `prompt:${prompt.prompt_text}`;
  }

  function buildAutoBoundSubstitutions(): Record<string, string> {
    return {
      company_name: formCompanyName.trim(),
      position_title: formPositionTitle.trim(),
      signer_name: profile?.full_name?.trim() || "",
      email: profile?.email?.trim() || "",
    };
  }

  function isAutoBoundVariable(name: string): boolean {
    return autoBoundVariableNames.has(name.trim());
  }

  function normalizePromptIdentifier(value: string): string {
    return value.trim().toLowerCase().replace(/[^a-z0-9]+/g, "");
  }

  function isAutoBoundPrompt(prompt: GuidedPrompt): boolean {
    const key = (prompt.key || "").trim();
    if (key && (isAutoBoundVariable(key) || removedVariableNames.has(key))) {
      return true;
    }
    const identifier = normalizePromptIdentifier(prompt.prompt_text || "");
    return autoBoundPromptIdentifiers.has(identifier);
  }

  function buildPromptValuesToSave(): Record<string, string> {
    const valuesToSave: Record<string, string> = {};
    for (const [key, value] of Object.entries(promptValues)) {
      const trimmedKey = key.trim();
      if (isAutoBoundVariable(trimmedKey) || removedVariableNames.has(trimmedKey)) {
        continue;
      }
      const trimmedValue = value.trim();
      if (trimmedValue) {
        valuesToSave[key] = trimmedValue;
      }
    }
    return valuesToSave;
  }

  async function loadPromptFields(): Promise<void> {
    const effectiveCoverTemplateID =
      formCoverLetterTemplateID ?? coverLetterTemplates[0]?.id ?? null;
    if (effectiveCoverTemplateID !== formCoverLetterTemplateID) {
      formCoverLetterTemplateID = effectiveCoverTemplateID;
    }

    if (!effectiveCoverTemplateID) {
      promptVariables = [];
      promptPrompts = [];
      promptValues = {};
      return;
    }

    loadingPromptFields = true;
    try {
      const varsPromise = parseTemplateVariables(effectiveCoverTemplateID);
      const savedPromise = editingApp
        ? getApplicationPromptValues(editingApp.id, effectiveCoverTemplateID)
        : Promise.resolve({} as Record<string, string>);
      const [vars, saved] = await Promise.all([varsPromise, savedPromise]);
      const autoSubs = buildAutoBoundSubstitutions();

      promptVariables = (vars.variables || []).filter((v) => {
        const key = v.name.trim();
        return !removedVariableNames.has(key) && !isAutoBoundVariable(key);
      });
      promptPrompts = (vars.prompts || []).filter((p) => !isAutoBoundPrompt(p));

      const initial: Record<string, string> = {};
      for (const v of promptVariables) {
        initial[v.name] = saved[v.name] ?? "";
      }
      for (const p of promptPrompts) {
        const key = promptKey(p);
        initial[key] = saved[key] ?? "";
      }

      for (const [key, value] of Object.entries(autoSubs)) {
        if (value) {
          initial[key] = value;
        }
      }

      delete initial.phone;
      promptValues = initial;
    } catch {
      // Toast already shown
    } finally {
      loadingPromptFields = false;
    }
  }

  function prettyName(value: string): string {
    return value
      .replace(/_/g, " ")
      .replace(/\b\w/g, (m) => m.toUpperCase());
  }

  function closeVersionChoiceModal(result: GenerationVersionChoice | null): void {
    showVersionChoiceModal = false;

    const resolver = versionChoiceResolver;
    versionChoiceResolver = null;
    if (resolver) {
      resolver(result);
    }
  }

  function confirmVersionChoiceModal(): void {
    closeVersionChoiceModal({
      resume: versionChoiceNeedsResume ? versionChoiceResume : "new",
      cover: versionChoiceNeedsCover ? versionChoiceCover : "new",
    });
  }

  function cancelVersionChoiceModal(): void {
    closeVersionChoiceModal(null);
  }

  async function chooseVersionModes(
    mode: GenerationMode
  ): Promise<GenerationVersionChoice | null> {
    const needsResumeChoice =
      (mode === "resume" || mode === "both") && !!formResumeExportID;
    const needsCoverChoice =
      (mode === "cover_letter" || mode === "both") &&
      !!formCoverLetterLatestExportID;

    if (!needsResumeChoice && !needsCoverChoice) {
      return {
        resume: "new",
        cover: "new",
      };
    }

    versionChoiceNeedsResume = needsResumeChoice;
    versionChoiceNeedsCover = needsCoverChoice;
    versionChoiceResume = "overwrite_latest";
    versionChoiceCover = "overwrite_latest";
    showVersionChoiceModal = true;

    return new Promise((resolve) => {
      versionChoiceResolver = resolve;
    });
  }

  async function generateDocuments(mode: GenerationMode): Promise<void> {
    if (creating || !editingApp) {
      addToast("info", "Save this application before generating documents");
      return;
    }

    const effectiveCoverTemplateID =
      formCoverLetterTemplateID ?? coverLetterTemplates[0]?.id ?? null;
    if (effectiveCoverTemplateID !== formCoverLetterTemplateID) {
      formCoverLetterTemplateID = effectiveCoverTemplateID;
    }

    if ((mode === "resume" || mode === "both") && !selectedResumeTemplateID) {
      addToast("error", "Select a resume template");
      return;
    }
    if ((mode === "resume" || mode === "both") && !selectedLensID) {
      addToast("error", "Select a lens for resume generation");
      return;
    }
    if ((mode === "cover_letter" || mode === "both") && !effectiveCoverTemplateID) {
      addToast("error", "Select a cover letter template");
      return;
    }

    for (const p of promptPrompts) {
      const key = promptKey(p);
      if (p.required && !promptValues[key]?.trim()) {
        addToast("error", `Required field missing: ${p.prompt_text}`);
        return;
      }
    }

    const versionChoice = await chooseVersionModes(mode);
    if (!versionChoice) {
      return;
    }

    generatingDocuments = true;
    try {
      let nextResumeExportID = formResumeExportID;
      let nextCoverLatestExportID = formCoverLetterLatestExportID;

      if (mode === "resume" || mode === "both") {
        const lensSelections = await getLensExportSelections(selectedLensID!);
        const resumeRequest: ExportRequest = {
          ...lensSelections,
          template_id: selectedResumeTemplateID!,
          lens_id: selectedLensID,
        };
        const resumeExport =
          versionChoice.resume === "overwrite_latest" && nextResumeExportID
            ? await overwriteExport(nextResumeExportID, resumeRequest)
            : await createExport(resumeRequest);
        nextResumeExportID = resumeExport.id;
      }

      if (mode === "cover_letter" || mode === "both") {
        const autoSubs = buildAutoBoundSubstitutions();
        const substitutions: Record<string, string> = {};

        for (const [key, value] of Object.entries(promptValues)) {
          if (removedVariableNames.has(key.trim())) {
            continue;
          }
          const trimmed = value.trim();
          if (trimmed) {
            substitutions[key] = trimmed;
          }
        }

        for (const [key, value] of Object.entries(autoSubs)) {
          if (value) {
            substitutions[key] = value;
          } else {
            delete substitutions[key];
          }
        }

        delete substitutions.phone;

        await saveApplicationPromptValues(
          editingApp.id,
          effectiveCoverTemplateID!,
          buildPromptValuesToSave()
        );

        const coverRequest: ExportRequest = {
          template_id: effectiveCoverTemplateID!,
          lens_id: null,
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

        const coverExport =
          versionChoice.cover === "overwrite_latest" && nextCoverLatestExportID
            ? await overwriteExport(nextCoverLatestExportID, coverRequest)
            : await createExport(coverRequest);
        nextCoverLatestExportID = coverExport.id;
      }

      const updated = await updateApplication(
        editingApp.id,
        buildApplicationInput({
          resume_export_id: nextResumeExportID,
          cover_letter_template_id: effectiveCoverTemplateID,
          cover_letter_latest_export_id: nextCoverLatestExportID,
        })
      );

      editingApp = updated;
      applyAppToForm(updated);
      await Promise.all([reloadPageData(), loadHistory()]);
      addToast("success", "Documents generated for this application");
    } catch {
      // Toast already shown
    } finally {
      generatingDocuments = false;
    }
  }
</script>

<div class="applications-page">
  <div class="page-header">
    <h2>Job Applications</h2>
    <button class="btn btn-primary" on:click={openCreateEditor}>+ New Application</button>
  </div>

  <p class="page-description">
    Read-only list view with quick access to latest documents. Edit an
    application to update details and generate new versions.
  </p>

  <div class="search-bar">
    <input
      type="text"
      class="form-input search-input"
      placeholder="Search by company or position..."
      bind:value={searchQuery}
      on:input={handleSearchInput}
    />

    <div class="list-filter-toggle" role="group" aria-label="Application status filter">
      <button
        class="toggle-btn"
        class:active={listFilterMode === "all"}
        on:click={() => (listFilterMode = "all")}
      >
        All
      </button>
      <button
        class="toggle-btn"
        class:active={listFilterMode === "active"}
        on:click={() => (listFilterMode = "active")}
      >
        Active Only
      </button>
    </div>
  </div>

  {#if loading}
    <LoadingSpinner />
  {:else if visibleApplications.length === 0}
    <div class="empty-state">
      <p>
        {#if searchQuery.trim()}
          No applications match your search.
        {:else if listFilterMode === "active"}
          No active applications right now.
        {:else}
          No applications yet.
        {/if}
      </p>
    </div>
  {:else}
    <div class="applications-table-wrap">
      <table class="applications-table">
        <thead>
          <tr>
            <th>Company / Role</th>
            <th>Status</th>
            <th>Fit</th>
            <th>Job URL</th>
            <th>Documents</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each visibleApplications as app (app.id)}
            <tr>
              <td>
                <div class="row-company">{app.company_name}</div>
                <div class="row-position">{app.position_title}</div>
                <div class="row-meta">Applied {formatDate(app.date_applied, "day")}</div>
              </td>
              <td>
                <span
                  class="status-badge"
                  style="color: {statusColor(app.status)}; border-color: {statusColor(app.status)}40;"
                >
                  {app.status}
                </span>
              </td>
              <td>
                {#if app.fit_indicator}
                  <span
                    class="fit-badge"
                    style="color: {fitColor(app.fit_indicator)}; border-color: {fitColor(app.fit_indicator)}40;"
                  >
                    {app.fit_indicator}
                  </span>
                {:else}
                  <span class="row-empty">-</span>
                {/if}
              </td>
              <td>
                {#if app.job_posting_url}
                  <button
                    class="job-link-btn"
                    on:click={() => openJobPostingURL(app.job_posting_url)}
                  >
                    Open
                  </button>
                {:else}
                  <span class="row-empty">-</span>
                {/if}
              </td>
              <td>
                <div class="doc-mini-actions">
                  <button
                    class="doc-mini-btn"
                    on:click={() => openLinkedExport(app.resume_export_id, "resume")}
                    disabled={!app.resume_export_id || openingExportID === app.resume_export_id}
                    title={exportLabel(app.resume_export_id)}
                    aria-label="Open resume"
                  >
                    R
                  </button>
                  <button
                    class="doc-mini-btn"
                    on:click={() =>
                      openLinkedExport(app.cover_letter_latest_export_id, "cover letter")}
                    disabled={
                      !app.cover_letter_latest_export_id ||
                      openingExportID === app.cover_letter_latest_export_id
                    }
                    title={exportLabel(app.cover_letter_latest_export_id)}
                    aria-label="Open cover letter"
                  >
                    CL
                  </button>
                </div>
              </td>
              <td>
                <button class="btn btn-primary btn-small" on:click={() => openEditEditor(app)}>
                  Edit
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

{#if showEditor}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="overlay" on:click|self={closeEditor}>
    <div class="editor-dialog">
      <div class="editor-header">
        <h3>{creating ? "New Application" : `Edit: ${formCompanyName || "Application"}`}</h3>
        <div class="editor-header-actions">
          <button class="btn btn-primary" on:click={saveEditor} disabled={saving}>
            {saving ? "Saving..." : "Save Changes"}
          </button>
          <button class="btn btn-ghost btn-small" on:click={closeEditor}>Close</button>
        </div>
      </div>

      <div class="editor-body">
        <section class="editor-section editor-main">
          <h4>Application Details</h4>

          <div class="form-grid">
            <div class="form-field">
              <label class="form-label" for="editor-company">Company</label>
              <input
                id="editor-company"
                class="form-input"
                type="text"
                bind:value={formCompanyName}
              />
            </div>

            <div class="form-field">
              <label class="form-label" for="editor-position">Position Title</label>
              <input
                id="editor-position"
                class="form-input"
                type="text"
                bind:value={formPositionTitle}
              />
            </div>

            <div class="form-field">
              <label class="form-label" for="editor-job-url">Job Posting URL</label>
              <input
                id="editor-job-url"
                class="form-input"
                type="url"
                placeholder="https://company.com/jobs/..."
                bind:value={formJobPostingURL}
              />
            </div>

            <div class="form-field">
              <label class="form-label" for="editor-date">Date Applied</label>
              <input
                id="editor-date"
                class="form-input"
                type="date"
                bind:value={formDateApplied}
              />
            </div>

            <div class="form-field">
              <label class="form-label" for="editor-status">Status</label>
              <select id="editor-status" class="form-input" bind:value={formStatus}>
                {#each statuses as s (s)}
                  <option value={s}>{s}</option>
                {/each}
              </select>
            </div>
          </div>

          <div class="form-field fit-field">
            <span class="form-label">Fit</span>
            <div class="fit-options">
              {#each fitIndicators as fit (fit)}
                <button
                  class="fit-option"
                  class:fit-active={formFitIndicator === fit}
                  style={formFitIndicator === fit
                    ? `color: ${fitColor(fit)}; border-color: ${fitColor(fit)};`
                    : ""}
                  on:click={() => (formFitIndicator = fit)}
                >
                  {fit}
                </button>
              {/each}
              {#if formFitIndicator}
                <button class="fit-option" on:click={() => (formFitIndicator = "")}>Clear</button>
              {/if}
            </div>
          </div>

          <div class="form-field">
            <label class="form-label" for="editor-notes">Notes</label>
            <textarea
              id="editor-notes"
              class="form-input form-textarea"
              rows="5"
              bind:value={formNotes}
            />
          </div>

          <div class="details-divider"></div>

          <div class="group-card">
            <div class="group-card-header">
              <h5>Resume</h5>
              <button
                class="btn btn-secondary btn-small"
                on:click={() => openLinkedExport(formResumeExportID, "resume")}
                disabled={!formResumeExportID || openingExportID === formResumeExportID}
              >
                Open Resume
              </button>
            </div>

            <div class="group-grid">
              <div class="form-field">
                <label class="form-label" for="editor-linked-resume">Linked Resume Export</label>
                <select id="editor-linked-resume" class="form-input" bind:value={formResumeExportID}>
                  <option value={null}>-- None --</option>
                  {#each exports as ex (ex.id)}
                    <option value={ex.id}>{exportLabel(ex.id)}</option>
                  {/each}
                </select>
              </div>

              <div class="form-field">
                <label class="form-label" for="gen-resume-template">Resume Template</label>
                <select
                  id="gen-resume-template"
                  class="form-input"
                  bind:value={selectedResumeTemplateID}
                >
                  <option value={null}>-- Select --</option>
                  {#each resumeTemplates as t (t.id)}
                    <option value={t.id}>{t.name}</option>
                  {/each}
                </select>
              </div>

              <div class="form-field">
                <label class="form-label" for="gen-lens">Lens (required)</label>
                <select id="gen-lens" class="form-input" bind:value={selectedLensID}>
                  <option value={null}>-- Select --</option>
                  {#each lenses as lens (lens.id)}
                    <option value={lens.id}>{lens.name}</option>
                  {/each}
                </select>
              </div>
            </div>

            <div class="group-actions">
              <button
                class="btn btn-secondary"
                on:click={() => generateDocuments("resume")}
                disabled={creating || generatingDocuments}
              >
                {generatingDocuments ? "Generating..." : "Generate Resume"}
              </button>
            </div>
          </div>

          <div class="details-divider"></div>

          <div class="group-card">
            <div class="group-card-header">
              <h5>Cover Letter</h5>
              <button
                class="btn btn-secondary btn-small"
                on:click={() => openLinkedExport(formCoverLetterLatestExportID, "cover letter")}
                disabled={
                  !formCoverLetterLatestExportID ||
                  openingExportID === formCoverLetterLatestExportID
                }
              >
                Open Cover Letter
              </button>
            </div>

            <div class="form-field">
              <label class="form-label" for="editor-cover-template">Cover Letter Template</label>
              <select
                id="editor-cover-template"
                class="form-input"
                bind:value={formCoverLetterTemplateID}
                on:change={loadPromptFields}
              >
                {#if coverLetterTemplates.length === 0}
                  <option value={null}>-- No Cover Letter Templates --</option>
                {:else}
                  {#each coverLetterTemplates as t (t.id)}
                    <option value={t.id}>{t.name}</option>
                  {/each}
                {/if}
              </select>
            </div>

            <p class="hint">
              Company, Position, Signer Name, and Email are strongly bound.
              Phone is excluded.
            </p>

            {#if loadingPromptFields}
              <p class="hint">Loading cover letter fields...</p>
            {:else}
              {#if promptVariables.length > 0}
                <h6 class="subhead">Template Variables</h6>
                {#each promptVariables as v (v.name)}
                  <div class="form-field">
                    <label class="form-label" for={`var-${v.name}`}>{prettyName(v.name)}</label>
                    <input
                      id={`var-${v.name}`}
                      class="form-input"
                      type="text"
                      bind:value={promptValues[v.name]}
                    />
                  </div>
                {/each}
              {/if}

              {#if promptPrompts.length > 0}
                <h6 class="subhead">Prompted Fields</h6>
                {#each promptPrompts as p (promptKey(p))}
                  <div class="form-field">
                    <label class="form-label" for={`prompt-${promptKey(p)}`}>
                      {p.prompt_text}{p.required ? " *" : ""}
                    </label>
                    {#if p.help_text}
                      <p class="hint">{p.help_text}</p>
                    {/if}
                    {#if p.multiline}
                      <textarea
                        id={`prompt-${promptKey(p)}`}
                        class="form-input form-textarea"
                        rows="4"
                        bind:value={promptValues[promptKey(p)]}
                      />
                    {:else}
                      <input
                        id={`prompt-${promptKey(p)}`}
                        class="form-input"
                        type="text"
                        bind:value={promptValues[promptKey(p)]}
                      />
                    {/if}
                  </div>
                {/each}
              {/if}

              {#if promptVariables.length === 0 && promptPrompts.length === 0}
                <p class="hint">No additional fields for this cover letter template.</p>
              {/if}
            {/if}

            <div class="group-actions">
              <button
                class="btn btn-secondary"
                on:click={() => generateDocuments("cover_letter")}
                disabled={creating || generatingDocuments}
              >
                {generatingDocuments ? "Generating..." : "Generate Cover Letter"}
              </button>
              <button
                class="btn btn-primary"
                on:click={() => generateDocuments("both")}
                disabled={creating || generatingDocuments}
              >
                {generatingDocuments ? "Generating..." : "Generate Both"}
              </button>
            </div>
          </div>

          {#if creating}
            <p class="hint">Save the application first to generate documents.</p>
          {/if}

          <div class="editor-actions">
            <button class="btn btn-primary" on:click={saveEditor} disabled={saving}>
              {saving ? "Saving..." : "Save Changes"}
            </button>
            {#if !creating}
              <button
                class="btn btn-danger"
                on:click={deleteFromEditor}
                disabled={deleting || saving}
              >
                {deleting ? "Deleting..." : "Delete Application"}
              </button>
            {/if}
          </div>
        </section>

        <section class="editor-section editor-side">
          <h4>Status Timeline</h4>
          {#if creating}
            <p class="hint">Timeline becomes available after the application is saved.</p>
          {:else if historyLoading}
            <p class="hint">Loading history...</p>
          {:else if timelineEntries.length === 0}
            <p class="hint">No status history available.</p>
          {:else}
            <div class="timeline">
              {#each timelineEntries as entry (entry.id)}
                <div class="timeline-entry">
                  <span
                    class="timeline-dot"
                    style="background-color: {statusColor(entry.to_status)};"
                  ></span>
                  {#if entry.is_baseline}
                    <span class="timeline-text">Created with status {entry.to_status}</span>
                  {:else}
                    <span class="timeline-text">{entry.from_status} &rarr; {entry.to_status}</span>
                  {/if}
                  <span class="timeline-date">{formatTimestamp(entry.changed_at)}</span>
                </div>
              {/each}
            </div>
          {/if}
        </section>
      </div>
    </div>
  </div>
{/if}

{#if showVersionChoiceModal}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="overlay" on:click|self={cancelVersionChoiceModal}>
    <div class="version-modal">
      <h3>Choose Version Action</h3>
      <p class="hint">
        Existing document versions were found. Choose whether to overwrite the
        latest version or create a new one.
      </p>

      {#if versionChoiceNeedsResume}
        <div class="version-choice-group">
          <p class="version-choice-title">Resume</p>
          <label class="choice-option">
            <input type="radio" bind:group={versionChoiceResume} value="overwrite_latest" />
            Overwrite latest resume
          </label>
          <label class="choice-option">
            <input type="radio" bind:group={versionChoiceResume} value="new" />
            Create new resume version
          </label>
        </div>
      {/if}

      {#if versionChoiceNeedsCover}
        <div class="version-choice-group">
          <p class="version-choice-title">Cover Letter</p>
          <label class="choice-option">
            <input type="radio" bind:group={versionChoiceCover} value="overwrite_latest" />
            Overwrite latest cover letter
          </label>
          <label class="choice-option">
            <input type="radio" bind:group={versionChoiceCover} value="new" />
            Create new cover letter version
          </label>
        </div>
      {/if}

      <div class="version-choice-actions">
        <button class="btn btn-cancel" on:click={cancelVersionChoiceModal}>Cancel</button>
        <button class="btn btn-primary" on:click={confirmVersionChoiceModal}>Continue</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .applications-page {
    max-width: 1120px;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
  }

  h2 {
    margin: 0;
    color: #e0e0e0;
    font-size: 1.5rem;
  }

  .page-description {
    margin: 0 0 16px;
    color: #7a8a9a;
    font-size: 0.92rem;
  }

  .search-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
    margin-bottom: 18px;
  }

  .search-input {
    flex: 1 1 380px;
    box-sizing: border-box;
  }

  .list-filter-toggle {
    display: inline-flex;
    border: 1px solid #2c4157;
    border-radius: 7px;
    overflow: hidden;
    background: #1a2534;
    flex-shrink: 0;
  }

  .toggle-btn {
    min-height: 44px;
    border: none;
    border-right: 1px solid #2c4157;
    background: transparent;
    color: #8fa3b8;
    font-size: 0.8rem;
    font-weight: 600;
    letter-spacing: 0.02em;
    padding: 0 12px;
    cursor: pointer;
    transition: background-color 0.15s, color 0.15s;
  }

  .toggle-btn:last-child {
    border-right: none;
  }

  .toggle-btn:hover {
    background: #223347;
    color: #d4e0ef;
  }

  .toggle-btn.active {
    background: #2a4d73;
    color: #edf4ff;
  }

  .empty-state {
    text-align: center;
    color: #5a6a7a;
    padding: 48px 0;
  }

  .applications-table-wrap {
    border: 1px solid #2a3a4a;
    border-radius: 8px;
    overflow: auto;
    background: #1e2d3d;
  }

  .applications-table {
    width: 100%;
    min-width: 860px;
    border-collapse: collapse;
  }

  .applications-table thead {
    background: #1a2534;
  }

  .applications-table th {
    text-align: left;
    color: #8ea2b8;
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 700;
    padding: 11px 12px;
    border-bottom: 1px solid #2a3a4a;
    white-space: nowrap;
  }

  .applications-table td {
    padding: 11px 12px;
    border-bottom: 1px solid #25384b;
    vertical-align: middle;
  }

  .applications-table tbody tr:last-child td {
    border-bottom: none;
  }

  .applications-table tbody tr:hover {
    background: #1b2a3b;
  }

  .row-company {
    color: #dce5f0;
    font-size: 0.9rem;
    font-weight: 600;
    line-height: 1.25;
  }

  .row-position {
    color: #9cb0c5;
    font-size: 0.78rem;
    margin-top: 2px;
    line-height: 1.3;
  }

  .row-meta {
    color: #6f8297;
    font-size: 0.73rem;
    margin-top: 3px;
  }

  .row-empty {
    color: #5f7288;
    font-size: 0.78rem;
  }

  .job-link-btn {
    color: #7eafff;
    font-size: 0.78rem;
    text-decoration: none;
    background: transparent;
    border: none;
    padding: 0;
    cursor: pointer;
  }

  .job-link-btn:hover {
    color: #a5c7ff;
    text-decoration: underline;
  }

  .status-badge,
  .fit-badge {
    border: 1px solid;
    border-radius: 4px;
    padding: 2px 8px;
    font-size: 0.72rem;
    font-weight: 600;
  }

  .doc-mini-actions {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .doc-mini-btn {
    min-width: 38px;
    height: 30px;
    border: 1px solid #324a63;
    border-radius: 6px;
    background: #1a2534;
    color: #d3deec;
    font-size: 0.72rem;
    font-weight: 700;
    cursor: pointer;
    transition: border-color 0.15s, background-color 0.15s, color 0.15s;
  }

  .doc-mini-btn:hover {
    border-color: #4f6d8e;
    background: #213246;
    color: #e4ecf7;
  }

  .doc-mini-btn:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.58);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 160;
    padding: 16px;
  }

  .editor-dialog {
    width: min(1120px, 96vw);
    max-height: 92vh;
    overflow: auto;
    background: #1c2b3b;
    border: 1px solid #2f4256;
    border-radius: 10px;
    padding: 16px;
  }

  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 12px;
  }

  .editor-header-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .editor-header h3 {
    margin: 0;
    color: #e0e0e0;
    font-size: 1.05rem;
  }

  .editor-body {
    display: grid;
    grid-template-columns: 1.45fr 0.95fr;
    gap: 12px;
  }

  .editor-main {
    min-width: 0;
  }

  .editor-side {
    min-width: 0;
  }

  .editor-section {
    background: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 8px;
    padding: 12px;
  }

  .editor-section h4 {
    margin: 0 0 10px;
    color: #c0d0e0;
    font-size: 0.92rem;
  }

  .editor-section h5 {
    margin: 0 0 8px;
    color: #a0b0c0;
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .subhead {
    margin: 10px 0 6px;
    color: #a0b0c0;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .form-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .form-label {
    color: #7a8a9a;
    font-size: 0.74rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-weight: 600;
  }

  .form-input {
    background: #1a2332;
    color: #e0e0e0;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    min-height: 44px;
    padding: 10px 12px;
    font-size: 0.86rem;
    line-height: 1.2;
  }

  select.form-input {
    height: 44px;
  }

  .form-input:focus {
    outline: none;
    border-color: #4a8af4;
    box-shadow: 0 0 0 2px rgba(74, 138, 244, 0.14);
  }

  .form-textarea {
    min-height: 96px;
    height: auto;
    padding-top: 10px;
    font-family: inherit;
    resize: vertical;
    line-height: 1.5;
  }

  .fit-field {
    margin-top: 10px;
  }

  .details-divider {
    margin: 12px 0;
    border-top: 1px solid #2a3a4a;
  }

  .group-card {
    border: 1px solid #2b3e53;
    background: #192636;
    border-radius: 8px;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .group-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .group-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .group-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .fit-options {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  .fit-option {
    border: 1px solid #2a3a4a;
    background: transparent;
    color: #7a8a9a;
    border-radius: 4px;
    padding: 4px 8px;
    font-size: 0.75rem;
    cursor: pointer;
  }

  .fit-option:hover {
    border-color: #3a4f62;
    color: #c0d0e0;
  }

  .fit-active {
    font-weight: 600;
  }

  .editor-actions {
    margin-top: 12px;
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .timeline {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-top: 6px;
  }

  .timeline-entry {
    display: grid;
    grid-template-columns: 10px 1fr auto;
    gap: 8px;
    align-items: center;
  }

  .timeline-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .timeline-text {
    color: #a0b0c0;
    font-size: 0.78rem;
  }

  .timeline-date {
    color: #5a6a7a;
    font-size: 0.74rem;
  }

  .hint {
    color: #7a8a9a;
    font-size: 0.78rem;
    margin: 8px 0 0;
    line-height: 1.45;
  }

  .btn {
    border: 1px solid #3a4a5a;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.82rem;
    padding: 8px 14px;
    transition: background-color 0.15s, border-color 0.15s, color 0.15s;
  }

  .btn:disabled {
    opacity: 0.56;
    cursor: not-allowed;
  }

  .btn-primary {
    background: #2a5090;
    border-color: #3a60a0;
    color: #e0e0e0;
  }

  .btn-primary:hover {
    background: #3a60a0;
  }

  .btn-secondary {
    background: #244364;
    border-color: #2e557c;
    color: #e0e0e0;
  }

  .btn-secondary:hover {
    background: #2e557c;
  }

  .btn-ghost {
    background: transparent;
    border-color: #3a4a5a;
    color: #7a8a9a;
  }

  .btn-ghost:hover {
    background: #2a3a4a;
    color: #c0d0e0;
  }

  .btn-danger {
    background: transparent;
    border-color: #69343a;
    color: #d27272;
  }

  .btn-danger:hover {
    background: #3a2024;
  }

  .btn-cancel {
    background: transparent;
    border-color: #3a4a5a;
    color: #7a8a9a;
  }

  .btn-cancel:hover {
    background: #2a3a4a;
    color: #c0d0e0;
  }

  .btn-small {
    padding: 4px 10px;
    font-size: 0.74rem;
  }

  .version-modal {
    width: min(520px, 92vw);
    background: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 8px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .version-modal h3 {
    margin: 0;
    color: #e0e0e0;
    font-size: 1rem;
  }

  .version-choice-group {
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .version-choice-title {
    margin: 0;
    color: #a0b0c0;
    font-size: 0.82rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .choice-option {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #c0d0e0;
    font-size: 0.86rem;
  }

  .version-choice-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 4px;
  }

  @media (max-width: 980px) {
    .editor-body,
    .form-grid,
    .group-grid {
      grid-template-columns: 1fr;
    }

    .group-card-header {
      flex-direction: column;
      align-items: flex-start;
    }
  }
</style>
