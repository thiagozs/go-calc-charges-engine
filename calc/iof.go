package calc

import (
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// CalculateIOF computes IOF (Imposto sobre Operacoes Financeiras) for the given
// principal over the specified number of days.
//
// Input validation (non-negative principal, valid days) is the caller's responsibility.
func CalculateIOF(principal domain.Money, days int, cfg config.IOFConfig) domain.Money {
	daily := mulRateDays(principal, cfg.DailyRate, days)
	additional := mulRate(principal, cfg.AdditionalRate)
	total := daily + additional

	maxVal := mulRate(principal, cfg.MaxAnnualRate)
	if total > maxVal {
		total = maxVal
	}

	return total
}
