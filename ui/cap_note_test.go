package ui

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// cappedState is a Modern Age player holding every tech so far and enough
// milestone rewards to put all production, gold and knowledge past +200%,
// where a bonus counts a quarter, with iron under it. The techs are in a
// layer of their own and fill no pool.
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

// earned is the "counts a quarter past +200%: +N% earned" note for pool as
// st has it. The fixture's epoch event may add to a pool, so the tests read
// the figure off the state instead of typing it. The rule itself is typed:
// +200% in full and a quarter of the rest.
func earned(t *testing.T, st game.GameState, pool string) string {
	t.Helper()
	p := st.Pools[pool]
	if want := 2 + (p.Earned-2)*0.25; !p.Limited || !p.Soft || p.Earned <= 2 || math.Abs(p.Applied-want) > 1e-9 {
		t.Fatalf("%s: earned %v, applied %v (want %v), limited %v, soft %v; this fixture needs it past +200%%", pool, p.Earned, p.Applied, want, p.Limited, p.Soft)
	}
	return "counts a quarter past +200%: " + textfmt.SignedPercent(p.Earned) + " earned"
}

// applied is the headline a pool past +200% shows: what the engine applies.
func applied(st game.GameState, pool string) string {
	return fmt.Sprintf("%+.0f%%", st.Pools[pool].Applied*100)
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
// the engine applies, and a pool past +200% says so with what was earned.
func TestStatsPanelShowsCappedPools(t *testing.T) {
	_, st := cappedState(t)
	out := renderActiveMultipliers(st)
	if got := applied(st, "gold_rate"); got != "+245%" {
		t.Fatalf("gold has earned +380%% and applies %s, want +245%% (+200%% and a quarter of the other +180%%)", got)
	}
	for _, want := range []string{
		"Gold production     [-] [green]+245%[-] [yellow]" + earned(t, st, "gold_rate") + "[-]",
		"Knowledge production[-] [green]" + applied(st, "knowledge_rate") + "[-] [yellow]" + earned(t, st, "knowledge_rate") + "[-]",
		"[yellow]" + earned(t, st, "production_all") + "[-]", // morale rides on top of its headline
		"Iron production     [-] [green]+70%[-]   ",          // under +200%: no note
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Active multipliers is missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "+380%[-]   ") || strings.Contains(out, "[green]+335%[-] ") {
		t.Errorf("a pool past +200%% still shows the raw sum as its headline:\n%s", out)
	}
	if strings.Contains(out, "[green]+200%[-] ") || strings.Contains(out, "capped at +200%") {
		t.Errorf("a pool past +200%% reads as stopped at +200%%:\n%s", out)
	}
}

// TestResearchPanelBonuses: the Stats panel's Research bonuses list is what the techs come
// to together. Their output bonuses are in a layer no cap holds, so none
// carries a note, even here, where the pools for all production and gold
// are past +200%. Cuts of a price or a time and the mechanic numbers
// read as what they are, and a first source's flat output as an amount.
func TestResearchPanelBonuses(t *testing.T) {
	_, st := cappedState(t)
	earned(t, st, "production_all") // the fixture's pools are past +200%
	earned(t, st, "gold_rate")
	out := section(t, statsProvider(st, 120), "Research bonuses", "Resource rates")
	for _, want := range []string{
		"[green]+5%    [-] All production\n",
		"[green]+31%   [-] Gold production\n",
		"[green]+86%   [-] Knowledge production\n",
		"[green]+25%   [-] Storage\n",
		"[green]+37%   [-] Housing\n",
		"[green]-19%   [-] Building costs\n",
		"[green]-27%   [-] Construction time\n",
		"[green]-23%   [-] Research time\n",
		"[green]-25%   [-] Wonder construction time\n",
		"[green]+25%   [-] Festival length\n",
		"[green]+20%   [-] Soldier storage\n",
		"[green]+21%   [-] Trade route income\n",
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
	if strings.Contains(out, "capped") || strings.Contains(out, "a quarter") {
		t.Errorf("a tech's bonus carries a cap note:\n%s", out)
	}
}

// TestFestivalSaysWhatItAddsPastTheKnee: the festival costs culture, so the
// command says before the player pays what it would add: with all production
// past +200% its +20% counts a quarter, +5%. It never reads as doing nothing.
func TestFestivalSaysWhatItAddsPastTheKnee(t *testing.T) {
	ge, _ := cappedState(t)
	if note := ge.FestivalStatus().CapNote; note != "counts a quarter past +200%: +5% now" {
		t.Fatalf("FestivalStatus().CapNote = %q past +200%%, want %q", note, "counts a quarter past +200%: +5% now")
	}
	out := cmdFestivalStatus(ge).Message
	if want := "Held now, the festival counts a quarter past +200%: +5% now (see stats)."; !strings.Contains(out, want) {
		t.Errorf("festival status is missing %q:\n%s", want, out)
	}
	if strings.Contains(out, "no effect") || strings.Contains(out, "capped") {
		t.Errorf("festival status reads as if the festival did nothing:\n%s", out)
	}
	fresh := game.NewGameEngine()
	if st := fresh.FestivalStatus(); st.CapNote != "" || st.CapPhrase != "" {
		t.Errorf("a new game's festival carries the note %q (%q), want none", st.CapNote, st.CapPhrase)
	}
	if out := cmdFestivalStatus(fresh).Message; strings.Contains(out, "capped") || strings.Contains(out, "a quarter") {
		t.Errorf("a new game's festival status mentions a cap:\n%s", out)
	}
}

// TestMilestonesAndWondersTagRewardsPastTheKnee: a milestone reward and a
// wonder effect in a pool past +200% say what they would add before the
// player works for them: a quarter of what they promise, never nothing and
// never the full figure.
func TestMilestonesAndWondersTagRewardsPastTheKnee(t *testing.T) {
	ge, st := cappedState(t)
	ms := milestonesProvider(st, 120)
	for _, want := range []string{
		"+20% all production[-] [yellow](counts a quarter past +200%: +5% now)[-]",
		"+10% all production[-] [yellow](counts a quarter past +200%: +2.5% now)[-]",
		"+5% all production[-] [yellow](counts a quarter past +200%: +1.25% now)[-]",
	} {
		if !strings.Contains(ms, want) {
			t.Errorf("the Milestones panel has no reward tagged %q:\n%s", want, ms)
		}
	}
	if strings.Contains(ms, "no effect") {
		t.Errorf("the Milestones panel shows a reward as doing nothing:\n%s", ms)
	}
	if strings.Contains(ms, "all production[-]\n") {
		t.Errorf("the Milestones panel lists an all-production reward with no note, as if all of it counted:\n%s", ms)
	}
	if want := "+5% research speed[-]\n"; !strings.Contains(ms, want) {
		t.Errorf("a research speed reward with room carries a note, or is gone (want %q)", want)
	}
	// The Information Age's wonder, the Global Network, adds knowledge
	// production to a pool that is already past +200%.
	if err := ge.EnterAgeForTest("information_age"); err != nil {
		t.Fatal(err)
	}
	wonders := wondersProvider(ge.GetState(), 120)
	if want := "+30% knowledge production[-] [yellow](counts a quarter past +200%: +7.5% now)[-]"; !strings.Contains(wonders, want) {
		t.Errorf("the Wonders panel has no wonder effect tagged %q:\n%s", want, wonders)
	}
	if want := "+30 data/tick[-]\n"; !strings.Contains(wonders, want) {
		t.Errorf("the Global Network's flat output carries a note, or is gone (want %q)", want)
	}
}

// TestRatesShowsEarnedAgainstApplied: the rates breakdown says what a pool
// has earned and what counts when the two differ: all production once, at
// the top, and a resource's own pool under that resource. A pool that
// applies all it has earned adds no line.
func TestRatesShowsEarnedAgainstApplied(t *testing.T) {
	ge, st := cappedState(t)
	out := HandleCommand("rates", ge).Message
	all := st.Pools["production_all"]
	for _, want := range []string{
		"  [yellow]All production: " + textfmt.SignedPercent(all.Earned) + " earned, " + applied(st, "production_all") + " counts. Past +200% a bonus counts a quarter.[-]",
		"    [yellow]Knowledge production: +335% earned, +234% counts. Past +200% a bonus counts a quarter.[-]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rates is missing %q:\n%s", want, out)
		}
	}
	// Nothing makes gold or iron in this fixture, so neither is listed, and
	// a pool is only ever shown under its resource.
	if strings.Contains(out, "Iron production:") || strings.Contains(out, "Gold production:") {
		t.Errorf("rates lists the pool of a resource it does not list:\n%s", out)
	}
	if fresh := HandleCommand("rates", game.NewGameEngine()).Message; strings.Contains(fresh, "earned") {
		t.Errorf("a new game's rates mention a pool:\n%s", fresh)
	}
}
