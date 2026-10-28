package game

import (
	"fmt"
	"slices"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// The legacy kit (Pacing v2, PR 6): three prestige shop items that carry a
// run's automation across the reset (site/docs/prestige.md, "The legacy
// kit").
//
//   - Plan Template (config.LegacyPlan). The engine records the plan as it is
//     written (ge.planLog): every build, research and advance item the
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
// was written in and the item (see PlanItem; Count is a build's copies). To
// and Amount were the retired sell orders'; a template read from a save
// skips those rows, and nothing sets the fields now.
type PlanTemplateItem struct {
	Age   string `json:"age"`
	Kind  string `json:"kind"`
	Key   string `json:"key,omitempty"`
	Count int    `json:"count,omitempty"`
}

// templateAgeLoad is how many of the items t holds for age count against
// an age's MaxPlanItems: every one but research, as in the plan itself.
func templateAgeLoad(t []PlanTemplateItem, age string) int {
	n := 0
	for _, it := range t {
		if it.Age == age && it.Kind != PlanResearch {
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
		if _, ok := set.Index(it.Age); !ok || it.Kind != PlanResearch && perAge[it.Age] >= MaxPlanItems {
			continue
		}
		switch it.Kind {
		case PlanBuild, PlanUpgrade:
			if it.Key == "" || it.Count <= 0 {
				continue
			}
			it.Count = min(it.Count, maxPlanCount)
		case PlanResearch:
			if it.Key == "" {
				continue
			}
			it.Count = 0
		case PlanAdvance:
			it.Key, it.Count = "", 0
		default:
			continue
		}
		if it.Kind != PlanResearch {
			perAge[it.Age]++
		}
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
// for the same building, a research or advance item already recorded
// for the age is not recorded twice, and an age holds at most MaxPlanItems
// entries. Caller holds the write lock.
func (ge *GameEngine) logPlanAddLocked(it PlanItem, count int) {
	if it.Kind == PlanDeal || ((it.Kind == PlanBuild || it.Kind == PlanUpgrade) && count <= 0) {
		return
	}
	for i := range ge.planLog {
		e := &ge.planLog[i]
		if e.Age != ge.age || e.Kind != it.Kind {
			continue
		}
		switch it.Kind {
		case PlanBuild, PlanUpgrade:
			if e.Key == it.Key {
				e.Count = min(e.Count+count, maxPlanCount)
				return
			}
		case PlanResearch:
			if e.Key == it.Key {
				return
			}
		case PlanAdvance:
			return
		}
	}
	if it.Kind != PlanResearch && templateAgeLoad(ge.planLog, ge.age) >= MaxPlanItems {
		return
	}
	t := PlanTemplateItem{Age: ge.age, Kind: it.Kind, Key: it.Key}
	switch it.Kind {
	case PlanBuild, PlanUpgrade:
		t.Count = count
	case PlanAdvance:
		t.Key = ""
	}
	ge.planLog = append(ge.planLog, t)
}

// unlogPlanItemLocked takes back what the player removed from the plan: a
// build loses the copies it never started (it.Count), and any other item
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
	case it.Kind == PlanBuild || it.Kind == PlanUpgrade:
		ge.planLog[idx].Count -= it.Count
		if ge.planLog[idx].Count > 0 {
			return
		}
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
	// Research takes no room in the plan: only the other kinds can find it
	// full.
	full = planLoad(ge.plan) >= MaxPlanItems
	switch t.Kind {
	case PlanBuild:
		def, ok := ge.Buildings.defs[t.Key]
		if !ok {
			return false, false
		}
		planned := plannedCount(ge.plan, PlanBuild, t.Key)
		if ge.planBuildInvalid(t.Key, planned) != "" {
			return false, false
		}
		count := min(t.Count, ge.planRoom(def, planned))
		if count <= 0 {
			return false, false
		}
		if n > 0 && ge.plan[n-1].Kind == PlanBuild && ge.plan[n-1].Key == t.Key {
			ge.plan[n-1].Count = min(ge.plan[n-1].Count+count, maxPlanCount)
			return true, false
		}
		if full {
			return false, true
		}
		ge.plan = append(ge.plan, PlanItem{Kind: PlanBuild, Key: t.Key, Count: count})
	case PlanUpgrade:
		newDef, pending := ge.Buildings.UpgradeTarget(t.Key)
		if !pending || ge.upgradeTechErr(newDef) != nil {
			return false, false
		}
		planned := plannedCount(ge.plan, PlanUpgrade, t.Key)
		count := min(t.Count, ge.Buildings.GetCount(t.Key)-planned, ge.planRoom(newDef, planned))
		if count <= 0 {
			return false, false
		}
		if n > 0 && ge.plan[n-1].Kind == PlanUpgrade && ge.plan[n-1].Key == t.Key {
			ge.plan[n-1].Count = min(ge.plan[n-1].Count+count, maxPlanCount)
			return true, false
		}
		if full {
			return false, true
		}
		ge.plan = append(ge.plan, PlanItem{Kind: PlanUpgrade, Key: t.Key, Count: count})
	case PlanResearch:
		if _, ok := ge.rules.Tech(t.Key); !ok {
			return false, false
		}
		if plannedCount(ge.plan, PlanResearch, t.Key) > 0 {
			return false, false
		}
		if ge.planResearchRefused(t.Key, n) != "" {
			return false, false
		}
		ge.plan = append(ge.plan, PlanItem{Kind: PlanResearch, Key: t.Key, Count: 1})
	case PlanAdvance:
		if plannedCount(ge.plan, PlanAdvance, "") > 0 {
			return false, false
		}
		if ge.progress.GetNextAge(ge.age) == "" {
			return false, false
		}
		if full {
			return false, true
		}
		ge.plan = append(ge.plan, PlanItem{Kind: PlanAdvance, Count: 1})
	default:
		return false, false
	}
	return true, false
}

// countPlanItems counts plan items.
func countPlanItems(n int) string {
	return textfmt.Count(n, vibe.Line("game.legacy.applyPlanTemplateLocked.02", "item"), vibe.Line("game.legacy.applyPlanTemplateLocked.03", "items"))
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
		item := PlanItem{Kind: t.Kind, Key: t.Key}
		ge.logPlanAddLocked(item, t.Count)
	}
	if added+skipped+full == 0 {
		return
	}
	line := vibe.Sprintf("game.legacy.applyPlanTemplateLocked.01", "Plan Template: added %s for the %s.", countPlanItems(added), ge.rules.Name(rules.KindAge, ge.age))
	if full > 0 {
		line += vibe.Sprintf("game.legacy.applyPlanTemplateLocked.04", " The plan is full (%d items, not counting research), so %s waited out.", MaxPlanItems, countPlanItems(full))
	}
	if skipped > 0 {
		line += vibe.Sprintf("game.legacy.applyPlanTemplateLocked.07", " %s can't be planned yet.", countPlanItems(skipped))
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
			ge.addLog("event", vibe.Sprintf("game.legacy.meetOldFriendsLocked.01", "Old friends: the %s remember your people and make contact again.", def.Name))
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
	ge.workerShares = cloneMap(ge.Prestige.legacyShares)
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
		pm.legacyShares = cloneMap(ge.workerShares)
	}
}

// startRunLegacyLocked applies the kit at the start of a run (after a
// prestige or a Succumb), once the managers are fresh: the shares, the
// first age's template slice, and old friends of the first age. Caller
// holds the write lock.
func (ge *GameEngine) startRunLegacyLocked() {
	if ge.applyLegacySharesLocked() {
		ge.addLogAbout("info", TagPeople, vibe.Line("game.legacy.startRunLegacyLocked.01", "Roster: your shares carry over. ")+sharesLine(ge.rules, ge.workerShares)+".")
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

// enteredAgesLocked is every age this run has entered: the first age, the
// ages it reached and the current one. Caller holds the lock.
func (ge *GameEngine) enteredAgesLocked() map[string]bool {
	out := map[string]bool{ge.rules.AgeKeys()[0]: true, ge.age: true}
	for _, a := range ge.Stats.AgesReached {
		out[a] = true
	}
	return out
}

// ===== Snapshot =====

// LegacyKitState is what the legacy kit remembers, for the prestige shop.
type LegacyKitState struct {
	// PlanItems is the template's size and PlanAges how many ages it spans.
	PlanItems int
	PlanAges  int
	// Factions is how many civilizations are remembered; Shares how many
	// worker domains have a remembered share.
	Factions int
	Shares   int
}

// kitState summarizes the kit's memory for the snapshot.
func (pm *PrestigeManager) kitState() LegacyKitState {
	byAge := map[string]int{}
	for _, it := range pm.legacyPlan {
		byAge[it.Age]++
	}
	return LegacyKitState{
		PlanItems: len(pm.legacyPlan),
		PlanAges:  len(byAge),
		Factions:  len(pm.legacyFactions),
		Shares:    len(pm.legacyShares),
	}
}

// LoadLegacy restores the kit's memory from a save, cleaning what a hand
// edit could have broken.
func (pm *PrestigeManager) LoadLegacy(plan []PlanTemplateItem, factions []string, shares map[string]float64) {
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

// EnterAgeForTest moves the game into age through the real advance (the
// carryover, the unlocks, the kit's template slice and old friends), with no
// requirements and no fate checked. A test hook for other packages; not
// reachable from play.
func (ge *GameEngine) EnterAgeForTest(age string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if _, ok := ge.rules.Index(age); !ok {
		return fmt.Errorf("Unknown age '%s'.", age)
	}
	ge.advanceAge(age)
	return nil
}
