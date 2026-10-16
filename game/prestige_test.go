package game

import (
	"math"
	"testing"
)

func TestPrestigeManager_CanPrestige(t *testing.T) {
	pm := NewPrestigeManager()
	ageOrder := map[string]int{}
	for i, a := range ageKeys() {
		ageOrder[a] = i
	}
	for _, age := range []string{"primitive_age", "bronze_age", "classical_age"} {
		if pm.CanPrestige(age, ageOrder) {
			t.Errorf("CanPrestige(%s) = true; prestige opens at the Medieval Age", age)
		}
	}
	for _, age := range []string{"medieval_age", "industrial_age", "modern_age", "transcendent_age"} {
		if !pm.CanPrestige(age, ageOrder) {
			t.Errorf("CanPrestige(%s) = false; prestige is open from the Medieval Age on", age)
		}
	}
}

func TestPrestigeManager_Floor(t *testing.T) {
	pm := NewPrestigeManager()
	order := map[string]int{}
	for i, age := range ageKeys() {
		order[age] = i
	}
	if !pm.CanPrestige("medieval_age", order) {
		t.Fatal("first prestige should open at the minimum age")
	}
	pm.floor = "medieval_age"
	for _, age := range []string{"primitive_age", "medieval_age"} {
		if pm.CanPrestige(age, order) {
			t.Errorf("CanPrestige(%s) = true below floor", age)
		}
	}
	if !pm.CanPrestige("renaissance_age", order) {
		t.Fatal("prestige should open one age past floor")
	}
	if got := pm.Snapshot().Floor; got != "medieval_age" {
		t.Fatalf("snapshot floor = %q", got)
	}
}

func TestPrestigeFloorAtFinalAgeAllowsClearedGame(t *testing.T) {
	order := map[string]int{}
	ages := ageKeys()
	for i, age := range ages {
		order[age] = i
	}
	pm := NewPrestigeManager()
	pm.floor = ages[len(ages)-1]
	if !pm.CanPrestige(PrestigeMinAge, order) {
		t.Fatal("a cleared game should be able to prestige again from its minimum age")
	}
}

func TestPrestigeFloorSurvivesSaveLoad(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := NewGameEngine()
	for _, age := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age"} {
		if err := ge.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if err := ge.SaveGame("floor"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("floor"); err != nil {
		t.Fatal(err)
	}
	if got := loaded.GetState().Prestige.Floor; got != "medieval_age" {
		t.Fatalf("loaded floor = %q", got)
	}
}

// TestPrestigeManager_CalculatePoints: a prestige pays the depth table,
// with no divisor and no milestone, tech or building terms.
func TestPrestigeManager_CalculatePoints(t *testing.T) {
	pm := NewPrestigeManager()
	want := map[string]int{
		"primitive_age": 0, "stone_age": 1, "bronze_age": 2, "iron_age": 3,
		"medieval_age": 9, "renaissance_age": 12, "victorian_age": 39,
		"modern_age": 120, "information_age": 201, "cyberpunk_age": 363,
		"interstellar_age": 1092, "transcendent_age": 3279,
	}
	for age, w := range want {
		if got := pm.CalculatePoints(age); got != w {
			t.Errorf("CalculatePoints(%s) = %d, want %d", age, got, w)
		}
	}
	// No divisor: ten levels later a Modern Age prestige still pays 120.
	for range 10 {
		pm.Prestige(120)
	}
	if got := pm.CalculatePoints("modern_age"); got != 120 {
		t.Errorf("at level %d a Modern Age prestige pays %d, want 120 (no divisor)", pm.GetLevel(), got)
	}
	if got := pm.CalculatePoints("not_an_age"); got != 0 {
		t.Errorf("an unknown age pays %d, want 0", got)
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

	if got, want := pm.ProductionMultiplier(), 1+0.282*math.Sqrt(5); math.Abs(got-want) > 1e-12 {
		t.Errorf("production multiplier = %v, want %v", got, want)
	}
	if p := pm.Snapshot().PassiveBonus; p != 0 {
		t.Errorf("PassiveBonus = %v, want 0", p)
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
