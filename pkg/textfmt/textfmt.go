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

// units are the suffixes Number prints, smallest first, each a thousand of
// the one before: thousand, million, billion, trillion, quadrillion, and
// from there quintillion, sextillion, septillion, octillion, nonillion and
// decillion. The largest store in the game is tens of quintillions, so the
// table runs five units past what an amount in play needs.
var units = []struct {
	threshold float64
	suffix    string
}{
	{1e3, "K"},
	{1e6, "M"},
	{1e9, "B"},
	{1e12, "T"},
	{1e15, "Q"},
	{1e18, "Qi"},
	{1e21, "Sx"},
	{1e24, "Sp"},
	{1e27, "Oc"},
	{1e30, "No"},
	{1e33, "Dc"},
}

// MaxNumber is the largest amount Number is sure to print with at most three
// digits before its unit: 999 of the last unit. Past it there is no unit
// left to move to, and the digits pile up in front of the last one
// (1000Dc, 25000Dc). Nothing a player can hold or be quoted should reach it.
func MaxNumber() float64 {
	return 999 * units[len(units)-1].threshold
}

// Number formats an amount or rate for display: three significant figures,
// trailing zeros dropped, and a unit from a thousand up, K M B T Q and then
// Qi Sx Sp Oc No Dc (950, 12.5, 0.711, 1.5K, 1.23M, 876M, 11.8Qi). Up to
// MaxNumber it never prints more than three digits before the unit. Rounding
// goes through strconv, so the output is identical on every architecture
// (config descriptions built with it are part of the determinism
// fingerprint).
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
	i := 0
	for i < len(units)-1 && abs >= units[i+1].threshold {
		i++
	}
	s := sig3(abs / units[i].threshold)
	if s == "1000" && i < len(units)-1 {
		return prefix + "1" + units[i+1].suffix // 999.6K rolls over to 1M
	}
	return prefix + s + units[i].suffix
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
