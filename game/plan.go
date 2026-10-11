package game

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

// The build plan: an ordered list of builds and techs the player wants, which
// the engine starts on its own as resources come in, during play and during
// offline catch-up. It is how a player who checks in a few times a day puts
// the hours between visits to work.
//
// Rules (site/docs/plan.md and the economy design):
//
//   - An item is paid for when it starts, not when it is queued, through the
//     same checks as `build` and `research` (age lock, MaxCount, one research
//     at a time, knowledge cost).
//   - Every tick, after production, the engine walks the plan in order and
//     starts everything it can. A build item with a count starts as many
//     copies as it can afford, one at a time along the cost curve.
//   - Skipping with reservation: an item that cannot start yet does not block
//     the items after it, but it reserves its next price. A later item only
//     starts if it can be paid from what is left after every earlier waiting
//     item's reservation, so a cheap item further down can never delay one
//     above it, while resources the top items don't need are not left idle.
//     Items that cannot start for a reason money won't fix reserve nothing: a
//     price over the current storage cap (the next storage building has to
//     come first), a resource it lacks that the current income won't bring
//     in within a day (it needs the market or a producer first), or a wonder
//     (see below).
//   - A wonder's price is what its bank still lacks. Once what is held,
//     after the reservations above it, covers all of that, the plan banks it
//     and starts the wonder. While it waits it reserves nothing: it is 40
//     price units of its age, and holding that back would stall everything
//     below it for hours. Deposits, wonder overflow and the plan's overflow
//     (a queued wonder takes it in plan order, with wonder overflow on or
//     off) fill the bank; a part bigger than a full store can only be banked
//     that way.
//   - Overflow pays the plan (overflow.go, bankPlanOverflow): what a full
//     store would throw away, after the age's wonder has taken what it needs,
//     goes into the banks of the plan's build items, in plan order, each up
//     to what its next copy still lacks. A banked item's price in the walk is
//     what its bank doesn't cover: that is what it reserves while it waits,
//     what must fit under the cap, and what it pays from the stores when it
//     starts (the bank pays the rest). So a copy priced over a cap becomes
//     buyable once overflow has banked the part over it. Only builds that can
//     start in this age bank: the next age's buildings wait for the advance.
//     A queued wonder banks into its own bank. Nothing is built that was not
//     queued.
//   - Deal items (plan_deal.go) take a civilization's trade deal once its
//     fixed price is free, reserving it while they wait, like one build.
//   - Techs start in plan order: only the first research item can take the
//     research slot when it frees up. Later research items still reserve
//     their knowledge.
//   - Items that can never start drop out with a log line: a building of an
//     earlier age after an advance, a building at its MaxCount, a tech
//     already researched.
//   - Planning a tech plans what it still needs before it
//     (plan_research.go), so a tech in the plan has what it needs
//     researched, being researched or planned above it. One that loses
//     that (its prerequisite was removed from above it or moved below it,
//     the research it waited on was canceled, or a game update changed
//     what the tech needs) stays and waits: it could still start, so it is
//     not dropped. While it waits it holds neither the research slot's
//     turn nor its knowledge, so a prerequisite planned below it can go
//     first. A tech of an age not reached yet waits for the age the same
//     way.
//
// The plan is saved, and it is cleared by prestige, Succumb and a new game.

// Plan item kinds.
const (
	PlanBuild    = "build"
	PlanResearch = "research"
	PlanTrade    = "trade"
	PlanAdvance  = "advance"
)

// MaxPlanItems caps the plan's length: 60 items, room for a day away on the
// one-week curve (an age's storage, its required buildings, a few producers
// per resource, the wonder and the advance, then the next age's opening
// moves) with some to spare. It was 30 before Pacing v2's away-proofing.
// Research items do not count against it (planLoad): a research queue is
// bounded by the tree, each tech once, and one command may queue a long
// chain. maxPlanCount caps one build item's count.
const (
	MaxPlanItems = 60
	maxPlanCount = 1000
)

// planLoad is how many of items count against MaxPlanItems: every item but
// the research ones.
func planLoad(items []PlanItem) int {
	n := 0
	for _, it := range items {
		if it.Kind != PlanResearch {
			n++
		}
	}
	return n
}

// PlanLoad is how many of a plan's items count against MaxPlanItems (every
// one but research), for the panels.
func PlanLoad(views []PlanItemView) int {
	n := 0
	for _, v := range views {
		if v.Kind != PlanResearch {
			n++
		}
	}
	return n
}

// PlanItem is one entry of the build plan, as saved. Count is how many copies
// of a build item are still to start; Started counts those already started
// by the plan. A research item has Count 1 until it starts, when it leaves
// the plan.
type PlanItem struct {
	Kind    string `json:"kind"`
	Key     string `json:"key"`
	Count   int    `json:"count"`
	Started int    `json:"started,omitempty"`
	// Trade items (plan_trade.go): Key is what is sold, To what is bought,
	// Amount how much of To is still wanted (0: no limit) and Got how much
	// the item has bought so far. Count stays 1 until Amount is met.
	To     string  `json:"to,omitempty"`
	Amount float64 `json:"amount,omitempty"`
	Got    float64 `json:"got,omitempty"`
	// Deal items (plan_deal.go): Key is the civilization, Deal the offer's ID.
	Deal int `json:"deal,omitempty"`
	// Banked is what overflow has set aside toward a build item's next copy,
	// by resource, at most that copy's price (overflow.go). The copy pays
	// from it first when it starts. Removing the item, a drop, an advance and
	// the last copy starting put what is left back in the stores, up to their
	// caps.
	Banked map[string]float64 `json:"banked,omitempty"`
}

// Plan item statuses, for the UI.
const (
	PlanStatusReady   = "ready"   // affordable now; starts on the next tick
	PlanStatusWaiting = "waiting" // saving up; Progress says how far along
	PlanStatusBlocked = "blocked" // waiting on something money won't fix; Note says what
)

// PlanItemView is one plan item for the UI: the item, the price of its next
// start, and whether it could start now after the items above it.
type PlanItemView struct {
	Kind    string
	Key     string
	Name    string
	Count   int
	Started int
	// To, Amount and Got are a trade item's (see PlanItem).
	To     string
	Amount float64
	Got    float64
	// Cost is the price of the next start (for a wonder, what its bank still lacks).
	Cost map[string]float64
	// Banked is the part of Cost a build item's overflow bank already holds
	// (nil when it holds nothing).
	Banked map[string]float64
	// Status is PlanStatusReady, PlanStatusWaiting or PlanStatusBlocked.
	Status string
	// Progress is how much of Cost is covered by the bank and what is free
	// after the reservations above this item, 0 to 1 (the least-covered
	// resource).
	Progress float64
	// Short is the resource furthest from covered ("" when ready or blocked).
	Short string
	// Note explains a blocked item ("research slot busy", ...).
	Note string
}

// clonePlan copies a plan so snapshots and saves share nothing with the live
// one, banks included.
func clonePlan(p []PlanItem) []PlanItem {
	if len(p) == 0 {
		return nil
	}
	out := append([]PlanItem(nil), p...)
	for i := range out {
		out[i].Banked = cloneBank(out[i].Banked)
	}
	return out
}

// cloneBank copies a plan item's bank (nil for an empty one).
func cloneBank(b map[string]float64) map[string]float64 {
	if len(b) == 0 {
		return nil
	}
	out := make(map[string]float64, len(b))
	for k, v := range b {
		out[k] = v
	}
	return out
}

// loadBank is a saved bank as the engine accepts it: positive, finite
// amounts of known resources only (nil if nothing is left).
func loadBank(b map[string]float64, known func(string) bool) map[string]float64 {
	var out map[string]float64
	for k, v := range b {
		if !(v > 0) || math.IsInf(v, 0) || !known(k) {
			continue
		}
		if out == nil {
			out = make(map[string]float64, len(b))
		}
		out[k] = v
	}
	return out
}

// loadPlanIn is a saved plan as the engine accepts it: known kinds only,
// counts in range, at most MaxPlanItems besides research and no more
// research items than the tree has techs, banks on build items only, holding
// positive, finite amounts of resources set knows. A plan the game wrote
// passes unchanged.
func loadPlanIn(set *rules.Set, saved []PlanItem) []PlanItem {
	known := func(res string) bool {
		_, ok := set.Resource(res)
		return ok
	}
	var out []PlanItem
	for _, it := range saved {
		if it.Kind == PlanBuild {
			it.Banked = loadBank(it.Banked, known)
		} else {
			it.Banked = nil
		}
		if (it.Kind != PlanBuild && it.Kind != PlanResearch && it.Kind != PlanTrade && it.Kind != PlanAdvance && it.Kind != PlanDeal) || it.Count <= 0 {
			continue
		}
		if load := planLoad(out); it.Kind != PlanResearch && load >= MaxPlanItems || it.Kind == PlanResearch && len(out)-load >= set.TechCount() {
			continue
		}
		if it.Kind == PlanDeal && (it.Key == "" || it.Deal <= 0) {
			continue
		}
		if it.Kind == PlanTrade && (it.To == "" || !(it.Amount >= 0) || math.IsInf(it.Amount, 0) || !(it.Got >= 0)) {
			continue
		}
		if it.Kind != PlanBuild {
			it.Count = 1
		}
		it.Count = min(it.Count, maxPlanCount)
		it.Started = max(it.Started, 0)
		out = append(out, it)
	}
	return out
}

// ===== Public API =====

// PlanAddBuild appends count copies of building key to the plan (merged into
// the last item when that is the same building) and returns how many it
// added: fewer than count when the building's MaxCount leaves less room. The
// building must be one `build` would accept in this age, ignoring cost.
func (ge *GameEngine) PlanAddBuild(key string, count int) (int, error) {
	if count <= 0 || count > maxPlanCount {
		return 0, fmt.Errorf("The count must be a whole number from 1 to %d.", maxPlanCount)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	def, ok := ge.Buildings.defs[key]
	if !ok {
		if s := ge.Buildings.SuggestKey(key); s != "" {
			return 0, fmt.Errorf("Unknown building '%s'. Did you mean '%s'?", key, s)
		}
		return 0, fmt.Errorf("Unknown building '%s'. Type build to see what you can build.", key)
	}
	planned := plannedCopies(ge.plan, key)
	if reason := ge.planBuildInvalid(key, planned); reason != "" {
		return 0, fmt.Errorf("Can't plan %s: %s.", def.Name, reason)
	}
	if def.MaxCount > 0 {
		room := def.MaxCount - ge.Buildings.GetCount(key) - ge.Buildings.GetQueueCount(key, ge.buildQueue) - planned
		count = min(count, room)
	}
	if n := len(ge.plan); n > 0 && ge.plan[n-1].Kind == PlanBuild && ge.plan[n-1].Key == key {
		add := min(count, maxPlanCount-ge.plan[n-1].Count)
		if add <= 0 {
			return 0, fmt.Errorf("That plan item already holds %d, the most one item can.", maxPlanCount)
		}
		ge.plan[n-1].Count += add
		ge.logPlanAddLocked(ge.plan[n-1], add)
		return add, nil
	}
	if planLoad(ge.plan) >= MaxPlanItems {
		return 0, errPlanFull()
	}
	ge.plan = append(ge.plan, PlanItem{Kind: PlanBuild, Key: key, Count: count})
	ge.notePlanSize()
	ge.logPlanAddLocked(ge.plan[len(ge.plan)-1], count)
	return count, nil
}

// PlanAddResearch plans tech key, with what it still needs before it
// (PlanAddResearchChain, plan_research.go, which also says what was added).
func (ge *GameEngine) PlanAddResearch(key string) error {
	_, err := ge.PlanAddResearchChain(key)
	return err
}

// PlanRemove removes item n (1-based) and returns a description of it. What
// overflow had banked for it goes back to the stores, up to the caps.
func (ge *GameEngine) PlanRemove(n int) (string, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if n < 1 || n > len(ge.plan) {
		return "", ge.planIndexErr(n)
	}
	it := ge.plan[n-1]
	ge.plan = append(ge.plan[:n-1:n-1], ge.plan[n:]...)
	ge.unlogPlanItemLocked(it)
	label := ge.planItemLabel(it)
	back := map[string]float64{}
	if ge.returnPlanBank(&it, back) {
		ge.logBankReturn("the bank of "+label, "", back)
	}
	return label, nil
}

// PlanClear empties the plan and returns how many items it held. What
// overflow had banked goes back to the stores, up to the caps.
func (ge *GameEngine) PlanClear() int {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	n := len(ge.plan)
	ge.returnPlanBanks("")
	for _, it := range ge.plan {
		ge.unlogPlanItemLocked(it)
	}
	ge.plan = nil
	return n
}

// PlanMove moves item n (1-based) by delta places (negative is up, toward
// the front) and returns its new position. Moves past either end stop there.
func (ge *GameEngine) PlanMove(n, delta int) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if n < 1 || n > len(ge.plan) {
		return 0, ge.planIndexErr(n)
	}
	to := min(max(n-1+delta, 0), len(ge.plan)-1)
	it := ge.plan[n-1]
	rest := append(ge.plan[:n-1:n-1], ge.plan[n:]...)
	out := make([]PlanItem, 0, len(ge.plan))
	out = append(out, rest[:to]...)
	out = append(out, it)
	out = append(out, rest[to:]...)
	ge.plan = out
	return to + 1, nil
}

func (ge *GameEngine) planIndexErr(n int) error {
	if len(ge.plan) == 0 {
		return fmt.Errorf("The plan is empty.")
	}
	return fmt.Errorf("There is no plan item %d (the plan has %s).", n, textfmt.Count(len(ge.plan), "item", "items"))
}

// errPlanFull is the refusal when the plan holds MaxPlanItems items besides
// its research.
func errPlanFull() error {
	return fmt.Errorf("The plan is full (%d items, not counting research). Remove one with plan remove <n> first.", MaxPlanItems)
}

// planItemLabel is "3 Huts", "research Pottery", "trade iron ore for food"
// or, with an amount left to buy, "trade iron ore for 500 food".
func (ge *GameEngine) planItemLabel(it PlanItem) string {
	if it.Kind == PlanTrade {
		if it.Amount > 0 {
			return "trade " + ResourceName(it.Key) + " for " + Amount(it.Amount, it.To)
		}
		return "trade " + ResourceName(it.Key) + " for " + ResourceName(it.To)
	}
	if it.Kind == PlanAdvance {
		return "advance when ready"
	}
	if it.Kind == PlanDeal {
		return ge.planDealLabel(it)
	}
	if it.Kind == PlanResearch {
		return "research " + techDefName(ge.rules, it.Key)
	}
	return buildingCountIn(ge.rules, it.Count, it.Key)
}

// ===== Validity =====

// plannedCopies is how many copies of key items holds.
func plannedCopies(items []PlanItem, key string) int {
	n := 0
	for _, it := range items {
		if it.Kind == PlanBuild && it.Key == key {
			n += it.Count
		}
	}
	return n
}

// planBuildInvalid is why a build of key can never start in this age ("" if
// it can). planned is how many copies earlier plan items will take first,
// which only matters for MaxCount.
func (ge *GameEngine) planBuildInvalid(key string, planned int) string {
	def, ok := ge.Buildings.defs[key]
	if !ok {
		return "unknown building"
	}
	// The next age's buildings may be planned ahead: they wait (checkPlanItem)
	// until an advance unlocks them.
	future := def.RequiredAge != "" && def.RequiredAge == ge.progress.GetNextAge(ge.age)
	// A building its tech still holds back may be planned too: it waits
	// (checkPlanItem) until the research is done.
	if !ge.Buildings.AgeUnlocked(key) && !future {
		return "not unlocked yet"
	}
	if def.RequiredAge != "" && def.RequiredAge != ge.age && !future {
		return "it belongs to another age"
	}
	if def.MaxCount > 0 {
		have := ge.Buildings.GetCount(key) + ge.Buildings.GetQueueCount(key, ge.buildQueue) + planned
		if have >= def.MaxCount {
			if def.Category == "wonder" {
				return "already built or planned"
			}
			return fmt.Sprintf("at its limit of %d", def.MaxCount)
		}
	}
	return ""
}

// planResearchInvalid is why tech key can never start from the plan ("" if
// it can): the plan drops an item it is true of. A tech of an age not
// reached yet is not one: it waits for the age (planTechAhead).
func (ge *GameEngine) planResearchInvalid(key string) string {
	if _, ok := ge.rules.Tech(key); !ok {
		return "unknown technology"
	}
	if ge.Research.IsResearched(key) {
		return "already researched"
	}
	if ge.Research.currentTech == key {
		return "already being researched"
	}
	return ""
}

// planResearchMissing names what tech key at plan position idx still needs
// that is not on its way: a prerequisite neither researched, being
// researched nor planned before idx ("Tool Making"), or its either-or group
// when no key of it is ("Map Making or Boatbuilding", either true). "" when
// everything it needs is on its way.
func (ge *GameEngine) planResearchMissing(key string, idx int) (names string, either bool) {
	def, ok := ge.rules.Tech(key)
	if !ok {
		return "", false
	}
	onItsWay := func(pre string) bool {
		if ge.Research.IsResearched(pre) || ge.Research.currentTech == pre {
			return true
		}
		for i := 0; i < idx && i < len(ge.plan); i++ {
			if ge.plan[i].Kind == PlanResearch && ge.plan[i].Key == pre {
				return true
			}
		}
		return false
	}
	if pre := def.MissingPrereq(onItsWay); pre != "" {
		return techDefName(ge.rules, pre), false
	}
	if !def.AnyOfMet(onItsWay) {
		return techNamesOr(ge.rules, def.AnyOf), true
	}
	return "", false
}

// planResearchRefused is why tech key cannot be added to the plan at
// position idx ("" if it can): it could never start, or what it needs is
// not researched, being researched or planned before it.
func (ge *GameEngine) planResearchRefused(key string, idx int) string {
	if reason := ge.planResearchInvalid(key); reason != "" {
		return reason
	}
	switch names, either := ge.planResearchMissing(key, idx); {
	case names == "":
		return ""
	case either:
		return fmt.Sprintf("it needs %s first, and none of them is researched or planned before it", names)
	default:
		return fmt.Sprintf("it needs %s first, which isn't researched or planned before it", names)
	}
}

// ===== Execution =====

// planStarts collects what the plan started, for one summary log line.
type planStarts struct {
	builds map[string]int
	order  []string // build keys in first-start order
	techs  []string
	// trades sums what trade items sold and bought, by "from>to".
	trades     map[string][2]float64
	tradeOrder []string
	// advanced is the age an advance item moved to ("" if none).
	advanced string
	// deals are the deal items taken, as labels.
	deals []string
}

func (s *planStarts) addTrade(from, to string, sold, got float64) {
	if s.trades == nil {
		s.trades = map[string][2]float64{}
	}
	k := from + ">" + to
	if _, ok := s.trades[k]; !ok {
		s.tradeOrder = append(s.tradeOrder, k)
	}
	t := s.trades[k]
	s.trades[k] = [2]float64{t[0] + sold, t[1] + got}
}

func (s *planStarts) addBuild(key string) {
	if s.builds == nil {
		s.builds = map[string]int{}
	}
	if s.builds[key] == 0 {
		s.order = append(s.order, key)
	}
	s.builds[key]++
}

func (s *planStarts) empty() bool {
	return len(s.order) == 0 && len(s.techs) == 0 && len(s.tradeOrder) == 0 && s.advanced == "" && len(s.deals) == 0
}

// describe renders the starts as verb-led clauses: "started building 2 Huts
// and 1 Farm, started researching Pottery, traded 5K coal for 1.2K food".
// Callers prefix it ("Plan: ", "While you were away, your plan "). set names
// the buildings and techs.
func (s *planStarts) describe(set *rules.Set) string {
	var parts []string
	if len(s.order) > 0 {
		builds := make([]string, 0, len(s.order))
		for _, k := range s.order {
			builds = append(builds, buildingCountIn(set, s.builds[k], k))
		}
		parts = append(parts, "started building "+textfmt.List(builds))
	}
	if len(s.techs) > 0 {
		names := make([]string, 0, len(s.techs))
		for _, k := range s.techs {
			names = append(names, techDefName(set, k))
		}
		parts = append(parts, "started researching "+textfmt.List(names))
	}
	for _, k := range s.tradeOrder {
		from, to, _ := strings.Cut(k, ">")
		t := s.trades[k]
		parts = append(parts, fmt.Sprintf("traded %s for %s", Amount(t[0], from), Amount(t[1], to)))
	}
	parts = append(parts, s.deals...)
	if s.advanced != "" {
		parts = append(parts, "advanced to the "+s.advanced)
	}
	return strings.Join(parts, ", ")
}

// techDefName is a tech's name in set ("" for a key set does not know).
func techDefName(set *rules.Set, key string) string {
	def, _ := set.Tech(key)
	return def.Name
}

// techNamesOr is the names of techs keys in set as a choice: "Map Making or
// Boatbuilding".
func techNamesOr(set *rules.Set, keys []string) string {
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = techDefName(set, k)
	}
	return orList(names)
}

// planCheck is the plan walk's verdict on one item's next start.
type planCheck struct {
	// cost is what the start takes from the stores: for a banked build, the
	// part of its price the bank doesn't cover.
	cost    map[string]float64
	blocked string // why it can't start for a reason money won't fix
	reserve bool   // whether a waiting item holds its price back from later items
	// price and banked are a banked build's whole price and the part of it
	// its bank covers (nil for an item with nothing banked).
	price, banked map[string]float64
}

// checkPlanItem works out whether item it could start now. researchFirst is
// whether it is the first research item still in the plan.
func (ge *GameEngine) checkPlanItem(it PlanItem, researchFirst bool) planCheck {
	switch it.Kind {
	case PlanBuild:
		def := ge.Buildings.defs[it.Key]
		if def.RequiredAge != "" && def.RequiredAge != ge.age {
			return planCheck{blocked: "waits for the " + ge.progress.GetAgeName(def.RequiredAge)}
		}
		if ge.Buildings.TechLocked(it.Key) {
			return planCheck{blocked: "waits for " + ge.techName(def.RequiredTech)}
		}
		if def.Category == "wonder" {
			if DevGodMode || ge.Buildings.IsWonderBankFull(it.Key) {
				return planCheck{}
			}
			// The rest of the bank is the price: it is paid from what is
			// held once that covers all of it. A part bigger than a full
			// store is banked in rounds (deposits, overflow), and a wonder
			// holds nothing back while it waits, like one whose bank the
			// income won't fill.
			rest := ge.wonderBankRest(it.Key)
			for _, res := range sortedKeys(rest) {
				if rest[res] > ge.Resources.GetStorage(res) {
					return planCheck{cost: rest, blocked: "bank not full"}
				}
			}
			return planCheck{cost: rest}
		}
		cost, _ := ge.Buildings.BuildBatchCost(it.Key, 1, ge.buildQueue)
		if DevGodMode {
			return planCheck{}
		}
		// Overflow's bank pays first: only the rest must fit under the caps,
		// come in, and be held back while the item waits.
		chk := planCheck{cost: cost}
		if len(it.Banked) > 0 {
			chk.price = cost
			chk.cost, chk.banked = splitBank(cost, it.Banked)
		}
		for _, res := range sortedKeys(chk.cost) {
			if chk.cost[res] > ge.Resources.GetStorage(res) && chk.cost[res] > ge.Resources.Get(res) {
				chk.blocked = "needs more " + res + " storage"
				return chk
			}
		}
		if res := ge.planUnfunded(chk.cost); res != "" {
			chk.blocked = "too little " + res + " coming in"
			return chk
		}
		chk.reserve = true
		return chk
	case PlanResearch:
		def, _ := ge.rules.Tech(it.Key)
		if order := ge.progress.GetAgeOrder(); order[def.Age] > order[ge.age] {
			return planCheck{blocked: "waits for the " + ge.progress.GetAgeName(def.Age)}
		}
		cost := map[string]float64{"knowledge": def.Cost}
		if DevGodMode {
			cost = nil
		}
		if def.Cost > ge.Resources.GetStorage("knowledge") && !DevGodMode {
			return planCheck{cost: cost, blocked: "needs more knowledge storage"}
		}
		if res := ge.planUnfunded(cost); res != "" {
			return planCheck{cost: cost, blocked: "too little " + res + " coming in"}
		}
		if ge.Research.currentTech != "" {
			return planCheck{cost: cost, blocked: "research slot busy", reserve: true}
		}
		if !researchFirst {
			return planCheck{cost: cost, blocked: "after the research above it", reserve: true}
		}
		if pre := def.MissingPrereq(ge.Research.IsResearched); pre != "" {
			return planCheck{cost: cost, blocked: "needs " + techDefName(ge.rules, pre) + " first", reserve: true}
		}
		if !def.AnyOfMet(ge.Research.IsResearched) {
			return planCheck{cost: cost, blocked: "needs " + techNamesOr(ge.rules, def.AnyOf) + " first", reserve: true}
		}
		return planCheck{cost: cost, reserve: true}
	}
	return planCheck{blocked: "unknown item"}
}

// wonderBankRest is what wonder w's bank still lacks, by resource.
func (ge *GameEngine) wonderBankRest(w string) map[string]float64 {
	bank := ge.Buildings.wonderBanks[w]
	rest := map[string]float64{}
	for res, need := range ge.Buildings.defs[w].BaseCost {
		if left := need - bank[res]; left > 0.001 {
			rest[res] = left
		}
	}
	return rest
}

// payWonderBank banks what wonder w still lacks from what is held, so a
// planned wonder can start. The caller has checked that it is covered.
func (ge *GameEngine) payWonderBank(w string) error {
	rest := ge.wonderBankRest(w)
	for _, res := range sortedKeys(rest) {
		if _, err := ge.Buildings.BankResource(w, res, rest[res], ge.Resources); err != nil {
			return err
		}
	}
	return nil
}

// planFundTicks is how far ahead the plan looks for income: MaxOfflineTime,
// the longest the plan ever runs unattended, in ticks at 1x.
var planFundTicks = MaxOfflineTime.Seconds() / BaseTickInterval.Seconds()

// planUnfunded is a resource cost needs more of than is held and that the
// current income won't bring in within planFundTicks (a day), or "" if there
// is none. Such an item can't start until the player acts (trades, builds a
// producer), so it reserves nothing: holding resources back for it would
// stall the plan behind it.
func (ge *GameEngine) planUnfunded(cost map[string]float64) string {
	for _, res := range sortedKeys(cost) {
		short := cost[res] - ge.Resources.Get(res)
		if short <= 0 {
			continue
		}
		if rate := ge.Resources.GetRate(res); rate <= 0 || short/rate > planFundTicks {
			return res
		}
	}
	return ""
}

// planFree is what is left of res after the reservations so far.
func (ge *GameEngine) planFree(res string, reserved map[string]float64) float64 {
	return ge.Resources.Get(res) - reserved[res]
}

// planCovers reports whether cost can be paid from what the reservations leave.
func (ge *GameEngine) planCovers(cost map[string]float64, reserved map[string]float64) bool {
	for res, c := range cost {
		if !(ge.planFree(res, reserved) >= c) {
			return false
		}
	}
	return true
}

// runPlan walks the plan and starts everything it can (see the rules at the
// top of this file). It drops items that can never start, logging each, and
// adds what it started to starts. Must be called with the write lock held.
// Reports whether anything started.
func (ge *GameEngine) runPlan(starts *planStarts) bool {
	if len(ge.plan) == 0 {
		return false
	}
	started := false
	reserved := map[string]float64{}
	researchSeen := false
	out := make([]PlanItem, 0, len(ge.plan))
	advanceAt := -1
	advanceIdx := 0
	for i, it := range ge.plan {
		if it.Kind == PlanAdvance {
			if ge.planAdvanceBlocker() == "" {
				// Ready: advance here, before the items below spend what the
				// requirements count. The rest waits for the next tick,
				// where what belonged to the old age drops out.
				advanceAt, advanceIdx = i, len(out)
				out = append(out, ge.plan[i+1:]...)
				break
			}
			out = append(out, it)
			continue
		}
		if it.Kind == PlanDeal {
			if gone := ge.planDealGone(it); gone != "" {
				ge.addLog("warning", fmt.Sprintf("Plan: dropped %s (%s).", ge.planItemLabel(it), gone))
				continue
			}
			if ge.runPlanDeal(it, reserved) {
				started = true
				ge.note(config.BadgeEvPlanStarted, PlanDeal)
				starts.deals = append(starts.deals, "took a "+ge.planItemLabel(it))
				continue
			}
			out = append(out, it)
			continue
		}
		if it.Kind == PlanTrade {
			if reason := ge.planTradeInvalid(it); reason != "" {
				ge.addLog("warning", fmt.Sprintf("Plan: dropped %s (%s).", ge.planItemLabel(it), reason))
				continue
			}
			if sold, got := ge.runPlanTrade(&it, reserved); got > 0 {
				started = true
				ge.note(config.BadgeEvPlanStarted, PlanTrade)
				starts.addTrade(it.Key, it.To, sold, got)
			}
			if it.Count > 0 {
				out = append(out, it)
			}
			continue
		}
		var reason, missing string
		ahead := false
		if it.Kind == PlanBuild {
			// Copies the surviving items above will take first (their
			// counts already net of what they started this walk).
			reason = ge.planBuildInvalid(it.Key, plannedCopies(out, it.Key))
		} else if reason = ge.planResearchInvalid(it.Key); reason == "" {
			// Prerequisites planned before this item count only while they
			// are still in the plan: out holds the survivors so far.
			saved := ge.plan
			ge.plan = out
			missing, _ = ge.planResearchMissing(it.Key, len(out))
			ge.plan = saved
			ahead = ge.planTechAhead(it.Key)
		}
		if (missing != "" || ahead) && it.Count > 0 {
			// It waits for something that is not on its way, or for its
			// age. It keeps its place, but not the research slot's turn or
			// its knowledge: a prerequisite planned below it, or a tech of
			// this age, can then go first.
			out = append(out, it)
			continue
		}
		if reason != "" || it.Count <= 0 {
			if reason == "" {
				reason = "nothing left to start"
			}
			label := ge.planItemLabel(it)
			ge.addLog("warning", fmt.Sprintf("Plan: dropped %s (%s).", label, reason))
			back := map[string]float64{}
			if ge.returnPlanBank(&it, back) {
				ge.logBankReturn("the bank of "+label, "", back)
			}
			continue
		}
		first := it.Kind == PlanResearch && !researchSeen
		if it.Kind == PlanResearch {
			researchSeen = true
		}
		for it.Count > 0 {
			chk := ge.checkPlanItem(it, first)
			if chk.blocked == "" && ge.planCovers(chk.cost, reserved) {
				var err error
				if it.Kind == PlanBuild && len(chk.cost) > 0 && ge.Buildings.defs[it.Key].Category == "wonder" {
					// A wonder is paid through its bank.
					err = ge.payWonderBank(it.Key)
				}
				if err != nil {
					ge.addLog("debug", fmt.Sprintf("Plan: %s refused: %v", ge.planItemLabel(it), err))
					break
				}
				if it.Kind == PlanBuild {
					err = ge.startPlanBuild(&it, chk.banked)
				} else {
					err = ge.startResearchLocked(it.Key, true)
				}
				if err != nil {
					// The walk's checks mirror the command's; if the command
					// still refuses, keep the item and try again next tick.
					ge.addLog("debug", fmt.Sprintf("Plan: %s refused: %v", ge.planItemLabel(it), err))
					break
				}
				started = true
				it.Count--
				it.Started++
				ge.note(config.BadgeEvPlanStarted, it.Kind)
				if it.Kind == PlanBuild {
					starts.addBuild(it.Key)
					// MaxCount may be reached mid-item.
					if it.Count > 0 {
						if why := ge.planBuildInvalid(it.Key, plannedCopies(out, it.Key)); why != "" {
							ge.addLog("warning", fmt.Sprintf("Plan: dropped the last %s (%s).", ge.planItemLabel(it), why))
							it.Count = 0
						}
					}
				} else {
					starts.techs = append(starts.techs, it.Key)
				}
				continue
			}
			if chk.reserve {
				for res, c := range chk.cost {
					reserved[res] += c
				}
			}
			break
		}
		if it.Count > 0 {
			out = append(out, it)
		} else {
			// Done: what is left of its bank (a price that fell) goes back.
			ge.returnPlanBank(&it, nil)
		}
	}
	if advanceAt >= 0 {
		next := ge.progress.GetNextAge(ge.age)
		// A fated doom cannot be outrun by the plan either: it strikes, or
		// its harbinger comes, first. If the advance must wait, the item goes
		// back where it was and tries again next tick.
		if err := ge.fateBeforeAdvance(next); err != nil {
			ge.addLog("debug", fmt.Sprintf("Plan: advance waits: %v", err))
			out = append(out[:advanceIdx], append([]PlanItem{ge.plan[advanceAt]}, out[advanceIdx:]...)...)
			advanceAt = -1
		}
	}
	if len(out) == 0 {
		out = nil
	}
	ge.plan = out
	if advanceAt >= 0 {
		next := ge.progress.GetNextAge(ge.age)
		ge.addLog("info", fmt.Sprintf("Plan: advancing to the %s.", ge.progress.GetAgeName(next)))
		// Reported before the advance, so a badge judged on reaching the age
		// sees that the plan did it.
		ge.note(config.BadgeEvPlanStarted, PlanAdvance)
		ge.advanceAge(next)
		starts.advanced = ge.progress.GetAgeName(ge.age)
		started = true
	}
	return started
}

// startPlanBuild starts the next copy of build item it: its overflow bank
// pays used (the part of the price checkPlanItem found the bank covers; nil
// for none) and the stores the rest. Then used comes out of the bank.
func (ge *GameEngine) startPlanBuild(it *PlanItem, used map[string]float64) error {
	if err := ge.startBuildPaid(it.Key, true, used); err != nil {
		return err
	}
	for res, u := range used {
		if left := it.Banked[res] - u; left > planBankEpsilon {
			it.Banked[res] = left
		} else {
			delete(it.Banked, res)
		}
	}
	if len(it.Banked) == 0 {
		it.Banked = nil
	}
	return nil
}

// staffPlanCopy fills key's empty worker slots. The plan calls it when a copy
// it started completes: a player who plans a producer while away wants it
// staffed, and otherwise it would run at the unstaffed 20% until they came
// back. Idle workers go first: into the copy, or, with worker shares set,
// wherever the shares say (shares.go; the copy's domain may already have its
// share). Then come workers in buildings an advance superseded (legacy: a
// higher tier of their lineage is open), the same lineage's first: after a
// `plan advance` the old age's copies have usually taken every idle hand,
// and the new age's producers would otherwise wait for the next visit. With
// shares set, only the copy's own domain gives workers, so the split holds.
// It never takes food workers (the move must not starve anyone) or this
// age's (they are where the player put them). Recruiting is the shares
// routine's job: it follows within a few ticks, as housing and food allow.
// Rates are recalculated by the caller.
func (ge *GameEngine) staffPlanCopy(key string) {
	def := ge.Buildings.defs[key]
	if def.WorkerCapacity <= 0 {
		return
	}
	shares := len(ge.workerShares) > 0
	if shares {
		ge.placeIdle(ge.newStaffView())
	} else if n := min(ge.slotsFree(key), ge.Workers.IdleCount("worker")); n > 0 {
		ge.Workers.Assign("worker", key, n)
	}
	free := ge.slotsFree(key)
	if p, set := ge.workerShares[def.WorkerDomain]; free <= 0 || ge.Buildings.IsLegacy(key) || set && p <= 0 {
		return
	}
	var same, other []string
	for _, src := range sortedKeys(ge.Buildings.legacyBuildings) {
		sdef := ge.Buildings.defs[src]
		if !ge.Buildings.legacyBuildings[src] || !planStaffSource(sdef) || shares && sdef.WorkerDomain != def.WorkerDomain {
			continue
		}
		if sdef.LineageKey == def.LineageKey {
			same = append(same, src)
		} else {
			other = append(other, src)
		}
	}
	for _, src := range append(same, other...) {
		if free <= 0 {
			break
		}
		if n := min(free, ge.Workers.GetAssignedCount("worker", src)); n > 0 {
			ge.Workers.Unassign("worker", src, n)
			ge.Workers.Assign("worker", key, n)
			free -= n
		}
	}
}

// planStaffSource reports whether staffPlanCopy may take workers from a
// legacy building of def: a producer (production or research) that makes no
// food. Soldiers, priests and the like stay where they are.
func planStaffSource(def config.BuildingDef) bool {
	if def.Category != "production" && def.Category != "research" {
		return false
	}
	for _, eff := range def.Effects {
		if eff.Type == "production" && eff.Target == "food" {
			return false
		}
	}
	return true
}

// runPlanTick runs the plan once during live play and logs what started.
func (ge *GameEngine) runPlanTick() {
	var s planStarts
	if ge.runPlan(&s) {
		ge.addLog(LogRoutine, "Plan: "+s.describe(ge.rules)+".")
	}
}

// planViews evaluates the plan without starting anything, for the UI: the
// same walk as runPlan, reservations included, so Status says what the next
// tick would do. Must be called with at least the read lock held.
func (ge *GameEngine) planViews() []PlanItemView {
	if len(ge.plan) == 0 {
		return nil
	}
	reserved := map[string]float64{}
	researchSeen := false
	out := make([]PlanItemView, 0, len(ge.plan))
	for i, it := range ge.plan {
		if it.Kind == PlanTrade {
			out = append(out, ge.planTradeView(it, reserved))
			continue
		}
		if it.Kind == PlanAdvance {
			out = append(out, ge.planAdvanceView(it))
			continue
		}
		if it.Kind == PlanDeal {
			out = append(out, ge.planDealView(it, reserved))
			continue
		}
		v := PlanItemView{Kind: it.Kind, Key: it.Key, Count: it.Count, Started: it.Started}
		if it.Kind == PlanResearch {
			v.Name = techDefName(ge.rules, it.Key)
		} else {
			v.Name = ge.Buildings.defs[it.Key].Name
		}
		// A tech waiting for something that is not on its way, or for its
		// age, holds neither the research slot's turn nor its knowledge, as
		// in the walk.
		missing := ""
		waits := false
		if it.Kind == PlanResearch {
			missing, _ = ge.planResearchMissing(it.Key, i)
			waits = missing != "" || ge.planTechAhead(it.Key)
		}
		first := it.Kind == PlanResearch && !researchSeen && !waits
		if it.Kind == PlanResearch && !waits {
			researchSeen = true
		}
		chk := ge.checkPlanItem(it, first)
		if missing != "" {
			chk.blocked, chk.reserve = "needs "+missing+" first", false
		}
		// A banked build shows its whole price, what its bank holds of it,
		// and progress counting the bank.
		price := chk.cost
		if chk.price != nil {
			price = chk.price
			v.Banked = cloneBank(chk.banked)
		}
		if len(price) > 0 {
			v.Cost = make(map[string]float64, len(price))
			for k, c := range price {
				v.Cost[k] = c
			}
		}
		// Progress: the least-covered resource after the reservations above.
		v.Progress = 1
		for _, res := range sortedKeys(price) {
			c := price[res]
			if c <= 0 {
				continue
			}
			have := chk.banked[res]
			if _, due := chk.cost[res]; due {
				have += math.Max(0, ge.planFree(res, reserved))
			} else {
				have = c // the bank covers it
			}
			f := math.Max(0, math.Min(1, have/c))
			if f < v.Progress {
				v.Progress, v.Short = f, res
			}
		}
		switch {
		case chk.blocked != "":
			v.Status, v.Note = PlanStatusBlocked, chk.blocked
		case v.Progress >= 1:
			v.Status, v.Short = PlanStatusReady, ""
		default:
			v.Status = PlanStatusWaiting
		}
		if v.Status == PlanStatusReady {
			// It would start and pay; later items see what's left.
			for res, c := range chk.cost {
				reserved[res] += c
			}
		} else if chk.reserve {
			for res, c := range chk.cost {
				reserved[res] += c
			}
		}
		out = append(out, v)
	}
	return out
}
