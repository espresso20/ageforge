package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The Gate Covenant (design-and-architecture/economy.md, Law 1 applied to age
// gates). Every advance must pass these against the most storage buildable in
// the age the player advances from, with no build_cost discounts assumed:
//
//   - a resource requirement fits with GateResourceMargin to spare;
//   - each required building can be built in that age (the age lock forbids
//     building an older age's buildings, and they may have been upgraded away);
//   - the last required copy of it costs at most 1/GateBuildingMargin of the
//     storage for every resource it costs;
//   - every resource the gate asks for, directly or in a required building's
//     price, has a source by then that does not need that resource first: a
//     building that doesn't cost it, hand gathering, an exchange, or techs
//     whose flat output alone covers the whole amount within GateTrickleHours.
const (
	GateBuildingMargin = 2.0
	GateResourceMargin = 1.25
	GateTrickleHours   = 48.0
)

// GateProblem is an age requirement that breaks the Gate Covenant, found from
// config alone. A run only ever meets the first one, so the report lists them
// all up front.
type GateProblem struct {
	From string `json:"from"`
	To   string `json:"to"`
	// Kind is "resource", "building", "unbuildable", "unsourced" (a gate) or
	// "dead_building" (a building whose price has no source in its own age;
	// From is that age, To is empty).
	Kind     string `json:"kind"`
	Key      string `json:"key"` // resource or building key
	Count    int    `json:"count,omitempty"`
	Resource string `json:"resource,omitempty"`
	// Need is the requirement amount, or the last required copy's undiscounted
	// price in Resource.
	Need       float64 `json:"need,omitempty"`
	MaxStorage float64 `json:"max_storage,omitempty"`
	Margin     float64 `json:"margin,omitempty"`   // the margin the rule asks for
	BuiltIn    string  `json:"built_in,omitempty"` // unbuildable: the only age the building can be built in
}

// GateSlack is the tightest requirement of one advance: MaxStorage / Need
// for whichever resource requirement or last required copy comes closest to
// the cap.
type GateSlack struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	Key        string  `json:"key"`
	Resource   string  `json:"resource"`
	Need       float64 `json:"need"`
	MaxStorage float64 `json:"max_storage"`
}

// Ratio is MaxStorage / Need.
func (g GateSlack) Ratio() float64 { return g.MaxStorage / g.Need }

// writeGates renders the static gate check.
func (s *Summary) writeGates(sb *strings.Builder) {
	sb.WriteString("\n## Static gate check\n\n")
	fmt.Fprintf(sb, "From config alone, against the most storage buildable in the age you advance from, with no build_cost discounts: every required building must be buildable in that age, its last required copy must cost at most 1/%g of the storage, and every resource requirement must fit with %gx to spare, and every resource the gate needs must have a source that doesn't cost it first (the Gate Covenant, economy.md). `go test ./smoke` fails on any row here; the runtime invariants are what fail a run.\n\n", GateBuildingMargin, GateResourceMargin)
	if len(s.Gates) == 0 {
		sb.WriteString("No problems.\n")
	} else {
		sb.WriteString("| advance | requirement | needs | max storage | rule |\n|---|---|---|---|---|\n")
		for _, g := range s.Gates {
			switch g.Kind {
			case "unbuildable":
				fmt.Fprintf(sb, "| %s → %s | %d %s | built only in %s | - | must be buildable in %s |\n", g.From, g.To, g.Count, g.Key, g.BuiltIn, g.From)
			case "unsourced":
				fmt.Fprintf(sb, "| %s → %s | %s (for %s) | no source | - | needs a source by %s |\n", g.From, g.To, g.Resource, g.Key, g.From)
			case "dead_building":
				fmt.Fprintf(sb, "| building in %s | %s costs %s | no source | - | price must be obtainable in its age |\n", g.From, g.Key, g.Resource)
			case "building":
				fmt.Fprintf(sb, "| %s → %s | %d %s (copy #%d price) | %s %s | %s | %gx |\n", g.From, g.To, g.Count, g.Key, g.Count, num(g.Need), g.Resource, num(g.MaxStorage), g.Margin)
			default:
				fmt.Fprintf(sb, "| %s → %s | %s | %s %s | %s | %gx |\n", g.From, g.To, g.Key, num(g.Need), g.Resource, num(g.MaxStorage), g.Margin)
			}
		}
	}
	if len(s.Slack) == 0 {
		return
	}
	sb.WriteString("\nTightest requirement per advance (max storage / need):\n\n| advance | requirement | needs | max storage | ratio |\n|---|---|---|---|---|\n")
	for _, g := range s.Slack {
		fmt.Fprintf(sb, "| %s → %s | %s | %s %s | %s | %.2fx |\n", g.From, g.To, g.Key, num(g.Need), g.Resource, num(g.MaxStorage), g.Ratio())
	}
}

// lastCopyPrice is the undiscounted price of copy #n of d in res.
func lastCopyPrice(d config.BuildingDef, res string, n int) float64 {
	return d.BaseCost[res] * math.Pow(d.CostScale, float64(n-1))
}

// StaticGates checks every advance against the Gate Covenant and returns
// what breaks it, plus the tightest requirement of each advance.
func StaticGates() ([]GateProblem, []GateSlack) {
	return staticGates(config.Ages(), config.BuildingByKey())
}

// handGatherable mirrors the gather command: food, wood and stone, by hand,
// up to and including the Medieval Age.
var handGatherable = map[string]bool{"food": true, "wood": true, "stone": true}

// sourced reports whether need of res can be obtained by the end of
// ages[idx] without already holding some of it: a building whose price does
// not include res, hand gathering, an exchange into res, or techs whose flat
// output covers need within GateTrickleHours at 1x. A producer that costs its
// own output (the Bronze Age smithy and its iron, the Renaissance mill and
// its steel) only counts once something else has supplied the first batch.
func sourced(res string, need float64, idx int, ages []config.AgeDef, defs map[string]config.BuildingDef) bool {
	order := map[string]int{}
	unlocked := false
	for i, a := range ages {
		order[a.Key] = i
		if i <= idx {
			for _, r := range a.UnlockResources {
				unlocked = unlocked || r == res
			}
		}
	}
	if !unlocked {
		return false
	}
	if handGatherable[res] && idx <= order["medieval_age"] {
		return true
	}
	trickle := 0.0 // flat per-tick output from techs
	for _, t := range config.Technologies() {
		if order[t.Age] > idx {
			continue
		}
		for _, e := range t.Effects {
			if e.Type == "production" && e.Target == res && e.Value > 0 {
				trickle += e.Value
			}
		}
	}
	if trickle*GateTrickleHours*3600/2 >= need { // 2s ticks at 1x
		return true
	}
	for _, d := range defs {
		if d.RequiredAge == "" || order[d.RequiredAge] > idx || d.Category == "wonder" {
			continue
		}
		if _, costsIt := d.BaseCost[res]; costsIt {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == res && e.Value > 0 {
				return true
			}
		}
	}
	for _, x := range config.BaseExchangeRates() {
		if x.To == res && order[x.MinAge] <= idx {
			return true
		}
	}
	return false
}

func staticGates(ages []config.AgeDef, defs map[string]config.BuildingDef) ([]GateProblem, []GateSlack) {
	var out []GateProblem
	var slack []GateSlack
	for i := 0; i+1 < len(ages); i++ {
		from, to := ages[i], ages[i+1]
		// Every resource the gate needs, what first needs it, and how much in
		// all (the requirement plus every required copy, undiscounted).
		needs := map[string]string{}
		amount := map[string]float64{}
		for res, v := range to.ResourceReqs {
			needs[res] = to.Key + " requirement"
			amount[res] += v
		}
		for _, bld := range sortedKeys(to.BuildingReqs) {
			for res := range defs[bld].BaseCost {
				if _, ok := needs[res]; !ok {
					needs[res] = bld
				}
				for c := 1; c <= to.BuildingReqs[bld]; c++ {
					amount[res] += lastCopyPrice(defs[bld], res, c)
				}
			}
		}
		for _, res := range sortedKeys(needs) {
			if !sourced(res, amount[res], i, ages, defs) {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "unsourced", Key: needs[res], Resource: res})
			}
		}
		tight := GateSlack{From: from.Key, To: to.Key, Need: 1, MaxStorage: math.Inf(1)}
		consider := func(key, res string, need, m float64) {
			if m/need < tight.Ratio() {
				tight.Key, tight.Resource, tight.Need, tight.MaxStorage = key, res, need, m
			}
		}
		for _, res := range sortedKeys(to.ResourceReqs) {
			need := to.ResourceReqs[res]
			m := MaxStorage(from.Key, res)
			consider(res, res, need, m)
			if need*GateResourceMargin > m {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "resource", Key: res, Resource: res,
					Need: need, MaxStorage: m, Margin: GateResourceMargin})
			}
		}
		for _, bld := range sortedKeys(to.BuildingReqs) {
			n := to.BuildingReqs[bld]
			d, ok := defs[bld]
			if !ok {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "unbuildable", Key: bld, Count: n, BuiltIn: "nowhere (unknown key)"})
				continue
			}
			if d.RequiredAge != "" && d.RequiredAge != from.Key {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "unbuildable", Key: bld, Count: n, BuiltIn: d.RequiredAge})
				continue
			}
			// Report the resource with the least headroom, so one row per building.
			var worst *GateProblem
			for _, res := range sortedKeys(d.BaseCost) {
				last := lastCopyPrice(d, res, n)
				m := MaxStorage(from.Key, res)
				consider(bld, res, last, m)
				if last*GateBuildingMargin <= m {
					continue
				}
				if worst == nil || m/last < worst.MaxStorage/worst.Need {
					worst = &GateProblem{From: from.Key, To: to.Key, Kind: "building", Key: bld, Count: n,
						Resource: res, Need: last, MaxStorage: m, Margin: GateBuildingMargin}
				}
			}
			if worst != nil {
				out = append(out, *worst)
			}
		}
		if tight.Key != "" {
			slack = append(slack, tight)
		}
	}
	// Buildings nobody can ever afford: a price in a resource that has no
	// source in the building's own age (coal before the Renaissance, crypto
	// before Cyberpunk). Not a gate, but dead content of the same kind.
	idx := map[string]int{}
	for i, a := range ages {
		idx[a.Key] = i
	}
	for _, k := range sortedKeys(defs) {
		d := defs[k]
		if d.RequiredAge == "" || d.Category == "wonder" {
			continue
		}
		for _, res := range sortedKeys(d.BaseCost) {
			if !sourced(res, d.BaseCost[res], idx[d.RequiredAge], ages, defs) {
				out = append(out, GateProblem{From: d.RequiredAge, Kind: "dead_building", Key: k, Resource: res})
			}
		}
	}
	return out, slack
}
