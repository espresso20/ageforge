package game

import (
	"slices"
	"time"
)

// The plan estimate: when each item of the plan, in its present order, is
// expected to start, and when the plan is done. It plays the plan's builds
// through the same model plan optimize does (plan_sim.go), so its Before is
// optimize's Before. What the model does not play (research, upgrades,
// deals, wonders, expeditions, anything after an advance) takes the time its
// own price gives (PlanItemView.StartsIn).

// PlanItemEstimate is one item's estimated start.
type PlanItemEstimate struct {
	// Modeled is whether the plan model played this item. When it did,
	// Start is when its first copy starts and End when its last does.
	Modeled    bool
	Start, End time.Duration
	// Never is set when a copy of it cannot be paid for by any income.
	Never bool
}

// PlanEstimate is the plan's estimate.
type PlanEstimate struct {
	// Finish is when the last item that can be timed starts.
	Finish time.Duration
	// Before is when the last building the model plays starts: the figure
	// PlanOptimizePreview calls Before.
	Before time.Duration
	// Untimed is how many items have no time (never, by hand, or waiting on
	// something the estimate cannot time).
	Untimed int
	// Items has one entry for each plan item, in order.
	Items []PlanItemEstimate
}

// estimateCache is the last estimate and what it was made from.
type estimateCache struct {
	plan []PlanItem
	tick int
	est  PlanEstimate
	ok   bool
}

// PlanEstimate plays the plan in its present order and returns each item's
// estimated start and the plan's finish. It changes nothing in the plan, and
// keeps its answer for a few ticks while the plan stays as it is, so a
// panel may call it on every refresh.
func (ge *GameEngine) PlanEstimate() PlanEstimate {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if c := ge.estimateCache; c.ok && ge.tick >= c.tick && ge.tick-c.tick < planCacheTicks && slices.Equal(c.plan, ge.plan) {
		return cloneEstimate(c.est)
	}
	est := ge.planEstimateLocked()
	ge.estimateCache = estimateCache{plan: slices.Clone(ge.plan), tick: ge.tick, est: est, ok: true}
	return cloneEstimate(est)
}

func cloneEstimate(e PlanEstimate) PlanEstimate {
	e.Items = slices.Clone(e.Items)
	return e
}

// planEstimateLocked is PlanEstimate without the cache. Caller holds the
// write lock.
func (ge *GameEngine) planEstimateLocked() PlanEstimate {
	est := PlanEstimate{Items: make([]PlanItemEstimate, len(ge.plan))}
	if len(ge.plan) == 0 {
		return est
	}
	ge.recalculateRates()
	iv := ge.tickIntervalLocked()
	views := ge.planViews()
	in := ge.optInputOf()
	if len(in.before) > 0 {
		m := ge.newPlanModel(in.keys)
		out := m.walk(in.before)
		est.Before = ticksToDuration(out.finish, iv)
		for k, at := range in.beforeAt {
			if out.first[k] < 0 {
				est.Items[at] = PlanItemEstimate{Modeled: true, Never: true}
				continue
			}
			e := PlanItemEstimate{Modeled: true, Start: ticksToDuration(out.first[k], iv), End: ticksToDuration(out.last[k], iv)}
			// A copy that never started leaves its item open-ended.
			if out.got[k] < in.before[k].left {
				e.Never = true
			}
			est.Items[at] = e
		}
	}
	est.Finish = est.Before
	for i, v := range views {
		e := &est.Items[i]
		if !e.Modeled {
			e.Start, e.End = v.StartsIn, v.StartsIn
			e.Never = v.When == PlanWhenNever
		}
		if e.Never || !e.Modeled && v.When != PlanWhenTimed {
			est.Untimed++
			continue
		}
		est.Finish = max(est.Finish, e.End)
	}
	return est
}
