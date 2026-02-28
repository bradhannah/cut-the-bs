<script lang="ts">
  import { onMount } from "svelte";
  import {
    getProfile,
    listProfileLinks,
    listWorkHistory,
    listSkillsByCategory,
    listAcademicCredentials,
    listCertifications,
    listSummaries,
    listDescriptors,
    listCoreExpertise,
    addToast,
    type UserProfile,
    type ProfileLink,
    type WorkHistoryEntry,
    type SkillCategoryWithSkills,
    type AcademicCredential,
    type Certification,
    type ProfessionalSummary,
    type RoleDescriptor,
    type CoreExpertise,
  } from "../services/api";
  import { ClipboardSetText } from "../../wailsjs/runtime/runtime";
  import LoadingSpinner from "../components/LoadingSpinner.svelte";

  type DateFormat = "long" | "short" | "numeric" | "iso";

  const dateFormats: { value: DateFormat; label: string; example: string }[] = [
    { value: "long", label: "January 2023", example: "January 2023" },
    { value: "short", label: "Jan 2023", example: "Jan 2023" },
    { value: "numeric", label: "01/2023", example: "01/2023" },
    { value: "iso", label: "2023-01", example: "2023-01" },
  ];

  let dateFormat: DateFormat = "long";

  let loading = true;
  let profile: UserProfile | null = null;
  let profileLinks: ProfileLink[] = [];
  let workHistory: WorkHistoryEntry[] = [];
  let skillCategories: SkillCategoryWithSkills[] = [];
  let academics: AcademicCredential[] = [];
  let certifications: Certification[] = [];
  let summaries: ProfessionalSummary[] = [];
  let descriptors: RoleDescriptor[] = [];
  let coreExpertise: CoreExpertise[] = [];

  let copiedKey: string | null = null;
  let copiedTimeout: ReturnType<typeof setTimeout> | null = null;

  // Collapsible section state — all open by default.
  let sectionsOpen: Record<string, boolean> = {
    personal: true,
    links: true,
    work: true,
    education: true,
    certifications: true,
    skills: true,
    summaries: true,
    descriptors: true,
    expertise: true,
  };

  const longMonths = [
    "January", "February", "March", "April", "May", "June",
    "July", "August", "September", "October", "November", "December",
  ];
  const shortMonths = [
    "Jan", "Feb", "Mar", "Apr", "May", "Jun",
    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
  ];

  onMount(async () => {
    try {
      const results = await Promise.all([
        getProfile(),
        listProfileLinks(),
        listWorkHistory(),
        listSkillsByCategory(),
        listAcademicCredentials(),
        listCertifications(),
        listSummaries(),
        listDescriptors(),
        listCoreExpertise(),
      ]);
      profile = results[0];
      profileLinks = results[1];
      workHistory = results[2];
      skillCategories = results[3];
      academics = results[4];
      certifications = results[5];
      summaries = results[6];
      descriptors = results[7];
      coreExpertise = results[8];
    } catch {
      addToast("error", "Failed to load helper data");
    } finally {
      loading = false;
    }
  });

  function formatDateValue(raw: string, granularity: string): string {
    if (!raw) return "";
    const parts = raw.split("-");
    const year = parts[0] || "";
    const monthNum = parseInt(parts[1] || "0", 10);

    if (granularity === "year" || !monthNum) {
      return year;
    }

    const dayNum = parseInt(parts[2] || "0", 10);

    switch (dateFormat) {
      case "long": {
        const base = `${longMonths[monthNum - 1]} ${year}`;
        return dayNum && granularity === "day" ? `${longMonths[monthNum - 1]} ${dayNum}, ${year}` : base;
      }
      case "short": {
        const base = `${shortMonths[monthNum - 1]} ${year}`;
        return dayNum && granularity === "day" ? `${shortMonths[monthNum - 1]} ${dayNum}, ${year}` : base;
      }
      case "numeric": {
        const mm = String(monthNum).padStart(2, "0");
        if (dayNum && granularity === "day") {
          return `${mm}/${String(dayNum).padStart(2, "0")}/${year}`;
        }
        return `${mm}/${year}`;
      }
      case "iso": {
        const mm = String(monthNum).padStart(2, "0");
        if (dayNum && granularity === "day") {
          return `${year}-${mm}-${String(dayNum).padStart(2, "0")}`;
        }
        return `${year}-${mm}`;
      }
      default:
        return raw;
    }
  }

  async function copyToClipboard(value: string, key: string): Promise<void> {
    if (!value) return;
    try {
      await ClipboardSetText(value);
      copiedKey = key;
      if (copiedTimeout) clearTimeout(copiedTimeout);
      copiedTimeout = setTimeout(() => {
        copiedKey = null;
      }, 1200);
    } catch {
      addToast("error", "Failed to copy to clipboard");
    }
  }

  function toggleSection(name: string): void {
    sectionsOpen[name] = !sectionsOpen[name];
  }

  function allSkillNames(): string {
    return skillCategories
      .flatMap((cat) => cat.skills)
      .filter((s) => !s.is_legacy)
      .map((s) => s.name)
      .join(", ");
  }

  function allCoreExpertiseLabels(): string {
    return coreExpertise.map((e) => e.label).join(" | ");
  }

  function allDescriptorTitles(): string {
    return descriptors.map((d) => d.title).join(", ");
  }

  function workBulletsCombined(entry: WorkHistoryEntry): string {
    return entry.bullets.map((b) => `- ${b.text}`).join("\n");
  }
</script>

<div class="helper-page">
  <div class="page-header">
    <div class="page-header-left">
      <h2>Application Helper</h2>
      <p class="page-description">
        Click any value to copy it to your clipboard. Use this while filling out job application portals.
      </p>
    </div>
    <div class="header-actions">
      <label class="date-format-label" for="date-fmt">Date format</label>
      <select
        id="date-fmt"
        class="date-format-select"
        bind:value={dateFormat}
      >
        {#each dateFormats as fmt (fmt.value)}
          <option value={fmt.value}>{fmt.label}</option>
        {/each}
      </select>
    </div>
  </div>

  {#if loading}
    <LoadingSpinner />
  {:else}
    <!-- Personal Info -->
    <section class="helper-section">
      <button class="section-toggle" on:click={() => toggleSection("personal")}>
        <span class="toggle-arrow" class:open={sectionsOpen.personal}>&#9654;</span>
        <h3>Personal Info</h3>
      </button>
      {#if sectionsOpen.personal && profile}
        <div class="copy-grid">
          {#if profile.full_name}
            <button
              class="copy-cell"
              class:copied={copiedKey === "profile-name"}
              on:click={() => copyToClipboard(profile.full_name, "profile-name")}
              title="Copy full name"
            >
              <span class="cell-label">Full Name</span>
              <span class="cell-value">{profile.full_name}</span>
              <span class="cell-status">{copiedKey === "profile-name" ? "Copied" : ""}</span>
            </button>
          {/if}
          {#if profile.email}
            <button
              class="copy-cell"
              class:copied={copiedKey === "profile-email"}
              on:click={() => copyToClipboard(profile.email, "profile-email")}
              title="Copy email"
            >
              <span class="cell-label">Email</span>
              <span class="cell-value">{profile.email}</span>
              <span class="cell-status">{copiedKey === "profile-email" ? "Copied" : ""}</span>
            </button>
          {/if}
          {#if profile.phone}
            <button
              class="copy-cell"
              class:copied={copiedKey === "profile-phone"}
              on:click={() => copyToClipboard(profile.phone, "profile-phone")}
              title="Copy phone"
            >
              <span class="cell-label">Phone</span>
              <span class="cell-value">{profile.phone}</span>
              <span class="cell-status">{copiedKey === "profile-phone" ? "Copied" : ""}</span>
            </button>
          {/if}
          {#if profile.location}
            <button
              class="copy-cell"
              class:copied={copiedKey === "profile-location"}
              on:click={() => copyToClipboard(profile.location, "profile-location")}
              title="Copy location"
            >
              <span class="cell-label">Location</span>
              <span class="cell-value">{profile.location}</span>
              <span class="cell-status">{copiedKey === "profile-location" ? "Copied" : ""}</span>
            </button>
          {/if}
        </div>
      {/if}
    </section>

    <!-- Links -->
    {#if profileLinks.length > 0}
      <section class="helper-section">
        <button class="section-toggle" on:click={() => toggleSection("links")}>
          <span class="toggle-arrow" class:open={sectionsOpen.links}>&#9654;</span>
          <h3>Links</h3>
        </button>
        {#if sectionsOpen.links}
          <div class="copy-grid">
            {#each profileLinks as pl (pl.id)}
              <button
                class="copy-cell"
                class:copied={copiedKey === `link-${pl.id}`}
                on:click={() => copyToClipboard(pl.url, `link-${pl.id}`)}
                title="Copy {pl.label} URL"
              >
                <span class="cell-label">{pl.label}</span>
                <span class="cell-value url-value">{pl.url}</span>
                <span class="cell-status">{copiedKey === `link-${pl.id}` ? "Copied" : ""}</span>
              </button>
            {/each}
          </div>
        {/if}
      </section>
    {/if}

    <!-- Work History -->
    {#if workHistory.length > 0}
      <section class="helper-section">
        <button class="section-toggle" on:click={() => toggleSection("work")}>
          <span class="toggle-arrow" class:open={sectionsOpen.work}>&#9654;</span>
          <h3>Work History</h3>
        </button>
        {#if sectionsOpen.work}
          {#each workHistory as entry (entry.id)}
            <div class="work-card">
              <div class="work-card-header">
                <span class="work-employer">{entry.employer_name}</span>
                <span class="work-title-sep">&mdash;</span>
                <span class="work-jobtitle">{entry.job_title}</span>
              </div>
              <div class="copy-grid">
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `work-employer-${entry.id}`}
                  on:click={() => copyToClipboard(entry.employer_name, `work-employer-${entry.id}`)}
                  title="Copy employer"
                >
                  <span class="cell-label">Employer</span>
                  <span class="cell-value">{entry.employer_name}</span>
                  <span class="cell-status">{copiedKey === `work-employer-${entry.id}` ? "Copied" : ""}</span>
                </button>
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `work-title-${entry.id}`}
                  on:click={() => copyToClipboard(entry.job_title, `work-title-${entry.id}`)}
                  title="Copy job title"
                >
                  <span class="cell-label">Job Title</span>
                  <span class="cell-value">{entry.job_title}</span>
                  <span class="cell-status">{copiedKey === `work-title-${entry.id}` ? "Copied" : ""}</span>
                </button>
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `work-start-${entry.id}`}
                  on:click={() => copyToClipboard(formatDateValue(entry.start_date, entry.date_granularity_start), `work-start-${entry.id}`)}
                  title="Copy start date"
                >
                  <span class="cell-label">Start Date</span>
                  <span class="cell-value">{formatDateValue(entry.start_date, entry.date_granularity_start)}</span>
                  <span class="cell-status">{copiedKey === `work-start-${entry.id}` ? "Copied" : ""}</span>
                </button>
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `work-end-${entry.id}`}
                  on:click={() => copyToClipboard(entry.end_date ? formatDateValue(entry.end_date, entry.date_granularity_end) : "Present", `work-end-${entry.id}`)}
                  title="Copy end date"
                >
                  <span class="cell-label">End Date</span>
                  <span class="cell-value">{entry.end_date ? formatDateValue(entry.end_date, entry.date_granularity_end) : "Present"}</span>
                  <span class="cell-status">{copiedKey === `work-end-${entry.id}` ? "Copied" : ""}</span>
                </button>
                {#if entry.summary}
                  <button
                    class="copy-cell copy-cell-wide"
                    class:copied={copiedKey === `work-summary-${entry.id}`}
                    on:click={() => copyToClipboard(entry.summary, `work-summary-${entry.id}`)}
                    title="Copy summary"
                  >
                    <span class="cell-label">Summary</span>
                    <span class="cell-value text-value">{entry.summary}</span>
                    <span class="cell-status">{copiedKey === `work-summary-${entry.id}` ? "Copied" : ""}</span>
                  </button>
                {/if}
              </div>
              {#if entry.bullets.length > 0}
                <div class="bullets-section">
                  <div class="bullets-header">
                    <span class="bullets-label">Achievements / Responsibilities</span>
                    <button
                      class="copy-all-btn"
                      class:copied={copiedKey === `work-allbullets-${entry.id}`}
                      on:click={() => copyToClipboard(workBulletsCombined(entry), `work-allbullets-${entry.id}`)}
                    >
                      {copiedKey === `work-allbullets-${entry.id}` ? "Copied" : "Copy All"}
                    </button>
                  </div>
                  {#each entry.bullets as bullet (bullet.id)}
                    <button
                      class="copy-bullet"
                      class:copied={copiedKey === `bullet-${bullet.id}`}
                      on:click={() => copyToClipboard(bullet.text, `bullet-${bullet.id}`)}
                      title="Copy bullet"
                    >
                      <span class="bullet-dot">&bull;</span>
                      <span class="bullet-text">{bullet.text}</span>
                      <span class="cell-status">{copiedKey === `bullet-${bullet.id}` ? "Copied" : ""}</span>
                    </button>
                  {/each}
                </div>
              {/if}
            </div>
          {/each}
        {/if}
      </section>
    {/if}

    <!-- Education -->
    {#if academics.length > 0}
      <section class="helper-section">
        <button class="section-toggle" on:click={() => toggleSection("education")}>
          <span class="toggle-arrow" class:open={sectionsOpen.education}>&#9654;</span>
          <h3>Education</h3>
        </button>
        {#if sectionsOpen.education}
          {#each academics as ac (ac.id)}
            <div class="entry-card">
              <div class="copy-grid">
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `edu-inst-${ac.id}`}
                  on:click={() => copyToClipboard(ac.institution, `edu-inst-${ac.id}`)}
                  title="Copy institution"
                >
                  <span class="cell-label">Institution</span>
                  <span class="cell-value">{ac.institution}</span>
                  <span class="cell-status">{copiedKey === `edu-inst-${ac.id}` ? "Copied" : ""}</span>
                </button>
                {#if ac.credential_type}
                  <button
                    class="copy-cell"
                    class:copied={copiedKey === `edu-type-${ac.id}`}
                    on:click={() => copyToClipboard(ac.credential_type, `edu-type-${ac.id}`)}
                    title="Copy credential type"
                  >
                    <span class="cell-label">Degree</span>
                    <span class="cell-value">{ac.credential_type}</span>
                    <span class="cell-status">{copiedKey === `edu-type-${ac.id}` ? "Copied" : ""}</span>
                  </button>
                {/if}
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `edu-field-${ac.id}`}
                  on:click={() => copyToClipboard(ac.field_of_study, `edu-field-${ac.id}`)}
                  title="Copy field of study"
                >
                  <span class="cell-label">Field of Study</span>
                  <span class="cell-value">{ac.field_of_study}</span>
                  <span class="cell-status">{copiedKey === `edu-field-${ac.id}` ? "Copied" : ""}</span>
                </button>
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `edu-date-${ac.id}`}
                  on:click={() => copyToClipboard(formatDateValue(ac.completion_date, ac.date_granularity || "month"), `edu-date-${ac.id}`)}
                  title="Copy completion date"
                >
                  <span class="cell-label">Completion Date</span>
                  <span class="cell-value">{formatDateValue(ac.completion_date, ac.date_granularity || "month")}</span>
                  <span class="cell-status">{copiedKey === `edu-date-${ac.id}` ? "Copied" : ""}</span>
                </button>
                <button
                  class="copy-cell copy-cell-wide"
                  class:copied={copiedKey === `edu-full-${ac.id}`}
                  on:click={() => copyToClipboard(`${ac.credential_type ? ac.credential_type + ", " : ""}${ac.field_of_study} — ${ac.institution}`, `edu-full-${ac.id}`)}
                  title="Copy full credential line"
                >
                  <span class="cell-label">Full Line</span>
                  <span class="cell-value">{ac.credential_type ? ac.credential_type + ", " : ""}{ac.field_of_study} — {ac.institution}</span>
                  <span class="cell-status">{copiedKey === `edu-full-${ac.id}` ? "Copied" : ""}</span>
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </section>
    {/if}

    <!-- Certifications -->
    {#if certifications.length > 0}
      <section class="helper-section">
        <button class="section-toggle" on:click={() => toggleSection("certifications")}>
          <span class="toggle-arrow" class:open={sectionsOpen.certifications}>&#9654;</span>
          <h3>Certifications</h3>
        </button>
        {#if sectionsOpen.certifications}
          {#each certifications as cert (cert.id)}
            <div class="entry-card">
              <div class="copy-grid">
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `cert-name-${cert.id}`}
                  on:click={() => copyToClipboard(cert.name, `cert-name-${cert.id}`)}
                  title="Copy certification name"
                >
                  <span class="cell-label">Certification</span>
                  <span class="cell-value">{cert.name}</span>
                  <span class="cell-status">{copiedKey === `cert-name-${cert.id}` ? "Copied" : ""}</span>
                </button>
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `cert-issuer-${cert.id}`}
                  on:click={() => copyToClipboard(cert.issuing_body, `cert-issuer-${cert.id}`)}
                  title="Copy issuing body"
                >
                  <span class="cell-label">Issuing Body</span>
                  <span class="cell-value">{cert.issuing_body}</span>
                  <span class="cell-status">{copiedKey === `cert-issuer-${cert.id}` ? "Copied" : ""}</span>
                </button>
                <button
                  class="copy-cell"
                  class:copied={copiedKey === `cert-earned-${cert.id}`}
                  on:click={() => copyToClipboard(formatDateValue(cert.date_earned, "month"), `cert-earned-${cert.id}`)}
                  title="Copy date earned"
                >
                  <span class="cell-label">Date Earned</span>
                  <span class="cell-value">{formatDateValue(cert.date_earned, "month")}</span>
                  <span class="cell-status">{copiedKey === `cert-earned-${cert.id}` ? "Copied" : ""}</span>
                </button>
                {#if cert.expiration_date}
                  <button
                    class="copy-cell"
                    class:copied={copiedKey === `cert-exp-${cert.id}`}
                    on:click={() => copyToClipboard(formatDateValue(cert.expiration_date, "month"), `cert-exp-${cert.id}`)}
                    title="Copy expiration date"
                  >
                    <span class="cell-label">Expiration</span>
                    <span class="cell-value">{formatDateValue(cert.expiration_date, "month")}</span>
                    <span class="cell-status">{copiedKey === `cert-exp-${cert.id}` ? "Copied" : ""}</span>
                  </button>
                {/if}
              </div>
            </div>
          {/each}
        {/if}
      </section>
    {/if}

    <!-- Skills -->
    {#if skillCategories.length > 0}
      <section class="helper-section">
        <button class="section-toggle" on:click={() => toggleSection("skills")}>
          <span class="toggle-arrow" class:open={sectionsOpen.skills}>&#9654;</span>
          <h3>Skills</h3>
        </button>
        {#if sectionsOpen.skills}
          <div class="skills-actions">
            <button
              class="copy-all-btn"
              class:copied={copiedKey === "all-skills"}
              on:click={() => copyToClipboard(allSkillNames(), "all-skills")}
            >
              {copiedKey === "all-skills" ? "Copied" : "Copy All Skills (comma-separated)"}
            </button>
          </div>
          {#each skillCategories as cat (cat.category.id)}
            {#if cat.skills.length > 0}
              <div class="skill-category-block">
                <span class="skill-category-name">{cat.category.name}</span>
                <div class="skill-chips">
                  {#each cat.skills as skill (skill.id)}
                    <button
                      class="skill-chip"
                      class:legacy={skill.is_legacy}
                      class:copied={copiedKey === `skill-${skill.id}`}
                      on:click={() => copyToClipboard(skill.name, `skill-${skill.id}`)}
                      title="{skill.name} — click to copy"
                    >
                      {skill.name}
                      {#if copiedKey === `skill-${skill.id}`}
                        <span class="chip-copied">&#10003;</span>
                      {/if}
                    </button>
                  {/each}
                </div>
              </div>
            {/if}
          {/each}
        {/if}
      </section>
    {/if}

    <!-- Professional Summaries -->
    {#if summaries.length > 0}
      <section class="helper-section">
        <button class="section-toggle" on:click={() => toggleSection("summaries")}>
          <span class="toggle-arrow" class:open={sectionsOpen.summaries}>&#9654;</span>
          <h3>Professional Summaries</h3>
        </button>
        {#if sectionsOpen.summaries}
          {#each summaries as s (s.id)}
            <button
              class="copy-cell copy-cell-block"
              class:copied={copiedKey === `summary-${s.id}`}
              on:click={() => copyToClipboard(s.body_text, `summary-${s.id}`)}
              title="Copy summary text"
            >
              <span class="cell-label">{s.label}</span>
              <span class="cell-value text-value">{s.body_text}</span>
              <span class="cell-status">{copiedKey === `summary-${s.id}` ? "Copied" : ""}</span>
            </button>
          {/each}
        {/if}
      </section>
    {/if}

    <!-- Role Descriptors -->
    {#if descriptors.length > 0}
      <section class="helper-section">
        <button class="section-toggle" on:click={() => toggleSection("descriptors")}>
          <span class="toggle-arrow" class:open={sectionsOpen.descriptors}>&#9654;</span>
          <h3>Role Descriptors</h3>
        </button>
        {#if sectionsOpen.descriptors}
          <div class="skills-actions">
            <button
              class="copy-all-btn"
              class:copied={copiedKey === "all-descriptors"}
              on:click={() => copyToClipboard(allDescriptorTitles(), "all-descriptors")}
            >
              {copiedKey === "all-descriptors" ? "Copied" : "Copy All (comma-separated)"}
            </button>
          </div>
          <div class="skill-chips">
            {#each descriptors as d (d.id)}
              <button
                class="skill-chip"
                class:copied={copiedKey === `desc-${d.id}`}
                on:click={() => copyToClipboard(d.title, `desc-${d.id}`)}
                title="{d.title} — click to copy"
              >
                {d.title}
                {#if copiedKey === `desc-${d.id}`}
                  <span class="chip-copied">&#10003;</span>
                {/if}
              </button>
            {/each}
          </div>
        {/if}
      </section>
    {/if}

    <!-- Core Expertise -->
    {#if coreExpertise.length > 0}
      <section class="helper-section">
        <button class="section-toggle" on:click={() => toggleSection("expertise")}>
          <span class="toggle-arrow" class:open={sectionsOpen.expertise}>&#9654;</span>
          <h3>Core Expertise</h3>
        </button>
        {#if sectionsOpen.expertise}
          <div class="skills-actions">
            <button
              class="copy-all-btn"
              class:copied={copiedKey === "all-expertise"}
              on:click={() => copyToClipboard(allCoreExpertiseLabels(), "all-expertise")}
            >
              {copiedKey === "all-expertise" ? "Copied" : "Copy All (pipe-separated)"}
            </button>
          </div>
          <div class="skill-chips">
            {#each coreExpertise as e (e.id)}
              <button
                class="skill-chip"
                class:copied={copiedKey === `exp-${e.id}`}
                on:click={() => copyToClipboard(e.label, `exp-${e.id}`)}
                title="{e.label} — click to copy"
              >
                {e.label}
                {#if copiedKey === `exp-${e.id}`}
                  <span class="chip-copied">&#10003;</span>
                {/if}
              </button>
            {/each}
          </div>
        {/if}
      </section>
    {/if}
  {/if}
</div>

<style>
  /* ------------------------------------------------------------ */
  /* Page layout                                                   */
  /* ------------------------------------------------------------ */

  .helper-page {
    max-width: 960px;
  }

  .page-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 24px;
  }

  .page-header h2 {
    margin: 0;
    font-size: 1.5rem;
    color: #e0e0e0;
  }

  .page-description {
    color: #7a8a9a;
    font-size: 0.9rem;
    margin: 4px 0 0;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .date-format-label {
    font-size: 0.78rem;
    color: #7a8a9a;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    white-space: nowrap;
  }

  .date-format-select {
    background-color: #1a2332;
    color: #e0e0e0;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 5px 8px;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .date-format-select:focus {
    outline: none;
    border-color: #4a8af4;
    box-shadow: 0 0 0 2px rgba(74, 138, 244, 0.15);
  }

  /* ------------------------------------------------------------ */
  /* Section toggles                                               */
  /* ------------------------------------------------------------ */

  .helper-section {
    margin-bottom: 8px;
  }

  .section-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    background: none;
    border: none;
    cursor: pointer;
    padding: 8px 0;
    width: 100%;
    text-align: left;
  }

  .section-toggle h3 {
    margin: 0;
    font-size: 1.05rem;
    color: #c0d0e0;
    font-weight: 600;
  }

  .section-toggle:hover h3 {
    color: #e0e0e0;
  }

  .toggle-arrow {
    font-size: 0.7rem;
    color: #5a6a7a;
    transition: transform 0.15s;
    display: inline-block;
  }

  .toggle-arrow.open {
    transform: rotate(90deg);
  }

  /* ------------------------------------------------------------ */
  /* Copy grid — the row/cell layout                               */
  /* ------------------------------------------------------------ */

  .copy-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 6px;
    padding: 4px 0 12px 20px;
  }

  .copy-cell {
    display: flex;
    flex-direction: column;
    gap: 2px;
    background: #1a2332;
    border: 1px solid #2a3a4a;
    border-radius: 5px;
    padding: 8px 10px;
    cursor: pointer;
    text-align: left;
    transition: border-color 0.15s, background-color 0.15s;
    position: relative;
    min-width: 0;
  }

  .copy-cell:hover {
    border-color: #4a8af4;
    background: #1e2d3d;
  }

  .copy-cell.copied {
    border-color: #40a060;
    background: #1a2e22;
  }

  .copy-cell-wide {
    grid-column: 1 / -1;
  }

  .copy-cell-block {
    display: flex;
    flex-direction: column;
    gap: 2px;
    background: #1a2332;
    border: 1px solid #2a3a4a;
    border-radius: 5px;
    padding: 8px 10px;
    cursor: pointer;
    text-align: left;
    transition: border-color 0.15s, background-color 0.15s;
    position: relative;
    margin: 0 0 6px 20px;
    width: calc(100% - 20px);
  }

  .copy-cell-block:hover {
    border-color: #4a8af4;
    background: #1e2d3d;
  }

  .copy-cell-block.copied {
    border-color: #40a060;
    background: #1a2e22;
  }

  .cell-label {
    font-size: 0.7rem;
    color: #5a7a8a;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-weight: 600;
  }

  .cell-value {
    font-size: 0.88rem;
    color: #d0e0f0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .cell-value.text-value {
    white-space: pre-wrap;
    line-height: 1.4;
    font-size: 0.82rem;
  }

  .cell-value.url-value {
    font-size: 0.8rem;
    color: #7eafff;
  }

  .cell-status {
    position: absolute;
    top: 6px;
    right: 8px;
    font-size: 0.68rem;
    color: #40a060;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  /* ------------------------------------------------------------ */
  /* Work history cards                                            */
  /* ------------------------------------------------------------ */

  .work-card {
    margin: 0 0 12px 20px;
    border: 1px solid #2a3a4a;
    border-radius: 6px;
    overflow: hidden;
  }

  .work-card .copy-grid {
    padding: 6px 10px 8px;
    margin: 0;
  }

  .work-card-header {
    display: flex;
    align-items: baseline;
    gap: 6px;
    padding: 10px 12px 2px;
    color: #e0e0e0;
    font-size: 0.95rem;
    font-weight: 600;
  }

  .work-employer { color: #c0d0e0; }
  .work-title-sep { color: #4a5a6a; font-weight: 400; }
  .work-jobtitle { color: #8ab0d0; font-weight: 400; }

  .entry-card {
    margin: 0 0 8px 20px;
  }

  .entry-card .copy-grid {
    padding: 0;
  }

  /* ------------------------------------------------------------ */
  /* Bullets                                                       */
  /* ------------------------------------------------------------ */

  .bullets-section {
    padding: 4px 12px 10px;
    border-top: 1px solid #222e3e;
  }

  .bullets-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 4px;
  }

  .bullets-label {
    font-size: 0.72rem;
    color: #5a7a8a;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-weight: 600;
  }

  .copy-all-btn {
    background: transparent;
    border: 1px solid #2a3a4a;
    border-radius: 3px;
    color: #7a8a9a;
    font-size: 0.72rem;
    padding: 2px 8px;
    cursor: pointer;
    transition: border-color 0.15s, color 0.15s;
  }

  .copy-all-btn:hover {
    border-color: #4a8af4;
    color: #c0d0e0;
  }

  .copy-all-btn.copied {
    border-color: #40a060;
    color: #40a060;
  }

  .copy-bullet {
    display: flex;
    align-items: flex-start;
    gap: 6px;
    background: transparent;
    border: 1px solid transparent;
    border-radius: 4px;
    padding: 3px 6px;
    cursor: pointer;
    text-align: left;
    width: 100%;
    transition: background-color 0.15s, border-color 0.15s;
    position: relative;
  }

  .copy-bullet:hover {
    background: #1e2d3d;
    border-color: #2a3a4a;
  }

  .copy-bullet.copied {
    background: #1a2e22;
    border-color: #40a060;
  }

  .bullet-dot {
    color: #4a5a6a;
    flex-shrink: 0;
    margin-top: 1px;
  }

  .bullet-text {
    color: #b0c4d8;
    font-size: 0.82rem;
    line-height: 1.4;
  }

  .copy-bullet .cell-status {
    top: 3px;
  }

  /* ------------------------------------------------------------ */
  /* Skill chips                                                   */
  /* ------------------------------------------------------------ */

  .skills-actions {
    padding: 0 0 6px 20px;
  }

  .skill-category-block {
    padding: 0 0 10px 20px;
  }

  .skill-category-name {
    display: block;
    font-size: 0.76rem;
    color: #5a7a8a;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-weight: 600;
    margin-bottom: 4px;
  }

  .skill-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
    padding-left: 20px;
  }

  .skill-category-block .skill-chips {
    padding-left: 0;
  }

  .skill-chip {
    background: #1a2332;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    color: #c0d0e0;
    font-size: 0.82rem;
    padding: 4px 10px;
    cursor: pointer;
    transition: border-color 0.15s, background-color 0.15s;
    white-space: nowrap;
  }

  .skill-chip:hover {
    border-color: #4a8af4;
    background: #1e2d3d;
  }

  .skill-chip.copied {
    border-color: #40a060;
    background: #1a2e22;
    color: #80d0a0;
  }

  .skill-chip.legacy {
    opacity: 0.5;
  }

  .chip-copied {
    margin-left: 4px;
    color: #40a060;
    font-size: 0.75rem;
  }

  /* ------------------------------------------------------------ */
  /* Responsive                                                    */
  /* ------------------------------------------------------------ */

  @media (max-width: 760px) {
    .page-header {
      flex-direction: column;
    }

    .copy-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
