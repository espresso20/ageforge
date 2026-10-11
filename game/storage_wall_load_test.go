package game

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A save from before the storage wall, loaded by the game after it. The
// fixture (testdata/pre_storage_wall_bronze.json) was written and signed by
// the old rules: 50 Stashes, 25 Storage Pits and 25 Warehouses, stores full
// to the old caps, and a plan with overflow banked into two of its items.

// wallFixtureSave is the fixture as the old game wrote it.
func wallFixtureSave(t *testing.T) GameSave {
	t.Helper()
	data, err := os.ReadFile(storageWallFixture)
	if err != nil {
		t.Fatal(err)
	}
	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		t.Fatal(err)
	}
	if save.StorageRules != 0 {
		t.Fatalf("the fixture carries storage_rules %d: it is not from before the wall", save.StorageRules)
	}
	return save
}

// wallFixtureView restores the fixture the way ViewSave does: the state and
// the load's own changes, with no time away played, so what the test reads
// is the load alone.
func wallFixtureView(t *testing.T) *GameEngine {
	t.Helper()
	isolateAccountDir(t)
	data, err := os.ReadFile(storageWallFixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath("wall"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	if err := ge.loadSave("wall", true); err != nil {
		t.Fatal(err)
	}
	return ge
}

func TestWallSave_KeepsEveryCopyAndEveryStock(t *testing.T) {
	save := wallFixtureSave(t)
	ge := wallFixtureView(t)
	if ge.cheaterBadge {
		t.Fatal("the old game's signature no longer verifies")
	}
	if len(save.Buildings) == 0 {
		t.Fatal("setup: the fixture holds no buildings")
	}
	for key, n := range save.Buildings {
		if got := ge.Buildings.GetCount(key); got != n {
			t.Errorf("%s: %d copies after the load, the save has %d", key, got, n)
		}
	}
	// A stock is what the save held plus what the plan had banked, put back.
	banked := map[string]float64{}
	for _, it := range save.Plan {
		for res, v := range it.Banked {
			banked[res] += v
		}
	}
	if len(banked) == 0 {
		t.Fatal("setup: the fixture's plan banked nothing")
	}
	over := 0
	for res, v := range save.Resources {
		want := v + banked[res]
		if got := ge.Resources.Get(res); got != want {
			t.Errorf("%s: %v after the load, want %v (saved %v, banked %v)", res, got, want, v, banked[res])
		}
		if ge.Resources.Get(res) > ge.Resources.GetStorage(res) {
			over++
		}
	}
	if over == 0 {
		t.Error("setup: no stock sits above its new store, so the fixture proves nothing about keeping it")
	}
	// The plan keeps its items and counts; the banks are empty.
	if len(ge.plan) != len(save.Plan) {
		t.Fatalf("%d plan items after the load, the save has %d", len(ge.plan), len(save.Plan))
	}
	for i, it := range ge.plan {
		if it.Key != save.Plan[i].Key || it.Count != save.Plan[i].Count {
			t.Errorf("plan item %d is %s x%d, the save has %s x%d", i, it.Key, it.Count, save.Plan[i].Key, save.Plan[i].Count)
		}
		if len(it.Banked) != 0 {
			t.Errorf("plan item %d still holds a bank: %v", i, it.Banked)
		}
	}
}

// Stock above a store gains nothing and is not cut by the ticks after the
// load; it falls only when spent, and a store it falls under fills again, to
// the cap and no further.
func TestWallSave_StockAboveAStoreHoldsUntilSpent(t *testing.T) {
	ge := wallFixtureView(t)
	ge.mu.Lock()
	ge.plan = nil // nothing of the plan's spends during the ticks
	ge.mu.Unlock()
	before := ge.Resources.GetAll()
	var over []string
	for res, v := range before {
		if v > ge.Resources.GetStorage(res) {
			over = append(over, res)
			if ge.Resources.GetRate(res) <= 0 {
				t.Fatalf("setup: %s is over its store with a rate of %v: the test could not tell holding from draining", res, ge.Resources.GetRate(res))
			}
		}
	}
	if len(over) == 0 {
		t.Fatal("setup: nothing is over its store")
	}
	ge.StepTicks(100)
	for _, res := range over {
		if got := ge.Resources.Get(res); got != before[res] {
			t.Errorf("%s was %v over a store of %v and is %v after 100 ticks: it must neither grow nor be cut", res, before[res], ge.Resources.GetStorage(res), got)
		}
	}
	for res, v := range before {
		if v <= ge.Resources.GetStorage(res) && ge.Resources.Get(res) > ge.Resources.GetStorage(res) {
			t.Errorf("%s grew past its store (%v of %v)", res, ge.Resources.Get(res), ge.Resources.GetStorage(res))
		}
	}

	// Spent: a price comes off the stock, and only that.
	cost := ge.Buildings.GetCost("house")
	if cost["wood"] <= 0 {
		t.Fatalf("setup: a house costs no wood (%v)", cost)
	}
	woodBefore := ge.Resources.Get("wood")
	if _, err := ge.BuildMultiple("house", 1); err != nil {
		t.Fatalf("a house: %v", err)
	}
	if got, want := ge.Resources.Get("wood"), woodBefore-cost["wood"]; got != want {
		t.Errorf("wood = %v after buying a house for %v, want %v", got, cost["wood"], want)
	}
	// Below its store, wood fills again and stops at the cap: a few ticks'
	// income short of it, twenty ticks later it is full and no more.
	store := ge.Resources.GetStorage("wood")
	low := store - float64(5*ge.Resources.GetRate("wood"))
	ge.mu.Lock()
	ge.Resources.Remove("wood", ge.Resources.Get("wood")-low)
	ge.mu.Unlock()
	ge.StepTicks(20)
	if got := ge.Resources.Get("wood"); got != store {
		t.Errorf("wood = %v after the stock fell to %v under a store of %v and 20 ticks, want it full at %v", got, low, store, store)
	}
}

// The first load says so once. The save it writes carries the storage rules,
// is signed, and loads without the line.
func TestWallSave_NoticeOnceAndSavedInTheNewShape(t *testing.T) {
	isolateAccountDir(t)
	count := func(ge *GameEngine) int {
		n := 0
		for _, l := range ge.GetLogs() {
			if strings.HasPrefix(l.Message, "Storage changed.") {
				n++
			}
		}
		return n
	}
	ge := loadFixture(t, storageWallFixture, "wall")
	if ge.cheaterBadge {
		t.Fatal("the old game's signature no longer verifies")
	}
	if n := count(ge); n != 1 {
		t.Fatalf("the notice appears %d times on the first load, want 1", n)
	}
	again := saveAndLoad(t, ge, "wall_again")
	if n := count(again); n != 0 {
		t.Errorf("the notice appears %d times on the second load, want 0", n)
	}
	data, err := os.ReadFile(savePath("wall_again"))
	if err != nil {
		t.Fatal(err)
	}
	var saved GameSave
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.StorageRules != config.StorageRules {
		t.Errorf("the re-saved game carries storage_rules %d, want %d", saved.StorageRules, config.StorageRules)
	}
	if again.cheaterBadge || saved.Signature == "" {
		t.Errorf("the re-saved game is not signed and valid (cheater %v, signature %q)", again.cheaterBadge, saved.Signature)
	}
}

// The next copy of the age's storage building is priced by the new curve, a
// rate of 1.75 on each copy owned, and it is refused when it is over its
// store, whatever the copies and stock the save kept.
func TestWallSave_NextStorageCopyIsOnTheNewCurveAndRefused(t *testing.T) {
	ge := wallFixtureView(t)
	var key string
	for _, d := range ge.rules.Buildings() {
		if d.Category == "storage" && d.RequiredAge == ge.age {
			key = d.Key
		}
	}
	if key == "" {
		t.Fatalf("setup: the %s has no storage building", ge.age)
	}
	def := ge.Buildings.defs[key]
	if def.CostScale != config.StorageRate {
		t.Fatalf("%s climbs at %v, want %v", key, def.CostScale, config.StorageRate)
	}
	n := ge.Buildings.GetCount(key)
	if n < 2 {
		t.Fatalf("setup: the save holds %d %s", n, key)
	}
	next := ge.Buildings.GetCost(key)
	ge.Buildings.counts[key] = n - 1
	prev := ge.Buildings.GetCost(key)
	ge.Buildings.counts[key] = n
	for res, p := range next {
		if ratio := p / prev[res]; math.Abs(ratio-config.StorageRate) > 1e-6 {
			t.Errorf("%s: copy %d costs %v of %s, copy %d %v: a ratio of %v, want %v", key, n+1, p, res, n, prev[res], ratio, config.StorageRate)
		}
	}
	res := ge.overStore(next)
	if res == "" {
		t.Fatalf("setup: the next %s (%v) fits every store the save holds", key, next)
	}
	err := ge.BuildBuilding(key)
	want := ge.storeTooSmallText(def.Name, res, next[res])
	if err == nil || err.Error() != want {
		t.Errorf("build %s = %v, want %q", key, err, want)
	}
	if ge.Buildings.GetCount(key) != n || queued(ge, key) != 0 {
		t.Errorf("the refused %s was bought", key)
	}
}
