package calc

import (
	"testing"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func TestCalculateRotative_FullScenario(t *testing.T) {
	balance := domain.RotativeBalance{
		Principal: 100_000,
		StartDate: utcDate(2024, 1, 1),
	}

	result := CalculateRotative(
		balance,
		utcDate(2024, 1, 31),
		defaultIOFConfig(),
		defaultInterestConfig(),
		defaultLateFeeConfig(),
		defaultLateInterestConfig(),
		defaultRotativeRulesConfig(),
	)

	if result.Interest != 12_000 {
		t.Fatalf("expected interest 12000 got %d", result.Interest)
	}
	if result.IOF != 626 {
		t.Fatalf("expected IOF 626 got %d", result.IOF)
	}
	if result.LateFee != 2_000 {
		t.Fatalf("expected late fee 2000 got %d", result.LateFee)
	}
	if result.LateInterest != 1_000 {
		t.Fatalf("expected late interest 1000 got %d", result.LateInterest)
	}
	if result.Total != balance.Principal+result.Interest+result.IOF+result.LateFee+result.LateInterest {
		t.Fatalf("total should match principal + charges")
	}
}

func TestCalculateRotative_RespectsMaxDays(t *testing.T) {
	balance := domain.RotativeBalance{
		Principal: 100_000,
		StartDate: utcDate(2024, 1, 1),
	}

	result := CalculateRotative(
		balance,
		utcDate(2024, 3, 15), // 74 dias
		defaultIOFConfig(),
		defaultInterestConfig(),
		defaultLateFeeConfig(),
		defaultLateInterestConfig(),
		defaultRotativeRulesConfig(),
	)

	if result.ChargedDays != 30 {
		t.Fatalf("expected charged days 30 got %d", result.ChargedDays)
	}
	if result.Interest != 12_000 {
		t.Fatalf("expected interest 12000 got %d", result.Interest)
	}
}

func TestCalculateRotative_ChargeCap(t *testing.T) {
	balance := domain.RotativeBalance{
		Principal: 100_000,
		StartDate: utcDate(2024, 1, 1),
	}

	highInterest := config.InterestConfig{MonthlyRate: 1_000_000}
	result := CalculateRotative(
		balance,
		utcDate(2024, 1, 31),
		defaultIOFConfig(),
		highInterest,
		defaultLateFeeConfig(),
		defaultLateInterestConfig(),
		defaultRotativeRulesConfig(),
	)

	if result.Charges != balance.Principal {
		t.Fatalf("expected charges capped at %d got %d", balance.Principal, result.Charges)
	}
	if !result.ChargeCapped {
		t.Fatalf("expected charge cap to be applied")
	}
}
