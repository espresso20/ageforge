package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// The Workbench drawing: header, queue, detail pane and the lines at the
// foot. Pure: a planScreen in, text out.

// planWideMin is the width from which the detail sits right of the queue;
// below it, under.
const planWideMin = 100

// planScreen is everything the panel draws from.
type planScreen struct {
	state       game.GameState
	sel         int
	note        string
	noteGood    bool
	clearArmed  bool
	width, rows int
	est         game.PlanEstimate
	haveEst     bool
	opt         game.PlanOptimizeResult
	haveOpt     bool
	mode        int
	input       string
	picks       []planPick
	pickSel     int
}

// padVis pads s with spaces to w visible cells (tags do not count).
func padVis(s string, w int) string {
	if n := visibleLen(s); n < w {
		return s + strings.Repeat(" ", w-n)
	}
	return s
}

// planSaving is how much sooner optimize would finish the plan, 0 when it
// would save under a minute or nothing would change.
func (sc planScreen) planSaving() (saved bool, d string) {
	if !sc.haveOpt || !sc.opt.Changed || sc.opt.Before-sc.opt.After < 60e9 {
		return false, ""
	}
	return true, textfmt.Duration(sc.opt.Before - sc.opt.After)
}

// planHeader is "Build plan" on the left and the estimate on the right.
func planHeader(sc planScreen) string {
	left := theme.Paint(theme.RoleAccent, vibe.Line("ui.overlay_plan.planPanelText.01", "═══ Build plan ═══"))
	if len(sc.state.Plan) == 0 {
		return left
	}
	var right string
	if sc.haveEst {
		right = vibe.Line("ui.overlay_plan.planHeader.01", "Plan finishes in ") + planTimeText(sc.est.Finish)
		if sc.est.Finish <= 0 && sc.est.Untimed > 0 {
			right = vibe.Line("ui.overlay_plan.planHeader.04", "Plan has no finish time yet")
		} else if sc.est.Untimed > 0 {
			right += vibe.Sprintf("ui.overlay_plan.planHeader.03", " (%d not timed)", sc.est.Untimed)
		}
		if ok, d := sc.planSaving(); ok {
			right += " · " + theme.Paint(theme.RoleHighlight, "O") + vibe.Line("ui.overlay_plan.planHeader.02", " optimize could save ") + d
		}
	}
	if gap := sc.width - visibleLen(left) - visibleLen(right); gap > 1 {
		return left + strings.Repeat(" ", gap) + theme.Paint(theme.RoleDim, right)
	}
	return left + "  " + theme.Paint(theme.RoleDim, right)
}

// planPanelScreen renders the Workbench.
func planPanelScreen(sc planScreen) string {
	var out []string
	out = append(out, planHeader(sc), "")
	// Rows left for the body, after the header pair and the foot pair.
	room := max(4, sc.rows-4)
	if sc.mode == planModeOptimize {
		out = append(out, planOptimizePopup(sc, room)...)
	} else if len(sc.state.Plan) == 0 {
		out = append(out, planEmptyLines()...)
	} else {
		out = append(out, planBody(sc, room)...)
	}
	out = append(out, "")
	out = append(out, planFootLines(sc)...)
	return strings.Join(out, "\n") + "\n"
}

// planEmptyLines says what an empty plan is and how to fill it.
func planEmptyLines() []string {
	return []string{
		strings.TrimRight(vibe.Line("ui.overlay_plan.planPanelText.03", " The plan is empty.\n\n"), "\n"),
		"",
		" " + theme.Keycap("A") + vibe.Line("ui.overlay_plan.planEmptyLines.01", " add an item by name, or type a command:"),
		"",
		" " + theme.Paint(theme.RoleLabel, vibe.Line("ui.plan_screen.planEmptyLines.01", "plan build <building> [count[]")) + theme.Paint(theme.RoleDim, vibe.Line("ui.plan_screen.planEmptyLines.02", "   e.g. plan build hut 10")),
		" " + theme.Paint(theme.RoleLabel, "plan research <tech>") + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_plan.planPanelText.06", "            with what it needs first, in order")),
		" " + theme.Paint(theme.RoleLabel, "plan expedition <mission>") + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_plan.planEmptyLines.02", "      a mission from the board")),
	}
}

// planBody is the queue and the detail, side by side when wide, stacked when
// not. room is the rows they have.
func planBody(sc planScreen, room int) []string {
	n := len(sc.state.Plan)
	if sc.width >= planWideMin {
		leftW := min(46, sc.width*2/5)
		rightW := sc.width - leftW - 3
		queue := planQueueLines(sc, leftW, room-1)
		detail := planDetailLines(sc, rightW)
		sep := theme.Paint(theme.RoleDim, " │ ")
		var out []string
		for i := range max(len(queue), min(len(detail), room-1)) {
			l, r := "", ""
			if i < len(queue) {
				l = queue[i]
			}
			if i < len(detail) {
				r = detail[i]
			}
			out = append(out, padVis(l, leftW)+sep+r)
		}
		return append(out, planCountLine(sc, n))
	}
	qRows := max(3, min(n, (room-2)/2))
	queue := planQueueLines(sc, sc.width, qRows)
	detail := planDetailLines(sc, sc.width)
	out := append(queue, theme.Paint(theme.RoleDim, strings.Repeat("─", min(sc.width, 60))))
	left := max(1, room-len(out)-1)
	if len(detail) > left {
		detail = detail[:left]
	}
	out = append(out, detail...)
	return append(out, planCountLine(sc, n))
}

// planCountLine is "12 of 60 items, and 3 techs queued".
func planCountLine(sc planScreen, n int) string {
	load := game.PlanLoad(sc.state.Plan)
	count := vibe.Sprintf("ui.overlay_plan.planPanelText.07", "%d of %d items", load, game.MaxPlanItems)
	if techs := n - load; techs > 0 {
		count += vibe.Sprintf("ui.overlay_plan.planPanelText.08", ", and %s queued (research takes no room)", textfmt.Count(techs, vibe.Line("ui.overlay_plan.planPanelText.09", "tech"), vibe.Line("ui.overlay_plan.planPanelText.10", "techs")))
	}
	return " " + theme.Paint(theme.RoleDim, count)
}

// planQueueLines is the queue: one line per item, scrolled so the selection
// shows, in rows lines at width cells.
func planQueueLines(sc planScreen, width, rows int) []string {
	n := len(sc.state.Plan)
	rows = max(1, rows)
	top := 0
	if n > rows {
		top = max(0, min(sc.sel-rows/2, n-rows))
	}
	const whenW, statusW = 11, 7
	titleW := max(8, width-2-4-1-statusW-1-whenW)
	var out []string
	for i := top; i < min(n, top+rows); i++ {
		v := sc.state.Plan[i]
		var est *game.PlanItemEstimate
		if sc.haveEst && i < len(sc.est.Items) {
			est = &sc.est.Items[i]
		}
		marker := "  "
		title := padVis(truncate(fmt.Sprintf("%2d. %s", i+1, planItemTitle(v)), titleW+4), titleW+4)
		when := planWhen(v, est)
		if i == sc.sel {
			marker = theme.Paint(theme.RoleAccent, "▸ ")
			title = theme.Selected(tview.Escape(title))
		} else {
			title = theme.Paint(theme.RoleText, tview.Escape(title))
		}
		out = append(out, marker+title+" "+padVis(planStatusText(v), statusW)+" "+theme.Paint(theme.RoleDim, truncate(when, whenW)))
	}
	return out
}

// planDetailLines is the selected item's pane at width cells.
func planDetailLines(sc planScreen, width int) []string {
	v := sc.state.Plan[sc.sel]
	var est *game.PlanItemEstimate
	if sc.haveEst && sc.sel < len(sc.est.Items) {
		est = &sc.est.Items[sc.sel]
	}
	var out []string
	head := theme.Paint(theme.RoleAccent, tview.Escape(planItemTitle(v))) + "  " + planStatusText(v)
	if when := planWhen(v, est); when != "" {
		head += theme.Paint(theme.RoleDim, "  ·  "+vibe.Line("ui.overlay_plan.planDetailLines.01", "starts ")+when)
	}
	out = append(out, head, "")
	if v.Status == game.PlanStatusBlocked && v.Note != "" {
		for _, l := range wrapWords(v.Note, width) {
			out = append(out, theme.Paint(theme.RoleWarning, tview.Escape(l)))
		}
		out = append(out, "")
	}
	out = append(out, planRowsTable(v, width)...)
	out = append(out, planHoldingLine(sc, width)...)
	out = append(out, planKindLines(sc, v, width)...)
	return out
}

// planRowsTable is the resource table: needs, have, held above, income and
// ready in, with the slowest row marked.
func planRowsTable(v game.PlanItemView, width int) []string {
	if len(v.Rows) == 0 {
		return nil
	}
	nameW := 8
	for _, r := range v.Rows {
		nameW = max(nameW, visibleLen(game.ResourceName(r.Res)))
	}
	nameW++
	full := width >= nameW+52
	hdr := padVis(vibe.Line("ui.overlay_plan.planRowsTable.01", "Resource"), nameW) + padVis(vibe.Line("ui.overlay_plan.planRowsTable.02", "Needs"), 8) + padVis(vibe.Line("ui.overlay_plan.planRowsTable.03", "Have"), 8)
	if full {
		hdr += padVis(vibe.Line("ui.overlay_plan.planRowsTable.04", "Held"), 8) + padVis(vibe.Line("ui.overlay_plan.planRowsTable.05", "Income"), 10)
	}
	hdr += vibe.Line("ui.overlay_plan.planRowsTable.06", "Ready in")
	out := []string{theme.Paint(theme.RoleDim, hdr)}
	for _, r := range v.Rows {
		line := padVis(game.ResourceName(r.Res), nameW) + padVis(textfmt.Number(r.Need), 8) + padVis(textfmt.Number(r.Have), 8)
		if full {
			held := "-"
			if r.Held >= 1 {
				held = textfmt.Number(r.Held)
			}
			income := vibe.Line("ui.overlay_plan.planRowsTable.07", "none")
			if r.PerSecond > 0 {
				income = "+" + textfmt.Number(r.PerSecond) + "/s"
			}
			line += padVis(held, 8) + padVis(income, 10)
		}
		switch {
		case r.Never:
			line += theme.Paint(theme.RoleNegative, txtPlanNever.String())
		case r.ReadyIn <= 0:
			line += theme.Paint(theme.RolePositive, txtPlanNow.String())
		default:
			line += textfmt.Duration(r.ReadyIn)
		}
		if r.Res == v.Bottleneck && v.Bottleneck != "" {
			line += theme.Paint(theme.RoleWarning, " ◂ "+vibe.Line("ui.overlay_plan.planRowsTable.08", "slowest"))
		}
		out = append(out, line)
	}
	return append(out, "")
}

// planHoldingLine is "Holding back: 180 stone from items 5 to 7".
func planHoldingLine(sc planScreen, width int) []string {
	v := sc.state.Plan[sc.sel]
	if len(v.HeldBack) == 0 || v.Status != game.PlanStatusWaiting {
		return nil
	}
	var parts []string
	for _, res := range sortedKeysOf(v.HeldBack) {
		var below []int
		for j := sc.sel + 1; j < len(sc.state.Plan); j++ {
			for _, r := range sc.state.Plan[j].Rows {
				if r.Res == res && r.Held > 0 {
					below = append(below, j+1)
					break
				}
			}
		}
		part := textfmt.Number(v.HeldBack[res]) + " " + game.ResourceName(res)
		if len(below) > 0 {
			part += vibe.Line("ui.overlay_plan.planHoldingLine.01", " from ") + planItemRange(below)
		}
		parts = append(parts, part)
	}
	var out []string
	for _, l := range wrapWords(vibe.Line("ui.overlay_plan.planHoldingLine.02", "Holding back: ")+strings.Join(parts, "; "), width) {
		out = append(out, theme.Paint(theme.RoleDim, tview.Escape(l)))
	}
	return append(out, "")
}

// planItemRange is "item 5", "items 5 and 6", "items 5 to 7" or "items 2, 5, 9"
// for the 1-based numbers in n (ascending).
func planItemRange(n []int) string {
	switch {
	case len(n) == 1:
		return fmt.Sprintf(vibe.Line("ui.overlay_plan.planItemRange.01", "item %d"), n[0])
	case n[len(n)-1]-n[0] == len(n)-1 && len(n) > 2:
		return vibe.Sprintf("ui.overlay_plan.planItemRange.02", "items %d to %d", n[0], n[len(n)-1])
	case len(n) == 2:
		return vibe.Sprintf("ui.overlay_plan.planItemRange.03", "items %d and %d", n[0], n[1])
	}
	s := make([]string, len(n))
	for i, x := range n {
		s[i] = fmt.Sprint(x)
	}
	return vibe.Line("ui.overlay_plan.planItemRange.04", "items ") + strings.Join(s, ", ")
}

// planKindLines are the lines only some items have: an advance's pillars, a
// wonder's bank, what holds the research slot, an expedition's mission.
func planKindLines(sc planScreen, v game.PlanItemView, width int) []string {
	var out []string
	switch {
	case v.Kind == game.PlanAdvance && len(v.Pillars) > 0:
		for _, p := range v.Pillars {
			line := padVis(truncate(p.Name, 12), 13) + wonderProgressBar(float64(p.Percent)/100, 10) + fmt.Sprintf(" %3d%%", p.Percent)
			switch {
			case p.Done:
				line += theme.Paint(theme.RolePositive, " "+vibe.Line("ui.overlay_plan.planKindLines.01", "done"))
			case p.Blocker != "":
				line += theme.Paint(theme.RoleDim, " "+tview.Escape(p.Blocker))
				if p.Item > 0 {
					line += theme.Paint(theme.RoleAccent, vibe.Sprintf("ui.overlay_plan.planKindLines.02", "  (item %d)", p.Item))
				} else {
					line += theme.Paint(theme.RoleWarning, vibe.Line("ui.overlay_plan.planKindLines.03", "  (not in the plan)"))
				}
			}
			out = append(out, line)
		}
	case v.When == game.PlanWhenHand:
		out = append(out, vibe.Line("ui.overlay_plan.planKindLines.04", "Bank ")+wonderProgressBar(v.BankShare, 14)+fmt.Sprintf(" %3.0f%%", v.BankShare*100))
		out = append(out, theme.Paint(theme.RoleWarning, vibe.Line("ui.overlay_plan.planKindLines.05", "Deposit by hand: wonder bank")))
	case v.Kind == game.PlanResearch:
		if v.Holds != "" {
			line := vibe.Line("ui.overlay_plan.planKindLines.06", "The research slot holds ") + tview.Escape(v.Holds)
			if v.Frees > 0 {
				line += vibe.Line("ui.overlay_plan.planKindLines.07", "; it frees in ") + textfmt.Duration(v.Frees)
			}
			out = append(out, line)
		}
	case v.Kind == game.PlanExpedition:
		out = append(out, vibe.Line("ui.overlay_plan.planKindLines.08", "Mission: ")+tview.Escape(v.Name))
		if len(v.Cost) > 0 {
			out = append(out, vibe.Line("ui.overlay_plan.planKindLines.09", "Price: ")+strings.ReplaceAll(planPriceWhole(v.Cost), glue, " "))
		}
		slot := vibe.Line("ui.overlay_plan.planKindLines.10", "Party: at home, free to go")
		if v.Status == game.PlanStatusBlocked && v.Note != "" {
			slot = vibe.Line("ui.overlay_plan.planKindLines.11", "Party: ") + tview.Escape(v.Note)
		}
		out = append(out, slot)
	}
	if len(out) > 0 {
		// A note under the lines, for a status the table above does not say.
		if v.Kind != game.PlanAdvance && v.Status != game.PlanStatusBlocked && v.Note != "" && v.When != game.PlanWhenHand {
			out = append(out, theme.Paint(theme.RoleDim, tview.Escape(v.Note)))
		}
	}
	if v.Kind == game.PlanBuild || v.Kind == game.PlanUpgrade {
		line := vibe.Sprintf("ui.overlay_plan.planKindLines.12", "Copies left: %d", v.Count)
		if v.Started > 0 {
			line += vibe.Sprintf("ui.overlay_plan.planKindLines.13", " (%d started)", v.Started)
		}
		out = append(out, theme.Paint(theme.RoleDim, line))
	}
	return out
}

// planFootLines are the lines above the footer keys: the add box with its
// matches, or the last action's note.
func planFootLines(sc planScreen) []string {
	switch sc.mode {
	case planModeAdd:
		return planAddLines(sc)
	case planModePick:
		return planPickLines(sc)
	}
	if sc.note == "" {
		return []string{"", ""}
	}
	role := theme.RoleNegative
	if sc.noteGood {
		role = theme.RolePositive
	} else if sc.clearArmed {
		role = theme.RoleWarning
	}
	return []string{" " + theme.Paint(role, tview.Escape(sc.note)), ""}
}
