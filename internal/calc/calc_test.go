package calc

import (
	"errors"
	"testing"
	"time"

	"fdats/internal/domain"
)

func defaultSlabs() []domain.RateSlab {
	return repoDefaults()
}

// repoDefaults mirrors the SRS §13 default configuration.
func repoDefaults() []domain.RateSlab {
	return []domain.RateSlab{
		{SortOrder: 1, MinDays: 1, MaxDays: 364, RatePercent: 4.00, Label: "1-364 Days"},
		{SortOrder: 2, MinDays: 365, MaxDays: 729, RatePercent: 8.00, Label: "365-729 Days"},
		{SortOrder: 3, MinDays: 730, MaxDays: 1094, RatePercent: 8.50, Label: "730-1094 Days"},
		{SortOrder: 4, MinDays: 1095, MaxDays: 1460, RatePercent: 9.50, Label: "1095-1460 Days"},
	}
}

func TestRateForDaysMatchesSRSSlabBoundaries(t *testing.T) {
	slabs := defaultSlabs()

	cases := []struct {
		days int
		want float64
	}{
		{1, 4.00},
		{100, 4.00},
		{364, 4.00},
		{365, 8.00},
		{500, 8.00},
		{729, 8.00},
		{730, 8.50},
		{1094, 8.50},
		{1095, 9.50},
		{1460, 9.50},
	}
	for _, c := range cases {
		got, err := RateForDays(slabs, c.days)
		if err != nil {
			t.Fatalf("days=%d unexpected error: %v", c.days, err)
		}
		if got != c.want {
			t.Errorf("days=%d rate=%v, want %v", c.days, got, c.want)
		}
	}
}

func TestRateForDaysOutsideConfiguredRange(t *testing.T) {
	for _, days := range []int{0, -5, 1461} {
		_, err := RateForDays(defaultSlabs(), days)
		if !errors.Is(err, domain.ErrNoRateForTenure) {
			t.Errorf("days=%d err=%v, want ErrNoRateForTenure", days, err)
		}
	}
}

func TestSimpleInterestSRSExample(t *testing.T) {
	// SRS §16: 1,00,000 x 8 x 365 / 365 / 100 = 8,000
	if got := SimpleInterest(100000, 8, 365); got != 8000 {
		t.Errorf("interest=%d, want 8000", got)
	}
}

func TestSimpleInterestRoundsToNearestRupee(t *testing.T) {
	// SRS §24: 50,000 x 4 x 200 / 365 / 100 = 1095.89 → 1096
	if got := SimpleInterest(50000, 4, 200); got != 1096 {
		t.Errorf("interest=%d, want 1096", got)
	}
}

func TestSimpleInterestGuards(t *testing.T) {
	if got := SimpleInterest(100000, 8, 0); got != 0 {
		t.Errorf("zero days: %d, want 0", got)
	}
	if got := SimpleInterest(0, 8, 100); got != 0 {
		t.Errorf("zero principal: %d, want 0", got)
	}
	if got := SimpleInterest(100000, 0, 100); got != 0 {
		t.Errorf("zero rate: %d, want 0", got)
	}
}

func TestComputeSRSPreviewExample(t *testing.T) {
	// SRS §10 preview: 1,00,000 on 01/10/2026 for 365 days at 8%.
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	res, err := Compute(100000, start, 365, defaultSlabs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RatePercent != 8.00 {
		t.Errorf("rate=%v, want 8.00", res.RatePercent)
	}
	if res.Interest != 8000 {
		t.Errorf("interest=%d, want 8000", res.Interest)
	}
	if got := domain.FormatDate(res.MaturityDate); got != "2027-10-01" {
		t.Errorf("maturity=%s, want 2027-10-01", got)
	}
	if res.MaturityAmount != 108000 {
		t.Errorf("maturity amount=%d, want 108000", res.MaturityAmount)
	}
}

func TestComputeRejectsInvalidInput(t *testing.T) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	if _, err := Compute(0, start, 365, defaultSlabs()); !errors.Is(err, domain.ErrInvalidAmount) {
		t.Errorf("zero principal err=%v", err)
	}
	if _, err := Compute(100000, time.Time{}, 365, defaultSlabs()); !errors.Is(err, domain.ErrStartDateRequired) {
		t.Errorf("zero start err=%v", err)
	}
	if _, err := Compute(100000, start, 0, defaultSlabs()); !errors.Is(err, domain.ErrInvalidTenure) {
		t.Errorf("zero tenure err=%v", err)
	}
	if _, err := Compute(100000, start, 5000, defaultSlabs()); !errors.Is(err, domain.ErrNoRateForTenure) {
		t.Errorf("unconfigured tenure err=%v", err)
	}
}

func TestComputeClosureMaturityCappedAtMaturityAmount(t *testing.T) {
	// SRS §25: closure on/after maturity pays exactly the maturity amount.
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	maturity := time.Date(2027, 10, 1, 0, 0, 0, 0, time.UTC)
	closure := time.Date(2027, 10, 20, 0, 0, 0, 0, time.UTC)

	res, err := ComputeClosure(100000, 8.00, start, maturity, 8000, 108000, closure, defaultSlabs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Type != domain.ClosureMatured {
		t.Errorf("type=%s, want MATURED", res.Type)
	}
	if res.Payable != 108000 {
		t.Errorf("payable=%d, want 108000", res.Payable)
	}
	if res.Interest != 8000 {
		t.Errorf("interest=%d, want 8000", res.Interest)
	}
}

func TestComputeClosureMaturityExactlyOnDate(t *testing.T) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	maturity := time.Date(2027, 10, 1, 0, 0, 0, 0, time.UTC)

	res, err := ComputeClosure(100000, 8.00, start, maturity, 8000, 108000, maturity, defaultSlabs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Type != domain.ClosureMatured {
		t.Errorf("type=%s, want MATURED", res.Type)
	}
	if res.Payable != 108000 {
		t.Errorf("payable=%d, want 108000", res.Payable)
	}
}

func TestComputeClosurePrematureUsesActualDaysHeld(t *testing.T) {
	// SRS §24: 50,000 held 200 days, rate for 200 days = 4%.
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	maturity := domain.AddDays(start, 365)
	closure := domain.AddDays(start, 200)

	res, err := ComputeClosure(50000, 8.00, start, maturity, 4000, 54000, closure, defaultSlabs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Type != domain.ClosurePremature {
		t.Errorf("type=%s, want PREMATURE", res.Type)
	}
	if !res.IsPremature {
		t.Error("IsPremature=false, want true")
	}
	if res.DaysHeld != 200 {
		t.Errorf("daysHeld=%d, want 200", res.DaysHeld)
	}
	if res.RatePercent != 4.00 {
		t.Errorf("rate=%v, want 4.00", res.RatePercent)
	}
	if res.Interest != 1096 {
		t.Errorf("interest=%d, want 1096", res.Interest)
	}
	if res.Payable != 51096 {
		t.Errorf("payable=%d, want 51096", res.Payable)
	}
}

func TestComputeClosureBeforeStartRejected(t *testing.T) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	maturity := domain.AddDays(start, 365)

	_, err := ComputeClosure(100000, 8.00, start, maturity, 8000, 108000, domain.AddDays(start, -1), defaultSlabs())
	if !errors.Is(err, domain.ErrClosureBeforeStart) {
		t.Errorf("err=%v, want ErrClosureBeforeStart", err)
	}
}

func TestComputeClosureSameDayPaysNoInterest(t *testing.T) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	maturity := domain.AddDays(start, 365)

	res, err := ComputeClosure(100000, 8.00, start, maturity, 8000, 108000, start, defaultSlabs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Interest != 0 || res.Payable != 100000 {
		t.Errorf("interest=%d payable=%d, want 0/100000", res.Interest, res.Payable)
	}
	if !res.IsPremature {
		t.Error("same-day closure should be premature")
	}
}

func TestValidateSlabs(t *testing.T) {
	if err := ValidateSlabs(defaultSlabs()); err != nil {
		t.Errorf("defaults should be valid: %v", err)
	}

	overlap := defaultSlabs()
	overlap[1].MinDays = 300
	if err := ValidateSlabs(overlap); !errors.Is(err, domain.ErrInvalidSlabConfig) {
		t.Errorf("overlap err=%v, want ErrInvalidSlabConfig", err)
	}

	negative := defaultSlabs()
	negative[0].RatePercent = -1
	if err := ValidateSlabs(negative); !errors.Is(err, domain.ErrInvalidSlabConfig) {
		t.Errorf("negative rate err=%v, want ErrInvalidSlabConfig", err)
	}

	invalidRange := defaultSlabs()
	invalidRange[0].MinDays = 0
	if err := ValidateSlabs(invalidRange); !errors.Is(err, domain.ErrInvalidSlabConfig) {
		t.Errorf("zero min days err=%v, want ErrInvalidSlabConfig", err)
	}

	if err := ValidateSlabs(nil); !errors.Is(err, domain.ErrInvalidSlabConfig) {
		t.Errorf("empty err=%v, want ErrInvalidSlabConfig", err)
	}
}

func TestDaysBetweenAndAddDays(t *testing.T) {
	a := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	b := time.Date(2027, 10, 1, 0, 0, 0, 0, time.UTC)
	if got := domain.DaysBetween(a, b); got != 365 {
		t.Errorf("days=%d, want 365", got)
	}
	if got := domain.DaysBetween(b, a); got != -365 {
		t.Errorf("reversed days=%d, want -365", got)
	}
	if got := domain.FormatDate(domain.AddDays(a, 365)); got != "2027-10-01" {
		t.Errorf("add days=%s", got)
	}
}
