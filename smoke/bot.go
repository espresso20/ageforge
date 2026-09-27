package smoke

import (
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
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
//     otherwise even out the slowest target; never undo a recent trade
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

	// Actions counts successful player actions by kind; Errors counts
	// rejected ones. Both feed the report.
	Actions map[string]int
	Errors  map[string]int

	// Trace, if set, receives one line per attempted action.
	Trace io.Writer
	// Harbinger is the policy for harbinger answers: ignore, appease, brace
	// or both (level 1 only).
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
	for bld, n := range st.NextAgeBldReqs {
		have := st.Buildings[bld].Count + p.queued[bld]
		if have < n {
			p.needBld[bld] = n - have
			for res, c := range st.Buildings[bld].NextCost {
				p.target[res] += c * float64(n-have)
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
		b.act("build_wonder", w, b.ge.BuildBuilding(w))
	}
	if b.answerHarbinger(st) {
		st = b.ge.GetState()
	}
	p := b.newPlan(st)
	if b.upgrade(p) {
		p = b.newPlan(b.ge.GetState())
	}
	pop := b.recruit(p)
	b.assign(p, pop)
	b.build(p)
	b.bankWonder(p)
	b.research(p)
	b.festival(p)
	b.trade(p)
	b.gather(p)
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
	if b.tradeForFood(p, rates) || b.tradeForBlocker(p, rates) {
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
				n = (dW*rS - dS*rW) / (x.Rate*rS + rW)
			case dS < 0:
				n = -dS // not produced: only the surplus is spare
			}
			n = math.Min(math.Min(n, p.amt[x.From]), 0.25*p.storage[x.From])
			n = math.Min(n, room/x.Rate)
			if n*x.Rate > sell*rate {
				from, rate, sell = x.From, x.Rate, n
			}
		}
		if from == "" || sell < 1 {
			continue
		}
		if got, err := b.ge.ExchangeResources(from, want, sell); b.act("trade", from+"->"+want, err) {
			p.amt[from] -= sell
			p.amt[want] += got
			b.sold[from], b.bought[want] = b.tick, b.tick
		}
		return
	}
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
	return b.tradeInto(p, rates, "food", 0.25*food.Storage-food.Amount, nil)
}

// tradeInto sells whatever the bot holds the most spare of for up to short
// of want, in one trade. keep is a price whose inputs must not be sold.
func (b *Bot) tradeInto(p *plan, rates map[string]game.ExchangeRateInfo, want string, short float64, keep map[string]float64) bool {
	if short <= 0 || b.recently(b.sold, want) {
		return false
	}
	from, best, sell := "", 0.0, 0.0
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
			from, best, sell = x.From, v, n
		}
	}
	if from == "" || sell < 1 {
		return false
	}
	got, err := b.ge.ExchangeResources(from, want, sell)
	if !b.act("trade", from+"->"+want, err) {
		return false
	}
	p.amt[from] -= sell
	p.amt[want] += got
	b.sold[from], b.bought[want] = b.tick, b.tick
	return true
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
	mult := math.Pow(b.defs[key].CostScale, float64(p.extra[key]))
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
	for r, v := range c {
		p.amt[r] -= v
	}
	p.extra[key]++
	for _, e := range b.defs[key].Effects {
		if e.Type == "storage" {
			if e.Target == "all" {
				for r := range p.storage {
					p.storage[r] += e.Value
				}
			} else {
				p.storage[e.Target] += e.Value
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
		p.foodRate -= per * float64(n)
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
			score += v * w[r]
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
		if p.invest && !onlyWonder {
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
		if b.act("bank_wonder", fmt.Sprintf("%s %.0f", res, dep), b.ge.BankWonderResource(w, res, dep)) {
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
