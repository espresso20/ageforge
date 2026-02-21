package game

import (
	"testing"
)

func TestPrestigeManager_CanPrestige(t *testing.T) {
	pm := NewPrestigeManager()
	ageOrder := map[string]int{
		"primitive_age": 0, "stone_age": 1, "bronze_age": 2,
		"iron_age": 3, "classical_age": 4, "medieval_age": 5,
	}

	if pm.CanPrestige("primitive_age", ageOrder) {
		t.Error("should not be able to prestige in primitive_age")
	}
	if pm.CanPrestige("bronze_age", ageOrder) {
		t.Error("should not be able to prestige in bronze_age")
	}
	if !pm.CanPrestige("medieval_age", ageOrder) {
		t.Error("should be able to prestige in medieval_age")
	}
}

func TestPrestigeManager_CalculatePoints(t *testing.T) {
	pm := NewPrestigeManager()
	ageOrder := map[string]int{
		"primitive_age": 0, "stone_age": 1, "bronze_age": 2,
		"iron_age": 3, "classical_age": 4, "medieval_age": 5,
	}

	pts := pm.CalculatePoints("medieval_age", ageOrder, 0, 0, 0)
	if pts < 1 {
		t.Errorf("medieval prestige points = %v, want >= 1", pts)
	}

	// More milestones/techs/buildings = more points
	pts2 := pm.CalculatePoints("medieval_age", ageOrder, 20, 15, 50)
	if pts2 <= pts {
		t.Errorf("points with bonuses (%v) should exceed base (%v)", pts2, pts)
	}
}

func TestPrestigeManager_DiminishingReturns(t *testing.T) {
	pm := NewPrestigeManager()
	ageOrder := map[string]int{"medieval_age": 5}

	pts1 := pm.CalculatePoints("medieval_age", ageOrder, 10, 10, 50)
	pm.Prestige(pts1)

	pts2 := pm.CalculatePoints("medieval_age", ageOrder, 10, 10, 50)
	if pts2 >= pts1 {
		t.Errorf("second prestige pts (%v) should be < first (%v) due to diminishing returns", pts2, pts1)
	}
}

func TestPrestigeManager_PrestigeGrantsLevel(t *testing.T) {
	pm := NewPrestigeManager()

	if pm.GetLevel() != 0 {
		t.Errorf("initial level = %v, want 0", pm.GetLevel())
	}

	pm.Prestige(5)
	if pm.GetLevel() != 1 {
		t.Errorf("level after prestige = %v, want 1", pm.GetLevel())
	}

	bonuses := pm.GetBonuses()
	if bonuses["production_all"] <= 0 {
		t.Error("production_all bonus should be > 0 after prestige")
	}
	if bonuses["tick_speed"] <= 0 {
		t.Error("tick_speed bonus should be > 0 after prestige")
	}
}

func TestPrestigeManager_SaveLoadRoundTrip(t *testing.T) {
	pm := NewPrestigeManager()
	pm.Prestige(10)
	pm.Prestige(5)

	snap := pm.Snapshot()

	pm2 := NewPrestigeManager()
	pm2.LoadState(snap.Level, snap.TotalEarned, snap.Available, nil)

	if pm2.GetLevel() != 2 {
		t.Errorf("loaded level = %v, want 2", pm2.GetLevel())
	}
}
