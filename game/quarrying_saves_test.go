package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Saves from before Quarrying existed, loaded by the game after it. Quarrying
// is an optional tech that locks nothing, so the tree's version does not go
// up (config.TechTreeVersion rises only when the tree starts locking
// something): an old save gets no grace age and no Research update line. A
// save in the Iron Age or any later one simply sees one more tech it may
// research. The fixtures are the raw-materials saves, written and signed by
// the game as it was the day before Quarrying: an Iron Age town and a Modern
// Age town, both on the tree's current version.
func TestQuarryingSave_AnOldSaveSeesOneMoreOptionalTech(t *testing.T) {
	for _, c := range []struct {
		name, path, age string
	}{
		{"iron", rawIronFixture, "iron_age"},
		{"modern", rawModernFixture, "modern_age"},
	} {
		t.Run(c.name, func(t *testing.T) {
			isolateAccountDir(t)
			save := loadSaveFile(t, c.path)
			if save.Age != c.age || save.TreeVersion != config.TechTreeVersion {
				t.Fatalf("setup: the fixture is in %s on tree version %d, want %s on %d", save.Age, save.TreeVersion, c.age, config.TechTreeVersion)
			}
			had := map[string]bool{}
			for _, k := range save.Research.Researched {
				had[k] = true
			}
			if had["quarrying"] || !had["bronze_working"] {
				t.Fatalf("setup: the fixture has Quarrying %v and Bronze Working %v", had["quarrying"], had["bronze_working"])
			}

			ge := loadFixtureAway(t, c.path, "quarrying_"+c.name, 0)
			if ge.Buildings.graceAge != "" {
				t.Errorf("an old save got the grace age %q: Quarrying locks nothing", ge.Buildings.graceAge)
			}
			if n, line := treeNotices(ge); n != 0 {
				t.Errorf("an old save got %d research update lines: %q", n, line)
			}
			for k := range had {
				if !ge.Research.IsResearched(k) {
					t.Errorf("%s was researched in the save and is not after the load", k)
				}
			}
			ts := ge.GetState().Research.Techs["quarrying"]
			if ts.Researched || !ts.PrereqsMet {
				t.Fatalf("Quarrying after the load: researched %v, prerequisites met %v; want it open to research", ts.Researched, ts.PrereqsMet)
			}

			// It can be researched, and then it adds its two rows.
			marble, ore := ge.Research.OutputFactor("marble"), ge.Research.OutputFactor("iron_ore")
			if err := ge.StartResearch("quarrying"); err != nil {
				t.Fatalf("starting Quarrying: %v", err)
			}
			researchOut(ge)
			if !ge.Research.IsResearched("quarrying") {
				t.Fatal("Quarrying did not finish")
			}
			// Tech bonuses add: the factor rises by the 8% of the row.
			if got := ge.Research.OutputFactor("marble") - marble; math.Abs(got-0.08) > 1e-9 {
				t.Errorf("marble output factor rose by %v with Quarrying, want 0.08", got)
			}
			if got := ge.Research.OutputFactor("iron_ore") - ore; math.Abs(got-0.08) > 1e-9 {
				t.Errorf("iron ore output factor rose by %v with Quarrying, want 0.08", got)
			}

			// And it saves and loads as researched.
			again := saveAndLoad(t, ge, "quarrying_again_"+c.name)
			if !again.Research.IsResearched("quarrying") {
				t.Error("Quarrying is not researched after a save and a load")
			}
			if v, grace := savedTree(t, "quarrying_again_"+c.name); v != config.TechTreeVersion || grace != "" {
				t.Errorf("the re-saved game carries tree version %d and grace age %q, want %d and none", v, grace, config.TechTreeVersion)
			}
		})
	}
}
