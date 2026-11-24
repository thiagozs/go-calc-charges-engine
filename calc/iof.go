package calc

import (
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// CalculateIOF computes IOF (Imposto sobre Operacoes Financeiras) for the given
// principal over the specified number of days:
//
//	IOF = principal * DailyRate * min(days, MaxDays) + principal * AdditionalRate
//
// Per Decreto 6.306/2007 art. 7, the daily rate applies for at most 365 days
// (par. 1) and the additional rate (0.38%) is charged once, regardless of term (par. 15).
// MaxDays <= 0 disables the day limit.
//
// Input validation (non-negative principal, valid days) is the caller's responsibility.
func CalculateIOF(principal domain.Money, days int, cfg config.IOFConfig) domain.Money {
	if cfg.MaxDays > 0 && days > cfg.MaxDays {
		days = cfg.MaxDays
	}
	return mulRateDays(principal, cfg.DailyRate, days) + mulRate(principal, cfg.AdditionalRate)
}
