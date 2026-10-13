package game

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// ResearchManager manages the tech tree and research progress.
// Only one technology can be in progress at a time. When research completes,
// its effects are accumulated into bonuses immediately so they are applied
// on the next recalculateRates call.
//
// NOTE: bonuses are rebuilt from scratch during LoadState by replaying all
// researched tech effects — do not persist the bonuses map independently.
type ResearchManager struct {
	rules      *rules.Set
	defs       map[string]config.TechDef
	researched map[string]bool
	// Currently in-progress tech key, or "" if idle. The tech's price
	// (TechDef.Cost) is the work it needs; progress is the work done on it.
	// The scholars' work each tick is added to it (AddWork) and the tech is
	// done when progress reaches the price. There is no separate wait.
	currentTech string
	progress    float64
	// slowed marks an Ancient Memory research: the cache recalls a
	// half-forgotten technology, so its scholars do half the work a tick.
	slowed bool
	// head is work done ahead of any tech, kept for the next tech started: a
	// gain that arrived with none in progress. A tech starts with it, never
	// with more than its price; the rest is lost.
	head float64
	// bonuses is what the researched techs come to together, by what each
	// effect changes: its kind and its target (config.TechEffectKey). Output
	// of gold and gold a tick are two keys, so neither can leak into the
	// other. A key that adds holds a sum, one that multiplies a product
	// (TechEffectKey.Fold), each held inside the key's limits. A key no
	// researched tech touches is absent: Term reads it as the key's start.
	bonuses map[config.TechEffectKey]float64
	// raw is bonuses before the limits, for the headroom test: a term that
	// differs from its raw value has run into a floor or a cap.
	raw map[config.TechEffectKey]float64
	// pools is the fractions in bonuses that join a pool the rest of the
	// game's bonuses share ("tick_speed", "military_power",
	// "expedition_reward"), by pool name. Rebuilt with bonuses (indexPools).
	// The layer's kinds join no pool.
	pools map[string]float64
	// order is every tech key, sorted, fixed at construction. Per-tick walks
	// over researched techs use it so summed effects never follow map order.
	order []string
}

// NewResearchManager creates a new research manager on the core ruleset.
func NewResearchManager() *ResearchManager { return NewResearchManagerWith(rules.Core()) }

// NewResearchManagerWith creates a new research manager with set's techs.
func NewResearchManagerWith(set *rules.Set) *ResearchManager {
	defs := set.TechMap()
	return &ResearchManager{
		rules:      set,
		defs:       defs,
		order:      sortedKeys(defs),
		researched: make(map[string]bool),
		bonuses:    make(map[config.TechEffectKey]float64),
		pools:      make(map[string]float64),
	}
}

// Rebind moves the manager onto set: it takes set's tech definitions and
// works the bonuses of the researched techs out again from them. What is
// researched and what is in progress stay.
func (rm *ResearchManager) Rebind(set *rules.Set) {
	rm.rules = set
	rm.defs = set.TechMap()
	rm.order = sortedKeys(rm.defs)
	rm.rebuildBonuses()
}

// StartResearch begins researching a technology. A tech can be started
// whenever its age and prerequisites allow: it has no price to afford. It
// starts with the head start the town holds, never more than its price.
func (rm *ResearchManager) StartResearch(key string, currentAge string, ageOrder map[string]int) error {
	def, ok := rm.defs[key]
	if !ok {
		return unknownKeyError("tech", key, rm.defs, "Type research list to see what you can research.")
	}
	if rm.researched[key] {
		return fmt.Errorf("%s is already researched.", def.Name)
	}
	if rm.currentTech != "" {
		currentDef := rm.defs[rm.currentTech]
		return fmt.Errorf("Already researching %s. Type research cancel to stop it.", currentDef.Name)
	}
	// Check age requirement
	if ageOrder[def.Age] > ageOrder[currentAge] {
		return fmt.Errorf("%s needs %s.", def.Name, laterAgeRef(rm.rules, currentAge, def.Age))
	}
	// Check prerequisites: every one of them, and one of the either-or
	// group when the tech has one.
	if prereq := def.MissingPrereq(rm.IsResearched); prereq != "" {
		return fmt.Errorf("%s needs %s researched first.", def.Name, rm.defs[prereq].Name)
	}
	if !def.AnyOfMet(rm.IsResearched) {
		names := make([]string, len(def.AnyOf))
		for i, k := range def.AnyOf {
			names[i] = rm.defs[k].Name
		}
		return fmt.Errorf("%s needs %s researched first.", def.Name, orList(names))
	}
	rm.begin(key, false)
	return nil
}

// begin puts tech key in progress with the head start the town holds.
func (rm *ResearchManager) begin(key string, slowed bool) {
	rm.currentTech = key
	rm.slowed = slowed
	rm.progress = min(rm.head, rm.defs[key].Cost)
	rm.head = 0
}

// orList joins names as a choice: "Map Making or Boatbuilding", "A, B or C".
func orList(names []string) string {
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1]
}

// techTimeTicks multiplies ticks by the techs' factor on a time, rounded
// down, never below one tick. A factor of 1 or more, or one nothing has set
// (0), leaves ticks as they are.
func techTimeTicks(ticks int, factor float64) int {
	if factor <= 0 || factor >= 1 || ticks <= 0 {
		return ticks
	}
	return max(1, int(float64(float64(ticks)*factor)))
}

// techGrowTicks is ticks × factor for a term that lengthens a time (a
// festival that lasts 25% longer), rounded down. A factor of 1 or less, or
// one nothing has set (0), leaves ticks as they are.
func techGrowTicks(ticks int, factor float64) int {
	if factor <= 1 || ticks <= 0 {
		return ticks
	}
	return int(float64(float64(ticks) * factor))
}

// memoryResearchSlowdown is the share of the work a tick that an Ancient
// Memory tech gets: the cache recalls a half-forgotten technology, so it
// researches at 50% speed. See StartMemoryResearch.
const memoryResearchSlowdown = 0.5

// StartMemoryResearch begins researching a technology recovered from an Ancient
// Memory cache (see GameEngine.AcceptAncientMemory). Unlike StartResearch it
// deliberately BYPASSES the prerequisite chain and the age gate: the cache
// hands you the tech "free" of those gates. The tradeoff is speed: its
// scholars do half the work a tick (slowed). It still refuses if the tech is
// unknown, already researched, or another research is in progress (one slot,
// same as normal research).
func (rm *ResearchManager) StartMemoryResearch(key string) error {
	def, ok := rm.defs[key]
	if !ok {
		return unknownKeyError("tech", key, rm.defs, "Type research list to see what you can research.")
	}
	if rm.researched[key] {
		return fmt.Errorf("%s is already researched.", def.Name)
	}
	if rm.currentTech != "" {
		currentDef := rm.defs[rm.currentTech]
		return fmt.Errorf("Already researching %s. Type research cancel to stop it.", currentDef.Name)
	}
	// NOTE: age gate and prerequisite loop are intentionally omitted — that
	// is the whole point of an Ancient Memory.
	rm.begin(key, true)
	return nil
}

// rebuildBonuses recomputes bonuses from every researched tech, folded in
// rm.order. Folding in completion order instead made the totals depend on
// the order techs were finished (0.1+0.3+0.4 is not 0.4+0.1+0.3 in floating
// point), so a loaded game, which can only replay them in one fixed order,
// came back with bonuses a few ulps off the live ones.
func (rm *ResearchManager) rebuildBonuses() {
	var held []config.TechDef
	for _, key := range rm.order {
		if rm.researched[key] {
			held = append(held, rm.defs[key])
		}
	}
	rm.raw = config.TechTerms(held)
	rm.bonuses = make(map[config.TechEffectKey]float64, len(rm.raw))
	for k, v := range rm.raw {
		rm.bonuses[k] = k.Clamp(v)
	}
	rm.indexPools()
}

// indexPools rebuilds pools from bonuses. Each pool has one key (a kind and
// target name one pool), so a pool's value is its key's sum, copied, never
// added up again.
func (rm *ResearchManager) indexPools() {
	rm.pools = make(map[string]float64)
	for k, v := range rm.bonuses {
		if pool, ok := k.Pool(); ok {
			rm.pools[pool] = v
		}
	}
}

// AddWork adds work the scholars did to the tech in progress and completes
// it when its price is reached. Returns the completed tech key, or "" if
// none completed. Work past the price is kept as a head start for the next
// tech.
func (rm *ResearchManager) AddWork(w float64) string {
	if rm.currentTech == "" || w <= 0 {
		return ""
	}
	rm.progress += w
	price := rm.defs[rm.currentTech].Cost
	if rm.progress < price {
		return ""
	}
	return rm.complete(rm.progress - price)
}

// complete finishes the tech in progress; over is the work past its price.
func (rm *ResearchManager) complete(over float64) string {
	done := rm.currentTech
	rm.researched[done] = true
	// Apply effects as permanent bonuses
	rm.rebuildBonuses()
	rm.currentTech = ""
	rm.progress = 0
	rm.slowed = false
	rm.head += max(over, 0)
	return done
}

// Grant adds work that did not come from the scholars' tick (an event, loot,
// a milestone). It goes to the tech in progress; with none in progress it is
// kept as a head start for the next tech started (never more than that
// tech's price). It returns the work that landed on the tech in progress
// (0 when it was kept), and the tech completed by it, if any.
func (rm *ResearchManager) Grant(w float64) (landed float64, completed string) {
	if w <= 0 {
		return 0, ""
	}
	if rm.currentTech == "" {
		rm.head += w
		return 0, ""
	}
	return w, rm.AddWork(w)
}

// CancelResearch cancels current research; the work done on it is lost.
func (rm *ResearchManager) CancelResearch() (string, bool) {
	if rm.currentTech == "" {
		return "", false
	}
	tech := rm.currentTech
	rm.currentTech = ""
	rm.progress = 0
	rm.slowed = false
	return tech, true
}

// ForceCompleteN instantly completes up to n unresearched techs available in
// the current age. Used by the "Grand Discovery" good epoch event to give the
// player a few free techs without queuing them. Any in-progress research that
// gets completed by this call is also cleared to avoid a stale state where
// currentTech is already in the researched map.
// Returns the keys of techs that were completed.
func (rm *ResearchManager) ForceCompleteN(n int, currentAge string, ageOrder map[string]int) []string {
	var completed []string
	for _, key := range sortedKeys(rm.defs) {
		def := rm.defs[key]
		if len(completed) >= n {
			break
		}
		if rm.researched[key] {
			continue
		}
		if ageOrder[def.Age] > ageOrder[currentAge] {
			continue
		}
		rm.researched[key] = true
		completed = append(completed, key)
	}
	if len(completed) > 0 {
		rm.rebuildBonuses()
	}
	// Also cancel any in-progress research to avoid state inconsistency
	if len(completed) > 0 && rm.currentTech != "" {
		if rm.researched[rm.currentTech] {
			rm.currentTech = ""
			rm.progress = 0
			rm.slowed = false
		}
	}
	return completed
}

// IsResearched returns whether a tech has been completed
func (rm *ResearchManager) IsResearched(key string) bool {
	return rm.researched[key]
}

// Bonus returns what the researched techs come to together on one thing: a
// kind and its target ("" for the kinds that take none). For a kind that
// adds it is the sum, 0 with no tech: military power is
// Bonus(config.EffectMilitaryPower, ""), the layer's share of gold output
// Bonus(config.EffectOutput, "gold"). For a kind that multiplies it is the
// factor, 1 with no tech: Bonus(config.EffectBuildTime, "") is 0.92 after
// one 8% cut. Either way it is inside the kind's limits.
func (rm *ResearchManager) Bonus(kind config.TechEffectKind, target string) float64 {
	k := config.TechEffectKey{Kind: kind, Target: target}
	if v, ok := rm.bonuses[k]; ok {
		return v
	}
	return k.Start()
}

// Mechanic is the tech term of the mechanic number key (config.Mechanics):
// the factor of a number that multiplies, the sum of one that adds.
func (rm *ResearchManager) Mechanic(key string) float64 {
	return rm.Bonus(config.EffectMechanic, key)
}

// MechanicValue is the mechanic number key with the tech term applied to
// base, its constant: base x the factor, or base + the sum. An unknown key
// leaves base alone.
func (rm *ResearchManager) MechanicValue(key string, base float64) float64 {
	def, ok := rm.rules.Mechanic(key)
	if !ok {
		return base
	}
	return def.Apply(base, rm.Mechanic(key))
}

// OutputFactor is the tech layer's multiplier on what resource res makes:
// 1 + the techs' bonus on res + their bonus on all production. Bonuses on
// the same thing add up, and nothing caps the sum.
func (rm *ResearchManager) OutputFactor(res string) float64 {
	return 1 + rm.Bonus(config.EffectOutput, res) + rm.Bonus(config.EffectAllOutput, "")
}

// TimeFactor is the techs' factor on research time (1 with none): the time
// a tech takes is multiplied by it, after the research speed pool.
func (rm *ResearchManager) TimeFactor() float64 {
	return rm.Bonus(config.EffectResearchTime, "")
}

// ResearchedCount returns how many techs have been researched
func (rm *ResearchManager) ResearchedCount() int {
	return len(rm.researched)
}

// GetResearched returns all researched tech keys, sorted. Callers sum tech
// effects in this order, so it must not follow map order.
func (rm *ResearchManager) GetResearched() []string {
	return sortedKeys(rm.researched)
}

// byTarget is the terms of one kind, by target: what researched techs add
// per tick to each resource (config.EffectFlatOutput), the layer's share of
// each resource's output (config.EffectOutput), each mechanic number's term
// (config.EffectMechanic).
func (rm *ResearchManager) byTarget(kind config.TechEffectKind) map[string]float64 {
	out := make(map[string]float64)
	for k, v := range rm.bonuses {
		if k.Kind == kind {
			out[k.Target] = v
		}
	}
	return out
}

// flatEffects is the flat output effects of the researched techs, one entry
// per effect, techs in rm.order. The engine adds each to its resource's
// rate in this order, so the sum never follows map order.
func (rm *ResearchManager) flatEffects() []config.TechEffect {
	var out []config.TechEffect
	for _, key := range rm.order {
		if !rm.researched[key] {
			continue
		}
		for _, eff := range rm.defs[key].Effects {
			if eff.Kind == config.EffectFlatOutput {
				out = append(out, eff)
			}
		}
	}
	return out
}

// GetBonuses returns a copy of the bonus pools the researched techs add to,
// by pool name ("military_power": 0.5 is +50%). Only game speed, military
// power and expedition rewards are pools; everything else a tech gives is
// in the layer (Bonus).
func (rm *ResearchManager) GetBonuses() map[string]float64 {
	return maps.Clone(rm.pools)
}

// Modifiers emits one OpAdd Modifier per pool the researched techs add to,
// attributed to Source "research". The targets are the pool names every
// other source uses ("tick_speed", "military_power", "expedition_reward").
// The layer's bonuses are no part of any pool, so none is emitted for them.
func (rm *ResearchManager) Modifiers() []Modifier {
	out := make([]Modifier, 0, len(rm.pools))
	for t, v := range rm.pools {
		out = append(out, Modifier{Source: "research", Target: t, Op: OpAdd, Value: v})
	}
	return out
}

// Snapshot returns research state for UI
func (rm *ResearchManager) Snapshot(currentAge string, ageOrder map[string]int) ResearchState {
	techs := make(map[string]TechState)

	for key, def := range rm.defs {
		// Available: the age is reached and the prerequisites are met (every
		// one of them, and one of the either-or group).
		prereqsMet := def.PrereqsMet(rm.IsResearched)
		available := prereqsMet && ageOrder[def.Age] <= ageOrder[currentAge]

		techs[key] = TechState{
			Name:          def.Name,
			Age:           def.Age,
			Cost:          def.Cost,
			Prerequisites: slices.Clone(def.Prerequisites), // def is the manager's table
			AnyOf:         slices.Clone(def.AnyOf),
			Description:   def.Description,
			Researched:    rm.researched[key],
			Available:     available && !rm.researched[key],
			PrereqsMet:    prereqsMet,
			Kind:          rm.rules.TechKind(key),
			KeystoneOf:    rm.rules.KeystoneOf(key),
		}
	}

	var currentName string
	if rm.currentTech != "" {
		currentName = rm.defs[rm.currentTech].Name
	}

	return ResearchState{
		Techs:           techs,
		CurrentTech:     rm.currentTech,
		CurrentTechName: currentName,
		Progress:        rm.progress,
		Price:           rm.currentPrice(),
		HeadStart:       rm.head,
		TotalResearched: len(rm.researched),
		Bonuses:         rm.GetBonuses(),
		Flat:            rm.byTarget(config.EffectFlatOutput),
		Output:          rm.byTarget(config.EffectOutput),
		AllOutput:       rm.Bonus(config.EffectAllOutput, ""),
		Storage:         rm.Bonus(config.EffectStorage, ""),
		Housing:         rm.Bonus(config.EffectHousing, ""),
		BuildCost:       rm.Bonus(config.EffectBuildCost, ""),
		BuildTime:       rm.Bonus(config.EffectBuildTime, ""),
		ResearchTime:    rm.TimeFactor(),
		Mechanics:       rm.byTarget(config.EffectMechanic),
	}
}

// currentPrice is the price of the tech in progress (0 with none).
func (rm *ResearchManager) currentPrice() float64 {
	if rm.currentTech == "" {
		return 0
	}
	return rm.defs[rm.currentTech].Cost
}

// LoadState restores research state from save data.
// Bonuses are always recomputed by replaying the effect list of every
// researched tech — this ensures correct values even if tech definitions
// changed between versions.
func (rm *ResearchManager) LoadState(researched []string, currentTech string, progress float64, slowed bool, head float64) {
	rm.researched = make(map[string]bool)
	for _, key := range researched {
		rm.researched[key] = true
	}
	rm.rebuildBonuses()
	rm.currentTech = currentTech
	rm.progress = max(progress, 0)
	rm.slowed = slowed && currentTech != ""
	rm.head = max(head, 0)
}
