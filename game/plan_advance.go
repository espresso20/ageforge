package game

import "github.com/espresso20/ageforge/vibe"

// Advance items: `plan advance` advances to the next age as soon as its
// requirements are met, instead of when the player next looks in. It is the
// `advance` command queued: it costs nothing and reserves nothing, it waits
// behind a pending catastrophe like the command does, and it is used up when
// it advances (the rest of the plan belongs to the old age and drops out).
// It takes effect at its place in the plan: the items above it run first,
// in the age they were planned for, and those below wait for the next tick,
// when what belonged to the old age drops out. The next age's
// buildings and techs may be planned ahead; they wait for the advance.

// PlanAddAdvance appends an advance item (at most one).
func (ge *GameEngine) PlanAddAdvance() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if plannedCount(ge.plan, PlanAdvance, "") > 0 {
		return vibe.Errorf("game.plan_advance.PlanAddAdvance.01", "The plan already advances when ready.")
	}
	if ge.progress.GetNextAge(ge.age) == "" {
		return vibe.Errorf("game.plan_advance.PlanAddAdvance.02", "This is the final age. There is nothing to advance to.")
	}
	if planLoad(ge.plan) >= MaxPlanItems {
		return errPlanFull()
	}
	ge.plan = append(ge.plan, PlanItem{Kind: PlanAdvance, Count: 1})
	ge.notePlanSize()
	ge.logPlanAddLocked(ge.plan[len(ge.plan)-1], 1)
	return nil
}

// planAdvanceBlocker is why an advance item can't advance now ("" if it
// can).
func (ge *GameEngine) planAdvanceBlocker() string {
	switch {
	case ge.progress.GetNextAge(ge.age) == "":
		return vibe.Line("game.plan_advance.planAdvanceBlocker.01", "this is the final age")
	case ge.pendingCatastrophe != "":
		return vibe.Line("game.plan_advance.planAdvanceBlocker.02", "a catastrophe is pending")
	case ge.progress.CheckAdvancement(ge.age, ge.Buildings) == "":
		return vibe.Line("game.plan_advance.planAdvanceBlocker.03", "waiting for the requirements")
	}
	return ""
}

// planAdvanceView is an advance item for the UI.
func (ge *GameEngine) planAdvanceView(it PlanItem) PlanItemView {
	v := PlanItemView{Kind: it.Kind, Count: it.Count, Name: txtAdvanceWhenReady.String()}
	if next := ge.progress.GetNextAge(ge.age); next != "" {
		v.Name = vibe.Line("game.plan_advance.planAdvanceView.02", "advance to the ") + ge.progress.GetAgeName(next)
	}
	if b := ge.planAdvanceBlocker(); b != "" {
		v.Status, v.Note, v.When = PlanStatusBlocked, b, PlanWhenUnknown
	} else {
		v.Status, v.Progress = PlanStatusReady, 1
	}
	return v
}

// fillPlanPillars gives the advance item of st.Plan the charter's pillars,
// from the same AdvanceView the charter panel reads, so the item says which
// pillar holds the advance back and what it lacks. A blocker that a build or
// research item of the plan fills carries that item's number. It runs on the
// finished snapshot (the view needs the whole state), only when the plan
// has an advance item.
func (st *GameState) fillPlanPillars() {
	at := -1
	for i, v := range st.Plan {
		if v.Kind == PlanAdvance {
			at = i
			break
		}
	}
	if at < 0 {
		return
	}
	view := st.AdvanceView()
	if view.Final || len(view.Pillars) == 0 {
		return
	}
	pillars := make([]PlanPillar, 0, len(view.Pillars))
	for _, p := range view.Pillars {
		pp := PlanPillar{Name: p.Name, Done: p.Met, Percent: int(p.Progress * 100)}
		if p.Met {
			pp.Percent = 100
		} else {
			pp.Percent = min(pp.Percent, 99)
			for _, row := range p.Rows {
				if !row.Met {
					pp.Blocker, pp.BlockerKey = row.Label, row.Key
					break
				}
			}
			pp.Item = st.planItemFor(pp.BlockerKey)
		}
		pillars = append(pillars, pp)
	}
	// st.Plan shares its array with nothing but this snapshot (planViews made
	// it), so the view can be edited in place.
	st.Plan[at].Pillars = pillars
	if v := st.Plan[at]; v.Status == PlanStatusBlocked && view.Blockers == nil {
		st.Plan[at].Note = vibe.Sprintf("game.plan_advance.fillPlanPillars.01", "%d of %d pillars", view.Stand, len(view.Pillars))
	}
}

// planItemFor is the number (1-based) of the first plan item that builds or
// researches key, 0 when none does.
func (st *GameState) planItemFor(key string) int {
	if key == "" {
		return 0
	}
	for i, v := range st.Plan {
		if (v.Kind == PlanBuild || v.Kind == PlanResearch) && v.Key == key {
			return i + 1
		}
	}
	return 0
}
