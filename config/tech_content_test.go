package config

import (
	"strings"
	"testing"

	"github.com/rivo/uniseg"
)

// The tree of the Stone and Iron Eras (the Primitive to Medieval Ages), as
// the tech tree's design draws it: each tech's age, lane, kind and what it
// needs ("a|b" is an either-or group). The first content batch filled these
// six ages in.
var earlyTree = []struct {
	key, age, lane string
	kind           TechKind
	needs          string
}{
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

// earlyAges are the six ages earlyTree covers.
var earlyAges = map[string]bool{
	"primitive_age": true, "stone_age": true, "bronze_age": true,
	"iron_age": true, "classical_age": true, "medieval_age": true,
}

// TestEarlyTreeIsAsDesigned: the Stone and Iron Eras hold exactly the techs
// of the design, each in its age and lane, of its kind, needing what the
// design says. Navigation, one age on, stands on Exploration.
func TestEarlyTreeIsAsDesigned(t *testing.T) {
	techs := Technologies()
	kinds := TechKinds(techs, BaseBuildings())
	byKey := TechByKey()
	listed := map[string]bool{}
	perAge := map[string]int{}
	for _, row := range earlyTree {
		listed[row.key] = true
		perAge[row.age]++
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
		if earlyAges[def.Age] && !listed[def.Key] {
			t.Errorf("%s is a tech of the %s the design does not list", def.Key, def.Age)
		}
	}
	for age, want := range map[string]int{"primitive_age": 3, "stone_age": 6, "bronze_age": 9, "iron_age": 7, "classical_age": 7, "medieval_age": 9} {
		if perAge[age] != want {
			t.Errorf("%s holds %d techs, want %d", age, perAge[age], want)
		}
	}
	if got := strings.Join(byKey["navigation"].Prerequisites, " "); got != "exploration mathematics" {
		t.Errorf("navigation needs %q, want Exploration and Mathematics", got)
	}
	if len(techs) != 94 {
		t.Errorf("the tree holds %d techs, want 94: the 77 it had and the 17 of the first content batch", len(techs))
	}
}

// TestEarlyTechsCarryTheirOwnArt: every tech of the Stone and Iron Eras has
// an emblem and a letter code of its own, written on the tech and not
// borrowed from its lane or made from its name by default. An emblem is one
// glyph that takes exactly one cell, and no two techs of a lane share one;
// the code's first letter is the tech's letter in the plain glyph tier.
func TestEarlyTechsCarryTheirOwnArt(t *testing.T) {
	raw := map[string]TechDef{}
	for _, def := range rawTechnologies() {
		raw[def.Key] = def
	}
	inLane := map[string]map[string]string{}
	for _, row := range earlyTree {
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
