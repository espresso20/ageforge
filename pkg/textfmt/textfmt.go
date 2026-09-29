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
	"strconv"
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

// Number formats an amount or rate for display: three significant figures,
// trailing zeros dropped, and a K/M/B/T/Q suffix from a thousand up
// (950, 12.5, 0.711, 1.5K, 1.23M, 876M). Rounding goes through strconv, so
// the output is identical on every architecture (config descriptions built
// with it are part of the determinism fingerprint).
func Number(n float64) string {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return "0"
	}
	prefix := ""
	if n < 0 {
		prefix = "-"
	}
	abs := math.Abs(n)
	if abs == 0 {
		return "0"
	}
	if abs < 1000 {
		s := sig3(abs)
		if s != "1000" {
			return prefix + s
		}
		abs = 1000 // 999.96 rolls over to 1K
	}
	for i := len(suffixes) - 1; i >= 0; i-- {
		// walk from K upwards so a rollover (999.6K) moves to the next suffix
		s := suffixes[i]
		if abs < s.threshold {
			continue
		}
		next := math.Inf(1)
		if i > 0 {
			next = suffixes[i-1].threshold
		}
		if abs >= next {
			continue
		}
		str := sig3(abs / s.threshold)
		if str == "1000" && i > 0 {
			return prefix + "1" + suffixes[i-1].suffix
		}
		return prefix + str + s.suffix
	}
	return prefix + sig3(abs)
}

// sig3 prints v (0 < v < ~1000) with three significant figures and no
// exponent or trailing zeros.
func sig3(v float64) string {
	if v < 0.001 {
		return trimZeros(strconv.FormatFloat(v, 'f', 6, 64))
	}
	s := strconv.FormatFloat(v, 'g', 3, 64)
	if strings.ContainsAny(s, "e") {
		f, _ := strconv.ParseFloat(s, 64)
		s = strconv.FormatFloat(f, 'f', -1, 64)
	}
	return s
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
// Small rates keep three significant figures (0.004, 0.711).
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
