package repo

import (
	"database/sql"

	"fdats/internal/domain"
)

// HistoryRepo stores the chronological event log for FDs (SRS §20).
type HistoryRepo struct {
	db *sql.DB
}

func NewHistoryRepo(db *sql.DB) *HistoryRepo { return &HistoryRepo{db: db} }

// Append records a new history entry.
func (h *HistoryRepo) Append(entry domain.HistoryEntry) error {
	_, err := h.db.Exec(
		`INSERT INTO fd_history(fd_number, event_date, event_type, amount, interest, reference_fd, remarks, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.FDNumber, entry.EventDate, entry.EventType, entry.Amount, entry.Interest,
		nullStr(entry.ReferenceFD), entry.Remarks, entry.CreatedAt,
	)
	return err
}

// ListForFD returns all history entries for an FD in chronological order.
func (h *HistoryRepo) ListForFD(fdNumber string) ([]domain.HistoryEntry, error) {
	rows, err := h.db.Query(
		`SELECT id, fd_number, event_date, event_type, amount, interest, reference_fd, remarks, created_at
		 FROM fd_history WHERE fd_number = ? ORDER BY id ASC`,
		fdNumber,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.HistoryEntry
	for rows.Next() {
		var e domain.HistoryEntry
		var reference sql.NullString
		if err := rows.Scan(&e.ID, &e.FDNumber, &e.EventDate, &e.EventType, &e.Amount, &e.Interest, &reference, &e.Remarks, &e.CreatedAt); err != nil {
			return nil, err
		}
		if reference.Valid {
			e.ReferenceFD = reference.String
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
