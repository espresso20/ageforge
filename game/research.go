package game

import (
	"fmt"

	"github.com/espresso20/ageforge/config"
)

// ResearchManager manages the tech tree and research progress.
// Only one technology can be in progress at a time. When research completes,
// its Effect entries are accumulated into bonuses immediately so they are
// applied on the next recalculateRates call.
//
// NOTE: bonuses are rebuilt from scratch during LoadState by replaying all
// researched tech effects — do not persist the bonuses map independently.
type ResearchManager struct {
	defs       map[string]config.TechDef
	researched map[string]bool
	// Currently in-progress tech key, or "" if idle.
	currentTech string
	ticksLeft   int
	totalTicks  int
	// bonuses accumulates effect values keyed by eff.Target (e.g. "tick_speed",
	// "production_all", "research_speed"). These feed into engine.recalculateRates.
	bonuses map[string]float64
}

// NewResearchManager creates a new research manager
func NewResearchManager() *ResearchManager {
	return &ResearchManager{
		defs:       config.TechByKey(),
		researched: make(map[string]bool),
		bonuses:    make(map[string]float64),
	}
}

// StartResearch begins researching a technology using only tech-derived bonuses.
// Prefer StartResearchWithSpeed to include permanent and prestige bonuses.
func (rm *ResearchManager) StartResearch(key string, currentAge string, ageOrder map[string]int, knowledge float64) error {
	return rm.StartResearchWithSpeed(key, currentAge, ageOrder, knowledge, rm.bonuses["research_speed"])
}

// StartResearchWithSpeed begins researching a technology, applying the given combined
// research speed bonus (from techs + permanent bonuses + prestige) to reduce tick count.
func (rm *ResearchManager) StartResearchWithSpeed(key string, currentAge string, ageOrder map[string]int, knowledge float64, speedBonus float64) error {
	def, ok := rm.defs[key]
	if !ok {
		return fmt.Errorf("unknown technology: %s", key)
	}
	if rm.researched[key] {
		return fmt.Errorf("%s is already researched", def.Name)
	}
	if rm.currentTech != "" {
		currentDef := rm.defs[rm.currentTech]
		return fmt.Errorf("already researching %s (%d ticks left)", currentDef.Name, rm.ticksLeft)
	}
	// Check age requirement
	if ageOrder[def.Age] > ageOrder[currentAge] {
		return fmt.Errorf("%s requires %s age", def.Name, def.Age)
	}
	// Check prerequisites
	for _, prereq := range def.Prerequisites {
		if !rm.researched[prereq] {
			prereqDef := rm.defs[prereq]
			return fmt.Errorf("%s requires %s to be researched first", def.Name, prereqDef.Name)
		}
	}
	// Check cost
	if knowledge < def.Cost {
		return fmt.Errorf("not enough knowledge (have: %.0f, need: %.0f)", knowledge, def.Cost)
	}

	rm.currentTech = key
	ticks := def.ResearchTicks
	// Apply combined research speed bonus (tech + permanent + prestige)
	if speedBonus > 0 {
		ticks = int(float64(ticks) * (1.0 - speedBonus))
		if ticks < 1 {
			ticks = 1
		}
	}
	rm.ticksLeft = ticks
	rm.totalTicks = ticks
	return nil
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
		def := rm.defs[completed]
		for _, eff := range def.Effects {
			rm.bonuses[eff.Target] += eff.Value
		}

		rm.currentTech = ""
		rm.ticksLeft = 0
		rm.totalTicks = 0
		return completed
	}
	return ""
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
	for key, def := range rm.defs {
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
		for _, eff := range def.Effects {
			rm.bonuses[eff.Target] += eff.Value
		}
		completed = append(completed, key)
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

// GetBonus returns the accumulated bonus for a target
func (rm *ResearchManager) GetBonus(target string) float64 {
	return rm.bonuses[target]
}

// ResearchedCount returns how many techs have been researched
func (rm *ResearchManager) ResearchedCount() int {
	return len(rm.researched)
}

// GetResearched returns all researched tech keys
func (rm *ResearchManager) GetResearched() []string {
	var keys []string
	for k := range rm.researched {
		keys = append(keys, k)
	}
	return keys
}

// GetBonuses returns a copy of all bonuses
func (rm *ResearchManager) GetBonuses() map[string]float64 {
	out := make(map[string]float64)
	for k, v := range rm.bonuses {
		out[k] = v
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
			Prerequisites: def.Prerequisites,
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
		Techs:            techs,
		CurrentTech:      rm.currentTech,
		CurrentTechName:  currentName,
		TicksLeft:        rm.ticksLeft,
		TotalTicks:       rm.totalTicks,
		TotalResearched:  len(rm.researched),
		Bonuses:          rm.GetBonuses(),
	}
}

// LoadState restores research state from save data.
// Bonuses are always recomputed by replaying the effect list of every
// researched tech — this ensures correct values even if tech definitions
// changed between versions.
func (rm *ResearchManager) LoadState(researched []string, currentTech string, ticksLeft, totalTicks int) {
	rm.researched = make(map[string]bool)
	rm.bonuses = make(map[string]float64)
	for _, key := range researched {
		rm.researched[key] = true
		// Re-apply bonuses
		if def, ok := rm.defs[key]; ok {
			for _, eff := range def.Effects {
				rm.bonuses[eff.Target] += eff.Value
			}
		}
	}
	rm.currentTech = currentTech
	rm.ticksLeft = ticksLeft
	rm.totalTicks = totalTicks
}
