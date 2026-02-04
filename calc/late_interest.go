package calc

import (
	"math"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func CalculateLateInterest(principal domain.Money, days int, cfg config.LateInterestConfig) domain.Money {
	if cfg.MonthlyRate <= 0 || days <= 0 {
		return 0
	}
	dailyRate := cfg.MonthlyRate / 30.0
	interest := float64(principal) * dailyRate * float64(days)
	return domain.Money(math.Round(interest))
}
