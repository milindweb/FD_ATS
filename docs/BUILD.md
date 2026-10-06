# Build Guide

How to set up the environment, run the app in development, run the test suites and
produce release builds of **Fixed Deposit Management**.

---

## 1. Prerequisites

| Tool | Version | Notes |
| --- | --- | --- |
| Go | 1.24+ | https://go.dev/dl/ — add `go` to `PATH` |
| Wails CLI | v2.x | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Node.js | 18+ | for the React frontend (`npm install`) |
| WebView2 Runtime | evergreen | preinstalled on Windows 10/11; otherwise from Microsoft |

Verify:

```powershell
go version
wails doctor      # should report "system ready"
node --version
```

Optional (installer packaging only, see `docs/INSTALLER.md`): NSIS ≥ 3, UPX.

## 2. One-time project setup

```powershell
cd <project root>
cd frontend
npm install       # installs React, react-router-dom, vitest, testing-library, …
```

Go dependencies download automatically on the first `go build`.

## 3. Development workflow

```powershell
wails dev
```

- Serves the frontend with hot reload and rebuilds the Go backend on change.
- Frontend-only work: `cd frontend; npm run dev` (backend calls need the Wails runtime,
  so full UI testing happens through `wails dev`).

### After changing `app.go`

Wails generates TypeScript bindings from the Go API. Regenerate them:

```powershell
wails generate module
```

This refreshes `frontend/wailsjs/go/**`. Frontend code must only import bindings via
`frontend/src/lib/api.ts`, so API changes are normally confined to that one file.

## 4. Tests

```powershell
# Backend (calculation rules, service flows, Excel export)
go vet ./...
go test ./...
go test ./internal/service -run TestRecommendedUserFlow -v   # full SRS §35 walk-through

# Frontend (formatting, status tones, components, pages)
cd frontend
npm run test          # vitest, single run
npm run build         # tsc type check + production bundle
```

All of the above must pass before a `wails build`.

## 5. Production build

```powershell
# From the project root
wails build
```

Produces `build\bin\FixedDepositManagement.exe` (single-file, frontend embedded,
WebView2 loader included). Useful flags:

```powershell
wails build -obfuscated     # obfuscate Go symbols
wails build -nsis           # additionally build the NSIS installer (needs NSIS)
wails build -skipbindings   # skip regenerating bindings (CI speed-up)
```

## 6. Project structure

```
app.go                     Wails bindings (thin facade over internal/service)
main.go                    Wails application entry point
internal/
  domain/                  Models, dates, centralized error messages (SRS §40)
  calc/                    Pure calculation engine (SRS §16, §24–§27)
  repo/                    SQLite access layer + schema migrations/seed data
  service/                 Business flows + integration tests
  report/                  Excel export builder (SRS §29)
  api/                     Shared DTOs (JSON tags = camelCase)
  brand/                   Branding constants (SRS §49)
frontend/
  src/lib/                 api.ts, format.ts, statusTone.ts, icons.tsx, theme.ts, hooks.ts
  src/components/ui/       Button, FormField, Card, Badge, Alert, Modal, States, …
  src/components/data/     DataTable, FilterBar, KpiCard, PaginationBar
  src/components/layout/   AppShell, AppHeader, AppFooter, MobileNav
  src/pages/               One component per screen (+ vitest tests)
  src/styles/              tokens.css → base.css → components.css → utilities.css → print.css
build/                     Wails platform assets and installer scripts
docs/                      Build, installer and user documentation
```

## 7. Data location & environment variables

| Variable | Purpose |
| --- | --- |
| `FD_ATS_DATA_DIR` | Overrides the folder holding `fd_management.db` (default: OS config dir + `FixedDepositManagement`) |

Useful for tests and portable installs; pointing it at a copy of the database is also a
simple manual backup strategy.
