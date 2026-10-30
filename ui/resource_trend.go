package ui

import (
	"math"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
)

// resource_trend.go is what the Resources box draws after an ordinary
// resource's amount and rate: a bar of how big the stock is and a small line
// of where it has been lately. The engine keeps no per-resource history, so
// the Economy tab keeps its own short one, here, outside the game state and
// the save.

const (
	resTrendCells  = 8                 // cells of trend line
	trendEvery     = time.Minute       // one sample this often
	trendSamples   = resTrendCells - 1 // kept samples; the live amount is the last point
	magBarFloorTop = 10.0              // the bar's top is never under this
)

// resHistory is the amounts of each resource at the last few sample times.
type resHistory struct {
	at   time.Time
	seen int // the game tick of the last snapshot, to spot a different game
	amts map[string][]float64
}

// note takes a snapshot in: at most one sample every trendEvery of wall
// time, and none carried over from a different game (the tick went back).
func (h *resHistory) note(state game.GameState, now time.Time) {
	if h.amts == nil || state.Tick < h.seen {
		h.amts, h.at = map[string][]float64{}, time.Time{}
	}
	h.seen = state.Tick
	if !h.at.IsZero() && now.Sub(h.at) < trendEvery {
		return
	}
	h.at = now
	for key, rs := range state.Resources {
		if !rs.Unlocked {
			continue
		}
		s := append(h.amts[key], rs.Amount)
		if len(s) > trendSamples {
			s = s[len(s)-trendSamples:]
		}
		h.amts[key] = s
	}
}

// points is the line's values for key: its samples, then the amount now.
func (h *resHistory) points(key string, now float64) []float64 {
	if h == nil {
		return nil
	}
	return append(append([]float64(nil), h.amts[key]...), now)
}

// magnitudeTop is the top of the magnitude bar for a set of amounts: the
// power of ten at or above the largest, and at least 10.
func magnitudeTop(amounts []float64) float64 {
	top := 0.0
	for _, a := range amounts {
		top = max(top, a)
	}
	if top <= magBarFloorTop {
		return magBarFloorTop
	}
	return math.Pow(10, math.Ceil(math.Log10(top)))
}

// magnitudeFill is how far along the bar an amount is, 0 to 1, on a log
// scale up to top.
func magnitudeFill(amount, top float64) float64 {
	if amount <= 0 || top <= 0 {
		return 0
	}
	return min(1, math.Log10(amount+1)/math.Log10(top+1))
}

var (
	eighths    = []rune("▏▎▍▌▋▊▉")
	trendRamp  = []rune("▁▂▃▄▅▆▇█")
	trendASCII = []rune("_.-~^")
)

// magnitudeCells is a bar of width cells, ratio of it filled. The Unicode
// bar has eighth-block partial cells; the plain one has whole cells of # and
// . only.
func magnitudeCells(ratio float64, width int, plain bool) string {
	ratio = min(1, max(0, ratio))
	if plain {
		n := int(math.Round(ratio * float64(width)))
		return BarFillColor() + strings.Repeat("#", n) + BarEmptyColor() + strings.Repeat(".", width-n) + "[-]"
	}
	eight := int(ratio * float64(width) * 8)
	whole, part := eight/8, eight%8
	var b strings.Builder
	b.WriteString(BarFillColor() + strings.Repeat("█", whole))
	used := whole
	if part > 0 {
		b.WriteRune(eighths[part-1])
		used++
	}
	b.WriteString(BarEmptyColor() + strings.Repeat("░", width-used) + "[-]")
	return b.String()
}

// trendCells is the line of vals in resTrendCells cells, the newest at the
// right, scaled between its own lowest and highest. Fewer than two points
// leave it blank; a flat run sits at the middle.
func trendCells(vals []float64, plain bool, color string) string {
	if len(vals) > resTrendCells {
		vals = vals[len(vals)-resTrendCells:]
	}
	if len(vals) < 2 {
		return strings.Repeat(" ", resTrendCells)
	}
	ramp := trendRamp
	if plain {
		ramp = trendASCII
	}
	lo, hi := vals[0], vals[0]
	for _, v := range vals {
		lo, hi = min(lo, v), max(hi, v)
	}
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", resTrendCells-len(vals)))
	b.WriteString("[" + color + "]")
	for _, v := range vals {
		i := len(ramp) / 2
		if hi > lo {
			i = int(math.Round((v - lo) / (hi - lo) * float64(len(ramp)-1)))
		}
		b.WriteRune(ramp[i])
	}
	b.WriteString("[-]")
	return b.String()
}
