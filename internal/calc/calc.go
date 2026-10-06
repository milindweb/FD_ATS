// Package calc implements the Fixed Deposit calculation rules from the SRS.
// It is pure logic: no database, no I/O, no configuration of its own — the
// caller always supplies the interest rate slabs to use (SRS §13–§26).
package calc

import (
	"fmt"
	"math"
	"sort"
	"time"

	"fdats/internal/domain"
)

// Result is the outcome of the standard FD calculation (SRS §10, §15–§18).
type Result struct {
	Principal      int64
	StartDate      time.Time
	TenureDays     int
	RatePercent    float64
	Days           int
	Interest       int64
	MaturityDate   time.Time
	MaturityAmount int64
}

// ClosureResult is the outcome of a closure calculation (SRS §23–§25).
type ClosureResult struct {
	Type         domain.ClosureType
	DaysHeld     int
	RatePercent  float64
	Interest     int64
	Payable      int64
	IsPremature  bool
}

// RateForDays returns the annual interest rate configured for the given
// tenure in days (SRS §14). Slabs are matched by their sort order so a
// validated configuration always yields one deterministic answer.
func RateForDays(slabs []domain.RateSlab, days int) (float64, error) {
	ordered := sortedSlabs(slabs)
	for _, s := range ordered {
		if days >= s.MinDays && days <= s.MaxDays {
			return s.RatePercent, nil
		}
	}
	return 0, domain.ErrNoRateForTenure
}

// SimpleInterest implements SRS §15:
//
//	Interest = Principal x Rate x Days / 365 / 100
//
// The result is rounded to the nearest rupee (SRS §17).
func SimpleInterest(principal int64, ratePercent float64, days int) int64 {
	if principal <= 0 || days <= 0 || ratePercent <= 0 {
		return 0
	}
	interest := float64(principal) * ratePercent * float64(days) / 365.0 / 100.0
	return int64(math.Round(interest))
}

// MaturityDate returns start + tenure (SRS §18).
func MaturityDate(start time.Time, tenureDays int) time.Time {
	return domain.AddDays(start, tenureDays)
}

// Compute performs the full open-FD calculation: rate selection from the
// configured slabs, interest, maturity date and maturity amount
// (SRS §10, §14–§18).
func Compute(principal int64, start time.Time, tenureDays int, slabs []domain.RateSlab) (Result, error) {
	if principal <= 0 {
		return Result{}, domain.ErrInvalidAmount
	}
	if start.IsZero() {
		return Result{}, domain.ErrStartDateRequired
	}
	if tenureDays <= 0 {
		return Result{}, domain.ErrInvalidTenure
	}

	rate, err := RateForDays(slabs, tenureDays)
	if err != nil {
		return Result{}, err
	}

	maturity := MaturityDate(start, tenureDays)
	interest := SimpleInterest(principal, rate, tenureDays)

	return Result{
		Principal:      principal,
		StartDate:      start,
		TenureDays:     tenureDays,
		RatePercent:    rate,
		Days:           tenureDays,
		Interest:       interest,
		MaturityDate:   maturity,
		MaturityAmount: principal + interest,
	}, nil
}

// ComputeClosure calculates the payable amount for a closure (SRS §23–§25).
//
//   - closure on/after maturity: payable is capped at the original maturity
//     amount and no extra interest accrues.
//   - closure before maturity: interest uses the actual days held, and the
//     rate is re-selected from the configured slabs for that day count.
func ComputeClosure(
	principal int64,
	originalRate float64,
	start, maturity time.Time,
	maturityInterest, maturityAmount int64,
	closureDate time.Time,
	slabs []domain.RateSlab,
) (ClosureResult, error) {
	if closureDate.IsZero() {
		return ClosureResult{}, domain.ErrClosureDateRequired
	}
	daysHeld := domain.DaysBetween(start, closureDate)
	if daysHeld < 0 {
		return ClosureResult{}, domain.ErrClosureBeforeStart
	}

	if !closureDate.Before(maturity) {
		return ClosureResult{
			Type:        domain.ClosureMatured,
			DaysHeld:    daysHeld,
			RatePercent: originalRate,
			Interest:    maturityInterest,
			Payable:     maturityAmount,
		}, nil
	}

	// Premature closure (SRS §24).
	if daysHeld == 0 {
		// Closed on the start date: no interest has accrued yet.
		return ClosureResult{
			Type:        domain.ClosurePremature,
			DaysHeld:    0,
			RatePercent: 0,
			Interest:    0,
			Payable:     principal,
			IsPremature: true,
		}, nil
	}

	rate, err := RateForDays(slabs, daysHeld)
	if err != nil {
		return ClosureResult{}, err
	}
	interest := SimpleInterest(principal, rate, daysHeld)

	return ClosureResult{
		Type:        domain.ClosurePremature,
		DaysHeld:    daysHeld,
		RatePercent: rate,
		Interest:    interest,
		Payable:     principal + interest,
		IsPremature: true,
	}, nil
}

// ValidateSlabs checks that a slab configuration is usable: positive ranges,
// sane rates and no overlapping ranges (called before saving settings).
func ValidateSlabs(slabs []domain.RateSlab) error {
	if len(slabs) == 0 {
		return domain.ErrInvalidSlabConfig
	}
	seen := make(map[int]bool, len(slabs))
	for _, s := range slabs {
		if s.MinDays < 1 || s.MaxDays < s.MinDays {
			return fmt.Errorf("%w: slab %q has an invalid day range", domain.ErrInvalidSlabConfig, s.Label)
		}
		if s.RatePercent <= 0 {
			return fmt.Errorf("%w: slab %q must have a positive rate", domain.ErrInvalidSlabConfig, s.Label)
		}
		if seen[s.MinDays] {
			return fmt.Errorf("%w: duplicate slab starting at day %d", domain.ErrInvalidSlabConfig, s.MinDays)
		}
		seen[s.MinDays] = true
	}

	ordered := sortedSlabs(slabs)
	for i := 1; i < len(ordered); i++ {
		prev, cur := ordered[i-1], ordered[i]
		if cur.MinDays <= prev.MaxDays {
			return fmt.Errorf("%w: slabs %q and %q overlap", domain.ErrInvalidSlabConfig, prev.Label, cur.Label)
		}
	}
	return nil
}

func sortedSlabs(slabs []domain.RateSlab) []domain.RateSlab {
	out := make([]domain.RateSlab, len(slabs))
	copy(out, slabs)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SortOrder != out[j].SortOrder {
			return out[i].SortOrder < out[j].SortOrder
		}
		return out[i].MinDays < out[j].MinDays
	})
	return out
}
