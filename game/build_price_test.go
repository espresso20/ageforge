package game

import (
	"math"
	"testing"
)

// pricedEngine is an engine in its first age with every resource it knows
// stocked, a build-cost factor in force and a copy of the gathering camp
// already queued.
func pricedEngine(t *testing.T) *GameEngine {
	t.Helper()
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := NewGameEngine()
	for k := range ge.GetState().Resources {
		ge.SetStockForTest(k, 1e6)
	}
	ge.Buildings.SetCostMultiplier(0.8)
	if err := ge.BuildBuilding("gathering_camp"); err != nil {
		t.Fatal(err)
	}
	return ge
}

func TestBuildBatchOfOneIsTheNextCost(t *testing.T) {
	st := pricedEngine(t).GetState()
	checked := 0
	for key, bs := range st.Buildings {
		if bs.Category == "wonder" {
			continue
		}
		got, ok := st.BuildBatch(key, 1)
		if !ok {
			t.Fatalf("%s: the snapshot's rules do not know it", key)
		}
		for res, want := range bs.NextCost {
			if got[res] != want {
				t.Errorf("%s %s: batch of one is %v, NextCost %v", key, res, got[res], want)
			}
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no building was checked")
	}
}

func TestBuildBatchIsWhatBuildMultiplePays(t *testing.T) {
	ge := pricedEngine(t)
	st := ge.GetState()
	want, _ := st.BuildBatch("gathering_camp", 5)
	before := map[string]float64{}
	for k, r := range st.Resources {
		before[k] = r.Amount
	}
	if n, err := ge.BuildMultiple("gathering_camp", 5); err != nil || n != 5 {
		t.Fatalf("built %d, err %v", n, err)
	}
	for k, r := range ge.GetState().Resources {
		if paid := before[k] - r.Amount; math.Abs(paid-want[k]) > 1e-6 {
			t.Errorf("%s: paid %v, batch said %v", k, paid, want[k])
		}
	}
}

func TestMaxBuildableIsWhatMaxBuilds(t *testing.T) {
	ge := pricedEngine(t)
	for k := range ge.GetState().Resources {
		ge.SetStockForTest(k, 0)
	}
	ge.SetStockForTest("food", 700)
	ge.SetStockForTest("wood", 700)
	st := ge.GetState()
	want := st.MaxBuildable("gathering_camp")
	if want < 2 {
		t.Fatalf("staging: %d copies affordable, want at least 2", want)
	}
	got, err := ge.BuildMultiple("gathering_camp", 10000)
	if err != nil || got != want {
		t.Errorf("max built %d (err %v), MaxBuildable said %d", got, err, want)
	}
	if n := ge.GetState().MaxBuildable("gathering_camp"); n != 0 {
		t.Errorf("after max, %d more are still affordable", n)
	}
}

func TestCanGatherByHandEndsAfterTheMedievalAge(t *testing.T) {
	st := pricedEngine(t).GetState()
	if !st.CanGatherByHand() {
		t.Error("the first age cannot gather by hand")
	}
	st.Age = "medieval_age"
	if !st.CanGatherByHand() {
		t.Error("the Medieval Age cannot gather by hand")
	}
	st.Age = "renaissance_age"
	if st.CanGatherByHand() {
		t.Error("the age after the Medieval Age can gather by hand")
	}
}
