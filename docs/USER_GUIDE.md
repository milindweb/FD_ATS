# User Guide — Fixed Deposit Management

Offline desktop software for creating, tracking, renewing and closing Fixed Deposits,
with Excel reporting. This guide follows the actual application screens.

> **Product:** Fixed Deposit Management · **Developer:** Aarti Tech Services
> Website: https://aartitechservices.pages.dev/ · Email: aartitechservices@gmail.com · Phone: +91 9869787575

---

## 1. Starting the application

Double-click `FixedDepositManagement.exe`. On the first run the app creates its local
database (a single file in your user configuration folder) — no internet required.

If the app cannot start, it shows a polite message with a **Try Again** button instead
of crashing; use it after fixing the reported problem.

## 2. Dashboard (home screen)

- **KPI cards** — Active FD, Total Deposit, Maturing Soon (30 days), Maturing Today
  (with 7-day and 90-day counts).
- **Search box** — type at least part of an FD number, customer name or member number;
  results update as you type.
- **Filter buttons** — All / Active / Closed / Maturing.
- **FD table** — FD Number, Customer/Member, Deposit Amount, Interest Rate, Start Date,
  Maturity Date, Maturity Amount, Status. **Click any row** to open its details.
- **Pagination** — Previous/Next below the table.
- **Upcoming Maturities** — FDs maturing in the next 90 days with days remaining.
- **Create New FD** (top right) — opens the new deposit form.

On narrow screens the table turns into labelled cards automatically.

## 3. Creating a new FD

1. **New FD** (or *Create New FD* on the dashboard).
2. Fill in:
   - Customer / Member Name *(required)*
   - Customer / Member Number *(optional)*
   - Deposit Amount (₹) *(required, whole rupees)*
   - Start Date *(required)*
   - Tenure in days *(required — use the preset buttons: 6 months / 1 / 2 / 3 years)*
3. Watch the **Calculation Preview** — interest rate (chosen from the slab that matches
   the tenure), interest and maturity amount update as you type.
4. Press **Save Fixed Deposit**. The FD details screen opens with the new
   **FD number** (`FD-YY-NNN`, where YY comes from the start date).

Calculation: `Interest = Principal × Rate × Days ÷ 365 ÷ 100`, rounded to the nearest
rupee. `Maturity Amount = Principal + Interest`.

## 4. FD details

Shows the complete deposit, status, closure information (when closed), links to related
renewal FDs, and the **History** table (every open/renew/close event with amounts and
remarks).

Actions for active FDs: **Renew** and **Close FD**.

## 5. Renewing an FD

1. Open an active FD → **Renew**.
2. Choose the mode:
   - **Principal Only** — only the deposit amount is renewed; interest is paid out.
   - **Principal + Interest** — earned interest is added to the new deposit.
3. Set the **New Start Date** and **New Tenure**, optionally a remark.
4. Check the **Renewal Preview** (new rate, maturity date, new maturity amount).
5. **Confirm Renewal**.

The old FD is closed as `RENEWED`, a **new FD number** is created, and both records link
to each other. Nothing is deleted — history stays intact.

## 6. Closing an FD

1. Open an active FD → **Close FD**.
2. **Closure Date** (defaults to today; cannot be before the start date) and an optional
   remark.
3. The **Payable Preview** shows days held, the rate applied (from the slab matching the
   actual days held), interest and the **Amount Payable**:
   - **Matured closure** (on/after maturity): payable equals the original maturity amount —
     no extra interest is added after maturity.
   - **Premature closure**: interest for the actual days held at the applicable slab rate.
4. Press **Close FD** → confirm in the dialog. The payable amount cannot be changed by
   the confirmation — it is exactly what the preview showed.

Closed FDs remain searchable under the **Closed** filter and in reports.

## 7. Reports

1. Open **Reports**.
2. Choose a report: **FD Register**, **Maturity Report**, **Active FD Report** or
   **Closed FD Report**.
3. For Register/Maturity: set an optional **From/To** date range.
4. **Export …** → pick a file location in the save dialog → success message shows the
   row count and saved path.

Excel output: header row with filters enabled; columns include FD number, customer,
principal, rate, dates, interest, maturity amount, status and closure details.

## 8. Settings — interest rate slabs

**Settings** shows the rate slabs used for every calculation (default: 1–364 days 4%,
365–729 8%, 730–1094 8.5%, 1095–1460 9.5%). Edit values, add a slab (a *Max Days* of 0
means open-ended), then **Save Changes**. Validation explains any invalid range.

Changes apply to future calculations; existing FDs keep their recorded rates.

## 9. Theme and navigation

- Header theme button cycles **Light → Dark → System**.
- Desktop: navigation links in the header. Small screens: the **☰** menu opens a drawer.
- Footer: privacy policy and About on every screen.

## 10. Messages you may see (and what to do)

| Message (typical wording) | Meaning / action |
| --- | --- |
| Customer name is required | Enter the name before saving. |
| Deposit amount must be greater than zero | Principal must be a positive whole amount. |
| Maturity date must be after the start date | Fix the tenure or start date. |
| Closure date cannot be before the start date | Pick a valid closure date. |
| FD number already exists | Retry — numbering is allocated automatically with retries. |
| Only active FDs can be renewed/closed | The FD is already closed — open it to see details. |
| Interest rate slabs overlap / are invalid | Fix slab day ranges in Settings. |
| Could not open the local database | Close other programs using the file, then **Try Again**. |

Errors never lose your data silently: forms keep your input so you can correct and retry.

## 11. Backing up your data

All data lives in one file (`fd_management.db` in your user configuration folder, e.g.
`%AppData%\FixedDepositManagement\`). To back up: close the app and copy that file.
To restore: close the app and put the copy back. `FD_ATS_DATA_DIR` can point the app at
a different folder (e.g. a USB drive) for portable use.

---

**© 2026 Aarti Tech Services. All rights reserved.**
