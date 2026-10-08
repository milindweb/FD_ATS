# User Guide — Fixed Deposit Management

Offline desktop software for managing members and Fixed Deposits, with Excel reporting.
This guide follows the actual application screens.

> **Product:** Fixed Deposit Management · **Developer:** Aarti Tech Services
> Website: https://aartitechservices.pages.dev/ · Email: aartitechservices@gmail.com · Phone: +91 9869787575

---

## 1. Starting the application

Double-click `FixedDepositManagement.exe`. On the first run the app creates its local
database (a single file in your user configuration folder) — no internet required.

If the app cannot start, it shows a polite message with a **Try Again** button instead
of crashing; use it after fixing the reported problem.

## 2. Dashboard (home screen)

Layout, top to bottom:

- **Quick search** — type a term and press **Enter**; the results open on the
  **FD Master** page with the query pre-filled.
- **Quick actions** — **+ New Member**, **+ Create New FD**, and *View full FD list*.
- **KPI cards** — Active FDs, Total Deposit, Total Interest, Financial Year deposits.
- **Maturity chart** — active FDs bucketed by maturity month.
- **Upcoming Maturities** — FDs maturing in the next 90 days with days remaining
  (member name, principal, dates, amount).

## 3. FD Master (the full FD list)

The **FD Master** navigation item opens the complete FD list:

- **Search** — FD Number, FD Form No., Member Name or GEN No.; results update as you type.
- **Filter buttons** — All / Active / Closed / Maturing.
- **Table** — FD Number, Member, Deposit Amount, Interest Rate, Start Date,
  Maturity Date, Maturity Amount, Status. **Click any row** to open its details.
- **Pagination** — Previous/Next below the table.
- **+ Create New FD** (top right).

On narrow screens the table turns into labelled cards automatically.

## 4. Members (Member page)

The **Member** navigation item is the home of everything member-related:

- **Search** — GEN No., Name, Token No., Designation, PAN, Aadhaar or Mobile No.
- **Member list** — GEN No., Name, Designation, Token No., Mobile, Active FDs and
  Total Active FD Amount. **View** opens the member profile; **Edit** opens the form.
  Click any row to open the profile.
- **+ Add Member** — the member form in SRS field groups (identification, personal,
  address, employment, nominee, government ID, banking, remarks). **GEN No.** and
  **Name** are required; GEN No. must be unique.
- **Bulk Excel Upload** — pick an `.xlsx` file whose first sheet has at least the
  columns `GEN No.` and `Name`. An **Import Preview** shows Total / Valid /
  Duplicate GEN No. / Missing Mandatory / Invalid / Existing in System counts plus a
  row-by-row report. Nothing is saved until you press **Import**; existing GEN Nos.
  are skipped and never overwritten.
- **Member profile** — the complete member record in grouped sections, two summary
  cards (Total Active FDs, Total Active FD Amount) and the member's FD table.
  Each FD opens its details screen directly.

A member can hold any number of FDs (1 member → many FDs).

## 5. Creating a new FD

Every FD must belong to a member:

1. Open **+ Create New FD** (Dashboard quick action or FD Master action).
2. **Linked Member** *(required)* — search by GEN No., Name, Token No., Designation,
   PAN, Aadhaar or Mobile and pick the member. No member yet? Create one first
   (**+ New Member**).
3. **FD Form No.** *(optional)* — a manual, free-text reference; it is not unique.
4. Fill in **Deposit Amount** (whole rupees), **Start Date** and **Tenure in days**
   (presets: 6 months / 1 / 2 / 3 years).
5. Watch the **Calculation Preview** — interest rate (slab matching the tenure),
   interest and maturity amount update as you type.
6. Press **Save Fixed Deposit**. The FD details screen opens with the new
   **FD number** (`FD-1`, `FD-2`, … — one running sequence for the whole system).

The member's name and GEN No. are copied onto the FD automatically.

Calculation: `Interest = Principal × Rate × Days ÷ 365 ÷ 100`, rounded to the nearest
rupee. `Maturity Amount = Principal + Interest`.

## 6. FD details

Shows the complete deposit — FD Number, Member Name, GEN No., FD Form No., amounts,
dates and status — plus closure information (when closed), links to related renewal
FDs, and the **History** table (every open/edit/renew/close event with amounts and
remarks).

Actions for active FDs: **Edit**, **Renew** and **Close FD**.
**Back to FD Master** returns to the list.

Actions for closed FDs: **Reopen FD** (matured/premature closures) and
**Reverse Renewal** (FDs closed by renewal, or the renewed FD itself while it
is still active). Both require a reason that is recorded in the history.

## 7. Renewing an FD

1. Open an active FD → **Renew**. The Renew button only becomes available on
   the FD's maturity date — before that it stays disabled with a note telling
   you when renewal opens.
2. Choose the mode:
   - **Principal Only** — only the deposit amount is renewed; interest is paid out.
   - **Principal + Interest** — earned interest is added to the new deposit.
3. Set the **New Start Date** and **New Tenure**, optionally a remark. The start
   date defaults to the FD's maturity date and cannot be set before it —
   renewing early is blocked, but back-dated renewals on/after maturity are
   allowed.
4. Check the **Renewal Preview** (new rate, maturity date, new maturity amount).
5. **Confirm Renewal**.

The old FD is closed as `RENEWED`, a **new FD number** is created (the running
sequence continues), and both records link to each other. The renewed FD keeps
the member link and starts with an empty FD Form No. Nothing is deleted —
history stays intact.

## 8. Closing an FD

1. Open an active FD → **Close FD**.
2. **Closure Date** (defaults to today; cannot be before the start date) and a
   remark — required for premature closures, optional when closing at/after
   maturity.
3. The **Payable Preview** shows days held, the rate applied (from the slab matching the
   actual days held), interest and the **Amount Payable**:
   - **Matured closure** (on/after maturity): payable equals the original maturity amount —
     no extra interest is added after maturity.
   - **Premature closure**: interest for the actual days held at the applicable slab rate.
4. Press **Close FD** → confirm in the dialog. The payable amount cannot be changed by
   the confirmation — it is exactly what the preview showed.

Closed FDs remain searchable under the **Closed** filter and in reports.

## 9. Correcting a mistake

- **Edit** (on an active FD): change the linked member, FD Form No., deposit amount,
  start date or tenure. The rate, interest and maturity amount are recalculated
  and an `EDIT` event is added to the history. FD numbers and closed FDs are
  never editable. Editing is also how an unassigned FD gets linked to a member.
- **Edit Member** (on the Member page): update any member field; the member's
  name and GEN No. copies on all of that member's FDs refresh automatically.
- **Reopen FD** (on a matured/premature closed FD): enter a reason; the FD
  returns to ACTIVE and the closure stays visible in the history as part of
  the audit trail, followed by a `REOPEN` event.
- **Reverse Renewal** (on the old FD closed as RENEWED, or on the renewed FD
  while it is still active): enter a reason; the renewed FD is withdrawn and
  the previous FD becomes active again. Blocked once the renewed FD has been
  closed or renewed itself.

Nothing is silently overwritten — every reversal records its reason.

## 10. Reports

1. Open **Reports**.
2. Choose a report: **FD Register**, **Maturity Report**, **Active FD Report**,
   **Closed FD Report** or **Member-wise FD Summary**.
3. For Register/Maturity: set an optional **From/To** date range.
4. **Export …** → pick a file location in the save dialog → success message shows the
   row count and saved path.

Excel output: header row with filters enabled. Register-style reports include FD
number, member name, GEN No., principal, rate, dates, interest, maturity amount,
status and closure details. The **Member-wise FD Summary** lists GEN No., Member
Name, Designation, Active FD Count and Total Active FD Amount per member, plus one
**Unassigned** row for active FDs that are not linked to a member yet.

## 11. Settings

Settings has three tabs:

- **Slab Rates** — the rate slabs used by every calculation (default: 1–364 days 4%,
  365–729 8%, 730–1094 8.5%, 1095–1460 9.5%). Edit values row by row, add a slab
  (a *Max Days* of 0 means open-ended), then **Save Changes**. Validation explains
  any invalid or overlapping range. Changes apply to future calculations; existing
  FDs keep their recorded rates.
- **Login Credentials** — two cards, each requiring your **current password**:
  **Change Username** and **Change Password** (minimum 4 characters, confirmed
  twice). Every change issues a **fresh recovery code** and invalidates the old one —
  store it safely; it is what lets you reset the login if you forget the password.
- **Backup & Restore** — export the database to a chosen file, restore from a
  backup (with confirmation), or load the built-in **Sample Data** when the
  database is empty.

## 12. Theme and navigation

- Header theme button cycles **Light → Dark → System**.
- Desktop navigation: **Dashboard · FD Master · Member · Reports · Settings**.
  Small screens: the **☰** menu opens a drawer.
- About and Privacy Policy are linked from the footer on every screen.

## 13. Messages you may see (and what to do)

| Message (typical wording) | Meaning / action |
| --- | --- |
| A linked member is required | Search and select a member before saving the FD. |
| GEN No. is required / Member name is required | Fill in the mandatory member fields. |
| This GEN No. already exists | GEN Nos. must be unique — use Edit on the existing member instead. |
| Deposit amount must be greater than zero | Principal must be a positive whole amount. |
| Maturity date must be after the start date | Fix the tenure or start date. |
| Closure date cannot be before the start date | Pick a valid closure date. |
| FD number already exists | Retry — numbering is allocated automatically with retries. |
| Only active FDs can be renewed/closed | The FD is already closed — open it to see details. |
| This FD has not matured yet | Renewal opens on the maturity date — until then the Renew button is disabled. |
| Closed FDs cannot be edited | Editing is only for active FDs — reopen it first if a correction is needed. |
| Interest rate slabs overlap / are invalid | Fix slab day ranges in Settings. |
| Current password is incorrect | Re-enter your existing password. |
| Could not open the local database | Close other programs using the file, then **Try Again**. |

Errors never lose your data silently: forms keep your input so you can correct and retry.

## 14. Backing up your data

All data lives in one file (`fd_management.db` in your user configuration folder, e.g.
`%AppData%\FixedDepositManagement\`). Use **Settings → Backup & Restore** to export or
restore a copy, or close the app and copy the file by hand. `FD_ATS_DATA_DIR` can point
the app at a different folder (e.g. a USB drive) for portable use.

---

**© 2026 Aarti Tech Services. All rights reserved.**
