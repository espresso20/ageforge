package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// statsProvider generates the stats overlay text from the current game state.
// Covers: game statistics, active epoch events, and prestige upgrades/points.
func statsProvider(state game.GameState, _ int) string {
	var sb strings.Builder

	// ─── Statistics ───
	sb.WriteString("[gold]═══ Statistics ═══[-]\n\n")
	s := state.Stats

	fmt.Fprintf(&sb, " [gold]Play Time:[-]          %s\n", s.PlayTime.Truncate(1e9))
	fmt.Fprintf(&sb, " [gold]Total Ticks:[-]        %d\n", state.Tick)
	fmt.Fprintf(&sb, " [gold]Buildings Built:[-]    %d\n", s.TotalBuilt)
	fmt.Fprintf(&sb, " [gold]Workers Recruited:[-]  %d\n", s.TotalRecruited)
	fmt.Fprintf(&sb, " [gold]Techs Researched:[-]   %d\n", state.Research.TotalResearched)
	fmt.Fprintf(&sb, " [gold]Expeditions Done:[-]   %d\n", state.Military.CompletedCount)

	sb.WriteString("\n [gold]Ages Reached:[-]\n")
	for _, age := range s.AgesReached {
		fmt.Fprintf(&sb, "   [green]✓[-] %s\n", age)
	}

	sb.WriteString("\n [gold]Total Gathered:[-]\n")
	gKeys := make([]string, 0, len(s.TotalGathered))
	for k := range s.TotalGathered {
		gKeys = append(gKeys, k)
	}
	sort.Strings(gKeys)
	for _, k := range gKeys {
		fmt.Fprintf(&sb, "   %-12s %s\n", k, FormatNumber(s.TotalGathered[k]))
	}

	// Epoch & Legacy section
	sb.WriteString("\n [yellow]── Epoch & Legacy ──[-]\n")

	epochDisplay := state.EpochIcon + " " + state.EpochName
	if strings.TrimSpace(epochDisplay) == "" {
		epochDisplay = "—"
	}
	fmt.Fprintf(&sb, " %-20s [cyan]%s[-]\n", "Current Epoch:", epochDisplay)

	epochsSurvivedCount := len(state.CatastropheHistory)
	fmt.Fprintf(&sb, " %-20s %d\n", "Epochs Survived:", epochsSurvivedCount)

	succumbedCount := 0
	for _, active := range state.LegacyBonuses {
		if active {
			succumbedCount++
		}
	}
	totalCatastrophes := len(state.CatastropheHistory)
	enduredCount := totalCatastrophes - succumbedCount
	fmt.Fprintf(&sb, " %-20s %d  (Endured: %d  Succumbed: %d)\n",
		"Catastrophes:", totalCatastrophes, enduredCount, succumbedCount)

	// Legacy Bonuses
	sb.WriteString("\n [gold]Legacy Bonuses:[-]\n")
	if len(state.LegacyBonuses) == 0 {
		sb.WriteString("  [gray]None[-]\n")
	} else {
		activeEpochs := make([]string, 0, len(state.LegacyBonuses))
		for epochKey, active := range state.LegacyBonuses {
			if active {
				activeEpochs = append(activeEpochs, epochKey)
			}
		}
		sort.Strings(activeEpochs)

		if len(activeEpochs) == 0 {
			sb.WriteString("  [gray]None[-]\n")
		} else {
			epochByKey := config.EpochByKey()
			for _, epochKey := range activeEpochs {
				bonuses := config.LegacyBonusForEpoch(epochKey)
				epochDef, hasEpoch := epochByKey[epochKey]
				epochLabel := epochKey
				if hasEpoch {
					epochLabel = epochDef.Name
				}

				bonusKeys := make([]string, 0, len(bonuses))
				for k := range bonuses {
					bonusKeys = append(bonusKeys, k)
				}
				sort.Strings(bonusKeys)
				parts := make([]string, 0, len(bonusKeys))
				for _, k := range bonusKeys {
					parts = append(parts, fmt.Sprintf("%s +%.0f%%", k, bonuses[k]*100))
				}
				fmt.Fprintf(&sb, "  %-16s %s\n", epochLabel+":", strings.Join(parts, ",  "))
			}
		}
	}

	// Milestone summary hint
	ms := state.Milestones
	fmt.Fprintf(&sb, "\n [gray]Milestones: %d/%d — type [white]milestones[-][gray] to view[-]\n",
		ms.CompletedCount, ms.TotalCount)

	// ─── Active Events ───
	sb.WriteString("\n[gold]═══ Active Events ═══[-]\n\n")
	if len(state.ActiveEvents) == 0 {
		sb.WriteString(" [gray]No active events[-]\n")
	} else {
		for _, evt := range state.ActiveEvents {
			fmt.Fprintf(&sb, " [yellow]⚡[-] [yellow]%s[-] (%d ticks left)\n", evt.Name, evt.TicksLeft)
		}
	}

	// ─── Prestige ───
	sb.WriteString("\n[gold]═══ Prestige ═══[-]\n\n")
	p := state.Prestige

	fmt.Fprintf(&sb, " [gold]Level:[-] [cyan]%d[-]", p.Level)
	if p.PassiveBonus > 0 {
		fmt.Fprintf(&sb, "  [green]+%.0f%% production[-]", p.PassiveBonus*100)
	}
	sb.WriteString("\n")
	fmt.Fprintf(&sb, " [gold]Points:[-] [cyan]%d[-] available / %d total\n", p.Available, p.TotalEarned)

	if p.CanPrestige {
		fmt.Fprintf(&sb, " [green]Can prestige for %d pts![-]\n", p.PendingPoints)
	} else if p.Level == 0 {
		sb.WriteString(" [gray]Reach Medieval Age to prestige[-]\n")
	} else {
		sb.WriteString(" [yellow]Reach Medieval Age to prestige again[-]\n")
	}

	upgradeKeys := []string{
		"gather_boost", "storage_bonus", "research_speed", "military_power",
		"starting_food", "starting_wood", "population_cap", "expedition_loot",
	}
	hasPurchased := false
	for _, key := range upgradeKeys {
		u, ok := p.Upgrades[key]
		if !ok || u.Tier == 0 {
			continue
		}
		if !hasPurchased {
			sb.WriteString("\n [gold]Upgrades:[-]\n")
			hasPurchased = true
		}
		bar := ProgressBar(float64(u.Tier), float64(u.MaxTier), 5)
		fmt.Fprintf(&sb, "  %s %s [green]%s[-]\n", u.Name, bar, u.Effect)
	}

	if !hasPurchased && p.Level > 0 {
		sb.WriteString("\n [gray]No upgrades purchased yet[-]\n")
		sb.WriteString(" [gray]Type 'prestige shop' to browse[-]\n")
	}

	// ─── Resource Rates ───
	sb.WriteString("\n[gold]═══ Resource Rates (per tick) ═══[-]\n\n")
	rateKeys := make([]string, 0, len(state.Resources))
	for k, rs := range state.Resources {
		if rs.Unlocked {
			rateKeys = append(rateKeys, k)
		}
	}
	sort.Strings(rateKeys)
	if len(rateKeys) == 0 {
		sb.WriteString(" [gray]No unlocked resources[-]\n")
	} else {
		for _, k := range rateKeys {
			rs := state.Resources[k]
			if rs.Rate >= 0 {
				fmt.Fprintf(&sb, "  [cyan]%-16s[-] [green]+%.2f /tick[-]\n", k, rs.Rate)
			} else {
				fmt.Fprintf(&sb, "  [cyan]%-16s[-] [red]%.2f /tick[-]\n", k, rs.Rate)
			}
		}
	}

	// ─── Active Multipliers ───
	sb.WriteString("\n[gold]═══ Active Multipliers ═══[-]\n\n")

	// Build attribution maps so we can label each bonus by source.
	// These are re-derived from config definitions purely for display;
	// the canonical totals always come from state.PermanentBonuses.
	type bonusAttrib struct {
		milestones float64
		research   float64
		prestige   float64
		legacy     float64
		wonders    float64
		epoch      float64
	}
	attrib := map[string]*bonusAttrib{}

	ensureTarget := func(target string) {
		if attrib[target] == nil {
			attrib[target] = &bonusAttrib{}
		}
	}

	// 1. Milestone permanent_bonus effects
	msDefs := config.MilestoneByKey()
	for msKey, msInfo := range state.Milestones.Milestones {
		if !msInfo.Completed {
			continue
		}
		def, ok := msDefs[msKey]
		if !ok {
			continue
		}
		for _, eff := range def.Rewards {
			if eff.Type == "permanent_bonus" {
				ensureTarget(eff.Target)
				attrib[eff.Target].milestones += eff.Value
			}
		}
	}

	// 2. Research bonus effects
	techDefs := config.TechByKey()
	for techKey, techState := range state.Research.Techs {
		if !techState.Researched {
			continue
		}
		def, ok := techDefs[techKey]
		if !ok {
			continue
		}
		for _, eff := range def.Effects {
			if eff.Type == "bonus" {
				ensureTarget(eff.Target)
				attrib[eff.Target].research += eff.Value
			}
		}
	}

	// 3. Prestige — passive production_all bonus
	if state.Prestige.PassiveBonus > 0 {
		ensureTarget("production_all")
		attrib["production_all"].prestige += state.Prestige.PassiveBonus
	}
	// Prestige — upgrade rate bonuses
	prestigeDefs := config.PrestigeUpgradeByKey()
	for key, uState := range state.Prestige.Upgrades {
		if uState.Tier <= 0 {
			continue
		}
		def, ok := prestigeDefs[key]
		if !ok {
			continue
		}
		if def.EffectType == "rate_bonus" {
			ensureTarget(def.EffectKey)
			attrib[def.EffectKey].prestige += def.PerTier * float64(uState.Tier)
		}
	}

	// 4. Legacy bonuses
	for epochKey, active := range state.LegacyBonuses {
		if !active {
			continue
		}
		legBonuses := config.LegacyBonusForEpoch(epochKey)
		for target, mult := range legBonuses {
			ensureTarget(target)
			attrib[target].legacy += mult
		}
	}

	// 5. Speed multiplier (from wonders)
	if state.SpeedMultiplier > 1.0 {
		ensureTarget("speed_multiplier")
		attrib["speed_multiplier"].wonders += state.SpeedMultiplier - 1.0
	}

	// 6. Wonder bonus effects (Type "bonus" — production_all, knowledge_rate, expedition_reward, etc.)
	// Iterate all built wonders and accumulate their percentage/multiplier bonuses.
	buildingDefs := config.BaseBuildings()
	buildingDefsByKey := make(map[string]config.BuildingDef, len(buildingDefs))
	for _, bd := range buildingDefs {
		buildingDefsByKey[bd.Key] = bd
	}
	for bKey, bState := range state.Buildings {
		if bState.Count == 0 {
			continue
		}
		def, ok := buildingDefsByKey[bKey]
		if !ok || def.Category != "wonder" {
			continue
		}
		for _, eff := range def.Effects {
			if eff.Type == "bonus" {
				ensureTarget(eff.Target)
				attrib[eff.Target].wonders += eff.Value * float64(bState.Count)
			}
		}
	}

	// Collect targets that have any nonzero contribution
	activeTargets := make([]string, 0, len(attrib))
	for target, bc := range attrib {
		total := bc.milestones + bc.research + bc.prestige + bc.legacy + bc.wonders
		if total > 0 {
			activeTargets = append(activeTargets, target)
		}
	}
	if state.SpeedMultiplier > 1.0 {
		if attrib["speed_multiplier"] == nil {
			activeTargets = append(activeTargets, "speed_multiplier")
		}
	}

	if len(activeTargets) == 0 {
		sb.WriteString(" [gray]No active multipliers[-]\n")
	} else {
		// Sort: production_all first, then alphabetical
		sort.Slice(activeTargets, func(i, j int) bool {
			if activeTargets[i] == "production_all" {
				return true
			}
			if activeTargets[j] == "production_all" {
				return false
			}
			return activeTargets[i] < activeTargets[j]
		})

		for _, target := range activeTargets {
			var total float64
			if target == "speed_multiplier" {
				total = state.SpeedMultiplier - 1.0
			} else {
				total = state.PermanentBonuses[target]
			}
			a := attrib[target]
			if a == nil {
				a = &bonusAttrib{}
			}

			if target == "speed_multiplier" {
				fmt.Fprintf(&sb, "  [cyan]%-20s[-] [yellow]+%.1fx[-]\n", target, total)
			} else {
				fmt.Fprintf(&sb, "  [cyan]%-20s[-] [yellow]+%.0f%%[-]\n", target, total*100)
			}
			if a.milestones > 0 {
				fmt.Fprintf(&sb, "  [gray]  milestones     +%.0f%%[-]\n", a.milestones*100)
			}
			if a.research > 0 {
				fmt.Fprintf(&sb, "  [gray]  research       +%.0f%%[-]\n", a.research*100)
			}
			if a.prestige > 0 {
				fmt.Fprintf(&sb, "  [gray]  prestige       +%.0f%%[-]\n", a.prestige*100)
			}
			if a.legacy > 0 {
				fmt.Fprintf(&sb, "  [gray]  legacy         +%.0f%%[-]\n", a.legacy*100)
			}
			if a.epoch > 0 {
				fmt.Fprintf(&sb, "  [gray]  epoch events   +%.0f%%[-]\n", a.epoch*100)
			}
			if a.wonders > 0 {
				if target == "speed_multiplier" {
					fmt.Fprintf(&sb, "  [gray]  wonders        +%.1fx[-]\n", a.wonders)
				} else {
					fmt.Fprintf(&sb, "  [gray]  wonders        +%.0f%%[-]\n", a.wonders*100)
				}
			}
		}
	}

	return sb.String()
}
