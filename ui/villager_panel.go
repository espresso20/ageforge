package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// WorkerPanel shows a compact workers-by-domain summary.
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

// domainLabel returns a capitalized display label for a domain key.
func domainLabel(domain string) string {
	if domain == "" {
		return ""
	}
	return strings.ToUpper(domain[:1]) + domain[1:]
}

// UpdateState refreshes the worker panel with a compact domain summary
func (vp *WorkerPanel) UpdateState(state game.GameState) {
	// Hash-based dirty check
	var h uint64
	h = hashKey(state.Age)
	h ^= uint64(state.Workers.TotalPop+1) * 31
	h ^= uint64(state.Workers.FoodDrain*1000) * 37
	for key, vt := range state.Workers.Types {
		if vt.Count > 0 {
			h ^= hashKey(key) * 13
			h ^= uint64(vt.Count+1) * 7
			h ^= uint64(vt.IdleCount+1) * 3
		}
	}
	if h == vp.lastHash {
		return
	}
	vp.lastHash = h

	var sb strings.Builder

	v := state.Workers
	fmt.Fprintf(&sb, " [gold]Total:[-] %d/%d  [gold]Idle:[-] %d  [gold]Food:[-] %.1f/tick\n",
		v.TotalPop, v.MaxPop, v.TotalIdle, v.FoodDrain)
	sb.WriteString("\n [gold]Workers by Domain[-]\n")
	sb.WriteString(" [gray]─────────────────────[-]\n")

	// Fixed domain order for deterministic display.
	domainOrder := []string{"food", "faith", "knowledge", "military", "trade", "engineering", "hacker", "astronaut"}

	anyDomain := false
	for _, domain := range domainOrder {
		vt, ok := v.Types[domain]
		if !ok || !vt.Unlocked || vt.Count == 0 {
			continue
		}
		anyDomain = true

		label := domainToLabel[domain]
		if label == "" {
			label = domainLabel(domain)
		}

		// Calculate assigned = total - idle
		assigned := vt.Count - vt.IdleCount
		if assigned < 0 {
			assigned = 0
		}

		// Determine class name from config for display.
		className := ""
		if wcd, found := config.WorkerClassByDomainAndAge(domain, state.Age); found {
			className = wcd.ClassName
		}

		if className != "" {
			fmt.Fprintf(&sb, " %-14s [white]%d[-] / [gray]%d[-]  [gray](%s)[-]\n", label, assigned, vt.Count, className)
		} else {
			fmt.Fprintf(&sb, " %-14s [white]%d[-] / [gray]%d[-]\n", label, assigned, vt.Count)
		}
	}

	if !anyDomain {
		sb.WriteString(" [gray](no workers yet)[-]\n")
	}

	vp.root.SetText(sb.String())
}
