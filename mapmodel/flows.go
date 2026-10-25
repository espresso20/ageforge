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

// Activity is what drives the maps' ambient life: traffic, crowds,
// patrols. Every field comes from real state, so a busier economy is a
// busier picture.
type Activity struct {
	Routes    int // trade routes running
	Disrupted int // routes blockaded
	Staffed   int // workers at work
	Idle      int
	Soldiers  int
	Wars      int // civs at war with you
	// Wealth is 0..1: how built-up and trading the settlement is.
	Wealth float64
	// Traffic is 0..1: the overall density of moving things, from routes,
	// labour and wealth together.
	Traffic float64
}

func activityFor(m *Model) Activity {
	a := Activity{Staffed: m.Workers.Staffed, Idle: m.Workers.Idle, Soldiers: m.Army.Soldiers}
	for _, r := range m.Routes {
		if r.Disrupted {
			a.Disrupted++
		} else {
			a.Routes++
		}
	}
	for _, f := range m.Factions {
		if f.Relation == RelWar {
			a.Wars++
		}
	}
	trade := 0
	if l := m.Lineage(LinTrade); l != nil {
		trade = l.Count
	}
	total := m.TotalBuildings()
	built := clamp01(Log2(float64(1+total)) / 10)
	a.Wealth = clamp01(float64(0.7*built) + float64(0.3*clamp01(float64(trade)/20)))
	labour := clamp01(Log2(float64(1+a.Staffed)) / 12)
	routes := clamp01(float64(a.Routes) / 4)
	a.Traffic = clamp01(float64(0.35*routes) + float64(0.35*labour) + float64(0.3*a.Wealth))
	return a
}
