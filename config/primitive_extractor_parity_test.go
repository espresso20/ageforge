package config

import "testing"

// TestWoodCampMatchesGatheringCampCost guards Ety5GDPw: the two primitive-age
// extractors (gathering_camp = food, wood_camp = wood) must cost the same at every
// count. They share raw inputs (wood:20 @ 1.12) so normalizeCostCurves resolves
// both to the same ~16 wood @ 1.15. A wood camp once cost ~2.9x a gathering camp
// because its raw CostScale (1.30) ballooned the normalized base via the
// (oldScale/newScale)^9 pivot factor; this test re-surfaces any such drift.
func TestWoodCampMatchesGatheringCampCost(t *testing.T) {
	defs := BuildingByKey()
	wood, ok := defs["wood_camp"]
	if !ok {
		t.Fatal("wood_camp not found")
	}
	gather, ok := defs["gathering_camp"]
	if !ok {
		t.Fatal("gathering_camp not found")
	}

	if wood.CostScale != gather.CostScale {
		t.Errorf("CostScale mismatch: wood_camp=%v gathering_camp=%v", wood.CostScale, gather.CostScale)
	}
	if len(wood.BaseCost) != len(gather.BaseCost) {
		t.Fatalf("BaseCost resource count mismatch: wood_camp=%v gathering_camp=%v", wood.BaseCost, gather.BaseCost)
	}
	for res, wv := range wood.BaseCost {
		gv, ok := gather.BaseCost[res]
		if !ok {
			t.Errorf("wood_camp has cost in %q that gathering_camp lacks", res)
			continue
		}
		if wv != gv {
			t.Errorf("BaseCost[%q] mismatch: wood_camp=%v gathering_camp=%v", res, wv, gv)
		}
	}
}
