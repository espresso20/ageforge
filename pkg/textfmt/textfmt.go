// Package textfmt is the one place player-facing numbers, counts and
// durations are formatted. ui, game, boon and flavor all import it, so a
// panel, a log line and a flavor sentence print the same amount the same way:
// 950, 12.5K, 1.23M.
//
// It is a leaf package (standard library only) so anything can import it.
package textfmt

import (
	"fmt"
	"math"
	"strings"
	"time"
)

var suffixes = []struct {
	threshold float64
	suffix    string
}{
	{1e15, "Q"},
	{1e12, "T"},
	{1e9, "B"},
	{1e6, "M"},
	{1e3, "K"},
}

// Number formats an amount for display: whole numbers below 1000 print as
// is (950), fractions below 1000 keep one decimal (12.5), and larger values
// use a K/M/B/T/Q suffix with three significant figures (12.5K, 1.23M, 876M).
func Number(n float64) string {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return "0"
	}
	prefix := ""
	if n < 0 {
		prefix = "-"
	}
	abs := math.Abs(n)

	if abs < 1000 {
		if abs == math.Floor(abs) {
			return fmt.Sprintf("%s%.0f", prefix, abs)
		}
		s := fmt.Sprintf("%.1f", abs)
		if s == "1000.0" {
			return prefix + "1.00K"
		}
		if s == "0.0" {
			// Keep small fractions visible (0.04 is not "0.0").
			return prefix + trimZeros(fmt.Sprintf("%.2f", abs))
		}
		return prefix + s
	}

	for _, s := range suffixes {
		if abs >= s.threshold {
			scaled := abs / s.threshold
			switch {
			case scaled >= 100:
				return fmt.Sprintf("%s%.0f%s", prefix, scaled, s.suffix)
			case scaled >= 10:
				return fmt.Sprintf("%s%.1f%s", prefix, scaled, s.suffix)
			default:
				return fmt.Sprintf("%s%.2f%s", prefix, scaled, s.suffix)
			}
		}
	}
	return fmt.Sprintf("%s%.0f", prefix, abs)
}

// Int formats a whole count with Number.
func Int(n int) string { return Number(float64(n)) }

// Signed is Number with an explicit sign: "+5", "-1.2K", "+0".
func Signed(n float64) string {
	if n < 0 {
		return Number(n)
	}
	return "+" + Number(n)
}

// Rate formats a per-tick rate with its unit: "+3.25/tick", "-0.12/tick".
// Rates below 1 keep enough decimals to show a non-zero digit.
func Rate(r float64) string {
	return RateValue(r) + "/tick"
}

// RateValue is Rate without the "/tick" unit, for tables whose header
// already names the unit.
func RateValue(r float64) string {
	abs := math.Abs(r)
	sign := "+"
	if r < 0 {
		sign = "-"
	}
	if abs == 0 {
		return "+0"
	}
	if abs < 1 {
		prec := 2
		for prec < 6 && math.Round(abs*math.Pow10(prec)) == 0 {
			prec++
		}
		return sign + fmt.Sprintf("%.*f", prec, abs)
	}
	if abs < 100 && abs != math.Floor(abs) {
		return sign + trimZeros(fmt.Sprintf("%.2f", abs))
	}
	return sign + Number(abs)
}

// Percent formats a fraction as a whole percentage: 0.125 → "13%".
func Percent(frac float64) string {
	return fmt.Sprintf("%.0f%%", frac*100)
}

// SignedPercent formats a fraction as a signed percentage: 0.1 → "+10%".
func SignedPercent(frac float64) string {
	if frac < 0 {
		return fmt.Sprintf("-%.0f%%", -frac*100)
	}
	return fmt.Sprintf("+%.0f%%", frac*100)
}

// Plural picks the singular or plural noun for n: Plural(1, "item", "items").
func Plural(n int, one, many string) string {
	if n == 1 || n == -1 {
		return one
	}
	return many
}

// Count prints n with the right noun: Count(1, "worker", "workers") is
// "1 worker", Count(3, ...) is "3 workers". Large counts use Number.
func Count(n int, one, many string) string {
	return Int(n) + " " + Plural(n, one, many)
}

// Duration renders d at two units of precision, largest first, dropping a
// zero second unit: "45s", "4m 44s", "5m", "1h 12m", "2h", "3d 4h".
// Zero and negative durations render "0s".
func Duration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	total := int(d.Round(time.Second) / time.Second)
	switch {
	case total < 60:
		return fmt.Sprintf("%ds", total)
	case total < 3600:
		if s := total % 60; s != 0 {
			return fmt.Sprintf("%dm %ds", total/60, s)
		}
		return fmt.Sprintf("%dm", total/60)
	case total < 86400:
		if m := (total % 3600) / 60; m != 0 {
			return fmt.Sprintf("%dh %dm", total/3600, m)
		}
		return fmt.Sprintf("%dh", total/3600)
	default:
		if h := (total % 86400) / 3600; h != 0 {
			return fmt.Sprintf("%dd %dh", total/86400, h)
		}
		return fmt.Sprintf("%dd", total/86400)
	}
}

// Ticks renders a tick count as an approximate wall-clock duration ("~2m")
// given the length of one tick. The tilde is honest: the reading moves with
// game speed.
func Ticks(ticks int, interval time.Duration) string {
	return "~" + Duration(time.Duration(ticks)*interval)
}

// Capitalize upper-cases the first letter of s.
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Sentence capitalizes s and makes sure it ends with terminal punctuation.
func Sentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	s = Capitalize(s)
	switch s[len(s)-1] {
	case '.', '?', '!', ')', ':':
		return s
	}
	return s + "."
}

// List joins items as "a, b and c".
func List(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

func trimZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}
