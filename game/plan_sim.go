package game

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
)

// The plan simulator: a small model of the town that `plan optimize`
// (plan_optimize.go) plays the build plan on, to price an order of builds.
//
// It reads the game's own numbers and recomputes only what a new copy
// changes:
//
//   - A copy's price is Buildings.CopyCost, so the cost curve, the techs'
//     cut and the copies under construction count as they do for `build`.
//   - A producer makes its listed output, 20% with nobody in it and up to
//     100% with a full crew (config.UnstaffedShare, StaffedShare), the same
//     rule productionWithWorkerOutput applies. The crews are the workers the
//     homes and the food can carry (below), spread evenly over the slots,
//     which is what the roster does on auto.
//   - Workers are limited by housing (popCapBaseLocked and the techs'
//     bonus, less the scholars and soldiers who live in it) and by the food
//     margin the recruiting routine keeps (recruitFoodMargin). People
//     already in the town stay.
//   - Every other bonus (morale, techs, pools, prestige, mastery) is read off
//     the rates the engine reports now, as a factor per resource, so the
//     model grows with the town's real multipliers and has no age tables.
//
// It models time in ticks. A copy is paid for when the plan walk starts it
// (the rules at the top of plan.go: in order, skipping with reservation) and
// produces after its construction time. It does not model research, trade,
// events, or what a copy that is not a producer or a home does.

// simArrival is a copy that stands at tick at.
type simArrival struct {
	at  float64
	key int
}

// simBuilding is one building key as the model sees it.
type simBuilding struct {
	key     string
	out     []float64 // listed output of one copy, by resource
	slots   int       // worker slots of one copy
	housing int       // people one copy houses
	people  int       // scholars and soldiers who come with one copy
	build   float64   // construction time, in ticks
	// grows reports whether a copy changes the town's income: it makes
	// something, or it houses people.
	grows bool
	// supports reports whether a copy lets producers be staffed: it houses
	// people or grows food.
	supports bool
}

// buildingGrows reports whether a copy of def changes the town's income: it
// makes something, or it houses people. Anything else (storage, knowledge,
// the army, a wonder) the optimizer leaves where the player put it.
func buildingGrows(def config.BuildingDef) bool {
	for _, eff := range def.Effects {
		if eff.Type == "production" && eff.Value > 0 || eff.Type == "capacity" && eff.Target == "population" && eff.Value > 0 {
			return true
		}
	}
	return false
}

// planModel is the town as it stands when `plan optimize` runs, with the
// buildings the plan names.
type planModel struct {
	ge   *GameEngine
	res  []string
	ridx map[string]int
	food int // index of food in res

	keys []string
	kidx map[string]int
	bld  []simBuilding

	stock0  []float64
	queue   []simArrival // copies under construction
	slots0  int          // worker slots in the buildings standing
	worker0 []float64    // output of the staffed kinds at a full crew
	fixed0  []float64    // output that has no crew: unstaffed kinds, ruins

	capBase int     // housing before the techs' bonus
	bonus   float64 // the techs' housing bonus
	others0 int     // scholars and soldiers
	people0 int     // workers and reserve
	hire    bool    // whether the routine recruits
	perHead float64 // food one person eats a tick, before prestige and mastery

	mult []float64 // by resource: what the bonuses make of listed output
	flat []float64 // by resource: income no building makes
	pk   float64   // prestige times mastery, which the rates carry

	fundTicks float64
	costs     [][][]float64 // costs[key][ahead]: the price of the copy that many bought first
	sums      [][][]float64 // sums[key][n]: the price of the first n copies together

	wbuf, fbuf []float64
}

// unfundedFloor is the share of the biggest income a resource's income is
// floored at when it weighs a price, so a resource nothing makes yet is dear
// but not infinitely so.
const unfundedFloor = 0.02

// newPlanModel reads the town. keys are the buildings the plan builds; the
// ones under construction are added. Caller holds the lock and has
// recalculated the rates.
func (ge *GameEngine) newPlanModel(keys []string) *planModel {
	bm := ge.Buildings
	m := &planModel{
		ge:        ge,
		kidx:      map[string]int{},
		ridx:      map[string]int{},
		hire:      !ge.autoRecruitOff,
		perHead:   ge.Workers.FoodCostPerWorker(),
		pk:        ge.speedK() * ge.Prestige.ProductionMultiplier(),
		fundTicks: planFundTicks,
		capBase:   ge.popCapBaseLocked(),
		bonus:     ge.Research.Bonus(config.EffectHousing, ""),
		others0:   ge.scholarTotal() + ge.soldierTotal(),
		people0:   ge.Workers.TotalPop(),
	}
	addKey := func(k string) {
		if _, ok := bm.defs[k]; !ok {
			return
		}
		if _, dup := m.kidx[k]; !dup {
			m.kidx[k] = len(m.keys)
			m.keys = append(m.keys, k)
		}
	}
	for _, k := range keys {
		addKey(k)
	}
	for _, q := range ge.buildQueue {
		addKey(q.BuildingKey)
	}

	// The resources the model tracks: what the buildings cost and make, what
	// the town makes now, and food.
	seen := map[string]bool{"food": true}
	produces := func(def config.BuildingDef) {
		for _, eff := range def.Effects {
			if eff.Type == "production" {
				seen[eff.Target] = true
			}
		}
	}
	for _, k := range m.keys {
		def := bm.defs[k]
		for res := range def.BaseCost {
			seen[res] = true
		}
		produces(def)
	}
	bm.eachBuilt(func(_ string, _ int, def config.BuildingDef) { produces(def) })
	for _, k := range sortedKeys(bm.ruins) {
		produces(bm.defs[k])
	}
	for res := range seen {
		m.res = append(m.res, res)
	}
	sort.Strings(m.res)
	for i, res := range m.res {
		m.ridx[res] = i
	}
	m.food = m.ridx["food"]
	n := len(m.res)

	m.bld = make([]simBuilding, len(m.keys))
	m.costs = make([][][]float64, len(m.keys))
	m.sums = make([][][]float64, len(m.keys))
	for i, k := range m.keys {
		def := bm.defs[k]
		b := simBuilding{key: k, out: make([]float64, n), slots: def.WorkerCapacity,
			people: def.Scholars + ge.soldierPerCopy(def), build: float64(ge.buildTicksLocked(def))}
		b.grows = buildingGrows(def)
		for _, eff := range def.Effects {
			switch {
			case eff.Type == "production" && eff.Value > 0:
				b.out[m.ridx[eff.Target]] += eff.Value
				if eff.Target == "food" {
					b.supports = true
				}
			case eff.Type == "capacity" && eff.Target == "population":
				b.housing += int(eff.Value)
				b.supports = true
			}
		}
		m.bld[i] = b
	}
	for _, q := range ge.buildQueue {
		if i, ok := m.kidx[q.BuildingKey]; ok {
			m.queue = append(m.queue, simArrival{at: float64(max(q.TicksLeft, 0)), key: i})
		}
	}

	m.stock0 = make([]float64, n)
	for i, res := range m.res {
		m.stock0[i] = ge.Resources.Get(res)
	}

	// What the standing buildings make, split as productionWithWorkerOutput
	// splits it.
	m.worker0 = make([]float64, n)
	m.fixed0 = make([]float64, n)
	bm.eachBuilt(func(_ string, count int, def config.BuildingDef) {
		m.slots0 += count * def.WorkerCapacity
		for _, eff := range def.Effects {
			if eff.Type != "production" {
				continue
			}
			if def.WorkerCapacity > 0 {
				m.worker0[m.ridx[eff.Target]] += eff.Value * float64(count)
			} else {
				m.fixed0[m.ridx[eff.Target]] += eff.Value * float64(count)
			}
		}
	})
	for _, k := range sortedKeys(bm.ruins) {
		for _, eff := range bm.defs[k].Effects {
			if eff.Type == "production" {
				m.fixed0[m.ridx[eff.Target]] += eff.Value * float64(bm.ruins[k]) * 0.50
			}
		}
	}

	// The bonuses, as the factor from listed output to the rate the engine
	// reports. A resource nothing makes yet takes the others' average, and
	// what the town makes of it without a building is a flat income.
	m.mult = make([]float64, n)
	m.flat = make([]float64, n)
	engine := bm.WorkerScaledProduction(ge.Workers.GetAssignedCount)
	drain := ge.foodDrainLocked()
	known, sum := make([]bool, n), 0.0
	for i, res := range m.res {
		gross := ge.Resources.GetRate(res) / m.pk
		if i == m.food {
			gross += drain
		}
		if raw := engine[res]; raw > 1e-12 {
			m.mult[i] = math.Max(0, gross/raw)
			known[i] = true
			sum += m.mult[i]
		} else {
			m.flat[i] = math.Max(0, gross)
		}
	}
	avg, cnt := 1.0, 0
	for i := range known {
		if known[i] {
			cnt++
		}
	}
	if cnt > 0 {
		avg = sum / float64(cnt)
	}
	for i := range known {
		if !known[i] {
			m.mult[i] = avg
		}
	}

	m.wbuf, m.fbuf = make([]float64, n), make([]float64, n)
	return m
}

// cost is the price of the copy of key k that ahead copies come before, by
// resource.
func (m *planModel) cost(k, ahead int) []float64 {
	for len(m.costs[k]) <= ahead {
		n := len(m.costs[k])
		price := m.ge.Buildings.CopyCost(m.keys[k], n, m.ge.buildQueue)
		c := make([]float64, len(m.res))
		for res, v := range price {
			if i, ok := m.ridx[res]; ok {
				c[i] = v
			}
		}
		m.costs[k] = append(m.costs[k], c)
	}
	return m.costs[k][ahead]
}

// addSpan adds to into the price of n copies of key k bought after ahead others.
func (m *planModel) addSpan(into []float64, k, ahead, n int) {
	if n <= 0 {
		return
	}
	m.cost(k, ahead+n-1)
	if len(m.sums[k]) == 0 {
		m.sums[k] = [][]float64{make([]float64, len(m.res))}
	}
	for len(m.sums[k]) <= ahead+n {
		last := m.sums[k][len(m.sums[k])-1]
		next := make([]float64, len(last))
		for i, c := range m.costs[k][len(m.sums[k])-1] {
			next[i] = last[i] + c
		}
		m.sums[k] = append(m.sums[k], next)
	}
	for i := range into {
		into[i] += m.sums[k][ahead+n][i] - m.sums[k][ahead][i]
	}
}

// income fills out with the net income per tick of every resource once the
// copies arrived (by key, on top of the town as it stands) stand and are
// staffed as far as the homes and the food allow.
func (m *planModel) income(arrived []int, out []float64) {
	slots, home, others := m.slots0, m.capBase, m.others0
	w, fx := m.wbuf, m.fbuf
	copy(w, m.worker0)
	copy(fx, m.fixed0)
	for k, n := range arrived {
		if n == 0 {
			continue
		}
		b := &m.bld[k]
		slots += n * b.slots
		home += n * b.housing
		others += n * b.people
		for r, v := range b.out {
			if v == 0 {
				continue
			}
			if b.slots > 0 {
				w[r] += float64(n) * v
			} else {
				fx[r] += float64(n) * v
			}
		}
	}
	fill := func(e int) float64 {
		if slots <= 0 {
			return 0
		}
		return float64(e) / float64(slots)
	}
	gross := func(r, e int) float64 {
		return m.flat[r] + m.mult[r]*(fx[r]+(config.UnstaffedShare+config.StaffedShare*fill(e))*w[r])
	}

	room := max(0, TechHousing(home, m.bonus)-others)
	top := min(slots, room)
	if !m.hire {
		top = min(top, m.people0)
	}
	stay := min(top, m.people0)
	// Hire into every slot the homes hold, then give back the workers the
	// food margin refuses.
	e := top
	for i := 0; i < 24 && e > stay; i++ {
		g := gross(m.food, e)
		net := g - m.perHead*float64(max(e, m.people0)+others)
		margin := math.Max(m.perHead, 0.05*g)
		if net >= margin {
			break
		}
		cut := 1
		if m.perHead > 0 {
			cut = max(1, int(math.Ceil((margin-net)/m.perHead)))
		}
		e = max(stay, e-cut)
	}
	for r := range out {
		out[r] = m.pk * gross(r, e)
	}
	out[m.food] -= m.pk * m.perHead * float64(max(e, m.people0)+others)
}

// simRun is the town played forward.
type simRun struct {
	m       *planModel
	t       float64
	stock   []float64
	bought  []int // copies bought in this run, by key
	arrived []int // copies standing on top of the town as it was, by key
	pending []simArrival
	inc     []float64
	fresh   bool
	scratch []float64
	price   []float64
	demand  []float64 // the price of everything still to buy, by resource
}

func (m *planModel) newRun() *simRun {
	r := &simRun{
		m:       m,
		stock:   append([]float64(nil), m.stock0...),
		bought:  make([]int, len(m.keys)),
		arrived: make([]int, len(m.keys)),
		pending: append([]simArrival(nil), m.queue...),
		inc:     make([]float64, len(m.res)),
		scratch: make([]float64, len(m.res)),
		price:   make([]float64, len(m.res)),
		demand:  make([]float64, len(m.res)),
	}
	sort.SliceStable(r.pending, func(i, j int) bool { return r.pending[i].at < r.pending[j].at })
	return r
}

// income is the net income per tick now.
func (r *simRun) income() []float64 {
	if !r.fresh {
		r.m.income(r.arrived, r.inc)
		r.fresh = true
	}
	return r.inc
}

// next is what the next copy of key k costs now: the copies standing and
// under construction when the run began, and the ones it has bought, come first.
func (r *simRun) next(k int) []float64 { return r.m.cost(k, r.bought[k]) }

func costTol(c float64) float64 { return 1e-9 * math.Max(1, c) }

// wait is how many ticks the income takes to cover cost, +Inf when a
// resource it is short of has no income.
func (r *simRun) wait(cost []float64) float64 {
	inc := r.income()
	w := 0.0
	for i, c := range cost {
		short := c - r.stock[i]
		if c <= 0 || short <= costTol(c) {
			continue
		}
		if inc[i] <= 0 {
			return math.Inf(1)
		}
		w = math.Max(w, short/inc[i])
	}
	return w
}

// flow runs dt ticks of income into the stores.
func (r *simRun) flow(dt float64) {
	if dt <= 0 {
		return
	}
	inc := r.income()
	for i := range r.stock {
		r.stock[i] = math.Max(0, r.stock[i]+inc[i]*dt)
	}
}

// advanceTo runs the town forward to tick t, copies arriving on the way.
func (r *simRun) advanceTo(t float64) {
	for len(r.pending) > 0 && r.pending[0].at <= t {
		a := r.pending[0]
		r.pending = r.pending[1:]
		r.flow(a.at - r.t)
		r.t = math.Max(r.t, a.at)
		r.arrived[a.key]++
		r.fresh = false
	}
	r.flow(t - r.t)
	r.t = math.Max(r.t, t)
}

// nextArrival is the ticks until the next copy stands (+Inf for none).
func (r *simRun) nextArrival() float64 {
	if len(r.pending) == 0 {
		return math.Inf(1)
	}
	return math.Max(0, r.pending[0].at-r.t)
}

// waitFor runs the town forward until cost is covered. It reports false when
// no income will ever cover it.
func (r *simRun) waitFor(cost []float64) bool {
	for {
		w := r.wait(cost)
		if w == 0 {
			return true
		}
		next := r.nextArrival()
		if math.IsInf(w, 1) && math.IsInf(next, 1) {
			return false
		}
		if w <= next {
			r.advanceTo(r.t + w)
			for i, c := range cost {
				if c > 0 && r.stock[i] < c && c-r.stock[i] <= costTol(c) {
					r.stock[i] = c
				}
			}
			return true
		}
		r.advanceTo(r.t + next)
	}
}

// buy pays for the next copy of key k, which the caller has made affordable,
// and sets it building.
func (r *simRun) buy(k int) {
	cost := r.next(k)
	for i, c := range cost {
		r.stock[i] = math.Max(0, r.stock[i]-c)
	}
	r.bought[k]++
	b := r.m.bld[k]
	at := r.t + b.build
	if b.build <= 0 {
		r.arrived[k]++
		r.fresh = false
		return
	}
	i := sort.Search(len(r.pending), func(i int) bool { return r.pending[i].at > at })
	r.pending = append(r.pending, simArrival{})
	copy(r.pending[i+1:], r.pending[i:])
	r.pending[i] = simArrival{at: at, key: k}
}

// funded mirrors planUnfunded: whether the income covers what cost is short
// of within a day, so a waiting item holds its price back.
func (r *simRun) funded(cost []float64) bool {
	inc := r.income()
	for i, c := range cost {
		short := c - r.stock[i]
		if c <= 0 || short <= costTol(c) {
			continue
		}
		if inc[i] <= 0 || short/inc[i] > r.m.fundTicks {
			return false
		}
	}
	return true
}

// covered reports whether the stores hold cost beyond what reserved holds back.
func (r *simRun) covered(cost, reserved []float64) bool {
	for i, c := range cost {
		if c > 0 && r.stock[i]-reserved[i] < c-costTol(c) {
			return false
		}
	}
	return true
}

// simItem is one plan item as the walk sees it: copies of a building.
type simItem struct {
	key  int
	left int
}

// simOutcome is how a walk ended.
type simOutcome struct {
	finish float64 // tick the last copy that could start started
	bought int
	total  int
}

// walk plays items (in plan order) as the plan does: every tick, in order,
// start what can be paid after what the items above it hold back, and let
// the rest save up. It ends when everything has started or nothing ever will.
func (m *planModel) walk(items []simItem) simOutcome {
	r := m.newRun()
	out := simOutcome{}
	for _, it := range items {
		out.total += it.left
	}
	left := append([]simItem(nil), items...)
	reserved := make([]float64, len(m.res))
	remaining := out.total
	for remaining > 0 {
		for started := true; started; {
			started = false
			for i := range reserved {
				reserved[i] = 0
			}
			for i := range left {
				it := &left[i]
				for it.left > 0 {
					cost := r.next(it.key)
					if r.covered(cost, reserved) {
						r.buy(it.key)
						it.left--
						remaining--
						out.bought++
						out.finish = r.t
						started = true
						continue
					}
					if r.funded(cost) {
						for j, c := range cost {
							reserved[j] += c
						}
					}
					break
				}
			}
		}
		if remaining == 0 {
			break
		}
		// Nothing starts now: run on to the first moment something can.
		dt := r.nextArrival()
		for i := range reserved {
			reserved[i] = 0
		}
		for i := range left {
			it := &left[i]
			if it.left == 0 {
				continue
			}
			cost := r.next(it.key)
			if !r.funded(cost) {
				continue
			}
			need := 0.0
			for j, c := range cost {
				if c <= 0 {
					continue
				}
				short := c + reserved[j] - r.stock[j]
				if short <= costTol(c) {
					continue
				}
				if inc := r.income()[j]; inc > 0 {
					need = math.Max(need, short/inc)
				} else {
					need = math.Inf(1)
				}
			}
			dt = math.Min(dt, need)
			for j, c := range cost {
				reserved[j] += c
			}
		}
		if math.IsInf(dt, 1) {
			break
		}
		// A step that would not move the clock (a start that rounding
		// puts a hair away) still has to make progress.
		r.advanceTo(r.t + math.Max(dt, 1e-9))
	}
	return out
}
