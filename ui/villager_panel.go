package ui

import (
	"fmt"
	"strings"
)

// villager_panel.go holds the morale and assignment helpers the Workers panel
// (overlay_workers.go) and the status bar share. The old standalone worker
// panel that lived here was unreachable and has been removed.

// panelDomainLabels maps domain key to display label for the Workers panel.
var panelDomainLabels = map[string]string{
	"food":        "Food",
	"lumber":      "Lumber",
	"masonry":     "Masonry",
	"knowledge":   "Knowledge",
	"faith":       "Faith",
	"military":    "Military",
	"trade":       "Trade",
	"engineering": "Engineering",
	"metallurgy":  "Metallurgy",
	"energy":      "Energy",
	"hacker":      "Hacker",
	"astronaut":   "Astronaut",
}

// buildingRow holds one row of the assignment display.
type buildingRow struct {
	Key             string
	Name            string
	WorkersAssigned int
	Capacity        int
}

// moraleBand describes how to render the banded morale state for the UI.
// It is derived purely from the live morale fraction and the production
// multiplier the engine's continuous curve produced (state.MoraleMultiplier),
// so the UI never re-derives the curve formula itself.
type moraleBand struct {
	Color      string // tview color for morale text + bar fill
	Status     string // short player-facing status line, e.g. "production +18%"
	Bonus      bool   // multiplier > 1.0 (above pivot)
	Penalty    bool   // multiplier < 1.0 (below pivot)
	DeltaPct   int    // signed production delta in percent, rounded
	DeltaLabel string // pre-formatted signed percent ("+18%", "-34%", "+<1%", "-<1%"); "" when steady
}

// computeMoraleBand turns morale% + multiplier into colors and copy.
//   - mult > 1.0 → green "▲ … production +N%"
//   - mult == 1.0 → neutral "… steady" (no penalty text)
//   - mult < 1.0 → red "▼ … production −N%"
//
// Because the morale curve is now continuous, a real bonus/penalty can round to
// 0%. Rather than printing a dishonest "+0%"/"-0%", DeltaLabel reads "+<1%" /
// "-<1%" in that case so the player sees the effect exists but is tiny.
//
// moralePct is the raw morale fraction (e.g. 0.52); mult is state.MoraleMultiplier.
func computeMoraleBand(moralePct, mult float64) moraleBand {
	// Round the production delta off the multiplier, not the raw morale.
	delta := int(roundHalf((mult - 1.0) * 100))
	switch {
	case mult > 1.0:
		label := fmt.Sprintf("+%d%%", delta)
		if delta == 0 {
			label = "+<1%"
		}
		return moraleBand{
			Color:      "green",
			Status:     fmt.Sprintf("▲ Morale %.0f%% (production %s)", moralePct*100, label),
			Bonus:      true,
			DeltaPct:   delta,
			DeltaLabel: label,
		}
	case mult < 1.0:
		// delta is negative here; %d already carries the ASCII minus sign.
		label := fmt.Sprintf("%d%%", delta)
		if delta == 0 {
			label = "-<1%"
		}
		return moraleBand{
			Color:      "red",
			Status:     fmt.Sprintf("▼ Morale %.0f%% (production %s)", moralePct*100, label),
			Penalty:    true,
			DeltaPct:   delta,
			DeltaLabel: label,
		}
	default:
		return moraleBand{
			Color:    "white",
			Status:   fmt.Sprintf("Morale %.0f%% (production steady)", moralePct*100),
			DeltaPct: 0,
		}
	}
}

// roundHalf rounds to nearest, halves away from zero (math.Round semantics
// without pulling math in for a single call site at -0.0 edge cases).
func roundHalf(f float64) float64 {
	if f < 0 {
		return float64(int(f - 0.5))
	}
	return float64(int(f + 0.5))
}

// moraleBandBar renders a width-w bar whose fill uses the given tview color
// name (the morale band color) rather than the fixed assignment fill color.
func moraleBandBar(filled, total, w int, fillColor string) string {
	if total <= 0 {
		return BarEmptyColor() + strings.Repeat("░", w) + "[-]"
	}
	f := (filled * w) / total
	if f > w {
		f = w
	}
	if f < 0 {
		f = 0
	}
	return "[" + fillColor + "]" + strings.Repeat("█", f) + BarEmptyColor() + strings.Repeat("░", w-f) + "[-]"
}

// assignBar renders a small tview-colored progress bar of width w.
func assignBar(filled, total, w int) string {
	if total <= 0 {
		return BarEmptyColor() + strings.Repeat("░", w) + "[-]"
	}
	f := (filled * w) / total
	if f > w {
		f = w
	}
	if f < 0 {
		f = 0
	}
	return BarFillColor() + strings.Repeat("█", f) + BarEmptyColor() + strings.Repeat("░", w-f) + "[-]"
}
