//go:build rawmaterialsfixture

package game

import (
	"os"
	"testing"
)

// The saves the raw materials have to load. Each is a town as the game wrote
// it before iron ore, marble, titanium ore and dark matter crystals were
// priced: a Modern Age town with Titanium Mines (titanium ore was still shut
// in that age) and Modern Depots, a Digital Age town with all three titanium
// mines and three kinds of depot, and an Iron
// Age town whose Colosseum bank was full at the old price, with a Building in
// the build queue and two plan items that cost a raw now.
//
// This file builds only with -tags rawmaterialsfixture. Run it on a checkout
// of the game from before the change, so the bytes are the old game's,
// written and signed by it:
//
//	UPDATE_FIXTURES=1 go test -tags rawmaterialsfixture ./game -run TestWriteRawMaterialsFixtures
//
// then commit testdata/pre_raw_materials_*.json.
func TestWriteRawMaterialsFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the pre raw materials saves")
	}
	reach := func(ge *GameEngine, age string) {
		ge.mu.Lock()
		defer ge.mu.Unlock()
		ge.Stats.AgesReached = nil
		for _, a := range ageKeys() {
			ge.Stats.AgesReached = append(ge.Stats.AgesReached, a)
			ge.applyAgeUnlocks(a)
			if a == age {
				break
			}
			if w := ge.progress.WonderForAge(a); w != "" {
				ge.Buildings.counts[w] = 1
			}
		}
		ge.age = age
		ge.currentEpoch = ge.rules.EraOf(age)
		ge.Workers.SetAge(age)
		// The research of the ages before this one is done, as in a town
		// that played them.
		before := map[string]bool{}
		for _, a := range ageKeys() {
			before[a] = true
			if a == age {
				break
			}
		}
		for _, tech := range ge.rules.Techs() {
			if before[tech.Age] {
				ge.Research.researched[tech.Key] = true
			}
		}
		ge.Research.rebuildBonuses()
		ge.recalculateRates()
	}
	write := func(ge *GameEngine, name string) {
		t.Helper()
		if err := ge.SaveGame(name); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(savePath(name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("testdata/"+name+".json", data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("modern", func(t *testing.T) {
		isolateAccountDir(t)
		ge := newSeededEngine(91)
		reach(ge, "modern_age")
		ge.mu.Lock()
		ge.Buildings.counts["titanium_mine"] = 4
		ge.Buildings.counts["modern_depot"] = 3
		ge.mu.Unlock()
		ge.recalculateRates()
		write(ge, "pre_raw_materials_modern")
	})

	t.Run("digital", func(t *testing.T) {
		isolateAccountDir(t)
		ge := newSeededEngine(92)
		reach(ge, "digital_age")
		ge.mu.Lock()
		ge.Buildings.counts["titanium_mine"] = 3
		ge.Buildings.counts["precision_mine"] = 3
		ge.Buildings.counts["nano_drill_complex"] = 3
		ge.Buildings.counts["modern_depot"] = 3
		ge.Buildings.counts["info_vault"] = 2
		ge.Buildings.counts["digital_archive"] = 2
		ge.mu.Unlock()
		ge.recalculateRates()
		write(ge, "pre_raw_materials_digital")
	})

	t.Run("iron", func(t *testing.T) {
		isolateAccountDir(t)
		ge := newSeededEngine(93)
		reach(ge, "iron_age")
		ge.mu.Lock()
		ge.Buildings.counts["marble_quarry"] = 3
		ge.Buildings.counts["ironworks"] = 2
		ge.Buildings.counts["granary"] = 4
		ge.Buildings.counts["warehouse"] = 3
		ge.mu.Unlock()
		ge.recalculateRates()
		// The Colosseum's bank, full at the old price (stone, iron, gold).
		for res, amt := range map[string]float64{"stone": 512000, "iron": 113600, "gold": 100800} {
			setResource(ge, res, amt)
			if _, err := ge.BankWonderMax("colosseum", res); err != nil {
				t.Fatal(err)
			}
		}
		if !ge.Buildings.IsWonderBankFull("colosseum") {
			t.Fatal("setup: the Colosseum's bank is not full at the old price")
		}
		// A little left in hand, so the plan has something to read.
		setResource(ge, "stone", 20000)
		setResource(ge, "iron", 12000)
		setResource(ge, "gold", 9000)
		setResource(ge, "wood", 40000)
		// A building in the queue (paid when it was queued) and two plan
		// items; the second had banked part of its price from overflow.
		if err := ge.BuildBuilding("agora"); err != nil {
			t.Fatal(err)
		}
		if len(ge.buildQueue) != 1 {
			t.Fatalf("setup: the build queue holds %d items, want 1", len(ge.buildQueue))
		}
		ge.mu.Lock()
		ge.plan = []PlanItem{
			{Kind: PlanBuild, Key: "timber_yard", Count: 2},
			{Kind: PlanBuild, Key: "legion_fort", Count: 1, Banked: map[string]float64{"stone": 14000, "iron": 7000}},
		}
		ge.mu.Unlock()
		write(ge, "pre_raw_materials_iron")
	})
}
