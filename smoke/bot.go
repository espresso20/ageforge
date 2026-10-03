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
//   - buy storage whenever a requirement or a needed cost exceeds a cap
//   - buy housing when the population is capped and worker slots are free
//   - while the slowest requirement is more than the horizon away, invest in
//     the best producer for it; otherwise save
//   - staff buildings by how much each worker moves the requirements forward
//   - prefer producers it can pay for within a few horizons over pricier
//     ones that need a resource nothing makes yet
//   - at the market: buy food when workers starve, buy whatever blocks the
//     most wanted purchase (storage included) with spare resources, and
//     otherwise even out the slowest target; never undo a recent trade, and
//     never sell what the most wanted purchase is being saved for unless it
//     sits at its cap
//   - bank surplus into the age's wonder, research affordable techs with
//     surplus knowledge (a tech that makes a missing target resource
//     first), and hand-gather while the game allows it
//
// Every map walk goes through sorted keys so a seed replays the same decisions.
type Bot struct {
	ge        *game.GameEngine
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
	// visit: storage for what production will bring in by then, and the
	// wonder banked with what would otherwise be lost at the cap. See
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

	// Trace, if set, receives one line per attempted action.
	Trace io.Writer
	// Harbinger is the policy for harbinger answers: ignore, appease, brace
	// or both (Appease up to level 2, Brace level 1).
	Harbinger string
	tick      int
	// sold and bought remember the tick each resource last left or entered
	// through the market, so two trading rules can't undo each other.
	sold, bought map[string]int
}

// gatherYield matches the command line's per-command cap (ui gatherMaxYield).
const gatherYield = 25.0

// NewBot returns a bot that plays ge.
func NewBot(ge *game.GameEngine) *Bot {
	defs := config.BuildingByKey()
	n2k := make(map[string]string, len(defs))
	for k, d := range defs {
		n2k[d.Name] = k
	}
	idx := make(map[string]int)
	for i, k := range config.AgeOrder() {
		idx[k] = i
	}
	return &Bot{
		ge:           ge,
		defs:         defs,
		nameToKey:    n2k,
		ageIdx:       idx,
		gatherMax:    idx["medieval_age"],
		HorizonTicks: 900, // 30 minutes at 1x
		Actions:      make(map[string]int),
		Errors:       make(map[string]int),
		sold:         make(map[string]int),
		bought:       make(map[string]int),
	}
}

// plan is the bot's working view of one decision.
type plan struct {
	// storeK is Era Mastery's speed in this age: a storage building adds
	// its effect × storeK (storage grows with k).
	storeK   float64
	st       game.GameState
	amt      map[string]float64 // spendable amounts, decremented as the bot spends
	storage  map[string]float64 // caps, raised locally as storage is bought
	target   map[string]float64 // total amount still to generate this age
	eta      map[string]float64 // ticks until target, +Inf if unreachable at the current rate
	maxEta   float64
	worst    []string           // target resources, slowest first
	needBld  map[string]int     // required building deficits (built + queued)
	queued   map[string]int     // queued copies per key
	extra    map[string]int     // copies bought this decision, for cost scaling
	capNeed  map[string]float64 // amounts that must fit under a cap
	invest   bool
	foodRate float64
	slots    int // worker slots across all buildings
	// freeFoodSlots is the unstaffed capacity of food producers.
	freeFoodSlots int
	// blocker is the resource keeping the most wanted purchase out of reach;
	// hand-gathering goes there first.
	blocker string
	// blockNeed is how much of blocker that purchase costs; blockCost is its
	// whole price, which trades for the blocker must not sell off.
	blockNeed float64
	blockCost map[string]float64
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (b *Bot) newPlan(st game.GameState) *plan {
	p := &plan{
		st:      st,
		amt:     make(map[string]float64),
		storage: make(map[string]float64),
		target:  make(map[string]float64),
		eta:     make(map[string]float64),
		needBld: make(map[string]int),
		queued:  make(map[string]int),
		extra:   make(map[string]int),
		capNeed: make(map[string]float64),
	}
	p.storeK = max(st.Mastery.K, 1)
	for k, r := range st.Resources {
		p.amt[k] = r.Amount
		p.storage[k] = r.Storage
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
	for res, v := range st.NextAgeResReqs {
		p.target[res] += v
		p.capNeed[res] = math.Max(p.capNeed[res], v)
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
	b.harbingerTargets(p)
	if b.CheckInTicks > 0 {
		// Room for what comes in before the next visit, up to what the age
		// still needs: anything over the cap by then is lost. A player who
		// leaves a build plan spends most of it as it comes in, so they
		// store an hour's worth (the Storage Covenant's scale) rather than
		// the whole interval's: storage bought for eight hours of income
		// would cost more than the income it saves.
		ahead := b.CheckInTicks
		if b.UsePlan {
			ahead = math.Min(ahead, planStoreTicks)
		}
		for _, res := range sortedKeys(p.target) {
			if rate := st.Resources[res].Rate; rate > 0 && p.target[res] > 0 {
				want := math.Min(p.target[res], p.amt[res]+float64(rate*ahead))
				p.capNeed[res] = math.Max(p.capNeed[res], want)
			}
		}
	}
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
	b.tick = st.Tick
	if st.PendingMemoryTech != "" {
		b.act("accept_memory", st.PendingMemoryTech, b.ge.AcceptAncientMemory())
	}
	if w := st.CurrentAgeWonderKey; w != "" && st.Buildings[w].WonderBankFull && b.queuedCount(st, w) == 0 {
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
	b.trade(p)
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
// army: once soldiers exist (Iron Age on), hold ArmyGarrison of the current
// age's newest military building, buying one per decision when its price is
// at most a quarter of what the bot holds of everything it costs. The
// buildings train soldiers unstaffed (20% rate), and the bot never spends
// them, so the stock is a garrison that blunts raids and catastrophes.
func (b *Bot) keepGarrison(p *plan) {
	if r, ok := p.st.Resources["soldiers"]; !ok || !r.Unlocked {
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
	for _, key := range sortedKeys(p.st.Diplomacy.Factions) {
		for _, d := range p.st.Diplomacy.Factions[key].Deals {
			if d.Taken || p.amt[d.Give] < d.GiveAmt {
				continue
			}
			atCap := p.amt[d.Give] >= 0.95*p.storage[d.Give]
			spare := p.amt[d.Give] - d.GiveAmt - p.target[d.Give] - p.blockCost[d.Give]
			switch {
			case atCap:
			case d.Give == "knowledge" || spare < 0:
				continue
			}
			if d.Kind == game.DealSell || d.Kind == game.DealWant {
				short := p.target[d.Get] - p.amt[d.Get]
				room := p.storage[d.Get] - p.amt[d.Get]
				if short <= 0 || room < d.GetAmt {
					continue
				}
			} else if !atCap {
				continue
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
// queues everything worthwhile (several producers, storage for the hours
// ahead, staffing, the wonder) instead of making one decision and leaving.
// No time passes during a visit. Returns the rounds played.
func (b *Bot) CheckIn(st game.GameState) int {
	if b.Trace != nil {
		b.traceCheckIn("checkin", st)
	}
	if b.UsePlan && b.expandStorage(st) {
		st = b.ge.GetState()
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

// planStoreTicks is how much of the income ahead a check-in player who
// leaves a plan buys storage for: an hour at 1x.
const planStoreTicks = 1800

// planAhead is how a check-in player leaves the game: a build plan
// (`plan build`, `plan research`, `plan trade`) for the hours until the next
// visit, so the income of those hours is spent as it comes in instead of
// piling up at the caps. It replaces the previous visit's plan. Wonder
// overflow stays on (the default), so what the plan doesn't spend and the
// stores can't hold goes into the wonder.
//
// The plan is budgeted: each resource's budget is what the bot holds plus
// what it will make before the next visit (with some slack, see
// planIncomeSlack), less what the next age asks for as a resource
// requirement. Items go in priority order, each with as many copies as the
// budget left covers (along the cost curve):
//
//  1. the age's wonder, which starts once overflow and deposits fill its bank;
//  2. storage for anything that must fit under a cap (requirements, the
//     income ahead, prices over a cap);
//  3. the next age's required buildings;
//  4. housing when the population is capped and worker slots stand empty;
//  5. producers of the slowest requirements, slowest first;
//  6. techs the knowledge budget covers, a missing producer's tech first;
//  7. trades for what the age buys rather than makes, from what overflows.
//
// While the goal is beyond the horizon (investing), producers go before
// storage and required buildings: a waiting item holds its price back, and a
// required building waiting on stone would hold back the stone pits that
// bootstrap it. Storage and required buildings get a copy even when the
// budget can't cover one, as long as what they lack is coming in: they are
// what the age needs, and the plan waits for them. Anything short of a
// resource nothing makes is first bought at the market (fundAtMarket).
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
			budget[res] = p.amt[res] + float64(income*b.CheckInTicks) - st.NextAgeResReqs[res]
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
	// tradeIn is what the plan's trade items buy: it comes in while the
	// player is away even though nothing makes it (the Bronze Age's gold
	// before a market stands, the later ages' market-only stone).
	tradeIn := map[string]bool{}
	accrues := func(c map[string]float64) bool {
		for r, v := range c {
			if budget[r] < v && st.Resources[r].Rate+ahead[r] <= 0 && !tradeIn[r] {
				return false
			}
		}
		return true
	}
	add := func(key, kind string, want int, essential bool) int {
		n := 0
		for n < want && n < maxPlanCopies && b.buildable(p, key) {
			c := b.cost(p, key)
			if !p.fits(c) {
				break
			}
			if !covers(c) && n == 0 {
				b.fundAtMarket(p, c, budget)
			}
			if essential {
				// What the age hardly makes and the budget lacks (the
				// Atomic Age's iron for its vaults) is kept topped up at
				// the market while the player is away.
				for _, r := range sortedKeys(c) {
					scarce := (st.Resources[r].Rate+ahead[r])*b.CheckInTicks < 0.1*c[r]
					if budget[r] < c[r] && scarce && !tradeIn[r] && b.planTopUp(p, st, r) {
						tradeIn[r] = true
					}
				}
			}
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

	storage := func() {
		for _, key := range sortedKeys(p.needBld) {
			p.fits(b.cost(p, key)) // record prices over a cap in capNeed
		}
		for _, res := range sortedKeys(p.capNeed) {
			if p.capNeed[res] <= p.storage[res]*0.98 {
				continue
			}
			key, ok := b.bestFor(p, storer(res))
			if !ok {
				continue
			}
			per := 0.0
			for _, e := range b.defs[key].Effects {
				per += float64(storer(res)(e) * p.storeK)
			}
			want := 1
			if per > 0 {
				want = int(math.Ceil((p.capNeed[res] - p.storage[res]) / per))
			}
			n := add(key, "storage", want, true)
			for _, e := range b.defs[key].Effects {
				if e.Type == "storage" {
					for r := range p.storage {
						if e.Target == r || e.Target == "all" {
							p.storage[r] += float64(float64(e.Value*float64(n)) * p.storeK)
						}
					}
				}
			}
		}
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

	// Trades go first: a build waiting on stone holds back the gold it also
	// costs, so a trade below it could never sell that gold for the stone.
	// Each is capped at what is missing.
	b.planTrades(p, st)
	if b.hasTrader(st) {
		for _, v := range b.ge.GetState().Plan {
			if v.Kind == game.PlanTrade {
				tradeIn[v.To] = true
			}
		}
	}
	if p.invest {
		producers()
		storage()
		required()
		housingFor()
	} else {
		storage()
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
// a few copies of its storage building, its wonder, and its share of the
// following gate (the next age's buildings the age after it requires). They
// wait, reserving nothing, until the advance unlocks them, so a long absence
// that finishes one age keeps going in the next.
func (b *Bot) planNextAge(next string) {
	ages := config.AgeByKey()
	// First, producers of the next age that today's income can pay for:
	// the best one per resource the next age makes, so its new resources
	// (the Stone Age's stone, the Bronze Age's iron) start flowing without
	// waiting for a visit.
	st := b.ge.GetState()
	best := map[string]string{}
	bestV := map[string]float64{}
	for _, k := range sortedKeys(b.defs) {
		d := b.defs[k]
		if d.RequiredAge != next || d.Category == "wonder" || d.Category == "storage" || d.MaxCount > 0 {
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
			units += d.BaseCost[r] / math.Max(st.Resources[r].Storage, 1)
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
	nd := ages[next]
	for _, k := range sortedKeys(b.defs) {
		if d := b.defs[k]; d.RequiredAge == next && d.Category == "wonder" {
			_, err := b.ge.PlanAddBuild(k, 1)
			b.act("plan_next_wonder", k, err)
		}
	}
	after := ""
	for i, a := range config.AgeOrder() {
		if a == nd.Key && i+1 < len(config.AgeOrder()) {
			after = config.AgeOrder()[i+1]
		}
	}
	if after == "" {
		return
	}
	reqs := ages[after].BuildingReqs
	for _, k := range sortedKeys(reqs) {
		if b.defs[k].RequiredAge == next {
			_, err := b.ge.PlanAddBuild(k, reqs[k])
			b.act("plan_next_required", fmt.Sprintf("%s %d", k, reqs[k]), err)
		}
	}
}

// maxPlanTrades bounds the trade items a check-in player leaves.
const maxPlanTrades = 4

// planTrades adds `plan trade` items for the resources the age still needs
// that the income before the next visit won't bring (the market-only stone
// and iron of the later ages): each sold from the resource that fills its
// store before the next visit with the most to spare, for what is missing.
// A seller can serve several targets while its spare lasts: what it makes
// before the next visit past a full store and past what the age still needs
// of it.
func (b *Bot) planTrades(p *plan, st game.GameState) {
	traders := 0
	for key, bs := range st.Buildings {
		if b.defs[key].LineageKey == "trade" {
			traders += bs.Count
		}
	}
	income := func(r string) float64 { return float64(math.Max(st.Resources[r].Rate, 0) * b.CheckInTicks) }
	if traders == 0 {
		// No market yet: if the age needs a resource nothing makes, plan
		// the cheapest trade building first; the trade items below wait
		// for it.
		missing := false
		for _, res := range p.worst {
			if st.Resources[res].Rate <= 0 && p.target[res] > p.amt[res] {
				missing = true
			}
		}
		key, cheapest := "", math.Inf(1)
		for _, k := range sortedKeys(st.Buildings) {
			if b.defs[k].LineageKey != "trade" || !b.buildable(p, k) {
				continue
			}
			if f := p.costFrac(b.cost(p, k)); f < cheapest {
				key, cheapest = k, f
			}
		}
		if !missing || key == "" {
			return
		}
		_, err := b.ge.PlanAddBuild(key, 1)
		b.act("plan_market", key, err)
	}
	spare := map[string]float64{}
	spareOf := func(r string) float64 {
		if v, ok := spare[r]; ok {
			return v
		}
		v := p.amt[r] + income(r) - math.Max(0, p.target[r])
		spare[r] = math.Max(v, 0)
		return spare[r]
	}
	added := 0
	for _, want := range p.worst {
		short := p.target[want] - p.amt[want] - income(want)
		// Several sellers may serve one target, best first, until what is
		// missing is covered.
		for short > 0 && added < maxPlanTrades {
			from, best, rate := "", 0.0, 0.0
			for _, k := range sortedKeys(st.Trade.ExchangeRates) {
				x := st.Trade.ExchangeRates[k]
				if x.To != want || x.Rate <= 0 || x.From == want {
					continue
				}
				if v := spareOf(x.From) * x.Rate; v > best && !b.planHasTrade(x.From, want) {
					from, best, rate = x.From, v, x.Rate
				}
			}
			if from == "" || best < 1 {
				break
			}
			amount := math.Min(short, best)
			// A resource the age makes little of is kept topped up instead
			// (no amount): what it buys also feeds the other items of the
			// plan that cost it, which an amount sized to the gate misses.
			ask := amount
			if income(want) < 0.5*(p.target[want]-p.amt[want]) {
				ask = 0
			}
			if !b.act("plan_trade", from+"->"+want, b.ge.PlanAddTrade(from, want, ask)) {
				break
			}
			added++
			spare[from] -= amount / rate
			short -= amount
		}
	}
}

// planTopUp adds a `plan trade` item that keeps res topped up (no amount),
// sold from the resource with the most to spare before the next visit (what
// it holds and makes, less what the age still needs of it). Reports whether
// it added one.
func (b *Bot) planTopUp(p *plan, st game.GameState, res string) bool {
	if !b.hasTrader(st) {
		return false
	}
	from, best := "", 0.0
	for _, k := range sortedKeys(st.Trade.ExchangeRates) {
		x := st.Trade.ExchangeRates[k]
		if x.To != res || x.Rate <= 0 || x.From == res || b.planHasTrade(x.From, res) {
			continue
		}
		spare := p.amt[x.From] + float64(math.Max(st.Resources[x.From].Rate, 0)*b.CheckInTicks) - math.Max(0, p.target[x.From])
		if v := spare * x.Rate; v > best {
			from, best = x.From, v
		}
	}
	if from == "" || best < 1 {
		return false
	}
	return b.act("plan_topup", from+"->"+res, b.ge.PlanAddTrade(from, res, 0))
}

// hasTrader reports whether a trade building stands, so the market is open
// and a plan trade item can sell while the player is away. (A trade item
// waiting for one still holds back what it would sell.)
func (b *Bot) hasTrader(st game.GameState) bool {
	for key, bs := range st.Buildings {
		if bs.Count > 0 && b.defs[key].LineageKey == "trade" {
			return true
		}
	}
	return false
}

// planHasTrade reports whether the plan already trades from for to.
func (b *Bot) planHasTrade(from, to string) bool {
	for _, v := range b.ge.GetState().Plan {
		if v.Kind == game.PlanTrade && v.Key == from && v.To == to {
			return true
		}
	}
	return false
}

// planTechs adds the techs the knowledge budget covers, a missing
// producer's tech first, if the research slot will be free before the next
// visit.
func (b *Bot) planTechs(p *plan, st game.GameState, budget map[string]float64) {
	rs := st.Research
	if rs.CurrentTech != "" && float64(rs.TicksLeft) > b.CheckInTicks {
		return
	}
	techs := config.TechByKey()
	keys := sortedKeys(rs.Techs)
	unblocks := func(key string) bool {
		for _, e := range techs[key].Effects {
			if e.Type == "production" && p.target[e.Target] > p.amt[e.Target] && st.Resources[e.Target].Rate <= 0 {
				return true
			}
		}
		return false
	}
	sort.SliceStable(keys, func(i, j int) bool {
		ui, uj := unblocks(keys[i]), unblocks(keys[j])
		if ui != uj {
			return ui
		}
		return rs.Techs[keys[i]].Cost < rs.Techs[keys[j]].Cost
	})
	planned := 0
	for _, key := range keys {
		t := rs.Techs[key]
		if planned >= 3 || !t.Available || t.Researched || key == rs.CurrentTech {
			continue
		}
		if t.Cost > p.storage["knowledge"] || budget["knowledge"] < t.Cost {
			continue
		}
		if b.act("plan_research", key, b.ge.PlanAddResearch(key)) {
			budget["knowledge"] -= t.Cost
			planned++
		}
	}
}

// expandStorage is the first thing a check-in player does on arriving to
// full stores: build the storage the hours until the next visit call for
// (newPlan's capNeed: requirements, prices, and the income ahead up to what
// the age still needs), buying at the market whatever storage input nothing
// makes, with what sits at the caps before anything else spends it. Reports
// whether it built anything.
func (b *Bot) expandStorage(st game.GameState) bool {
	p := b.newPlan(st)
	built := false
	skip := map[string]bool{}
	for i := 0; i < 25; i++ {
		res := ""
		for _, r := range sortedKeys(p.capNeed) {
			if !skip[r] && p.capNeed[r] > p.storage[r]*0.98 {
				res = r
				break
			}
		}
		if res == "" {
			break
		}
		key, ok := b.bestFor(p, storer(res))
		if !ok {
			skip[res] = true
			continue
		}
		c := b.cost(p, key)
		if !p.affordable(c) {
			have := make(map[string]float64, len(p.amt))
			for k, v := range p.amt {
				have[k] = v
			}
			b.fundAtMarket(p, c, have)
		}
		if !b.tryBuild(p, key, "build_storage") {
			skip[res] = true
			continue
		}
		built = true
	}
	return built
}

// fundAtMarket buys, before leaving, what a plan item's price c needs and the
// hours ahead won't bring: each resource the budget lacks (the Industrial
// Age's stone and iron, the Atomic Age's iron and stone come only from the
// market). It sells what refills before the next visit anyway (income over
// the interval at least a full store), never what c itself needs, at most a
// quarter of a store per trade and a few trades per resource. The trades are
// real (`trade`); budget follows what changed hands.
func (b *Bot) fundAtMarket(p *plan, c map[string]float64, budget map[string]float64) {
	traders := 0
	for key, bs := range p.st.Buildings {
		if b.defs[key].LineageKey == "trade" {
			traders += bs.Count
		}
	}
	if traders == 0 {
		return
	}
	rates := p.st.Trade.ExchangeRates
	for _, r := range sortedKeys(c) {
		for i := 0; i < 6; i++ {
			short := math.Min(c[r]-budget[r], p.storage[r]-p.amt[r])
			if short <= 0 {
				break
			}
			from, best, sell := "", 0.0, 0.0
			for _, k := range sortedKeys(rates) {
				x := rates[k]
				if x.To != r || x.Rate <= 0 || p.st.Resources[x.From].Rate*b.CheckInTicks < p.storage[x.From] {
					continue
				}
				n := math.Min(math.Min(p.amt[x.From]-c[x.From], 0.25*p.storage[x.From]), short/x.Rate)
				if v := n * x.Rate; n >= 1 && v > best {
					from, best, sell = x.From, v, n
				}
			}
			if from == "" {
				break
			}
			got, err := b.exchange(from, r, sell)
			if !b.act("trade_plan", from+"->"+r, err) {
				break
			}
			p.amt[from] -= sell
			p.amt[r] += got
			budget[from] -= sell
			budget[r] += got
		}
	}
}

// traceCheckIn writes what a visit finds or leaves: each resource the age
// still needs, as held/cap (+rate per tick, still to make), slowest first.
func (b *Bot) traceCheckIn(what string, st game.GameState) {
	p := b.newPlan(st)
	var parts []string
	for _, res := range p.worst {
		r := st.Resources[res]
		parts = append(parts, fmt.Sprintf("%s %s/%s (%+.3g/t, need %s)", res, num(r.Amount), num(r.Storage), r.Rate, num(p.target[res])))
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
			case game.PlanTrade:
				s = fmt.Sprintf("trade %s->%s (got %s, want %s more) %s", v.Key, v.To, num(v.Got), num(v.Amount), v.Status)
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

// trade sells one resource for the slowest target resource at the market, as
// a player would with `trade`. It picks the amount that evens out the two
// resources' times to target (never selling one that would then become the
// slower of the two), skips rates pushed below 60% of base by recent trades,
// and makes at most one exchange per decision.
func (b *Bot) trade(p *plan) {
	traders := 0
	for key, bs := range p.st.Buildings {
		if b.defs[key].LineageKey == "trade" {
			traders += bs.Count
		}
	}
	if traders == 0 {
		return
	}
	rates := p.st.Trade.ExchangeRates
	if b.tradeForFood(p, rates) || b.tradeForBlocker(p, rates) || b.tradeOverflow(p, rates) {
		return
	}
	for _, want := range p.worst {
		room := math.Min(p.target[want], p.storage[want]) - p.amt[want]
		if room <= 0 || b.recently(b.sold, want) {
			continue
		}
		dW := p.target[want] - p.amt[want]
		rW := math.Max(p.st.Resources[want].Rate, 1e-6)
		from, rate, sell := "", 0.0, 0.0
		for _, k := range sortedKeys(rates) {
			x := rates[k]
			if x.To != want || x.Rate < x.BaseRate*0.6 || p.amt[x.From] < 1 || b.recently(b.bought, x.From) {
				continue
			}
			dS := p.target[x.From] - p.amt[x.From]
			rS := p.st.Resources[x.From].Rate
			var n float64
			switch {
			case rS > 0:
				// (dW - r n)/rW == (dS + n)/rS
				n = (float64(dW*rS) - float64(dS*rW)) / (float64(x.Rate*rS) + rW)
			case dS < 0:
				n = -dS // not produced: only the surplus is spare
			}
			// Don't sell what the most wanted purchase is being saved for,
			// unless it sits at its cap: with a target nothing produces, this
			// trade would otherwise drain the purchase's inputs as fast as
			// they come in (the Space Age's plasma and electricity, traded
			// away for titanium while the producers they would buy wait).
			spare := p.amt[x.From]
			if spare < 0.95*p.storage[x.From] {
				spare -= p.blockCost[x.From]
			}
			n = math.Min(math.Min(n, spare), 0.25*p.storage[x.From])
			n = math.Min(n, room/x.Rate)
			if n*x.Rate > sell*rate {
				from, rate, sell = x.From, x.Rate, n
			}
		}
		if from != "" && sell > 0 && sell < 1 && room < rate && p.amt[from]-p.blockCost[from] >= 1 {
			sell = 1 // less than one unit short: buy a little over (see tradeInto)
		}
		if from == "" || sell < 1 {
			continue
		}
		if got, err := b.exchange(from, want, sell); b.act("trade", from+"->"+want, err) {
			p.amt[from] -= sell
			p.amt[want] += got
			b.sold[from], b.bought[want] = b.tick, b.tick
		}
		return
	}
}

// tradeOverflow is a check-in player's trade: whatever will be lost at a cap
// before the next visit (held now plus what comes in by then, over the cap)
// is sold for the slowest target resource that still has room by then.
// Selling it costs nothing, since it would be gone anyway. Only with
// CheckInTicks set; reports whether it traded.
func (b *Bot) tradeOverflow(p *plan, rates map[string]game.ExchangeRateInfo) bool {
	if b.CheckInTicks <= 0 {
		return false
	}
	ahead := func(res string) float64 {
		return p.amt[res] + float64(math.Max(p.st.Resources[res].Rate, 0)*b.CheckInTicks)
	}
	for _, want := range p.worst {
		room := math.Min(p.target[want], p.storage[want]) - ahead(want)
		if room <= 0 || b.recently(b.sold, want) {
			continue
		}
		from, best, sell := "", 0.0, 0.0
		for _, k := range sortedKeys(rates) {
			x := rates[k]
			if x.To != want || x.Rate <= 0 || x.Rate < x.BaseRate*0.6 || b.recently(b.bought, x.From) {
				continue
			}
			over := math.Min(ahead(x.From)-p.storage[x.From], p.amt[x.From])
			n := math.Min(over, room/x.Rate)
			if v := n * x.Rate; n >= 1 && v > best {
				from, best, sell = x.From, v, n
			}
		}
		if from == "" {
			continue
		}
		got, err := b.exchange(from, want, sell)
		if !b.act("trade_overflow", from+"->"+want, err) {
			return false
		}
		p.amt[from] -= sell
		p.amt[want] += got
		b.sold[from], b.bought[want] = b.tick, b.tick
		return true
	}
	return false
}

// tradeForBlocker buys the resource blocking the most wanted purchase when
// the market sells it: it sells whatever the bot holds the most spare of
// (above what this age's targets need), at most a quarter of that cap per
// trade. Reports whether it traded.
func (b *Bot) tradeForBlocker(p *plan, rates map[string]game.ExchangeRateInfo) bool {
	want := p.blocker
	if want == "" {
		return false
	}
	return b.tradeInto(p, rates, want, math.Min(p.blockNeed, p.storage[want])-p.amt[want], p.blockCost)
}

// tradeForFood buys food when workers are starving and no food building is
// affordable (the Iron Age field works costs iron the bot may not have).
// Starvation drags morale down and every rate with it, a spiral a player
// would break at the market.
func (b *Bot) tradeForFood(p *plan, rates map[string]game.ExchangeRateInfo) bool {
	food := p.st.Resources["food"]
	if p.foodRate >= 0 || food.Amount > 0.1*food.Storage {
		return false
	}
	return b.tradeInto(p, rates, "food", float64(0.25*food.Storage)-food.Amount, nil)
}

// tradeInto sells whatever the bot holds the most spare of for up to short
// of want, in one trade. keep is a price whose inputs must not be sold.
func (b *Bot) tradeInto(p *plan, rates map[string]game.ExchangeRateInfo, want string, short float64, keep map[string]float64) bool {
	if short <= 0 || b.recently(b.sold, want) {
		return false
	}
	from, best, sell, fromSpare, fromRate := "", 0.0, 0.0, 0.0, 0.0
	for _, k := range sortedKeys(rates) {
		x := rates[k]
		if x.To != want || x.Rate < x.BaseRate*0.6 || x.Rate <= 0 || b.recently(b.bought, x.From) {
			continue
		}
		// Spare is what this age's targets and the purchase itself don't
		// need, or, for a resource sitting at its cap, whatever the cap is
		// wasting anyway. (Selling the purchase's other inputs would just
		// swap which one blocks it.)
		spare := p.amt[x.From] - p.target[x.From] - keep[x.From]
		if p.amt[x.From] >= 0.95*p.storage[x.From] {
			spare = math.Max(spare, math.Min(0.25*p.storage[x.From], p.amt[x.From]-keep[x.From]))
		}
		if spare <= 0 {
			continue
		}
		n := math.Min(math.Min(spare, 0.25*p.storage[x.From]), short/x.Rate)
		if v := n * x.Rate; v > best {
			from, best, sell, fromSpare, fromRate = x.From, v, n, spare, x.Rate
		}
	}
	// The market trades whole units. Short by less than one unit of the best
	// source (the last stone a wonder bank lacks, bought with gold at several
	// stone a coin), a player sells one and buys a little over; the bot does
	// too, when that unit is spare, instead of waiting for income that may
	// never come (an Endure can take an old resource's last producers).
	if from != "" && sell > 0 && sell < 1 && short < fromRate && fromSpare >= 1 {
		sell = 1
	}
	if from == "" || sell < 1 {
		return false
	}
	got, err := b.exchange(from, want, sell)
	if !b.act("trade", from+"->"+want, err) {
		return false
	}
	p.amt[from] -= sell
	p.amt[want] += got
	b.sold[from], b.bought[want] = b.tick, b.tick
	return true
}

// exchange trades amount of from for to at the market. With RecordPlan it
// also writes the trade into the run's plan log, as a trade item buying
// what this one got (GameEngine.NoteTradeForTest).
func (b *Bot) exchange(from, to string, amount float64) (float64, error) {
	got, err := b.ge.ExchangeResources(from, to, amount)
	if err == nil && b.RecordPlan && got > 0 {
		b.ge.NoteTradeForTest(from, to, got)
	}
	return got, err
}

// recently reports whether m records res within the last 150 ticks (5 min).
// The game tick restarts at prestige, so a record from a later tick than now
// is from the previous run and doesn't count.
func (b *Bot) recently(m map[string]int, res string) bool {
	t, ok := m[res]
	d := b.tick - t
	return ok && d >= 0 && d < 150
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

func (b *Bot) act(kind, detail string, err error) bool {
	if b.Trace != nil {
		res := "ok"
		if err != nil {
			res = err.Error()
		}
		fmt.Fprintf(b.Trace, "tick %d %s %s: %s\n", b.tick, kind, detail, res)
	}
	if err != nil {
		b.Errors[kind]++
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
	if def.MaxCount > 0 && bs.Count+p.queued[key]+p.extra[key] >= def.MaxCount {
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

// fits reports whether every part of cost fits under its cap; misses are
// recorded in capNeed so the storage pass raises them.
func (p *plan) fits(cost map[string]float64) bool {
	ok := true
	for r, c := range cost {
		if c > p.storage[r] {
			p.capNeed[r] = math.Max(p.capNeed[r], c)
			ok = false
		}
	}
	return ok
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

// costFrac is the price as the largest fraction of any cap it touches.
func (p *plan) costFrac(cost map[string]float64) float64 {
	f := 0.0
	for r, c := range cost {
		s := math.Max(p.storage[r], 1)
		f = math.Max(f, c/s)
	}
	return math.Max(f, 1e-9)
}

func (b *Bot) tryBuild(p *plan, key, kind string) bool {
	c := b.cost(p, key)
	if !p.fits(c) || !p.affordable(c) {
		return false
	}
	if !b.act(kind, key, b.ge.BuildBuilding(key)) {
		return false
	}
	if b.RecordPlan {
		b.ge.NotePlanForTest(game.PlanBuild, key, 1)
	}
	for r, v := range c {
		p.amt[r] -= v
	}
	p.extra[key]++
	for _, e := range b.defs[key].Effects {
		if e.Type == "storage" {
			if e.Target == "all" {
				for r := range p.storage {
					p.storage[r] += float64(e.Value * p.storeK)
				}
			} else {
				p.storage[e.Target] += float64(e.Value * p.storeK)
			}
		}
	}
	return true
}

// buyOrBootstrap buys key if it can. If not, it records the resource that
// blocks it longest as the plan's blocker and, with depth left, tries to buy
// a producer of that resource instead. Reports whether key itself was bought.
func (b *Bot) buyOrBootstrap(p *plan, key, kind string, depth int) bool {
	c := b.cost(p, key)
	if !p.fits(c) {
		// Over a cap: fits noted it in capNeed and the storage pass comes
		// first. Chasing the capped resource itself would only waste it.
		return false
	}
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
		p.blocker, p.blockNeed, p.blockCost = block, c[block], c
	}
	if depth > 0 {
		if k2, ok := b.bestFor(p, producer(block)); ok && k2 != key {
			b.buyOrBootstrap(p, k2, "build_bootstrap", depth-1)
		}
	}
	return false
}

// upgrade converts legacy buildings to their next tier, one copy at a time.
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
		per := make(map[string]float64, len(u.Cost))
		for r, c := range u.Cost {
			per[r] = c / float64(u.Count)
		}
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
		goal := math.Max(math.Max(p.target[res], r.Storage), 1)
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
	if p.foodRate < 0 || food.Amount < 0.2*food.Storage {
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
	if food.Amount < 0.3*food.Storage {
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
		if !p.fits(c) {
			continue
		}
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

func storer(res string) func(config.Effect) float64 {
	return func(e config.Effect) float64 {
		if e.Type == "storage" && (e.Target == res || e.Target == "all") && e.Value > 0 {
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

	// 3. Storage for anything that must fit under a cap. A storage building
	// the bot can't pay for records its blocker, so the market can supply it
	// (the Renaissance vault's stone, which nothing in that age produces).
	for _, res := range sortedKeys(p.capNeed) {
		for i := 0; i < 50 && p.capNeed[res] > p.storage[res]*0.98; i++ {
			key, ok := b.bestFor(p, storer(res))
			if !ok || !b.buyOrBootstrap(p, key, "build_storage", 0) {
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
		keep := p.st.NextAgeResReqs[res]
		onlyWonder := p.target[res]-left <= keep
		switch {
		case p.invest && !onlyWonder && b.CheckInTicks > 0:
			// Bank what would be lost at the cap before the next visit. With
			// overflow on the game does that as it happens, and what the
			// store holds now is better left to the build plan.
			if p.st.WonderOverflow {
				continue
			}
			over := p.amt[res] + float64(math.Max(p.st.Resources[res].Rate, 0)*b.CheckInTicks) - p.storage[res]
			if over <= 0 {
				continue
			}
			keep = math.Max(keep, p.amt[res]-over)
		case p.invest && !onlyWonder:
			if p.amt[res] < 0.9*p.storage[res] {
				continue
			}
			keep = math.Max(keep, 0.5*p.storage[res])
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

// research starts the cheapest available tech the knowledge surplus pays
// for, except that a tech producing a target resource nothing else is
// producing goes first (steel forging when the Renaissance asks for steel
// and no Medieval building makes it).
func (b *Bot) research(p *plan) {
	rs := p.st.Research
	if rs.CurrentTech != "" {
		return
	}
	k := p.amt["knowledge"]
	req := p.st.NextAgeResReqs["knowledge"]
	capK := p.storage["knowledge"]
	rate := p.st.Resources["knowledge"].Rate
	keys := sortedKeys(rs.Techs)
	techs := config.TechByKey()
	unblocks := func(key string) bool {
		for _, e := range techs[key].Effects {
			if e.Type == "production" && p.target[e.Target] > p.amt[e.Target] && p.st.Resources[e.Target].Rate <= 0 {
				return true
			}
		}
		return false
	}
	sort.SliceStable(keys, func(i, j int) bool {
		ui, uj := unblocks(keys[i]), unblocks(keys[j])
		if ui != uj {
			return ui
		}
		return rs.Techs[keys[i]].Cost < rs.Techs[keys[j]].Cost
	})
	for _, key := range keys {
		t := rs.Techs[key]
		if !t.Available || t.Researched || t.Cost > k {
			continue
		}
		ok := k-t.Cost >= req
		if !ok && k >= 0.98*capK && rate > 0 {
			// Knowledge is being wasted at the cap; spend it if refilling
			// takes less time than the slowest other requirement.
			other := 0.0
			for _, res := range p.worst {
				if res != "knowledge" {
					other = math.Max(other, p.eta[res])
				}
			}
			ok = t.Cost/rate < other
		}
		if ok && b.act("research", key, b.ge.StartResearch(key)) {
			p.amt["knowledge"] -= t.Cost
		}
		return
	}
}

// festival spends culture on a production buff when this age does not ask
// for culture and it is piling up.
func (b *Bot) festival(p *plan) {
	c, ok := p.st.Resources["culture"]
	if !ok || !c.Unlocked || p.st.NextAgeResReqs["culture"] > 0 || c.Amount < 0.8*c.Storage {
		return
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
	if g := p.blocker; (g == "food" || g == "wood" || g == "stone") && p.st.Resources[g].Unlocked && p.amt[g] < p.storage[g] {
		pick = g
	}
	for _, res := range p.worst {
		if pick != "" {
			break
		}
		if res == "food" || res == "wood" || res == "stone" {
			if p.st.Resources[res].Unlocked && p.amt[res] < p.storage[res] {
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
		p.amt[pick] += gatherYield
	}
}
