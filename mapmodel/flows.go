package mapmodel

import (
	"cmp"
	"maps"
	"slices"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// flows.go is the "flows" summary the maps can overlay: where the economy is
// stuck. Draining stocks are about to run dry, understaffed buildings and idle workers are labour in the
// wrong place, and the per-lineage output share says what the settlement
// actually runs on.

// Store is one resource's stock.
type Store struct {
	Key, Name string
	Rate      float64 // per second
	// EmptyIn is the seconds until a draining stock runs dry (0 when not
	// draining).
	EmptyIn float64
}

// Flows is the bottleneck summary.
type Flows struct {
	Draining     []Store // falling, dry within about a minute
	Understaffed []*Building
	IdleWorkers  int
	Idle         []*Building // take workers and have none
	// Shares is every lineage's share of working output, largest first.
	Shares []LineageShare
	// Worst is the one-line verdict ("iron runs dry · 12 idle workers"),
	// "" when nothing is stuck.
	Worst string
}

// LineageShare is a lineage's share of output.
type LineageShare struct {
	Lineage string
	Share   float64
}

func flowsFor(m *Model, st *game.GameState) Flows {
	f := Flows{IdleWorkers: m.Workers.Idle}
	perSec := 1 / game.BaseTickInterval.Seconds()
	for _, k := range slices.Sorted(maps.Keys(st.Resources)) {
		r := st.Resources[k]
		if !r.Unlocked {
			continue
		}
		if r.Rate < 0 && r.Amount > 0 {
			s := Store{Key: k, Name: r.Name, Rate: float64(r.Rate * perSec)}
			s.EmptyIn = r.Amount / -s.Rate
			if s.EmptyIn < 60 {
				f.Draining = append(f.Draining, s)
			}
		}
	}
	for _, b := range m.Buildings {
		if b.Count == 0 || b.Staffing < 0 {
			continue
		}
		if b.Workers == 0 {
			f.Idle = append(f.Idle, b)
		} else if b.Understaffed() {
			f.Understaffed = append(f.Understaffed, b)
		}
	}
	for _, l := range m.Lineages {
		f.Shares = append(f.Shares, LineageShare{l.Key, l.Share})
	}
	slices.SortStableFunc(f.Shares, func(a, b LineageShare) int { return cmp.Compare(b.Share, a.Share) })
	var worst []string
	if len(f.Draining) > 0 {
		worst = append(worst, f.Draining[0].Name+vibe.Line("mapmodel.flows.flowsFor.01", " runs dry"))
	}
	if f.IdleWorkers > 0 {
		worst = append(worst, plural(f.IdleWorkers, vibe.Line("mapmodel.flows.flowsFor.02", "idle worker"), vibe.Line("mapmodel.flows.flowsFor.03", "idle workers")))
	}
	if n := len(f.Understaffed) + len(f.Idle); n > 0 {
		worst = append(worst, plural(n, vibe.Line("mapmodel.flows.flowsFor.04", "building short of hands"), vibe.Line("mapmodel.flows.flowsFor.05", "buildings short of hands")))
	}
	for i, w := range worst {
		if i == 0 {
			f.Worst = w
		} else {
			f.Worst += " · " + w
		}
	}
	return f
}
