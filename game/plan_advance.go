package game

import "fmt"

// Advance items: `plan advance` advances to the next age as soon as its
// requirements are met, instead of when the player next looks in. It is the
// `advance` command queued: it costs nothing and reserves nothing, it waits
// behind a pending catastrophe like the command does, and it is used up when
// it advances (the rest of the plan belongs to the old age and drops out).
// It takes effect at its place in the plan: the items above it run first,
// in the age they were planned for, and those below wait for the next tick,
// so they cannot spend what the requirements count. The next age's
// buildings and techs may be planned ahead; they wait for the advance.

// PlanAddAdvance appends an advance item (at most one).
func (ge *GameEngine) PlanAddAdvance() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	for _, it := range ge.plan {
		if it.Kind == PlanAdvance {
			return fmt.Errorf("the plan already advances when ready")
		}
	}
	if ge.progress.GetNextAge(ge.age) == "" {
		return fmt.Errorf("this is the final age")
	}
	if len(ge.plan) >= MaxPlanItems {
		return fmt.Errorf("the plan is full (%d items) — remove one first", MaxPlanItems)
	}
	ge.plan = append(ge.plan, PlanItem{Kind: PlanAdvance, Count: 1})
	return nil
}

// planAdvanceBlocker is why an advance item can't advance now ("" if it
// can).
func (ge *GameEngine) planAdvanceBlocker() string {
	switch {
	case ge.progress.GetNextAge(ge.age) == "":
		return "this is the final age"
	case ge.pendingCatastrophe != "":
		return "a catastrophe is pending"
	case ge.progress.CheckAdvancement(ge.age, ge.Resources, ge.Buildings) == "":
		return "waiting for the requirements"
	}
	return ""
}

// planAdvanceView is an advance item for the UI.
func (ge *GameEngine) planAdvanceView(it PlanItem) PlanItemView {
	v := PlanItemView{Kind: it.Kind, Count: it.Count, Name: "advance when ready"}
	if next := ge.progress.GetNextAge(ge.age); next != "" {
		v.Name = "advance to the " + ge.progress.GetAgeName(next)
	}
	if b := ge.planAdvanceBlocker(); b != "" {
		v.Status, v.Note = PlanStatusBlocked, b
	} else {
		v.Status, v.Progress = PlanStatusReady, 1
	}
	return v
}
