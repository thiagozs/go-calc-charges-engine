package calc

import (
	"time"

	"github.com/thiagozs/go-calc-charges-engine/config"
)

func defaultIOFConfig() config.IOFConfig {
	return config.IOFConfig{
		DailyRate:      0.000082,
		AdditionalRate: 0.0038,
		MaxAnnualRate:  0.0408,
	}
}

func defaultInterestConfig() config.InterestConfig {
	return config.InterestConfig{
		MonthlyRate: 0.12,
	}
}

func defaultLateFeeConfig() config.LateFeeConfig {
	return config.LateFeeConfig{
		Rate: 0.02,
	}
}

func defaultLateInterestConfig() config.LateInterestConfig {
	return config.LateInterestConfig{
		MonthlyRate: 0.01,
	}
}

func defaultRotativeRulesConfig() config.RotativeRulesConfig {
	return config.RotativeRulesConfig{
		MaxDays:       30,
		MaxChargeRate: 1.0,
	}
}

func utcDate(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
