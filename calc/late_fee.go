package calc

import (
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// CalculateLateFee computes the late payment fee (multa).
//
// Input validation (non-negative principal) is the caller's responsibility.
func CalculateLateFee(principal domain.Money, cfg config.LateFeeConfig) domain.Money {
	if cfg.Rate <= 0 {
		return 0
	}
	return mulRate(principal, cfg.Rate)
}
