package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// The upgrade engine: an older building keeps producing and is upgraded, one
// copy at a time, to its line's building for the town's age at
// config.UpgradeShare of what the next new copy of that building costs.

// richTown gives the town stores that hold far more than any price here (the
// storage comes from buildings, so a recount keeps it) and fills them.
func richTown(ge *GameEngine) {
	ge.Buildings.counts["stash"] = 500
	ge.recalculateRates()
	for res, store := range ge.Resources.GetAllStorage() {
		setAmount(ge, res, math.Min(store, 40000))
	}
}

// campsInTheStoneAge is a Stone Age game holding n Gathering Camps (older: the
// Forager Post is the food line's building for the age), with money for any
// price.
func campsInTheStoneAge(t *testing.T, n int) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	ge.Buildings.counts["gathering_camp"] = n
	ge.advanceAge("stone_age")
	ge.pendingCatastrophe = ""
	if to, ok := ge.Buildings.UpgradeTarget("gathering_camp"); !ok || to.Key != "forager_post" {
		t.Fatalf("setup: gathering_camp upgrades to %q (%v) in the Stone Age", to.Key, ok)
	}
	richTown(ge)
	return ge
}

func halfOf(m map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(m))
	for res, v := range m {
		out[res] = math.Floor(float64(v * 0.5))
	}
	return out
}

func TestUpgradeShareIsHalf(t *testing.T) {
	if config.UpgradeShare != 0.5 {
		t.Errorf("UpgradeShare = %v, want 0.5", config.UpgradeShare)
	}
}

func TestUpgrade_OneCopyCostsHalfANewCopyAndMovesCounts(t *testing.T) {
	ge := campsInTheStoneAge(t, 4)
	newCost, _ := ge.Buildings.BuildBatchCost("forager_post", 1, ge.buildQueue)
	want := halfOf(newCost)
	if got := ge.Buildings.UpgradeCopyCost("gathering_camp", 0, ge.buildQueue); !mapsEqualF(got, want) {
		t.Fatalf("the price of an upgrade is %v, want half of a new Forager Post, %v", got, want)
	}
	// A priced list row says the same.
	rows := ge.GetAvailableUpgrades()
	if len(rows) == 0 || !mapsEqualF(rows[0].Cost, want) {
		t.Fatalf("upgrade list rows = %+v, want next price %v", rows, want)
	}

	before := map[string]float64{}
	for res := range want {
		before[res] = ge.Resources.Get(res)
	}
	if err := ge.UpgradeBuilding("gathering_camp", 1, false); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.GetCount("gathering_camp") != 3 || ge.Buildings.GetCount("forager_post") != 1 {
		t.Errorf("counts after one upgrade: camps %d, posts %d; want 3 and 1", ge.Buildings.GetCount("gathering_camp"), ge.Buildings.GetCount("forager_post"))
	}
	for res, price := range want {
		if got := before[res] - ge.Resources.Get(res); math.Abs(got-price) > 1e-6 {
			t.Errorf("%s paid %v, want %v", res, got, price)
		}
	}
	if !logHas(ge, "Upgraded 1 Gathering Camp to Forager Post") {
		t.Error("no line says what was upgraded")
	}
}

func mapsEqualF(a, b map[string]float64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}

// Each copy is priced after the last: the target's count is one higher, so the
// next upgrade costs more, and a batch costs the sum of its copies.
func TestUpgrade_CopiesArePricedOneAtATime(t *testing.T) {
	ge := campsInTheStoneAge(t, 5)
	var sum float64
	var prev float64
	for i := 0; i < 3; i++ {
		price := ge.Buildings.UpgradeCopyCost("gathering_camp", i, nil)
		newCost := ge.Buildings.CopyCost("forager_post", i, nil)
		if !mapsEqualF(price, halfOf(newCost)) {
			t.Fatalf("copy %d: price %v, want half of %v", i, price, newCost)
		}
		total := 0.0
		for _, v := range price {
			total += v
		}
		if total <= prev {
			t.Errorf("copy %d costs %v, no more than copy %d (%v)", i, total, i-1, prev)
		}
		prev = total
		sum += total
	}
	before := map[string]float64{}
	for _, res := range []string{"wood", "stone", "food", "gold", "iron"} {
		before[res] = ge.Resources.Get(res)
	}
	if err := ge.UpgradeBuilding("gathering_camp", 3, false); err != nil {
		t.Fatal(err)
	}
	paid := 0.0
	for res, v := range before {
		paid += v - ge.Resources.Get(res)
	}
	if math.Abs(paid-sum) > 1e-6 {
		t.Errorf("three upgrades paid %v, want the sum of three single prices %v", paid, sum)
	}
	// The next copy is the fourth: priced with three posts standing.
	next := ge.Buildings.UpgradeCopyCost("gathering_camp", 0, nil)
	if want := halfOf(ge.Buildings.CopyCost("forager_post", 0, nil)); !mapsEqualF(next, want) {
		t.Errorf("next price %v, want %v", next, want)
	}
}

// `upgrade <building>` with no count does as many copies as the town can pay
// for, and says what the next one costs or why it stopped.
func TestUpgrade_DefaultTakesWhatTheTownCanPay(t *testing.T) {
	ge := campsInTheStoneAge(t, 6)
	// Money for exactly two copies of every resource, not three.
	need := map[string]float64{}
	for i := 0; i < 2; i++ {
		for res, v := range ge.Buildings.UpgradeCopyCost("gathering_camp", i, nil) {
			need[res] += v
		}
	}
	for res := range ge.Resources.GetAllStorage() {
		setAmount(ge, res, 0)
	}
	for res, v := range need {
		setAmount(ge, res, v)
	}
	rows := ge.GetAvailableUpgrades()
	if len(rows) != 1 || rows[0].Affordable != 2 || rows[0].Count != 6 {
		t.Fatalf("list = %+v, want 6 held and 2 affordable", rows)
	}
	if err := ge.UpgradeBuilding("gathering_camp", 0, false); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.GetCount("forager_post") != 2 || ge.Buildings.GetCount("gathering_camp") != 4 {
		t.Errorf("posts %d, camps %d; want 2 and 4", ge.Buildings.GetCount("forager_post"), ge.Buildings.GetCount("gathering_camp"))
	}
	if !logHasAll(ge, "Upgraded 2 Gathering Camps to Forager Posts", "Stopped: Cannot afford to upgrade a Gathering Camp") {
		t.Errorf("the reply does not say how many were done and why it stopped: %v", lastLine(ge))
	}
	// With nothing left to spend, nothing is done and the reason is the error.
	for res := range ge.Resources.GetAllStorage() {
		setAmount(ge, res, 0)
	}
	err := ge.UpgradeBuilding("gathering_camp", 0, false)
	if err == nil || !strings.Contains(err.Error(), "Cannot afford to upgrade a Gathering Camp") {
		t.Errorf("with no money: err = %v, want a cannot-afford refusal", err)
	}
}

func lastLine(ge *GameEngine) string {
	if len(ge.log) == 0 {
		return ""
	}
	return ge.log[len(ge.log)-1].Message
}

func TestUpgrade_CountAndAll(t *testing.T) {
	ge := campsInTheStoneAge(t, 5)
	if err := ge.UpgradeBuilding("gathering_camp", 2, false); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.GetCount("forager_post") != 2 {
		t.Errorf("upgrade <2> made %d posts", ge.Buildings.GetCount("forager_post"))
	}
	if !strings.Contains(lastLine(ge), "The next copy costs") {
		t.Errorf("a counted upgrade does not say what the next copy costs: %q", lastLine(ge))
	}
	// A count above what is held does what is held.
	if err := ge.UpgradeBuilding("gathering_camp", 50, false); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.GetCount("gathering_camp") != 0 || ge.Buildings.GetCount("forager_post") != 5 {
		t.Errorf("after upgrading past what is held: camps %d, posts %d", ge.Buildings.GetCount("gathering_camp"), ge.Buildings.GetCount("forager_post"))
	}
	if err := ge.UpgradeBuilding("gathering_camp", 0, true); err == nil || !strings.Contains(err.Error(), "You have no Gathering Camp to upgrade") {
		t.Errorf("upgrade all with none held: err = %v", err)
	}

	all := campsInTheStoneAge(t, 7)
	if err := all.UpgradeBuilding("gathering_camp", 0, true); err != nil {
		t.Fatal(err)
	}
	if all.Buildings.GetCount("forager_post") != 7 || all.Buildings.IsOlder("forager_post") {
		t.Errorf("upgrade all: %d posts", all.Buildings.GetCount("forager_post"))
	}
	if got := all.GetAvailableUpgrades(); len(got) != 0 {
		t.Errorf("the list still holds %+v after everything was upgraded", got)
	}
}

// `all` stops at the copy it cannot pay for, however large the store: the
// price is judged against what is held, not against what a store holds. It
// says how many it did and what the next copy lacks.
func TestUpgrade_AllStopsAtTheCopyYouCannotPay(t *testing.T) {
	ge := campsInTheStoneAge(t, 6)
	p0 := ge.Buildings.UpgradeCopyCost("gathering_camp", 0, nil)
	p1 := ge.Buildings.UpgradeCopyCost("gathering_camp", 1, nil)
	var dear string
	for res := range p1 {
		if p1[res] > p0[res] && (dear == "" || p1[res] > p1[dear]) {
			dear = res
		}
	}
	if dear == "" {
		t.Fatal("setup: no resource gets dearer from the first copy to the second")
	}
	// Enough of every resource for six copies but one: the first copy's price
	// of the dear one, and less than the two first copies' together.
	for res, v := range p1 {
		setAmount(ge, res, v*100)
	}
	have := p0[dear] + p1[dear] - 1
	setAmount(ge, dear, have)
	rows := ge.GetAvailableUpgrades()
	if len(rows) != 1 || rows[0].Affordable != 1 || strings.Contains(rows[0].Lock, "storage") {
		t.Fatalf("list = %+v, want 1 affordable and no lock about a store", rows)
	}
	if err := ge.UpgradeBuilding("gathering_camp", 0, true); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.GetCount("forager_post") != 1 {
		t.Errorf("upgrade all did %d copies, want the 1 it could pay", ge.Buildings.GetCount("forager_post"))
	}
	if !logHasAll(ge, "Upgraded 1 Gathering Camp to Forager Post", "Stopped: Cannot afford to upgrade") || logHas(ge, "storage holds") {
		t.Errorf("the reply does not say what the next copy lacks: %q", lastLine(ge))
	}
	// A copy that cannot be paid from the start is refused the same way and
	// nothing moves.
	fresh := campsInTheStoneAge(t, 3)
	setAmount(fresh, dear, p0[dear]-1)
	err := fresh.UpgradeBuilding("gathering_camp", 0, true)
	if err == nil || !strings.HasPrefix(err.Error(), "Cannot afford to upgrade") {
		t.Errorf("upgrade without the price: err = %v", err)
	}
	if fresh.Buildings.GetCount("gathering_camp") != 3 || fresh.Buildings.GetCount("forager_post") != 0 {
		t.Error("the refused upgrade moved a copy")
	}
}

// A target that waits for a tech refuses, naming the tech; it opens when the
// tech is done.
func TestUpgrade_TechLockNamesTheTech(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["shrine"] = 3
	ge.advanceAge("stone_age")
	ge.pendingCatastrophe = ""
	richTown(ge)
	to, ok := ge.Buildings.UpgradeTarget("shrine")
	if !ok || to.RequiredTech == "" {
		t.Fatalf("setup: shrine upgrades to %q needing %q", to.Key, to.RequiredTech)
	}
	err := ge.UpgradeBuilding("shrine", 0, true)
	tech := ge.techName(to.RequiredTech)
	if err == nil || !strings.Contains(err.Error(), tech) {
		t.Fatalf("upgrade with the tech undone: err = %v, want the tech %q named", err, tech)
	}
	if ge.Buildings.GetCount("shrine") != 3 {
		t.Error("a locked upgrade moved copies")
	}
	rows := ge.GetAvailableUpgrades()
	if len(rows) != 1 || rows[0].Affordable != 0 || !strings.Contains(rows[0].Lock, tech) {
		t.Errorf("list = %+v, want the lock to name %q", rows, tech)
	}
	learn(ge, to.RequiredTech)
	if err := ge.UpgradeBuilding("shrine", 0, true); err != nil {
		t.Fatalf("upgrade after the tech: %v", err)
	}
	if ge.Buildings.GetCount(to.Key) != 3 {
		t.Errorf("%d %s after the upgrade, want 3", ge.Buildings.GetCount(to.Key), to.Key)
	}
}

// A hut goes straight to the building of the age, not a rung at a time.
func TestUpgrade_GoesStraightToTheCurrentRung(t *testing.T) {
	ge := newSeededEngine(1)
	ge.age = "iron_age"
	ge.Buildings.counts["hut"] = 3
	richTown(ge)
	to, ok := ge.Buildings.UpgradeTarget("hut")
	if !ok || to.Key != "townhouse" {
		t.Fatalf("a hut upgrades to %q (%v) in the Iron Age, want townhouse", to.Key, ok)
	}
	if to.RequiredTech != "" {
		learn(ge, to.RequiredTech)
	}
	if err := ge.UpgradeBuilding("hut", 2, false); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.GetCount("townhouse") != 2 || ge.Buildings.GetCount("hut") != 1 || ge.Buildings.GetCount("longhouse") != 0 || ge.Buildings.GetCount("house") != 0 {
		t.Errorf("huts %d, townhouses %d, longhouses %d, houses %d", ge.Buildings.GetCount("hut"), ge.Buildings.GetCount("townhouse"), ge.Buildings.GetCount("longhouse"), ge.Buildings.GetCount("house"))
	}
	// An upgraded copy is not older and has nothing to upgrade to.
	if err := ge.UpgradeBuilding("townhouse", 1, false); err == nil || !strings.Contains(err.Error(), "has no upgrade this age") {
		t.Errorf("upgrading the current rung: err = %v", err)
	}
}

// A line that pauses has no target for its last rung while it pauses, and the
// older copies of the line still go to that rung; it has one again when the
// line resumes.
func TestUpgrade_PausedLine(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["aerospace_foundry"] = 2
	ge.Buildings.counts["nano_alloy_plant"] = 2
	richTown(ge)
	ge.age = "fusion_age" // the titanium line has no rung of its own here
	if _, ok := ge.Buildings.UpgradeTarget("nano_alloy_plant"); ok {
		t.Error("nano_alloy_plant is the titanium line's rung for the Fusion Age, but has a target")
	}
	if err := ge.UpgradeBuilding("nano_alloy_plant", 0, true); err == nil || !strings.Contains(err.Error(), "has no upgrade this age") {
		t.Errorf("upgrade while paused: err = %v", err)
	}
	to, ok := ge.Buildings.UpgradeTarget("aerospace_foundry")
	if !ok || to.Key != "nano_alloy_plant" {
		t.Fatalf("aerospace_foundry upgrades to %q (%v) while the line is paused, want nano_alloy_plant", to.Key, ok)
	}
	ge.age = "space_age"
	to, ok = ge.Buildings.UpgradeTarget("nano_alloy_plant")
	if !ok || to.Key != "orbital_refinery" {
		t.Errorf("nano_alloy_plant upgrades to %q (%v) when the line resumes, want orbital_refinery", to.Key, ok)
	}
}

// The cap of the target holds for an upgrade.
func TestUpgrade_StopsAtTheTargetsMaxCount(t *testing.T) {
	ge := campsInTheStoneAge(t, 5)
	limited := ge.Buildings.defs["forager_post"]
	limited.MaxCount = 2
	ge.Buildings.defs["forager_post"] = limited
	ge.Buildings.rebuildOrder()
	if err := ge.UpgradeBuilding("gathering_camp", 0, true); err != nil {
		t.Fatal(err)
	}
	if got := ge.Buildings.GetCount("forager_post"); got != 2 {
		t.Errorf("%d posts after upgrading all, want the cap 2", got)
	}
	if !strings.Contains(lastLine(ge), "Stopped: Forager Post is at its max count of 2") {
		t.Errorf("the reply does not say the cap stopped it: %q", lastLine(ge))
	}
	if err := ge.UpgradeBuilding("gathering_camp", 1, false); err == nil || !strings.Contains(err.Error(), "max count") {
		t.Errorf("upgrade into a full target: err = %v", err)
	}
}

// The copy keeps its workers: they move from the old building's slots to the
// target's, and what the target's slots can't hold goes idle.
func TestUpgrade_CopyKeepsItsWorkers(t *testing.T) {
	ge := campsInTheStoneAge(t, 4)
	oldCap := ge.Buildings.defs["gathering_camp"].WorkerCapacity
	ge.Workers.UnlockType("worker")
	ge.Workers.Recruit("worker", 4*oldCap+5, 1<<20)
	if !ge.Workers.Assign("worker", "gathering_camp", 4*oldCap) {
		t.Fatal("setup: assign failed")
	}
	newCap := ge.Buildings.defs["forager_post"].WorkerCapacity
	if err := ge.UpgradeBuilding("gathering_camp", 1, false); err != nil {
		t.Fatal(err)
	}
	moved := min(oldCap, newCap)
	if got := ge.Workers.GetAssignedCount("worker", "forager_post"); got != moved {
		t.Errorf("the upgraded copy has %d workers, want %d (its %d, to a building of %d slots)", got, moved, oldCap, newCap)
	}
	if got := ge.Workers.GetAssignedCount("worker", "gathering_camp"); got != 3*oldCap {
		t.Errorf("the three camps left have %d workers, want %d", got, 3*oldCap)
	}

	// A target with fewer slots: the surplus goes idle, not lost.
	few := campsInTheStoneAge(t, 2)
	post := few.Buildings.defs["forager_post"]
	post.WorkerCapacity = 1
	few.Buildings.defs["forager_post"] = post
	few.Workers.UnlockType("worker")
	few.Workers.Recruit("worker", 2*oldCap, 1<<20)
	few.Workers.Assign("worker", "gathering_camp", 2*oldCap)
	pop := few.Workers.TotalPop()
	if err := few.UpgradeBuilding("gathering_camp", 1, false); err != nil {
		t.Fatal(err)
	}
	if got := few.Workers.GetAssignedCount("worker", "forager_post"); got != 1 {
		t.Errorf("the upgraded copy has %d workers in 1 slot, want 1", got)
	}
	if few.Workers.TotalPop() != pop {
		t.Errorf("population %d -> %d: an upgrade must not lose workers", pop, few.Workers.TotalPop())
	}
	if idle := few.Workers.IdleCount("worker"); idle != oldCap-1 {
		t.Errorf("idle = %d, want the %d the new copy had no slot for", idle, oldCap-1)
	}
}

// Old buildings may no longer be built new; the refusal names the target.
func TestBuildOlderBuilding_NamesTheUpgrade(t *testing.T) {
	ge := campsInTheStoneAge(t, 2)
	err := ge.BuildBuilding("gathering_camp")
	if err == nil || !strings.Contains(err.Error(), "belongs to a previous age") || !strings.Contains(err.Error(), "upgrade gathering_camp") || !strings.Contains(err.Error(), "Forager Post") {
		t.Errorf("build an older building: err = %v, want the previous-age refusal naming the upgrade and its target", err)
	}
	ge.Buildings.counts["gathering_camp"] = 0
	if err := ge.BuildBuilding("gathering_camp"); err == nil || strings.Contains(err.Error(), "upgrade") {
		t.Errorf("with no copies held the refusal still sends the player to upgrade: %v", err)
	}
}

// Older is read from the line and the age every time: it follows the age, it
// is not saved, and a loaded game reads it the same.
func TestOlderIsReadLive(t *testing.T) {
	isolateAccountDir(t)
	ge := olderBuildingsGame(t)
	ge.age = "stone_age"
	if ge.Buildings.IsOlder("gathering_camp") != true {
		t.Error("gathering_camp is not older in the Stone Age")
	}
	ge.age = "primitive_age"
	if ge.Buildings.IsOlder("gathering_camp") {
		t.Error("gathering_camp is older in the Primitive Age")
	}
	ge.age = "bronze_age"
	before := append([]string(nil), ge.Buildings.OlderKeys()...)
	if err := ge.SaveGame("older-live"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("older-live"); err != nil {
		t.Fatal(err)
	}
	after := loaded.Buildings.OlderKeys()
	if strings.Join(before, ",") != strings.Join(after, ",") || len(after) == 0 {
		t.Errorf("older buildings after a load: %v, before %v", after, before)
	}
	if err := loaded.UpgradeBuilding("gathering_camp", 1, false); err != nil && strings.Contains(err.Error(), "has no upgrade") {
		t.Errorf("a loaded game cannot upgrade: %v", err)
	}
	// The save carries no flags and no offers.
	if data := loaded.buildSaveSnapshot(); len(data.LegacyBuildings) != 0 || len(data.PendingUpgrades) != 0 {
		t.Errorf("the save holds legacy %v and offers %v", data.LegacyBuildings, data.PendingUpgrades)
	}
}

// Counters that name a building by key keep their progress when copies are
// upgraded away: the upgraded copy is the building it became.
func TestUpgrade_MilestoneKeepsItsProgress(t *testing.T) {
	ge := newSeededEngine(1)
	ge.age = "stone_age"
	ge.Buildings.counts["wood_camp"] = 25 // Lumber Operation asks for 25
	richTown(ge)
	var lumber config.MilestoneDef
	for _, m := range config.Milestones() {
		if m.Key == "lumber_operation" {
			lumber = m
		}
	}
	if err := ge.UpgradeBuilding("wood_camp", 10, false); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.GetCount("wood_camp") != 15 {
		t.Fatalf("%d wood camps left, want 15", ge.Buildings.GetCount("wood_camp"))
	}
	// Progress reads 25 of 25.
	for _, p := range ge.Milestones.computeProgress(lumber, MilestoneSnapshotParams{Age: "stone_age", AgeOrder: ge.progress.GetAgeOrder(), Buildings: ge.Buildings.GetAll(), held: ge.Buildings}) {
		if p.Label == ge.rules.Name(rules.KindBuilding, "wood_camp") && (p.Current != 25 || !p.Met) {
			t.Errorf("progress shows %v of %v after the upgrade, want 25 of 25", p.Current, p.Target)
		}
	}
	done := ge.Milestones.CheckMilestones(10000, "stone_age", ge.progress.GetAgeOrder(), ge.Resources, ge.Buildings, 0, 0, 100, map[string]bool{}, 0, 0, 0)
	var got bool
	for _, d := range done {
		got = got || d.Key == "lumber_operation"
	}
	if !got {
		t.Error("Lumber Operation did not complete: the upgraded copies stopped counting")
	}

	// The same for a sum across two rungs, each building counted once.
	sum := []string{"trading_post", "merchant_quarter"}
	ge.Buildings.counts["trading_post"] = 3
	ge.Buildings.counts["merchant_quarter"] = 4
	if n := ge.Buildings.CountFromAny(sum); n != 7 {
		t.Errorf("trading posts and merchant quarters count %d, want 7 (no copy counted twice)", n)
	}
	ge.Buildings.counts["guildhall"] = 2 // a later rung of the same line
	if n := ge.Buildings.CountFromAny(sum); n != 9 {
		t.Errorf("with 2 guildhalls the sum is %d, want 9", n)
	}
}

// An upgraded copy counts toward the next age's gate. The Bronze Age gate asks
// for Forager Posts; Gathering Camps upgraded to them are posts. A gate that
// asks for an older rung (Stone Camp) counts the Stone Pits too.
func TestUpgrade_UpgradedCopiesCountForTheGate(t *testing.T) {
	ge := newSeededEngine(1)
	ge.age = "stone_age"
	for key, n := range ge.progress.GetRequirementsForNext("stone_age") {
		ge.Buildings.counts[key] = n
	}
	need := ge.progress.GetRequirementsForNext("stone_age")["forager_post"]
	if need == 0 {
		t.Fatal("setup: the Bronze Age gate does not ask for Forager Posts")
	}
	ge.Buildings.counts["forager_post"] = 0
	ge.Buildings.counts["gathering_camp"] = need
	wonder := ge.progress.WonderForAge("stone_age")
	if wonder != "" {
		ge.Buildings.counts[wonder] = 1
	}
	richTown(ge)
	if got := ge.progress.CheckAdvancement("stone_age", ge.Buildings); got != "" {
		t.Fatalf("the gate is open with no Forager Posts: %q", got)
	}
	if err := ge.UpgradeBuilding("gathering_camp", 0, true); err != nil {
		t.Fatal(err)
	}
	if got := ge.progress.CheckAdvancement("stone_age", ge.Buildings); got != "bronze_age" {
		t.Errorf("the gate after upgrading %d camps to posts: %q, want bronze_age", need, got)
	}

	// A gate that names an older rung counts the later ones.
	stoneCamps := ge.progress.GetRequirementsForNext("stone_age")["stone_camp"]
	if stoneCamps == 0 {
		t.Fatal("setup: the Bronze Age gate does not ask for Stone Camps")
	}
	ge.Buildings.counts["stone_camp"] = 0
	ge.Buildings.counts["stone_pit"] += stoneCamps
	if got := ge.progress.CheckAdvancement("stone_age", ge.Buildings); got != "bronze_age" {
		t.Errorf("the gate with Stone Pits for Stone Camps: %q, want bronze_age", got)
	}
	// And the state shows it.
	if bs := ge.GetState().Buildings["stone_camp"]; bs.HeldWithLater < stoneCamps {
		t.Errorf("the state shows %d held with later rungs, want %d", bs.HeldWithLater, stoneCamps)
	}
}

// A trade route that needs a building keeps running on its upgraded copies.
func TestUpgrade_TradeRouteCountsLaterRungs(t *testing.T) {
	ge := newSeededEngine(1)
	ge.age = "classical_age"
	ge.Buildings.counts["trading_post"] = 3
	if got := ge.Buildings.CountFrom("market"); got != 3 {
		t.Errorf("markets and every later rung of their line count %d, want the 3 trading posts", got)
	}
}
