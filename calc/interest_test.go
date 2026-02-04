package calc

import (
	"testing"

	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func TestCalculateRotativeInterest(t *testing.T) {
	principal := domain.Money(100_000)
	cfg := defaultInterestConfig()

	tests := []struct {
		name     string
		days     int
		expected domain.Money
	}{
		{
			name:     "0 days",
			days:     0,
			expected: 0,
		},
		{
			name:     "15 days",
			days:     15,
			expected: domain.Money(6_000),
		},
		{
			name:     "30 days",
			days:     30,
			expected: domain.Money(12_000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interest := CalculateRotativeInterest(principal, tt.days, cfg)
			if interest != tt.expected {
				t.Fatalf("expected %d got %d", tt.expected, interest)
			}
		})
	}
}
