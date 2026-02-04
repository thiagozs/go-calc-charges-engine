package calc

import (
	"testing"

	"github.com/thiagozs/go-credit-card-engine/domain"
)

func TestApplyPayment(t *testing.T) {
	total := domain.Money(123_000) // principal + encargos
	iof := domain.Money(6_000)
	interest := domain.Money(14_000)
	lateInterest := domain.Money(1_000)
	lateFee := domain.Money(2_000)
	principal := domain.Money(100_000)

	tests := []struct {
		name         string
		payment      domain.Money
		expectedIOF  domain.Money
		expectedInt  domain.Money
		expectedMora domain.Money
		expectedFee  domain.Money
		expectedPrin domain.Money
		expectedRem  domain.Money
	}{
		{
			name:         "payment covers IOF only",
			payment:      domain.Money(5_000),
			expectedIOF:  domain.Money(5_000),
			expectedInt:  0,
			expectedMora: 0,
			expectedFee:  0,
			expectedPrin: 0,
			expectedRem:  domain.Money(118_000),
		},
		{
			name:         "payment covers IOF + interest",
			payment:      domain.Money(20_000),
			expectedIOF:  domain.Money(6_000),
			expectedInt:  domain.Money(14_000),
			expectedMora: 0,
			expectedFee:  0,
			expectedPrin: 0,
			expectedRem:  domain.Money(103_000),
		},
		{
			name:         "payment reaches principal",
			payment:      domain.Money(40_000),
			expectedIOF:  domain.Money(6_000),
			expectedInt:  domain.Money(14_000),
			expectedMora: domain.Money(1_000),
			expectedFee:  domain.Money(2_000),
			expectedPrin: domain.Money(17_000),
			expectedRem:  domain.Money(83_000),
		},
		{
			name:         "payment exceeds total",
			payment:      domain.Money(200_000),
			expectedIOF:  domain.Money(6_000),
			expectedInt:  domain.Money(14_000),
			expectedMora: domain.Money(1_000),
			expectedFee:  domain.Money(2_000),
			expectedPrin: domain.Money(100_000),
			expectedRem:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ApplyPayment(total, iof, interest, lateInterest, lateFee, principal, tt.payment)
			if r.PaidIOF != tt.expectedIOF {
				t.Fatalf("expected IOF %d got %d", tt.expectedIOF, r.PaidIOF)
			}
			if r.PaidInterest != tt.expectedInt {
				t.Fatalf("expected interest %d got %d", tt.expectedInt, r.PaidInterest)
			}
			if r.PaidLateInterest != tt.expectedMora {
				t.Fatalf("expected late interest %d got %d", tt.expectedMora, r.PaidLateInterest)
			}
			if r.PaidLateFee != tt.expectedFee {
				t.Fatalf("expected late fee %d got %d", tt.expectedFee, r.PaidLateFee)
			}
			if r.PaidPrincipal != tt.expectedPrin {
				t.Fatalf("expected principal %d got %d", tt.expectedPrin, r.PaidPrincipal)
			}
			if r.Remaining != tt.expectedRem {
				t.Fatalf("expected remaining %d got %d", tt.expectedRem, r.Remaining)
			}
		})
	}
}

// Regra bancária:
// Pagamento abate na ordem:
// 1) IOF
// 2) Juros
// 3) Principal
func TestApplyPayment_WithRotativeResult(t *testing.T) {
	balance := domain.RotativeBalance{
		Principal: 100_000, // R$ 1.000,00
		StartDate: utcDate(2024, 1, 1),
	}

	result := CalculateRotative(
		balance,
		utcDate(2024, 1, 31),
		defaultIOFConfig(),
		defaultInterestConfig(),
		defaultLateFeeConfig(),
		defaultLateInterestConfig(),
		defaultRotativeRulesConfig(),
	)

	payment := domain.Money(20_000) // Cliente paga R$ 200,00
	applied := ApplyPayment(
		result.Total,
		result.IOF,
		result.Interest,
		result.LateInterest,
		result.LateFee,
		result.Principal,
		payment,
	)

	if applied.Remaining != result.Total-payment {
		t.Fatalf("expected remaining %d got %d", result.Total-payment, applied.Remaining)
	}
	if applied.Remaining < 0 {
		t.Fatalf("remaining balance cannot be negative")
	}
}
