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
  start_date: string;
  end_date: string;
  date_granularity_start: string;
  date_granularity_end: string;
}

export interface WorkHistoryEntry {
  id: number;
  employer_name: string;
  job_title: string;
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
  text: string
): Promise<AchievementBullet> {
  return call<AchievementBullet>("CreateBullet", workHistoryID, text);
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

export async function reorderProfileLinks(
  orderedIDs: number[]
): Promise<void> {
  await call<void>("ReorderProfileLinks", orderedIDs);
}
