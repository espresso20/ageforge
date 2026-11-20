package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// plan_text.go is the plan as words: the lines `plan list` and `plan
// optimize` reply with, and the titles and times the Plan panel shares.

// planItemTitle is "Hut ×3 (2 started)" or "research Tool Making".
func planItemTitle(v game.PlanItemView) string {
	switch v.Kind {
	case game.PlanResearch:
		return rowid.Line("ui.overlay_plan.planItemTitle.01", "research ") + v.Name
	case game.PlanAdvance:
		return v.Name
	}
	if v.Kind == game.PlanUpgrade {
		s := rowid.Sprintf("ui.overlay_plan.planItemTitle.02", "upgrade %d %s → %s", v.Count, v.Name, v.Target)
		if v.Started > 0 {
			s += rowid.Sprintf("ui.overlay_plan.planItemTitle.03", " (%d started)", v.Started)
		}
		return s
	}
	s := v.Name
	if v.Count > 1 {
		s += fmt.Sprintf(" ×%d", v.Count)
	}
	if v.Started > 0 {
		s += rowid.Sprintf("ui.overlay_plan.planItemTitle.04", " (%d started)", v.Started)
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
	txtPlanNever = rowid.E("ui.overlay_plan.planWhen.02", "never")
	txtPlanNow   = rowid.E("ui.overlay_plan.planWhen.03", "now")
)

// planWhen is when an item starts, as a word: "now", "in 4m", "never" or
// "needs you". Empty when nothing can be said. est is the plan estimate (nil
// without one): the walk's figure wins over the item's own price when it
// played the item.
func planWhen(v game.PlanItemView, est *game.PlanItemEstimate) string {
	switch v.When {
	case game.PlanWhenHand:
		return rowid.Line("ui.overlay_plan.planWhen.01", "needs you")
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
	return rowid.Line("ui.overlay_plan.planWhen.04", "in ") + textfmt.Duration(at)
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
		return rowid.Line("ui.overlay_plan.planListText.01", "The plan is empty. Add items with: plan build <building> [count[], or plan research <tech>.")
	}
	var sb strings.Builder
	sb.WriteString(theme.Paint(theme.RoleAccent, rowid.Line("ui.overlay_plan.planListText.02", "Build plan")) + "\n")
	if state.PlanPaused {
		sb.WriteString(theme.Paint(theme.RoleDim, rowid.Line("ui.overlay_plan.planListText.07", "Paused: nothing starts until you resume it.")) + "\n")
	}
	for i, v := range state.Plan {
		status := v.Status
		switch v.Status {
		case game.PlanStatusWaiting:
			status = rowid.Sprintf("ui.overlay_plan.planListText.03", "waiting, %.0f%%", v.Progress*100)
			if v.Short != "" {
				status += rowid.Line("ui.overlay_plan.planListText.04", " (short of ") + game.ResourceName(v.Short) + ")"
			}
		case game.PlanStatusBlocked:
			status = rowid.Line("ui.overlay_plan.planListText.05", "blocked: ") + v.Note
		}
		if when := planWhen(v, nil); when != "" && v.Status != game.PlanStatusBlocked {
			status += "  ·  " + rowid.Line("ui.overlay_plan.planListText.06", "starts ") + when
		}
		fmt.Fprintf(&sb, "  %d. %s  %s\n", i+1, kit.Escape(planItemTitle(v)), theme.Paint(theme.RoleDim, status))
	}
	return strings.TrimRight(sb.String(), "\n")
}

const planOptimizeShown = 12

// planOptimizeText is the reply to `plan optimize`: what happened, the new
// order when there is one, and the plan's finish time before and after.
func planOptimizeText(res game.PlanOptimizeResult) string {
	if res.Reason == game.PlanOptimizeNothing {
		return rowid.Line("ui.plan_optimize.planOptimizeText.01", "Nothing to reorder: none of the buildings in the plan changes how the town grows.")
	}
	var sb strings.Builder
	if res.Changed {
		sb.WriteString(rowid.Line("ui.plan_optimize.planOptimizeText.02", "Reordered the plan for faster growth."))
		sb.WriteString(planOptimizeTimes(res))
		sb.WriteString("\n")
		for i, label := range res.Order {
			if i == planOptimizeShown {
				fmt.Fprint(&sb, rowid.Sprintf("ui.plan_optimize.planOptimizeText.03", "  and %d more", len(res.Order)-planOptimizeShown))
				break
			}
			fmt.Fprintf(&sb, "  %d. %s\n", i+1, kit.Escape(label))
		}
	} else {
		sb.WriteString(rowid.Line("ui.plan_optimize.planOptimizeText.04", "The plan is already in the fastest order I could find, so I left it alone."))
		sb.WriteString(planOptimizeTimes(res))
	}
	if res.Beyond > 0 {
		fmt.Fprint(&sb, rowid.Sprintf("ui.plan_optimize.planOptimizeText.05", "\nOnly the first %s copies were reordered; the other %s keep their order after them.", textfmt.Int(game.PlanOptimizeCopies), textfmt.Int(res.Beyond)))
	}
	if res.Waiting > 0 {
		fmt.Fprint(&sb, rowid.Sprintf("ui.plan_optimize.planOptimizeText.06", "\n%s cannot start until the town makes what they cost.", textfmt.Count(res.Waiting, txtCopy.String(), txtCopies.String())))
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
		return rowid.Sprintf("ui.plan_optimize.planOptimizeTimes.01", " The last building should start in about %s.", textfmt.Duration(res.Before))
	}
	return rowid.Sprintf("ui.plan_optimize.planOptimizeTimes.02", " The last building should start in about %s instead of %s.", textfmt.Duration(res.After), textfmt.Duration(res.Before))
}
