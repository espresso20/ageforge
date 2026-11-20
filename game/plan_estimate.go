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
	// Copies is when each copy starts, for a modelled item of 2 to 12 copies
	// (shorter when some copy never starts).
	Copies []time.Duration
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
	// Needs is what the plan up to its first advance waits on, by resource.
	Needs []PlanNeed
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
			for _, t := range out.starts[k] {
				e.Copies = append(e.Copies, ticksToDuration(t, iv))
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
		if v.Kind == PlanResearch {
			if def, ok := ge.rules.Tech(v.Key); ok {
				e.End = e.Start + ticksToDuration(float64(ge.researchTicksIfStarted(def.Cost)), iv)
			}
		}
		est.Finish = max(est.Finish, e.End)
	}
	est.Needs = ge.planNeeds(views, iv)
	return est
}

// PlanNeed is one resource the plan up to its first advance has too little
// of: what every copy left would cost, what is held, and when the income
// covers the rest.
type PlanNeed struct {
	Res        string
	Need, Have float64
	PerSecond  float64
	CoverIn    time.Duration
	Never      bool // the income is zero or negative
	Waiting    int  // items not ready whose price includes it
}

// maxPlanNeeds is how many resources PlanEstimate.Needs lists.
const maxPlanNeeds = 8

// planNeeds sums the price of every copy left of the items up to and
// including the first advance (rising prices counted, a wonder at what its
// bank lacks), and keeps the resources the stores fall short of: those the
// income never covers first, then the longest to cover.
func (ge *GameEngine) planNeeds(views []PlanItemView, iv time.Duration) []PlanNeed {
	need := map[string]float64{}
	waiting := map[string]int{}
	ahead := map[string]int{}
	for i, v := range views {
		if v.Kind == PlanAdvance {
			break
		}
		it := ge.plan[i]
		cost := func() map[string]float64 {
			switch {
			case it.Kind == PlanBuild && ge.Buildings.defs[it.Key].Category == "wonder", it.Kind == PlanDeal, it.Kind == PlanExpedition:
				return v.Cost
			case it.Kind == PlanBuild:
				return ge.Buildings.CopyCost(it.Key, ahead[it.Key], ge.buildQueue)
			case it.Kind == PlanUpgrade:
				return ge.Buildings.UpgradeCopyCost(it.Key, ahead["up:"+it.Key], ge.buildQueue)
			}
			return nil
		}
		copies := 1
		if it.Kind == PlanBuild || it.Kind == PlanUpgrade {
			copies = it.Count
		}
		for n := 0; n < copies; n++ {
			for res, c := range cost() {
				need[res] += c
				if n == 0 && v.Status != PlanStatusReady && c > 0 {
					waiting[res]++
				}
			}
			switch it.Kind {
			case PlanBuild:
				ahead[it.Key]++
			case PlanUpgrade:
				ahead["up:"+it.Key]++
			}
		}
	}
	var out []PlanNeed
	for _, res := range sortedKeys(need) {
		have := ge.Resources.Get(res)
		if need[res] <= have {
			continue
		}
		n := PlanNeed{Res: res, Need: need[res], Have: have, Waiting: waiting[res]}
		if iv > 0 {
			n.PerSecond = ge.Resources.GetRate(res) / iv.Seconds()
		}
		if n.PerSecond <= 0 {
			n.Never = true
		} else {
			n.CoverIn = time.Duration(float64((need[res] - have) / n.PerSecond * float64(time.Second)))
		}
		out = append(out, n)
	}
	slices.SortStableFunc(out, func(a, b PlanNeed) int {
		switch {
		case a.Never != b.Never:
			if a.Never {
				return -1
			}
			return 1
		case a.CoverIn > b.CoverIn:
			return -1
		case a.CoverIn < b.CoverIn:
			return 1
		}
		return 0
	})
	return out[:min(len(out), maxPlanNeeds)]
}
