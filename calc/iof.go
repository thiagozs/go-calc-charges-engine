package calc

import (
	"math"

	"github.com/thiagozs/go-credit-card-engine/config"
	"github.com/thiagozs/go-credit-card-engine/domain"
)

func CalculateIOF(principal domain.Money, days int, cfg config.IOFConfig) domain.Money {
	daily := float64(principal) * cfg.DailyRate * float64(days)
	additional := float64(principal) * cfg.AdditionalRate
	total := daily + additional

	max := float64(principal) * cfg.MaxAnnualRate
	if total > max {
		total = max
	}

	return domain.Money(math.Round(total))
}
