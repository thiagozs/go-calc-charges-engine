package calc

import (
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// CalculateInternationalIOF calcula o IOF de compras internacionais por transacao.
//
// Input validation (non-negative amount) is the caller's responsibility.
func CalculateInternationalIOF(amount domain.Money, cfg config.InternationalIOFConfig) domain.Money {
	return mulRate(amount, cfg.Rate)
}

// CalculateInternationalIOFChecked is CalculateInternationalIOF with input validation.
// Returns ErrInvalidAmount, ErrInvalidRate or ErrOverflow.
func CalculateInternationalIOFChecked(amount domain.Money, cfg config.InternationalIOFConfig) (domain.Money, error) {
	if amount < 0 {
		return 0, ErrInvalidAmount
	}
	if cfg.Rate < 0 {
		return 0, ErrInvalidRate
	}
	if !mulFits(int64(amount), int64(cfg.Rate)) {
		return 0, ErrOverflow
	}
	return CalculateInternationalIOF(amount, cfg), nil
}
