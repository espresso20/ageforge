package ui

import (
	"fmt"
	"math"

	"github.com/espresso20/ageforge/vibe"
)

// villager_panel.go holds the morale band the status bar and the People panel
// share. The old standalone worker panel that lived here was unreachable and
// has been removed, and so has the Workers text panel the People panel replaced.

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
	delta := int(math.Round((mult - 1.0) * 100))
	switch {
	case mult > 1.0:
		label := fmt.Sprintf("+%d%%", delta)
		if delta == 0 {
			label = "+<1%"
		}
		return moraleBand{
			Color:      "green",
			Status:     vibe.Sprintf("ui.villager_panel.computeMoraleBand.01", "▲ Morale %.0f%% (production %s)", moralePct*100, label),
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
			Status:     vibe.Sprintf("ui.villager_panel.computeMoraleBand.02", "▼ Morale %.0f%% (production %s)", moralePct*100, label),
			Penalty:    true,
			DeltaPct:   delta,
			DeltaLabel: label,
		}
	default:
		return moraleBand{
			Color:    "white",
			Status:   vibe.Sprintf("ui.villager_panel.computeMoraleBand.03", "Morale %.0f%% (production steady)", moralePct*100),
			DeltaPct: 0,
		}
	}
}
