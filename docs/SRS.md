# Fixed Deposit Management Software
## Software Requirements Specification (SRS)

as per srs.md and design.md write code in phase wise step by step Making Modular fixed deposit software sor me,Install superpowers gitnexus, , req Packaging documets, and verify actual used flow,test it and fix bug, dont hardcode code make reusable, frontend us css req.

**Version:** 1.0  
**Year:** 2026  
**Status:** Initial Specification  
**Product Type:** Standalone Desktop Application  
**Primary Platform:** Windows  
**Cross-Platform Target:** Windows, macOS, Linux  

**Designed by:** Aarti Tech Services  
**Website:** https://aartitechservices.pages.dev/  
**Email:** aartitechservices@gmail.com  
**Phone:** +91 9869787575  

---

# 1. Introduction

## 1.1 Purpose

The **Fixed Deposit Management Software** is a simple, user-friendly desktop application designed to manage Fixed Deposits (FDs).

The software will help users:

- Create Fixed Deposits
- Calculate interest
- Calculate maturity dates
- Calculate maturity amounts
- Search and manage FD records
- Renew Fixed Deposits
- Close Fixed Deposits
- Handle premature closure
- Maintain FD history
- Track upcoming maturities
- Export FD reports to Excel

The application should be easy to operate for users who may not have technical knowledge.

---

# 2. Product Objectives

The main objectives are:

1. Keep FD management simple.
2. Reduce manual calculation errors.
3. Automatically calculate interest and maturity.
4. Maintain complete FD history.
5. Make renewal and closure easy.
6. Provide quick access to upcoming maturities.
7. Generate useful Excel reports.
8. Work as a standalone desktop application.
9. Provide a responsive interface suitable for desktop and small screen sizes.
10. Keep the software lightweight and easy to maintain.

---

# 3. Scope

## 3.1 Included

The first version shall include:

- Dashboard overview (KPIs, chart, upcoming maturities)
- FD Master (FD search and listing)
- FD creation
- FD calculation
- FD details
- FD history
- FD renewal
- FD closure
- Premature closure
- Maturity tracking
- Excel reports
- Interest-rate configuration
- Member master (add / view / edit / search)
- Bulk member import from Excel
- Member-wise FD summary
- FD-to-member linkage
- About page
- Privacy Policy
- Responsive Header and Footer

## 3.2 Not Included in Version 1

The following are intentionally outside the initial scope:

- Online banking integration
- Payment gateway
- SMS service
- WhatsApp integration
- Email automation
- Cloud synchronization
- Multi-organization management
- Complex accounting
- TDS management
- Loan management
- Customer portal
- Mobile Android/iOS application
- Online user registration

These may be considered in future versions if required.

---

# 4. Application Type

The application shall be a standalone desktop application.

It should operate primarily using local data.

The application should not require:

- XAMPP
- Apache
- MySQL
- PostgreSQL
- PHP installation
- Node.js installation
- Internet connection for normal FD operations

The software should be installable and usable like a normal desktop application.

---

# 5. Supported Platforms

The application should be designed for cross-platform desktop support:

- Windows
- macOS
- Linux

The initial release may prioritize Windows.

The same application architecture and user interface should be designed so that future builds can support macOS and Linux without redesigning the complete application.

---

# 6. User Interface Requirements

## 6.1 General Design

The interface shall be:

- Simple
- Clean
- Professional
- Responsive
- Mobile-first
- Easy to understand
- Touch-friendly
- Keyboard-friendly
- Suitable for small and large screens

The application should not feel like a complex accounting system.

---

# 7. Responsive Design

The interface must work properly at different screen sizes.

Minimum design target:

- 320px width
- 375px width
- 480px width
- Tablet sizes
- Laptop screens
- Desktop monitors
- Large desktop screens

## 7.1 Desktop

Desktop screens may use:

- Sidebar or horizontal navigation
- Multi-column layouts
- Tables
- Dashboard cards

## 7.2 Small Screens

On smaller screens:

- Tables should transform into cards where appropriate.
- Forms should become single-column.
- Navigation should become a mobile menu.
- Buttons should remain touch-friendly.
- Text should remain readable.
- Horizontal scrolling should be avoided wherever practical.

---

# 8. Application Layout

Every page shall use a common:

- Header
- Main content area
- Footer

## 8.1 Header

The header shall contain:

- Application logo
- Application name: **Fixed Deposit Management**
- Main navigation menu

Main menu:

- Dashboard
- FD Master
- Member
- Reports
- Settings

The About page remains reachable from the footer.

The header should remain consistent across the application.

## 8.2 Small Screen Header

On small screens, the navigation shall collapse into a menu button.

Example:

```text
┌─────────────────────────────┐
│ ☰  Fixed Deposit Management │
└─────────────────────────────┘
```

## 8.3 Footer

The footer shall contain:

- Application name
- Designed by Aarti Tech Services
- Privacy Policy
- About
- Copyright information

Example:

```text
Fixed Deposit Management

Designed by Aarti Tech Services
© 2026 All rights reserved.
```

---

# 9. Main Screens

The application shall have the following functional screens.

## 9.1 Dashboard

The Dashboard shall be the application's overview screen. The FD list itself lives on the FD Master page (§9.2).

It shall provide:

- Total active FDs
- Total deposit amount
- Upcoming maturities
- Maturity chart
- Search (quick search — results open on the FD Master page with the query pre-filled)
- Quick action to create a new member
- Quick action to create a new FD
- A link to open the full FD list (FD Master)

Example:

```text
Active FD       125
Total Deposit   ₹1,25,00,000
Maturing Soon   8
```

Layout order (top to bottom):

1. Page header
2. Quick search box
3. Quick action buttons
4. KPI cards, maturity chart and upcoming maturities

Quick actions (shown side by side, directly under the search box):

**+ New Member** · **+ Create New FD**

The **+ New Member** action opens the Add Member form.

The **+ Create New FD** process first lets the user search for and select an existing member (see §10).

---

## 9.2 FD Master

A dedicated **FD Master** page shall contain the full FD list.

The user should be able to search using:

- FD Number
- FD Form No.
- Member Name
- GEN No.

Filters:

- All
- Active
- Closed
- Maturing

The FD list should show:

- FD Number
- Member
- Deposit Amount
- Interest Rate
- Start Date
- Maturity Date
- Maturity Amount
- Status

The page shall also provide the **+ Create New FD** quick action.

Clicking an FD shall open its FD Details.

---

## 9.3 Member Page

A dedicated **Member** page shall contain all member-related functionality (see §50–§53):

- Add Member
- Bulk Excel Upload
- Member Search
- Member List
- View Member Profile
- Edit / Modify Member
- Member-wise FD summary
- View all FDs linked to the member

A single member can have multiple FDs (**1 Member → Many FDs**). The Member page is the central place for managing all member information.

---

# 10. Create FD

The Create FD screen shall allow the user to create a new Fixed Deposit.

Required information:

- Linked Member (selected from the Member master — mandatory)
- FD Form No. (optional, manual, not unique)
- Deposit Amount
- Start Date
- Tenure

### Member selection

- The user searches for an existing member by **GEN No., Name, Token No., Designation, PAN, Aadhaar or Mobile No.**
- The required member is selected and the FD is automatically associated with that member.
- No second member record is ever created for an existing GEN No.
- If no member exists yet, the user creates one first (Dashboard → **+ New Member** or Member page → Add Member).
- The member's name and GEN No. are recorded on the FD as a readable copy of the link.

The system shall automatically determine:

- Interest Rate
- Interest Amount
- Maturity Date
- Maturity Amount

The user should be able to preview the calculation before saving.

Example:

```text
Principal Amount      ₹1,00,000
Start Date             01/10/2026
Tenure                 365 Days
Interest Rate          8.00%

Interest               ₹8,000
Maturity Date          01/10/2027
Maturity Amount        ₹1,08,000
```

---

# 11. FD Number

The system shall automatically generate a unique FD number.

Format:

```text
FD-1
FD-2
FD-3
FD-4
...
```

Rules:

- Format must be **FD-N**.
- Sequential and unique across the whole application.
- Automatically generated by the system.
- The user must not manually enter or modify the FD Number.
- Existing FD numbers must never be reused.
- FDs created under the earlier `FD-YY-NNN` format are renumbered once to `FD-N` during the upgrade; all history and renewal links are rewritten with them.

---

# 12. FD Master Information

Each FD record shall contain:

### Identification

- FD Number (system-generated, see §11)
- Linked Member — GEN No. (mandatory; every FD belongs to one member, see §53)
- Member Name (readable copy of the linked member)
- FD Form No. (optional, manual, not unique, searchable)

### Deposit

- Principal Amount
- Start Date
- Tenure
- Interest Rate

### Maturity

- Maturity Date
- Interest Amount
- Maturity Amount

### Status

- Active
- Closed

### Closure

- Closure Date
- Closure Type
- Closure Remark

### History

- Creation
- Renewal
- Closure

---

# 13. Interest Rate Configuration

Interest rates shall be maintained through the Settings screen.

Default configuration:

| Tenure | Interest Rate |
|---|---:|
| 1–364 Days | 4.00% |
| 365–729 Days | 8.00% |
| 730–1094 Days | 8.50% |
| 1095–1460 Days | 9.50% |

The user/administrator should be able to modify the interest percentage associated with the configured tenure slabs.

The calculation system shall always use the currently configured rate.

---

# 14. Interest Rate Selection

The applicable rate shall be selected according to the number of FD days.

Rules:

```text
1–364 days       → 4.00%
365–729 days     → 8.00%
730–1094 days    → 8.50%
1095–1460 days   → 9.50%
```

Examples:

```text
100 days   → 4.00%
364 days   → 4.00%

365 days   → 8.00%
500 days   → 8.00%
729 days   → 8.00%

730 days   → 8.50%
1094 days  → 8.50%

1095 days  → 9.50%
1460 days  → 9.50%
```

A tenure outside the configured range should not be accepted unless the rate configuration is updated.

---

# 15. Interest Calculation

The basic FD calculation shall use simple interest.

Formula:

```text
Interest =
Principal × Rate × Days / 365 / 100
```

Where:

- Principal = FD deposit amount
- Rate = annual interest rate
- Days = applicable number of days

---

# 16. Maturity Amount

The maturity amount shall be:

```text
Maturity Amount =
Principal + Interest
```

Example:

```text
Principal = ₹1,00,000
Rate = 8%
Days = 365

Interest =
1,00,000 × 8 × 365 / 365 / 100

Interest = ₹8,000

Maturity Amount =
₹1,00,000 + ₹8,000

Maturity Amount = ₹1,08,000
```

---

# 17. Rounding

The system should retain sufficient calculation precision internally.

The final payable/maturity amount shall be rounded to the nearest rupee.

Example:

```text
₹1,08,123.47
        ↓
₹1,08,123
```

---

# 18. Maturity Date

The maturity date shall be calculated using the FD start date and selected tenure.

Conceptually:

```text
Maturity Date =
Start Date + Tenure
```

The system shall display the maturity date automatically.

---

# 19. FD Details

The FD Details screen shall show all information for a selected FD.

Example:

```text
FD-1
ACTIVE

Member:
ABC (GEN-001)

Principal:
₹1,00,000

Interest Rate:
8.00%

Start Date:
01/10/2026

Maturity Date:
01/10/2027

Interest:
₹8,000

Maturity Amount:
₹1,08,000
```

Available actions:

- Edit
- Renew
- Close

The user shall also be able to view the FD history.

---

## 19.1 Editing an FD

The software shall allow an active FD to be edited after it has been created.

Editable fields: linked member (re-selection), FD Form No., deposit amount,
start date and tenure. The FD number, status, closure information and history
shall never be editable.

On save, the interest rate, interest and maturity amount shall be recalculated
using the same calculation rules as creation, and an `EDIT` event shall be
recorded in the FD history with the date and the new principal amount.

Validation shall match creation (a member selected, positive amount, valid
dates, tenure of at least one day). Like creation, the start date may be in
the future.

Only active FDs may be edited. Editing a closed FD shall be rejected with an
error. An FD that is past its maturity date but still open (active) remains
editable until it is closed.

---

# 20. FD History

Each FD shall maintain a simple chronological history.

Example:

```text
01/10/2026    OPEN
01/10/2027    RENEW
01/10/2028    CLOSE
```

History should contain relevant:

- Date
- Transaction type
- Amount
- Interest
- Reference/New FD Number
- Remarks

---

# 21. FD Renewal

The software shall support renewal of an FD.

Two renewal options shall be available.

## 21.1 Principal Only

Only the original principal is used for the new FD.

Example:

```text
Original Principal = ₹1,00,000
Earned Interest    = ₹8,000

New FD Principal   = ₹1,00,000
```

The earned interest is not added to the new FD principal.

## 21.2 Principal + Interest

The principal and earned interest are combined.

Example:

```text
Principal = ₹1,00,000
Interest  = ₹8,000

New FD Principal =
₹1,08,000
```

A renewal shall create a new FD number.

The previous FD shall remain available in history.

The renewal start date shall be on or after the previous FD's maturity date.
Renewals starting before maturity shall be rejected; back-dated renewals on or
after the maturity date are allowed. The default start date is the maturity
date.

Independently of the chosen start date, the renewal action itself shall only
be available on or after the maturity date: while today's date is before the
maturity date, renewal shall be rejected with an error stating that renewal
becomes available on the maturity date. This prevents future-dated renewals
from being submitted early.

---

# 22. Renewal Link

The system shall maintain a link between the old FD and the new FD.

Example:

```text
FD-1
     ↓
Renewed as
     ↓
FD-15
```

The user should be able to identify the previous and renewed FD.

## 22.1 Reversing a renewal

A renewal submitted by mistake shall be reversible while the renewed FD is
still active (not closed or renewed again). The reversal withdraws the renewed
FD with its history and reopens the previous FD, and shall require a reason
that is recorded in the previous FD's history as a `REVERSE` event.

If the renewed FD has already been closed or renewed, the reversal shall be
blocked.

---

# 23. FD Closure

An active FD may be closed.

The closure screen shall require:

- Closure Date
- Closure Remark (mandatory for premature closures; optional for closures at
  or after maturity)

The system shall automatically calculate the payable amount.

Two basic closure situations shall be supported:

1. Premature closure
2. Closure at/after maturity

## 23.1 Reopening a closed FD

A closed FD (matured or premature) shall be reopenable when the closure was
submitted by mistake. Reopening requires a reason, returns the FD to `ACTIVE`,
clears the closure fields, and appends a `REOPEN` event to the history — the
original closure events are never deleted.

FDs closed by renewal cannot be reopened directly; they must be reversed per
§22.1.

---

# 24. Premature Closure

Premature closure means:

```text
Closure Date < Maturity Date
```

For premature closure, interest shall be calculated only for the actual number of days the FD was held.

Formula:

```text
Days Held =
Closure Date - Start Date
```

The applicable interest rate shall be determined using the actual days held.

Example:

```text
Principal = ₹50,000
Original Tenure = 365 days
Original Rate = 8%

Actual Days Held = 200 days

Applicable Rate for 200 days = 4%
```

Then:

```text
Interest =
₹50,000 × 4 × 200 / 365 / 100
```

The resulting interest shall be added to the principal to determine the payable amount.

---

# 25. Closure at Maturity

If:

```text
Closure Date >= Maturity Date
```

the system shall not calculate additional interest beyond the original maturity amount.

The payable amount shall be capped at the calculated maturity amount.

Example:

```text
Principal         ₹1,00,000
Maturity Interest ₹8,000
Maturity Amount   ₹1,08,000
```

The payable amount shall remain:

```text
₹1,08,000
```

unless another business rule is introduced in a future version.

---

# 26. Closure Validation

The system shall:

- Require closure date
- Require a closure remark for premature closures
- Prevent closure before the FD start date
- Prevent closing an already closed FD
- Calculate actual days held for premature closure
- Calculate the correct applicable rate
- Display the final payable amount before confirmation

---

# 27. FD Status

Each FD shall have one primary status:

```text
ACTIVE
CLOSED
```

The system may also display closure type:

```text
MATURED
PREMATURE
RENEWED
```

where applicable.

---

# 28. Maturity Tracking

The Dashboard should show upcoming FD maturities.

The user should be able to identify:

- Maturing today
- Maturing within 7 days
- Maturing within 30 days
- Maturing within 90 days
- Custom date range

The maturity information should include:

- FD Number
- Member
- Principal
- Maturity Date
- Maturity Amount
- Days Remaining
- Status

Closed FDs shall not appear as active upcoming maturities.

---

# 29. Reports

Reports shall be generated primarily as **Excel files (.xlsx)**.

The application does not require a complex built-in report viewer.

The Reports screen shall allow the user to select the required report and export it to Excel.

## 29.1 FD Register

The FD Register Excel file shall contain:

- FD Number
- Member Name
- GEN No.
- Deposit Amount
- Interest Rate
- Start Date
- Tenure
- Maturity Date
- Interest Amount
- Maturity Amount
- Status

## 29.2 Maturity Report

The Maturity Report shall contain:

- FD Number
- Member Name
- Deposit Amount
- Interest Rate
- Maturity Date
- Days Remaining
- Maturity Amount
- Status

The user should be able to select a date range before exporting.

## 29.3 Active FD Report

The report shall contain only active FDs.

## 29.4 Closed FD Report

The report shall contain closed FDs.

## 29.5 Excel Requirements

Generated Excel files should:

- Use `.xlsx` format
- Have clear column headings
- Use appropriate date formatting
- Use appropriate currency formatting
- Allow Excel filtering
- Include a report title
- Include report generation date
- Be easy to print

## 29.6 Member-wise FD Summary

The report shall contain one row per member:

- GEN No.
- Member Name
- Designation
- Active FD Count
- Total Active FD Amount

It makes it easy to identify members holding multiple FDs, the total active amount per member and member-wise FD exposure.

---

# 30. Settings

Settings shall contain only configuration and administrative functions.

The Settings page shall use tabs — **only one tab's content is visible at a time**:

1. **Slab Rates**
2. **Login Credentials**
3. **Backup & Restore**

The user should not need to understand technical configuration.

## 30.1 Slab Rates

Rate rules shall initially be displayed in **read-only mode**:

```text
1–364 Days       4.00%    [Edit]
365–729 Days     8.00%    [Edit]
730–1094 Days    8.50%    [Edit]
1095–1460 Days   9.50%    [Edit]
```

Each rate rule has its own **Edit / Save / Cancel** action:

- **Edit** makes only that particular rule editable.
- **Save** persists the change and returns the rule to read-only.
- **Cancel** discards the change and returns the rule to read-only.

All rules must never be permanently editable at once.

## 30.2 Login Credentials

Two separate cards:

### Card 1 — Change Username

- Current Username
- Current Password (verification)
- New Username
- Save / Change Username button
- Validation and success/error message

### Card 2 — Change Password

- Current Password
- New Password
- Confirm New Password
- Change Password button
- Password validation (minimum 4 characters, matching confirmation)
- Success/error message

Username and password management remain visually and functionally separate. Both actions verify the current password.

## 30.3 Backup & Restore

- Create/Download Backup
- Restore Backup (with confirmation before restoring)
- Backup information/status
- Sample data loader (temporary demo data, available only while the database is empty)

---

# 31. About Page

The About page shall provide information about the software and developer.

## About Fixed Deposit Management

**Fixed Deposit Management Software**

Designed by:

**Aarti Tech Services**

The page may describe:

> Simple and user-friendly software for managing Fixed Deposits, calculating interest and maturity amounts, handling renewals and closures, and generating Excel reports.

### Aarti Tech Services

**FULL STACK DEVELOPMENT, SEO DIGITAL MARKETING, ENGINEERING SOLUTIONS**

Build, grow, and transform your business with modern websites, custom software, SEO, digital marketing, and innovative technology solutions.

Website:

https://aartitechservices.pages.dev/

Email:

aartitechservices@gmail.com

Phone:

+91 9869787575

```text
© 2026 Aarti Tech Services.
All rights reserved.
```

---

# 32. Privacy Policy

The application shall include a simple Privacy Policy page.

The Privacy Policy shall explain:

- What FD information is stored
- How the information is used
- That the application primarily stores data locally
- That FD data is not automatically sent to Aarti Tech Services
- User responsibility for backups
- Data security responsibility
- Third-party service handling
- Data sharing
- Policy updates
- Contact details

The Privacy Policy shall be accessible from the About page and/or Footer.

---

# 33. Footer

The footer shall appear consistently throughout the application.

Example:

```text
Fixed Deposit Management

Designed by Aarti Tech Services

Privacy Policy | About

© 2026 Aarti Tech Services. All rights reserved.
```

---

# 34. Main Navigation

The application shall have a simple navigation menu:

```text
Dashboard
FD Master
Member
Reports
Settings
```

The About and Privacy pages remain reachable from the footer.

The FD Details, Renew/Close and Add/Edit Member screens are opened from their list/profile records and do not need separate permanent menu items.

---

# 35. Recommended User Flow

The main workflow should be:

```text
Open Application
       ↓
Dashboard
       │
       ├── + New Member
       │        ↓
       │   Add Member (or Bulk Excel Upload)
       │        ↓
       │   Member saved (GEN No. unique)
       │
       └── FD Master
                ↓
        Search / View FDs
                │
                ├── View FD Details → Renew / Close
                │
                └── + Create New FD
                         ↓
                 Search & Select Member
                         ↓
                   Enter Deposit Details
                         ↓
                  Calculate Preview
                         ↓
                    Confirm & Save
                         │
                         ├── View / Renew / Close FD
                         └── Export Excel Reports
```

---

# 36. Simple Screen Structure

The final application should contain these primary pages:

| # | Screen | Purpose |
|---|---|---|
| 1 | Dashboard | Overview: KPIs, chart, upcoming maturities, quick actions |
| 2 | FD Master | FD search/list/filter + create FD |
| 3 | Member List | Search/list members + member-wise FD summary |
| 4 | Add / Edit Member | Create or modify a member record |
| 5 | Member Profile | Full member profile + linked FDs |
| 6 | Bulk Excel Upload | Import members from Excel with validation preview |
| 7 | Create FD | Select member, create and calculate FD |
| 8 | FD Details | View FD + history |
| 9 | Renew / Close | Renew or close selected FD |
| 10 | Reports | Export Excel reports (incl. member-wise summary) |
| 11 | Settings | Slab Rates / Login Credentials / Backup & Restore |
| 12 | About | Software and developer information |
| 13 | Privacy Policy | Privacy information |

The last two pages are informational rather than operational FD screens.

---

# 37. Data Requirements

The application shall maintain sufficient information to reproduce every FD calculation.

Each FD should retain:

- FD number (FD-N)
- FD Form No.
- Linked member (GEN No. + member name copy)
- Principal
- Start date
- Original tenure
- Original interest rate
- Maturity date
- Calculated interest
- Maturity amount
- Current status
- Closure information
- Renewal information
- Transaction history

Each member should retain:

- System Member ID
- GEN No. (unique)
- Personal, address, employment, nominee, government ID and banking fields (see §50)
- Created date, updated date, profile remarks

Historical FD information should not be overwritten when an FD is renewed or closed.

---

# 38. Calculation Integrity

The software shall prioritize calculation accuracy.

The system must:

- Use the configured interest rate
- Use the correct number of days
- Use the correct tenure slab
- Calculate premature closure using actual days held
- Prevent additional interest after maturity
- Apply final rounding consistently
- Store the calculation inputs used for the FD

A previously calculated FD should remain traceable even if interest-rate settings are changed later.

---

# 39. Usability Requirements

The application should be usable by a first-time user without technical training.

The UI should:

- Use simple labels
- Avoid unnecessary terminology
- Show calculations clearly
- Provide confirmation before important actions
- Show validation messages near the relevant field
- Avoid unnecessary popups
- Keep frequently used actions easily accessible
- Use consistent buttons and colors
- Clearly distinguish Active and Closed FDs

---

# 40. Error Handling

The software shall provide clear messages for invalid operations.

Examples:

```text
Please enter a valid deposit amount.

Please select a start date.

Please select a valid tenure.

No interest rate is configured for this tenure.

Closure date cannot be before the FD start date.

This FD is already closed.

This FD cannot be renewed because it is already closed.
```

Technical errors should not be shown directly to normal users.

---

# 41. Data Safety

The application should protect stored FD data against accidental loss.

The design should support reliable local data storage and backup.

The software should not delete historical FD information as part of normal renewal or closure operations.

---

# 42. Performance

The application should open quickly and remain responsive during normal usage.

Search, filtering, FD creation, calculation, renewal, closure and report generation should complete without unnecessary delays for normal-sized datasets.

The application should remain usable with a large number of FD records.

---

# 43. Security

The application should follow basic desktop application security practices.

Important requirements:

- Protect local application data
- Avoid exposing sensitive information unnecessarily
- Validate all financial inputs
- Validate dates
- Prevent invalid calculations
- Prevent duplicate FD numbers
- Prevent accidental duplicate transactions
- Protect exported files as user data

---

# 44. Technology Direction

The recommended technology direction for the standalone cross-platform application is:

```text
Desktop Framework : Wails
Backend           : Go
Frontend          : React + TypeScript
UI                : Tailwind CSS
Database          : SQLite
Excel Export      : XLSX
```

The application should be designed so that the frontend remains fully responsive and the business/calculation logic remains independent from the UI.

---

# 45. Cross-Platform Requirement

The application should be capable of producing platform-specific builds for:

```text
Windows → .exe
macOS   → Application package
Linux   → Linux desktop package
```

The FD calculation rules and data structure should remain consistent across platforms.

---

# 46. Future Expansion

The application should be structured so that future features can be added without redesigning the core FD module.

Possible future features may include:

- Advanced interest schemes
- Additional FD types
- TDS
- Senior citizen rates
- Multiple organizations
- Cloud backup
- Synchronization
- Mobile application
- Online access
- Additional financial modules

These features are **not part of Version 1**.

---

# 47. Core Business Rules Summary

The most important rules are:

### Interest

```text
Interest =
Principal × Rate × Days / 365 / 100
```

### Maturity

```text
Maturity Amount =
Principal + Interest
```

### Rate Slabs

```text
1–364 days       → 4.00%
365–729 days     → 8.00%
730–1094 days    → 8.50%
1095–1460 days   → 9.50%
```

### Premature Closure

```text
Actual Days =
Closure Date - Start Date
```

The rate is determined from the **actual days held**.

### Maturity Closure

If the FD is closed on or after maturity:

```text
Payable Amount ≤ Original Maturity Amount
```

No additional interest is automatically added after maturity.

### Renewal

Two options:

```text
Principal Only

OR

Principal + Interest
```

Renewal always creates a **new FD number**.

### FD Status

```text
ACTIVE
CLOSED
```

### Member Rules

```text
GEN No. = unique member identifier (mandatory, no duplicates)
One Member → Many FDs
Every FD is linked to exactly one member
FD Number = system-generated unique FD-N (never reused)
Bulk Excel import never overwrites existing members
```

---

# 48. Final Product Vision

The Fixed Deposit Management Software should be:

**Simple. Fast. Accurate. Responsive. Offline-friendly.**

The user should be able to open the software and perform the most common task in only a few steps:

```text
Open
 ↓
Search/Create FD
 ↓
View Calculation
 ↓
Save
 ↓
Manage / Renew / Close
 ↓
Export Excel Report
```

The application should avoid unnecessary complexity and focus on making Fixed Deposit management easy and reliable.

---

# 49. Branding

## Product

**Fixed Deposit Management**

## Developer

**Aarti Tech Services**

## Services

**FULL STACK DEVELOPMENT, SEO DIGITAL MARKETING, ENGINEERING SOLUTIONS**

## Website

https://aartitechservices.pages.dev/

## Contact

aartitechservices@gmail.com  
+91 9869787575

## Copyright

**© 2026 Aarti Tech Services. All rights reserved.**

---

# 50. Member Master

Each member record shall contain the following fields.

### Identification

- **GEN No.** — unique member ID and primary identifier
  - Mandatory
  - Must be unique — duplicate GEN Nos. shall not be allowed (enforced on add, edit and import)
- System Member ID (internal, system-assigned)
- Created Date
- Updated Date

### Personal Information

- Name (mandatory)
- Date of Birth
- Mobile No.
- Email ID

### Address

- Present Address
- Permanent Address

### Employment Information

- Employer Name
- Department
- Designation
- Token No.

### Nominee Information

- Nominee Name
- Nominee Relationship

### Government ID

- Aadhaar No.
- PAN No.

### Banking Information

- Bank Name
- Account No.
- IFSC Code

### System Information

- System Member ID
- Created Date
- Updated Date
- Profile Remarks

The system shall provide **Add New Member**, **View Member**, **Edit / Modify Member** and **Search Member** operations. GEN No. remains the primary unique identifier at all times.

---

# 51. Member Page and Profile

## 51.1 Member Search

A prominent search facility on the Member page shall support:

- GEN No.
- Name
- Token No.
- Designation
- PAN
- Aadhaar
- Mobile No.

## 51.2 Member List

All members shall be displayed in a searchable table:

| GEN No. | Name | Designation | Token No. | Mobile | Active FDs | Total Active FD Amount | Actions |
|---|---|---|---|---|---:|---:|---|
| GEN-001 | Member Name | Clerk | 125 | XXXXXXXX | 3 | ₹5,00,000 | View / Edit |

Actions: **View** and **Edit / Modify**.

## 51.3 Member Profile View

The View Member page shall display the complete member profile in a clean, grouped layout (§50 field groups).

It shall also display the member's FD summary:

- Total Active FDs
- Total Active FD Amount
- FD Number
- FD Amount (principal)
- FD Start Date
- FD Maturity Date
- FD Status

Each FD in the summary opens the individual FD details screen directly from the member profile.

---

# 52. Bulk Excel Member Import

A **Bulk Excel Upload** facility on the Member page shall import existing member records without creating each member manually.

## 52.1 Mandatory Excel Columns

| Column | Requirement |
|---|---|
| GEN No. | Mandatory and unique |
| Name | Mandatory |

## 52.2 Optional Excel Columns

DOB, Mobile, Email, Present Address, Permanent Address, Employer Name, Department, Designation, Token No., Nominee Name, Nominee Relationship, Aadhaar, PAN, Bank Name, Account No., IFSC Code, Profile Remarks.

## 52.3 Import Validation

All records shall be validated before importing:

- GEN No. cannot be blank
- Name cannot be blank
- GEN No. must be unique
- Duplicate GEN Nos. within the uploaded file must be detected
- GEN Nos. already existing in the system must be identified
- Invalid or incorrectly formatted data must be reported

An **Import Preview** shall be shown before anything is saved:

```text
Import Summary
Total Records:      500
Valid Records:      485
Duplicate GEN No.:  10
Missing Mandatory:  5
Existing in System: 8
```

The user can review the errors, correct the Excel file and repeat the preview before completing the import.

## 52.4 Existing Member Handling

If a GEN No. already exists:

- The member is **not** duplicated
- The GEN No. is clearly identified in the preview
- The record is **skipped** — existing member information is never overwritten by default

An explicit "Update Existing Members" option may be added in a future version.

---

# 53. FD–Member Relationship

Every FD shall be linked to an existing Member Master record:

```text
GEN-001
   ├── FD-1
   ├── FD-8
   └── FD-15

GEN-002
   ├── FD-2
   └── FD-11
```

One member has one master profile but can hold any number of FDs.

Rules:

- Creating an FD requires selecting an existing member (§10).
- A renewed FD inherits the member link of the FD it replaces; the new FD's FD Form No. starts empty.
- FDs created before members existed are matched automatically when their stored customer number equals a member's GEN No.
- FDs that match no member are shown as **Unassigned** in the member-wise report and can be linked later by editing the FD and selecting a member.
- The FD stores a readable copy of the member's name and GEN No., refreshed whenever the member record changes.

---

# 54. End of SRS

This SRS defines the Version 1 functional and business requirements for the Fixed Deposit Management Software.

The implementation should follow these requirements while keeping the application simple, responsive, maintainable and user-friendly.