package game

import (
	"os"
	"testing"
)

// The save the storage wall has to load: a Bronze Age town built the way the
// old rules allowed and a patient player built it. It holds every storage
// copy the old limits sold (50 Stashes, 25 Storage Pits, 25 Warehouses), 20
// Forager Posts, stores full to the old caps in wood and stone and a large
// knowledge stock, and a plan whose next Warehouse and next House had already
// banked part of their price from overflow.
const storageWallFixture = "testdata/pre_storage_wall_bronze.json"

// TestWriteStorageWallFixture regenerates the save. Run it with
// UPDATE_FIXTURES=1 on a checkout of the game from before the storage wall:
// the point of the committed bytes is that the old game wrote and signed
// them.
func TestWriteStorageWallFixture(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the pre storage wall save")
	}
	isolateAccountDir(t)
	ge := newSeededEngine(73)
	ge.StepTicks(30)
	ge.mu.Lock()
	ge.Stats.AgesReached = nil
	const age = "bronze_age"
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
	early := map[string]bool{"primitive_age": true, "stone_age": true, "bronze_age": true}
	for _, tech := range ge.rules.Techs() {
		if early[tech.Age] {
			ge.Research.researched[tech.Key] = true
		}
	}
	for key, n := range map[string]int{
		"stash": 50, "storage_pit": 25, "warehouse": 25,
		"hut": 32, "longhouse": 40, "house": 12,
		"gathering_camp": 15, "forager_post": 20, "story_circle": 15, "elders_hall": 21,
		"wood_camp": 15, "woodcutter_camp": 21, "stone_camp": 30, "stone_pit": 22,
		"farm": 6, "quarry": 6, "lumber_mill": 6, "scriptorium": 5, "smithy": 4, "market": 3,
	} {
		ge.Buildings.counts[key] = n
	}
	ge.Research.rebuildBonuses()
	ge.recalculateRates()
	for _, res := range []string{"wood", "stone", "food"} {
		ge.Resources.Add(res, ge.Resources.GetStorage(res))
	}
	ge.Resources.Add("knowledge", 372655-ge.Resources.Get("knowledge"))
	ge.Resources.Add("iron", 4000-ge.Resources.Get("iron"))
	ge.Resources.Add("gold", 2500-ge.Resources.Get("gold"))
	ge.plan = []PlanItem{
		{Kind: PlanBuild, Key: "warehouse", Count: 3, Banked: map[string]float64{"wood": 21000, "stone": 9000}},
		{Kind: PlanBuild, Key: "house", Count: 10, Banked: map[string]float64{"wood": 1500}},
		{Kind: PlanBuild, Key: "farm", Count: 20},
	}
	ge.mu.Unlock()
	if err := ge.SaveGame("pre_storage_wall_bronze"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(savePath("pre_storage_wall_bronze"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(storageWallFixture, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
