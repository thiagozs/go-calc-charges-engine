package service

import (
	"time"

	"github.com/thiagozs/go-credit-card-engine/calc"
	"github.com/thiagozs/go-credit-card-engine/config"
	"github.com/thiagozs/go-credit-card-engine/domain"
)

type RotativeService struct {
	IOFConfig          config.IOFConfig
	InterestConfig     config.InterestConfig
	LateFeeConfig      config.LateFeeConfig
	LateInterestConfig config.LateInterestConfig
	RulesConfig        config.RotativeRulesConfig
}

func (s *RotativeService) Calculate(balance domain.RotativeBalance,
	at time.Time) calc.RotativeResult {
	return calc.CalculateRotative(
		balance,
		at,
		s.IOFConfig,
		s.InterestConfig,
		s.LateFeeConfig,
		s.LateInterestConfig,
		s.RulesConfig,
	)
}

func NewRotativeService(cfg config.EngineConfig) *RotativeService {
	return &RotativeService{
		IOFConfig:          cfg.IOF,
		InterestConfig:     cfg.Interest,
		LateFeeConfig:      cfg.LateFee,
		LateInterestConfig: cfg.LateInterest,
		RulesConfig:        cfg.Rules,
	}
}

func NewRotativeServiceFromEnv() (*RotativeService, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}
	return NewRotativeService(cfg), nil
}
