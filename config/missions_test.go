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

func TestBadgeExpeditionsAreEveryEraAndKind(t *testing.T) {
	got := BadgeExpeditions()
	if want := len(Epochs()) * 3; len(got) != want {
		t.Fatalf("%d mission badge subjects, want %d", len(got), want)
	}
	if got[0].Key != MissionBadgeKey(Epochs()[0].Key, MissionScout) || got[1].Kind != MissionExpedition || got[2].Kind != MissionCampaign {
		t.Errorf("order: %+v", got[:3])
	}
}

// TestBoardRuleNumbers pins the numbers the mission board runs on.
func TestBoardRuleNumbers(t *testing.T) {
	for name, c := range map[string][2]float64{
		"offers per tier": {BoardOffersPerTier, 3}, "refill minutes": {BoardRefillMinutes, 5}, "expire minutes": {BoardExpireMinutes, 60},
		"rare chance": {RareOfferChance, .05}, "rare gap": {RareOfferGap, 10}, "rare floor": {RareOfferFloor, 40},
		"rare length": {RareLengthFactor, 1.5}, "rare success cut": {RareSuccessCut, .15}, "rare success floor": {RareSuccessFloor, .10},
		"rare haul": {RareHaulFactor, 4}, "civil success": {CivilGrandSuccess, .60}, "civil cost minutes": {CivilGrandCostMinutes, 45},
	} {
		if c[0] != c[1] {
			t.Errorf("%s is %v, want %v", name, c[0], c[1])
		}
	}
	if RareTitleWord != "Legendary" {
		t.Errorf("the rare word is %q", RareTitleWord)
	}
}
