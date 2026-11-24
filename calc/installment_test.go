package calc

import (
	"errors"
	"math/big"
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

	plan := mustPlan(t, 100_000, 10, purchaseDate, firstDueDate, iofCfg, instCfg)

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
	plan := mustPlan(t, 10_000, 3, purchaseDate, firstDueDate, iofCfg, instCfg)

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

	plan := mustPlan(t, 100_000, 3, purchaseDate, firstDueDate, iofCfg, instCfg)

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

	plan := mustPlan(t, 100_000, 1, purchaseDate, firstDueDate, iofCfg, instCfg)

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

	plan := mustPlan(t, 100_000, 12, purchaseDate, firstDueDate, iofCfg, instCfg)

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
	plan := mustPlan(t, 50_000_000, 12, purchaseDate, firstDueDate, iofCfg, instCfg)

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

	plan := mustPlan(t, 100_000, 3, purchaseDate, firstDueDate, iofCfg, instCfg)

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

func mustPlan(t *testing.T, amount domain.Money, n int, purchase, first time.Time, iof config.IOFConfig, inst config.InstallmentConfig) domain.InstallmentPlan {
	t.Helper()
	plan, err := CalculateInstallmentPlan(amount, n, purchase, first, iof, inst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return plan
}

func TestInstallment_InvalidInput(t *testing.T) {
	d := utcDate(2024, 2, 10)
	p := utcDate(2024, 1, 5)
	iof := defaultIOFConfig()
	free := config.InstallmentConfig{}
	cases := []struct {
		name   string
		amount domain.Money
		n      int
		pur    time.Time
		first  time.Time
		iof    config.IOFConfig
		inst   config.InstallmentConfig
		want   error
	}{
		{"zero installments", 100, 0, p, d, iof, free, ErrInvalidInstallments},
		{"negative installments", 100, -1, p, d, iof, free, ErrInvalidInstallments},
		{"too many installments", 100, MaxInstallments + 1, p, d, iof, free, ErrInvalidInstallments},
		{"zero amount", 0, 3, p, d, iof, free, ErrInvalidAmount},
		{"negative amount", -5, 3, p, d, iof, free, ErrInvalidAmount},
		{"zero purchase date", 100, 3, time.Time{}, d, iof, free, ErrInvalidDate},
		{"due before purchase", 100, 3, d, p, iof, free, ErrInvalidDate},
		{"negative rate", 100, 3, p, d, iof, config.InstallmentConfig{MonthlyRate: -1}, ErrInvalidRate},
		{"amount overflow", 1 << 62, 3, p, d, iof, free, ErrOverflow},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := CalculateInstallmentPlan(c.amount, c.n, c.pur, c.first, c.iof, c.inst)
			if !errors.Is(err, c.want) {
				t.Fatalf("expected %v, got %v", c.want, err)
			}
		})
	}
}

func TestInstallment_LargeAmountNoOverflow(t *testing.T) {
	// R$ 5,000,000.00 at 5% a.m. used to overflow int64 in the PMT calculation.
	plan := mustPlan(t, 500_000_000, 12, utcDate(2025, 1, 31), utcDate(2025, 1, 31),
		config.IOFConfig{}, config.InstallmentConfig{MonthlyRate: 50_000})
	var principal domain.Money
	for _, in := range plan.Installments {
		principal += in.Principal
		if in.Amount < in.Principal || in.Amount > 70_000_000 {
			t.Fatalf("installment %d amount out of range: %d", in.Number, in.Amount)
		}
	}
	if principal != 500_000_000 {
		t.Fatalf("principal sum: got %d", principal)
	}
	// Exact PMT is ~R$564,100; total interest ~R$1.77M, nowhere near the overflowed 3.58M.
	if plan.TotalInterest > 200_000_000 {
		t.Fatalf("implausible total interest %d", plan.TotalInterest)
	}
}

func TestInstallment_EndOfMonthDueDates(t *testing.T) {
	plan := mustPlan(t, 30_000, 3, utcDate(2025, 1, 10), utcDate(2025, 1, 31), config.IOFConfig{}, config.InstallmentConfig{})
	want := []time.Time{utcDate(2025, 1, 31), utcDate(2025, 2, 28), utcDate(2025, 3, 31)}
	for i, in := range plan.Installments {
		if !in.DueDate.Equal(want[i]) {
			t.Fatalf("installment %d: want %v, got %v", i+1, want[i], in.DueDate)
		}
	}
}

func TestInstallment_InfeasibleLongPlanRejected(t *testing.T) {
	// R$ 1.000,00 in 360x at 1.99% a.m.: the payment barely covers interest, so rounding
	// drives the balance negative mid-plan. Found by FuzzInstallmentPlan.
	_, err := CalculateInstallmentPlan(100_000, 360, utcDate(2024, 1, 1), utcDate(2024, 2, 1),
		config.IOFConfig{}, config.InstallmentConfig{MonthlyRate: 19_900})
	if !errors.Is(err, ErrInfeasiblePlan) {
		t.Fatalf("expected ErrInfeasiblePlan, got %v", err)
	}
}

// referencePMT is an independent big.Rat evaluation of the Price formula, rounded half up.
func referencePMT(pv int64, rate int64, n int) int64 {
	r := big.NewRat(rate, 1_000_000)
	onePlus := new(big.Rat).Add(big.NewRat(1, 1), r)
	pow := big.NewRat(1, 1)
	for i := 0; i < n; i++ {
		pow.Mul(pow, onePlus)
	}
	// PV * r * pow / (pow - 1)
	v := new(big.Rat).Mul(big.NewRat(pv, 1), r)
	v.Mul(v, pow)
	v.Quo(v, new(big.Rat).Sub(pow, big.NewRat(1, 1)))
	// round half up
	v.Add(v, big.NewRat(1, 2))
	return new(big.Int).Quo(v.Num(), v.Denom()).Int64()
}

func TestInstallment_PMTMatchesExactFormula(t *testing.T) {
	for _, pv := range []int64{10_000, 100_000, 1_000_000, 50_000_000, 500_000_000} {
		for _, rate := range []int64{1, 5_000, 19_900, 50_000, 120_000} {
			for _, n := range []int{1, 2, 3, 6, 12, 24, 48, 60} {
				plan, err := CalculateInstallmentPlan(domain.Money(pv), n, utcDate(2024, 1, 1), utcDate(2024, 2, 1),
					config.IOFConfig{}, config.InstallmentConfig{MonthlyRate: domain.Rate(rate)})
				if err != nil {
					continue
				}
				first := plan.Installments[0]
				if n == 1 {
					continue
				}
				got := int64(first.Principal + first.Interest)
				if want := referencePMT(pv, rate, n); got != want {
					t.Fatalf("pv=%d rate=%d n=%d: PMT %d, exact %d", pv, rate, n, got, want)
				}
			}
		}
	}
}

func TestInstallment_KnownValue(t *testing.T) {
	// R$ 1.000,00 in 12x at 1.99% a.m. -> PMT = R$ 94,50 (1.0199^12 ≈ 1.266743 -> 1000*0.0199*1.266743/0.266743).
	plan := mustPlan(t, 100_000, 12, utcDate(2024, 1, 1), utcDate(2024, 2, 1), config.IOFConfig{}, config.InstallmentConfig{MonthlyRate: 19_900})
	in := plan.Installments[0]
	if in.Principal+in.Interest != 9_450 {
		t.Fatalf("expected PMT 9450, got %d", in.Principal+in.Interest)
	}
	if in.Interest != 1_990 {
		t.Fatalf("expected first interest 1990, got %d", in.Interest)
	}
}
