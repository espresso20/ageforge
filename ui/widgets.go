package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ProgressBar returns a text-based progress bar with colored segments.
// Filled portion uses BarFillColor (purple), empty portion uses BarEmptyColor (dark gray).
func ProgressBar(current, max float64, width int) string {
	if max <= 0 {
		return "[" + BarEmptyColor + "]" + strings.Repeat("░", width) + "[-]"
	}
	ratio := current / max
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}
	filled := int(ratio * float64(width))
	empty := width - filled
	return "[" + BarFillColor + "]" + strings.Repeat("█", filled) + "[" + BarEmptyColor + "]" + strings.Repeat("░", empty) + "[-]"
}

// suffixes for large number formatting
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

// FormatNumber formats a number with suffix notation for large values (K/M/B/T/Q)
func FormatNumber(n float64) string {
	negative := n < 0
	abs := math.Abs(n)

	prefix := ""
	if negative {
		prefix = "-"
	}

	if abs < 1000 {
		if abs == math.Floor(abs) {
			return fmt.Sprintf("%s%.0f", prefix, abs)
		}
		return fmt.Sprintf("%s%.1f", prefix, abs)
	}

	for _, s := range suffixes {
		if abs >= s.threshold {
			scaled := abs / s.threshold
			var formatted string
			if scaled >= 100 {
				formatted = fmt.Sprintf("%.0f%s", scaled, s.suffix)
			} else if scaled >= 10 {
				formatted = fmt.Sprintf("%.1f%s", scaled, s.suffix)
			} else {
				formatted = fmt.Sprintf("%.2f%s", scaled, s.suffix)
			}
			return prefix + formatted
		}
	}

	return fmt.Sprintf("%s%.0f", prefix, abs)
}

// FormatRate formats a rate with sign and suffix notation.
// Uses higher precision for small rates so values like 0.02 show as +0.02
// rather than rounding to +0.0.
func FormatRate(rate float64) string {
	if rate == 0 {
		return "[gray]+0.0[-]"
	}
	abs := math.Abs(rate)
	sign := "+"
	color := "green"
	if rate < 0 {
		sign = ""
		color = "red"
	}
	// For small rates (< 1), use enough decimal places to show a non-zero digit
	if abs < 1 {
		// Find how many decimals we need so it doesn't round to zero
		prec := 2
		for prec < 6 {
			if math.Round(abs*math.Pow10(prec))/math.Pow10(prec) > 0 {
				break
			}
			prec++
		}
		return fmt.Sprintf("[%s]%s%.*f[-]", color, sign, prec, rate)
	}
	return fmt.Sprintf("[%s]%s%s[-]", color, sign, FormatNumber(rate))
}

// FormatCost formats a cost map as a string with stable ordering
func FormatCost(cost map[string]float64) string {
	if len(cost) == 0 {
		return "free"
	}
	keys := make([]string, 0, len(cost))
	for k := range cost {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, FormatNumber(cost[k])))
	}
	return strings.Join(parts, " ")
}

// FormatETA formats milliseconds into a human-readable duration string
func FormatETA(ms int) string {
	secs := ms / 1000
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	mins := secs / 60
	remainSecs := secs % 60
	if mins < 60 {
		return fmt.Sprintf("%dm%02ds", mins, remainSecs)
	}
	hours := mins / 60
	remainMins := mins % 60
	return fmt.Sprintf("%dh%02dm", hours, remainMins)
}

// Pad right-pads a string to a minimum width
func Pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
