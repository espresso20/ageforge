package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
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
	vp.root.SetBorder(true).SetTitle(" Workers ").SetTitleColor(ColorVillager)
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
	Name            string
	WorkersAssigned int
	Capacity        int
}

func renderVillagerPanel(tv *tview.TextView, state *game.GameState) {
	var sb strings.Builder

	wt := state.Workers.Types["worker"]
	total := state.Workers.TotalPop
	maxPop := state.Workers.MaxPop
	foodDrain := state.Workers.FoodDrain

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

	// Build domain groups from buildings with assigned workers.
	type domainGroup struct {
		Domain string
		Rows   []buildingRow
		Total  int
	}
	groupMap := map[string]*domainGroup{}
	groupOrder := []string{}

	for _, bs := range state.Buildings {
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
			Name:            bs.Name,
			WorkersAssigned: bs.WorkersAssigned,
			Capacity:        cap,
		})
		groupMap[domain].Total += bs.WorkersAssigned
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
			fmt.Fprintf(&sb, "[green][%s][-][white]%s\n", label, classDisplay)
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

// assignBar renders a small tview-colored progress bar of width w.
func assignBar(filled, total, w int) string {
	if total <= 0 {
		return "[gray]" + strings.Repeat("░", w) + "[-]"
	}
	f := (filled * w) / total
	if f > w {
		f = w
	}
	if f < 0 {
		f = 0
	}
	return "[green]" + strings.Repeat("█", f) + "[gray]" + strings.Repeat("░", w-f) + "[-]"
}

