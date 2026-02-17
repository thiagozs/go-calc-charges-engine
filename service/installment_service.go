package service

import (
	"time"

	"github.com/thiagozs/go-calc-charges-engine/calc"
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

type InstallmentService struct {
	IOFConfig         config.IOFConfig
	InstallmentConfig config.InstallmentConfig
}

func (s *InstallmentService) Calculate(
	amount domain.Money,
	numInstallments int,
	purchaseDate time.Time,
	firstDueDate time.Time,
) domain.InstallmentPlan {
	return calc.CalculateInstallmentPlan(
		amount, numInstallments,
		purchaseDate, firstDueDate,
		s.IOFConfig, s.InstallmentConfig,
	)
}

func NewInstallmentService(cfg config.EngineConfig) *InstallmentService {
	return &InstallmentService{
		IOFConfig:         cfg.IOF,
		InstallmentConfig: cfg.Installment,
	}
}
