package calc

import (
	"testing"

	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func TestCalculateIOF(t *testing.T) {
	principal := domain.Money(100_000) // R$ 1.000,00
	cfg := defaultIOFConfig()

	tests := []struct {
		name     string
		days     int
		expected domain.Money
	}{
		{
			name:     "30 days within cap",
			days:     30,
			expected: domain.Money(626), // R$ 6,26
		},
		{
			name:     "365 days within cap",
			days:     365,
			expected: domain.Money(3_373), // R$ 33,73
		},
		{
			name:     "days beyond 365 are not charged",
			days:     1_000,
			expected: domain.Money(3_373), // same as 365 days: 82*365 + 3800 = 33_730 / 1e6
		},
		{
			name:     "zero days only pays the additional rate",
			days:     0,
			expected: domain.Money(380),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iof := CalculateIOF(principal, tt.days, cfg)
			if iof != tt.expected {
				t.Fatalf("expected %d got %d", tt.expected, iof)
			}
		})
	}
}

func TestCalculateIOF_MaxDaysZeroDisablesLimit(t *testing.T) {
	cfg := defaultIOFConfig()
	cfg.MaxDays = 0
	// 100_000 * 82 * 1000 / 1e6 = 8_200, + 380
	if got := CalculateIOF(100_000, 1_000, cfg); got != 8_580 {
		t.Fatalf("expected 8580, got %d", got)
	}
}
