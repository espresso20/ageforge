package game

import (
	"fmt"
	"maps"
	"slices"

	"github.com/espresso20/ageforge/rules"
)

// The plan as a research queue (the tech tree, part 5).
//
//   - `plan research <tech>` takes any tech the player may see named
//     (spoilers.go): every age up to the next one, and on known ground every
//     age the account has reached. A tech of an age still hidden is
//     answered like a key that does not exist.
//   - It adds what the tech still needs first, in order (researchChain):
//     every prerequisite that is neither researched, being researched nor
//     planned, each after its own, then the tech. For an either-or group
//     none of whose techs is on its way it takes the branch that costs the
//     least knowledge to add. Research does not count against the plan's
//     60 items (MaxPlanItems), so a chain always fits.
//   - A tech of an age not reached yet waits in the plan for the age. While
//     it waits it holds neither the research slot's turn nor its knowledge,
//     like a tech whose prerequisite is not on its way (plan.go), so this
//     age's techs planned below it still run. Nothing is dropped for
//     waiting: a research item leaves the plan when it starts, when it is
//     researched some other way, or when the player removes it.

// researchChain is what planning tech key adds, in plan order: every tech it
// still needs that is not on its way, each after what it needs itself, then
// key. onWay reports a tech that is researched, being researched or planned
// already. For an either-or group none of whose techs is on its way (or in
// the chain so far) it takes the branch that adds the least knowledge cost,
// the first listed among equals. Empty when key is unknown or on its way.
func researchChain(set *rules.Set, key string, onWay func(string) bool) []string {
	var walk func(k string, seen map[string]bool, out *[]string)
	walk = func(k string, seen map[string]bool, out *[]string) {
		if seen[k] || onWay(k) {
			return
		}
		def, ok := set.Tech(k)
		if !ok {
			return
		}
		seen[k] = true
		for _, pre := range def.Prerequisites {
			walk(pre, seen, out)
		}
		covered := func(a string) bool { return seen[a] || onWay(a) }
		if len(def.AnyOf) > 0 && !slices.ContainsFunc(def.AnyOf, covered) {
			best, bestCost := "", 0.0
			for _, a := range def.AnyOf {
				var trial []string
				walk(a, maps.Clone(seen), &trial)
				cost := 0.0
				for _, t := range trial {
					d, _ := set.Tech(t)
					cost += d.Cost
				}
				if len(trial) > 0 && (best == "" || cost < bestCost) {
					best, bestCost = a, cost
				}
			}
			if best != "" {
				walk(best, seen, out)
			}
		}
		*out = append(*out, k)
	}
	var chain []string
	walk(key, map[string]bool{}, &chain)
	return chain
}

// PlanResearchChain is what `plan research key` would add to the plan in the
// snapshot st, in order, key last: nil when the tech is unknown, researched,
// being researched or planned already. The UI reads it to say what a card's
// Enter will do; the engine decides with the same walk (PlanAddResearchChain).
func PlanResearchChain(st GameState, key string) []string {
	planned := map[string]bool{}
	for _, it := range st.Plan {
		if it.Kind == PlanResearch {
			planned[it.Key] = true
		}
	}
	return researchChain(st.Ruleset(), key, func(k string) bool {
		return st.Research.Techs[k].Researched || (k != "" && st.Research.CurrentTech == k) || planned[k]
	})
}

// researchOnWay reports whether tech key is researched, being researched or
// in the plan. Caller holds the lock.
func (ge *GameEngine) researchOnWay(key string) bool {
	if ge.Research.IsResearched(key) || ge.Research.currentTech == key {
		return true
	}
	for _, it := range ge.plan {
		if it.Kind == PlanResearch && it.Key == key {
			return true
		}
	}
	return false
}

// planTechAhead reports whether tech key belongs to an age not reached yet:
// in the plan it waits for the age.
func (ge *GameEngine) planTechAhead(key string) bool {
	def, ok := ge.rules.Tech(key)
	if !ok {
		return false
	}
	order := ge.progress.GetAgeOrder()
	return order[def.Age] > order[ge.age]
}

// errTechOutOfSight answers a tech key the plan cannot take because the
// player may not see it: "No tech 'x' in sight. Did you mean 'y'? Type
// research to see the tree." A key that does not exist and a tech of an age
// still hidden read the same, and the suggestion only ever names a tech in
// sight.
func (ge *GameEngine) errTechOutOfSight(key string, sight AgeSight) error {
	seen := map[string]bool{}
	for _, t := range ge.rules.Techs() {
		if sight.Age(t.Age) {
			seen[t.Key] = true
		}
	}
	msg := fmt.Sprintf("No tech '%s' in sight.", key)
	if s := closestKey(key, seen); s != "" {
		msg += fmt.Sprintf(" Did you mean '%s'?", s)
	}
	return fmt.Errorf("%s Type research to see the tree.", msg)
}

// PlanAddResearchChain plans tech key and, before it, everything it still
// needs that is not researched, being researched or planned (researchChain).
// The tech must be one the player may see and not researched, in progress
// or planned already. It returns the techs it added, in plan order, key
// last. Research takes no room in the plan, so the chain always fits.
func (ge *GameEngine) PlanAddResearchChain(key string) ([]string, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	def, ok := ge.rules.Tech(key)
	if sight := ge.ageSightLocked(); !ok || !sight.Age(def.Age) {
		return nil, ge.errTechOutOfSight(key, sight)
	}
	for _, it := range ge.plan {
		if it.Kind == PlanResearch && it.Key == key {
			return nil, fmt.Errorf("%s is already in the plan.", def.Name)
		}
	}
	if reason := ge.planResearchInvalid(key); reason != "" {
		return nil, fmt.Errorf("Can't plan %s: %s.", def.Name, reason)
	}
	chain := researchChain(ge.rules, key, ge.researchOnWay)
	for _, k := range chain {
		ge.plan = append(ge.plan, PlanItem{Kind: PlanResearch, Key: k, Count: 1})
		ge.logPlanAddLocked(ge.plan[len(ge.plan)-1], 1)
	}
	ge.notePlanSize()
	return chain, nil
}
