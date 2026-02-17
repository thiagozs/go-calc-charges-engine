package calc

import (
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// CalculateLateInterest computes late payment interest (juros de mora).
// The daily rate is derived from MonthlyRate / 30.
//
// Input validation (non-negative principal, valid days) is the caller's responsibility.
func CalculateLateInterest(principal domain.Money, days int, cfg config.LateInterestConfig) domain.Money {
	if cfg.MonthlyRate <= 0 || days <= 0 {
		return 0
	}
	denom := int64(30) * domain.RateDenominator
	num := int64(principal) * int64(cfg.MonthlyRate) * int64(days)
	return domain.Money((num + denom/2) / denom)
}
