package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TestMilestonesPageListsEveryCategory: the page's Progress total counts
// every milestone, so every milestone has to be a card in a section. The
// epoch milestones were counted and never listed.
func TestMilestonesPageListsEveryCategory(t *testing.T) {
	inOrder := map[string]bool{}
	names := config.MilestoneCategoryNames()
	for _, cat := range config.MilestoneCategoryOrder() {
		inOrder[cat] = true
		if names[cat] == "" {
			t.Errorf("milestone category %q has no display name", cat)
		}
	}

	var st game.GameState
	st.Milestones.Milestones = map[string]game.MilestoneInfo{}
	for _, m := range config.Milestones() {
		if !inOrder[m.Category] {
			t.Errorf("milestone %s is in category %q, which config.MilestoneCategoryOrder leaves out", m.Key, m.Category)
		}
		st.Milestones.Milestones[m.Key] = game.MilestoneInfo{
			Name: m.Name, Category: m.Category, Visible: true, Completed: true,
		}
	}
	st.Milestones.TotalCount = len(st.Milestones.Milestones)
	st.Milestones.CompletedCount = st.Milestones.TotalCount

	pg := milestonesPage(st)
	if pg.count() != st.Milestones.TotalCount {
		t.Errorf("%d cards for %d milestones", pg.count(), st.Milestones.TotalCount)
	}
	for _, m := range config.Milestones() {
		found := false
		for _, c := range deckSectionNamed(t, pg, names[m.Category]).Cards {
			if c.Title == m.Name && c.Chip == "✓ done" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s (%s) is not a done card under %s", m.Name, m.Key, names[m.Category])
		}
	}
}

// TestMilestonesPageEpochGroup: a new game lists the epoch milestones it can
// see in their own section and counts the rest as hidden without naming them.
// Age Hopper needs the Classical Age, which a Primitive Age player cannot see
// named yet (game/spoilers.go), so it stays hidden though it is not marked
// Hidden.
func TestMilestonesPageEpochGroup(t *testing.T) {
	engine := freshSpoilerEngine(t)
	pg := milestonesPage(engine.GetState())
	shown := deckText(pg)

	epoch := deckSectionNamed(t, pg, "Epoch")
	var body []string
	for _, c := range epoch.Cards {
		body = append(body, c.Title)
		if c.Chip == "✓ done" {
			t.Errorf("a new game has done %s", c.Title)
		}
	}
	for _, name := range []string{"First Farmers", "Survivor"} {
		if !strings.Contains(strings.Join(body, "\n"), name) {
			t.Errorf("the Epoch section should list %s: %v", name, body)
		}
	}
	for _, name := range []string{"Age Hopper", "Enduring Civilization", "Industrial Titan", "Power Grid"} {
		if strings.Contains(shown, name) {
			t.Errorf("a new game's milestones page names %s:\n%s", name, shown)
		}
	}
	hidden := ""
	for _, n := range epoch.Notes {
		hidden += n.Value
	}
	if !strings.Contains(hidden, "+ 4 hidden milestones") {
		t.Errorf("the Epoch section should count its 4 hidden milestones: %q", hidden)
	}
	checkSpoilers(t, "milestones page", shown, freshSpoilerTerms())
}

// milestoneState is a settlement chain half done: one done, two with
// conditions (one of them half met), and one hidden.
func milestoneState() game.GameState {
	var s game.GameState
	s.Milestones.CompletedCount, s.Milestones.TotalCount = 5, 77
	s.Milestones.CurrentTitle = "Aspiring"
	s.Milestones.Chains = []game.ChainInfo{{Category: "settlement", CompletedCount: 1, TotalCount: 3}, {Category: "builder", CompletedCount: 2, TotalCount: 2, Complete: true, Title: "Master Builder", BoostActive: true}}
	s.Milestones.Milestones = map[string]game.MilestoneInfo{
		"a": {Name: "First Shelter", Category: "settlement", Visible: true, Completed: true, RewardText: "+10 food"},
		"b": {Name: "Small Village", Category: "settlement", Visible: true, Description: "Reach a population of 250.", RewardText: "+50 wood",
			Progress: []game.MilestoneProgress{{Label: "people", Current: 200, Target: 250}}},
		"c": {Name: "Bustling Town", Category: "settlement", Visible: true, Description: "Reach 1,500 and build.", RewardText: "+5% food production",
			Progress: []game.MilestoneProgress{{Label: "people", Current: 0, Target: 1500}, {Label: "huts", Met: true, Current: 5, Target: 5}, {Label: "wells", Current: 1, Target: 4}}},
		"d": {Name: "Secret", Category: "settlement", Hidden: true},
		"e": {Name: "Early Builder", Category: "builder", Visible: true, Completed: true, RewardText: "+20 stone"},
	}
	return s
}

// TestMilestonesPage_Cards: the sections, chips and facts the old page showed:
// progress and title, the chain's count and title and boost, each done
// milestone with its reward, each open one with its description, every
// condition and its reward, and the hidden count. No card has a command.
func TestMilestonesPage_Cards(t *testing.T) {
	pg := milestonesPage(milestoneState())
	set := deckSectionNamed(t, pg, "Settlement")
	if set.Meta != "1 of 3" || len(set.Cards) != 3 || set.MinW != 34 {
		t.Errorf("Settlement = meta %q, %d cards, minW %d", set.Meta, len(set.Cards), set.MinW)
	}
	if b := deckSectionNamed(t, pg, "Builder"); b.Meta != "2 of 2" {
		t.Errorf("Builder meta = %q", b.Meta)
	}
	wantAll(t, "milestones page", deckText(pg),
		"Progress: 5/77", "\"Aspiring\"", "✓ Master Builder", "⚡ boost active", "+ 1 hidden milestone",
		"First Shelter ✓ done", "Reward +10 food", "Reward +50 wood", "Reward +5% food production",
		"Reach a population of 250.", "200/250 people", "0/1.5K people", "✓ huts", "1/4 wells")
	done, _ := deckCardNamed(t, pg, "First Shelter")
	village, _ := deckCardNamed(t, pg, "Small Village")
	town, _ := deckCardNamed(t, pg, "Bustling Town")
	if !done.Quiet || done.Chip != "✓ done" || done.Meter {
		t.Errorf("a done card = %+v", done)
	}
	if village.Chip != "now" || !village.Meter || village.Share < 0.79 || village.Share > 0.81 {
		t.Errorf("the milestone in hand = %+v", village)
	}
	if town.Chip != "" || !town.Meter || town.Share != 0 {
		t.Errorf("a milestone not begun = %+v", town)
	}
	for n := 0; n < pg.count(); n++ {
		if c, _ := pg.card(n); c.Command != "" || c.Action != "" {
			t.Errorf("%s has a command %q", c.Title, c.Command)
		}
	}
	if strings.Contains(deckText(pg), "Secret") {
		t.Error("a hidden milestone is named")
	}
}

func TestMilestonesPage_Fits(t *testing.T) {
	wantDeckFits(t, milestonesPage, milestoneState())
	wantDeckFits(t, milestonesPage, game.GameState{})
}
