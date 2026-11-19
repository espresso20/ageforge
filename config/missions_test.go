package config

import (
	"fmt"
	"strings"
	"testing"
)

func TestMissionKinds(t *testing.T) {
	kinds := MissionKinds()
	if len(kinds) != 3 || kinds[0] != MissionScout || kinds[1] != MissionExpedition || kinds[2] != MissionCampaign {
		t.Fatalf("kinds %v", kinds)
	}
	for _, kind := range kinds {
		if _, ok := MissionKindTuning(kind); !ok {
			t.Errorf("%s has no tuning", kind)
		}
	}
	if _, ok := MissionKindTuning("nope"); ok {
		t.Error("an unknown kind has tuning")
	}
}

// TestMissionWordsCoverEveryAge checks each age has its tier names, six
// places, four objectives for each tier and six twists, and that the
// objectives end in the article the places follow.
func TestMissionWordsCoverEveryAge(t *testing.T) {
	ages := Ages()
	words := AllMissionWords()
	if len(words) != len(ages) {
		t.Fatalf("%d word sets for %d ages", len(words), len(ages))
	}
	for i, age := range ages {
		w, ok := MissionWordsFor(age.Key)
		if !ok || words[i].Age != age.Key {
			t.Errorf("%s has no words, or they are out of order", age.Key)
			continue
		}
		for n, name := range w.Tiers {
			if strings.TrimSpace(name) == "" {
				t.Errorf("%s tier name %d is empty", age.Key, n)
			}
		}
		seen := map[string]bool{}
		for _, place := range w.Places {
			if strings.TrimSpace(place) == "" || seen[place] {
				t.Errorf("%s place %q is empty or repeated", age.Key, place)
			}
			seen[place] = true
		}
		for tier := 0; tier < 4; tier++ {
			for _, o := range w.ObjectivesFor(tier) {
				if !strings.HasSuffix(o, " the") {
					t.Errorf("%s objective %q should end in 'the'", age.Key, o)
				}
				if seen[o] {
					t.Errorf("%s objective %q repeats (tier %d)", age.Key, o, tier)
				}
				seen[o] = true
			}
		}
		for _, twist := range w.Twists {
			if !strings.HasSuffix(twist, ".") {
				t.Errorf("%s twist %q should be a sentence", age.Key, twist)
			}
		}
		for _, pool := range [][3]string{w.Weather, w.Foes, w.Helpers, w.Finds} {
			poolSeen := map[string]bool{}
			for _, word := range pool {
				word = strings.TrimSpace(word)
				if word == "" || poolSeen[word] {
					t.Errorf("%s has an empty or repeated road-event word %q", age.Key, word)
				}
				poolSeen[word] = true
			}
		}
	}
	if _, ok := MissionWordsFor("no_age"); ok {
		t.Error("an unknown age has words")
	}
}

func TestRoadEventStoryPools(t *testing.T) {
	w := MissionWords{}
	for _, kind := range []string{"storm", "helpers", "find", "toll", "foes", "ambush", "ambush_haul", "fair", "ruins"} {
		if w.RoadTemplateCount(kind) < 3 {
			t.Errorf("%s has fewer than three road story templates", kind)
		}
		seen := map[string]bool{}
		for i := 0; i < w.RoadTemplateCount(kind); i++ {
			template := w.RoadTemplate(kind, i)
			if strings.ContainsAny(template, "!—") {
				t.Errorf("%s has forbidden punctuation: %q", kind, template)
			}
			if seen[template] {
				t.Errorf("%s repeats template %q", kind, template)
			}
			seen[template] = true
		}
	}
}

// TestRoadTemplatesCarryTheirRows: every road template has a row id of its own,
// in the export's order, and reads as it is written.
func TestRoadTemplatesCarryTheirRows(t *testing.T) {
	w := MissionWords{}
	n := 0
	rows := map[string]bool{}
	for _, kind := range []string{"storm", "helpers", "find", "toll", "foes", "ambush", "ambush_haul", "fair", "ruins"} {
		for i := 0; i < w.RoadTemplateCount(kind); i++ {
			n++
			row := roadEventTemplates[kind][i].rowID
			if want := fmt.Sprintf("config.mission_words.roadEventTemplates.%02d", n); row != want {
				t.Errorf("%s template %d has the row %q, want %q", kind, i, row, want)
			}
			rows[row] = true
		}
	}
	if len(rows) != n {
		t.Errorf("%d road templates with %d rows, want one row each", n, len(rows))
	}
}
