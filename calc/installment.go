package calc

import (
	"math/big"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// priceInstallment computes the Tabela Price payment (PMT) exactly with integer
// arithmetic, rounded half-up to the centavo:
//
//	PMT = PV * r * (1+r)^n / ((1+r)^n - 1),  r = rate / RateDenominator
//
// Writing (1+r)^n = N / D^n with N = (D+rate)^n, this becomes
//
//	PMT = PV * rate * N / (D * (N - D^n))
//
// computed with a single division at the end. (A fixed-point power rounding at every
// step loses precision that the small denominator (1+r)^n - 1 then amplifies.)
// Returns ErrOverflow if the result does not fit in int64.
func priceInstallment(pv domain.Money, rate domain.Rate, n int) (domain.Money, error) {
	d := big.NewInt(domain.RateDenominator)
	exp := big.NewInt(int64(n))
	pow := new(big.Int).Exp(big.NewInt(domain.RateDenominator+int64(rate)), exp, nil) // N
	base := new(big.Int).Exp(d, exp, nil)                                             // D^n

	den := new(big.Int).Sub(pow, base)
	den.Mul(den, d)
	if den.Sign() <= 0 {
		return 0, ErrInvalidRate
	}
	num := new(big.Int).Mul(big.NewInt(int64(pv)), big.NewInt(int64(rate)))
	num.Mul(num, pow)
	num.Add(num, new(big.Int).Rsh(den, 1)) // + den/2: round half up
	num.Quo(num, den)
	if !num.IsInt64() {
		return 0, ErrOverflow
	}
	return domain.Money(num.Int64()), nil
}

// validateInstallmentInput checks inputs and guarantees that every int64 product
// computed by the plan calculation fits, so the arithmetic below cannot overflow silently.
func validateInstallmentInput(
	totalAmount domain.Money,
	n int,
	purchaseDate, firstDueDate time.Time,
	iofCfg config.IOFConfig,
	instCfg config.InstallmentConfig,
) error {
	if n < 1 || n > MaxInstallments {
		return ErrInvalidInstallments
	}
	if totalAmount <= 0 {
		return ErrInvalidAmount
	}
	if purchaseDate.IsZero() || firstDueDate.IsZero() || daysBetween(firstDueDate, purchaseDate) > 0 {
		return ErrInvalidDate
	}
	if iofCfg.DailyRate < 0 || iofCfg.AdditionalRate < 0 || iofCfg.MaxDays < 0 || instCfg.MonthlyRate < 0 {
		return ErrInvalidRate
	}

	maxDays := int64(daysBetween(purchaseDate, addMonths(firstDueDate, n-1)))
	amount := int64(totalAmount)
	if !mulFits(amount, int64(iofCfg.DailyRate), maxDays) ||
		!mulFits(amount, int64(iofCfg.AdditionalRate)) ||
		!mulFits(amount, int64(instCfg.MonthlyRate)) ||
		!mulFits(amount, domain.RateDenominator) {
		return ErrOverflow
	}
	return nil
}

// CalculateInstallmentPlan generates an installment plan for a credit card purchase.
//
// Parameters:
//   - totalAmount: purchase amount in centavos
//   - numInstallments: number of installments (typically 1-12)
//   - purchaseDate: date of purchase
//   - firstDueDate: due date of the first installment
//   - iofCfg: IOF configuration for per-installment IOF calculation
//   - instCfg: installment interest rate config (MonthlyRate 0 = sem juros)
//
// Returns ErrInvalidInstallments, ErrInvalidAmount, ErrInvalidDate, ErrInvalidRate,
// ErrInfeasiblePlan or ErrOverflow when the input is invalid or too large to compute exactly in int64.
func CalculateInstallmentPlan(
	totalAmount domain.Money,
	numInstallments int,
	purchaseDate time.Time,
	firstDueDate time.Time,
	iofCfg config.IOFConfig,
	instCfg config.InstallmentConfig,
) (domain.InstallmentPlan, error) {
	if err := validateInstallmentInput(totalAmount, numInstallments, purchaseDate, firstDueDate, iofCfg, instCfg); err != nil {
		return domain.InstallmentPlan{}, err
	}
	installments := make([]domain.Installment, numInstallments)

	if instCfg.MonthlyRate == 0 {
		return calculateInterestFree(totalAmount, numInstallments, purchaseDate, firstDueDate, iofCfg, installments), nil
	}
	return calculateWithInterest(totalAmount, numInstallments, purchaseDate, firstDueDate, iofCfg, instCfg, installments)
}

// calculateInterestFree computes an interest-free installment plan (sem juros).
// Principal is divided equally; remainder centavos go to the first installment.
func calculateInterestFree(
	totalAmount domain.Money,
	n int,
	purchaseDate time.Time,
	firstDueDate time.Time,
	iofCfg config.IOFConfig,
	installments []domain.Installment,
) domain.InstallmentPlan {
	base := totalAmount / domain.Money(n)
	remainder := totalAmount - base*domain.Money(n)

	var totalIOF domain.Money
	for i := 0; i < n; i++ {
		dueDate := addMonths(firstDueDate, i)
		days := daysBetween(purchaseDate, dueDate)

		principal := base
		if i == 0 {
			principal += remainder
		}

		iof := CalculateIOF(principal, days, iofCfg)
		totalIOF += iof

		installments[i] = domain.Installment{
			Number:    i + 1,
			DueDate:   dueDate,
			Principal: principal,
			Interest:  0,
			IOF:       iof,
			Amount:    principal + iof,
		}
	}

	return domain.InstallmentPlan{
		TotalAmount:   totalAmount,
		TotalIOF:      totalIOF,
		TotalInterest: 0,
		TotalWithIOF:  totalAmount + totalIOF,
		Installments:  installments,
	}
}

// calculateWithInterest computes an installment plan using Tabela Price (PMT formula).
func calculateWithInterest(
	totalAmount domain.Money,
	n int,
	purchaseDate time.Time,
	firstDueDate time.Time,
	iofCfg config.IOFConfig,
	instCfg config.InstallmentConfig,
	installments []domain.Installment,
) (domain.InstallmentPlan, error) {
	r := instCfg.MonthlyRate

	pmt, err := priceInstallment(totalAmount, r, n)
	if err != nil {
		return domain.InstallmentPlan{}, err
	}

	balance := totalAmount
	var totalInterest, totalIOF domain.Money

	for i := range n {
		dueDate := addMonths(firstDueDate, i)
		days := daysBetween(purchaseDate, dueDate)

		interest := mulRate(balance, r)
		principal := pmt - interest

		// The payment is rounded to the centavo and that error compounds over the term.
		// If a non-final installment cannot amortize (or overshoots) the balance, the plan
		// is degenerate: reject it instead of producing negative balances or interest.
		if i < n-1 && (principal <= 0 || principal > balance) {
			return domain.InstallmentPlan{}, ErrInfeasiblePlan
		}

		// Last installment: adjust for rounding to ensure
		// balance reaches zero
		if i == n-1 {
			principal = balance
			interest = pmt - principal
			if interest < 0 {
				interest = mulRate(balance, r)
				principal = balance
			}
		}

		iof := CalculateIOF(principal, days, iofCfg)
		totalIOF += iof
		totalInterest += interest

		installments[i] = domain.Installment{
			Number:    i + 1,
			DueDate:   dueDate,
			Principal: principal,
			Interest:  interest,
			IOF:       iof,
			Amount:    principal + interest + iof,
		}

		balance -= principal
	}

	return domain.InstallmentPlan{
		TotalAmount:   totalAmount,
		TotalIOF:      totalIOF,
		TotalInterest: totalInterest,
		TotalWithIOF:  totalAmount + totalInterest + totalIOF,
		Installments:  installments,
	}, nil
}
