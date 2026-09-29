package game

import "testing"

// Upgrading your markets into trading posts (the log offers it on reaching the
// Iron Age) used to shut the exchange, which only looked for "market" or
// "port". Any building of the trade lineage now opens it.
func TestExchange_AnyTradeBuildingOpensMarket(t *testing.T) {
	ge := NewGameEngine()
	ge.Resources.UnlockResource("iron")
	ge.Resources.UnlockResource("gold")
	ge.Resources.AddStorage("iron", 1e6)
	ge.Resources.AddStorage("gold", 1e6)
	ge.Resources.Add("iron", 1000)

	if _, err := ge.Trade.Exchange("iron", "gold", 100, ge.Resources, ge.Buildings, 0); err == nil {
		t.Fatal("exchange worked with no trade building at all")
	}
	ge.Buildings.counts["trading_post"] = 1 // upgraded market; no "market" left
	got, err := ge.Trade.Exchange("iron", "gold", 100, ge.Resources, ge.Buildings, 0)
	if err != nil {
		t.Fatalf("exchange with a trading post: %v", err)
	}
	if got <= 0 {
		t.Errorf("exchange returned %v gold", got)
	}
}

// The Trade panel shows what you sold and what you bought separately; the
// totals survive a save, and an old save without them still verifies.
func TestExchange_TracksSoldAndBought(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := NewGameEngine()
	for _, r := range []string{"iron", "gold"} {
		ge.Resources.UnlockResource(r)
		ge.Resources.AddStorage(r, 1e6)
	}
	ge.Resources.Add("iron", 1000)
	ge.Buildings.counts["market"] = 1
	got, err := ge.Trade.Exchange("iron", "gold", 100, ge.Resources, ge.Buildings, 0)
	if err != nil {
		t.Fatal(err)
	}
	st := ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil)
	if st.TotalSold["iron"] != 100 || st.TotalSold["gold"] != 0 {
		t.Errorf("sold = %v, want 100 iron only", st.TotalSold)
	}
	if st.TotalBought["gold"] != got || st.TotalBought["iron"] != 0 {
		t.Errorf("bought = %v, want %v gold only", st.TotalBought, got)
	}

	if err := ge.SaveGame("sold_bought"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("sold_bought"); err != nil {
		t.Fatal(err)
	}
	if ge2.Trade.totalSold["iron"] != 100 || ge2.Trade.totalBought["gold"] != got {
		t.Errorf("after load: sold %v, bought %v", ge2.Trade.totalSold, ge2.Trade.totalBought)
	}

	// A save written before the split has neither field.
	old := ge.buildSaveSnapshot()
	old.Trade.TotalSold, old.Trade.TotalBought = nil, nil
	old.Signature = signSave(old, saveHMACKey)
	writeRawSave(t, "pre_split", old)
	ge3 := NewGameEngine()
	if err := ge3.LoadGame("pre_split"); err != nil {
		t.Fatal(err)
	}
	if ge3.cheaterBadge {
		t.Error("a save without sold/bought totals failed signature verification")
	}
}
