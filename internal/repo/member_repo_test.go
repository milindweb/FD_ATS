package repo

import (
	"errors"
	"testing"

	"fdats/internal/domain"
)

func memberFixture(gen, ts string) domain.Member {
	return domain.Member{
		GENNo:     gen,
		Name:      "Member " + gen,
		CreatedAt: ts,
		UpdatedAt: ts,
	}
}

func TestMemberInsertGetUpdate(t *testing.T) {
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	repo := NewMemberRepo(db)
	m := memberFixture("GEN-1001", "2026-01-01 10:00:00")
	m.Name = "Priya Sharma"
	m.Mobile = "9876543210"
	m.TokenNo = "T-42"
	m.Designation = "Clerk"
	m.PAN = "ABCDE1234F"
	m.ProfileRemarks = "VIP"

	id, err := repo.Insert(m)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if id <= 0 {
		t.Fatalf("id = %d, want > 0", id)
	}

	got, err := repo.Get(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.GENNo != "GEN-1001" || got.Name != "Priya Sharma" || got.TokenNo != "T-42" || got.ProfileRemarks != "VIP" {
		t.Errorf("get = %+v, want inserted values", got)
	}

	byGEN, err := repo.GetByGEN("GEN-1001")
	if err != nil || byGEN.ID != id {
		t.Errorf("GetByGEN = %+v err %v, want id %d", byGEN, err, id)
	}

	if _, err := repo.Get(99999); !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("Get missing = %v, want ErrMemberNotFound", err)
	}
	if _, err := repo.GetByGEN("GEN-NOPE"); !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("GetByGEN missing = %v, want ErrMemberNotFound", err)
	}

	// Update persists changes.
	got.Name = "Priya S. Rao"
	got.UpdatedAt = "2026-01-02 10:00:00"
	if err := repo.Update(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	again, err := repo.Get(id)
	if err != nil || again.Name != "Priya S. Rao" {
		t.Errorf("after update = %+v err %v, want renamed", again, err)
	}

	// Duplicate GEN rejected on insert and update.
	if _, err := repo.Insert(memberFixture("GEN-1001", "2026-01-01 10:00:00")); !errors.Is(err, ErrDuplicateGEN) {
		t.Errorf("duplicate insert = %v, want ErrDuplicateGEN", err)
	}
	other := memberFixture("GEN-1002", "2026-01-01 10:00:00")
	otherID, err := repo.Insert(other)
	if err != nil {
		t.Fatalf("insert other: %v", err)
	}
	again.GENNo = "GEN-1002"
	if err := repo.Update(again); !errors.Is(err, ErrDuplicateGEN) {
		t.Errorf("duplicate update = %v, want ErrDuplicateGEN", err)
	}
	_ = otherID
}

func TestMemberListSearch(t *testing.T) {
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	repo := NewMemberRepo(db)
	base := memberFixture("GEN-1", "2026-01-01 10:00:00")
	base.Name = "Asha Nair"
	base.TokenNo = "TK-111"
	base.Designation = "Officer"
	base.PAN = "AAAAA1111A"
	base.Aadhaar = "111122223333"
	base.Mobile = "9000000001"
	if _, err := repo.Insert(base); err != nil {
		t.Fatalf("insert base: %v", err)
	}
	second := memberFixture("GEN-2", "2026-01-02 10:00:00")
	second.Name = "Ravi Kumar"
	second.TokenNo = "TK-222"
	second.Designation = "Manager"
	second.PAN = "BBBBB2222B"
	second.Aadhaar = "444455556666"
	second.Mobile = "9000000002"
	if _, err := repo.Insert(second); err != nil {
		t.Fatalf("insert second: %v", err)
	}
	third := memberFixture("GEN-3", "2026-01-03 10:00:00")
	third.Name = "Meera Das"
	if _, err := repo.Insert(third); err != nil {
		t.Fatalf("insert third: %v", err)
	}

	cases := []struct {
		search string
		want   string
	}{
		{"gen-2", "GEN-2"},
		{"asha", "GEN-1"},
		{"tk-222", "GEN-2"},
		{"manager", "GEN-2"},
		{"aaaaa1111a", "GEN-1"},
		{"444455556666", "GEN-2"},
		{"9000000001", "GEN-1"},
	}
	for _, c := range cases {
		items, total, err := repo.List(MemberListParams{Search: c.search})
		if err != nil {
			t.Fatalf("list %q: %v", c.search, err)
		}
		if total != 1 || len(items) != 1 || items[0].GENNo != c.want {
			t.Errorf("search %q → total=%d items=%v, want exactly %s", c.search, total, items, c.want)
		}
	}

	// Empty search returns everything.
	items, total, err := repo.List(MemberListParams{})
	if err != nil || total != 3 || len(items) != 3 {
		t.Errorf("all = total %d len %d err %v, want 3/3/nil", total, len(items), err)
	}

	// Pagination.
	page1, total, err := repo.List(MemberListParams{Page: 1, PageSize: 2})
	if err != nil || total != 3 || len(page1) != 2 {
		t.Errorf("page1 = total %d len %d err %v, want 3/2/nil", total, len(page1), err)
	}
	page2, _, err := repo.List(MemberListParams{Page: 2, PageSize: 2})
	if err != nil || len(page2) != 1 || page2[0].GENNo != "GEN-3" {
		t.Errorf("page2 = %+v err %v, want GEN-3", page2, err)
	}

	// PageSize 0 returns all matches (SRS §51.2).
	all, _, err := repo.List(MemberListParams{PageSize: 0})
	if err != nil || len(all) != 3 {
		t.Errorf("pagesize 0 = %d err %v, want 3", len(all), err)
	}
}

func TestMemberActiveFDAggregates(t *testing.T) {
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	repo := NewMemberRepo(db)
	id, err := repo.Insert(memberFixture("GEN-500", "2026-01-01 10:00:00"))
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	otherID, err := repo.Insert(memberFixture("GEN-501", "2026-01-01 10:00:00"))
	if err != nil {
		t.Fatalf("insert other: %v", err)
	}
	_ = otherID

	insertRawFD(t, db, "FD-1", "GEN-500", "2026-01-01 10:00:00")
	insertRawFD(t, db, "FD-2", "GEN-500", "2026-01-02 10:00:00")
	insertRawFD(t, db, "FD-3", "GEN-500", "2026-01-03 10:00:00")
	mustExec(t, db, "UPDATE fds SET member_id = ? WHERE fd_number IN ('FD-1','FD-2','FD-3')", id)
	mustExec(t, db, "UPDATE fds SET status = 'CLOSED', principal = 50000 WHERE fd_number = 'FD-3'")

	items, _, err := repo.List(MemberListParams{Search: "GEN-500"})
	if err != nil || len(items) != 1 {
		t.Fatalf("list = %v err %v, want 1 member", items, err)
	}
	got := items[0]
	if got.ActiveFDCount != 2 {
		t.Errorf("ActiveFDCount = %d, want 2 (closed FD excluded)", got.ActiveFDCount)
	}
	if got.ActiveFDAmount != 20000 {
		t.Errorf("ActiveFDAmount = %d, want 20000", got.ActiveFDAmount)
	}

	// Member without FDs reports zero aggregates.
	empty, _, err := repo.List(MemberListParams{Search: "GEN-501"})
	if err != nil || len(empty) != 1 {
		t.Fatalf("list empty = %v err %v", empty, err)
	}
	if empty[0].ActiveFDCount != 0 || empty[0].ActiveFDAmount != 0 {
		t.Errorf("aggregates = %d/%d, want 0/0", empty[0].ActiveFDCount, empty[0].ActiveFDAmount)
	}
}
