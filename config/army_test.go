package config

import "testing"

// TestArmyExperienceFor: a win earns experience, a loss costs less than a win
// earned, and a kind that is no mission moves nothing.
func TestArmyExperienceFor(t *testing.T) {
	for _, kind := range MissionKinds() {
		win, loss := ArmyExperienceFor(kind, true), ArmyExperienceFor(kind, false)
		if win <= 0 || loss >= 0 || -loss >= win {
			t.Errorf("%s: a win is %v and a loss %v; want a gain and a smaller loss", kind, win, loss)
		}
	}
	for _, success := range []bool{true, false} {
		if got := ArmyExperienceFor("no_such_kind", success); got != 0 {
			t.Errorf("no_such_kind success=%v: %v, want 0", success, got)
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
