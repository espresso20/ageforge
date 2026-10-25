package config

import (
	"maps"
	"slices"
)

// The maker rule: with no market swap, every material a town is charged for
// has to come from a building, at the same standard as every other producer.
// The tables are authored (the building files, MakerPaybackTicks and
// MakerFills below); this file is the derivation the cold start reads
// (MakerStandard), and a test recomputes the standard, the holes and every
// filled output from it (maker_audit_test.go holds the audit: MakerHoles,
// MakerPaybacks and what they call).
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

// MakerPaybackTicks is payback(A) for every age, in ticks: the median payback
// of the age's healthy makers, typed here and recomputed by a test. It is what
// the old payback rule set the ages' buildings by (DealPaybackTicks still
// holds that rule's figures, for the flow resources). Outputs did not move when
// new copies took NewCopyPriceFactor, so every payback is that much longer: the
// figures below are the ones from before times the factor, which is also what
// the median of the healthy makers comes to now.
var MakerPaybackTicks = map[string]float64{
	"primitive_age":    28.12 * NewCopyPriceFactor,
	"stone_age":        109.38 * NewCopyPriceFactor,
	"bronze_age":       764.57 * NewCopyPriceFactor,
	"iron_age":         1365.7 * NewCopyPriceFactor,
	"classical_age":    2193.1 * NewCopyPriceFactor,
	"medieval_age":     3183.9 * NewCopyPriceFactor,
	"renaissance_age":  9438.1 * NewCopyPriceFactor,
	"colonial_age":     6056.4 * NewCopyPriceFactor,
	"industrial_age":   8667.1 * NewCopyPriceFactor,
	"victorian_age":    14896 * NewCopyPriceFactor,
	"electric_age":     20116 * NewCopyPriceFactor,
	"atomic_age":       25407 * NewCopyPriceFactor,
	"modern_age":       25950 * NewCopyPriceFactor,
	"information_age":  15227 * LatePriceStep * NewCopyPriceFactor,
	"digital_age":      40929 * LatePriceStep * LatePriceStep * NewCopyPriceFactor,
	"cyberpunk_age":    50249 * LatePriceStep * LatePriceStep * LatePriceStep * NewCopyPriceFactor,
	"fusion_age":       129379 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * NewCopyPriceFactor,
	"space_age":        96530 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * NewCopyPriceFactor,
	"interstellar_age": 79991 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * NewCopyPriceFactor,
	"galactic_age":     169450 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * NewCopyPriceFactor,
	"quantum_age":      51348 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * NewCopyPriceFactor,
	"transcendent_age": 47204 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * NewCopyPriceFactor,
}

// MakerFill is one hole the rules fill: in Age, Building also makes Material,
// by Rule 1, 2 or 3. Rule 5 is a maker placed by the owner's decision where
// the rules found no building to fill a hole (the Modern Age's data): a new
// building, put on the standard like any maker; the rules read it as a maker
// that is there. Rule 6 is a maker that was already there, whose price was
// re-authored (the cold start's first makers, the Server Farm, the Black Market
// Hub) or whose age's price levels moved under it: its output is re-derived
// from its price, so it sits on the standard again; it is read as a maker that
// is there too. The building tables hold the output; MakerFills is the
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
		{Age: "industrial_age", Material: "iron_ore", Building: "steam_mine", Rule: 1},
		{Age: "victorian_age", Material: "iron", Building: "bessemer_plant", Rule: 3},
		{Age: "victorian_age", Material: "stone", Building: "uranium_mine", Rule: 3},
		{Age: "electric_age", Material: "electricity", Building: "power_generator", Rule: 6},
		{Age: "electric_age", Material: "iron", Building: "electric_arc_furnace", Rule: 3},
		{Age: "atomic_age", Material: "iron", Building: "advanced_alloy_plant", Rule: 3},
		{Age: "atomic_age", Material: "stone", Building: "uranium_processing_works", Rule: 3},
		{Age: "modern_age", Material: "data", Building: "mainframe_hall", Rule: 5},
		{Age: "modern_age", Material: "electricity", Building: "power_grid_hub", Rule: 6},
		{Age: "modern_age", Material: "oil", Building: "oil_platform", Rule: 6},
		{Age: "modern_age", Material: "oil", Building: "oil_refinery", Rule: 6},
		{Age: "modern_age", Material: "electricity", Building: "oil_refinery", Rule: 6},
		{Age: "modern_age", Material: "gold", Building: "seaport", Rule: 6},
		{Age: "modern_age", Material: "steel", Building: "titanium_smelter", Rule: 3},
		{Age: "modern_age", Material: "titanium_ore", Building: "titanium_mine", Rule: 1},
		{Age: "information_age", Material: "data", Building: "server_farm", Rule: 6},
		{Age: "information_age", Material: "steel", Building: "aerospace_foundry", Rule: 3},
		{Age: "information_age", Material: "titanium_ore", Building: "precision_mine", Rule: 1},
		{Age: "digital_age", Material: "steel", Building: "nano_alloy_plant", Rule: 3},
		{Age: "digital_age", Material: "titanium_ore", Building: "nano_drill_complex", Rule: 1},
		{Age: "cyberpunk_age", Material: "crypto", Building: "black_market", Rule: 6},
		{Age: "cyberpunk_age", Material: "steel", Building: "dark_matter_refinery", Rule: 3},
		{Age: "cyberpunk_age", Material: "dark_matter_crystals", Building: "dark_crystal_mine", Rule: 1},
		{Age: "fusion_age", Material: "steel", Building: "exotic_matter_forge", Rule: 3},
		{Age: "fusion_age", Material: "dark_matter_crystals", Building: "exotic_mineral_extractor", Rule: 1},
		{Age: "space_age", Material: "steel", Building: "orbital_refinery", Rule: 3},
		{Age: "space_age", Material: "dark_matter_crystals", Building: "asteroid_crystal_mine", Rule: 1},
		{Age: "interstellar_age", Material: "plasma", Building: "warp_drive_plant", Rule: 3},
		{Age: "interstellar_age", Material: "titanium", Building: "antimatter_forge", Rule: 3},
		{Age: "galactic_age", Material: "titanium", Building: "stellar_metallurgy", Rule: 3},
		{Age: "quantum_age", Material: "antimatter", Building: "reality_excavator", Rule: 6},
		{Age: "quantum_age", Material: "quantum_flux", Building: "reality_harvester", Rule: 6},
		{Age: "quantum_age", Material: "dark_matter", Building: "reality_forge", Rule: 3},
	}
}

// MakerOpenHoles lists every hole the rules leave open.
func MakerOpenHoles() []MakerOpenHole {
	return []MakerOpenHole{
		{Age: "transcendent_age", Material: "quantum_flux", Why: "the last age has no makers of its own: its only building that a run can build without an optional tech is housing, so the Singularity Core's quantum flux has no lineage to fill it."},
	}
}

// MakerAsks is what the gate out of each age asks of each material the age
// prices, in units of that material's level: every building the gate
// requires, every required copy at the price it is bought at. The wonder's
// asks come back separately. Keyed by the age the gate leaves.
func MakerAsks(defs []BuildingDef, ages []AgeDef) (gate, wonder map[string]map[string]float64) {
	levels := priceLevels(defs)
	byKey := indexBy(defs, func(d BuildingDef) string { return d.Key })
	gate = map[string]map[string]float64{}
	wonder = map[string]map[string]float64{}
	for i, a := range ages {
		lv := levels[a.Key]
		gate[a.Key] = map[string]float64{}
		wonder[a.Key] = map[string]float64{}
		if i+1 < len(ages) {
			for _, k := range slices.Sorted(maps.Keys(ages[i+1].BuildingReqs)) {
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

// RawMaterials are the four raw materials, in the order the tables list them.
// A raw is a material like wood or stone in the ages that make it: buildings
// of the age are priced in it, half of what they would cost in its worked form.
var RawMaterials = []string{"iron_ore", "marble", "titanium_ore", "dark_matter_crystals"}

// rawPartnerForms are, for each raw, the worked forms it can pair with, in
// order. In an age the raw's partner is the first of them that the age's
// buildings are priced in (a building keeps half of its partner part, so the
// answer is the same before any raw was priced). makerRaw is the first form
// of each.
var rawPartnerForms = map[string][]string{
	"iron_ore":             {"iron"},
	"marble":               {"stone"},
	"titanium_ore":         {"titanium", "steel"},
	"dark_matter_crystals": {"dark_matter", "electricity"},
}

// ageMakesRaw reports whether a building of age makes raw (a production
// effect with a positive value).
func ageMakesRaw(defs []BuildingDef, age, raw string) bool {
	for _, d := range defs {
		if d.RequiredAge != age {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 && e.Target == raw {
				return true
			}
		}
	}
	return false
}

// RawPartner is the worked form raw is priced alongside in age: the first of
// its forms that a (non-wonder) building of the age is priced in. It is "" in
// an age with no maker of the raw, where the raw is not priced and nothing
// pairs, and in an age that prices none of its forms.
func RawPartner(defs []BuildingDef, age, raw string) string {
	if !ageMakesRaw(defs, age, raw) {
		return ""
	}
	for _, form := range rawPartnerForms[raw] {
		for _, d := range defs {
			if d.RequiredAge == age && d.Category != "wonder" && d.BaseCost[form] > 0 {
				return form
			}
		}
	}
	return ""
}

// UnstaffedShare is what a building with worker slots makes with nobody in it,
// as a share of its listed output; workers add StaffedShare at a full crew.
// The game's own rates read these (game/buildings.go).
const (
	UnstaffedShare = 0.20
	StaffedShare   = 0.80
)

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
		{Age: "electric_age", Material: "coal", Targets: 3.608},
	}
}

func ageKeys(ages []AgeDef) []string {
	out := make([]string, len(ages))
	for i, a := range ages {
		out[i] = a.Key
	}
	return out
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
	byKey := indexBy(defs, func(d BuildingDef) string { return d.Key })
	touched := map[string]bool{}
	rederived := map[string]bool{}
	for _, f := range fills {
		touched[f.Building] = true
		if f.Rule == 6 {
			rederived[f.Building] = true
		}
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
				if weight[e.Target] <= 0 || rederived[key] {
					// (A maker that was already there keeps the split it was
					// written with, equal units: every dual maker in the
					// tables splits so.)
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
