package config

import "github.com/caarlos0/env/v11"

type EngineConfig struct {
	IOF              IOFConfig
	Interest         InterestConfig
	LateFee          LateFeeConfig
	LateInterest     LateInterestConfig
	Rules            RotativeRulesConfig
	InternationalIOF InternationalIOFConfig
}

func LoadFromEnv() (EngineConfig, error) {
	var cfg EngineConfig
	if err := env.Parse(&cfg); err != nil {
		return EngineConfig{}, err
	}
	return cfg, nil
}
