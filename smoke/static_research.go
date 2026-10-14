package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The Research Covenant: what the tech tree's lock may ask of an age, from
// config alone. Each age's wonder needs its keystone tech, and the advance
// needs the wonder, so a keystone is a gate like any other and gets the same
// kind of check. Knowledge is not held or spent: a tech is work, and the
// scholars of the knowledge buildings do it, so there is no price to afford
// and no store to fit it in.
//
//   - Time: the keystone is done well inside the age's target. The work it
//     takes, with the techs of its own age it stands on, as a share of what
//     the reference town's scholars do in the age's target time
//     (referenceKnowledge), plus the time to build the wonder as a share of
//     the target, come to at most 1.
//
// The Time rule counts a run that kept up: one that arrives with the spine
// techs of earlier ages researched, as the smoke bot does (it takes the
// cheapest tech once the keystone is done, and the spine is the cheap part).
// A run that researched nothing an earlier wonder did not ask for owes those
// too, and the row reports them (Carried): with today's tree that is more
// than the Interstellar and Galactic Ages make in their target time, where
// every tech is on the spine. It is reported, not failed: the ages still
// finish, late.
//
// (That the keystone can be reached in its age at all is the Gate
// Covenant's check.)
// ResearchRow is one age against the Research Covenant.
type ResearchRow struct {
	Age string `json:"age"`
	// First is the age's cheapest tech ("" for an age with none).
	First     string  `json:"first_tech,omitempty"`
	FirstCost float64 `json:"first_cost,omitempty"`
	// Keystone is the tech the age's wonder needs ("" for none). Chain is
	// the keystone and the techs of its own age it stands on, what it stands
	// on first: what a run that kept up still has to research for the
	// wonder. Carried is the earlier ages' techs a run that researched
	// nothing an earlier wonder did not ask for owes on top.
	Keystone string   `json:"keystone,omitempty"`
	Wonder   string   `json:"wonder,omitempty"`
	Chain    []string `json:"chain,omitempty"`
	Carried  []string `json:"carried,omitempty"`
	// ChainCost is the chain's work and ChainMax the priciest tech of the
	// chain and of what is carried.
	ChainCost float64 `json:"chain_cost,omitempty"`
	ChainMax  float64 `json:"chain_max,omitempty"`
	// The Time rule's two parts, as shares of the age: the chain's work over
	// what the scholars do in the target, and the wonder's build time over
	// the target. CarriedShare is what is carried, over what the scholars do
	// (reported, not part of the rule).
	KnowledgeShare float64 `json:"knowledge_share,omitempty"`
	BuildShare     float64 `json:"build_share,omitempty"`
	CarriedShare   float64 `json:"carried_share,omitempty"`
}

// TimeShare is the share of the age's target the Time rule counts: the
// research and the wonder's construction.
func (r ResearchRow) TimeShare() float64 {
	return r.KnowledgeShare + r.BuildShare
}

// Problems lists what the row breaks of the covenant, one line each (nil
// when it keeps it).
func (r ResearchRow) Problems() []string {
	var out []string
	if r.Keystone == "" {
		return out
	}
	known, isKnown := ResearchTimeKnown[r.Age]
	if isKnown && r.TimeShare() <= 1 {
		out = append(out, fmt.Sprintf("%s: the keystone, its research and the wonder now fit the age (%.0f%% of it); take %s out of ResearchTimeKnown (it was listed at %.0f%%)",
			r.Age, 100*r.TimeShare(), r.Age, 100*known))
	}
	if r.TimeShare() > 1 && !isKnown {
		out = append(out, fmt.Sprintf("%s: %s waits for %s (%s of work, %.0f%% of what the scholars do in the age's target time); with %.0f%% of the target to build the wonder that is %.0f%% of the age, over 100%%",
			r.Age, r.Wonder, strings.Join(r.Chain, ", "), num(r.ChainCost), 100*r.KnowledgeShare, 100*r.BuildShare, 100*r.TimeShare()))
	}
	return out
}

// ResearchTimeKnown lists the ages where, for the reference town, the
// keystone, its research and the wonder take more than the age's target, with
// the share they take today. The Time rule used to count knowledge at a rate
// read off bot runs; it counts what eight staffed copies of the knowledge
// buildings make now. One age did not fit with five copies (the Electric Age,
// at 102%); the town of eight makes 1.6 times the knowledge its buildings
// give and the age takes 72%, so none is listed. An age listed here that
// starts to fit, or an age not listed that stops fitting, fails the check, so
// this table only ever gets shorter on purpose.
var ResearchTimeKnown = map[string]float64{}

// StaticResearch checks every age against the Research Covenant and returns
// one row per age, failing or not.
func StaticResearch() []ResearchRow {
	return staticResearch(config.Ages(), config.BuildingByKey(), config.Technologies(), referenceKnowledge)
}

// referenceKnowledge is what the reference town makes of knowledge in an
// age's target time: eight fully staffed copies of every knowledge building
// so far, with the bonuses the tables give that town, for as many ticks as
// the age is meant to last. It is worked out from the building and tech
// tables alone.
func referenceKnowledge(age string) float64 {
	return float64(config.TypicalIncome(config.ResearchTarget, age) * config.LatePriceTargetTicks(age))
}

// staticResearch is StaticResearch over the given tables; knowledge is what
// an age makes in its target time.
func staticResearch(ages []config.AgeDef, defs map[string]config.BuildingDef, techs []config.TechDef, knowledge func(age string) float64) []ResearchRow {
	byKey := make(map[string]config.TechDef, len(techs))
	for _, t := range techs {
		byKey[t.Key] = t
	}
	have := map[string]bool{} // what a keystones-only run has researched so far
	var out []ResearchRow
	for _, a := range ages {
		row := ResearchRow{Age: a.Key}
		for _, t := range techs {
			if t.Age == a.Key && (row.First == "" || t.Cost < row.FirstCost) {
				row.First, row.FirstCost = t.Key, t.Cost
			}
		}

		wonder := ageWonder(a, defs)
		if key := defs[wonder].RequiredTech; key != "" {
			row.Keystone, row.Wonder = key, wonder
			carried := 0.0
			for _, k := range configChain(byKey, key, have, nil) {
				t := byKey[k]
				row.ChainMax = math.Max(row.ChainMax, t.Cost)
				if t.Age != a.Key {
					row.Carried = append(row.Carried, k)
					carried += t.Cost
					continue
				}
				row.Chain = append(row.Chain, k)
				row.ChainCost += t.Cost
			}
			if made := knowledge(a.Key); made > 0 {
				row.KnowledgeShare = row.ChainCost / made
				row.CarriedShare = carried / made
			} else {
				row.KnowledgeShare = math.Inf(1)
			}
			if target := config.LatePriceTargetTicks(a.Key); target > 0 {
				row.BuildShare = float64(defs[wonder].BuildTicks) / target
			}
		}
		out = append(out, row)
	}
	return out
}

// configChain appends key and every tech it still needs to out, what it
// needs first, and marks them in have: every prerequisite not in have, all
// the way down, and for an either-or group have does not satisfy, the
// branch that costs the least to finish.
func configChain(techs map[string]config.TechDef, key string, have map[string]bool, out []string) []string {
	t, ok := techs[key]
	if !ok || have[key] {
		return out
	}
	have[key] = true
	for _, pre := range t.Prerequisites {
		out = configChain(techs, pre, have, out)
	}
	if len(t.AnyOf) > 0 && !t.AnyOfMet(func(k string) bool { return have[k] }) {
		best, bestCost := "", math.Inf(1)
		for _, alt := range t.AnyOf {
			trial := make(map[string]bool, len(have))
			for k := range have {
				trial[k] = true
			}
			cost := 0.0
			for _, k := range configChain(techs, alt, trial, nil) {
				cost += techs[k].Cost
			}
			if _, known := techs[alt]; known && cost < bestCost {
				best, bestCost = alt, cost
			}
		}
		if best != "" {
			out = configChain(techs, best, have, out)
		}
	}
	return append(out, key)
}

// writeResearch renders the Research Covenant check.
func writeResearch(sb *strings.Builder, rows []ResearchRow) {
	fmt.Fprintf(sb, "Each age's wonder needs its keystone tech (the Research Covenant). A tech is work the scholars do, so there is no price to afford and no store to fit it in. The keystone must be done well inside the age: its work, with the techs of its own age it stands on, as a share of what the reference town's scholars (the copies the gates ask for of every knowledge building so far, all researching) do in the age's target time, plus the wonder's build time as a share of the target, may not pass 100%%. \"Carried in\" is what a run that researched nothing an earlier wonder did not ask for owes on top, as a share of what the scholars do: reported, not failed. `go test ./smoke` fails on any row marked ✗.\n\n")
	sb.WriteString("| age | first tech | work | the wonder waits for | work | of the age's research | + build | of the age | carried in | |\n|---|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		mark := "✓"
		if len(r.Problems()) > 0 {
			mark = "✗"
		}
		first, firstCost := "-", "-"
		if r.First != "" {
			first, firstCost = r.First, num(r.FirstCost)
		}
		if r.Keystone == "" {
			fmt.Fprintf(sb, "| %s | %s | %s | - | - | - | - | - | - | %s |\n", r.Age, first, firstCost, mark)
			continue
		}
		carried := "-"
		if len(r.Carried) > 0 {
			carried = fmt.Sprintf("%s (+%.0f%%)", strings.Join(r.Carried, ", "), 100*r.CarriedShare)
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %.0f%% | %.0f%% | %.0f%% | %s | %s |\n", r.Age, first, firstCost,
			strings.Join(r.Chain, ", "), num(r.ChainCost), 100*r.KnowledgeShare, 100*r.BuildShare, 100*r.TimeShare(), carried, mark)
	}
}
