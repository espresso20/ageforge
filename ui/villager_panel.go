package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// WorkerPanel shows a compact workers summary.
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

// UpdateState refreshes the worker panel with a compact worker summary
func (vp *WorkerPanel) UpdateState(state game.GameState) {
	// Hash-based dirty check
	var h uint64
	h = hashKey(state.Age)
	h ^= uint64(state.Workers.TotalPop+1) * 31
	h ^= uint64(state.Workers.FoodDrain*1000) * 37
	wt := state.Workers.Types["worker"]
	h ^= uint64(wt.Count+1) * 7
	h ^= uint64(wt.IdleCount+1) * 3
	if h == vp.lastHash {
		return
	}
	vp.lastHash = h

	var sb strings.Builder

	v := state.Workers
	fmt.Fprintf(&sb, " [gold]Total:[-] %d/%d  [gold]Idle:[-] %d  [gold]Food:[-] %.1f/tick\n",
		v.TotalPop, v.MaxPop, v.TotalIdle, v.FoodDrain)

	if !wt.Unlocked || wt.Count == 0 {
		sb.WriteString("\n [gray](no workers yet)[-]\n")
		vp.root.SetText(sb.String())
		return
	}

	sb.WriteString("\n [gold]Workers[-]\n")
	sb.WriteString(" [gray]─────────────────────[-]\n")

	// Calculate assigned = total - idle
	assigned := wt.Count - wt.IdleCount
	if assigned < 0 {
		assigned = 0
	}

	// Determine class name from config for display (use "food" domain progression).
	className := wt.Name
	if wcd, found := config.WorkerClassByDomainAndAge("food", state.Age); found {
		className = wcd.ClassName
	}

	if className != "" {
		fmt.Fprintf(&sb, " %-14s [white]%d[-] / [gray]%d[-]  [gray](%s)[-]\n", "Workers", assigned, wt.Count, className)
	} else {
		fmt.Fprintf(&sb, " %-14s [white]%d[-] / [gray]%d[-]\n", "Workers", assigned, wt.Count)
	}

	vp.root.SetText(sb.String())
}
