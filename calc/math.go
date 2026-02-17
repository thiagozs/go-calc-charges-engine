package calc

import (
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

// addMonths adds the given number of months to a time.
func addMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// daysBetween returns the number of whole days between two times.
// Returns 0 if to is before from.
func daysBetween(from, to time.Time) int {
	d := int(to.Sub(from).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}
