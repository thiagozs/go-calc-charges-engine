package calc

import (
	"math"

	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// Debt is the outstanding amount broken down by component.
// Total is always derived from the components, so it cannot diverge from them.
type Debt struct {
	IOF          domain.Money
	Interest     domain.Money
	LateInterest domain.Money
	LateFee      domain.Money
	Principal    domain.Money
}

// Total returns the sum of all components.
func (d Debt) Total() domain.Money {
	return d.IOF + d.Interest + d.LateInterest + d.LateFee + d.Principal
}

// Debt returns the amount owed in the result, ready to be passed to ApplyPayment.
func (r RotativeResult) Debt() Debt {
	return Debt{
		IOF:          r.IOF,
		Interest:     r.Interest,
		LateInterest: r.LateInterest,
		LateFee:      r.LateFee,
		Principal:    r.Principal,
	}
}

// AmortizationResult detalha como o pagamento foi aplicado seguindo a regra bancaria.
// The caller (ledger) should persist this struct for audit trail purposes.
//
// Invariants: Paid* + Remaining == debt.Total() and
// Paid* + Excess == payment (Excess is the part of the payment that exceeded the debt).
type AmortizationResult struct {
	PaidIOF          domain.Money
	PaidInterest     domain.Money
	PaidLateInterest domain.Money
	PaidLateFee      domain.Money
	PaidPrincipal    domain.Money
	Remaining        domain.Money
	Excess           domain.Money
}

// TotalPaid returns the amount of the payment that was applied to the debt.
func (r AmortizationResult) TotalPaid() domain.Money {
	return r.PaidIOF + r.PaidInterest + r.PaidLateInterest + r.PaidLateFee + r.PaidPrincipal
}

// ApplyPayment aplica a regra bancária:
// IOF -> Juros -> Juros de Mora -> Multa -> Principal
//
// Any part of the payment above the total debt is returned in Excess (the caller
// decides whether to refund or credit it). Returns ErrNegativeAmount if the payment
// or any debt component is negative, and ErrOverflow if the debt total does not fit in int64.
func ApplyPayment(debt Debt, payment domain.Money) (AmortizationResult, error) {
	components := [...]domain.Money{debt.IOF, debt.Interest, debt.LateInterest, debt.LateFee, debt.Principal}
	if payment < 0 {
		return AmortizationResult{}, ErrNegativeAmount
	}
	var total int64
	for _, c := range components {
		if c < 0 {
			return AmortizationResult{}, ErrNegativeAmount
		}
		if int64(c) > math.MaxInt64-total {
			return AmortizationResult{}, ErrOverflow
		}
		total += int64(c)
	}

	left := payment
	apply := func(owed domain.Money) domain.Money {
		paid := min(left, owed)
		left -= paid
		return paid
	}

	result := AmortizationResult{
		PaidIOF:          apply(debt.IOF),
		PaidInterest:     apply(debt.Interest),
		PaidLateInterest: apply(debt.LateInterest),
		PaidLateFee:      apply(debt.LateFee),
		PaidPrincipal:    apply(debt.Principal),
		Excess:           left,
	}
	result.Remaining = domain.Money(total) - result.TotalPaid()
	return result, nil
}
