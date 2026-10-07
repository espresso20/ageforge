package config

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/rivo/uniseg"
)

// TestTechTreeKeepsItsRules: the real tree breaks none of its rules. No
// cycle, no tech that needs a later age's tech, every lane, code and emblem
// in place and every code its own, keystones in their wonders' ages, no
// capstone on the spine.
func TestTechTreeKeepsItsRules(t *testing.T) {
	for _, problem := range TechTreeProblems(Technologies(), BaseBuildings(), AgeOrder()) {
		t.Error(problem)
	}
}

// TestEveryTechHasLaneCodeAndEmblem spells out what TechTreeProblems checks
// about a tech's place and art, and adds what it leaves to the terminal: an
// emblem is one cell wide.
func TestEveryTechHasLaneCodeAndEmblem(t *testing.T) {
	lanes := TechLaneByKey()
	perLane := map[string]int{}
	codes := map[string]string{}
	for _, tech := range Technologies() {
		if _, ok := lanes[tech.Lane]; !ok {
			t.Errorf("%s: lane %q is not a lane", tech.Key, tech.Lane)
		}
		perLane[tech.Lane]++
		if !validTechCode(tech.Code) {
			t.Errorf("%s: code %q is not 2 to %d capital letters", tech.Key, tech.Code, TechCodeMax)
		}
		if other, taken := codes[tech.Code]; taken {
			t.Errorf("%s and %s share the code %s", other, tech.Key, tech.Code)
		}
		codes[tech.Code] = tech.Key
		if n := len([]rune(tech.Emblem)); n != 1 || uniseg.StringWidth(tech.Emblem) != 1 {
			t.Errorf("%s: emblem %q is %d glyphs, %d cells wide; want one glyph, one cell", tech.Key, tech.Emblem, n, uniseg.StringWidth(tech.Emblem))
		}
	}
	var split []string
	for _, l := range TechLanes() {
		if perLane[l.Key] == 0 {
			t.Errorf("lane %s holds no tech", l.Key)
		}
		split = append(split, l.Key+" "+strings.Repeat("#", perLane[l.Key]))
	}
	t.Logf("%d techs by lane:\n  %s", len(codes), strings.Join(split, "\n  "))
}

// TestTechLanes: ten lanes, each with its own key, name and one-cell glyph.
func TestTechLanes(t *testing.T) {
	lanes := TechLanes()
	if len(lanes) != 10 {
		t.Errorf("%d lanes, want 10", len(lanes))
	}
	seen := map[string]bool{}
	for _, l := range lanes {
		for _, v := range []string{"key " + l.Key, "name " + l.Name, "emblem " + l.Emblem} {
			if seen[v] {
				t.Errorf("two lanes share the %s", v)
			}
			seen[v] = true
		}
		if l.Key == "" || l.Name == "" {
			t.Errorf("a lane is missing its key or name: %+v", l)
		}
		if len([]rune(l.Emblem)) != 1 || uniseg.StringWidth(l.Emblem) != 1 {
			t.Errorf("lane %s: emblem %q is not one glyph, one cell wide", l.Key, l.Emblem)
		}
	}
}

// TestTechCodeFor: the rule that makes a code from a name.
func TestTechCodeFor(t *testing.T) {
	for name, want := range map[string]string{
		"Printing Press":       "PRINT",
		"Iron Smelting":        "IRON",
		"The Wheel":            "WHEEL",
		"Zero-G Manufacturing": "ZERO",
		"Self-Replication":     "SELF",
		"Radio":                "RADIO",
		"Internet of Things":   "INTER",
		"A.I. Research":        "AI",
		"":                     "",
		"The":                  "",
	} {
		if got := TechCodeFor(name); got != want {
			t.Errorf("TechCodeFor(%q) = %q, want %q", name, got, want)
		}
	}
	// A tech that sets no code or emblem gets the rule's code and its
	// lane's glyph; one that sets them keeps them.
	got := fillTechArt([]TechDef{
		{Key: "a", Name: "Crop Rotation", Lane: LaneAgriculture},
		{Key: "b", Name: "Crop Rotation", Lane: LaneAgriculture, Code: "ROTA", Emblem: "≈"},
	})
	if got[0].Code != "CROP" || got[0].Emblem != "♣" {
		t.Errorf("a tech with no art got code %q and emblem %q, want CROP and ♣", got[0].Code, got[0].Emblem)
	}
	if got[1].Code != "ROTA" || got[1].Emblem != "≈" {
		t.Errorf("a tech with its own art got code %q and emblem %q, want ROTA and ≈", got[1].Code, got[1].Emblem)
	}
}

// treeFixture is a small tree with everything the real one will have once
// wonders require techs: two wonders with a keystone each, a chain under
// each keystone, an either-or group on the way to the second, a capstone on
// top, and techs off to the side.
//
//	a1:  root ── mid          side
//	a2:  key_one*   left   right(needs side)   opt_gate
//	a3:  join (key_one, and left or right) ── key_two* ── cap (capstone, also needs side)
func treeFixture() (techs []TechDef, buildings []BuildingDef, ages []string) {
	ages = []string{"a1", "a2", "a3"}
	techs = fillTechArt([]TechDef{
		{Key: "root", Name: "Root", Age: "a1", Lane: LaneCraft},
		{Key: "mid", Name: "Mid", Age: "a1", Lane: LaneCraft, Prerequisites: []string{"root"}},
		{Key: "side", Name: "Side", Age: "a1", Lane: LaneTrade, Prerequisites: []string{"root"}},
		{Key: "key_one", Name: "Keyone", Age: "a2", Lane: LaneCraft, Prerequisites: []string{"mid"}},
		{Key: "left", Name: "Left", Age: "a2", Lane: LaneKnowledge},
		{Key: "right", Name: "Right", Age: "a2", Lane: LaneTrade, Prerequisites: []string{"side"}},
		{Key: "opt_gate", Name: "Gate", Age: "a2", Lane: LaneEnergy},
		{Key: "join", Name: "Join", Age: "a3", Lane: LaneTrade, Prerequisites: []string{"key_one"}, AnyOf: []string{"left", "right"}},
		{Key: "key_two", Name: "Keytwo", Age: "a3", Lane: LaneTrade, Prerequisites: []string{"join"}},
		{Key: "cap", Name: "Cap", Age: "a3", Lane: LaneTrade, Capstone: true, Prerequisites: []string{"key_two", "side"}},
	})
	buildings = []BuildingDef{
		{Key: "wonder_one", Category: "wonder", RequiredAge: "a2", RequiredTech: "key_one"},
		{Key: "wonder_two", Category: "wonder", RequiredAge: "a3", RequiredTech: "key_two"},
		// A tech that opens an ordinary building is not a keystone.
		{Key: "mill", Category: "production", RequiredAge: "a2", RequiredTech: "opt_gate"},
	}
	return techs, buildings, ages
}

// TestTechKindsAreDerived: nothing but the capstone flag is written down. A
// keystone is a wonder's tech, the spine is all a keystone stands on, and an
// either-or group pulls neither of its members onto the spine.
func TestTechKindsAreDerived(t *testing.T) {
	techs, buildings, ages := treeFixture()
	if problems := TechTreeProblems(techs, buildings, ages); len(problems) != 0 {
		t.Fatalf("the fixture breaks the rules it is there to show:\n  %s", strings.Join(problems, "\n  "))
	}
	got := TechKinds(techs, buildings)
	want := map[string]TechKind{
		"root": TechSpine, "mid": TechSpine, "key_one": TechKeystone,
		"join": TechSpine, "key_two": TechKeystone,
		"left": TechOptional, "right": TechOptional, "side": TechOptional, "opt_gate": TechOptional,
		"cap": TechCapstone,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("kinds:\n got %v\nwant %v", got, want)
	}

	// The spine only depends on the spine: everything a keystone or a
	// spine tech lists in Prerequisites is a keystone or a spine tech.
	required := func(key string) bool { return got[key] == TechKeystone || got[key] == TechSpine }
	for _, tech := range techs {
		if !required(tech.Key) {
			continue
		}
		for _, p := range tech.Prerequisites {
			if !required(p) {
				t.Errorf("%s is %s but needs %s, which is %s", tech.Key, got[tech.Key], p, got[p])
			}
		}
	}

	// One more prerequisite on a spine tech pulls that tech onto the spine:
	// the spine is whatever the keystones stand on, however it changes.
	for i := range techs {
		if techs[i].Key == "mid" {
			techs[i].Prerequisites = []string{"root", "side"}
		}
	}
	if k := TechKinds(techs, buildings)["side"]; k != TechSpine {
		t.Errorf("side is %s once a spine tech needs it, want spine", k)
	}
}

// realKeystones is each wonder's keystone tech, as the tree's design names
// them. The Sacred Grove has none by design: nothing blocks the first age.
// Stonehenge has none yet: its keystone (Calendar) is a tech still to come.
var realKeystones = map[string]string{
	"sacred_grove":         "",
	"great_monolith":       "stoneworking",
	"stonehenge":           "calendar",
	"colosseum":            "mathematics",
	"parthenon":            "philosophy",
	"great_library":        "theology",
	"sistine_chapel":       "patronage",
	"grand_lighthouse":     "cartography",
	"crystal_palace":       "industrialization",
	"eiffel_tower":         "mass_production",
	"hoover_dam":           "power_distribution",
	"particle_accelerator": "nuclear_fission",
	"space_program":        "satellite_tech",
	"global_network":       "internet",
	"world_simulation":     "machine_learning",
	"neon_citadel":         "cybernetics",
	"stellar_cradle":       "fusion_power",
	"dyson_scaffold":       "orbital_mechanics",
	"warp_nexus":           "warp_drive",
	"cosmic_beacon":        "galactic_navigation",
	"reality_anchor":       "quantum_mechanics",
	"singularity_core":     "transcendence",
}

// TestRealTechKinds: on the real tables every tech has a kind, there is one
// keystone per tech a wonder requires, and the spine only depends on the
// spine. The counts are pinned: 21 keystones (every wonder but the Sacred
// Grove's) and the 26 techs they stand on, and the two capstones of the Iron
// Era.
func TestRealTechKinds(t *testing.T) {
	techs, buildings := Technologies(), BaseBuildings()
	kinds := TechKinds(techs, buildings)
	keystones := map[string]bool{}
	wonders := 0
	for _, b := range buildings {
		if b.Category != "wonder" {
			continue
		}
		wonders++
		want, listed := realKeystones[b.Key]
		if !listed || b.RequiredTech != want {
			t.Errorf("%s requires %q, want %q (listed %v)", b.Key, b.RequiredTech, want, listed)
		}
		if b.RequiredTech != "" {
			keystones[b.RequiredTech] = true
		}
	}
	if wonders != len(realKeystones) {
		t.Errorf("%d wonders, the keystone table lists %d", wonders, len(realKeystones))
	}
	count := map[TechKind]int{}
	for _, tech := range techs {
		k, ok := kinds[tech.Key]
		if !ok {
			t.Errorf("%s has no kind", tech.Key)
		}
		count[k]++
		if (k == TechKeystone) != keystones[tech.Key] {
			t.Errorf("%s is %s, but a wonder requires it: %v", tech.Key, k, keystones[tech.Key])
		}
		if k != TechKeystone && k != TechSpine {
			continue
		}
		for _, p := range tech.Prerequisites {
			if kinds[p] != TechKeystone && kinds[p] != TechSpine {
				t.Errorf("%s is %s but needs %s, which is %s", tech.Key, k, p, kinds[p])
			}
		}
	}
	if count[TechKeystone] != 21 || count[TechSpine] != 27 || count[TechCapstone] != 6 || count[TechOptional] != 74 {
		t.Errorf("kinds: %d keystone, %d spine, %d capstone, %d optional; want 21, 27, 6 and 74",
			count[TechKeystone], count[TechSpine], count[TechCapstone], count[TechOptional])
	}
}

// TestTechKindsNeedOnlyTheRawWonders: Technologies works the kinds out from
// the raw storage-and-wonder table, to stay off the building normalizers.
// That is only right while every wonder is defined there: the kinds it gives
// must be the kinds the whole building table gives.
func TestTechKindsNeedOnlyTheRawWonders(t *testing.T) {
	techs := Technologies()
	raw, full := TechKinds(techs, baseBuildingsRaw()), TechKinds(techs, BaseBuildings())
	if !reflect.DeepEqual(raw, full) {
		t.Errorf("kinds from the raw table differ from kinds from BaseBuildings:\n raw  %v\n full %v", raw, full)
	}
}

// TestTechTreeProblemsCatchEachRule breaks the fixture one rule at a time
// and expects the check to say so.
func TestTechTreeProblemsCatchEachRule(t *testing.T) {
	edit := func(key string, change func(*TechDef)) func([]TechDef, []BuildingDef) ([]TechDef, []BuildingDef) {
		return func(techs []TechDef, buildings []BuildingDef) ([]TechDef, []BuildingDef) {
			for i := range techs {
				if techs[i].Key == key {
					change(&techs[i])
				}
			}
			return techs, buildings
		}
	}
	cases := []struct {
		name  string
		wants string // what the problem must say
		edit  func([]TechDef, []BuildingDef) ([]TechDef, []BuildingDef)
	}{
		{"a cycle through prerequisites", "a cycle: root needs mid needs root",
			edit("root", func(d *TechDef) { d.Prerequisites = []string{"mid"} })},
		{"a cycle through an either-or group", "a cycle: left needs right needs left",
			// right is one of two ways to left, and right needs left. That
			// left could be reached through opt_gate instead does not
			// excuse the loop: an either-or key is an edge like any other.
			func(techs []TechDef, buildings []BuildingDef) ([]TechDef, []BuildingDef) {
				for i := range techs {
					switch techs[i].Key {
					case "left":
						techs[i].AnyOf = []string{"right", "opt_gate"}
					case "right":
						techs[i].Prerequisites = []string{"side", "left"}
					}
				}
				return techs, buildings
			}},
		{"a tech that needs itself", "side needs itself",
			edit("side", func(d *TechDef) { d.Prerequisites = []string{"side"} })},
		{"a prerequisite from a later age", "mid (a1) needs key_two, a tech of a later age (a3)",
			edit("mid", func(d *TechDef) { d.Prerequisites = []string{"root", "key_two"} })},
		{"an either-or key from a later age", "right (a2) needs cap, a tech of a later age (a3)",
			edit("right", func(d *TechDef) { d.AnyOf = []string{"left", "cap"} })},
		{"a prerequisite that is no tech", "side needs wheel, which is not a tech",
			edit("side", func(d *TechDef) { d.Prerequisites = []string{"wheel"} })},
		{"an either-or key that is no tech", "join needs boats, which is not a tech",
			edit("join", func(d *TechDef) { d.AnyOf = []string{"left", "boats"} })},
		{"a key in both lists", "join names left twice in what it needs",
			edit("join", func(d *TechDef) { d.Prerequisites = []string{"key_one", "left"} })},
		{"an either-or group of one", "join has an either-or group of one (left)",
			edit("join", func(d *TechDef) { d.AnyOf = []string{"left"} })},
		{"no lane", `opt_gate has no lane, or one that does not exist ("")`,
			edit("opt_gate", func(d *TechDef) { d.Lane = "" })},
		{"a lane that does not exist", `opt_gate has no lane, or one that does not exist ("diplomacy")`,
			edit("opt_gate", func(d *TechDef) { d.Lane = "diplomacy" })},
		{"no code", `side has the code ""`,
			edit("side", func(d *TechDef) { d.Code = "" })},
		{"a code too long", `side has the code "SIDEWAYS"`,
			edit("side", func(d *TechDef) { d.Code = "SIDEWAYS" })},
		{"a code in lower case", `side has the code "side"`,
			edit("side", func(d *TechDef) { d.Code = "side" })},
		{"two techs with one code", "root and side share the code ROOT",
			edit("side", func(d *TechDef) { d.Code = "ROOT" })},
		{"no emblem", `side has the emblem ""`,
			edit("side", func(d *TechDef) { d.Emblem = "" })},
		{"an emblem of two glyphs", `side has the emblem "$$"`,
			edit("side", func(d *TechDef) { d.Emblem = "$$" })},
		{"a capstone on the spine", "mid is flagged a capstone but a keystone stands on it",
			edit("mid", func(d *TechDef) { d.Capstone = true })},
		{"a capstone that is a keystone", "key_two is flagged a capstone but a wonder requires it",
			edit("key_two", func(d *TechDef) { d.Capstone = true })},
		{"a keystone outside its wonder's age", "key_one is the keystone of wonder_one (a3) but opens in a2",
			func(techs []TechDef, buildings []BuildingDef) ([]TechDef, []BuildingDef) {
				buildings[0].RequiredAge = "a3"
				return techs, buildings
			}},
	}
	for _, c := range cases {
		techs, buildings, ages := treeFixture()
		techs, buildings = c.edit(techs, buildings)
		problems := TechTreeProblems(techs, buildings, ages)
		found := false
		for _, p := range problems {
			found = found || strings.Contains(p, c.wants)
		}
		if !found {
			t.Errorf("%s: no problem says %q; got %q", c.name, c.wants, problems)
		}
		if !sort.StringsAreSorted(problems) {
			t.Errorf("%s: the problems are not sorted: %q", c.name, problems)
		}
	}
}

// TestPrereqsMet: every prerequisite, and one key of the either-or group.
func TestPrereqsMet(t *testing.T) {
	have := func(keys ...string) func(string) bool {
		return func(k string) bool {
			for _, h := range keys {
				if h == k {
					return true
				}
			}
			return false
		}
	}
	both := TechDef{Prerequisites: []string{"a", "b"}, AnyOf: []string{"x", "y"}}
	cases := []struct {
		def     TechDef
		have    []string
		met     bool
		missing string
		anyOf   bool
	}{
		{TechDef{}, nil, true, "", true},
		{TechDef{Prerequisites: []string{"a", "b"}}, []string{"a"}, false, "b", true},
		{TechDef{Prerequisites: []string{"a", "b"}}, []string{"b", "a"}, true, "", true},
		{TechDef{AnyOf: []string{"x", "y"}}, nil, false, "", false},
		{TechDef{AnyOf: []string{"x", "y"}}, []string{"y"}, true, "", true},
		{both, []string{"a", "b"}, false, "", false},
		{both, []string{"a", "x", "y"}, false, "b", true},
		{both, []string{"a", "b", "x"}, true, "", true},
	}
	for _, c := range cases {
		h := have(c.have...)
		if got := c.def.PrereqsMet(h); got != c.met {
			t.Errorf("%+v with %v: met %v, want %v", c.def, c.have, got, c.met)
		}
		if got := c.def.MissingPrereq(h); got != c.missing {
			t.Errorf("%+v with %v: missing %q, want %q", c.def, c.have, got, c.missing)
		}
		if got := c.def.AnyOfMet(h); got != c.anyOf {
			t.Errorf("%+v with %v: either-or met %v, want %v", c.def, c.have, got, c.anyOf)
		}
	}
}
