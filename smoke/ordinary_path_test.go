package smoke

import (
	"fmt"
	"math"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The tests in this file drive the real engine along the ordinary player's
// purchase list (reference.go), with the resources granted instead of waited
// for. They take over what the bot's long runs used to show: that nothing on
// the way through the 22 ages is refused, that a save made in any age loads
// back to the same game, and that time spent away pays sanely from any of
// them.

// ordinaryPurchases is what the ordinary player buys in age i, in the order
// he buys it: the cheapest next copy, up to the reference town.
func ordinaryPurchases(t *refTables, i int) []config.BuildingDef {
	owned := map[string]int{}
	var out []config.BuildingDef
	for {
		best, bestCost := -1, 0.0
		for bi, d := range t.byAge[i] {
			if owned[d.Key] >= t.refCount(d) {
				continue
			}
			if u := units(copyPrice(d, owned[d.Key]), t.levels[i]); best < 0 || u < bestCost {
				best, bestCost = bi, u
			}
		}
		if best < 0 {
			return out
		}
		out = append(out, t.byAge[i][best])
		owned[t.byAge[i][best].Key]++
	}
}

// grant sets the engine's stock to what price asks, with room to spare.
func grant(ge *game.GameEngine, price map[string]float64) {
	for res, v := range price {
		ge.SetStockForTest(res, float64(2*v)+10)
	}
}

// walkOrdinary plays the ordinary player's list through every age and calls
// each on the engine as it stands in that age, before the advance out of it.
func walkOrdinary(t *testing.T, each func(age config.AgeDef, ge *game.GameEngine)) {
	t.Helper()
	restore := game.SetDataDirForTest(t.TempDir())
	defer restore()
	tables := newRefTables()
	wonders := map[string]config.BuildingDef{}
	for _, d := range config.BaseBuildings() {
		if d.Category == "wonder" {
			wonders[d.RequiredAge] = d
		}
	}
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	if !longTests() {
		mastered := map[string]int{}
		for _, age := range tables.ages {
			mastered[age.Key] = config.MasteryCap
		}
		ge.SetMasteryForTest(mastered, tables.ages[len(tables.ages)-1].Key)
	}
	for i, age := range tables.ages {
		if got := ge.GetState().Age; got != age.Key {
			t.Fatalf("the engine is in %s, the walk in %s", got, age.Key)
		}
		_ = ge.ForceQuietFateForTest(age.EpochKey)
		ge.GrantTechsForTest()
		longest := 0
		for n, d := range ordinaryPurchases(tables, i) {
			grant(ge, copyPrice(d, tables.refCount(d)))
			if err := ge.BuildBuilding(d.Key); err != nil {
				t.Fatalf("%s: purchase %d (%s) was refused: %v", age.Name, n+1, d.Key, err)
			}
			longest = max(longest, d.BuildTicks)
		}
		if w, ok := wonders[age.Key]; ok {
			grant(ge, w.BaseCost)
			for res := range w.BaseCost {
				if _, err := ge.BankWonderMax(w.Key, res); err != nil {
					t.Fatalf("%s: banking %s for the wonder (%s) was refused: %v", age.Name, res, w.Key, err)
				}
			}
			if err := ge.BuildBuilding(w.Key); err != nil {
				t.Fatalf("%s: the wonder (%s) was refused: %v", age.Name, w.Key, err)
			}
			longest = max(longest, w.BuildTicks)
		}
		if !longTests() {
			longest = int(math.Ceil(float64(longest) / config.MasteryK(config.MasteryCap)))
		}
		ge.StepTicks(longest + 5)
		if each != nil {
			each(age, ge)
		}
		if i+1 == len(tables.ages) {
			break
		}
		if err := ge.AdvanceAge(); err != nil {
			t.Fatalf("%s: the gate to the %s did not open: %v", age.Name, tables.ages[i+1].Name, err)
		}
	}
}

// longTests asks for the long form of the tests in this file: the walk on
// the frontier (every build takes its full time) and time away checked from
// every age. By default the walk runs on mastered ground, where the same
// purchases build about four times as fast, and time away is checked from
// every third age.
func longTests() bool { return os.Getenv("AGEFORGE_LONG_TESTS") != "" }

// TestOrdinaryPath walks the ordinary player's purchase list through all 22
// ages on the engine itself. Every purchase must be accepted, every wonder
// must build and every gate must open; a save made in each age must load
// into a fresh engine as the same game; and time away must pay sanely.
func TestOrdinaryPath(t *testing.T) {
	ages := 0
	walkOrdinary(t, func(age config.AgeDef, ge *game.GameEngine) {
		savedAndLoaded(t, age, ge)
		if longTests() || ages%3 == 0 {
			timeAway(t, age, ge)
		}
		ages++
	})
	if ages != 22 {
		t.Fatalf("walked %d ages, want 22", ages)
	}
}

// savedAndLoaded: a save made here loads into a fresh engine as the same
// game.
func savedAndLoaded(t *testing.T, age config.AgeDef, ge *game.GameEngine) {
	t.Helper()
	name := "walk_" + age.Key
	if err := ge.SaveGame(name); err != nil {
		t.Fatalf("%s: save: %v", age.Name, err)
	}
	back := game.NewGameEngine()
	if err := back.LoadGame(name); err != nil {
		t.Fatalf("%s: load: %v", age.Name, err)
	}
	if a, b := ge.StateDigest(), back.StateDigest(); a != b {
		t.Errorf("%s: the game loaded is not the game saved (digest %s, loaded %s)", age.Name, a, b)
	}
}

// timeAway: from a save made here, an hour away never pays more than the
// same hour played, and thirty hours away pay no more than the twenty-four
// the game allows.
func timeAway(t *testing.T, age config.AgeDef, ge *game.GameEngine) {
	t.Helper()
	const hour = int(3600 / config.TickSeconds)
	{
		name := "away_" + age.Key
		if err := ge.SaveGame(name); err != nil {
			t.Fatalf("%s: save: %v", age.Name, err)
		}
		load := func() *game.GameEngine {
			e := game.NewGameEngine()
			if err := e.LoadGame(name); err != nil {
				t.Fatalf("%s: load: %v", age.Name, err)
			}
			e.SeedRNG(7)
			return e
		}
		stock := func(e *game.GameEngine) map[string]float64 {
			out := map[string]float64{}
			for k, r := range e.GetState().Resources {
				out[k] = r.Amount
			}
			return out
		}
		start := stock(load())
		played := load()
		played.StepTicks(hour)
		away := load()
		away.SimulateOffline(time.Hour)
		day, more := load(), load()
		day.SimulateOffline(24 * time.Hour)
		more.SimulateOffline(30 * time.Hour)
		p, a, d, m := stock(played), stock(away), stock(day), stock(more)
		keys := make([]string, 0, len(start))
		for k := range start {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var notes []string
		for _, k := range keys {
			gainPlayed, gainAway := p[k]-start[k], a[k]-start[k]
			if gainPlayed > 0 && gainAway > float64(gainPlayed*1.05)+1 {
				notes = append(notes, fmt.Sprintf("%s: an hour away paid %.4g, the hour played %.4g", k, gainAway, gainPlayed))
			}
			if m[k] > float64(d[k]*1.001)+1 {
				notes = append(notes, fmt.Sprintf("%s: thirty hours away left %.4g, twenty-four left %.4g", k, m[k], d[k]))
			}
		}
		for _, n := range notes {
			t.Errorf("%s: %s", age.Name, n)
		}
	}
}
