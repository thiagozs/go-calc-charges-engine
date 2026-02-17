package domain

import "time"

type Money int64

// Rate represents a fixed-point rate with 6 decimal places.
// The value is the numerator; the implicit denominator is RateDenominator (1_000_000).
// For example, 0.000082 is represented as Rate(82), and 0.12 as Rate(120_000).
type Rate int64

// RateDenominator is the implicit denominator for Rate values.
const RateDenominator int64 = 1_000_000

type Transaction struct {
	ID            string
	Amount        Money
	Date          time.Time
	International bool
}
