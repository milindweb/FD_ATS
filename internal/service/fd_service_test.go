package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fdats/internal/api"
	"fdats/internal/domain"
	"fdats/internal/repo"
	"fdats/internal/service"
)

// fixedNow is the deterministic "today" used by every test: 06/10/2026.
var fixedNow = time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC)

func newTestService(t *testing.T) *service.FDService {
	t.Helper()
	db, err := repo.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	svc, err := service.NewFDService(db)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	svc.SetClock(func() time.Time { return fixedNow })
	return svc
}

func baseRequest() api.PreviewRequest {
	return api.PreviewRequest{
		CustomerName:   "ABC",
		CustomerNumber: "CUST-001",
		Principal:      100000,
		StartDate:      "2026-10-01",
		TenureDays:     365,
	}
}

func TestPreviewMatchesSRSExample(t *testing.T) {
	svc := newTestService(t)

	res, err := svc.Preview(baseRequest())
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if res.RatePercent != 8.00 || res.Interest != 8000 || res.MaturityAmount != 108000 {
		t.Errorf("preview = %+v, want rate 8 / interest 8000 / maturity 108000", res)
	}
	if res.MaturityDate != "2027-10-01" {
		t.Errorf("maturity=%s, want 2027-10-01", res.MaturityDate)
	}
}

func TestCreateAssignsSequentialFDNumbers(t *testing.T) {
	svc := newTestService(t)

	first, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatalf("create 1: %v", err)
	}
	if first.FDNumber != "FD-26-001" {
		t.Errorf("first number=%s, want FD-26-001", first.FDNumber)
	}
	if first.Status != domain.StatusActive {
		t.Errorf("status=%s, want ACTIVE", first.Status)
	}

	req2 := baseRequest()
	req2.CustomerName = "PQR"
	second, err := svc.Create(req2)
	if err != nil {
		t.Fatalf("create 2: %v", err)
	}
	if second.FDNumber != "FD-26-002" {
		t.Errorf("second number=%s, want FD-26-002", second.FDNumber)
	}
}

func TestCreateUsesStartDateYearForPeriod(t *testing.T) {
	svc := newTestService(t)

	req := baseRequest()
	req.StartDate = "2027-01-15"
	fd, err := svc.Create(req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if fd.FDNumber != "FD-27-001" {
		t.Errorf("number=%s, want FD-27-001", fd.FDNumber)
	}
}

func TestCreateRecordsOpenHistory(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	detail, err := svc.Get(created.FDNumber)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(detail.History) != 1 {
		t.Fatalf("history length=%d, want 1", len(detail.History))
	}
	h := detail.History[0]
	if h.EventType != domain.EventOpen || h.Amount != 100000 || h.EventDate != "2026-10-01" {
		t.Errorf("history=%+v, want OPEN 100000 on 2026-10-01", h)
	}
}

func TestCreateValidationMessages(t *testing.T) {
	svc := newTestService(t)

	cases := []struct {
		name string
		mut  func(*api.PreviewRequest)
		want error
	}{
		{"missing name", func(r *api.PreviewRequest) { r.CustomerName = " " }, domain.ErrCustomerNameRequired},
		{"missing number", func(r *api.PreviewRequest) { r.CustomerNumber = "" }, domain.ErrCustomerNumberRequired},
		{"zero amount", func(r *api.PreviewRequest) { r.Principal = 0 }, domain.ErrInvalidAmount},
		{"negative amount", func(r *api.PreviewRequest) { r.Principal = -500 }, domain.ErrInvalidAmount},
		{"missing start", func(r *api.PreviewRequest) { r.StartDate = "" }, domain.ErrStartDateRequired},
		{"bad start", func(r *api.PreviewRequest) { r.StartDate = "31/12/2026" }, domain.ErrInvalidDate},
		{"zero tenure", func(r *api.PreviewRequest) { r.TenureDays = 0 }, domain.ErrInvalidTenure},
		{"unconfigured tenure", func(r *api.PreviewRequest) { r.TenureDays = 5000 }, domain.ErrNoRateForTenure},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := baseRequest()
			c.mut(&req)
			if _, err := svc.Create(req); !errors.Is(err, c.want) {
				t.Errorf("err=%v, want %v", err, c.want)
			}
			if _, err := svc.Preview(req); !errors.Is(err, c.want) {
				t.Errorf("preview err=%v, want %v", err, c.want)
			}
		})
	}
}

func TestListSearchAndFilters(t *testing.T) {
	svc := newTestService(t)

	// FD-26-001: ABC, active, matures 2027-10-01.
	// FD-26-002: XYZ, active, short tenure matures 2026-10-10 (maturing soon).
	// FD-26-003: PQR, closed early.
	req1 := baseRequest()
	if _, err := svc.Create(req1); err != nil {
		t.Fatal(err)
	}
	req2 := baseRequest()
	req2.CustomerName = "XYZ"
	req2.CustomerNumber = "CUST-002"
	req2.TenureDays = 9 // matures 2026-10-10, within 30 days
	if _, err := svc.Create(req2); err != nil {
		t.Fatal(err)
	}
	req3 := baseRequest()
	req3.CustomerName = "PQR"
	req3.CustomerNumber = "CUST-003"
	fd3, err := svc.Create(req3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Close(api.CloseRequest{
		FDNumber: fd3.FDNumber, ClosureDate: "2026-10-05", Remark: "closed early",
	}); err != nil {
		t.Fatal(err)
	}

	t.Run("search by fd number", func(t *testing.T) {
		res, err := svc.List(api.ListRequest{Search: "FD-26-001", Filter: "ALL"})
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != 1 || res.Items[0].FDNumber != "FD-26-001" {
			t.Errorf("total=%d items=%+v", res.Total, res.Items)
		}
	})

	t.Run("search by customer name is case insensitive", func(t *testing.T) {
		res, err := svc.List(api.ListRequest{Search: "xyz"})
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != 1 {
			t.Errorf("total=%d, want 1", res.Total)
		}
	})

	t.Run("search by customer number", func(t *testing.T) {
		res, err := svc.List(api.ListRequest{Search: "CUST-003"})
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != 1 {
			t.Errorf("total=%d, want 1", res.Total)
		}
	})

	t.Run("active filter", func(t *testing.T) {
		res, err := svc.List(api.ListRequest{Filter: domain.FilterActive})
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != 2 {
			t.Errorf("active total=%d, want 2", res.Total)
		}
	})

	t.Run("closed filter", func(t *testing.T) {
		res, err := svc.List(api.ListRequest{Filter: domain.FilterClosed})
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != 1 {
			t.Errorf("closed total=%d, want 1", res.Total)
		}
	})

	t.Run("maturing filter excludes closed and long dated", func(t *testing.T) {
		res, err := svc.List(api.ListRequest{Filter: domain.FilterMaturing})
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != 1 || res.Items[0].FDNumber != "FD-26-002" {
			t.Errorf("maturing total=%d items=%+v", res.Total, res.Items)
		}
	})

	t.Run("invalid filter rejected", func(t *testing.T) {
		if _, err := svc.List(api.ListRequest{Filter: "NONSENSE"}); !errors.Is(err, domain.ErrInvalidFilter) {
			t.Errorf("err=%v, want ErrInvalidFilter", err)
		}
	})

	t.Run("pagination totals", func(t *testing.T) {
		res, err := svc.List(api.ListRequest{Page: 1, PageSize: 2})
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != 3 || len(res.Items) != 2 {
			t.Errorf("total=%d len=%d, want 3/2", res.Total, len(res.Items))
		}
	})
}

func TestDashboardStats(t *testing.T) {
	svc := newTestService(t)

	if _, err := svc.Create(baseRequest()); err != nil {
		t.Fatal(err)
	}
	req2 := baseRequest()
	req2.TenureDays = 9 // matures 2026-10-10 (within 7 days of 2026-10-06)
	fd2, err := svc.Create(req2)
	if err != nil {
		t.Fatal(err)
	}

	stats, err := svc.Dashboard()
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalFDs != 2 || stats.ActiveFDs != 2 {
		t.Errorf("counts=%+v, want 2/2", stats)
	}
	if stats.ActivePrincipal != 200000 {
		t.Errorf("principal=%d, want 200000", stats.ActivePrincipal)
	}
	if stats.TotalInterest != fd2.MaturityAmount-100000+8000 {
		t.Errorf("totalInterest=%d, want %d", stats.TotalInterest, fd2.MaturityAmount-100000+8000)
	}
	if stats.FYLabel != "FY 2026-27" {
		t.Errorf("fyLabel=%q, want FY 2026-27", stats.FYLabel)
	}
	if stats.FYDeposits != 200000 {
		t.Errorf("fyDeposits=%d, want 200000 (both started 2026-10-01)", stats.FYDeposits)
	}
	if stats.MaturingToday != 0 {
		t.Errorf("today=%d, want 0", stats.MaturingToday)
	}
	if stats.Maturing7 != 1 || stats.Maturing30 != 1 || stats.Maturing90 != 1 {
		t.Errorf("windows=%d/%d/%d, want 1/1/1", stats.Maturing7, stats.Maturing30, stats.Maturing90)
	}
}

func TestUpcomingMaturities(t *testing.T) {
	svc := newTestService(t)

	if _, err := svc.Create(baseRequest()); err != nil {
		t.Fatal(err)
	}
	req2 := baseRequest()
	req2.TenureDays = 9
	if _, err := svc.Create(req2); err != nil {
		t.Fatal(err)
	}

	upcoming, err := svc.Upcoming(api.UpcomingRequest{})
	if err != nil {
		t.Fatal(err)
	}
	// Default window is today .. +90 days: only FD-26-002 (matures 2026-10-10)
	// falls inside; FD-26-001 matures 2027-10-01, far beyond it.
	if len(upcoming) != 1 {
		t.Fatalf("default window len=%d, want 1", len(upcoming))
	}
	if upcoming[0].FDNumber != "FD-26-002" || upcoming[0].DaysRemaining != 4 {
		t.Errorf("first=%+v, want FD-26-002 with 4 days remaining", upcoming[0])
	}

	// An explicit range can reach further out (SRS §28 custom range).
	upcoming, err = svc.Upcoming(api.UpcomingRequest{FromDate: "2026-10-06", ToDate: "2027-12-31"})
	if err != nil {
		t.Fatal(err)
	}
	if len(upcoming) != 2 {
		t.Fatalf("custom range len=%d, want 2", len(upcoming))
	}
	if upcoming[0].FDNumber != "FD-26-002" || upcoming[1].FDNumber != "FD-26-001" {
		t.Errorf("order=%s,%s, want FD-26-002,FD-26-001", upcoming[0].FDNumber, upcoming[1].FDNumber)
	}
}

func TestRenewPrincipalOnly(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatal(err)
	}

	res, err := svc.Renew(api.RenewRequest{
		FDNumber:   created.FDNumber,
		Mode:       domain.RenewPrincipalOnly,
		TenureDays: 365,
	})
	if err != nil {
		t.Fatalf("renew: %v", err)
	}

	if res.PreviousFD.Status != domain.StatusClosed {
		t.Errorf("previous status=%s, want CLOSED", res.PreviousFD.Status)
	}
	if res.PreviousFD.ClosureType == nil || *res.PreviousFD.ClosureType != domain.ClosureRenewed {
		t.Errorf("closure type=%v, want RENEWED", res.PreviousFD.ClosureType)
	}
	if res.NewFD.Principal != 100000 {
		t.Errorf("new principal=%d, want 100000 (interest not carried)", res.NewFD.Principal)
	}
	if res.NewFD.FDNumber == created.FDNumber {
		t.Error("renewal must create a new FD number")
	}
	if res.NewFD.RenewedFrom == nil || *res.NewFD.RenewedFrom != created.FDNumber {
		t.Errorf("renewedFrom=%v, want %s", res.NewFD.RenewedFrom, created.FDNumber)
	}
	if res.PreviousFD.RenewedTo == nil || *res.PreviousFD.RenewedTo != res.NewFD.FDNumber {
		t.Errorf("renewedTo=%v, want %s", res.PreviousFD.RenewedTo, res.NewFD.FDNumber)
	}

	// New FD starts at the old maturity date by default (SRS §22).
	if res.NewFD.StartDate != "2027-10-01" {
		t.Errorf("new start=%s, want 2027-10-01", res.NewFD.StartDate)
	}
	if res.NewFD.FDNumber != "FD-27-001" {
		t.Errorf("new number=%s, want FD-27-001", res.NewFD.FDNumber)
	}

	// History of the old FD contains OPEN + RENEW.
	detail, err := svc.Get(created.FDNumber)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.History) != 2 {
		t.Fatalf("old history=%d entries, want 2", len(detail.History))
	}
	if detail.History[1].EventType != domain.EventRenew || detail.History[1].ReferenceFD != res.NewFD.FDNumber {
		t.Errorf("old history[1]=%+v, want RENEW referencing new FD", detail.History[1])
	}
}

func TestRenewPrincipalPlusInterest(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatal(err)
	}

	res, err := svc.Renew(api.RenewRequest{
		FDNumber:   created.FDNumber,
		Mode:       domain.RenewPrincipalPlusInterest,
		TenureDays: 365,
	})
	if err != nil {
		t.Fatalf("renew: %v", err)
	}
	if res.NewFD.Principal != 108000 {
		t.Errorf("new principal=%d, want 108000 (principal + interest)", res.NewFD.Principal)
	}
	if res.NewFD.MaturityAmount != 108000+8640 {
		t.Errorf("new maturity=%d, want 116640", res.NewFD.MaturityAmount)
	}
}

func TestRenewValidation(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("unknown fd", func(t *testing.T) {
		_, err := svc.Renew(api.RenewRequest{FDNumber: "FD-99-999", Mode: domain.RenewPrincipalOnly, TenureDays: 365})
		if !errors.Is(err, domain.ErrFDNotFound) {
			t.Errorf("err=%v, want ErrFDNotFound", err)
		}
	})

	t.Run("bad mode", func(t *testing.T) {
		_, err := svc.Renew(api.RenewRequest{FDNumber: created.FDNumber, Mode: "HALF", TenureDays: 365})
		if !errors.Is(err, domain.ErrInvalidRenewalMode) {
			t.Errorf("err=%v, want ErrInvalidRenewalMode", err)
		}
	})

	t.Run("bad tenure", func(t *testing.T) {
		_, err := svc.Renew(api.RenewRequest{FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly, TenureDays: 0})
		if !errors.Is(err, domain.ErrInvalidTenure) {
			t.Errorf("err=%v, want ErrInvalidTenure", err)
		}
	})

	t.Run("start before fd start", func(t *testing.T) {
		_, err := svc.Renew(api.RenewRequest{
			FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly,
			StartDate: "2026-01-01", TenureDays: 365,
		})
		if !errors.Is(err, domain.ErrClosureBeforeStart) {
			t.Errorf("err=%v, want ErrClosureBeforeStart", err)
		}
	})

	t.Run("renew twice blocked", func(t *testing.T) {
		res, err := svc.Renew(api.RenewRequest{FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly, TenureDays: 365})
		if err != nil {
			t.Fatal(err)
		}
		_, err = svc.Renew(api.RenewRequest{FDNumber: res.PreviousFD.FDNumber, Mode: domain.RenewPrincipalOnly, TenureDays: 365})
		if !errors.Is(err, domain.ErrClosedCannotRenew) {
			t.Errorf("err=%v, want ErrClosedCannotRenew", err)
		}
	})
}

func TestClosePremature(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatal(err)
	}

	preview, err := svc.PreviewClosure(api.CloseRequest{
		FDNumber: created.FDNumber, ClosureDate: "2026-10-05", Remark: "need money",
	})
	if err != nil {
		t.Fatalf("preview closure: %v", err)
	}
	if !preview.IsPremature {
		t.Error("closure on 2026-10-05 must be premature (maturity 2027-10-01)")
	}
	if preview.DaysHeld != 4 {
		t.Errorf("daysHeld=%d, want 4", preview.DaysHeld)
	}
	if preview.RatePercent != 4.00 {
		t.Errorf("rate=%v, want 4.00 (slab for 4 days)", preview.RatePercent)
	}
	// 100000 x 4 x 4 / 365 / 100 = 43.83 → 44
	if preview.Interest != 44 {
		t.Errorf("interest=%d, want 44", preview.Interest)
	}
	if preview.Payable != 100044 {
		t.Errorf("payable=%d, want 100044", preview.Payable)
	}

	closed, err := svc.Close(api.CloseRequest{
		FDNumber: created.FDNumber, ClosureDate: "2026-10-05", Remark: "need money",
	})
	if err != nil {
		t.Fatalf("close: %v", err)
	}
	if closed.Status != domain.StatusClosed {
		t.Errorf("status=%s, want CLOSED", closed.Status)
	}
	if closed.ClosureType == nil || *closed.ClosureType != domain.ClosurePremature {
		t.Errorf("closure type=%v, want PREMATURE", closed.ClosureType)
	}

	detail, err := svc.Get(created.FDNumber)
	if err != nil {
		t.Fatal(err)
	}
	last := detail.History[len(detail.History)-1]
	if last.EventType != domain.EventClose || last.Amount != 100044 {
		t.Errorf("last history=%+v, want CLOSE 100044", last)
	}
}

func TestCloseAtMaturityPaysMaturityAmount(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest()) // matures 2027-10-01
	if err != nil {
		t.Fatal(err)
	}

	closed, err := svc.Close(api.CloseRequest{
		FDNumber: created.FDNumber, ClosureDate: "2027-11-15", Remark: "after maturity",
	})
	if err != nil {
		t.Fatalf("close: %v", err)
	}
	if closed.ClosurePayable == nil || *closed.ClosurePayable != 108000 {
		t.Errorf("payable=%v, want 108000 (capped)", closed.ClosurePayable)
	}
	if closed.ClosureType == nil || *closed.ClosureType != domain.ClosureMatured {
		t.Errorf("type=%v, want MATURED", closed.ClosureType)
	}
}

func TestCloseValidation(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		req  api.CloseRequest
		want error
	}{
		{"missing date", api.CloseRequest{FDNumber: created.FDNumber, Remark: "x"}, domain.ErrClosureDateRequired},
		{"bad date", api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "not-a-date", Remark: "x"}, domain.ErrInvalidDate},
		{"missing remark", api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-10-05"}, domain.ErrClosureRemarkRequired},
		{"before start", api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-09-01", Remark: "x"}, domain.ErrClosureBeforeStart},
		{"unknown fd", api.CloseRequest{FDNumber: "FD-99-999", ClosureDate: "2026-10-05", Remark: "x"}, domain.ErrFDNotFound},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := svc.Close(c.req); !errors.Is(err, c.want) {
				t.Errorf("close err=%v, want %v", err, c.want)
			}
			if _, err := svc.PreviewClosure(c.req); !errors.Is(err, c.want) {
				t.Errorf("preview err=%v, want %v", err, c.want)
			}
		})
	}

	// Close it once, then ensure double closure is blocked.
	if _, err := svc.Close(api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-10-05", Remark: "first"}); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Close(api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-10-06", Remark: "second"})
	if !errors.Is(err, domain.ErrAlreadyClosed) {
		t.Errorf("double close err=%v, want ErrAlreadyClosed", err)
	}
	_, err = svc.Renew(api.RenewRequest{FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly, TenureDays: 365})
	if !errors.Is(err, domain.ErrClosedCannotRenew) {
		t.Errorf("renew closed err=%v, want ErrClosedCannotRenew", err)
	}
}

func TestRateSlabSettings(t *testing.T) {
	svc := newTestService(t)

	slabs, err := svc.RateSlabs()
	if err != nil {
		t.Fatal(err)
	}
	if len(slabs) != 4 {
		t.Fatalf("default slabs=%d, want 4", len(slabs))
	}

	// Change the 365-729 slab from 8% to 7% and confirm it is used.
	for i := range slabs {
		if slabs[i].MinDays == 365 {
			slabs[i].RatePercent = 7.00
		}
	}
	if err := svc.SaveRateSlabs(slabs); err != nil {
		t.Fatalf("save slabs: %v", err)
	}

	res, err := svc.Preview(baseRequest())
	if err != nil {
		t.Fatal(err)
	}
	if res.RatePercent != 7.00 {
		t.Errorf("rate=%v, want updated 7.00", res.RatePercent)
	}
	if res.Interest != 7000 {
		t.Errorf("interest=%d, want 7000", res.Interest)
	}

	// Invalid configuration is rejected and does not change stored slabs.
	bad := []domain.RateSlab{{MinDays: 1, MaxDays: 500, RatePercent: 5, Label: "wide"}, {MinDays: 400, MaxDays: 600, RatePercent: 6, Label: "overlap"}}
	if err := svc.SaveRateSlabs(bad); !errors.Is(err, domain.ErrInvalidSlabConfig) {
		t.Errorf("bad slabs err=%v, want ErrInvalidSlabConfig", err)
	}
	after, err := svc.RateSlabs()
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 4 {
		t.Errorf("slabs changed after failed save: %d", len(after))
	}
}

func TestExportReportCreatesXlsx(t *testing.T) {
	svc := newTestService(t)

	if _, err := svc.Create(baseRequest()); err != nil {
		t.Fatal(err)
	}
	req2 := baseRequest()
	req2.CustomerName = "XYZ"
	closed, err := svc.Create(req2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Close(api.CloseRequest{FDNumber: closed.FDNumber, ClosureDate: "2026-10-05", Remark: "test"}); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	kinds := map[string]int{
		api.ReportRegister: 2,
		api.ReportActive:   1,
		api.ReportClosed:   1,
	}
	for kind, wantRows := range kinds {
		path := filepath.Join(dir, kind+".xlsx")
		res, err := svc.ExportReport(api.ReportRequest{Kind: kind, Path: path})
		if err != nil {
			t.Fatalf("export %s: %v", kind, err)
		}
		if res.RowCount != wantRows {
			t.Errorf("%s rows=%d, want %d", kind, res.RowCount, wantRows)
		}
		if _, err := os.Stat(res.Path); err != nil {
			t.Errorf("%s file missing: %v", kind, err)
		}
	}

	// Maturity report with an explicit range covering both FDs.
	res, err := svc.ExportReport(api.ReportRequest{
		Kind: api.ReportMaturity, FromDate: "2026-01-01", ToDate: "2027-12-31",
		Path: filepath.Join(dir, "maturity.xlsx"),
	})
	if err != nil {
		t.Fatalf("export maturity: %v", err)
	}
	if res.RowCount != 2 {
		t.Errorf("maturity rows=%d, want 2", res.RowCount)
	}

	// Path without extension is completed automatically.
	res, err = svc.ExportReport(api.ReportRequest{Kind: api.ReportActive, Path: filepath.Join(dir, "noext")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(res.Path, ".xlsx") {
		t.Errorf("path=%s, want .xlsx suffix", res.Path)
	}

	if _, err := svc.ExportReport(api.ReportRequest{Kind: "WEIRD", Path: filepath.Join(dir, "x.xlsx")}); !errors.Is(err, domain.ErrInvalidReport) {
		t.Errorf("bad kind err=%v, want ErrInvalidReport", err)
	}
	if _, err := svc.ExportReport(api.ReportRequest{Kind: api.ReportRegister, Path: "  "}); !errors.Is(err, domain.ErrExportPathRequired) {
		t.Errorf("no path err=%v, want ErrExportPathRequired", err)
	}
}

// TestRecommendedUserFlow walks the SRS §35 workflow end to end.
func TestRecommendedUserFlow(t *testing.T) {
	svc := newTestService(t)

	// Open Application → Dashboard
	stats, err := svc.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if stats.TotalFDs != 0 {
		t.Fatalf("expected empty dashboard, got %+v", stats)
	}

	// Create New FD → Enter Details → Calculate Preview → Confirm & Save
	req := baseRequest()
	preview, err := svc.Preview(req)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if preview.MaturityAmount != 108000 {
		t.Fatalf("preview maturity=%d, want 108000", preview.MaturityAmount)
	}
	created, err := svc.Create(req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Search / View Existing FD
	found, err := svc.List(api.ListRequest{Search: created.FDNumber})
	if err != nil || found.Total != 1 {
		t.Fatalf("search failed: err=%v total=%d", err, found.Total)
	}

	// View FD Details
	detail, err := svc.Get(created.FDNumber)
	if err != nil {
		t.Fatalf("details: %v", err)
	}
	if detail.FD.Status != domain.StatusActive || len(detail.History) != 1 {
		t.Fatalf("details=%+v", detail)
	}

	// Renew
	renewed, err := svc.Renew(api.RenewRequest{
		FDNumber: created.FDNumber, Mode: domain.RenewPrincipalPlusInterest, TenureDays: 365,
	})
	if err != nil {
		t.Fatalf("renew: %v", err)
	}
	if renewed.PreviousFD.Status != domain.StatusClosed || renewed.NewFD.Status != domain.StatusActive {
		t.Fatalf("renew result=%+v", renewed)
	}

	// Close the renewed FD at maturity
	closed, err := svc.Close(api.CloseRequest{
		FDNumber: renewed.NewFD.FDNumber, ClosureDate: "2028-10-01", Remark: "matured and withdrawn",
	})
	if err != nil {
		t.Fatalf("close: %v", err)
	}
	if closed.ClosurePayable == nil {
		t.Fatal("payable missing")
	}

	// Export Excel Report
	res, err := svc.ExportReport(api.ReportRequest{
		Kind: api.ReportRegister, Path: filepath.Join(t.TempDir(), "register.xlsx"),
	})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if res.RowCount != 2 {
		t.Errorf("register rows=%d, want 2", res.RowCount)
	}
}
