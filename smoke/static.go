package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/rules"
)

// The Gate Covenant (the economy design's Law 1 applied to age gates). Every
// advance must pass these against the most storage buildable in the age the
// player advances from, with no build_cost discounts assumed:
//
//   - each required building can be built in that age (the age lock forbids
//     building an older age's buildings, and they may have been upgraded away),
//     and the tech it waits for, if any, can be finished in that age;
//   - the last required copy of it costs at most 1/GateBuildingMargin of the
//     store the gates force (gateStore) for every resource it costs: the
//     storage rule's own room (config.StorageRuleRoom), against the store it
//     sizes for exactly this;
//   - the age's wonder, which every advance also requires, is the one price
//     allowed over a store: it is paid into its bank in rounds (deposits,
//     overflow), so it is held to no storage margin (it was held to one full
//     store while a store was 25 copies; the Dyson Scaffold's steel is two
//     and a half of the Space Age's walled stores). Its parts are checked
//     for a source, below;
//   - every resource the gate asks for, directly, in a required building's
//     price, in the wonder's or in the price of the wonder's keystone tech
//     (knowledge), can be had in that age from a cold start (see
//     coldStart): by a player who skipped every building no gate required,
//     through that age's own buildings (bought with what is reachable), the
//     buildings the earlier gates required, hand gathering, techs whose flat
//     output alone covers the whole amount within GateTrickleHours. A price
//     paid once (the wonder, its keystone tech, the first copy of a building)
//     is held to a bound on a carried source: the copies the gates required,
//     at their listed output, must make that price in one target of the age
//     (CarriedSupplyTargets);
//   - a town arrives in the next age with empty stores and the buildings the
//     gates so far asked for, so the first copy of every building of that age
//     a run can build, and its first tech, must be payable from that cold
//     start, and a flow resource in one is held to the bound below, with its
//     constant (firstCopyProblems);
//   - every flow resource (food, faith, culture, soldiers) the gate asks for is
//     made within the age's pacing target at config.FlowIncome;
//   - the storage ladder: the first copy of the new age's storage building
//     fits, with GateLadderMargin to spare, in the least storage the gate
//     forces a player to hold (see ladderForced and gateStore). Storage is
//     never lost (no catastrophe takes it and it can't be sold), so that is
//     the least anyone enters the age with, and the age lock leaves the new
//     age's storage as the only storage they can build there.
//
// Storage is a wall since the storage rule (config/storage_rule.go), and two
// of these are stated in its terms. A gate asks for five copies of the
// storage building of the age it leaves, so the least storage it forces is
// the store those copies give on top of every earlier gate's (gateStore): the
// rule's reference store. And a required storage copy is not held to
// GateBuildingMargin against the most storage of the age: it is the building
// that makes that storage, so each copy up to the last one asked for must
// fit, with the rule's own margin, under the store the copies before it
// give. The ladder's margin is the rule's too (it was 1.25 while storage was
// sized to hours of income and five times what any price needed): the rule
// sizes five copies to hold the next age's first storage building and
// config.StorageRuleMargin more.
const (
	GateBuildingMargin = config.StorageRuleRoom
	GateResourceMargin = 1.25
	GateWonderMargin   = 1.0 // no longer checked: see the wonder, above
	GateTrickleHours   = 48.0
	GateLadderMargin   = config.StorageRuleMargin
)

// GateProblem is an age requirement that breaks the Gate Covenant, found from
// config alone. A run only ever meets the first one, so the report lists them
// all up front.
type GateProblem struct {
	From string `json:"from"`
	To   string `json:"to"`
	// Kind is "building", "wonder", "unbuildable", "unsourced",
	// "flow", "ladder", "tech_locked" (a gate needs a building whose tech
	// can't be finished in the age it is built in) or "dead_building" (a building whose price has
	// no source in its own age; From is that age, To is empty).
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
	// Flow problems: what FlowIncome makes over the age's target.
	Made float64 `json:"made,omitempty"`
	// Ladder problems: Key is the new age's storage building, Need its first
	// copy's price in Resource, MaxStorage the least storage the gate forces,
	// and ForcedBy what forces it (a requirement or a required building).
	ForcedBy string `json:"forced_by,omitempty"`
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
	fmt.Fprintf(sb, "From config alone, against the most storage buildable in the age you advance from, with no build_cost discounts. Every building a gate asks for must be buildable in that age, and the tech it waits for, if any, must be finishable in that age. The last required copy of each must cost at most 1/%g of the storage. The wonder is paid through its bank, so it is held to no storage margin. Every resource a gate needs, the wonder's included, must be obtainable in that age from a cold start: a town with empty stores and only the buildings the gates so far asked for. Every flow resource a gate needs must be made within the age's target at a moderate income. The first copy of the new age's storage must fit with %gx to spare in the least storage the gate forces a player to hold. On arrival, the first copy of every building of the new age that a run can build, and the cheapest tech of the new age, must be payable from that same cold start. No building may cost a resource with no source in its own age. Together these checks are the Gate Covenant. `go test ./smoke` fails on any row here; the runtime invariants are what fail a run.\n\n", GateBuildingMargin, GateLadderMargin)
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
			case "ladder":
				fmt.Fprintf(sb, "| %s → %s | first %s | %s %s | - | %gx under the %s the gate forces (%s) |\n", g.From, g.To, g.Key, num(g.Need), g.Resource, g.Margin, num(g.MaxStorage), g.ForcedBy)
			case "flow":
				fmt.Fprintf(sb, "| %s → %s | %s (for %s) | %s %s | - | makes %s in the target |\n", g.From, g.To, g.Resource, g.Key, num(g.Need), g.Resource, num(g.Made))
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
	return float64(d.BaseCost[res] * detmath.Pow(d.CostScale, float64(n-1)))
}

// StaticGates checks every advance against the Gate Covenant and returns
// what breaks it, plus the tightest requirement of each advance.
func StaticGates() ([]GateProblem, []GateSlack) {
	return staticGates(config.Ages(), config.BuildingByKey())
}

// coldStart is what a player entering an age is guaranteed to have, and
// what they can reach from there by building in that age alone.
// It is the Gate Covenant's model of sourcing ("from a cold start"): a player
// may skip any building no gate required, so nothing optional from an earlier
// age may be relied on. The Iron Age trading post was the case that forced
// it: it cost gold and was the Iron Age's only gold producer, and the Bronze
// Age building that would have broken the loop was optional. Since the gate
// rule (config/gate_rule.go) it is not: the gate out of the Bronze Age asks
// for the market building, and every other building of the age that makes
// something.
//
// A town arrives with empty stores and, of what it owns, only what the gates
// so far asked for. Nothing else carries over into age A:
//
//   - every building an earlier gate required, and every earlier age's
//     wonder (each advance requires them; an upgrade keeps the lineage and
//     its output);
//   - so a resource is sourced on arrival when one of those buildings makes
//     it. Nothing is banked: a gate asks for no amount of any resource, and
//     what a town holds when it advances is whatever it did not spend.
//
// A building or a tech priced in a sourced resource is payable. One priced in
// a resource first made in A is payable once a building of A that makes it is
// itself payable: solve follows that chain, a building at a time, to a fixed
// point.
//
// A building that waits for a tech (BuildingDef.RequiredTech) is one of A's
// producers only when the tech is one no run leaves A without: the keystone
// of A's wonder or of an earlier one, or a tech one of those stands on. A
// building behind any other tech may never be built, so nothing a gate asks
// for may hang on it. This is the content rule for moving a building onto a
// tech: the age must keep another way to make what it makes, or its tech
// must be on the spine in time.
//
// A resource is a way in when A's own producers (bought with what is
// reachable) make it, when a building an earlier gate required makes it,
// when it is hand gathered (through the Medieval Age), or when techs whose
// flat output covers the amount within GateTrickleHours do. How thin a carried
// supply is against a gate's total is not judged for a resource that is not a
// flow resource: the flow bound is the covenant's only test of supply against
// a total. The first copy of each building of A, and the first tech, must be
// payable within that bound (firstCopyProblems), and a carried supply counts
// toward a first copy only if it is not thin: the copies the gates required,
// at their listed output, must make the copy's price in one target of A
// (CarriedSupplyTargets).
type coldStart struct {
	age      string
	unlocked map[string]bool
	trickle  map[string]float64 // flat per-tick tech output up to the age
	boot     map[string]bool    // steady carried supply: bootstraps a gate's total and a first copy it is not too thin for
	made     map[string]float64 // per tick, at listed output: the carried required copies of what they make
	reach    map[string]bool    // resources the age itself supplies in bulk
	carried  map[string]bool    // buildings carried over
	held     map[string]bool    // techs no run leaves the age without
}

// CarriedSupplyTargets is how many targets of the new age the carried required
// copies get to make a first copy's price in: one. A carried supply that
// cannot make the price in that time is a trickle against it, and does not
// count.
const CarriedSupplyTargets = 1.0

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

// techReachable reports whether tech can be finished by the end of age: it
// and what it needs, all the way down, open in age or earlier. A gate that
// needs a building behind a tech that isn't would soft-lock the run.
func techReachable(tech, age string) bool {
	return techReachableIn(config.TechByKey(), config.AgeOrder(), tech, age)
}

// techReachableIn is techReachable over the given techs and age order. A
// tech needs every one of its Prerequisites and, when it has an either-or
// group, one of its keys: one branch that opens in time is enough.
func techReachableIn(techs map[string]config.TechDef, ages []string, tech, age string) bool {
	order := map[string]int{}
	for i, a := range ages {
		order[a] = i
	}
	limit, known := order[age]
	if !known {
		return false
	}
	verdict := map[string]bool{}
	walking := map[string]bool{}
	var ok func(string) bool
	ok = func(k string) bool {
		if v, done := verdict[k]; done {
			return v
		}
		t, found := techs[k]
		at, placed := order[t.Age]
		if !found || !placed || at > limit || walking[k] {
			return false // unknown, too late, or a loop: it can never be finished
		}
		walking[k] = true
		v := t.PrereqsMet(ok)
		walking[k] = false
		verdict[k] = v
		return v
	}
	return ok(tech)
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
	copies := map[string]int{} // how many copies of each carried building the gates required
	trickle := map[string]float64{}
	techs := config.Technologies()
	// held is, age by age, the techs no run leaves the age without: the gate
	// rule's own set (config.GateTownHeldTechs).
	defList := make([]config.BuildingDef, 0, len(defs))
	for _, k := range sortedKeys(defs) {
		defList = append(defList, defs[k])
	}
	held := config.GateTownHeldTechs(defList, ages, techs)
	for i, a := range ages {
		for _, r := range a.UnlockResources {
			unlocked[r] = true
		}
		if i > 0 {
			for bld, n := range a.BuildingReqs {
				carried[bld] = true
				copies[bld] = n
			}
			if w := ageWonder(ages[i-1], defs); w != "" {
				carried[w] = true
				copies[w] = 1
			}
		}
		for _, t := range techs {
			if t.Age != a.Key {
				continue
			}
			for _, e := range t.Effects {
				if e.Kind == config.EffectFlatOutput && e.Value > 0 {
					trickle[e.Target] += e.Value
				}
			}
		}
		cs := &coldStart{age: a.Key, unlocked: cloneSet(unlocked),
			trickle: cloneMap(trickle), boot: map[string]bool{}, made: map[string]float64{}, reach: map[string]bool{},
			carried: cloneSet(carried), held: held[i]}
		for _, k := range sortedKeys(carried) {
			d := defs[k]
			for _, e := range d.Effects {
				if e.Type == "production" && e.Value > 0 && unlocked[e.Target] {
					cs.boot[e.Target] = true
					cs.made[e.Target] += float64(float64(copies[k]) * e.Value)
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

// affords reports whether amount of res can be paid once in the age, as the
// first copy of a building or a tech: the age supplies it, the carried
// required copies make it in CarriedSupplyTargets targets of the age (a
// carried supply that cannot is a trickle against the price and does not
// count), or techs cover it.
func (cs *coldStart) affords(res string, amount float64) bool {
	if !cs.unlocked[res] {
		return false
	}
	return cs.reach[res] || cs.carriedMakes(res, amount) || cs.covers(res, amount)
}

// carriedMakes reports whether the copies of the buildings the gates so far
// required, at their listed output, make amount of res within
// CarriedSupplyTargets targets of the age.
func (cs *coldStart) carriedMakes(res string, amount float64) bool {
	return cs.boot[res] && float64(cs.made[res]*config.AgeTargetTicks(cs.age)*CarriedSupplyTargets) >= amount
}

// covers reports whether techs within GateTrickleHours cover amount of res
// without any producer.
func (cs *coldStart) covers(res string, amount float64) bool {
	return cs.trickle[res]*GateTrickleHours*3600/config.TickSeconds >= amount
}

// sources reports whether a town has a way to make res, for the amount a gate
// asks for: a building of the age that can be paid for makes it, a building an
// earlier gate required (and so every town carries) makes it, or techs cover
// it. With the market gone this is the only way in, so the carried makers
// count here as they do for a first copy (affords).
func (cs *coldStart) sources(res string, amount float64) bool {
	return cs.unlocked[res] && (cs.reach[res] || cs.boot[res] || cs.covers(res, amount))
}

// solve grows reach to a fixed point: buy every building of the age whose
// first copy is affordable. A resource is sourced when a building of the age
// that can be paid for makes it; nothing is bought from anyone.
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
			if d.RequiredTech != "" && !cs.held[d.RequiredTech] {
				continue // it waits for a tech a run may never research
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
			for _, e := range d.Effects {
				if e.Type == "production" && e.Value > 0 {
					add(e.Target)
				}
			}
		}
	}
}

// ColdStartOpen is the one cold-start problem the tables may have, named here
// and not skipped: the Transcendent Age's Singularity Core is priced in 576
// quadrillion quantum flux, and the last age has no makers of its own (its only
// building a run can build without an optional tech is housing), so the carried
// copies cannot make it in a target of the age. The maker rule lists it as an
// open hole (config.MakerOpenHoles). TestColdStartOpenIsTheOneException proves
// the row is there and that nothing else is.
var ColdStartOpen = []GateProblem{
	{From: "transcendent_age", Kind: "dead_building", Key: "singularity_core", Resource: "quantum_flux"},
}

// staticGates is staticGatesAll without the rows ColdStartOpen names.
func staticGates(ages []config.AgeDef, defs map[string]config.BuildingDef) ([]GateProblem, []GateSlack) {
	all, slack := staticGatesAll(ages, defs)
	var out []GateProblem
	for _, g := range all {
		open := false
		for _, o := range ColdStartOpen {
			open = open || (g.From == o.From && g.Kind == o.Kind && g.Key == o.Key && g.Resource == o.Resource)
		}
		if !open {
			out = append(out, g)
		}
	}
	return out, slack
}

func staticGatesAll(ages []config.AgeDef, defs map[string]config.BuildingDef) ([]GateProblem, []GateSlack) {
	var out []GateProblem
	cold := coldStarts(ages, defs)
	techs := config.TechByKey()
	var slack []GateSlack
	for i := 0; i+1 < len(ages); i++ {
		from, to := ages[i], ages[i+1]
		wonder := ageWonder(from, defs)
		// Every resource the gate needs, what first needs it, and how much in
		// all (the requirement, every required copy undiscounted, the wonder).
		needs := map[string]string{}
		amount := map[string]float64{}
		thin := map[string]bool{} // parts of a once-bought price the cold start cannot make in time
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
			// The wonder's keystone tech is paid in knowledge, which no gate
			// asks for itself: the age must make it.
			if t, ok := techs[defs[wonder].RequiredTech]; ok && t.Cost > 0 {
				if _, ok := needs["knowledge"]; !ok {
					needs["knowledge"] = t.Key
				}
				amount["knowledge"] += t.Cost
				if !cold[i].affords("knowledge", t.Cost) {
					thin["knowledge"] = true
				}
			}
			// The wonder and its keystone tech are bought once, so each part
			// of their price is a first copy: a carried supply too thin to
			// make it in a target of the age does not count (affords).
			for res, c := range defs[wonder].BaseCost {
				if c > 0 && !config.IsFlowResource(res) && !cold[i].affords(res, c) {
					thin[res] = true
				}
			}
		}
		gated := sortedKeys(to.BuildingReqs)
		if wonder != "" {
			gated = append(gated, wonder)
		}
		for _, bld := range gated {
			if t := defs[bld].RequiredTech; t != "" && !techReachable(t, from.Key) {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "tech_locked", Key: bld})
			}
		}
		for _, res := range sortedKeys(needs) {
			if !cold[i].sources(res, amount[res]) || thin[res] {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "unsourced", Key: needs[res], Resource: res})
				continue
			}
			if !config.IsFlowResource(res) || amount[res] <= 0 {
				continue
			}
			made := float64(rules.Core().FlowIncome(res, from.Key) * config.AgeTargetTicks(from.Key))
			if short := amount[res] - made; short > 0 {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "flow", Key: needs[res], Resource: res,
					Need: amount[res], Made: made})
			}
		}
		tight := GateSlack{From: from.Key, To: to.Key, Need: 1, MaxStorage: math.Inf(1)}
		consider := func(key, res string, need, m float64) {
			if m/need < tight.Ratio() {
				tight.Key, tight.Resource, tight.Need, tight.MaxStorage = key, res, need, m
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
				if d.Category == "storage" {
					// Each copy asked for, under the store the copies
					// before it give: the storage every earlier gate
					// forced, and this building's own earlier copies.
					before := gateStore(ages[:i+1], defs) + baseStorage(res)
					for k := 1; k <= n; k++ {
						price := lastCopyPrice(d, res, k)
						m := before + float64(float64(k-1)*storagePerCopy(d))
						if price*config.StorageRuleMargin <= m {
							continue
						}
						if worst == nil || m/price < worst.MaxStorage/worst.Need {
							worst = &GateProblem{From: from.Key, To: to.Key, Kind: "building", Key: bld, Count: k,
								Resource: res, Need: price, MaxStorage: m, Margin: config.StorageRuleMargin}
						}
					}
					continue
				}
				// Under the store the gates force by the time the copy is
				// bought: this gate's own storage copies included.
				last := lastCopyPrice(d, res, n)
				m := gateStore(ages[:i+2], defs) + baseStorage(res)
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
		// The wonder is not held to a store: it is the one price allowed
		// over one, paid into its bank in rounds (deposits, overflow), and
		// the storage rule leaves it out of what a store must hold. What its
		// price is made of is checked above, with the gate's other needs:
		// every part of it must have a source from a cold start.
		out = append(out, ladderProblems(ages[:i+2], from, to, defs)...)
		out = append(out, firstCopyProblems(from, to, cold[i+1], defs, config.Technologies())...)
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

// firstCopyProblems checks what a town can do on arriving in `to` from `from`
// with empty stores and the buildings the gates so far asked for: the first
// copy of every building of `to` that a run can build, and the first tech of
// `to` (the cheapest), must each be payable in every resource they cost. A
// resource is payable when the cold start sources it (dead_building says so
// for a building; the first tech is checked here). A flow resource is held to
// the bound the covenant puts on a gate's flow resource: what the reference
// income makes in the age's pacing target. No number is new.
func firstCopyProblems(from, to config.AgeDef, cs *coldStart, defs map[string]config.BuildingDef, techs []config.TechDef) []GateProblem {
	var out []GateProblem
	check := func(key, res string, price float64) {
		if price <= 0 {
			return
		}
		if !cs.affords(res, price) {
			if _, isBuilding := defs[key]; !isBuilding {
				out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "unsourced", Key: key, Resource: res})
			}
			return // a building's missing source is a dead_building row already
		}
		if !config.IsFlowResource(res) {
			return
		}
		made := float64(rules.Core().FlowIncome(res, to.Key) * config.AgeTargetTicks(to.Key))
		if short := price - made; short > 0 {
			out = append(out, GateProblem{From: from.Key, To: to.Key, Kind: "flow", Key: key, Resource: res,
				Need: price, Made: made})
		}
	}
	for _, k := range sortedKeys(defs) {
		d := defs[k]
		if d.RequiredAge != to.Key || d.Category == "wonder" {
			continue
		}
		if d.RequiredTech != "" && !cs.held[d.RequiredTech] {
			continue // it waits for a tech a run may never research
		}
		for _, res := range sortedKeys(d.BaseCost) {
			check(k, res, d.BaseCost[res])
		}
	}
	first := config.TechDef{}
	for _, t := range techs {
		if t.Age == to.Key && t.Cost > 0 && (first.Key == "" || t.Cost < first.Cost) {
			first = t
		}
	}
	if first.Key != "" {
		check(first.Key, "knowledge", first.Cost)
	}
	return out
}

// ladderForced is the least storage the gate from `from` into `to` forces a
// player to hold: its biggest single price, whether a resource requirement
// (held all at once) or the last required copy of a building buildable in
// `from` (paid all at once), undiscounted. The wonder forces nothing: it is
// banked a deposit at a time. Every storage building raises every cap alike
// (Target "all"), so the caps move together and one figure stands for all of
// them. by names what forces it.
func ladderForced(from, to config.AgeDef, defs map[string]config.BuildingDef) (forced float64, by string) {
	for _, bld := range sortedKeys(to.BuildingReqs) {
		d, ok := defs[bld]
		if !ok || d.Category == "wonder" || (d.RequiredAge != "" && d.RequiredAge != from.Key) {
			continue // unbuildable requirements are their own problem
		}
		n := to.BuildingReqs[bld]
		for _, res := range sortedKeys(d.BaseCost) {
			if last := lastCopyPrice(d, res, n); last > forced {
				forced, by = last, fmt.Sprintf("%s #%d, %s %s", bld, n, num(last), res)
			}
		}
	}
	return forced, by
}

// storagePerCopy is what one copy of d adds to every store (its "all"
// storage).
func storagePerCopy(d config.BuildingDef) float64 {
	per := 0.0
	for _, e := range d.Effects {
		if e.Type == "storage" && e.Target == "all" {
			per += e.Value
		}
	}
	return per
}

// baseStorage is what res holds before any building.
func baseStorage(res string) float64 {
	for _, r := range config.BaseResources() {
		if r.Key == res {
			return r.BaseStorage
		}
	}
	return 0
}

// gateStore is the storage, on top of a resource's base, that the gates into
// ages[1:] force between them: every storage copy they ask for, at what a
// copy holds. A town that has advanced into the last of ages holds at least
// this, in the one place a town can get it: an older age's storage cannot be
// built later.
func gateStore(ages []config.AgeDef, defs map[string]config.BuildingDef) float64 {
	total := 0.0
	for _, a := range ages[1:] {
		for _, bld := range sortedKeys(a.BuildingReqs) {
			if d, ok := defs[bld]; ok && d.Category == "storage" {
				total += float64(float64(a.BuildingReqs[bld]) * storagePerCopy(d))
			}
		}
	}
	return total
}

// ladderProblems checks the storage ladder for one advance, the last of
// ages: the first copy of each storage building of `to` must fit, with
// GateLadderMargin to spare, in the storage the gates force: the larger of
// the gate's biggest single price (ladderForced) and the store the storage
// copies asked for so far give (gateStore, on the resource's base). One row
// per storage building, for the resource with the least headroom. A player
// who met the gate with nothing to spare could otherwise never raise a cap
// in `to`: the age lock forbids every older storage building.
func ladderProblems(ages []config.AgeDef, from, to config.AgeDef, defs map[string]config.BuildingDef) []GateProblem {
	forced, by := ladderForced(from, to, defs)
	store := gateStore(ages, defs)
	var out []GateProblem
	for _, k := range sortedKeys(defs) {
		d := defs[k]
		if d.Category != "storage" || d.RequiredAge != to.Key {
			continue
		}
		var worst *GateProblem
		for _, res := range sortedKeys(d.BaseCost) {
			c := d.BaseCost[res]
			held, heldBy := forced, by
			if s := store + baseStorage(res); s > held {
				held, heldBy = s, "the storage copies the gates ask for"
			}
			if c*GateLadderMargin <= held || (worst != nil && c <= worst.Need) {
				continue
			}
			worst = &GateProblem{From: from.Key, To: to.Key, Kind: "ladder", Key: k, Resource: res,
				Need: c, MaxStorage: held, Margin: GateLadderMargin, ForcedBy: heldBy}
		}
		if worst != nil {
			out = append(out, *worst)
		}
	}
	return out
}
