package config

import (
	"fmt"
	"reflect"

	"github.com/caarlos0/env/v11"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

type EngineConfig struct {
	IOF              IOFConfig
	Interest         InterestConfig
	LateFee          LateFeeConfig
	LateInterest     LateInterestConfig
	Rules            RotativeRulesConfig
	InternationalIOF InternationalIOFConfig
	Installment      InstallmentConfig
}

func LoadFromEnv() (EngineConfig, error) {
	var cfg EngineConfig
	opts := env.Options{FuncMap: map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(domain.Rate(0)): func(v string) (any, error) { return domain.ParseRate(v) },
	}}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return EngineConfig{}, err
	}
	if err := cfg.Validate(); err != nil {
		return EngineConfig{}, err
	}
	return cfg, nil
}

// Validate rejects negative rates and a negative MaxDays.
func (c EngineConfig) Validate() error {
	checks := []struct {
		name  string
		value int64
	}{
		{"IOF_DAILY_RATE", int64(c.IOF.DailyRate)},
		{"IOF_ADDITIONAL_RATE", int64(c.IOF.AdditionalRate)},
		{"IOF_MAX_DAYS", int64(c.IOF.MaxDays)},
		{"ROTATIVE_MONTHLY_RATE", int64(c.Interest.MonthlyRate)},
		{"LATE_FEE_RATE", int64(c.LateFee.Rate)},
		{"LATE_INTEREST_MONTHLY_RATE", int64(c.LateInterest.MonthlyRate)},
		{"ROTATIVE_MAX_DAYS", int64(c.Rules.MaxDays)},
		{"ROTATIVE_MAX_CHARGE_RATE", int64(c.Rules.MaxChargeRate)},
		{"INTERNATIONAL_IOF_RATE", int64(c.InternationalIOF.Rate)},
		{"INSTALLMENT_MONTHLY_RATE", int64(c.Installment.MonthlyRate)},
	}
	for _, ch := range checks {
		if ch.value < 0 {
			return fmt.Errorf("config: %s must not be negative (got %d)", ch.name, ch.value)
		}
	}
	return nil
}
