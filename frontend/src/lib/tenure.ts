// tenure.ts — calendar tenure helpers.
//
// The FD engine stores tenure in days, but the UI offers year/month presets.
// A "3 years" tenure must land on the same calendar date, not on a fixed
// 1,095-day count: 01-10-2023 + 3 years is 01-10-2026 (1,096 days, because
// 2024 is a leap year). These helpers convert a calendar tenure into the
// exact day count for the chosen start date.

export interface TenurePreset {
  months: number;
  label: string;
}

export const TENURE_PRESETS: readonly TenurePreset[] = [
  { months: 6, label: "6 months" },
  { months: 12, label: "1 year" },
  { months: 24, label: "2 years" },
  { months: 36, label: "3 years" },
];

const FALLBACK_DAYS: Record<number, number> = {
  6: 180,
  12: 365,
  24: 730,
  36: 1095,
};

/**
 * Calendar days from `startISO` to the same day `months` months later.
 * The day is clamped to the target month's last day, so 31 Jan + 1 month
 * becomes 28/29 Feb. Returns 0 when the date or tenure is invalid.
 */
export function calendarTenureDays(startISO: string, months: number): number {
  if (months <= 0 || !/^\d{4}-\d{2}-\d{2}$/.test(startISO)) return 0;

  const [year, month, day] = startISO.split("-").map(Number);
  const start = new Date(Date.UTC(year, month - 1, day));
  if (
    start.getUTCFullYear() !== year ||
    start.getUTCMonth() !== month - 1 ||
    start.getUTCDate() !== day
  ) {
    return 0;
  }

  const totalMonths = month - 1 + months;
  const targetYear = year + Math.floor(totalMonths / 12);
  const targetMonth = totalMonths % 12;
  const lastDay = new Date(Date.UTC(targetYear, targetMonth + 1, 0)).getUTCDate();
  const end = Date.UTC(targetYear, targetMonth, Math.min(day, lastDay));
  return Math.round((end - start.getTime()) / 86_400_000);
}

/** Days for a preset at the given start date, with a sensible fallback. */
export function tenurePresetDays(startISO: string, months: number): number {
  return calendarTenureDays(startISO, months) || FALLBACK_DAYS[months] || 0;
}

/** Returns the preset that matches `tenureDays` for `startISO`, if any. */
export function activeTenurePreset(startISO: string, tenureDays: string): TenurePreset | undefined {
  const days = Number(tenureDays);
  if (!Number.isFinite(days) || days <= 0) return undefined;
  return TENURE_PRESETS.find((preset) => tenurePresetDays(startISO, preset.months) === days);
}
