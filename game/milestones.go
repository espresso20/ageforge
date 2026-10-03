package game

import (
	"slices"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// MilestoneManager tracks milestone completion, chain progress, and the
// civilisation title awarded to the player.
//
// Milestones are checked each tick via CheckMilestones; once met they are
// permanently flagged in the completed map and never re-checked. Chains are
// unlocked when all their constituent milestones are completed and grant a
// title and a temporary tick-speed boost via InjectEvent.
//
// Title priority: chain titles override the count-based fallback title.
// The most recently completed chain's title wins (last writer in the loop).
type MilestoneManager struct {
	defs            []config.MilestoneDef
	completed       map[string]bool
	chains          []config.MilestoneChainDef
	chainsCompleted map[string]bool
	currentTitle    string
	// milestoneToChain is a reverse-index built at construction time for O(1)
	// chain lookup when rendering the milestone list in the UI.
	milestoneToChain map[string]string // milestone key -> chain key
	// titles is the count-based fallback title ladder, held so
	// recalculateTitle (run every tick from checkMilestones) doesn't rebuild it.
	titles []config.TitleDef
}

// NewMilestoneManager creates a new milestone manager
func NewMilestoneManager() *MilestoneManager {
	chains := config.MilestoneChains()
	m2c := make(map[string]string)
	for _, c := range chains {
		for _, mk := range c.MilestoneKeys {
			m2c[mk] = c.Key
		}
	}
	return &MilestoneManager{
		defs:             config.Milestones(),
		completed:        make(map[string]bool),
		chains:           chains,
		chainsCompleted:  make(map[string]bool),
		milestoneToChain: m2c,
		titles:           config.MilestoneTitles(),
	}
}

// CheckMilestones checks all milestones against current state.
// Returns list of newly completed milestones.
func (mm *MilestoneManager) CheckMilestones(
	tick int,
	age string,
	ageOrder map[string]int,
	resources *ResourceManager,
	buildings *BuildingManager,
	population int,
	techCount int,
	totalBuilt int,
	researchedTechs map[string]bool,
	soldiersTrained int,
	wonderCount int,
	knowledgeCount int,
) []config.MilestoneDef {
	var completed []config.MilestoneDef

	for _, def := range mm.defs {
		if mm.completed[def.Key] {
			continue
		}

		if mm.checkMilestone(def, tick, age, ageOrder, resources, buildings, population, techCount, totalBuilt, researchedTechs, soldiersTrained, wonderCount, knowledgeCount) {
			mm.completed[def.Key] = true
			completed = append(completed, def)
		}
	}

	return completed
}

func (mm *MilestoneManager) checkMilestone(
	def config.MilestoneDef,
	tick int,
	age string,
	ageOrder map[string]int,
	resources *ResourceManager,
	buildings *BuildingManager,
	population int,
	techCount int,
	totalBuilt int,
	researchedTechs map[string]bool,
	soldiersTrained int,
	wonderCount int,
	knowledgeCount int,
) bool {
	// Check min tick
	if def.MinTick > 0 && tick < def.MinTick {
		return false
	}

	// Check age
	if def.MinAge != "" {
		if ageOrder[age] < ageOrder[def.MinAge] {
			return false
		}
	}

	// Check resources
	for res, required := range def.MinResources {
		if resources.Get(res) < required {
			return false
		}
	}

	// Check buildings
	for bld, required := range def.MinBuildings {
		if buildings.GetCount(bld) < required {
			return false
		}
	}
	if sum := def.MinBuildingSum; sum.Count > 0 {
		have := 0
		for _, bld := range sum.Keys {
			have += buildings.GetCount(bld)
		}
		if have < sum.Count {
			return false
		}
	}

	// Check population
	if def.MinPopulation > 0 && population < def.MinPopulation {
		return false
	}

	// Check tech count
	if def.MinTechCount > 0 && techCount < def.MinTechCount {
		return false
	}

	// Check specific techs
	for _, tech := range def.RequiredTechs {
		if !researchedTechs[tech] {
			return false
		}
	}

	// Run counters (zero asks for nothing)
	if totalBuilt < def.MinTotalBuilt || soldiersTrained < def.MinSoldiersTrained ||
		wonderCount < def.MinWonders || knowledgeCount < def.MinKnowledgeWorkers {
		return false
	}

	return true
}

// CheckChains checks for newly completed chains. Returns newly completed chain defs.
func (mm *MilestoneManager) CheckChains() []config.MilestoneChainDef {
	var newlyCompleted []config.MilestoneChainDef
	for _, chain := range mm.chains {
		if mm.chainsCompleted[chain.Key] {
			continue
		}
		allDone := true
		for _, mk := range chain.MilestoneKeys {
			if !mm.completed[mk] {
				allDone = false
				break
			}
		}
		if allDone {
			mm.chainsCompleted[chain.Key] = true
			newlyCompleted = append(newlyCompleted, chain)
		}
	}
	return newlyCompleted
}

// recalculateTitle selects the current title. Chain titles take precedence over
// the count-based fallback sequence. Among chain titles the last completed
// chain (in config order) wins, giving later chains a natural prestige upgrade.
func (mm *MilestoneManager) recalculateTitle() {
	// Chain titles take priority (use latest completed chain's title)
	bestChainTitle := ""
	for _, chain := range mm.chains {
		if mm.chainsCompleted[chain.Key] {
			bestChainTitle = chain.Title
		}
	}
	if bestChainTitle != "" {
		mm.currentTitle = bestChainTitle
		return
	}

	// Fallback to count-based titles
	count := len(mm.completed)
	mm.currentTitle = ""
	for _, t := range mm.titles {
		if count >= t.MinMilestones {
			mm.currentTitle = t.Title
		}
	}
}

// IsCompleted checks if a milestone has been achieved
func (mm *MilestoneManager) IsCompleted(key string) bool {
	return mm.completed[key]
}

// CompletedCount returns how many milestones are completed
func (mm *MilestoneManager) CompletedCount() int {
	return len(mm.completed)
}

// GetCompleted returns all completed milestone keys
func (mm *MilestoneManager) GetCompleted() []string {
	return sortedKeys(mm.completed)
}

// GetChainsCompleted returns all completed chain keys
func (mm *MilestoneManager) GetChainsCompleted() []string {
	return sortedKeys(mm.chainsCompleted)
}

// GetCurrentTitle returns the current civilization title
func (mm *MilestoneManager) GetCurrentTitle() string {
	return mm.currentTitle
}

// computeProgress computes progress indicators for a milestone definition
func (mm *MilestoneManager) computeProgress(def config.MilestoneDef, params MilestoneSnapshotParams) []MilestoneProgress {
	var progress []MilestoneProgress

	if def.MinTick > 0 {
		progress = append(progress, MilestoneProgress{
			Label:   "Ticks survived",
			Current: float64(params.Tick),
			Target:  float64(def.MinTick),
			Met:     params.Tick >= def.MinTick,
		})
	}

	if def.MinAge != "" {
		currentOrder := params.AgeOrder[params.Age]
		targetOrder := params.AgeOrder[def.MinAge]
		met := currentOrder >= targetOrder
		progress = append(progress, MilestoneProgress{
			Label:   "Age: " + AgeName(def.MinAge),
			Current: float64(currentOrder),
			Target:  float64(targetOrder),
			Met:     met,
		})
	}

	// Sort resource keys before iterating — map iteration order is random,
	// which would cause progress bar lines to flicker each tick.
	resKeys := make([]string, 0, len(def.MinResources))
	for res := range def.MinResources {
		resKeys = append(resKeys, res)
	}
	sort.Strings(resKeys)
	for _, res := range resKeys {
		required := def.MinResources[res]
		current := params.Resources[res]
		progress = append(progress, MilestoneProgress{
			Label:   textfmt.Capitalize(ResourceName(res)),
			Current: current,
			Target:  required,
			Met:     current >= required,
		})
	}

	// Sort building keys before iterating — same reason as resources.
	bldKeys := make([]string, 0, len(def.MinBuildings))
	for bld := range def.MinBuildings {
		bldKeys = append(bldKeys, bld)
	}
	sort.Strings(bldKeys)
	for _, bld := range bldKeys {
		required := def.MinBuildings[bld]
		current := float64(params.Buildings[bld])
		progress = append(progress, MilestoneProgress{
			Label:   BuildingName(bld),
			Current: current,
			Target:  float64(required),
			Met:     int(current) >= required,
		})
	}
	// A building sum is one row: "Trading Posts and Merchant Quarters".
	if sum := def.MinBuildingSum; sum.Count > 0 {
		have := 0
		names := make([]string, 0, len(sum.Keys))
		for _, bld := range sum.Keys {
			have += params.Buildings[bld]
			names = append(names, pluralName(2, BuildingName(bld)))
		}
		progress = append(progress, MilestoneProgress{
			Label:   textfmt.List(names),
			Current: float64(have),
			Target:  float64(sum.Count),
			Met:     have >= sum.Count,
		})
	}

	if def.MinPopulation > 0 {
		progress = append(progress, MilestoneProgress{
			Label:   "Population",
			Current: float64(params.Population),
			Target:  float64(def.MinPopulation),
			Met:     params.Population >= def.MinPopulation,
		})
	}

	if def.MinTechCount > 0 {
		progress = append(progress, MilestoneProgress{
			Label:   "Technologies",
			Current: float64(params.TechCount),
			Target:  float64(def.MinTechCount),
			Met:     params.TechCount >= def.MinTechCount,
		})
	}

	// Run counters
	for _, c := range []struct {
		label         string
		current, need int
	}{
		{"Buildings built", params.TotalBuilt, def.MinTotalBuilt},
		{"Soldiers trained", params.SoldiersTrained, def.MinSoldiersTrained},
		{"Wonders", params.WonderCount, def.MinWonders},
		{"Knowledge workers", params.KnowledgeCount, def.MinKnowledgeWorkers},
	} {
		if c.need > 0 {
			progress = append(progress, MilestoneProgress{
				Label:   c.label,
				Current: float64(c.current),
				Target:  float64(c.need),
				Met:     c.current >= c.need,
			})
		}
	}

	return progress
}

// overallProgress returns 0.0-1.0 progress ratio for a milestone
func overallProgress(progress []MilestoneProgress) float64 {
	if len(progress) == 0 {
		return 0
	}
	total := 0.0
	for _, p := range progress {
		if p.Target <= 0 {
			if p.Met {
				total += 1.0
			}
			continue
		}
		ratio := p.Current / p.Target
		if ratio > 1.0 {
			ratio = 1.0
		}
		total += ratio
	}
	return total / float64(len(progress))
}

// formatRewards formats effects into a human-readable reward string
func formatRewards(effects []config.Effect) string {
	var parts []string
	for _, e := range effects {
		switch e.Type {
		case "instant_resource":
			parts = append(parts, "+"+Amount(e.Value, e.Target))
		case "permanent_bonus":
			parts = append(parts, textfmt.SignedPercent(e.Value)+" "+EffectTargetName(e.Target))
		}
	}
	return strings.Join(parts, ", ")
}

// Snapshot returns milestone state for UI with progress, chains, and titles
func (mm *MilestoneManager) Snapshot(params MilestoneSnapshotParams) MilestoneState {
	milestones := make(map[string]MilestoneInfo)
	visibleCount := 0

	for _, def := range mm.defs {
		completed := mm.completed[def.Key]
		progress := mm.computeProgress(def, params)
		ratio := overallProgress(progress)

		// Visibility: completed || !hidden || progress > 0.5
		// Age milestones: visible when player is in preceding age or later
		visible := completed || !def.Hidden || ratio > 0.5
		if def.Hidden && def.MinAge != "" && !completed {
			// For hidden age milestones, show when in preceding age
			targetOrder := params.AgeOrder[def.MinAge]
			currentOrder := params.AgeOrder[params.Age]
			if currentOrder >= targetOrder-1 {
				visible = true
			}
		}
		// No spoilers: an unfinished milestone that needs an age the player
		// cannot see named yet stays hidden, whatever its progress.
		if !completed && def.MinAge != "" && params.Sight != nil && !params.Sight.Age(def.MinAge) {
			visible = false
		}

		if visible {
			visibleCount++
		}

		chainKey := mm.milestoneToChain[def.Key]

		milestones[def.Key] = MilestoneInfo{
			Name:        def.Name,
			Description: def.Description,
			Category:    def.Category,
			Hidden:      def.Hidden,
			Visible:     visible,
			Completed:   completed,
			RewardText:  formatRewards(def.Rewards),
			Rewards:     slices.Clone(def.Rewards), // def is the manager's table
			Progress:    progress,
			ChainKey:    chainKey,
		}
	}

	// Build chain info
	var chains []ChainInfo
	// Check which chains have active boosts
	activeBoosts := make(map[string]bool)
	for _, ae := range params.activeEvents {
		if strings.HasSuffix(ae.Key, "_chain_boost") {
			activeBoosts[strings.TrimSuffix(ae.Key, "_boost")] = true
		}
	}

	for _, chain := range mm.chains {
		completedCount := 0
		for _, mk := range chain.MilestoneKeys {
			if mm.completed[mk] {
				completedCount++
			}
		}
		chains = append(chains, ChainInfo{
			Name:           chain.Name,
			Key:            chain.Key,
			Category:       chain.Category,
			CompletedCount: completedCount,
			TotalCount:     len(chain.MilestoneKeys),
			Complete:       mm.chainsCompleted[chain.Key],
			Title:          chain.Title,
			BoostActive:    activeBoosts[chain.Key],
		})
	}

	return MilestoneState{
		Milestones:     milestones,
		CompletedCount: len(mm.completed),
		TotalCount:     len(mm.defs),
		VisibleCount:   visibleCount,
		Chains:         chains,
		CurrentTitle:   mm.currentTitle,
	}
}

// LoadState restores milestone state from save
func (mm *MilestoneManager) LoadState(completed []string, chainsCompleted []string, title string) {
	mm.completed = make(map[string]bool)
	for _, key := range completed {
		mm.completed[key] = true
	}
	mm.chainsCompleted = make(map[string]bool)
	for _, key := range chainsCompleted {
		mm.chainsCompleted[key] = true
	}
	mm.currentTitle = title
}
