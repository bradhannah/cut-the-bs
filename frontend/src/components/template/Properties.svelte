<script lang="ts">
  // Properties.svelte — T041-T046
  // Right panel in the template builder. When an element is selected,
  // shows context-sensitive property editors based on element_type.
  // When no element is selected, shows template-level margin editors.
  // Changes are debounced at 300ms before persisting via API.

  import { onDestroy } from "svelte";
  import {
    canvasElements,
    selectedElement,
    currentTemplate,
    builderReadOnly,
    markSaving,
    markSaved,
    markSaveError,
  } from "../../stores/templateBuilder";
  import {
    updateTemplateElement,
    updateDocumentTemplate,
    addToast,
    type TemplateElementInput,
    type DocumentTemplateInput,
  } from "../../services/api";
  import { elementLabels, elementIcons } from "./elementTypes";

  // --- Config Parsing ---

  // Parsed config for the currently selected element.
  let config: Record<string, any> = {};

  // When the selected element changes, re-parse config.
  $: if ($selectedElement) {
    try {
      config = JSON.parse($selectedElement.config);
    } catch {
      config = {};
    }
  } else {
    config = {};
  }

  $: elementType = $selectedElement?.element_type || "";
  $: elementLabel = elementLabels[elementType] || elementType;
  $: elementIcon = elementIcons[elementType] || "?";

  // --- Debounced Save (T045) ---

  $: isReadOnly = $builderReadOnly || $currentTemplate?.is_builtin;

  let debounceTimer: ReturnType<typeof setTimeout> | null = null;

  onDestroy(() => {
    if (debounceTimer) clearTimeout(debounceTimer);
  });

  function updateConfigField(field: string, value: any): void {
    if (isReadOnly) return;
    config = { ...config, [field]: value };
    debounceSaveElement();
  }

  function debounceSaveElement(): void {
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      saveElement();
    }, 300);
  }

  async function saveElement(): Promise<void> {
    if (isReadOnly) return;
    const el = $selectedElement;
    if (!el) return;

    const input: TemplateElementInput = {
      parent_id: el.parent_id ?? null,
      element_type: el.element_type,
      config: JSON.stringify(config),
    };

    try {
      markSaving();
      const updated = await updateTemplateElement(el.id, input);
      // Update the element in the canvasElements store.
      canvasElements.update((els) =>
        els.map((e) => (e.id === updated.id ? updated : e))
      );
      markSaved();
    } catch (err: any) {
      markSaveError();
      addToast("error", err?.message || "Failed to save element properties");
    }
  }

  // --- Template Margin Save (T046) ---

  let marginDebounceTimer: ReturnType<typeof setTimeout> | null = null;

  onDestroy(() => {
    if (marginDebounceTimer) clearTimeout(marginDebounceTimer);
  });

  function updateMargin(field: string, value: number): void {
    if (isReadOnly) return;
    if (!$currentTemplate) return;
    // Clamp to 0-288
    const clamped = Math.max(0, Math.min(288, value));
    currentTemplate.update((t) => {
      if (!t) return t;
      return { ...t, [field]: clamped };
    });
    debounceMarginSave();
  }

  function debounceMarginSave(): void {
    if (marginDebounceTimer) clearTimeout(marginDebounceTimer);
    marginDebounceTimer = setTimeout(() => {
      saveMargins();
    }, 300);
  }

  async function saveMargins(): Promise<void> {
    if (isReadOnly) return;
    const tmpl = $currentTemplate;
    if (!tmpl) return;

    const input: DocumentTemplateInput = {
      name: tmpl.name,
      description: tmpl.description,
      template_type: tmpl.template_type,
      margin_top: tmpl.margin_top,
      margin_bottom: tmpl.margin_bottom,
      margin_left: tmpl.margin_left,
      margin_right: tmpl.margin_right,
    };

    try {
      markSaving();
      const updated = await updateDocumentTemplate(tmpl.id, input);
      currentTemplate.set(updated);
      markSaved();
    } catch (err: any) {
      markSaveError();
      addToast("error", err?.message || "Failed to save template margins");
    }
  }

  // Points to inches conversion for display.
  function ptsToInches(pts: number): string {
    return (pts / 72).toFixed(2);
  }

  // --- Data Binding Options ---

  const dataBindingOptions = [
    { value: "", label: "Always show" },
    { value: "summaries", label: "Summaries" },
    { value: "work_history", label: "Work History" },
    { value: "skills", label: "Skills" },
    { value: "core_expertise", label: "Core Expertise" },
    { value: "academics", label: "Academics" },
    { value: "certifications", label: "Certifications" },
  ];

  // --- Font Style Options ---

  const fontStyleOptions = [
    { value: "normal", label: "Normal" },
    { value: "bold", label: "Bold" },
    { value: "italic", label: "Italic" },
    { value: "bold_italic", label: "Bold Italic" },
  ];

  // --- Alignment Options ---

  const alignmentOptions = [
    { value: "left", label: "Left" },
    { value: "center", label: "Center" },
    { value: "right", label: "Right" },
  ];

  const workTitleRowLayoutOptions = [
    { value: "inline_with_dates", label: "Inline with dates" },
    { value: "stack_dates_below", label: "Dates on next line" },
  ];

  const paragraphSegmentTypeOptions = [
    { value: "static", label: "Static text" },
    { value: "profile", label: "Profile token" },
    { value: "application", label: "Application token" },
    { value: "adhoc", label: "Ad-hoc prompt" },
  ];

  const profileTokenOptions = [
    { value: "signer_name", label: "Signer Name" },
    { value: "email", label: "Email" },
    { value: "location", label: "Location" },
    { value: "profile_links", label: "Profile Links" },
  ];

  const applicationTokenOptions = [
    { value: "company_name", label: "Company Name" },
    { value: "position_title", label: "Position Title" },
    { value: "hiring_manager", label: "Hiring Manager" },
    { value: "recipient_address", label: "Recipient Address" },
  ];

  function paragraphSegments(): Record<string, any>[] {
    if (!Array.isArray(config.segments)) {
      return [];
    }
    return config.segments as Record<string, any>[];
  }

  function updateParagraphSegment(
    index: number,
    field: string,
    value: any
  ): void {
    const segments = paragraphSegments().map((segment) => ({ ...segment }));
    if (!segments[index]) return;
    segments[index][field] = value;
    updateConfigField("segments", segments);
  }

  function addParagraphSegment(): void {
    const segments = paragraphSegments().map((segment) => ({ ...segment }));
    segments.push({ type: "static", text: "" });
    updateConfigField("segments", segments);
  }

  function removeParagraphSegment(index: number): void {
    const segments = paragraphSegments().filter((_, i) => i !== index);
    updateConfigField("segments", segments);
  }

  function onParagraphSegmentTypeChange(index: number, nextType: string): void {
    const segments = paragraphSegments().map((segment) => ({ ...segment }));
    if (!segments[index]) return;

    if (nextType === "static") {
      segments[index] = { type: "static", text: "" };
    } else if (nextType === "profile") {
      segments[index] = { type: "profile", token: "signer_name" };
    } else if (nextType === "application") {
      segments[index] = { type: "application", token: "company_name" };
    } else {
      segments[index] = {
        type: "adhoc",
        key: "",
        label: "",
        help_text: "",
        required: false,
        multiline: true,
      };
    }

    updateConfigField("segments", segments);
  }
</script>

<div class="properties-root" class:read-only={isReadOnly}>
{#if $selectedElement}
  <!-- Element Properties -->
  <div class="properties-header">
    <div class="header-left">
      <span class="header-icon">{elementIcon}</span>
      <h3>{elementLabel}</h3>
    </div>
    <span class="element-type-badge">{elementType}</span>
  </div>

  <div class="properties-body">
    <!-- ========== FORMATTING ELEMENTS (T042) ========== -->

    {#if elementType === "section_heading"}
      <div class="prop-group">
        <label class="prop-label" for="sh-text">Text</label>
        <input
          id="sh-text"
          type="text"
          class="prop-input"
          value={config.text || ""}
          on:input={(e) => updateConfigField("text", e.currentTarget.value)}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="sh-font-size">Font Size (pt)</label>
        <input
          id="sh-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="72"
          step="0.5"
          value={config.font_size ?? 12}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={(config.font_style ?? "bold") === "bold"}
            on:change={(e) =>
              updateConfigField(
                "font_style",
                e.currentTarget.checked ? "bold" : "normal"
              )}
          />
          Bold
        </label>
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.uppercase ?? true}
            on:change={(e) =>
              updateConfigField("uppercase", e.currentTarget.checked)}
          />
          Uppercase
        </label>
      </div>
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.underline ?? true}
            on:change={(e) =>
              updateConfigField("underline", e.currentTarget.checked)}
          />
          Underline
        </label>
      </div>
      {#if config.underline}
        <div class="prop-group">
          <label class="prop-label" for="sh-ul-weight"
            >Underline Weight (pt)</label
          >
          <input
            id="sh-ul-weight"
            type="number"
            class="prop-input prop-input-sm"
            min="0.1"
            max="5"
            step="0.1"
            value={config.underline_weight ?? 0.5}
            on:input={(e) =>
              updateConfigField(
                "underline_weight",
                parseFloat(e.currentTarget.value)
              )}
          />
        </div>
      {/if}
      <div class="prop-group">
        <label class="prop-label" for="sh-space-before">Space Before (pt)</label
        >
        <input
          id="sh-space-before"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_before ?? 8}
          on:input={(e) =>
            updateConfigField(
              "space_before",
              parseFloat(e.currentTarget.value)
            )}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="sh-space-after">Space After (pt)</label>
        <input
          id="sh-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 4}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="sh-data-binding">Data Binding</label>
        <select
          id="sh-data-binding"
          class="prop-select"
          value={config.data_binding || ""}
          on:change={(e) =>
            updateConfigField("data_binding", e.currentTarget.value)}
        >
          {#each dataBindingOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
        <span class="prop-hint"
          >When bound data is empty, heading is hidden</span
        >
      </div>
    {:else if elementType === "horizontal_rule"}
      <div class="prop-group">
        <label class="prop-label" for="hr-weight">Weight (pt)</label>
        <input
          id="hr-weight"
          type="number"
          class="prop-input prop-input-sm"
          min="0.1"
          max="10"
          step="0.1"
          value={config.weight ?? 0.5}
          on:input={(e) =>
            updateConfigField("weight", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="hr-space-before">Space Before (pt)</label
        >
        <input
          id="hr-space-before"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_before ?? 4}
          on:input={(e) =>
            updateConfigField(
              "space_before",
              parseFloat(e.currentTarget.value)
            )}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="hr-space-after">Space After (pt)</label>
        <input
          id="hr-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 4}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "spacer"}
      <div class="prop-group">
        <label class="prop-label" for="sp-height">Height (pt)</label>
        <input
          id="sp-height"
          type="number"
          class="prop-input prop-input-sm"
          min="1"
          max="200"
          step="1"
          value={config.height ?? 10}
          on:input={(e) =>
            updateConfigField("height", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "static_text"}
      <div class="prop-group">
        <label class="prop-label" for="st-text">Text</label>
        <textarea
          id="st-text"
          class="prop-textarea"
          rows="3"
          value={config.text || ""}
          on:input={(e) => updateConfigField("text", e.currentTarget.value)}
        ></textarea>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="st-font-size">Font Size (pt)</label>
        <input
          id="st-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="72"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="st-font-style">Font Style</label>
        <select
          id="st-font-style"
          class="prop-select"
          value={config.font_style || "normal"}
          on:change={(e) =>
            updateConfigField("font_style", e.currentTarget.value)}
        >
          {#each fontStyleOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="st-alignment">Alignment</label>
        <select
          id="st-alignment"
          class="prop-select"
          value={config.alignment || "left"}
          on:change={(e) =>
            updateConfigField("alignment", e.currentTarget.value)}
        >
          {#each alignmentOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="st-space-after">Space After (pt)</label>
        <input
          id="st-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 4}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>

      <!-- ========== DATA-BOUND ELEMENTS (T043) ========== -->
    {:else if elementType === "profile_header"}
      <div class="prop-group">
        <label class="prop-label" for="ph-name-fs">Name Font Size (pt)</label>
        <input
          id="ph-name-fs"
          type="number"
          class="prop-input prop-input-sm"
          min="8"
          max="48"
          step="0.5"
          value={config.name_font_size ?? 18}
          on:input={(e) =>
            updateConfigField(
              "name_font_size",
              parseFloat(e.currentTarget.value)
            )}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ph-detail-fs"
          >Detail Font Size (pt)</label
        >
        <input
          id="ph-detail-fs"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.detail_font_size ?? 10}
          on:input={(e) =>
            updateConfigField(
              "detail_font_size",
              parseFloat(e.currentTarget.value)
            )}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ph-alignment">Alignment</label>
        <select
          id="ph-alignment"
          class="prop-select"
          value={config.alignment || "center"}
          on:change={(e) =>
            updateConfigField("alignment", e.currentTarget.value)}
        >
          {#each alignmentOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ph-link-sep">Link Separator</label>
        <input
          id="ph-link-sep"
          type="text"
          class="prop-input"
          value={config.link_separator || " | "}
          on:input={(e) =>
            updateConfigField("link_separator", e.currentTarget.value)}
        />
      </div>
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.show_links ?? true}
            on:change={(e) =>
              updateConfigField("show_links", e.currentTarget.checked)}
          />
          Show Links
        </label>
      </div>
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.show_links_inline ?? false}
            disabled={!(config.show_links ?? true)}
            on:change={(e) =>
              updateConfigField("show_links_inline", e.currentTarget.checked)}
          />
          Links Inline (wrap)
        </label>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ph-space-after">Space After (pt)</label>
        <input
          id="ph-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 6}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "role_descriptors"}
      <div class="prop-group">
        <label class="prop-label" for="rd-separator">Separator</label>
        <input
          id="rd-separator"
          type="text"
          class="prop-input"
          value={config.separator || " | "}
          on:input={(e) =>
            updateConfigField("separator", e.currentTarget.value)}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="rd-font-size">Font Size (pt)</label>
        <input
          id="rd-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="rd-font-style">Font Style</label>
        <select
          id="rd-font-style"
          class="prop-select"
          value={config.font_style || "italic"}
          on:change={(e) =>
            updateConfigField("font_style", e.currentTarget.value)}
        >
          {#each fontStyleOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="rd-alignment">Alignment</label>
        <select
          id="rd-alignment"
          class="prop-select"
          value={config.alignment || "center"}
          on:change={(e) =>
            updateConfigField("alignment", e.currentTarget.value)}
        >
          {#each alignmentOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="rd-space-after">Space After (pt)</label>
        <input
          id="rd-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 4}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "professional_summary"}
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.show_master ?? true}
            on:change={(e) =>
              updateConfigField("show_master", e.currentTarget.checked)}
          />
          Show Master Summary
        </label>
      </div>
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.show_bullet_summaries ?? true}
            on:change={(e) =>
              updateConfigField(
                "show_bullet_summaries",
                e.currentTarget.checked
              )}
          />
          Show Bullet Summaries
        </label>
      </div>
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.enable_bullets ?? true}
            disabled={!(config.show_bullet_summaries ?? true)}
            on:change={(e) =>
              updateConfigField("enable_bullets", e.currentTarget.checked)}
          />
          Render Bullet Markers
        </label>
      </div>

      {#if (config.show_bullet_summaries ?? true) && (config.enable_bullets ?? true)}
        <div class="prop-group">
          <label class="prop-label" for="ps-bullet">Bullet Character</label>
          <input
            id="ps-bullet"
            type="text"
            class="prop-input prop-input-sm"
            maxlength="2"
            value={config.bullet_char || "\u2022"}
            on:input={(e) =>
              updateConfigField("bullet_char", e.currentTarget.value)}
          />
        </div>
      {/if}

      <div class="prop-group">
        <label class="prop-label" for="ps-font-size">Font Size (pt)</label>
        <input
          id="ps-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ps-space-before">Space Before (pt)</label
        >
        <input
          id="ps-space-before"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_before ?? 0}
          on:input={(e) =>
            updateConfigField(
              "space_before",
              parseFloat(e.currentTarget.value)
            )}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ps-space-after">Space After (pt)</label>
        <input
          id="ps-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 2}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "skills"}
      <div class="prop-group">
        <label class="prop-label" for="sk-font-size">Font Size (pt)</label>
        <input
          id="sk-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.group_by_category ?? true}
            on:change={(e) =>
              updateConfigField("group_by_category", e.currentTarget.checked)}
          />
          Group by Category
        </label>
      </div>
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.include_legacy ?? true}
            on:change={(e) =>
              updateConfigField("include_legacy", e.currentTarget.checked)}
          />
          Include Legacy Skills
        </label>
      </div>
      {#if config.include_legacy}
        <div class="prop-group">
          <label class="prop-label" for="sk-legacy-suffix">Legacy Suffix</label>
          <input
            id="sk-legacy-suffix"
            type="text"
            class="prop-input"
            value={config.legacy_suffix || " (Legacy)"}
            on:input={(e) =>
              updateConfigField("legacy_suffix", e.currentTarget.value)}
          />
        </div>
      {/if}
      <div class="prop-group">
        <label class="prop-label" for="sk-separator">Skill Separator</label>
        <input
          id="sk-separator"
          type="text"
          class="prop-input"
          value={config.skill_separator || ", "}
          on:input={(e) =>
            updateConfigField("skill_separator", e.currentTarget.value)}
        />
      </div>
      {#if config.group_by_category}
        <div class="prop-group">
          <label class="prop-label" for="sk-cat-font-style"
            >Category Font Style</label
          >
          <select
            id="sk-cat-font-style"
            class="prop-select"
            value={config.category_font_style || "bold"}
            on:change={(e) =>
              updateConfigField("category_font_style", e.currentTarget.value)}
          >
            {#each fontStyleOptions as opt}
              <option value={opt.value}>{opt.label}</option>
            {/each}
          </select>
        </div>
      {/if}
    {:else if elementType === "core_expertise"}
      <div class="prop-group">
        <label class="prop-label" for="ce-font-size">Font Size (pt)</label>
        <input
          id="ce-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ce-separator">Separator</label>
        <input
          id="ce-separator"
          type="text"
          class="prop-input"
          value={config.separator || " \u00B7 "}
          on:input={(e) =>
            updateConfigField("separator", e.currentTarget.value)}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ce-space-after">Space After (pt)</label>
        <input
          id="ce-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 4}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>

      <!-- ========== LOOP CONTAINERS & SUB-ELEMENTS (T044) ========== -->
    {:else if elementType === "work_history_loop"}
      <div class="prop-group">
        <label class="prop-label" for="whl-entry-gap">Entry Gap (pt)</label>
        <input
          id="whl-entry-gap"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.entry_gap ?? 8}
          on:input={(e) =>
            updateConfigField("entry_gap", parseFloat(e.currentTarget.value))}
        />
        <span class="prop-hint">Spacing between work history entries</span>
      </div>
    {:else if elementType === "education_loop"}
      <div class="prop-group">
        <label class="prop-label" for="el-entry-gap">Entry Gap (pt)</label>
        <input
          id="el-entry-gap"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.entry_gap ?? 4}
          on:input={(e) =>
            updateConfigField("entry_gap", parseFloat(e.currentTarget.value))}
        />
        <span class="prop-hint">Spacing between education entries</span>
      </div>
    {:else if elementType === "certifications_loop"}
      <div class="prop-group">
        <label class="prop-label" for="cl-entry-gap">Entry Gap (pt)</label>
        <input
          id="cl-entry-gap"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.entry_gap ?? 2}
          on:input={(e) =>
            updateConfigField("entry_gap", parseFloat(e.currentTarget.value))}
        />
        <span class="prop-hint">Spacing between certification entries</span>
      </div>
    {:else if elementType === "work_title"}
      <div class="prop-group">
        <label class="prop-label" for="wt-font-size">Font Size (pt)</label>
        <input
          id="wt-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 11}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wt-font-style">Font Style</label>
        <select
          id="wt-font-style"
          class="prop-select"
          value={config.font_style || "bold"}
          on:change={(e) =>
            updateConfigField("font_style", e.currentTarget.value)}
        >
          {#each fontStyleOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wt-row-layout">Title Row Layout</label>
        <select
          id="wt-row-layout"
          class="prop-select"
          value={config.title_row_layout || "inline_with_dates"}
          on:change={(e) =>
            updateConfigField("title_row_layout", e.currentTarget.value)}
        >
          {#each workTitleRowLayoutOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-row">
        <label class="prop-toggle">
          <input
            type="checkbox"
            checked={config.include_employer ?? false}
            on:change={(e) =>
              updateConfigField("include_employer", e.currentTarget.checked)}
          />
          Include Employer
        </label>
      </div>
      {#if config.include_employer}
        <div class="prop-group">
          <label class="prop-label" for="wt-emp-sep">Employer Separator</label>
          <input
            id="wt-emp-sep"
            type="text"
            class="prop-input"
            value={config.employer_separator || " - "}
            on:input={(e) =>
              updateConfigField("employer_separator", e.currentTarget.value)}
          />
        </div>
        <div class="prop-group">
          <label class="prop-label" for="wt-emp-style"
            >Employer Font Style</label
          >
          <select
            id="wt-emp-style"
            class="prop-select"
            value={config.employer_font_style || "normal"}
            on:change={(e) =>
              updateConfigField("employer_font_style", e.currentTarget.value)}
          >
            {#each fontStyleOptions as opt}
              <option value={opt.value}>{opt.label}</option>
            {/each}
          </select>
        </div>
      {/if}
      <div class="prop-group">
        <label class="prop-label" for="wt-space-after">Space After (pt)</label>
        <input
          id="wt-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 0}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "work_employer"}
      <div class="prop-group">
        <label class="prop-label" for="we-font-size">Font Size (pt)</label>
        <input
          id="we-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="we-font-style">Font Style</label>
        <select
          id="we-font-style"
          class="prop-select"
          value={config.font_style || "normal"}
          on:change={(e) =>
            updateConfigField("font_style", e.currentTarget.value)}
        >
          {#each fontStyleOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="we-space-after">Space After (pt)</label>
        <input
          id="we-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 0}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "work_dates"}
      <div class="prop-group">
        <label class="prop-label" for="wd-font-size">Font Size (pt)</label>
        <input
          id="wd-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wd-alignment">Alignment</label>
        <select
          id="wd-alignment"
          class="prop-select"
          value={config.alignment || "right"}
          on:change={(e) =>
            updateConfigField("alignment", e.currentTarget.value)}
        >
          {#each alignmentOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wd-space-after">Space After (pt)</label>
        <input
          id="wd-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 2}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "work_summary"}
      <div class="prop-group">
        <label class="prop-label" for="ws-font-size">Font Size (pt)</label>
        <input
          id="ws-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ws-font-style">Font Style</label>
        <select
          id="ws-font-style"
          class="prop-select"
          value={config.font_style || "italic"}
          on:change={(e) =>
            updateConfigField("font_style", e.currentTarget.value)}
        >
          {#each fontStyleOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ws-space-after">Space After (pt)</label>
        <input
          id="ws-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 2}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "work_bullets"}
      <div class="prop-group">
        <label class="prop-label" for="wb-bullet">Bullet Character</label>
        <input
          id="wb-bullet"
          type="text"
          class="prop-input prop-input-sm"
          maxlength="2"
          value={config.bullet_char || "\u2022"}
          on:input={(e) =>
            updateConfigField("bullet_char", e.currentTarget.value)}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wb-font-size">Font Size (pt)</label>
        <input
          id="wb-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wb-indent">Indent (pt)</label>
        <input
          id="wb-indent"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.indent ?? 15}
          on:input={(e) =>
            updateConfigField("indent", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wb-space-after">Space After (pt)</label>
        <input
          id="wb-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 2}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "work_outcomes"}
      <div class="prop-group">
        <label class="prop-label" for="wo-bullet">Bullet Character</label>
        <input
          id="wo-bullet"
          type="text"
          class="prop-input prop-input-sm"
          maxlength="2"
          value={config.bullet_char || "\u2022"}
          on:input={(e) =>
            updateConfigField("bullet_char", e.currentTarget.value)}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wo-font-size">Font Size (pt)</label>
        <input
          id="wo-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wo-indent">Indent (pt)</label>
        <input
          id="wo-indent"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.indent ?? 15}
          on:input={(e) =>
            updateConfigField("indent", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wo-outcomes-label">Outcomes Label</label>
        <input
          id="wo-outcomes-label"
          type="text"
          class="prop-input"
          value={config.outcomes_label || "Key Outcomes:"}
          on:input={(e) =>
            updateConfigField("outcomes_label", e.currentTarget.value)}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="wo-space-after">Space After (pt)</label>
        <input
          id="wo-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 2}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "edu_credential"}
      <div class="prop-group">
        <label class="prop-label" for="ec-font-size">Font Size (pt)</label>
        <input
          id="ec-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ec-font-style">Font Style</label>
        <select
          id="ec-font-style"
          class="prop-select"
          value={config.font_style || "bold"}
          on:change={(e) =>
            updateConfigField("font_style", e.currentTarget.value)}
        >
          {#each fontStyleOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ec-space-after">Space After (pt)</label>
        <input
          id="ec-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 0}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "edu_institution"}
      <div class="prop-group">
        <label class="prop-label" for="ei-font-size">Font Size (pt)</label>
        <input
          id="ei-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ei-font-style">Font Style</label>
        <select
          id="ei-font-style"
          class="prop-select"
          value={config.font_style || "normal"}
          on:change={(e) =>
            updateConfigField("font_style", e.currentTarget.value)}
        >
          {#each fontStyleOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ei-space-after">Space After (pt)</label>
        <input
          id="ei-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 0}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "edu_date"}
      <div class="prop-group">
        <label class="prop-label" for="ed-font-size">Font Size (pt)</label>
        <input
          id="ed-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ed-alignment">Alignment</label>
        <select
          id="ed-alignment"
          class="prop-select"
          value={config.alignment || "right"}
          on:change={(e) =>
            updateConfigField("alignment", e.currentTarget.value)}
        >
          {#each alignmentOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ed-space-after">Space After (pt)</label>
        <input
          id="ed-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 0}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "cert_name"}
      <div class="prop-group">
        <label class="prop-label" for="cn-font-size">Font Size (pt)</label>
        <input
          id="cn-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="cn-font-style">Font Style</label>
        <select
          id="cn-font-style"
          class="prop-select"
          value={config.font_style || "bold"}
          on:change={(e) =>
            updateConfigField("font_style", e.currentTarget.value)}
        >
          {#each fontStyleOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="cn-space-after">Space After (pt)</label>
        <input
          id="cn-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 0}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "cert_detail"}
      <div class="prop-group">
        <label class="prop-label" for="cd-font-size">Font Size (pt)</label>
        <input
          id="cd-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 10}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="cd-font-style">Font Style</label>
        <select
          id="cd-font-style"
          class="prop-select"
          value={config.font_style || "normal"}
          on:change={(e) =>
            updateConfigField("font_style", e.currentTarget.value)}
        >
          {#each fontStyleOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="cd-space-after">Space After (pt)</label>
        <input
          id="cd-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 0}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>

      <!-- ========== COVER LETTER ELEMENTS ========== -->
    {:else if elementType === "body_text"}
      <div class="prop-group">
        <label class="prop-label" for="bt-font-size">Font Size (pt)</label>
        <input
          id="bt-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 11}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="bt-line-spacing">Line Spacing</label>
        <input
          id="bt-line-spacing"
          type="number"
          class="prop-input prop-input-sm"
          min="0.5"
          max="3"
          step="0.05"
          value={config.line_spacing ?? 1.15}
          on:input={(e) =>
            updateConfigField(
              "line_spacing",
              parseFloat(e.currentTarget.value)
            )}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="bt-space-after">Space After (pt)</label>
        <input
          id="bt-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 12}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "paragraph"}
      <div class="prop-group">
        <label class="prop-label" for="pg-font-size">Font Size (pt)</label>
        <input
          id="pg-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 11}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="pg-line-spacing">Line Spacing</label>
        <input
          id="pg-line-spacing"
          type="number"
          class="prop-input prop-input-sm"
          min="0.5"
          max="3"
          step="0.05"
          value={config.line_spacing ?? 1.15}
          on:input={(e) =>
            updateConfigField(
              "line_spacing",
              parseFloat(e.currentTarget.value)
            )}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="pg-space-after">Space After (pt)</label>
        <input
          id="pg-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 12}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>

      <div class="prop-group">
        <div class="prop-label-row">
          <span class="prop-label">Segments</span>
          <button
            type="button"
            class="prop-button"
            on:click={addParagraphSegment}>Add Segment</button
          >
        </div>

        {#if paragraphSegments().length === 0}
          <div class="prop-hint">Add segments to build this paragraph.</div>
        {/if}

        {#each paragraphSegments() as segment, index (index)}
          <div class="segment-card">
            <div class="segment-card-head">
              <span class="segment-index">#{index + 1}</span>
              <button
                type="button"
                class="segment-remove"
                on:click={() => removeParagraphSegment(index)}
              >
                Remove
              </button>
            </div>

            <div class="prop-group">
              <div class="prop-label">Type</div>
              <select
                class="prop-select"
                value={segment.type || "static"}
                on:change={(e) =>
                  onParagraphSegmentTypeChange(index, e.currentTarget.value)}
              >
                {#each paragraphSegmentTypeOptions as opt}
                  <option value={opt.value}>{opt.label}</option>
                {/each}
              </select>
            </div>

            {#if segment.type === "static"}
              <div class="prop-group">
                <div class="prop-label">Text</div>
                <textarea
                  class="prop-textarea"
                  rows="2"
                  value={segment.text || ""}
                  on:input={(e) =>
                    updateParagraphSegment(
                      index,
                      "text",
                      e.currentTarget.value
                    )}
                ></textarea>
              </div>
            {:else if segment.type === "profile"}
              <div class="prop-group">
                <div class="prop-label">Profile Token</div>
                <select
                  class="prop-select"
                  value={segment.token || "signer_name"}
                  on:change={(e) =>
                    updateParagraphSegment(
                      index,
                      "token",
                      e.currentTarget.value
                    )}
                >
                  {#each profileTokenOptions as opt}
                    <option value={opt.value}>{opt.label}</option>
                  {/each}
                </select>
              </div>
            {:else if segment.type === "application"}
              <div class="prop-group">
                <div class="prop-label">Application Token</div>
                <select
                  class="prop-select"
                  value={segment.token || "company_name"}
                  on:change={(e) =>
                    updateParagraphSegment(
                      index,
                      "token",
                      e.currentTarget.value
                    )}
                >
                  {#each applicationTokenOptions as opt}
                    <option value={opt.value}>{opt.label}</option>
                  {/each}
                </select>
              </div>
            {:else if segment.type === "adhoc"}
              <div class="prop-group">
                <div class="prop-label">Key</div>
                <input
                  class="prop-input"
                  type="text"
                  value={segment.key || ""}
                  on:input={(e) =>
                    updateParagraphSegment(index, "key", e.currentTarget.value)}
                />
                <span class="prop-hint"
                  >Stable id used to store answers (e.g. why_company).</span
                >
              </div>
              <div class="prop-group">
                <div class="prop-label">Prompt Label</div>
                <input
                  class="prop-input"
                  type="text"
                  value={segment.label || ""}
                  on:input={(e) =>
                    updateParagraphSegment(
                      index,
                      "label",
                      e.currentTarget.value
                    )}
                />
              </div>
              <div class="prop-group">
                <div class="prop-label">Help Text</div>
                <textarea
                  class="prop-textarea"
                  rows="2"
                  value={segment.help_text || ""}
                  on:input={(e) =>
                    updateParagraphSegment(
                      index,
                      "help_text",
                      e.currentTarget.value
                    )}
                ></textarea>
              </div>
              <div class="prop-row">
                <label class="prop-toggle">
                  <input
                    type="checkbox"
                    checked={segment.required ?? false}
                    on:change={(e) =>
                      updateParagraphSegment(
                        index,
                        "required",
                        e.currentTarget.checked
                      )}
                  />
                  Required
                </label>
                <label class="prop-toggle">
                  <input
                    type="checkbox"
                    checked={segment.multiline ?? true}
                    on:change={(e) =>
                      updateParagraphSegment(
                        index,
                        "multiline",
                        e.currentTarget.checked
                      )}
                  />
                  Multiline
                </label>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {:else if elementType === "date"}
      <div class="prop-group">
        <label class="prop-label" for="dt-font-size">Font Size (pt)</label>
        <input
          id="dt-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 11}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="dt-format">Date Format</label>
        <input
          id="dt-format"
          type="text"
          class="prop-input"
          value={config.format || "January 2, 2006"}
          on:input={(e) => updateConfigField("format", e.currentTarget.value)}
        />
        <span class="prop-hint">Go date format (e.g. January 2, 2006)</span>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="dt-alignment">Alignment</label>
        <select
          id="dt-alignment"
          class="prop-select"
          value={config.alignment || "left"}
          on:change={(e) =>
            updateConfigField("alignment", e.currentTarget.value)}
        >
          {#each alignmentOptions as opt}
            <option value={opt.value}>{opt.label}</option>
          {/each}
        </select>
      </div>
      <div class="prop-group">
        <label class="prop-label" for="dt-space-after">Space After (pt)</label>
        <input
          id="dt-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 12}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "greeting"}
      <div class="prop-group">
        <label class="prop-label" for="gr-text">Text</label>
        <input
          id="gr-text"
          type="text"
          class="prop-input"
          value={config.text || "Dear Hiring Manager,"}
          on:input={(e) => updateConfigField("text", e.currentTarget.value)}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="gr-font-size">Font Size (pt)</label>
        <input
          id="gr-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 11}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="gr-space-after">Space After (pt)</label>
        <input
          id="gr-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 12}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "closing"}
      <div class="prop-group">
        <label class="prop-label" for="cl2-text">Text</label>
        <input
          id="cl2-text"
          type="text"
          class="prop-input"
          value={config.text || "Sincerely,"}
          on:input={(e) => updateConfigField("text", e.currentTarget.value)}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="cl2-font-size">Font Size (pt)</label>
        <input
          id="cl2-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 11}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="cl2-space-after">Space After (pt)</label>
        <input
          id="cl2-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 24}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else if elementType === "recipient_address"}
      <div class="prop-group">
        <label class="prop-label" for="ra-font-size">Font Size (pt)</label>
        <input
          id="ra-font-size"
          type="number"
          class="prop-input prop-input-sm"
          min="6"
          max="36"
          step="0.5"
          value={config.font_size ?? 11}
          on:input={(e) =>
            updateConfigField("font_size", parseFloat(e.currentTarget.value))}
        />
      </div>
      <div class="prop-group">
        <label class="prop-label" for="ra-space-after">Space After (pt)</label>
        <input
          id="ra-space-after"
          type="number"
          class="prop-input prop-input-sm"
          min="0"
          max="100"
          step="1"
          value={config.space_after ?? 12}
          on:input={(e) =>
            updateConfigField("space_after", parseFloat(e.currentTarget.value))}
        />
      </div>
    {:else}
      <!-- Fallback: show raw config JSON -->
      <div class="prop-group">
        <label class="prop-label">Raw Config</label>
        <pre class="config-preview">{JSON.stringify(config, null, 2)}</pre>
      </div>
    {/if}
  </div>
{:else if $currentTemplate}
  <!-- Template-Level Properties (T046) -->
  <div class="properties-header">
    <div class="header-left">
      <h3>Template</h3>
    </div>
    <span class="element-type-badge">margins</span>
  </div>

  <div class="properties-body">
    <p class="section-title">Page Margins</p>

    {#if $currentTemplate.is_builtin}
      <p class="builtin-notice">
        Built-in templates cannot be edited. Duplicate this template to
        customize margins.
      </p>
    {/if}

    <div class="prop-group">
      <label class="prop-label" for="margin-top">
        Top
        <span class="margin-inches"
          >{ptsToInches($currentTemplate.margin_top)}in</span
        >
      </label>
      <input
        id="margin-top"
        type="number"
        class="prop-input prop-input-sm"
        min="0"
        max="288"
        step="1"
        value={$currentTemplate.margin_top}
        disabled={$currentTemplate.is_builtin}
        on:input={(e) =>
          updateMargin("margin_top", parseFloat(e.currentTarget.value))}
      />
    </div>

    <div class="prop-group">
      <label class="prop-label" for="margin-bottom">
        Bottom
        <span class="margin-inches"
          >{ptsToInches($currentTemplate.margin_bottom)}in</span
        >
      </label>
      <input
        id="margin-bottom"
        type="number"
        class="prop-input prop-input-sm"
        min="0"
        max="288"
        step="1"
        value={$currentTemplate.margin_bottom}
        disabled={$currentTemplate.is_builtin}
        on:input={(e) =>
          updateMargin("margin_bottom", parseFloat(e.currentTarget.value))}
      />
    </div>

    <div class="prop-group">
      <label class="prop-label" for="margin-left">
        Left
        <span class="margin-inches"
          >{ptsToInches($currentTemplate.margin_left)}in</span
        >
      </label>
      <input
        id="margin-left"
        type="number"
        class="prop-input prop-input-sm"
        min="0"
        max="288"
        step="1"
        value={$currentTemplate.margin_left}
        disabled={$currentTemplate.is_builtin}
        on:input={(e) =>
          updateMargin("margin_left", parseFloat(e.currentTarget.value))}
      />
    </div>

    <div class="prop-group">
      <label class="prop-label" for="margin-right">
        Right
        <span class="margin-inches"
          >{ptsToInches($currentTemplate.margin_right)}in</span
        >
      </label>
      <input
        id="margin-right"
        type="number"
        class="prop-input prop-input-sm"
        min="0"
        max="288"
        step="1"
        value={$currentTemplate.margin_right}
        disabled={$currentTemplate.is_builtin}
        on:input={(e) =>
          updateMargin("margin_right", parseFloat(e.currentTarget.value))}
      />
    </div>

    <span class="prop-hint"
      >Values in points (72pt = 1 inch). Range: 0-288pt.</span
    >
  </div>
{/if}
</div>

<style>
  /* --- Header --- */
  .properties-header {
    padding: 12px 16px;
    border-bottom: 1px solid #2a3a4a;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .header-left h3 {
    margin: 0;
    font-size: 0.95rem;
    color: #e0e0e0;
  }

  .header-icon {
    font-size: 0.95rem;
  }

  .element-type-badge {
    font-size: 0.7rem;
    padding: 2px 6px;
    border-radius: 3px;
    background-color: #2a3a4a;
    color: #7a8a9a;
    font-family: monospace;
  }

  /* --- Body --- */
  .properties-body {
    padding: 12px 16px;
    overflow-y: auto;
  }

  .section-title {
    font-size: 0.82rem;
    font-weight: 600;
    color: #c0d0e0;
    margin: 0 0 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  /* --- Property Groups --- */
  .prop-group {
    margin-bottom: 12px;
  }

  .prop-label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 0.78rem;
    color: #c0d0e0;
    margin-bottom: 4px;
  }

  .prop-label-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .prop-button {
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    background-color: #1e2d3d;
    color: #c0d0e0;
    font-size: 0.72rem;
    padding: 4px 8px;
    cursor: pointer;
  }

  .prop-button:hover {
    border-color: #4a8af4;
  }

  .prop-input {
    width: 100%;
    padding: 6px 8px;
    border: 1px solid #2a3a4a;
    border-radius: 3px;
    background-color: #1e2d3d;
    color: #e0e0e0;
    font-size: 0.82rem;
    outline: none;
    transition: border-color 0.15s;
  }

  .prop-input:focus {
    border-color: #4a8af4;
  }

  .prop-input-sm {
    width: 80px;
  }

  .prop-textarea {
    width: 100%;
    padding: 6px 8px;
    border: 1px solid #2a3a4a;
    border-radius: 3px;
    background-color: #1e2d3d;
    color: #e0e0e0;
    font-size: 0.82rem;
    font-family: inherit;
    outline: none;
    resize: vertical;
    min-height: 60px;
    transition: border-color 0.15s;
  }

  .prop-textarea:focus {
    border-color: #4a8af4;
  }

  .prop-select {
    width: 100%;
    padding: 6px 8px;
    border: 1px solid #2a3a4a;
    border-radius: 3px;
    background-color: #1e2d3d;
    color: #e0e0e0;
    font-size: 0.82rem;
    outline: none;
    cursor: pointer;
    transition: border-color 0.15s;
  }

  .prop-select:focus {
    border-color: #4a8af4;
  }

  .prop-row {
    display: flex;
    gap: 16px;
    margin-bottom: 10px;
  }

  .segment-card {
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 8px;
    margin-bottom: 10px;
    background-color: #172635;
  }

  .segment-card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .segment-index {
    font-size: 0.72rem;
    color: #7a8a9a;
    font-family: monospace;
  }

  .segment-remove {
    border: 1px solid #2a3a4a;
    border-radius: 3px;
    background-color: transparent;
    color: #c0d0e0;
    font-size: 0.72rem;
    padding: 2px 6px;
    cursor: pointer;
  }

  .segment-remove:hover {
    border-color: #c87070;
    color: #e09a9a;
  }

  .prop-toggle {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.78rem;
    color: #c0d0e0;
    cursor: pointer;
    user-select: none;
  }

  .prop-toggle input[type="checkbox"] {
    accent-color: #4a8af4;
    cursor: pointer;
  }

  .prop-hint {
    display: block;
    font-size: 0.7rem;
    color: #5a6a7a;
    margin-top: 3px;
  }

  .margin-inches {
    font-size: 0.7rem;
    color: #5a6a7a;
    font-weight: 400;
  }

  .config-preview {
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 10px;
    font-size: 0.72rem;
    color: #c0d0e0;
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
    max-height: 300px;
    overflow-y: auto;
  }

  .builtin-notice {
    font-size: 0.75rem;
    color: #7a8a9a;
    background-color: #1e2d3d;
    border: 1px solid #2a3a4a;
    border-radius: 4px;
    padding: 8px 10px;
    margin: 8px 0 4px;
    line-height: 1.4;
  }

  .properties-root.read-only :global(input),
  .properties-root.read-only :global(select),
  .properties-root.read-only :global(textarea),
  .properties-root.read-only :global(button) {
    pointer-events: none;
    opacity: 0.72;
  }
</style>
