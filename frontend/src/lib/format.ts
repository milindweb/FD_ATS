// format.ts — presentation helpers. All formatting decisions live here so
// pages never hand-roll numbers or dates.

const inr = new Intl.NumberFormat("en-IN", {
  style: "currency",
  currency: "INR",
  maximumFractionDigits: 0,
  minimumFractionDigits: 0,
});

const plainIN = new Intl.NumberFormat("en-IN", { maximumFractionDigits: 0 });

/** ₹1,25,00,000 — Indian grouping, no paise (amounts are whole rupees). */
export function formatMoney(value: number | null | undefined): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "₹0";
  return inr.format(Math.round(value));
}

/** 1,25,00,000 without the currency symbol. */
export function formatNumber(value: number | null | undefined): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "0";
  return plainIN.format(Math.round(value));
}

/** 01/10/2026 — the display date format used across the app. */
export function formatDate(iso: string | null | undefined): string {
  if (!iso) return "—";
  const [y, m, d] = iso.split("-");
  if (!y || !m || !d) return iso;
  return `${d}/${m}/${y}`;
}

/** 8.00% */
export function formatPercent(rate: number | null | undefined): string {
  if (rate === null || rate === undefined || Number.isNaN(rate)) return "—";
  return `${rate.toFixed(2)}%`;
}

/** 365 Days */
export function formatTenure(days: number | null | undefined): string {
  if (!days && days !== 0) return "—";
  return `${formatNumber(days)} Days`;
}

/** Today in ISO format (local time). */
export function todayISO(): string {
  const d = new Date();
  const month = `${d.getMonth() + 1}`.padStart(2, "0");
  const day = `${d.getDate()}`.padStart(2, "0");
  return `${d.getFullYear()}-${month}-${day}`;
}

/** Parses a date input value (YYYY-MM-DD) into a local Date, or null. */
export function parseISO(value: string): Date | null {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return null;
  const [y, m, d] = value.split("-").map(Number);
  const date = new Date(y, m - 1, d);
  return Number.isNaN(date.getTime()) ? null : date;
}

/** Human day label for maturity tracking: "Today", "Tomorrow", "in 12 Days". */
export function daysLabel(days: number): string {
  if (days < 0) return `${formatNumber(-days)} Days ago`;
  if (days === 0) return "Today";
  if (days === 1) return "Tomorrow";
  return `in ${formatNumber(days)} Days`;
}
