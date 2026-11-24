package calc

import (
	"errors"
	"math"
	"testing"

	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func TestApplyPayment(t *testing.T) {
	debt := Debt{
		IOF:          6_000,
		Interest:     14_000,
		LateInterest: 1_000,
		LateFee:      2_000,
		Principal:    100_000,
	} // total 123_000

	tests := []struct {
		name         string
		payment      domain.Money
		expectedIOF  domain.Money
		expectedInt  domain.Money
		expectedMora domain.Money
		expectedFee  domain.Money
		expectedPrin domain.Money
		expectedRem  domain.Money
		expectedExc  domain.Money
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
			expectedExc:  77_000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ApplyPayment(debt, tt.payment)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
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
			if r.Excess != tt.expectedExc {
				t.Fatalf("expected excess %d got %d", tt.expectedExc, r.Excess)
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

	result := mustRotative(t,
		balance,
		utcDate(2024, 1, 31),
		defaultIOFConfig(),
		defaultInterestConfig(),
		defaultLateFeeConfig(),
		defaultLateInterestConfig(),
		defaultRotativeRulesConfig(),
	)

	payment := domain.Money(20_000) // Cliente paga R$ 200,00
	applied, err := ApplyPayment(result.Debt(), payment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applied.Remaining != result.Total-payment {
		t.Fatalf("expected remaining %d got %d", result.Total-payment, applied.Remaining)
	}
	if applied.Remaining < 0 {
		t.Fatalf("remaining balance cannot be negative")
	}
}

func TestApplyPayment_InvalidInput(t *testing.T) {
	if _, err := ApplyPayment(Debt{Principal: 100}, -1); !errors.Is(err, ErrNegativeAmount) {
		t.Fatalf("negative payment: got %v", err)
	}
	if _, err := ApplyPayment(Debt{IOF: -1, Principal: 100}, 10); !errors.Is(err, ErrNegativeAmount) {
		t.Fatalf("negative component: got %v", err)
	}
	if _, err := ApplyPayment(Debt{Interest: math.MaxInt64, Principal: 1}, 10); !errors.Is(err, ErrOverflow) {
		t.Fatalf("overflow: got %v", err)
	}
}

func TestApplyPayment_ZeroPaymentAndZeroDebt(t *testing.T) {
	r, err := ApplyPayment(Debt{IOF: 5, Principal: 10}, 0)
	if err != nil || r.Remaining != 15 || r.TotalPaid() != 0 || r.Excess != 0 {
		t.Fatalf("zero payment: %+v %v", r, err)
	}
	r, err = ApplyPayment(Debt{}, 50)
	if err != nil || r.Remaining != 0 || r.Excess != 50 {
		t.Fatalf("zero debt: %+v %v", r, err)
	}
}

// FuzzApplyPayment checks the conservation invariants for arbitrary inputs.
func FuzzApplyPayment(f *testing.F) {
	f.Add(int64(6_000), int64(14_000), int64(1_000), int64(2_000), int64(100_000), int64(40_000))
	f.Add(int64(0), int64(0), int64(0), int64(0), int64(0), int64(0))
	f.Add(int64(1), int64(2), int64(3), int64(4), int64(5), int64(1_000))
	f.Fuzz(func(t *testing.T, iof, interest, lateInterest, lateFee, principal, payment int64) {
		debt := Debt{domain.Money(iof), domain.Money(interest), domain.Money(lateInterest), domain.Money(lateFee), domain.Money(principal)}
		r, err := ApplyPayment(debt, domain.Money(payment))
		if err != nil {
			return // invalid input is rejected, nothing to check
		}
		total := debt.Total()
		if r.Remaining < 0 || r.Excess < 0 {
			t.Fatalf("negative remaining/excess: %+v", r)
		}
		if r.TotalPaid()+r.Remaining != total {
			t.Fatalf("paid+remaining != total: %+v total=%d", r, total)
		}
		if r.TotalPaid()+r.Excess != domain.Money(payment) {
			t.Fatalf("paid+excess != payment: %+v payment=%d", r, payment)
		}
		if r.Remaining > 0 && r.Excess > 0 {
			t.Fatalf("excess while debt remains: %+v", r)
		}
		// Priority order: a later component is only paid once earlier ones are fully paid.
		parts := []struct{ paid, owed domain.Money }{
			{r.PaidIOF, debt.IOF}, {r.PaidInterest, debt.Interest}, {r.PaidLateInterest, debt.LateInterest},
			{r.PaidLateFee, debt.LateFee}, {r.PaidPrincipal, debt.Principal},
		}
		for i := 0; i < len(parts)-1; i++ {
			if parts[i].paid < parts[i].owed {
				for _, later := range parts[i+1:] {
					if later.paid != 0 {
						t.Fatalf("priority violated: %+v", r)
					}
				}
			}
		}
	})
}
