package calc

import (
	"testing"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func TestCalculateInternationalIOF(t *testing.T) {
	cfg := config.InternationalIOFConfig{Rate: 35_000}

	tests := []struct {
		name     string
		amount   domain.Money
		expected domain.Money
	}{
		{
			name:     "basic rate",
			amount:   domain.Money(100_000),
			expected: domain.Money(3_500),
		},
		{
			name:     "rounding",
			amount:   domain.Money(12_345),
			expected: domain.Money(432), // 12_345 * 3.5% = 432.075 -> 432
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateInternationalIOF(tt.amount, cfg)
			if got != tt.expected {
				t.Fatalf("expected %d got %d", tt.expected, got)
			}
		})
	}
}
