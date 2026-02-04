package calc

import (
	"math"

	"github.com/thiagozs/go-credit-card-engine/config"
	"github.com/thiagozs/go-credit-card-engine/domain"
)

// CalculateInternationalIOF calcula o IOF de compras internacionais por transação.
func CalculateInternationalIOF(amount domain.Money, cfg config.InternationalIOFConfig) domain.Money {
	iof := float64(amount) * cfg.Rate
	return domain.Money(math.Round(iof))
}
