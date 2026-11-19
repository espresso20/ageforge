package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// multSrc is one source's part of a target's multiplier: its label, the
// amount as written ("+10%" or "×1.18"), and whether it lowers the target.
type multSrc struct {
	label, amount string
	neg           bool
}

// text is the source as one phrase.
func (s multSrc) text() string { return s.label + " " + s.amount }

// multRow is one line of the Active multipliers table: the target, what the
// engine applies to it in percent, the pool's note, and each source.
type multRow struct {
	target  string
	pct     float64
	note    string
	sources []multSrc
}

// activeMultipliers reads the resolver snapshot carried on the state
// (state.Modifiers), the single source of truth the engine also uses to
// compute its rates: Total drives each headline, Breakdown the per-source
// attribution, and the two can never disagree because they read the same
// contributions.
func activeMultipliers(state game.GameState) []multRow {
	r := game.NewResolver()
	r.AddAll(state.Modifiers) // nil/empty slice is fine: it yields no targets
	var rows []multRow
	for _, target := range r.Targets() {
		if !isPanelMultiplier(target) {
			continue // capacity/flat value (population, all), not a rate multiplier
		}
		// A target is listed if ANY source contributes beyond epsilon, even when
		// they net to x1.0, so an opposing +10%/-10% pair stays visible.
		sources := summarizeBreakdown(r.Breakdown(target))
		if len(sources) == 0 {
			continue
		}
		// The headline is what the engine applies. A pool a limit holds
		// (research speed past +100%, a price cut past its floor) shows the
		// applied total, with a note that says how much was earned: the
		// sources beside it still list everything, so they can add up to more.
		total := r.Total(target)
		if pool, ok := state.Pools[target]; ok && pool.Limited {
			total = 1 + pool.Applied
			for _, m := range r.Breakdown(target) {
				if m.Op == game.OpMul {
					total *= m.Value // morale, an ally: their own multipliers, outside the pool
				}
			}
		}
		rows = append(rows, multRow{
			target:  multiplierTargetLabel(target),
			pct:     (total - 1.0) * 100,
			note:    kit.Strip(poolTag(state, target)),
			sources: sources,
		})
	}
	return rows
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

// summarizeBreakdown collapses a target's per-modifier contributions into
// source-labelled amounts. Modifiers from the same source are merged
// (additive points summed, multipliers producted) so a source shows once, in
// the order the resolver met them. No-op contributions (OpAdd 0, OpMul 1) are
// dropped.
func summarizeBreakdown(mods []game.Modifier) []multSrc {
	type agg struct {
		addSum  float64
		mulProd float64
		hasMul  bool
		order   int
	}
	bySource := map[string]*agg{}
	for _, m := range mods {
		a := bySource[m.Source]
		if a == nil {
			a = &agg{mulProd: 1.0, order: len(bySource)}
			bySource[m.Source] = a
		}
		if m.Op == game.OpMul {
			a.mulProd *= m.Value
			a.hasMul = true
		} else {
			a.addSum += m.Value
		}
	}
	sources := make([]string, 0, len(bySource))
	for src := range bySource {
		sources = append(sources, src)
	}
	sort.Slice(sources, func(i, j int) bool { return bySource[sources[i]].order < bySource[sources[j]].order })

	var parts []multSrc
	for _, src := range sources {
		a, label := bySource[src], multiplierSourceLabel(src)
		// A multiplicative display only when the source contributed a genuine
		// OpMul (morale x1.18) and no additive points alongside.
		if a.hasMul && math.Abs(a.addSum) <= multEpsilon {
			if math.Abs(a.mulProd-1.0) > multEpsilon {
				parts = append(parts, multSrc{label, fmt.Sprintf("×%.2f", a.mulProd), a.mulProd < 1.0})
			}
			continue
		}
		// Additive (the common case). Any stray OpMul folds into the percent so
		// nothing is silently dropped.
		if eff := (1+a.addSum)*a.mulProd - 1.0; math.Abs(eff) > multEpsilon {
			parts = append(parts, multSrc{label, fmt.Sprintf("%+.0f%%", eff*100), eff < 0})
		}
	}
	return parts
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
		return rowid.Line("ui.overlay_stats.multiplierSourceLabel.01", "Event: ") + name
	}
	switch src {
	case "research":
		return rowid.Line("ui.overlay_stats.multiplierSourceLabel.02", "Research")
	case "prestige":
		return rowid.Line("ui.overlay_stats.multiplierSourceLabel.03", "Prestige")
	case "wonders":
		return rowid.Line("ui.overlay_stats.multiplierSourceLabel.04", "Wonders")
	case "permanent":
		return rowid.Line("ui.overlay_stats.multiplierSourceLabel.05", "Permanent")
	case "morale":
		return txtMorale.String()
	case "event":
		return rowid.Line("ui.overlay_stats.multiplierSourceLabel.07", "Event")
	case "diplomacy":
		return rowid.Line("ui.overlay_stats.multiplierSourceLabel.08", "Diplomacy")
	case "legacy":
		return rowid.Line("ui.overlay_stats.multiplierSourceLabel.09", "Legacy")
	case "cosmic_legacy":
		return rowid.Line("ui.overlay_stats.multiplierSourceLabel.10", "Cosmic Legacy")
	}
	return textfmt.Capitalize(src)
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
