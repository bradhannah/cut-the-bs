<script lang="ts">
  import { onMount, tick } from "svelte";
  import {
    listSkillsByCategory,
    createSkill,
    updateSkill,
    deleteSkill,
    checkSkillLensReferences,
    splitSkillsText,
    getCompetenceLevels,
    listSkillCategories,
    createSkillCategory,
    renameSkillCategory,
    deleteSkillCategory,
    reorderSkillCategories,
    listLenses,
    getSkillLensTags,
    setSkillLensTags,
    addToast,
    type Skill,
    type SkillInput,
    type SkillCategory,
    type SkillCategoryWithSkills,
    type CompetenceLevel,
    type Lens,
  } from "../services/api";
  import DragHandle from "../components/DragHandle.svelte";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";

  let categoriesWithSkills: SkillCategoryWithSkills[] = [];
  let categories: SkillCategory[] = [];
  let competenceLevels: CompetenceLevel[] = [];
  let loading = true;
  let pageRootEl: HTMLDivElement | null = null;

  // Mass edit mode
  let massEditMode = false;

  // Collapsible categories (persisted in localStorage)
  const COLLAPSED_KEY = "skills-collapsed-categories";
  let collapsedCategories: Set<number> = new Set(
    JSON.parse(localStorage.getItem(COLLAPSED_KEY) || "[]")
  );

  function toggleCategory(categoryId: number): void {
    const next = new Set(collapsedCategories);
    if (next.has(categoryId)) {
      next.delete(categoryId);
    } else {
      next.add(categoryId);
    }
    collapsedCategories = next;
    localStorage.setItem(COLLAPSED_KEY, JSON.stringify([...next]));
  }

  // Skill form state
  let showSkillForm = false;
  let editingSkill: Skill | null = null;
  let formName = "";
  let formCategoryId = 0;
  let formCompetenceLevel = 5;
  let formIsLegacy = false;

  // Category form state
  let showCategoryForm = false;
  let editingCategory: SkillCategory | null = null;
  let categoryName = "";

  // Paste dialog state
  let showPasteDialog = false;
  let pasteText = "";
  let pastePreview: string[] = [];
  let pasteCategoryId = 0;
  let pasteCompetenceLevel = 5;

  // Delete confirmation state
  let deleteConfirmSkill: Skill | null = null;
  let lensReferences: string[] = [];

  // Lens tags state
  let lenses: Lens[] = [];
  let skillLensTagIds: Set<number> = new Set();
  let lensTagRequestToken = 0;

  onMount(async () => {
    await Promise.all([loadData(), loadCompetenceLevels(), loadLenses()]);
  });

  async function loadData(showLoading = true): Promise<void> {
    if (showLoading) {
      loading = true;
    }
    try {
      const [cats, catList] = await Promise.all([
        listSkillsByCategory(),
        listSkillCategories(),
      ]);
      categoriesWithSkills = cats || [];
      categories = catList || [];
    } finally {
      if (showLoading) {
        loading = false;
      }
    }
  }

  function getContentScrollContainer(): HTMLElement | null {
    if (!pageRootEl) {
      return null;
    }

    const container = pageRootEl.closest(".content");
    return container instanceof HTMLElement ? container : null;
  }

  async function loadDataPreservingScroll(showLoading = false): Promise<void> {
    const container = getContentScrollContainer();
    const scrollTop = container ? container.scrollTop : 0;

    await loadData(showLoading);
    await tick();

    if (container) {
      container.scrollTop = scrollTop;
    }
  }

  async function loadCompetenceLevels(): Promise<void> {
    try {
      competenceLevels = (await getCompetenceLevels()) || [];
    } catch {
      // Toast already shown
    }
  }

  async function loadLenses(): Promise<void> {
    try {
      lenses = (await listLenses()) || [];
    } catch {
      // Toast already shown
    }
  }

  // --- Skill CRUD ---

  function openAddSkill(): void {
    lensTagRequestToken += 1;
    editingSkill = null;
    deleteConfirmSkill = null;
    lensReferences = [];
    formName = "";
    formCategoryId = categories.length > 0 ? categories[0].id : 0;
    formCompetenceLevel = 5;
    formIsLegacy = false;
    skillLensTagIds = new Set();
    showSkillForm = true;
  }

  async function openEditSkill(skill: Skill): Promise<void> {
    const requestToken = ++lensTagRequestToken;
    deleteConfirmSkill = null;
    lensReferences = [];
    editingSkill = skill;
    formName = skill.name;
    formCategoryId = skill.category_id;
    formCompetenceLevel = skill.competence_level;
    formIsLegacy = skill.is_legacy;
    showSkillForm = true;

    // Load lens tags for this skill
    if (lenses.length > 0) {
      try {
        const tagIds = await getSkillLensTags(skill.id);
        if (
          requestToken !== lensTagRequestToken ||
          !editingSkill ||
          editingSkill.id !== skill.id
        ) {
          return;
        }
        skillLensTagIds = new Set(tagIds || []);
      } catch {
        if (
          requestToken !== lensTagRequestToken ||
          !editingSkill ||
          editingSkill.id !== skill.id
        ) {
          return;
        }
        skillLensTagIds = new Set();
      }
    }
  }

  function cancelSkillForm(): void {
    lensTagRequestToken += 1;
    showSkillForm = false;
    editingSkill = null;
    deleteConfirmSkill = null;
    lensReferences = [];
    skillLensTagIds = new Set();
  }

  function findSkillById(skillId: number): Skill | null {
    for (const group of categoriesWithSkills) {
      const found = group.skills.find((skill) => skill.id === skillId);
      if (found) {
        return found;
      }
    }
    return null;
  }

  async function handleSkillSubmit(): Promise<void> {
    const input: SkillInput = {
      name: formName.trim(),
      category_id: formCategoryId,
      competence_level: formCompetenceLevel,
      is_legacy: formIsLegacy,
    };

    if (!input.name) {
      addToast("error", "Skill name is required");
      return;
    }
    if (!input.category_id) {
      addToast("error", "Please select a category");
      return;
    }

    try {
      let savedSkill: Skill;
      if (editingSkill) {
        savedSkill = await updateSkill(editingSkill.id, input);
      } else {
        savedSkill = await createSkill(input);
      }

      // Save lens tags if there are lenses
      if (lenses.length > 0) {
        await setSkillLensTags(savedSkill.id, [...skillLensTagIds]);
      }

      await loadDataPreservingScroll(false);

      const refreshedSkill = findSkillById(savedSkill.id);
      const formSkill = refreshedSkill ?? savedSkill;
      editingSkill = formSkill;
      formName = formSkill.name;
      formCategoryId = formSkill.category_id;
      formCompetenceLevel = formSkill.competence_level;
      formIsLegacy = formSkill.is_legacy;
      showSkillForm = true;
    } catch {
      // Toast already shown
    }
  }

  async function confirmDeleteSkill(skill: Skill): Promise<void> {
    try {
      lensReferences = (await checkSkillLensReferences(skill.id)) || [];
      deleteConfirmSkill = skill;
    } catch {
      // Toast already shown
    }
  }

  async function handleDeleteSkill(): Promise<void> {
    if (!deleteConfirmSkill) return;
    const deletedSkillId = deleteConfirmSkill.id;
    const deletedSkillName = deleteConfirmSkill.name;
    try {
      await deleteSkill(deletedSkillId, deletedSkillName);
      deleteConfirmSkill = null;
      lensReferences = [];
      await loadDataPreservingScroll(false);

      if (editingSkill && editingSkill.id === deletedSkillId) {
        cancelSkillForm();
      }
    } catch {
      // Toast already shown
    }
  }

  async function requestDeleteSkill(skill: Skill): Promise<void> {
    if (!showSkillForm || !editingSkill || editingSkill.id !== skill.id) {
      await openEditSkill(skill);
    }

    await confirmDeleteSkill(skill);
  }

  function cancelDelete(): void {
    deleteConfirmSkill = null;
    lensReferences = [];
  }

  // --- Category CRUD ---

  function openAddCategory(): void {
    editingCategory = null;
    categoryName = "";
    showCategoryForm = true;
  }

  function openEditCategory(cat: SkillCategory): void {
    editingCategory = cat;
    categoryName = cat.name;
    showCategoryForm = true;
  }

  function cancelCategoryForm(): void {
    showCategoryForm = false;
    editingCategory = null;
  }

  async function handleCategorySubmit(): Promise<void> {
    const name = categoryName.trim();
    if (!name) {
      addToast("error", "Category name is required");
      return;
    }

    try {
      if (editingCategory) {
        await renameSkillCategory(editingCategory.id, name);
      } else {
        await createSkillCategory(name);
      }
      showCategoryForm = false;
      editingCategory = null;
      await loadData();
    } catch {
      // Toast already shown
    }
  }

  async function handleDeleteCategory(id: number): Promise<void> {
    try {
      await deleteSkillCategory(id);
      await loadData();
    } catch {
      // Toast already shown
    }
  }

  async function handleCategoryReorder(orderedIDs: number[]): Promise<void> {
    try {
      await reorderSkillCategories(orderedIDs);
      await loadData();
    } catch {
      // Toast already shown
    }
  }

  // --- Paste Dialog ---

  function openPasteDialog(): void {
    pasteText = "";
    pastePreview = [];
    pasteCategoryId = categories.length > 0 ? categories[0].id : 0;
    pasteCompetenceLevel = 5;
    showPasteDialog = true;
  }

  async function handlePastePreview(): Promise<void> {
    try {
      pastePreview = (await splitSkillsText(pasteText)) || [];
    } catch {
      // Toast already shown
    }
  }

  async function handlePasteConfirm(): Promise<void> {
    if (!pasteCategoryId) {
      addToast("error", "Please select a category");
      return;
    }
    showPasteDialog = false;
    try {
      for (const name of pastePreview) {
        await createSkill({
          name,
          category_id: pasteCategoryId,
          competence_level: pasteCompetenceLevel,
          is_legacy: false,
        });
      }
      addToast(
        "success",
        `Added ${pastePreview.length} skill${pastePreview.length !== 1 ? "s" : ""}`
      );
      await loadData();
    } catch {
      await loadData();
    }
  }

  function closePasteDialog(): void {
    showPasteDialog = false;
  }

  // Helpers
  function competenceLabel(level: number): string {
    const cl = competenceLevels.find((c) => c.level === level);
    return cl ? cl.label : String(level);
  }

  function getCategory(id: number): SkillCategory {
    return categories.find((c) => c.id === id) as SkillCategory;
  }

  function toggleLensTag(lensId: number): void {
    const next = new Set(skillLensTagIds);
    if (next.has(lensId)) {
      next.delete(lensId);
    } else {
      next.add(lensId);
    }
    skillLensTagIds = next;
  }

  $: sortedCategories = [...categories].sort(
    (a, b) => a.sort_order - b.sort_order
  );

  // --- Mass Edit Auto-Save ---

  async function handleMassEditCompetence(skill: Skill, level: number): Promise<void> {
    if (level === skill.competence_level) return;
    try {
      await updateSkill(skill.id, {
        name: skill.name,
        category_id: skill.category_id,
        competence_level: level,
        is_legacy: skill.is_legacy,
      });
      // Update local state without full reload for responsiveness
      skill.competence_level = level;
      categoriesWithSkills = categoriesWithSkills;
    } catch {
      // Toast already shown
    }
  }

  async function handleMassEditLegacy(skill: Skill, legacy: boolean): Promise<void> {
    if (legacy === skill.is_legacy) return;
    try {
      await updateSkill(skill.id, {
        name: skill.name,
        category_id: skill.category_id,
        competence_level: skill.competence_level,
        is_legacy: legacy,
      });
      // Update local state without full reload for responsiveness
      skill.is_legacy = legacy;
      categoriesWithSkills = categoriesWithSkills;
    } catch {
      // Toast already shown
    }
  }
</script>

<div class="skills-page" bind:this={pageRootEl}>
  <div class="page-header">
    <h2>Skills</h2>
    <div class="header-actions">
      <button
        class="btn btn-small"
        class:btn-active={massEditMode}
        on:click={() => (massEditMode = !massEditMode)}
      >
        {massEditMode ? "Done Editing" : "Mass Edit"}
      </button>
      <button class="btn btn-small" on:click={openPasteDialog}>
        Paste Skills
      </button>
      <button class="btn btn-primary" on:click={openAddSkill}>
        + Add Skill
      </button>
    </div>
  </div>
  <p class="page-description">
    Manage your skills organized by category with competence levels.
  </p>

  <div class="skills-layout">
    <div class="skills-main">
      <!-- Category Management -->
      <section class="section">
        <div class="section-header">
          <h3 class="section-title">Categories</h3>
          <button class="btn btn-small" on:click={openAddCategory}>
            + Add Category
          </button>
        </div>

        {#if showCategoryForm}
          <div class="entry-form compact">
            <div class="form-row">
              <div class="form-field">
                <label class="form-label" for="cat-name">
                  {editingCategory ? "Rename Category" : "New Category"}
                </label>
                <input
                  id="cat-name"
                  type="text"
                  class="form-input"
                  bind:value={categoryName}
                  placeholder="e.g. Programming Languages"
                />
              </div>
            </div>
            <div class="form-actions">
              <button class="btn btn-primary" on:click={handleCategorySubmit}>
                {editingCategory ? "Rename" : "Create"}
              </button>
              <button class="btn btn-cancel" on:click={cancelCategoryForm}>
                Cancel
              </button>
            </div>
          </div>
        {/if}

        {#if sortedCategories.length > 0}
          <div class="category-list">
            <DragHandle
              items={sortedCategories}
              on:reorder={(e) => handleCategoryReorder(e.detail.orderedIDs)}
              let:item
            >
              {@const cat = getCategory(item.id)}
              <div class="category-chip">
                <span class="category-name">{cat.name}</span>
                <div class="category-actions">
                  <button
                    class="btn btn-small btn-ghost"
                    on:click={() => openEditCategory(cat)}
                  >
                    Rename
                  </button>
                  <button
                    class="btn btn-small btn-danger"
                    on:click={() => handleDeleteCategory(cat.id)}
                  >
                    Delete
                  </button>
                </div>
              </div>
            </DragHandle>
          </div>
        {/if}
      </section>

      <!-- Skills by Category -->
      {#if loading}
        <LoadingSpinner />
      {:else if categoriesWithSkills.length === 0}
        <div class="empty-state">
          <p>No skills yet.</p>
          <p class="empty-hint">Create a category first, then add skills to it.</p>
        </div>
      {:else}
        {#each categoriesWithSkills as group (group.category.id)}
          <section class="section">
            <button
              class="section-title-toggle"
              on:click={() => toggleCategory(group.category.id)}
            >
              <span
                class="chevron"
                class:collapsed={collapsedCategories.has(group.category.id)}
              >&#9660;</span>
              <h3 class="section-title">{group.category.name}</h3>
              <span class="skill-count-badge">{group.skills.length}</span>
            </button>
            {#if !collapsedCategories.has(group.category.id)}
              {#if group.skills.length === 0}
                <p class="empty-hint">No skills in this category.</p>
              {:else}
                <div class="skill-grid">
                  {#each group.skills as skill (skill.id)}
                    <div
                      class="skill-card"
                      class:legacy={skill.is_legacy}
                      class:selected={showSkillForm && editingSkill?.id === skill.id}
                    >
                      <div class="skill-info">
                        <span class="skill-name">{skill.name}</span>
                        {#if !massEditMode}
                          <span class="competence-badge level-{skill.competence_level}">
                            {skill.competence_level}/10
                          </span>
                          <span class="competence-rating-label">
                            {competenceLabel(skill.competence_level)}
                          </span>
                          {#if skill.is_legacy}
                            <span class="legacy-badge">Legacy</span>
                          {/if}
                        {/if}
                      </div>
                      {#if massEditMode}
                        <div class="mass-edit-controls">
                          <div class="mass-edit-slider">
                            <input
                              type="range"
                              min="1"
                              max="10"
                              value={skill.competence_level}
                              class="form-range"
                              on:change={(e) =>
                                handleMassEditCompetence(
                                  skill,
                                  parseInt(e.currentTarget.value)
                                )}
                            />
                            <div class="mass-edit-level-wrap">
                              <span class="mass-edit-level">{skill.competence_level}/10</span>
                              <span class="mass-edit-rating">
                                {competenceLabel(skill.competence_level)}
                              </span>
                            </div>
                          </div>
                          <label class="mass-edit-legacy">
                            <input
                              type="checkbox"
                              checked={skill.is_legacy}
                              on:change={(e) =>
                                handleMassEditLegacy(skill, e.currentTarget.checked)}
                            />
                            Legacy
                          </label>
                        </div>
                      {:else}
                        <div class="skill-actions">
                          <button
                            class="btn btn-small btn-ghost"
                            on:click={() => openEditSkill(skill)}
                          >
                            Edit
                          </button>
                          <button
                            class="btn btn-small btn-danger"
                            on:click={() => requestDeleteSkill(skill)}
                          >
                            Delete
                          </button>
                        </div>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            {/if}
          </section>
        {/each}
      {/if}
    </div>

    <aside class="skill-details-pane">
      <div class="skill-details-header">
        <h3>Skill Details</h3>
        {#if showSkillForm}
          <button class="btn btn-small" on:click={openAddSkill}>
            + New Skill
          </button>
        {/if}
      </div>

      {#if showSkillForm}
        <div class="entry-form detail-form">
          <h4 class="form-title">
            {editingSkill ? `Editing ${editingSkill.name}` : "New Skill"}
          </h4>
          <div class="form-row">
            <div class="form-field">
              <label class="form-label" for="skill-name">Name</label>
              <input
                id="skill-name"
                type="text"
                class="form-input"
                bind:value={formName}
                placeholder="e.g. TypeScript"
              />
            </div>
          </div>
          <div class="form-row">
            <div class="form-field">
              <label class="form-label" for="skill-category">Category</label>
              <select
                id="skill-category"
                class="form-input"
                bind:value={formCategoryId}
              >
                {#each sortedCategories as cat (cat.id)}
                  <option value={cat.id}>{cat.name}</option>
                {/each}
              </select>
            </div>
          </div>
          <div class="form-row">
            <div class="form-field">
              <label class="form-label" for="skill-competence">
                Competence Level ({formCompetenceLevel}/10)
              </label>
              <input
                id="skill-competence"
                type="range"
                min="1"
                max="10"
                class="form-range"
                bind:value={formCompetenceLevel}
              />
              <span class="competence-hint">
                {competenceLabel(formCompetenceLevel)}
              </span>
            </div>
          </div>

          <div class="form-row form-row-tight">
            <div class="form-field form-field-checkbox">
              <label class="form-label">
                <input type="checkbox" bind:checked={formIsLegacy} />
                Legacy skill (no longer actively used)
              </label>
            </div>
          </div>

          {#if lenses.length > 0}
            <fieldset class="form-field lens-tags-field">
              <legend class="form-label">Lens Tags</legend>
              <div class="lens-tag-list">
                {#each lenses as lens (lens.id)}
                  <label class="lens-tag-item">
                    <input
                      type="checkbox"
                      checked={skillLensTagIds.has(lens.id)}
                      on:change={() => toggleLensTag(lens.id)}
                    />
                    <span class="lens-tag-label">{lens.name}</span>
                  </label>
                {/each}
              </div>
            </fieldset>
          {/if}

          <div class="form-actions detail-form-actions">
            <button class="btn btn-primary" on:click={handleSkillSubmit}>
              {editingSkill ? "Save Changes" : "Create Skill"}
            </button>
            {#if editingSkill}
              <button
                class="btn btn-danger"
                on:click={() => confirmDeleteSkill(editingSkill)}
              >
                Delete
              </button>
            {/if}
            <button class="btn btn-cancel" on:click={cancelSkillForm}>
              Close
            </button>
          </div>

          {#if deleteConfirmSkill && editingSkill && deleteConfirmSkill.id === editingSkill.id}
            <div class="confirm-dialog detail-delete-confirm">
              <p>
                Delete skill <strong>{deleteConfirmSkill.name}</strong>?
              </p>
              <p class="confirm-note">
                This only deletes the selected skill.
              </p>
              {#if lensReferences.length > 0}
                <p class="warn-text">
                  This skill is referenced by {lensReferences.length} lens{lensReferences.length !==
                  1
                    ? "es"
                    : ""}:
                  {lensReferences.join(", ")}
                </p>
              {/if}
              <div class="form-actions">
                <button class="btn btn-danger-solid" on:click={handleDeleteSkill}>
                  Confirm Delete
                </button>
                <button class="btn btn-cancel" on:click={cancelDelete}>Cancel</button>
              </div>
            </div>
          {/if}
        </div>
      {:else}
        <div class="detail-empty">
          <p class="detail-empty-title">Select a skill to edit it here.</p>
          <p class="detail-empty-hint">
            The detail pane stays visible so you can edit without scrolling back
            up.
          </p>
          <button class="btn btn-primary" on:click={openAddSkill}>
            + Add Skill
          </button>
        </div>
      {/if}
    </aside>
  </div>

  <!-- Paste Dialog -->
  {#if showPasteDialog}
    <div class="modal-overlay">
      <div class="modal">
        <h3 class="form-title">Paste Skills</h3>
        <p class="modal-hint">
          Enter comma-separated skill names. They will be split and previewed
          before creation.
        </p>
        <div class="form-field">
          <label class="form-label" for="paste-text"
            >Skills (comma-separated)</label
          >
          <textarea
            id="paste-text"
            class="form-input form-textarea"
            bind:value={pasteText}
            placeholder="React, TypeScript, Node.js, PostgreSQL"
            rows="4"
          />
        </div>
        <div class="form-row">
          <div class="form-field">
            <label class="form-label" for="paste-category">Category</label>
            <select
              id="paste-category"
              class="form-input"
              bind:value={pasteCategoryId}
            >
              {#each sortedCategories as cat (cat.id)}
                <option value={cat.id}>{cat.name}</option>
              {/each}
            </select>
          </div>
          <div class="form-field">
            <label class="form-label" for="paste-competence">
              Competence ({pasteCompetenceLevel}/10)
            </label>
            <input
              id="paste-competence"
              type="range"
              min="1"
              max="10"
              class="form-range"
              bind:value={pasteCompetenceLevel}
            />
          </div>
        </div>
        <button class="btn btn-small" on:click={handlePastePreview}>
          Preview
        </button>
        {#if pastePreview.length > 0}
          <div class="paste-preview">
            <p class="preview-count">
              {pastePreview.length} skill{pastePreview.length !== 1 ? "s" : ""} to
              add:
            </p>
            <ul class="preview-list">
              {#each pastePreview as name, i (i)}
                <li>{name}</li>
              {/each}
            </ul>
          </div>
        {/if}
        <div class="form-actions">
          <button
            class="btn btn-primary"
            on:click={handlePasteConfirm}
            disabled={pastePreview.length === 0}
          >
            Add All
          </button>
          <button class="btn btn-cancel" on:click={closePasteDialog}>
            Cancel
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .skills-page {
    min-width: 0;
  }

  .skills-layout {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 390px;
    gap: 20px;
    align-items: start;
  }

  .skills-main {
    min-width: 0;
  }

  .skill-details-pane {
    position: sticky;
    top: 0;
    align-self: start;
    max-height: calc(100vh - 120px);
    overflow-y: auto;
    padding-right: 2px;
  }

  .skill-details-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 10px;
  }

  .skill-details-header h3 {
    margin: 0;
    font-size: 1rem;
    color: #c0d0e0;
    letter-spacing: 0.01em;
  }

  .detail-form {
    margin-bottom: 0;
  }

  .detail-form-actions {
    flex-wrap: wrap;
  }

  .detail-empty {
    background-color: #1e2d3d;
    border: 1px solid #3a4a5a;
    border-radius: 6px;
    padding: 18px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .detail-empty-title {
    margin: 0;
    color: #d0dce8;
    font-size: 0.92rem;
    font-weight: 600;
  }

  .detail-empty-hint {
    margin: 0;
    color: #7a8a9a;
    font-size: 0.84rem;
    line-height: 1.4;
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

  .header-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    justify-content: flex-end;
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

  /* --- Sections --- */
  .section {
    margin-bottom: 24px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .section-title {
    margin: 0 0 8px;
    font-size: 1.1rem;
    color: #c0d0e0;
    font-weight: 600;
  }

  .section-header .section-title {
    margin-bottom: 0;
  }

  /* --- Collapsible Category Toggle --- */
  .section-title-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    background: none;
    border: none;
    padding: 4px 0;
    margin: 0 0 8px;
    cursor: pointer;
    width: 100%;
    text-align: left;
  }

  .section-title-toggle .section-title {
    margin: 0;
  }

  .chevron {
    font-size: 0.7rem;
    color: #5a6a7a;
    transition: transform 0.15s ease;
    flex-shrink: 0;
  }

  .chevron.collapsed {
    transform: rotate(-90deg);
  }

  .skill-count-badge {
    font-size: 0.7rem;
    padding: 1px 7px;
    border-radius: 10px;
    background-color: #2a3a4a;
    color: #7a8a9a;
    font-weight: 600;
    flex-shrink: 0;
  }

  /* --- Forms --- */
  .entry-form {
    background-color: #1e2d3d;
    border: 1px solid #3a4a5a;
    border-radius: 6px;
    padding: 20px;
    margin-bottom: 24px;
  }

  .entry-form.compact {
    padding: 16px;
    margin-bottom: 16px;
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

  .form-row-tight {
    margin-bottom: 8px;
  }

  .form-field {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .form-field-checkbox {
    justify-content: center;
  }

  .form-field-checkbox .form-label {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    text-transform: none;
    font-weight: 400;
    font-size: 0.85rem;
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
  }

  .form-range {
    width: 100%;
    accent-color: #4a8af4;
  }

  .competence-hint {
    font-size: 0.8rem;
    color: #7a8a9a;
  }

  .form-actions {
    display: flex;
    gap: 8px;
    margin-top: 8px;
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

  .btn-danger-solid {
    background-color: #802020;
    border-color: #a03030;
    color: #e0e0e0;
  }

  .btn-danger-solid:hover {
    background-color: #a03030;
  }

  /* --- Category List --- */
  .category-list {
    display: flex;
    flex-direction: column;
    margin-bottom: 16px;
  }

  .category-chip {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    margin-bottom: 4px;
  }

  .category-name {
    font-size: 0.9rem;
    color: #e0e0e0;
    font-weight: 600;
  }

  .category-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  /* --- Skill Grid --- */
  .skill-grid {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .skill-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px;
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    transition:
      border-color 0.14s,
      box-shadow 0.14s;
  }

  .skill-card.selected {
    border-color: #4a8af4;
    box-shadow: 0 0 0 1px rgba(74, 138, 244, 0.2);
  }

  .skill-card.legacy {
    opacity: 0.6;
  }

  .skill-info {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }

  .skill-name {
    font-size: 0.9rem;
    color: #e0e0e0;
  }

  .competence-badge {
    font-size: 0.75rem;
    padding: 2px 8px;
    border-radius: 10px;
    font-weight: 600;
    background-color: #2a3a4a;
    color: #a0b0c0;
  }

  .competence-rating-label {
    font-size: 0.76rem;
    color: #8fa4ba;
    letter-spacing: 0.01em;
  }

  .legacy-badge {
    font-size: 0.7rem;
    padding: 2px 6px;
    border-radius: 8px;
    background-color: #3a3020;
    color: #c0a060;
  }

  .skill-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  /* --- Mass Edit Mode --- */
  .btn-active {
    background-color: #2a5090;
    border-color: #3a60a0;
    color: #e0e0e0;
  }

  .btn-active:hover {
    background-color: #3a60a0;
  }

  .mass-edit-controls {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-shrink: 0;
  }

  .mass-edit-slider {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .mass-edit-level-wrap {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 82px;
  }

  .mass-edit-slider .form-range {
    width: 120px;
    accent-color: #4a8af4;
  }

  .mass-edit-level {
    font-size: 0.8rem;
    color: #a0b0c0;
    font-weight: 600;
    min-width: 32px;
  }

  .mass-edit-rating {
    font-size: 0.72rem;
    color: #8fa4ba;
    line-height: 1.15;
    white-space: nowrap;
  }

  .mass-edit-legacy {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.8rem;
    color: #7a8a9a;
    cursor: pointer;
    white-space: nowrap;
  }

  /* --- Confirm Dialog --- */
  .confirm-dialog {
    background-color: #2a1a1a;
    border: 1px solid #5a3030;
    border-radius: 6px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .detail-delete-confirm {
    margin-top: 12px;
    margin-bottom: 0;
  }

  .confirm-dialog p {
    margin: 0 0 8px;
    color: #e0e0e0;
    font-size: 0.9rem;
  }

  .confirm-note {
    color: #9fb2c7 !important;
    font-size: 0.82rem !important;
  }

  .warn-text {
    color: #e0a060 !important;
    font-size: 0.85rem !important;
  }

  /* --- Modal --- */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal {
    background-color: #1e2d3d;
    border: 1px solid #3a4a5a;
    border-radius: 8px;
    padding: 24px;
    width: 500px;
    max-width: 90vw;
    max-height: 80vh;
    overflow-y: auto;
  }

  .modal-hint {
    color: #7a8a9a;
    font-size: 0.85rem;
    margin: 0 0 16px;
  }

  .paste-preview {
    margin-top: 12px;
    padding: 12px;
    background-color: #1a2332;
    border-radius: 4px;
  }

  .preview-count {
    font-size: 0.85rem;
    color: #7a8a9a;
    margin: 0 0 8px;
  }

  .preview-list {
    margin: 0;
    padding-left: 20px;
    color: #e0e0e0;
    font-size: 0.85rem;
  }

  .preview-list li {
    margin-bottom: 4px;
  }

  /* --- Lens Tags --- */
  .lens-tags-field {
    margin-bottom: 12px;
  }

  .lens-tag-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 4px;
  }

  .lens-tag-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    background-color: #1a2332;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    cursor: pointer;
    transition:
      background-color 0.15s,
      border-color 0.15s;
  }

  .lens-tag-item:hover {
    background-color: #2a3a4a;
    border-color: #3a4a5a;
  }

  .lens-tag-label {
    font-size: 0.85rem;
    color: #c0d0e0;
    user-select: none;
  }

  @media (max-width: 1120px) {
    .skills-layout {
      grid-template-columns: 1fr;
    }

    .skill-details-pane {
      position: static;
      max-height: none;
      overflow: visible;
      padding-right: 0;
    }
  }

  @media (max-width: 760px) {
    .page-header {
      align-items: flex-start;
      flex-direction: column;
      gap: 12px;
    }

    .header-actions {
      justify-content: flex-start;
    }

    .form-row {
      flex-direction: column;
      gap: 10px;
    }

    .skill-card {
      align-items: flex-start;
      flex-direction: column;
      gap: 10px;
    }

    .skill-actions,
    .mass-edit-controls {
      width: 100%;
      justify-content: flex-start;
    }
  }
</style>
