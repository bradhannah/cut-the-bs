/**
 * Shared date formatting utilities for consistent display across
 * all frontend pages.
 */

const monthNames = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

/**
 * Format a date string based on its granularity for display.
 *
 * - "year"  -> "2024"
 * - "month" -> "Jun 2024"
 * - "day"   -> "Jun 15, 2024"
 *
 * If the date string is empty, returns the fallback (default "").
 */
export function formatDate(
  dateStr: string,
  granularity: string = "month",
  fallback: string = ""
): string {
  if (!dateStr) return fallback;

  if (granularity === "year") {
    return dateStr.substring(0, 4);
  }

  const parts = dateStr.split("-");

  if (granularity === "month" && parts.length >= 2) {
    const monthIdx = parseInt(parts[1], 10) - 1;
    if (monthIdx >= 0 && monthIdx < 12) {
      return `${monthNames[monthIdx]} ${parts[0]}`;
    }
    return dateStr.substring(0, 7);
  }

  // day or fallback: YYYY-MM-DD -> "Mon DD, YYYY"
  if (parts.length >= 3) {
    const monthIdx = parseInt(parts[1], 10) - 1;
    const day = parseInt(parts[2], 10);
    if (monthIdx >= 0 && monthIdx < 12 && !isNaN(day)) {
      return `${monthNames[monthIdx]} ${day}, ${parts[0]}`;
    }
  }

  return dateStr;
}

/**
 * Format an ISO timestamp for display. Used for generated_at,
 * changed_at, and other full timestamps.
 *
 * Example: "2024-06-15T14:30:00Z" -> "Jun 15, 2024 2:30 PM"
 */
export function formatTimestamp(isoStr: string): string {
  if (!isoStr) return "";
  try {
    const d = new Date(isoStr);
    return d.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return isoStr;
  }
}

/**
 * Format a date range string from start/end dates with their
 * granularities. Used by work history cards.
 *
 * Example: "Jun 2020 - Present"
 */
export function formatDateRange(
  startDate: string,
  startGranularity: string,
  endDate: string,
  endGranularity: string
): string {
  const start = formatDate(startDate, startGranularity, "Unknown");
  const end = formatDate(endDate, endGranularity, "Present");
  return `${start} - ${end}`;
}
