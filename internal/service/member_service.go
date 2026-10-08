// MemberService implements the member master use cases of the SRS: member
// CRUD, search/list with FD aggregates, the member profile view and the bulk
// Excel member import with its validation preview (SRS §50–§52).
package service

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"fdats/internal/api"
	"fdats/internal/domain"
	"fdats/internal/repo"
)

// MemberService exposes the member master use cases.
type MemberService struct {
	db      *sql.DB
	members *repo.MemberRepo
	fds     *repo.FDRepo
	now     func() time.Time
}

// NewMemberService wires the member service against an open database.
func NewMemberService(db *sql.DB) *MemberService {
	return &MemberService{
		db:      db,
		members: repo.NewMemberRepo(db),
		fds:     repo.NewFDRepo(db),
		now:     time.Now,
	}
}

// SetClock replaces the wall clock (tests only).
func (s *MemberService) SetClock(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func (s *MemberService) nowUTC() time.Time { return s.now().UTC() }

// ---------------------------------------------------------------------------
// CRUD & search
// ---------------------------------------------------------------------------

// List returns one page of members with active-FD aggregates (SRS §51.1,
// §51.2). An empty search returns every member.
func (s *MemberService) List(req api.MemberListRequest) (api.MemberListResponse, error) {
	items, total, err := s.members.List(repo.MemberListParams{
		Search:   req.Search,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return api.MemberListResponse{}, err
	}
	out := make([]api.Member, 0, len(items))
	for _, m := range items {
		out = append(out, toMemberWithStats(m))
	}
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	return api.MemberListResponse{Items: out, Total: total, Page: page, PageSize: size}, nil
}

// Get returns one member master record by system ID — used by the FD flows
// to resolve the selected member (SRS §10, §53).
func (s *MemberService) Get(id int64) (domain.Member, error) {
	return s.members.Get(id)
}

// GetProfile returns one member together with every FD linked to it
// (SRS §51.3).
func (s *MemberService) GetProfile(id int64) (api.MemberProfile, error) {
	m, err := s.members.Get(id)
	if err != nil {
		return api.MemberProfile{}, err
	}
	fds, err := s.fds.ListByMember(id)
	if err != nil {
		return api.MemberProfile{}, err
	}
	out := make([]api.FD, 0, len(fds))
	for _, fd := range fds {
		out = append(out, toFD(fd))
	}
	return api.MemberProfile{Member: toMember(m), FDs: out}, nil
}

// Save creates (ID = 0) or updates a member and refreshes the readable name
// and GEN No. copies stored on its linked FDs (SRS §50, §53).
func (s *MemberService) Save(req api.SaveMemberRequest) (api.Member, error) {
	m := memberFromRequest(req)
	m.GENNo = strings.TrimSpace(m.GENNo)
	m.Name = strings.TrimSpace(m.Name)
	if m.GENNo == "" {
		return api.Member{}, domain.ErrGENRequired
	}
	if m.Name == "" {
		return api.Member{}, domain.ErrMemberNameRequired
	}

	if m.ID == 0 {
		m.CreatedAt = formatTimestamp(s.nowUTC())
		m.UpdatedAt = m.CreatedAt
		id, err := s.members.Insert(m)
		if err != nil {
			if errors.Is(err, repo.ErrDuplicateGEN) {
				return api.Member{}, domain.ErrDuplicateGEN
			}
			return api.Member{}, err
		}
		m.ID = id
		return s.saved(m.ID)
	}

	existing, err := s.members.Get(m.ID)
	if err != nil {
		return api.Member{}, err
	}
	m.CreatedAt = existing.CreatedAt
	m.UpdatedAt = formatTimestamp(s.nowUTC())
	if err := s.members.Update(m); err != nil {
		if errors.Is(err, repo.ErrDuplicateGEN) {
			return api.Member{}, domain.ErrDuplicateGEN
		}
		return api.Member{}, err
	}
	// Refresh the readable copies on every linked FD (SRS §53).
	if err := s.members.SyncFDMemberSnapshots(m); err != nil {
		return api.Member{}, err
	}
	return s.saved(m.ID)
}

// saved re-reads a member after a save so the caller receives real
// timestamps and the aggregate fields.
func (s *MemberService) saved(id int64) (api.Member, error) {
	m, err := s.members.Get(id)
	if err != nil {
		return api.Member{}, err
	}
	return toMember(m), nil
}

// AllWithStats returns every member with active-FD aggregates, ordered by
// GEN No. — the dataset for the member-wise report (SRS §29.6).
func (s *MemberService) AllWithStats() ([]domain.MemberWithStats, error) {
	items, _, err := s.members.List(repo.MemberListParams{})
	return items, err
}

// ---------------------------------------------------------------------------
// Bulk Excel import (SRS §52)
// ---------------------------------------------------------------------------

// PreviewImport validates an Excel file and returns the import summary
// without saving anything (SRS §52.3).
func (s *MemberService) PreviewImport(path string) (api.MemberImportPreview, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return api.MemberImportPreview{}, domain.ErrImportPathRequired
	}
	rows, err := s.parseImport(path)
	if err != nil {
		return api.MemberImportPreview{}, err
	}
	if len(rows) == 0 {
		return api.MemberImportPreview{}, domain.ErrNoImportableRows
	}

	prev := api.MemberImportPreview{Total: len(rows), Rows: make([]api.MemberImportRow, 0, len(rows))}
	for _, r := range rows {
		prev.Rows = append(prev.Rows, api.MemberImportRow{
			Row:    r.excelRow,
			GENNo:  r.member.GENNo,
			Name:   r.member.Name,
			Status: r.status,
			Error:  r.errText,
		})
		switch r.status {
		case api.ImportRowNew:
			prev.Valid++
		case api.ImportRowExisting:
			prev.Valid++
			prev.Existing++
		case api.ImportRowDuplicate:
			prev.DuplicateGEN++
		case api.ImportRowInvalid:
			if r.missingMandatory {
				prev.MissingMandatory++
			} else {
				prev.Invalid++
			}
		}
	}
	return prev, nil
}

// CommitImport re-validates the file against the current database, inserts
// every still-new member and skips the rest (SRS §52.4). Imported members
// are immediately matched against existing FDs (§53).
func (s *MemberService) CommitImport(path string) (api.MemberImportResult, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return api.MemberImportResult{}, domain.ErrImportPathRequired
	}
	rows, err := s.parseImport(path)
	if err != nil {
		return api.MemberImportResult{}, err
	}
	if len(rows) == 0 {
		return api.MemberImportResult{}, domain.ErrNoImportableRows
	}

	res := api.MemberImportResult{Total: len(rows)}
	for _, r := range rows {
		if r.status != api.ImportRowNew {
			res.Skipped++
			continue
		}
		if _, err := s.members.Insert(r.member); err != nil {
			if errors.Is(err, repo.ErrDuplicateGEN) {
				// Changed since the preview — never overwrite (SRS §52.4).
				res.Skipped++
				continue
			}
			return res, fmt.Errorf("Import stopped after %d members: %v", res.Imported, err)
		}
		res.Imported++
	}
	if _, err := repo.LinkFDsToMembers(s.db); err != nil {
		return res, err
	}
	return res, nil
}

// importRow is one validated Excel data row.
type importRow struct {
	excelRow         int
	member           domain.Member
	status           string
	errText          string
	missingMandatory bool
}

// parseImport reads and validates the first sheet of an Excel file against
// the current member set (SRS §52.1–§52.4). It performs no writes, so
// preview and commit share exactly the same validation.
func (s *MemberService) parseImport(path string) ([]importRow, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not open the Excel file: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, domain.ErrNoImportableRows
	}
	grid, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("Could not read the Excel file: %v", err)
	}
	if len(grid) == 0 {
		return nil, domain.ErrNoImportableRows
	}

	cols, err := importHeaderColumns(grid[0])
	if err != nil {
		return nil, err
	}

	existing, err := s.existingGENs()
	if err != nil {
		return nil, err
	}

	var rows []importRow
	seen := map[string]bool{}
	for i := 1; i < len(grid); i++ {
		cells := grid[i]
		if rowEmpty(cells) {
			continue // blank separator rows are not records
		}
		r := importRow{excelRow: i + 1}
		cellAt := func(field string) string {
			c, ok := cols[field]
			if !ok || c >= len(cells) {
				return ""
			}
			return strings.TrimSpace(cells[c])
		}

		r.member = domain.Member{
			GENNo:               cellAt("gen"),
			Name:                cellAt("name"),
			DOB:                 cellAt("dob"),
			Mobile:              cellAt("mobile"),
			Email:               cellAt("email"),
			PresentAddress:      cellAt("present"),
			PermanentAddress:    cellAt("permanent"),
			EmployerName:        cellAt("employer"),
			Department:          cellAt("department"),
			Designation:         cellAt("designation"),
			TokenNo:             cellAt("token"),
			NomineeName:         cellAt("nominee"),
			NomineeRelationship: cellAt("relationship"),
			Aadhaar:             cellAt("aadhaar"),
			PAN:                 cellAt("pan"),
			BankName:            cellAt("bank"),
			AccountNo:           cellAt("account"),
			IFSC:                cellAt("ifsc"),
			ProfileRemarks:      cellAt("remarks"),
		}

		// Track GEN Nos. already encountered in this file. A present GEN
		// counts as seen even when Name is blank, so later repeats are
		// still flagged as duplicates (SRS §52.3).
		dupInFile := false
		if r.member.GENNo != "" {
			key := strings.ToUpper(r.member.GENNo)
			dupInFile = seen[key]
			seen[key] = true
		}

		// Mandatory data (SRS §52.1, §52.3).
		if r.member.GENNo == "" || r.member.Name == "" {
			r.status = api.ImportRowInvalid
			r.missingMandatory = true
			if r.member.GENNo == "" {
				r.errText = "GEN No. is required."
			} else {
				r.errText = "Name is required."
			}
			rows = append(rows, r)
			continue
		}

		// Duplicate GEN No. within the file (SRS §52.3).
		if dupInFile {
			r.status = api.ImportRowDuplicate
			r.errText = "GEN No. is repeated in this file."
			rows = append(rows, r)
			continue
		}

		// Already in the system — skipped, never overwritten (SRS §52.4).
		if existing[strings.ToUpper(r.member.GENNo)] {
			r.status = api.ImportRowExisting
			r.errText = "GEN No. already exists in the system."
			rows = append(rows, r)
			continue
		}

		// Format validation (SRS §52.3).
		if msg := validateImportFormats(&r.member); msg != "" {
			r.status = api.ImportRowInvalid
			r.errText = msg
			rows = append(rows, r)
			continue
		}

		r.status = api.ImportRowNew
		rows = append(rows, r)
	}
	return rows, nil
}

// existingGENs loads every GEN No. currently in the member master, keyed
// case-insensitively so a GEN matching only in case is still recognised.
func (s *MemberService) existingGENs() (map[string]bool, error) {
	rows, err := s.db.Query("SELECT gen_no FROM members")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	set := map[string]bool{}
	for rows.Next() {
		var gen string
		if err := rows.Scan(&gen); err != nil {
			return nil, err
		}
		set[strings.ToUpper(gen)] = true
	}
	return set, rows.Err()
}

// importHeaderAliases maps normalized header text (lowercased, with spaces,
// dots, underscores and hyphens removed) to import field keys. Headers come
// straight from user files, so common spellings are accepted.
var importHeaderAliases = map[string]string{
	"genno": "gen", "gennum": "gen", "gen": "gen", "genumber": "gen",
	"name": "name", "membername": "name",
	"dob": "dob", "dateofbirth": "dob",
	"mobile": "mobile", "mobileno": "mobile", "phonenumber": "mobile",
	"email": "email", "emailid": "email",
	"presentaddress": "present",
	"permanentaddress": "permanent",
	"employername": "employer", "employer": "employer",
	"department": "department",
	"designation": "designation",
	"tokenno": "token", "token": "token",
	"nomineename": "nominee",
	"nomineerelationship": "relationship",
	"aadhaar": "aadhaar", "aadhaarno": "aadhaar", "aadhaarnumber": "aadhaar",
	"pan": "pan", "panno": "pan", "pannumber": "pan",
	"bankname": "bank", "bank": "bank",
	"accountno": "account", "accountnumber": "account",
	"ifsc": "ifsc", "ifsccode": "ifsc",
	"profileremarks": "remarks", "remarks": "remarks",
}

// importHeaderColumns maps field keys to column indexes from the header row,
// requiring the two mandatory columns of SRS §52.1.
func importHeaderColumns(header []string) (map[string]int, error) {
	cols := map[string]int{}
	for i, h := range header {
		norm := normalizeHeader(h)
		if norm == "" {
			continue
		}
		if field, ok := importHeaderAliases[norm]; ok {
			if _, dup := cols[field]; !dup {
				cols[field] = i
			}
		}
	}
	if _, ok := cols["gen"]; !ok {
		return nil, fmt.Errorf(`The Excel file must contain the column "GEN No.".`)
	}
	if _, ok := cols["name"]; !ok {
		return nil, fmt.Errorf(`The Excel file must contain the column "Name".`)
	}
	return cols, nil
}

func normalizeHeader(h string) string {
	r := strings.NewReplacer(" ", "", "\t", "", ".", "", "_", "", "-", "")
	return strings.ToLower(r.Replace(strings.TrimSpace(h)))
}

func rowEmpty(cells []string) bool {
	for _, c := range cells {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

var (
	dobLayouts = []string{
		"2006-01-02", "2006/01/02", "02-01-2006", "02/01/2006", "02.01.2006",
		"02-Jan-2006", "02-January-2006", "2 Jan 2006", "02 Jan 2006",
		"January 2, 2006", "Jan 2, 2006",
	}
	mobilePattern   = regexp.MustCompile(`^\d{6,15}$`)
	aadhaarPattern  = regexp.MustCompile(`^\d{12}$`)
	panPattern      = regexp.MustCompile(`^[A-Za-z]{5}\d{4}[A-Za-z]$`)
	ifscPattern     = regexp.MustCompile(`^[A-Za-z]{4}0[A-Za-z0-9]{6}$`)
	accountPattern  = regexp.MustCompile(`^[A-Za-z0-9-]{4,30}$`)
	excelSerialPat  = regexp.MustCompile(`^\d{4,6}$`)
)

// validateImportFormats checks and normalises the optional fields that carry
// a format (SRS §52.3). It returns a user-facing message for the first
// problem found, or "" when the row is clean.
func validateImportFormats(m *domain.Member) string {
	if m.DOB != "" {
		iso, err := normalizeDOB(m.DOB)
		if err != nil {
			return "DOB is not a valid date."
		}
		m.DOB = iso
	}
	if m.Mobile != "" {
		digits := strings.NewReplacer(" ", "", "-", "", "+", "").Replace(m.Mobile)
		if !mobilePattern.MatchString(digits) {
			return "Mobile No. must be 6 to 15 digits."
		}
	}
	if m.Email != "" && !validEmail(m.Email) {
		return "Email ID is not valid."
	}
	if m.Aadhaar != "" {
		digits := strings.ReplaceAll(m.Aadhaar, " ", "")
		if !aadhaarPattern.MatchString(digits) {
			return "Aadhaar No. must be 12 digits."
		}
		m.Aadhaar = digits
	}
	if m.PAN != "" && !panPattern.MatchString(m.PAN) {
		return "PAN No. is not valid."
	}
	m.PAN = strings.ToUpper(m.PAN)
	if m.IFSC != "" && !ifscPattern.MatchString(strings.ToUpper(m.IFSC)) {
		return "IFSC Code is not valid."
	}
	m.IFSC = strings.ToUpper(m.IFSC)
	if m.AccountNo != "" {
		acct := strings.ReplaceAll(m.AccountNo, " ", "")
		if !accountPattern.MatchString(acct) {
			return "Account No. is not valid."
		}
	}
	return ""
}

// normalizeDOB parses common date spellings (including Excel serials) into
// the ISO form stored by the application.
func normalizeDOB(s string) (string, error) {
	s = strings.TrimSpace(s)
	if excelSerialPat.MatchString(s) {
		serial, err := strconv.Atoi(s)
		if err == nil && serial > 0 {
			base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
			return base.AddDate(0, 0, serial).Format("2006-01-02"), nil
		}
	}
	for _, layout := range dobLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("invalid date: %s", s)
}

func validEmail(s string) bool {
	at := strings.Index(s, "@")
	if at <= 0 || at == len(s)-1 {
		return false
	}
	domainPart := s[at+1:]
	return strings.Contains(domainPart, ".") && !strings.HasPrefix(domainPart, ".") && !strings.HasSuffix(domainPart, ".")
}

// ---------------------------------------------------------------------------
// Converters
// ---------------------------------------------------------------------------

func memberFromRequest(req api.SaveMemberRequest) domain.Member {
	return domain.Member{
		ID:                  req.ID,
		GENNo:               req.GENNo,
		Name:                req.Name,
		DOB:                 req.DOB,
		Mobile:              req.Mobile,
		Email:               req.Email,
		PresentAddress:      req.PresentAddress,
		PermanentAddress:    req.PermanentAddress,
		EmployerName:        req.EmployerName,
		Department:          req.Department,
		Designation:         req.Designation,
		TokenNo:             req.TokenNo,
		NomineeName:         req.NomineeName,
		NomineeRelationship: req.NomineeRelationship,
		Aadhaar:             req.Aadhaar,
		PAN:                 req.PAN,
		BankName:            req.BankName,
		AccountNo:           req.AccountNo,
		IFSC:                req.IFSC,
		ProfileRemarks:      req.ProfileRemarks,
	}
}

func toMember(m domain.Member) api.Member {
	return api.Member{
		ID:                  m.ID,
		GENNo:               m.GENNo,
		Name:                m.Name,
		DOB:                 m.DOB,
		Mobile:              m.Mobile,
		Email:               m.Email,
		PresentAddress:      m.PresentAddress,
		PermanentAddress:    m.PermanentAddress,
		EmployerName:        m.EmployerName,
		Department:          m.Department,
		Designation:         m.Designation,
		TokenNo:             m.TokenNo,
		NomineeName:         m.NomineeName,
		NomineeRelationship: m.NomineeRelationship,
		Aadhaar:             m.Aadhaar,
		PAN:                 m.PAN,
		BankName:            m.BankName,
		AccountNo:           m.AccountNo,
		IFSC:                m.IFSC,
		ProfileRemarks:      m.ProfileRemarks,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func toMemberWithStats(m domain.MemberWithStats) api.Member {
	out := toMember(m.Member)
	out.ActiveFDCount = m.ActiveFDCount
	out.ActiveFDAmount = m.ActiveFDAmount
	return out
}
