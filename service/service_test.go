package service

import (
	"errors"
	"testing"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/calc"
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func defaultCfg(t *testing.T) config.EngineConfig {
	t.Helper()
	return config.EngineConfig{
		IOF:              config.IOFConfig{DailyRate: 82, AdditionalRate: 3_800, MaxDays: 365},
		Interest:         config.InterestConfig{MonthlyRate: 120_000},
		LateFee:          config.LateFeeConfig{Rate: 20_000},
		LateInterest:     config.LateInterestConfig{MonthlyRate: 10_000},
		Rules:            config.RotativeRulesConfig{MaxDays: 30, MaxChargeRate: 1_000_000},
		InternationalIOF: config.InternationalIOFConfig{Rate: 35_000},
	}
}

func TestRotativeService_PropagatesErrors(t *testing.T) {
	svc := NewRotativeService(defaultCfg(t))
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := svc.Calculate(domain.RotativeBalance{Principal: -1, StartDate: start}, start); !errors.Is(err, calc.ErrInvalidAmount) {
		t.Fatalf("got %v", err)
	}
	res, err := svc.Calculate(domain.RotativeBalance{Principal: 100_000, StartDate: start}, start.AddDate(0, 0, 15))
	if err != nil || res.Total <= 100_000 {
		t.Fatalf("unexpected: %+v %v", res, err)
	}
}

func TestInstallmentService_PropagatesErrors(t *testing.T) {
	svc := NewInstallmentService(defaultCfg(t))
	d := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	if _, err := svc.Calculate(100_000, 0, d.AddDate(0, -1, 0), d); !errors.Is(err, calc.ErrInvalidInstallments) {
		t.Fatalf("got %v", err)
	}
}

func TestInternationalIOFService(t *testing.T) {
	svc := NewInternationalIOFService(defaultCfg(t))
	got, err := svc.Calculate(100_000)
	if err != nil || got != 3_500 { // 3.5%
		t.Fatalf("got %d, %v", got, err)
	}
	if _, err := svc.Calculate(-1); !errors.Is(err, calc.ErrInvalidAmount) {
		t.Fatalf("got %v", err)
	}
}
