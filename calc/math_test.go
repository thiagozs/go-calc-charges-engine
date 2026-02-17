package calc

import (
	"testing"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/domain"
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

func TestFixedPow(t *testing.T) {
	// (1 + 0.0199)^1 = 1.0199 => fixedPow(19900, 1) should be ~1_019_900
	got := fixedPow(19_900, 1)
	if got != 1_019_900 {
		t.Fatalf("expected 1019900, got %d", got)
	}

	// (1 + 0.0199)^2 ≈ 1.040196 => ~1_040_196
	got2 := fixedPow(19_900, 2)
	// Allow small rounding: should be 1_040_196 or 1_040_197
	if got2 < 1_040_195 || got2 > 1_040_197 {
		t.Fatalf("expected ~1040196, got %d", got2)
	}
}

func TestFixedPow_Zero(t *testing.T) {
	got := fixedPow(120_000, 0)
	if got != domain.RateDenominator {
		t.Fatalf("expected %d, got %d", domain.RateDenominator, got)
	}
}
