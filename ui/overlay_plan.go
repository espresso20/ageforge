package ui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
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
func (p *planPanel) provider(state game.GameState, screenW int) string {
	p.sel = clampSel(p.sel, len(state.Plan))
	return planPanelText(state, p.sel, p.note, p.noteGood, p.clearArmed, overlayTextWidth(screenW))
}

func clampSel(sel, n int) int {
	return max(0, min(sel, n-1))
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
			p.note, p.noteGood = vibe.Line("ui.overlay_plan.handleKey.01", "The plan is empty."), false
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
			p.note, p.noteGood = vibe.Line("ui.overlay_plan.handleKey.02", "The plan is empty."), false
			break
		}
		what, err := engine.PlanRemove(p.sel + 1)
		if err != nil {
			p.note, p.noteGood = err.Error(), false
			break
		}
		p.note, p.noteGood = vibe.Line("ui.overlay_plan.handleKey.03", "Removed ")+what+".", true
		p.sel = clampSel(p.sel, n-1)
	case isRune && event.Rune() == 'C':
		if n == 0 {
			p.note, p.noteGood = vibe.Line("ui.overlay_plan.handleKey.04", "The plan is empty."), false
			break
		}
		if !p.clearArmed {
			arm = true
			p.note, p.noteGood = vibe.Line("ui.overlay_plan.handleKey.05", "Press C again to clear the whole plan."), false
			break
		}
		engine.PlanClear()
		p.sel = 0
		p.note, p.noteGood = vibe.Line("ui.overlay_plan.handleKey.06", "Plan cleared."), true
	default:
		return false
	}
	p.clearArmed = arm
	return true
}

// planFooter is the keys on the panel's foot.
func planFooter() [][2]string {
	return [][2]string{{"↑↓", "pick"}, {"U D", "move"}, {"X", "remove"}, {"C", "clear"}, {"Esc", "close"}}
}

// planStatusText is an item's status word in its role color.
func planStatusText(v game.PlanItemView) string {
	switch v.Status {
	case game.PlanStatusReady:
		return theme.Paint(theme.RolePositive, padRight(vibe.Line("ui.overlay_plan.planStatusText.01", "ready"), 7))
	case game.PlanStatusWaiting:
		return theme.Paint(theme.RoleHighlight, "waiting")
	default:
		return theme.Paint(theme.RoleWarning, "blocked")
	}
}

// planItemTitle is "Hut ×3 (2 started)" or "research Tool Making".
func planItemTitle(v game.PlanItemView) string {
	switch v.Kind {
	case game.PlanResearch:
		return vibe.Line("ui.overlay_plan.planItemTitle.01", "research ") + v.Name
	case game.PlanAdvance:
		return v.Name
	}
	if v.Kind == game.PlanUpgrade {
		s := vibe.Sprintf("ui.overlay_plan.planItemTitle.02", "upgrade %d %s → %s", v.Count, v.Name, v.Target)
		if v.Started > 0 {
			s += vibe.Sprintf("ui.overlay_plan.planItemTitle.03", " (%d started)", v.Started)
		}
		return s
	}
	s := v.Name
	if v.Count > 1 {
		s += fmt.Sprintf(" ×%d", v.Count)
	}
	if v.Started > 0 {
		s += vibe.Sprintf("ui.overlay_plan.planItemTitle.04", " (%d started)", v.Started)
	}
	return s
}

// planItemDetailGlued is planItemDetail with the spaces inside each material
// of a price glued, so planDetailRows can wrap between materials.
func planItemDetailGlued(v game.PlanItemView, state game.GameState) string {
	switch v.Status {
	case game.PlanStatusBlocked:
		return theme.Paint(theme.RoleDim, v.Note)
	case game.PlanStatusWaiting:
		detail := wonderProgressBar(v.Progress, 8) + fmt.Sprintf(" %3.0f%%", v.Progress*100)
		if v.Short != "" {
			cost := v.Cost[v.Short]
			covered := v.Progress * cost
			line := fmt.Sprintf("  %s %s / %s", game.ResourceName(v.Short), textfmt.Number(covered), textfmt.Number(cost))
			var notes []string
			if held := state.Resources[v.Short].Amount - covered; held >= 1 {
				notes = append(notes, textfmt.Number(held)+vibe.Line("ui.overlay_plan.planItemDetailGlued.03", " held for items above"))
			}
			if len(notes) > 0 {
				line += " (" + strings.Join(notes, ", ") + ")"
			}
			detail += theme.Paint(theme.RoleDim, line)
		}
		return detail
	default:
		if len(v.Cost) == 0 {
			return theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_plan.planItemDetailGlued.04", "starts next tick"))
		}
		return theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_plan.planItemDetailGlued.08", "starts next tick for ")+planPriceWhole(v.Cost))
	}
}

// planDetailLead is the cells before an item's detail: the seven that indent
// the line, the status word (seven) and two spaces.
const planDetailLead = 16

// planDetailRows is detail on one row when it fits width cells, otherwise
// wrapped at its spaces onto further rows under the first. The materials of a
// price are joined with glue, so a material ("49.8T dark matter crystals")
// is not broken while it fits a row of its own.
func planDetailRows(detail string, width int) string {
	if visibleLen(strings.ReplaceAll(detail, glue, " ")) <= width {
		return strings.ReplaceAll(detail, glue, " ")
	}
	return strings.Join(wrapWords(detail, width), "\n"+strings.Repeat(" ", planDetailLead))
}

// planPriceWhole is "48 wood, 20 stone", resources in key order, with each
// material's spaces glued.
func planPriceWhole(cost map[string]float64) string {
	var parts []string
	for _, k := range sortedKeysOf(cost) {
		parts = append(parts, strings.ReplaceAll(game.Amount(cost[k], k), " ", glue))
	}
	return strings.Join(parts, ", ")
}

// planPanelText renders the panel. Pure: state in, text out.
// width is the cells the text has; a detail that is longer than the room left
// of the status word goes on over further lines under its first.
func planPanelText(state game.GameState, sel int, note string, noteGood, clearArmed bool, width int) string {
	var sb strings.Builder
	sb.WriteString(theme.Paint(theme.RoleAccent, vibe.Line("ui.overlay_plan.planPanelText.01", "═══ Build plan ═══")) + "\n")
	sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_plan.planPanelText.02", " Started in order as resources come in, while you play and while you are away.\n Each item is paid when it starts. A waiting item holds its price back from the\n items below it; items below may still start with what it doesn't need.")) + "\n\n")

	if len(state.Plan) == 0 {
		sb.WriteString(vibe.Line("ui.overlay_plan.planPanelText.03", " The plan is empty.\n\n"))
		sb.WriteString(" " + theme.Paint(theme.RoleLabel, "plan build <building> [count[]") + theme.Paint(theme.RoleDim, "   e.g. plan build hut 10") + "\n")
		sb.WriteString(" " + theme.Paint(theme.RoleLabel, "plan research <tech>") + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_plan.planPanelText.06", "            with what it needs first, in order")) + "\n")
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
			fmt.Fprintf(&sb, " %s%s\n       %s  %s\n", marker, title, planStatusText(v), planDetailRows(planItemDetailGlued(v, state), width-planDetailLead))
		}
		// Research takes no room: the 60 are for everything else.
		load := game.PlanLoad(state.Plan)
		count := vibe.Sprintf("ui.overlay_plan.planPanelText.07", "%d of %d items", load, game.MaxPlanItems)
		if techs := len(state.Plan) - load; techs > 0 {
			count += vibe.Sprintf("ui.overlay_plan.planPanelText.08", ", and %s queued (research takes no room)", textfmt.Count(techs, vibe.Line("ui.overlay_plan.planPanelText.09", "tech"), vibe.Line("ui.overlay_plan.planPanelText.10", "techs")))
		}
		fmt.Fprintf(&sb, "\n %s\n", theme.Paint(theme.RoleDim, count))
	}

	sb.WriteString("\n " + theme.Keycap("↑↓") + vibe.Line("ui.overlay_plan.planPanelText.11", " select  ") + theme.Keycap("U") + " up  " + theme.Keycap("D") + " down  " +
		theme.Keycap("X") + vibe.Line("ui.overlay_plan.planPanelText.14", " remove  ") + theme.Keycap("C") + vibe.Line("ui.overlay_plan.planPanelText.15", " clear\n"))
	sb.WriteString(" " + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_plan.planPanelText.16", "Add with plan build / plan upgrade / plan research; plan list, remove, up, down and clear work as commands too.")) + "\n")

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
		return vibe.Line("ui.overlay_plan.planListText.01", "The plan is empty. Add items with: plan build <building> [count[], or plan research <tech>.")
	}
	var sb strings.Builder
	sb.WriteString(theme.Paint(theme.RoleAccent, vibe.Line("ui.overlay_plan.planListText.02", "Build plan")) + "\n")
	for i, v := range state.Plan {
		status := v.Status
		switch v.Status {
		case game.PlanStatusWaiting:
			status = vibe.Sprintf("ui.overlay_plan.planListText.03", "waiting, %.0f%%", v.Progress*100)
			if v.Short != "" {
				status += vibe.Line("ui.overlay_plan.planListText.04", " (short of ") + game.ResourceName(v.Short) + ")"
			}
		case game.PlanStatusBlocked:
			status = vibe.Line("ui.overlay_plan.planListText.05", "blocked: ") + v.Note
		}
		fmt.Fprintf(&sb, "  %d. %s  %s\n", i+1, tview.Escape(planItemTitle(v)), theme.Paint(theme.RoleDim, status))
	}
	return strings.TrimRight(sb.String(), "\n")
}
