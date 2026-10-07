package game

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

// Worker shares: how the workforce splits across the worker domains, and the
// routine that keeps to it, live and offline (site/docs/workers-and-domains.md,
// "Worker shares").
//
//   - A share is a percent of the workforce for one domain (`workers share
//     knowledge 40`). Domains without one are on auto: they split what the
//     set shares leave in proportion to their worker slots. With no shares
//     set, the default, every domain is on auto, so the workers follow the
//     buildings you have. A share of 0 keeps a domain empty.
//   - The routine runs every staffEveryTicks ticks and once per offline step.
//     It puts idle workers to work, then, with auto-recruit on (the default),
//     recruits into the empty worker slots while housing allows and the food
//     income stays positive. Each worker goes to the domain furthest below its
//     share; within a domain, this age's buildings fill first, newest first.
//   - It never takes a worker out of a building, except out of a domain whose
//     share is 0. A share command moves workers once, to match the new
//     shares; after that the player's own assign and unassign stick.
//   - After a worker command (recruit, assign, unassign, dismiss, sell,
//     upgrade) it waits staffHoldTicks, so it never grabs the workers the
//     player is moving by hand.
//   - Food: a recruit must leave the net food rate at or above a margin (one
//     worker's food, or recruitFoodMarginShare of the food production if that
//     is more). When the next worker can't be fed, a worker for a food
//     building that grows more than it eats can still come, so a small food
//     share never stops growth. While food is short, idle workers go to food
//     buildings first. Nothing recruits while the food store is empty.
//
// Shares, the auto-recruit switch and the wait are saved. Prestige, Succumb
// and a new game put the shares back on auto (startRunShares, where Pacing
// v2's legacy kit will carry them across a prestige instead). Auto-recruit is
// a preference, kept across prestige and Succumb like wonder overflow; a new
// game turns it back on.

const (
	// staffEveryTicks is how often the routine runs in live play: every 10
	// seconds at 1x. Offline it runs once per catch-up step.
	staffEveryTicks = 5
	// staffHoldTicks is how long the routine waits after a worker command: a
	// minute at 1x.
	staffHoldTicks = 30
	// recruitFoodMarginShare is the part of the food production the routine
	// keeps as surplus when it recruits (at least one worker's food).
	recruitFoodMarginShare = 0.05
)

// shareDomain is one worker domain as the routine sees it.
type shareDomain struct {
	key      string
	slots    int     // worker slots in the domain's built buildings
	assigned int     // workers in them
	weight   float64 // effective share of the workforce, 0 to 1 (shareWeightsFor)
	set      bool    // the player set this domain's share
	zero     bool    // set to 0: the domain gets no workers
}

func (d *shareDomain) free() int { return d.slots - d.assigned }

// shareWeightsFor fills in each domain's weight, set and zero from shares: a
// set share is its percent (scaled down when the set shares add up to more
// than 100), and the domains on auto split what the set shares leave by
// worker slots.
func shareWeightsFor(doms []shareDomain, shares map[string]float64) {
	total := 0.0
	for i := range doms {
		p, ok := shares[doms[i].key]
		doms[i].set, doms[i].zero = ok, ok && p <= 0
		if ok {
			total += p
		}
	}
	scale := math.Max(100, total)
	left := math.Max(0, 100-total) / 100 // exactly 0 when the set shares take it all
	autoSlots := 0
	for i := range doms {
		if doms[i].set {
			doms[i].weight = shares[doms[i].key] / scale
		} else {
			autoSlots += doms[i].slots
		}
	}
	for i := range doms {
		if doms[i].set {
			continue
		}
		doms[i].weight = 0
		if autoSlots > 0 {
			doms[i].weight = float64(left*float64(doms[i].slots)) / float64(autoSlots)
		}
	}
}

// nextShareDomain is the domain the next worker goes to, -1 when none has a
// free slot: of the domains with a share left, the one furthest below it
// ((workers + ½) ÷ share lowest, the earlier domain on a tie); when those are
// all full, the auto domain with the fewest workers for its slots. A domain
// set to 0 never gets one, nor does skip.
func nextShareDomain(doms []shareDomain, skip int) int {
	best, bestV := -1, 0.0
	for i := range doms {
		d := &doms[i]
		if i == skip || d.zero || d.weight <= 0 || d.free() <= 0 {
			continue
		}
		if v := (float64(d.assigned) + 0.5) / d.weight; best < 0 || v < bestV {
			best, bestV = i, v
		}
	}
	if best >= 0 {
		return best
	}
	for i := range doms {
		d := &doms[i]
		if i == skip || d.set || d.free() <= 0 {
			continue
		}
		if v := (float64(d.assigned) + 0.5) / float64(d.slots); best < 0 || v < bestV {
			best, bestV = i, v
		}
	}
	return best
}

// recruitFoodMargin is the net food rate a recruit must leave: one worker's
// food, or recruitFoodMarginShare of the food production if that is more.
func recruitFoodMargin(net, drain, perWorker float64) float64 {
	return math.Max(perWorker, float64(recruitFoodMarginShare*(net+drain)))
}

// cleanSharesIn is a saved or typed shares map as the engine keeps it: set's
// worker domains only, each percent finite, from 0 to 100, to a tenth. nil
// when nothing is left.
func cleanSharesIn(set *rules.Set, in map[string]float64) map[string]float64 {
	var out map[string]float64
	for _, d := range set.WorkerDomains() {
		p, ok := in[d]
		if !ok || math.IsNaN(p) || math.IsInf(p, 0) {
			continue
		}
		if out == nil {
			out = make(map[string]float64)
		}
		out[d] = math.Round(math.Min(100, math.Max(0, p))*10) / 10
	}
	return out
}

// cloneShares copies a shares map so snapshots and saves share nothing with
// the live one.
func cloneShares(m map[string]float64) map[string]float64 {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]float64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// IsWorkerDomain reports whether key is one of set's worker domains.
func IsWorkerDomain(set *rules.Set, key string) bool {
	for _, d := range set.WorkerDomains() {
		if d == key {
			return true
		}
	}
	return false
}

// SharePercent prints a share's percent: "40%", "12.5%".
func SharePercent(p float64) string {
	if p == math.Trunc(p) {
		return fmt.Sprintf("%.0f%%", p)
	}
	return fmt.Sprintf("%.1f%%", p)
}

// DomainName is a worker domain's display name: "Knowledge".
func DomainName(key string) string {
	if key == "" {
		return key
	}
	return strings.ToUpper(key[:1]) + key[1:]
}

// ===== Engine state =====

// workersHeld reports whether the routine is waiting after a worker command.
func (ge *GameEngine) workersHeld() bool { return ge.tick < ge.staffHoldUntil }

// holdStaffing starts the routine's wait after a worker command. Caller
// holds the write lock.
func (ge *GameEngine) holdStaffing() { ge.staffHoldUntil = ge.tick + staffHoldTicks }

// startRunShares sets the shares for a new run (prestige and Succumb): back
// to auto. Pacing v2's legacy kit will carry them across from here instead.
func (ge *GameEngine) startRunShares() {
	ge.workerShares = nil
	ge.staffHoldUntil = 0
}

// staffView is the routine's working copy of the worker domains: each one's
// slots, workers and weight, and its built buildings in fill order.
type staffView struct {
	doms []shareDomain
	blds [][]string // per domain: built buildings that take workers, in fill order
	food int        // index of the food domain, -1 if none
}

// newStaffView reads the worker domains off the built buildings. Caller holds
// the lock.
func (ge *GameEngine) newStaffView() *staffView {
	domains := ge.rules.WorkerDomains()
	sv := &staffView{doms: make([]shareDomain, len(domains)), blds: make([][]string, len(domains)), food: -1}
	idx := make(map[string]int, len(domains))
	for i, d := range domains {
		sv.doms[i].key = d
		idx[d] = i
		if d == "food" {
			sv.food = i
		}
	}
	// Fill order: this age's buildings before superseded ones, then the
	// newest age first, then the higher tier of its line, then by key.
	type fillKey struct {
		key       string
		legacy    bool
		age, tier int
	}
	order := make([][]fillKey, len(domains))
	ages := ge.progress.ageIndex
	ge.Buildings.eachBuilt(func(key string, count int, def config.BuildingDef) {
		i, ok := idx[def.WorkerDomain]
		if !ok || def.WorkerCapacity <= 0 {
			return
		}
		sv.doms[i].slots += def.WorkerCapacity * count
		sv.doms[i].assigned += ge.Workers.GetAssignedCount("worker", key)
		order[i] = append(order[i], fillKey{key, ge.Buildings.IsLegacy(key), ages[def.RequiredAge], def.LineageTier})
	})
	for i, keys := range order {
		sort.Slice(keys, func(a, b int) bool {
			ka, kb := keys[a], keys[b]
			switch {
			case ka.legacy != kb.legacy:
				return !ka.legacy
			case ka.age != kb.age:
				return ka.age > kb.age
			case ka.tier != kb.tier:
				return ka.tier > kb.tier
			}
			return ka.key < kb.key
		})
		sv.blds[i] = make([]string, len(keys))
		for j, k := range keys {
			sv.blds[i][j] = k.key
		}
	}
	shareWeightsFor(sv.doms, ge.workerShares)
	return sv
}

// freeSlots is how many more workers the built buildings take, all domains
// together. It allocates nothing, so the routine checks it before building a
// staffView.
func (ge *GameEngine) freeSlots() int {
	free := 0
	ge.Buildings.eachBuilt(func(key string, count int, def config.BuildingDef) {
		if def.WorkerCapacity > 0 {
			free += def.WorkerCapacity*count - ge.Workers.GetAssignedCount("worker", key)
		}
	})
	return free
}

// slotsFree is how many more workers building key takes.
func (ge *GameEngine) slotsFree(key string) int {
	return ge.Buildings.defs[key].WorkerCapacity*ge.Buildings.GetCount(key) - ge.Workers.GetAssignedCount("worker", key)
}

// staffDomain assigns n idle workers to domain i's buildings in fill order
// and returns how many it placed.
func (ge *GameEngine) staffDomain(sv *staffView, i, n int) int {
	placed := 0
	for _, key := range sv.blds[i] {
		if placed >= n {
			break
		}
		if k := min(n-placed, ge.slotsFree(key)); k > 0 && ge.Workers.Assign("worker", key, k) {
			placed += k
		}
	}
	return placed
}

// unstaffOne takes one worker out of domain i, from the last building in fill
// order that has one (superseded and older buildings first). It reports
// whether it found one.
func (ge *GameEngine) unstaffOne(sv *staffView, i int) bool {
	for j := len(sv.blds[i]) - 1; j >= 0; j-- {
		if key := sv.blds[i][j]; ge.Workers.GetAssignedCount("worker", key) > 0 {
			if ge.Workers.Unassign("worker", key, 1) {
				sv.doms[i].assigned--
				return true
			}
		}
	}
	return false
}

// applyStaffPlan assigns plan[i] idle workers to each domain i. The view's
// counts already include them.
func (ge *GameEngine) applyStaffPlan(sv *staffView, plan []int) {
	for i, n := range plan {
		if n > 0 {
			ge.staffDomain(sv, i, n)
		}
	}
}

// recruitFood is what the routine plans recruits with: the net food rate now
// (the last rates, corrected for workers who came or went since), each
// worker's food and the margin a recruit must leave. All three are in the
// units a worker eats in, before Era Mastery's speed-up (foodBeforeMastery).
func (ge *GameEngine) recruitFood() (net, perWorker, margin float64) {
	drain := ge.Workers.FoodDrain()
	if r := ge.Resources.resources["food"]; r != nil {
		net = foodBeforeMastery(r.Rate, r.Breakdown) - drain
	}
	perWorker = ge.Workers.FoodCostPerWorker()
	return net, perWorker, recruitFoodMargin(net, drain, perWorker)
}

// foodBeforeMastery is the food made per tick before any worker eats and
// before Era Mastery multiplies the net rate, read off a food rate and its
// breakdown. On known ground the rate is k times what is made less what is
// eaten, while a worker's food (FoodCostPerWorker) and what a food worker
// grows (foodPerWorker) are per tick before k. Compared as they stood, the
// rate looked k times richer than it was, and the routine recruited more
// workers than the food fed. Breakdown.FoodDrain is minus the drain the
// last rates counted.
func foodBeforeMastery(rate float64, b RateBreakdown) float64 {
	return rate - b.MasteryRate - b.FoodDrain
}

// FoodBeforeMastery is foodBeforeMastery for the Workers panel.
func FoodBeforeMastery(rate float64, b RateBreakdown) float64 { return foodBeforeMastery(rate, b) }

// foodWorkerFactor is what multiplies the food a worker adds to a food
// building: morale, times the all-production and food bonuses plus the
// worker output bonus (which is added to them, not multiplied), read off the
// last rates, times the tech layer's factor on food, times the Cosmic
// Legacy's.
func (ge *GameEngine) foodWorkerFactor() float64 {
	bonuses := 1.0
	if r := ge.Resources.resources["food"]; r != nil && r.Breakdown.BuildingRate > 0 {
		bonuses = math.Max(0, 1+r.Breakdown.BonusRate/r.Breakdown.BuildingRate)
	}
	return float64(float64(float64(ge.moraleMultiplier()*math.Max(0, bonuses+ge.workerBonus))*ge.Research.OutputFactor("food")) * ge.cosmicLegacyFactor())
}

// foodPerWorker is the food one more worker in building key grows a tick: a
// full building makes five times an empty one, so each slot adds 80% of the
// building's food over its slots, times factor (foodWorkerFactor).
func (ge *GameEngine) foodPerWorker(key string, factor float64) float64 {
	def := ge.Buildings.defs[key]
	if def.WorkerCapacity <= 0 {
		return 0
	}
	out := 0.0
	for _, eff := range def.Effects {
		if eff.Type == "production" && eff.Target == "food" {
			out += eff.Value
		}
	}
	return float64(float64(out*0.8) / float64(def.WorkerCapacity) * factor)
}

// foodGainAt is the food the k-th new food worker of this run (from 0) would
// grow, in the food building it would fill; 0 when no food slot is left.
func (ge *GameEngine) foodGainAt(sv *staffView, k int, factor float64) float64 {
	if sv.food < 0 {
		return 0
	}
	for _, key := range sv.blds[sv.food] {
		free := ge.slotsFree(key)
		if k < free {
			return ge.foodPerWorker(key, factor)
		}
		k = max(0, k-free)
	}
	return 0
}

// staffCounts is what one run of the routine did.
type staffCounts struct{ moved, placed, hired int }

func (c staffCounts) any() bool { return c.moved+c.placed+c.hired > 0 }

func (c *staffCounts) add(o staffCounts) {
	c.moved += o.moved
	c.placed += o.placed
	c.hired += o.hired
}

// staffByShares runs the routine once, without the wait: workers in domains
// set to 0 move out, idle workers go to work, and with recruit set (and
// auto-recruit on) recruits fill the slots left. Caller holds the write lock
// and recalculates the rates when it did anything.
func (ge *GameEngine) staffByShares(recruit bool) staffCounts {
	var c staffCounts
	if !ge.Workers.IsUnlocked("worker") {
		return c
	}
	// Cheap checks first: in a settled game every slot is filled or nobody
	// can come, and this runs every few ticks.
	zero := ge.zeroShareStaffed()
	if !zero && ge.freeSlots() <= 0 {
		return c
	}
	idle := ge.Workers.IdleCount("worker")
	room := 0
	if recruit && !ge.autoRecruitOff && ge.Resources.Get("food") > 0 {
		room = ge.popCapLocked() - ge.Workers.TotalPop()
	}
	if idle <= 0 && room <= 0 && !zero {
		return c
	}
	sv := ge.newStaffView()
	c.moved = ge.drainZeroShares(sv)
	c.placed = ge.placeIdle(sv)
	if room > 0 {
		c.hired = ge.recruitByShares(sv, room)
	}
	return c
}

// zeroShareStaffed reports whether a domain set to 0 still has workers, so
// the routine has some to move.
func (ge *GameEngine) zeroShareStaffed() bool {
	for d, p := range ge.workerShares {
		if p <= 0 && ge.Workers.GetDomainCount(d) > 0 {
			return true
		}
	}
	return false
}

// drainZeroShares moves the workers of domains set to 0 into free slots
// elsewhere, by the shares. Returns how many moved.
func (ge *GameEngine) drainZeroShares(sv *staffView) int {
	moved := 0
	for i := range sv.doms {
		for sv.doms[i].zero && sv.doms[i].assigned > 0 {
			to := nextShareDomain(sv.doms, i)
			if to < 0 || !ge.unstaffOne(sv, i) {
				break
			}
			if ge.staffDomain(sv, to, 1) == 1 {
				sv.doms[to].assigned++
			}
			moved++
		}
	}
	return moved
}

// placeIdle puts the idle workers to work, each in the domain furthest below
// its share, food buildings first while food is short. Returns how many it
// placed.
func (ge *GameEngine) placeIdle(sv *staffView) int {
	idle := ge.Workers.IdleCount("worker")
	if idle <= 0 {
		return 0
	}
	net, _, margin := ge.recruitFood()
	factor := ge.foodWorkerFactor()
	plan := make([]int, len(sv.doms))
	n, food := 0, 0
	for n < idle {
		i := nextShareDomain(sv.doms, -1)
		if f := sv.food; net < margin && f >= 0 && !sv.doms[f].zero && sv.doms[f].free() > 0 {
			i = f
		}
		if i < 0 {
			break
		}
		if i == sv.food {
			net += ge.foodGainAt(sv, food, factor)
			food++
		}
		plan[i]++
		sv.doms[i].assigned++
		n++
	}
	ge.applyStaffPlan(sv, plan)
	return n
}

// recruitByShares recruits up to room workers into the free slots, by the
// shares, while the food rate stays at or above the margin. A recruit the
// food won't feed goes to a food building instead when that worker grows
// more than they eat. Returns how many it recruited.
func (ge *GameEngine) recruitByShares(sv *staffView, room int) int {
	net, per, margin := ge.recruitFood()
	factor := ge.foodWorkerFactor()
	plan := make([]int, len(sv.doms))
	n, food := 0, 0
	for n < room {
		i := nextShareDomain(sv.doms, -1)
		if i < 0 {
			break
		}
		gain := 0.0
		if i == sv.food {
			gain = ge.foodGainAt(sv, food, factor)
		}
		if net+gain-per < margin {
			f := sv.food
			if f < 0 || f == i || sv.doms[f].zero || sv.doms[f].free() <= 0 {
				break
			}
			g := ge.foodGainAt(sv, food, factor)
			if g <= per || net+g-per < margin {
				break
			}
			i, gain = f, g
		}
		if i == sv.food {
			food++
		}
		net = net + gain - per
		plan[i]++
		sv.doms[i].assigned++
		n++
	}
	if n == 0 || !ge.Workers.Recruit("worker", n, ge.popCapLocked()) {
		return 0
	}
	ge.Stats.RecordRecruit(n)
	ge.applyStaffPlan(sv, plan)
	return n
}

// rebalanceShares moves workers between domains toward the shares, once, as
// a share command does: one at a time from the domain furthest over its share
// to the one furthest under it with a free slot, while that evens them out.
// Domains set to 0 and auto domains the set shares leave nothing go first.
// Food workers stay where moving them would push the food rate below zero.
// Returns how many moved.
func (ge *GameEngine) rebalanceShares(sv *staffView) int {
	net, _, _ := ge.recruitFood()
	factor := ge.foodWorkerFactor()
	foodStays := false
	moved := 0
	for limit := ge.Workers.TotalPop() + len(sv.doms); limit > 0; limit-- {
		to := nextShareDomain(sv.doms, -1)
		if to < 0 {
			break
		}
		toW := sv.doms[to].weight
		from, fromV := -1, 0.0
		for i := range sv.doms {
			d := &sv.doms[i]
			if i == to || d.assigned <= 0 || (i == sv.food && foodStays) {
				continue
			}
			v := math.Inf(1)
			if !d.zero && d.weight > 0 {
				v = (float64(d.assigned) - 0.5) / d.weight
			}
			if toW <= 0 && !d.zero {
				continue // an auto domain the shares leave nothing only takes from a domain set to 0
			}
			if from < 0 || v > fromV {
				from, fromV = i, v
			}
		}
		if from < 0 || (!math.IsInf(fromV, 1) && fromV <= (float64(sv.doms[to].assigned)+0.5)/toW) {
			break
		}
		if from == sv.food {
			loss := ge.foodLossOfOne(sv, factor)
			if net-loss < 0 {
				foodStays = true
				continue
			}
			net -= loss
		}
		if to == sv.food {
			net += ge.foodGainAt(sv, 0, factor)
		}
		if !ge.unstaffOne(sv, from) {
			break
		}
		if ge.staffDomain(sv, to, 1) == 1 {
			sv.doms[to].assigned++
		}
		moved++
	}
	return moved
}

// foodLossOfOne is the food the next worker unstaffOne would take out of the
// food domain grows.
func (ge *GameEngine) foodLossOfOne(sv *staffView, factor float64) float64 {
	keys := sv.blds[sv.food]
	for j := len(keys) - 1; j >= 0; j-- {
		if ge.Workers.GetAssignedCount("worker", keys[j]) > 0 {
			return ge.foodPerWorker(keys[j], factor)
		}
	}
	return 0
}

// popCapLocked is the housing: what the housing buildings hold plus the
// population bonuses, raised by the techs' housing bonus. Caller holds the
// lock.
func (ge *GameEngine) popCapLocked() int {
	base := ge.Buildings.GetPopCapacity() + int(ge.permanentBonuses["population"]+ge.Prestige.GetBonuses()["population"])
	return TechHousing(base, ge.Research.Bonus(config.EffectHousing, ""))
}

// TechHousing is housing of base with the techs' housing bonus (a fraction:
// 0.05 is +5%): base × (1 + bonus), rounded up to a whole person, so the
// first small bonus on a small village still houses one more.
func TechHousing(base int, bonus float64) int {
	if bonus <= 0 || base <= 0 {
		return base
	}
	// The epsilon keeps a product that is whole on paper (20 × 1.05 = 21)
	// from rounding up a second time on floating point's last digit.
	return int(math.Ceil(float64(float64(base)*(1+bonus)) - 1e-9))
}

// keepSharesLive is the routine's live run, from doTick: unless it is waiting
// after a worker command, it staffs and recruits, recalculates the rates
// when that changed anything, and logs what it did as a routine line.
func (ge *GameEngine) keepSharesLive() {
	if ge.workersHeld() {
		return
	}
	before := ge.Workers.TotalPop()
	c := ge.staffByShares(true)
	if !c.any() {
		return
	}
	ge.recalculateRates()
	ge.addLog(LogRoutine, "Shares: "+c.describe(ge.Workers.TotalPop(), ge.popCapLocked())+".")
	if before == 0 && c.hired > 0 {
		// The first recruits of a run (or after everyone died): say once,
		// as a note rather than a routine line, that this happens on its
		// own.
		ge.addLog("info", "Workers arrive on their own: the game recruits into empty worker slots while housing and food allow, and puts them to work by your worker shares. Type workers to see them; workers auto-recruit off to recruit by hand.")
	}
}

// describe is what a run did, for the log: "recruited 3 workers (population
// 12/20), put 2 idle workers to work".
func (c staffCounts) describe(pop, popCap int) string {
	var parts []string
	if c.hired > 0 {
		parts = append(parts, fmt.Sprintf("recruited %s (population %d/%d)", textfmt.Count(c.hired, "worker", "workers"), pop, popCap))
	}
	if c.placed > 0 {
		parts = append(parts, "put "+textfmt.Count(c.placed, "idle worker", "idle workers")+" to work")
	}
	if c.moved > 0 {
		parts = append(parts, "moved "+textfmt.Count(c.moved, "worker", "workers")+" out of domains set to 0%")
	}
	return strings.Join(parts, ", ")
}

// ===== Public API =====

// WorkerShares returns the shares the player set (domain → percent); nil when
// every domain is on auto.
func (ge *GameEngine) WorkerShares() map[string]float64 {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return cloneShares(ge.workerShares)
}

// AutoRecruit reports whether the shares routine recruits.
func (ge *GameEngine) AutoRecruit() bool {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return !ge.autoRecruitOff
}

// SetAutoRecruit turns the routine's recruiting on or off. Turning it on
// recruits at once, as far as housing and food allow.
func (ge *GameEngine) SetAutoRecruit(on bool) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.autoRecruitOff = !on
	if on {
		ge.staffHoldUntil = 0
		if c := ge.staffByShares(true); c.any() {
			ge.recalculateRates()
		}
	}
}

// ShareReply is what a share command did, for its reply.
type ShareReply struct {
	Line    string // what the shares are now and what moved
	Warning bool   // the shares leave the food buildings without workers
}

// UnknownDomainError refuses a word that names none of set's worker domains.
func UnknownDomainError(set *rules.Set, domain string) error {
	return fmt.Errorf("Unknown worker domain '%s'. The domains are %s.", domain, strings.Join(set.WorkerDomains(), ", "))
}

// SetWorkerShare sets domain's share of the workforce to percent (0 to 100),
// then moves workers once to match the shares and staffs and recruits as the
// routine does.
func (ge *GameEngine) SetWorkerShare(domain string, percent float64) (ShareReply, error) {
	if !IsWorkerDomain(ge.rules, domain) {
		return ShareReply{}, UnknownDomainError(ge.rules, domain)
	}
	if math.IsNaN(percent) || math.IsInf(percent, 0) || percent < 0 || percent > 100 {
		return ShareReply{}, fmt.Errorf("A share is a percent from 0 to 100 (got %v).", percent)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	shares := cloneShares(ge.workerShares)
	if shares == nil {
		shares = make(map[string]float64)
	}
	shares[domain] = percent
	ge.workerShares = cleanSharesIn(ge.rules, shares)
	return ge.applySharesLocked(domain), nil
}

// ClearWorkerShare puts domain back on auto, or every domain when domain is
// "", then moves workers once to match, as SetWorkerShare does.
func (ge *GameEngine) ClearWorkerShare(domain string) (ShareReply, error) {
	if domain != "" && !IsWorkerDomain(ge.rules, domain) {
		return ShareReply{}, UnknownDomainError(ge.rules, domain)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if domain == "" {
		ge.workerShares = nil
	} else {
		shares := cloneShares(ge.workerShares)
		delete(shares, domain)
		ge.workerShares = cleanSharesIn(ge.rules, shares)
	}
	return ge.applySharesLocked(domain), nil
}

// applySharesLocked moves workers to match changed shares, staffs and
// recruits, and returns the reply saying what the shares are now and what
// moved. changed is the domain the command named ("" for all).
func (ge *GameEngine) applySharesLocked(changed string) ShareReply {
	ge.staffHoldUntil = 0
	sv := ge.newStaffView()
	slots := 0
	if i := domainIndex(ge.rules, changed); i >= 0 {
		slots = sv.doms[i].slots
	}
	moved := ge.rebalanceShares(sv)
	c := ge.staffByShares(true)
	if moved > 0 || c.any() {
		ge.recalculateRates()
	}
	line := "Worker shares: " + sharesLine(ge.rules, ge.workerShares) + "."
	if moved > 0 {
		line += " Moved " + textfmt.Count(moved, "worker", "workers") + " to match."
	}
	if c.any() {
		line += " " + textfmt.Capitalize(c.describe(ge.Workers.TotalPop(), ge.popCapLocked())) + "."
	}
	total := 0.0
	for _, d := range ge.rules.WorkerDomains() {
		total += ge.workerShares[d]
	}
	if total = math.Round(total*10) / 10; total > 100 {
		line += fmt.Sprintf(" The shares add up to %s, so each is scaled down to fit.", SharePercent(total))
	}
	if p, ok := ge.workerShares[changed]; ok && p > 0 && slots == 0 {
		line += fmt.Sprintf(" You have no %s buildings yet; the share applies once you build one.", strings.ToLower(DomainName(changed)))
	}
	r := ShareReply{Line: line}
	if p, ok := ge.workerShares["food"]; ok && p <= 0 {
		r.Line += " Food buildings get no workers now: watch your food."
		r.Warning = true
	}
	return r
}

// domainIndex is key's place in set's worker domains, -1 if it is none.
func domainIndex(set *rules.Set, key string) int {
	for i, d := range set.WorkerDomains() {
		if d == key {
			return i
		}
	}
	return -1
}

// sharesLine names the shares: "Knowledge 40%, Food 25%, the rest on auto",
// "all on auto".
func sharesLine(set *rules.Set, shares map[string]float64) string {
	if len(shares) == 0 {
		return "all on auto (workers follow your buildings' slots)"
	}
	var parts []string
	for _, d := range set.WorkerDomains() {
		if p, ok := shares[d]; ok {
			parts = append(parts, DomainName(d)+" "+SharePercent(p))
		}
	}
	if len(shares) < len(set.WorkerDomains()) {
		parts = append(parts, "the rest on auto")
	}
	return strings.Join(parts, ", ")
}

// ===== Workers panel =====

// ShareRow is one worker domain in the Workers panel's shares table.
type ShareRow struct {
	Domain  string
	Name    string
	Percent float64 // effective share of the workforce, 0 to 100
	Set     bool    // the player set it; otherwise it is on auto
	Slots   int     // worker slots in the domain's built buildings
	Workers int     // workers in them
}

// ShareRows is the shares table for a snapshot: each domain with worker slots
// or a share set, in domain order, with the share it gets.
func ShareRows(st GameState) []ShareRow {
	domains := st.Ruleset().WorkerDomains()
	doms := make([]shareDomain, len(domains))
	idx := make(map[string]int, len(domains))
	for i, d := range domains {
		doms[i].key = d
		idx[d] = i
	}
	for _, bs := range st.Buildings {
		if i, ok := idx[bs.WorkerDomain]; ok && bs.WorkerCapacity > 0 && bs.Count > 0 {
			doms[i].slots += bs.WorkerCapacity * bs.Count
			doms[i].assigned += bs.WorkersAssigned
		}
	}
	shareWeightsFor(doms, st.Workers.Shares)
	var rows []ShareRow
	for _, d := range doms {
		if d.slots == 0 && !d.set {
			continue
		}
		rows = append(rows, ShareRow{Domain: d.key, Name: DomainName(d.key), Percent: float64(d.weight * 100),
			Set: d.set, Slots: d.slots, Workers: d.assigned})
	}
	return rows
}

// Recruit statuses, for the Workers panel (RecruitStatus).
const (
	RecruitActive  = ""        // recruiting as slots open
	RecruitOff     = "off"     // auto-recruit is off
	RecruitHeld    = "held"    // waiting after a worker command
	RecruitHousing = "housing" // no housing left
	RecruitSlots   = "slots"   // every worker slot is filled or has an idle worker for it
	RecruitFood    = "food"    // the food income won't feed another worker
	RecruitStarve  = "starve"  // workers are starving
)

// RecruitStatus says whether the routine can recruit right now, and if not
// why, from a snapshot.
func RecruitStatus(st GameState) string {
	ws := st.Workers
	switch {
	case !ws.AutoRecruit:
		return RecruitOff
	case ws.HoldTicks > 0:
		return RecruitHeld
	case ws.MaxPop-ws.TotalPop <= 0:
		return RecruitHousing
	}
	free := 0
	for _, r := range ShareRows(st) {
		if !(r.Set && r.Percent <= 0) {
			free += max(0, r.Slots-r.Workers)
		}
	}
	if free-ws.TotalIdle <= 0 {
		return RecruitSlots
	}
	food := st.Resources["food"]
	if food.Amount <= 0 {
		return RecruitStarve
	}
	per := 0.0
	if cls, ok := st.Ruleset().WorkerClass("food", st.Age); ok {
		per = cls.FoodCost
	}
	net := foodBeforeMastery(food.Rate, food.Breakdown) - ws.FoodDrain
	if net-per < recruitFoodMargin(net, ws.FoodDrain, per) {
		return RecruitFood
	}
	return RecruitActive
}
