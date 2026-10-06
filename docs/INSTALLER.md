# Installer & Packaging Guide

How to turn the built application into redistributable packages for end users
(SRS §45: Windows `.exe`, macOS application package, Linux desktop package).

---

## 1. Windows — portable executable (always available)

```powershell
wails build
```

Output: `build\bin\FixedDepositManagement.exe`

- Single self-contained file; the React frontend is embedded in the binary.
- Requires the Microsoft **WebView2 Runtime** (preinstalled on Windows 10/11).
- The app creates its database on first run under
  `%AppData%\FixedDepositManagement\fd_management.db`.
- Distribution: copy the `.exe` (optionally zipped). This is the zero-dependency path.

## 2. Windows — NSIS installer

The repository already contains the Wails NSIS project under `build/windows/installer/`
(`project.nsi`, `wails_tools.nsh`). The installer:

- installs the app to `Program Files`,
- writes Start Menu / Desktop shortcuts,
- bundles the WebView2 bootstrapper when the runtime is missing,
- creates an uninstaller.

### Prerequisites

1. **NSIS ≥ 3** — https://nsis.sourceforge.io/Download — add `makensis` to `PATH`.
2. (Optional) **UPX** for compressing the exe — https://upx.github.io/.
3. (Optional) Code-signing certificates for `signtool`.

### Build

```powershell
wails build -nsis
```

Output: `build\bin\*-installer.exe` (name from `build/windows/installer/project.nsi`
`OutFile`, using the product name in `build/windows/info.json`).

### Installer metadata

- Product branding/version: `build/windows/info.json` (used to generate the NSIS defines).
- To enable the license page, un-comment `MUI_PAGE_LICENSE` in `project.nsi` and provide
  `resources\eula.txt`.
- Signing: un-comment the `Sign` macros in `project.nsi` and set your certificate.

> Note: NSIS is **not installed** on the current development machine, so the installer
> target has not been produced here — the portable `.exe` above is the verified artifact.
> Install NSIS and re-run `wails build -nsis` to generate it.

## 3. macOS

```powershell
# On a macOS host with Xcode command line tools installed:
wails build -platform darwin/universal
```

Output: `build/bin/<AppName>.app` (unsigned). For distribution outside the App Store,
sign (`codesign`) and notarise the bundle.

## 4. Linux

```powershell
# On a Linux host:
wails build -platform linux/amd64
```

Output: `build/bin/<AppName>`. Package as you prefer:

- a `.deb` wrapping the binary + a `.desktop` entry, or
- an AppImage, or
- a tarball.

The WebView2 equivalent on Linux is WebKitGTK (`webkit2gtk` package).

## 5. Distribution checklist

- [ ] `go vet ./...` and `go test ./...` pass.
- [ ] `cd frontend; npm run test` and `npm run build` pass.
- [ ] `wails build` succeeds; app launches and creates its database (smoke test).
- [ ] Fresh-profile test: run with an empty `FD_ATS_DATA_DIR`, create an FD, export a report.
- [ ] Version bumped in `internal/brand/brand.go` (`Version`) and `build/windows/info.json`.
- [ ] Installer (if produced) tested on a clean VM: install → run → uninstall.
- [ ] Backup note included: copying `fd_management.db` is a complete backup.

## 6. Backup & data portability (support notes)

- All data is in one SQLite file; no server, no account.
- Restore = replace `fd_management.db` while the app is closed.
- `FD_ATS_DATA_DIR` can point to a USB/network folder for a portable installation.
