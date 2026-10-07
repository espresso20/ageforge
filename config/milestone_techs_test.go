package config

import "testing"

// techMilestoneAges is the age each tech-count milestone is first within
// reach in. They were set there when the tree held 77 techs, and they keep
// their place as it grows.
var techMilestoneAges = []struct{ key, age string }{
	{"tech_pioneer", "iron_age"},
	{"deep_thinker", "renaissance_age"},
	{"philosophes", "industrial_age"},
	{"renaissance_mind", "electric_age"},
	{"tech_master", "information_age"},
}

// TestTechCountMilestonesKeepTheirAge: each tech-count milestone asks for
// more techs than every age before its own holds, and no more than the tree
// holds by the end of that age: the count before it plus half of the age's
// own techs, rounded up. Adding techs to an age moves the line; this test
// then says what each threshold should be.
func TestTechCountMilestonesKeepTheirAge(t *testing.T) {
	order := AgeOrder()
	upTo := map[string]int{} // techs of the ages up to and including each age
	per := map[string]int{}
	for _, tech := range Technologies() {
		per[tech.Age]++
	}
	sum := 0
	before := map[string]int{}
	for _, a := range order {
		before[a] = sum
		sum += per[a]
		upTo[a] = sum
	}
	byKey := MilestoneByKey()
	for _, m := range techMilestoneAges {
		def, ok := byKey[m.key]
		if !ok {
			t.Errorf("no milestone %s", m.key)
			continue
		}
		want := before[m.age] + (per[m.age]+1)/2
		if def.MinTechCount != want {
			t.Errorf("%s asks for %d techs; to be first within reach in the %s (%d techs before it, %d by its end) it should ask for %d",
				m.key, def.MinTechCount, m.age, before[m.age], upTo[m.age], want)
		}
		if def.MinTechCount <= before[m.age] || def.MinTechCount > upTo[m.age] {
			t.Errorf("%s (%d techs) is not first within reach in the %s: %d techs come before it and %d by its end", m.key, def.MinTechCount, m.age, before[m.age], upTo[m.age])
		}
	}
}
