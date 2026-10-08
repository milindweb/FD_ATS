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

// defaultMemberID is the member created by newTestService; baseRequest and
// maturedRequest attach their FDs to it.
var defaultMemberID int64

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

	m, err := svc.Members().Save(api.SaveMemberRequest{GENNo: "CUST-001", Name: "ABC"})
	if err != nil {
		t.Fatalf("create default member: %v", err)
	}
	defaultMemberID = m.ID
	return svc
}

// memberFor creates an additional member and returns its ID.
func memberFor(t *testing.T, svc *service.FDService, gen, name string) int64 {
	t.Helper()
	m, err := svc.Members().Save(api.SaveMemberRequest{GENNo: gen, Name: name})
	if err != nil {
		t.Fatalf("create member %s: %v", gen, err)
	}
	return m.ID
}

func baseRequest() api.PreviewRequest {
	return api.PreviewRequest{
		MemberID:   defaultMemberID,
		Principal:  100000,
		StartDate:  "2026-10-01",
		TenureDays: 365,
	}
}

// maturedRequest is baseRequest shifted back one year: start 2025-10-01,
// matures 2026-10-01 — five days before fixedNow, so renewals are allowed.
func maturedRequest() api.PreviewRequest {
	req := baseRequest()
	req.StartDate = "2025-10-01"
	return req
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
	if first.FDNumber != "FD-1" {
		t.Errorf("first number=%s, want FD-1", first.FDNumber)
	}
	if first.Status != domain.StatusActive {
		t.Errorf("status=%s, want ACTIVE", first.Status)
	}

	req2 := baseRequest()
	req2.MemberID = memberFor(t, svc, "CUST-002", "PQR")
	second, err := svc.Create(req2)
	if err != nil {
		t.Fatalf("create 2: %v", err)
	}
	if second.FDNumber != "FD-2" {
		t.Errorf("second number=%s, want FD-2", second.FDNumber)
	}
}

// FD-N numbers come from one global sequence regardless of the start year
// (SRS §11: no year component).
func TestCreateUsesGlobalSequenceAcrossStartYears(t *testing.T) {
	svc := newTestService(t)

	req := baseRequest()
	req.StartDate = "2026-10-01"
	if _, err := svc.Create(req); err != nil {
		t.Fatalf("create 1: %v", err)
	}

	req2 := baseRequest()
	req2.StartDate = "2027-01-15"
	fd, err := svc.Create(req2)
	if err != nil {
		t.Fatalf("create 2: %v", err)
	}
	if fd.FDNumber != "FD-2" {
		t.Errorf("number=%s, want FD-2 (global sequence continues across years)", fd.FDNumber)
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
		{"missing member", func(r *api.PreviewRequest) { r.MemberID = 0 }, domain.ErrMemberRequired},
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

func TestEditFDUpdatesActiveFD(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatal(err)
	}

	edited, err := svc.EditFD(api.EditFDRequest{
		FDNumber:   created.FDNumber,
		MemberID:   memberFor(t, svc, "CUST-999", "Renamed Member"),
		Principal:  200000,
		StartDate:  "2026-10-01",
		TenureDays: 365,
	})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	if edited.CustomerName != "Renamed Member" || edited.CustomerNumber != "CUST-999" {
		t.Errorf("identity=%s/%s, want Renamed Member/CUST-999", edited.CustomerName, edited.CustomerNumber)
	}
	if edited.Principal != 200000 || edited.InterestAmount != 16000 || edited.MaturityAmount != 216000 {
		t.Errorf("amounts p=%d i=%d m=%d, want 200000/16000/216000",
			edited.Principal, edited.InterestAmount, edited.MaturityAmount)
	}
	if edited.Status != domain.StatusActive {
		t.Errorf("status=%s, want ACTIVE", edited.Status)
	}

	detail, err := svc.Get(created.FDNumber)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.History) != 2 {
		t.Fatalf("history=%d entries, want 2 (OPEN + EDIT)", len(detail.History))
	}
	last := detail.History[1]
	if last.EventType != domain.EventEdit || last.Remarks != "FD details edited" {
		t.Errorf("last history=%+v, want EDIT \"FD details edited\"", last)
	}
}

func TestEditFDBlockedOnClosedFD(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(maturedRequest())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Close(api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-10-01"}); err != nil {
		t.Fatalf("close: %v", err)
	}

	_, err = svc.EditFD(api.EditFDRequest{
		FDNumber: created.FDNumber, MemberID: defaultMemberID,
		Principal: 1, StartDate: "2026-10-01", TenureDays: 1,
	})
	if !errors.Is(err, domain.ErrEditClosed) {
		t.Errorf("err=%v, want ErrEditClosed", err)
	}
}

func TestEditFDValidation(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatal(err)
	}

	base := func() api.EditFDRequest {
		return api.EditFDRequest{
			FDNumber: created.FDNumber, MemberID: defaultMemberID,
			Principal: 100000, StartDate: "2026-10-01", TenureDays: 365,
		}
	}

	t.Run("unknown fd", func(t *testing.T) {
		req := base()
		req.FDNumber = "FD-99-999"
		if _, err := svc.EditFD(req); !errors.Is(err, domain.ErrFDNotFound) {
			t.Errorf("err=%v, want ErrFDNotFound", err)
		}
	})

	t.Run("member required", func(t *testing.T) {
		req := base()
		req.MemberID = 0
		if _, err := svc.EditFD(req); !errors.Is(err, domain.ErrMemberRequired) {
			t.Errorf("err=%v, want ErrMemberRequired", err)
		}
	})

	t.Run("member not found", func(t *testing.T) {
		req := base()
		req.MemberID = 99999
		if _, err := svc.EditFD(req); !errors.Is(err, domain.ErrMemberNotFound) {
			t.Errorf("err=%v, want ErrMemberNotFound", err)
		}
	})

	t.Run("amount", func(t *testing.T) {
		req := base()
		req.Principal = 0
		if _, err := svc.EditFD(req); !errors.Is(err, domain.ErrInvalidAmount) {
			t.Errorf("err=%v, want ErrInvalidAmount", err)
		}
	})

	t.Run("bad date", func(t *testing.T) {
		req := base()
		req.StartDate = "not-a-date"
		if _, err := svc.EditFD(req); !errors.Is(err, domain.ErrInvalidDate) {
			t.Errorf("err=%v, want ErrInvalidDate", err)
		}
	})

	t.Run("tenure", func(t *testing.T) {
		req := base()
		req.TenureDays = 0
		if _, err := svc.EditFD(req); !errors.Is(err, domain.ErrInvalidTenure) {
			t.Errorf("err=%v, want ErrInvalidTenure", err)
		}
	})
}

func TestListSearchAndFilters(t *testing.T) {
	svc := newTestService(t)

	// FD-1: ABC, active, matures 2027-10-01.
	// FD-2: XYZ, active, short tenure matures 2026-10-10 (maturing soon).
	// FD-3: PQR, closed early.
	req1 := baseRequest()
	if _, err := svc.Create(req1); err != nil {
		t.Fatal(err)
	}
	req2 := baseRequest()
	req2.MemberID = memberFor(t, svc, "CUST-002", "XYZ")
	req2.TenureDays = 9 // matures 2026-10-10, within 30 days
	if _, err := svc.Create(req2); err != nil {
		t.Fatal(err)
	}
	req3 := baseRequest()
	req3.MemberID = memberFor(t, svc, "CUST-003", "PQR")
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
		res, err := svc.List(api.ListRequest{Search: "FD-1", Filter: "ALL"})
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != 1 || res.Items[0].FDNumber != "FD-1" {
			t.Errorf("total=%d items=%+v", res.Total, res.Items)
		}
	})

	t.Run("search by member name is case insensitive", func(t *testing.T) {
		res, err := svc.List(api.ListRequest{Search: "xyz"})
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != 1 {
			t.Errorf("total=%d, want 1", res.Total)
		}
	})

	t.Run("search by member GEN number", func(t *testing.T) {
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
		if res.Total != 1 || res.Items[0].FDNumber != "FD-2" {
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
	// Default window is today .. +90 days: only FD-2 (matures 2026-10-10)
	// falls inside; FD-1 matures 2027-10-01, far beyond it.
	if len(upcoming) != 1 {
		t.Fatalf("default window len=%d, want 1", len(upcoming))
	}
	if upcoming[0].FDNumber != "FD-2" || upcoming[0].DaysRemaining != 4 {
		t.Errorf("first=%+v, want FD-2 with 4 days remaining", upcoming[0])
	}

	// An explicit range can reach further out (SRS §28 custom range).
	upcoming, err = svc.Upcoming(api.UpcomingRequest{FromDate: "2026-10-06", ToDate: "2027-12-31"})
	if err != nil {
		t.Fatal(err)
	}
	if len(upcoming) != 2 {
		t.Fatalf("custom range len=%d, want 2", len(upcoming))
	}
	if upcoming[0].FDNumber != "FD-2" || upcoming[1].FDNumber != "FD-1" {
		t.Errorf("order=%s,%s, want FD-2,FD-1", upcoming[0].FDNumber, upcoming[1].FDNumber)
	}
}

func TestRenewPrincipalOnly(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(maturedRequest())
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
	if res.NewFD.StartDate != "2026-10-01" {
		t.Errorf("new start=%s, want 2026-10-01", res.NewFD.StartDate)
	}
	if res.NewFD.FDNumber != "FD-2" {
		t.Errorf("new number=%s, want FD-2", res.NewFD.FDNumber)
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

	created, err := svc.Create(maturedRequest())
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

	created, err := svc.Create(maturedRequest())
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
			StartDate: "2025-05-01", TenureDays: 365,
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

func TestRenewBeforeMaturityBlocked(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(maturedRequest()) // start 2025-10-01, matures 2026-10-01
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Renew(api.RenewRequest{
		FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly,
		StartDate: "2026-09-30", TenureDays: 365,
	})
	if !errors.Is(err, domain.ErrRenewBeforeMaturity) {
		t.Errorf("err=%v, want ErrRenewBeforeMaturity", err)
	}

	if _, err := svc.Get(created.FDNumber); err != nil {
		t.Fatalf("fd must stay active after rejected renewal: %v", err)
	}
}

func TestRenewOnOrAfterMaturityAllowed(t *testing.T) {
	svc := newTestService(t)

	for _, start := range []string{"2026-10-01", "2026-12-01"} {
		t.Run("start "+start, func(t *testing.T) {
			created, err := svc.Create(maturedRequest()) // matures 2026-10-01
			if err != nil {
				t.Fatal(err)
			}
			res, err := svc.Renew(api.RenewRequest{
				FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly,
				StartDate: start, TenureDays: 365,
			})
			if err != nil {
				t.Fatalf("renew at %s: %v", start, err)
			}
			if res.NewFD.StartDate != start {
				t.Errorf("new start=%s, want %s", res.NewFD.StartDate, start)
			}
			if res.PreviousFD.Status != domain.StatusClosed {
				t.Errorf("previous status=%s, want CLOSED", res.PreviousFD.Status)
			}
		})
	}
}

func TestRenewBackdatedOnMaturityAllowed(t *testing.T) {
	svc := newTestService(t)

	req := baseRequest()
	req.StartDate = "2025-01-01"
	req.TenureDays = 365 // matures 2026-01-01, already past
	created, err := svc.Create(req)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Renew(api.RenewRequest{
		FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly,
		StartDate: "2025-12-30", TenureDays: 365,
	})
	if !errors.Is(err, domain.ErrRenewBeforeMaturity) {
		t.Errorf("start after fd start but before maturity err=%v, want ErrRenewBeforeMaturity", err)
	}

	res, err := svc.Renew(api.RenewRequest{
		FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly,
		StartDate: "2026-01-01", TenureDays: 365,
	})
	if err != nil {
		t.Fatalf("back-dated renewal on maturity: %v", err)
	}
	if res.NewFD.StartDate != "2026-01-01" {
		t.Errorf("new start=%s, want 2026-01-01", res.NewFD.StartDate)
	}
}

func TestRenewBlockedUntilMaturity(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest()) // start 2026-10-01, matures 2027-10-01 (future)
	if err != nil {
		t.Fatal(err)
	}

	// Even with a valid start date on the maturity date, the renewal itself
	// must not be possible before the maturity date has been reached.
	_, err = svc.Renew(api.RenewRequest{
		FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly,
		StartDate: "2027-10-01", TenureDays: 365,
	})
	if !errors.Is(err, domain.ErrRenewNotMatured) {
		t.Errorf("err=%v, want ErrRenewNotMatured", err)
	}

	if _, err := svc.Get(created.FDNumber); err != nil {
		t.Fatalf("fd must stay active after rejected renewal: %v", err)
	}
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
		name        string
		req         api.CloseRequest
		want        error // Close expectation
		previewWant error // PreviewClosure expectation; nil = preview must succeed
	}{
		{"missing date", api.CloseRequest{FDNumber: created.FDNumber, Remark: "x"}, domain.ErrClosureDateRequired, domain.ErrClosureDateRequired},
		{"bad date", api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "not-a-date", Remark: "x"}, domain.ErrInvalidDate, domain.ErrInvalidDate},
		{"missing remark on premature close", api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-10-05"}, domain.ErrClosureRemarkRequired, nil},
		{"before start", api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-09-01", Remark: "x"}, domain.ErrClosureBeforeStart, domain.ErrClosureBeforeStart},
		{"unknown fd", api.CloseRequest{FDNumber: "FD-99-999", ClosureDate: "2026-10-05", Remark: "x"}, domain.ErrFDNotFound, domain.ErrFDNotFound},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := svc.Close(c.req); !errors.Is(err, c.want) {
				t.Errorf("close err=%v, want %v", err, c.want)
			}
			_, perr := svc.PreviewClosure(c.req)
			if c.previewWant == nil {
				if perr != nil {
					t.Errorf("preview err=%v, want success (preview never requires a remark)", perr)
				}
			} else if !errors.Is(perr, c.previewWant) {
				t.Errorf("preview err=%v, want %v", perr, c.previewWant)
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

func TestCloseMaturedWithoutRemark(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest()) // matures 2027-10-01
	if err != nil {
		t.Fatal(err)
	}

	// A matured payout needs no remark: preview and close both succeed.
	if _, err := svc.PreviewClosure(api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2027-10-01"}); err != nil {
		t.Errorf("preview without remark: %v, want success", err)
	}
	closed, err := svc.Close(api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2027-10-01"})
	if err != nil {
		t.Fatalf("close matured without remark: %v, want success", err)
	}
	if closed.ClosureType == nil || *closed.ClosureType != domain.ClosureMatured {
		t.Errorf("type=%v, want MATURED", closed.ClosureType)
	}
}

func TestReopenClosedFD(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Reopen(api.ReopenRequest{FDNumber: created.FDNumber, Remark: "x"}); !errors.Is(err, domain.ErrNotClosed) {
		t.Errorf("reopen active fd err=%v, want ErrNotClosed", err)
	}

	if _, err := svc.Close(api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-10-05", Remark: "need money"}); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Reopen(api.ReopenRequest{FDNumber: created.FDNumber}); !errors.Is(err, domain.ErrReversalReasonRequired) {
		t.Errorf("reopen without reason err=%v, want ErrReversalReasonRequired", err)
	}

	reopened, err := svc.Reopen(api.ReopenRequest{FDNumber: created.FDNumber, Remark: "wrong closure date"})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if reopened.Status != domain.StatusActive {
		t.Errorf("status=%s, want ACTIVE", reopened.Status)
	}
	if reopened.ClosureDate != nil || reopened.ClosureType != nil || reopened.ClosurePayable != nil {
		t.Errorf("closure fields not cleared: %+v", reopened)
	}
	if reopened.ClosureRemark != "" {
		t.Errorf("closure remark=%q, want empty", reopened.ClosureRemark)
	}

	detail, err := svc.Get(created.FDNumber)
	if err != nil {
		t.Fatal(err)
	}
	last := detail.History[len(detail.History)-1]
	if last.EventType != domain.EventReopen || last.Remarks != "wrong closure date" {
		t.Errorf("last history=%+v, want REOPEN with reason", last)
	}

	if _, err := svc.Close(api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-10-06", Remark: "second try"}); err != nil {
		t.Errorf("close after reopen: %v", err)
	}
}

func TestReopenRenewedFDBlocked(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(maturedRequest())
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.Renew(api.RenewRequest{FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly, TenureDays: 365})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Reopen(api.ReopenRequest{FDNumber: res.PreviousFD.FDNumber, Remark: "undo renewal"})
	if !errors.Is(err, domain.ErrRenewedCannotReopen) {
		t.Errorf("reopen renewed fd err=%v, want ErrRenewedCannotReopen", err)
	}
}

func TestReverseRenewal(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(maturedRequest())
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.Renew(api.RenewRequest{FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly, TenureDays: 365})
	if err != nil {
		t.Fatal(err)
	}
	newNumber := res.NewFD.FDNumber

	if _, err := svc.ReverseRenewal(api.ReverseRenewalRequest{FDNumber: res.PreviousFD.FDNumber}); !errors.Is(err, domain.ErrReversalReasonRequired) {
		t.Errorf("reverse without reason err=%v, want ErrReversalReasonRequired", err)
	}

	old, err := svc.ReverseRenewal(api.ReverseRenewalRequest{FDNumber: res.PreviousFD.FDNumber, Remark: "renewed by mistake"})
	if err != nil {
		t.Fatalf("reverse renewal: %v", err)
	}
	if old.Status != domain.StatusActive {
		t.Errorf("old status=%s, want ACTIVE", old.Status)
	}
	if old.ClosureType != nil || old.RenewedTo != nil {
		t.Errorf("old closure/renewal links not cleared: %+v", old)
	}

	if _, err := svc.Get(newNumber); !errors.Is(err, domain.ErrFDNotFound) {
		t.Errorf("withdrawn fd err=%v, want ErrFDNotFound", err)
	}

	detail, err := svc.Get(old.FDNumber)
	if err != nil {
		t.Fatal(err)
	}
	last := detail.History[len(detail.History)-1]
	if last.EventType != domain.EventReverse || last.ReferenceFD != newNumber {
		t.Errorf("last history=%+v, want REVERSE referencing %s", last, newNumber)
	}

	if _, err := svc.Renew(api.RenewRequest{FDNumber: old.FDNumber, Mode: domain.RenewPrincipalOnly, TenureDays: 365}); err != nil {
		t.Errorf("renew after reversal: %v", err)
	}
}

func TestReverseRenewalBlockedAfterNewFDUsed(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(maturedRequest())
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.Renew(api.RenewRequest{FDNumber: created.FDNumber, Mode: domain.RenewPrincipalOnly, TenureDays: 365})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Close(api.CloseRequest{FDNumber: res.NewFD.FDNumber, ClosureDate: "2027-10-01", Remark: "paid out"}); err != nil {
		t.Fatal(err)
	}

	_, err = svc.ReverseRenewal(api.ReverseRenewalRequest{FDNumber: res.PreviousFD.FDNumber, Remark: "undo"})
	if !errors.Is(err, domain.ErrRenewalAlreadyUsed) {
		t.Errorf("err=%v, want ErrRenewalAlreadyUsed", err)
	}
}

func TestReverseRenewalOnNonRenewedFDBlocked(t *testing.T) {
	svc := newTestService(t)

	created, err := svc.Create(baseRequest())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.ReverseRenewal(api.ReverseRenewalRequest{FDNumber: created.FDNumber, Remark: "x"}); !errors.Is(err, domain.ErrNotRenewed) {
		t.Errorf("active fd err=%v, want ErrNotRenewed", err)
	}

	if _, err := svc.Close(api.CloseRequest{FDNumber: created.FDNumber, ClosureDate: "2026-10-05", Remark: "need money"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReverseRenewal(api.ReverseRenewalRequest{FDNumber: created.FDNumber, Remark: "x"}); !errors.Is(err, domain.ErrNotRenewed) {
		t.Errorf("premature-closed fd err=%v, want ErrNotRenewed", err)
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
	req2.MemberID = memberFor(t, svc, "CUST-002", "XYZ")
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
		api.ReportMembers:  2, // default member + XYZ; no unassigned FDs
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
	req := maturedRequest()
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
