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
