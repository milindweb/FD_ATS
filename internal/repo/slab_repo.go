package repo

import (
	"database/sql"
	"fmt"

	"fdats/internal/domain"
)

// DefaultRateSlabs are the SRS §13 defaults. They are seeded into an empty
// database once; afterwards the stored configuration is authoritative.
func DefaultRateSlabs() []domain.RateSlab {
	return []domain.RateSlab{
		{SortOrder: 1, MinDays: 1, MaxDays: 364, RatePercent: 4.00, Label: "1-364 Days"},
		{SortOrder: 2, MinDays: 365, MaxDays: 729, RatePercent: 8.00, Label: "365-729 Days"},
		{SortOrder: 3, MinDays: 730, MaxDays: 1094, RatePercent: 8.50, Label: "730-1094 Days"},
		{SortOrder: 4, MinDays: 1095, MaxDays: 1460, RatePercent: 9.50, Label: "1095-1460 Days"},
	}
}

// SlabRepo manages the interest rate configuration (SRS §13, §30).
type SlabRepo struct {
	db *sql.DB
}

func NewSlabRepo(db *sql.DB) *SlabRepo { return &SlabRepo{db: db} }

// List returns all configured slabs ordered for display.
func (s *SlabRepo) List() ([]domain.RateSlab, error) {
	rows, err := s.db.Query(
		`SELECT id, sort_order, min_days, max_days, rate_percent, label
		 FROM rate_slabs ORDER BY sort_order ASC, min_days ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.RateSlab
	for rows.Next() {
		var slab domain.RateSlab
		if err := rows.Scan(&slab.ID, &slab.SortOrder, &slab.MinDays, &slab.MaxDays, &slab.RatePercent, &slab.Label); err != nil {
			return nil, err
		}
		out = append(out, slab)
	}
	return out, rows.Err()
}

// EnsureSeeded inserts the default slabs when the table is empty.
func (s *SlabRepo) EnsureSeeded() error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM rate_slabs`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return s.Replace(DefaultRateSlabs())
}

// Replace swaps the whole slab configuration in one transaction.
func (s *SlabRepo) Replace(slabs []domain.RateSlab) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM rate_slabs`); err != nil {
		return err
	}
	for i, slab := range slabs {
		if slab.SortOrder == 0 {
			slab.SortOrder = i + 1
		}
		if _, err := tx.Exec(
			`INSERT INTO rate_slabs(sort_order, min_days, max_days, rate_percent, label) VALUES(?, ?, ?, ?, ?)`,
			slab.SortOrder, slab.MinDays, slab.MaxDays, slab.RatePercent, slab.Label,
		); err != nil {
			return fmt.Errorf("insert slab %q: %w", slab.Label, err)
		}
	}
	return tx.Commit()
}
