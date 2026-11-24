package calc

import (
	"testing"
	"time"
)

func TestMulRate_ZeroPrincipal(t *testing.T) {
	got := mulRate(0, 120_000)
	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestMulRate_ZeroRate(t *testing.T) {
	got := mulRate(100_000, 0)
	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestMulRate_ExactDivision(t *testing.T) {
	// 100_000 * 20_000 / 1_000_000 = 2_000
	got := mulRate(100_000, 20_000)
	if got != 2_000 {
		t.Fatalf("expected 2000, got %d", got)
	}
}

func TestMulRate_RoundsUp(t *testing.T) {
	// 12_345 * 35_000 = 432_075_000 + 500_000 = 432_575_000 / 1_000_000 = 432
	got := mulRate(12_345, 35_000)
	if got != 432 {
		t.Fatalf("expected 432, got %d", got)
	}
}

func TestMulRate_LargeValue(t *testing.T) {
	// R$1 billion = 100_000_000_000 centavos, rate = 1.0 (1_000_000)
	got := mulRate(100_000_000_000, 1_000_000)
	if got != 100_000_000_000 {
		t.Fatalf("expected 100_000_000_000, got %d", got)
	}
}

func TestMulRateDays_ZeroDays(t *testing.T) {
	got := mulRateDays(100_000, 82, 0)
	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestMulRateDays_IOFDaily(t *testing.T) {
	// 100_000 * 82 * 30 = 246_000_000 + 500_000 = 246_500_000 / 1_000_000 = 246
	got := mulRateDays(100_000, 82, 30)
	if got != 246 {
		t.Fatalf("expected 246, got %d", got)
	}
}

func TestDaysBetween(t *testing.T) {
	from := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 2, 25, 0, 0, 0, 0, time.UTC)
	got := daysBetween(from, to)
	if got != 15 {
		t.Fatalf("expected 15, got %d", got)
	}
}

func TestDaysBetween_Negative(t *testing.T) {
	from := time.Date(2024, 2, 25, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	got := daysBetween(from, to)
	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestAddMonths(t *testing.T) {
	base := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	got := addMonths(base, 3)
	expected := time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestAddMonths_ClampsEndOfMonth(t *testing.T) {
	cases := []struct {
		base   time.Time
		months int
		want   time.Time
	}{
		{utcDate(2025, 1, 31), 1, utcDate(2025, 2, 28)},
		{utcDate(2024, 1, 31), 1, utcDate(2024, 2, 29)},
		{utcDate(2025, 1, 31), 2, utcDate(2025, 3, 31)},
		{utcDate(2025, 11, 30), 3, utcDate(2026, 2, 28)},
	}
	for _, c := range cases {
		if got := addMonths(c.base, c.months); !got.Equal(c.want) {
			t.Fatalf("addMonths(%v,%d): want %v, got %v", c.base, c.months, c.want, got)
		}
	}
}

func TestDaysBetween_IgnoresTimeZoneAndDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("tzdata unavailable")
	}
	// Spans the 2024-03-10 DST jump: 23h day must still count as one day.
	from := time.Date(2024, 3, 9, 12, 0, 0, 0, loc)
	to := time.Date(2024, 3, 11, 12, 0, 0, 0, loc)
	if got := daysBetween(from, to); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}
