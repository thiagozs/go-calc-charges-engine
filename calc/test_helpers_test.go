package calc

import (
	"testing"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func defaultIOFConfig() config.IOFConfig {
	return config.IOFConfig{
		DailyRate:      82,
		AdditionalRate: 3_800,
		MaxDays:        365,
	}
}

func defaultInterestConfig() config.InterestConfig {
	return config.InterestConfig{
		MonthlyRate: 120_000,
	}
}

func defaultLateFeeConfig() config.LateFeeConfig {
	return config.LateFeeConfig{
		Rate: 20_000,
	}
}

func defaultLateInterestConfig() config.LateInterestConfig {
	return config.LateInterestConfig{
		MonthlyRate: 10_000,
	}
}

func defaultRotativeRulesConfig() config.RotativeRulesConfig {
	return config.RotativeRulesConfig{
		MaxDays:       30,
		MaxChargeRate: 1_000_000,
	}
}

func utcDate(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func mustRotative(
	t *testing.T,
	balance domain.RotativeBalance,
	calcDate time.Time,
	iofCfg config.IOFConfig,
	intCfg config.InterestConfig,
	lateFeeCfg config.LateFeeConfig,
	lateInterestCfg config.LateInterestConfig,
	rulesCfg config.RotativeRulesConfig,
) RotativeResult {
	t.Helper()
	r, err := CalculateRotative(balance, calcDate, iofCfg, intCfg, lateFeeCfg, lateInterestCfg, rulesCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return r
}
