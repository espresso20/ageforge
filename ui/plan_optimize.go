package ui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// planOptimizeShown is how many plan items the `plan optimize` reply lists.
const planOptimizeShown = 12

// planOptimizeText is the reply to `plan optimize`: what happened, the new
// order when there is one, and the plan's finish time before and after.
func planOptimizeText(res game.PlanOptimizeResult) string {
	if res.Reason == game.PlanOptimizeNothing {
		return vibe.Line("ui.plan_optimize.planOptimizeText.01", "Nothing to reorder: none of the buildings in the plan changes how the town grows.")
	}
	var sb strings.Builder
	if res.Changed {
		sb.WriteString(vibe.Line("ui.plan_optimize.planOptimizeText.02", "Reordered the plan for faster growth."))
		sb.WriteString(planOptimizeTimes(res))
		sb.WriteString("\n")
		for i, label := range res.Order {
			if i == planOptimizeShown {
				fmt.Fprint(&sb, vibe.Sprintf("ui.plan_optimize.planOptimizeText.03", "  and %d more", len(res.Order)-planOptimizeShown))
				break
			}
			fmt.Fprintf(&sb, "  %d. %s\n", i+1, tview.Escape(label))
		}
	} else {
		sb.WriteString(vibe.Line("ui.plan_optimize.planOptimizeText.04", "The plan is already in the fastest order I could find, so I left it alone."))
		sb.WriteString(planOptimizeTimes(res))
	}
	if res.Beyond > 0 {
		fmt.Fprint(&sb, vibe.Sprintf("ui.plan_optimize.planOptimizeText.05", "\nOnly the first %s copies were reordered; the other %s keep their order after them.", textfmt.Int(game.PlanOptimizeCopies), textfmt.Int(res.Beyond)))
	}
	if res.Waiting > 0 {
		fmt.Fprint(&sb, vibe.Sprintf("ui.plan_optimize.planOptimizeText.06", "\n%s cannot start until the town makes what they cost.", textfmt.Count(res.Waiting, txtCopy.String(), txtCopies.String())))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// planOptimizeTimes is the estimate sentence: when the last building of the
// plan starts, before and after. Empty when nothing has to be waited for.
func planOptimizeTimes(res game.PlanOptimizeResult) string {
	if res.Before <= 0 && res.After <= 0 {
		return ""
	}
	if !res.Changed {
		return vibe.Sprintf("ui.plan_optimize.planOptimizeTimes.01", " The last building should start in about %s.", textfmt.Duration(res.Before))
	}
	return vibe.Sprintf("ui.plan_optimize.planOptimizeTimes.02", " The last building should start in about %s instead of %s.", textfmt.Duration(res.After), textfmt.Duration(res.Before))
}

// handleOptimizeKey is the key handler while the optimize popup is open:
// Enter keeps the new order, Esc closes it, other keys do nothing.
func (p *planPanel) handleOptimizeKey(event *tcell.EventKey, engine *game.GameEngine) bool {
	switch event.Key() {
	case tcell.KeyEscape:
		p.mode = planModeQueue
	case tcell.KeyEnter:
		p.mode = planModeQueue
		prev, err := engine.PlanOptimizePreview()
		if err != nil || !prev.Changed {
			break
		}
		res, err := engine.PlanOptimize()
		if err != nil {
			p.note, p.noteGood = err.Error(), false
			break
		}
		p.sel = 0
		p.note, p.noteGood = vibe.Sprintf("ui.plan_optimize.handleOptimizeKey.01", "Kept the new order. The plan finishes in %s instead of %s.", planTimeText(res.After), planTimeText(res.Before)), true
	}
	return true
}

// planOptimizePopup is the preview: the plan as it stands and as optimize
// would leave it, side by side with the moved items marked, then the finish
// before and after and the time saved. room is the rows it may use.
func planOptimizePopup(sc planScreen, room int) []string {
	title := theme.Paint(theme.RoleAccent, vibe.Line("ui.plan_optimize.planOptimizePopup.01", "Optimize preview"))
	pad := strings.Repeat(" ", 2)
	if !sc.haveOpt || !sc.opt.Changed {
		why := vibe.Line("ui.plan_optimize.planOptimizePopup.02", "Nothing would change: the plan is already in the fastest order I could find.")
		if sc.haveOpt && sc.opt.Reason == game.PlanOptimizeNothing {
			why = vibe.Line("ui.plan_optimize.planOptimizePopup.03", "Nothing to reorder: none of the buildings in the plan changes how the town grows.")
		}
		out := []string{pad + title, "", pad + why}
		if sc.haveOpt && sc.opt.Before > 0 {
			out = append(out, "", pad+vibe.Line("ui.plan_optimize.planOptimizePopup.04", "Plan finishes in ")+planTimeText(sc.opt.Before))
		}
		return append(out, "", pad+theme.Paint(theme.RoleDim, vibe.Line("ui.plan_optimize.planOptimizePopup.05", "Esc closes")))
	}
	half := max(14, (sc.width-8)/2)
	shown := max(3, room-8)
	n := len(sc.opt.Order)
	out := []string{pad + title, ""}
	out = append(out, pad+padVis(theme.Paint(theme.RoleDim, vibe.Line("ui.plan_optimize.planOptimizePopup.06", "Now")), half)+theme.Paint(theme.RoleDim, vibe.Line("ui.plan_optimize.planOptimizePopup.07", "After")))
	for i := 0; i < min(n, shown); i++ {
		old := ""
		if i < len(sc.opt.Previous) {
			old = sc.opt.Previous[i]
		}
		mark := "  "
		right := tview.Escape(sc.opt.Order[i])
		if i < len(sc.opt.Moved) && sc.opt.Moved[i] {
			mark = theme.Paint(theme.RoleHighlight, "* ")
			right = theme.Paint(theme.RoleHighlight, right)
		}
		out = append(out, pad+padVis(fmt.Sprintf("%2d. %s", i+1, truncate(tview.Escape(old), half-6)), half)+mark+right)
	}
	if n > shown {
		out = append(out, pad+theme.Paint(theme.RoleDim, vibe.Sprintf("ui.plan_optimize.planOptimizePopup.08", "and %d more", n-shown)))
	}
	out = append(out, "",
		pad+vibe.Line("ui.plan_optimize.planOptimizePopup.09", "Finishes in ")+planTimeText(sc.opt.Before)+vibe.Line("ui.plan_optimize.planOptimizePopup.10", " now, ")+planTimeText(sc.opt.After)+vibe.Line("ui.plan_optimize.planOptimizePopup.11", " after")+
			vibe.Line("ui.plan_optimize.planOptimizePopup.12", ": saves ")+planTimeText(max(0, sc.opt.Before-sc.opt.After)),
		pad+theme.Keycap("Enter")+vibe.Line("ui.plan_optimize.planOptimizePopup.13", " keep it   ")+theme.Keycap("Esc")+vibe.Line("ui.plan_optimize.planOptimizePopup.14", " close"))
	return out
}
