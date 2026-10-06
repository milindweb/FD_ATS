package service

import (
	"time"

	"fdats/internal/api"
	"fdats/internal/domain"
)

// MaturityChart buckets active FD maturities by calendar month for the next
// 12 months, starting with the current month. FDs past maturity that are
// still active appear in the current month's bucket.
func (s *FDService) MaturityChart() ([]api.MaturityBucket, error) {
	today := s.nowUTC()
	from := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 12, 0).AddDate(0, 0, -1)

	buckets := make([]api.MaturityBucket, 0, 12)
	index := make(map[string]int, 12)
	for i := 0; i < 12; i++ {
		m := from.AddDate(0, i, 0)
		period := m.Format("2006-01")
		index[period] = len(buckets)
		buckets = append(buckets, api.MaturityBucket{
			Period: period,
			Label:  m.Format("Jan"),
		})
	}

	fds, err := s.fds.ListUpcoming(domain.FormatDate(from), domain.FormatDate(to))
	if err != nil {
		return nil, err
	}
	for _, fd := range fds {
		if len(fd.MaturityDate) < 7 {
			continue
		}
		if i, ok := index[fd.MaturityDate[:7]]; ok {
			buckets[i].Count++
			buckets[i].Amount += fd.MaturityAmount
		}
	}
	return buckets, nil
}
