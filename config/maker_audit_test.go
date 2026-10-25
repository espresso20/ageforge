package config

import (
	"maps"
	"math"
	"slices"
	"sort"
)

// The maker-rule audit: the derivation of the paybacks and the holes from the
// building tables, which the tests of maker_rule_test.go, first_makers_test.go
// and raw_materials_test.go recompute against the authored tables. The rule is
// described at the top of maker_rule.go. Nothing the game runs reads these.

// MakerHealthyShare is how much of the age's best output (units per tick) a
// maker must make to count as healthy.
const MakerHealthyShare = 0.01

// makerSkipped are the materials the maker rule leaves alone.
func makerSkipped(res string) bool { return res == "knowledge" || flowResources[res] }

// makerUnits is the price of d's first copy, and its output per tick, in the
// price units of levels. Output of a material in skip is not counted.
func makerUnits(d BuildingDef, lv map[string]float64, skip map[string]bool) (price, out float64) {
	price = priceUnits(d.BaseCost, lv)
	for _, e := range d.Effects {
		if e.Type != "production" || e.Value <= 0 || lv[e.Target] <= 0 || skip[d.Key+"/"+e.Target] {
			continue
		}
		out += float64(e.Value / lv[e.Target])
	}
	return price, out
}

// MakerPayback is the payback figures of one age: the median of its healthy
// makers, the lowest and highest of them, and how many of the age's makers
// are healthy.
type MakerPayback struct {
	Median, Lo, Hi  float64
	Healthy, Makers int
}

// MakerPaybacks recomputes payback(A) for every age in order from defs. The
// makers are the age's non-wonder buildings with a priced output, optional
// tech or not.
func MakerPaybacks(defs []BuildingDef, order []string) map[string]MakerPayback {
	levels := priceLevels(defs)
	out := map[string]MakerPayback{}
	for _, age := range order {
		lv := levels[age]
		type row struct{ out, payback float64 }
		var rows []row
		best := 0.0
		for _, d := range defs {
			if d.RequiredAge != age || d.Category == "wonder" {
				continue
			}
			price, o := makerUnits(d, lv, nil)
			if o <= 0 {
				continue
			}
			rows = append(rows, row{o, float64(price / o)})
			best = math.Max(best, o)
		}
		var paybacks []float64
		for _, r := range rows {
			if r.out >= float64(MakerHealthyShare*best) {
				paybacks = append(paybacks, r.payback)
			}
		}
		if len(paybacks) == 0 {
			continue
		}
		sort.Float64s(paybacks)
		out[age] = MakerPayback{Median: median(paybacks), Lo: paybacks[0], Hi: paybacks[len(paybacks)-1],
			Healthy: len(paybacks), Makers: len(rows)}
	}
	return out
}

// MakerHole is one (age, material) the maker rule looks at, with what the
// rules decide: Rule 1, 2 or 3 and the Building that fills it, or Rule 4 with
// no building (nothing fits), or Rule 0 with Why saying the rules do not
// decide (a tie).
type MakerHole struct {
	Age, Material, Building string
	Rule                    int
	Why                     string
}

// makerRaw is the raw form of each material that has one.
var makerRaw = map[string]string{"iron": "iron_ore", "stone": "marble", "titanium": "titanium_ore", "dark_matter": "dark_matter_crystals"}

// RawLeftOut names the buildings the raw rule would price in a raw and does
// not, by building key, with the reason: the Modern Age's data and the
// Cyberpunk Age's crypto each have one source in their age, and the raw's
// first maker is priced in it. A raw paid by that source would be a loop (the
// Mainframe Hall costs titanium ore, which the Titanium Mine makes, and the
// mine costs the Hall's data), so a town could never start. A test proves each
// entry is that loop; nothing here is a tuning knob.
func RawLeftOut() map[string]string {
	return map[string]string{
		"mainframe_hall": "the Modern Age's only maker of data, which the Titanium Mine (the first maker of titanium ore) is priced in",
		"black_market":   "the Cyberpunk Age's only maker of crypto, which the Dark Crystal Mine (the first maker of dark matter crystals) is priced in",
	}
}

// FullCrewShare is what a building with worker slots makes at a full crew, as
// a share of its listed output: the unstaffed share plus what workers add. The
// game pays a building listed x count x (UnstaffedShare + StaffedShare x crew
// fill), so the listed output of a table is already the fully staffed rate and
// FullCrewShare is 1. The shortage test of the maker rule counts a carried
// building that has worker slots at its listed output times this (staffed
// output); the cold start's thin-source bound reads the same listed output. A
// full crew makes 1 / UnstaffedShare = 5 times what an empty building does,
// which is not a multiple of the listed output.
const FullCrewShare = UnstaffedShare + StaffedShare

// MakerTieBreaks are the ties between lineages the rules cannot decide, as the
// owner decided them: "age/material" -> the building that fills it.
var MakerTieBreaks = map[string]string{
	"electric_age/iron": "electric_arc_furnace",
}

// makerCarried is, for the age at position i, the copies of each building the
// gates so far required and the wonders of the earlier ages: what every town
// brings into the age (the cold start's carried set).
func makerCarried(defs []BuildingDef, ages []AgeDef, i int) map[string]int {
	copies := map[string]int{}
	for j := 1; j <= i && j < len(ages); j++ {
		for k, n := range ages[j].BuildingReqs {
			copies[k] = n
		}
		for _, d := range defs {
			if d.Category == "wonder" && d.RequiredAge == ages[j-1].Key {
				copies[d.Key] = 1
			}
		}
	}
	return copies
}

// MakerHoles works the maker rule out from defs, in age order.
//
// Rule 2 first: every building of an age whose output is the raw form of a
// material the age prices also makes that material, hole or not (Rule 2
// entries, one per building and material). Then the holes: a material the
// gate or the wonder is priced in that no healthy maker a run can build makes,
// and that is short: the copies every town carries into the age (the gates'
// required copies and the earlier wonders, as the tables hold them, with a
// building that has worker slots counted staffed, at FullCrewShare
// times its listed output) cannot make what the gate and the wonder ask of it
// in one target of the age. A material that is not short is not a hole, and neither is a
// shortage listed in MakerAcceptedShortages.
//
// "Last made" in rule 3 is read on the tables without the pairs in fills, as
// they stood before the fills were made; the carried supply is read on the
// tables as they are. Pass MakerFills() to read tables that hold them; the
// holes of tables that hold none come from passing nil.
func MakerHoles(defs []BuildingDef, ages []AgeDef, techs []TechDef, fills []MakerFill) []MakerHole {
	levels := priceLevels(defs)
	held := GateTownHeldTechs(defs, ages, techs)
	gate, wonder := MakerAsks(defs, ages)
	pos := AgePositions(ageKeys(ages))
	// skip are the outputs a fill put there: read without them, the tables are
	// as they stood before the fills. skipMade is the same without the Rule 1
	// pairs: a Rule 1 building made the material already, at a trickle, so it
	// is still a maker of it (the trickle Rule 1 finds), while its output,
	// which the fill put on the standard, counts for nothing.
	skip, skipMade := map[string]bool{}, map[string]bool{}
	for _, f := range fills {
		if f.Rule < 5 { // a placed maker is a maker that is there
			skip[f.Building+"/"+f.Material] = true
			if f.Rule > 1 {
				skipMade[f.Building+"/"+f.Material] = true
			}
		}
	}
	var out []MakerHole
	for i, a := range ages {
		lv := levels[a.Key]
		buildable := func(d BuildingDef) bool { return d.RequiredTech == "" || held[i][d.RequiredTech] }
		makes := func(d BuildingDef, res string) bool {
			if skipMade[d.Key+"/"+res] {
				return false
			}
			for _, e := range d.Effects {
				if e.Type == "production" && e.Value > 0 && e.Target == res {
					return true
				}
			}
			return false
		}
		supply := func(res string) float64 { return makerSupply(defs, ages, i, res) }
		best := 0.0
		outUnits := map[string]float64{}
		for _, d := range defs {
			if d.RequiredAge == a.Key && d.Category != "wonder" {
				_, o := makerUnits(d, lv, skip)
				outUnits[d.Key] = o
				best = math.Max(best, o)
			}
		}
		var rows []MakerHole
		// Rule 2, for every material the age prices that has a raw form.
		rawFilled := map[string]bool{}
		for res := range lv {
			rw := makerRaw[res]
			if rw == "" {
				continue
			}
			for _, d := range defs {
				if d.RequiredAge == a.Key && d.Category != "wonder" && buildable(d) && makes(d, rw) {
					rows = append(rows, MakerHole{Age: a.Key, Material: res, Building: d.Key, Rule: 2})
					rawFilled[res] = true
				}
			}
		}
		// A wonder part in a material the age does not price has no level to
		// put a standard on: the rules cannot decide it.
		for _, d := range defs {
			if d.RequiredAge != a.Key || d.Category != "wonder" {
				continue
			}
			for res, c := range d.BaseCost {
				if c > 0 && lv[res] <= 0 && !makerSkipped(res) {
					rows = append(rows, MakerHole{Age: a.Key, Material: res,
						Why: "the wonder is priced in " + res + " and no building of the age is, so the age has no level to put a standard on"})
				}
			}
		}
		asked := map[string]bool{}
		for res := range gate[a.Key] {
			asked[res] = true
		}
		for res := range wonder[a.Key] {
			asked[res] = true
		}
		var mats []string
		for res := range asked {
			if !makerSkipped(res) && !rawFilled[res] {
				mats = append(mats, res)
			}
		}
		sort.Strings(mats)
		for _, res := range mats {
			healthy := false
			var trickle []string
			for _, d := range defs {
				if d.RequiredAge != a.Key || d.Category == "wonder" || !buildable(d) || !makes(d, res) {
					continue
				}
				if outUnits[d.Key] >= float64(MakerHealthyShare*best) {
					healthy = true
				} else {
					trickle = append(trickle, d.Key)
				}
			}
			if healthy {
				continue
			}
			ask := float64((gate[a.Key][res] + wonder[a.Key][res]) * lv[res])
			if float64(supply(res)*LatePriceTargetTicks(a.Key)) >= ask {
				continue // the carried copies make what is asked: no shortage
			}
			if makerAccepted(a.Key, res) {
				continue // a shortage the owner accepts: a few more of the older buildings cover it
			}
			h := MakerHole{Age: a.Key, Material: res}
			switch {
			case len(trickle) == 1:
				h.Rule, h.Building = 1, trickle[0]
			case len(trickle) > 1:
				h.Why = "several trickle makers: " + joinKeys(trickle)
			default:
				// Rule 3: the lineages that made it in the latest earlier age.
				last := -1
				lineages := map[string]bool{}
				for _, d := range defs {
					j, ok := pos[d.RequiredAge]
					if !ok || j >= i || d.Category == "wonder" || d.LineageKey == "" || !makes(d, res) {
						continue
					}
					if j > last {
						last, lineages = j, map[string]bool{}
					}
					if j == last {
						lineages[d.LineageKey] = true
					}
				}
				var fit []string
				for _, d := range defs {
					if d.RequiredAge == a.Key && d.Category != "wonder" && lineages[d.LineageKey] && buildable(d) {
						fit = append(fit, d.Key)
					}
				}
				sort.Strings(fit)
				switch {
				case len(fit) == 1:
					h.Rule, h.Building = 3, fit[0]
				case len(fit) > 1 && slices.Contains(fit, MakerTieBreaks[a.Key+"/"+res]):
					h.Rule, h.Building = 3, MakerTieBreaks[a.Key+"/"+res]
				case len(fit) > 1:
					h.Why = "two lineages made it last, and both have a building this age: " + joinKeys(fit)
				default:
					h.Rule = 4
					h.Why = "nothing fits"
				}
			}
			rows = append(rows, h)
		}
		sort.SliceStable(rows, func(x, y int) bool {
			if rows[x].Material != rows[y].Material {
				return rows[x].Material < rows[y].Material
			}
			return rows[x].Building < rows[y].Building
		})
		out = append(out, rows...)
	}
	return out
}

// makerSupply is what the copies every town carries into the age at position
// i make of res per tick: the gates' required copies and the earlier wonders,
// as the tables hold them, at listed output (the listed output of a building
// with worker slots is its full-crew rate: FullCrewShare of it).
func makerSupply(defs []BuildingDef, ages []AgeDef, i int, res string) float64 {
	byKey := indexBy(defs, func(d BuildingDef) string { return d.Key })
	carried := makerCarried(defs, ages, i)
	total := 0.0
	for _, k := range slices.Sorted(maps.Keys(carried)) {
		d := byKey[k]
		mult := 1.0
		if d.WorkerCapacity > 0 {
			mult = FullCrewShare
		}
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 && e.Target == res {
				total += float64(float64(carried[k]) * e.Value * mult)
			}
		}
	}
	return total
}

// MakerShortageTargets is how many targets of the age the carried copies need
// to make what the gate and the wonder ask of res: the ask over what the
// carried copies make in one target. Under 1 the carried copies cover it; 0
// where nothing is asked or the supply is none (an unreachable ask reads as
// +Inf only when something is asked).
func MakerShortageTargets(defs []BuildingDef, ages []AgeDef, age, res string) float64 {
	pos := AgePositions(ageKeys(ages))
	i, ok := pos[age]
	if !ok {
		return 0
	}
	gate, wonder := MakerAsks(defs, ages)
	lv := priceLevels(defs)[age]
	ask := float64((gate[age][res] + wonder[age][res]) * lv[res])
	made := float64(makerSupply(defs, ages, i, res) * LatePriceTargetTicks(age))
	if ask <= 0 {
		return 0
	}
	if made <= 0 {
		return math.Inf(1)
	}
	return ask / made
}

func makerAccepted(age, res string) bool {
	for _, a := range MakerAcceptedShortages() {
		if a.Age == age && a.Material == res {
			return true
		}
	}
	return false
}

func joinKeys(keys []string) string {
	s := ""
	for i, k := range keys {
		if i > 0 {
			s += ", "
		}
		s += k
	}
	return s
}
