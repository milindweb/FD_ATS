package report_test

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"fdats/internal/api"
	"fdats/internal/domain"
	"fdats/internal/report"
)

func sampleFDs() []domain.FixedDeposit {
	closureDate := "2026-10-05"
	closureType := domain.ClosurePremature
	payable := int64(100044)
	return []domain.FixedDeposit{
		{
			FDNumber: "FD-1", CustomerName: "ABC", CustomerNumber: "CUST-001",
			Principal: 100000, StartDate: "2026-10-01", TenureDays: 365, InterestRate: 8.00,
			MaturityDate: "2027-10-01", InterestAmount: 8000, MaturityAmount: 108000,
			Status: domain.StatusActive,
		},
		{
			FDNumber: "FD-2", CustomerName: "XYZ", CustomerNumber: "CUST-002",
			Principal: 50000, StartDate: "2026-01-01", TenureDays: 100, InterestRate: 4.00,
			MaturityDate: "2026-04-11", InterestAmount: 548, MaturityAmount: 50548,
			Status: domain.StatusClosed, ClosureDate: &closureDate, ClosureType: &closureType,
			ClosurePayable: &payable, ClosureRemark: "early withdrawal",
		},
	}
}

func export(t *testing.T, kind string, fds []domain.FixedDeposit) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), kind+".xlsx")
	err := report.Export(report.Options{
		Kind:        kind,
		Title:       "Test Report",
		FDs:         fds,
		GeneratedAt: time.Date(2026, 10, 6, 12, 0, 0, 0, time.Local),
		Path:        path,
	})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	return path
}

func readSheet(t *testing.T, path string) *excelize.File {
	t.Helper()
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open exported file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func cell(t *testing.T, f *excelize.File, ref string) string {
	t.Helper()
	v, err := f.GetCellValue("Report", ref)
	if err != nil {
		t.Fatalf("read %s: %v", ref, err)
	}
	return v
}

// numCell reads a money cell as an integer, tolerating thousands separators
// introduced by the cell number format.
func numCell(t *testing.T, f *excelize.File, ref string) int64 {
	t.Helper()
	raw := strings.NewReplacer(",", "", " ", "").Replace(cell(t, f, ref))
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		t.Fatalf("parse %s=%q: %v", ref, raw, err)
	}
	return v
}

func TestRegisterReportContents(t *testing.T) {
	path := export(t, api.ReportRegister, sampleFDs())
	f := readSheet(t, path)

	if got := cell(t, f, "A1"); got != "Test Report" {
		t.Errorf("title=%q, want Test Report", got)
	}
	if got := cell(t, f, "A2"); got != "Generated: 06/10/2026" {
		t.Errorf("generated=%q", got)
	}

	wantHeaders := []string{"A4", "B4", "C4", "D4", "E4", "F4", "G4", "H4", "I4", "J4", "K4"}
	wantValues := []string{"FD Number", "Member Name", "GEN No.", "Deposit Amount",
		"Interest Rate", "Start Date", "Tenure", "Maturity Date", "Interest Amount", "Maturity Amount", "Status"}
	for i, ref := range wantHeaders {
		if got := cell(t, f, ref); got != wantValues[i] {
			t.Errorf("%s=%q, want %q", ref, got, wantValues[i])
		}
	}

	if got := cell(t, f, "A5"); got != "FD-1" {
		t.Errorf("row1 fd=%q", got)
	}
	if got := numCell(t, f, "D5"); got != 100000 {
		t.Errorf("row1 principal=%d, want 100000", got)
	}
	if got := numCell(t, f, "J5"); got != 108000 {
		t.Errorf("row1 maturity amount=%d, want 108000", got)
	}
	if got := cell(t, f, "A6"); got != "FD-2" {
		t.Errorf("row2 fd=%q", got)
	}
}

func TestClosedReportHasClosureColumns(t *testing.T) {
	path := export(t, api.ReportClosed, sampleFDs())
	f := readSheet(t, path)

	if got := cell(t, f, "L4"); got != "Closure Date" {
		t.Errorf("L4=%q, want Closure Date", got)
	}
	if got := cell(t, f, "A5"); got != "FD-2" {
		t.Errorf("first closed fd=%q, want FD-2", got)
	}
	if got := numCell(t, f, "N5"); got != 100044 {
		t.Errorf("payable=%d, want 100044", got)
	}
}

func TestActiveReportOnlyActiveRows(t *testing.T) {
	path := export(t, api.ReportActive, sampleFDs())
	f := readSheet(t, path)

	if got := cell(t, f, "A5"); got != "FD-1" {
		t.Errorf("first active=%q, want FD-1", got)
	}
	if got := cell(t, f, "A6"); got != "" {
		t.Errorf("second row=%q, want empty (closed FD excluded)", got)
	}
}

func TestMaturityReportDaysRemaining(t *testing.T) {
	path := export(t, api.ReportMaturity, sampleFDs())
	f := readSheet(t, path)

	if got := cell(t, f, "F4"); got != "Days Remaining" {
		t.Errorf("F4=%q, want Days Remaining", got)
	}
	// Generated 06/10/2026; FD-1 matures 01/10/2027 → 360 days.
	if got := cell(t, f, "F5"); got != "360" {
		t.Errorf("days remaining=%q, want 360", got)
	}
}

// TestMembersReportContents covers the member-wise summary (SRS §29.6),
// including the trailing "Unassigned" row for FDs without a member link.
func TestMembersReportContents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "members.xlsx")
	err := report.Export(report.Options{
		Kind:        api.ReportMembers,
		Title:       "Member-wise FD Summary",
		GeneratedAt: time.Date(2026, 10, 6, 12, 0, 0, 0, time.Local),
		Members: []domain.MemberWithStats{
			{Member: domain.Member{GENNo: "GEN-1", Name: "Alpha", Designation: "Teacher"},
				ActiveFDCount: 2, ActiveFDAmount: 150000},
		},
		UnassignedActiveFDCount: 1,
		UnassignedActiveAmount:  50000,
		Path:                    path,
	})
	if err != nil {
		t.Fatalf("export members: %v", err)
	}
	f := readSheet(t, path)

	wantValues := []string{"GEN No.", "Member Name", "Designation", "Active FD Count", "Total Active FD Amount"}
	for i, ref := range []string{"A4", "B4", "C4", "D4", "E4"} {
		if got := cell(t, f, ref); got != wantValues[i] {
			t.Errorf("%s=%q, want %q", ref, got, wantValues[i])
		}
	}
	if got := cell(t, f, "A5"); got != "GEN-1" {
		t.Errorf("row1 gen=%q, want GEN-1", got)
	}
	if got := numCell(t, f, "D5"); got != 2 {
		t.Errorf("row1 count=%d, want 2", got)
	}
	if got := numCell(t, f, "E5"); got != 150000 {
		t.Errorf("row1 amount=%d, want 150000", got)
	}
	if got := cell(t, f, "A6"); got != "Unassigned" {
		t.Errorf("row2=%q, want Unassigned", got)
	}
	if got := numCell(t, f, "D6"); got != 1 {
		t.Errorf("row2 count=%d, want 1", got)
	}
	if got := numCell(t, f, "E6"); got != 50000 {
		t.Errorf("row2 amount=%d, want 50000", got)
	}
}

func TestExportAppendsExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "noext")
	if err := report.Export(report.Options{Kind: api.ReportRegister, Title: "T", GeneratedAt: time.Now(), Path: path}); err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := excelize.OpenFile(path + ".xlsx"); err != nil {
		t.Errorf("file with appended .xlsx not found: %v", err)
	}
}
