package ui

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// cappedState is a Modern Age player holding every tech so far and enough
// milestone rewards to put all production, gold and knowledge far past +200%,
// where a production pool still counts in full, with iron under it. The
// techs are in a layer of their own and fill no pool.
func cappedState(t *testing.T) (*game.GameEngine, game.GameState) {
	t.Helper()
	ge := game.NewGameEngine()
	ge.SeedRNG(1) // entering an era rolls an epoch event: the same one every run
	must(t, ge.EnterAgeForTest("modern_age"))
	ge.GrantTechsForTest()
	ge.GrantBonusForTest("production_all", 2.4)
	ge.GrantBonusForTest("gold_rate", 3.8)
	ge.GrantBonusForTest("knowledge_rate", 3.35)
	ge.GrantBonusForTest("iron_rate", 0.7)
	return ge, ge.GetState()
}

// large reads pool off st and fails unless it is far past +200% and applied
// in full: nothing is held back.
func large(t *testing.T, st game.GameState, pool string) game.BonusPool {
	t.Helper()
	p := st.Pools[pool]
	if p.Limited || p.Earned <= 2 || math.Abs(p.Applied-p.Earned) > 1e-9 {
		t.Fatalf("%s: earned %v, applied %v, limited %v; this fixture needs it past +200%% and in full", pool, p.Earned, p.Applied, p.Limited)
	}
	return p
}

// applied is the headline a pool shows: what the engine applies.
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

// TestStatsPanelShowsLargePoolsInFull: the Active multipliers headline is the
// whole of a production pool, however far past +200% it is, and carries no
// note. Only a pool a limit holds says so.
func TestStatsPanelShowsLargePoolsInFull(t *testing.T) {
	_, st := cappedState(t)
	for _, pool := range []string{"production_all", "gold_rate", "knowledge_rate"} {
		large(t, st, pool)
	}
	out := renderActiveMultipliers(st)
	if got := applied(st, "gold_rate"); got != "+380%" {
		t.Fatalf("gold has earned +380%% and applies %s, want +380%%", got)
	}
	for _, want := range []string{
		"Gold production     [-] [green]+380%[-]   ",
		"Scholar output      [-] [green]" + applied(st, "knowledge_rate") + "[-]   ",
		"Iron production     [-] [green]+70%[-]   ",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Active multipliers is missing %q:\n%s", want, out)
		}
	}
	for _, bad := range []string{"quarter", "earned", "capped at +200%", "[green]+200%[-] "} {
		if strings.Contains(out, bad) {
			t.Errorf("a pool far past +200%% carries %q:\n%s", bad, out)
		}
	}
}

// TestResearchPanelBonuses: the Stats panel's Research bonuses list is what the techs come
// to together. Their output bonuses are in a layer no cap holds, so none
// carries a note, even here, where the pools for all production and gold
// are past +200%. Cuts of a price or a time and the mechanic numbers
// read as what they are, and a first source's flat output as an amount.
func TestResearchPanelBonuses(t *testing.T) {
	_, st := cappedState(t)
	large(t, st, "production_all") // the fixture's pools are past +200%
	large(t, st, "gold_rate")
	out := section(t, statsProvider(st, 120), "Research bonuses", "Resource rates")
	for _, want := range []string{
		"[green]+7%    [-] All production\n",
		"[green]+36%   [-] Gold production\n",
		"[green]+86%   [-] Research production\n",
		"[green]+37%   [-] Housing\n",
		"[green]-19%   [-] Building costs\n",
		"[green]-27%   [-] Construction time\n",
		"[green]-23%   [-] Research time\n",
		"[green]-25%   [-] Wonder construction time\n",
		"[green]+25%   [-] Festival length\n",
		"[green]+10%   [-] Soldier posts\n",
		"[green]+21%   [-] Trade route income\n",
		"[green]+5     [-] Morale ceiling, in points\n",
		"[green]-28%   [-] Trade route time\n",
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
	if strings.Contains(out, "Market fee") {
		t.Errorf("the market has no fee to cut:\n%s", out)
	}
}

// TestFestivalSaysNothingOnALargePool: the festival's +20% counts in full
// whatever all production holds, so the command has no note to show before
// the player pays.
func TestFestivalSaysNothingOnALargePool(t *testing.T) {
	ge, st := cappedState(t)
	large(t, st, "production_all")
	if note := ge.FestivalStatus().CapNote; note != "" {
		t.Fatalf("FestivalStatus().CapNote = %q on a large pool, want none", note)
	}
	out := cmdFestivalStatus(ge).Message
	for _, bad := range []string{"Held now", "no effect", "capped", "a quarter"} {
		if strings.Contains(out, bad) {
			t.Errorf("festival status says %q on a pool that counts in full:\n%s", bad, out)
		}
	}
}

// TestMilestonesAndWondersTagOnlyLimitedRewards: a reward in a production
// pool far past +200% carries no note, since all of it counts. A reward in a
// pool a limit holds back (research speed, past +100%) still says what it
// would add: nothing.
func TestMilestonesAndWondersTagOnlyLimitedRewards(t *testing.T) {
	ge, _ := cappedState(t)
	ge.GrantBonusForTest("research_speed", 1.2)
	ms := milestonesProvider(ge.GetState(), 120)
	for _, want := range []string{
		"+20% all production[-]\n",
		"+10% all production[-]\n",
		"+5% all production[-]\n",
		"+5% research speed[-] [yellow](capped: no effect now)[-]",
	} {
		if !strings.Contains(ms, want) {
			t.Errorf("the Milestones panel is missing %q:\n%s", want, ms)
		}
	}
	if strings.Contains(ms, "a quarter") {
		t.Errorf("the Milestones panel still talks of a quarter:\n%s", ms)
	}
	// The Information Age's wonder, the Global Network, adds scholar
	// output to a pool that is already far past +200%.
	must(t, ge.EnterAgeForTest("information_age"))
	wonders := wondersProvider(ge.GetState(), 120)
	if want := "+30% scholar output[-]\n"; !strings.Contains(wonders, want) {
		t.Errorf("the Wonders panel tags a scholar output effect that counts in full (want %q):\n%s", want, wonders)
	}
	if strings.Contains(wonders, "a quarter") {
		t.Errorf("the Wonders panel still talks of a quarter:\n%s", wonders)
	}
	if want := "+30 data/tick[-]\n"; !strings.Contains(wonders, want) {
		t.Errorf("the Global Network's flat output carries a note, or is gone (want %q)", want)
	}
}

// TestRatesShowsEarnedAgainstApplied: the rates breakdown says what a pool
// has earned and what counts when the two differ, and only then: a pool far
// past +200% applies all it has earned and adds no line, and one held at the
// floor says so.
func TestRatesShowsEarnedAgainstApplied(t *testing.T) {
	ge, _ := cappedState(t)
	if out := HandleCommand("rates", ge).Message; strings.Contains(out, "earned") || strings.Contains(out, "counts.") {
		t.Errorf("rates lists a pool that applies all it has earned:\n%s", out)
	}
	// A debuff stack the floor holds: -140% earned, the pool counts -90%.
	floored := game.NewGameEngine()
	floored.GrantBonusForTest("production_all", -1.4)
	out := HandleCommand("rates", floored).Message
	if want := "  [yellow]All production: -140% earned, -90% counts.[-]"; !strings.Contains(out, want) {
		t.Errorf("rates is missing %q:\n%s", want, out)
	}
	if fresh := HandleCommand("rates", game.NewGameEngine()).Message; strings.Contains(fresh, "earned") {
		t.Errorf("a new game's rates mention a pool:\n%s", fresh)
	}
}
