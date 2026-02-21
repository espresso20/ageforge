package game

import (
	"fmt"
	"strings"

	"github.com/user/ageforge/config"
)

// MilestoneManager tracks and checks milestones, chains, and titles
type MilestoneManager struct {
	defs             []config.MilestoneDef
	completed        map[string]bool
	chains           []config.MilestoneChainDef
	chainsCompleted  map[string]bool
	currentTitle     string
	milestoneToChain map[string]string // milestone key -> chain key
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
	soldierCount int,
	wonderCount int,
) []config.MilestoneDef {
	var completed []config.MilestoneDef

	for _, def := range mm.defs {
		if mm.completed[def.Key] {
			continue
		}

		if mm.checkMilestone(def, tick, age, ageOrder, resources, buildings, population, techCount, totalBuilt, researchedTechs, soldierCount, wonderCount) {
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
	soldierCount int,
	wonderCount int,
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

	// Special checks based on milestone key
	switch def.Key {
	case "master_builder":
		if totalBuilt < 20 {
			return false
		}
	case "war_machine":
		if soldierCount < 10 {
			return false
		}
	case "wonder_builder":
		if wonderCount < 1 {
			return false
		}
	case "scholars_haven":
		// This checks for 5+ scholars — we need villager data
		// For simplicity, use population >= 10 as proxy (already set in def)
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

// recalculateTitle picks the best title: chain titles override count-based fallback titles.
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
	for _, t := range config.MilestoneTitles() {
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
	var keys []string
	for k := range mm.completed {
		keys = append(keys, k)
	}
	return keys
}

// GetChainsCompleted returns all completed chain keys
func (mm *MilestoneManager) GetChainsCompleted() []string {
	var keys []string
	for k := range mm.chainsCompleted {
		keys = append(keys, k)
	}
	return keys
}

// GetCurrentTitle returns the current civilization title
func (mm *MilestoneManager) GetCurrentTitle() string {
	return mm.currentTitle
}

// computeProgress computes progress indicators for a milestone definition
func (mm *MilestoneManager) computeProgress(def config.MilestoneDef, params MilestoneSnapshotParams) []MilestoneProgress {
	var progress []MilestoneProgress

	if def.MinAge != "" {
		currentOrder := params.AgeOrder[params.Age]
		targetOrder := params.AgeOrder[def.MinAge]
		met := currentOrder >= targetOrder
		progress = append(progress, MilestoneProgress{
			Label:   fmt.Sprintf("Age: %s", def.MinAge),
			Current: float64(currentOrder),
			Target:  float64(targetOrder),
			Met:     met,
		})
	}

	for res, required := range def.MinResources {
		current := params.Resources[res]
		progress = append(progress, MilestoneProgress{
			Label:   fmt.Sprintf("%s", res),
			Current: current,
			Target:  required,
			Met:     current >= required,
		})
	}

	for bld, required := range def.MinBuildings {
		current := float64(params.Buildings[bld])
		progress = append(progress, MilestoneProgress{
			Label:   fmt.Sprintf("%s", bld),
			Current: current,
			Target:  float64(required),
			Met:     int(current) >= required,
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

	// Special conditions
	switch def.Key {
	case "master_builder":
		progress = append(progress, MilestoneProgress{
			Label:   "Buildings built",
			Current: float64(params.TotalBuilt),
			Target:  20,
			Met:     params.TotalBuilt >= 20,
		})
	case "war_machine":
		progress = append(progress, MilestoneProgress{
			Label:   "Soldiers",
			Current: float64(params.SoldierCount),
			Target:  10,
			Met:     params.SoldierCount >= 10,
		})
	case "wonder_builder":
		progress = append(progress, MilestoneProgress{
			Label:   "Wonders",
			Current: float64(params.WonderCount),
			Target:  1,
			Met:     params.WonderCount >= 1,
		})
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
			parts = append(parts, fmt.Sprintf("+%.0f %s", e.Value, e.Target))
		case "permanent_bonus":
			if e.Value < 0 {
				parts = append(parts, fmt.Sprintf("%.0f%% %s", e.Value*100, e.Target))
			} else {
				parts = append(parts, fmt.Sprintf("+%.0f%% %s", e.Value*100, e.Target))
			}
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
