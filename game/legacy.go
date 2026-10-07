package game

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

// The legacy kit (Pacing v2, PR 6): three prestige shop items that carry a
// run's automation across the reset (site/docs/prestige.md, "The legacy
// kit").
//
//   - Plan Template (config.LegacyPlan). The engine records the plan as it is
//     written (ge.planLog): every build, research, trade and advance item the
//     player adds, tagged with the age it was added in. A build counts the
//     copies added; removing an item takes back the copies it never started,
//     so a build's count ends as what was started plus what is still
//     waiting. Deals are left out (their offers die with the run). At each
//     prestige and Succumb the log becomes the template (age by age: an age
//     the run entered takes this run's slice, an age it never reached keeps
//     the old one, so a short run never forgets a deeper one). With the
//     template bought, the start of a run and every advance add that age's
//     slice to the plan, the advance item included, so a plan written once
//     chains ages while the player is away. The slice is logged again as
//     written, so the template carries forward. The techs in it are the
//     ones the player planned (`plan research`): the kit never picks a
//     research path by itself.
//   - Worker Shares (config.LegacyWorkers). The shares the player set carry
//     into the next run instead of going back to auto (startRunShares).
//   - Old Friends (config.LegacyFactions). Every civilization met is
//     remembered; with it bought, each is met again as soon as the age
//     reaches its own, with no expedition and no fallback wait, at neutral
//     opinion (meetOldFriendsLocked).
//
// The kit remembers whether or not it is bought, so an item bought right
// after a prestige works on the run that just began.
//
// A fourth item, Research Memory, which replayed the last run's research
// order whenever the research slot was idle, was cut before release (the
// owner, 2026-10-03): research is getting its own redesign, a full tech
// tree, and nothing here should choose a research path for the player.

// PlanTemplateItem is one item of the plan template, as saved: the age it
// was written in and the item (see PlanItem; Count is a build's copies).
type PlanTemplateItem struct {
	Age    string  `json:"age"`
	Kind   string  `json:"kind"`
	Key    string  `json:"key,omitempty"`
	Count  int     `json:"count,omitempty"`
	To     string  `json:"to,omitempty"`
	Amount float64 `json:"amount,omitempty"`
}

// clonePlanTemplate copies a template (nil for an empty one).
func clonePlanTemplate(t []PlanTemplateItem) []PlanTemplateItem {
	if len(t) == 0 {
		return nil
	}
	return append([]PlanTemplateItem(nil), t...)
}

// templateAgeCount is how many items t holds for age.
func templateAgeCount(t []PlanTemplateItem, age string) int {
	n := 0
	for _, it := range t {
		if it.Age == age {
			n++
		}
	}
	return n
}

// loadPlanTemplate is a saved template as the engine accepts it: ages set
// knows and known kinds (deals never), counts in range, at most MaxPlanItems
// per age.
func loadPlanTemplate(set *rules.Set, saved []PlanTemplateItem) []PlanTemplateItem {
	var out []PlanTemplateItem
	perAge := map[string]int{}
	for _, it := range saved {
		if _, ok := set.Index(it.Age); !ok || perAge[it.Age] >= MaxPlanItems {
			continue
		}
		switch it.Kind {
		case PlanBuild:
			if it.Key == "" || it.Count <= 0 {
				continue
			}
			it.Count = min(it.Count, maxPlanCount)
		case PlanResearch:
			if it.Key == "" {
				continue
			}
			it.Count = 0
		case PlanTrade:
			if it.Key == "" || it.To == "" || !(it.Amount >= 0) || it.Amount > 1e18 {
				continue
			}
			it.Count = 0
		case PlanAdvance:
			it.Key, it.Count = "", 0
		default:
			continue
		}
		perAge[it.Age]++
		out = append(out, it)
	}
	return out
}

// mergePlanTemplate is the template after a run: for every age in touched,
// the run's slice (empty if it wrote nothing there); for every other age,
// the old template's slice. Ages come out in set's age order.
func mergePlanTemplate(set *rules.Set, old, run []PlanTemplateItem, touched map[string]bool) []PlanTemplateItem {
	var out []PlanTemplateItem
	for _, a := range set.AgeKeys() {
		src := old
		if touched[a] {
			src = run
		}
		for _, it := range src {
			if it.Age == a {
				out = append(out, it)
			}
		}
	}
	return out
}

// uniqueKeys is keys with every repeat after the first dropped.
func uniqueKeys(keys []string) []string {
	seen := make(map[string]bool, len(keys))
	var out []string
	for _, k := range keys {
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}

// ===== Recording the plan as written =====

// logPlanAddLocked records that the player added it to the plan (count
// copies of a build) in the current age. A build merges into the age's entry
// for the same building, a research, trade or advance item already recorded
// for the age is not recorded twice, and an age holds at most MaxPlanItems
// entries. Caller holds the write lock.
func (ge *GameEngine) logPlanAddLocked(it PlanItem, count int) {
	if it.Kind == PlanDeal || (it.Kind == PlanBuild && count <= 0) {
		return
	}
	for i := range ge.planLog {
		e := &ge.planLog[i]
		if e.Age != ge.age || e.Kind != it.Kind {
			continue
		}
		switch it.Kind {
		case PlanBuild:
			if e.Key == it.Key {
				e.Count = min(e.Count+count, maxPlanCount)
				return
			}
		case PlanResearch:
			if e.Key == it.Key {
				return
			}
		case PlanTrade:
			if e.Key == it.Key && e.To == it.To {
				e.Amount = it.Amount
				return
			}
		case PlanAdvance:
			return
		}
	}
	if templateAgeCount(ge.planLog, ge.age) >= MaxPlanItems {
		return
	}
	t := PlanTemplateItem{Age: ge.age, Kind: it.Kind, Key: it.Key}
	switch it.Kind {
	case PlanBuild:
		t.Count = count
	case PlanTrade:
		t.To, t.Amount = it.To, it.Amount
	case PlanAdvance:
		t.Key = ""
	}
	ge.planLog = append(ge.planLog, t)
}

// unlogPlanItemLocked takes back what the player removed from the plan: a
// build loses the copies it never started (it.Count), a trade that bought
// something keeps what it bought (it.Got) as its amount, and any other item
// loses its entry. The entry looked for is the current age's, then the
// latest earlier one. Caller holds the write lock.
func (ge *GameEngine) unlogPlanItemLocked(it PlanItem) {
	if it.Kind == PlanDeal {
		return
	}
	match := func(e PlanTemplateItem) bool {
		if e.Kind != it.Kind {
			return false
		}
		switch it.Kind {
		case PlanTrade:
			return e.Key == it.Key && e.To == it.To
		case PlanAdvance:
			return true
		}
		return e.Key == it.Key
	}
	idx := -1
	order := ge.rules.Indexes()
	cur := order[ge.age]
	for i := len(ge.planLog) - 1; i >= 0; i-- {
		e := ge.planLog[i]
		if !match(e) || order[e.Age] > cur {
			continue
		}
		if e.Age == ge.age {
			idx = i
			break
		}
		if idx < 0 {
			idx = i
		}
	}
	if idx < 0 {
		return
	}
	switch {
	case it.Kind == PlanBuild:
		ge.planLog[idx].Count -= it.Count
		if ge.planLog[idx].Count > 0 {
			return
		}
	case it.Kind == PlanTrade && it.Got > 0:
		ge.planLog[idx].Amount = it.Got
		return
	}
	ge.planLog = slices.Delete(ge.planLog, idx, idx+1)
	if len(ge.planLog) == 0 {
		ge.planLog = nil
	}
}

// ===== Applying the template =====

// addTemplateItemLocked adds one template item to the plan, through the
// same checks the plan commands make. Reports whether it went in and
// whether the plan was full. Caller holds the write lock.
func (ge *GameEngine) addTemplateItemLocked(t PlanTemplateItem) (added, full bool) {
	n := len(ge.plan)
	switch t.Kind {
	case PlanBuild:
		def, ok := ge.Buildings.defs[t.Key]
		if !ok {
			return false, false
		}
		planned := plannedCopies(ge.plan, t.Key)
		if ge.planBuildInvalid(t.Key, planned) != "" {
			return false, false
		}
		count := t.Count
		if def.MaxCount > 0 {
			room := def.MaxCount - ge.Buildings.GetCount(t.Key) - ge.Buildings.GetQueueCount(t.Key, ge.buildQueue) - planned
			count = min(count, room)
		}
		if count <= 0 {
			return false, false
		}
		if n > 0 && ge.plan[n-1].Kind == PlanBuild && ge.plan[n-1].Key == t.Key {
			ge.plan[n-1].Count = min(ge.plan[n-1].Count+count, maxPlanCount)
			return true, false
		}
		if n >= MaxPlanItems {
			return false, true
		}
		ge.plan = append(ge.plan, PlanItem{Kind: PlanBuild, Key: t.Key, Count: count})
	case PlanResearch:
		if _, ok := ge.rules.Tech(t.Key); !ok {
			return false, false
		}
		for _, it := range ge.plan {
			if it.Kind == PlanResearch && it.Key == t.Key {
				return false, false
			}
		}
		if ge.planResearchRefused(t.Key, n) != "" {
			return false, false
		}
		if n >= MaxPlanItems {
			return false, true
		}
		ge.plan = append(ge.plan, PlanItem{Kind: PlanResearch, Key: t.Key, Count: 1})
	case PlanTrade:
		it := PlanItem{Kind: PlanTrade, Key: t.Key, To: t.To, Count: 1, Amount: t.Amount}
		if ge.planTradeInvalid(it) != "" {
			return false, false
		}
		for _, p := range ge.plan {
			if p.Kind == PlanTrade && p.Key == t.Key && p.To == t.To {
				return false, false
			}
		}
		if n >= MaxPlanItems {
			return false, true
		}
		ge.plan = append(ge.plan, it)
	case PlanAdvance:
		for _, it := range ge.plan {
			if it.Kind == PlanAdvance {
				return false, false
			}
		}
		if ge.progress.GetNextAge(ge.age) == "" {
			return false, false
		}
		if n >= MaxPlanItems {
			return false, true
		}
		ge.plan = append(ge.plan, PlanItem{Kind: PlanAdvance, Count: 1})
	default:
		return false, false
	}
	return true, false
}

// applyPlanTemplateLocked adds the template's slice for the current age to
// the plan, when Plan Template is bought, and logs it as written so the
// template carries forward. One log line says what went in. Caller holds
// the write lock.
func (ge *GameEngine) applyPlanTemplateLocked() {
	pm := ge.Prestige
	if !pm.Owns(config.LegacyPlan) {
		return
	}
	added, skipped, full := 0, 0, 0
	for _, t := range pm.legacyPlan {
		if t.Age != ge.age {
			continue
		}
		ok, isFull := ge.addTemplateItemLocked(t)
		switch {
		case ok:
			added++
		case isFull:
			full++
		default:
			skipped++
		}
		// Logged as written whatever happened: a full plan or an item that
		// can't start now doesn't make the template forget it.
		item := PlanItem{Kind: t.Kind, Key: t.Key, To: t.To, Amount: t.Amount}
		ge.logPlanAddLocked(item, t.Count)
	}
	if added+skipped+full == 0 {
		return
	}
	pm.templateApplied = ge.age
	line := fmt.Sprintf("Plan Template: added %s for the %s.", textfmt.Count(added, "item", "items"), ge.rules.Name(rules.KindAge, ge.age))
	if full > 0 {
		line += fmt.Sprintf(" The plan is full (%d items), so %s waited out.", MaxPlanItems, textfmt.Count(full, "item", "items"))
	}
	if skipped > 0 {
		line += fmt.Sprintf(" %s can't be planned yet.", textfmt.Count(skipped, "item", "items"))
	}
	ge.addLog(LogRoutine, line)
}

// ===== Old Friends =====

// meetOldFriendsLocked meets again every remembered civilization whose age
// the player has reached, when Old Friends is bought: at neutral opinion, as
// first contact would, with one log line each. Caller holds the write lock.
func (ge *GameEngine) meetOldFriendsLocked() {
	if len(ge.Prestige.legacyFactions) == 0 || !ge.Prestige.Owns(config.LegacyFactions) {
		return
	}
	order := ge.rules.Indexes()
	cur := order[ge.age]
	for _, key := range ge.Prestige.legacyFactions {
		def, ok := ge.Diplomacy.factionDefs[key]
		if !ok || ge.Diplomacy.IsDiscovered(key) || order[def.MinAge] > cur {
			continue
		}
		if _, ok := ge.Diplomacy.DiscoverFaction(key); ok {
			ge.addLog("event", fmt.Sprintf("Old friends: the %s remember your people and make contact again.", def.Name))
		}
	}
}

// ===== Worker Shares =====

// applyLegacySharesLocked puts the remembered worker shares back when Worker
// Shares is bought and no shares are set. Reports whether it set any.
// Caller holds the write lock.
func (ge *GameEngine) applyLegacySharesLocked() bool {
	if !ge.Prestige.Owns(config.LegacyWorkers) || len(ge.workerShares) > 0 || len(ge.Prestige.legacyShares) == 0 {
		return false
	}
	ge.workerShares = cloneShares(ge.Prestige.legacyShares)
	return true
}

// ===== Run boundaries =====

// captureLegacyLocked is what the kit remembers when a run ends (prestige)
// or falls (Succumb): the plan as written becomes the template age by age,
// the civilizations met are added to those remembered, and the worker
// shares are kept (an unset split keeps the old one unless Worker Shares is
// bought, when clearing them is the player's choice). The run's plan log starts over. Caller holds the write lock,
// before the managers are reset.
func (ge *GameEngine) captureLegacyLocked() {
	pm := ge.Prestige
	// The ages whose slice this run rewrites: those it wrote something in,
	// and, with the template bought, every age it entered (the template was
	// added there and logged, so an emptied slice was the player's doing).
	touched := map[string]bool{}
	for _, it := range ge.planLog {
		touched[it.Age] = true
	}
	if pm.Owns(config.LegacyPlan) {
		for a := range ge.enteredAgesLocked() {
			touched[a] = true
		}
	}
	pm.legacyPlan = mergePlanTemplate(ge.rules, pm.legacyPlan, ge.planLog, touched)
	if len(pm.legacyPlan) == 0 {
		pm.legacyPlan = nil
	}
	ge.planLog = nil
	met := map[string]bool{}
	for _, k := range pm.legacyFactions {
		met[k] = true
	}
	for _, def := range ge.Diplomacy.factionList {
		if ge.Diplomacy.IsDiscovered(def.Key) {
			met[def.Key] = true
		}
	}
	pm.legacyFactions = nil
	for _, def := range ge.Diplomacy.factionList { // roster order, so the log reads the same every time
		if met[def.Key] {
			pm.legacyFactions = append(pm.legacyFactions, def.Key)
		}
	}
	if len(ge.workerShares) > 0 || pm.Owns(config.LegacyWorkers) {
		pm.legacyShares = cloneShares(ge.workerShares)
	}
}

// startRunLegacyLocked applies the kit at the start of a run (after a
// prestige or a Succumb), once the managers are fresh: the shares, the
// first age's template slice, and old friends of the first age. Caller
// holds the write lock.
func (ge *GameEngine) startRunLegacyLocked() {
	if ge.applyLegacySharesLocked() {
		ge.addLog("info", "Worker Shares: your shares carry over. "+sharesLine(ge.rules, ge.workerShares)+".")
	}
	ge.applyPlanTemplateLocked()
	ge.meetOldFriendsLocked()
}

// legacyOnAgeEnteredLocked applies the kit on entering a new age: its
// template slice and its old friends. Caller holds the write lock.
func (ge *GameEngine) legacyOnAgeEnteredLocked() {
	ge.applyPlanTemplateLocked()
	ge.meetOldFriendsLocked()
}

// legacyOnPurchaseLocked puts a kit item to work the moment it is bought:
// the template's slice for this age (unless the plan log already holds
// something written in this age), the shares (unless some are set), and old
// friends already within reach. Caller holds the write lock.
func (ge *GameEngine) legacyOnPurchaseLocked(key string) {
	switch key {
	case config.LegacyPlan:
		// The ages this run already passed keep their slices as written
		// (the run's log takes them over), and this age's slice goes into
		// the plan unless something was written here already.
		for _, t := range ge.Prestige.legacyPlan {
			if a := t.Age; a != ge.age && ge.enteredAgesLocked()[a] && templateAgeCount(ge.planLog, a) == 0 {
				ge.copyTemplateSliceLocked(a)
			}
		}
		if templateAgeCount(ge.planLog, ge.age) == 0 {
			ge.applyPlanTemplateLocked()
		}
	case config.LegacyWorkers:
		if ge.applyLegacySharesLocked() {
			ge.addLog("info", "Your remembered worker shares are set again. "+ge.applySharesLocked("").Line)
		}
	case config.LegacyFactions:
		ge.meetOldFriendsLocked()
	}
}

// enteredAgesLocked is every age this run has entered: the first age, the
// ages it reached and the current one. Caller holds the lock.
func (ge *GameEngine) enteredAgesLocked() map[string]bool {
	out := map[string]bool{ge.rules.AgeKeys()[0]: true, ge.age: true}
	for _, a := range ge.Stats.AgesReached {
		out[a] = true
	}
	return out
}

// copyTemplateSliceLocked writes the template's slice for age into the
// run's plan log without adding it to the plan. Caller holds the write lock.
func (ge *GameEngine) copyTemplateSliceLocked(age string) {
	for _, t := range ge.Prestige.legacyPlan {
		if t.Age == age {
			ge.planLog = append(ge.planLog, t)
		}
	}
}

// ===== Shop refund =====

// refundShopLocked is the one-time step that moves a save from the first
// prestige shop to the legacy kit: the retired perks' tiers go to 0 (their
// keys stay), and the points they cost plus the points left are refunded at
// the new rate (config.ShopRefund). Available and the lifetime total become
// the refund, in depth points from here on. Runs once, after the signature
// check, and marks the save (PrestigeSave.ShopVersion). A save with nothing
// to refund is marked silently. Caller holds the write lock.
func (ge *GameEngine) refundShopLocked() {
	pm := ge.Prestige
	if pm.shopVersion >= config.PrestigeShopVersion {
		return
	}
	pm.shopVersion = config.PrestigeShopVersion
	old, refund := config.ShopRefund(pm.level, pm.available, pm.upgrades)
	for _, def := range pm.upgradeList {
		if _, held := pm.upgrades[def.Key]; held && def.Retired {
			pm.upgrades[def.Key] = 0
		}
	}
	if pm.level <= 0 && old <= 0 {
		return
	}
	pm.available, pm.totalEarned = refund, refund
	how := fmt.Sprintf("%s at %d each", textfmt.Count(pm.level, "prestige", "prestiges"), config.RefundPerPrestige)
	if refund != pm.level*config.RefundPerPrestige {
		how = fmt.Sprintf("your %s old points at the new rate", textfmt.Int(old))
	}
	ge.addLog("info", fmt.Sprintf("The prestige shop changed. Your old perks were refunded as %s (%s).",
		textfmt.Count(refund, "point", "points"), how))
}

// ===== Snapshot =====

// LegacyKitState is what the legacy kit remembers, for the prestige shop.
type LegacyKitState struct {
	// PlanItems is the template's size and PlanAges how many ages it spans;
	// PlanByAge is its items per age. PlanAppliedAge is the age whose slice
	// Plan Template last added to the plan ("" before any this run).
	PlanItems      int
	PlanAges       int
	PlanByAge      map[string]int
	PlanAppliedAge string
	// Factions is how many civilizations are remembered (FactionKeys, in
	// roster order); Shares how many worker domains have a remembered share.
	Factions    int
	FactionKeys []string
	Shares      int
}

// kitState summarizes the kit's memory for the snapshot.
func (pm *PrestigeManager) kitState() LegacyKitState {
	byAge := map[string]int{}
	for _, it := range pm.legacyPlan {
		byAge[it.Age]++
	}
	return LegacyKitState{
		PlanItems:      len(pm.legacyPlan),
		PlanAges:       len(byAge),
		PlanByAge:      byAge,
		PlanAppliedAge: pm.templateApplied,
		Factions:       len(pm.legacyFactions),
		FactionKeys:    slices.Clone(pm.legacyFactions),
		Shares:         len(pm.legacyShares),
	}
}

// LoadLegacy restores the shop version and the kit's memory from a save,
// cleaning what a hand edit could have broken.
func (pm *PrestigeManager) LoadLegacy(shopVersion int, plan []PlanTemplateItem, factions []string, shares map[string]float64) {
	pm.shopVersion = shopVersion
	pm.legacyPlan = loadPlanTemplate(pm.rules, plan)
	known := map[string]bool{}
	for _, f := range pm.rules.Factions() {
		known[f.Key] = true
	}
	pm.legacyFactions = nil
	for _, k := range uniqueKeys(factions) {
		if known[k] {
			pm.legacyFactions = append(pm.legacyFactions, k)
		}
	}
	pm.legacyShares = cleanSharesIn(pm.rules, shares)
}

// ===== Test hooks =====

// LegacyKit is the legacy kit's memory as a save carries it, for tests and
// the smoke suite's canned veteran kit.
type LegacyKit struct {
	Plan     []PlanTemplateItem `json:"plan,omitempty"`
	Factions []string           `json:"factions,omitempty"`
	Shares   map[string]float64 `json:"shares,omitempty"`
}

// LegacyForTest returns what the kit remembers now. A test hook for other
// packages (the smoke suite dumps it to make the canned veteran kit); not
// reachable from play.
func (ge *GameEngine) LegacyForTest() LegacyKit {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	pm := ge.Prestige
	return LegacyKit{
		Plan:     clonePlanTemplate(pm.legacyPlan),
		Factions: slices.Clone(pm.legacyFactions),
		Shares:   cloneShares(pm.legacyShares),
	}
}

// SetLegacyForTest gives the engine kit as its memory and, with own, buys
// every kit item for free, then puts them to work as a purchase would (the
// current age's template slice, the shares, old friends). A test hook for
// other packages (the smoke suite's veteran with the kit); not reachable
// from play.
func (ge *GameEngine) SetLegacyForTest(kit LegacyKit, own bool) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.Prestige.LoadLegacy(ge.Prestige.shopVersion, kit.Plan, kit.Factions, kit.Shares)
	if !own {
		return
	}
	for _, key := range ge.rules.LegacyKit() {
		ge.Prestige.upgrades[key] = 1
		ge.legacyOnPurchaseLocked(key)
	}
}

// NotePlanForTest records kind and key (count copies of a build) in the
// run's plan log as if the player had planned them in the current age. A
// test hook for other packages (the smoke suite records a bot's builds to
// make the canned veteran template); not reachable from play.
func (ge *GameEngine) NotePlanForTest(kind, key string, count int) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.logPlanAddLocked(PlanItem{Kind: kind, Key: key, Count: count}, count)
}

// WriteShopV1SaveForTest writes the running game, with the given prestige
// state, as a signed save from the first prestige shop: no shop version and
// no kit memory, as a build from before the legacy kit wrote it. A test hook
// for other packages (the smoke suite's refund check loads it back); not
// reachable from play.
func (ge *GameEngine) WriteShopV1SaveForTest(name string, level, totalEarned, available int, upgrades map[string]int) error {
	ge.mu.Lock()
	ge.Prestige.LoadState(level, totalEarned, available, maps.Clone(upgrades))
	gs := ge.buildSaveSnapshot()
	ge.mu.Unlock()
	stripShopV2(&gs)
	gs.Signature = signSave(gs, saveHMACKey)
	data, err := json.MarshalIndent(gs, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(savePath(name), data, 0o644)
}

// stripShopV2 removes what the legacy kit added to a save, leaving the
// bytes a build from before it would have written.
func stripShopV2(gs *GameSave) {
	gs.Prestige.ShopVersion = 0
	gs.Prestige.LegacyPlan, gs.Prestige.LegacyFactions, gs.Prestige.LegacyShares = nil, nil, nil
	gs.PlanLog = nil
}

// EnterAgeForTest moves the game into age through the real advance (the
// carryover, the unlocks, the kit's template slice and old friends), with no
// requirements and no fate checked. A test hook for other packages (the
// smoke suite's kit checks); not reachable from play.
func (ge *GameEngine) EnterAgeForTest(age string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if _, ok := ge.rules.Index(age); !ok {
		return fmt.Errorf("Unknown age '%s'.", age)
	}
	ge.advanceAge(age)
	return nil
}

// NoteAdvanceForTest records in the run's plan log an advance item for age
// (the age just left), as if the player had planned the advance they made.
// A test hook for other packages (the smoke suite records a bot's advances
// to make the canned veteran template); not reachable from play.
func (ge *GameEngine) NoteAdvanceForTest(age string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if _, ok := ge.rules.Index(age); !ok || templateAgeCount(ge.planLog, age) >= MaxPlanItems {
		return
	}
	for _, e := range ge.planLog {
		if e.Age == age && e.Kind == PlanAdvance {
			return
		}
	}
	ge.planLog = append(ge.planLog, PlanTemplateItem{Age: age, Kind: PlanAdvance})
}

// NoteTradeForTest records in the run's plan log a trade item selling give
// for get, adding got to what the current age's item buys, as if the player
// had planned the trades they made. A test hook for other packages (the
// smoke suite records a bot's trades to make the canned veteran template);
// not reachable from play.
func (ge *GameEngine) NoteTradeForTest(give, get string, got float64) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	for i := range ge.planLog {
		e := &ge.planLog[i]
		if e.Age == ge.age && e.Kind == PlanTrade && e.Key == give && e.To == get {
			e.Amount = float64(e.Amount + got)
			return
		}
	}
	ge.logPlanAddLocked(PlanItem{Kind: PlanTrade, Key: give, To: get, Count: 1, Amount: got}, 1)
}
