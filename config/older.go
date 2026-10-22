package config

import "sort"

// UpgradeShare is what one upgraded copy costs, as a share of the price the
// next new copy of its target would cost right now.
const UpgradeShare = 0.5

// InLine reports whether a building belongs to a building line, and so can be
// an older building. Storage, wonders, monuments and the standalone buildings
// are in no line, and a knowledge building keeps its tier: the knowledge store
// the tech prices are set against is made of the copies a player holds.
func InLine(d BuildingDef) bool {
	switch d.LineageKey {
	case "", "storage", "wonder", "monument":
		return false
	}
	switch d.Category {
	case "storage", "wonder", "monument":
		return false
	}
	return d.Scholars == 0
}

// LineRungs groups the buildings that are in a line by line, each line's
// rungs ordered from the first rung up.
func LineRungs(defs []BuildingDef) map[string][]BuildingDef {
	lines := map[string][]BuildingDef{}
	for _, d := range defs {
		if InLine(d) {
			lines[d.Line()] = append(lines[d.Line()], d)
		}
	}
	for _, rungs := range lines {
		sort.SliceStable(rungs, func(i, j int) bool { return rungs[i].LineageTier < rungs[j].LineageTier })
	}
	return lines
}

// CurrentRung is the rung of a line (rungs, ordered from the first) that
// belongs to age: the rung with the highest required age that is no later than
// age. A line that pauses in age keeps its last rung. ok is false when the
// line has no rung yet. ages gives each age's place in the order.
func CurrentRung(rungs []BuildingDef, ages map[string]int, age string) (cur BuildingDef, ok bool) {
	at, known := ages[age]
	if !known {
		return BuildingDef{}, false
	}
	best := -1
	for _, r := range rungs {
		pos, has := ages[r.RequiredAge]
		if !has || pos > at {
			continue
		}
		if pos >= best {
			cur, ok, best = r, true, pos
		}
	}
	return cur, ok
}

// UpgradeTarget is the building d upgrades to in age: the current rung of its
// line, when that rung belongs to a later age than d's own. ok is false when d
// is not older: it is the current rung, or it shares the current rung's age (a
// line with two rungs in one age, Stone Camp and Stone Pit, offers both new in
// that age), or it is in no line. rungs is d's line, ordered from the first.
func UpgradeTarget(rungs []BuildingDef, ages map[string]int, d BuildingDef, age string) (BuildingDef, bool) {
	if !InLine(d) {
		return BuildingDef{}, false
	}
	cur, ok := CurrentRung(rungs, ages, age)
	if !ok || ages[cur.RequiredAge] <= ages[d.RequiredAge] || cur.LineageTier <= d.LineageTier {
		return BuildingDef{}, false
	}
	return cur, true
}
