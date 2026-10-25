package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// statsProvider generates the stats overlay text from the current game state.
// Covers: game statistics, active epoch events, and prestige upgrades/points.
func statsProvider(state game.GameState, _ int) string {
	var sb strings.Builder

	// ─── Statistics ───
	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.01", "[gold]═══ Statistics ═══[-]\n\n"))
	s := state.Stats

	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.02", " [gold]Play time:[-]          %s\n", s.PlayTime.Truncate(config.Billion)))
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.03", " [gold]Game ticks:[-]         %s\n", textfmt.Int(state.Tick)))
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.04", " [gold]Buildings built:[-]    %s\n", textfmt.Int(s.TotalBuilt)))
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.05", " [gold]Workers recruited:[-]  %s\n", textfmt.Int(s.TotalRecruited)))
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.06", " [gold]Techs researched:[-]   %s\n", textfmt.Int(state.Research.TotalResearched)))
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.07", " [gold]Expeditions and campaigns done:[-] %s\n", textfmt.Int(state.Military.CompletedCount)))

	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.08", "\n [gold]Ages reached:[-]\n"))
	for _, age := range s.AgesReached {
		fmt.Fprintf(&sb, "   [green]✓[-] %s\n", game.AgeName(age))
	}

	// Time spent in each age of this run. The last entry is the current age, so
	// its time runs to the current tick. Nothing is shown for an old save with no log.
	if len(s.AgeLog) > 0 {
		sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.09", "\n [gold]Ages this run[-]\n"))
		iv := tickInterval(state)
		for i, e := range s.AgeLog {
			if i == len(s.AgeLog)-1 {
				fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.10", "   %-14s %s so far\n", game.AgeName(e.Age),
					game.DurationText(state.Tick-e.Tick, iv)))
				continue
			}
			fmt.Fprintf(&sb, "   %-14s %s\n", game.AgeName(e.Age),
				game.DurationText(s.AgeLog[i+1].Tick-e.Tick, iv))
		}
	}

	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.11", "\n [gold]Total gathered:[-]\n"))
	gKeys := make([]string, 0, len(s.TotalGathered))
	for k := range s.TotalGathered {
		gKeys = append(gKeys, k)
	}
	sort.Strings(gKeys)
	for _, k := range gKeys {
		fmt.Fprintf(&sb, "   %-14s %s\n", textfmt.Capitalize(game.ResourceName(k)), FormatNumber(s.TotalGathered[k]))
	}

	// ─── Lifetime (Account) ───
	// Cross-save aggregates persisted on the account (the accounts design §3.3, Phase 6),
	// distinct from the per-save Statistics above. Renders a placeholder when no
	// account is wired (AccountStats nil) so the panel never blanks out.
	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.12", "\n [yellow]── Lifetime (account) ──[-]\n"))
	if state.AccountStats == nil {
		sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.13", " [gray]No account loaded.[-]\n"))
	} else {
		as := state.AccountStats
		if as.DisplayName != "" {
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.14", " [gray]Account:[-] [white]%s[-]\n", as.DisplayName))
		}
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.15", " [gold]Total prestiges:[-]        %d\n", as.TotalPrestiges))
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.16", " [gold]Civilizations started:[-]  %d\n", as.CivilizationsStarted))

		highestAge := "none"
		if as.HighestAge != "" {
			highestAge = as.HighestAge
			if def, ok := state.Ruleset().Age(as.HighestAge); ok {
				highestAge = def.Name
			}
		}
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.18", " [gold]Highest age ever:[-]       %s\n", highestAge))

		// The badges in a few lines; the badge case has them all. The views
		// come with the spoiler rules applied.
		for _, line := range badgeDigestLines(as.Badges, as.BadgeSummary) {
			sb.WriteString(" " + line + "\n")
		}
	}

	// Epoch & Legacy section
	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.19", "\n [yellow]── Epoch and legacy ──[-]\n"))

	epochDisplay := state.EpochIcon + " " + state.EpochName
	if strings.TrimSpace(epochDisplay) == "" {
		epochDisplay = "none"
	}
	fmt.Fprintf(&sb, " %-20s [cyan]%s[-]\n", vibe.Line("ui.overlay_stats.statsProvider.21", "Current epoch:"), epochDisplay)

	// Both counts come from the civilization log, so a pending catastrophe is
	// in neither and repeat succumbs all count.
	totalCatastrophes := state.CatastrophesEndured + state.CatastrophesSuccumbed
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.22", " %-20s %d  (endured %d, succumbed %d)\n",
		vibe.Line("ui.overlay_stats.statsProvider.23", "Catastrophes:"), totalCatastrophes, state.CatastrophesEndured, state.CatastrophesSuccumbed))
	if state.PendingCatastrophe != "" {
		sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.24", " [red]  One catastrophe is pending. Type catastrophe to choose.[-]\n"))
	} else if state.LastPassage.Pending {
		sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.25", " [red]  The Last Passage is pending. Type catastrophe to choose.[-]\n"))
	}

	// Legacy bonuses
	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.26", "\n [gold]Legacy bonuses:[-]\n"))
	if len(state.LegacyBonuses) == 0 {
		sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.27", "  [gray]None[-]\n"))
	} else {
		activeEpochs := make([]string, 0, len(state.LegacyBonuses))
		for epochKey, active := range state.LegacyBonuses {
			if active {
				activeEpochs = append(activeEpochs, epochKey)
			}
		}
		sort.Strings(activeEpochs)

		if len(activeEpochs) == 0 {
			sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.28", "  [gray]None[-]\n"))
		} else {
			set := state.Ruleset()
			for _, epochKey := range activeEpochs {
				bonuses := set.LegacyBonus(epochKey)
				epochDef, hasEpoch := set.Era(epochKey)
				epochLabel := epochKey
				if hasEpoch {
					epochLabel = epochDef.Name
				}

				fmt.Fprintf(&sb, "  %-16s %s\n", epochLabel+":", legacyBonusLine(state, bonuses))
			}
		}
	}
	if f := state.SuccumbResearchFactor; f > 0 && f < 1 {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.29", "  %-16s research time %s (permanent, %s for each epoch succumbed in)\n", vibe.Line("ui.overlay_stats.statsProvider.30", "Ancient Knowledge:"), game.ResearchFactorText(f), game.ResearchFactorText(game.SuccumbResearchTimeFactor)))
	}
	if state.LastPassage.CosmicLegacy {
		// Applied after the pools, so it never carries a note.
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.31", "  %-16s all production %s, counted in full after every other bonus (permanent, through every prestige)\n", vibe.Line("ui.overlay_stats.statsProvider.32", "Cosmic Legacy:"), textfmt.SignedPercent(game.CosmicLegacyProductionBonus)))
	}

	// Milestone summary hint
	ms := state.Milestones
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.33", "\n [gray]Milestones: %d/%d. Type [white]milestones[-][gray] to see them.[-]\n",
		ms.CompletedCount, ms.TotalCount))

	// ─── Active Events ───
	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.34", "\n[gold]═══ Active events ═══[-]\n\n"))
	if len(state.ActiveEvents) == 0 {
		sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.35", " [gray]No active events.[-]\n"))
	} else {
		for _, evt := range state.ActiveEvents {
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.36", " [yellow]⚡[-] [yellow]%s[-] (%s left)\n", evt.Name, formatTicks(evt.TicksLeft, state)))
			for _, eff := range evt.Effects {
				color := "green"
				if eff.Value < 0 {
					color = "red"
				}
				// The "<res>_rate" case is what a civilization specialty boon or
				// setback arrives as, and it is the common case: without it a
				// boon renders as a name with no magnitude at all.
				if plain, _ := effectMagnitude(eff); plain != "" {
					held := config.Effect{Type: eff.Type, Target: eff.Target, Value: eff.Value}
					fmt.Fprintf(&sb, " [%s]    %s[-]%s\n", color, plain, capTag(state, held, true, "-"))
				}
			}
		}
	}

	// ─── Prestige ───
	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.37", "\n[gold]═══ Prestige ═══[-]\n\n"))
	p := state.Prestige

	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.38", " [gold]Level:[-] [cyan]%d[-]\n", p.Level))
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.39", " [gold]Prestige:[-] [green]x%.2f[-] production\n", p.ProductionMultiplier))
	if cue := state.PrestigeCue; cue.Ready {
		iv := tickInterval(state)
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.40", " [gray]A new run would pass this age in about %s; finishing it takes about %s at today's income.[-]\n",
			game.DurationText(cue.FreshTicks, iv), game.DurationText(cue.FinishTicks, iv)))
	}
	if p.CanPrestige {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.41", " [gold]After this prestige:[-] [green]x%.2f[-] production\n", p.NextProductionMultiplier))
	} else if p.FloorBlocks {
		if p.Floor != "" {
			floorName := state.Ruleset().Name(rules.KindAge, p.Floor)
			if !strings.HasPrefix(strings.ToLower(floorName), "the ") {
				floorName = "the " + floorName
			}
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.43", " [yellow]Prestige opens once you pass %s, the furthest an earlier run reached.[-]\n", floorName))
		}
	}
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.44", " [gold]Era Mastery:[-] %s\n", masteryNowText(state)))
	if next := masteryNextText(state); next != "" {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.45", " [gray]Next prestige: %s.[-]\n", next))
	}
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.46", " [gold]Points earned:[-] [cyan]%d[-] total\n", p.TotalEarned))

	if p.CanPrestige {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.47", " [green]Prestige now for %s.[-]\n", textfmt.Count(p.PendingPoints, vibe.Line("ui.overlay_stats.statsProvider.48", "point"), vibe.Line("ui.overlay_stats.statsProvider.49", "points"))))
		if p.NextAge != "" {
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.50", " [gray]From %s: %s.[-]\n", ageRef(state, p.NextAge), textfmt.Count(p.NextAgePoints, vibe.Line("ui.overlay_stats.statsProvider.51", "point"), vibe.Line("ui.overlay_stats.statsProvider.52", "points"))))
		}
	} else if p.Level == 0 {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.53", " [gray]Reach %s to prestige.[-]\n", ageRef(state, game.PrestigeMinAge)))
	} else {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_stats.statsProvider.54", " [yellow]Reach %s to prestige again.[-]\n", ageRef(state, game.PrestigeMinAge)))
	}

	for _, line := range carryoverStatusLines(state) {
		sb.WriteString(" " + line + "\n")
	}

	// ─── Resource Rates ───
	// What research adds up to: the tech tree is a map, so the totals
	// live here.
	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.55", "\n[gold]═══ Research bonuses ═══[-]\n\n"))
	sb.WriteString(researchSummaryLines(state))

	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.56", "\n[gold]═══ Resource rates ═══[-]\n\n"))
	rateKeys := make([]string, 0, len(state.Resources))
	for k, rs := range state.Resources {
		if rs.Unlocked {
			rateKeys = append(rateKeys, k)
		}
	}
	sort.Strings(rateKeys)
	if len(rateKeys) == 0 {
		sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.57", " [gray]No unlocked resources.[-]\n"))
	} else {
		for _, k := range rateKeys {
			rs := state.Resources[k]
			color := "green"
			if rs.Rate < 0 {
				color = "red"
			}
			fmt.Fprintf(&sb, "  [cyan]%-16s[-] [%s]%s[-]\n", textfmt.Capitalize(game.ResourceName(k)), color, textfmt.Rate(rs.Rate))
		}
	}

	// ─── Active Multipliers ───
	sb.WriteString(vibe.Line("ui.overlay_stats.statsProvider.58", "\n[gold]═══ Active multipliers ═══[-]\n\n"))
	sb.WriteString(renderActiveMultipliers(state))

	return sb.String()
}

// renderActiveMultipliers renders the Active Multipliers section from the
// resolver snapshot carried on the state (state.Modifiers), the single source
// of truth the engine also uses to compute its rates. We rebuild a Resolver
// here rather than re-deriving bonuses from config: Total drives each headline,
// Breakdown drives the per-source attribution, and the two can never disagree
// because they read the same contributions.
func renderActiveMultipliers(state game.GameState) string {
	var sb strings.Builder

	r := game.NewResolver()
	r.AddAll(state.Modifiers) // nil/empty slice is fine — yields no targets

	wrote := false
	for _, target := range r.Targets() {
		if !isPanelMultiplier(target) {
			continue // capacity/flat value (population, all) — not a rate multiplier
		}
		// Build the per-source breakdown FIRST. We render a target if ANY source
		// contributes beyond epsilon — even when those sources net to ×1.0 — so an
		// opposing +10%/-10% pair stays visible instead of silently collapsing.
		breakdown := summarizeBreakdown(r.Breakdown(target))
		if breakdown == "" {
			continue // genuinely empty — every contribution was a no-op
		}
		// The headline is what the engine applies. A pool a limit holds
		// (research speed past +100%, a price cut past its floor) shows the
		// applied total, with a note that says how much was earned: the
		// sources beside it still list everything, so they can add up to
		// more.
		total := r.Total(target)
		if pool, ok := state.Pools[target]; ok && pool.Limited {
			total = 1 + pool.Applied
			for _, m := range r.Breakdown(target) {
				if m.Op == game.OpMul {
					total *= m.Value // morale, an ally: their own multipliers, outside the pool
				}
			}
		}
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
		fmt.Fprintf(&sb, "  [cyan]%-20s[-] [%s]%+.0f%%[-]%s   %s\n",
			multiplierTargetLabel(target), headColor, netPct, poolTag(state, target), breakdown)
		wrote = true
	}

	if !wrote {
		return vibe.Line("ui.overlay_stats.renderActiveMultipliers.01", " [gray]No active multipliers.[-]\n")
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
// — population and "all" — are NOT rate multipliers and must not render as
// percentages here.
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
		if a.hasMul && math.Abs(a.addSum) <= multEpsilon {
			if math.Abs(a.mulProd-1.0) <= multEpsilon {
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
		if math.Abs(eff) <= multEpsilon {
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

// multiplierTargetLabel maps a resolver target id to a panel label in
// glossary words: "All production", "Worker output", "Food production",
// "Game speed".
func multiplierTargetLabel(target string) string {
	return formatBonusName(target)
}

// multiplierSourceLabel maps a modifier Source id to a friendly label.
// Known sources are title-cased; "event:<name>" becomes "Event: <name>".
func multiplierSourceLabel(src string) string {
	if name, ok := strings.CutPrefix(src, "event:"); ok {
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.01", "Event: ") + name
	}
	switch src {
	case "research":
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.02", "Research")
	case "prestige":
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.03", "Prestige")
	case "wonders":
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.04", "Wonders")
	case "permanent":
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.05", "Permanent")
	case "morale":
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.06", "Morale")
	case "event":
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.07", "Event")
	case "diplomacy":
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.08", "Diplomacy")
	case "legacy":
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.09", "Legacy")
	case "cosmic_legacy":
		return vibe.Line("ui.overlay_stats.multiplierSourceLabel.10", "Cosmic Legacy")
	}
	return capitalize(src)
}

// rateEffectLabel names the thing a "<res>_rate" active-event effect acts on,
// in glossary words. Target carries the resource key for the effects
// boon/apply.go builds, but event defs may leave it empty, so fall back to the
// type: "iron_rate" reads "iron production", "gather_rate" "worker output".
func rateEffectLabel(eff game.EventEffectInfo) string {
	if eff.Target != "" {
		return game.EffectTargetName(eff.Target)
	}
	return game.EffectTargetName(eff.Type)
}

// legacyBonusLine renders an epoch's legacy bonus map as
// "Stone production +20%, wood production +20%", sorted by the display
// text so the line never reorders between refreshes.
func legacyBonusLine(state game.GameState, bonuses map[string]float64) string {
	parts := make([]string, 0, len(bonuses))
	for k, v := range bonuses {
		held := config.Effect{Type: "permanent_bonus", Target: k + "_rate", Value: v}
		parts = append(parts, game.EffectTargetName(k)+" "+textfmt.SignedPercent(v)+capTag(state, held, true, "-"))
	}
	sort.Strings(parts)
	return textfmt.Capitalize(strings.Join(parts, ", "))
}
