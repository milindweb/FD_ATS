package repo

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
)

// legacySchema is the exact schema the previous release created: no members
// table and no member_id/fd_form_no columns on fds.
const legacySchema = `
CREATE TABLE IF NOT EXISTS rate_slabs (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	sort_order  INTEGER NOT NULL,
	min_days    INTEGER NOT NULL,
	max_days    INTEGER NOT NULL,
	rate_percent REAL NOT NULL,
	label       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS fds (
	fd_number       TEXT PRIMARY KEY,
	customer_name   TEXT NOT NULL,
	customer_number TEXT NOT NULL,
	principal       INTEGER NOT NULL,
	start_date      TEXT NOT NULL,
	tenure_days     INTEGER NOT NULL,
	interest_rate   REAL NOT NULL,
	maturity_date   TEXT NOT NULL,
	interest_amount INTEGER NOT NULL,
	maturity_amount INTEGER NOT NULL,
	status          TEXT NOT NULL,
	closure_date    TEXT,
	closure_type    TEXT,
	closure_remark  TEXT,
	closure_rate    REAL,
	closure_days    INTEGER,
	closure_interest INTEGER,
	closure_payable  INTEGER,
	renewed_from    TEXT,
	renewed_to      TEXT,
	created_at      TEXT NOT NULL,
	updated_at      TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_fds_status ON fds(status);
CREATE INDEX IF NOT EXISTS idx_fds_maturity ON fds(maturity_date);
CREATE INDEX IF NOT EXISTS idx_fds_customer_number ON fds(customer_number);

CREATE TABLE IF NOT EXISTS fd_history (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	fd_number   TEXT NOT NULL,
	event_date  TEXT NOT NULL,
	event_type  TEXT NOT NULL,
	amount      INTEGER NOT NULL DEFAULT 0,
	interest    INTEGER NOT NULL DEFAULT 0,
	reference_fd TEXT,
	remarks     TEXT NOT NULL DEFAULT '',
	created_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_fd_history_number ON fd_history(fd_number, id);

CREATE TABLE IF NOT EXISTS fd_sequence (
	period       TEXT PRIMARY KEY,
	last_number  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS auth (
	id            INTEGER PRIMARY KEY CHECK (id = 1),
	username      TEXT NOT NULL,
	password_hash TEXT NOT NULL,
	recovery_code TEXT NOT NULL,
	updated_at    TEXT NOT NULL
);
`

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

// insertRawFD stores an FD with a caller-chosen number and marker, bypassing
// InsertWithNumber so tests can simulate pre-FD-N databases.
func insertRawFD(t *testing.T, db *sql.DB, number, marker, createdAt string) {
	t.Helper()
	mustExec(t, db, `INSERT INTO fds (fd_number, customer_name, customer_number, principal, start_date,
		tenure_days, interest_rate, maturity_date, interest_amount, maturity_amount, status, created_at, updated_at)
		VALUES (?, 'Test Customer', ?, 10000, '2026-01-01', 365, 7.0, '2027-01-01', 700, 10700, 'ACTIVE', ?, ?)`,
		number, marker, createdAt, createdAt)
}

func fdNumberByMarker(t *testing.T, db *sql.DB, marker string) string {
	t.Helper()
	var n string
	if err := db.QueryRow("SELECT fd_number FROM fds WHERE customer_number = ?", marker).Scan(&n); err != nil {
		t.Fatalf("lookup fd %q: %v", marker, err)
	}
	return n
}

func TestMigrateFDNumberingRenumbersLegacy(t *testing.T) {
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// Creation order A, B, C must become FD-1, FD-2, FD-3 regardless of
	// the legacy numbers being out of sequence.
	insertRawFD(t, db, "FD-26-002", "A", "2026-01-01 10:00:00")
	insertRawFD(t, db, "FD-25-007", "B", "2026-02-01 10:00:00")
	insertRawFD(t, db, "FD-26-001", "C", "2026-03-01 10:00:00")

	// Renewal chain A → B → C plus history references.
	mustExec(t, db, "UPDATE fds SET renewed_to = 'FD-25-007' WHERE customer_number = 'A'")
	mustExec(t, db, "UPDATE fds SET renewed_from = 'FD-26-002', renewed_to = 'FD-26-001' WHERE customer_number = 'B'")
	mustExec(t, db, "UPDATE fds SET renewed_from = 'FD-25-007' WHERE customer_number = 'C'")
	mustExec(t, db, `INSERT INTO fd_history (fd_number, event_date, event_type, reference_fd, created_at)
		VALUES ('FD-25-007', '2026-02-01', 'OPEN', 'FD-26-002', '2026-02-01 10:00:00')`)

	if err := migrateFDNumbering(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if got := fdNumberByMarker(t, db, "A"); got != "FD-1" {
		t.Errorf("A = %s, want FD-1", got)
	}
	if got := fdNumberByMarker(t, db, "B"); got != "FD-2" {
		t.Errorf("B = %s, want FD-2", got)
	}
	if got := fdNumberByMarker(t, db, "C"); got != "FD-3" {
		t.Errorf("C = %s, want FD-3", got)
	}

	// Renewal chain rewritten to the new numbers.
	var from, to sql.NullString
	if err := db.QueryRow("SELECT renewed_from, renewed_to FROM fds WHERE customer_number = 'B'").Scan(&from, &to); err != nil {
		t.Fatalf("read chain: %v", err)
	}
	if !from.Valid || from.String != "FD-1" || !to.Valid || to.String != "FD-3" {
		t.Errorf("B chain = %v → %v, want FD-1 → FD-3", from.String, to.String)
	}

	// History references rewritten.
	var histFD, histRef string
	if err := db.QueryRow("SELECT fd_number, reference_fd FROM fd_history").Scan(&histFD, &histRef); err != nil {
		t.Fatalf("read history: %v", err)
	}
	if histFD != "FD-2" || histRef != "FD-1" {
		t.Errorf("history = %s ref %s, want FD-2 ref FD-1", histFD, histRef)
	}

	// Sequence reseeded as one global counter at the highest number.
	var period string
	var last int
	if err := db.QueryRow("SELECT period, last_number FROM fd_sequence").Scan(&period, &last); err != nil {
		t.Fatalf("read sequence: %v", err)
	}
	if period != "GLOBAL" || last != 3 {
		t.Errorf("sequence = %s/%d, want GLOBAL/3", period, last)
	}

	// Idempotent: a second run changes nothing.
	if err := migrateFDNumbering(db); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if got := fdNumberByMarker(t, db, "A"); got != "FD-1" {
		t.Errorf("after re-run A = %s, want FD-1", got)
	}
	if err := db.QueryRow("SELECT last_number FROM fd_sequence").Scan(&last); err != nil || last != 3 {
		t.Errorf("after re-run sequence = %d (err %v), want 3", last, err)
	}
}

func TestMigrateFDNumberingNoopOnNewFormat(t *testing.T) {
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	insertRawFD(t, db, "FD-1", "A", "2026-01-01 10:00:00")
	insertRawFD(t, db, "FD-2", "B", "2026-02-01 10:00:00")

	if err := migrateFDNumbering(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if got := fdNumberByMarker(t, db, "A"); got != "FD-1" {
		t.Errorf("A = %s, want FD-1 (unchanged)", got)
	}
	if got := fdNumberByMarker(t, db, "B"); got != "FD-2" {
		t.Errorf("B = %s, want FD-2 (unchanged)", got)
	}
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM fd_sequence").Scan(&n); err != nil || n != 0 {
		t.Errorf("fd_sequence rows = %d (err %v), want 0 (migration must not touch sequence)", n, err)
	}
}

func TestMigrateFDNumberingSwapsInterleavedFormat(t *testing.T) {
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// Mixed formats where new-format rows must swap: FD-2 (t1) → FD-1 and
	// FD-1 (t2) → FD-2, with a legacy row taking FD-3. Requires parking.
	insertRawFD(t, db, "FD-2", "A", "2026-01-01 10:00:00")
	insertRawFD(t, db, "FD-1", "B", "2026-02-01 10:00:00")
	insertRawFD(t, db, "FD-26-001", "C", "2026-03-01 10:00:00")

	if err := migrateFDNumbering(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if got := fdNumberByMarker(t, db, "A"); got != "FD-1" {
		t.Errorf("A = %s, want FD-1", got)
	}
	if got := fdNumberByMarker(t, db, "B"); got != "FD-2" {
		t.Errorf("B = %s, want FD-2", got)
	}
	if got := fdNumberByMarker(t, db, "C"); got != "FD-3" {
		t.Errorf("C = %s, want FD-3", got)
	}
}

// TestMigrateOnPreMemberDatabase runs the full Migrate against a database
// created by the previous release (fds without member_id/fd_form_no). This
// is the real-user upgrade path the app executes on startup.
func TestMigrateOnPreMemberDatabase(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)", filepath.ToSlash(filepath.Join(t.TempDir(), "legacy.db")))
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	mustExec(t, db, legacySchema)
	insertRawFD(t, db, "FD-26-001", "LEGACY", "2026-01-01 10:00:00")
	mustExec(t, db, `INSERT INTO fd_history (fd_number, event_date, event_type, created_at)
		VALUES ('FD-26-001', '2026-01-01', 'OPEN', '2026-01-01 10:00:00')`)

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate legacy database: %v", err)
	}

	// New columns and index exist.
	cols := map[string]bool{}
	rows, err := db.Query("PRAGMA table_info(fds)")
	if err != nil {
		t.Fatalf("pragma: %v", err)
	}
	for rows.Next() {
		var cid, notNull, pk int
		var name, colType string
		var dflt any
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			t.Fatalf("scan pragma: %v", err)
		}
		cols[name] = true
	}
	rows.Close()
	for _, col := range []string{"member_id", "fd_form_no"} {
		if !cols[col] {
			t.Errorf("fds.%s missing after migrate", col)
		}
	}
	var idxCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_fds_member'").Scan(&idxCount); err != nil || idxCount != 1 {
		t.Errorf("idx_fds_member count = %d (err %v), want 1", idxCount, err)
	}

	// Legacy data preserved and renumbered.
	if got := fdNumberByMarker(t, db, "LEGACY"); got != "FD-1" {
		t.Errorf("legacy FD = %s, want FD-1", got)
	}
	var history int
	if err := db.QueryRow("SELECT COUNT(*) FROM fd_history").Scan(&history); err != nil || history != 1 {
		t.Errorf("history rows = %d (err %v), want 1", history, err)
	}

	// Idempotent on the migrated database.
	if err := Migrate(db); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}

func TestLinkFDsToMembers(t *testing.T) {
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	members := NewMemberRepo(db)
	idA, err := members.Insert(memberFixture("GEN-A", "2026-01-01 10:00:00"))
	if err != nil {
		t.Fatalf("insert A: %v", err)
	}
	idB, err := members.Insert(memberFixture("GEN-B", "2026-01-01 10:00:00"))
	if err != nil {
		t.Fatalf("insert B: %v", err)
	}

	insertRawFD(t, db, "FD-1", "GEN-A", "2026-01-01 10:00:00") // unlinked, matches GEN-A
	insertRawFD(t, db, "FD-2", "GEN-B", "2026-01-01 10:00:00") // already linked, must not move
	insertRawFD(t, db, "FD-3", "GEN-Z", "2026-01-01 10:00:00") // no member match
	mustExec(t, db, "UPDATE fds SET member_id = ? WHERE fd_number = 'FD-2'", idA)

	n, err := LinkFDsToMembers(db)
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	if n != 1 {
		t.Errorf("linked = %d, want 1", n)
	}

	var memberID sql.NullInt64
	if err := db.QueryRow("SELECT member_id FROM fds WHERE fd_number = 'FD-1'").Scan(&memberID); err != nil {
		t.Fatalf("read FD-1: %v", err)
	}
	if !memberID.Valid || memberID.Int64 != idA {
		t.Errorf("FD-1 member = %v, want %d", memberID, idA)
	}
	if err := db.QueryRow("SELECT member_id FROM fds WHERE fd_number = 'FD-2'").Scan(&memberID); err != nil {
		t.Fatalf("read FD-2: %v", err)
	}
	if !memberID.Valid || memberID.Int64 != idA {
		t.Errorf("FD-2 member changed = %v, want preserved %d", memberID, idA)
	}
	if err := db.QueryRow("SELECT member_id FROM fds WHERE fd_number = 'FD-3'").Scan(&memberID); err != nil {
		t.Fatalf("read FD-3: %v", err)
	}
	if memberID.Valid {
		t.Errorf("FD-3 member = %d, want NULL (unassigned)", memberID.Int64)
	}
	_ = idB

	// Idempotent: nothing left to link.
	n, err = LinkFDsToMembers(db)
	if err != nil {
		t.Fatalf("second link: %v", err)
	}
	if n != 0 {
		t.Errorf("second link = %d, want 0", n)
	}
}
