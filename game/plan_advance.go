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
	v := PlanItemView{Kind: it.Kind, Count: it.Count, Name: vibe.Line("game.plan_advance.planAdvanceView.01", "advance when ready")}
	if next := ge.progress.GetNextAge(ge.age); next != "" {
		v.Name = vibe.Line("game.plan_advance.planAdvanceView.02", "advance to the ") + ge.progress.GetAgeName(next)
	}
	if b := ge.planAdvanceBlocker(); b != "" {
		v.Status, v.Note = PlanStatusBlocked, b
	} else {
		v.Status, v.Progress = PlanStatusReady, 1
	}
	return v
}
