package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// The Plan panel: the build plan in order, what each item's next start costs
// and whether it could start now, with keys to reorder and remove items. A
// text overlay with a key handler; the handler runs on the tview goroutine and
// never under the engine lock, so it may call the engine's plan methods
// (write lock) and GetState (read lock). Items are added with commands
// (`plan build`, `plan research`), which autocomplete.

// planPanel holds the panel's UI-only state. Owned by the tview goroutine.
type planPanel struct {
	// sel is the selected item, 0-based; clamped to the plan on every render.
	sel int
	// note is the last action's feedback line ("" for none).
	note     string
	noteGood bool
	// clearArmed is set by the first C press; the second one clears.
	clearArmed bool
}

// reset clears the feedback and the Clear confirmation, for a fresh open.
func (p *planPanel) reset() {
	p.note, p.noteGood, p.clearArmed = "", false, false
}

// provider renders the panel for OverlayManager.
func (p *planPanel) provider(state game.GameState, _ int) string {
	p.sel = clampSel(p.sel, len(state.Plan))
	return planPanelText(state, p.sel, p.note, p.noteGood, p.clearArmed)
}

func clampSel(sel, n int) int {
	if n == 0 || sel < 0 {
		return 0
	}
	if sel >= n {
		return n - 1
	}
	return sel
}

// handleKey acts on event against engine. Reports whether the key was
// consumed. Must not be called under the engine lock.
//
//	↑/↓ or k/j   select
//	u / d        move the selected item up / down
//	x or Delete  remove the selected item
//	C, C         clear the plan (press twice)
func (p *planPanel) handleKey(event *tcell.EventKey, engine *game.GameEngine) bool {
	n := len(engine.GetState().Plan)
	r := unicode.ToLower(event.Rune())
	isRune := event.Key() == tcell.KeyRune
	arm := false
	switch {
	case event.Key() == tcell.KeyUp || (isRune && r == 'k'):
		p.sel = clampSel(p.sel-1, n)
	case event.Key() == tcell.KeyDown || (isRune && r == 'j'):
		p.sel = clampSel(p.sel+1, n)
	case isRune && (r == 'u' || r == 'd'):
		if n == 0 {
			p.note, p.noteGood = "The plan is empty.", false
			break
		}
		delta := -1
		if r == 'd' {
			delta = 1
		}
		to, err := engine.PlanMove(p.sel+1, delta)
		if err != nil {
			p.note, p.noteGood = err.Error(), false
			break
		}
		p.sel = to - 1
		p.note = ""
	case event.Key() == tcell.KeyDelete || (isRune && r == 'x'):
		if n == 0 {
			p.note, p.noteGood = "The plan is empty.", false
			break
		}
		what, err := engine.PlanRemove(p.sel + 1)
		if err != nil {
			p.note, p.noteGood = err.Error(), false
			break
		}
		p.note, p.noteGood = "Removed "+what+".", true
		p.sel = clampSel(p.sel, n-1)
	case isRune && event.Rune() == 'C':
		if n == 0 {
			p.note, p.noteGood = "The plan is empty.", false
			break
		}
		if !p.clearArmed {
			arm = true
			p.note, p.noteGood = "Press C again to clear the whole plan.", false
			break
		}
		engine.PlanClear()
		p.sel = 0
		p.note, p.noteGood = "Plan cleared.", true
	default:
		return false
	}
	p.clearArmed = arm
	return true
}

// planStatusText is an item's status word in its role color.
func planStatusText(v game.PlanItemView) string {
	switch v.Status {
	case game.PlanStatusReady:
		return theme.Paint(theme.RolePositive, "ready  ")
	case game.PlanStatusWaiting:
		return theme.Paint(theme.RoleHighlight, "waiting")
	default:
		return theme.Paint(theme.RoleWarning, "blocked")
	}
}

// planItemTitle is "Hut ×3 (2 started)", "research Tool Making" or
// "trade food for wood: 500 wood still to buy, 300 bought".
func planItemTitle(v game.PlanItemView) string {
	switch v.Kind {
	case game.PlanResearch:
		return "research " + v.Name
	case game.PlanAdvance:
		return v.Name
	case game.PlanTrade:
		s := "trade " + game.ResourceName(v.Key) + " for " + game.ResourceName(v.To)
		if v.Amount > 0 {
			s += ": " + game.Amount(v.Amount, v.To) + " still to buy"
		} else {
			s += ": keeps " + game.ResourceName(v.To) + " topped up"
		}
		if v.Got > 0 {
			s += ", " + FormatNumber(v.Got) + " bought"
		}
		return s
	}
	s := v.Name
	if v.Count > 1 {
		s += fmt.Sprintf(" ×%d", v.Count)
	}
	if v.Started > 0 {
		s += fmt.Sprintf(" (%d started)", v.Started)
	}
	return s
}

// planItemDetail is the second part of an item's line: the progress bar and
// what it is short of, the reason it is blocked, or what it will pay. A build
// item's overflow bank shows wherever it holds something.
func planItemDetail(v game.PlanItemView, state game.GameState) string {
	switch v.Status {
	case game.PlanStatusBlocked:
		note := v.Note
		if len(v.Banked) > 0 {
			note += " (" + formatPlanCost(v.Banked) + " banked from overflow)"
		}
		return theme.Paint(theme.RoleDim, note)
	case game.PlanStatusWaiting:
		detail := wonderProgressBar(v.Progress, 8) + fmt.Sprintf(" %3.0f%%", v.Progress*100)
		if v.Short != "" {
			cost := v.Cost[v.Short]
			covered := v.Progress * cost
			banked := v.Banked[v.Short]
			line := fmt.Sprintf("  %s %s / %s", game.ResourceName(v.Short), FormatNumber(covered), FormatNumber(cost))
			var notes []string
			if banked >= 1 {
				notes = append(notes, FormatNumber(banked)+" banked")
			}
			if held := state.Resources[v.Short].Amount - (covered - banked); held >= 1 {
				notes = append(notes, FormatNumber(held)+" held for items above")
			}
			if len(notes) > 0 {
				line += " (" + strings.Join(notes, ", ") + ")"
			}
			detail += theme.Paint(theme.RoleDim, line)
		}
		return detail
	default:
		due := planDue(v)
		switch {
		case len(v.Cost) == 0:
			return theme.Paint(theme.RoleDim, "starts next tick")
		case len(due) == 0:
			return theme.Paint(theme.RoleDim, "starts next tick, paid from its bank")
		case len(v.Banked) > 0:
			return theme.Paint(theme.RoleDim, "starts next tick for "+formatPlanCost(due)+" and its bank")
		}
		return theme.Paint(theme.RoleDim, "starts next tick for "+formatPlanCost(v.Cost))
	}
}

// planDue is what an item's next start takes from the stores: its price less
// what its bank covers (parts under 1 left out).
func planDue(v game.PlanItemView) map[string]float64 {
	if len(v.Banked) == 0 {
		return v.Cost
	}
	due := map[string]float64{}
	for res, c := range v.Cost {
		if d := c - v.Banked[res]; d >= 1 {
			due[res] = d
		}
	}
	return due
}

// formatPlanCost is "48 wood, 20 stone", resources in key order.
func formatPlanCost(cost map[string]float64) string {
	var parts []string
	for _, k := range sortedMapKeys(cost) {
		parts = append(parts, game.Amount(cost[k], k))
	}
	return strings.Join(parts, ", ")
}

func sortedMapKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// planPanelText renders the panel. Pure: state in, text out.
func planPanelText(state game.GameState, sel int, note string, noteGood, clearArmed bool) string {
	var sb strings.Builder
	sb.WriteString(theme.Paint(theme.RoleAccent, "═══ Build plan ═══") + "\n")
	sb.WriteString(theme.Paint(theme.RoleDim, " Started in order as resources come in, while you play and while you are away.\n Each item is paid when it starts. A waiting item holds its price back from the\n items below it; items below may still start with what it doesn't need.\n Overflow from full storage is banked toward items' next copies, in order.") + "\n\n")

	if len(state.Plan) == 0 {
		sb.WriteString(" The plan is empty.\n\n")
		sb.WriteString(" " + theme.Paint(theme.RoleLabel, "plan build <building> [count[]") + theme.Paint(theme.RoleDim, "   e.g. plan build hut 10") + "\n")
		sb.WriteString(" " + theme.Paint(theme.RoleLabel, "plan research <tech>") + theme.Paint(theme.RoleDim, "            with what it needs first, in order") + "\n")
	} else {
		for i, v := range state.Plan {
			marker := "  "
			title := fmt.Sprintf("%2d. %s", i+1, planItemTitle(v))
			if i == sel {
				marker = theme.Paint(theme.RoleAccent, "▸ ")
				title = theme.Selected(tview.Escape(title))
			} else {
				title = theme.Paint(theme.RoleText, tview.Escape(title))
			}
			fmt.Fprintf(&sb, " %s%s\n       %s  %s\n", marker, title, planStatusText(v), planItemDetail(v, state))
		}
		fmt.Fprintf(&sb, "\n %s\n", theme.Paint(theme.RoleDim, fmt.Sprintf("%d of %d items", len(state.Plan), game.MaxPlanItems)))
	}

	sb.WriteString("\n " + theme.Keycap("↑↓") + " select  " + theme.Keycap("U") + " up  " + theme.Keycap("D") + " down  " +
		theme.Keycap("X") + " remove  " + theme.Keycap("C") + " clear\n")
	sb.WriteString(" " + theme.Paint(theme.RoleDim, "Add with plan build / plan research; plan list, remove, up, down and clear work as commands too.") + "\n")

	if note != "" {
		role := theme.RoleNegative
		if noteGood {
			role = theme.RolePositive
		} else if clearArmed {
			role = theme.RoleWarning
		}
		sb.WriteString("\n " + theme.Paint(role, tview.Escape(note)) + "\n")
	}
	return sb.String()
}

// planListText is `plan list`: the plan as plain log lines.
func planListText(state game.GameState) string {
	if len(state.Plan) == 0 {
		return "The plan is empty. Add items with: plan build <building> [count[], or plan research <tech>."
	}
	var sb strings.Builder
	sb.WriteString(theme.Paint(theme.RoleAccent, "Build plan") + "\n")
	for i, v := range state.Plan {
		status := v.Status
		switch v.Status {
		case game.PlanStatusWaiting:
			status = fmt.Sprintf("waiting, %.0f%%", v.Progress*100)
			if v.Short != "" {
				status += " (short of " + game.ResourceName(v.Short) + ")"
			}
		case game.PlanStatusBlocked:
			status = "blocked: " + v.Note
		}
		if len(v.Banked) > 0 {
			status += ", " + formatPlanCost(v.Banked) + " banked"
		}
		fmt.Fprintf(&sb, "  %d. %s  %s\n", i+1, tview.Escape(planItemTitle(v)), theme.Paint(theme.RoleDim, status))
	}
	return strings.TrimRight(sb.String(), "\n")
}
