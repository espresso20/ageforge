package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
)

// ProgressBar returns a text-based progress bar with colored segments.
// Filled portion uses BarFillColor (Accent role), empty uses BarEmptyColor (Dim role).
func ProgressBar(current, max float64, width int) string {
	if max <= 0 {
		return BarEmptyColor() + strings.Repeat("░", width) + "[-]"
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
	return BarFillColor() + strings.Repeat("█", filled) + BarEmptyColor() + strings.Repeat("░", empty) + "[-]"
}

// FormatCost formats a cost map as "50 food, 30 wood" (display names, sorted
// by key so the order is stable). An empty cost is "free".
func FormatCost(cost map[string]float64) string {
	if len(cost) == 0 {
		return "free"
	}
	return game.Amounts(cost)
}
