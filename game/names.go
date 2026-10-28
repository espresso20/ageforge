package game

import (
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// names.go turns config keys into the names a player reads. Every log line,
// refusal, toast and panel that mentions a resource, building, age, tech,
// civilization, trade route or effect target goes through these helpers, so
// "iron_ore" never reaches the screen where "iron ore" should.
//
// The names are a ruleset's (rules.Set.Name). The functions below read the
// core set, for callers with no engine or snapshot to hand; the engine and
// its managers name things from their own set.

// ResourceName returns the lowercase display name of a resource for use
// mid-sentence ("iron ore", "dark matter"): the wording config.ResourceLabel
// gives, read from the core ruleset.
func ResourceName(key string) string {
	return rules.Core().ResourceLabel(key)
}

// BuildingName returns a building's display name ("Lumber Mill").
func BuildingName(key string) string {
	return rules.Core().Name(rules.KindBuilding, key)
}

// BuildingNames returns the plural-aware display name for n copies of a
// building: "1 Farm", "3 Farms".
func BuildingCount(n int, key string) string {
	return buildingCountIn(rules.Core(), n, key)
}

// buildingCountIn is BuildingCount with the building named by set.
func buildingCountIn(set *rules.Set, n int, key string) string {
	return textfmt.Int(n) + " " + pluralName(n, set.Name(rules.KindBuilding, key))
}

// pluralName pluralizes a display name for a count. "X of Y" pluralizes its
// head ("Monuments of Ages"). A name ending in s is already plural or
// invariant (Barracks, Ironworks, Standing Stones) unless it ends in ss or
// us, the singular endings that take "es" (Fortresses, Campuses, Nexuses).
func pluralName(n int, name string) string {
	if n == 1 || n == -1 {
		return name
	}
	if head, rest, ok := strings.Cut(name, " of "); ok {
		return pluralName(n, head) + " of " + rest
	}
	switch {
	case strings.HasSuffix(name, "ss"), strings.HasSuffix(name, "us"),
		strings.HasSuffix(name, "x"), strings.HasSuffix(name, "ch"),
		strings.HasSuffix(name, "sh"):
		return name + "es"
	case strings.HasSuffix(name, "s"):
		return name
	case strings.HasSuffix(name, "y") && len(name) > 1 && !strings.ContainsRune("aeiou", rune(name[len(name)-2])):
		return name[:len(name)-1] + "ies"
	}
	return name + "s"
}

// AgeName returns an age's display name ("Industrial Age").
func AgeName(key string) string {
	return rules.Core().Name(rules.KindAge, key)
}

// TechName returns a tech's display name ("Steam Power").
func TechName(key string) string {
	return rules.Core().Name(rules.KindTech, key)
}

// CivName returns a civilization's display name ("Merchant Guild").
func CivName(key string) string {
	return rules.Core().Name(rules.KindCiv, key)
}

// EffectTargetName names a bonus target the way the glossary does:
// production_all → "all production", gather_rate → "worker output",
// food_rate → "food production", tick_speed → "game speed". It extends
// config.EffectTargetLabel (the one source of target wording) with the
// bare-resource targets that milestones and events use.
func EffectTargetName(target string) string {
	if _, ok := rules.Core().Resource(target); ok {
		return ResourceName(target) + vibe.Line("game.names.EffectTargetName.02", " production")
	}
	return config.EffectTargetLabel(target)
}

// Amount formats one amount of a resource: "23 food", "1.23M iron ore".
func Amount(v float64, res string) string {
	if res == config.ResearchTarget {
		// Work done on a tech is a progress bar, never a number of its own.
		return txtResearchProgress.String()
	}
	return textfmt.Number(v) + " " + ResourceName(res)
}

// Amounts formats a resource map as "23 food, 45 wood", sorted by resource
// key so the order is stable. Zero amounts are skipped. An empty map is
// "nothing".
func Amounts(m map[string]float64) string {
	parts := amountParts(m, func(v float64) bool { return v != 0 })
	if len(parts) == 0 {
		return "nothing"
	}
	return strings.Join(parts, ", ")
}

// DurationText renders a tick count as wall-clock time ("~2m") at the given
// tick interval. Logs and refusals use it instead of raw tick counts.
func DurationText(ticks int, interval time.Duration) string {
	if interval <= 0 {
		interval = BaseTickInterval
	}
	return textfmt.Ticks(ticks, interval)
}

// durationLocked is DurationText at the engine's current speed. Caller holds
// ge.mu (read or write).
func (ge *GameEngine) durationLocked(ticks int) string {
	return DurationText(ticks, ge.tickIntervalLocked())
}
