package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/vibe"
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

func TestBadgeExpeditionsAreEveryEraAndKind(t *testing.T) {
	got := BadgeExpeditions()
	if want := len(Epochs()) * 3; len(got) != want {
		t.Fatalf("%d mission badge subjects, want %d", len(got), want)
	}
	if got[0].Key != MissionBadgeKey(Epochs()[0].Key, MissionScout) || got[1].Kind != MissionExpedition || got[2].Kind != MissionCampaign {
		t.Errorf("order: %+v", got[:3])
	}
}

// slotsOf lists the <slots> of a road story template, in order.
func slotsOf(s string) []string { return regexp.MustCompile(`<[a-z]+>`).FindAllString(s, -1) }

// TestRoadTemplatesCarryTheirRowsAndFollowTheSwitch: every road template has a
// row id of its own, in the export's order, and is told by the script's line
// for it while VibeMode is on. Off, or with no line, it reads as it is written.
func TestRoadTemplatesCarryTheirRowsAndFollowTheSwitch(t *testing.T) {
	was := vibe.On()
	defer vibe.Set(was)
	w := MissionWords{}
	n := 0
	rows := map[string]bool{}
	for _, kind := range []string{"storm", "helpers", "find", "toll", "foes", "ambush", "ambush_haul", "fair", "ruins"} {
		for i := 0; i < w.RoadTemplateCount(kind); i++ {
			n++
			template := w.RoadTemplate(kind, i)
			row := roadEventTemplates[kind][i].vibeRow
			if want := fmt.Sprintf("config.mission_words.roadEventTemplates.%02d", n); row != want {
				t.Errorf("%s template %d has the row %q, want %q", kind, i, row, want)
			}
			rows[row] = true

			restore := vibe.UseScriptForTest(map[string]string{row: "SCRIPT " + row})
			vibe.Set(false)
			if got := w.RoadLine(template); got != template {
				t.Errorf("%s: VibeMode off reads %q, want the template", row, got)
			}
			vibe.Set(true)
			if got := w.RoadLine(template); got != "SCRIPT "+row {
				t.Errorf("%s: VibeMode on reads %q, want the script's line", row, got)
			}
			restore()
			restore = vibe.UseScriptForTest(map[string]string{})
			if got := w.RoadLine(template); got != template {
				t.Errorf("%s: VibeMode on with no line reads %q, want the template", row, got)
			}
			restore()
			vibe.Set(false)
		}
	}
	if len(rows) != n {
		t.Errorf("%d road templates with %d rows, want one row each", n, len(rows))
	}
	if got := w.RoadLine("A story written by hand."); got != "A story written by hand." {
		t.Errorf("a story that is no template reads %q", got)
	}
}

// TestTheScriptTellsRoadTemplatesWithTheirSlots: the script's line for a road
// template has every <slot> the template has, none more, because the number
// or the word that goes in a slot is put in by whoever tells the story, and
// the lint cannot read <slots>.
func TestTheScriptTellsRoadTemplatesWithTheirSlots(t *testing.T) {
	told, total := 0, 0
	for kind, pool := range roadEventTemplates {
		for _, tpl := range pool {
			total++
			line, ok := vibe.Text(tpl.vibeRow)
			if !ok {
				continue
			}
			told++
			want, got := slotsOf(tpl.text), slotsOf(line)
			sort.Strings(want)
			sort.Strings(got)
			if strings.Join(want, "") != strings.Join(got, "") {
				t.Errorf("%s (%s): the script's line has the slots %v, the template %v:\n  %s\n  %s", tpl.vibeRow, kind, got, want, line, tpl.text)
			}
			if strings.ContainsAny(line, "—") {
				t.Errorf("%s: the script's line has an em dash: %s", tpl.vibeRow, line)
			}
		}
	}
	if told != total {
		t.Errorf("%d of the %d road templates have a line in the script", told, total)
	}
}
