package game

import (
	"strings"
	"testing"
)

func TestSellRefusal_PrimitiveAgeAndHeldCounts(t *testing.T) {
	ge := newSeededEngine(1)
	if r := ge.GetState().SellRefusal("gathering_camp"); !strings.Contains(r, "cannot sell") {
		t.Errorf("primitive age refusal = %q", r)
	}
	ge.Buildings.counts["gathering_camp"] = 2
	ge.advanceAge("stone_age")
	ge.pendingCatastrophe = ""
	st := ge.GetState()
	if r := st.SellRefusal("gathering_camp"); r != "" {
		t.Errorf("a held building refuses: %q", r)
	}
	if r := st.SellRefusal("forager_post"); !strings.Contains(r, "no ") {
		t.Errorf("none held refusal = %q", r)
	}
}

func TestQueuedRefusal_UniqueBuildingOnly(t *testing.T) {
	ge := newSeededEngine(1)
	ge.advanceAge("classical_age")
	ge.pendingCatastrophe = ""
	richTown(ge)
	if r := ge.GetState().QueuedRefusal("cultural_obelisk"); r != "" {
		t.Fatalf("refusal before building: %q", r)
	}
	noErr(t, ge.BuildBuilding("cultural_obelisk"))
	st := ge.GetState()
	want := st.QueuedRefusal("cultural_obelisk")
	if !strings.Contains(want, "under construction") {
		t.Fatalf("queued refusal = %q", want)
	}
	err := ge.BuildBuilding("cultural_obelisk")
	if err == nil || err.Error() != want {
		t.Errorf("engine says %v, state says %q", err, want)
	}
	if r := st.QueuedRefusal("gathering_camp"); r != "" {
		t.Errorf("a non-unique building refuses: %q", r)
	}
}

func TestUpgradeOffer_OlderBuildingsOnly(t *testing.T) {
	ge := campsInTheStoneAge(t, 3)
	o := ge.UpgradeOffer("gathering_camp")
	if o.ToKey != "forager_post" || o.Count != 3 || o.Refusal != "" || len(o.Cost) == 0 {
		t.Fatalf("offer = %+v", o)
	}
	if r := ge.UpgradeOffer("forager_post").Refusal; !strings.Contains(r, "no upgrade") {
		t.Errorf("a current building: %q", r)
	}
	for res := range o.Cost {
		setAmount(ge, res, 0)
	}
	if o := ge.UpgradeOffer("gathering_camp"); o.Refusal == "" || len(o.Cost) == 0 {
		t.Errorf("an unaffordable upgrade: %+v", o)
	}
	err := ge.UpgradeBuilding("gathering_camp", 1, false)
	if err == nil || err.Error() != ge.UpgradeOffer("gathering_camp").Refusal {
		t.Errorf("engine says %v, offer says %q", err, ge.UpgradeOffer("gathering_camp").Refusal)
	}
}
