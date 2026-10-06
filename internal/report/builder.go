package report

import (
	"time"

	"fdats/internal/api"
	"fdats/internal/domain"
)

// buildSheet turns FD records into report headings and rows (SRS §29.1–§29.4).
func buildSheet(opts Options) ([]string, [][]any) {
	switch opts.Kind {
	case api.ReportMaturity:
		return maturitySheet(opts)
	case api.ReportActive:
		return registerSheet(selectStatus(opts.FDs, domain.StatusActive), false)
	case api.ReportClosed:
		return registerSheet(selectStatus(opts.FDs, domain.StatusClosed), true)
	default:
		return registerSheet(opts.FDs, true)
	}
}

// Sheet returns the headings and rows for a report without writing a file,
// so the same layout can be previewed on screen before exporting.
func Sheet(opts Options) ([]string, [][]any) {
	return buildSheet(opts)
}

func selectStatus(fds []domain.FixedDeposit, status domain.FDStatus) []domain.FixedDeposit {
	out := make([]domain.FixedDeposit, 0, len(fds))
	allActive := true
	allClosed := true
	for _, fd := range fds {
		if fd.Status != domain.StatusActive {
			allActive = false
		}
		if fd.Status != domain.StatusClosed {
			allClosed = false
		}
	}
	if (status == domain.StatusActive && allActive) || (status == domain.StatusClosed && allClosed) {
		return fds
	}
	for _, fd := range fds {
		if fd.Status == status {
			out = append(out, fd)
		}
	}
	return out
}

func registerSheet(fds []domain.FixedDeposit, withClosure bool) ([]string, [][]any) {
	headers := []string{
		"FD Number", "Customer/Member", "Customer/Member Number", "Deposit Amount",
		"Interest Rate", "Start Date", "Tenure (Days)", "Maturity Date",
		"Interest Amount", "Maturity Amount", "Status",
	}
	if withClosure {
		headers = append(headers, "Closure Date", "Closure Type", "Payable Amount", "Remark")
	}

	rows := make([][]any, 0, len(fds))
	for _, fd := range fds {
		row := []any{
			fd.FDNumber,
			fd.CustomerName,
			fd.CustomerNumber,
			fd.Principal,
			fd.InterestRate,
			mustDate(fd.StartDate),
			int64(fd.TenureDays),
			mustDate(fd.MaturityDate),
			fd.InterestAmount,
			fd.MaturityAmount,
			string(fd.Status),
		}
		if withClosure {
			row = append(row,
				optionalDate(fd.ClosureDate),
				optionalClosureType(fd.ClosureType),
				optionalInt(fd.ClosurePayable),
				fd.ClosureRemark,
			)
		}
		rows = append(rows, row)
	}
	return headers, rows
}

func maturitySheet(opts Options) ([]string, [][]any) {
	headers := []string{
		"FD Number", "Customer/Member", "Deposit Amount", "Interest Rate",
		"Maturity Date", "Days Remaining", "Maturity Amount", "Status",
	}

	rows := make([][]any, 0, len(opts.FDs))
	for _, fd := range opts.FDs {
		maturity := mustDate(fd.MaturityDate)
		daysRemaining := int64(daysBetween(opts.GeneratedAt, maturity))
		rows = append(rows, []any{
			fd.FDNumber,
			fd.CustomerName,
			fd.Principal,
			fd.InterestRate,
			maturity,
			daysRemaining,
			fd.MaturityAmount,
			string(fd.Status),
		})
	}
	return headers, rows
}

func mustDate(s string) time.Time {
	t, err := domain.ParseDate(s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func optionalDate(s *string) any {
	if s == nil || *s == "" {
		return ""
	}
	return mustDate(*s)
}

func optionalClosureType(t *domain.ClosureType) string {
	if t == nil {
		return ""
	}
	return string(*t)
}

func optionalInt(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func daysBetween(a, b time.Time) int {
	au := time.Date(a.UTC().Year(), a.UTC().Month(), a.UTC().Day(), 0, 0, 0, 0, time.UTC)
	bu := time.Date(b.UTC().Year(), b.UTC().Month(), b.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return int(bu.Sub(au).Hours() / 24)
}
