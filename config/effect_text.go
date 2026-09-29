package config

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Player-facing wording for effects.
//
// Descriptions that quote numbers go stale the moment a balance pass touches
// the data behind them (the Payback Rule rewrites most production rates on
// every call to BaseBuildings), so the mechanical half of a building
// description is built here from the building's runtime Effects. The lineage
// files hold only the flavor sentence. Event, epoch-event and awakening text is
// still written by hand, and the guard tests in effect_text_test.go check it
// against the same helpers.

// WonderSpeedCapStep is how far each built wonder raises the cap on the
// `speed` setting. game.MaxSpeedForAge applies it; a game test pins the two
// together.
const WonderSpeedCapStep = 0.5

// ResourceLabel is a resource's display name in running text: "iron ore",
// "dark matter". Unknown keys fall back to the key with spaces.
func ResourceLabel(key string) string {
	if r, ok := resourceNames()[key]; ok {
		return strings.ToLower(r)
	}
	return strings.ReplaceAll(key, "_", " ")
}

func resourceNames() map[string]string {
	m := make(map[string]string)
	for _, r := range BaseResources() {
		m[r.Key] = r.Name
	}
	return m
}

// EffectTargetLabel names the target of a "bonus"-style effect in running
// text, following the glossary: production_all is "all production",
// gather_rate is "worker output", <res>_rate is "<res> production".
func EffectTargetLabel(target string) string {
	switch target {
	case "production_all", "all":
		return "all production"
	case "gather_rate":
		return "worker output"
	case "research_speed":
		return "research speed"
	case "expedition_reward":
		return "expedition rewards"
	case "military_power":
		return "military power"
	case "build_cost":
		return "building costs"
	case "tick_speed":
		return "game speed"
	case "trade_route_income":
		return "trade route income"
	case "population":
		return "housing"
	}
	if res, ok := strings.CutSuffix(target, "_rate"); ok {
		return ResourceLabel(res) + " production"
	}
	return strings.ReplaceAll(target, "_", " ")
}

// FormatAmount prints an amount or rate the way descriptions show it:
// 3 significant figures and a K/M/B/T/Q suffix from a thousand up, no sign.
func FormatAmount(v float64) string {
	return FormatRateValue(math.Abs(v))
}

// FormatPercent prints a fraction as a whole-ish percentage: 0.3 -> "30%".
func FormatPercent(frac float64) string {
	return FormatRateValue(math.Abs(frac)*100) + "%"
}

// signed prefixes "+" or "-" to an already formatted magnitude.
func signed(v float64, s string) string {
	if v < 0 {
		return "-" + s
	}
	return "+" + s
}

// DurationText renders a tick count as wall-clock time at 1x speed, in the
// same shape the UI's formatTicks uses: "~30s", "~4m 48s", "~1h 12m".
func DurationText(ticks int) string {
	return textfmt.Ticks(ticks, time.Duration(TickSeconds*float64(time.Second)))
}

// RateText is a per-tick change to one resource: "gold +5/tick".
func RateText(res string, v float64) string {
	return ResourceLabel(res) + " " + signed(v, FormatAmount(v)) + "/tick"
}

// buildingEffectParts lists the player-visible effects of a building, one
// phrase each, in Effects order. Effects the engine never reads
// (capacity:military) and per-tick morale nudges are left out.
func buildingEffectParts(d BuildingDef) []string {
	var parts []string
	for _, e := range d.Effects {
		switch e.Type {
		case "production":
			if e.Value != 0 {
				parts = append(parts, signed(e.Value, FormatAmount(e.Value))+" "+ResourceLabel(e.Target)+"/tick")
			}
		case "capacity":
			if e.Target == "population" {
				parts = append(parts, "+"+FormatAmount(e.Value)+" housing")
			}
		case "storage":
			switch e.Target {
			case "all":
				parts = append(parts, "+"+FormatAmount(e.Value)+" storage for every resource")
			case "soldiers":
				parts = append(parts, "+"+FormatAmount(e.Value)+" soldier storage")
			default:
				parts = append(parts, "+"+FormatAmount(e.Value)+" "+ResourceLabel(e.Target)+" storage")
			}
		case "bonus", "trade_route_income":
			parts = append(parts, signed(e.Value, FormatPercent(e.Value))+" "+EffectTargetLabel(e.Target))
		case "opinion":
			parts = append(parts, "+"+FormatAmount(e.Value)+" opinion/tick per worker with every civilization that is not hostile")
		}
	}
	return parts
}

// buildingEffectText is the mechanical sentence appended to a building's
// flavor: "+1 food/tick (3 workers)." Wonders add the speed-cap line. It is
// empty for a building with nothing to report (the Geographic Society, whose
// work is done by the engine, not by an Effect).
func buildingEffectText(d BuildingDef) string {
	parts := buildingEffectParts(d)
	var sb strings.Builder
	if len(parts) > 0 {
		sb.WriteString(strings.Join(parts, ", "))
		if d.WorkerCapacity > 0 && d.WorkerDomain != "" {
			fmt.Fprintf(&sb, " (%d workers)", d.WorkerCapacity)
		}
		sb.WriteString(".")
	}
	if d.Category == "wonder" {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		fmt.Fprintf(&sb, "Raises the speed cap by %sx.", FormatAmount(WonderSpeedCapStep))
	}
	return sb.String()
}

// appendEffectText finishes every building's Description: the flavor
// sentence from the lineage file, then the mechanical sentence built from the
// runtime Effects. It runs last in BaseBuildings, after the Payback Rule has
// set the rates, so the numbers a player reads are the numbers the engine
// uses.
func appendEffectText(defs []BuildingDef) []BuildingDef {
	for i := range defs {
		d := &defs[i]
		txt := buildingEffectText(*d)
		switch {
		case txt == "":
		case d.Description == "":
			d.Description = txt
		default:
			d.Description = d.Description + " " + txt
		}
	}
	return defs
}
