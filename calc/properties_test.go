package calc

import (
	"testing"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// FuzzInstallmentPlan checks plan-wide invariants for arbitrary inputs.
// Inputs rejected by validation are skipped; everything accepted must be consistent.
func FuzzInstallmentPlan(f *testing.F) {
	f.Add(int64(100_000), 10, int64(0), 30, 5)
	f.Add(int64(100_000), 12, int64(19_900), 30, 5)
	f.Add(int64(500_000_000), 12, int64(50_000), 0, 0)
	f.Add(int64(1), 7, int64(1), 1, 0)
	f.Add(int64(999_999_999), 360, int64(120_000), 400, 31)
	f.Fuzz(func(t *testing.T, amount int64, n int, rate int64, firstOffsetDays int, purchaseDay int) {
		if firstOffsetDays < 0 || firstOffsetDays > 3650 || purchaseDay < 0 || purchaseDay > 400 {
			t.Skip()
		}
		purchase := utcDate(2024, 1, 1).AddDate(0, 0, purchaseDay)
		first := purchase.AddDate(0, 0, firstOffsetDays)
		iof := defaultIOFConfig()
		plan, err := CalculateInstallmentPlan(domain.Money(amount), n, purchase, first, iof, config.InstallmentConfig{MonthlyRate: domain.Rate(rate)})
		if err != nil {
			return
		}
		if len(plan.Installments) != n {
			t.Fatalf("expected %d installments, got %d", n, len(plan.Installments))
		}
		var principal, interest, iofSum, amounts domain.Money
		prev := time.Time{}
		for i, in := range plan.Installments {
			if in.Number != i+1 {
				t.Fatalf("installment numbering: %d at index %d", in.Number, i)
			}
			if in.Principal < 0 || in.Interest < 0 || in.IOF < 0 {
				t.Fatalf("negative component in %+v", in)
			}
			if in.Amount != in.Principal+in.Interest+in.IOF {
				t.Fatalf("amount mismatch in %+v", in)
			}
			if i > 0 && !in.DueDate.After(prev) {
				t.Fatalf("due dates not increasing: %v then %v", prev, in.DueDate)
			}
			prev = in.DueDate
			principal += in.Principal
			interest += in.Interest
			iofSum += in.IOF
			amounts += in.Amount
		}
		if principal != domain.Money(amount) {
			t.Fatalf("sum of principals %d != amount %d", principal, amount)
		}
		if interest != plan.TotalInterest || iofSum != plan.TotalIOF {
			t.Fatalf("totals mismatch: %+v", plan)
		}
		if plan.TotalWithIOF != amounts {
			t.Fatalf("TotalWithIOF %d != sum of installments %d", plan.TotalWithIOF, amounts)
		}
		if rate == 0 && interest != 0 {
			t.Fatalf("interest-free plan has interest %d", interest)
		}
	})
}

// FuzzRotative checks the charge breakdown invariants.
func FuzzRotative(f *testing.F) {
	f.Add(int64(100_000), 15)
	f.Add(int64(100_000), 90)
	f.Add(int64(0), 10)
	f.Add(int64(1), 1)
	f.Add(int64(9_000_000_000_000), 30)
	f.Fuzz(func(t *testing.T, principal int64, days int) {
		if days < -400 || days > 4000 {
			t.Skip()
		}
		start := utcDate(2024, 1, 1)
		rules := defaultRotativeRulesConfig()
		r, err := CalculateRotative(domain.RotativeBalance{Principal: domain.Money(principal), StartDate: start},
			start.AddDate(0, 0, days), defaultIOFConfig(), defaultInterestConfig(),
			defaultLateFeeConfig(), defaultLateInterestConfig(), rules)
		if err != nil {
			return
		}
		if r.Interest < 0 || r.IOF < 0 || r.LateFee < 0 || r.LateInterest < 0 {
			t.Fatalf("negative component: %+v", r)
		}
		if r.FinancialCharges != r.Interest+r.LateFee+r.LateInterest || r.Charges != r.FinancialCharges+r.IOF {
			t.Fatalf("charges mismatch: %+v", r)
		}
		if r.Total != r.Principal+r.Charges {
			t.Fatalf("total mismatch: %+v", r)
		}
		if r.ChargedDays > r.Days || r.ChargedDays > rules.MaxDays && rules.MaxDays > 0 {
			t.Fatalf("charged days out of range: %+v", r)
		}
		// The cap covers financial charges only (never IOF).
		if limit := mulRate(r.Principal, rules.MaxChargeRate); r.FinancialCharges > limit {
			t.Fatalf("financial charges %d exceed cap %d: %+v", r.FinancialCharges, limit, r)
		}
		// The result must feed ApplyPayment without error and be fully payable.
		a, err := ApplyPayment(r.Debt(), r.Total)
		if err != nil || a.Remaining != 0 || a.Excess != 0 {
			t.Fatalf("paying the total should settle the debt: %+v %v", a, err)
		}
	})
}

// FuzzDates checks that date helpers stay monotonic and non-negative.
func FuzzDates(f *testing.F) {
	f.Add(2024, 1, 31, 1)
	f.Add(2025, 2, 28, 12)
	f.Add(2024, 12, 31, 359)
	f.Fuzz(func(t *testing.T, y, m, d, months int) {
		if y < 1970 || y > 2200 || m < 1 || m > 12 || d < 1 || d > 31 || months < 0 || months > 1200 {
			t.Skip()
		}
		base := utcDate(y, time.Month(m), d)
		if base.Day() != d { // e.g. Feb 31 normalised by time.Date: not a real date
			t.Skip()
		}
		got := addMonths(base, months)
		if got.Day() > d {
			t.Fatalf("addMonths(%v,%d)=%v moved the day forward", base, months, got)
		}
		if got.Before(base) {
			t.Fatalf("addMonths(%v,%d)=%v went backwards", base, months, got)
		}
		next := addMonths(base, months+1)
		if !next.After(got) {
			t.Fatalf("addMonths not strictly increasing: %v then %v", got, next)
		}
		if daysBetween(base, got) < 0 || daysBetween(got, base) != 0 {
			t.Fatalf("daysBetween sign wrong for %v, %v", base, got)
		}
	})
}
