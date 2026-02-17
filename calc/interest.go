package calc

import (
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// CalculateRotativeInterest computes rotative credit interest.
// The daily rate is derived from MonthlyRate / 30.
//
// Input validation (non-negative principal, valid days) is the caller's responsibility.
func CalculateRotativeInterest(principal domain.Money, days int, cfg config.InterestConfig) domain.Money {
	denom := int64(30) * domain.RateDenominator
	num := int64(principal) * int64(cfg.MonthlyRate) * int64(days)
	return domain.Money((num + denom/2) / denom)
}
