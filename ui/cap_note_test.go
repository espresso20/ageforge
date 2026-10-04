package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// cappedState is a Modern Age player holding every tech so far: the techs
// alone put all production, gold and knowledge past their +200% caps.
func cappedState(t *testing.T) (*game.GameEngine, game.GameState) {
	t.Helper()
	ge := game.NewGameEngine()
	if err := ge.EnterAgeForTest("modern_age"); err != nil {
		t.Fatal(err)
	}
	ge.GrantTechsForTest()
	return ge, ge.GetState()
}

func section(t *testing.T, out, from, to string) string {
	t.Helper()
	i := strings.Index(out, from)
	if i < 0 {
		t.Fatalf("no %q in:\n%s", from, out)
	}
	out = out[i:]
	if j := strings.Index(out, to); to != "" && j >= 0 {
		out = out[:j]
	}
	return out
}

// TestStatsPanelShowsCappedPools: the Active multipliers headline is what
// the engine applies, and a pool past its cap says so with what was earned.
func TestStatsPanelShowsCappedPools(t *testing.T) {
	_, st := cappedState(t)
	out := renderActiveMultipliers(st)
	for _, pool := range []string{"production_all", "gold_rate", "knowledge_rate"} {
		p := st.Pools[pool]
		if !p.Limited || p.Applied != 2 {
			t.Fatalf("%s: earned %v, applied %v, limited %v; this fixture needs it past the +200%% cap", pool, p.Earned, p.Applied, p.Limited)
		}
	}
	for _, want := range []string{
		"Gold production     [-] [green]+200%[-] [yellow]capped at +200%: +380% earned[-]",
		"Knowledge production[-] [green]+200%[-] [yellow]capped at +200%: +335% earned[-]",
		"[yellow]capped at +200%: +240% earned[-]",  // all production (morale rides on top of its headline)
		"Iron production     [-] [green]+70%[-]   ", // under its cap: no note
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Active multipliers is missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "+380%[-]   ") || strings.Contains(out, "[green]+335%[-] ") {
		t.Errorf("a capped pool's headline still shows the raw sum:\n%s", out)
	}
}

// TestResearchPanelBonuses: the Research bonuses list shows percentages
// only for the bonus pools, with the cap note, and flat output, storage and
// housing as amounts. It used to print every tech effect as a percentage:
// "+37500% All production" for +375 storage.
func TestResearchPanelBonuses(t *testing.T) {
	_, st := cappedState(t)
	out := section(t, researchProvider(st, 120), "Research bonuses", "Available now")
	for _, want := range []string{
		"+240%", "All production [yellow]capped at +200%: +240% earned[-]",
		"Gold production [yellow]capped at +200%: +380% earned[-]",
		"[gray]Output:[-]", "+2.8 food/tick", "+19 electricity/tick",
		"[gray]Storage:[-] +375 for every resource, +100 gold",
		"[gray]Housing:[-] +5",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Research bonuses is missing %q:\n%s", want, out)
		}
	}
	for _, bad := range []string{"37500%", "10210%", "1900%", "+500%", "+280%"} {
		if strings.Contains(out, bad) {
			t.Errorf("Research bonuses shows an amount as a percentage (%s):\n%s", bad, out)
		}
	}
}

// TestResearchTreeTagsCappedTechs: a researched tech whose pool is past its
// cap carries the pool's note, a tech still to research says what the cap
// would leave of it, and a bonus with room carries nothing.
func TestResearchTreeTagsCappedTechs(t *testing.T) {
	ge, st := cappedState(t)
	tree := section(t, researchProvider(st, 120), "Tech tree", "")
	for _, want := range []string{
		"+30% all production [yellow](capped at +200%: +240% earned)[gray]",
		"+30% gold production [yellow](capped at +200%: +380% earned)[gray]",
		"+40% iron production,", // Iron Smelting: iron is under its cap
	} {
		if !strings.Contains(tree, want) {
			t.Errorf("the tech tree is missing %q", want)
		}
	}
	// One age on, the Information Age techs are offered: Internet's
	// knowledge bonus would add nothing.
	if err := ge.EnterAgeForTest("information_age"); err != nil {
		t.Fatal(err)
	}
	st = ge.GetState()
	avail := section(t, researchProvider(st, 120), "Available now", "Tech tree")
	if want := "+120% knowledge production [yellow](capped: no effect now)[gray]"; !strings.Contains(avail, want) {
		t.Errorf("Available now is missing %q:\n%s", want, avail)
	}
	if want := "+3 data/tick,"; !strings.Contains(avail, want) && !strings.Contains(avail, "+3 data/tick[") {
		t.Errorf("Available now lost Internet's flat output:\n%s", avail)
	}
}

// TestFestivalWarnsWhenCapped: the festival costs culture, so the command
// says before the player pays that the cap would leave nothing of it.
func TestFestivalWarnsWhenCapped(t *testing.T) {
	ge, _ := cappedState(t)
	if note := ge.FestivalStatus().CapNote; note != "capped: no effect now" {
		t.Fatalf("FestivalStatus().CapNote = %q past the cap, want %q", note, "capped: no effect now")
	}
	out := cmdFestivalStatus(ge).Message
	if want := "Right now it is capped: no effect now."; !strings.Contains(out, want) {
		t.Errorf("festival status is missing %q:\n%s", want, out)
	}
	fresh := game.NewGameEngine()
	if note := fresh.FestivalStatus().CapNote; note != "" {
		t.Errorf("a new game's festival carries the note %q, want none", note)
	}
	if out := cmdFestivalStatus(fresh).Message; strings.Contains(out, "capped") {
		t.Errorf("a new game's festival status mentions a cap:\n%s", out)
	}
}
