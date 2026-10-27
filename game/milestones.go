package game

import (
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
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
	rules           *rules.Set
	defs            []config.MilestoneDef
	completed       map[string]bool
	chains          []config.MilestoneChainDef
	chainsCompleted map[string]bool
	currentTitle    string
	// titles is the count-based fallback title ladder, held so
	// recalculateTitle (run every tick from checkMilestones) doesn't rebuild it.
	titles []config.TitleDef
	// plans holds, for each milestone in defs, the text of its snapshot row that
	// only the ruleset decides, built once in Rebind (see milestonePlan).
	plans []milestonePlan
}

// NewMilestoneManagerWith creates a new milestone manager with set's
// milestones, chains and titles.
func NewMilestoneManagerWith(set *rules.Set) *MilestoneManager {
	mm := &MilestoneManager{
		completed:       make(map[string]bool),
		chainsCompleted: make(map[string]bool),
	}
	mm.Rebind(set)
	return mm
}

// Rebind moves the manager onto set: it takes set's milestones, chains and
// titles. What is completed stays.
func (mm *MilestoneManager) Rebind(set *rules.Set) {
	mm.rules = set
	mm.defs = set.Milestones()
	mm.chains = set.MilestoneChains()
	mm.titles = set.MilestoneTitles()
	mm.plans = make([]milestonePlan, len(mm.defs))
	for i, def := range mm.defs {
		mm.plans[i] = mm.planFor(def)
	}
	mm.recalculateTitle() // the title in the words of this set, at once
}

// CheckMilestones checks all milestones against the town's numbers in params
// (the same ones the snapshot reads; params.Resources holds the work done on
// techs as "research") and the researched techs, which it only reads.
// Returns list of newly completed milestones.
func (mm *MilestoneManager) CheckMilestones(params MilestoneSnapshotParams, researched map[string]bool) []config.MilestoneDef {
	var completed []config.MilestoneDef

	for _, def := range mm.defs {
		if mm.completed[def.Key] {
			continue
		}

		if mm.checkMilestone(def, params, researched) {
			mm.completed[def.Key] = true
			completed = append(completed, def)
		}
	}

	return completed
}

func (mm *MilestoneManager) checkMilestone(def config.MilestoneDef, params MilestoneSnapshotParams, researched map[string]bool) bool {
	// Check min tick
	if def.MinTick > 0 && params.Tick < def.MinTick {
		return false
	}

	// Check age
	if def.MinAge != "" {
		if params.AgeOrder[params.Age] < params.AgeOrder[def.MinAge] {
			return false
		}
	}

	// Check resources
	for res, required := range def.MinResources {
		if params.Resources[res] < required {
			return false
		}
	}

	// Check buildings
	// A building named by key counts the building and every later rung of its
	// line: upgrading copies away never takes the progress back.
	for bld, required := range def.MinBuildings {
		if params.heldFrom(bld) < required {
			return false
		}
	}
	if sum := def.MinBuildingSum; sum.Count > 0 {
		if params.heldFromAny(sum.Keys) < sum.Count {
			return false
		}
	}

	// Check population
	if def.MinPopulation > 0 && params.Population < def.MinPopulation {
		return false
	}

	// Check tech count
	if def.MinTechCount > 0 && params.TechCount < def.MinTechCount {
		return false
	}

	// Check specific techs
	for _, tech := range def.RequiredTechs {
		if !researched[tech] {
			return false
		}
	}

	// Run counters (zero asks for nothing)
	if params.TotalBuilt < def.MinTotalBuilt || params.SoldiersTrained < def.MinSoldiersTrained ||
		params.WonderCount < def.MinWonders || params.KnowledgeCount < def.MinKnowledgeWorkers {
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
	mm.currentTitle = mm.titleFor(mm.chains, mm.titles)
}

// titleFor is the title the completed milestones earn from these chains and
// this ladder of count-based titles.
func (mm *MilestoneManager) titleFor(chains []config.MilestoneChainDef, titles []config.TitleDef) string {
	// Chain titles take priority (use latest completed chain's title)
	bestChainTitle := ""
	for _, chain := range chains {
		if mm.chainsCompleted[chain.Key] {
			bestChainTitle = chain.Title
		}
	}
	if bestChainTitle != "" {
		return bestChainTitle
	}

	// Fallback to count-based titles
	count := len(mm.completed)
	title := ""
	for _, t := range titles {
		if count >= t.MinMilestones {
			title = t.Title
		}
	}
	return title
}

// savedTitle is the title a save holds: in plain English whatever script the
// game is read in, so a save made under VibeMode reads the same as one made
// without it. The title in play is worked out again every tick.
func (mm *MilestoneManager) savedTitle() string {
	if !rules.IsCore(mm.rules) {
		return mm.currentTitle
	}
	// Worked out from what is completed, not read from currentTitle, which
	// is the other script's until the next tick after a switch.
	plain := rules.CoreFor(false)
	return mm.titleFor(plain.MilestoneChains(), plain.MilestoneTitles())
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

// milestonePlan is what a milestone's snapshot row holds that only the ruleset
// decides: the reward line, the label of each progress row and the order of the
// resource and building rows. A snapshot is taken twice a second; this is built
// once (Rebind), so a snapshot only fills in the numbers.
type milestonePlan struct {
	rewardText string
	ageLabel   string
	resKeys    []string // sorted: map order would make the bars flicker
	resLabels  []string
	bldKeys    []string // sorted, same reason
	bldLabels  []string
	sumLabel   string
	rows       int // how many progress rows the milestone has
}

// planFor builds def's plan.
func (mm *MilestoneManager) planFor(def config.MilestoneDef) milestonePlan {
	plan := milestonePlan{rewardText: formatRewards(def.Rewards)}
	if def.MinTick > 0 {
		plan.rows++
	}
	if def.MinAge != "" {
		plan.ageLabel = vibe.Line("game.milestones.planFor.01", "Age: ") + mm.rules.Name(rules.KindAge, def.MinAge)
		plan.rows++
	}
	// Sort resource keys before iterating — map iteration order is random,
	// which would cause progress bar lines to flicker each tick.
	plan.resKeys = make([]string, 0, len(def.MinResources))
	for res := range def.MinResources {
		plan.resKeys = append(plan.resKeys, res)
	}
	sort.Strings(plan.resKeys)
	for _, res := range plan.resKeys {
		plan.resLabels = append(plan.resLabels, textfmt.Capitalize(ResourceName(res)))
	}
	// Sort building keys before iterating — same reason as resources.
	plan.bldKeys = make([]string, 0, len(def.MinBuildings))
	for bld := range def.MinBuildings {
		plan.bldKeys = append(plan.bldKeys, bld)
	}
	sort.Strings(plan.bldKeys)
	for _, bld := range plan.bldKeys {
		plan.bldLabels = append(plan.bldLabels, mm.rules.Name(rules.KindBuilding, bld))
	}
	plan.rows += len(plan.resKeys) + len(plan.bldKeys)
	// A building sum is one row: "Trading Posts and Merchant Quarters".
	if sum := def.MinBuildingSum; sum.Count > 0 {
		names := make([]string, 0, len(sum.Keys))
		for _, bld := range sum.Keys {
			names = append(names, pluralName(2, mm.rules.Name(rules.KindBuilding, bld)))
		}
		plan.sumLabel = textfmt.List(names)
		plan.rows++
	}
	for _, need := range []int{def.MinPopulation, def.MinTechCount, def.MinTotalBuilt, def.MinSoldiersTrained, def.MinWonders, def.MinKnowledgeWorkers} {
		if need > 0 {
			plan.rows++
		}
	}
	return plan
}

// progressFor fills in def's progress rows from its plan and the town's
// numbers, in room taken from rows (nil: its own).
func (mm *MilestoneManager) progressFor(def config.MilestoneDef, plan *milestonePlan, params MilestoneSnapshotParams, rows *sliceArena[MilestoneProgress]) []MilestoneProgress {
	progress := rows.take(plan.rows)

	if def.MinTick > 0 {
		progress = append(progress, row(vibe.Line("game.milestones.progressFor.01", "Ticks survived"), params.Tick, def.MinTick))
	}

	if def.MinAge != "" {
		progress = append(progress, row(plan.ageLabel, params.AgeOrder[params.Age], params.AgeOrder[def.MinAge]))
	}

	for i, res := range plan.resKeys {
		progress = append(progress, row(plan.resLabels[i], params.Resources[res], def.MinResources[res]))
	}

	for i, bld := range plan.bldKeys {
		progress = append(progress, row(plan.bldLabels[i], params.heldFrom(bld), def.MinBuildings[bld]))
	}
	if sum := def.MinBuildingSum; sum.Count > 0 {
		progress = append(progress, row(plan.sumLabel, params.heldFromAny(sum.Keys), sum.Count))
	}

	if def.MinPopulation > 0 {
		progress = append(progress, row(vibe.Line("game.milestones.progressFor.02", "Population"), params.Population, def.MinPopulation))
	}

	if def.MinTechCount > 0 {
		progress = append(progress, row(vibe.Line("game.milestones.progressFor.03", "Technologies"), params.TechCount, def.MinTechCount))
	}

	// Run counters
	progress = appendCounter(progress, vibe.Line("game.milestones.progressFor.04", "Buildings built"), params.TotalBuilt, def.MinTotalBuilt)
	progress = appendCounter(progress, vibe.Line("game.milestones.progressFor.05", "Soldiers trained"), params.SoldiersTrained, def.MinSoldiersTrained)
	progress = appendCounter(progress, vibe.Line("game.milestones.progressFor.06", "Wonders"), params.WonderCount, def.MinWonders)
	progress = appendCounter(progress, vibe.Line("game.milestones.progressFor.07", "Knowledge workers"), params.KnowledgeCount, def.MinKnowledgeWorkers)

	return progress
}

// row is one progress row: how far cur is toward need, met once it gets there.
// Ints compare as ints, floats as floats.
func row[N int | float64](label string, cur, need N) MilestoneProgress {
	return MilestoneProgress{Label: label, Current: float64(cur), Target: float64(need), Met: cur >= need}
}

// appendCounter adds a run-counter row when the milestone asks for any.
func appendCounter(progress []MilestoneProgress, label string, current, need int) []MilestoneProgress {
	if need <= 0 {
		return progress
	}
	return append(progress, row(label, current, need))
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
			if e.Target == config.ResearchTarget {
				parts = append(parts, vibe.Line("game.milestones.formatRewards.01", "research progress"))
				continue
			}
			parts = append(parts, "+"+Amount(e.Value, e.Target))
		case "permanent_bonus":
			parts = append(parts, textfmt.SignedPercent(e.Value)+" "+EffectTargetName(e.Target))
		}
	}
	return strings.Join(parts, ", ")
}

// Snapshot returns milestone state for UI with progress, chains, and titles
func (mm *MilestoneManager) Snapshot(params MilestoneSnapshotParams) MilestoneState {
	milestones := make(map[string]MilestoneInfo, len(mm.defs))
	var rewards sliceArena[config.Effect] // one backing array for every milestone's copy
	var rows sliceArena[MilestoneProgress]
	for i, def := range mm.defs {
		rewards.reserve(def.Rewards)
		rows.want += mm.plans[i].rows
	}

	for i, def := range mm.defs {
		completed := mm.completed[def.Key]
		plan := &mm.plans[i]
		progress := mm.progressFor(def, plan, params, &rows)
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

		milestones[def.Key] = MilestoneInfo{
			Name:        def.Name,
			Description: def.Description,
			Category:    def.Category,
			Hidden:      def.Hidden,
			Visible:     visible,
			Completed:   completed,
			RewardText:  plan.rewardText,
			Rewards:     rewards.clone(def.Rewards), // def is the manager's table
			Progress:    progress,
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
