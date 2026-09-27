package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// GateProblem is an age requirement that cannot fit under the most storage
// the player can build before advancing, found from config alone. A run only
// ever meets the first one, so the report lists them all up front.
type GateProblem struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	Kind       string  `json:"kind"` // "resource" or "building"
	Key        string  `json:"key"`  // resource or building key
	Count      int     `json:"count,omitempty"`
	Resource   string  `json:"resource"`
	Need       float64 `json:"need"` // amount, or the last copy's price after the discount
	MaxStorage float64 `json:"max_storage"`
	// Discount is the best build_cost multiplier reachable by From (every
	// build_cost tech and milestone up to that age), already applied to Need.
	Discount float64 `json:"build_cost_multiplier,omitempty"`
}

// writeGates renders the static gate check.
func (s *Summary) writeGates(sb *strings.Builder) {
	sb.WriteString("\n## Static gate check\n\n")
	sb.WriteString("From config alone: every advance whose resource requirement, or whose last required building's price (after the best build_cost discount reachable by then), is larger than the most storage buildable in the age you advance from. A run stops at the first of these, so they are all listed here. Report only; the runtime invariants are what fail the run.\n\n")
	if len(s.Gates) == 0 {
		sb.WriteString("None.\n")
		return
	}
	sb.WriteString("| advance | requirement | needs | max storage | discount |\n|---|---|---|---|---|\n")
	for _, g := range s.Gates {
		req := g.Key
		if g.Kind == "building" {
			req = fmt.Sprintf("%d %s (copy #%d price)", g.Count, g.Key, g.Count)
		}
		disc := "-"
		if g.Kind == "building" {
			disc = fmt.Sprintf("x%.2f", g.Discount)
		}
		fmt.Fprintf(sb, "| %s → %s | %s | %s %s | %s | %s |\n", g.From, g.To, req, num(g.Need), g.Resource, num(g.MaxStorage), disc)
	}
}

// bestBuildCostMult is the lowest build cost multiplier reachable in ageIdx:
// every build_cost tech and milestone available by then, floored like the
// engine (buildCostFloor = 0.10).
func bestBuildCostMult(ageIdx int, order map[string]int) float64 {
	sum := 0.0
	for _, t := range config.Technologies() {
		if order[t.Age] > ageIdx {
			continue
		}
		for _, e := range t.Effects {
			if e.Target == "build_cost" {
				sum += e.Value
			}
		}
	}
	for _, m := range config.Milestones() {
		if m.MinAge != "" && order[m.MinAge] > ageIdx {
			continue
		}
		for _, e := range m.Rewards {
			if e.Target == "build_cost" {
				sum += e.Value
			}
		}
	}
	return math.Min(1, math.Max(0.10, 1+sum))
}

// StaticGates checks every advance: each resource requirement, and the price
// of the last copy of each required building, against MaxStorage in the age
// the player advances from.
func StaticGates() []GateProblem {
	order := map[string]int{}
	for i, k := range config.AgeOrder() {
		order[k] = i
	}
	defs := config.BuildingByKey()
	ages := config.Ages()
	var out []GateProblem
	for i := 0; i+1 < len(ages); i++ {
		from, to := ages[i], ages[i+1]
		for _, res := range sortedKeys(to.ResourceReqs) {
			need := to.ResourceReqs[res]
			if m := MaxStorage(from.Key, res); need > m {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "resource", Key: res, Resource: res, Need: need, MaxStorage: m})
			}
		}
		disc := bestBuildCostMult(i, order)
		for _, bld := range sortedKeys(to.BuildingReqs) {
			n := to.BuildingReqs[bld]
			d := defs[bld]
			for _, res := range sortedKeys(d.BaseCost) {
				last := d.BaseCost[res] * math.Pow(d.CostScale, float64(n-1)) * disc
				if m := MaxStorage(from.Key, res); last > m {
					out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "building", Key: bld, Count: n,
						Resource: res, Need: last, MaxStorage: m, Discount: disc})
					break // one resource is enough to block it
				}
			}
		}
	}
	return out
}
