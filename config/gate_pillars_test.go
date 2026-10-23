package config

import (
	"slices"
	"testing"
)

// Every building any age's gate asks for lands in exactly one of the three
// pillars that take buildings, and no gate asks for a wonder by count (the
// age's wonder is the Monument pillar's row).
func TestEveryGateBuildingHasOnePillar(t *testing.T) {
	byKey := map[string]BuildingDef{}
	for _, b := range BaseBuildings() {
		byKey[b.Key] = b
	}
	pillars := GatePillars()
	taken := map[string]int{}
	asked := 0
	for _, age := range Ages() {
		for key := range age.BuildingReqs {
			def, ok := byKey[key]
			if !ok {
				t.Errorf("%s asks for %q, which is not a building", age.Key, key)
				continue
			}
			asked++
			got := GatePillar(def)
			n := 0
			for _, p := range pillars {
				if p == got {
					n++
				}
			}
			if n != 1 {
				t.Errorf("%s: %s (%s/%s/%s) is in %d pillars (%q)", age.Key, key, def.Category, def.LineageKey, def.WorkerDomain, n, got)
			}
			if got == PillarMonument {
				t.Errorf("%s asks for %s, which sorts to the Monument pillar", age.Key, key)
			}
			taken[got]++
		}
	}
	if asked == 0 {
		t.Fatal("no gate asks for a building")
	}
	for _, p := range []string{PillarKnowledge, PillarPeople, PillarIndustry} {
		if taken[p] == 0 {
			t.Errorf("no gate row lands in %s: the mapping no longer sorts the gates into three groups", p)
		}
	}
}

// The mapping reads category, lineage and domain: a few buildings pinned so
// a change to it is a decision.
func TestGatePillarSamples(t *testing.T) {
	byKey := map[string]BuildingDef{}
	for _, b := range BaseBuildings() {
		byKey[b.Key] = b
	}
	for key, want := range map[string]string{
		"story_circle":   PillarKnowledge,
		"scriptorium":    PillarKnowledge,
		"shrine":         PillarPeople,
		"temple":         PillarPeople,
		"cathedral":      PillarPeople,
		"hut":            PillarPeople,
		"longhouse":      PillarPeople,
		"gathering_camp": PillarPeople,
		"farm":           PillarPeople,
		"wood_camp":      PillarIndustry,
		"stash":          PillarIndustry,
		"war_camp":       PillarIndustry,
		"market":         PillarIndustry,
		"sacred_grove":   PillarMonument,
		"stonehenge":     PillarMonument,
	} {
		def, ok := byKey[key]
		if !ok {
			t.Errorf("no building %q", key)
			continue
		}
		if got := GatePillar(def); got != want {
			t.Errorf("%s is in %s, want %s", key, got, want)
		}
	}
	for _, b := range BaseBuildings() {
		if b.Category == "wonder" && GatePillar(b) != PillarMonument {
			t.Errorf("wonder %s is not in the Monument pillar", b.Key)
		}
	}
}

func TestPillarNames(t *testing.T) {
	want := []string{"Knowledge", "People", "Industry", "Monument"}
	var got []string
	for _, p := range GatePillars() {
		got = append(got, PillarName(p))
	}
	if !slices.Equal(got, want) {
		t.Errorf("pillar names %v, want %v", got, want)
	}
}
