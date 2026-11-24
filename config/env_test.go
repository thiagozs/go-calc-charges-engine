package config

import "testing"

var envDefaults = map[string]string{
	"IOF_DAILY_RATE":             "82",
	"IOF_ADDITIONAL_RATE":        "3800",
	"IOF_MAX_DAYS":               "365",
	"ROTATIVE_MONTHLY_RATE":      "120000",
	"LATE_FEE_RATE":              "20000",
	"LATE_INTEREST_MONTHLY_RATE": "10000",
	"ROTATIVE_MAX_DAYS":          "30",
	"ROTATIVE_MAX_CHARGE_RATE":   "1000000",
	"INTERNATIONAL_IOF_RATE":     "35000",
	"INSTALLMENT_MONTHLY_RATE":   "0",
}

// setDefaultEnv pins every variable so tests do not depend on the caller's shell.
func setDefaultEnv(t *testing.T) {
	t.Helper()
	for k, v := range envDefaults {
		t.Setenv(k, v)
	}
}

func TestLoadFromEnv_Defaults(t *testing.T) {
	setDefaultEnv(t)
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Interest.MonthlyRate != 120_000 || cfg.Rules.MaxDays != 30 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadFromEnv_RejectsNegativeRate(t *testing.T) {
	setDefaultEnv(t)
	t.Setenv("LATE_FEE_RATE", "-1")
	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("expected error for negative rate")
	}
}

func TestLoadFromEnv_RejectsNegativeMaxDays(t *testing.T) {
	setDefaultEnv(t)
	t.Setenv("ROTATIVE_MAX_DAYS", "-5")
	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("expected error for negative max days")
	}
}

func TestLoadFromEnv_AcceptsDecimalRates(t *testing.T) {
	setDefaultEnv(t)
	t.Setenv("ROTATIVE_MONTHLY_RATE", "0.12")
	t.Setenv("IOF_DAILY_RATE", "0.000082")
	t.Setenv("ROTATIVE_MAX_CHARGE_RATE", "1.0")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Interest.MonthlyRate != 120_000 || cfg.IOF.DailyRate != 82 || cfg.Rules.MaxChargeRate != 1_000_000 {
		t.Fatalf("unexpected: %+v", cfg)
	}
}

func TestLoadFromEnv_RejectsMalformedRate(t *testing.T) {
	setDefaultEnv(t)
	t.Setenv("LATE_FEE_RATE", "0.0000001")
	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("expected error for >6 decimal places")
	}
}
