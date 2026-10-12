package config

import (
	"math"
	"slices"
	"sort"
)

// The maker rule: with no market swap, every material a town is charged for
// has to come from a building, at the same standard as every other producer.
// The tables are authored (the building files, MakerPaybackTicks and
// MakerFills below); this file is the derivation, and a test recomputes the
// standard, the holes and every filled output from it.
//
// The standard. A maker of an age pays back its own price in the same number
// of ticks. Measure a building's first copy in the age's price units
// (priceUnits: 1.0 is the median first-copy price of one resource) and its
// output in units per tick (the output of each material the age prices, over
// that material's level, added up). Its payback is the one over the other.
// A maker is healthy when its output is at least MakerHealthyShare of the
// age's best. payback(A) is the median payback of the age's healthy makers;
// MakerPaybackTicks writes it down. A maker put on the standard makes
// (first-copy price in units / payback(A)) x the level of its material.
//
// The holes. An (age, material) is a hole when a building the gate out of the
// age requires, or the age's wonder, is priced in the material, and no
// building of the age that a run with no optional tech can build is a healthy
// maker of it. Knowledge is left alone (it is a later step), and so are the
// flow resources, which have no level. Who fills a hole, in this order:
//
//  1. a building of the age that already makes the material, at a trickle:
//     its output goes to the standard;
//  2. else a building of the age whose output is the raw form of the material
//     (iron_ore for iron, marble for stone, titanium_ore for titanium,
//     dark_matter_crystals for dark matter): it also makes the material;
//  3. else the age's building in the lineage that last made the material in an
//     earlier age, on the tables as they stood before any fill (a fill is not
//     a lineage's history: it would hand iron to the mines, which made iron ore
//     for it): it also makes the material;
//  4. else nothing fits, and the hole is listed (MakerOpenHoles).
//
// A building that ends up making two or more of the materials its age charges
// for makes the standard once, split between them in proportion to what the
// age's gate asks of each, in units (the buildings the gate requires, every
// required copy at the price it is bought at; the wonder is not part of what
// the gate asks). A material the gate asks
// nothing of counts as one unit (a single median price). A building keeps the output it has when it is not touched
// by a fill.

// MakerHealthyShare is how much of the age's best output (units per tick) a
// maker must make to count as healthy.
const MakerHealthyShare = 0.01

// MakerPaybackTicks is payback(A) for every age, in ticks: the median payback
// of the age's healthy makers, typed here and recomputed by a test. It is what
// the old payback rule set the ages' buildings by (DealPaybackTicks still
// holds that rule's figures, for the flow resources).
var MakerPaybackTicks = map[string]float64{
	"primitive_age":    28.12,
	"stone_age":        109.38,
	"bronze_age":       764.57,
	"iron_age":         1365.7,
	"classical_age":    2193.1,
	"medieval_age":     3183.9,
	"renaissance_age":  9438.1,
	"colonial_age":     6056.4,
	"industrial_age":   8667.1,
	"victorian_age":    14896,
	"electric_age":     20116,
	"atomic_age":       25407,
	"modern_age":       25950,
	"information_age":  15227,
	"digital_age":      40929,
	"cyberpunk_age":    50249,
	"fusion_age":       129379,
	"space_age":        96530,
	"interstellar_age": 79991,
	"galactic_age":     169450,
	"quantum_age":      51348,
	"transcendent_age": 47204,
}

// MakerFill is one hole the rules fill: in Age, Building also makes Material,
// by Rule 1, 2 or 3. Rule 5 is a maker placed by the owner's decision where
// the rules found no building to fill a hole (the Modern Age's data): a new
// building, put on the standard like any maker; the rules read it as a maker
// that is there. The building tables hold the output; MakerFills is the
// record of why, and a test proves it against the rules.
type MakerFill struct {
	Age, Material, Building string
	Rule                    int
}

// MakerOpenHole is a hole the rules leave open, with the reason.
type MakerOpenHole struct {
	Age, Material, Why string
}

// MakerFills lists every hole the rules fill, in age order.
func MakerFills() []MakerFill {
	return []MakerFill{
		{Age: "iron_age", Material: "iron", Building: "marble_quarry", Rule: 2},
		{Age: "iron_age", Material: "stone", Building: "marble_quarry", Rule: 2},
		{Age: "classical_age", Material: "iron", Building: "marble_works", Rule: 2},
		{Age: "classical_age", Material: "stone", Building: "marble_works", Rule: 2},
		{Age: "medieval_age", Material: "iron", Building: "stonemasons_guild", Rule: 2},
		{Age: "medieval_age", Material: "stone", Building: "stonemasons_guild", Rule: 3},
		{Age: "renaissance_age", Material: "iron", Building: "iron_mine", Rule: 2},
		{Age: "renaissance_age", Material: "stone", Building: "iron_mine", Rule: 3},
		{Age: "colonial_age", Material: "iron", Building: "deep_iron_mine", Rule: 2},
		{Age: "colonial_age", Material: "stone", Building: "deep_iron_mine", Rule: 3},
		{Age: "industrial_age", Material: "iron", Building: "steam_mine", Rule: 2},
		{Age: "industrial_age", Material: "stone", Building: "steam_mine", Rule: 3},
		{Age: "victorian_age", Material: "iron", Building: "bessemer_plant", Rule: 3},
		{Age: "victorian_age", Material: "stone", Building: "uranium_mine", Rule: 3},
		{Age: "electric_age", Material: "iron", Building: "electric_arc_furnace", Rule: 3},
		{Age: "atomic_age", Material: "iron", Building: "advanced_alloy_plant", Rule: 3},
		{Age: "atomic_age", Material: "stone", Building: "uranium_processing_works", Rule: 3},
		{Age: "modern_age", Material: "data", Building: "mainframe_hall", Rule: 5},
		{Age: "modern_age", Material: "steel", Building: "titanium_smelter", Rule: 3},
		{Age: "information_age", Material: "steel", Building: "aerospace_foundry", Rule: 3},
		{Age: "digital_age", Material: "steel", Building: "nano_alloy_plant", Rule: 3},
		{Age: "cyberpunk_age", Material: "steel", Building: "dark_matter_refinery", Rule: 3},
		{Age: "fusion_age", Material: "steel", Building: "exotic_matter_forge", Rule: 3},
		{Age: "space_age", Material: "steel", Building: "orbital_refinery", Rule: 3},
		{Age: "interstellar_age", Material: "plasma", Building: "warp_drive_plant", Rule: 3},
		{Age: "interstellar_age", Material: "titanium", Building: "antimatter_forge", Rule: 3},
		{Age: "galactic_age", Material: "titanium", Building: "stellar_metallurgy", Rule: 3},
		{Age: "quantum_age", Material: "dark_matter", Building: "reality_forge", Rule: 3},
	}
}

// MakerOpenHoles lists every hole the rules leave open.
func MakerOpenHoles() []MakerOpenHole {
	return []MakerOpenHole{
		{Age: "transcendent_age", Material: "quantum_flux", Why: "the last age has no makers of its own: its only building that a run can build without an optional tech is housing, so the Singularity Core's quantum flux has no lineage to fill it."},
	}
}

// makerSkipped are the materials the maker rule leaves alone.
func makerSkipped(res string) bool { return res == "knowledge" || flowResources[res] }

// MakerExcluded is the set of (building, material) pairs MakerFills put into
// the tables, keyed "building/material": the tables as they stood before the
// fills read without them.
func MakerExcluded() map[string]bool {
	out := map[string]bool{}
	for _, f := range MakerFills() {
		out[f.Building+"/"+f.Material] = true
	}
	return out
}

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

// MakerAsks is what the gate out of each age asks of each material the age
// prices, in units of that material's level: every building the gate
// requires, every required copy at the price it is bought at. The wonder's
// asks come back separately. Keyed by the age the gate leaves.
func MakerAsks(defs []BuildingDef, ages []AgeDef) (gate, wonder map[string]map[string]float64) {
	levels := priceLevels(defs)
	byKey := make(map[string]BuildingDef, len(defs))
	for _, d := range defs {
		byKey[d.Key] = d
	}
	gate = map[string]map[string]float64{}
	wonder = map[string]map[string]float64{}
	for i, a := range ages {
		lv := levels[a.Key]
		gate[a.Key] = map[string]float64{}
		wonder[a.Key] = map[string]float64{}
		if i+1 < len(ages) {
			keys := make([]string, 0, len(ages[i+1].BuildingReqs))
			for k := range ages[i+1].BuildingReqs {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				d := byKey[k]
				for res, c := range d.BaseCost {
					if lv[res] <= 0 || c <= 0 {
						continue
					}
					total, copyPrice := 0.0, c
					for n := 0; n < ages[i+1].BuildingReqs[k]; n++ {
						total += copyPrice
						copyPrice = float64(copyPrice * d.CostScale)
					}
					gate[a.Key][res] += float64(total / lv[res])
				}
			}
		}
		for _, d := range defs {
			if d.RequiredAge != a.Key || d.Category != "wonder" {
				continue
			}
			for res, c := range d.BaseCost {
				if lv[res] > 0 && c > 0 {
					wonder[a.Key][res] += float64(c / lv[res])
				}
			}
		}
	}
	return gate, wonder
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

// UnstaffedShare is what a building with worker slots makes with nobody in it,
// as a share of its listed output; workers add StaffedShare at a full crew.
// The game's own rates read these (game/buildings.go).
const (
	UnstaffedShare = 0.20
	StaffedShare   = 0.80
)

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
	byKey := make(map[string]BuildingDef, len(defs))
	for _, d := range defs {
		byKey[d.Key] = d
	}
	skip := map[string]bool{}
	for _, f := range fills {
		if f.Rule < 5 { // a placed maker is a maker that is there
			skip[f.Building+"/"+f.Material] = true
		}
	}
	var out []MakerHole
	for i, a := range ages {
		lv := levels[a.Key]
		buildable := func(d BuildingDef) bool { return d.RequiredTech == "" || held[i][d.RequiredTech] }
		makes := func(d BuildingDef, res string) bool {
			if skip[d.Key+"/"+res] {
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
			if float64(supply(res)*AgeTargetTicks(a.Key)) >= ask {
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
	byKey := make(map[string]BuildingDef, len(defs))
	for _, d := range defs {
		byKey[d.Key] = d
	}
	carried := makerCarried(defs, ages, i)
	total := 0.0
	for _, k := range sortedBuildingKeys(carried) {
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
	made := float64(makerSupply(defs, ages, i, res) * AgeTargetTicks(age))
	if ask <= 0 {
		return 0
	}
	if made <= 0 {
		return math.Inf(1)
	}
	return ask / made
}

// MakerShortage is a shortage the owner accepts: the carried required copies,
// at listed output, need Targets targets of Age to make what the gate and the
// wonder ask of Material, and the player covers the rest with a few more of the
// older buildings. A test holds each to the lingering reference player's stay
// (smoke) and to the tables (config).
type MakerShortage struct {
	Age, Material string
	Targets       float64
}

// MakerAcceptedShortages lists them. The hole test treats a listed pair as not
// a hole, so no building is made to carry a material it does not read as
// making (a Coal Works making wood, a Power Generator making coal).
func MakerAcceptedShortages() []MakerShortage {
	return []MakerShortage{
		{Age: "colonial_age", Material: "wood", Targets: 2.16},
		{Age: "electric_age", Material: "coal", Targets: 2.51},
	}
}

func makerAccepted(age, res string) bool {
	for _, a := range MakerAcceptedShortages() {
		if a.Age == age && a.Material == res {
			return true
		}
	}
	return false
}

func sortedBuildingKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func ageKeys(ages []AgeDef) []string {
	out := make([]string, len(ages))
	for i, a := range ages {
		out[i] = a.Key
	}
	return out
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

// MakerStandard is what the maker rule gives every building that a fill
// touches, as building -> material -> output per tick: the building's standard
// (its first-copy price in units over payback(A)) split in proportion to what
// the age's gate asks of the priced materials the building makes, each turned
// back into the material by its level. Unrounded. paybacks is payback(A) per
// age (MakerPaybackTicks).
func MakerStandard(defs []BuildingDef, ages []AgeDef, paybacks map[string]float64, fills []MakerFill) map[string]map[string]float64 {
	levels := priceLevels(defs)
	gate, _ := MakerAsks(defs, ages)
	byKey := make(map[string]BuildingDef, len(defs))
	for _, d := range defs {
		byKey[d.Key] = d
	}
	touched := map[string]bool{}
	for _, f := range fills {
		touched[f.Building] = true
	}
	out := map[string]map[string]float64{}
	for key := range touched {
		d := byKey[key]
		lv := levels[d.RequiredAge]
		price := priceUnits(d.BaseCost, lv)
		var mats []string
		weight := map[string]float64{}
		share := 0.0
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 && lv[e.Target] > 0 {
				mats = append(mats, e.Target)
				// What the gate asks of it; where the gate asks nothing of a
				// priced material (the Colonial Age's gate asks no iron) one
				// unit of it, a single median price, so that a raw form's
				// maker still makes it.
				weight[e.Target] = gate[d.RequiredAge][e.Target]
				if weight[e.Target] <= 0 {
					weight[e.Target] = 1
				}
				share += weight[e.Target]
			}
		}
		out[key] = map[string]float64{}
		for _, res := range mats {
			units := float64(float64(price/paybacks[d.RequiredAge]) * weight[res] / share)
			out[key][res] = float64(units * lv[res])
		}
	}
	return out
}
