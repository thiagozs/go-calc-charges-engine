package service

import (
	"github.com/thiagozs/go-calc-charges-engine/calc"
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

type InternationalIOFService struct {
	cfg config.InternationalIOFConfig
}

// Calculate returns the IOF for an international transaction.
// Returns calc.ErrInvalidAmount for negative amounts and calc.ErrOverflow when
// the amount is too large to compute exactly.
func (s *InternationalIOFService) Calculate(amount domain.Money) (domain.Money, error) {
	return calc.CalculateInternationalIOFChecked(amount, s.cfg)
}

func NewInternationalIOFService(cfg config.EngineConfig) *InternationalIOFService {
	return &InternationalIOFService{cfg: cfg.InternationalIOF}
}
