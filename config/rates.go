package config

import "github.com/thiagozs/go-calc-charges-engine/domain"

type IOFConfig struct {
	DailyRate      domain.Rate `env:"IOF_DAILY_RATE" envDefault:"82"`
	AdditionalRate domain.Rate `env:"IOF_ADDITIONAL_RATE" envDefault:"3800"`
	// MaxDays limits the days used by the daily IOF (Decreto 6.306/2007, art. 7, par. 1: 365 days).
	// 0 disables the limit.
	MaxDays int `env:"IOF_MAX_DAYS" envDefault:"365"`
}

type InterestConfig struct {
	MonthlyRate domain.Rate `env:"ROTATIVE_MONTHLY_RATE" envDefault:"120000"`
}

type LateFeeConfig struct {
	Rate domain.Rate `env:"LATE_FEE_RATE" envDefault:"20000"`
}

type LateInterestConfig struct {
	MonthlyRate domain.Rate `env:"LATE_INTEREST_MONTHLY_RATE" envDefault:"10000"`
}

type RotativeRulesConfig struct {
	MaxDays       int         `env:"ROTATIVE_MAX_DAYS" envDefault:"30"`
	MaxChargeRate domain.Rate `env:"ROTATIVE_MAX_CHARGE_RATE" envDefault:"1000000"`
}

type InternationalIOFConfig struct {
	Rate domain.Rate `env:"INTERNATIONAL_IOF_RATE" envDefault:"35000"`
}

type InstallmentConfig struct {
	MonthlyRate domain.Rate `env:"INSTALLMENT_MONTHLY_RATE" envDefault:"0"`
}
