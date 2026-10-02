package game

import "testing"

// TestStorageShrink_ClampsAmounts: losing a storage building lowered the caps
// but left stock above them; a resource with no production never passed
// through Add again, so it stayed over its cap (the fuzzer's "sell
// renaissance_vault": electricity 69.1B over a 69.1B cap). Storage can no
// longer be sold or destroyed (storage_permanent_test.go), but the clamp in
// recalculateRates still guards any cap that drops, so it is checked by
// removing the copies directly.
func TestStorageShrink_ClampsAmounts(t *testing.T) {
	ge := newSeededEngine(1)
	ge.age = "stone_age"
	ge.Buildings.counts["storage_pit"] = 4
	ge.recalculateRates()
	full := ge.Resources.GetStorage("wood")
	ge.Resources.Add("wood", full)
	if n := ge.Buildings.RemoveBuilding("storage_pit", 2); n != 2 {
		t.Fatalf("removed %d storage pits, want 2", n)
	}
	ge.recalculateRates()
	if got, cap := ge.Resources.Get("wood"), ge.Resources.GetStorage("wood"); got > cap || cap >= full {
		t.Errorf("after losing storage: wood %v, cap %v (was %v)", got, cap, full)
	}
}
