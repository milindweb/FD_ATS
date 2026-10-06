package service_test

import (
	"errors"
	"testing"

	"fdats/internal/api"
	"fdats/internal/domain"
)

// LoadSampleData builds 12 FDs, closes two (one premature, one at maturity)
// and renews one more, so the database ends up with 13 FD rows.
func TestLoadSampleDataSeedsOnlyEmptyDatabase(t *testing.T) {
	svc := newTestService(t)

	count, err := svc.LoadSampleData()
	if err != nil {
		t.Fatalf("load sample: %v", err)
	}
	if count != 13 {
		t.Errorf("count=%d, want 13 FDs (12 created + 1 renewal)", count)
	}

	if _, err := svc.LoadSampleData(); !errors.Is(err, domain.ErrSampleDataNotEmpty) {
		t.Errorf("second load err=%v, want ErrSampleDataNotEmpty", err)
	}

	list, err := svc.List(api.ListRequest{Filter: "ALL", Page: 1, PageSize: 100})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	var active, closed, renewedFrom, overdue, maturingWeek int
	for _, fd := range list.Items {
		switch fd.Status {
		case domain.StatusActive:
			active++
			if fd.RenewedFrom != nil {
				renewedFrom++
			}
			if fd.MaturityDate < "2026-10-06" {
				overdue++ // matured but still active
			}
			if fd.MaturityDate >= "2026-10-06" && fd.MaturityDate <= "2026-10-13" {
				maturingWeek++
			}
		case domain.StatusClosed:
			closed++
		}
	}
	if active != 10 || closed != 3 {
		t.Errorf("active=%d closed=%d, want 10/3", active, closed)
	}
	if renewedFrom != 1 {
		t.Errorf("renewed new FDs=%d, want 1", renewedFrom)
	}
	if overdue != 1 {
		t.Errorf("overdue active FDs=%d, want 1", overdue)
	}
	if maturingWeek != 2 {
		t.Errorf("maturing within 7 days=%d, want 2", maturingWeek)
	}
}

// The preview carries the same headings and rows the export writes.
func TestPreviewReportReturnsSheetPreview(t *testing.T) {
	svc := newTestService(t)

	if _, err := svc.Create(baseRequest()); err != nil {
		t.Fatalf("create 1: %v", err)
	}
	second := baseRequest()
	second.CustomerName = "PQR"
	if _, err := svc.Create(second); err != nil {
		t.Fatalf("create 2: %v", err)
	}

	prev, err := svc.PreviewReport(api.ReportRequest{Kind: api.ReportRegister})
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if prev.Kind != api.ReportRegister {
		t.Errorf("kind=%s, want REGISTER", prev.Kind)
	}
	if len(prev.Headers) == 0 {
		t.Fatal("headers empty")
	}
	if prev.Total != 2 || len(prev.Rows) != 2 {
		t.Errorf("total=%d rows=%d, want 2/2", prev.Total, len(prev.Rows))
	}
	for i, row := range prev.Rows {
		if len(row) != len(prev.Headers) {
			t.Errorf("row %d has %d cells, headers have %d", i, len(row), len(prev.Headers))
		}
	}
	if prev.Rows[0][0] != "FD-26-001" {
		t.Errorf("first cell=%q, want FD-26-001", prev.Rows[0][0])
	}

	if _, err := svc.PreviewReport(api.ReportRequest{Kind: "NOPE"}); !errors.Is(err, domain.ErrInvalidReport) {
		t.Errorf("invalid kind err=%v, want ErrInvalidReport", err)
	}
}

// The chart buckets active maturities into the current and next month.
func TestMaturityChartBucketsByMonth(t *testing.T) {
	svc := newTestService(t)

	near := baseRequest()
	near.TenureDays = 5 // maturity 2026-10-06 (this month)
	if _, err := svc.Create(near); err != nil {
		t.Fatalf("create near: %v", err)
	}
	far := baseRequest()
	far.CustomerName = "NEXT"
	far.StartDate = "2026-09-30"
	far.TenureDays = 60 // maturity 2026-11-29 (next month)
	if _, err := svc.Create(far); err != nil {
		t.Fatalf("create far: %v", err)
	}

	buckets, err := svc.MaturityChart()
	if err != nil {
		t.Fatalf("chart: %v", err)
	}
	if len(buckets) != 12 {
		t.Fatalf("buckets=%d, want 12", len(buckets))
	}
	if buckets[0].Period != "2026-10" || buckets[1].Period != "2026-11" {
		t.Errorf("periods=%s,%s, want 2026-10,2026-11", buckets[0].Period, buckets[1].Period)
	}
	if buckets[0].Count != 1 || buckets[1].Count != 1 {
		t.Errorf("counts=%d,%d, want 1,1", buckets[0].Count, buckets[1].Count)
	}
	if buckets[0].Amount <= 0 || buckets[1].Amount <= 0 {
		t.Errorf("amounts=%d,%d, want both > 0", buckets[0].Amount, buckets[1].Amount)
	}
	for i, b := range buckets[2:] {
		if b.Count != 0 {
			t.Errorf("bucket %d count=%d, want 0", i+2, b.Count)
		}
	}
}
