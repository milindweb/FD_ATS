# Fixed Deposit Management

Desktop software for managing Fixed Deposits: creation with live interest calculation,
renewals, closures (matured and premature), maturity tracking and Excel report export.
Works fully offline. Built with **Wails + Go + React + TypeScript + SQLite** per `docs/SRS.md`
and `docs/design.md`.

> Developer: **Aarti Tech Services** — https://aartitechservices.pages.dev/

---

## Features

| Area | What it does |
| --- | --- |
| Dashboard | KPIs (active FDs, total deposit, maturing today/soon), searchable & filterable FD list, pagination, upcoming maturities (next 90 days) |
| New FD | Live calculation preview while typing (SRS §10), client-side validation, tenure presets |
| FD Details | Full deposit info, status, closure details, linked renewal FDs, chronological history |
| Renew | Principal-only or principal + interest; old FD closed as `RENEWED`, new FD number issued (SRS §21–§22) |
| Close | Payable preview before confirmation, premature (actual days) or matured closure (SRS §23–§26) |
| Reports | FD Register, Maturity, Active and Closed reports as `.xlsx` with autofilters (SRS §29) |
| Settings | Editable interest rate slabs (SRS §13, §30) |
| About / Privacy | Branding (SRS §31, §49) and privacy policy (SRS §33) |

### Business rules (SRS §47)

```
Interest          = Principal × Rate × Days / 365 / 100   (rounded to nearest rupee)
Maturity Amount   = Principal + Interest
Rate slabs        1–364 → 4.00% · 365–729 → 8.00% · 730–1094 → 8.50% · 1095–1460 → 9.50%
Premature closure rate is picked from the ACTUAL days held
Matured closure   payable ≤ original maturity amount (no extra interest)
Renewal           creates a NEW FD number (principal only, or principal + interest)
FD numbers        FD-YY-NNN where YY comes from the start date
```

---

## Quick start (development)

Prerequisites: Go 1.24+, Node.js 18+, [Wails CLI v2](https://wails.io) (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).

```powershell
# Backend tests
go vet ./...
go test ./...

# Frontend tests + type check
cd frontend
npm install
npm run test
npm run build

# Run the app in dev mode
cd ..
wails dev
```

## Production build

```powershell
wails build
```

Output: `build\bin\FixedDepositManagement.exe` — a single self-contained executable.
Installer packaging is described in [docs/INSTALLER.md](docs/INSTALLER.md).

---

## Documentation

| Document | Contents |
| --- | --- |
| [docs/BUILD.md](docs/BUILD.md) | Environment setup, dev workflow, test commands, project structure |
| [docs/INSTALLER.md](docs/INSTALLER.md) | Windows installer (NSIS), cross-platform packaging, distribution checklist |
| [docs/USER_GUIDE.md](docs/USER_GUIDE.md) | End-user manual: every screen, step-by-step flows, error messages |
| [docs/SRS.md](docs/SRS.md) | Software requirements specification (source of truth for behaviour) |
| [docs/design.md](docs/design.md) | Design system: tokens, components, CSS rules (source of truth for UI) |

---

## Architecture

```
main.go / app.go            Wails binding surface (thin: validates readiness, delegates)
internal/domain             Entities, date helpers, centralized error messages
internal/calc               Interest/maturity/closure engine (pure functions, fully tested)
internal/repo               SQLite persistence (modernc.org/sqlite — no CGO)
internal/service            Business flows: create, list, renew, close, dashboard, reports
internal/report             Excel (.xlsx) export via excelize
internal/api                DTOs shared by service and bindings (JSON = camelCase)
frontend/src/lib            api.ts (single backend import surface), formatting, icons, theme
frontend/src/components     Reusable UI (Button, FormField, DataTable, Modal, …), layout
frontend/src/pages          One file per screen
frontend/src/styles         tokens → base → components → utilities → print (design.md §8)
```

Data lives in the OS user-config folder (`%AppData%\FixedDepositManagement\fd_management.db`);
override with the `FD_ATS_DATA_DIR` environment variable.

## Tests

```powershell
go test ./...                          # domain, calculation, service flows, Excel export
cd frontend; npm run test              # 28 vitest tests: format, tones, components, pages
```

`internal/service` includes `TestRecommendedUserFlow`, which walks the complete
SRS §35 flow: create → verify calculation → renew → close → report.
