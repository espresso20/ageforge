package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// WorkerPanel shows a compact workers summary grouped by building domain.
type WorkerPanel struct {
	root     *tview.TextView
	lastHash uint64
}

// NewWorkerPanel creates the worker info panel
func NewWorkerPanel() *WorkerPanel {
	vp := &WorkerPanel{}
	vp.root = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	vp.root.SetBorder(true).SetTitle(" Workers ")
	theme.Track(func() { vp.root.SetTitleColor(theme.Color(theme.RoleAccent)) })
	return vp
}

// Primitive returns the underlying tview primitive
func (vp *WorkerPanel) Primitive() tview.Primitive {
	return vp.root
}

// UpdateState refreshes the worker panel with workers grouped by building domain.
func (vp *WorkerPanel) UpdateState(state game.GameState) {
	// Hash-based dirty check
	var h uint64
	h = hashKey(state.Age)
	h ^= uint64(state.Workers.TotalPop+1) * 31
	h ^= uint64(state.Workers.FoodDrain*1000) * 37
	h ^= uint64(state.Morale*1000) * 41
	h ^= uint64(state.MoraleMultiplier*1000) * 43
	wt := state.Workers.Types["worker"]
	h ^= uint64(wt.Count+1) * 7
	h ^= uint64(wt.IdleCount+1) * 3
	for _, bs := range state.Buildings {
		if bs.WorkersAssigned > 0 {
			h ^= uint64(bs.WorkersAssigned+1) * 13
		}
	}
	if h == vp.lastHash {
		return
	}
	vp.lastHash = h

	renderVillagerPanel(vp.root, &state)
}

// panelDomainLabels maps domain key to display label for the worker panel.
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

// computeMoraleBand turns morale% + multiplier into colours and copy.
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
			Status:     fmt.Sprintf("▲ Morale %.0f%% — production %s", moralePct*100, label),
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
			Status:     fmt.Sprintf("▼ Morale %.0f%% — production %s", moralePct*100, label),
			Penalty:    true,
			DeltaPct:   delta,
			DeltaLabel: label,
		}
	default:
		return moraleBand{
			Color:    "white",
			Status:   fmt.Sprintf("Morale %.0f%% — steady", moralePct*100),
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

func renderVillagerPanel(tv *tview.TextView, state *game.GameState) {
	var sb strings.Builder

	wt := state.Workers.Types["worker"]
	total := state.Workers.TotalPop
	maxPop := state.Workers.MaxPop
	foodDrain := state.Workers.FoodDrain

	// Morale bar (always shown) — recoloured by band (green high / neutral mid /
	// red low) off the engine's banded multiplier, not the old 0.50/0.80 cutoffs.
	band := computeMoraleBand(state.Morale, state.MoraleMultiplier)
	moraleBar := moraleBandBar(int(state.Morale*20), 20, 20, band.Color)
	capStr := ""
	if state.MoraleCap > 1.0 {
		capStr = fmt.Sprintf(" [gray](cap: %.2f)[-]", state.MoraleCap)
	}
	fmt.Fprintf(&sb, "[white]Morale:[white] [%s]%.0f%%[-]%s  %s\n\n",
		band.Color, state.Morale*100, capStr, moraleBar)

	if !wt.Unlocked || total == 0 {
		fmt.Fprintf(&sb, "[white]Workers  [yellow]0[white] / [green]%d[white]\n\n", maxPop)
		fmt.Fprint(&sb, "[gray]No workers yet.\n")
		fmt.Fprint(&sb, "Recruit with: [cyan]recruit [count|max][-]\n")
		tv.SetText(sb.String())
		return
	}

	idle := wt.IdleCount
	fmt.Fprintf(&sb, "[white]Workers  [yellow]%d[white] / [green]%d[white]   Idle: [cyan]%d[white]   Food: [red]%.2f/tick[white]\n\n",
		total, maxPop, idle, foodDrain)

	// Morale band line in summary section — bonus (green), steady (neutral), or
	// penalty (red). No "penalty" text unless we're actually in the low band.
	if band.Bonus || band.Penalty {
		fmt.Fprintf(&sb, "  [%s]%s[-]\n\n", band.Color, band.Status)
	}

	// Build domain groups from buildings with assigned workers.
	type domainGroup struct {
		Domain string
		Rows   []buildingRow
		Total  int
	}
	groupMap := map[string]*domainGroup{}
	groupOrder := []string{}

	for bKey, bs := range state.Buildings {
		if bs.WorkersAssigned <= 0 || bs.WorkerDomain == "" {
			continue
		}
		domain := bs.WorkerDomain
		if _, exists := groupMap[domain]; !exists {
			groupMap[domain] = &domainGroup{Domain: domain}
			groupOrder = append(groupOrder, domain)
		}
		// Total capacity = WorkerCapacity per instance × number of built instances.
		count := bs.Count
		if count <= 0 {
			count = 1
		}
		cap := bs.WorkerCapacity * count
		groupMap[domain].Rows = append(groupMap[domain].Rows, buildingRow{
			Key:             bKey,
			Name:            bs.Name,
			WorkersAssigned: bs.WorkersAssigned,
			Capacity:        cap,
		})
		groupMap[domain].Total += bs.WorkersAssigned
	}

	// Sort domain groups alphabetically for stable display.
	sort.Strings(groupOrder)
	// Sort buildings within each domain group alphabetically by key.
	for _, grp := range groupMap {
		sort.Slice(grp.Rows, func(i, j int) bool {
			return grp.Rows[i].Key < grp.Rows[j].Key
		})
	}

	if len(groupOrder) == 0 {
		fmt.Fprintf(&sb, "[cyan]Idle: %d[white] — assign with: [cyan]assign <building> [count|all][-]\n", idle)
	} else {
		for _, domain := range groupOrder {
			grp := groupMap[domain]
			label, ok2 := panelDomainLabels[domain]
			if !ok2 {
				label = capitalize(domain)
			}
			// Class name for this domain at current age.
			classDisplay := ""
			if cls, found := config.WorkerClassByDomainAndAge(domain, state.Age); found && cls.ClassName != "" {
				classDisplay = fmt.Sprintf("  %s × %d", cls.ClassName, grp.Total)
			}
			fmt.Fprintf(&sb, "[cyan][%s][-][white]%s\n", label, classDisplay)
			for _, row := range grp.Rows {
				bar := assignBar(row.WorkersAssigned, row.Capacity, 8)
				fmt.Fprintf(&sb, "  %-20s %s [cyan]%d[white]/[green]%d[white]\n",
					row.Name, bar, row.WorkersAssigned, row.Capacity)
			}
			fmt.Fprintln(&sb)
		}
		if idle > 0 {
			fmt.Fprintf(&sb, "[cyan]Idle: %d[white] — assign with: [cyan]assign <building> [count|all][-]\n", idle)
		}
	}

	tv.SetText(sb.String())
}

// moraleBandBar renders a width-w bar whose fill uses the given tview color
// name (the morale band colour) rather than the fixed assignment fill colour.
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
