// Package api defines the data contracts shared between the service layer
// and the Wails binding layer. All JSON tags become the field names used by
// the frontend.
package api

import "fdats/internal/domain"

// PreviewRequest calculates an FD without saving it (SRS §10). The member
// is mandatory: the FD's readable name/GEN copies come from the selected
// member, and FD Form No. is the optional manual form number (§12).
type PreviewRequest struct {
	MemberID   int64  `json:"memberId"`
	FDFormNo   string `json:"fdFormNo"`
	Principal  int64  `json:"principal"`
	StartDate  string `json:"startDate"`
	TenureDays int    `json:"tenureDays"`
}

// EditFDRequest updates the editable fields of an existing ACTIVE FD;
// derived amounts are recomputed by the service. MemberID relinks the FD
// (SRS §53); an empty FDFormNo clears the manual form number.
type EditFDRequest struct {
	FDNumber   string `json:"fdNumber"`
	MemberID   int64  `json:"memberId"`
	FDFormNo   string `json:"fdFormNo"`
	Principal  int64  `json:"principal"`
	StartDate  string `json:"startDate"`
	TenureDays int    `json:"tenureDays"`
}

// Calculation is a full interest/maturity calculation result (SRS §10).
type Calculation struct {
	Principal      int64   `json:"principal"`
	StartDate      string  `json:"startDate"`
	TenureDays     int     `json:"tenureDays"`
	Days           int     `json:"days"`
	RatePercent    float64 `json:"ratePercent"`
	Interest       int64   `json:"interest"`
	MaturityDate   string  `json:"maturityDate"`
	MaturityAmount int64   `json:"maturityAmount"`
}

// FD is the FD master record as seen by the UI (SRS §12).
type FD struct {
	FDNumber        string              `json:"fdNumber"`
	MemberID        *int64              `json:"memberId,omitempty"`
	FDFormNo        string              `json:"fdFormNo"`
	CustomerName    string              `json:"customerName"`
	CustomerNumber  string              `json:"customerNumber"`
	Principal       int64               `json:"principal"`
	StartDate       string              `json:"startDate"`
	TenureDays      int                 `json:"tenureDays"`
	InterestRate    float64             `json:"interestRate"`
	MaturityDate    string              `json:"maturityDate"`
	InterestAmount  int64               `json:"interestAmount"`
	MaturityAmount  int64               `json:"maturityAmount"`
	Status          domain.FDStatus     `json:"status"`
	ClosureDate     *string             `json:"closureDate,omitempty"`
	ClosureType     *domain.ClosureType `json:"closureType,omitempty"`
	ClosureRemark   string              `json:"closureRemark,omitempty"`
	ClosureRate     *float64            `json:"closureRate,omitempty"`
	ClosureDays     *int                `json:"closureDays,omitempty"`
	ClosureInterest *int64              `json:"closureInterest,omitempty"`
	ClosurePayable  *int64              `json:"closurePayable,omitempty"`
	RenewedFrom     *string             `json:"renewedFrom,omitempty"`
	RenewedTo       *string             `json:"renewedTo,omitempty"`
	CreatedAt       string              `json:"createdAt"`
	UpdatedAt       string              `json:"updatedAt"`
}

// FDDetail is an FD together with its chronological history (SRS §19–§20).
type FDDetail struct {
	FD      FD                    `json:"fd"`
	History []domain.HistoryEntry `json:"history"`
}

// ListRequest drives the dashboard FD list (SRS §9.2).
type ListRequest struct {
	Search   string `json:"search"`
	Filter   string `json:"filter"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

// ListResponse is one page of FD results.
type ListResponse struct {
	Items    []FD `json:"items"`
	Total    int  `json:"total"`
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
}

// DashboardStats are the KPI numbers shown on the dashboard (SRS §9.1, §28).
type DashboardStats struct {
	TotalFDs        int    `json:"totalFds"`
	ActiveFDs       int    `json:"activeFds"`
	ActivePrincipal int64  `json:"activePrincipal"`
	TotalInterest   int64  `json:"totalInterest"`
	FYLabel         string `json:"fyLabel"`
	FYDeposits      int64  `json:"fyDeposits"`
	MaturingToday   int    `json:"maturingToday"`
	Maturing7       int    `json:"maturing7"`
	Maturing30      int    `json:"maturing30"`
	Maturing90      int    `json:"maturing90"`
}

// UpcomingFD is one row of the maturity tracker (SRS §28).
type UpcomingFD struct {
	FDNumber       string          `json:"fdNumber"`
	CustomerName   string          `json:"customerName"`
	Principal      int64           `json:"principal"`
	MaturityDate   string          `json:"maturityDate"`
	MaturityAmount int64           `json:"maturityAmount"`
	DaysRemaining  int             `json:"daysRemaining"`
	Status         domain.FDStatus `json:"status"`
}

// UpcomingRequest selects a maturity window (SRS §28).
type UpcomingRequest struct {
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
}

// RenewRequest renews an FD (SRS §21).
type RenewRequest struct {
	FDNumber   string `json:"fdNumber"`
	Mode       string `json:"mode"`
	StartDate  string `json:"startDate"`
	TenureDays int    `json:"tenureDays"`
	Remark     string `json:"remark"`
}

// RenewResult links the closed FD to the newly created one (SRS §22).
type RenewResult struct {
	PreviousFD FD `json:"previousFd"`
	NewFD      FD `json:"newFd"`
}

// CloseRequest closes an FD (SRS §23, §26).
type CloseRequest struct {
	FDNumber    string `json:"fdNumber"`
	ClosureDate string `json:"closureDate"`
	Remark      string `json:"remark"`
}

// ReopenRequest returns a closed FD to ACTIVE status; the remark is the
// mandatory audit reason for the reversal.
type ReopenRequest struct {
	FDNumber string `json:"fdNumber"`
	Remark   string `json:"remark"`
}

// ReverseRenewalRequest withdraws the renewed FD and reopens the previous one.
type ReverseRenewalRequest struct {
	FDNumber string `json:"fdNumber"`
	Remark   string `json:"remark"`
}

// ClosurePreview shows the payable amount before confirmation (SRS §26).
type ClosurePreview struct {
	FDNumber     string             `json:"fdNumber"`
	ClosureDate  string             `json:"closureDate"`
	IsPremature  bool               `json:"isPremature"`
	ClosureType  domain.ClosureType `json:"closureType"`
	DaysHeld     int                `json:"daysHeld"`
	RatePercent  float64            `json:"ratePercent"`
	Interest     int64              `json:"interest"`
	Payable      int64              `json:"payable"`
	MaturityDate string             `json:"maturityDate"`
}

// SaveSlabsRequest replaces the interest rate configuration (SRS §30).
type SaveSlabsRequest struct {
	Slabs []domain.RateSlab `json:"slabs"`
}

// Report kinds (SRS §29).
const (
	ReportRegister = "REGISTER"
	ReportMaturity = "MATURITY"
	ReportActive   = "ACTIVE"
	ReportClosed   = "CLOSED"
	ReportMembers  = "MEMBERS"
)

// ReportRequest selects a report and its export location (SRS §29).
type ReportRequest struct {
	Kind     string `json:"kind"`
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
	Path     string `json:"path"`
}

// ReportResult confirms a successful export.
type ReportResult struct {
	Path     string `json:"path"`
	RowCount int    `json:"rowCount"`
}

// ReportPreview is the on-screen preview of a report before exporting.
type ReportPreview struct {
	Kind    string     `json:"kind"`
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
	Total   int        `json:"total"`
}

// MaturityBucket aggregates active FD maturities for one chart column.
type MaturityBucket struct {
	Period string `json:"period"` // YYYY-MM
	Label  string `json:"label"`  // short month label, e.g. "Nov"
	Count  int    `json:"count"`
	Amount int64  `json:"amount"`
}

// SystemStatus tells the UI whether the backend is ready (SRS §40: technical
// errors are surfaced politely rather than crashing).
type SystemStatus struct {
	Ready  bool   `json:"ready"`
	Error  string `json:"error"`
	Authed bool   `json:"authed"`
}

// LoginRequest signs in with username and password.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangeCredentialsRequest updates the login from Settings while signed in.
// An empty new username or password keeps the current value.
type ChangeCredentialsRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewUsername     string `json:"newUsername"`
	NewPassword     string `json:"newPassword"`
}

// ChangeUsernameRequest replaces just the username (SRS §30.2: the current
// password is always required).
type ChangeUsernameRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewUsername     string `json:"newUsername"`
}

// ChangePasswordRequest replaces just the password (SRS §30.2).
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ResetCredentialsRequest recovers a forgotten login using the recovery code.
type ResetCredentialsRequest struct {
	RecoveryCode string `json:"recoveryCode"`
	NewUsername  string `json:"newUsername"`
	NewPassword  string `json:"newPassword"`
}

// AuthInfo exposes the signed-in account details for Settings.
type AuthInfo struct {
	Username     string `json:"username"`
	RecoveryCode string `json:"recoveryCode"`
}

// Member is the member master record as seen by the UI, including the
// active-FD aggregates shown in the member list (SRS §50, §51.2).
type Member struct {
	ID                  int64  `json:"id"`
	GENNo               string `json:"genNo"`
	Name                string `json:"name"`
	DOB                 string `json:"dob"`
	Mobile              string `json:"mobile"`
	Email               string `json:"email"`
	PresentAddress      string `json:"presentAddress"`
	PermanentAddress    string `json:"permanentAddress"`
	EmployerName        string `json:"employerName"`
	Department          string `json:"department"`
	Designation         string `json:"designation"`
	TokenNo             string `json:"tokenNo"`
	NomineeName         string `json:"nomineeName"`
	NomineeRelationship string `json:"nomineeRelationship"`
	Aadhaar             string `json:"aadhaar"`
	PAN                 string `json:"pan"`
	BankName            string `json:"bankName"`
	AccountNo           string `json:"accountNo"`
	IFSC                string `json:"ifsc"`
	ProfileRemarks      string `json:"profileRemarks"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
	ActiveFDCount       int    `json:"activeFdCount"`
	ActiveFDAmount      int64  `json:"activeFdAmount"`
}

// SaveMemberRequest creates (ID = 0) or updates a member (SRS §50).
// GEN No. and Name are mandatory.
type SaveMemberRequest struct {
	ID                  int64  `json:"id"`
	GENNo               string `json:"genNo"`
	Name                string `json:"name"`
	DOB                 string `json:"dob"`
	Mobile              string `json:"mobile"`
	Email               string `json:"email"`
	PresentAddress      string `json:"presentAddress"`
	PermanentAddress    string `json:"permanentAddress"`
	EmployerName        string `json:"employerName"`
	Department          string `json:"department"`
	Designation         string `json:"designation"`
	TokenNo             string `json:"tokenNo"`
	NomineeName         string `json:"nomineeName"`
	NomineeRelationship string `json:"nomineeRelationship"`
	Aadhaar             string `json:"aadhaar"`
	PAN                 string `json:"pan"`
	BankName            string `json:"bankName"`
	AccountNo           string `json:"accountNo"`
	IFSC                string `json:"ifsc"`
	ProfileRemarks      string `json:"profileRemarks"`
}

// MemberListRequest drives the member list (SRS §51.1, §51.2).
type MemberListRequest struct {
	Search   string `json:"search"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

// MemberListResponse is one page of member results.
type MemberListResponse struct {
	Items    []Member `json:"items"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
}

// MemberProfile is a member together with every FD linked to it (SRS §51.3).
type MemberProfile struct {
	Member Member `json:"member"`
	FDs    []FD   `json:"fds"`
}

// Row statuses in an import preview (SRS §52.3, §52.4).
const (
	ImportRowNew       = "new"       // valid and not yet in the system
	ImportRowExisting  = "existing"  // GEN No. already in the system — will be skipped
	ImportRowDuplicate = "duplicate" // GEN No. repeated within the file
	ImportRowInvalid   = "invalid"   // missing mandatory data or bad format
)

// MemberImportRow reports the outcome for one Excel data row (SRS §52.3).
type MemberImportRow struct {
	Row    int    `json:"row"` // Excel row number
	GENNo  string `json:"genNo"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// MemberImportPreview is the validation summary shown before importing
// (SRS §52.3). Nothing is saved until the preview is accepted.
type MemberImportPreview struct {
	Total           int               `json:"total"`
	Valid           int               `json:"valid"`
	DuplicateGEN    int               `json:"duplicateGen"`
	MissingMandatory int              `json:"missingMandatory"`
	Invalid         int               `json:"invalid"`
	Existing        int               `json:"existing"`
	Rows            []MemberImportRow `json:"rows"`
}

// MemberImportResult confirms a completed import (SRS §52.4).
type MemberImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Total    int `json:"total"`
}
