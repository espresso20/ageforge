package game

import (
	"fmt"
	"math"
	"time"
)

// plan optimize (card 6PEgLKVM): reorder the plan's builds so the town grows
// as fast as it can. The rules:
//
//   - Only builds move. An advance and everything after it stay put: the
//     builds before an advance are what it waits for, and the ones after it
//     belong to an age the model cannot play yet.
//   - It never adds or removes a copy. A key keeps its total; the items are
//     split and merged to fit the new order, never past MaxPlanItems.
//   - Greedy by payback. Playing the town forward, each step takes the copy
//     (or the home or food producer that lets a producer be staffed, with
//     that producer) with the least of: the wait to afford it, its
//     construction time, and its price over the income it adds. Prices and
//     income are counted per resource, weighted by how close the resource
//     is to being the slowest to pay for what is left of the plan
//     (plan_sim.go). Housing and food count through the workers they let
//     the producers have.
//   - What the model cannot value keeps its place by rank: after as many
//     copies that grow the town as stood before it. That is research, deals
//     and upgrades (the player's priorities, which a build neither needs nor
//     gives), storage, knowledge and the army, and the items that cannot take
//     part now (wonders, banked items, another age's buildings, a tech not
//     researched). Only storage, knowledge and the army are paid for in the
//     model.
//   - The first PlanOptimizeCopies copies are reordered; the rest follow in
//     their order.
//   - The old order and the new are both played through the plan's own rules
//     (planModel.walk); the new one is kept only if it is faster.

// PlanOptimizeCopies is how many copies of the plan's builds the optimizer
// plays through.
const PlanOptimizeCopies = 2000

// Why a plan optimize left the plan alone.
const (
	PlanOptimizeNothing = "nothing" // no building in it grows the town
	PlanOptimizeBest    = "best"    // no order found is faster
)

// PlanOptimizeResult is what `plan optimize` did.
type PlanOptimizeResult struct {
	// Changed is whether the plan was reordered; Reason says why not.
	Changed bool
	Reason  string
	// Order labels the plan's items as they now stand ("3 Huts").
	Order []string
	// Before and After are the estimated times until the last building of
	// the plan's order starts, in the old order and the new.
	Before, After time.Duration
	// Beyond is how many copies lay past the simulation and kept their order.
	Beyond int
	// Waiting is how many copies cannot start at all until the player makes
	// what they cost.
	Waiting int
}

// optHeld is an item that keeps its place: a build the model does not value
// (play is true), or one that cannot take part. rank is how many copies that
// grow the town stood before it.
type optHeld struct {
	it   PlanItem
	rank int
	play bool
	key  int
}

// optRun is a stretch of one building the greedy walk bought.
type optRun struct {
	key, count int
}

// planOptimizeKind sorts a plan build item for the optimizer.
const (
	optMover  = iota // grows the town: moves
	optAnchor        // costs something and is not valued: keeps its rank
	optInert         // cannot take part now: keeps its rank, costs nothing
)

func (ge *GameEngine) planOptimizeKind(it PlanItem) int {
	def, ok := ge.Buildings.defs[it.Key]
	if !ok || len(it.Banked) > 0 || def.Category == "wonder" || ge.planBuildInvalid(it.Key, 0) != "" ||
		def.RequiredAge != "" && def.RequiredAge != ge.age || ge.Buildings.TechLocked(it.Key) {
		return optInert
	}
	if buildingGrows(def) {
		return optMover
	}
	return optAnchor
}

// PlanOptimize reorders the plan's builds for the fastest growth, or leaves
// the plan as it is when no order is faster. See the rules at the top of this
// file.
func (ge *GameEngine) PlanOptimize() (PlanOptimizeResult, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if len(ge.plan) == 0 {
		return PlanOptimizeResult{}, fmt.Errorf("The plan is empty.")
	}
	ge.recalculateRates()

	end := len(ge.plan)
	for i, it := range ge.plan {
		if it.Kind == PlanAdvance {
			end = i
			break
		}
	}
	region, tail := ge.plan[:end], ge.plan[end:]

	var (
		held    []optHeld
		rest    []PlanItem // past the simulation
		keys    []string
		counts  = map[string]int{}
		started = map[string]int{}
		before  []simItem
		moved   int
		copies  int
		capped  bool
	)
	keyOf := func(k string) int {
		for i, have := range keys {
			if have == k {
				return i
			}
		}
		keys = append(keys, k)
		return len(keys) - 1
	}
	for _, it := range region {
		if capped {
			rest = append(rest, it)
			continue
		}
		if it.Kind != PlanBuild {
			// Research, deals and upgrades neither need nor give a building:
			// they keep their place by rank, like the builds the model
			// cannot value.
			held = append(held, optHeld{it: it, rank: moved})
			continue
		}
		switch kind := ge.planOptimizeKind(it); kind {
		case optMover:
			take := min(it.Count, PlanOptimizeCopies-copies)
			counts[it.Key] += take
			started[it.Key] += it.Started
			before = append(before, simItem{key: keyOf(it.Key), left: take})
			moved += take
			copies += take
			if take < it.Count {
				cut := it
				cut.Count, cut.Started, cut.Banked = it.Count-take, 0, nil
				rest = append(rest, cut)
				capped = true
			}
		default:
			h := optHeld{it: it, rank: moved, play: kind == optAnchor}
			if h.play {
				h.key = keyOf(it.Key)
				before = append(before, simItem{key: h.key, left: it.Count})
			}
			held = append(held, h)
			copies += it.Count
		}
		if copies >= PlanOptimizeCopies {
			capped = true
		}
	}
	beyond := 0
	for _, it := range rest {
		if it.Kind == PlanBuild {
			beyond += it.Count
		}
	}
	if moved == 0 {
		return PlanOptimizeResult{Reason: PlanOptimizeNothing, Order: ge.planLabels(ge.plan)}, nil
	}

	// The model lists keys in the order given, so an index into keys is an
	// index into the model's buildings.
	m := ge.newPlanModel(keys)
	old := m.walk(before)

	var best []PlanItem
	var bestOut simOutcome
	for grain := 1; ; grain *= 2 {
		remaining := make([]int, len(m.keys))
		for k, key := range m.keys {
			remaining[k] = counts[key]
		}
		runs := m.greedy(remaining, held, grain)
		mid, play := ge.optAssemble(m, runs, held, started)
		plan := make([]PlanItem, 0, len(ge.plan)+len(runs))
		plan = append(plan, mid...)
		plan = append(plan, rest...)
		plan = append(plan, tail...)
		if planLoad(plan) > MaxPlanItems && grain < maxPlanCount {
			// Too many items for the plan: buy in bigger steps.
			continue
		}
		best, bestOut = plan, m.walk(play)
		break
	}

	res := PlanOptimizeResult{Beyond: beyond, Waiting: old.total - old.bought}
	iv := ge.tickIntervalLocked()
	res.Before = time.Duration(old.finish * float64(iv))
	res.After = time.Duration(bestOut.finish * float64(iv))
	faster := bestOut.bought > old.bought ||
		bestOut.bought == old.bought && old.finish-bestOut.finish > math.Max(1, 0.001*old.finish)
	if !faster || planItemsEqual(best, ge.plan) || planLoad(best) > MaxPlanItems {
		res.Reason = PlanOptimizeBest
		res.After = res.Before
		res.Order = ge.planLabels(ge.plan)
		return res, nil
	}
	ge.plan = best
	res.Changed = true
	res.Waiting = bestOut.total - bestOut.bought
	res.Order = ge.planLabels(ge.plan)
	return res, nil
}

// planLabels names every item of plan, in order.
func (ge *GameEngine) planLabels(plan []PlanItem) []string {
	out := make([]string, len(plan))
	for i, it := range plan {
		out[i] = ge.planItemLabel(it)
	}
	return out
}

// planItemsEqual reports whether two plans hold the same items in the same
// order.
func planItemsEqual(a, b []PlanItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Kind != b[i].Kind || a[i].Key != b[i].Key || a[i].Count != b[i].Count || a[i].Deal != b[i].Deal {
			return false
		}
	}
	return true
}

// greedy plays the town forward and returns the order it bought remaining
// (copies per key) in, with held items placed by rank. grain is how many
// copies of a pick it buys at once: the optimizer widens it when the order
// would not fit the plan's item limit.
func (m *planModel) greedy(remaining []int, held []optHeld, grain int) []optRun {
	run := m.newRun()
	var out []optRun
	total := 0
	for _, n := range remaining {
		total += n
	}
	placed, h := 0, 0
	dead := false // nothing more can be paid for, ever
	flush := func() {
		for h < len(held) && held[h].rank <= placed {
			out = append(out, optRun{key: -1 - h})
			if hd := held[h]; hd.play && !dead {
				for c := 0; c < hd.it.Count; c++ {
					if !run.waitFor(run.next(hd.key)) {
						break
					}
					run.buy(hd.key)
				}
			}
			h++
		}
	}
	for {
		flush()
		if total == 0 {
			break
		}
		k := -1
		if !dead {
			run.setDemand(remaining, held[h:])
			k = run.choose(remaining)
		}
		if k < 0 {
			// Nothing can be paid for. The rest stays in the plan's order
			// of keys.
			dead = true
			for key, n := range remaining {
				if n > 0 {
					out = append(out, optRun{key: key, count: n})
					remaining[key] = 0
				}
			}
			total = 0
			placed = math.MaxInt32
			continue
		}
		n := min(grain, remaining[k])
		if h < len(held) {
			n = min(n, max(1, held[h].rank-placed))
		}
		bought := 0
		for ; bought < n; bought++ {
			if !run.waitFor(run.next(k)) {
				break
			}
			run.buy(k)
		}
		if bought == 0 {
			// choose promised a copy it could not pay for: give up on it.
			dead = true
			continue
		}
		out = append(out, optRun{key: k, count: bought})
		remaining[k] -= bought
		total -= bought
		placed += bought
	}
	return out
}

// setDemand sets what is left to buy: remaining copies of each key, and the
// held builds that are still to come.
func (r *simRun) setDemand(remaining []int, held []optHeld) {
	for i := range r.demand {
		r.demand[i] = 0
	}
	for k, n := range remaining {
		r.m.addSpan(r.demand, k, r.bought[k], n)
	}
	for _, hd := range held {
		if hd.play {
			r.m.addSpan(r.demand, hd.key, r.bought[hd.key]+remaining[hd.key], hd.it.Count)
		}
	}
}

// choose is the key to buy next from remaining, or -1 when no copy can ever
// be paid for. It takes the least payback of a copy alone, or of a home or
// food producer together with a producer that needs its people (then it is
// the home or the food producer that comes first); when nothing pays back
// it takes the first copy in plan order that can be paid for at all.
func (r *simRun) choose(remaining []int) int {
	best, bestScore := -1, math.Inf(1)
	for k, n := range remaining {
		if n == 0 {
			continue
		}
		if s := r.payback(k, -1); s < bestScore {
			best, bestScore = k, s
		}
	}
	for p, n := range remaining {
		if n == 0 || r.m.bld[p].supports && r.m.bld[p].slots == 0 {
			continue
		}
		for s, ns := range remaining {
			if s == p || ns == 0 || !r.m.bld[s].supports {
				continue
			}
			if sc := r.payback(s, p); sc < bestScore {
				best, bestScore = s, sc
			}
		}
	}
	if best >= 0 {
		return best
	}
	for k, n := range remaining {
		if n > 0 && r.waitFor(r.next(k)) {
			return k
		}
	}
	return -1
}

// payback scores buying one copy of key a, with one of key b when b is not
// negative: the ticks to afford them, the longest construction, and the time
// they take to earn their price back. A price and an income are counted in
// ticks of the town's income of each resource, so what the town is short of
// weighs most. +Inf when they cannot be paid for or add nothing.
func (r *simRun) payback(a, b int) float64 {
	m := r.m
	price := r.price
	copy(price, r.next(a))
	build := m.bld[a].build
	if b >= 0 {
		for i, c := range r.next(b) {
			price[i] += c
		}
		build = math.Max(build, m.bld[b].build)
	}
	w := r.wait(price)
	if math.IsInf(w, 1) {
		return w
	}
	r.arrived[a]++
	if b >= 0 {
		r.arrived[b]++
	}
	m.income(r.arrived, r.scratch)
	r.arrived[a]--
	if b >= 0 {
		r.arrived[b]--
	}
	inc := r.income()
	// What is worth earning: a resource weighs by how far it is from being the
	// slowest to pay for what is left of the plan (the time the income needs
	// to cover its demand, against the slowest resource's). One nothing is
	// left to buy with weighs nothing, so food counts only through the
	// workers it feeds, which the income above already holds.
	floor := 0.0
	for _, v := range inc {
		floor = math.Max(floor, v)
	}
	if floor *= unfundedFloor; floor <= 0 {
		floor = 1
	}
	slowest := 0.0
	for i, d := range r.demand {
		slowest = math.Max(slowest, math.Max(0, d-r.stock[i])/math.Max(inc[i], floor))
	}
	num, den := 0.0, 0.0
	for i := range price {
		if r.demand[i] <= 0 {
			continue
		}
		per := math.Max(inc[i], floor)
		weight := 1.0
		if slowest > 0 {
			weight = math.Pow(math.Max(0, r.demand[i]-r.stock[i])/per/slowest, 3)
		}
		num += weight * price[i] / per
		den += weight * (r.scratch[i] - inc[i]) / per
	}
	if den <= 1e-12 {
		return math.Inf(1)
	}
	return w + build + num/den
}

// optAssemble turns the greedy's runs into plan items (consecutive copies of
// a key merged, no item over maxPlanCount) and the items the walk plays.
// started puts a key's Started count on its first item.
func (ge *GameEngine) optAssemble(m *planModel, runs []optRun, held []optHeld, started map[string]int) ([]PlanItem, []simItem) {
	var mid []PlanItem
	var play []simItem
	var isRun []bool // whether mid[i] is a run of movers, which a later run may extend
	startedLeft := make(map[string]int, len(started))
	for k, v := range started {
		startedLeft[k] = v
	}
	add := func(key string, n int) {
		k := m.kidx[key]
		for n > 0 {
			if last := len(mid) - 1; last >= 0 && isRun[last] && mid[last].Key == key && mid[last].Count < maxPlanCount {
				c := min(n, maxPlanCount-mid[last].Count)
				mid[last].Count += c
				play[len(play)-1].left += c
				n -= c
				continue
			}
			c := min(n, maxPlanCount)
			mid = append(mid, PlanItem{Kind: PlanBuild, Key: key, Count: c, Started: startedLeft[key]})
			isRun = append(isRun, true)
			startedLeft[key] = 0
			play = append(play, simItem{key: k, left: c})
			n -= c
		}
	}
	for _, r := range runs {
		if r.key < 0 {
			hd := held[-1-r.key]
			mid = append(mid, hd.it)
			isRun = append(isRun, false)
			if hd.play {
				play = append(play, simItem{key: hd.key, left: hd.it.Count})
			}
			continue
		}
		add(m.keys[r.key], r.count)
	}
	return mid, play
}
