package ui

import (
	"fmt"
	"sort"
	"strings"
)

// ProgressBar returns a text-based progress bar
func ProgressBar(current, max float64, width int) string {
	if max <= 0 {
		return strings.Repeat("░", width)
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
	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// FormatNumber formats a number with 1 decimal if not whole
func FormatNumber(n float64) string {
	if n == float64(int(n)) {
		return fmt.Sprintf("%.0f", n)
	}
	return fmt.Sprintf("%.1f", n)
}

// FormatRate formats a rate with sign
func FormatRate(rate float64) string {
	if rate == 0 {
		return "[gray]+0.0[-]"
	}
	if rate > 0 {
		return fmt.Sprintf("[green]+%.1f[-]", rate)
	}
	return fmt.Sprintf("[red]%.1f[-]", rate)
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
		parts = append(parts, fmt.Sprintf("%s:%.0f", k, cost[k]))
	}
	return strings.Join(parts, " ")
}

// Pad right-pads a string to a minimum width
func Pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
