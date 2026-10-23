package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
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
		fmt.Fprint(&sb, vibe.Sprintf("ui.plan_optimize.planOptimizeText.06", "\n%s cannot start until the town makes what they cost.", textfmt.Count(res.Waiting, vibe.Line("ui.plan_optimize.planOptimizeText.07", "copy"), vibe.Line("ui.plan_optimize.planOptimizeText.08", "copies"))))
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
