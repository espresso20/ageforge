package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
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
//     price or in the wonder's, can be had in that age from a cold start (see
//     coldStart): by a player who skipped every building no gate required,
//     through that age's own buildings (bought with what is reachable), hand
//     gathering, the market (only once a trade building can stand), techs
//     whose flat output alone covers the whole amount within
//     GateTrickleHours, or the stock carried over. Older ages' producers only
//     pay for first copies and seed market parity;
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
	fmt.Fprintf(sb, "From config alone, against the most storage buildable in the age you advance from, with no build_cost discounts: every required building must be buildable in that age, its last required copy must cost at most 1/%g of the storage, each part of the age's wonder at most 1/%g of it, every resource requirement must fit with %gx to spare, every resource the gate needs (the wonder's included) must be obtainable in that age from a cold start (by a player who skipped every building no gate required), and every flow resource it needs must be made within the age's target at a moderate income or bought for at most %g price units (the Gate Covenant, economy.md). No building may cost a resource with no source in its own age. `go test ./smoke` fails on any row here; the runtime invariants are what fail a run.\n\n", GateBuildingMargin, GateWonderMargin, GateResourceMargin, GateFlowMarketUnits)
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

// coldStart is what a player entering an age is guaranteed to have, and
// what they can reach from there by building and trading in that age alone.
// It is the Gate Covenant's model of sourcing ("from a cold start"): a player
// may skip any building no gate required, so nothing optional from an earlier
// age may be relied on. The Iron Age trading post was the case that forced
// it: it cost gold, it was the Iron Age's only gold producer and its only
// trade building (which the market needs), and the Bronze Age market that
// would have broken the loop is optional.
//
// What carries over into age A, and nothing else:
//
//   - every building an earlier gate required, and every earlier age's
//     wonder (each advance requires them; an upgrade keeps the lineage and
//     its output);
//   - the resources the gate into A required, held at the advance, up to the
//     advance's carryover cap (game.CarryoverStarterBuildings copies of A's
//     cheapest building priced in it; faith is kept whole);
//   - a steady supply of every resource an earlier gate required as a
//     resource requirement, unless it could be hand-gathered then or the
//     stock carried into that age already met it: thousands of iron can't be
//     banked without making it, and whatever made it (a producer, or a trade
//     building and the market) still stands. Prices of required buildings
//     don't imply a supply: a few scriptoriums' gold can come from events.
//
// Carried producers are sized to older prices, so, as in the old rule, they
// only bootstrap A's buildings (pay for a first copy) and seed market parity;
// a gate's total must come from A itself: A's own producers (bought with
// what is reachable), hand gathering (through the Medieval Age), the market
// (only with a trade building, carried or buildable in A), techs whose flat
// output covers the amount within GateTrickleHours, or the carried stock.
type coldStart struct {
	idx      int
	age      string
	unlocked map[string]bool
	levels   map[string]float64 // construction resources of the age
	trickle  map[string]float64 // flat per-tick tech output up to the age
	stock    map[string]float64 // carried resources
	boot     map[string]bool    // steady carried supply: bootstraps and seeds parity only
	reach    map[string]bool    // resources the age itself supplies in bulk
	market   bool               // a trade building stands or can be built
	carried  map[string]bool    // buildings carried over
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

// carryoverStock is what the advance into age leaves of amount of res
// (game.GameEngine.advanceAge): at most CarryoverStarterBuildings copies of
// the age's cheapest building priced in it, a residual share of a resource
// no building of the age costs, and faith whole.
func carryoverStock(res string, amount float64, age string, defs map[string]config.BuildingDef) float64 {
	if res == "faith" {
		return amount
	}
	entry := 0.0
	for _, d := range defs {
		if d.RequiredAge != age || d.Category == "wonder" {
			continue
		}
		if c := d.BaseCost[res]; c > 0 && (entry == 0 || c < entry) {
			entry = c
		}
	}
	if entry > 0 {
		return math.Min(amount, game.CarryoverStarterBuildings*entry)
	}
	return amount * game.CarryoverResidualPct
}

// coldStarts computes the cold start of every age, in order.
func coldStarts(ages []config.AgeDef, defs map[string]config.BuildingDef) []*coldStart {
	medieval := len(ages)
	for i, a := range ages {
		if a.Key == "medieval_age" {
			medieval = i
		}
	}
	out := make([]*coldStart, len(ages))
	unlocked := map[string]bool{}
	carried := map[string]bool{}
	supply := map[string]bool{}
	trickle := map[string]float64{}
	techs := config.Technologies()
	for i, a := range ages {
		for _, r := range a.UnlockResources {
			unlocked[r] = true
		}
		stock := map[string]float64{}
		if i > 0 {
			for bld := range a.BuildingReqs {
				carried[bld] = true
			}
			if w := ageWonder(ages[i-1], defs); w != "" {
				carried[w] = true
			}
			for _, res := range sortedKeys(a.ResourceReqs) {
				v := a.ResourceReqs[res]
				stock[res] = carryoverStock(res, v, a.Key, defs)
				if handGatherable[res] && i-1 <= medieval {
					continue // could have been gathered by hand
				}
				if out[i-1].stock[res] >= v {
					continue // the stock carried into that age met it
				}
				supply[res] = true
			}
		}
		for _, t := range techs {
			if t.Age != a.Key {
				continue
			}
			for _, e := range t.Effects {
				if e.Type == "production" && e.Value > 0 {
					trickle[e.Target] += e.Value
				}
			}
		}
		cs := &coldStart{idx: i, age: a.Key, unlocked: cloneSet(unlocked), levels: config.PriceLevels(a.Key),
			trickle: cloneMap(trickle), stock: stock, boot: map[string]bool{}, reach: map[string]bool{},
			carried: cloneSet(carried)}
		for r, ok := range supply {
			if ok && unlocked[r] {
				cs.boot[r] = true
			}
		}
		for _, k := range sortedKeys(carried) {
			d := defs[k]
			if d.LineageKey == "trade" {
				cs.market = true
			}
			for _, e := range d.Effects {
				if e.Type == "production" && e.Value > 0 && unlocked[e.Target] {
					cs.boot[e.Target] = true
				}
			}
		}
		if i <= medieval {
			for r := range handGatherable {
				if unlocked[r] {
					cs.reach[r] = true
				}
			}
		}
		cs.solve(defs)
		out[i] = cs
	}
	return out
}

func cloneSet(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		if v {
			out[k] = true
		}
	}
	return out
}

func cloneMap(m map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// affords reports whether amount of res can be paid once in the age: the age
// supplies it, a carried supply does, or techs or the carried stock cover it.
func (cs *coldStart) affords(res string, amount float64) bool {
	if !cs.unlocked[res] {
		return false
	}
	return cs.reach[res] || cs.boot[res] || cs.covers(res, amount)
}

// covers reports whether techs within GateTrickleHours, or the carried
// stock, cover amount of res without any producer.
func (cs *coldStart) covers(res string, amount float64) bool {
	return cs.trickle[res]*GateTrickleHours*3600/config.TickSeconds >= amount || cs.stock[res] >= amount
}

// sources reports whether the age supplies the whole amount of res a gate
// asks for.
func (cs *coldStart) sources(res string, amount float64) bool {
	return cs.unlocked[res] && (cs.reach[res] || cs.covers(res, amount))
}

// solve grows reach to a fixed point: buy every building of the age whose
// first copy is affordable, open the market once a trade building stands or
// can be bought, and trade into what the market sells.
func (cs *coldStart) solve(defs map[string]config.BuildingDef) {
	keys := sortedKeys(defs)
	for changed := true; changed; {
		changed = false
		add := func(r string) {
			if cs.unlocked[r] && !cs.reach[r] {
				cs.reach[r], changed = true, true
			}
		}
		for _, k := range keys {
			d := defs[k]
			if d.RequiredAge != cs.age || d.Category == "wonder" {
				continue
			}
			ok := true
			for res, c := range d.BaseCost {
				if c > 0 && !cs.affords(res, c) {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			if d.LineageKey == "trade" && !cs.market {
				cs.market, changed = true, true
			}
			for _, e := range d.Effects {
				if e.Type == "production" && e.Value > 0 {
					add(e.Target)
				}
			}
		}
		if !cs.market {
			continue
		}
		// A construction resource of the age trades at parity with the
		// others, so any supply of one (carried supplies included) is a way
		// into all of them.
		parity := false
		for r := range cs.levels {
			parity = parity || cs.reach[r] || cs.boot[r]
		}
		if parity {
			for _, r := range sortedKeys(cs.levels) {
				add(r)
			}
		}
		for _, x := range config.BaseExchangeRates() {
			if (cs.reach[x.From] || cs.boot[x.From]) && minAgeReached(x.MinAge, cs.idx) {
				add(x.To)
			}
		}
	}
}

// minAgeReached reports whether age comes at or before position idx.
func minAgeReached(age string, idx int) bool {
	for i, k := range config.AgeOrder() {
		if k == age {
			return i <= idx
		}
	}
	return false
}

// flowMarketUnits is what short of the flow resource res costs at the
// market in cs's age, in price units of that age: the cheapest listed
// exchange into res from one of the age's construction resources. -1 if the
// market is out of reach or sells res for none of them.
func flowMarketUnits(res string, short float64, cs *coldStart) float64 {
	if !cs.market {
		return -1
	}
	best := -1.0
	for _, x := range config.BaseExchangeRates() {
		if x.To != res || !minAgeReached(x.MinAge, cs.idx) || cs.levels[x.From] <= 0 {
			continue
		}
		rate := config.ExchangeRate(x, cs.age)
		if rate <= 0 {
			continue
		}
		if u := short / rate / cs.levels[x.From]; best < 0 || u < best {
			best = u
		}
	}
	return best
}

func staticGates(ages []config.AgeDef, defs map[string]config.BuildingDef) ([]GateProblem, []GateSlack) {
	var out []GateProblem
	cold := coldStarts(ages, defs)
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
			if !cold[i].sources(res, amount[res]) {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "unsourced", Key: needs[res], Resource: res})
				continue
			}
			if !config.IsFlowResource(res) || amount[res] <= 0 {
				continue
			}
			made := config.FlowIncome(res, from.Key) * config.AgeTargetTicks(from.Key)
			if short := amount[res] - made; short > 0 {
				if u := flowMarketUnits(res, short, cold[i]); u < 0 || u > GateFlowMarketUnits {
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
			if !cold[idx[d.RequiredAge]].affords(res, d.BaseCost[res]) {
				out = append(out, GateProblem{From: d.RequiredAge, Kind: "dead_building", Key: k, Resource: res})
			}
		}
	}
	return out, slack
}
