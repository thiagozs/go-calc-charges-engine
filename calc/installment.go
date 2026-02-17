package calc

import (
	"time"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// fixedPow computes (1 + rate/RateDenominator)^n in fixed point with denominator = RateDenominator.
func fixedPow(rate domain.Rate, n int) int64 {
	result := domain.RateDenominator
	base := domain.RateDenominator + int64(rate)
	for i := 0; i < n; i++ {
		result = (result*base + domain.RateDenominator/2) / domain.RateDenominator
	}
	return result
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
// Input validation (positive amount, valid dates, n >= 1) is the caller's responsibility.
func CalculateInstallmentPlan(
	totalAmount domain.Money,
	numInstallments int,
	purchaseDate time.Time,
	firstDueDate time.Time,
	iofCfg config.IOFConfig,
	instCfg config.InstallmentConfig,
) domain.InstallmentPlan {
	installments := make([]domain.Installment, numInstallments)

	if instCfg.MonthlyRate == 0 {
		return calculateInterestFree(totalAmount, numInstallments, purchaseDate, firstDueDate, iofCfg, installments)
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
) domain.InstallmentPlan {
	r := instCfg.MonthlyRate

	// PMT = PV * r * (1+r)^n / ((1+r)^n - 1)
	pow := fixedPow(r, n)
	pmtNum := int64(r) * pow
	pmtDen := (pow - domain.RateDenominator) * domain.RateDenominator

	pmt := domain.Money((int64(totalAmount)*pmtNum + pmtDen/2) / pmtDen)

	balance := totalAmount
	var totalInterest, totalIOF domain.Money

	for i := range n {
		dueDate := addMonths(firstDueDate, i)
		days := daysBetween(purchaseDate, dueDate)

		interest := mulRate(balance, r)
		principal := pmt - interest

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
	}
}
