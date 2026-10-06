package service

import (
	"fdats/internal/api"
	"fdats/internal/domain"
	"fdats/internal/repo"
)

// sampleSpec describes one demo FD to create. Offsets are relative to today
// so the dataset always demonstrates near/far maturities.
type sampleSpec struct {
	name         string
	number       string
	principal    int64
	startDaysAgo int
	tenureDays   int
}

// LoadSampleData inserts a small demo dataset for evaluating the UI before
// real data exists.
//
// TEMPORARY: this whole method (plus its Settings button) is planned for
// complete removal once the application holds real data. It only runs on a
// completely empty database.
func (s *FDService) LoadSampleData() (int, error) {
	_, total, err := s.fds.List(repo.ListParams{Filter: domain.FilterAll, Page: 1, PageSize: 1})
	if err != nil {
		return 0, err
	}
	if total > 0 {
		return 0, domain.ErrSampleDataNotEmpty
	}

	today := s.nowUTC()
	todayISO := domain.FormatDate(today)

	specs := []sampleSpec{
		{"Asha Patil", "M-1001", 150000, 358, 365},    // matures in 7 days
		{"Ravi Kumar", "M-1002", 250000, 1090, 1095},  // matures in 5 days
		{"Meera Joshi", "M-1003", 50000, 344, 365},    // matures in 21 days
		{"Suresh Patil", "M-1004", 500000, 700, 730},  // matures in 30 days
		{"Anita Deshmukh", "M-1005", 75000, 275, 365}, // matures in 90 days
		{"Vikram Rao", "M-1006", 1000000, 640, 730},   // matures in 90 days
		{"Sneha Kulkarni", "M-1007", 350000, 100, 1095},
		{"Amit Sharma", "M-1008", 20000, 20, 180},
		{"Pooja Mehta", "M-1009", 125000, 400, 730},
		{"Rahul Verma", "M-1010", 60000, 500, 365},  // matured 135 days ago, still active
		{"Divya Nair", "M-1011", 85000, 300, 365},   // closed prematurely below
		{"Karan Singh", "M-1012", 300000, 800, 730}, // closed at maturity below
	}

	created := make(map[string]api.FD, len(specs))
	for _, sp := range specs {
		fd, err := s.Create(api.PreviewRequest{
			CustomerName:   sp.name,
			CustomerNumber: sp.number,
			Principal:      sp.principal,
			StartDate:      domain.FormatDate(domain.AddDays(today, -sp.startDaysAgo)),
			TenureDays:     sp.tenureDays,
		})
		if err != nil {
			return s.countFDs(), err
		}
		created[sp.number] = fd
	}

	// One premature closure (held 300 of 365 days).
	if _, err := s.Close(api.CloseRequest{
		FDNumber:    created["M-1011"].FDNumber,
		ClosureDate: todayISO,
		Remark:      "Closed early for a family expense",
	}); err != nil {
		return s.countFDs(), err
	}

	// One closure at maturity (matured 70 days ago).
	if _, err := s.Close(api.CloseRequest{
		FDNumber:    created["M-1012"].FDNumber,
		ClosureDate: domain.FormatDate(domain.AddDays(today, -70)),
		Remark:      "Matured deposit paid out",
	}); err != nil {
		return s.countFDs(), err
	}

	// One renewal chain: principal + interest for another two years.
	if _, err := s.Renew(api.RenewRequest{
		FDNumber:   created["M-1009"].FDNumber,
		Mode:       domain.RenewPrincipalPlusInterest,
		StartDate:  todayISO,
		TenureDays: 730,
		Remark:     "Renewed for another two years",
	}); err != nil {
		return s.countFDs(), err
	}

	return s.countFDs(), nil
}

// countFDs returns the total number of FD rows currently stored.
func (s *FDService) countFDs() int {
	_, total, err := s.fds.List(repo.ListParams{Filter: domain.FilterAll, Page: 1, PageSize: 1})
	if err != nil {
		return 0
	}
	return total
}
