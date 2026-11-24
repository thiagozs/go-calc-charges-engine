package domain

import (
	"fmt"
	"testing"
)

func TestParseRate(t *testing.T) {
	ok := map[string]Rate{
		"0.12":      120_000,
		"0.000082":  82,
		"1.0":       1_000_000,
		"0.0":       0,
		"0.035":     35_000,
		"12.5":      12_500_000,
		".5":        500_000,
		"1.":        1_000_000,
		"0.1200000": 120_000, // trailing zeros beyond 6 places are fine
		"-0.01":     -10_000,
		"+0.02":     20_000,
		" 0.02 ":    20_000,
		"120000":    120_000, // raw fixed-point integer
		"82":        82,
		"0":         0,
		"-5":        -5,
	}
	for in, want := range ok {
		got, err := ParseRate(in)
		if err != nil || got != want {
			t.Errorf("ParseRate(%q) = %d, %v; want %d", in, got, err, want)
		}
	}

	bad := []string{"", " ", ".", "-.", "abc", "0.1.2", "0.0000001", "1e3", "0.1a", "--1", "0.-1", "9223372036855.0", "99999999999999999999"}
	for _, in := range bad {
		if got, err := ParseRate(in); err == nil {
			t.Errorf("ParseRate(%q) = %d; want error", in, got)
		}
	}
}

// FuzzParseRate checks that parsing never panics and that decimal output round-trips.
func FuzzParseRate(f *testing.F) {
	for _, s := range []string{"0.12", "120000", "-0.01", "1.", ".5", "abc", "0.0000001", "9223372036855.0"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		r, err := ParseRate(s)
		if err != nil {
			return
		}
		// Re-render as decimal and parse again: must give the same value.
		neg := r < 0
		abs := int64(r)
		if neg {
			abs = -abs
		}
		if abs < 0 { // MinInt64
			return
		}
		text := fmt.Sprintf("%d.%06d", abs/RateDenominator, abs%RateDenominator)
		if neg {
			text = "-" + text
		}
		back, err := ParseRate(text)
		if err != nil || back != r {
			t.Fatalf("round trip of %q: %d -> %q -> %d, %v", s, r, text, back, err)
		}
	})
}
