package ui

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// newCaseTestEngine is an engine in the given age (the Primitive Age when
// empty) where every key-taking command in TestCommandKeysIgnoreCase can
// succeed: two huts, a market, a gathering camp with two of three workers on
// it, a known civilization, prestige points and deep stocks. Each call builds
// the same engine from the same seed, so two of them run the same command to
// the same result.
func newCaseTestEngine(t *testing.T, age string) *game.GameEngine {
	t.Helper()
	prevDev := game.DevModeActive
	game.DevModeActive = true
	t.Cleanup(func() { game.DevModeActive = prevDev })

	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	if age != "" {
		if msg := game.DevExecCommand("/age "+age, ge); msg != "jumped to "+age {
			t.Fatalf("dev /age %s: %q", age, msg)
		}
	}
	for _, b := range []string{"hut", "hut", "market", "gathering_camp"} {
		if msg := game.DevExecCommand("/build "+b, ge); msg != "built "+b {
			t.Fatalf("dev /build %s: %q", b, msg)
		}
	}
	for _, c := range []string{"recruit 3", "assign gathering_camp 2"} {
		if res := HandleCommand(c, ge); res.Type == "error" {
			t.Fatalf("%s: %s", c, res.Message)
		}
	}
	ge.Diplomacy.DiscoverFaction("riverlands_tribes")
	ge.Prestige.LoadState(1, 20, 20, nil)
	stock := map[string]float64{}
	for _, k := range []string{"food", "wood", "stone", "gold", "iron", "knowledge", "culture", "soldiers"} {
		stock[k] = 50000
	}
	ge.Resources.LoadStorage(stock)
	ge.Resources.LoadAmounts(stock)
	return ge
}

// comparableState is the engine's state without the wall-clock start and
// play time, which differ between any two engines.
func comparableState(ge *game.GameEngine) game.GameState {
	st := ge.GetState()
	st.Stats.GameStarted = time.Time{}
	st.Stats.PlayTime = 0
	return st
}

// TestCommandKeysIgnoreCase runs every command that takes a building, tech,
// resource, civilization, theme, expedition, trade route or prestige upgrade
// key twice, on two identical engines: once with the lowercase key and once in
// another case. Both must give the same result and leave the same game state.
// `build Hut` always worked; `sell Hut` and `dismiss Hut` used to fail.
func TestCommandKeysIgnoreCase(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	prevTheme := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prevTheme) })

	const stone, bronze, colonial = "stone_age", "bronze_age", "colonial_age"
	cases := []struct {
		age, lower, mixed string
		refused           bool // no setup here lets it succeed; the refusals, which name the key, are compared
	}{
		{"", "build hut", "build Hut", false},
		{"", "build hut 2", "build HUT 2", false},
		{stone, "sell hut", "sell Hut", false},
		{stone, "sell hut 1", "sell HUT 1", false},
		{"", "dismiss gathering_camp", "dismiss Gathering_Camp", false},
		{"", "dismiss gathering_camp all", "dismiss GATHERING_CAMP all", false},
		{"", "assign gathering_camp", "assign Gathering_Camp", false},
		{"", "unassign gathering_camp", "unassign GATHERING_CAMP", false},
		{"", "plan build hut", "plan build Hut", false},
		{"", "gather food", "gather Food", false},
		{"", "trade food wood 10", "trade Food WOOD 10", false},
		{"", "plan trade food wood", "plan trade FOOD Wood", false},
		{"", "theme " + theme.DefaultKey, "theme " + strings.ToUpper(theme.DefaultKey), false},
		{"", "prestige buy gather_boost", "prestige buy Gather_Boost", false},
		{bronze, "upgrade hut", "upgrade Hut", true},
		{bronze, "research tool_making", "research Tool_Making", false},
		{bronze, "research tool making", "research Tool Making", false},
		{bronze, "plan research tool_making", "plan research TOOL_MAKING", false},
		{bronze, "wonder collect stone 10", "wonder collect Stone 10", false},
		{bronze, "expedition scout_party", "expedition Scout_Party", false},
		{bronze, "expedition scout party", "expedition SCOUT PARTY", false},
		{bronze, "campaign raid_bandits", "campaign Raid_Bandits", false},
		{bronze, "trade route start local_barter", "trade route start Local_Barter", false},
		{bronze, "diplomacy gift riverlands_tribes", "diplomacy gift Riverlands_Tribes", false},
		{bronze, "diplomacy rival riverlands_tribes", "diplomacy rival RIVERLANDS_TRIBES", false},
		{bronze, "trade route stop local_barter", "trade route stop LOCAL_BARTER", true},
		{bronze, "diplomacy ally riverlands_tribes", "diplomacy ally Riverlands_Tribes", true},
		{bronze, "diplomacy neutral riverlands_tribes", "diplomacy neutral RIVERLANDS_TRIBES", false},
		{bronze, "diplomacy tribute riverlands_tribes", "diplomacy tribute Riverlands_Tribes", true},
		{bronze, "diplomacy raid riverlands_tribes", "diplomacy raid RIVERLANDS_TRIBES", false},
		{bronze, "diplomacy embargo riverlands_tribes", "diplomacy embargo Riverlands Tribes", false},
		{colonial, "blackmarket gold", "blackmarket Gold", false},
		{colonial, "trade black gold", "trade black GOLD", false},
	}
	for _, c := range cases {
		t.Run(c.mixed, func(t *testing.T) {
			want, got := newCaseTestEngine(t, c.age), newCaseTestEngine(t, c.age)
			wantRes := HandleCommand(c.lower, want)
			if wantRes.Type == "error" && !c.refused {
				t.Fatalf("%q failed, so the case check proves nothing: %s", c.lower, wantRes.Message)
			}
			gotRes := HandleCommand(c.mixed, got)
			if gotRes != wantRes {
				t.Errorf("%q = %+v, want the same as %q: %+v", c.mixed, gotRes, c.lower, wantRes)
			}
			if !reflect.DeepEqual(comparableState(got), comparableState(want)) {
				t.Errorf("%q left a different game state from %q", c.mixed, c.lower)
			}
		})
	}
}
