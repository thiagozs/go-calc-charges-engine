package calc

import (
	"math"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func CalculateRotativeInterest(principal domain.Money, days int, cfg config.InterestConfig) domain.Money {
	dailyRate := cfg.MonthlyRate / 30.0
	interest := float64(principal) * dailyRate * float64(days)
	return domain.Money(math.Round(interest))
}
