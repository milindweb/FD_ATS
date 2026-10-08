// Package repo provides SQLite persistence for the FD application.
// Storage is a single local database file; no server is required (SRS §4).
package repo

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const driverName = "sqlite"

// Open creates (if needed) and opens the SQLite database inside dataDir,
// applying the schema migrations.
func Open(dataDir string) (*sql.DB, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "fd_management.db")
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", filepath.ToSlash(dbPath))

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	// SQLite benefits from a single writer connection.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := Migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// Migrate creates the schema. It is idempotent so it can run on every start.
func Migrate(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS rate_slabs (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	sort_order  INTEGER NOT NULL,
	min_days    INTEGER NOT NULL,
	max_days    INTEGER NOT NULL,
	rate_percent REAL NOT NULL,
	label       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS members (
	id                  INTEGER PRIMARY KEY AUTOINCREMENT,
	gen_no              TEXT NOT NULL UNIQUE,
	name                TEXT NOT NULL,
	dob                 TEXT NOT NULL DEFAULT '',
	mobile              TEXT NOT NULL DEFAULT '',
	email               TEXT NOT NULL DEFAULT '',
	present_address     TEXT NOT NULL DEFAULT '',
	permanent_address   TEXT NOT NULL DEFAULT '',
	employer_name       TEXT NOT NULL DEFAULT '',
	department          TEXT NOT NULL DEFAULT '',
	designation         TEXT NOT NULL DEFAULT '',
	token_no            TEXT NOT NULL DEFAULT '',
	nominee_name        TEXT NOT NULL DEFAULT '',
	nominee_relationship TEXT NOT NULL DEFAULT '',
	aadhaar             TEXT NOT NULL DEFAULT '',
	pan                 TEXT NOT NULL DEFAULT '',
	bank_name           TEXT NOT NULL DEFAULT '',
	account_no          TEXT NOT NULL DEFAULT '',
	ifsc                TEXT NOT NULL DEFAULT '',
	profile_remarks     TEXT NOT NULL DEFAULT '',
	created_at          TEXT NOT NULL,
	updated_at          TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS fds (
	fd_number       TEXT PRIMARY KEY,
	member_id       INTEGER,
	fd_form_no      TEXT NOT NULL DEFAULT '',
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
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	// Databases created before the member feature lack the new fds columns.
	if err := addColumnIfMissing(db, "fds", "member_id", "INTEGER"); err != nil {
		return err
	}
	if err := addColumnIfMissing(db, "fds", "fd_form_no", `TEXT NOT NULL DEFAULT ''`); err != nil {
		return err
	}
	// Created after the column adds: on pre-member databases the schema batch
	// above cannot create this index because fds.member_id does not exist yet.
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_fds_member ON fds(member_id)"); err != nil {
		return fmt.Errorf("create index idx_fds_member: %w", err)
	}

	if err := migrateFDNumbering(db); err != nil {
		return err
	}
	if _, err := LinkFDsToMembers(db); err != nil {
		return err
	}
	return nil
}

// addColumnIfMissing adds a column to an existing table when it is absent,
// so CREATE TABLE IF NOT EXISTS stays idempotent for already-created tables.
func addColumnIfMissing(db *sql.DB, table, column, decl string) error {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("inspect table %s: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return fmt.Errorf("inspect table %s: %w", table, err)
		}
		if name == column {
			return rows.Err()
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("inspect table %s: %w", table, err)
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, decl)); err != nil {
		return fmt.Errorf("add column %s.%s: %w", table, column, err)
	}
	return nil
}
