package service

import (
	"time"

	"github.com/thiagozs/go-calc-charges-engine/calc"
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

type InstallmentService struct {
	iofConfig         config.IOFConfig
	installmentConfig config.InstallmentConfig
}

func (s *InstallmentService) Calculate(
	amount domain.Money,
	numInstallments int,
	purchaseDate time.Time,
	firstDueDate time.Time,
) (domain.InstallmentPlan, error) {
	return calc.CalculateInstallmentPlan(
		amount, numInstallments,
		purchaseDate, firstDueDate,
		s.iofConfig, s.installmentConfig,
	)
}

func NewInstallmentService(cfg config.EngineConfig) *InstallmentService {
	return &InstallmentService{
		iofConfig:         cfg.IOF,
		installmentConfig: cfg.Installment,
	}
}

func NewInstallmentServiceFromEnv() (*InstallmentService, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}
	return NewInstallmentService(cfg), nil
}
