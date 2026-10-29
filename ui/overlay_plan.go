package ui

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// The Plan panel (the Workbench): the build plan as a queue on the left and
// the selected item's detail on the right, with an optimize preview and an
// add box. A text overlay with a key handler; the handler runs on the tview
// goroutine and never under the engine lock, so it may call the engine's plan
// methods (write lock) and GetState (read lock). The drawing is plan_screen.go,
// the add box plan_add.go and the preview popup plan_optimize.go.

// Modes of the panel: the queue, or one of its three overlays on it.
const (
	planModeQueue = iota
	planModeAdd   // typing the name of an item to add
	planModePick  // the name fits several items; pick one
	planModeOptimize
)

// planPanel holds the panel's UI-only state. Owned by the tview goroutine.
type planPanel struct {
	// sel is the selected item, 0-based; clamped to the plan on every render.
	sel int
	// note is the last action's feedback line ("" for none).
	note     string
	noteGood bool
	// clearArmed is set by the first C press; the second one clears.
	clearArmed bool

	mode int
	// input is the name typed into the add box; picks are the items it can
	// mean when several, pickSel the one chosen.
	input   string
	picks   []planPick
	pickSel int

	// engine is read for the estimate and the optimize preview; the panel
	// draws without them when it is nil. height says how many rows the screen
	// has (40 when nil).
	engine *game.GameEngine
	height func() int
}

// reset clears the feedback, the Clear confirmation and any overlay, for a
// fresh open.
func (p *planPanel) reset() {
	p.note, p.noteGood, p.clearArmed = "", false, false
	p.mode, p.input, p.picks, p.pickSel = planModeQueue, "", nil, 0
}

// provider renders the panel for OverlayManager.
func (p *planPanel) provider(state game.GameState, screenW int) string {
	p.sel = clampSel(p.sel, len(state.Plan))
	sc := planScreen{
		state: state, sel: p.sel, note: p.note, noteGood: p.noteGood, clearArmed: p.clearArmed,
		width: overlayTextWidth(screenW), rows: p.bodyRows(),
		mode: p.mode, input: p.input, picks: p.picks, pickSel: p.pickSel,
	}
	if p.engine != nil {
		sc.est = p.engine.PlanEstimate()
		sc.haveEst = len(sc.est.Items) == len(state.Plan)
		if prev, err := p.engine.PlanOptimizePreview(); err == nil {
			sc.opt, sc.haveOpt = prev, true
		}
	}
	return planPanelScreen(sc)
}

// bodyRows is the rows the panel's text may use.
func (p *planPanel) bodyRows() int {
	h := 40
	if p.height != nil {
		h = p.height()
	}
	return max(12, h-panelChromeRows-promptRows)
}

func clampSel(sel, n int) int {
	return max(0, min(sel, n-1))
}

// handleKey acts on event against engine. Reports whether the key was
// consumed. Must not be called under the engine lock.
//
//	↑/↓ or k/j   select; PgUp/PgDn page
//	a            add an item (a name box; Enter adds below the selection)
//	+ / -        the count of a build or upgrade item
//	u / d        move the selected item up / down
//	x or Delete  remove the selected item
//	o            preview plan optimize (Enter keeps it)
//	C, C         clear the plan (press twice)
func (p *planPanel) handleKey(event *tcell.EventKey, engine *game.GameEngine) bool {
	p.engine = engine
	switch p.mode {
	case planModeAdd, planModePick:
		return p.handleAddKey(event, engine)
	case planModeOptimize:
		return p.handleOptimizeKey(event, engine)
	}
	n := len(engine.GetState().Plan)
	r := unicode.ToLower(event.Rune())
	isRune := event.Key() == tcell.KeyRune
	arm := false
	switch {
	case event.Key() == tcell.KeyUp || (isRune && r == 'k'):
		p.sel = clampSel(p.sel-1, n)
	case event.Key() == tcell.KeyDown || (isRune && r == 'j'):
		p.sel = clampSel(p.sel+1, n)
	case event.Key() == tcell.KeyPgUp:
		p.sel = clampSel(p.sel-planPageRows(p.bodyRows()), n)
	case event.Key() == tcell.KeyPgDn:
		p.sel = clampSel(p.sel+planPageRows(p.bodyRows()), n)
	case isRune && r == 'a':
		p.mode, p.input, p.picks, p.pickSel = planModeAdd, "", nil, 0
		p.note = ""
	case isRune && r == 'o':
		p.mode = planModeOptimize
		p.note = ""
	case isRune && (event.Rune() == '+' || event.Rune() == '=' || event.Rune() == '-' || event.Rune() == '_' || event.Rune() == '−'):
		if n == 0 {
			p.note, p.noteGood = vibe.Line("ui.overlay_plan.handleKey.07", "The plan is empty."), false
			break
		}
		delta := 1
		if event.Rune() == '-' || event.Rune() == '_' || event.Rune() == '−' {
			delta = -1
		}
		count, err := engine.PlanAdjust(p.sel+1, delta)
		if err != nil {
			p.note, p.noteGood = err.Error(), false
			break
		}
		p.note, p.noteGood = vibe.Sprintf("ui.overlay_plan.handleKey.08", "Count is now %d.", count), true
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
		p.note, p.noteGood = txtRemoved.String()+what+".", true
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

// planPageRows is how far PgUp and PgDn move the selection: a queue page.
func planPageRows(bodyRows int) int {
	return max(1, bodyRows-6)
}

// planFooter is the keys on the panel's foot: the one place they are listed.
func planFooter() [][2]string {
	return [][2]string{{"A", "add"}, {"+/−", "count"}, {"U/D", "move"}, {"X", "remove"}, {"O", "optimize"}, {"C", "clear"}, {"Esc", "close"}}
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

// planPriceWhole is "48 wood, 20 stone", resources in key order, with each
// material's spaces glued.
func planPriceWhole(cost map[string]float64) string {
	var parts []string
	for _, k := range sortedKeysOf(cost) {
		parts = append(parts, strings.ReplaceAll(game.Amount(cost[k], k), " ", glue))
	}
	return strings.Join(parts, ", ")
}

// txtPlanNever and txtPlanNow are said in several places, so each has one id.
var (
	txtPlanNever = vibe.E("ui.overlay_plan.planWhen.02", "never")
	txtPlanNow   = vibe.E("ui.overlay_plan.planWhen.03", "now")
)

// planWhen is when an item starts, as a word: "now", "in 4m", "never" or
// "needs you". Empty when nothing can be said. est is the plan estimate (nil
// without one): the walk's figure wins over the item's own price when it
// played the item.
func planWhen(v game.PlanItemView, est *game.PlanItemEstimate) string {
	switch v.When {
	case game.PlanWhenHand:
		return vibe.Line("ui.overlay_plan.planWhen.01", "needs you")
	case game.PlanWhenNever:
		return txtPlanNever.String()
	case game.PlanWhenUnknown:
		return ""
	}
	at := v.StartsIn
	if v.Status == game.PlanStatusReady {
		at = 0
	} else if est != nil && est.Modeled {
		if est.Never {
			return txtPlanNever.String()
		}
		at = est.Start
	}
	if at <= 0 {
		return txtPlanNow.String()
	}
	return vibe.Line("ui.overlay_plan.planWhen.04", "in ") + textfmt.Duration(at)
}

// planTimeText is a span for the header: textfmt.Duration, "now" for none.
func planTimeText(d time.Duration) string {
	if d <= 0 {
		return txtPlanNow.String()
	}
	return textfmt.Duration(d)
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
		if when := planWhen(v, nil); when != "" && v.Status != game.PlanStatusBlocked {
			status += "  ·  " + vibe.Line("ui.overlay_plan.planListText.06", "starts ") + when
		}
		fmt.Fprintf(&sb, "  %d. %s  %s\n", i+1, tview.Escape(planItemTitle(v)), theme.Paint(theme.RoleDim, status))
	}
	return strings.TrimRight(sb.String(), "\n")
}
