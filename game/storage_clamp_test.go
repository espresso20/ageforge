package game

import "testing"

// TestSellStorage_ClampsAmounts: selling a storage building lowered the caps
// but left stock above them; a resource with no production never passed
// through Add again, so it stayed over its cap (the fuzzer's "sell
// renaissance_vault": electricity 69.1B over a 69.1B cap).
func TestSellStorage_ClampsAmounts(t *testing.T) {
	ge := newSeededEngine(1)
	ge.age = "stone_age"
	ge.Buildings.counts["storage_pit"] = 4
	ge.recalculateRates()
	full := ge.Resources.GetStorage("wood")
	ge.Resources.Add("wood", full)
	if err := ge.SellBuilding("storage_pit", 2); err != nil {
		t.Fatal(err)
	}
	if got, cap := ge.Resources.Get("wood"), ge.Resources.GetStorage("wood"); got > cap || cap >= full {
		t.Errorf("after selling storage: wood %v, cap %v (was %v)", got, cap, full)
	}
}
