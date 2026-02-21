// API service layer — typed wrappers around Wails-generated bindings.
// Provides error handling and a consistent interface for all pages.
//
// Note: The Wails bindings (wailsjs/go/main/App.js) are auto-generated
// at build time. This module re-exports them with proper typing and
// wraps calls with error handling for the toast notification system.

// Re-export types used across the frontend.
export interface WorkHistoryInput {
  employer_name: string;
  job_title: string;
  summary: string;
  start_date: string;
  end_date: string;
  date_granularity_start: string;
  date_granularity_end: string;
}

export interface WorkHistoryEntry {
  id: number;
  employer_name: string;
  job_title: string;
  summary: string;
  start_date: string;
  end_date: string;
  date_granularity_start: string;
  date_granularity_end: string;
  sort_order: number;
  bullets: AchievementBullet[];
  created_at: string;
  updated_at: string;
}

export interface AchievementBullet {
  id: number;
  work_history_id: number;
  text: string;
  bullet_type: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

// --- Profile Types ---

export interface UserProfile {
  id: number;
  full_name: string;
  email: string;
  phone: string;
  location: string;
}

export interface ProfileLink {
  id: number;
  label: string;
  url: string;
  sort_order: number;
}

export interface ProfileLinkInput {
  label: string;
  url: string;
}

// Toast notification support
export type ToastLevel = "info" | "success" | "error";

export interface ToastMessage {
  id: number;
  level: ToastLevel;
  message: string;
}

type ToastSubscriber = (toasts: ToastMessage[]) => void;

let nextToastId = 0;
let toasts: ToastMessage[] = [];
const subscribers: Set<ToastSubscriber> = new Set();

function notifySubscribers(): void {
  for (const sub of subscribers) {
    sub([...toasts]);
  }
}

export function subscribeToasts(fn: ToastSubscriber): () => void {
  subscribers.add(fn);
  fn([...toasts]);
  return () => {
    subscribers.delete(fn);
  };
}

export function addToast(
  level: ToastLevel,
  message: string,
  autoRemoveMs = 4000
): void {
  const id = nextToastId++;
  toasts = [...toasts, { id, level, message }];
  notifySubscribers();

  if (autoRemoveMs > 0) {
    setTimeout(() => removeToast(id), autoRemoveMs);
  }
}

export function removeToast(id: number): void {
  toasts = toasts.filter((t) => t.id !== id);
  notifySubscribers();
}

// --- Wails binding wrappers ---
// These call window['go']['main']['App'][method] directly.
// At build time, Wails injects these bindings onto the window object.
// We wrap them here with error handling and toast notifications.

/* eslint-disable @typescript-eslint/no-explicit-any */
function getBinding(method: string): (...args: any[]) => Promise<any> {
  return (window as any)?.go?.main?.App?.[method];
}
/* eslint-enable @typescript-eslint/no-explicit-any */

async function call<T>(method: string, ...args: unknown[]): Promise<T> {
  const fn = getBinding(method);
  if (!fn) {
    const msg = `Wails binding "${method}" not available`;
    addToast("error", msg);
    throw new Error(msg);
  }
  try {
    return (await fn(...args)) as T;
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err);
    addToast("error", message);
    throw err;
  }
}

// --- Work History API ---

export async function listWorkHistory(): Promise<WorkHistoryEntry[]> {
  return call<WorkHistoryEntry[]>("ListWorkHistory");
}

export async function getWorkHistory(id: number): Promise<WorkHistoryEntry> {
  return call<WorkHistoryEntry>("GetWorkHistory", id);
}

export async function createWorkHistory(
  input: WorkHistoryInput
): Promise<WorkHistoryEntry> {
  const result = await call<WorkHistoryEntry>("CreateWorkHistory", input);
  addToast("success", "Work history entry created");
  return result;
}

export async function updateWorkHistory(
  id: number,
  input: WorkHistoryInput
): Promise<WorkHistoryEntry> {
  const result = await call<WorkHistoryEntry>("UpdateWorkHistory", id, input);
  addToast("success", "Work history entry updated");
  return result;
}

export async function deleteWorkHistory(id: number): Promise<void> {
  await call<void>("DeleteWorkHistory", id);
  addToast("success", "Work history entry deleted");
}

export async function reorderWorkHistory(orderedIDs: number[]): Promise<void> {
  await call<void>("ReorderWorkHistory", orderedIDs);
}

// --- Bullet API ---

export async function createBullet(
  workHistoryID: number,
  text: string,
  bulletType: string = "primary"
): Promise<AchievementBullet> {
  return call<AchievementBullet>("CreateBullet", workHistoryID, text, bulletType);
}

export async function updateBullet(
  id: number,
  text: string
): Promise<AchievementBullet> {
  return call<AchievementBullet>("UpdateBullet", id, text);
}

export async function deleteBullet(id: number): Promise<void> {
  await call<void>("DeleteBullet", id);
}

export async function reorderBullets(
  workHistoryID: number,
  orderedIDs: number[]
): Promise<void> {
  await call<void>("ReorderBullets", workHistoryID, orderedIDs);
}

export async function splitBulletText(text: string): Promise<string[]> {
  return call<string[]>("SplitBulletText", text);
}

// --- Profile API ---

export async function getProfile(): Promise<UserProfile> {
  return call<UserProfile>("GetProfile");
}

export async function updateProfile(
  profile: UserProfile
): Promise<UserProfile> {
  const result = await call<UserProfile>("UpdateProfile", profile);
  addToast("success", "Profile updated");
  return result;
}

// --- Profile Links API ---

export async function listProfileLinks(): Promise<ProfileLink[]> {
  return call<ProfileLink[]>("ListProfileLinks");
}

export async function createProfileLink(
  input: ProfileLinkInput
): Promise<ProfileLink> {
  const result = await call<ProfileLink>("CreateProfileLink", input);
  addToast("success", "Link added");
  return result;
}

export async function updateProfileLink(
  id: number,
  input: ProfileLinkInput
): Promise<ProfileLink> {
  const result = await call<ProfileLink>("UpdateProfileLink", id, input);
  addToast("success", "Link updated");
  return result;
}

export async function deleteProfileLink(id: number): Promise<void> {
  await call<void>("DeleteProfileLink", id);
  addToast("success", "Link deleted");
}

export async function reorderProfileLinks(orderedIDs: number[]): Promise<void> {
  await call<void>("ReorderProfileLinks", orderedIDs);
}

// --- Skill Types ---

export interface SkillInput {
  name: string;
  category_id: number;
  competence_level: number;
  is_legacy: boolean;
}

export interface Skill {
  id: number;
  name: string;
  category_id: number;
  competence_level: number;
  is_legacy: boolean;
}

export interface SkillCategory {
  id: number;
  name: string;
  sort_order: number;
}

export interface SkillCategoryWithSkills {
  category: SkillCategory;
  skills: Skill[];
}

export interface CompetenceLevel {
  level: number;
  label: string;
  description: string;
}

// --- Skills API ---

export async function listSkills(): Promise<Skill[]> {
  return call<Skill[]>("ListSkills");
}

export async function listSkillsByCategory(): Promise<
  SkillCategoryWithSkills[]
> {
  return call<SkillCategoryWithSkills[]>("ListSkillsByCategory");
}

export async function createSkill(input: SkillInput): Promise<Skill> {
  const result = await call<Skill>("CreateSkill", input);
  addToast("success", "Skill created");
  return result;
}

export async function updateSkill(
  id: number,
  input: SkillInput
): Promise<Skill> {
  const result = await call<Skill>("UpdateSkill", id, input);
  addToast("success", "Skill updated");
  return result;
}

export async function deleteSkill(id: number): Promise<void> {
  await call<void>("DeleteSkill", id);
  addToast("success", "Skill deleted");
}

export async function checkSkillLensReferences(id: number): Promise<string[]> {
  return call<string[]>("CheckSkillLensReferences", id);
}

export async function splitSkillsText(text: string): Promise<string[]> {
  return call<string[]>("SplitSkillsText", text);
}

export async function getCompetenceLevels(): Promise<CompetenceLevel[]> {
  return call<CompetenceLevel[]>("GetCompetenceLevels");
}

// --- Skill Categories API ---

export async function listSkillCategories(): Promise<SkillCategory[]> {
  return call<SkillCategory[]>("ListSkillCategories");
}

export async function createSkillCategory(
  name: string
): Promise<SkillCategory> {
  const result = await call<SkillCategory>("CreateSkillCategory", name);
  addToast("success", "Category created");
  return result;
}

export async function renameSkillCategory(
  id: number,
  name: string
): Promise<SkillCategory> {
  const result = await call<SkillCategory>("RenameSkillCategory", id, name);
  addToast("success", "Category renamed");
  return result;
}

export async function deleteSkillCategory(id: number): Promise<void> {
  await call<void>("DeleteSkillCategory", id);
  addToast("success", "Category deleted");
}

export async function reorderSkillCategories(
  orderedIDs: number[]
): Promise<void> {
  await call<void>("ReorderSkillCategories", orderedIDs);
}

// --- Academic Types ---

export interface AcademicInput {
  institution: string;
  credential_type: string;
  field_of_study: string;
  completion_date: string;
  date_granularity: string;
}

export interface AcademicCredential {
  id: number;
  institution: string;
  credential_type: string;
  field_of_study: string;
  completion_date: string;
  date_granularity: string;
  sort_order: number;
}

// --- Academic API ---

export async function listAcademicCredentials(): Promise<AcademicCredential[]> {
  return call<AcademicCredential[]>("ListAcademicCredentials");
}

export async function createAcademicCredential(
  input: AcademicInput
): Promise<AcademicCredential> {
  const result = await call<AcademicCredential>(
    "CreateAcademicCredential",
    input
  );
  addToast("success", "Academic credential created");
  return result;
}

export async function updateAcademicCredential(
  id: number,
  input: AcademicInput
): Promise<AcademicCredential> {
  const result = await call<AcademicCredential>(
    "UpdateAcademicCredential",
    id,
    input
  );
  addToast("success", "Academic credential updated");
  return result;
}

export async function deleteAcademicCredential(id: number): Promise<void> {
  await call<void>("DeleteAcademicCredential", id);
  addToast("success", "Academic credential deleted");
}

export async function reorderAcademicCredentials(
  orderedIDs: number[]
): Promise<void> {
  await call<void>("ReorderAcademicCredentials", orderedIDs);
}

// --- Certification Types ---

export interface CertificationInput {
  name: string;
  issuing_body: string;
  date_earned: string;
  expiration_date: string;
}

export interface Certification {
  id: number;
  name: string;
  issuing_body: string;
  date_earned: string;
  expiration_date: string;
  is_active: boolean;
  sort_order: number;
}

// --- Certification API ---

export async function listCertifications(): Promise<Certification[]> {
  return call<Certification[]>("ListCertifications");
}

export async function createCertification(
  input: CertificationInput
): Promise<Certification> {
  const result = await call<Certification>("CreateCertification", input);
  addToast("success", "Certification created");
  return result;
}

export async function updateCertification(
  id: number,
  input: CertificationInput
): Promise<Certification> {
  const result = await call<Certification>("UpdateCertification", id, input);
  addToast("success", "Certification updated");
  return result;
}

export async function deleteCertification(id: number): Promise<void> {
  await call<void>("DeleteCertification", id);
  addToast("success", "Certification deleted");
}

export async function reorderCertifications(
  orderedIDs: number[]
): Promise<void> {
  await call<void>("ReorderCertifications", orderedIDs);
}

// --- Summary Types ---

export interface SummaryInput {
  label: string;
  body_text: string;
}

export interface ProfessionalSummary {
  id: number;
  label: string;
  body_text: string;
}

// --- Summary API ---

export async function listSummaries(): Promise<ProfessionalSummary[]> {
  return call<ProfessionalSummary[]>("ListSummaries");
}

export async function createSummary(
  input: SummaryInput
): Promise<ProfessionalSummary> {
  const result = await call<ProfessionalSummary>("CreateSummary", input);
  addToast("success", "Summary created");
  return result;
}

export async function updateSummary(
  id: number,
  input: SummaryInput
): Promise<ProfessionalSummary> {
  const result = await call<ProfessionalSummary>("UpdateSummary", id, input);
  addToast("success", "Summary updated");
  return result;
}

export async function deleteSummary(id: number): Promise<void> {
  await call<void>("DeleteSummary", id);
  addToast("success", "Summary deleted");
}

// --- Descriptor Types ---

export interface RoleDescriptor {
  id: number;
  title: string;
  sort_order: number;
}

// --- Descriptor API ---

export async function listDescriptors(): Promise<RoleDescriptor[]> {
  return call<RoleDescriptor[]>("ListDescriptors");
}

export async function createDescriptor(title: string): Promise<RoleDescriptor> {
  const result = await call<RoleDescriptor>("CreateDescriptor", title);
  addToast("success", "Descriptor created");
  return result;
}

export async function updateDescriptor(
  id: number,
  title: string
): Promise<RoleDescriptor> {
  const result = await call<RoleDescriptor>("UpdateDescriptor", id, title);
  addToast("success", "Descriptor updated");
  return result;
}

export async function deleteDescriptor(id: number): Promise<void> {
  await call<void>("DeleteDescriptor", id);
  addToast("success", "Descriptor deleted");
}

export async function reorderDescriptors(orderedIDs: number[]): Promise<void> {
  await call<void>("ReorderDescriptors", orderedIDs);
}

// --- Core Expertise Types ---

export interface CoreExpertise {
  id: number;
  label: string;
  sort_order: number;
}

// --- Core Expertise API ---

export async function listCoreExpertise(): Promise<CoreExpertise[]> {
  return call<CoreExpertise[]>("ListCoreExpertise");
}

export async function createCoreExpertise(label: string): Promise<CoreExpertise> {
  const result = await call<CoreExpertise>("CreateCoreExpertise", label);
  addToast("success", "Core expertise created");
  return result;
}

export async function updateCoreExpertise(
  id: number,
  label: string
): Promise<CoreExpertise> {
  const result = await call<CoreExpertise>("UpdateCoreExpertise", id, label);
  addToast("success", "Core expertise updated");
  return result;
}

export async function deleteCoreExpertise(id: number): Promise<void> {
  await call<void>("DeleteCoreExpertise", id);
  addToast("success", "Core expertise deleted");
}

export async function reorderCoreExpertise(orderedIDs: number[]): Promise<void> {
  await call<void>("ReorderCoreExpertise", orderedIDs);
}

export async function checkCoreExpertiseLensReferences(id: number): Promise<string[]> {
  return call<string[]>("CheckCoreExpertiseLensReferences", id);
}

export async function splitCoreExpertiseText(text: string): Promise<string[]> {
  return call<string[]>("SplitCoreExpertiseText", text);
}

// --- Resume Export Types ---

export interface ResumeTemplate {
  id: string;
  name: string;
  description: string;
  preview_url: string;
}

export interface ExportRequest {
  template_id: string;
  lens_id?: number | null;
  summary_ids: number[];
  master_summary_id?: number | null;
  work_history_ids: number[];
  bullet_ids: number[];
  skill_ids: number[];
  skill_sort_overrides: Record<number, number>;
  academic_ids: number[];
  certification_ids: number[];
  descriptor_ids: number[];
  core_expertise_ids: number[];
}

export interface ResumeExport {
  id: number;
  template_id: string;
  file_path: string;
  summary_id: number | null;
  lens_id: number | null;
  generated_at: string;
}

// --- Resume Export API ---

export async function listTemplates(): Promise<ResumeTemplate[]> {
  return call<ResumeTemplate[]>("ListTemplates");
}

export async function previewExport(req: ExportRequest): Promise<string> {
  return call<string>("PreviewExport", req);
}

export async function createExport(req: ExportRequest): Promise<ResumeExport> {
  const result = await call<ResumeExport>("CreateExport", req);
  addToast("success", "Resume exported successfully");
  return result;
}

export async function listExports(): Promise<ResumeExport[]> {
  return call<ResumeExport[]>("ListExports");
}

export async function openExportFile(exportID: number): Promise<void> {
  await call<void>("OpenExportFile", exportID);
}

// --- Cover Letter Types ---

export interface CoverLetterInput {
  title: string;
  body_text: string;
}

export interface CoverLetter {
  id: number;
  title: string;
  body_text: string;
  file_path: string;
}

// --- Cover Letter API ---

export async function listCoverLetters(): Promise<CoverLetter[]> {
  return call<CoverLetter[]>("ListCoverLetters");
}

export async function createCoverLetter(
  input: CoverLetterInput
): Promise<CoverLetter> {
  const result = await call<CoverLetter>("CreateCoverLetter", input);
  addToast("success", "Cover letter created");
  return result;
}

export async function updateCoverLetter(
  id: number,
  input: CoverLetterInput
): Promise<CoverLetter> {
  const result = await call<CoverLetter>("UpdateCoverLetter", id, input);
  addToast("success", "Cover letter updated");
  return result;
}

export async function deleteCoverLetter(id: number): Promise<void> {
  await call<void>("DeleteCoverLetter", id);
  addToast("success", "Cover letter deleted");
}

export async function exportCoverLetter(id: number): Promise<string> {
  const filePath = await call<string>("ExportCoverLetter", id);
  addToast("success", "Cover letter exported to PDF");
  return filePath;
}

// --- Job Application Types ---

export interface ApplicationInput {
  company_name: string;
  position_title: string;
  date_applied: string;
  fit_indicator: string;
  resume_export_id: number | null;
  cover_letter_id: number | null;
  notes: string;
}

export interface JobApplication {
  id: number;
  company_name: string;
  position_title: string;
  date_applied: string;
  status: string;
  fit_indicator: string;
  resume_export_id: number | null;
  cover_letter_id: number | null;
  notes: string;
}

export interface StatusChange {
  id: number;
  from_status: string;
  to_status: string;
  changed_at: string;
}

// --- Job Application API ---

export async function listApplications(): Promise<JobApplication[]> {
  return call<JobApplication[]>("ListApplications");
}

export async function searchApplications(
  query: string
): Promise<JobApplication[]> {
  return call<JobApplication[]>("SearchApplications", query);
}

export async function createApplication(
  input: ApplicationInput
): Promise<JobApplication> {
  const result = await call<JobApplication>("CreateApplication", input);
  addToast("success", "Application created");
  return result;
}

export async function updateApplication(
  id: number,
  input: ApplicationInput
): Promise<JobApplication> {
  const result = await call<JobApplication>("UpdateApplication", id, input);
  addToast("success", "Application updated");
  return result;
}

export async function updateApplicationStatus(
  id: number,
  newStatus: string
): Promise<JobApplication> {
  const result = await call<JobApplication>(
    "UpdateApplicationStatus",
    id,
    newStatus
  );
  addToast("success", `Status changed to ${newStatus}`);
  return result;
}

export async function updateApplicationFit(
  id: number,
  fitIndicator: string
): Promise<JobApplication> {
  const result = await call<JobApplication>(
    "UpdateApplicationFit",
    id,
    fitIndicator
  );
  addToast("success", "Fit indicator updated");
  return result;
}

export async function getApplicationHistory(
  id: number
): Promise<StatusChange[]> {
  return call<StatusChange[]>("GetApplicationHistory", id);
}

export async function deleteApplication(id: number): Promise<void> {
  await call<void>("DeleteApplication", id);
  addToast("success", "Application deleted");
}

export async function getApplicationStatuses(): Promise<string[]> {
  return call<string[]>("GetApplicationStatuses");
}

export async function getFitIndicators(): Promise<string[]> {
  return call<string[]>("GetFitIndicators");
}

// --- Lens Types ---

export interface LensInput {
  name: string;
}

export interface Lens {
  id: number;
  name: string;
}

export interface LensWorkHistoryItem {
  work_history_id: number;
  sort_order: number;
}

export interface LensBulletItem {
  bullet_id: number;
  sort_order: number;
}

export interface LensSkillItem {
  skill_id: number;
  custom_sort_order: number | null;
}

export interface LensDescriptorItem {
  descriptor_id: number;
  sort_order: number;
}

export interface LensCoreExpertiseItem {
  core_expertise_id: number;
  sort_order: number;
}

export interface LensSummaryItem {
  summary_id: number;
  sort_order: number;
  is_master: boolean;
}

export interface LensDetail {
  id: number;
  name: string;
  summaries: LensSummaryItem[];
  work_history: LensWorkHistoryItem[];
  bullets: LensBulletItem[];
  skills: LensSkillItem[];
  academic_ids: number[];
  cert_ids: number[];
  descriptors: LensDescriptorItem[];
  core_expertise: LensCoreExpertiseItem[];
}

export interface SkillWithTags {
  id: number;
  name: string;
  category_id: number;
  competence_level: number;
  is_legacy: boolean;
  lens_ids: number[];
}

// --- Lens API ---

export async function listLenses(): Promise<Lens[]> {
  return call<Lens[]>("ListLenses");
}

export async function getLens(id: number): Promise<LensDetail> {
  return call<LensDetail>("GetLens", id);
}

export async function createLens(input: LensInput): Promise<Lens> {
  const result = await call<Lens>("CreateLens", input);
  addToast("success", "Lens created");
  return result;
}

export async function updateLens(
  id: number,
  input: LensInput
): Promise<Lens> {
  const result = await call<Lens>("UpdateLens", id, input);
  addToast("success", "Lens updated");
  return result;
}

export async function deleteLens(id: number): Promise<void> {
  await call<void>("DeleteLens", id);
  addToast("success", "Lens deleted");
}

export async function setLensWorkHistory(
  lensID: number,
  selections: LensWorkHistoryItem[]
): Promise<void> {
  await call<void>("SetLensWorkHistory", lensID, selections);
}

export async function setLensBullets(
  lensID: number,
  selections: LensBulletItem[]
): Promise<void> {
  await call<void>("SetLensBullets", lensID, selections);
}

export async function setLensSkills(
  lensID: number,
  selections: LensSkillItem[]
): Promise<void> {
  await call<void>("SetLensSkills", lensID, selections);
}

export async function setLensAcademics(
  lensID: number,
  academicIDs: number[]
): Promise<void> {
  await call<void>("SetLensAcademics", lensID, academicIDs);
}

export async function setLensCerts(
  lensID: number,
  certIDs: number[]
): Promise<void> {
  await call<void>("SetLensCerts", lensID, certIDs);
}

export async function setLensDescriptors(
  lensID: number,
  selections: LensDescriptorItem[]
): Promise<void> {
  await call<void>("SetLensDescriptors", lensID, selections);
}

export async function setLensCoreExpertise(
  lensID: number,
  selections: LensCoreExpertiseItem[]
): Promise<void> {
  await call<void>("SetLensCoreExpertise", lensID, selections);
}

export async function setLensSummaries(
  lensID: number,
  selections: LensSummaryItem[]
): Promise<void> {
  await call<void>("SetLensSummaries", lensID, selections);
}

export async function getLensExportSelections(
  lensID: number
): Promise<ExportRequest> {
  return call<ExportRequest>("GetLensExportSelections", lensID);
}

// --- Lens Reference Check API ---

export async function checkWorkHistoryLensReferences(
  id: number
): Promise<string[]> {
  return call<string[]>("CheckWorkHistoryLensReferences", id);
}

export async function checkBulletLensReferences(
  id: number
): Promise<string[]> {
  return call<string[]>("CheckBulletLensReferences", id);
}

export async function checkAcademicLensReferences(
  id: number
): Promise<string[]> {
  return call<string[]>("CheckAcademicLensReferences", id);
}

export async function checkCertLensReferences(
  id: number
): Promise<string[]> {
  return call<string[]>("CheckCertLensReferences", id);
}

export async function checkDescriptorLensReferences(
  id: number
): Promise<string[]> {
  return call<string[]>("CheckDescriptorLensReferences", id);
}

export async function checkSummaryLensReferences(
  id: number
): Promise<string[]> {
  return call<string[]>("CheckSummaryLensReferences", id);
}

// --- Skill Lens Tag API ---

export async function getSkillLensTags(skillID: number): Promise<number[]> {
  return call<number[]>("GetSkillLensTags", skillID);
}

export async function setSkillLensTags(
  skillID: number,
  lensIDs: number[]
): Promise<void> {
  await call<void>("SetSkillLensTags", skillID, lensIDs);
  addToast("success", "Skill lens tags updated");
}

export async function listSkillsWithLensTags(): Promise<SkillWithTags[]> {
  return call<SkillWithTags[]>("ListSkillsWithLensTags");
}

// --- Data Management Types ---

export interface ImportResult {
  records_imported: number;
  records_skipped: number;
  errors: string[];
}

export interface BackupSettings {
  rolling_backup_count: number;
}

// --- Data Management API ---

export async function exportAllData(outputPath: string): Promise<string> {
  const result = await call<string>("ExportAllData", outputPath);
  addToast("success", "Data exported successfully");
  return result;
}

export async function importAllData(inputPath: string): Promise<void> {
  await call<void>("ImportAllData", inputPath);
  addToast("success", "Data imported successfully — please reload");
}

export async function importCSV(
  filePath: string,
  dataType: string
): Promise<ImportResult> {
  const result = await call<ImportResult>("ImportCSV", filePath, dataType);
  addToast(
    "success",
    `Imported ${result.records_imported} records (${result.records_skipped} skipped)`
  );
  return result;
}

export async function importJSON(
  filePath: string,
  dataType: string
): Promise<ImportResult> {
  const result = await call<ImportResult>("ImportJSON", filePath, dataType);
  addToast(
    "success",
    `Imported ${result.records_imported} records (${result.records_skipped} skipped)`
  );
  return result;
}

export async function getDataDirectory(): Promise<string> {
  return call<string>("GetDataDirectory");
}

export async function setDataDirectory(path: string): Promise<void> {
  await call<void>("SetDataDirectory", path);
  addToast("success", "Data directory updated — restart to apply");
}

export async function getBackupSettings(): Promise<BackupSettings> {
  return call<BackupSettings>("GetBackupSettings");
}

export async function updateBackupSettings(
  settings: BackupSettings
): Promise<void> {
  await call<void>("UpdateBackupSettings", settings);
  addToast("success", "Backup settings updated");
}

export async function openDataDirectory(): Promise<void> {
  await call<void>("OpenDataDirectory");
}
