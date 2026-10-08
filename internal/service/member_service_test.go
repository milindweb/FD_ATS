package service

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"fdats/internal/api"
	"fdats/internal/domain"
	"fdats/internal/repo"
)

func newMemberService(t *testing.T) (*MemberService, *sql.DB) {
	t.Helper()
	db, err := repo.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewMemberService(db), db
}

func saveMember(t *testing.T, svc *MemberService, gen, name string) api.Member {
	t.Helper()
	m, err := svc.Save(api.SaveMemberRequest{GENNo: gen, Name: name})
	if err != nil {
		t.Fatalf("save %s: %v", gen, err)
	}
	return m
}

func TestMemberSaveValidation(t *testing.T) {
	svc, _ := newMemberService(t)

	if _, err := svc.Save(api.SaveMemberRequest{Name: "No GEN"}); err != domain.ErrGENRequired {
		t.Errorf("missing GEN = %v, want ErrGENRequired", err)
	}
	if _, err := svc.Save(api.SaveMemberRequest{GENNo: "GEN-1"}); err != domain.ErrMemberNameRequired {
		t.Errorf("missing Name = %v, want ErrMemberNameRequired", err)
	}

	m := saveMember(t, svc, "GEN-1", "Asha Nair")
	if m.ID == 0 || m.CreatedAt == "" || m.UpdatedAt == "" {
		t.Errorf("created member = %+v, want id and timestamps", m)
	}

	if _, err := svc.Save(api.SaveMemberRequest{GENNo: "GEN-1", Name: "Impostor"}); err != domain.ErrDuplicateGEN {
		t.Errorf("duplicate GEN = %v, want domain.ErrDuplicateGEN", err)
	}
	if _, err := svc.Save(api.SaveMemberRequest{ID: 99999, GENNo: "GEN-2", Name: "Ghost"}); err != domain.ErrMemberNotFound {
		t.Errorf("update missing = %v, want ErrMemberNotFound", err)
	}
}

func TestMemberSaveUpdateSyncsSnapshots(t *testing.T) {
	svc, db := newMemberService(t)

	m := saveMember(t, svc, "GEN-77", "Old Name")
	insertLinkedFD(t, db, "FD-1", m.ID, "Old Name", "GEN-77")

	updated, err := svc.Save(api.SaveMemberRequest{
		ID: m.ID, GENNo: "GEN-78", Name: "New Name", Designation: "Officer",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.CreatedAt != m.CreatedAt {
		t.Errorf("createdAt changed on update: %s → %s", m.CreatedAt, updated.CreatedAt)
	}
	if updated.Designation != "Officer" {
		t.Errorf("designation = %q, want Officer", updated.Designation)
	}

	var custName, custNumber string
	if err := db.QueryRow("SELECT customer_name, customer_number FROM fds WHERE fd_number = 'FD-1'").Scan(&custName, &custNumber); err != nil {
		t.Fatalf("read fd: %v", err)
	}
	if custName != "New Name" || custNumber != "GEN-78" {
		t.Errorf("snapshot = %s/%s, want New Name/GEN-78 (SRS §53)", custName, custNumber)
	}
}

func insertLinkedFD(t *testing.T, db *sql.DB, fdNumber string, memberID int64, custName, custNumber string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO fds (fd_number, member_id, customer_name, customer_number, principal,
		start_date, tenure_days, interest_rate, maturity_date, interest_amount, maturity_amount, status,
		created_at, updated_at)
		VALUES (?, ?, ?, ?, 10000, '2026-01-01', 365, 7.0, '2027-01-01', 700, 10700, 'ACTIVE', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`,
		fdNumber, memberID, custName, custNumber)
	if err != nil {
		t.Fatalf("insert fd: %v", err)
	}
}

func TestMemberListAndProfile(t *testing.T) {
	svc, db := newMemberService(t)

	saveMember(t, svc, "GEN-10", "Asha Nair")
	saveMember(t, svc, "GEN-20", "Ravi Kumar")

	var id int64
	if err := db.QueryRow("SELECT id FROM members WHERE gen_no = 'GEN-10'").Scan(&id); err != nil {
		t.Fatalf("lookup: %v", err)
	}
	insertLinkedFD(t, db, "FD-1", id, "Asha Nair", "GEN-10")
	insertLinkedFD(t, db, "FD-2", id, "Asha Nair", "GEN-10")
	_, err := db.Exec("UPDATE fds SET status = 'CLOSED' WHERE fd_number = 'FD-2'")
	if err != nil {
		t.Fatalf("close: %v", err)
	}

	list, err := svc.List(api.MemberListRequest{Search: "asha"})
	if err != nil || list.Total != 1 {
		t.Fatalf("list = %+v err %v, want 1 match", list, err)
	}
	if list.Items[0].ActiveFDCount != 1 || list.Items[0].ActiveFDAmount != 10000 {
		t.Errorf("aggregates = %d/%d, want 1/10000", list.Items[0].ActiveFDCount, list.Items[0].ActiveFDAmount)
	}

	profile, err := svc.GetProfile(id)
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if profile.Member.GENNo != "GEN-10" || len(profile.FDs) != 2 {
		t.Errorf("profile = %s with %d fds, want GEN-10 with 2", profile.Member.GENNo, len(profile.FDs))
	}
	if profile.FDs[0].FDNumber == "" {
		t.Error("profile FDs missing fdNumber")
	}
}

func writeImportXLSX(t *testing.T, headers []string, rows [][]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "members.xlsx")
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			t.Fatalf("set header: %v", err)
		}
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				t.Fatalf("set cell: %v", err)
			}
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("save xlsx: %v", err)
	}
	_ = f.Close()
	return path
}

func TestMemberImportPreview(t *testing.T) {
	svc, db := newMemberService(t)

	saveMember(t, svc, "GEN-EXIST", "Existing Member")

	path := writeImportXLSX(t,
		[]string{"GEN No.", "Name", "DOB", "Mobile", "Aadhaar", "PAN"},
		[][]string{
			{"GEN-100", "Asha Nair", "15/06/1990", "9876543210", "123456789012", "ABCDE1234F"},
			{"GEN-EXIST", "Renamed Person", "", "", "", ""},
			{"GEN-100", "Duplicate Row", "", "", "", ""},
			{"GEN-101", "", "", "", "", ""},
			{"GEN-102", "Kiran Das", "not-a-date", "", "", ""},
			{"GEN-103", "Sunil Roy", "", "12ab", "", ""},
		},
	)

	prev, err := svc.PreviewImport(path)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if prev.Total != 6 {
		t.Errorf("total = %d, want 6", prev.Total)
	}
	if prev.Valid != 2 || prev.Existing != 1 {
		t.Errorf("valid/existing = %d/%d, want 2/1 (existing counts as valid, SRS §52.3)", prev.Valid, prev.Existing)
	}
	if prev.DuplicateGEN != 1 {
		t.Errorf("duplicate = %d, want 1", prev.DuplicateGEN)
	}
	if prev.MissingMandatory != 1 {
		t.Errorf("missing = %d, want 1", prev.MissingMandatory)
	}
	if prev.Invalid != 2 {
		t.Errorf("invalid = %d, want 2 (bad DOB and bad mobile)", prev.Invalid)
	}
	// Row statuses must be reported per row for review (SRS §52.3).
	if prev.Rows[0].Status != api.ImportRowNew || prev.Rows[1].Status != api.ImportRowExisting ||
		prev.Rows[2].Status != api.ImportRowDuplicate || prev.Rows[3].Status != api.ImportRowInvalid ||
		prev.Rows[4].Status != api.ImportRowInvalid || prev.Rows[5].Status != api.ImportRowInvalid {
		t.Errorf("row statuses = %+v, want new/existing/duplicate/invalid/invalid/invalid", prev.Rows)
	}
	if prev.Rows[3].Error == "" || prev.Rows[4].Error == "" {
		t.Error("invalid rows must carry an error message")
	}

	// Preview must not save anything (SRS §52.3).
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM members").Scan(&count); err != nil || count != 1 {
		t.Errorf("members after preview = %d (err %v), want 1", count, err)
	}
}

func TestMemberImportCommit(t *testing.T) {
	svc, db := newMemberService(t)

	saveMember(t, svc, "GEN-EXIST", "Existing Member")
	// Unlinked FD waiting for its member (SRS §53 auto-link).
	if _, err := db.Exec(`INSERT INTO fds (fd_number, customer_name, customer_number, principal,
		start_date, tenure_days, interest_rate, maturity_date, interest_amount, maturity_amount, status,
		created_at, updated_at)
		VALUES ('FD-9', 'Unknown', 'GEN-100', 10000, '2026-01-01', 365, 7.0, '2027-01-01', 700, 10700, 'ACTIVE',
		'2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("insert fd: %v", err)
	}

	path := writeImportXLSX(t,
		[]string{"GEN No.", "Name", "DOB"},
		[][]string{
			{"GEN-100", "Asha Nair", "1990-06-15"},
			{"GEN-EXIST", "Renamed Person", ""},
			{"GEN-101", "", ""},
			{"GEN-101", "Phantom", ""},
		},
	)

	res, err := svc.CommitImport(path)
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if res.Imported != 1 || res.Skipped != 3 || res.Total != 4 {
		t.Errorf("result = %+v, want imported 1, skipped 3, total 4", res)
	}

	// Imported member exists with normalised DOB.
	var name, dob string
	if err := db.QueryRow("SELECT name, dob FROM members WHERE gen_no = 'GEN-100'").Scan(&name, &dob); err != nil {
		t.Fatalf("read imported: %v", err)
	}
	if name != "Asha Nair" || dob != "1990-06-15" {
		t.Errorf("imported = %s/%s, want Asha Nair/1990-06-15", name, dob)
	}

	// Existing member never overwritten (SRS §52.4).
	if err := db.QueryRow("SELECT name FROM members WHERE gen_no = 'GEN-EXIST'").Scan(&name); err != nil {
		t.Fatalf("read existing: %v", err)
	}
	if name != "Existing Member" {
		t.Errorf("existing member = %q, want 'Existing Member' (never overwritten)", name)
	}

	// Imported GEN auto-linked to the pre-existing FD (SRS §53).
	var memberID sql.NullInt64
	if err := db.QueryRow("SELECT member_id FROM fds WHERE fd_number = 'FD-9'").Scan(&memberID); err != nil {
		t.Fatalf("read fd link: %v", err)
	}
	if !memberID.Valid {
		t.Error("FD-9 not auto-linked to the imported member")
	}
}

func TestMemberImportFileErrors(t *testing.T) {
	svc, _ := newMemberService(t)

	if _, err := svc.PreviewImport(""); err != domain.ErrImportPathRequired {
		t.Errorf("empty path = %v, want ErrImportPathRequired", err)
	}
	if _, err := svc.PreviewImport(filepath.Join(t.TempDir(), "missing.xlsx")); err == nil {
		t.Error("missing file: want error")
	}

	noGen := writeImportXLSX(t, []string{"Name"}, [][]string{{"Asha"}})
	_, err := svc.PreviewImport(noGen)
	if err == nil || !strings.Contains(err.Error(), "GEN No.") {
		t.Errorf("missing column = %v, want GEN No. column error", err)
	}

	headersOnly := writeImportXLSX(t, []string{"GEN No.", "Name"}, nil)
	if _, err := svc.PreviewImport(headersOnly); err != domain.ErrNoImportableRows {
		t.Errorf("headers only = %v, want ErrNoImportableRows", err)
	}
}

// TestWriteMemberTemplate pins the sample template to the SRS §52.1/§52.2
// column contract: the exact canonical header row, headers only (no data
// rows), on a sheet the importer will read.
func TestWriteMemberTemplate(t *testing.T) {
	svc, _ := newMemberService(t)

	path := filepath.Join(t.TempDir(), "template.xlsx")
	if err := svc.WriteMemberTemplate(path); err != nil {
		t.Fatalf("write template: %v", err)
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open template: %v", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	if sheet != "Members" {
		t.Errorf("sheet = %q, want Members", sheet)
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		t.Fatalf("read rows: %v", err)
	}
	want := []string{
		"GEN No.", "Name", "DOB", "Mobile", "Email", "Present Address",
		"Permanent Address", "Employer Name", "Department", "Designation",
		"Token No.", "Nominee Name", "Nominee Relationship", "Aadhaar", "PAN",
		"Bank Name", "Account No.", "IFSC Code", "Profile Remarks",
	}
	if len(rows) != 1 {
		t.Errorf("row count = %d, want 1 (headers only)", len(rows))
	}
	if len(rows) > 0 {
		got := rows[0]
		if len(got) != len(want) {
			t.Errorf("header count = %d, want %d: %v", len(got), len(want), got)
		}
		for i := range want {
			if i >= len(got) || got[i] != want[i] {
				t.Errorf("header[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	}

	// The template must round-trip through the importer's header parser.
	if _, err := importHeaderColumns(rows[0]); err != nil {
		t.Errorf("header parser rejected template: %v", err)
	}
}
