package calc

import "github.com/thiagozs/go-credit-card-engine/domain"

// AmortizationResult detalha como o pagamento foi aplicado
type AmortizationResult struct {
	PaidIOF          domain.Money
	PaidInterest     domain.Money
	PaidLateInterest domain.Money
	PaidLateFee      domain.Money
	PaidPrincipal    domain.Money
	Remaining        domain.Money
}

// ApplyPayment aplica a regra bancária:
// IOF -> Juros -> Juros de Mora -> Multa -> Principal
func ApplyPayment(
	total domain.Money,
	iof domain.Money,
	interest domain.Money,
	lateInterest domain.Money,
	lateFee domain.Money,
	principal domain.Money,
	payment domain.Money,
) AmortizationResult {

	remaining := total
	p := payment

	result := AmortizationResult{}

	// 1) IOF
	if p > 0 {
		paid := min(p, iof)
		result.PaidIOF = paid
		p -= paid
		remaining -= paid
	}

	// 2) Juros
	if p > 0 {
		paid := min(p, interest)
		result.PaidInterest = paid
		p -= paid
		remaining -= paid
	}

	// 3) Juros de Mora
	if p > 0 {
		paid := min(p, lateInterest)
		result.PaidLateInterest = paid
		p -= paid
		remaining -= paid
	}

	// 4) Multa
	if p > 0 {
		paid := min(p, lateFee)
		result.PaidLateFee = paid
		p -= paid
		remaining -= paid
	}

	// 5) Principal
	if p > 0 {
		paid := min(p, principal)
		result.PaidPrincipal = paid
		remaining -= paid
	}

	result.Remaining = remaining
	return result
}

func min(a, b domain.Money) domain.Money {
	if a < b {
		return a
	}
	return b
}
