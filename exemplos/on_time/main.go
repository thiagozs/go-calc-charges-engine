package main

import (
	"fmt"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/calc"
	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func main() {
	cycleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	closingDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC) // fecha na madrugada
	dueDate := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	paymentDate := time.Date(2024, 2, 10, 10, 30, 0, 0, time.UTC) // pago no mesmo dia do vencimento

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

	payment := invoice.TotalAmount // cliente paga o total no vencimento
	if isSameDay(paymentDate, dueDate) {
		invoice.PaidAmount = payment
	}

	printSection("Linha do tempo")
	fmt.Printf("Ciclo: %s -> %s\n", cycleStart.Format("2006-01-02"), closingDate.Format("2006-01-02"))
	fmt.Printf("Vencimento: %s\n", dueDate.Format("2006-01-02"))
	fmt.Printf("Pagamento: %s\n", paymentDate.Format("2006-01-02 15:04"))

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

	printSection("Pagamento no vencimento (sem encargos)")
	fmt.Printf("- Pago: %s\n", formatMoney(invoice.PaidAmount))
	fmt.Printf("- Em aberto: %s\n", formatMoney(invoice.TotalAmount-invoice.PaidAmount))
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
