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

- FD Dashboard
- FD search and listing
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
- New FD
- Reports
- Settings
- About

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

## 9.1 FD Dashboard

The Dashboard shall also function as the main FD List.

It shall provide:

- Total active FDs
- Total deposit amount
- Upcoming maturities
- Search
- Filters
- FD list
- Quick action to create a new FD

Example:

```text
Active FD       125
Total Deposit   ₹1,25,00,000
Maturing Soon   8
```

Quick action:

**+ Create New FD**

---

## 9.2 FD Search and List

The Dashboard shall contain the FD list.

The user should be able to search using:

- FD Number
- Customer/Member Name
- Customer/Member Number

Filters:

- All
- Active
- Closed
- Maturing

The FD list should show:

- FD Number
- Customer/Member
- Deposit Amount
- Interest Rate
- Start Date
- Maturity Date
- Maturity Amount
- Status

Clicking an FD shall open its FD Details.

---

# 10. Create FD

The Create FD screen shall allow the user to create a new Fixed Deposit.

Required information:

- Customer/Member
- Customer/Member Number
- Deposit Amount
- Start Date
- Tenure

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

Default format:

```text
FD-YY-001
FD-YY-002
FD-YY-003
```

Example:

```text
FD-26-001
```

The numbering should automatically increment.

The numbering/reset period should be configurable in the future if required.

---

# 12. FD Master Information

Each FD record shall contain:

### Identification

- FD Number
- Customer/Member Name
- Customer/Member Number

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
FD-26-001
ACTIVE

Customer:
ABC

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

Editable fields: customer name, customer number, deposit amount, start date
and tenure. The FD number, status, closure information and history shall never
be editable.

On save, the interest rate, interest and maturity amount shall be recalculated
using the same calculation rules as creation, and an `EDIT` event shall be
recorded in the FD history with the date and the new principal amount.

Validation shall match creation (name required, positive amount, valid dates,
tenure of at least one day). Like creation, the start date may be in the
future.

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
FD-26-001
     ↓
Renewed as
     ↓
FD-27-015
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
- Customer
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
- Customer/Member
- Customer/Member Number
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
- Customer/Member
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

---

# 30. Settings

The Settings page shall remain simple.

Primary setting:

### Interest Rate Configuration

```text
1–364 Days       [4.00%]
365–729 Days     [8.00%]
730–1094 Days    [8.50%]
1095–1460 Days   [9.50%]

[Save Changes]
```

The user should not need to understand technical configuration.

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

The application shall have a simple navigation menu.

```text
Dashboard
New FD
Reports
Settings
About
```

The FD Details and Renew/Close screens are opened from the Dashboard/FD record and do not need separate permanent menu items.

---

# 35. Recommended User Flow

The main workflow should be:

```text
Open Application
       ↓
Dashboard
       ↓
Search / View Existing FD
       │
       ├── View FD Details
       │       ├── Renew
       │       └── Close
       │
       └── Create New FD
               ↓
          Enter Details
               ↓
        Calculate Preview
               ↓
          Confirm & Save
```

---

# 36. Simple Screen Structure

The final application should contain these primary pages:

| # | Screen | Purpose |
|---|---|---|
| 1 | Dashboard | Overview + FD search/list/manage |
| 2 | Create FD | Create and calculate FD |
| 3 | FD Details | View FD + history |
| 4 | Renew / Close | Renew or close selected FD |
| 5 | Reports | Export Excel reports |
| 6 | Settings | Configure interest rates |
| 7 | About | Software and developer information |
| 8 | Privacy Policy | Privacy information |

The last two pages are informational rather than operational FD screens.

---

# 37. Data Requirements

The application shall maintain sufficient information to reproduce every FD calculation.

Each FD should retain:

- FD number
- Customer/member information
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

# 50. End of SRS

This SRS defines the Version 1 functional and business requirements for the Fixed Deposit Management Software.

The implementation should follow these requirements while keeping the application simple, responsive, maintainable and user-friendly.