package calc

import (
	"errors"
	"testing"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func TestCalculateRotative_FullScenario(t *testing.T) {
	balance := domain.RotativeBalance{
		Principal: 100_000,
		StartDate: utcDate(2024, 1, 1),
	}

	result := mustRotative(t,
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

	result := mustRotative(t,
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

func rotativeWith(t *testing.T, b domain.RotativeBalance, days int, in config.InterestConfig) RotativeResult {
	t.Helper()
	return mustRotative(t, b, b.StartDate.AddDate(0, 0, days), defaultIOFConfig(), in,
		defaultLateFeeConfig(), defaultLateInterestConfig(), defaultRotativeRulesConfig())
}

func TestCalculateRotative_ChargeCap(t *testing.T) {
	balance := domain.RotativeBalance{Principal: 100_000, StartDate: utcDate(2024, 1, 1)}

	// 100% a.m.: interest 100_000 + late fee 2_000 + late interest 1_000 = 103_000 > cap.
	result := rotativeWith(t, balance, 30, config.InterestConfig{MonthlyRate: 1_000_000})

	if !result.ChargeCapped {
		t.Fatalf("expected charge cap to be applied")
	}
	// Interest absorbs the cut, so late fee and late interest stay whole.
	if result.FinancialCharges != 100_000 || result.Interest != 97_000 || result.LateFee != 2_000 || result.LateInterest != 1_000 {
		t.Fatalf("unexpected capped breakdown: %+v", result)
	}
	// IOF is a tax: it sits outside the 100% cap.
	if result.IOF != 626 || result.Charges != 100_626 || result.Total != 200_626 {
		t.Fatalf("IOF must not count toward the cap: %+v", result)
	}
}

func TestCalculateRotative_CapIsCumulativeAndUsesOriginalDebt(t *testing.T) {
	// Original debt 100_000 -> cap 100_000. 97_000 already charged -> only 3_000 left,
	// even though the current principal (40_000) would generate 6_000 of charges.
	balance := domain.RotativeBalance{
		Principal:             40_000,
		StartDate:             utcDate(2024, 1, 1),
		OriginalDebt:          100_000,
		ChargesAlreadyApplied: 97_000,
	}
	r := rotativeWith(t, balance, 30, config.InterestConfig{MonthlyRate: 120_000})
	if !r.ChargeCapped || r.FinancialCharges != 3_000 {
		t.Fatalf("expected cumulative cap of 3000, got %+v", r)
	}
}

func TestCalculateRotative_CapExhaustedCutsEverything(t *testing.T) {
	balance := domain.RotativeBalance{Principal: 100_000, StartDate: utcDate(2024, 1, 1), ChargesAlreadyApplied: 150_000}
	r := rotativeWith(t, balance, 30, config.InterestConfig{MonthlyRate: 120_000})
	if r.FinancialCharges != 0 || r.Interest != 0 || r.LateFee != 0 || r.LateInterest != 0 {
		t.Fatalf("cap already exhausted, expected no financial charges: %+v", r)
	}
	if r.IOF == 0 {
		t.Fatalf("IOF is still due: %+v", r)
	}
}

func TestCalculateRotative_CapCutOrder(t *testing.T) {
	// Limit 2_500 against interest 12_000 + late interest 1_000 + late fee 2_000:
	// interest is cut to zero first, then late interest is trimmed to 500; the fee stays whole.
	balance := domain.RotativeBalance{Principal: 100_000, StartDate: utcDate(2024, 1, 1),
		OriginalDebt: 100_000, ChargesAlreadyApplied: 97_500}
	r := rotativeWith(t, balance, 30, config.InterestConfig{MonthlyRate: 120_000})
	if r.Interest != 0 || r.LateInterest != 500 || r.LateFee != 2_000 {
		t.Fatalf("unexpected cut order: %+v", r)
	}
}

func TestCalculateRotative_NegativeCumulativeInputs(t *testing.T) {
	for _, b := range []domain.RotativeBalance{
		{Principal: 1, StartDate: utcDate(2024, 1, 1), OriginalDebt: -1},
		{Principal: 1, StartDate: utcDate(2024, 1, 1), ChargesAlreadyApplied: -1},
	} {
		_, err := CalculateRotative(b, utcDate(2024, 1, 31), defaultIOFConfig(), defaultInterestConfig(),
			defaultLateFeeConfig(), defaultLateInterestConfig(), defaultRotativeRulesConfig())
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount for %+v, got %v", b, err)
		}
	}
}

func TestCalculateRotative_InvalidInput(t *testing.T) {
	start, end := utcDate(2024, 1, 1), utcDate(2024, 1, 31)
	run := func(b domain.RotativeBalance, at time.Time, iof config.IOFConfig, in config.InterestConfig) error {
		_, err := CalculateRotative(b, at, iof, in, defaultLateFeeConfig(), defaultLateInterestConfig(), defaultRotativeRulesConfig())
		return err
	}
	ok := domain.RotativeBalance{Principal: 100_000, StartDate: start}

	if err := run(domain.RotativeBalance{Principal: -1, StartDate: start}, end, defaultIOFConfig(), defaultInterestConfig()); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("negative principal: got %v", err)
	}
	if err := run(domain.RotativeBalance{Principal: 1}, end, defaultIOFConfig(), defaultInterestConfig()); !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("zero start date: got %v", err)
	}
	if err := run(ok, end, defaultIOFConfig(), config.InterestConfig{MonthlyRate: -1}); !errors.Is(err, ErrInvalidRate) {
		t.Fatalf("negative rate: got %v", err)
	}
	huge := domain.RotativeBalance{Principal: 1 << 60, StartDate: start}
	if err := run(huge, end, defaultIOFConfig(), defaultInterestConfig()); !errors.Is(err, ErrOverflow) {
		t.Fatalf("huge principal: got %v", err)
	}
}
