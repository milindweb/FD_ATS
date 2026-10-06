package domain

import (
	"fmt"
	"time"
)

// DateLayout is the canonical date format used across storage and the UI.
const DateLayout = "2006-01-02"

// ParseDate parses an ISO date string. Empty input yields the zero time and
// no error so optional dates can be passed through.
func ParseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.ParseInLocation(DateLayout, s, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q", ErrInvalidDate, s)
	}
	return t, nil
}

// FormatDate renders a time.Time as an ISO date string.
func FormatDate(t time.Time) string {
	return t.UTC().Format(DateLayout)
}

// DaysBetween returns the number of calendar days from a to b (b - a).
// Both values are normalised to UTC midnight, so the result is exact.
func DaysBetween(a, b time.Time) int {
	au := time.Date(a.UTC().Year(), a.UTC().Month(), a.UTC().Day(), 0, 0, 0, 0, time.UTC)
	bu := time.Date(b.UTC().Year(), b.UTC().Month(), b.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return int(bu.Sub(au).Hours() / 24)
}

// AddDays returns t plus n calendar days.
func AddDays(t time.Time, n int) time.Time {
	return t.UTC().AddDate(0, 0, n)
}

// PeriodForYear builds the FD numbering period key for a year (SRS §11).
func PeriodForYear(year int) string {
	return fmt.Sprintf("%02d", year%100)
}

// FormatFDNumber builds an FD number such as FD-26-001 (SRS §11).
func FormatFDNumber(period string, sequence int) string {
	return fmt.Sprintf("FD-%s-%03d", period, sequence)
}
