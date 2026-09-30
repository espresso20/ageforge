package mapmodel_test

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// TestUndiscoveredCivsAnonymous: a civ not yet met is only a key and a site
// slot in the model, so no style can print its name or traits.
func TestUndiscoveredCivsAnonymous(t *testing.T) {
	st := fixture.State(fixture.Options{Age: "bronze_age", Seed: 7})
	met := 0
	for k, f := range st.Diplomacy.Factions {
		f.Discovered = met == 0
		if f.Discovered {
			met++
		}
		st.Diplomacy.Factions[k] = f
	}
	m := build(t, st, nil)
	if len(m.Factions) != len(st.Diplomacy.Factions) {
		t.Fatalf("%d civs in the model, %d in the state", len(m.Factions), len(st.Diplomacy.Factions))
	}
	for _, f := range m.Factions {
		switch {
		case f.Discovered && f.Name == "":
			t.Errorf("%s is met but has no name", f.Key)
		case !f.Discovered && (f.Name != "" || f.Personality != "" || f.Specialty != "" || f.Strength != 0 || f.Opinion != 0):
			t.Errorf("%s is not met but the model knows %+v", f.Key, f)
		}
	}
	if m.DiscoveredCount() != 1 {
		t.Errorf("DiscoveredCount %d, want 1", m.DiscoveredCount())
	}
}
