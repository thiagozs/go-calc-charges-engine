package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// rateScale is the number of decimal places of a Rate (RateDenominator = 10^rateScale).
const rateScale = 6

// ParseRate parses a rate from text.
//
// Two formats are accepted:
//   - Decimal, when the text contains a '.': "0.12" -> Rate(120_000), "0.000082" -> Rate(82),
//     "1.0" -> Rate(1_000_000). At most 6 fractional digits (trailing zeros beyond that are fine).
//   - Raw fixed-point integer, when there is no '.': "120000" -> Rate(120_000).
//
// A bare integer is always the raw numerator, never a fraction: "1" is Rate(1), not 100%.
// Write "1.0" for 100%. Parsing is exact (no floating point); values that would
// lose precision or overflow int64 return an error.
func ParseRate(s string) (Rate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("rate: empty value")
	}
	if !strings.Contains(s, ".") {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("rate: invalid value %q", s)
		}
		return Rate(n), nil
	}

	neg := false
	switch s[0] {
	case '-':
		neg, s = true, s[1:]
	case '+':
		s = s[1:]
	}
	intPart, fracPart, _ := strings.Cut(s, ".")
	if intPart == "" && fracPart == "" {
		return 0, fmt.Errorf("rate: invalid value %q", s)
	}
	if !allDigits(intPart) || !allDigits(fracPart) {
		return 0, fmt.Errorf("rate: invalid value %q", s)
	}
	if len(fracPart) > rateScale {
		if strings.Trim(fracPart[rateScale:], "0") != "" {
			return 0, fmt.Errorf("rate: %q has more than %d decimal places", s, rateScale)
		}
		fracPart = fracPart[:rateScale]
	}
	fracPart += strings.Repeat("0", rateScale-len(fracPart))

	var whole int64
	if intPart != "" {
		var err error
		whole, err = strconv.ParseInt(intPart, 10, 64)
		if err != nil || whole > (math.MaxInt64-RateDenominator)/RateDenominator {
			return 0, fmt.Errorf("rate: %q out of range", s)
		}
	}
	frac, _ := strconv.ParseInt(fracPart, 10, 64)
	v := whole*RateDenominator + frac
	if neg {
		v = -v
	}
	return Rate(v), nil
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
