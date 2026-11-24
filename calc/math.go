package calc

import (
	"math"
	"math/big"
	"time"

	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// mulRate computes (principal * rate) rounded to the nearest centavo.
// Uses round-half-up: (principal * rate + RateDenominator/2) / RateDenominator.
func mulRate(principal domain.Money, rate domain.Rate) domain.Money {
	return domain.Money((int64(principal)*int64(rate) + domain.RateDenominator/2) / domain.RateDenominator)
}

// mulRateDays computes (principal * rate * days) rounded to the nearest centavo.
// The full product is accumulated before dividing to avoid premature truncation.
func mulRateDays(principal domain.Money, rate domain.Rate, days int) domain.Money {
	return domain.Money((int64(principal)*int64(rate)*int64(days) + domain.RateDenominator/2) / domain.RateDenominator)
}

// addMonths adds the given number of months to a time, clamping the day to the
// last day of the target month (Jan 31 + 1 month = Feb 28/29, not Mar 3).
// Always compute from the original anchor date to avoid drift (Jan 31 -> Feb 28 -> Mar 31).
func addMonths(t time.Time, months int) time.Time {
	y, m, d := t.Date()
	lastDay := time.Date(y, m+time.Month(months)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if d > lastDay {
		d = lastDay
	}
	return time.Date(y, m+time.Month(months), d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// daysBetween returns the number of calendar days between two times, using each
// time's own civil date (year/month/day). This is independent of time zone offsets
// and DST transitions. Returns 0 if to is before from.
func daysBetween(from, to time.Time) int {
	fy, fm, fd := from.Date()
	ty, tm, td := to.Date()
	a := time.Date(fy, fm, fd, 0, 0, 0, 0, time.UTC)
	b := time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
	d := int(b.Sub(a) / (24 * time.Hour))
	if d < 0 {
		return 0
	}
	return d
}

// maxSafeProduct leaves headroom for the rounding term (denominator/2) added
// before the final division in mulRate-style helpers.
var maxSafeProduct = big.NewInt(math.MaxInt64 - domain.DaysInMonth*domain.RateDenominator)

// mulFits reports whether the product of all non-negative factors, plus the
// rounding headroom, fits in int64.
func mulFits(factors ...int64) bool {
	acc := new(big.Int).SetInt64(1)
	for _, f := range factors {
		acc.Mul(acc, big.NewInt(f))
	}
	return acc.Cmp(maxSafeProduct) <= 0
}
