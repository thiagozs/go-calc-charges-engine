package config

type IOFConfig struct {
	DailyRate      float64 `env:"IOF_DAILY_RATE" envDefault:"0.000082"`
	AdditionalRate float64 `env:"IOF_ADDITIONAL_RATE" envDefault:"0.0038"`
	MaxAnnualRate  float64 `env:"IOF_MAX_ANNUAL_RATE" envDefault:"0.0408"`
}

type InterestConfig struct {
	MonthlyRate float64 `env:"ROTATIVE_MONTHLY_RATE" envDefault:"0.12"`
}

type LateFeeConfig struct {
	Rate float64 `env:"LATE_FEE_RATE" envDefault:"0.02"`
}

type LateInterestConfig struct {
	MonthlyRate float64 `env:"LATE_INTEREST_MONTHLY_RATE" envDefault:"0.01"`
}

type RotativeRulesConfig struct {
	MaxDays       int     `env:"ROTATIVE_MAX_DAYS" envDefault:"30"`
	MaxChargeRate float64 `env:"ROTATIVE_MAX_CHARGE_RATE" envDefault:"1.0"`
}

type InternationalIOFConfig struct {
	Rate float64 `env:"INTERNATIONAL_IOF_RATE" envDefault:"0.035"`
}
