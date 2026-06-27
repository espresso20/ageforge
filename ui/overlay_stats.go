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
	sb.WriteString(renderActiveMultipliers(state))

	return sb.String()
}

// renderActiveMultipliers renders the Active Multipliers section from the
// resolver snapshot carried on the state (state.Modifiers), the single source
// of truth the engine also uses to compute its rates. We rebuild a Resolver
// here rather than re-deriving bonuses from config: Total drives each headline,
// Breakdown drives the per-source attribution, and the two can never disagree
// because they read the same contributions.
//
// SpeedMultiplier is intentionally NOT a resolver modifier — it's a wonder-gate
// concept layered on top of tick speed — so it gets its own line, as before.
func renderActiveMultipliers(state game.GameState) string {
	var sb strings.Builder

	r := game.NewResolver()
	r.AddAll(state.Modifiers) // nil/empty slice is fine — yields no targets

	wrote := false
	for _, target := range r.Targets() {
		if !isPanelMultiplier(target) {
			continue // capacity/flat value (population, all, bare storage key) — not a rate multiplier
		}
		// Build the per-source breakdown FIRST. We render a target if ANY source
		// contributes beyond epsilon — even when those sources net to ×1.0 — so an
		// opposing +10%/-10% pair stays visible instead of silently collapsing.
		breakdown := summarizeBreakdown(r.Breakdown(target))
		if breakdown == "" {
			continue // genuinely empty — every contribution was a no-op
		}
		total := r.Total(target)
		netPct := (total - 1.0) * 100
		// Headline color by net sign: green bonus, red penalty, white if the row
		// only renders because opposing sources cancel out.
		headColor := "white"
		switch {
		case netPct > 0.5:
			headColor = "green"
		case netPct < -0.5:
			headColor = "red"
		}
		fmt.Fprintf(&sb, "  [cyan]%-20s[-] [%s]%+.0f%%[-]   %s\n",
			multiplierTargetLabel(target), headColor, netPct, breakdown)
		wrote = true
	}

	// SpeedMultiplier: a wonder gate, not a resolver modifier. Render it on its
	// own line so it stays visible.
	if state.SpeedMultiplier > 1.0 {
		fmt.Fprintf(&sb, "  [cyan]%-20s[-] [yellow]×%.2f[-]   [gray]wonders[-]\n",
			"Game Speed", state.SpeedMultiplier)
		wrote = true
	}

	if !wrote {
		return " [gray]No active multipliers[-]\n"
	}
	return sb.String()
}

// multEpsilon is the tolerance below which a single source's contribution is
// treated as a no-op and dropped from a target's per-source breakdown. (A
// target row itself is no longer omitted on net ≈ 1.0 — opposing sources must
// stay visible — so this guards fragment-level noise only.)
const multEpsilon = 0.0005

// isPanelMultiplier reports whether a resolver target is a genuine rate
// multiplier that belongs in the Active Multipliers panel. Capacity/flat values
// — population, "all", and bare-resource storage keys like "food"/"culture" —
// are NOT rate multipliers and must not render as percentages here.
func isPanelMultiplier(target string) bool {
	switch target {
	case "production_all", "gather_rate", "tick_speed", "military_power",
		"expedition_reward", "research_speed", "build_cost":
		return true
	}
	return strings.HasSuffix(target, "_rate")
}

// summarizeBreakdown collapses a target's per-modifier contributions into a
// compact, source-labelled string like
// "Morale ×1.18 · Research +10% · Wonders +5%". Modifiers from the same source
// are merged (additive points summed, multipliers producted) so a source shows
// once. No-op contributions (OpAdd 0, OpMul 1) are dropped. Returns "" if every
// contribution is a no-op.
func summarizeBreakdown(mods []game.Modifier) string {
	type agg struct {
		addSum  float64
		mulProd float64
		hasMul  bool
		order   int
	}
	bySource := map[string]*agg{}
	order := 0
	for _, m := range mods {
		a := bySource[m.Source]
		if a == nil {
			a = &agg{mulProd: 1.0, order: order}
			bySource[m.Source] = a
			order++
		}
		if m.Op == game.OpMul {
			a.mulProd *= m.Value
			a.hasMul = true
		} else {
			a.addSum += m.Value
		}
	}

	// Stable order: first-seen (matches resolver insertion order).
	sources := make([]string, 0, len(bySource))
	for src := range bySource {
		sources = append(sources, src)
	}
	sort.Slice(sources, func(i, j int) bool {
		return bySource[sources[i]].order < bySource[sources[j]].order
	})

	parts := make([]string, 0, len(sources))
	for _, src := range sources {
		a := bySource[src]
		label := multiplierSourceLabel(src)
		// Prefer a multiplicative display only when the source contributed a
		// genuine OpMul (e.g. morale ×1.18) and no additive points alongside.
		if a.hasMul && absFloat(a.addSum) <= multEpsilon {
			if absFloat(a.mulProd-1.0) <= multEpsilon {
				continue // ×1.00 — no-op
			}
			// Color by sign: a multiplier above 1.0 is a bonus, below is a penalty.
			color := "green"
			if a.mulProd < 1.0 {
				color = "red"
			}
			parts = append(parts, fmt.Sprintf("[%s]%s ×%.2f[-]", color, label, a.mulProd))
			continue
		}
		// Additive (the common case). Fold any stray OpMul into the percent so
		// nothing is silently dropped.
		eff := (1+a.addSum)*a.mulProd - 1.0
		if absFloat(eff) <= multEpsilon {
			continue // +0% — no-op
		}
		// Color by sign: positive contribution is a bonus, negative a penalty.
		color := "green"
		if eff < 0 {
			color = "red"
		}
		parts = append(parts, fmt.Sprintf("[%s]%s %+.0f%%[-]", color, label, eff*100))
	}
	return strings.Join(parts, " [gray]·[-] ")
}

// multiplierTargetLabel maps a resolver target id to a friendly panel label.
// Reuses formatBonusName (overlay_research.go) which already handles
// production_all, gather_rate, <res>_rate, military_power, etc.
func multiplierTargetLabel(target string) string {
	switch target {
	case "tick_speed":
		return "Tick Speed"
	}
	return formatBonusName(target)
}

// multiplierSourceLabel maps a modifier Source id to a friendly label.
// Known sources are title-cased; "event:<name>" becomes "Event: <name>".
func multiplierSourceLabel(src string) string {
	if name, ok := strings.CutPrefix(src, "event:"); ok {
		return "Event: " + name
	}
	switch src {
	case "research":
		return "Research"
	case "prestige":
		return "Prestige"
	case "wonders":
		return "Wonders"
	case "permanent":
		return "Permanent"
	case "morale":
		return "Morale"
	case "event":
		return "Event"
	}
	return capitalize(src)
}

// absFloat is a tiny abs helper for epsilon comparisons (avoids importing math
// for a single call site).
func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
