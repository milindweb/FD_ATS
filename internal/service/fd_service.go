// Package service implements the FD business flows of the SRS: creation,
// search, renewal, closure, maturity tracking and rate configuration.
// Everything here is UI-independent so it can be tested directly.
package service

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"fdats/internal/api"
	"fdats/internal/calc"
	"fdats/internal/domain"
	"fdats/internal/repo"
)

// MaturingSoonDays is the dashboard "maturing soon" window used by the
// MATURING list filter (SRS §9.1, §28).
const MaturingSoonDays = 30

// insertAttempts bounds the FD number allocation retries under concurrency.
const insertAttempts = 5

// FDService exposes the application use cases.
type FDService struct {
	fds     *repo.FDRepo
	history *repo.HistoryRepo
	slabs   *repo.SlabRepo
	auth    *repo.AuthRepo
	now     func() time.Time
}

// NewFDService wires the service against an open database and seeds the
// default interest rate configuration on first run.
func NewFDService(db *sql.DB) (*FDService, error) {
	slabs := repo.NewSlabRepo(db)
	if err := slabs.EnsureSeeded(); err != nil {
		return nil, fmt.Errorf("seed rate slabs: %w", err)
	}
	return &FDService{
		fds:     repo.NewFDRepo(db),
		history: repo.NewHistoryRepo(db),
		slabs:   slabs,
		auth:    repo.NewAuthRepo(db),
		now:     time.Now,
	}, nil
}

// SetClock replaces the wall clock (tests only).
func (s *FDService) SetClock(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func (s *FDService) nowUTC() time.Time { return s.now().UTC() }

// ---------------------------------------------------------------------------
// Calculation & creation
// ---------------------------------------------------------------------------

// Preview calculates an FD without persisting anything (SRS §10).
func (s *FDService) Preview(req api.PreviewRequest) (api.Calculation, error) {
	res, err := s.computeFromRequest(req)
	if err != nil {
		return api.Calculation{}, err
	}
	return toCalculation(res), nil
}

// Create validates, calculates and saves a new FD (SRS §10, §11).
func (s *FDService) Create(req api.PreviewRequest) (api.FD, error) {
	if err := validateCreate(req); err != nil {
		return api.FD{}, err
	}
	res, err := s.computeFromRequest(req)
	if err != nil {
		return api.FD{}, err
	}

	now := s.nowUTC()
	period := domain.PeriodForYear(res.StartDate.Year())

	fd := domain.FixedDeposit{
		CustomerName:   strings.TrimSpace(req.CustomerName),
		CustomerNumber: strings.TrimSpace(req.CustomerNumber),
		Principal:      res.Principal,
		StartDate:      domain.FormatDate(res.StartDate),
		TenureDays:     res.TenureDays,
		InterestRate:   res.RatePercent,
		MaturityDate:   domain.FormatDate(res.MaturityDate),
		InterestAmount: res.Interest,
		MaturityAmount: res.MaturityAmount,
		Status:         domain.StatusActive,
		CreatedAt:      formatTimestamp(now),
		UpdatedAt:      formatTimestamp(now),
	}

	number, err := s.fds.InsertWithNumber(fd, period, insertAttempts)
	if err != nil {
		return api.FD{}, err
	}
	fd.FDNumber = number

	if err := s.history.Append(domain.HistoryEntry{
		FDNumber:  number,
		EventDate: fd.StartDate,
		EventType: domain.EventOpen,
		Amount:    fd.Principal,
		Remarks:   "FD opened",
		CreatedAt: fd.CreatedAt,
	}); err != nil {
		return api.FD{}, fmt.Errorf("record history: %w", err)
	}

	return toFD(fd), nil
}

// EditFD updates the editable fields of an ACTIVE FD and recomputes the
// derived rate/interest/maturity amounts. Closed FDs are locked (audit).
func (s *FDService) EditFD(req api.EditFDRequest) (api.FD, error) {
	fd, err := s.fds.Get(strings.TrimSpace(req.FDNumber))
	if err != nil {
		return api.FD{}, err
	}
	if fd.Status != domain.StatusActive {
		return api.FD{}, domain.ErrEditClosed
	}

	res, err := s.computeFromRequest(api.PreviewRequest{
		CustomerName:   req.CustomerName,
		CustomerNumber: req.CustomerNumber,
		Principal:      req.Principal,
		StartDate:      req.StartDate,
		TenureDays:     req.TenureDays,
	})
	if err != nil {
		return api.FD{}, err
	}

	fd.CustomerName = strings.TrimSpace(req.CustomerName)
	fd.CustomerNumber = strings.TrimSpace(req.CustomerNumber)
	fd.Principal = res.Principal
	fd.StartDate = domain.FormatDate(res.StartDate)
	fd.TenureDays = res.TenureDays
	fd.InterestRate = res.RatePercent
	fd.MaturityDate = domain.FormatDate(res.MaturityDate)
	fd.InterestAmount = res.Interest
	fd.MaturityAmount = res.MaturityAmount
	fd.UpdatedAt = formatTimestamp(s.nowUTC())

	if err := s.fds.Update(fd); err != nil {
		return api.FD{}, err
	}
	if err := s.history.Append(domain.HistoryEntry{
		FDNumber:  fd.FDNumber,
		EventDate: domain.FormatDate(s.nowUTC()),
		EventType: domain.EventEdit,
		Amount:    fd.Principal,
		Remarks:   "FD details edited",
		CreatedAt: fd.UpdatedAt,
	}); err != nil {
		return api.FD{}, fmt.Errorf("record history: %w", err)
	}
	return toFD(fd), nil
}

func (s *FDService) computeFromRequest(req api.PreviewRequest) (calc.Result, error) {
	if err := validateCreate(req); err != nil {
		return calc.Result{}, err
	}
	start, _ := domain.ParseDate(req.StartDate)
	slabs, err := s.slabs.List()
	if err != nil {
		return calc.Result{}, err
	}
	return calc.Compute(req.Principal, start, req.TenureDays, slabs)
}

func validateCreate(req api.PreviewRequest) error {
	if strings.TrimSpace(req.CustomerName) == "" {
		return domain.ErrCustomerNameRequired
	}
	if strings.TrimSpace(req.CustomerNumber) == "" {
		return domain.ErrCustomerNumberRequired
	}
	if req.Principal <= 0 {
		return domain.ErrInvalidAmount
	}
	if strings.TrimSpace(req.StartDate) == "" {
		return domain.ErrStartDateRequired
	}
	if _, err := domain.ParseDate(req.StartDate); err != nil {
		return domain.ErrInvalidDate
	}
	if req.TenureDays <= 0 {
		return domain.ErrInvalidTenure
	}
	return nil
}

// ---------------------------------------------------------------------------
// Search & list
// ---------------------------------------------------------------------------

// List returns one page of the FD list (SRS §9.2).
func (s *FDService) List(req api.ListRequest) (api.ListResponse, error) {
	filter := strings.ToUpper(strings.TrimSpace(req.Filter))
	if filter == "" {
		filter = domain.FilterAll
	}
	switch filter {
	case domain.FilterAll, domain.FilterActive, domain.FilterClosed, domain.FilterMaturing:
	default:
		return api.ListResponse{}, domain.ErrInvalidFilter
	}

	cutoff := domain.FormatDate(domain.AddDays(s.nowUTC(), MaturingSoonDays))
	items, total, err := s.fds.List(repo.ListParams{
		Search:   req.Search,
		Filter:   filter,
		Cutoff:   cutoff,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return api.ListResponse{}, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 25
	}

	out := make([]api.FD, 0, len(items))
	for _, fd := range items {
		out = append(out, toFD(fd))
	}
	return api.ListResponse{Items: out, Total: total, Page: page, PageSize: pageSize}, nil
}

// Get returns one FD with its full history (SRS §19, §20).
func (s *FDService) Get(fdNumber string) (api.FDDetail, error) {
	fd, err := s.fds.Get(strings.TrimSpace(fdNumber))
	if err != nil {
		return api.FDDetail{}, err
	}
	history, err := s.history.ListForFD(fd.FDNumber)
	if err != nil {
		return api.FDDetail{}, err
	}
	if history == nil {
		history = []domain.HistoryEntry{}
	}
	return api.FDDetail{FD: toFD(fd), History: history}, nil
}

// ---------------------------------------------------------------------------
// Dashboard & maturity tracking
// ---------------------------------------------------------------------------

// Dashboard returns the KPI figures (SRS §9.1, §28).
func (s *FDService) Dashboard() (api.DashboardStats, error) {
	today := s.nowUTC()
	var stats api.DashboardStats

	_, total, err := s.fds.List(repo.ListParams{Filter: domain.FilterAll, Page: 1, PageSize: 1})
	if err != nil {
		return stats, err
	}
	stats.TotalFDs = total

	active, err := s.fds.CountActive()
	if err != nil {
		return stats, err
	}
	stats.ActiveFDs = active

	principal, err := s.fds.SumActivePrincipal()
	if err != nil {
		return stats, err
	}
	stats.ActivePrincipal = principal

	interest, err := s.fds.SumActiveInterest()
	if err != nil {
		return stats, err
	}
	stats.TotalInterest = interest

	stats.FYLabel, stats.FYDeposits, err = s.financialYear(today)
	if err != nil {
		return stats, err
	}

	todayISO := domain.FormatDate(today)
	stats.MaturingToday, err = s.fds.CountMaturingWithin(todayISO)
	if err != nil {
		return stats, err
	}
	for days, field := range map[int]*int{7: &stats.Maturing7, 30: &stats.Maturing30, 90: &stats.Maturing90} {
		cutoff := domain.FormatDate(domain.AddDays(today, days))
		n, err := s.fds.CountMaturingWithin(cutoff)
		if err != nil {
			return stats, err
		}
		*field = n
	}
	return stats, nil
}

// financialYear derives the current Indian financial year (1 Apr – 31 Mar)
// label and the total principal of FDs started inside it.
func (s *FDService) financialYear(now time.Time) (string, int64, error) {
	year := now.Year()
	if now.Month() < time.April {
		year--
	}
	label := fmt.Sprintf("FY %d-%02d", year, (year+1)%100)
	deposits, err := s.fds.SumPrincipalBetween(
		domain.FormatDate(time.Date(year, time.April, 1, 0, 0, 0, 0, time.UTC)),
		domain.FormatDate(time.Date(year+1, time.March, 31, 0, 0, 0, 0, time.UTC)),
	)
	if err != nil {
		return "", 0, err
	}
	return label, deposits, nil
}

// Upcoming lists active FDs maturing in the requested range (SRS §28).
// An empty request defaults to the next 90 days.
func (s *FDService) Upcoming(req api.UpcomingRequest) ([]api.UpcomingFD, error) {
	today := s.nowUTC()
	from := req.FromDate
	if strings.TrimSpace(from) == "" {
		from = domain.FormatDate(today)
	}
	to := req.ToDate
	if strings.TrimSpace(to) == "" {
		to = domain.FormatDate(domain.AddDays(today, 90))
	}
	if _, err := domain.ParseDate(from); err != nil {
		return nil, domain.ErrInvalidDate
	}
	if _, err := domain.ParseDate(to); err != nil {
		return nil, domain.ErrInvalidDate
	}

	fds, err := s.fds.ListUpcoming(from, to)
	if err != nil {
		return nil, err
	}

	out := make([]api.UpcomingFD, 0, len(fds))
	for _, fd := range fds {
		maturity, _ := domain.ParseDate(fd.MaturityDate)
		out = append(out, api.UpcomingFD{
			FDNumber:       fd.FDNumber,
			CustomerName:   fd.CustomerName,
			Principal:      fd.Principal,
			MaturityDate:   fd.MaturityDate,
			MaturityAmount: fd.MaturityAmount,
			DaysRemaining:  domain.DaysBetween(today, maturity),
			Status:         fd.Status,
		})
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Renewal
// ---------------------------------------------------------------------------

// Renew closes the source FD as RENEWED and opens a new FD (SRS §21, §22).
func (s *FDService) Renew(req api.RenewRequest) (api.RenewResult, error) {
	fd, err := s.fds.Get(strings.TrimSpace(req.FDNumber))
	if err != nil {
		return api.RenewResult{}, err
	}
	if fd.Status != domain.StatusActive {
		return api.RenewResult{}, domain.ErrClosedCannotRenew
	}

	mode := strings.ToUpper(strings.TrimSpace(req.Mode))
	if mode != domain.RenewPrincipalOnly && mode != domain.RenewPrincipalPlusInterest {
		return api.RenewResult{}, domain.ErrInvalidRenewalMode
	}
	if req.TenureDays <= 0 {
		return api.RenewResult{}, domain.ErrInvalidTenure
	}
	// Renewal itself is only possible once the maturity date has been
	// reached; until then the deposit is still invested.
	if domain.FormatDate(s.nowUTC()) < fd.MaturityDate {
		return api.RenewResult{}, domain.ErrRenewNotMatured
	}

	startISO := strings.TrimSpace(req.StartDate)
	if startISO == "" {
		startISO = fd.MaturityDate // default: renew from maturity (SRS §22)
	}
	start, err := domain.ParseDate(startISO)
	if err != nil {
		return api.RenewResult{}, domain.ErrInvalidDate
	}
	if domain.DaysBetween(mustParse(fd.StartDate), start) < 0 {
		return api.RenewResult{}, domain.ErrClosureBeforeStart
	}
	oldMaturity, _ := domain.ParseDate(fd.MaturityDate)
	if domain.DaysBetween(oldMaturity, start) < 0 {
		return api.RenewResult{}, domain.ErrRenewBeforeMaturity
	}

	slabs, err := s.slabs.List()
	if err != nil {
		return api.RenewResult{}, err
	}
	newRate, err := calc.RateForDays(slabs, req.TenureDays)
	if err != nil {
		return api.RenewResult{}, err
	}

	// Earned interest on the old FD as of the renewal date (SRS §21).
	earned, err := calc.ComputeClosure(
		fd.Principal, fd.InterestRate,
		mustParse(fd.StartDate), oldMaturity,
		fd.InterestAmount, fd.MaturityAmount,
		start, slabs,
	)
	if err != nil {
		return api.RenewResult{}, err
	}

	newPrincipal := fd.Principal
	if mode == domain.RenewPrincipalPlusInterest {
		newPrincipal = fd.Principal + earned.Interest
	}
	newInterest := calc.SimpleInterest(newPrincipal, newRate, req.TenureDays)

	now := s.nowUTC()
	period := domain.PeriodForYear(start.Year())

	newFD := domain.FixedDeposit{
		CustomerName:   fd.CustomerName,
		CustomerNumber: fd.CustomerNumber,
		Principal:      newPrincipal,
		StartDate:      domain.FormatDate(start),
		TenureDays:     req.TenureDays,
		InterestRate:   newRate,
		MaturityDate:   domain.FormatDate(calc.MaturityDate(start, req.TenureDays)),
		InterestAmount: newInterest,
		MaturityAmount: newPrincipal + newInterest,
		Status:         domain.StatusActive,
		RenewedFrom:    &fd.FDNumber,
		CreatedAt:      formatTimestamp(now),
		UpdatedAt:      formatTimestamp(now),
	}

	newNumber, err := s.fds.InsertWithNumber(newFD, period, insertAttempts)
	if err != nil {
		return api.RenewResult{}, err
	}
	newFD.FDNumber = newNumber

	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		remark = fmt.Sprintf("Renewed as %s", newNumber)
	} else {
		remark = fmt.Sprintf("%s (renewed as %s)", remark, newNumber)
	}

	fd.Status = domain.StatusClosed
	closureType := domain.ClosureRenewed
	fd.ClosureType = &closureType
	fd.ClosureDate = &newFD.StartDate
	fd.ClosureRemark = remark
	rate := earned.RatePercent
	days := earned.DaysHeld
	interest := earned.Interest
	payable := earned.Payable
	fd.ClosureRate = &rate
	fd.ClosureDays = &days
	fd.ClosureInterest = &interest
	fd.ClosurePayable = &payable
	fd.RenewedTo = &newNumber
	fd.UpdatedAt = formatTimestamp(now)

	if err := s.fds.Update(fd); err != nil {
		return api.RenewResult{}, err
	}

	if err := s.history.Append(domain.HistoryEntry{
		FDNumber:    fd.FDNumber,
		EventDate:   newFD.StartDate,
		EventType:   domain.EventRenew,
		Amount:      fd.Principal,
		Interest:    interest,
		ReferenceFD: newNumber,
		Remarks:     remark,
		CreatedAt:   formatTimestamp(now),
	}); err != nil {
		return api.RenewResult{}, err
	}

	if err := s.history.Append(domain.HistoryEntry{
		FDNumber:  newNumber,
		EventDate: newFD.StartDate,
		EventType: domain.EventOpen,
		Amount:    newPrincipal,
		Remarks:   fmt.Sprintf("Renewed from %s", fd.FDNumber),
		CreatedAt: formatTimestamp(now),
	}); err != nil {
		return api.RenewResult{}, err
	}

	return api.RenewResult{PreviousFD: toFD(fd), NewFD: toFD(newFD)}, nil
}

// ---------------------------------------------------------------------------
// Closure
// ---------------------------------------------------------------------------

// PreviewClosure calculates the payable amount for a prospective closure
// without saving (SRS §26).
func (s *FDService) PreviewClosure(req api.CloseRequest) (api.ClosurePreview, error) {
	fd, res, err := s.prepareClosure(req)
	if err != nil {
		return api.ClosurePreview{}, err
	}
	return api.ClosurePreview{
		FDNumber:     fd.FDNumber,
		ClosureDate:  req.ClosureDate,
		IsPremature:  res.IsPremature,
		ClosureType:  res.Type,
		DaysHeld:     res.DaysHeld,
		RatePercent:  res.RatePercent,
		Interest:     res.Interest,
		Payable:      res.Payable,
		MaturityDate: fd.MaturityDate,
	}, nil
}

// Close closes an FD either prematurely or at maturity (SRS §23–§26).
func (s *FDService) Close(req api.CloseRequest) (api.FD, error) {
	fd, res, err := s.prepareClosure(req)
	if err != nil {
		return api.FD{}, err
	}
	if res.Type == domain.ClosurePremature && strings.TrimSpace(req.Remark) == "" {
		return api.FD{}, domain.ErrClosureRemarkRequired
	}

	now := s.nowUTC()
	closureType := res.Type
	payable := res.Payable
	interest := res.Interest
	rate := res.RatePercent
	days := res.DaysHeld

	fd.Status = domain.StatusClosed
	fd.ClosureDate = &req.ClosureDate
	fd.ClosureType = &closureType
	fd.ClosureRemark = strings.TrimSpace(req.Remark)
	fd.ClosureRate = &rate
	fd.ClosureDays = &days
	fd.ClosureInterest = &interest
	fd.ClosurePayable = &payable
	fd.UpdatedAt = formatTimestamp(now)

	if err := s.fds.Update(fd); err != nil {
		return api.FD{}, err
	}

	if err := s.history.Append(domain.HistoryEntry{
		FDNumber:  fd.FDNumber,
		EventDate: req.ClosureDate,
		EventType: domain.EventClose,
		Amount:    payable,
		Interest:  interest,
		Remarks:   fd.ClosureRemark,
		CreatedAt: formatTimestamp(now),
	}); err != nil {
		return api.FD{}, err
	}

	return toFD(fd), nil
}

func (s *FDService) prepareClosure(req api.CloseRequest) (domain.FixedDeposit, calc.ClosureResult, error) {
	fd, err := s.fds.Get(strings.TrimSpace(req.FDNumber))
	if err != nil {
		return domain.FixedDeposit{}, calc.ClosureResult{}, err
	}
	if fd.Status != domain.StatusActive {
		return domain.FixedDeposit{}, calc.ClosureResult{}, domain.ErrAlreadyClosed
	}
	if strings.TrimSpace(req.ClosureDate) == "" {
		return domain.FixedDeposit{}, calc.ClosureResult{}, domain.ErrClosureDateRequired
	}
	closureDate, err := domain.ParseDate(req.ClosureDate)
	if err != nil {
		return domain.FixedDeposit{}, calc.ClosureResult{}, domain.ErrInvalidDate
	}

	slabs, err := s.slabs.List()
	if err != nil {
		return domain.FixedDeposit{}, calc.ClosureResult{}, err
	}

	res, err := calc.ComputeClosure(
		fd.Principal, fd.InterestRate,
		mustParse(fd.StartDate), mustParse(fd.MaturityDate),
		fd.InterestAmount, fd.MaturityAmount,
		closureDate, slabs,
	)
	if err != nil {
		return domain.FixedDeposit{}, calc.ClosureResult{}, err
	}
	return fd, res, nil
}

// ---------------------------------------------------------------------------
// Reversals
// ---------------------------------------------------------------------------

// Reopen returns a closed FD to ACTIVE status. The remark is the mandatory
// audit reason; renewals must be undone with ReverseRenewal instead.
func (s *FDService) Reopen(req api.ReopenRequest) (api.FD, error) {
	fd, err := s.fds.Get(strings.TrimSpace(req.FDNumber))
	if err != nil {
		return api.FD{}, err
	}
	if fd.Status != domain.StatusClosed {
		return api.FD{}, domain.ErrNotClosed
	}
	if fd.ClosureType != nil && *fd.ClosureType == domain.ClosureRenewed {
		return api.FD{}, domain.ErrRenewedCannotReopen
	}
	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		return api.FD{}, domain.ErrReversalReasonRequired
	}

	now := s.nowUTC()
	fd.Status = domain.StatusActive
	fd.ClosureDate = nil
	fd.ClosureType = nil
	fd.ClosureRemark = ""
	fd.ClosureRate = nil
	fd.ClosureDays = nil
	fd.ClosureInterest = nil
	fd.ClosurePayable = nil
	fd.UpdatedAt = formatTimestamp(now)

	if err := s.fds.Update(fd); err != nil {
		return api.FD{}, err
	}
	if err := s.history.Append(domain.HistoryEntry{
		FDNumber:  fd.FDNumber,
		EventDate: domain.FormatDate(now),
		EventType: domain.EventReopen,
		Remarks:   remark,
		CreatedAt: formatTimestamp(now),
	}); err != nil {
		return api.FD{}, err
	}
	return toFD(fd), nil
}

// ReverseRenewal withdraws the renewed FD and reopens the previous one in a
// single transaction. It only works while the renewed FD is still untouched.
func (s *FDService) ReverseRenewal(req api.ReverseRenewalRequest) (api.FD, error) {
	fd, err := s.fds.Get(strings.TrimSpace(req.FDNumber))
	if err != nil {
		return api.FD{}, err
	}
	if fd.Status != domain.StatusClosed || fd.ClosureType == nil || *fd.ClosureType != domain.ClosureRenewed {
		return api.FD{}, domain.ErrNotRenewed
	}
	if fd.RenewedTo == nil || *fd.RenewedTo == "" {
		return api.FD{}, domain.ErrNotRenewed
	}

	withdrawn, err := s.fds.Get(*fd.RenewedTo)
	if err != nil {
		return api.FD{}, err
	}
	if withdrawn.Status != domain.StatusActive {
		return api.FD{}, domain.ErrRenewalAlreadyUsed
	}

	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		return api.FD{}, domain.ErrReversalReasonRequired
	}

	now := s.nowUTC()
	fd.Status = domain.StatusActive
	fd.ClosureDate = nil
	fd.ClosureType = nil
	fd.ClosureRemark = ""
	fd.ClosureRate = nil
	fd.ClosureDays = nil
	fd.ClosureInterest = nil
	fd.ClosurePayable = nil
	fd.RenewedTo = nil
	fd.UpdatedAt = formatTimestamp(now)

	entry := domain.HistoryEntry{
		FDNumber:    fd.FDNumber,
		EventDate:   domain.FormatDate(now),
		EventType:   domain.EventReverse,
		ReferenceFD: withdrawn.FDNumber,
		Remarks:     remark,
		CreatedAt:   formatTimestamp(now),
	}
	if err := s.fds.ReverseRenewal(fd, withdrawn.FDNumber, entry); err != nil {
		return api.FD{}, err
	}
	return toFD(fd), nil
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

// RateSlabs returns the current interest rate configuration (SRS §13, §30).
func (s *FDService) RateSlabs() ([]domain.RateSlab, error) {
	return s.slabs.List()
}

// SaveRateSlabs validates and replaces the interest rate configuration.
func (s *FDService) SaveRateSlabs(slabs []domain.RateSlab) error {
	normalised := make([]domain.RateSlab, 0, len(slabs))
	for i, slab := range slabs {
		slab.SortOrder = i + 1
		if slab.Label == "" {
			slab.Label = fmt.Sprintf("%d-%d Days", slab.MinDays, slab.MaxDays)
		}
		normalised = append(normalised, slab)
	}
	if err := calc.ValidateSlabs(normalised); err != nil {
		return err
	}
	return s.slabs.Replace(normalised)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mustParse(s string) time.Time {
	t, _ := domain.ParseDate(s)
	return t
}

func formatTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func toFD(fd domain.FixedDeposit) api.FD {
	return api.FD{
		FDNumber:        fd.FDNumber,
		CustomerName:    fd.CustomerName,
		CustomerNumber:  fd.CustomerNumber,
		Principal:       fd.Principal,
		StartDate:       fd.StartDate,
		TenureDays:      fd.TenureDays,
		InterestRate:    fd.InterestRate,
		MaturityDate:    fd.MaturityDate,
		InterestAmount:  fd.InterestAmount,
		MaturityAmount:  fd.MaturityAmount,
		Status:          fd.Status,
		ClosureDate:     fd.ClosureDate,
		ClosureType:     fd.ClosureType,
		ClosureRemark:   fd.ClosureRemark,
		ClosureRate:     fd.ClosureRate,
		ClosureDays:     fd.ClosureDays,
		ClosureInterest: fd.ClosureInterest,
		ClosurePayable:  fd.ClosurePayable,
		RenewedFrom:     fd.RenewedFrom,
		RenewedTo:       fd.RenewedTo,
		CreatedAt:       fd.CreatedAt,
		UpdatedAt:       fd.UpdatedAt,
	}
}

func toCalculation(res calc.Result) api.Calculation {
	return api.Calculation{
		Principal:      res.Principal,
		StartDate:      domain.FormatDate(res.StartDate),
		TenureDays:     res.TenureDays,
		Days:           res.Days,
		RatePercent:    res.RatePercent,
		Interest:       res.Interest,
		MaturityDate:   domain.FormatDate(res.MaturityDate),
		MaturityAmount: res.MaturityAmount,
	}
}
