package config

import (
	"strings"
	"testing"

	"github.com/rivo/uniseg"
)

// techRow is one tech as the tech tree's design draws it: its age, lane and
// kind and what it needs ("a|b" is an either-or group).
type techRow struct {
	key, age, lane string
	kind           TechKind
	needs          string
}

// The tree of the Stone and Iron Eras (the Primitive to Medieval Ages), as
// the tech tree's design draws it: each tech's age, lane, kind and what it
// needs ("a|b" is an either-or group). The first content batch filled these
// six ages in.
var earlyTree = []techRow{
	{"language", "primitive_age", LaneKnowledge, TechSpine, ""},
	{"fire_mastery", "primitive_age", LaneAgriculture, TechOptional, ""},
	{"tool_making", "primitive_age", LaneCraft, TechSpine, ""},

	{"ritual", "stone_age", LaneFaith, TechSpine, "language"},
	{"primitive_writing", "stone_age", LaneKnowledge, TechSpine, "language"},
	{"pottery", "stone_age", LaneTrade, TechOptional, "fire_mastery"},
	{"animal_husbandry", "stone_age", LaneAgriculture, TechOptional, "fire_mastery"},
	{"woodworking", "stone_age", LaneCraft, TechOptional, "tool_making"},
	{"stoneworking", "stone_age", LaneMaterials, TechKeystone, "tool_making"},

	{"calendar", "bronze_age", LaneFaith, TechKeystone, "ritual"},
	{"map_making", "bronze_age", LaneKnowledge, TechOptional, "primitive_writing"},
	{"currency", "bronze_age", LaneTrade, TechSpine, "primitive_writing"},
	{"boatbuilding", "bronze_age", LaneTrade, TechOptional, "woodworking"},
	{"agriculture", "bronze_age", LaneAgriculture, TechOptional, "animal_husbandry"},
	{"the_wheel", "bronze_age", LaneCraft, TechOptional, "woodworking"},
	{"masonry", "bronze_age", LaneCraft, TechOptional, "stoneworking"},
	{"bronze_working", "bronze_age", LaneMaterials, TechSpine, "stoneworking"},
	{"military_tactics", "bronze_age", LaneMilitary, TechOptional, "bronze_working"},

	{"priesthood", "iron_age", LaneFaith, TechOptional, "calendar"},
	{"mathematics", "iron_age", LaneKnowledge, TechKeystone, "primitive_writing"},
	{"exploration", "iron_age", LaneTrade, TechSpine, "map_making|boatbuilding"},
	{"irrigation", "iron_age", LaneAgriculture, TechOptional, "agriculture"},
	{"road_building", "iron_age", LaneCraft, TechOptional, "masonry the_wheel"},
	{"iron_smelting", "iron_age", LaneMaterials, TechSpine, "bronze_working"},
	{"siege_warfare", "iron_age", LaneMilitary, TechOptional, "military_tactics"},

	{"drama", "classical_age", LaneFaith, TechOptional, "priesthood"},
	{"philosophy", "classical_age", LaneKnowledge, TechKeystone, "mathematics"},
	{"envoys", "classical_age", LaneTrade, TechOptional, "exploration"},
	{"the_plough", "classical_age", LaneAgriculture, TechOptional, "irrigation"},
	{"civil_engineering", "classical_age", LaneCraft, TechOptional, "road_building"},
	{"metal_casting", "classical_age", LaneMaterials, TechOptional, "iron_smelting"},
	{"imperial_legions", "classical_age", LaneMilitary, TechOptional, "siege_warfare iron_smelting"},

	{"theology", "medieval_age", LaneFaith, TechKeystone, "philosophy"},
	{"alchemy", "medieval_age", LaneKnowledge, TechOptional, "philosophy"},
	{"scholasticism", "medieval_age", LaneKnowledge, TechCapstone, "alchemy theology"},
	{"banking", "medieval_age", LaneTrade, TechSpine, "currency mathematics"},
	{"feudalism", "medieval_age", LaneAgriculture, TechOptional, "the_plough"},
	{"chronometry", "medieval_age", LaneCraft, TechOptional, ""},
	{"guilds", "medieval_age", LaneCraft, TechCapstone, "civil_engineering metal_casting"},
	{"steel_forging", "medieval_age", LaneMaterials, TechSpine, "iron_smelting"},
	{"fortification", "medieval_age", LaneMilitary, TechOptional, "imperial_legions"},
}

// midTree is the same for the Steel and Electric Eras (the Renaissance to
// Atomic Ages), which the second content batch filled in.
var midTree = []techRow{
	{"patronage", "renaissance_age", LaneFaith, TechKeystone, "banking"},
	{"printing_press", "renaissance_age", LaneKnowledge, TechOptional, "alchemy theology"},
	{"navigation", "renaissance_age", LaneTrade, TechSpine, "exploration mathematics"},
	{"crop_rotation", "renaissance_age", LaneAgriculture, TechOptional, "feudalism"},
	{"architecture", "renaissance_age", LaneCraft, TechOptional, "civil_engineering"},
	{"blast_furnace", "renaissance_age", LaneMaterials, TechOptional, "steel_forging"},
	{"gunpowder", "renaissance_age", LaneMilitary, TechOptional, "alchemy siege_warfare"},

	{"baroque_arts", "colonial_age", LaneFaith, TechOptional, "patronage"},
	{"scientific_method", "colonial_age", LaneKnowledge, TechOptional, "printing_press"},
	{"cartography", "colonial_age", LaneTrade, TechKeystone, "navigation"},
	{"mercantilism", "colonial_age", LaneTrade, TechOptional, "banking navigation"},
	{"embassies", "colonial_age", LaneTrade, TechOptional, "envoys"},
	{"new_world_crops", "colonial_age", LaneAgriculture, TechOptional, "crop_rotation"},
	{"surveying", "colonial_age", LaneCraft, TechOptional, "architecture"},
	{"coke_smelting", "colonial_age", LaneMaterials, TechOptional, "blast_furnace"},
	{"colonialism", "colonial_age", LaneMilitary, TechOptional, "cartography gunpowder"},

	{"romanticism", "industrial_age", LaneFaith, TechOptional, "baroque_arts"},
	{"encyclopedia", "industrial_age", LaneKnowledge, TechOptional, "scientific_method"},
	{"railroads", "industrial_age", LaneTrade, TechOptional, "steam_power road_building"},
	{"geographic_societies", "industrial_age", LaneTrade, TechOptional, "cartography"},
	{"concert_of_nations", "industrial_age", LaneTrade, TechCapstone, "embassies geographic_societies"},
	{"seed_drill", "industrial_age", LaneAgriculture, TechOptional, "new_world_crops"},
	{"industrialization", "industrial_age", LaneCraft, TechKeystone, "steam_power"},
	{"clockwork_automation", "industrial_age", LaneCraft, TechOptional, "chronometry"},
	{"interchangeable_parts", "industrial_age", LaneCraft, TechCapstone, "industrialization clockwork_automation"},
	{"steam_pumps", "industrial_age", LaneMaterials, TechOptional, "coke_smelting"},
	{"rifling", "industrial_age", LaneMilitary, TechOptional, "gunpowder"},
	{"steam_power", "industrial_age", LaneEnergy, TechSpine, "steel_forging"},

	{"museums", "victorian_age", LaneFaith, TechOptional, "romanticism"},
	{"public_education", "victorian_age", LaneKnowledge, TechOptional, "encyclopedia"},
	{"telecommunications", "victorian_age", LaneTrade, TechOptional, "electrification"},
	{"sanitation", "victorian_age", LaneAgriculture, TechOptional, "seed_drill"},
	{"mass_production", "victorian_age", LaneCraft, TechKeystone, "industrialization"},
	{"geology", "victorian_age", LaneMaterials, TechOptional, "steam_pumps"},
	{"general_staff", "victorian_age", LaneMilitary, TechOptional, "rifling"},
	{"electrification", "victorian_age", LaneEnergy, TechSpine, "industrialization"},

	{"radio", "electric_age", LaneFaith, TechOptional, "telecommunications"},
	{"modern_physics", "electric_age", LaneKnowledge, TechOptional, "public_education"},
	{"wire_transfers", "electric_age", LaneTrade, TechOptional, "telecommunications"},
	{"fertilizers", "electric_age", LaneAgriculture, TechOptional, "sanitation"},
	{"assembly_line", "electric_age", LaneCraft, TechOptional, "mass_production"},
	{"chemical_engineering", "electric_age", LaneMaterials, TechSpine, "mass_production"},
	{"mechanized_warfare", "electric_age", LaneMilitary, TechOptional, "general_staff"},
	{"power_distribution", "electric_age", LaneEnergy, TechKeystone, "electrification"},
	{"aviation", "electric_age", LaneSpace, TechSpine, "mass_production"},

	{"cinema", "atomic_age", LaneFaith, TechOptional, "radio"},
	{"big_science", "atomic_age", LaneKnowledge, TechCapstone, "modern_physics"},
	{"corporations", "atomic_age", LaneTrade, TechOptional, "mercantilism wire_transfers"},
	{"green_revolution", "atomic_age", LaneAgriculture, TechOptional, "fertilizers"},
	{"prefabrication", "atomic_age", LaneCraft, TechOptional, "assembly_line"},
	{"plastics", "atomic_age", LaneMaterials, TechOptional, "chemical_engineering"},
	{"nuclear_deterrence", "atomic_age", LaneMilitary, TechOptional, "nuclear_fission rocketry"},
	{"military_industrial_complex", "atomic_age", LaneMilitary, TechCapstone, "mechanized_warfare nuclear_deterrence"},
	{"nuclear_fission", "atomic_age", LaneEnergy, TechKeystone, "power_distribution chemical_engineering"},
	{"civilian_reactors", "atomic_age", LaneEnergy, TechOptional, "nuclear_deterrence"},
	{"rocketry", "atomic_age", LaneSpace, TechSpine, "aviation"},
}

// lateTree is the same for the Digital, Neon and Cosmic Eras (the Modern to
// Transcendent Ages), which the third content batch filled in: the whole
// tree is pinned now.
var lateTree = []techRow{
	{"television", "modern_age", LaneFaith, TechOptional, "cinema"},
	{"information_theory", "modern_age", LaneKnowledge, TechOptional, "modern_physics"},
	{"containerization", "modern_age", LaneTrade, TechOptional, "corporations"},
	{"suburbs", "modern_age", LaneAgriculture, TechOptional, "green_revolution"},
	{"nanofabrication", "modern_age", LaneCraft, TechOptional, "computers"},
	{"titanium_alloys", "modern_age", LaneMaterials, TechOptional, "chemical_engineering"},
	{"special_forces", "modern_age", LaneMilitary, TechOptional, "nuclear_deterrence"},
	{"electricity_tech", "modern_age", LaneEnergy, TechSpine, "nuclear_fission"},
	{"satellite_tech", "modern_age", LaneSpace, TechKeystone, "rocketry electricity_tech"},
	{"computers", "modern_age", LaneComputing, TechSpine, "electricity_tech"},

	{"social_media", "information_age", LaneFaith, TechOptional, "internet"},
	{"search_engines", "information_age", LaneKnowledge, TechOptional, "internet"},
	{"e_commerce", "information_age", LaneTrade, TechOptional, "containerization internet"},
	{"medical_nanobots", "information_age", LaneAgriculture, TechOptional, "nanofabrication"},
	{"internet_of_things", "information_age", LaneAgriculture, TechOptional, "social_media cybersecurity medical_nanobots"},
	{"embedded_systems", "information_age", LaneCraft, TechOptional, "nanofabrication"},
	{"precision_mining", "information_age", LaneMaterials, TechOptional, "titanium_alloys"},
	{"cybersecurity", "information_age", LaneMilitary, TechSpine, "computers"},
	{"smart_grid", "information_age", LaneEnergy, TechOptional, "electricity_tech"},
	{"space_stations", "information_age", LaneSpace, TechOptional, "satellite_tech"},
	{"internet", "information_age", LaneComputing, TechKeystone, "computers satellite_tech"},

	{"virtual_reality", "digital_age", LaneFaith, TechOptional, "social_media"},
	{"open_science", "digital_age", LaneKnowledge, TechOptional, "search_engines"},
	{"automated_logistics", "digital_age", LaneTrade, TechOptional, "e_commerce"},
	{"global_village", "digital_age", LaneTrade, TechCapstone, "e_commerce social_media"},
	{"gene_editing", "digital_age", LaneAgriculture, TechOptional, "medical_nanobots"},
	{"self_replication", "digital_age", LaneCraft, TechOptional, "medical_nanobots machine_learning"},
	{"nano_alloys", "digital_age", LaneMaterials, TechOptional, "precision_mining"},
	{"drone_warfare", "digital_age", LaneMilitary, TechOptional, "cybersecurity"},
	{"grid_storage", "digital_age", LaneEnergy, TechOptional, "smart_grid"},
	{"reusable_launchers", "digital_age", LaneSpace, TechOptional, "space_stations"},
	{"machine_learning", "digital_age", LaneComputing, TechKeystone, "internet cybersecurity"},
	{"cloud_computing", "digital_age", LaneComputing, TechOptional, "internet"},
	{"general_ai", "digital_age", LaneComputing, TechCapstone, "machine_learning cloud_computing"},

	{"holography", "cyberpunk_age", LaneFaith, TechOptional, "cybernetics blockchain"},
	{"neural_interface", "cyberpunk_age", LaneKnowledge, TechSpine, "machine_learning"},
	{"blockchain", "cyberpunk_age", LaneTrade, TechOptional, "cybersecurity cloud_computing"},
	{"synthetic_food", "cyberpunk_age", LaneAgriculture, TechOptional, "gene_editing"},
	{"cybernetics", "cyberpunk_age", LaneCraft, TechKeystone, "neural_interface"},
	{"dark_crystal_mining", "cyberpunk_age", LaneMaterials, TechOptional, "nano_alloys"},
	{"augmented_soldiers", "cyberpunk_age", LaneMilitary, TechOptional, "cybernetics drone_warfare"},
	{"dark_energy", "cyberpunk_age", LaneEnergy, TechOptional, "grid_storage"},
	{"lunar_outposts", "cyberpunk_age", LaneSpace, TechOptional, "reusable_launchers"},
	{"darknets", "cyberpunk_age", LaneComputing, TechOptional, "cloud_computing"},

	{"neural_art", "fusion_age", LaneFaith, TechOptional, "holography"},
	{"unified_theory", "fusion_age", LaneKnowledge, TechOptional, "neural_interface plasma_physics"},
	{"maglev_transit", "fusion_age", LaneTrade, TechOptional, "superconductors"},
	{"closed_biospheres", "fusion_age", LaneAgriculture, TechOptional, "synthetic_food"},
	{"molecular_assembly", "fusion_age", LaneCraft, TechOptional, "cybernetics self_replication"},
	{"superconductors", "fusion_age", LaneMaterials, TechSpine, "plasma_physics"},
	{"plasma_weapons", "fusion_age", LaneMilitary, TechOptional, "augmented_soldiers plasma_physics"},
	{"fusion_power", "fusion_age", LaneEnergy, TechKeystone, "nuclear_fission cybernetics"},
	{"plasma_physics", "fusion_age", LaneEnergy, TechSpine, "fusion_power"},
	{"fusion_drives", "fusion_age", LaneSpace, TechOptional, "fusion_power lunar_outposts"},
	{"quantum_networking", "fusion_age", LaneComputing, TechOptional, "darknets"},

	{"overview_effect", "space_age", LaneFaith, TechOptional, "neural_art"},
	{"deep_space_astronomy", "space_age", LaneKnowledge, TechOptional, "unified_theory"},
	{"asteroid_claims", "space_age", LaneTrade, TechOptional, "maglev_transit"},
	{"stellar_cartography", "space_age", LaneTrade, TechCapstone, "asteroid_claims deep_space_astronomy"},
	{"hydroponics", "space_age", LaneAgriculture, TechOptional, "closed_biospheres"},
	{"zero_g_manufacturing", "space_age", LaneCraft, TechSpine, "orbital_mechanics superconductors"},
	{"asteroid_refining", "space_age", LaneMaterials, TechOptional, "space_mining"},
	{"orbital_defense", "space_age", LaneMilitary, TechOptional, "plasma_weapons orbital_mechanics"},
	{"orbital_solar", "space_age", LaneEnergy, TechOptional, "plasma_physics"},
	{"orbital_mechanics", "space_age", LaneSpace, TechKeystone, "rocketry plasma_physics"},
	{"space_mining", "space_age", LaneSpace, TechSpine, "orbital_mechanics"},
	{"space_elevator", "space_age", LaneSpace, TechCapstone, "zero_g_manufacturing space_mining"},
	{"orbital_relays", "space_age", LaneComputing, TechOptional, "quantum_networking orbital_mechanics"},

	{"void_contemplation", "interstellar_age", LaneFaith, TechOptional, "overview_effect"},
	{"xenology", "interstellar_age", LaneKnowledge, TechOptional, "deep_space_astronomy"},
	{"interstellar_trade", "interstellar_age", LaneTrade, TechOptional, "asteroid_claims warp_drive"},
	{"protein_synthesis", "interstellar_age", LaneAgriculture, TechOptional, "hydroponics"},
	{"hull_printing", "interstellar_age", LaneCraft, TechOptional, "zero_g_manufacturing"},
	{"stellar_core_mining", "interstellar_age", LaneMaterials, TechOptional, "asteroid_refining"},
	{"fleet_doctrine", "interstellar_age", LaneMilitary, TechOptional, "orbital_defense"},
	{"stellar_engineering", "interstellar_age", LaneEnergy, TechSpine, "warp_drive"},
	{"warp_drive", "interstellar_age", LaneSpace, TechKeystone, "space_mining zero_g_manufacturing"},
	{"galactic_network", "interstellar_age", LaneComputing, TechOptional, "orbital_relays"},

	{"galactic_memory", "galactic_age", LaneFaith, TechOptional, "void_contemplation"},
	{"cosmology", "galactic_age", LaneKnowledge, TechOptional, "xenology"},
	{"federation_charter", "galactic_age", LaneTrade, TechOptional, "interstellar_trade xenology"},
	{"matter_conversion", "galactic_age", LaneAgriculture, TechOptional, "protein_synthesis"},
	{"dyson_engineering", "galactic_age", LaneCraft, TechOptional, "stellar_engineering hull_printing"},
	{"neutron_mining", "galactic_age", LaneMaterials, TechOptional, "stellar_core_mining"},
	{"armada_command", "galactic_age", LaneMilitary, TechOptional, "fleet_doctrine"},
	{"antimatter_synthesis", "galactic_age", LaneEnergy, TechSpine, "galactic_navigation"},
	{"galactic_navigation", "galactic_age", LaneSpace, TechKeystone, "warp_drive stellar_engineering"},
	{"terraforming", "galactic_age", LaneSpace, TechOptional, "warp_drive"},
	{"mind_uploading", "galactic_age", LaneComputing, TechOptional, "galactic_network"},

	{"reality_art", "quantum_age", LaneFaith, TechOptional, "galactic_memory"},
	{"quantum_mechanics", "quantum_age", LaneKnowledge, TechKeystone, "antimatter_synthesis"},
	{"timeless_archive", "quantum_age", LaneKnowledge, TechCapstone, "cosmology quantum_mechanics"},
	{"probability_markets", "quantum_age", LaneTrade, TechOptional, "federation_charter"},
	{"quantum_cultivation", "quantum_age", LaneAgriculture, TechOptional, "matter_conversion"},
	{"reality_manipulation", "quantum_age", LaneCraft, TechSpine, "quantum_mechanics"},
	{"reality_engineering", "quantum_age", LaneCraft, TechCapstone, "reality_manipulation dyson_engineering"},
	{"quantum_metallurgy", "quantum_age", LaneMaterials, TechOptional, "neutron_mining"},
	{"probability_warfare", "quantum_age", LaneMilitary, TechOptional, "armada_command"},
	{"zero_point_energy", "quantum_age", LaneEnergy, TechOptional, "antimatter_synthesis"},
	{"wormholes", "quantum_age", LaneSpace, TechOptional, "galactic_navigation"},
	{"quantum_computing", "quantum_age", LaneComputing, TechOptional, "clockwork_automation quantum_mechanics"},

	{"transcendence", "transcendent_age", LaneFaith, TechKeystone, "reality_manipulation"},
	{"omniversal_exchange", "transcendent_age", LaneTrade, TechOptional, "probability_markets"},
	{"singularity_engineering", "transcendent_age", LaneCraft, TechOptional, "reality_manipulation"},
	{"omniversal_command", "transcendent_age", LaneMilitary, TechOptional, "probability_warfare"},
}

// designedTree is every age a content batch has filled in.
func designedTree() []techRow {
	return append(append(append([]techRow(nil), earlyTree...), midTree...), lateTree...)
}

// designedTechs is how many techs each filled-in age holds.
var designedTechs = map[string]int{
	"primitive_age": 3, "stone_age": 6, "bronze_age": 9, "iron_age": 7, "classical_age": 7, "medieval_age": 9,
	"renaissance_age": 7, "colonial_age": 9, "industrial_age": 12, "victorian_age": 8, "electric_age": 9, "atomic_age": 11,
	"modern_age": 10, "information_age": 11, "digital_age": 13, "cyberpunk_age": 10, "fusion_age": 11, "space_age": 13,
	"interstellar_age": 10, "galactic_age": 11, "quantum_age": 12, "transcendent_age": 4,
}

// TestTreeIsAsDesigned: every age holds exactly the techs of the design,
// each in its age and lane, of its kind, needing what the design says, and
// no lane of an age holds more than three.
func TestTreeIsAsDesigned(t *testing.T) {
	techs := Technologies()
	kinds := TechKinds(techs, BaseBuildings())
	byKey := TechByKey()
	listed := map[string]bool{}
	perAge := map[string]int{}
	perLane := map[string]int{}
	for _, row := range designedTree() {
		listed[row.key] = true
		perAge[row.age]++
		perLane[row.age+" "+row.lane]++
		def, ok := byKey[row.key]
		if !ok {
			t.Errorf("%s is not a tech", row.key)
			continue
		}
		needs := strings.Join(def.Prerequisites, " ")
		if len(def.AnyOf) > 0 {
			needs = strings.TrimSpace(needs + " " + strings.Join(def.AnyOf, "|"))
		}
		if def.Age != row.age || def.Lane != row.lane || kinds[row.key] != row.kind || needs != row.needs {
			t.Errorf("%s: %s, %s, %s, needs %q; want %s, %s, %s, needs %q",
				row.key, def.Age, def.Lane, kinds[row.key], needs, row.age, row.lane, row.kind, row.needs)
		}
	}
	for _, def := range techs {
		if _, designed := designedTechs[def.Age]; designed && !listed[def.Key] {
			t.Errorf("%s is a tech of the %s the design does not list", def.Key, def.Age)
		}
	}
	for age, want := range designedTechs {
		if perAge[age] != want {
			t.Errorf("%s holds %d techs, want %d", age, perAge[age], want)
		}
	}
	for laneOfAge, n := range perLane {
		if n > 3 {
			t.Errorf("%s holds %d techs: a lane takes three an age at most", laneOfAge, n)
		}
	}
	if len(techs) != 202 || len(designedTechs) != len(Ages()) {
		t.Errorf("the tree holds %d techs in %d pinned ages, want 202 in all %d: the 77 it had and the 17, 34 and 74 of the three content batches", len(techs), len(designedTechs), len(Ages()))
	}
}

// TestDesignedTechsCarryTheirOwnArt: every tech has an emblem and a letter
// code of its own, written on the tech and not borrowed
// from its lane or made from its name by default. An emblem is one glyph
// that takes exactly one cell, and no two techs of a lane share one; the
// code's first letter is the tech's letter in the plain glyph tier.
func TestDesignedTechsCarryTheirOwnArt(t *testing.T) {
	raw := map[string]TechDef{}
	for _, def := range rawTechnologies() {
		raw[def.Key] = def
	}
	inLane := map[string]map[string]string{}
	for _, row := range designedTree() {
		def := TechByKey()[row.key]
		if raw[row.key].Emblem == "" {
			t.Errorf("%s sets no emblem of its own: it would wear its lane's", row.key)
		}
		rs := []rune(def.Emblem)
		if len(rs) != 1 || uniseg.StringWidth(def.Emblem) != 1 {
			t.Errorf("%s has the emblem %q: %d glyphs, %d cells wide; an emblem is one glyph in one cell", row.key, def.Emblem, len(rs), uniseg.StringWidth(def.Emblem))
		}
		if len(rs) == 1 && (rs[0] >= 0x1F000 || rs[0] == 0xFE0F) {
			t.Errorf("%s has the emblem %q, from the emoji planes: terminals draw those two cells wide", row.key, def.Emblem)
		}
		if inLane[def.Lane] == nil {
			inLane[def.Lane] = map[string]string{}
		}
		if other, taken := inLane[def.Lane][def.Emblem]; taken {
			t.Errorf("%s and %s, both in the %s lane, share the emblem %q", other, row.key, def.Lane, def.Emblem)
		}
		inLane[def.Lane][def.Emblem] = row.key
		if !validTechCode(def.Code) || def.Code[0] < 'A' || def.Code[0] > 'Z' {
			t.Errorf("%s has the code %q: no letter for the plain glyph tier", row.key, def.Code)
		}
	}
}
