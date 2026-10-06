// statusTone.ts — one map from status key → visual tone (design.md §8).
// The status key always comes from the API; colour never encodes meaning on
// its own because StatusBadge always renders the text too.

export type Tone = "neutral" | "info" | "success" | "warning" | "danger" | "brand";

const toneMap: Record<string, Tone> = {
  // FD status
  ACTIVE: "success",
  CLOSED: "neutral",
  // Closure types
  MATURED: "success",
  PREMATURE: "warning",
  RENEWED: "info",
  // History events
  OPEN: "brand",
  RENEW: "info",
  CLOSE: "neutral",
};

export function statusTone(key: string | null | undefined): Tone {
  if (!key) return "neutral";
  return toneMap[key.toUpperCase()] ?? "neutral";
}

/** Short human label for keys shown in the UI. */
export function statusLabel(key: string | null | undefined): string {
  if (!key) return "—";
  const upper = key.toUpperCase();
  switch (upper) {
    case "OPEN":
      return "Opened";
    case "RENEW":
      return "Renewed";
    case "CLOSE":
      return "Closed";
    default:
      // ACTIVE / CLOSED / MATURED / PREMATURE / RENEWED read fine as-is.
      return key;
  }
}
