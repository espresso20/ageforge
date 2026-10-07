package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// cappedState is a Modern Age player holding every tech so far and enough
// milestone rewards to put all production, gold and knowledge past their
// +200% caps, with iron under its own. The techs are in a layer of their
// own and fill no pool.
func cappedState(t *testing.T) (*game.GameEngine, game.GameState) {
	t.Helper()
	ge := game.NewGameEngine()
	ge.SeedRNG(1) // entering an era rolls an epoch event: the same one every run
	if err := ge.EnterAgeForTest("modern_age"); err != nil {
		t.Fatal(err)
	}
	ge.GrantTechsForTest()
	ge.GrantBonusForTest("production_all", 2.4)
	ge.GrantBonusForTest("gold_rate", 3.8)
	ge.GrantBonusForTest("knowledge_rate", 3.35)
	ge.GrantBonusForTest("iron_rate", 0.7)
	return ge, ge.GetState()
}

// earned is the "capped at +200%: +N% earned" note for pool as st has it.
// The fixture's epoch event may add to a pool, so the tests read the figure
// off the state instead of typing it.
func earned(t *testing.T, st game.GameState, pool string) string {
	t.Helper()
	p := st.Pools[pool]
	if !p.Limited || p.Applied != 2 {
		t.Fatalf("%s: earned %v, applied %v, limited %v; this fixture needs it past the +200%% cap", pool, p.Earned, p.Applied, p.Limited)
	}
	return "capped at +200%: " + textfmt.SignedPercent(p.Earned) + " earned"
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
	for _, want := range []string{
		"Gold production     [-] [green]+200%[-] [yellow]" + earned(t, st, "gold_rate") + "[-]",
		"Knowledge production[-] [green]+200%[-] [yellow]" + earned(t, st, "knowledge_rate") + "[-]",
		"[yellow]" + earned(t, st, "production_all") + "[-]", // morale rides on top of its headline
		"Iron production     [-] [green]+70%[-]   ",          // under its cap: no note
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Active multipliers is missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "+380%[-]   ") || strings.Contains(out, "[green]+335%[-] ") {
		t.Errorf("a capped pool's headline still shows the raw sum:\n%s", out)
	}
}

// TestResearchPanelBonuses: the Stats panel's Research bonuses list is what the techs come
// to together. Their output bonuses are in a layer no cap holds, so none
// carries a cap note, even here, where the pools for all production and gold
// are past their caps. Cuts of a price or a time and the mechanic numbers
// read as what they are, and a first source's flat output as an amount.
func TestResearchPanelBonuses(t *testing.T) {
	_, st := cappedState(t)
	earned(t, st, "production_all") // the fixture's pools are past their caps
	earned(t, st, "gold_rate")
	out := section(t, statsProvider(st, 120), "Research bonuses", "Resource rates")
	for _, want := range []string{
		"[green]+5%    [-] All production\n",
		"[green]+31%   [-] Gold production\n",
		"[green]+81%   [-] Knowledge production\n",
		"[green]+25%   [-] Storage\n",
		"[green]+32%   [-] Housing\n",
		"[green]-19%   [-] Building costs\n",
		"[green]-27%   [-] Construction time\n",
		"[green]-23%   [-] Research time\n",
		"[green]-50%   [-] Wonder construction time\n",
		"[green]+25%   [-] Festival length\n",
		"[green]+20%   [-] Soldier storage\n",
		"[green]+10%   [-] Trade route income\n",
		"[green]+5     [-] Morale ceiling, in points\n",
		"[green]-28%   [-] Trade route time\n",
		"[green]-8     [-] Market fee, in points\n",
		"[green]+2     [-] Hand gathering\n",
		"[green]+15%   [-] Game speed\n",
		"[gray]Output:[-]  +1 data/tick, +0.25 steel/tick",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Research bonuses is missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "capped") {
		t.Errorf("a tech's bonus carries a cap note:\n%s", out)
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

// TestMilestonesAndWondersTagCappedRewards: a milestone reward and a wonder
// effect the cap would swallow say so before the player works for them.
func TestMilestonesAndWondersTagCappedRewards(t *testing.T) {
	ge, st := cappedState(t)
	ms := milestonesProvider(st, 120)
	if want := "+20% all production[-] [yellow](capped: no effect now)[-]"; !strings.Contains(ms, want) {
		t.Errorf("the Milestones panel has no capped reward %q:\n%s", want, ms)
	}
	if want := "+5% research speed[-]\n"; !strings.Contains(ms, want) {
		t.Errorf("a research speed reward with room carries a note, or is gone (want %q)", want)
	}
	// The Information Age's wonder, the Global Network, adds knowledge
	// production to a pool that is already full.
	if err := ge.EnterAgeForTest("information_age"); err != nil {
		t.Fatal(err)
	}
	wonders := wondersProvider(ge.GetState(), 120)
	if want := "+30% knowledge production[-] [yellow](capped: no effect now)[-]"; !strings.Contains(wonders, want) {
		t.Errorf("the Wonders panel has no capped wonder effect %q:\n%s", want, wonders)
	}
	if want := "+30 data/tick[-]\n"; !strings.Contains(wonders, want) {
		t.Errorf("the Global Network's flat output carries a note, or is gone (want %q)", want)
	}
}
