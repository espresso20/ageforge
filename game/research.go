package game

import (
	"fmt"
	"maps"
	"slices"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
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
	// Currently in-progress tech key, or "" if idle.
	currentTech string
	ticksLeft   int
	totalTicks  int
	// bonuses is what the researched techs add together, by what each
	// effect changes: its kind and its target (config.TechEffectKey). Output
	// of gold, storage of gold and gold a tick are three keys, so none can
	// leak into another. Fractions and flat amounts both live here; the
	// kind says which a sum is.
	bonuses map[config.TechEffectKey]float64
	// pools is the fractions in bonuses by the pool each adds to
	// ("production_all", "<res>_rate", "gather_rate", "tick_speed",
	// "military_power", ...): the names the resolver and the panels share
	// with every other source of bonuses. Rebuilt with bonuses (indexPools).
	pools map[string]float64
	// order is every tech key, sorted, fixed at construction. Per-tick walks
	// over researched techs use it so summed effects never follow map order.
	order []string
	// timeK is the Era Mastery speed the next research starts at: its time
	// is divided by it after the research-speed step (MasteryTicks). The
	// engine sets it before each start; 0 or 1 leaves times alone.
	timeK float64
	// timeMult is Ancient Knowledge's factor for the next research
	// (GameEngine.succumbResearchFactor): its time is multiplied by it after
	// the research-speed step and before Era Mastery. The engine sets it
	// before each start; 0 or 1 leaves times alone.
	timeMult float64
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

// StartResearch begins researching a technology using only tech-derived bonuses.
// Prefer StartResearchWithSpeed to include permanent and prestige bonuses.
func (rm *ResearchManager) StartResearch(key string, currentAge string, ageOrder map[string]int, knowledge float64) error {
	return rm.StartResearchWithSpeed(key, currentAge, ageOrder, knowledge, rm.Bonus(config.EffectResearchTime, ""))
}

// StartResearchWithSpeed begins researching a technology, applying the given combined
// research speed bonus (from techs + permanent bonuses + prestige) to reduce tick count.
func (rm *ResearchManager) StartResearchWithSpeed(key string, currentAge string, ageOrder map[string]int, knowledge float64, speedBonus float64) error {
	def, ok := rm.defs[key]
	if !ok {
		return unknownKeyError("tech", key, rm.defs, "Type research list to see what you can research.")
	}
	if rm.researched[key] {
		return fmt.Errorf("%s is already researched.", def.Name)
	}
	if rm.currentTech != "" {
		currentDef := rm.defs[rm.currentTech]
		return fmt.Errorf("Already researching %s (%s left). Type research cancel to stop it.", currentDef.Name, DurationText(rm.ticksLeft, BaseTickInterval))
	}
	// Check age requirement
	if ageOrder[def.Age] > ageOrder[currentAge] {
		return fmt.Errorf("%s needs %s.", def.Name, laterAgeRef(rm.rules, currentAge, def.Age))
	}
	// Check prerequisites
	for _, prereq := range def.Prerequisites {
		if !rm.researched[prereq] {
			prereqDef := rm.defs[prereq]
			return fmt.Errorf("%s needs %s researched first.", def.Name, prereqDef.Name)
		}
	}
	// Check cost
	if knowledge < def.Cost {
		return fmt.Errorf("Not enough knowledge for %s: need %s, have %s.", def.Name, textfmt.Number(def.Cost), textfmt.Number(knowledge))
	}

	rm.currentTech = key
	ticks := ResearchTicks(def.ResearchTicks, speedBonus, rm.timeMult, rm.timeK)
	rm.ticksLeft = ticks
	rm.totalTicks = ticks
	return nil
}

// ResearchTicks is how long a tech listed at base ticks takes to research
// with a research speed bonus of speed and Ancient Knowledge's factor mult,
// on ground of speed k: research speed takes its share off the listed time
// (+30% leaves 70% of it, rounded down, one tick at least), Ancient
// Knowledge multiplies what is left (x0.8 per epoch succumbed in, rounded
// down, one tick at least), then Era Mastery divides that by k
// (MasteryTicks). The engine starts research with it and the Research panel
// lists times with it, so the time a tech shows is the time it takes.
func ResearchTicks(base int, speed, mult, k float64) int {
	ticks := base
	if speed > 0 {
		ticks = int(float64(ticks) * (1.0 - speed))
		if ticks < 1 {
			ticks = 1
		}
	}
	return MasteryTicks(ancientKnowledgeTicks(ticks, mult), k)
}

// ancientKnowledgeTicks multiplies ticks by Ancient Knowledge's factor mult,
// rounded down, never below one tick. A factor of 1 or more, or one no
// engine has set (0), leaves ticks as they are.
func ancientKnowledgeTicks(ticks int, mult float64) int {
	if mult <= 0 || mult >= 1 || ticks <= 0 {
		return ticks
	}
	return max(1, int(float64(float64(ticks)*mult)))
}

// memoryResearchSlowdown is the tick multiplier applied to an Ancient Memory
// tech: the cache recalls a half-forgotten technology, so it researches at 50%
// speed (double the tick count). See StartMemoryResearch.
const memoryResearchSlowdown = 2.0

// StartMemoryResearch begins researching a technology recovered from an Ancient
// Memory cache (see GameEngine.AcceptAncientMemory). Unlike StartResearchWithSpeed
// it deliberately BYPASSES the prerequisite chain, the age gate, and the knowledge
// cost — the cache hands you the tech "free" of those gates. The tradeoff is speed:
// the tick count is doubled (50% rate). Because the slowdown is baked into ticksLeft
// /totalTicks here, it persists through save/load with no extra research-save field.
//
// It still refuses if the tech is unknown, already researched, or another research
// is in progress (one slot, same as normal research).
func (rm *ResearchManager) StartMemoryResearch(key string, speedBonus float64) error {
	def, ok := rm.defs[key]
	if !ok {
		return unknownKeyError("tech", key, rm.defs, "Type research list to see what you can research.")
	}
	if rm.researched[key] {
		return fmt.Errorf("%s is already researched.", def.Name)
	}
	if rm.currentTech != "" {
		currentDef := rm.defs[rm.currentTech]
		return fmt.Errorf("Already researching %s (%s left). Type research cancel to stop it.", currentDef.Name, DurationText(rm.ticksLeft, BaseTickInterval))
	}
	// NOTE: age gate, prerequisite loop, and knowledge cost check are intentionally
	// omitted — that is the whole point of an Ancient Memory.

	rm.currentTech = key
	ticks := def.ResearchTicks
	// Apply the same combined research speed bonus a normal research gets...
	if speedBonus > 0 {
		ticks = int(float64(ticks) * (1.0 - speedBonus))
	}
	// ...then halve the rate (double the ticks) for the memory penalty.
	ticks = int(float64(ticks) * memoryResearchSlowdown)
	if ticks < 1 {
		ticks = 1
	}
	ticks = MasteryTicks(ancientKnowledgeTicks(ticks, rm.timeMult), rm.timeK)
	rm.ticksLeft = ticks
	rm.totalTicks = ticks
	return nil
}

// rebuildBonuses recomputes bonuses from every researched tech, summed in
// rm.order. Summing in completion order instead made the totals depend on the
// order techs were finished (0.1+0.3+0.4 is not 0.4+0.1+0.3 in floating
// point), so a loaded game, which can only replay them in one fixed order,
// came back with bonuses a few ulps off the live ones.
func (rm *ResearchManager) rebuildBonuses() {
	rm.bonuses = make(map[config.TechEffectKey]float64)
	for _, key := range rm.order {
		if !rm.researched[key] {
			continue
		}
		for _, eff := range rm.defs[key].Effects {
			rm.bonuses[eff.Key()] += eff.Value
		}
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

// Tick processes one tick of research. Returns completed tech key or empty string.
func (rm *ResearchManager) Tick() string {
	if rm.currentTech == "" {
		return ""
	}
	rm.ticksLeft--
	if rm.ticksLeft <= 0 {
		completed := rm.currentTech
		rm.researched[completed] = true

		// Apply effects as permanent bonuses
		rm.rebuildBonuses()

		rm.currentTech = ""
		rm.ticksLeft = 0
		rm.totalTicks = 0
		return completed
	}
	return ""
}

// Advance moves research on by n ticks at once (offline catch-up). Returns
// the completed tech key, or "" if none completed.
func (rm *ResearchManager) Advance(n int) string {
	if rm.currentTech == "" || n <= 0 {
		return ""
	}
	if rm.ticksLeft > n {
		rm.ticksLeft -= n
		return ""
	}
	rm.ticksLeft = 1
	return rm.Tick()
}

// CancelResearch cancels current research
func (rm *ResearchManager) CancelResearch() (string, bool) {
	if rm.currentTech == "" {
		return "", false
	}
	tech := rm.currentTech
	rm.currentTech = ""
	rm.ticksLeft = 0
	rm.totalTicks = 0
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
			rm.ticksLeft = 0
			rm.totalTicks = 0
		}
	}
	return completed
}

// IsResearched returns whether a tech has been completed
func (rm *ResearchManager) IsResearched(key string) bool {
	return rm.researched[key]
}

// Bonus returns what the researched techs add together to one thing: a kind
// and its target ("" for the kinds that take none). Military power is
// Bonus(config.EffectMilitaryPower, ""), gold storage
// Bonus(config.EffectFlatStorage, "gold"), storage for every resource
// Bonus(config.EffectFlatStorage, config.AllResources).
func (rm *ResearchManager) Bonus(kind config.TechEffectKind, target string) float64 {
	return rm.bonuses[config.TechEffectKey{Kind: kind, Target: target}]
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

// byTarget is the sums of one kind, by target: what researched techs add per
// tick to each resource (config.EffectFlatOutput), or to each store
// (config.EffectFlatStorage).
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
// by pool name ("production_all": 0.5 is +50%). The flat amounts (output,
// storage, housing) are not pools and are not in it.
func (rm *ResearchManager) GetBonuses() map[string]float64 {
	return maps.Clone(rm.pools)
}

// Modifiers emits one OpAdd Modifier per pool the researched techs add to,
// attributed to Source "research". The targets are the pool names every
// other source uses ("production_all", "<res>_rate", "gather_rate",
// "tick_speed"). Per-tech attribution is deferred; the summed per-target
// view is golden-equal to the current scattered math.
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
		available := true
		// Check age
		if ageOrder[def.Age] > ageOrder[currentAge] {
			available = false
		}
		// Check prereqs
		prereqsMet := true
		for _, prereq := range def.Prerequisites {
			if !rm.researched[prereq] {
				prereqsMet = false
				available = false
				break
			}
		}

		techs[key] = TechState{
			Name:          def.Name,
			Age:           def.Age,
			Cost:          def.Cost,
			Prerequisites: slices.Clone(def.Prerequisites), // def is the manager's table
			Description:   def.Description,
			Researched:    rm.researched[key],
			Available:     available && !rm.researched[key],
			PrereqsMet:    prereqsMet,
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
		TicksLeft:       rm.ticksLeft,
		TotalTicks:      rm.totalTicks,
		TotalResearched: len(rm.researched),
		Bonuses:         rm.GetBonuses(),
		Flat:            rm.byTarget(config.EffectFlatOutput),
		Storage:         rm.byTarget(config.EffectFlatStorage),
		Housing:         rm.Bonus(config.EffectFlatHousing, ""),
	}
}

// LoadState restores research state from save data.
// Bonuses are always recomputed by replaying the effect list of every
// researched tech — this ensures correct values even if tech definitions
// changed between versions.
func (rm *ResearchManager) LoadState(researched []string, currentTech string, ticksLeft, totalTicks int) {
	rm.researched = make(map[string]bool)
	for _, key := range researched {
		rm.researched[key] = true
	}
	rm.rebuildBonuses()
	rm.currentTech = currentTech
	rm.ticksLeft = ticksLeft
	rm.totalTicks = totalTicks
}
