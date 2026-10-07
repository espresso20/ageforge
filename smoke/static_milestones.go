package smoke

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

// The Milestone Covenant: every milestone, every chain and the title ladder
// can be completed. It is proven from config alone with the Gate Covenant's
// model: the most storage buildable in an age (MaxStorage), undiscounted
// prices, and the cold start for what an age can supply.
//
// Milestones reset with the run, so a milestone is due by the end of the last
// age a normal run plays through (the age before game.PrestigeRunAge, the Modern Age), or by
// the end of its own MinAge when that is later: a milestone that names a
// later age is deep-run content and must be doable in that age. Every
// requirement must fit by then:
//
//   - a building count: a building can only be built in its own age (the age
//     lock), so the last required copy must cost no more than the most storage
//     buildable in that age, in every resource it costs, and stay within its
//     max count. A building sum fits the sum of its buildings' ceilings.
//   - a population: at most MilestonePopShare of the housing ceiling, which is
//     every housing building of every age so far raised to its storage limit
//     and never upgraded, plus the techs' housing.
//   - structures built in the run: at most MilestoneBuildShare of the build
//     ceiling, which is every building of every age so far raised to its
//     storage limit, each copy built once. Selling and rebuilding (the
//     counter counts those too) and rebuilding after a catastrophe are churn,
//     not play, so they are left out.
//   - a resource amount: fits the most storage buildable with
//     GateResourceMargin to spare, in an age from MinAge on that supplies it.
//   - a tech count or named techs: researchable by then, counting
//     prerequisites and the knowledge storage each price needs.
//   - wonders: built by then, one per age, each part fitting one full store
//     (the Gate Covenant's wonder rule).
//   - knowledge workers: at most MilestonePopShare of the worker slots in
//     knowledge buildings (every copy at its storage limit), and no more
//     than housing holds.
//   - soldiers trained: made by then at a moderate soldier income
//     (config.FlowIncome, the Gate Covenant's flow model) over each age's
//     pacing target.
//   - play time (MinTick) always passes, so it never limits anything.
//
// Where the model simplifies, it errs low: build_cost discounts, the refund
// an upgrade gets, and upgrades that add copies of a building after its own
// age can only raise what is possible. A failure names the number that
// breaks.
const (
	// MilestonePopShare is the most of the housing ceiling a population
	// milestone may ask for. The ceiling keeps every old housing tier
	// standing at its storage limit. A player upgrades old tiers and seldom
	// maxes storage, and the newest tier alone is about half the ceiling.
	MilestonePopShare = 0.4
	// MilestoneBuildShare is the most of the build ceiling a structures-built
	// milestone may ask for: half of everything a run can hold.
	MilestoneBuildShare = 0.5
)

// MilestoneProblem is a milestone, chain or title the static check proves
// out of reach, with the reason in plain words.
type MilestoneProblem struct {
	// Key is the milestone key, the chain key (Kind "chain") or the title
	// (Kind "title").
	Key string `json:"key"`
	// Kind is the requirement that fails: "age", "building",
	// "building_sum", "population", "builds", "techs", "tech", "resource",
	// "wonders", "knowledge_workers", "soldiers", "unknown", "chain" or
	// "title".
	Kind string `json:"kind"`
	Why  string `json:"why"`
}

// MilestoneReach is one milestone's verdict: the first age where every
// requirement fits, the age it is due by, and the requirement closest to
// its limit at that age.
type MilestoneReach struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	// Earliest is "" when some requirement never fits.
	Earliest string `json:"earliest,omitempty"`
	DueBy    string `json:"due_by"`
	// Tightest names the requirement with the highest Need/Limit at DueBy.
	// Empty for a milestone that only asks for an age or play time.
	Tightest string  `json:"tightest,omitempty"`
	Need     float64 `json:"need,omitempty"`
	Limit    float64 `json:"limit,omitempty"`
}

// Share is Need / Limit: +Inf when a requirement has nothing to fit in, 0
// when there is none.
func (r MilestoneReach) Share() float64 {
	switch {
	case r.Need <= 0:
		return 0
	case r.Limit <= 0:
		return math.Inf(1)
	}
	return r.Need / r.Limit
}

// StaticMilestones checks every milestone, chain and title against the
// Milestone Covenant. It returns what breaks it and every milestone's
// verdict, in config order.
func StaticMilestones() ([]MilestoneProblem, []MilestoneReach) {
	return staticMilestones(config.Milestones(), config.MilestoneChains(), config.MilestoneTitles(), config.BuildingByKey(), game.PrestigeRunAge)
}

// staticMilestones is StaticMilestones over the given data, with prestige
// opening at prestigeAge (the tests move it and break the numbers).
func staticMilestones(ms []config.MilestoneDef, chains []config.MilestoneChainDef, titles []config.TitleDef, defs map[string]config.BuildingDef, prestigeAge string) ([]MilestoneProblem, []MilestoneReach) {
	m := newMilestoneModel(defs, prestigeAge)
	var out []MilestoneProblem
	reach := make([]MilestoneReach, 0, len(ms))
	byKey := map[string]MilestoneReach{}
	names := map[string]string{}
	failed := map[string]bool{}
	for _, def := range ms {
		r, probs := m.check(def)
		out = append(out, probs...)
		reach = append(reach, r)
		byKey[def.Key] = r
		names[def.Key] = def.Name
		failed[def.Key] = len(probs) > 0
	}
	for _, c := range chains {
		var blocked []string
		for _, k := range c.MilestoneKeys {
			if _, ok := byKey[k]; !ok {
				out = append(out, MilestoneProblem{Key: c.Key, Kind: "chain",
					Why: fmt.Sprintf("%s (%s) lists the milestone %q, which does not exist.", c.Name, c.Key, k)})
				continue
			}
			if failed[k] {
				blocked = append(blocked, names[k])
			}
		}
		if len(blocked) > 0 {
			out = append(out, MilestoneProblem{Key: c.Key, Kind: "chain",
				Why: fmt.Sprintf("%s (%s) can't be finished, so its title %q is out of reach: %s %s.",
					c.Name, c.Key, c.Title, textfmt.List(blocked), textfmt.Plural(len(blocked), "fails its check", "fail their checks"))})
		}
	}
	ok := 0
	for _, def := range ms {
		if !failed[def.Key] {
			ok++
		}
	}
	for _, t := range titles {
		if t.MinMilestones > ok {
			out = append(out, MilestoneProblem{Key: t.Title, Kind: "title",
				Why: fmt.Sprintf("The title %q needs %d milestones, but only %d of the %d can be completed.", t.Title, t.MinMilestones, ok, len(ms))})
		}
	}
	return out, reach
}

// milestoneModel holds the ceilings the checks compare against, from config.
type milestoneModel struct {
	ages   []config.AgeDef
	idx    map[string]int
	defs   map[string]config.BuildingDef
	cold   []*coldStart
	runEnd int // the last age a normal run plays through

	store  map[[2]string]float64 // MaxStorage by age and resource
	copies map[string]copyCeiling
	unlock map[string]int // building: the first age that unlocks it
	resAge map[string]int // resource: the age it unlocks in
	techAt map[string]int // tech: the first age it can be researched (len(ages): never)

	// By age, counting everything up to and including it.
	builds, techs, wonders   []int
	housing, staff, soldiers []float64
}

// copyCeiling is how many copies of one building can ever stand.
type copyCeiling struct {
	age int // the only age it can be built in; -1 when never
	n   int // the most copies that can stand
	// What stops copy n+1: a resource that runs over storage, with the
	// copy's price and that storage, or "" for the max count.
	res          string
	price, store float64
}

// copyLimit caps the copy count of a building that nothing else caps.
const copyLimit = 100000

// newMilestoneModel builds the ceilings from defs; a run ends just before
// prestigeAge.
func newMilestoneModel(defs map[string]config.BuildingDef, prestigeAge string) *milestoneModel {
	ages := config.Ages()
	n := len(ages)
	m := &milestoneModel{ages: ages, idx: map[string]int{}, defs: defs, runEnd: n - 1,
		store: map[[2]string]float64{}, copies: map[string]copyCeiling{},
		unlock: map[string]int{}, resAge: map[string]int{}, techAt: map[string]int{},
		builds: make([]int, n), techs: make([]int, n), wonders: make([]int, n),
		housing: make([]float64, n), staff: make([]float64, n), soldiers: make([]float64, n)}
	for i, a := range ages {
		m.idx[a.Key] = i
		for _, k := range a.UnlockBuildings {
			if _, seen := m.unlock[k]; !seen {
				m.unlock[k] = i
			}
		}
	}
	if p, ok := m.idx[prestigeAge]; ok && p > 0 {
		m.runEnd = p - 1
	}
	for _, r := range config.BaseResources() {
		m.resAge[r.Key] = m.idx[r.Age]
	}
	m.cold = coldStarts(ages, defs)

	for _, k := range sortedKeys(defs) {
		d := defs[k]
		c := m.ceiling(k)
		if c.age < 0 {
			continue
		}
		house := 0.0
		for _, e := range d.Effects {
			if e.Type == "capacity" && e.Target == "population" {
				house += e.Value
			}
		}
		for i := c.age; i < n; i++ {
			m.builds[i] += c.n
			if d.Category == "wonder" && c.n > 0 {
				m.wonders[i]++
			}
			m.housing[i] += float64(house * float64(c.n))
			if d.WorkerDomain == "knowledge" {
				m.staff[i] += float64(d.WorkerCapacity * c.n)
			}
		}
	}

	techs := config.Technologies()
	byKey := make(map[string]config.TechDef, len(techs))
	for _, t := range techs {
		byKey[t.Key] = t
	}
	for _, t := range techs {
		for i := m.techAge(t.Key, byKey, 0); i < n; i++ {
			m.techs[i]++
			for _, e := range t.Effects {
				if e.Kind == config.EffectFlatHousing {
					m.housing[i] += e.Value
				}
			}
		}
	}

	made := 0.0
	for i, a := range ages {
		if i >= m.resAge["soldiers"] {
			made += float64(rules.Core().FlowIncome("soldiers", a.Key) * config.AgeTargetTicks(a.Key))
		}
		m.soldiers[i] = made
	}
	return m
}

// maxStorage is MaxStorage over the model's buildings, cached.
func (m *milestoneModel) maxStorage(i int, res string) float64 {
	k := [2]string{m.ages[i].Key, res}
	v, ok := m.store[k]
	if !ok {
		v = maxStorageIn(m.defs, m.ages[i].Key, res)
		m.store[k] = v
	}
	return v
}

// ceiling is how many copies of building k can ever stand: copies are built
// only in the building's own age, so each must fit the most storage
// buildable there. A wonder is 1 when each part of its price fits one store.
func (m *milestoneModel) ceiling(k string) copyCeiling {
	if c, ok := m.copies[k]; ok {
		return c
	}
	d := m.defs[k]
	c := copyCeiling{age: -1}
	if u, ok := m.unlock[k]; ok {
		if d.RequiredAge == "" {
			c.age = u // exempt from the age lock: buildable from its unlock on
		} else if a, known := m.idx[d.RequiredAge]; known && u <= a {
			c.age = a
		}
	}
	switch {
	case c.age < 0:
		// never buildable: no copies
	case d.Category == "wonder":
		c.n = 1
		for _, res := range sortedKeys(d.BaseCost) {
			if s := m.maxStorage(c.age, res); d.BaseCost[res]*GateWonderMargin > s {
				c.n, c.res, c.price, c.store = 0, res, d.BaseCost[res], s
				break
			}
		}
	default:
		for c.n < copyLimit && (d.MaxCount <= 0 || c.n < d.MaxCount) {
			res, price, store := m.hardest(d, c.n+1, c.age)
			if price > store {
				c.res, c.price, c.store = res, price, store
				break
			}
			c.n++
		}
	}
	m.copies[k] = c
	return c
}

// hardest is the resource copy n of d comes closest to running out of
// storage in, in age i: the copy's undiscounted price in it and the most
// storage buildable for it.
func (m *milestoneModel) hardest(d config.BuildingDef, n, i int) (res string, price, store float64) {
	best := math.Inf(1)
	for _, r := range sortedKeys(d.BaseCost) {
		p := lastCopyPrice(d, r, n)
		if p <= 0 {
			continue
		}
		s := m.maxStorage(i, r)
		if s/p < best {
			best, res, price, store = s/p, r, p, s
		}
	}
	return res, price, store
}

// techAge is the first age tech k can be researched in: its own age, after
// its prerequisites, and once the most knowledge storage holds its price.
// len(ages) means never.
func (m *milestoneModel) techAge(k string, byKey map[string]config.TechDef, depth int) int {
	if a, ok := m.techAt[k]; ok {
		return a
	}
	n := len(m.ages)
	t, ok := byKey[k]
	if !ok || depth > len(byKey) {
		return n // unknown, or a prerequisite loop
	}
	a, ok := m.idx[t.Age]
	if !ok {
		a = n
	}
	for _, p := range t.Prerequisites {
		if pa := m.techAge(p, byKey, depth+1); pa > a {
			a = pa
		}
	}
	for a < n && t.Cost > m.maxStorage(a, "knowledge") {
		a++
	}
	m.techAt[k] = a
	return a
}

// supplied reports whether age i can supply res at all (the Gate Covenant's
// cold start: its own buildings, the market, a carried supply).
func (m *milestoneModel) supplied(i int, res string) bool {
	cs := m.cold[i]
	return cs.unlocked[res] && (cs.reach[res] || cs.boot[res])
}

func (m *milestoneModel) ageName(i int) string { return game.AgeName(m.ages[i].Key) }

// due says what "due" means for the age a milestone is due by.
func (m *milestoneModel) due(i int) string {
	if i == m.runEnd {
		return fmt.Sprintf("by the end of the %s (the last age a run plays before prestige opens)", m.ageName(i))
	}
	return fmt.Sprintf("by the end of the %s (its own age)", m.ageName(i))
}

// firstAge is the first age from `from` on where limit(i) >= need, or -1.
func (m *milestoneModel) firstAge(from int, need float64, limit func(int) float64) int {
	for i := from; i < len(m.ages); i++ {
		if limit(i) >= need {
			return i
		}
	}
	return -1
}

// milestoneReq is one requirement of a milestone.
type milestoneReq struct {
	kind  string
	what  string // the report's name for it: "Stone Pits", "population"
	need  float64
	first int                  // the first age it fits; -1 when never
	limit func(int) float64    // the most it may ask for at an age; nil: no limit to show
	why   func(due int) string // why it doesn't fit by due (first < 0 or first > due)
}

// check proves one milestone.
func (m *milestoneModel) check(def config.MilestoneDef) (MilestoneReach, []MilestoneProblem) {
	due, first := m.runEnd, 0
	var probs []MilestoneProblem
	fail := func(kind, why string) {
		probs = append(probs, MilestoneProblem{Key: def.Key, Kind: kind, Why: fmt.Sprintf("%s (%s) %s", def.Name, def.Key, why)})
	}
	if def.MinAge != "" {
		a, ok := m.idx[def.MinAge]
		if !ok {
			fail("age", fmt.Sprintf("asks for the age %q, which does not exist.", def.MinAge))
			return MilestoneReach{Key: def.Key, Name: def.Name, DueBy: m.ages[due].Key}, probs
		}
		first = a
		due = max(due, a)
	}
	r := MilestoneReach{Key: def.Key, Name: def.Name, DueBy: m.ages[due].Key}
	never := false
	for _, q := range m.requirements(def, first) {
		if q.first < 0 || q.first > due {
			fail(q.kind, q.why(due))
		}
		never = never || q.first < 0
		first = max(first, q.first)
		if q.limit == nil {
			continue
		}
		cur := MilestoneReach{Need: q.need, Limit: q.limit(due)}
		if r.Tightest == "" || cur.Share() > r.Share() {
			r.Tightest, r.Need, r.Limit = q.what, cur.Need, cur.Limit
		}
	}
	if !never {
		r.Earliest = m.ages[first].Key
	}
	return r, probs
}

// requirements lists every requirement of def but its age and play time.
// from is the MinAge's index (0 without one).
func (m *milestoneModel) requirements(def config.MilestoneDef, from int) []milestoneReq {
	var reqs []milestoneReq
	for _, k := range sortedKeys(def.MinBuildings) {
		reqs = append(reqs, m.buildingReq(k, def.MinBuildings[k]))
	}
	if def.MinBuildingSum.Count > 0 || len(def.MinBuildingSum.Keys) > 0 {
		reqs = append(reqs, m.buildingSumReq(def.MinBuildingSum))
	}
	if def.MinPopulation > 0 {
		reqs = append(reqs, m.shareReq("population", "population", float64(def.MinPopulation), MilestonePopShare, m.housing,
			"a population of %s", "housing tops out at %s", "every housing building %s raised to its storage limit and never upgraded", "a population milestone"))
	}
	if def.MinTechCount > 0 {
		reqs = append(reqs, m.countReq("techs", "technologies", def.MinTechCount, m.techs, "can be researched"))
	}
	for _, k := range def.RequiredTechs {
		reqs = append(reqs, m.techReq(k))
	}
	for _, res := range sortedKeys(def.MinResources) {
		reqs = append(reqs, m.resourceReq(res, def.MinResources[res], from))
	}
	if def.MinTotalBuilt > 0 {
		reqs = append(reqs, m.shareReq("builds", "structures built", float64(def.MinTotalBuilt), MilestoneBuildShare, intsToFloats(m.builds),
			"%s structures built in one run", "building every structure once, up to its storage limit, makes %s", "every building %s, each copy built once", "a build-count milestone"))
	}
	if def.MinSoldiersTrained > 0 {
		reqs = append(reqs, m.soldiersReq(def.MinSoldiersTrained))
	}
	if def.MinWonders > 0 {
		reqs = append(reqs, m.countReq("wonders", "wonders", def.MinWonders, m.wonders, "can be built (one per age)"))
	}
	if def.MinKnowledgeWorkers > 0 {
		slots := make([]float64, len(m.ages))
		for i := range slots {
			slots[i] = math.Min(m.staff[i], m.housing[i])
		}
		reqs = append(reqs, m.shareReq("knowledge_workers", "knowledge workers", float64(def.MinKnowledgeWorkers), MilestonePopShare, slots,
			"%s knowledge workers", "knowledge buildings (capped by housing) hold %s workers", "every knowledge building %s raised to its storage limit", "a staffing milestone"))
	}
	return reqs
}

func intsToFloats(v []int) []float64 {
	out := make([]float64, len(v))
	for i, x := range v {
		out[i] = float64(x)
	}
	return out
}

// unknownReq is a requirement naming a key that does not exist.
func unknownReq(what, key string) milestoneReq {
	return milestoneReq{kind: "unknown", what: what, first: -1,
		why: func(int) string { return fmt.Sprintf("asks for the %s %q, which does not exist.", what, key) }}
}

// buildingReq: need copies of building k standing.
func (m *milestoneModel) buildingReq(k string, need int) milestoneReq {
	d, ok := m.defs[k]
	if !ok {
		return unknownReq("building", k)
	}
	c := m.ceiling(k)
	count := game.BuildingCount(need, k)
	q := milestoneReq{kind: "building", what: pluralName(k), need: float64(need), first: c.age,
		limit: func(i int) float64 {
			if c.age < 0 || c.age > i {
				return 0
			}
			return float64(c.n)
		}}
	switch {
	case c.age < 0:
		q.first = -1
		q.why = func(int) string {
			return fmt.Sprintf("needs %s, but %s is never unlocked in time to be built (it can only be built in the %s).", count, d.Name, game.AgeName(d.RequiredAge))
		}
	case need > c.n && c.res == "":
		q.first = -1
		q.why = func(int) string {
			return fmt.Sprintf("needs %s, but at most %d may exist (its max count).", count, c.n)
		}
	case need > c.n && d.Category == "wonder":
		q.first = -1
		q.why = func(int) string {
			return fmt.Sprintf("needs %s, but it costs %s, over the most %s storage buildable in the %s (%s).", count, num(c.price), game.ResourceName(c.res), m.ageName(c.age), num(c.store))
		}
	case need > c.n:
		q.first = -1
		res, price, store := m.hardest(d, need, c.age)
		q.why = func(int) string {
			return fmt.Sprintf("needs %s, but %s can only be built in the %s, and copy #%d costs %s %s, over the most %s storage buildable there (%s). At most %d can ever stand.",
				count, pluralName(k), m.ageName(c.age), need, num(price), game.ResourceName(res), game.ResourceName(res), num(store), c.n)
		}
	default:
		q.why = func(due int) string {
			return fmt.Sprintf("needs %s %s, but %s can only be built in the %s.", count, m.due(due), pluralName(k), m.ageName(c.age))
		}
	}
	return q
}

// buildingSumReq: sum.Count buildings across sum.Keys, in any mix.
func (m *milestoneModel) buildingSumReq(sum config.BuildingSum) milestoneReq {
	var names []string
	for _, k := range sum.Keys {
		if _, ok := m.defs[k]; !ok {
			return unknownReq("building", k)
		}
		names = append(names, pluralName(k))
	}
	what := textfmt.List(names)
	if len(sum.Keys) == 0 || sum.Count <= 0 {
		return milestoneReq{kind: "building_sum", what: what, first: -1,
			why: func(int) string { return "has a building sum without both buildings and a count." }}
	}
	standing := func(i int) float64 {
		n := 0
		for _, k := range sum.Keys {
			if c := m.ceiling(k); c.age >= 0 && c.age <= i {
				n += c.n
			}
		}
		return float64(n)
	}
	need := float64(sum.Count)
	q := milestoneReq{kind: "building_sum", what: what, need: need, limit: standing, first: m.firstAge(0, need, standing)}
	q.why = func(due int) string {
		var parts []string
		for _, k := range sum.Keys {
			c := m.ceiling(k)
			if c.age < 0 {
				parts = append(parts, fmt.Sprintf("no %s (never unlocked in time)", pluralName(k)))
				continue
			}
			parts = append(parts, fmt.Sprintf("%s in the %s", game.BuildingCount(c.n, k), m.ageName(c.age)))
		}
		last := len(m.ages) - 1
		if q.first < 0 {
			return fmt.Sprintf("needs %d %s between them, but at most %s can ever stand (%s).", sum.Count, what, grouped(int(standing(last))), textfmt.List(parts))
		}
		return fmt.Sprintf("needs %d %s between them %s, but only %s can stand by then (%s).", sum.Count, what, m.due(due), grouped(int(standing(due))), textfmt.List(parts))
	}
	return q
}

// shareReq: need of something whose ceiling at age i is ceil[i], of which a
// milestone may ask for at most share. The format pieces fill in the plain
// words: needF takes the need, ceilF the ceiling, basisF where the ceiling
// comes from ("so far" or "of every age"), and who names the milestone kind.
func (m *milestoneModel) shareReq(kind, what string, need, share float64, ceil []float64, needF, ceilF, basisF, who string) milestoneReq {
	limit := func(i int) float64 { return float64(share * ceil[i]) }
	q := milestoneReq{kind: kind, what: what, need: need, limit: limit, first: m.firstAge(0, need, limit)}
	pct := strconv.Itoa(int(math.Round(share * 100)))
	needText := fmt.Sprintf(needF, grouped(int(need)))
	q.why = func(due int) string {
		last := len(m.ages) - 1
		if q.first < 0 {
			return fmt.Sprintf("needs %s, but %s in the whole game (%s), and %s may ask for at most %s%% of that (%s).",
				needText, fmt.Sprintf(ceilF, grouped(int(ceil[last]))), fmt.Sprintf(basisF, "of every age"), who, pct, grouped(int(limit(last))))
		}
		return fmt.Sprintf("needs %s %s, but by then %s (%s), and %s may ask for at most %s%% of that (%s). The first age where it fits is the %s.",
			needText, m.due(due), fmt.Sprintf(ceilF, grouped(int(ceil[due]))), fmt.Sprintf(basisF, "so far"), who, pct, grouped(int(limit(due))), m.ageName(q.first))
	}
	return q
}

// countReq: need of something counted per age (techs, wonders).
func (m *milestoneModel) countReq(kind, what string, need int, have []int, verb string) milestoneReq {
	limit := func(i int) float64 { return float64(have[i]) }
	q := milestoneReq{kind: kind, what: what, need: float64(need), limit: limit, first: m.firstAge(0, float64(need), limit)}
	q.why = func(due int) string {
		if q.first < 0 {
			return fmt.Sprintf("needs %d %s, but only %d %s in the whole game.", need, what, have[len(have)-1], verb)
		}
		return fmt.Sprintf("needs %d %s %s, but only %d %s by then. The first age with %d is the %s.", need, what, m.due(due), have[due], verb, need, m.ageName(q.first))
	}
	return q
}

// techReq: tech k researched.
func (m *milestoneModel) techReq(k string) milestoneReq {
	a, ok := m.techAt[k]
	if !ok {
		return unknownReq("technology", k)
	}
	name := game.TechName(k)
	q := milestoneReq{kind: "tech", what: name, first: a}
	if a >= len(m.ages) {
		q.first = -1
		q.why = func(int) string { return fmt.Sprintf("needs the technology %s, which can never be researched.", name) }
		return q
	}
	q.why = func(due int) string {
		return fmt.Sprintf("needs the technology %s %s, but it can't be researched before the %s.", name, m.due(due), m.ageName(a))
	}
	return q
}

// resourceReq: amount of res held at once, in an age from `from` on.
func (m *milestoneModel) resourceReq(res string, amount float64, from int) milestoneReq {
	unlock, ok := m.resAge[res]
	if !ok {
		return unknownReq("resource", res)
	}
	name := game.ResourceName(res)
	start := max(from, unlock)
	// The amount has to be held once, in any age from start to i that
	// supplies it; storage only grows, so the latest such age is the best.
	limit := func(i int) float64 {
		for a := i; a >= start; a-- {
			if m.supplied(a, res) {
				return m.maxStorage(a, res) / GateResourceMargin
			}
		}
		return 0
	}
	q := milestoneReq{kind: "resource", what: name, need: amount, limit: limit, first: m.firstAge(start, amount, limit)}
	q.why = func(due int) string {
		hold := fmt.Sprintf("to hold %s", game.Amount(amount, res))
		last := len(m.ages) - 1
		if q.first < 0 {
			if l := limit(last); l > 0 {
				return fmt.Sprintf("needs %s, but the most %s storage buildable (%s) holds less than %gx that.", hold, name, num(float64(l*GateResourceMargin)), GateResourceMargin)
			}
			return fmt.Sprintf("needs %s, but nothing supplies %s from the %s on.", hold, name, m.ageName(start))
		}
		if l := limit(due); l > 0 {
			return fmt.Sprintf("needs %s %s, but the most %s storage buildable by then (%s) holds less than %gx that. It first fits in the %s.",
				hold, m.due(due), name, num(float64(l*GateResourceMargin)), GateResourceMargin, m.ageName(q.first))
		}
		return fmt.Sprintf("needs %s %s, but nothing supplies %s by then. It first fits in the %s.", hold, m.due(due), name, m.ageName(q.first))
	}
	return q
}

// soldiersReq: need soldiers trained in the run.
func (m *milestoneModel) soldiersReq(need int) milestoneReq {
	limit := func(i int) float64 { return m.soldiers[i] }
	q := milestoneReq{kind: "soldiers", what: "soldiers trained", need: float64(need), limit: limit, first: m.firstAge(0, float64(need), limit)}
	q.why = func(due int) string {
		last := len(m.ages) - 1
		if q.first < 0 {
			return fmt.Sprintf("needs %s soldiers trained, but a moderate soldier income makes only %s in the whole game.", grouped(need), grouped(int(m.soldiers[last])))
		}
		return fmt.Sprintf("needs %s soldiers trained %s, but a moderate soldier income makes only %s by then. The first age where it fits is the %s.", grouped(need), m.due(due), grouped(int(m.soldiers[due])), m.ageName(q.first))
	}
	return q
}

// pluralName is a building's name for several copies ("Stone Pits").
func pluralName(k string) string {
	s := game.BuildingCount(2, k)
	return s[strings.IndexByte(s, ' ')+1:]
}

// grouped prints a whole count with thousands separators ("4,361").
func grouped(n int) string {
	s := strconv.Itoa(n)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		return "-" + s
	}
	return s
}

// writeMilestones renders the milestone feasibility check.
func writeMilestones(sb *strings.Builder, problems []MilestoneProblem, reach []MilestoneReach) {
	fmt.Fprintf(sb, "From config alone: every milestone must be completable by the end of the last age a run plays before prestige opens, or of its own age when that is later. A building count must fit the most storage buildable in the building's own age (the only age it can be built in), undiscounted, within its max count; a population at most %g%% of the housing ceiling (every housing building so far at its storage limit, never upgraded); structures built at most %g%% of the build ceiling (every building so far at its storage limit, built once); a resource the most storage buildable with %gx to spare; techs and wonders what can be researched and built by then; soldiers what a moderate income trains. Every chain and the title ladder must be completable too. `go test ./smoke` fails on any problem here.\n\n",
		MilestonePopShare*100, MilestoneBuildShare*100, GateResourceMargin)
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		for _, p := range problems {
			fmt.Fprintf(sb, "- %s\n", p.Why)
		}
		sb.WriteString("\n")
	}
	rows := make([]MilestoneReach, 0, len(reach))
	for _, r := range reach {
		if r.Tightest != "" {
			rows = append(rows, r)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Share() > rows[j].Share() })
	sb.WriteString("Every milestone with a count, tightest first (need / limit at the age it is due by):\n\n| milestone | earliest | due by | tightest requirement | need | limit | share |\n|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		earliest := "never"
		if r.Earliest != "" {
			earliest = game.AgeName(r.Earliest)
		}
		share := "no room"
		if sh := r.Share(); !math.IsInf(sh, 1) {
			share = fmt.Sprintf("%.0f%%", sh*100)
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s |\n", r.Key, earliest, game.AgeName(r.DueBy), r.Tightest, num(r.Need), num(r.Limit), share)
	}
}
