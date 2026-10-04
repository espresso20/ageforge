package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// cappedState is a Modern Age player holding every tech so far: the techs
// alone put all production, gold and knowledge past their +200% caps.
func cappedState(t *testing.T) (*game.GameEngine, game.GameState) {
	t.Helper()
	ge := game.NewGameEngine()
	ge.SeedRNG(1) // entering an era rolls an epoch event: the same one every run
	if err := ge.EnterAgeForTest("modern_age"); err != nil {
		t.Fatal(err)
	}
	ge.GrantTechsForTest()
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

// TestResearchPanelBonuses: the Research bonuses list shows percentages
// only for the bonus pools, with the cap note, and flat output, storage and
// housing as amounts. It used to print every tech effect as a percentage:
// "+37500% All production" for +375 storage.
func TestResearchPanelBonuses(t *testing.T) {
	_, st := cappedState(t)
	out := section(t, researchProvider(st, 120), "Research bonuses", "Available now")
	for _, want := range []string{
		"+240%", "All production [yellow]" + earned(t, st, "production_all") + "[-]",
		"Gold production [yellow]" + earned(t, st, "gold_rate") + "[-]",
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
		"+30% all production [yellow](" + earned(t, st, "production_all") + ")[gray]",
		"+30% gold production [yellow](" + earned(t, st, "gold_rate") + ")[gray]",
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

// TestResearchPanelShowsRealTimes: a tech's listed time is the time it
// takes to research now, research speed and Era Mastery included, and the
// panel says what research speed does. It used to list times without
// research speed, so the bonus never showed.
func TestResearchPanelShowsRealTimes(t *testing.T) {
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	out := researchProvider(ge.GetState(), 120)
	base := config.TechByKey()["tool_making"].ResearchTicks
	if want := formatTicks(base, ge.GetState()); !strings.Contains(out, "800 knowledge · "+want) {
		t.Fatalf("with no bonus Tool Making should list its base time %s:\n%s", want, section(t, out, "Available now", "Tech tree"))
	}
	if strings.Contains(out, "Research speed") {
		t.Errorf("a new game mentions research speed:\n%s", out)
	}
	// One epoch succumbed in: Ancient Knowledge, +25% research speed.
	ge.SetLegacyBonusForTest("iron_era")
	st := ge.GetState()
	out = researchProvider(st, 120)
	quick := game.ResearchTicks(base, 0.25, 1)
	if quick >= base {
		t.Fatalf("+25%% research speed leaves Tool Making at %d of %d ticks", quick, base)
	}
	if want := "800 knowledge · " + formatTicks(quick, st); !strings.Contains(out, want) {
		t.Errorf("with +25%% research speed Tool Making should list %q:\n%s", want, section(t, out, "Available now", "Tech tree"))
	}
	if want := "Research speed +25%: techs take 75% of their base time."; !strings.Contains(out, want) {
		t.Errorf("the Research panel does not say %q", want)
	}
}
