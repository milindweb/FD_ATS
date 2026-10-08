package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"fdats/internal/domain"
)

// ErrDuplicateGEN is returned when a GEN No. already exists (SRS §50).
var ErrDuplicateGEN = errors.New("gen no already exists")

// MemberRepo reads and writes member master records (SRS §50).
type MemberRepo struct {
	db *sql.DB
}

func NewMemberRepo(db *sql.DB) *MemberRepo { return &MemberRepo{db: db} }

const memberColumns = `id, gen_no, name, dob, mobile, email, present_address, permanent_address,
	employer_name, department, designation, token_no, nominee_name, nominee_relationship,
	aadhaar, pan, bank_name, account_no, ifsc, profile_remarks, created_at, updated_at`

func scanMember(scan func(dest ...any) error) (domain.Member, error) {
	var m domain.Member
	err := scan(
		&m.ID, &m.GENNo, &m.Name, &m.DOB, &m.Mobile, &m.Email, &m.PresentAddress, &m.PermanentAddress,
		&m.EmployerName, &m.Department, &m.Designation, &m.TokenNo, &m.NomineeName, &m.NomineeRelationship,
		&m.Aadhaar, &m.PAN, &m.BankName, &m.AccountNo, &m.IFSC, &m.ProfileRemarks, &m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

// Insert stores a new member and returns its system ID.
func (r *MemberRepo) Insert(m domain.Member) (int64, error) {
	res, err := r.db.Exec(`INSERT INTO members (gen_no, name, dob, mobile, email, present_address, permanent_address,
		employer_name, department, designation, token_no, nominee_name, nominee_relationship,
		aadhaar, pan, bank_name, account_no, ifsc, profile_remarks, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.GENNo, m.Name, m.DOB, m.Mobile, m.Email, m.PresentAddress, m.PermanentAddress,
		m.EmployerName, m.Department, m.Designation, m.TokenNo, m.NomineeName, m.NomineeRelationship,
		m.Aadhaar, m.PAN, m.BankName, m.AccountNo, m.IFSC, m.ProfileRemarks, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, fmt.Errorf("%w: %s", ErrDuplicateGEN, err.Error())
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Update applies changes to an existing member.
func (r *MemberRepo) Update(m domain.Member) error {
	res, err := r.db.Exec(`UPDATE members SET
		gen_no = ?, name = ?, dob = ?, mobile = ?, email = ?, present_address = ?, permanent_address = ?,
		employer_name = ?, department = ?, designation = ?, token_no = ?, nominee_name = ?, nominee_relationship = ?,
		aadhaar = ?, pan = ?, bank_name = ?, account_no = ?, ifsc = ?, profile_remarks = ?, updated_at = ?
		WHERE id = ?`,
		m.GENNo, m.Name, m.DOB, m.Mobile, m.Email, m.PresentAddress, m.PermanentAddress,
		m.EmployerName, m.Department, m.Designation, m.TokenNo, m.NomineeName, m.NomineeRelationship,
		m.Aadhaar, m.PAN, m.BankName, m.AccountNo, m.IFSC, m.ProfileRemarks, m.UpdatedAt,
		m.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return fmt.Errorf("%w: %s", ErrDuplicateGEN, err.Error())
		}
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

// Get fetches one member by system ID.
func (r *MemberRepo) Get(id int64) (domain.Member, error) {
	row := r.db.QueryRow("SELECT "+memberColumns+" FROM members WHERE id = ?", id)
	m, err := scanMember(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Member{}, domain.ErrMemberNotFound
	}
	return m, err
}

// GetByGEN fetches one member by GEN No.
func (r *MemberRepo) GetByGEN(genNo string) (domain.Member, error) {
	row := r.db.QueryRow("SELECT "+memberColumns+" FROM members WHERE gen_no = ?", genNo)
	m, err := scanMember(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Member{}, domain.ErrMemberNotFound
	}
	return m, err
}

// SyncFDMemberSnapshots refreshes the readable copies of the member's name
// and GEN No. on every FD linked to the member (SRS §53). FD audit fields
// are deliberately left untouched — only the snapshot columns change.
func (r *MemberRepo) SyncFDMemberSnapshots(m domain.Member) error {
	_, err := r.db.Exec(
		"UPDATE fds SET customer_name = ?, customer_number = ? WHERE member_id = ?",
		m.Name, m.GENNo, m.ID,
	)
	return err
}

// MemberListParams filters and pages the member list (SRS §51).
type MemberListParams struct {
	Search   string
	Page     int
	PageSize int
}

func memberSearchClause(search string) (string, []any) {
	s := strings.TrimSpace(search)
	if s == "" {
		return "", nil
	}
	like := "%" + strings.ToLower(s) + "%"
	return ` WHERE (LOWER(gen_no) LIKE ? OR LOWER(name) LIKE ? OR LOWER(token_no) LIKE ?
		OR LOWER(designation) LIKE ? OR LOWER(pan) LIKE ? OR LOWER(aadhaar) LIKE ? OR LOWER(mobile) LIKE ?)`,
		[]any{like, like, like, like, like, like, like}
}

// List returns a page of members with their active-FD aggregates plus the
// total match count (SRS §51.2). PageSize <= 0 returns every match.
func (r *MemberRepo) List(p MemberListParams) ([]domain.MemberWithStats, int, error) {
	where, args := memberSearchClause(p.Search)

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM members"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT m.id, m.gen_no, m.name, m.dob, m.mobile, m.email, m.present_address, m.permanent_address,
		m.employer_name, m.department, m.designation, m.token_no, m.nominee_name, m.nominee_relationship,
		m.aadhaar, m.pan, m.bank_name, m.account_no, m.ifsc, m.profile_remarks, m.created_at, m.updated_at,
		COALESCE(SUM(CASE WHEN f.status = 'ACTIVE' THEN 1 ELSE 0 END), 0) AS active_count,
		COALESCE(SUM(CASE WHEN f.status = 'ACTIVE' THEN f.principal ELSE 0 END), 0) AS active_amount
		FROM members m
		LEFT JOIN fds f ON f.member_id = m.id` + where +
		" GROUP BY m.id ORDER BY m.gen_no ASC"

	qArgs := append([]any{}, args...)
	if p.PageSize > 0 {
		if p.Page < 1 {
			p.Page = 1
		}
		query += " LIMIT ? OFFSET ?"
		qArgs = append(qArgs, p.PageSize, (p.Page-1)*p.PageSize)
	}

	rows, err := r.db.Query(query, qArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []domain.MemberWithStats
	for rows.Next() {
		var mw domain.MemberWithStats
		err := rows.Scan(
			&mw.ID, &mw.GENNo, &mw.Name, &mw.DOB, &mw.Mobile, &mw.Email, &mw.PresentAddress, &mw.PermanentAddress,
			&mw.EmployerName, &mw.Department, &mw.Designation, &mw.TokenNo, &mw.NomineeName, &mw.NomineeRelationship,
			&mw.Aadhaar, &mw.PAN, &mw.BankName, &mw.AccountNo, &mw.IFSC, &mw.ProfileRemarks, &mw.CreatedAt, &mw.UpdatedAt,
			&mw.ActiveFDCount, &mw.ActiveFDAmount,
		)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mw)
	}
	return items, total, rows.Err()
}
