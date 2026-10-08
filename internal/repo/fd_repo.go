package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"fdats/internal/domain"
)

// ErrDuplicateFDNumber is returned when a generated FD number already exists.
var ErrDuplicateFDNumber = errors.New("fd number already exists")

// FDRepo reads and writes Fixed Deposit records.
type FDRepo struct {
	db *sql.DB
}

func NewFDRepo(db *sql.DB) *FDRepo { return &FDRepo{db: db} }

func (r *FDRepo) DB() *sql.DB { return r.db }

const fdColumns = `fd_number, member_id, fd_form_no, customer_name, customer_number, principal, start_date,
	tenure_days, interest_rate, maturity_date, interest_amount, maturity_amount, status,
	closure_date, closure_type, closure_remark, closure_rate, closure_days,
	closure_interest, closure_payable, renewed_from, renewed_to, created_at, updated_at`

func scanFD(scan func(dest ...any) error) (domain.FixedDeposit, error) {
	var fd domain.FixedDeposit
	var memberID sql.NullInt64
	var closureDate, closureType, closureRemark sql.NullString
	var closureRate sql.NullFloat64
	var closureDays sql.NullInt64
	var closureInterest, closurePayable sql.NullInt64
	var renewedFrom, renewedTo sql.NullString

	err := scan(
		&fd.FDNumber, &memberID, &fd.FDFormNo, &fd.CustomerName, &fd.CustomerNumber, &fd.Principal, &fd.StartDate,
		&fd.TenureDays, &fd.InterestRate, &fd.MaturityDate, &fd.InterestAmount, &fd.MaturityAmount,
		&fd.Status, &closureDate, &closureType, &closureRemark, &closureRate, &closureDays,
		&closureInterest, &closurePayable, &renewedFrom, &renewedTo, &fd.CreatedAt, &fd.UpdatedAt,
	)
	if err != nil {
		return fd, err
	}
	fd.ClosureRemark = closureRemark.String
	if memberID.Valid {
		v := memberID.Int64
		fd.MemberID = &v
	}
	if closureDate.Valid {
		fd.ClosureDate = &closureDate.String
	}
	if closureType.Valid {
		t := domain.ClosureType(closureType.String)
		fd.ClosureType = &t
	}
	if closureRate.Valid {
		fd.ClosureRate = &closureRate.Float64
	}
	if closureDays.Valid {
		d := int(closureDays.Int64)
		fd.ClosureDays = &d
	}
	if closureInterest.Valid {
		v := closureInterest.Int64
		fd.ClosureInterest = &v
	}
	if closurePayable.Valid {
		v := closurePayable.Int64
		fd.ClosurePayable = &v
	}
	if renewedFrom.Valid {
		fd.RenewedFrom = &renewedFrom.String
	}
	if renewedTo.Valid {
		fd.RenewedTo = &renewedTo.String
	}
	return fd, nil
}

// Get fetches one FD by number.
func (r *FDRepo) Get(fdNumber string) (domain.FixedDeposit, error) {
	row := r.db.QueryRow("SELECT "+fdColumns+" FROM fds WHERE fd_number = ?", fdNumber)
	fd, err := scanFD(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.FixedDeposit{}, domain.ErrFDNotFound
	}
	return fd, err
}

// InsertWithNumber allocates the next FD-N and inserts the record in one
// transaction, retrying if a concurrent writer took the number first.
func (r *FDRepo) InsertWithNumber(fd domain.FixedDeposit, attempts int) (string, error) {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		number, err := r.insertOnce(fd)
		if err == nil {
			return number, nil
		}
		if errors.Is(err, ErrDuplicateFDNumber) {
			lastErr = err
			continue
		}
		return "", err
	}
	return "", fmt.Errorf("could not allocate an FD number after %d attempts: %w", attempts, lastErr)
}

func (r *FDRepo) insertOnce(fd domain.FixedDeposit) (string, error) {
	period := domain.FDSequenceKey
	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	var last int
	err = tx.QueryRow("SELECT last_number FROM fd_sequence WHERE period = ?", period).Scan(&last)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	next := last + 1
	if errors.Is(err, sql.ErrNoRows) {
		if _, err := tx.Exec("INSERT INTO fd_sequence(period, last_number) VALUES(?, ?)", period, next); err != nil {
			return "", err
		}
	} else {
		if _, err := tx.Exec("UPDATE fd_sequence SET last_number = ? WHERE period = ?", next, period); err != nil {
			return "", err
		}
	}

	fd.FDNumber = domain.FormatFDNumber(next)
	if _, err := tx.Exec(`INSERT INTO fds (`+fdColumns+`)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fd.FDNumber, fd.MemberID, fd.FDFormNo, fd.CustomerName, fd.CustomerNumber, fd.Principal, fd.StartDate,
		fd.TenureDays, fd.InterestRate, fd.MaturityDate, fd.InterestAmount, fd.MaturityAmount,
		fd.Status, fd.ClosureDate, fd.ClosureType, fd.ClosureRemark, fd.ClosureRate, fd.ClosureDays,
		fd.ClosureInterest, fd.ClosurePayable, fd.RenewedFrom, fd.RenewedTo, fd.CreatedAt, fd.UpdatedAt,
	); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return "", fmt.Errorf("%w: %s", ErrDuplicateFDNumber, err.Error())
		}
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return fd.FDNumber, nil
}

// Update applies changes to an existing FD record.
func (r *FDRepo) Update(fd domain.FixedDeposit) error {
	res, err := r.db.Exec(`UPDATE fds SET
		member_id = ?, fd_form_no = ?, customer_name = ?, customer_number = ?, principal = ?, start_date = ?, tenure_days = ?,
		interest_rate = ?, maturity_date = ?, interest_amount = ?, maturity_amount = ?, status = ?,
		closure_date = ?, closure_type = ?, closure_remark = ?, closure_rate = ?, closure_days = ?,
		closure_interest = ?, closure_payable = ?, renewed_from = ?, renewed_to = ?, updated_at = ?
		WHERE fd_number = ?`,
		fd.MemberID, fd.FDFormNo, fd.CustomerName, fd.CustomerNumber, fd.Principal, fd.StartDate, fd.TenureDays,
		fd.InterestRate, fd.MaturityDate, fd.InterestAmount, fd.MaturityAmount, fd.Status,
		fd.ClosureDate, fd.ClosureType, fd.ClosureRemark, fd.ClosureRate, fd.ClosureDays,
		fd.ClosureInterest, fd.ClosurePayable, fd.RenewedFrom, fd.RenewedTo, fd.UpdatedAt,
		fd.FDNumber,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrFDNotFound
	}
	return nil
}

// ReverseRenewal atomically withdraws the renewed FD together with its
// history rows and reopens the previous FD, appending the audit entry.
func (r *FDRepo) ReverseRenewal(reopened domain.FixedDeposit, withdrawnFDNumber string, entry domain.HistoryEntry) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec("DELETE FROM fd_history WHERE fd_number = ?", withdrawnFDNumber); err != nil {
		return err
	}
	res, err := tx.Exec("DELETE FROM fds WHERE fd_number = ?", withdrawnFDNumber)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrFDNotFound
	}

	if _, err := tx.Exec(`UPDATE fds SET
		member_id = ?, fd_form_no = ?, customer_name = ?, customer_number = ?, principal = ?, start_date = ?, tenure_days = ?,
		interest_rate = ?, maturity_date = ?, interest_amount = ?, maturity_amount = ?, status = ?,
		closure_date = ?, closure_type = ?, closure_remark = ?, closure_rate = ?, closure_days = ?,
		closure_interest = ?, closure_payable = ?, renewed_from = ?, renewed_to = ?, updated_at = ?
		WHERE fd_number = ?`,
		reopened.MemberID, reopened.FDFormNo, reopened.CustomerName, reopened.CustomerNumber, reopened.Principal, reopened.StartDate, reopened.TenureDays,
		reopened.InterestRate, reopened.MaturityDate, reopened.InterestAmount, reopened.MaturityAmount, reopened.Status,
		reopened.ClosureDate, reopened.ClosureType, reopened.ClosureRemark, reopened.ClosureRate, reopened.ClosureDays,
		reopened.ClosureInterest, reopened.ClosurePayable, reopened.RenewedFrom, reopened.RenewedTo, reopened.UpdatedAt,
		reopened.FDNumber,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`INSERT INTO fd_history(fd_number, event_date, event_type, amount, interest, reference_fd, remarks, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.FDNumber, entry.EventDate, entry.EventType, entry.Amount, entry.Interest,
		nullStr(entry.ReferenceFD), entry.Remarks, entry.CreatedAt,
	); err != nil {
		return err
	}

	return tx.Commit()
}

// ListParams filters and pages the FD list (SRS §9.2).
type ListParams struct {
	Search   string
	Filter   string
	Cutoff   string // used by the MATURING filter: maturity_date <= Cutoff
	Page     int
	PageSize int
}

// List returns a page of FDs matching the params plus the total match count.
func (r *FDRepo) List(p ListParams) ([]domain.FixedDeposit, int, error) {
	where, args := buildWhere(p)

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM fds"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 25
	}
	offset := (p.Page - 1) * p.PageSize

	query := "SELECT " + fdColumns + " FROM fds" + where + " ORDER BY created_at DESC, fd_number DESC LIMIT ? OFFSET ?"
	qArgs := append(append([]any{}, args...), p.PageSize, offset)

	rows, err := r.db.Query(query, qArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []domain.FixedDeposit
	for rows.Next() {
		fd, err := scanFD(rows.Scan)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, fd)
	}
	return items, total, rows.Err()
}

// ListUpcoming returns active FDs maturing between fromISO and toISO inclusive.
func (r *FDRepo) ListUpcoming(fromISO, toISO string) ([]domain.FixedDeposit, error) {
	rows, err := r.db.Query(
		"SELECT "+fdColumns+" FROM fds WHERE status = ? AND maturity_date >= ? AND maturity_date <= ? ORDER BY maturity_date ASC",
		domain.StatusActive, fromISO, toISO,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.FixedDeposit
	for rows.Next() {
		fd, err := scanFD(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, fd)
	}
	return items, rows.Err()
}

// CountActive returns the number of active FDs.
func (r *FDRepo) CountActive() (int, error) {
	var n int
	err := r.db.QueryRow("SELECT COUNT(*) FROM fds WHERE status = ?", domain.StatusActive).Scan(&n)
	return n, err
}

// SumActivePrincipal returns the total principal held in active FDs.
func (r *FDRepo) SumActivePrincipal() (int64, error) {
	var v sql.NullInt64
	err := r.db.QueryRow("SELECT SUM(principal) FROM fds WHERE status = ?", domain.StatusActive).Scan(&v)
	if err != nil {
		return 0, err
	}
	return v.Int64, nil
}

// SumActiveInterest returns total interest payable at maturity across active FDs.
func (r *FDRepo) SumActiveInterest() (int64, error) {
	var v sql.NullInt64
	err := r.db.QueryRow(
		"SELECT SUM(maturity_amount - principal) FROM fds WHERE status = ?",
		domain.StatusActive,
	).Scan(&v)
	if err != nil {
		return 0, err
	}
	return v.Int64, nil
}

// SumPrincipalBetween sums principal of FDs started within [startISO, endISO].
func (r *FDRepo) SumPrincipalBetween(startISO, endISO string) (int64, error) {
	var v sql.NullInt64
	err := r.db.QueryRow(
		"SELECT SUM(principal) FROM fds WHERE start_date >= ? AND start_date <= ?",
		startISO, endISO,
	).Scan(&v)
	if err != nil {
		return 0, err
	}
	return v.Int64, nil
}

// CountMaturingWithin counts active FDs maturing on or before cutoffISO.
func (r *FDRepo) CountMaturingWithin(cutoffISO string) (int, error) {
	var n int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM fds WHERE status = ? AND maturity_date <= ?",
		domain.StatusActive, cutoffISO,
	).Scan(&n)
	return n, err
}

// ListAllByStatus returns every FD with the given status (used by reports).
func (r *FDRepo) ListAllByStatus(status string) ([]domain.FixedDeposit, error) {
	rows, err := r.db.Query("SELECT "+fdColumns+" FROM fds WHERE status = ? ORDER BY fd_number ASC", status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.FixedDeposit
	for rows.Next() {
		fd, err := scanFD(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, fd)
	}
	return items, rows.Err()
}

// ListAll returns every FD ordered by number (used by the FD register).
func (r *FDRepo) ListAll() ([]domain.FixedDeposit, error) {
	rows, err := r.db.Query("SELECT " + fdColumns + " FROM fds ORDER BY fd_number ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.FixedDeposit
	for rows.Next() {
		fd, err := scanFD(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, fd)
	}
	return items, rows.Err()
}

// ListByMaturityRange returns active FDs maturing inside a date range.
func (r *FDRepo) ListByMaturityRange(fromISO, toISO string) ([]domain.FixedDeposit, error) {
	rows, err := r.db.Query(
		"SELECT "+fdColumns+" FROM fds WHERE maturity_date >= ? AND maturity_date <= ? ORDER BY maturity_date ASC",
		fromISO, toISO,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.FixedDeposit
	for rows.Next() {
		fd, err := scanFD(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, fd)
	}
	return items, rows.Err()
}

// UnassignedActiveSummary totals active FDs not linked to any member; the
// member-wise report shows them as one "Unassigned" row (SRS §53).
func (r *FDRepo) UnassignedActiveSummary() (int64, int64, error) {
	var count, amount int64
	err := r.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(principal), 0) FROM fds
		 WHERE member_id IS NULL AND status = ?`,
		string(domain.StatusActive),
	).Scan(&count, &amount)
	if err != nil {
		return 0, 0, err
	}
	return count, amount, nil
}

// ListByMember returns every FD linked to a member (SRS §51.3), oldest first.
func (r *FDRepo) ListByMember(memberID int64) ([]domain.FixedDeposit, error) {
	rows, err := r.db.Query(
		"SELECT "+fdColumns+" FROM fds WHERE member_id = ? ORDER BY start_date ASC, fd_number ASC",
		memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.FixedDeposit
	for rows.Next() {
		fd, err := scanFD(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, fd)
	}
	return items, rows.Err()
}

func buildWhere(p ListParams) (string, []any) {
	var clauses []string
	var args []any

	if s := strings.TrimSpace(p.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		clauses = append(clauses, "(LOWER(fd_number) LIKE ? OR LOWER(fd_form_no) LIKE ? OR LOWER(customer_name) LIKE ? OR LOWER(customer_number) LIKE ?)")
		args = append(args, like, like, like, like)
	}

	switch strings.ToUpper(p.Filter) {
	case "", domain.FilterAll:
	case domain.FilterActive:
		clauses = append(clauses, "status = ?")
		args = append(args, domain.StatusActive)
	case domain.FilterClosed:
		clauses = append(clauses, "status = ?")
		args = append(args, domain.StatusClosed)
	case domain.FilterMaturing:
		clauses = append(clauses, "status = ? AND maturity_date <= ?")
		args = append(args, domain.StatusActive, p.Cutoff)
	}

	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}
