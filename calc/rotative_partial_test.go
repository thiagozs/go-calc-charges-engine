package calc

import (
	"testing"

	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// Cenário real:
// Fatura vence dia 01/01
// Cliente paga parte dia 10/01
// Cliente quita restante dia 31/01
func TestRotativeWithPartialPayments(t *testing.T) {
	// Saldo inicial em rotativo
	balance := domain.RotativeBalance{
		Principal: 100_000, // R$ 1.000,00
		StartDate: utcDate(2024, 1, 1),
	}

	// Primeiro período: 9 dias (01 -> 10)
	result1 := CalculateRotative(
		balance,
		utcDate(2024, 1, 10),
		defaultIOFConfig(),
		defaultInterestConfig(),
		defaultLateFeeConfig(),
		defaultLateInterestConfig(),
		defaultRotativeRulesConfig(),
	)

	// Cliente paga R$ 400,00
	payment := domain.Money(40_000)
	applied := ApplyPayment(
		result1.Total,
		result1.IOF,
		result1.Interest,
		result1.LateInterest,
		result1.LateFee,
		result1.Principal,
		payment,
	)
	remainingPrincipal := result1.Principal - applied.PaidPrincipal

	if remainingPrincipal <= 0 {
		t.Fatalf("remaining principal should be positive")
	}

	// Segundo período: 21 dias (10 -> 31)
	balance2 := domain.RotativeBalance{
		Principal: remainingPrincipal,
		StartDate: utcDate(2024, 1, 10),
	}

	result2 := CalculateRotative(
		balance2,
		utcDate(2024, 1, 31),
		defaultIOFConfig(),
		defaultInterestConfig(),
		defaultLateFeeConfig(),
		defaultLateInterestConfig(),
		defaultRotativeRulesConfig(),
	)

	if result2.Total <= balance2.Principal {
		t.Fatalf("expected interest + IOF after partial payment")
	}
	if applied.PaidPrincipal <= 0 {
		t.Fatalf("expected payment to amortize principal")
	}
}
