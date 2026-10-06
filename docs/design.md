# design.md — Project Design Authority (Housing Society Web App)

> This file is the **project-level design authority** named by `AGENTS.md`.
> The **universal visual specification** is `DESIGN-General.md` (2462 lines) — read it first; it is
> binding for every visual decision. This file maps that specification onto *this* application so
> no implementer has to reinterpret it.

---

## 1. Authority Order

```text
Doc/DESIGN.md          = universal design specification (structure, tokens, components, a11y, motion)
design.md (this file)  = project mapping: the concrete tokens, component inventory, layouts, states
frontend/src/styles/*  = the ACTUAL implemented tokens/components  (code = current reality)
```

If `DESIGN.md` describes something this project does not implement, it is **not implemented**.
`DESIGN.md` §97 (AI Must Not Guess) and §98 (Docs vs Implementation) apply at all times.
Actual implementation state is recorded in `Project_Status.md`.

---

## 2. Stack-Specific Decisions

| DESIGN.md topic | This project's implementation |
|---|---|
| §5 Design tokens | CSS custom properties in `../frontend/src/styles/tokens.css` — single source of truth |
| §89 CSS architecture | `tokens.css` → `base.css` → `components.css` → `utilities.css` → `print.css` (no CSS-in-JS, no Tailwind, no UI kit) |
| §86 Reusable components | `../frontend/src/components/ui/**`, one class family `hs-<component>` each |
| §63 Icons | `../frontend/src/components/ui/Icon.tsx` inline SVG registry (no icon dependency) |
| §80 Charts | `../frontend/src/components/data/*Chart.tsx` small SVG charts using chart colour tokens |
| §11 Theme modes | `light` / `dark` / `system`, applied via `data-theme` on `<html>`, controlled by `state/themeStore` |
| §19 Density modes | **Not implemented** — only default compact density (documented, not claimed) |
| §70–72 Motion | `--motion-fast/normal/slow` + `prefers-reduced-motion` guard in `base.css` |

---

## 3. Concrete Token Set (implemented in `styles/tokens.css`)

```text
COLOUR (semantic roles from DESIGN.md §7)
  --color-bg, --color-surface, --color-surface-raised, --color-surface-hover,
  --color-border, --color-border-strong, --color-text, --color-text-secondary,
  --color-text-muted, --color-brand, --color-brand-hover, --color-brand-active,
  --color-success, --color-warning, --color-danger, --color-info,
  --color-focus-ring
  (+ *-soft variants used for badges/alerts)
  NAVY ACCENT (project palette)
    --color-brand #1f4e79  --color-brand-hover #1a4267  --color-brand-active #153755
    --color-table-head-bg / --color-table-head-text
      light #1f4e79 / #ffffff      dark #222834 / #e7ecf3
    --color-violet / --color-violet-soft
      light #6d4fd0 / #efeaff      dark #9b8af5 / #241f35

TYPOGRAPHY (DESIGN.md §14–§15)
  --font-sans          13px body / 12px metadata / 14px controls / 22px page title
  Font file: Nunito variable woff2, font-weight: 200 1000 (one file serves all weights)
  Never add -webkit-font-smoothing: antialiased — it makes bold text look blurred
  --font-mono          IDs, numbers, audit values only
  --text-xs .75rem  --text-sm .8125rem  --text-base .875rem  --text-lg .9375rem
  --text-xl 1.0625rem  --text-2xl 1.25rem  --text-3xl 1.5rem
  --weight-normal 400  --weight-medium 500  --weight-semibold 600  --weight-bold 700
  --leading-tight 1.25  --leading-normal 1.5

SPACING (DESIGN.md §17)  --space-1 4px … --space-10 48px
RADIUS (DESIGN.md §20)   --radius-sm 4px  --radius-md 4px  --radius-lg 6px  --radius-xl 8px  --radius-full 999px
                         (squared-off: badges are 4px rects; --radius-full only for spinner + timeline dot)
KPI/TINTED CARDS   1px border in a darker shade of the card's own colour
                   (color-mix 35% of the semantic colour)
DASHBOARD 4 CARDS  tone sequence: danger → success → info → violet
                   (red · green · blue · violet)
                   bg = *-soft tint, value + icon = tone colour, label/meta = muted
CONTROLS (DESIGN.md §41/§43)  --control-h-sm 32px  --control-h-md 38px  --control-h-lg 44px
LAYOUT (DESIGN.md §24/§25)
  --header-h 60px  --sidebar-w-collapsed 64px  --sidebar-w-expanded 240px
  --footer-h 40px  --content-max 1600px
MOTION (DESIGN.md §71)   --motion-fast 130ms  --motion-normal 200ms  --motion-slow 280ms  --ease-standard
LAYERS (DESIGN.md §85)
  --z-base 0  --z-sticky 40  --z-dropdown 50  --z-drawer 60  --z-modal 70  --z-toast 80
BREAKPOINTS (DESIGN.md §32)  sm 640px  md 768px  lg 1024px  xl 1280px  (media queries use these)
```

**Light/dark:** both themes redefine **only** the colour block; every other token is theme-independent
(DESIGN.md §11 — one component + theme tokens, no duplicated components). Sidebar uses neutral surfaces
in dark mode (DESIGN.md §28), never a saturated brand block.

---

## 4. Component Inventory (must exist and be reused — DESIGN.md §86)

```text
ui/     Button, IconButton, Input, Textarea, Select, Checkbox, Radio, Switch, FormField,
        Card, Badge, StatusBadge, Alert, Toast, Modal, ConfirmDialog, Drawer, Dropdown,
        Tabs, Tooltip, Breadcrumb, PageHeader, Toolbar, EmptyState, ErrorState, Skeleton,
        Spinner, Icon, Avatar, Pagination, SegmentedControl
data/   DataTable, DataListMobile, FilterBar, KpiCard, PaginationBar,
        BarChart, LineChart, DonutChart, ProgressBar, AmountText, LookupSelect
layout/ AppShell, AppHeader, AppSidebar, AppFooter, MobileNav, PageContainer
```

Every page composes these. A page must not define its own button/input/table/badge styling
(DESIGN.md §92). Variant names follow DESIGN.md §87–§88: `variant | size | disabled | loading | className`.

---

## 5. Navigation Model (DESIGN.md §25–§29)

- Desktop: compact icon-first sidebar (64px), expands to 240px on pointer-over/click-lock; icons stay
  visible when collapsed; tooltips for collapsed items; active state obvious.
- Mobile (`< 1024px`): the **same** nav data renders into a `Drawer` with full labels, overlay,
  close action and Escape support. One nav source: `state/navStore` built from the user's permission set.
- Navigation items are **derived from permissions** (`../frontend/src/lib/permission.ts`), never from a hardcoded
  module/role list. A user sees only modules their permissions allow.

---

## 6. Page Patterns (DESIGN.md §36–§38, §77–§79)

```text
List page      PageHeader (title, context, actions) → FilterBar → DataTable → PaginationBar
Detail page    PageHeader → Summary card → Details → History/timeline → Related records
Create/Edit    PageHeader → form sections (1–3 col desktop, 1 col mobile) → sticky action bar
Dashboard      PageHeader → KPI row → primary data → recent activity tables (compact, not big cards)
Auth screen    brand mark → auth surface → form → primary action → supporting links
```

States every page must implement: **loading** (skeleton), **empty**, **error** (with retry),
**permission-denied** (explains and offers safe navigation back).

---

## 7. Accessibility & Responsive Commitments (DESIGN.md §81–§84)

- All interactive elements keyboard reachable; visible focus ring from `--color-focus-ring`.
- Every input has a real `<label>`; icon-only controls have `aria-label`.
- Status is never conveyed by colour alone — `StatusBadge` always shows text.
- Tables: real `<table>` semantics inside an `overflow-x-auto` container; no sub-12px text to force fit.
- Touch targets ≥ 40px on mobile controls (`--control-h-lg`).
- No horizontal page scroll at 360px; verified by responsive checks in Step 08.
- Verified at 360 / 768 / 1280 / 1536 px in light, dark and system themes.

---

## 8. Status Colour Mapping (DESIGN.md §55 — identical meaning everywhere)

| Semantic state | Token | Used by |
|---|---|---|
| New / informational | `--color-info` | new complaint, new notice, info alerts |
| Pending / attention | `--color-warning` | PENDING, PARTIAL, IN_PROGRESS, GRACE |
| Success / completed | `--color-success` | PAID, RESOLVED, PRESENT, PUBLISHED, ACTIVE |
| Error / overdue / rejected | `--color-danger` | OVERDUE, CANCELLED, ABSENT, EXPIRED, LOCKED |
| Neutral | `--color-text-muted` | CLOSED, ARCHIVED, INACTIVE, DRAFT |

`StatusBadge` resolves a status key → tone through a single map in
`../frontend/src/lib/statusTone.ts`, driven by the status key returned by the API (never by module).