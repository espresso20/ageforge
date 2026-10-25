package game

import (
	"math"
	"testing"
)

func prestigeCueEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := NewGameEngine()
	ge.age = "medieval_age"
	ge.Stats.AgeLog[0] = AgeEntry{Age: "primitive_age", Tick: 0}
	for key, count := range ge.progress.GetRequirementsForNext(ge.age) {
		ge.Buildings.counts[key] = count
	}
	if keystone := ge.rules.Keystone(ge.age); keystone != "" {
		ge.Research.researched[keystone] = true
	}
	for key, resource := range ge.Resources.resources {
		resource.Amount = 0
		resource.Rate = 10
		ge.Resources.resources[key] = resource
	}
	ge.Prestige.prestigeCoefficient = 1
	return ge
}

func TestPrestigeCueUsesGateShortfallAndRunElapsed(t *testing.T) {
	ge := prestigeCueEngine(t)
	wonder := ge.progress.WonderForAge(ge.age)
	if wonder == "" {
		t.Fatal("setup has no current-age wonder")
	}
	for res := range ge.Buildings.defs[wonder].BaseCost {
		resource := ge.Resources.resources[res]
		resource.Rate = 10
		ge.Resources.resources[res] = resource
	}
	var missingKey string
	for key := range ge.progress.GetRequirementsForNext(ge.age) {
		missingKey = key
		break
	}
	ge.Buildings.counts[missingKey]--
	ge.tick = 900
	cue := ge.prestigeCueLocked()
	need := map[string]float64{}
	buildingCost, ok := ge.Buildings.BuildBatchCost(missingKey, 1, ge.buildQueue)
	if !ok {
		t.Fatalf("unknown staged gate building %q", missingKey)
	}
	for res, cost := range buildingCost {
		need[res] += cost
	}
	for res, cost := range ge.Buildings.defs[wonder].BaseCost {
		need[res] += cost
	}
	remaining := 0.0
	for _, cost := range need {
		remaining = math.Max(remaining, cost/10)
	}
	wantFinish := int(math.Ceil(remaining))
	wantFresh := int(math.Ceil((900 + remaining) * cue.Multiplier / cue.NextMultiplier))
	if cue.FinishTicks != wantFinish || cue.FreshTicks != wantFresh {
		t.Fatalf("cue ticks finish=%d fresh=%d; want %d and %d", cue.FinishTicks, cue.FreshTicks, wantFinish, wantFresh)
	}
}

func TestPrestigeCueIncludesQueuedGateConstruction(t *testing.T) {
	ge := prestigeCueEngine(t)
	reqs := ge.progress.GetRequirementsForNext(ge.age)
	var want int
	for i, key := range sortedKeys(reqs) {
		ge.Buildings.counts[key]--
		ticks := 120 + i
		ge.buildQueue = append(ge.buildQueue, BuildQueueItem{BuildingKey: key, TicksLeft: ticks, TotalTicks: ticks})
		want = max(want, ticks)
	}
	wonder := ge.progress.WonderForAge(ge.age)
	ge.buildQueue = append(ge.buildQueue, BuildQueueItem{BuildingKey: wonder, TicksLeft: 250, TotalTicks: 250})
	want = max(want, 250)
	if got := ge.prestigeCueLocked().FinishTicks; got < want {
		t.Fatalf("FinishTicks = %d, want at least %d", got, want)
	}
}

func TestPrestigeCueClearsOnAgeAdvance(t *testing.T) {
	ge := prestigeCueEngine(t)
	ge.prestigeCue = PrestigeCue{Age: ge.age, Ready: true, FinishTicks: 12, Seen: true}
	ge.prestigeCueSeenAge = ge.age
	ge.advanceAge("renaissance_age")
	state := ge.GetState()
	if state.PrestigeCue.Ready || state.PrestigeCue.FinishTicks != 0 || state.PrestigeCue.Seen {
		t.Fatalf("cue after age advance = %+v, want cleared", state.PrestigeCue)
	}
}

func TestPrestigeCueReadinessGates(t *testing.T) {
	t.Run("floor blocks cue despite faster fresh run", func(t *testing.T) {
		ge := prestigeCueEngine(t)
		ge.Prestige.floor = ge.age
		cue := ge.prestigeCueLocked()
		if cue.FreshTicks >= cue.FinishTicks {
			t.Fatalf("test setup lacks favorable time comparison: %+v", cue)
		}
		if cue.Ready {
			t.Fatalf("cue is ready below prestige floor: %+v", cue)
		}
	})
	t.Run("cannot prestige", func(t *testing.T) {
		ge := prestigeCueEngine(t)
		ge.age = "stone_age"
		if ge.prestigeCueLocked().Ready {
			t.Fatal("cue ready before prestige is available")
		}
	})
	t.Run("equal multipliers", func(t *testing.T) {
		ge := prestigeCueEngine(t)
		ge.Prestige.prestigeCoefficient = 0
		if ge.prestigeCueLocked().Ready {
			t.Fatal("cue ready when next multiplier equals current")
		}
	})
	t.Run("zero income", func(t *testing.T) {
		ge := prestigeCueEngine(t)
		wonder := ge.progress.WonderForAge(ge.age)
		for res := range ge.Buildings.defs[wonder].BaseCost {
			resource := ge.Resources.resources[res]
			resource.Rate = 0
			ge.Resources.resources[res] = resource
			break
		}
		if ge.prestigeCueLocked().Ready {
			t.Fatal("cue ready with required resource shortfall and no income")
		}
	})
	t.Run("fresh is not sooner", func(t *testing.T) {
		ge := prestigeCueEngine(t)
		ge.tick = 100000000
		if ge.prestigeCueLocked().Ready {
			t.Fatal("cue ready when fresh run is not faster")
		}
	})
	t.Run("long finish and large multiplier", func(t *testing.T) {
		ge := prestigeCueEngine(t)
		ge.Prestige.prestigeCoefficient = 10
		for res := range ge.Buildings.defs[ge.progress.WonderForAge(ge.age)].BaseCost {
			resource := ge.Resources.resources[res]
			resource.Rate = 0.001
			ge.Resources.resources[res] = resource
		}
		if cue := ge.prestigeCueLocked(); !cue.Ready || cue.FinishTicks <= 0 {
			t.Fatalf("long gate cue = %+v, want ready", cue)
		}
	})
}

func TestPrestigeCueSeenPersistsAndResets(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := prestigeCueEngine(t)
	ge.MarkPrestigeCueSeen()
	if !ge.GetState().PrestigeCue.Seen {
		t.Fatal("seen cue missing from state")
	}
	loaded := saveAndLoad(t, ge, "prestige_cue")
	if !loaded.GetState().PrestigeCue.Seen {
		t.Fatal("seen cue did not survive save and load")
	}
	if err := loaded.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige: %v", err)
	}
	if loaded.GetState().PrestigeCue.Seen {
		t.Fatal("seen cue survived prestige")
	}
	loaded.MarkPrestigeCueSeen()
	succumbIn(t, loaded, "iron_age")
	if loaded.prestigeCueSeenAge != "" {
		t.Fatalf("seen age after Succumb = %q, want empty", loaded.prestigeCueSeenAge)
	}
	loaded.age = "medieval_age"
	loaded.MarkPrestigeCueSeen()
	loaded.Reset()
	if loaded.prestigeCueSeenAge != "" {
		t.Fatalf("seen age after Reset = %q, want empty", loaded.prestigeCueSeenAge)
	}
}

func TestPrestigeCueRunsOnlyEverySixtyTicks(t *testing.T) {
	ge := NewGameEngine()
	ge.prestigeCue = PrestigeCue{Age: "sentinel"}
	for i := 0; i < 59; i++ {
		ge.doTick()
	}
	if got := ge.prestigeCue.Age; got != "sentinel" {
		t.Fatalf("cue recomputed before 60 ticks: age %q", got)
	}
	ge.doTick()
	if got := ge.prestigeCue.Age; got != ge.age {
		t.Fatalf("cue age after 60 ticks = %q, want %q", got, ge.age)
	}
}
