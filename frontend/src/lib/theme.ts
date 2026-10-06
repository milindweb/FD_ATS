// theme.ts — light / dark / system modes applied via data-theme on <html>
// (design.md §11). One component + theme tokens; no duplicated components.

export type ThemeMode = "light" | "dark" | "system";

const STORAGE_KEY = "fdats.theme";
const mediaQuery = () => window.matchMedia("(prefers-color-scheme: dark)");

function resolve(mode: ThemeMode): "light" | "dark" {
  if (mode === "system") return mediaQuery().matches ? "dark" : "light";
  return mode;
}

export function currentTheme(): ThemeMode {
  const stored = localStorage.getItem(STORAGE_KEY);
  return stored === "light" || stored === "dark" || stored === "system" ? stored : "light";
}

export function applyTheme(mode: ThemeMode): void {
  localStorage.setItem(STORAGE_KEY, mode);
  document.documentElement.setAttribute("data-theme", resolve(mode));
}

/** Applies the stored theme and keeps "system" mode in sync. Call once. */
export function initTheme(): void {
  applyTheme(currentTheme());
  mediaQuery().addEventListener("change", () => {
    if (currentTheme() === "system") applyTheme("system");
  });
}

/** Cycles light → dark → system and returns the new mode. */
export function cycleTheme(): ThemeMode {
  const order: ThemeMode[] = ["light", "dark", "system"];
  const next = order[(order.indexOf(currentTheme()) + 1) % order.length];
  applyTheme(next);
  return next;
}
