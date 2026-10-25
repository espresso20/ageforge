package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// armyGarrison is a Classical Age garrison: 2,000 soldiers, defense 4,000,
// against the age's threat of 1.94K (the threat is measured in soldiers at
// their posts): it blunts 30% of a raid.
const armyGarrison = 2000

func TestArmyPanel_NoGarrison(t *testing.T) {
	txt := untag(militaryProvider(game.NewGameEngine().GetState(), 0))
	for _, want := range []string{
		"Threat:", "You have no garrison: raids hit you with full force.",
		"Soldiers blunt raids, war raids and what an Endure takes (up to 45%).",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("Army panel missing %q:\n%s", want, txt)
		}
	}
	if strings.Contains(txt, "Your garrison would blunt") {
		t.Errorf("Army panel claims a garrison with no soldiers:\n%s", txt)
	}
}

// Until a military building stands the panel says nobody is posted yet, and
// what brings soldiers. A War Camp comes with its soldiers from the day it
// stands, two ages before the age that brings them to every town: the panel
// then shows the posts and what the garrison does.
func TestArmyPanel_PostsBeforeSoldiers(t *testing.T) {
	engine := game.NewGameEngine()
	engine.EnsureMissionBoard()
	txt := untag(militaryProvider(engine.GetState(), 0))
	if !strings.Contains(txt, "Posts:     none yet. A military building comes with its soldiers.") || strings.Contains(txt, "/tick") {
		t.Errorf("a new game's Army panel should say nobody is posted yet, with no rate:\n%s", txt)
	}
	// A new game's board may offer a military grand mission, whose entry says
	// what it waits for, or none, and the empty-list wording says so.
	if !strings.Contains(txt, "come with military buildings") && !strings.Contains(txt, "needs a party of") && !strings.Contains(txt, "A military building comes with its soldiers") {
		t.Errorf("the Army panel should say where soldiers come from:\n%s", txt)
	}
	must(t, engine.EnterAgeForTest("stone_age"))
	if txt = untag(militaryProvider(engine.GetState(), 0)); !strings.Contains(txt, "Posts:     none yet.") {
		t.Errorf("a Stone Age town with no War Camp should still say nobody is posted:\n%s", txt)
	}
	// A garrison in the Stone Age: the panel shows the army, and what the
	// garrison does against a Stone Age raid.
	engine.SetGarrisonForTest("stone_age", 1500)
	state := engine.GetState()
	txt = untag(militaryProvider(state, 0))
	blunt := config.DefenseMitigation(state.Military.DefenseRating, config.AgeThreat(config.AgeByKey()["stone_age"].Order)) * 100
	for _, want := range []string{"Soldiers:  1.5K posted of", "(raids in the Stone Age)", fmt.Sprintf("Your garrison would blunt about %.0f%% of a raid.", blunt)} {
		if !strings.Contains(txt, want) {
			t.Errorf("the Stone Age Army panel with a garrison is missing %q:\n%s", want, txt)
		}
	}
	must(t, engine.EnterAgeForTest("bronze_age"))
	must(t, engine.EnterAgeForTest("iron_age"))
	if txt = untag(militaryProvider(engine.GetState(), 0)); strings.Contains(txt, "Training:") || strings.Contains(txt, "/tick") {
		t.Errorf("nothing trains soldiers any more, yet the Iron Age Army panel shows a rate:\n%s", txt)
	}
}

func TestArmyPanel_Garrison(t *testing.T) {
	engine := game.NewGameEngine()
	engine.SetGarrisonForTest("classical_age", armyGarrison)
	txt := untag(militaryProvider(engine.GetState(), 0))
	for _, want := range []string{
		// The Classical Age's threat in soldiers at their posts: 1.94K.
		"Threat:    1.94K (raids in the Classical Age)",
		"Your garrison would blunt about 30% of a raid.",
		"Raids, war raids and an Endure's losses all hit you that much softer.",
		"Twice the garrison: about 36%. No army blunts more than 45%.",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("Army panel missing %q:\n%s", want, txt)
		}
	}
}

// The same garrison a few ages on is outmatched: the threat doubles per age.
func TestArmyPanel_GarrisonOutmatchedLater(t *testing.T) {
	engine := game.NewGameEngine()
	engine.SetGarrisonForTest("industrial_age", armyGarrison)
	txt := untag(militaryProvider(engine.GetState(), 0))
	if !strings.Contains(txt, "Your garrison would blunt about 17% of a raid.") {
		t.Errorf("an outmatched garrison should blunt about 17%%:\n%s", txt)
	}
}

func TestGarrisonSavedSummary(t *testing.T) {
	if got := garrisonSavedSummary(nil); got != "" {
		t.Errorf("nothing saved = %q, want empty", got)
	}
	got := untag(garrisonSavedSummary(&game.DefenseTally{
		Buildings: 3, Workers: 1, Raids: 2,
		Resources: map[string]float64{"food": 1200, "gold": 300, "wood": 50, "stone": 40, "iron": 30, "dust": 0.2},
	}))
	want := "3 buildings, 1 worker, 1.2K food, 300 gold, 50 wood, 40 stone (2 raids blunted)"
	if got != want {
		t.Errorf("summary = %q, want %q", got, want)
	}
}

// The Army panel on a real screen: the garrison line is drawn, not just built.
func TestArmyPanel_OnScreen(t *testing.T) {
	const w, h = 160, 50
	engine := game.NewGameEngine()
	engine.SetGarrisonForTest("classical_age", armyGarrison)
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), engine, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	if !d.overlayMgr.Show("army", engine.GetState()) {
		t.Fatal("overlay army not registered")
	}
	if !gridHas(screenGrid(t, pages, w, h), "Your garrison would blunt about 30% of a raid.") {
		t.Error("the garrison line is not on screen")
	}
}

// The catastrophe modal's Endure preview counts the garrison and the Brace.
func TestCatastropheModal_EndureCountsGarrison(t *testing.T) {
	const w, h = 160, 50
	engine := game.NewGameEngine()
	engine.SetGarrisonForTest("iron_age", config.AgeThreat(3)/2) // defense == the Iron Age threat
	must(t, engine.ForceCatastropheForTest())
	st := engine.GetState()
	if st.PendingEndure.Garrison <= 0 {
		t.Fatalf("pending Endure has no garrison: %+v", st.PendingEndure)
	}
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), engine, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	d.showCatastropheModal(st.PendingCatastrophe)
	grid := screenGrid(t, pages, w, h)
	for _, want := range []string{
		"All resources reduced to 34.1%",
		"Your garrison keeps 34.1% of stock, not 15%",
	} {
		if !gridHas(grid, want) {
			t.Errorf("modal missing %q", want)
		}
	}
	if gridHas(grid, "No garrison") {
		t.Error("modal says no garrison with 640,000 soldiers standing")
	}

	// Without soldiers the modal shows the old numbers and says so.
	l := buildCatastropheModalLayout("iron_era", false, 0, game.DefaultEndureOutcome(), game.GameState{})
	endure := untag(l.endure)
	for _, want := range []string{"20% of buildings destroyed", "All resources reduced to 15%", "No garrison: soldiers would soften this"} {
		if !strings.Contains(endure, want) {
			t.Errorf("garrison-free Endure missing %q:\n%s", want, endure)
		}
	}
}

// Every Endure variant (Brace level x garrison, capped or not) fits the box.
func TestCatastropheModal_DefenseLinesFit(t *testing.T) {
	inner := catastropheModalWidth - 2
	for _, o := range []game.EndureOutcome{
		{BraceLevel: 2, Garrison: 0.45, DestroyPct: 8, KeepFrac: 0.66, BracedDestroyPct: 10, BracedKeepFrac: 0.45, Capped: true, BuildingsSaved: 1234},
		{BraceLevel: 1, Garrison: 0.2, DestroyPct: 12, KeepFrac: 0.44, BracedDestroyPct: 15, BracedKeepFrac: 0.30},
		{Garrison: 0.1, DestroyPct: 18, KeepFrac: 0.235, BracedDestroyPct: 20, BracedKeepFrac: 0.15, BuildingsSaved: 1},
	} {
		l := buildCatastropheModalLayout("steel_era", false, 0, o, game.GameState{})
		for _, line := range strings.Split(l.endure, "\n") {
			if w := tview.TaggedStringWidth(line); w > inner {
				t.Errorf("line %q is %d cells, box interior is %d", line, w, inner)
			}
		}
	}
}

// gridHas reports whether any screen row contains s.
func gridHas(grid [][]rune, s string) bool {
	for _, row := range grid {
		if strings.Contains(string(row), s) {
			return true
		}
	}
	return false
}
