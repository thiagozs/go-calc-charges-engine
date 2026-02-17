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
