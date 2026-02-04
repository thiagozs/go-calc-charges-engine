package calc

import (
	"math"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func CalculateLateFee(principal domain.Money, cfg config.LateFeeConfig) domain.Money {
	if cfg.Rate <= 0 {
		return 0
	}
	fee := float64(principal) * cfg.Rate
	return domain.Money(math.Round(fee))
}
