package game

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A gate met exactly leaves a town that can build the next age.
//
// The gate asks for a town, not for a pile of resources (config/gate_rule.go):
// the buildings it lists and the age's wonder. A town that met it with
// nothing to spare arrives in the next age with empty stores. This plays that
// town on the real engine: a new game, the town stood up as finished builds
// would (the listed buildings, the wonder, the techs a run cannot leave the
// age without, and nothing else), every store drained to zero just before
// 'advance', and then no help of any kind. A player buys the first copy of
// every building of the new age that a player with no optional tech can build
// on arrival (no tech, or one a run cannot leave the age behind without), cheapest
// first, as soon as the stores allow; each must be bought within the bound the
// gate covenant uses for what the town makes (smoke/static.go,
// firstCopyProblems): the age's pacing target, config.AgeTargetTicks. A
// building behind the new age's own keystone (the Bronze Age altar) waits for
// that tech, whose price and time are research and not measured here.
//
// The first gate is the short form. The next two are behind
// AGEFORGE_LONG_TESTS=1.

// gateTownStep is how many ticks the player waits between looks at his
// stores. The tick of a purchase is therefore known to this many.
const gateTownStep = 10

func TestFirstGateMetExactlyLeavesATownThatCanBuild(t *testing.T) {
	gateTownArrives(t, 0)
}

func TestLaterGatesMetExactlyLeaveATownThatCanBuild(t *testing.T) {
	if os.Getenv("AGEFORGE_LONG_TESTS") == "" {
		t.Skip("set AGEFORGE_LONG_TESTS=1 for the next two gates")
	}
	for g := 1; g <= 2; g++ {
		t.Run(config.Ages()[g].Key, func(t *testing.T) { gateTownArrives(t, g) })
	}
}

// gateTownArrives meets the gate that leaves ages[g] exactly, advances, and
// buys the first copy of everything buildable in ages[g+1].
func gateTownArrives(t *testing.T, g int) {
	t.Helper()
	ages, defs, techs := config.Ages(), config.BaseBuildings(), config.Technologies()
	from, to := ages[g], ages[g+1]
	held := config.GateTownHeldTechs(defs, ages, techs)

	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	ge := NewGameEngine()
	ge.SeedRNG(1)
	for j := 1; j <= g; j++ {
		if err := ge.EnterAgeForTest(ages[j].Key); err != nil {
			t.Fatalf("entering the %s: %v", ages[j].Name, err)
		}
	}
	_ = ge.ForceQuietFateForTest(from.EpochKey)
	if keys := sortedTrueKeys(held[g]); len(keys) > 0 {
		ge.GrantTechsForTest(keys...)
	}

	// The town: every wonder up to the age's and every building the gates so
	// far asked for, at the counts they asked.
	town := map[string]int{}
	for j := 0; j <= g; j++ {
		for _, d := range defs {
			if d.Category == "wonder" && d.RequiredAge == ages[j].Key {
				town[d.Key] = 1
			}
		}
		for k, n := range ages[j+1].BuildingReqs {
			town[k] = max(town[k], n)
		}
	}
	ge.SetTownForTest(town, 0)
	ge.SetTownForTest(nil, ge.GetState().Workers.MaxPop) // housed to the roof
	for _, res := range unlockedResources(ge) {
		ge.SetStockForTest(res, 0)
	}
	if err := ge.AdvanceAge(); err != nil {
		t.Fatalf("a town that met the gate exactly could not advance: %v", err)
	}
	st := ge.GetState()
	if st.Age != to.Key {
		t.Fatalf("advanced to %s, want %s", st.Age, to.Key)
	}
	for res, r := range st.Resources {
		if r.Amount != 0 {
			t.Fatalf("arrived with %v %s: the stores were drained before the advance", r.Amount, res)
		}
	}

	// What a player with no optional tech can build here.
	want := map[string]bool{}
	for _, d := range defs {
		if d.RequiredAge != to.Key || d.Category == "wonder" {
			continue
		}
		if d.RequiredTech == "" || held[g][d.RequiredTech] {
			want[d.Key] = true
		} else {
			t.Logf("%s: the first %s waits for %s, not measured", to.Name, d.Key, d.RequiredTech)
		}
	}
	if len(want) == 0 {
		t.Fatalf("no building of the %s to buy", to.Name)
	}

	bound := int(config.AgeTargetTicks(to.Key))
	bought := map[string]int{}
	for len(bought) < len(want) && st.Tick <= bound {
		var ready []string
		for k := range want {
			if _, done := bought[k]; !done && st.Buildings[k].CanBuild {
				ready = append(ready, k)
			}
		}
		sort.Slice(ready, func(a, b int) bool {
			ca, cb := costSum(st.Buildings[ready[a]].NextCost), costSum(st.Buildings[ready[b]].NextCost)
			if ca != cb {
				return ca < cb
			}
			return ready[a] < ready[b]
		})
		for _, k := range ready {
			if !ge.GetState().Buildings[k].CanBuild {
				continue // an earlier purchase this step took the stores
			}
			if err := ge.BuildBuilding(k); err != nil {
				t.Fatalf("buying the first %s: %v", k, err)
			}
			bought[k] = st.Tick
		}
		ge.StepTicks(gateTownStep)
		st = ge.GetState()
	}
	var late []string
	for k := range want {
		if _, done := bought[k]; !done {
			late = append(late, fmt.Sprintf("%s (%v)", k, ge.GetState().Buildings[k].NextCost))
		}
	}
	sort.Strings(late)
	if len(late) > 0 {
		t.Errorf("%s: first copies not affordable within the age's pacing target (%d ticks): %v", to.Name, bound, late)
	}
	for _, k := range sortedTrueKeys(want) {
		if tick, ok := bought[k]; ok {
			t.Logf("%s: the first %s bought at tick %d of %d", to.Name, k, tick, bound)
		}
	}
}

func sortedTrueKeys(m map[string]bool) []string {
	var out []string
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func unlockedResources(ge *GameEngine) []string {
	var out []string
	for res, r := range ge.GetState().Resources {
		if r.Unlocked {
			out = append(out, res)
		}
	}
	sort.Strings(out)
	return out
}

func costSum(cost map[string]float64) float64 {
	total := 0.0
	for _, v := range cost {
		total += v
	}
	return total
}
