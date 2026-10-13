package config

// The gate rule: what the gate out of an age asks for in buildings, besides
// the keystone and the wonder. It is written in the tables (the gates in
// Ages(), authored literals); this file is the derivation, and a test
// recomputes what each gate must ask for from it and fails when a gate asks
// for less.
//
// No gate asks for an amount of a resource. A gate asks for the town its age
// is priced for: the reference town, which holds TownCopies copies of every
// building that is not storage (the storage rule's town, and smoke's
// RefCopies; the town rule, town_rule.go, says why eight). For the
// gate that leaves age A, on top of whatever the gate asks for today (a
// count is never lowered):
//
//  1. TownCopies copies of every building introduced in A that makes
//     something: a production effect with a positive value. That is food,
//     wood, stone, metals, gold, power and the rest; knowledge, faith and
//     culture; and soldiers, for a military building. The gate keeps the
//     larger of this and the count it already asks.
//  2. Not asked: the wonder (the gate asks for it already, apart from the
//     buildings), a monument or an embassy (they make no resource), housing
//     and storage (the gate asks for them under their own rules: housing at
//     the count it always had, storage at GateStorageCopies), and a building
//     a run may not have the tech for. A gate may only ask for what a player
//     who researched nothing optional can build: a building that waits for a
//     tech is asked for only when the tech is one no run leaves A without
//     (GateTownHeldTechs). That leaves out a building behind an optional
//     tech, and one behind a spine tech that A's own keystone does not stand
//     on.
//  3. A building can only be asked for as many copies as it can be built in
//     (its MaxCount, when it has one).
//  4. A knowledge building is not asked for TownCopies: it climbs at its own
//     rate and the gate asks the spend it always did (knowledge_rule.go).
//
// The covenants on gates (smoke.StaticGates) hold the result: each building
// asked for is buildable in A, and its last required copy fits the store with
// the covenant's margin. A building a covenant rules out is not silently
// dropped: it is named in GateTownLeftOut, with the reason, and the test
// proves the list is true.

// GateTownLeftOut names the buildings the gate rule would ask for and a
// covenant stops it asking for, by building key, with the reason. Nothing
// here is a tuning knob: an entry needs the numbers that break the covenant.
func GateTownLeftOut() map[string]string {
	return map[string]string{}
}

// GateTownHeldTechs is the techs a run cannot leave age A without: the
// keystone of the wonder of every age up to and including A, and everything
// each one stands on (Prerequisites, all the way down; an either-or group
// puts neither branch in it). Entry i is the set for ages[i].
func GateTownHeldTechs(defs []BuildingDef, ages []AgeDef, techs []TechDef) []map[string]bool {
	byKey := make(map[string]TechDef, len(techs))
	for _, t := range techs {
		byKey[t.Key] = t
	}
	wonderTech := map[string]string{}
	for _, d := range defs {
		if d.Category == "wonder" && d.RequiredTech != "" {
			if _, taken := wonderTech[d.RequiredAge]; !taken {
				wonderTech[d.RequiredAge] = d.RequiredTech
			}
		}
	}
	held := map[string]bool{}
	var hold func(key string)
	hold = func(key string) {
		t, ok := byKey[key]
		if !ok || held[key] {
			return
		}
		held[key] = true
		for _, p := range t.Prerequisites {
			hold(p)
		}
	}
	out := make([]map[string]bool, len(ages))
	for i, a := range ages {
		hold(wonderTech[a.Key])
		out[i] = make(map[string]bool, len(held))
		for k := range held {
			out[i][k] = true
		}
	}
	return out
}

// makesSomething reports whether d has a production effect: it makes a
// resource (food, gold, knowledge, faith, culture, soldiers and the rest).
func makesSomething(d BuildingDef) bool {
	for _, e := range d.Effects {
		if _, ok := producedBy(e); ok && e.Value > 0 {
			return true
		}
	}
	return false
}

// GateTown works the gate rule out: what the gate that leaves each age must
// ask for besides what it asks for today, by building key, keyed by the age
// the gate leads into (as the ages' BuildingReqs are read). Each count is the
// reference town's, TownCopies, or the building's MaxCount when that
// is smaller; a knowledge building's is the knowledge rule's
// (KnowledgeGateCount). Buildings in GateTownLeftOut are not in it.
func GateTown(defs []BuildingDef, ages []AgeDef, techs []TechDef) map[string]map[string]int {
	held := GateTownHeldTechs(defs, ages, techs)
	left := GateTownLeftOut()
	out := map[string]map[string]int{}
	for i := 0; i+1 < len(ages); i++ {
		for _, d := range defs {
			if d.RequiredAge != ages[i].Key {
				continue
			}
			switch d.Category {
			case "wonder", "monument", "storage", "housing":
				continue
			}
			if !makesSomething(d) {
				continue
			}
			if d.RequiredTech != "" && !held[i][d.RequiredTech] {
				continue
			}
			if _, skip := left[d.Key]; skip {
				continue
			}
			n := TownCopies
			if IsKnowledgeBuilding(d) {
				// A knowledge building climbs at KnowledgeRate: the gate
				// asks the spend it asked before (knowledge_rule.go).
				n = KnowledgeGateCount(KnowledgeGateWas()[d.Key])
			}
			if d.MaxCount > 0 {
				n = min(n, d.MaxCount)
			}
			if out[ages[i+1].Key] == nil {
				out[ages[i+1].Key] = map[string]int{}
			}
			out[ages[i+1].Key][d.Key] = n
		}
	}
	return out
}
