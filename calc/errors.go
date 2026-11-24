package calc

import "errors"

// MaxInstallments is the upper bound accepted by CalculateInstallmentPlan.
const MaxInstallments = 360

var (
	// ErrInvalidInstallments is returned when the number of installments is outside [1, MaxInstallments].
	ErrInvalidInstallments = errors.New("calc: number of installments must be between 1 and 360")
	// ErrInvalidAmount is returned when the amount is not strictly positive.
	ErrInvalidAmount = errors.New("calc: amount must be positive")
	// ErrNegativeAmount is returned when a payment or debt component is negative.
	ErrNegativeAmount = errors.New("calc: amount must not be negative")
	// ErrInvalidRate is returned when a rate is negative.
	ErrInvalidRate = errors.New("calc: rate must not be negative")
	// ErrInvalidDate is returned for zero dates or a first due date before the purchase date.
	ErrInvalidDate = errors.New("calc: invalid dates")
	// ErrInfeasiblePlan is returned when the rounded payment does not amortize the balance
	// (for example a very long term with a tiny principal), which would yield invalid installments.
	ErrInfeasiblePlan = errors.New("calc: installment plan is infeasible for this amount, rate and term")
	// ErrOverflow is returned when an intermediate result does not fit in int64.
	ErrOverflow = errors.New("calc: arithmetic overflow")
)
