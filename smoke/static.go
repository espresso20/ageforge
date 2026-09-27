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
//   - the age's wonder, which every advance also requires, costs at most
//     1/GateWonderMargin of the storage in each resource (wonders are banked
//     a deposit at a time, so no margin: each part must fit one full store);
//   - every resource the gate asks for, directly, in a required building's
//     price or in the wonder's, has a source in that age that does not need
//     that resource first: a building of that age that doesn't cost it, hand
//     gathering, a market exchange, or techs whose flat output alone covers
//     the whole amount within GateTrickleHours. Older ages' producers don't
//     count: the age lock stops the player building more of them;
//   - every flow resource (food, faith, culture, soldiers) the gate asks for is
//     made within the age's pacing target at config.FlowIncome, or the rest
//     can be bought at the market for at most GateFlowMarketUnits price units
//     of the age.
const (
	GateBuildingMargin  = 2.0
	GateResourceMargin  = 1.25
	GateWonderMargin    = 1.0
	GateTrickleHours    = 48.0
	GateFlowMarketUnits = 10.0
)

// GateProblem is an age requirement that breaks the Gate Covenant, found from
// config alone. A run only ever meets the first one, so the report lists them
// all up front.
type GateProblem struct {
	From string `json:"from"`
	To   string `json:"to"`
	// Kind is "resource", "building", "wonder", "unbuildable", "unsourced",
	// "flow" (a gate) or "dead_building" (a building whose price has no source
	// in its own age; From is that age, To is empty).
	Kind     string `json:"kind"`
	Key      string `json:"key"` // resource, building or wonder key
	Count    int    `json:"count,omitempty"`
	Resource string `json:"resource,omitempty"`
	// Need is the requirement amount, the last required copy's undiscounted
	// price in Resource, the wonder's price in it, or (flow) the whole amount
	// of the flow resource the gate asks for.
	Need       float64 `json:"need,omitempty"`
	MaxStorage float64 `json:"max_storage,omitempty"`
	Margin     float64 `json:"margin,omitempty"`   // the margin the rule asks for
	BuiltIn    string  `json:"built_in,omitempty"` // unbuildable: the only age the building can be built in
	// Flow problems: what FlowIncome makes over the age's target, and what the
	// rest would cost at the market in price units of the age (-1: no route).
	Made        float64 `json:"made,omitempty"`
	MarketUnits float64 `json:"market_units,omitempty"`
}

// GateSlack is the tightest requirement of one advance: MaxStorage / Need
// for whichever resource requirement, last required copy or wonder part comes
// closest to the cap.
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
	fmt.Fprintf(sb, "From config alone, against the most storage buildable in the age you advance from, with no build_cost discounts: every required building must be buildable in that age, its last required copy must cost at most 1/%g of the storage, each part of the age's wonder at most 1/%g of it, every resource requirement must fit with %gx to spare, every resource the gate needs (the wonder's included) must have a source in that age that doesn't cost it first, and every flow resource it needs must be made within the age's target at a moderate income or bought for at most %g price units (the Gate Covenant, economy.md). No building may cost a resource with no source in its own age. `go test ./smoke` fails on any row here; the runtime invariants are what fail a run.\n\n", GateBuildingMargin, GateWonderMargin, GateResourceMargin, GateFlowMarketUnits)
	if len(s.Gates) == 0 {
		sb.WriteString("No problems.\n")
	} else {
		sb.WriteString("| advance | requirement | needs | max storage | rule |\n|---|---|---|---|---|\n")
		for _, g := range s.Gates {
			switch g.Kind {
			case "unbuildable":
				fmt.Fprintf(sb, "| %s → %s | %d %s | built only in %s | - | must be buildable in %s |\n", g.From, g.To, g.Count, g.Key, g.BuiltIn, g.From)
			case "unsourced":
				fmt.Fprintf(sb, "| %s → %s | %s (for %s) | no source | - | needs a source in %s |\n", g.From, g.To, g.Resource, g.Key, g.From)
			case "dead_building":
				fmt.Fprintf(sb, "| building in %s | %s costs %s | no source | - | price must be obtainable in its age |\n", g.From, g.Key, g.Resource)
			case "building":
				fmt.Fprintf(sb, "| %s → %s | %d %s (copy #%d price) | %s %s | %s | %gx |\n", g.From, g.To, g.Count, g.Key, g.Count, num(g.Need), g.Resource, num(g.MaxStorage), g.Margin)
			case "wonder":
				fmt.Fprintf(sb, "| %s → %s | wonder %s | %s %s | %s | %gx |\n", g.From, g.To, g.Key, num(g.Need), g.Resource, num(g.MaxStorage), g.Margin)
			case "flow":
				market := "no market route"
				if g.MarketUnits >= 0 {
					market = fmt.Sprintf("the rest costs %.1f price units", g.MarketUnits)
				}
				fmt.Fprintf(sb, "| %s → %s | %s (for %s) | %s %s | - | makes %s in the target; %s (max %g) |\n", g.From, g.To, g.Resource, g.Key, num(g.Need), g.Resource, num(g.Made), market, GateFlowMarketUnits)
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

// ageWonder is the wonder an age unlocks, which the advance out of it
// requires (game.ProgressManager picks it the same way), or "".
func ageWonder(a config.AgeDef, defs map[string]config.BuildingDef) string {
	for _, k := range a.UnlockBuildings {
		if d, ok := defs[k]; ok && d.Category == "wonder" {
			return k
		}
	}
	return ""
}

// sourced reports whether need of res can be obtained in ages[idx] without
// already holding some of it: a building of that age whose price does not
// include res, hand gathering, a market exchange into res (a listed pair
// open by then, or market parity when res is a construction resource of the
// age and something produces one), or techs whose flat output covers need
// within GateTrickleHours at 1x. Older ages' producers of a resource that is
// not a construction resource of the age don't count: the age lock stops the
// player building more, and their output is sized to an older age's prices
// (the Stellar Cradle's 940T uranium against Atomic Age uranium mines). A
// producer that costs its own output (the Bronze Age smithy and its iron, the
// Renaissance mill and its steel) only counts once something else has
// supplied the first batch.
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
	age := ages[idx].Key
	lv := config.PriceLevels(age)
	parity := false // some construction resource of the age is produced
	for _, d := range defs {
		j, ok := order[d.RequiredAge]
		if !ok || j > idx || d.Category == "wonder" {
			continue
		}
		_, costsIt := d.BaseCost[res]
		for _, e := range d.Effects {
			if e.Type != "production" || e.Value <= 0 {
				continue
			}
			if e.Target == res && !costsIt && j == idx {
				return true
			}
			if lv[e.Target] > 0 && (e.Target != res || !costsIt) {
				parity = true
			}
		}
	}
	// A construction resource of the age trades at parity with the others,
	// so any producer of one of them, old copies included, is a way in.
	if parity && lv[res] > 0 {
		return true
	}
	for _, x := range config.BaseExchangeRates() {
		if x.To == res && order[x.MinAge] <= idx {
			return true
		}
	}
	return false
}

// flowMarketUnits is what short of the flow resource res costs at the
// market in ages[idx], in price units of that age: the cheapest listed
// exchange into res from one of the age's construction resources. -1 if the
// market sells res for none of them.
func flowMarketUnits(res string, short float64, idx int, ages []config.AgeDef) float64 {
	order := map[string]int{}
	for i, a := range ages {
		order[a.Key] = i
	}
	age := ages[idx].Key
	lv := config.PriceLevels(age)
	best := -1.0
	for _, x := range config.BaseExchangeRates() {
		if x.To != res || order[x.MinAge] > idx || lv[x.From] <= 0 {
			continue
		}
		rate := config.ExchangeRate(x, age)
		if rate <= 0 {
			continue
		}
		if u := short / rate / lv[x.From]; best < 0 || u < best {
			best = u
		}
	}
	return best
}

func staticGates(ages []config.AgeDef, defs map[string]config.BuildingDef) ([]GateProblem, []GateSlack) {
	var out []GateProblem
	var slack []GateSlack
	for i := 0; i+1 < len(ages); i++ {
		from, to := ages[i], ages[i+1]
		wonder := ageWonder(from, defs)
		// Every resource the gate needs, what first needs it, and how much in
		// all (the requirement, every required copy undiscounted, the wonder).
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
		if wonder != "" {
			for res, c := range defs[wonder].BaseCost {
				if _, ok := needs[res]; !ok {
					needs[res] = wonder
				}
				amount[res] += c
			}
		}
		for _, res := range sortedKeys(needs) {
			if !sourced(res, amount[res], i, ages, defs) {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "unsourced", Key: needs[res], Resource: res})
				continue
			}
			if !config.IsFlowResource(res) || amount[res] <= 0 {
				continue
			}
			made := config.FlowIncome(res, from.Key) * config.AgeTargetTicks(from.Key)
			if short := amount[res] - made; short > 0 {
				if u := flowMarketUnits(res, short, i, ages); u < 0 || u > GateFlowMarketUnits {
					out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "flow", Key: needs[res], Resource: res,
						Need: amount[res], Made: made, MarketUnits: u})
				}
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
		if wonder != "" {
			var worst *GateProblem
			for _, res := range sortedKeys(defs[wonder].BaseCost) {
				c := defs[wonder].BaseCost[res]
				m := MaxStorage(from.Key, res)
				consider(wonder, res, c, m)
				if c*GateWonderMargin <= m {
					continue
				}
				if worst == nil || m/c < worst.MaxStorage/worst.Need {
					worst = &GateProblem{From: from.Key, To: to.Key, Kind: "wonder", Key: wonder,
						Resource: res, Need: c, MaxStorage: m, Margin: GateWonderMargin}
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
	// before Cyberpunk). Not a gate, but dead content of the same kind. The
	// wonders the gates require were checked above; the last age's wonder,
	// which no gate requires, is checked here.
	idx := map[string]int{}
	gated := map[string]bool{}
	for i, a := range ages {
		idx[a.Key] = i
		if w := ageWonder(a, defs); w != "" && i+1 < len(ages) {
			gated[w] = true
		}
	}
	for _, k := range sortedKeys(defs) {
		d := defs[k]
		if d.RequiredAge == "" || gated[k] {
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
