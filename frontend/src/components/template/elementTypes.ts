// Default element configurations for newly created template elements.
// Used by both Palette (for display) and Canvas (on drop) to initialize
// new elements with sensible defaults.

export const defaultConfigs: Record<string, string> = {
  profile_header: JSON.stringify({
    name_font_size: 18,
    detail_font_size: 10,
    alignment: "center",
    link_separator: " | ",
    show_links: true,
    show_links_inline: false,
    space_after: 6,
  }),
  role_descriptors: JSON.stringify({
    separator: " | ",
    // "normal" (not italic) matches the Professional template and renders cleaner.
    font_style: "normal",
    font_size: 10,
    alignment: "center",
    space_after: 4,
  }),
  professional_summary: JSON.stringify({
    show_master: true,
    show_bullet_summaries: true,
    enable_bullets: true,
    bullet_char: "\u2022",
    font_size: 10,
    space_before: 0,
    space_after: 2,
  }),
  skills: JSON.stringify({
    font_size: 10,
    group_by_category: true,
    // Makes category labels render bold — matches Professional template.
    category_font_style: "bold",
    include_legacy: true,
    legacy_suffix: " (Legacy)",
    skill_separator: ", ",
    space_after: 2,
  }),
  core_expertise: JSON.stringify({
    font_size: 10,
    separator: " | ",
    alignment: "center",
    space_after: 4,
  }),
  section_heading: JSON.stringify({
    text: "Section Title",
    font_size: 12,
    // Go reads font_style ("bold"/"normal"/"italic"), NOT a boolean "bold" field.
    font_style: "bold",
    uppercase: true,
    underline: true,
    underline_weight: 0.5,
    space_before: 10,
    space_after: 4,
    data_binding: "",
  }),
  horizontal_rule: JSON.stringify({
    weight: 0.5,
    space_before: 4,
    space_after: 4,
  }),
  spacer: JSON.stringify({ height: 10 }),
  static_text: JSON.stringify({
    text: "Enter text here",
    font_size: 10,
    font_style: "normal",
    alignment: "left",
    space_after: 4,
  }),
  work_history_loop: JSON.stringify({ entry_gap: 8 }),
  // 0 gap matches Professional — edu/cert entries are tight by design.
  education_loop: JSON.stringify({ entry_gap: 0 }),
  certifications_loop: JSON.stringify({ entry_gap: 0 }),
  work_title: JSON.stringify({
    font_size: 10,
    font_style: "bold",
    title_row_layout: "inline_with_dates",
    // Show employer inline with job title — matches Professional template.
    include_employer: true,
    employer_separator: " - ",
    employer_font_style: "italic",
    // Space between title row and next element (dates/bullets).
    // 13pt prevents dates from colliding with the title text.
    space_after: 13,
  }),
  work_employer: JSON.stringify({
    font_size: 10,
    font_style: "normal",
    space_after: 0,
  }),
  work_dates: JSON.stringify({
    font_size: 9,
    alignment: "right",
  }),
  work_summary: JSON.stringify({
    font_size: 10,
    font_style: "italic",
    space_after: 2,
  }),
  work_bullets: JSON.stringify({
    bullet_char: "\u2022",
    font_size: 10,
    font_style: "normal",
    indent: 12,
    bullet_sym_width: 10,
  }),
  work_outcomes: JSON.stringify({
    bullet_char: "\u2022",
    font_size: 10,
    font_style: "normal",
    indent: 12,
    bullet_sym_width: 10,
    outcomes_label: "Key Outcomes:",
    outcomes_gap: 2,
  }),
  edu_credential: JSON.stringify({
    font_size: 10,
    font_style: "bold",
  }),
  edu_institution: JSON.stringify({
    font_size: 10,
    font_style: "normal",
  }),
  edu_date: JSON.stringify({
    font_size: 9,
    alignment: "right",
  }),
  cert_name: JSON.stringify({
    font_size: 10,
    font_style: "bold",
  }),
  cert_detail: JSON.stringify({
    font_size: 10,
    font_style: "normal",
  }),
  // Cover letter types
  body_text: JSON.stringify({
    font_size: 11,
    line_spacing: 1.15,
    space_after: 12,
  }),
  date: JSON.stringify({
    font_size: 11,
    format: "January 2, 2006",
    alignment: "left",
    space_after: 12,
  }),
  greeting: JSON.stringify({
    text: "Dear Hiring Manager,",
    font_size: 11,
    space_after: 12,
  }),
  closing: JSON.stringify({
    text: "Sincerely,",
    font_size: 11,
    space_after: 24,
  }),
  recipient_address: JSON.stringify({
    font_size: 11,
    space_after: 12,
  }),
};

// Human-readable labels for element types.
export const elementLabels: Record<string, string> = {
  profile_header: "Profile Header",
  role_descriptors: "Role Descriptors",
  professional_summary: "Summary",
  skills: "Skills",
  core_expertise: "Core Expertise",
  section_heading: "Section Heading",
  horizontal_rule: "Horizontal Rule",
  spacer: "Spacer",
  static_text: "Static Text",
  work_history_loop: "Work History Loop",
  education_loop: "Education Loop",
  certifications_loop: "Certifications Loop",
  work_title: "Work Title",
  work_employer: "Work Employer",
  work_dates: "Work Dates",
  work_summary: "Work Summary",
  work_bullets: "Work Bullets",
  work_outcomes: "Work Outcomes",
  edu_credential: "Education Credential",
  edu_institution: "Education Institution",
  edu_date: "Education Date",
  cert_name: "Certification Name",
  cert_detail: "Certification Detail",
  body_text: "Body Text",
  date: "Date",
  greeting: "Greeting",
  closing: "Closing",
  recipient_address: "Recipient Address",
};

// Icons for element types.
export const elementIcons: Record<string, string> = {
  profile_header: "\u{1F464}",
  role_descriptors: "\u{1F4CB}",
  professional_summary: "\u{1F4DD}",
  skills: "\u{1F527}",
  core_expertise: "\u{2B50}",
  section_heading: "\u{1F524}",
  horizontal_rule: "\u{2500}",
  spacer: "\u{2195}",
  static_text: "\u{1F4C4}",
  work_history_loop: "\u{1F504}",
  education_loop: "\u{1F504}",
  certifications_loop: "\u{1F504}",
  work_title: "\u{1F4BC}",
  work_employer: "\u{1F3E2}",
  work_dates: "\u{1F4C5}",
  work_summary: "\u{1F4DD}",
  work_bullets: "\u{2022}",
  work_outcomes: "\u{1F3AF}",
  edu_credential: "\u{1F393}",
  edu_institution: "\u{1F3EB}",
  edu_date: "\u{1F4C5}",
  cert_name: "\u{1F4DC}",
  cert_detail: "\u{1F4DC}",
  body_text: "\u{1F4DD}",
  date: "\u{1F4C5}",
  greeting: "\u{1F44B}",
  closing: "\u{270D}",
  recipient_address: "\u{1F4EC}",
};

// Formatting element types that can be used multiple times on the canvas.
export const repeatableElementTypes = new Set([
  "section_heading",
  "horizontal_rule",
  "spacer",
  "static_text",
]);

// Loop container types.
export const loopElementTypes = new Set([
  "work_history_loop",
  "education_loop",
  "certifications_loop",
]);

// Valid children per loop type (mirrors Go domain.ValidLoopChildren).
export const validLoopChildren: Record<string, Set<string>> = {
  work_history_loop: new Set([
    "work_title", "work_employer", "work_dates",
    "work_summary", "work_bullets", "work_outcomes",
    "section_heading", "horizontal_rule", "spacer", "static_text",
  ]),
  education_loop: new Set([
    "edu_credential", "edu_institution", "edu_date",
    "section_heading", "horizontal_rule", "spacer", "static_text",
  ]),
  certifications_loop: new Set([
    "cert_name", "cert_detail",
    "section_heading", "horizontal_rule", "spacer", "static_text",
  ]),
};

// Loop-specific child elements shown in dynamic "* Items" categories.
// Shared formatting types are intentionally excluded (they remain in
// the Formatting palette category only).
export const loopSpecificChildren: Record<string, string[]> = {
  work_history_loop: [
    "work_title",
    "work_employer",
    "work_dates",
    "work_summary",
    "work_bullets",
    "work_outcomes",
  ],
  education_loop: ["edu_credential", "edu_institution", "edu_date"],
  certifications_loop: ["cert_name", "cert_detail"],
};

// Labels for dynamic palette categories shown when loop containers exist.
export const loopCategoryNames: Record<string, string> = {
  work_history_loop: "Work History Items",
  education_loop: "Education Items",
  certifications_loop: "Certification Items",
};

// Loop iteration labels for UI display.
export const loopIterationLabels: Record<string, string> = {
  work_history_loop: "Repeats for each work history entry",
  education_loop: "Repeats for each education entry",
  certifications_loop: "Repeats for each certification",
};

// Shared formatting element types valid for both template types.
const sharedTypes = new Set([
  "profile_header",
  "section_heading",
  "horizontal_rule",
  "spacer",
  "static_text",
]);

// Top-level element types valid for resume templates.
export const resumeElementTypes = new Set([
  ...sharedTypes,
  "role_descriptors",
  "professional_summary",
  "skills",
  "core_expertise",
  "work_history_loop",
  "education_loop",
  "certifications_loop",
]);

// Top-level element types valid for cover letter templates.
export const coverLetterElementTypes = new Set([
  ...sharedTypes,
  "body_text",
  "date",
  "greeting",
  "closing",
  "recipient_address",
]);

// Returns whether an element type is valid for a given template type.
export function isElementTypeValid(
  elementType: string,
  templateType: string
): boolean {
  if (templateType === "cover_letter") {
    return coverLetterElementTypes.has(elementType);
  }
  return resumeElementTypes.has(elementType);
}
