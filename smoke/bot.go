package smoke

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// Bot is a greedy, legitimate player. It only calls the same GameEngine
// methods the command line reaches (build, recruit, assign/unassign, upgrade,
// research, wonder collect, festival, gather); no dev commands, no god mode, no
// speed changes. Flow control (advance, prestige, catastrophe choices) lives in
// the runner so it can check invariants around each transition.
//
// The strategy, per decision:
//   - keep food positive: food producers get workers first, and a food
//     building is bought whenever the net food rate goes negative
//   - buy the next age's required buildings as soon as they are affordable
//   - buy housing when the population is capped and worker slots are free
//   - while the slowest requirement is more than the horizon away, invest in
//     the best producer for it; otherwise save
//   - staff buildings by how much each worker moves the requirements forward
//   - prefer producers it can pay for within a few horizons over pricier
//     ones that need a resource nothing makes yet
//   - bank surplus into the age's wonder, research affordable techs with
//     surplus knowledge (a tech that makes a missing target resource
//     first), and hand-gather while the game allows it
//
// Every map walk goes through sorted keys so a seed replays the same decisions.
type Bot struct {
	ge *game.GameEngine
	// rules is the engine's ruleset, read once: what the bot knows of the
	// game's buildings, ages and techs.
	rules     *rules.Set
	defs      map[string]config.BuildingDef
	nameToKey map[string]string
	ageIdx    map[string]int
	gatherMax int // index of the last age that allows hand-gathering

	// HorizonTicks is how far away (in ticks) the slowest requirement may be
	// before the bot stops investing in production and starts saving.
	HorizonTicks float64
	// CheckInTicks is the time to the next decision for a player who only
	// checks in now and then (the idle style); 0 for one who is always
	// there. A check-in player plans for what happens before the next
	// visit: the buildings and techs the income of those hours pays for. See
	// CheckIn.
	CheckInTicks float64
	// Deals makes the bot take faction trade deals (takeDeals). Off by
	// default: the greedy bot ignores them, and -deals=on measures what they
	// are worth.
	Deals bool
	// Army makes the bot keep a modest garrison (keepGarrison): a few of the
	// age's newest military buildings, bought from spare stock. Off by
	// default: the greedy bot ignores the army beyond the buildings an age
	// gate asks for, and -army=on measures what a garrison is worth.
	Army bool
	// UsePlan makes a check-in player leave a build plan for the hours until
	// the next visit (planAhead). The idle style sets it; the greedy bot,
	// always there, has no use for one.
	UsePlan bool
	// RecordPlan writes every build the bot makes into the run's plan log
	// (GameEngine.NotePlanForTest), as if it had planned it: how the canned
	// veteran kit's template is made (Config.DumpLegacy).
	RecordPlan bool
	// UseShares makes a check-in player leave its workers to the game's
	// worker shares, on auto: the game recruits and staffs, between visits
	// too, instead of the bot recruiting and assigning by hand at each
	// visit. The idle style sets it (-no-shares turns it off). Setting
	// shares once per age (domains the age doesn't need at 0, or a lean
	// toward the slowest requirement's domain) measured no better than auto
	// on the idle targets, so the bot keeps the default players get.
	UseShares bool

	// Actions counts successful player actions by kind; Errors counts
	// rejected ones. Both feed the report.
	Actions map[string]int
	Errors  map[string]int
	// Refusals keeps what the first few rejected actions were told, so the
	// report can say what went wrong, not only how often (refusalsKept at
	// most: a bot that is refused every tick must not fill the report).
	Refusals []Refusal

	// Trace, if set, receives one line per attempted action.
	Trace io.Writer
	// Harbinger is the policy for harbinger answers: ignore, appease, brace
	// or both (Appease up to level 2, Brace level 1).
	Harbinger string
	tick      int
	age       string
}

// gatherYield matches the command line's per-command cap (ui gatherMaxYield).
// A tech that adds to hand gathering adds to each of these on top.
const gatherYield = 25.0

// NewBot returns a bot that plays ge.
func NewBot(ge *game.GameEngine) *Bot {
	set := ge.Rules()
	defs := set.BuildingMap()
	n2k := make(map[string]string, len(defs))
	for k, d := range defs {
		n2k[d.Name] = k
	}
	idx := make(map[string]int)
	for i, k := range set.AgeKeys() {
		idx[k] = i
	}
	return &Bot{
		ge:           ge,
		rules:        set,
		defs:         defs,
		nameToKey:    n2k,
		ageIdx:       idx,
		gatherMax:    idx["medieval_age"],
		HorizonTicks: 900, // 30 minutes at 1x
		Actions:      make(map[string]int),
		Errors:       make(map[string]int),
	}
}

// plan is the bot's working view of one decision.
type plan struct {
	st  game.GameState
	amt map[string]float64 // spendable amounts, decremented as the bot spends
	// stock is a good stock of each resource for the town (goodStock): what
	// the bot treats as a full store. Nothing caps stock, so it holds the
	// bot to nothing; it is the scale the bot weighs a pile by.
	stock    map[string]float64
	target   map[string]float64 // total amount still to generate this age
	eta      map[string]float64 // ticks until target, +Inf if unreachable at the current rate
	maxEta   float64
	worst    []string       // target resources, slowest first
	needBld  map[string]int // required building deficits (built + queued)
	queued   map[string]int // queued copies per key
	extra    map[string]int // copies bought this decision, for cost scaling
	invest   bool
	foodRate float64
	slots    int // worker slots across all buildings
	// freeFoodSlots is the unstaffed capacity of food producers.
	freeFoodSlots int
	// must is the techs the age cannot be left without that are still to
	// research, in the order to research them (mustTechs); hold is what the
	// ones not started yet cost, by resource: the bot saves it, and the
	// wonder bank does not take it.
	must []string
	hold map[string]float64
	// blocker is the resource keeping the most wanted purchase out of reach;
	// hand-gathering goes there first.
	blocker string
	// blockCost is that purchase's whole price, which a deal must not give away.
	blockCost map[string]float64
}

// goodStock is what the bot counts as a full store of res: a good stock of it
// for the town as it stands (rules.Set.StockOf): the age's stock scale in
// minutes of the town's own income of res. Stock has no ceiling, so nothing
// holds the bot to it; it is the scale the bot weighs a pile of res by.
func (b *Bot) goodStock(st game.GameState, res string) float64 {
	return goodStockOf(b.rules, st, res)
}

func (b *Bot) newPlan(st game.GameState) *plan {
	p := &plan{
		st:      st,
		amt:     make(map[string]float64),
		stock:   make(map[string]float64),
		target:  make(map[string]float64),
		eta:     make(map[string]float64),
		needBld: make(map[string]int),
		queued:  make(map[string]int),
		extra:   make(map[string]int),
		hold:    make(map[string]float64),
	}
	for k, r := range st.Resources {
		p.amt[k] = r.Amount
		p.stock[k] = b.goodStock(st, k)
	}
	p.foodRate = st.Resources["food"].Rate
	for key, bs := range st.Buildings {
		def := b.defs[key]
		c := bs.Count * def.WorkerCapacity
		p.slots += c
		if c > bs.WorkersAssigned && b.perWorkerYield(key)["food"] > 0 {
			p.freeFoodSlots += c - bs.WorkersAssigned
		}
	}
	for _, q := range st.BuildQueue {
		if k, ok := b.nameToKey[q.Name]; ok {
			p.queued[k]++
		}
	}
	// Sorted: several buildings add to one resource's target, and a float
	// sum in map order moves the last bit from run to run.
	for _, bld := range sortedKeys(st.NextAgeBldReqs) {
		n := st.NextAgeBldReqs[bld]
		have := st.Buildings[bld].Count + p.queued[bld]
		if have < n {
			p.needBld[bld] = n - have
			for res, c := range st.Buildings[bld].NextCost {
				p.target[res] += float64(c * float64(n-have))
			}
		}
	}
	if w := st.CurrentAgeWonderKey; w != "" {
		bank := st.Buildings[w].WonderBank
		for res, c := range b.defs[w].BaseCost {
			if left := c - bank[res]; left > 0 {
				p.target[res] += left
			}
		}
	}
	// The techs the age cannot be left without. A tech has no price: the
	// scholars do its work.
	p.must = b.mustTechs(p)
	b.harbingerTargets(p)
	p.maxEta = 0
	for _, res := range sortedKeys(p.target) {
		deficit := p.target[res] - p.amt[res]
		switch {
		case deficit <= 0:
			p.eta[res] = 0
		case st.Resources[res].Rate <= 0:
			p.eta[res] = math.Inf(1)
		default:
			p.eta[res] = deficit / st.Resources[res].Rate
		}
		if p.eta[res] > 0 {
			p.worst = append(p.worst, res)
		}
		if p.eta[res] > p.maxEta {
			p.maxEta = p.eta[res]
		}
	}
	sort.SliceStable(p.worst, func(i, j int) bool {
		ei, ej := p.eta[p.worst[i]], p.eta[p.worst[j]]
		if ei != ej {
			return ei > ej
		}
		// Both unreachable: the larger remaining fraction first.
		fi := 1 - p.amt[p.worst[i]]/p.target[p.worst[i]]
		fj := 1 - p.amt[p.worst[j]]/p.target[p.worst[j]]
		return fi > fj
	})
	p.invest = p.maxEta > b.HorizonTicks
	return p
}

// Play makes one round of economic decisions from a fresh snapshot.
func (b *Bot) Play(st game.GameState) {
	b.tick, b.age = st.Tick, st.Age
	if st.PendingMemoryTech != "" {
		b.act("accept_memory", st.PendingMemoryTech, b.ge.AcceptAncientMemory())
	}
	// The wonder, once its bank is full and its keystone tech is researched.
	if w := st.CurrentAgeWonderKey; w != "" && st.Buildings[w].WonderBankFull && st.Buildings[w].NeedsTech == "" && b.queuedCount(st, w) == 0 {
		if b.act("build_wonder", w, b.ge.BuildBuilding(w)) && b.RecordPlan {
			b.ge.NotePlanForTest(game.PlanBuild, w, 1)
		}
	}
	if b.answerHarbinger(st) {
		st = b.ge.GetState()
	}
	p := b.newPlan(st)
	if b.upgrade(p) {
		p = b.newPlan(b.ge.GetState())
	}
	if !b.UseShares {
		pop := b.recruit(p)
		b.assign(p, pop)
	}
	b.build(p)
	b.bankWonder(p)
	b.research(p)
	b.festival(p)
	if b.Deals {
		b.takeDeals(p)
	}
	if b.Army {
		b.keepGarrison(p)
	}
	b.gather(p)
}

// ArmyGarrison is how many of the age's newest military buildings the
// -army=on bot keeps: a modest garrison, not a war machine.
const ArmyGarrison = 4

// keepGarrison is the -army=on policy, a player who keeps a modest standing
// army: once soldiers begin (Iron Age on), hold ArmyGarrison of the current
// age's newest military building, buying one per decision when its price is
// at most a quarter of what the bot holds of everything it costs. The
// soldiers come with the buildings, and the bot never sends them out, so they
// are a garrison that blunts raids and catastrophes.
func (b *Bot) keepGarrison(p *plan) {
	if b.ageIdx[p.st.Age] < b.ageIdx[config.SoldierBeginAge] {
		return
	}
	best, tier := "", -1
	for _, key := range sortedKeys(p.st.Buildings) {
		def := b.defs[key]
		if def.LineageKey != "military" || def.LineageTier <= tier || !b.buildable(p, key) {
			continue
		}
		best, tier = key, def.LineageTier
	}
	if best == "" || p.st.Buildings[best].Count+b.queuedCount(p.st, best)+p.extra[best] >= ArmyGarrison {
		return
	}
	if c := b.cost(p, best); p.cheapFor(c, 0.25) {
		b.tryBuild(p, best, "build_army")
	}
}

// takeDeals is the -deals=on policy, a player who reads the Factions panel:
// take an open goods deal whose goods a target of the age still lacks (with
// room for them) and whose price is spare: held beyond its own target, not
// what the most wanted purchase is being saved for, and not knowledge, which
// research spends, unless it sits at its cap. Favor and rare deals only when
// their price would otherwise be lost at the cap. One deal per decision.
func (b *Bot) takeDeals(p *plan) {
	if !p.open(config.FeatureDiplomacy) {
		return // deals wait for a tech the bot has not researched yet
	}
	for _, key := range sortedKeys(p.st.Diplomacy.Factions) {
		for _, d := range p.st.Diplomacy.Factions[key].Deals {
			if d.Taken || p.amt[d.Give] < d.GiveAmt {
				continue
			}
			spare := p.amt[d.Give] - d.GiveAmt - p.target[d.Give] - p.blockCost[d.Give]
			if spare < 0 {
				continue
			}
			if d.Kind == game.DealSell || d.Kind == game.DealWant {
				if short := p.target[d.Get] - p.amt[d.Get]; short <= 0 {
					continue
				}
			} else {
				continue // favor and rare deals are for what a full store would lose
			}
			if got, err := b.ge.AcceptFactionDeal(key, d.Num); b.act("deal", fmt.Sprintf("%s#%d %s->%s", key, d.Num, d.Give, d.Get), err) {
				p.amt[d.Give] -= got.GiveAmt
				if got.Get != "" {
					p.amt[got.Get] += got.GetAmt
				}
			}
			return
		}
	}
}

// maxCheckInRounds bounds the rounds of one check-in.
const maxCheckInRounds = 30

// CheckIn is one visit by a player who checks in now and then: rounds of
// Play on a fresh snapshot until one changes nothing, so a visit spends and
// queues everything worthwhile (several producers, staffing, the wonder)
// instead of making one decision and leaving.
// No time passes during a visit. Returns the rounds played.
func (b *Bot) CheckIn(st game.GameState) int {
	if b.Trace != nil {
		b.traceCheckIn("checkin", st)
	}
	for round := 1; ; round++ {
		before := b.actionCount()
		b.Play(st)
		if round >= maxCheckInRounds || b.actionCount() == before {
			if b.UsePlan {
				b.planAhead(b.ge.GetState())
			}
			if b.Trace != nil {
				b.traceCheckIn(fmt.Sprintf("leaves after %d round(s)", round), b.ge.GetState())
			}
			return round
		}
		st = b.ge.GetState()
	}
}

// maxPlanCopies bounds one plan item the bot queues.
const maxPlanCopies = 40

// planIncomeSlack scales the income a check-in player budgets the plan
// against: production grows while they are away (the plan adds producers),
// and an item the income doesn't reach just waits for the next visit.
const planIncomeSlack = 1.5

// planAhead is how a check-in player leaves the game: a build plan
// (`plan build`, `plan research`) for the hours until the next
// visit, so the income of those hours is spent as it comes in instead of
// piling up unspent. It replaces the previous visit's plan.
//
// The plan is budgeted: each resource's budget is what the bot holds plus
// what it will make before the next visit (with some slack, see
// planIncomeSlack), less what the next age asks for as a resource
// requirement. Items go in priority order, each with as many copies as the
// budget left covers (along the cost curve):
//
//  1. the age's wonder, which starts once deposits fill its bank;
//  2. the next age's required buildings;
//  3. housing when the population is capped and worker slots stand empty;
//  4. producers of the slowest requirements, slowest first;
//  5. techs the knowledge budget covers, a missing producer's tech first.
//
// While the goal is beyond the horizon (investing), producers go before
// required buildings: a waiting item holds its price back, and a required
// building waiting on stone would hold back the stone pits that bootstrap
// it. Required buildings get a copy even when the budget can't cover one, as
// long as what they lack is coming in: they are what the age needs, and the
// plan waits for them.
func (b *Bot) planAhead(st game.GameState) {
	if b.CheckInTicks <= 0 {
		return
	}
	b.ge.PlanClear()
	p := b.newPlan(st)
	// Income ahead: today's rate plus what the buildings under construction
	// will add (they finish within minutes), with slack for the producers
	// the plan itself adds on the way.
	ahead := map[string]float64{}
	for _, q := range st.BuildQueue {
		for _, e := range b.defs[b.nameToKey[q.Name]].Effects {
			if e.Type == "production" && e.Value > 0 {
				ahead[e.Target] += e.Value
			}
		}
	}
	budget := map[string]float64{}
	for res, r := range st.Resources {
		if r.Unlocked {
			income := (math.Max(r.Rate, 0) + ahead[res]) * planIncomeSlack
			budget[res] = p.amt[res] + float64(income*b.CheckInTicks)
		}
	}
	covers := func(c map[string]float64) bool {
		for r, v := range c {
			if budget[r] < v {
				return false
			}
		}
		return true
	}
	// accrues reports whether every part of c the budget lacks is being
	// made, so the plan will get there by waiting.
	accrues := func(c map[string]float64) bool {
		for r, v := range c {
			if budget[r] < v && st.Resources[r].Rate+ahead[r] <= 0 {
				return false
			}
		}
		return true
	}
	add := func(key, kind string, want int, essential bool) int {
		n := 0
		for n < want && n < maxPlanCopies && b.buildable(p, key) {
			c := b.cost(p, key)
			if !covers(c) && !(essential && accrues(c)) {
				break
			}
			for r, v := range c {
				budget[r] -= v
			}
			p.extra[key]++
			n++
		}
		if n == 0 {
			return 0
		}
		got, err := b.ge.PlanAddBuild(key, n)
		b.act("plan_"+kind, fmt.Sprintf("%s %d", key, n), err)
		if err != nil {
			got = 0
		}
		p.extra[key] -= n - got
		for _, e := range b.defs[key].Effects {
			if e.Type == "production" && e.Value > 0 {
				ahead[e.Target] += float64(e.Value * float64(got)) // planned producers count as coming in
			}
		}
		return got
	}

	required := func() {
		for _, key := range sortedKeys(p.needBld) {
			add(key, "required", p.needBld[key], true)
		}
	}
	producers := func() {
		for i, res := range p.worst {
			if i >= 3 {
				break
			}
			if key, ok := b.bestFor(p, producer(res)); ok {
				add(key, "production", maxPlanCopies, false)
			}
		}
	}
	housingFor := func() {
		ws := st.Workers
		if ws.TotalPop >= ws.MaxPop && p.slots > ws.MaxPop {
			if key, ok := b.bestFor(p, housing); ok {
				add(key, "housing", 5, false)
			}
		}
	}

	if p.invest {
		producers()
		required()
		housingFor()
	} else {
		required()
		housingFor()
		producers()
	}
	b.planTechs(p, st, budget)
	// The wonder goes last before the advance: it pays the rest of its bank
	// from what the items above leave, so it takes the stock once the age's
	// investments are made rather than before (overflow fills it meanwhile).
	if w := st.CurrentAgeWonderKey; w != "" && st.Buildings[w].Count == 0 && b.queuedCount(st, w) == 0 {
		_, err := b.ge.PlanAddBuild(w, 1)
		b.act("plan_wonder", w, err)
	}
	// 8. Advance as soon as the age is ready, not at the next visit, and
	// 9. a start on the next age, which waits for the advance: its storage,
	// its wonder and the buildings the age after it requires.
	if st.NextAge != "" {
		b.act("plan_advance", st.NextAge, b.ge.PlanAddAdvance())
		b.planNextAge(st.NextAge)
	}
}

// planNextAge queues the next age's opening moves behind the advance item:
// a few copies of its storage building (a gate asks for them), its wonder, and its share of the
// following gate (the next age's buildings the age after it requires). They
// wait, reserving nothing, until the advance unlocks them, so a long absence
// that finishes one age keeps going in the next.
func (b *Bot) planNextAge(next string) {
	// First, producers of the next age that today's income can pay for:
	// the best one per resource the next age makes, so its new resources
	// (the Stone Age's stone, the Bronze Age's iron) start flowing without
	// waiting for a visit.
	st := b.ge.GetState()
	best := map[string]string{}
	bestV := map[string]float64{}
	for _, k := range sortedKeys(b.defs) {
		d := b.defs[k]
		if d.RequiredAge != next || d.Category == "wonder" || d.Category == "storage" || d.ModelCopyLimit() > 0 {
			continue
		}
		fundable := true
		for r := range d.BaseCost {
			if st.Resources[r].Rate <= 0 {
				fundable = false
			}
		}
		if !fundable {
			continue
		}
		units := 0.0
		for _, r := range sortedKeys(d.BaseCost) { // a float sum: sorted, not map order
			units += d.BaseCost[r] / math.Max(b.goodStock(st, r), 1)
		}
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 && !config.IsFlowResource(e.Target) {
				if v := e.Value / math.Max(units, 1e-9); v > bestV[e.Target] {
					best[e.Target], bestV[e.Target] = k, v
				}
			}
		}
	}
	seen := map[string]bool{}
	for _, res := range sortedKeys(best) {
		if k := best[res]; !seen[k] {
			seen[k] = true
			_, err := b.ge.PlanAddBuild(k, 5)
			b.act("plan_next_production", k, err)
		}
	}
	var keys []string
	for _, k := range sortedKeys(b.defs) {
		if d := b.defs[k]; d.RequiredAge == next && d.Category == "storage" {
			keys = append(keys, k)
		}
	}
	for _, k := range keys {
		_, err := b.ge.PlanAddBuild(k, 3)
		b.act("plan_next_storage", k, err)
	}
	nd, _ := b.rules.Age(next)
	for _, k := range sortedKeys(b.defs) {
		if d := b.defs[k]; d.RequiredAge == next && d.Category == "wonder" {
			_, err := b.ge.PlanAddBuild(k, 1)
			b.act("plan_next_wonder", k, err)
		}
	}
	after := b.rules.Next(nd.Key)
	if after == "" {
		return
	}
	gate, _ := b.rules.Age(after)
	reqs := gate.BuildingReqs
	for _, k := range sortedKeys(reqs) {
		if b.defs[k].RequiredAge == next {
			_, err := b.ge.PlanAddBuild(k, reqs[k])
			b.act("plan_next_required", fmt.Sprintf("%s %d", k, reqs[k]), err)
		}
	}
}

// planTechs adds research to the plan if the slot will be free before the
// next visit: first the techs the age cannot be left without, in order (the
// keystone's chain, then what unlocks something the bot wants to build),
// since a planned tech waits for the tech planned before it; then, once
// none of those is left, the cheapest techs. Three at most.
func (b *Bot) planTechs(p *plan, st game.GameState, budget map[string]float64) {
	rs := st.Research
	if rs.CurrentTech != "" && float64(rs.TicksLeft) > b.CheckInTicks {
		return
	}
	planned := 0
	for _, key := range p.must {
		t := rs.Techs[key]
		_ = t
		if planned >= 3 || key == rs.CurrentTech {
			break
		}
		if !b.act("plan_research", key, b.ge.PlanAddResearch(key)) {
			break
		}
		planned++
	}
	if len(p.must) > 0 {
		return
	}
	for _, key := range b.restOrder(p) {
		t := rs.Techs[key]
		if planned >= 3 || !t.Available || t.Researched || key == rs.CurrentTech {
			continue
		}
		if b.act("plan_research", key, b.ge.PlanAddResearch(key)) {
			planned++
		}
	}
}

// traceCheckIn writes what a visit finds or leaves: each resource the age
// still needs, as held/a good stock (+rate per tick, still to make), slowest first.
func (b *Bot) traceCheckIn(what string, st game.GameState) {
	p := b.newPlan(st)
	var parts []string
	for _, res := range p.worst {
		r := st.Resources[res]
		parts = append(parts, fmt.Sprintf("%s %s/%s (%+.3g/t, need %s)", res, num(r.Amount), num(p.stock[res]), r.Rate, num(p.target[res])))
	}
	// Required buildings not yet built or queued, whose price is in hand
	// (the ones it isn't show up above as resources).
	for _, k := range sortedKeys(p.needBld) {
		parts = append(parts, fmt.Sprintf("%s %d more", k, p.needBld[k]))
	}
	if w := st.CurrentAgeWonderKey; w != "" && st.Buildings[w].Count == 0 && b.queuedCount(st, w) == 0 {
		parts = append(parts, w+" unbuilt")
	}
	fmt.Fprintf(b.Trace, "tick %d %s %s: queue %d, research %q; waiting on: %s\n", st.Tick, what, st.Age, len(st.BuildQueue),
		st.Research.CurrentTech, strings.Join(parts, "; "))
	if len(st.Plan) > 0 {
		var items []string
		for _, v := range st.Plan {
			s := fmt.Sprintf("%s x%d %s", v.Key, v.Count, v.Status)
			switch v.Kind {
			case game.PlanAdvance:
				s = "advance " + v.Status
			}
			switch {
			case v.Note != "":
				s += " (" + v.Note + ")"
			case v.Short != "":
				s += fmt.Sprintf(" (%s %.0f%%)", v.Short, v.Progress*100)
			}
			items = append(items, s)
		}
		fmt.Fprintf(b.Trace, "tick %d   plan: %s\n", st.Tick, strings.Join(items, "; "))
	}
}

// actionCount is the number of successful actions so far, less hand
// gathering, which a player can always do once more.
func (b *Bot) actionCount() int {
	n := 0
	for k, v := range b.Actions {
		if k != "gather" {
			n += v
		}
	}
	return n
}

// queuedCount is how many copies of key are under construction.
func (b *Bot) queuedCount(st game.GameState, key string) int {
	n := 0
	for _, q := range st.BuildQueue {
		if b.nameToKey[q.Name] == key {
			n++
		}
	}
	return n
}

// Refusal is one action the game refused: when, in which age, what the bot
// asked for and what it was told.
type Refusal struct {
	Tick    int    `json:"tick"`
	Age     string `json:"age"`
	Kind    string `json:"kind"`
	Detail  string `json:"detail,omitempty"`
	Message string `json:"message"`
}

// refusalsKept is how many refusals a run's report carries in full.
const refusalsKept = 20

// refuse counts a rejected action of kind and keeps what it was told.
func (b *Bot) refuse(kind, detail, message string) {
	b.Errors[kind]++
	if len(b.Refusals) < refusalsKept {
		b.Refusals = append(b.Refusals, Refusal{Tick: b.tick, Age: b.age, Kind: kind, Detail: detail, Message: message})
	}
}

func (b *Bot) act(kind, detail string, err error) bool {
	if b.Trace != nil {
		res := "ok"
		if err != nil {
			res = err.Error()
		}
		fmt.Fprintf(b.Trace, "tick %d %s %s: %s\n", b.tick, kind, detail, res)
	}
	if err != nil {
		b.refuse(kind, detail, err.Error())
		return false
	}
	b.Actions[kind]++
	return true
}

// buildable reports whether key can be bought in the current age at all,
// ignoring cost.
func (b *Bot) buildable(p *plan, key string) bool {
	bs, ok := p.st.Buildings[key]
	def := b.defs[key]
	if !ok || !bs.Unlocked || bs.IsLegacy || bs.AtMaxCount || def.Category == "wonder" {
		return false
	}
	if def.RequiredAge != "" && def.RequiredAge != p.st.Age {
		return false
	}
	if limit := def.ModelCopyLimit(); limit > 0 && bs.Count+p.queued[key]+p.extra[key] >= limit {
		return false
	}
	return true
}

// cost returns the price of the next copy of key, accounting for copies
// already bought this decision.
func (b *Bot) cost(p *plan, key string) map[string]float64 {
	base := p.st.Buildings[key].NextCost
	mult := detmath.Pow(b.defs[key].CostScale, float64(p.extra[key]))
	out := make(map[string]float64, len(base))
	for r, c := range base {
		out[r] = c * mult
	}
	return out
}

func (p *plan) affordable(cost map[string]float64) bool {
	for r, c := range cost {
		if p.amt[r] < c {
			return false
		}
	}
	return true
}

// cheapFor reports whether cost is at most frac of what the bot holds.
func (p *plan) cheapFor(cost map[string]float64, frac float64) bool {
	for r, c := range cost {
		if c > p.amt[r]*frac {
			return false
		}
	}
	return true
}

// costFrac is the price as the largest fraction of a good stock of any
// resource it touches.
func (p *plan) costFrac(cost map[string]float64) float64 {
	f := 0.0
	for r, c := range cost {
		s := math.Max(p.stock[r], 1)
		f = math.Max(f, c/s)
	}
	return math.Max(f, 1e-9)
}

func (b *Bot) tryBuild(p *plan, key, kind string) bool {
	c := b.cost(p, key)
	if !p.affordable(c) {
		return false
	}
	// What the copy really costs is read off the stores: the game floors
	// each copy's price, so the second copy bought in one decision can cost
	// a fraction more than the first's price times the scale (cost), and a
	// plan that booked the estimate went on to sell wood it did not have.
	held := make(map[string]float64, len(c))
	for r := range c {
		held[r] = b.ge.Resources.Get(r)
	}
	if !b.act(kind, key, b.ge.BuildBuilding(key)) {
		return false
	}
	if b.RecordPlan {
		b.ge.NotePlanForTest(game.PlanBuild, key, 1)
	}
	for r := range c {
		p.amt[r] -= held[r] - b.ge.Resources.Get(r)
	}
	p.extra[key]++
	return true
}

// buyOrBootstrap buys key if it can. If not, it records the resource that
// blocks it longest as the plan's blocker and, with depth left, tries to buy
// a producer of that resource instead. Reports whether key itself was bought.
func (b *Bot) buyOrBootstrap(p *plan, key, kind string, depth int) bool {
	c := b.cost(p, key)
	if p.affordable(c) {
		return b.tryBuild(p, key, kind)
	}
	block, worst := "", -1.0
	for _, r := range sortedKeys(c) {
		deficit := c[r] - p.amt[r]
		if deficit <= 0 {
			continue
		}
		t := math.Inf(1)
		if rate := p.st.Resources[r].Rate; rate > 0 {
			t = deficit / rate
		}
		if t > worst {
			block, worst = r, t
		}
	}
	if block == "" {
		return false
	}
	if p.blocker == "" {
		p.blocker, p.blockCost = block, c
	}
	if depth > 0 {
		if k2, ok := b.bestFor(p, producer(block)); ok && k2 != key {
			b.buyOrBootstrap(p, k2, "build_bootstrap", depth-1)
		}
	}
	return false
}

// upgrade converts older buildings to their line's building, one copy at a time.
// It upgrades toward a required building whenever it can afford to, and
// otherwise only while investing and only when a copy is cheap. It never
// upgrades into a building with a MaxCount: an upgrade would spend one of the
// target's capped slots. (The game no longer offers storage upgrades and
// stops any upgrade at the cap; this is belt and braces.) Reports whether
// anything changed.
func (b *Bot) upgrade(p *plan) bool {
	changed := false
	for _, u := range b.ge.GetAvailableUpgrades() {
		to := b.defs[u.ToKey]
		if to.MaxCount > 0 || u.Count <= 0 {
			continue
		}
		per := u.Cost // the next copy's price
		n := 0
		switch {
		case p.needBld[u.ToKey] > 0:
			n = p.needBld[u.ToKey]
		case p.invest && p.cheapFor(per, 0.25):
			n = u.Count
		}
		for i := 0; i < n && i < u.Count && i < 200; i++ {
			if b.ge.UpgradeBuilding(u.FromKey, 1, false) != nil {
				break
			}
			b.act("upgrade", u.FromKey, nil)
			if b.RecordPlan {
				// The plan can't upgrade: a template written from this run
				// builds the new tier instead.
				b.ge.NotePlanForTest(game.PlanBuild, u.ToKey, 1)
			}
			changed = true
		}
	}
	return changed
}

// perWorkerYield is the production one more worker adds to building key.
func (b *Bot) perWorkerYield(key string) map[string]float64 {
	def := b.defs[key]
	out := make(map[string]float64)
	if def.WorkerCapacity <= 0 {
		return out
	}
	for _, e := range def.Effects {
		if e.Type == "production" {
			out[e.Target] += e.Value * 0.8 / float64(def.WorkerCapacity)
		}
	}
	return out
}

// weights values one unit of each resource per tick.
func (b *Bot) weights(p *plan) map[string]float64 {
	w := make(map[string]float64)
	finiteMax := 0.0
	for _, res := range p.worst {
		if !math.IsInf(p.eta[res], 1) {
			finiteMax = math.Max(finiteMax, p.eta[res])
		}
	}
	for res, r := range p.st.Resources {
		if !r.Unlocked {
			continue
		}
		goal := math.Max(math.Max(p.target[res], p.stock[res]), 1)
		v := 1 / goal
		if p.target[res] > 0 && p.eta[res] > 0 {
			boost := 5.0
			if !math.IsInf(p.eta[res], 1) && finiteMax > 0 {
				boost = 1 + 4*p.eta[res]/finiteMax
			}
			v *= 1 + boost
		} else if p.target[res] == 0 {
			v *= 0.25 // nothing this age asks for it
		}
		w[res] = v
	}
	food := p.st.Resources["food"]
	if p.foodRate < 0 || food.Amount < 0.2*p.stock["food"] {
		w["food"] *= 20
	}
	return w
}

// recruit fills empty worker slots as far as housing and food allow; idle
// workers only eat. Returns the new population.
func (b *Bot) recruit(p *plan) int {
	ws := p.st.Workers
	room := min(ws.MaxPop, p.slots) - ws.TotalPop
	if b.UsePlan && ws.MaxPop > p.slots {
		// A check-in player keeps some idle hands for the producers the
		// plan will finish before the next visit (it staffs them from idle
		// workers): as many as half the food surplus feeds, within housing.
		per := 0.1
		if ws.TotalPop > 0 && ws.FoodDrain > 0 {
			per = ws.FoodDrain / float64(ws.TotalPop)
		}
		spare := int(math.Max(0, p.foodRate) / per / 2)
		room = max(room, 0) + min(ws.MaxPop-max(p.slots, ws.TotalPop), spare)
	}
	if room <= 0 {
		return ws.TotalPop
	}
	per := 0.1
	if ws.TotalPop > 0 && ws.FoodDrain > 0 {
		per = ws.FoodDrain / float64(ws.TotalPop)
	}
	food := p.st.Resources["food"]
	n := room
	if food.Amount < 0.3*p.stock["food"] {
		// Thin buffer: as many as the surplus feeds, plus whoever fills the
		// empty food slots (they grow more than they eat).
		n = int(math.Max(0, p.foodRate)/per) + p.freeFoodSlots
		if n > room {
			n = room
		}
	}
	if n <= 0 {
		return ws.TotalPop
	}
	if b.act("recruit", fmt.Sprint(n), b.ge.RecruitWorker("worker", n)) {
		p.foodRate -= float64(per * float64(n))
		return ws.TotalPop + n
	}
	return ws.TotalPop
}

// assign staffs buildings by value per worker, then applies the difference
// from the current assignment with unassign/assign calls.
func (b *Bot) assign(p *plan, pop int) {
	w := b.weights(p)
	type slot struct {
		key   string
		cap   int
		score float64
	}
	var slots []slot
	for _, key := range sortedKeys(p.st.Buildings) {
		bs := p.st.Buildings[key]
		def := b.defs[key]
		if bs.Count <= 0 || def.WorkerCapacity <= 0 {
			continue
		}
		score := 0.0
		for r, v := range b.perWorkerYield(key) {
			score += float64(v * w[r])
		}
		slots = append(slots, slot{key: key, cap: bs.Count * def.WorkerCapacity, score: score})
	}
	sort.SliceStable(slots, func(i, j int) bool { return slots[i].score > slots[j].score })
	want := make(map[string]int)
	left := pop
	for _, s := range slots {
		if left <= 0 || s.score <= 0 {
			break
		}
		n := s.cap
		if n > left {
			n = left
		}
		want[s.key] = n
		left -= n
	}
	// Release first so the assigns below find idle workers.
	for _, s := range slots {
		cur := p.st.Buildings[s.key].WorkersAssigned
		if cur > want[s.key] {
			b.act("unassign", fmt.Sprintf("%s %d", s.key, cur-want[s.key]), b.ge.UnassignWorker(s.key, cur-want[s.key]))
		}
	}
	for _, s := range slots {
		cur := p.st.Buildings[s.key].WorkersAssigned
		if want[s.key] > cur {
			b.act("assign", fmt.Sprintf("%s %d", s.key, want[s.key]-cur), b.ge.AssignWorker(s.key, want[s.key]-cur))
		}
	}
}

// bestFor picks the buildable building with the best effect value per unit
// of cost for (typ, target). match decides whether an effect counts. It
// prefers buildings the bot can pay for within a few horizons at current
// rates: a pricier producer that needs a resource nothing makes yet (the
// Renaissance foundry's steel) loses to one it can actually buy (the mill),
// as it would for a human. If nothing is reachable it falls back to the best
// score, so buyOrBootstrap can go after the blocker.
func (b *Bot) bestFor(p *plan, match func(config.Effect) float64) (string, bool) {
	best, bestScore := "", 0.0
	far, farScore := "", 0.0
	for _, key := range sortedKeys(p.st.Buildings) {
		if !b.buildable(p, key) {
			continue
		}
		v := 0.0
		for _, e := range b.defs[key].Effects {
			v += match(e)
		}
		if v <= 0 {
			continue
		}
		c := b.cost(p, key)
		s := v / p.costFrac(c)
		if b.payEta(p, c) > 4*b.HorizonTicks {
			if s > farScore {
				far, farScore = key, s
			}
			continue
		}
		if s > bestScore {
			best, bestScore = key, s
		}
	}
	if best == "" {
		best = far
	}
	return best, best != ""
}

// payEta is how many ticks until the bot holds cost at current rates (0 if
// it already does, +Inf if a missing resource is not being produced).
func (b *Bot) payEta(p *plan, cost map[string]float64) float64 {
	eta := 0.0
	for r, c := range cost {
		deficit := c - p.amt[r]
		if deficit <= 0 {
			continue
		}
		rate := p.st.Resources[r].Rate
		if rate <= 0 {
			return math.Inf(1)
		}
		eta = math.Max(eta, deficit/rate)
	}
	return eta
}

func producer(res string) func(config.Effect) float64 {
	return func(e config.Effect) float64 {
		if e.Type == "production" && e.Target == res && e.Value > 0 {
			return e.Value
		}
		return 0
	}
}

func housing(e config.Effect) float64 {
	if e.Type == "capacity" && e.Target == "population" && e.Value > 0 {
		return e.Value
	}
	return 0
}

func (b *Bot) build(p *plan) {
	// 1. Food first: a negative net rate starves workers.
	if p.foodRate < 0 {
		if key, ok := b.bestFor(p, producer("food")); ok {
			b.buyOrBootstrap(p, key, "build_food", 2)
		}
	}

	// 2. The next age's required buildings.
	for _, key := range sortedKeys(p.needBld) {
		for i := 0; i < p.needBld[key] && i < 200; i++ {
			if !b.buildable(p, key) || !b.buyOrBootstrap(p, key, "build_required", 0) {
				break
			}
		}
	}

	// 4. Housing when the population is capped and slots stand empty.
	ws := p.st.Workers
	if ws.TotalPop >= ws.MaxPop && p.slots > ws.MaxPop {
		if key, ok := b.bestFor(p, housing); ok {
			if c := b.cost(p, key); p.invest || p.cheapFor(c, 0.5) {
				b.tryBuild(p, key, "build_housing")
			}
		}
	}

	// 5. Invest in the slowest requirement while it is beyond the horizon.
	if !p.invest {
		return
	}
	for _, res := range p.worst {
		key, ok := b.bestFor(p, producer(res))
		if !ok {
			continue // nothing buildable makes it; try the next slowest
		}
		// Buy it if we can; otherwise bootstrap whatever blocks it, and save
		// rather than spend the same resources on something slower to matter.
		b.buyOrBootstrap(p, key, "build_production", 2)
		return
	}
}

// bankWonder moves surplus above the age requirements into the wonder bank.
// While investing it only banks what would otherwise be lost at the cap,
// unless nothing but the wonder wants that resource (the Hoover Dam's
// stone): holding it back then only delays the advance.
func (b *Bot) bankWonder(p *plan) {
	w := p.st.CurrentAgeWonderKey
	if w == "" {
		return
	}
	bank := p.st.Buildings[w].WonderBank
	cost := b.defs[w].BaseCost
	for _, res := range sortedKeys(cost) {
		left := cost[res] - bank[res]
		if left <= 0 {
			continue
		}
		keep := p.hold[res]
		onlyWonder := p.target[res]-left <= keep
		switch {
		case p.invest && !onlyWonder && b.CheckInTicks > 0:
			// Nothing is lost at a cap, so what the store holds now is
			// better left to the build plan.
			continue
		case p.invest && !onlyWonder:
			if p.amt[res] < 0.9*p.stock[res] {
				continue
			}
			keep = math.Max(keep, 0.5*p.stock[res])
		}
		dep := math.Min(left, p.amt[res]-keep)
		if dep < left {
			dep = math.Floor(dep) // whole units, except the last fraction
		}
		if dep <= 0 || (dep < 1 && dep < left) {
			continue
		}
		_, err := b.ge.BankWonderResource(w, res, dep)
		if b.act("bank_wonder", fmt.Sprintf("%s %.0f", res, dep), err) {
			p.amt[res] -= dep
		}
	}
}

// research starts the next tech: the next of the techs the age cannot be
// left without (the keystone's chain, then what unlocks something the bot
// wants to build; see bot_research.go), and once none is left the cheapest
// available tech. A tech has no price, so it starts when the slot is free.
func (b *Bot) research(p *plan) {
	rs := p.st.Research
	if rs.CurrentTech != "" {
		return
	}
	if key := p.nextMust(); key != "" {
		b.startResearch(p, key)
		return
	}
	for _, key := range b.restOrder(p) {
		t := rs.Techs[key]
		if !t.Available || t.Researched {
			continue
		}
		b.startResearch(p, key)
		return
	}
}

// startResearch starts tech key. Reports whether it started.
func (b *Bot) startResearch(p *plan, key string) bool {
	if !b.act("research", key, b.ge.StartResearch(key)) {
		return false
	}
	if b.RecordPlan {
		// The tech the bot chose, written as a planned research item: its
		// own research path, which the template keeps.
		b.ge.NotePlanForTest(game.PlanResearch, key, 1)
	}
	return true
}

// festival spends culture on a production buff when it is piling up.
func (b *Bot) festival(p *plan) {
	c, ok := p.st.Resources["culture"]
	if !ok || !c.Unlocked || c.Amount < 0.8*p.stock["culture"] {
		return
	}
	if !p.open(config.FeatureFestivals) {
		return // festivals wait for a tech the bot has not researched yet
	}
	fs := b.ge.FestivalStatus()
	if fs.Ready && fs.Culture >= fs.Cost {
		b.act("festival", "", b.ge.DoFestival())
	}
}

// gather hand-gathers the slowest of food/wood/stone, once per decision,
// while the game still allows it.
func (b *Bot) gather(p *plan) {
	if b.ageIdx[p.st.Age] > b.gatherMax {
		return
	}
	pick := ""
	if g := p.blocker; (g == "food" || g == "wood" || g == "stone") && p.st.Resources[g].Unlocked && p.amt[g] < p.stock[g] {
		pick = g
	}
	for _, res := range p.worst {
		if pick != "" {
			break
		}
		if res == "food" || res == "wood" || res == "stone" {
			if p.st.Resources[res].Unlocked && p.amt[res] < p.stock[res] {
				pick = res
				break
			}
		}
	}
	if pick == "" && p.foodRate < 0 {
		pick = "food"
	}
	if pick == "" {
		return
	}
	if _, err := b.ge.GatherResource(pick, gatherYield); b.act("gather", pick, err) {
		p.amt[pick] += gatherYield + p.st.Research.Mechanics[config.MechanicGatherAmount]
	}
}
