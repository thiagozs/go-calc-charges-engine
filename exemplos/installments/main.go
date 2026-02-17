package main

import (
	"fmt"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/calc"
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func main() {
	purchaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	firstDueDate := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	amount := domain.Money(100_000) // R$ 1.000,00

	iofCfg := config.IOFConfig{DailyRate: 82, AdditionalRate: 3_800, MaxAnnualRate: 40_800}

	// === Sem juros (10x) ===
	printSection("Parcelamento sem juros - 10x de R$ 1.000,00")

	plan := calc.CalculateInstallmentPlan(amount, 10, purchaseDate, firstDueDate, iofCfg, config.InstallmentConfig{MonthlyRate: 0})

	for _, inst := range plan.Installments {
		fmt.Printf("  Parcela %2d | Venc: %s | Principal: %s | IOF: %s | Total: %s\n",
			inst.Number,
			inst.DueDate.Format("2006-01-02"),
			formatMoney(inst.Principal),
			formatMoney(inst.IOF),
			formatMoney(inst.Amount),
		)
	}
	fmt.Printf("\n  Total compra: %s\n", formatMoney(plan.TotalAmount))
	fmt.Printf("  Total IOF: %s\n", formatMoney(plan.TotalIOF))
	fmt.Printf("  Total com IOF: %s\n", formatMoney(plan.TotalWithIOF))

	// === Com juros (12x a 1.99% a.m.) ===
	printSection("Parcelamento com juros - 12x de R$ 1.000,00 a 1.99% a.m.")

	planJuros := calc.CalculateInstallmentPlan(amount, 12, purchaseDate, firstDueDate, iofCfg, config.InstallmentConfig{MonthlyRate: 19_900})

	for _, inst := range planJuros.Installments {
		fmt.Printf("  Parcela %2d | Venc: %s | Principal: %s | Juros: %s | IOF: %s | Total: %s\n",
			inst.Number,
			inst.DueDate.Format("2006-01-02"),
			formatMoney(inst.Principal),
			formatMoney(inst.Interest),
			formatMoney(inst.IOF),
			formatMoney(inst.Amount),
		)
	}
	fmt.Printf("\n  Total compra: %s\n", formatMoney(planJuros.TotalAmount))
	fmt.Printf("  Total juros: %s\n", formatMoney(planJuros.TotalInterest))
	fmt.Printf("  Total IOF: %s\n", formatMoney(planJuros.TotalIOF))
	fmt.Printf("  Total com juros + IOF: %s\n", formatMoney(planJuros.TotalWithIOF))
}

func formatMoney(value domain.Money) string {
	return fmt.Sprintf("R$ %.2f", float64(value)/100.0)
}

func printSection(title string) {
	fmt.Printf("\n=== %s ===\n\n", title)
}
