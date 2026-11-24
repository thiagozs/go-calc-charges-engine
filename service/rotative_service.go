package service

import (
	"time"

	"github.com/thiagozs/go-calc-charges-engine/calc"
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

type RotativeService struct {
	iofConfig          config.IOFConfig
	interestConfig     config.InterestConfig
	lateFeeConfig      config.LateFeeConfig
	lateInterestConfig config.LateInterestConfig
	rulesConfig        config.RotativeRulesConfig
}

func (s *RotativeService) Calculate(balance domain.RotativeBalance,
	at time.Time) (calc.RotativeResult, error) {
	return calc.CalculateRotative(
		balance,
		at,
		s.iofConfig,
		s.interestConfig,
		s.lateFeeConfig,
		s.lateInterestConfig,
		s.rulesConfig,
	)
}

func NewRotativeService(cfg config.EngineConfig) *RotativeService {
	return &RotativeService{
		iofConfig:          cfg.IOF,
		interestConfig:     cfg.Interest,
		lateFeeConfig:      cfg.LateFee,
		lateInterestConfig: cfg.LateInterest,
		rulesConfig:        cfg.Rules,
	}
}

func NewRotativeServiceFromEnv() (*RotativeService, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}
	return NewRotativeService(cfg), nil
}
