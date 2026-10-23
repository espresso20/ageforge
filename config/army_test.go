package config

import "testing"

func TestArmyExperienceFor(t *testing.T) {
	for _, c := range []struct {
		kind    string
		success bool
		want    float64
	}{
		{MissionScout, true, 1}, {MissionExpedition, true, 3}, {MissionCampaign, true, 8},
		{MissionScout, false, -.5}, {MissionExpedition, false, -1.5}, {MissionCampaign, false, -4},
		{"no_such_kind", true, 0}, {"no_such_kind", false, 0},
	} {
		if got := ArmyExperienceFor(c.kind, c.success); got != c.want {
			t.Errorf("%s success=%v: %v, want %v", c.kind, c.success, got, c.want)
		}
	}
}

func TestArmyRanks(t *testing.T) {
	for _, c := range []struct {
		xp   float64
		rank int
	}{{0, 0}, {9.5, 0}, {10, 1}, {29, 1}, {30, 2}, {79.5, 2}, {80, 3}, {199, 3}, {200, 4}, {5000, 4}} {
		if got := ArmyRankOf(c.xp); got != c.rank {
			t.Errorf("%v experience is rank %d, want %d", c.xp, got, c.rank)
		}
	}
	if ArmyRankCount != 5 {
		t.Errorf("%d ranks, want 5", ArmyRankCount)
	}
	for r, want := range []string{"Green", "Seasoned", "Veteran", "Elite", "Legendary"} {
		key := ArmyRankKey(r)
		if got := Term("bronze_age", key).One; got != want {
			t.Errorf("rank %d (%s) reads %q, want %q", r, key, got, want)
		}
	}
	// What a rank gives: two points of odds, 5% of haul, 25% of Legendary
	// offers for each rank above Green.
	for r := 0; r < ArmyRankCount; r++ {
		f := float64(r)
		if got := ArmyOddsBonus(r); got != .02*f {
			t.Errorf("rank %d odds %v", r, got)
		}
		if got := ArmyHaulFactor(r); got != 1+.05*f {
			t.Errorf("rank %d haul %v", r, got)
		}
		if got := ArmyRareFactor(r); got != 1+.25*f {
			t.Errorf("rank %d rare %v", r, got)
		}
	}
}

// TestMissionTiersOpen: the Stone Era offers short missions, the Iron Era
// adds long ones, the Steel Era and later grand ones; a Veteran army or better
// sees the next tier an era early.
func TestMissionTiersOpen(t *testing.T) {
	for _, e := range Epochs() {
		want := min(3, e.Order+1)
		for rank := 0; rank < ArmyRankCount; rank++ {
			w := want
			if rank >= 2 {
				w = min(3, e.Order+2)
			}
			if got := MissionTiersOpen(e.Order, rank); got != w {
				t.Errorf("%s at rank %d: %d tiers, want %d", e.Name, rank, got, w)
			}
		}
	}
}
