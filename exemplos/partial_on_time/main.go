package main

import (
	"fmt"
	"time"

	"github.com/thiagozs/go-credit-card-engine/calc"
	"github.com/thiagozs/go-credit-card-engine/config"
	"github.com/thiagozs/go-credit-card-engine/domain"
)

func main() {
	cycleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	closingDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC) // fecha na madrugada
	dueDate := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	partialPaymentDate := time.Date(2024, 2, 10, 18, 45, 0, 0, time.UTC) // pago no mesmo dia do vencimento
	finalPaymentDate := time.Date(2024, 2, 25, 10, 0, 0, 0, time.UTC)    // quitação após vencimento

	transactions := []domain.Transaction{
		{ID: "t1", Amount: 35_000, Date: time.Date(2024, 1, 5, 13, 0, 0, 0, time.UTC)},
		{ID: "t2", Amount: 12_500, Date: time.Date(2024, 1, 12, 9, 0, 0, 0, time.UTC), International: true},
		{ID: "t3", Amount: 8_900, Date: time.Date(2024, 1, 20, 18, 0, 0, 0, time.UTC)},
		{ID: "t4", Amount: 5_200, Date: time.Date(2024, 1, 30, 21, 0, 0, 0, time.UTC)},
	}

	principal := sumTransactionsInCycle(transactions, cycleStart, closingDate)
	internationalIOF := sumInternationalIOFInCycle(transactions, cycleStart, closingDate, config.InternationalIOFConfig{Rate: 0.035})

	invoice := domain.Invoice{
		ID:          "inv-2024-01",
		ClosingDate: closingDate,
		DueDate:     dueDate,
		TotalAmount: principal + internationalIOF,
	}

	partialPayment := domain.Money(40_000) // cliente paga parte do total no vencimento
	if isSameDay(partialPaymentDate, dueDate) {
		invoice.PaidAmount = partialPayment
	}

	printSection("Linha do tempo")
	fmt.Printf("Ciclo: %s -> %s\n", cycleStart.Format("2006-01-02"), closingDate.Format("2006-01-02"))
	fmt.Printf("Vencimento: %s\n", dueDate.Format("2006-01-02"))
	fmt.Printf("Pagamento parcial: %s\n", partialPaymentDate.Format("2006-01-02 15:04"))
	fmt.Printf("Pagamento final: %s\n", finalPaymentDate.Format("2006-01-02 15:04"))

	printSection("Transacoes no ciclo")
	for _, tx := range transactions {
		if inCycle(tx.Date, cycleStart, closingDate) {
			note := ""
			if tx.International {
				note = " (internacional)"
			}
			fmt.Printf("- %s | %s%s\n", tx.Date.Format("2006-01-02"), formatMoney(tx.Amount), note)
		}
	}

	printSection("Fatura no fechamento")
	fmt.Printf("- Principal: %s\n", formatMoney(principal))
	fmt.Printf("- IOF internacional: %s\n", formatMoney(internationalIOF))
	fmt.Printf("- Total no fechamento: %s\n", formatMoney(invoice.TotalAmount))
	fmt.Printf("- Fechamento: %s\n", invoice.ClosingDate.Format("2006-01-02 15:04"))
	fmt.Printf("- Vencimento: %s\n", invoice.DueDate.Format("2006-01-02"))

	printSection("Pagamento parcial no vencimento")
	fmt.Printf("- Pago: %s\n", formatMoney(invoice.PaidAmount))
	fmt.Printf("- Em aberto: %s\n", formatMoney(invoice.TotalAmount-invoice.PaidAmount))

	remaining := invoice.TotalAmount - invoice.PaidAmount
	if remaining > 0 && finalPaymentDate.After(dueDate) {
		iofCfg := config.IOFConfig{DailyRate: 0.000082, AdditionalRate: 0.0038, MaxAnnualRate: 0.0408}
		intCfg := config.InterestConfig{MonthlyRate: 0.12}
		lateFeeCfg := config.LateFeeConfig{Rate: 0.02}
		lateInterestCfg := config.LateInterestConfig{MonthlyRate: 0.01}
		rulesCfg := config.RotativeRulesConfig{MaxDays: 30, MaxChargeRate: 1.0}

		rotative := calc.CalculateRotative(
			domain.RotativeBalance{Principal: remaining, StartDate: dueDate},
			finalPaymentDate,
			iofCfg,
			intCfg,
			lateFeeCfg,
			lateInterestCfg,
			rulesCfg,
		)

		invoice.TotalAmount = rotative.Total

		printSection("Encargos ate a quitacao")
		fmt.Printf("- Dias em atraso: %d (cobrados %d)\n", rotative.Days, rotative.ChargedDays)
		fmt.Printf("- Saldo em atraso: %s\n", formatMoney(remaining))
		fmt.Printf("- IOF: %s\n", formatMoney(rotative.IOF))
		fmt.Printf("- Juros rotativo: %s\n", formatMoney(rotative.Interest))
		fmt.Printf("- Juros de mora: %s\n", formatMoney(rotative.LateInterest))
		fmt.Printf("- Multa: %s\n", formatMoney(rotative.LateFee))
		fmt.Printf("- Total encargos: %s\n", formatMoney(rotative.Charges))
		if rotative.ChargeCapped {
			fmt.Printf("- Teto aplicado: sim\n")
		} else {
			fmt.Printf("- Teto aplicado: nao\n")
		}

		printSection("Fatura final apos encargos")
		fmt.Printf("- Total: %s\n", formatMoney(invoice.TotalAmount))
	}
}

func sumTransactionsInCycle(transactions []domain.Transaction, start, end time.Time) domain.Money {
	var total domain.Money
	for _, tx := range transactions {
		if inCycle(tx.Date, start, end) {
			total += tx.Amount
		}
	}
	return total
}

func sumInternationalIOFInCycle(transactions []domain.Transaction, start, end time.Time, cfg config.InternationalIOFConfig) domain.Money {
	var total domain.Money
	for _, tx := range transactions {
		if tx.International && inCycle(tx.Date, start, end) {
			total += calc.CalculateInternationalIOF(tx.Amount, cfg)
		}
	}
	return total
}

func inCycle(date, start, end time.Time) bool {
	if date.Before(start) {
		return false
	}
	if !date.Before(end) && !date.Equal(end) {
		return false
	}
	return true
}

func isSameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func formatMoney(value domain.Money) string {
	return fmt.Sprintf("R$ %.2f", float64(value)/100.0)
}

func printSection(title string) {
	fmt.Printf("\n=== %s ===\n", title)
}
