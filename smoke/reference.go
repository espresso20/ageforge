package smoke

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
)

// The reference players: three ways of playing, each a rule anyone can
// recompute from the game's tables. Nothing here runs the engine or a bot.
//
// Common to all three. Income each tick is what the building table gives for
// the copies owned, with the workforce spread evenly over every job, times
// the bonuses the tables give the reference town in that age
// (config.IncomeFactor). Materials convert into one another at the age's
// price levels, so a town's income and a price are both counted in price
// units. A wonder costs RefWonderUnits of them.
//
//   - The ordinary player buys the cheapest thing he can afford, only up to
//     the reference town (RefCopies of every building, or the gate's count
//     where the gate asks for more), then pays the wonder, and advances the
//     moment both are done.
//   - The lingering player buys the cheapest thing he can afford and that
//     fits under his stores, with no other limit, pays the wonder once the reference counts are met, and stays
//     RefLinger times as long as it took him to open the gate.
//   - The check-in player is always away: his income runs at RefAwayRate (the
//     offline rule), the plan buys the cheapest thing with no limit, and he
//     advances at the first visit, RefVisitHours apart, after the gate opens.
//
// An age's length for a player is the first moment the town has made what
// that player bought. It is an output of the tables, not a target.
const (
	RefCopies      = 5
	RefWonderUnits = 40.0
	RefLinger      = 3.0
	RefAwayRate    = 0.5
	RefVisitHours  = 8.0
	// RefTopMastery is the speed of an age at the top of the Era Mastery
	// table (1 + the square root of 10): production and storage are
	// multiplied by it and prices are not.
	RefTopMastery = 4.16
	// RefHandGather is the wood and food per tick a player gathers by hand
	// in the first age. It is an assumption of the reference players, not a
	// number from the tables: without it an empty town never starts.
	RefHandGather = 0.5
)

const refTicksPerHour = 3600.0 / config.TickSeconds

// refTables is everything the reference players read, gathered once.
type refTables struct {
	ages   []config.AgeDef
	idx    map[string]int
	byAge  [][]config.BuildingDef // an age's buildings, wonders and monuments aside
	levels []map[string]float64   // an age's price levels
	factor []map[string]float64   // the reference town's income bonus, by resource
	food   []float64              // food one worker eats per tick
	base   map[string]float64     // a resource's storage before any building
	techs  [][]float64            // an age's tech prices, cheapest first
	defs   map[string]config.BuildingDef
}

func newRefTables() *refTables {
	t := &refTables{ages: config.Ages(), idx: map[string]int{}, base: map[string]float64{}, defs: map[string]config.BuildingDef{}}
	order := make([]string, len(t.ages))
	for i, a := range t.ages {
		t.idx[a.Key] = i
		order[i] = a.Key
	}
	all := config.BaseBuildings()
	techs := config.Technologies()
	pos := config.AgePositions(order)
	lv := config.PriceLevelsByAge(all)
	resources := config.BaseResources()
	t.byAge = make([][]config.BuildingDef, len(t.ages))
	t.techs = make([][]float64, len(t.ages))
	for _, d := range all {
		i, ok := t.idx[d.RequiredAge]
		if !ok || d.Category == "wonder" || d.Category == "monument" {
			continue
		}
		t.byAge[i] = append(t.byAge[i], d)
		t.defs[d.Key] = d
	}
	for _, tc := range techs {
		if i, ok := t.idx[tc.Age]; ok && tc.Cost > 0 {
			t.techs[i] = append(t.techs[i], tc.Cost)
		}
	}
	for _, r := range resources {
		t.base[r.Key] = r.BaseStorage
	}
	for i, a := range t.ages {
		sort.Float64s(t.techs[i])
		t.levels = append(t.levels, lv[a.Key])
		f := map[string]float64{}
		for _, r := range resources {
			f[r.Key] = config.IncomeFactor(techs, pos, a.Key, r.Key)
		}
		t.factor = append(t.factor, f)
		fc := 0.1
		if c, ok := config.WorkerClassByDomainAndAge("food", a.Key); ok {
			fc = c.FoodCost
		}
		t.food = append(t.food, fc)
	}
	return t
}

// gate is what the game asks to leave age i.
func (t *refTables) gate(i int) (map[string]float64, map[string]int) {
	if i+1 >= len(t.ages) {
		return nil, nil
	}
	return t.ages[i+1].ResourceReqs, t.ages[i+1].BuildingReqs
}

// refCount is how many copies of d the reference town holds.
func (t *refTables) refCount(d config.BuildingDef) int {
	_, counts := t.gate(t.idx[d.RequiredAge])
	return max(RefCopies, counts[d.Key])
}

// refTown is the reference town standing in age i: every age so far at its
// reference counts.
func (t *refTables) refTown(i int) map[string]int {
	c := map[string]int{}
	for j := 0; j <= i; j++ {
		for _, d := range t.byAge[j] {
			c[d.Key] = t.refCount(d)
		}
	}
	return c
}

func copyPrice(d config.BuildingDef, owned int) map[string]float64 {
	f := detmath.Pow(d.CostScale, float64(owned))
	out := make(map[string]float64, len(d.BaseCost))
	for r, v := range d.BaseCost {
		out[r] = math.Max(1, math.Floor(float64(v*f)))
	}
	return out
}

// units counts amounts of several resources in the age's price units.
func units(vals, levels map[string]float64) float64 {
	keys := make([]string, 0, len(vals))
	for r := range vals {
		keys = append(keys, r)
	}
	sort.Strings(keys)
	u := 0.0
	for _, r := range keys {
		if l := levels[r]; l > 0 {
			u += vals[r] / l
		}
	}
	return u
}

// townState is what a town makes and holds.
type townState struct {
	income  map[string]float64
	caps    map[string]float64
	workers float64 // people at work
	room    float64 // people the houses hold
	jobs    float64
	eaten   float64 // food the workers eat per tick
	grown   float64 // food made per tick, before they eat
}

// state works out a town standing in age i. speed is the Era Mastery speed
// of the age: production and storage are multiplied by it.
func (t *refTables) state(counts map[string]int, i int, speed float64) townState {
	staffed, free := map[string]float64{}, map[string]float64{}
	capAll := 0.0
	own := map[string]float64{}
	s := townState{income: map[string]float64{}, caps: map[string]float64{}}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		n := float64(counts[k])
		d, ok := t.defs[k]
		if !ok || n <= 0 || t.idx[d.RequiredAge] > i {
			continue
		}
		s.jobs += float64(n * float64(d.WorkerCapacity))
		for _, e := range d.Effects {
			switch {
			case e.Type == "production" && d.WorkerCapacity > 0:
				staffed[e.Target] += float64(n * e.Value)
			case e.Type == "production":
				free[e.Target] += float64(n * e.Value)
			case e.Type == "capacity" && e.Target == "population":
				s.room += float64(n * e.Value)
			case e.Type == "storage" && e.Target == "all":
				capAll += float64(n * e.Value)
			case e.Type == "storage":
				own[e.Target] += float64(n * e.Value)
			}
		}
	}
	fac := t.factor[i]
	fc := t.food[i]
	s.workers = math.Min(s.room, s.jobs)
	fst, fun := float64(staffed["food"]*fac["food"]), float64(free["food"]*fac["food"])
	if s.jobs > 0 {
		// A town recruits no one it cannot feed.
		if net := fun + float64(fst*(0.2+float64(0.8*s.workers/s.jobs))) - float64(s.workers*fc); net < 0 {
			if den := fc - float64(0.8*fst/s.jobs); den > 0 {
				s.workers = math.Max(0, math.Min(s.workers, (fun+float64(0.2*fst))/den))
			}
		}
	}
	fill := 1.0
	if s.jobs > 0 {
		fill = s.workers / s.jobs
	}
	for r, v := range free {
		s.income[r] += float64(v * fac[r] * speed)
	}
	for r, v := range staffed {
		s.income[r] += float64(v * fac[r] * (0.2 + float64(0.8*fill)) * speed)
	}
	if i == 0 {
		// Gathering by hand, so a town with no buildings can start.
		s.income["wood"] += float64(RefHandGather * speed)
		s.income["food"] += float64(RefHandGather * speed)
	}
	s.grown = s.income["food"]
	s.eaten = float64(s.workers * fc * speed)
	s.income["food"] -= s.eaten
	for r, b := range t.base {
		s.caps[r] = float64((b + capAll + own[r]) * speed)
	}
	return s
}

// refRun is one stay in one age.
type refRun struct {
	counts  map[string]int
	end     townState
	ticks   float64   // how long the stay lasted
	gate    float64   // when the reference counts were met and the wonder paid
	wall    bool      // nothing more could be bought (only without overflow)
	doubles []float64 // ticks at which income first reached 2, 4, 8... times refIncome
}

// refPlay is how a stay is played.
type refPlay struct {
	only      map[string]int // buy only up to these counts (the ordinary player)
	wonder    bool           // pay the wonder once the reference counts are met
	overflow  bool           // a price over the store can still be bought (the plan banks overflow)
	rate      float64        // share of income earned (the away player earns RefAwayRate)
	speed     float64        // Era Mastery speed of the age
	leave     func(gate, now float64) bool
	refIncome float64 // income, in price units, that doublings are counted against
	horizon   float64
}

// stay plays age i from start. Time moves in steps that lengthen as the stay
// does, so a stay of a month costs no more than a stay of an hour.
func (t *refTables) stay(i int, start map[string]int, p refPlay) refRun {
	lv := t.levels[i]
	blds := t.byAge[i]
	counts := make(map[string]int, len(start)+len(blds))
	for k, v := range start {
		counts[k] = v
	}
	for _, d := range blds {
		if _, ok := counts[d.Key]; !ok {
			counts[d.Key] = 0
		}
	}
	ref := map[string]int{}
	for _, d := range blds {
		ref[d.Key] = t.refCount(d)
	}
	run := refRun{gate: -1}
	st := t.state(counts, i, p.speed)
	stock, wonder, now, next := 0.0, RefWonderUnits, 0.0, 2.0
	if !p.wonder {
		wonder = 0
	}
	for now < p.horizon {
		dt := math.Max(1, now/200)
		incU := float64(units(st.income, lv) * p.rate)
		capU := units(st.caps, lv)
		met := true
		for k, n := range ref {
			if counts[k] < n {
				met = false
				break
			}
		}
		gain := float64(incU * dt)
		if met && wonder > 0 {
			pay := math.Min(stock+gain, wonder)
			wonder -= pay
			stock += gain - pay
		} else {
			stock += gain
		}
		if !p.overflow {
			stock = math.Min(stock, capU)
		}
		now += dt
		feasible := false
		for !(met && wonder > 0) {
			best, bestCost := -1, 0.0
			blocked := false
			if p.only != nil && !p.overflow {
				for _, d := range blds {
					if d.Category != "storage" && counts[d.Key] < p.only[d.Key] && overCap(copyPrice(d, counts[d.Key]), st.caps) {
						blocked = true
						break
					}
				}
			}
			for bi, d := range blds {
				n := counts[d.Key]
				if d.MaxCount > 0 && n >= d.MaxCount {
					continue
				}
				if p.only != nil {
					limit := p.only[d.Key]
					if d.Category == "storage" && blocked {
						limit = max(limit, n+1)
					}
					if n >= limit {
						continue
					}
				}
				price := copyPrice(d, n)
				if !p.overflow && overCap(price, st.caps) {
					continue
				}
				feasible = true
				if u := units(price, lv); u <= stock && (best < 0 || u < bestCost) {
					best, bestCost = bi, u
				}
			}
			if best < 0 {
				break
			}
			stock -= bestCost
			counts[blds[best].Key]++
			st = t.state(counts, i, p.speed)
		}
		if p.refIncome > 0 {
			for units(st.income, lv)/p.refIncome >= next {
				run.doubles = append(run.doubles, now)
				next *= 2
			}
		}
		if run.gate < 0 && wonder <= 0 {
			done := true
			for k, n := range ref {
				if counts[k] < n {
					done = false
					break
				}
			}
			if done {
				run.gate = now
			}
		}
		if p.leave != nil && run.gate >= 0 && p.leave(run.gate, now) {
			break
		}
		if p.leave == nil && !p.overflow && !feasible {
			run.wall = true
			break
		}
	}
	run.counts, run.end, run.ticks = counts, st, now
	return run
}

func overCap(price, caps map[string]float64) bool {
	for r, v := range price {
		if v > caps[r] {
			return true
		}
	}
	return false
}

// refWalk is one player's path through every age: how long each stay lasted
// and what the town made on leaving, against the reference town's income.
type refWalk struct {
	Hours  []float64
	Income []float64
	// Runs is each age's stay as it ended: the town the player left with.
	Runs []refRun
}

// walk sends a player through all 22 ages at the given Era Mastery speed.
// overflow plays the rule the game had before the storage wall, where a
// price over the store could still be bought; the properties read the walk
// without it.
func (t *refTables) walk(player string, speed float64, refIncome []float64, overflow bool) refWalk {
	var w refWalk
	counts := map[string]int{}
	for i := range t.ages {
		p := refPlay{wonder: true, overflow: overflow, rate: 1, speed: speed, horizon: 4e7}
		switch player {
		case "ordinary":
			p.only = map[string]int{}
			for _, d := range t.byAge[i] {
				p.only[d.Key] = t.refCount(d)
			}
			p.leave = func(gate, now float64) bool { return true }
		case "lingering":
			p.leave = func(gate, now float64) bool { return now >= float64(RefLinger*gate) }
		case "check-in":
			p.rate = RefAwayRate
			visit := float64(RefVisitHours * refTicksPerHour)
			p.leave = func(gate, now float64) bool { return now >= float64(math.Ceil(gate/visit)*visit) }
		}
		run := t.stay(i, counts, p)
		counts = run.counts
		w.Hours = append(w.Hours, run.ticks/refTicksPerHour)
		w.Income = append(w.Income, units(run.end.income, t.levels[i])/refIncome[i])
		w.Runs = append(w.Runs, run)
	}
	return w
}
