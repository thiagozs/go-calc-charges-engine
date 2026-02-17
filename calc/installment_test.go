package calc

import (
	"testing"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

func TestInstallment_InterestFree_EqualDivision(t *testing.T) {
	purchaseDate := utcDate(2024, 1, 5)
	firstDueDate := utcDate(2024, 2, 10)
	iofCfg := defaultIOFConfig()
	instCfg := config.InstallmentConfig{MonthlyRate: 0}

	plan := CalculateInstallmentPlan(100_000, 10, purchaseDate, firstDueDate, iofCfg, instCfg)

	if len(plan.Installments) != 10 {
		t.Fatalf("expected 10 installments, got %d", len(plan.Installments))
	}

	// Each installment should be R$100 (10_000 centavos)
	for i, inst := range plan.Installments {
		if inst.Principal != 10_000 {
			t.Fatalf("installment %d: expected principal 10000, got %d", i+1, inst.Principal)
		}
		if inst.Interest != 0 {
			t.Fatalf("installment %d: expected interest 0, got %d", i+1, inst.Interest)
		}
	}

	if plan.TotalAmount != 100_000 {
		t.Fatalf("expected total amount 100000, got %d", plan.TotalAmount)
	}
	if plan.TotalInterest != 0 {
		t.Fatalf("expected total interest 0, got %d", plan.TotalInterest)
	}
}

func TestInstallment_InterestFree_RemainderInFirst(t *testing.T) {
	purchaseDate := utcDate(2024, 1, 5)
	firstDueDate := utcDate(2024, 2, 10)
	iofCfg := defaultIOFConfig()
	instCfg := config.InstallmentConfig{MonthlyRate: 0}

	// R$100 in 3 installments: 34 + 33 + 33
	plan := CalculateInstallmentPlan(10_000, 3, purchaseDate, firstDueDate, iofCfg, instCfg)

	if plan.Installments[0].Principal != 3_334 {
		t.Fatalf("first installment: expected principal 3334, got %d", plan.Installments[0].Principal)
	}
	if plan.Installments[1].Principal != 3_333 {
		t.Fatalf("second installment: expected principal 3333, got %d", plan.Installments[1].Principal)
	}
	if plan.Installments[2].Principal != 3_333 {
		t.Fatalf("third installment: expected principal 3333, got %d", plan.Installments[2].Principal)
	}

	// Verify sum
	var sum domain.Money
	for _, inst := range plan.Installments {
		sum += inst.Principal
	}
	if sum != 10_000 {
		t.Fatalf("sum of principals: expected 10000, got %d", sum)
	}
}

func TestInstallment_InterestFree_IOFPerInstallment(t *testing.T) {
	purchaseDate := utcDate(2024, 1, 5)
	firstDueDate := utcDate(2024, 2, 10)
	iofCfg := defaultIOFConfig()
	instCfg := config.InstallmentConfig{MonthlyRate: 0}

	plan := CalculateInstallmentPlan(100_000, 3, purchaseDate, firstDueDate, iofCfg, instCfg)

	// Each installment should have IOF > 0, and later installments have more IOF (more days)
	for i, inst := range plan.Installments {
		if inst.IOF <= 0 {
			t.Fatalf("installment %d: expected IOF > 0, got %d", i+1, inst.IOF)
		}
	}

	// IOF should increase with time (more days = more IOF)
	if plan.Installments[2].IOF <= plan.Installments[0].IOF {
		t.Fatalf("last installment IOF should be >= first installment IOF")
	}

	if plan.TotalIOF <= 0 {
		t.Fatalf("expected total IOF > 0, got %d", plan.TotalIOF)
	}
}

func TestInstallment_SingleInstallment(t *testing.T) {
	purchaseDate := utcDate(2024, 1, 5)
	firstDueDate := utcDate(2024, 2, 10)
	iofCfg := defaultIOFConfig()
	instCfg := config.InstallmentConfig{MonthlyRate: 0}

	plan := CalculateInstallmentPlan(100_000, 1, purchaseDate, firstDueDate, iofCfg, instCfg)

	if len(plan.Installments) != 1 {
		t.Fatalf("expected 1 installment, got %d", len(plan.Installments))
	}
	if plan.Installments[0].Principal != 100_000 {
		t.Fatalf("expected principal 100000, got %d", plan.Installments[0].Principal)
	}
	if plan.TotalInterest != 0 {
		t.Fatalf("expected 0 interest, got %d", plan.TotalInterest)
	}
}

func TestInstallment_WithInterest_PMT(t *testing.T) {
	purchaseDate := utcDate(2024, 1, 5)
	firstDueDate := utcDate(2024, 2, 10)
	iofCfg := defaultIOFConfig()
	// 1.99% monthly
	instCfg := config.InstallmentConfig{MonthlyRate: 19_900}

	plan := CalculateInstallmentPlan(100_000, 12, purchaseDate, firstDueDate, iofCfg, instCfg)

	if len(plan.Installments) != 12 {
		t.Fatalf("expected 12 installments, got %d", len(plan.Installments))
	}

	// Total interest should be positive
	if plan.TotalInterest <= 0 {
		t.Fatalf("expected positive interest, got %d", plan.TotalInterest)
	}

	// Sum of principal portions should equal the original amount
	var principalSum domain.Money
	for _, inst := range plan.Installments {
		principalSum += inst.Principal
	}
	if principalSum != 100_000 {
		t.Fatalf("sum of principals: expected 100000, got %d", principalSum)
	}

	// Each installment amount should include interest
	for i, inst := range plan.Installments {
		if inst.Interest < 0 {
			t.Fatalf("installment %d: negative interest %d", i+1, inst.Interest)
		}
	}
}

func TestInstallment_WithInterest_LargeAmount(t *testing.T) {
	purchaseDate := utcDate(2024, 1, 5)
	firstDueDate := utcDate(2024, 2, 10)
	iofCfg := defaultIOFConfig()
	instCfg := config.InstallmentConfig{MonthlyRate: 19_900}

	// R$500,000.00
	plan := CalculateInstallmentPlan(50_000_000, 12, purchaseDate, firstDueDate, iofCfg, instCfg)

	var principalSum domain.Money
	for _, inst := range plan.Installments {
		principalSum += inst.Principal
	}
	if principalSum != 50_000_000 {
		t.Fatalf("sum of principals: expected 50000000, got %d", principalSum)
	}
}

func TestInstallment_DueDateProgression(t *testing.T) {
	purchaseDate := utcDate(2024, 1, 5)
	firstDueDate := utcDate(2024, 2, 10)
	iofCfg := defaultIOFConfig()
	instCfg := config.InstallmentConfig{MonthlyRate: 0}

	plan := CalculateInstallmentPlan(100_000, 3, purchaseDate, firstDueDate, iofCfg, instCfg)

	expected := []time.Time{
		utcDate(2024, 2, 10),
		utcDate(2024, 3, 10),
		utcDate(2024, 4, 10),
	}

	for i, inst := range plan.Installments {
		if !inst.DueDate.Equal(expected[i]) {
			t.Fatalf("installment %d: expected due date %v, got %v", i+1, expected[i], inst.DueDate)
		}
	}
}
