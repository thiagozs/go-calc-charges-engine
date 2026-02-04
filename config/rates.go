package config

type IOFConfig struct {
	DailyRate      float64
	AdditionalRate float64
	MaxAnnualRate  float64
}

type InterestConfig struct {
	MonthlyRate float64
}

type LateFeeConfig struct {
	Rate float64
}

type LateInterestConfig struct {
	MonthlyRate float64
}

type RotativeRulesConfig struct {
	MaxDays       int
	MaxChargeRate float64
}

type InternationalIOFConfig struct {
	Rate float64
}
