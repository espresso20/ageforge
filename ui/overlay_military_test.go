package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// armyGarrison is a Classical Age garrison: 2,000 soldiers, defense 4,000,
// against the age's threat of 1.94K (the threat is measured in soldiers at
// their posts): it blunts 30% of a raid.
const armyGarrison = 2000

func TestArmyPage_NoGarrison(t *testing.T) {
	pg := armyPage(game.NewGameEngine().GetState())
	txt := deckText(pg)
	wantAll(t, "Army panel", txt,
		"Threat", "You have no garrison: raids hit you with full force.",
		"Soldiers blunt raids, war raids and what an Endure takes (up to 45%).",
	)
	if strings.Contains(txt, "Your garrison would blunt") {
		t.Errorf("Army panel claims a garrison with no soldiers:\n%s", txt)
	}
	if n := pg.Notes[0]; n.K != xBad {
		t.Errorf("the garrison warning should be in the bad colour: %+v", n)
	}
}

// Until a military building stands the panel says nobody is posted yet, and
// what brings soldiers. A War Camp comes with its soldiers from the day it
// stands, two ages before the age that brings them to every town: the panel
// then shows the posts and what the garrison does.
func TestArmyPage_PostsBeforeSoldiers(t *testing.T) {
	engine := game.NewGameEngine()
	engine.EnsureMissionBoard()
	txt := deckText(armyPage(engine.GetState()))
	if !strings.Contains(txt, "none yet. A military building comes with its soldiers.") || strings.Contains(txt, "/tick") {
		t.Errorf("a new game's Army panel should say nobody is posted yet, with no rate:\n%s", txt)
	}
	// A new game's board may offer a military grand mission, whose entry says
	// what it waits for, or none, and the empty-list wording says so.
	if !strings.Contains(txt, "come with military buildings") && !strings.Contains(txt, "Soldiers") && !strings.Contains(txt, "A military building comes with its soldiers") {
		t.Errorf("the Army panel should say where soldiers come from:\n%s", txt)
	}
	must(t, engine.EnterAgeForTest("stone_age"))
	if txt = deckText(armyPage(engine.GetState())); !strings.Contains(txt, "none yet.") {
		t.Errorf("a Stone Age town with no War Camp should still say nobody is posted:\n%s", txt)
	}
	// A garrison in the Stone Age: the panel shows the army, and what the
	// garrison does against a Stone Age raid.
	engine.SetGarrisonForTest("stone_age", 1500)
	state := engine.GetState()
	txt = deckText(armyPage(state))
	blunt := config.DefenseMitigation(state.Military.DefenseRating, config.AgeThreat(config.AgeByKey()["stone_age"].Order)) * 100
	wantAll(t, "the Stone Age Army panel with a garrison", txt, "1.5K posted of", "raids in the Stone Age", fmt.Sprintf("Your garrison would blunt about %.0f%% of a raid.", blunt))
	must(t, engine.EnterAgeForTest("bronze_age"))
	must(t, engine.EnterAgeForTest("iron_age"))
	if txt = deckText(armyPage(engine.GetState())); strings.Contains(txt, "Training:") || strings.Contains(txt, "/tick") {
		t.Errorf("nothing trains soldiers any more, yet the Iron Age Army panel shows a rate:\n%s", txt)
	}
}

func TestArmyPage_Garrison(t *testing.T) {
	engine := game.NewGameEngine()
	engine.SetGarrisonForTest("classical_age", armyGarrison)
	txt := deckText(armyPage(engine.GetState()))
	wantAll(t, "Army panel", txt,
		// The Classical Age's threat in soldiers at their posts: 1.94K.
		"Threat", "1.94K", "raids in the Classical Age",
		"Your garrison would blunt about 30% of a raid.",
		"Raids, war raids and an Endure's losses all hit you that much softer.",
		"Twice the garrison: about 36%. No army blunts more than 45%.",
	)
}

// The same garrison a few ages on is outmatched: the threat doubles per age.
func TestArmyPage_GarrisonOutmatchedLater(t *testing.T) {
	engine := game.NewGameEngine()
	engine.SetGarrisonForTest("industrial_age", armyGarrison)
	txt := deckText(armyPage(engine.GetState()))
	if !strings.Contains(txt, "Your garrison would blunt about 17% of a raid.") {
		t.Errorf("an outmatched garrison should blunt about 17%%:\n%s", txt)
	}
}

// armyState is a mid-game army: posts, bonuses, a campaign under way, what
// the garrison has saved, and two campaigns on the board.
func armyState() game.GameState {
	s := game.GameState{Age: "iron_age", AgeName: "Iron Age"}
	s.Military = game.MilitaryState{
		SoldierCount: 120, SoldierCap: 200, DefenseRating: 4000, Threat: 1940, Mitigation: 0.3,
		MilitaryBonus: 0.1, ExpeditionBonus: 0.05, CompletedCount: 3,
		Posts:          []game.SoldierPost{{Name: "War Camp", Count: 2, Posted: 80, Posts: 100}, {Name: "Barracks", Count: 1, Posted: 40, Posts: 100}},
		Saved:          &game.DefenseTally{Buildings: 3, Raids: 2},
		ActiveMilitary: &game.ExpeditionSnapshot{Name: "Siege of Ashford", Soldiers: 60, TicksLeft: 30, TotalTicks: 120},
		Expeditions: []game.ExpeditionInfo{
			{Name: "Border Skirmish", Category: game.ExpeditionMilitary, Description: "Push the raiders back past the river.", SoldiersNeeded: 40,
				DurationMin: 100, DurationMax: 200, Difficulty: 0.4, Chance: 0.25, Cost: map[string]float64{"food": 30}, CanLaunch: true},
			{Name: "Hill Raid", Category: game.ExpeditionMilitary, Description: "A hill fort.", SoldiersNeeded: 500,
				DurationMin: 100, DurationMax: 200, Difficulty: 0.6, Chance: 0.6, LaunchBlockReason: "need 500 soldiers"},
			{Name: "Scout the Coast", Category: "scouting", Description: "Not a campaign.", CanLaunch: true},
		},
	}
	return s
}

// TestArmyPage_Cards: a mid-game state gives the overview, the campaign under
// way and a card to each military campaign (and none to a scouting offer),
// with every fact the old page showed.
func TestArmyPage_Cards(t *testing.T) {
	pg := armyPage(armyState())
	if n := len(deckSectionNamed(t, pg, "Overview").Cards); n != 4 {
		t.Errorf("Overview has %d cards, want 4", n)
	}
	if n := len(deckSectionNamed(t, pg, "Active Campaign").Cards); n != 1 {
		t.Errorf("Active Campaign has %d cards, want 1", n)
	}
	if n := len(deckSectionNamed(t, pg, "Campaigns").Cards); n != 2 {
		t.Errorf("Campaigns has %d cards, want 2", n)
	}
	out := deckText(pg)
	wantAll(t, "army page", out,
		"120 posted of 200", "Military Bonus +10%", "Expedition Bonus +5%",
		"War Camp x2 80/100", "Barracks x1 40/100", "4K", "1.94K", "raids in the Iron Age",
		"Your garrison would blunt about 30% of a raid.", "Saved this run: 3 buildings (2 raids blunted)",
		"Siege of Ashford", "60 soldiers deployed", "left",
		"Completed: 3 (expeditions and campaigns)",
		"Push the raiders back past the river.", "Soldiers 40", "Duration", "Difficulty 40%", "(chance of failure; 25% with your army)",
		"Cost 30 food", "need 500 soldiers", "(chance of failure before bonuses)",
		"Commands: expedition opens the Expeditions panel; expedition <name> sends an offer")
	if strings.Contains(out, "Scout the Coast") || strings.Contains(out, "No active campaign") {
		t.Errorf("a scouting offer is on the Army page, or it denies the campaign under way:\n%s", out)
	}
	if c, _ := deckCardNamed(t, pg, "Siege of Ashford"); !c.Meter || c.Share < 0.74 || c.Share > 0.76 {
		t.Errorf("the campaign's meter = %+v", c)
	}
	if c, _ := deckCardNamed(t, pg, "Border Skirmish"); c.Command != "expedition Border Skirmish" || c.Quiet {
		t.Errorf("Border Skirmish = %+v", c)
	}
	if c, _ := deckCardNamed(t, pg, "Hill Raid"); c.Command != "" || !c.Quiet || c.Why != "need 500 soldiers" {
		t.Errorf("Hill Raid should only show: %+v", c)
	}
	s := armyState()
	s.Military.ActiveMilitary = nil
	s.Military.Expeditions = nil
	out = deckText(armyPage(s))
	wantAll(t, "army page without a campaign", out, "No active campaign", "No campaigns available yet.",
		"Campaigns are grand missions and cost soldiers,", "which come with military buildings. They open as your civilization grows.")
}

// TestArmyPage_RunsItsCommands: Enter on a campaign runs what the old page
// told the player to type.
func TestArmyPage_RunsItsCommands(t *testing.T) {
	wantDeckEnter(t, armyPage, armyState(), "Border Skirmish", "expedition Border Skirmish")
}

func TestArmyPage_Fits(t *testing.T) {
	wantDeckFits(t, armyPage, armyState())
	wantDeckFits(t, armyPage, game.GameState{})
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
	d, pages := newTestDashboard(engine)
	if !d.overlayMgr.Show("army", engine.GetState()) {
		t.Fatal("overlay army not registered")
	}
	if !strings.Contains(drawPrimitive(t, pages, w, h).Text(), "Your garrison would blunt about 30% of a raid.") {
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
	d, pages := newTestDashboard(engine)
	d.showCatastropheModal(st.PendingCatastrophe)
	screen := drawPrimitive(t, pages, w, h).Text()
	for _, want := range []string{
		"All resources reduced to 34.1%",
		"Your garrison keeps 34.1% of stock, not 15%",
	} {
		if !strings.Contains(screen, want) {
			t.Errorf("modal missing %q", want)
		}
	}
	if strings.Contains(screen, "No garrison") {
		t.Error("modal says no garrison with 640,000 soldiers standing")
	}

	// Without soldiers the modal shows the old numbers and says so.
	l := buildCatastropheModalLayout("iron_era", false, 0, game.DefaultEndureOutcome(), game.GameState{})
	endure := untag(l.endure)
	wantAll(t, "garrison-free Endure", endure, "20% of buildings destroyed", "All resources reduced to 15%", "No garrison: soldiers would soften this")
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
			if w := kit.Width(line); w > inner {
				t.Errorf("line %q is %d cells, box interior is %d", line, w, inner)
			}
		}
	}
}
