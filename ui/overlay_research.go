package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// formatTechEffect names a config.Effect in glossary words: "+10% all
// production", "+0.5 food/tick", "+500 storage for every resource",
// "+10 housing", "unlocks Lumber Mill".
func formatTechEffect(eff config.Effect) string {
	switch eff.Type {
	case "bonus":
		return textfmt.SignedPercent(eff.Value) + " " + game.EffectTargetName(eff.Target)
	case "production":
		return rateNumber(eff.Value) + " " + game.ResourceName(eff.Target) + "/tick"
	case "storage":
		if eff.Target == "all" {
			return textfmt.Signed(eff.Value) + " storage for every resource"
		}
		return textfmt.Signed(eff.Value) + " " + game.ResourceName(eff.Target) + " storage"
	case "unlock":
		return "unlocks " + game.BuildingName(eff.Target)
	case "capacity":
		switch eff.Target {
		case "population":
			return textfmt.Signed(eff.Value) + " housing"
		case "military":
			return textfmt.Signed(eff.Value) + " soldier storage"
		}
		return textfmt.Signed(eff.Value) + " " + game.EffectTargetName(eff.Target)
	default:
		return game.EffectTargetName(eff.Target) + " " + textfmt.Signed(eff.Value)
	}
}

// rateNumber prints a signed per-tick amount with the decimals a small rate
// needs ("+0.05", "+0.5", "+2.5", "-3"): textfmt.RateValue without its
// trailing zeros. textfmt.Signed would round 0.05 up to "0.1".
func rateNumber(v float64) string {
	s := textfmt.RateValue(v)
	if strings.Contains(s, ".") && !strings.ContainsAny(s, "KMBTQ") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}

// researchSummaryLines is the Stats panel's Research section: what research
// speed, the techs' own cut of research time and Ancient Knowledge do to
// the times the tech tree shows, then what the researched techs come to
// together.
func researchSummaryLines(state game.GameState) string {
	var sb strings.Builder
	if p, ok := state.Pools["research_speed"]; ok && p.Earned != 0 {
		fmt.Fprintf(&sb, "  [gray]Research speed %s: techs take %s of their base time. The times on the tech tree include it.[-]%s\n",
			textfmt.SignedPercent(p.Earned), textfmt.Percent(math.Max(0, 1-p.Applied)), poolTag(state, "research_speed"))
	}
	if f := state.Research.ResearchTime; f > 0 && f < 1 {
		fmt.Fprintf(&sb, "  [gray]Research techs: research time %s. The times on the tech tree include it.[-]\n", game.ResearchFactorText(f))
	}
	if f := state.SuccumbResearchFactor; f > 0 && f < 1 {
		fmt.Fprintf(&sb, "  [gray]Ancient Knowledge: research time %s. The times on the tech tree include it.[-]\n", game.ResearchFactorText(f))
	}
	sb.WriteString(researchBonusLines(state))
	return sb.String()
}

// researchBonusLines lists what the researched techs come to together: the
// tech layer first (output, storage, housing, the cuts of a price or a
// time, the mechanic numbers), which no cap holds, then the pools techs
// share with the rest of the game (with a note when a limit holds one),
// then the flat output of a first source.
func researchBonusLines(state game.GameState) string {
	rs := state.Research
	var sb strings.Builder
	line := func(value, what string) {
		fmt.Fprintf(&sb, "  [green]%-7s[-] %s\n", value, what)
	}
	if rs.AllOutput != 0 {
		line(textfmt.SignedPercent(rs.AllOutput), "All production")
	}
	for _, res := range sortedKeysOf(rs.Output) {
		if v := rs.Output[res]; v != 0 {
			line(textfmt.SignedPercent(v), textfmt.Capitalize(game.ResourceName(res))+" production")
		}
	}
	if rs.Storage != 0 {
		line(textfmt.SignedPercent(rs.Storage), "Storage")
	}
	if rs.Housing != 0 {
		line(textfmt.SignedPercent(rs.Housing), "Housing")
	}
	for _, cut := range []struct {
		factor float64
		what   string
	}{{rs.BuildCost, "Building costs"}, {rs.BuildTime, "Construction time"}, {rs.ResearchTime, "Research time"}} {
		if cut.factor > 0 && cut.factor != 1 {
			line(textfmt.SignedPercent(cut.factor-1), cut.what)
		}
	}
	set := state.Ruleset()
	for _, key := range sortedKeysOf(rs.Mechanics) {
		def, ok := set.Mechanic(key)
		if !ok {
			continue
		}
		term := rs.Mechanics[key]
		switch {
		case def.Multiplies && term != 1:
			line(textfmt.SignedPercent(term-1), textfmt.Capitalize(def.Name))
		case !def.Multiplies && term != 0 && def.Unit == config.UnitPoints:
			line(textfmt.Signed(term*100), textfmt.Capitalize(def.Name)+", in points")
		case !def.Multiplies && term != 0:
			line(textfmt.Signed(term), textfmt.Capitalize(def.Name))
		}
	}
	for _, key := range sortedKeysOf(rs.Bonuses) {
		if value := rs.Bonuses[key]; value != 0 {
			fmt.Fprintf(&sb, "  [green]%-7s[-] %s%s\n", textfmt.SignedPercent(value), formatBonusName(key), poolTag(state, key))
		}
	}
	var flat []string
	for _, res := range sortedKeysOf(rs.Flat) {
		if v := rs.Flat[res]; v != 0 {
			flat = append(flat, rateNumber(v)+" "+game.ResourceName(res)+"/tick")
		}
	}
	if len(flat) > 0 {
		fmt.Fprintf(&sb, "  [gray]Output:[-]  %s\n", strings.Join(flat, ", "))
	}
	if sb.Len() == 0 {
		return "  [gray]No research bonuses yet.[-]\n"
	}
	return sb.String()
}

// formatBonusName names a research/milestone bonus key in glossary words, as
// a label: "All production", "Worker output", "Iron production", "Housing".
func formatBonusName(key string) string {
	return textfmt.Capitalize(game.EffectTargetName(key))
}

// capitalize upper-cases the first letter of s.
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
