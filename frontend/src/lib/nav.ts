// nav.ts — one navigation source used by both the header and the mobile
// drawer (design.md §5). Adding a screen means adding one entry here.

import type { IconName } from "./icons";

export interface NavItem {
  to: string;
  label: string;
  icon: IconName;
}

export const navItems: NavItem[] = [
  { to: "/", label: "Dashboard", icon: "grid" },
  { to: "/new", label: "New FD", icon: "plus" },
  { to: "/reports", label: "Reports", icon: "sheet" },
  { to: "/settings", label: "Settings", icon: "gear" },
];
