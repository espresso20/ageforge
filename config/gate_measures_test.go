package config

import (
	"reflect"
	"testing"
)

func mk(key, age, cat string, cost map[string]float64, produces ...string) BuildingDef {
	d := BuildingDef{Key: key, RequiredAge: age, Category: cat, BaseCost: cost}
	for _, p := range produces {
		d.Effects = append(d.Effects, Effect{Type: "production", Target: p, Value: 1})
	}
	return d
}

// The main materials are the resources the age's makers produce that the
// age's prices use most, in price units, two at most; the wonder and the
// monuments do not count, and nor does another age's building.
func TestGateMainMaterialsRule(t *testing.T) {
	levels := map[string]float64{"wood": 10, "stone": 100, "iron": 1000, "gold": 10, "salt": 10}
	blds := []BuildingDef{
		mk("camp", "a", "production", map[string]float64{"wood": 30, "stone": 50}, "wood"),          // wood 3, stone 0.5
		mk("pit", "a", "production", map[string]float64{"stone": 100, "iron": 500}, "stone"),        // stone 1, iron 0.5
		mk("mine", "a", "production", map[string]float64{"iron": 2000, "wood": 10}, "iron", "gold"), // iron 2, wood 1
		mk("hall", "a", "housing", map[string]float64{"salt": 100}),                                 // salt 10 (nobody makes it)
		mk("wonder", "a", "wonder", map[string]float64{"gold": 1000}, "gold"),                       // not counted
		mk("obelisk", "a", "monument", map[string]float64{"gold": 1000}, "gold"),                    // not counted
		mk("later", "b", "production", map[string]float64{"gold": 1000}, "gold"),                    // another age
	}
	// weights: wood 3+1 = 4, stone 0.5+1 = 1.5, iron 0.5+2 = 2.5, gold 0 (no price in age a)
	got := GateMainMaterials("a", blds, levels)
	if want := []string{"wood", "iron"}; !reflect.DeepEqual(got, want) {
		t.Errorf("main materials = %v, want %v", got, want)
	}
	// A resource nobody makes is not a main material, however much prices use it.
	if got := GateMainMaterials("a", blds[3:4], levels); len(got) != 0 {
		t.Errorf("a resource nobody makes was a main material: %v", got)
	}
	// Ties go to the key.
	tie := []BuildingDef{mk("x", "t", "production", map[string]float64{"wood": 10, "salt": 10}, "wood", "salt")}
	if got, want := GateMainMaterials("t", tie, levels), []string{"salt", "wood"}; !reflect.DeepEqual(got, want) {
		t.Errorf("tied materials = %v, want %v", got, want)
	}
}

// The key tech is, of the age's techs the keystone does not stand on, the one
// most of the age's buildings need; failing that, the one most techs build on;
// then the cheaper.
func TestGateKeyTechRule(t *testing.T) {
	techs := []TechDef{
		{Key: "root", Age: "old", Cost: 1},
		{Key: "base", Age: "a", Cost: 10, Prerequisites: []string{"root"}}, // the keystone stands on it
		{Key: "keystone", Age: "a", Cost: 20, Prerequisites: []string{"base"}},
		{Key: "two", Age: "a", Cost: 50},      // two buildings need it
		{Key: "one", Age: "a", Cost: 5},       // one building
		{Key: "none", Age: "a", Cost: 1},      // none
		{Key: "elsewhere", Age: "b", Cost: 1}, // another age
		{Key: "child1", Age: "b", Prerequisites: []string{"none"}},
	}
	bld := func(key, age, tech string) BuildingDef {
		return BuildingDef{Key: key, RequiredAge: age, RequiredTech: tech}
	}
	blds := []BuildingDef{
		bld("wonder", "a", "keystone"), bld("m1", "a", "two"), bld("m2", "a", "two"), bld("m3", "a", "one"),
		bld("other", "b", "none"), bld("free", "a", ""),
	}
	if got := GateKeyTech("a", "wonder", blds, techs); got != "two" {
		t.Errorf("key tech = %q, want two (the most buildings need it)", got)
	}
	// With no building needing a tech of the age, the tech most others stand
	// on: "none" has a dependent (child1), the rest have none; "base" is the
	// keystone's own and is left out.
	if got := GateKeyTech("a", "wonder", []BuildingDef{bld("wonder", "a", "keystone")}, techs); got != "none" {
		t.Errorf("key tech = %q, want none (the one a tech builds on)", got)
	}
	// With nothing to tell them apart, the cheaper tech, then the key.
	flat := []TechDef{{Key: "z", Age: "a", Cost: 3}, {Key: "y", Age: "a", Cost: 3}, {Key: "x", Age: "a", Cost: 4}}
	if got := GateKeyTech("a", "", nil, flat); got != "y" {
		t.Errorf("key tech = %q, want y (cheapest, then by key)", got)
	}
	if got := GateKeyTech("c", "", nil, flat); got != "" {
		t.Errorf("an age with no tech has key tech %q", got)
	}
}

// The table in this package is written by the reference model (package
// smoke holds it to the model); here it is held to the rules above.
func TestGateMeasureTableFollowsTheRules(t *testing.T) {
	all := BaseBuildings()
	techs := Technologies()
	levels := PriceLevelsByAge(all)
	ages := Ages()
	table := GateMeasures()
	if len(table) != len(ages)-1 {
		t.Fatalf("%d measures for %d ages", len(table), len(ages))
	}
	for i, m := range table {
		a := ages[i]
		if m.Age != a.Key {
			t.Fatalf("measure %d is for %s, want %s", i, m.Age, a.Key)
		}
		if got := GateMainMaterials(a.Key, all, levels[a.Key]); !reflect.DeepEqual(got, m.Materials) {
			t.Errorf("%s: materials %v, the rule says %v", a.Key, m.Materials, got)
		}
		wonder := ""
		for _, key := range a.UnlockBuildings {
			for _, d := range all {
				if d.Key == key && d.Category == "wonder" && wonder == "" {
					wonder = key
				}
			}
		}
		if got := GateKeyTech(a.Key, wonder, all, techs); got != m.KeyTech {
			t.Errorf("%s: key tech %s, the rule says %s", a.Key, m.KeyTech, got)
		}
	}
	// A copy of the table is the caller's.
	table[0].Materials[0] = "changed"
	if GateMeasures()[0].Materials[0] == "changed" {
		t.Error("GateMeasures hands out the table's own slice")
	}
}
