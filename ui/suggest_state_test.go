package ui

// Game-aware suggestions: argument completion reads the live game state.

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// checkBuildSuggestions: `build` offers only buildings buildable in this
// age (unlocked, not superseded, short of their limit), never a locked or a
// next-age one, the affordable ones first; `plan build` ranks the same way
// and then adds the next age's.
func checkBuildSuggestions(t *testing.T, eng *game.GameEngine) {
	t.Helper()
	st := eng.GetState()
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	comp := NewAutoCompleter(eng)
	got := comp("build ")
	if len(got) == 0 {
		t.Fatalf("%s: build offers nothing", st.Age)
	}
	affordable, seenUnaffordable := 0, false
	for _, line := range got {
		key := strings.TrimPrefix(line, "build ")
		bs, ok := st.Buildings[key]
		switch {
		case !ok:
			t.Errorf("%s: build offers %q, not a building", st.Age, key)
		case !bs.Unlocked:
			t.Errorf("%s: build offers locked %s", st.Age, key)
		case bs.IsLegacy || bs.AtMaxCount:
			t.Errorf("%s: build offers %s, which can't be built (legacy %v, at max %v)", st.Age, key, bs.IsLegacy, bs.AtMaxCount)
		case order[bs.AgeKey] > order[st.Age]:
			t.Errorf("%s: build offers %s from the later %s", st.Age, key, bs.AgeKey)
		}
		if bs.CanBuild {
			affordable++
			if seenUnaffordable {
				t.Errorf("%s: affordable %s sorts after an unaffordable building: %v", st.Age, key, got)
			}
		} else {
			seenUnaffordable = true
		}
	}
	if affordable == 0 || !seenUnaffordable {
		t.Errorf("%s: want affordable and unaffordable buildings to check the order, got %d of %d affordable", st.Age, affordable, len(got))
	}

	rank := func(key string) int {
		bs := st.Buildings[key]
		switch {
		case bs.AgeKey == st.NextAge && !bs.Unlocked:
			return 2
		case bs.CanBuild:
			return 0
		}
		return 1
	}
	last := 0
	for _, line := range comp("plan build ") {
		key := strings.TrimPrefix(line, "plan build ")
		r := rank(key)
		if r < last {
			t.Errorf("%s: plan build offers %s (rank %d) after rank %d", st.Age, key, r, last)
		}
		last = r
	}
}

func TestSuggestBuildGameAware(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	eng.Resources.Add("wood", 30)
	eng.Resources.Add("food", 30)
	checkBuildSuggestions(t, eng)

	// Later on, superseded buildings drop out and the age's new ones come in.
	if err := eng.SummonHarbingerForTest("bronze_age"); err != nil {
		t.Fatal(err)
	}
	for _, r := range []string{"wood", "stone", "food"} {
		eng.Resources.Add(r, 200)
	}
	checkBuildSuggestions(t, eng)
}

func TestSuggestResearchGameAware(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	// The first age with available techs of different costs; enough
	// knowledge for the cheapest but not the dearest.
	var st game.GameState
	lo, hi := -1.0, 0.0
	for _, age := range config.AgeOrder() {
		if err := eng.SummonHarbingerForTest(age); err != nil {
			t.Fatal(err)
		}
		st = eng.GetState()
		lo, hi = -1.0, 0.0
		for _, ts := range st.Research.Techs {
			if ts.Available {
				if lo < 0 || ts.Cost < lo {
					lo = ts.Cost
				}
				if ts.Cost > hi {
					hi = ts.Cost
				}
			}
		}
		if lo >= 0 && lo < hi {
			break
		}
	}
	if lo < 0 || lo == hi {
		t.Fatalf("no age has available techs of different costs")
	}
	eng.Resources.Add("knowledge", lo-st.Resources["knowledge"].Amount)
	st = eng.GetState()
	got := NewAutoCompleter(eng)("research ")
	seenDear := false
	techs := 0
	for _, line := range got {
		key := strings.TrimPrefix(line, "research ")
		ts, ok := st.Research.Techs[key]
		if !ok {
			continue // list, cancel
		}
		techs++
		if !ts.Available {
			t.Errorf("research offers %s, which can't start now", key)
		}
		if affordable := ts.Cost <= st.Resources["knowledge"].Amount; affordable && seenDear {
			t.Errorf("affordable %s sorts after an unaffordable tech: %v", key, got)
		} else if !affordable {
			seenDear = true
		}
	}
	if techs == 0 || !seenDear {
		t.Errorf("research offers %v; want available techs, affordable first", got)
	}
}

// TestSuggestFromGameState: the other game-aware slots.
func TestSuggestFromGameState(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	comp := NewAutoCompleter(eng)

	// Nothing is built yet: assign, unassign, dismiss and sell offer nothing.
	for _, line := range []string{"assign ", "unassign ", "sell ", "dismiss "} {
		if got := comp(line); len(got) != 0 {
			t.Errorf("%q on a fresh game = %v, want nothing", line, got)
		}
	}
	eng.Resources.Add("wood", 100)
	eng.Resources.Add("food", 100)
	var worker string
	for key, bs := range eng.GetState().Buildings {
		if bs.CanBuild && bs.WorkerDomain != "" && bs.WorkerCapacity > 0 {
			worker = key
			break
		}
	}
	if worker == "" {
		t.Fatal("no affordable building that takes workers")
	}
	if res := HandleCommand("build "+worker, eng); res.Type == "error" {
		t.Fatalf("build %s: %s", worker, res.Message)
	}
	for i := 0; i < 400 && eng.GetState().Buildings[worker].Count == 0; i++ {
		eng.StepTicks(1)
	}
	if got := comp("assign "); !slices.Contains(got, "assign "+worker) {
		t.Errorf("assign offers %v, want the built %s", got, worker)
	}
	// Nothing can be sold in the Primitive Age, so sell still offers nothing
	// (TestSuggestSellOnlyWhatSellTakes covers the later ages).
	if got := comp("sell "); len(got) != 0 {
		t.Errorf("sell offers %v in the Primitive Age, want nothing", got)
	}

	// Themes: the unlocked ones and list; diplomacy: only civilizations met.
	if got := comp("theme "); !slices.Contains(got, "theme "+theme.DefaultKey) || !slices.Contains(got, "theme list") {
		t.Errorf("theme offers %v", got)
	}
	for _, line := range comp("diplomacy ally ") {
		key := strings.TrimPrefix(line, "diplomacy ally ")
		if !eng.GetState().Diplomacy.Factions[key].Discovered {
			t.Errorf("diplomacy ally offers undiscovered %s", key)
		}
	}
	// load: the saves on disk.
	if err := eng.SaveGame("alpha"); err != nil {
		t.Fatal(err)
	}
	if got := comp("load al"); !slices.Equal(got, []string{"load alpha"}) {
		t.Errorf("load al = %v, want the alpha save", got)
	}
}

// TestSuggestSellOnlyWhatSellTakes: sell offers exactly the built buildings
// SellBuilding takes. It used to offer every built building, wonders and
// storage included, and then refuse them (storage became permanent in #163).
// The guard builds two of every building, takes the suggestions, then tries
// to sell one copy of each: a building is offered if and only if the sale
// goes through.
func TestSuggestSellOnlyWhatSellTakes(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	counts := map[string]int{}
	for key := range config.BuildingByKey() {
		counts[key] = 2
	}
	eng.Buildings.LoadCounts(counts)
	comp := NewAutoCompleter(eng)

	// The Primitive Age refuses every sale, so nothing is offered.
	if got := comp("sell "); len(got) != 0 {
		t.Errorf("sell offers %d buildings in the Primitive Age, want none: %v", len(got), got)
	}

	if err := eng.SummonHarbingerForTest("iron_age"); err != nil {
		t.Fatal(err)
	}
	offered := map[string]bool{}
	for _, line := range comp("sell ") {
		offered[strings.TrimPrefix(line, "sell ")] = true
	}
	if len(offered) == 0 {
		t.Fatal("sell offers nothing with every building built")
	}
	byKey := config.BuildingByKey()
	sawWonder, sawStorage := false, false
	for key := range counts {
		cat := byKey[key].Category
		sawWonder = sawWonder || cat == "wonder"
		sawStorage = sawStorage || cat == "storage"
		err := eng.SellBuilding(key, 1)
		switch {
		case offered[key] && err != nil:
			t.Errorf("sell offers %s (%s), but selling it is refused: %v", key, cat, err)
		case !offered[key] && err == nil:
			t.Errorf("sell does not offer %s (%s), but selling it goes through", key, cat)
		}
	}
	if !sawWonder || !sawStorage {
		t.Errorf("the guard built no wonder (%v) or no storage (%v); it checks nothing", sawWonder, sawStorage)
	}
}
