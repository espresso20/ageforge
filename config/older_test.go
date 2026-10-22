package config

import "testing"

func olderAgeIndex() map[string]int {
	ages := map[string]int{}
	for i, a := range AgeOrder() {
		ages[a] = i
	}
	return ages
}

// currentKey is the key of a line's current rung in age, "" when it has none.
func currentKey(rungs []BuildingDef, ages map[string]int, age string) string {
	if cur, ok := CurrentRung(rungs, ages, age); ok {
		return cur.Key
	}
	return ""
}

// TestCurrentRungIsTheHighestRungOfTheAge: for every line in every age the
// current rung is the highest rung that belongs to that age or an earlier one.
// Tiers climb with the ages and no two rungs share a tier, so the answer is
// one building.
func TestCurrentRungIsTheHighestRungOfTheAge(t *testing.T) {
	ages := olderAgeIndex()
	lines := LineRungs(BaseBuildings())
	if len(lines) != 38 {
		t.Fatalf("%d lines with rungs a building can be upgraded along, want 38 (39 lines, less knowledge)", len(lines))
	}
	for line, rungs := range lines {
		for i := 1; i < len(rungs); i++ {
			if rungs[i].LineageTier == rungs[i-1].LineageTier {
				t.Errorf("line %s: %s and %s share tier %d", line, rungs[i-1].Key, rungs[i].Key, rungs[i].LineageTier)
			}
			if ages[rungs[i].RequiredAge] < ages[rungs[i-1].RequiredAge] {
				t.Errorf("line %s: %s (%s) is above %s (%s) but belongs to an earlier age", line, rungs[i].Key, rungs[i].RequiredAge, rungs[i-1].Key, rungs[i-1].RequiredAge)
			}
		}
		for _, age := range AgeOrder() {
			want := ""
			for i := len(rungs) - 1; i >= 0; i-- {
				if ages[rungs[i].RequiredAge] <= ages[age] {
					want = rungs[i].Key
					break
				}
			}
			if got := currentKey(rungs, ages, age); got != want {
				t.Errorf("line %s in %s: current rung %q, want %q", line, age, got, want)
			}
		}
	}
}

// TestEveryOlderBuildingHasATargetInItsOwnLine: in every age, a building that
// belongs to it or an earlier age is older exactly when its line has a rung of
// a later age than its own that the age has, and the target is the current
// rung of its own line: a later rung, never one of another line, never a rung
// of a later age.
func TestEveryOlderBuildingHasATargetInItsOwnLine(t *testing.T) {
	ages := olderAgeIndex()
	lines := LineRungs(BaseBuildings())
	older := 0
	for _, age := range AgeOrder() {
		for line, rungs := range lines {
			cur, hasCur := CurrentRung(rungs, ages, age)
			for _, d := range rungs {
				if ages[d.RequiredAge] > ages[age] {
					continue // not unlocked yet; nobody holds it
				}
				if !hasCur {
					t.Fatalf("line %s in %s: %s is held but the line has no current rung", line, age, d.Key)
				}
				wantOlder := false
				for _, r := range rungs {
					if ages[r.RequiredAge] > ages[d.RequiredAge] && ages[r.RequiredAge] <= ages[age] {
						wantOlder = true
					}
				}
				to, isOlder := UpgradeTarget(rungs, ages, d, age)
				if isOlder != wantOlder {
					t.Errorf("%s in %s: older = %v, want %v", d.Key, age, isOlder, wantOlder)
				}
				if d.Key == cur.Key && isOlder {
					t.Errorf("%s is the current rung of %s in %s but older", d.Key, line, age)
				}
				if !isOlder {
					continue
				}
				older++
				if to.Key != cur.Key || to.Line() != d.Line() || to.LineageTier <= d.LineageTier || ages[to.RequiredAge] > ages[age] {
					t.Errorf("%s is older in %s but its target %s is not the current rung %s of line %s", d.Key, age, to.Key, cur.Key, line)
				}
			}
		}
	}
	if older == 0 {
		t.Fatal("no older building in any age")
	}
}

// TestOnlyLineBuildingsCanBeOlder: storage, wonders, monuments, knowledge
// buildings and the standalone buildings are in no line.
func TestOnlyLineBuildingsCanBeOlder(t *testing.T) {
	inLine := map[string]bool{}
	for _, rungs := range LineRungs(BaseBuildings()) {
		for _, d := range rungs {
			inLine[d.Key] = true
		}
	}
	for _, d := range BaseBuildings() {
		switch {
		case d.Category == "storage" || d.Category == "wonder" || d.Category == "monument" || d.Scholars > 0 || d.LineageKey == "":
			if inLine[d.Key] {
				t.Errorf("%s (%s, lineage %q) is in line %s", d.Key, d.Category, d.LineageKey, d.Line())
			}
		default:
			if !inLine[d.Key] {
				t.Errorf("%s (%s, lineage %q) is in no line", d.Key, d.Category, d.LineageKey)
			}
		}
	}
}

// TestUpgradeTargetCases pins the ones that read oddly.
func TestUpgradeTargetCases(t *testing.T) {
	ages := olderAgeIndex()
	lines := LineRungs(BaseBuildings())
	byKey := BuildingByKey()
	for _, c := range []struct{ key, age, want string }{
		{"hut", "primitive_age", ""},
		{"hut", "stone_age", "longhouse"},
		// Straight to the age's rung, not one rung at a time.
		{"hut", "iron_age", "townhouse"},
		{"gathering_camp", "bronze_age", "farm"},
		// Two rungs in one age are both new in it; the next age's rung is
		// the target of both.
		{"stone_camp", "stone_age", ""},
		{"stone_pit", "stone_age", ""},
		{"stone_camp", "bronze_age", "quarry"},
		{"hunting_lodge", "iron_age", ""},
		{"hunting_lodge", "classical_age", "military_academy"},
		// Paused: the current rung stays current and is not older.
		{"nano_alloy_plant", "fusion_age", ""},
		{"aerospace_foundry", "fusion_age", "nano_alloy_plant"},
		{"nano_alloy_plant", "space_age", "orbital_refinery"},
		// The last of a line is never older; storage and knowledge are in
		// no line.
		{"sawmill", "transcendent_age", ""},
		{"wood_workshop", "transcendent_age", "sawmill"},
		{"stash", "iron_age", ""},
		{"story_circle", "iron_age", ""},
	} {
		d, ok := byKey[c.key]
		if !ok {
			t.Fatalf("no building %q", c.key)
		}
		got := ""
		if to, older := UpgradeTarget(lines[d.Line()], ages, d, c.age); older {
			got = to.Key
		}
		if got != c.want {
			t.Errorf("%s in %s upgrades to %q, want %q", c.key, c.age, got, c.want)
		}
	}
}

// TestCurrentRungCases pins the ones that read oddly: a line that has not
// started, a line that pauses, two rungs in one age, a line of one rung.
func TestCurrentRungCases(t *testing.T) {
	ages := olderAgeIndex()
	lines := LineRungs(BaseBuildings())
	for _, c := range []struct{ line, age, want string }{
		{"housing", "primitive_age", "hut"},
		{"housing", "stone_age", "longhouse"},
		{"housing", "iron_age", "townhouse"},
		{"housing", "transcendent_age", "transcendent_nexus"},
		// Not started: the line's first rung comes later.
		{"harbor", "iron_age", ""},
		{"metallurgy_titanium", "industrial_age", ""},
		// Two rungs in one age: the higher one is current.
		{"geological_extraction_stone", "stone_age", "stone_pit"},
		{"military", "iron_age", "legion_fort"},
		// Paused: titanium has no rung in the Cyberpunk or Fusion Age.
		{"metallurgy_titanium", "digital_age", "nano_alloy_plant"},
		{"metallurgy_titanium", "cyberpunk_age", "nano_alloy_plant"},
		{"metallurgy_titanium", "fusion_age", "nano_alloy_plant"},
		{"metallurgy_titanium", "space_age", "orbital_refinery"},
		// It ended: the last rung stays current in every later age.
		{"organic_extraction_wood", "medieval_age", "sawmill"},
		{"organic_extraction_wood", "transcendent_age", "sawmill"},
		{"energy_coal", "transcendent_age", "coal_plant"},
	} {
		if got := currentKey(lines[c.line], ages, c.age); got != c.want {
			t.Errorf("line %s in %s: current rung %q, want %q", c.line, c.age, got, c.want)
		}
	}
}
