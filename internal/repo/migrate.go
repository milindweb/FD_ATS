package repo

import (
	"database/sql"
	"fmt"
)

// legacyFDPattern matches the pre-FD-N numbering format (FD-YY-NNN).
// New-format numbers (FD-N) never contain a second dash, so the pattern
// only fires while old-format rows exist — keeping the migration idempotent.
const legacyFDPattern = "FD-[0-9][0-9]-[0-9]*"

// migrateFDNumbering performs the one-time clean break to FD-N (SRS §11):
// every FD is renumbered to FD-1…FD-N in creation order, history and
// renewal links are rewritten, and the sequence table is reseeded as a
// single global counter. It runs inside Migrate and is a no-op once no
// legacy-format FD numbers remain.
func migrateFDNumbering(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM fds WHERE fd_number GLOB ?", legacyFDPattern).Scan(&count); err != nil {
		return fmt.Errorf("check legacy fd numbers: %w", err)
	}
	if count == 0 {
		return nil
	}

	rows, err := db.Query("SELECT fd_number FROM fds ORDER BY created_at ASC, fd_number ASC")
	if err != nil {
		return fmt.Errorf("list fd numbers: %w", err)
	}
	var olds []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			rows.Close()
			return fmt.Errorf("list fd numbers: %w", err)
		}
		olds = append(olds, n)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Bijection old → FD-1…FD-N by creation order.
	mapping := make(map[string]string, len(olds))
	var order [][2]string
	for i, old := range olds {
		next := fmt.Sprintf("FD-%d", i+1)
		if old == next {
			continue
		}
		mapping[old] = next
		order = append(order, [2]string{old, next})
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("renumber fds: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Park every renamed value on a collision-free temp before taking the
	// final name: rename chains (A→B while B→C) must never observe a value
	// another row has just vacated or is about to claim.
	parkAndRename := func(table, column string) error {
		for old := range mapping {
			q := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", table, column, column)
			if _, err := tx.Exec(q, "#T#"+old, old); err != nil {
				return fmt.Errorf("renumber %s.%s: %w", table, column, err)
			}
		}
		for _, e := range order {
			q := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", table, column, column)
			if _, err := tx.Exec(q, e[1], "#T#"+e[0]); err != nil {
				return fmt.Errorf("renumber %s.%s: %w", table, column, err)
			}
		}
		return nil
	}

	if err := parkAndRename("fds", "fd_number"); err != nil {
		return err
	}
	if err := parkAndRename("fd_history", "fd_number"); err != nil {
		return err
	}
	if err := parkAndRename("fd_history", "reference_fd"); err != nil {
		return err
	}
	if err := parkAndRename("fds", "renewed_from"); err != nil {
		return err
	}
	if err := parkAndRename("fds", "renewed_to"); err != nil {
		return err
	}

	// One global counter for future FD-N allocations.
	if _, err := tx.Exec("DELETE FROM fd_sequence"); err != nil {
		return fmt.Errorf("renumber fds: %w", err)
	}
	if _, err := tx.Exec("INSERT INTO fd_sequence(period, last_number) VALUES('GLOBAL', ?)", len(olds)); err != nil {
		return fmt.Errorf("renumber fds: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("renumber fds: %w", err)
	}
	return nil
}

// LinkFDsToMembers matches unlinked FDs to members whose GEN No. equals the
// FD's stored customer number (SRS §53). It is safe to run repeatedly and
// never touches FDs that are already linked. Returns the number of FDs linked.
func LinkFDsToMembers(db *sql.DB) (int64, error) {
	res, err := db.Exec(`
		UPDATE fds SET member_id = (SELECT m.id FROM members m WHERE m.gen_no = fds.customer_number)
		WHERE member_id IS NULL
		  AND customer_number <> ''
		  AND EXISTS (SELECT 1 FROM members m WHERE m.gen_no = fds.customer_number)`)
	if err != nil {
		return 0, fmt.Errorf("link fds to members: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("link fds to members: %w", err)
	}
	return n, nil
}
