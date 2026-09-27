package game

import "testing"

// TestUpgradeBuilding_PartialUpgradeRehomesWorkers is the smoke fuzzer's
// "upgrade wood_camp 7" (seed 2): a partial upgrade left every worker on the
// old building, 138 workers in 39 copies x 3 slots = 117. Workers over the
// old building's remaining capacity follow the upgraded copies, and go idle
// if those are full.
func TestUpgradeBuilding_PartialUpgradeRehomesWorkers(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["gathering_camp"] = 10
	ge.advanceAge("stone_age")
	ge.pendingCatastrophe = ""
	newKey, ok := ge.Buildings.GetPendingUpgrade("gathering_camp")
	if !ok {
		t.Fatal("setup: no pending upgrade for gathering_camp")
	}
	oldCap := ge.Buildings.defs["gathering_camp"].WorkerCapacity
	newCap := ge.Buildings.defs[newKey].WorkerCapacity
	ge.Workers.UnlockType("worker")
	ge.Workers.Recruit("worker", 10*oldCap, 1<<20)
	if !ge.Workers.Assign("worker", "gathering_camp", 10*oldCap) {
		t.Fatal("setup: assign failed")
	}
	for key := range ge.Resources.resources {
		ge.Resources.AddStorage(key, 1e9)
		ge.Resources.Add(key, 1e9)
	}
	pop := ge.Workers.TotalPop()

	if err := ge.UpgradeBuilding("gathering_camp", 3, false); err != nil {
		t.Fatal(err)
	}
	oldAssigned := ge.Workers.GetAssignedCount("worker", "gathering_camp")
	newAssigned := ge.Workers.GetAssignedCount("worker", newKey)
	if max := oldCap * ge.Buildings.GetCount("gathering_camp"); oldAssigned > max {
		t.Errorf("gathering_camp: %d workers in %d slots", oldAssigned, max)
	}
	if max := newCap * ge.Buildings.GetCount(newKey); newAssigned > max {
		t.Errorf("%s: %d workers in %d slots", newKey, newAssigned, max)
	}
	if want := min(3*oldCap, newCap*3); newAssigned != want {
		t.Errorf("%s: %d workers followed the upgraded copies, want %d", newKey, newAssigned, want)
	}
	if got := ge.Workers.TotalPop(); got != pop {
		t.Errorf("population %d -> %d; an upgrade must not create or destroy workers", pop, got)
	}
	if idle := ge.Workers.IdleCount("worker"); idle != pop-oldAssigned-newAssigned {
		t.Errorf("idle = %d, want %d", idle, pop-oldAssigned-newAssigned)
	}
}
