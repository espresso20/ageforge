package config

import (
	"strings"
	"testing"
)

func TestMissionKinds(t *testing.T) {
	want := map[string]MissionTuning{
		MissionScout:      {5, 30, .85, 2, 5},
		MissionExpedition: {30, 60, .75, 3, 15},
		MissionCampaign:   {60, 120, .65, 4, 10},
	}
	kinds := MissionKinds()
	if len(kinds) != 3 || kinds[0] != MissionScout || kinds[1] != MissionExpedition || kinds[2] != MissionCampaign {
		t.Fatalf("kinds %v", kinds)
	}
	for _, kind := range kinds {
		got, ok := MissionKindTuning(kind)
		if !ok || got != want[kind] {
			t.Errorf("%s has tuning %+v, want %+v", kind, got, want[kind])
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
		for tier, objs := range w.Objectives {
			for _, o := range objs {
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
	}
	if _, ok := MissionWordsFor("no_age"); ok {
		t.Error("an unknown age has words")
	}
}

func TestBadgeExpeditionsAreEveryEraAndKind(t *testing.T) {
	got := BadgeExpeditions()
	if want := len(Epochs()) * 3; len(got) != want {
		t.Fatalf("%d mission badge subjects, want %d", len(got), want)
	}
	if got[0].Key != MissionBadgeKey(Epochs()[0].Key, MissionScout) || got[1].Kind != MissionExpedition || got[2].Kind != MissionCampaign {
		t.Errorf("order: %+v", got[:3])
	}
}
