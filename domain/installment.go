package domain

import "time"

// InstallmentPlan represents a complete installment plan for a credit card purchase.
// The caller (ledger) should persist this struct for audit trail purposes.
type InstallmentPlan struct {
	TotalAmount   Money
	TotalIOF      Money
	TotalInterest Money
	TotalWithIOF  Money
	Installments  []Installment
}

// Installment represents a single installment in a plan.
type Installment struct {
	Number    int
	DueDate   time.Time
	Principal Money
	Interest  Money
	IOF       Money
	Amount    Money
}
