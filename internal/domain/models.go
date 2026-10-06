package domain

// FDStatus is the primary lifecycle status of a Fixed Deposit (SRS §27).
type FDStatus string

const (
	StatusActive FDStatus = "ACTIVE"
	StatusClosed FDStatus = "CLOSED"
)

// ClosureType describes why an FD was closed (SRS §27).
type ClosureType string

const (
	ClosureMatured   ClosureType = "MATURED"
	ClosurePremature ClosureType = "PREMATURE"
	ClosureRenewed   ClosureType = "RENEWED"
)

// HistoryEventType is a chronological FD history entry (SRS §20).
type HistoryEventType string

const (
	EventOpen  HistoryEventType = "OPEN"
	EventRenew HistoryEventType = "RENEW"
	EventClose HistoryEventType = "CLOSE"
)

// Renewal modes (SRS §21).
const (
	RenewPrincipalOnly        = "PRINCIPAL_ONLY"
	RenewPrincipalPlusInterest = "PRINCIPAL_PLUS_INTEREST"
)

// FD filters used by the dashboard list (SRS §9.2).
const (
	FilterAll      = "ALL"
	FilterActive   = "ACTIVE"
	FilterClosed   = "CLOSED"
	FilterMaturing = "MATURING"
)

// FixedDeposit is the FD master record (SRS §12).
// Monetary values are whole rupees; dates are ISO YYYY-MM-DD strings.
type FixedDeposit struct {
	FDNumber       string
	CustomerName   string
	CustomerNumber string
	Principal      int64
	StartDate      string
	TenureDays     int
	InterestRate   float64
	MaturityDate   string
	InterestAmount int64
	MaturityAmount int64
	Status         FDStatus
	ClosureDate    *string
	ClosureType    *ClosureType
	ClosureRemark  string
	ClosureRate    *float64
	ClosureDays    *int
	ClosureInterest *int64
	ClosurePayable  *int64
	RenewedFrom    *string
	RenewedTo      *string
	CreatedAt      string
	UpdatedAt      string
}

// HistoryEntry is one chronological event recorded for an FD (SRS §20).
type HistoryEntry struct {
	ID          int64            `json:"id"`
	FDNumber    string           `json:"fdNumber"`
	EventDate   string           `json:"eventDate"`
	EventType   HistoryEventType `json:"eventType"`
	Amount      int64            `json:"amount"`
	Interest    int64            `json:"interest"`
	ReferenceFD string           `json:"referenceFd"`
	Remarks     string           `json:"remarks"`
	CreatedAt   string           `json:"createdAt"`
}

// RateSlab maps a tenure range (in days) to an annual interest rate (SRS §13).
type RateSlab struct {
	ID          int64   `json:"id"`
	SortOrder   int     `json:"sortOrder"`
	MinDays     int     `json:"minDays"`
	MaxDays     int     `json:"maxDays"`
	RatePercent float64 `json:"ratePercent"`
	Label       string  `json:"label"`
}
