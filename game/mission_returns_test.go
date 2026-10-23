package game

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// sendAndResolve sends a short mission and runs it to its end, whatever
// the roll.
func sendAndResolve(t *testing.T, ge *GameEngine) {
	t.Helper()
	if err := ge.LaunchExpedition(keyOffer(ge, "scouting_party")); err != nil {
		t.Fatalf("LaunchExpedition: %v", err)
	}
	ge.mu.Lock()
	active := ge.Military.ActiveByTier(TierShort)
	active.RoadEvents = nil // This helper forces an immediate return for the log test.
	active.TicksLeft = 1
	ge.processExpeditions()
	ge.mu.Unlock()
}

// TestReturnsLogKeepsTheLastTen verifies a mission that comes home is
// logged newest first, with its tier name and place, and only the last
// config.MissionReturnsKept are kept.
func TestReturnsLogKeepsTheLastTen(t *testing.T) {
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")
	sendAndResolve(t, ge)

	st := ge.GetState().Military
	if len(st.Returns) != 1 || st.ReturnsMade != 1 || st.ReturnsSeen != 0 {
		t.Fatalf("after one return: log %d, made %d, seen %d; want 1, 1, 0", len(st.Returns), st.ReturnsMade, st.ReturnsSeen)
	}
	got := st.Returns[0]
	if got.Label == "" || got.Place == "" || got.Title == "" || got.Tier != TierShort {
		t.Errorf("the log entry is missing its words: %+v", got)
	}
	if got.Success && got.Haul == "" {
		t.Errorf("a success with no haul: %+v", got)
	}

	for i := 0; i < config.MissionReturnsKept+3; i++ {
		setResource(ge, "stone", 1000)
		setResource(ge, "wood", 1000)
		sendAndResolve(t, ge)
	}
	st = ge.GetState().Military
	if len(st.Returns) != config.MissionReturnsKept {
		t.Errorf("the log holds %d entries, want %d", len(st.Returns), config.MissionReturnsKept)
	}
	if st.ReturnsMade != config.MissionReturnsKept+4 {
		t.Errorf("made = %d, want %d", st.ReturnsMade, config.MissionReturnsKept+4)
	}

	ge.MarkReturnsSeen()
	if st = ge.GetState().Military; st.ReturnsSeen != st.ReturnsMade {
		t.Errorf("after marking them seen: seen %d, made %d", st.ReturnsSeen, st.ReturnsMade)
	}
}

// TestReturnsLogIsSaved verifies the log, and how many of it was seen,
// survive a save and a load.
func TestReturnsLogIsSaved(t *testing.T) {
	isolateAccountDir(t)
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")
	sendAndResolve(t, ge)
	want := ge.GetState().Military.Returns
	if err := ge.SaveGame("returns_log"); err != nil {
		t.Fatalf("SaveGame: %v", err)
	}

	loaded := NewGameEngine()
	if err := loaded.LoadGame("returns_log"); err != nil {
		t.Fatalf("LoadGame: %v", err)
	}
	st := loaded.GetState().Military
	if len(st.Returns) != 1 || !reflect.DeepEqual(st.Returns[0], want[0]) {
		t.Errorf("loaded log = %+v, want %+v", st.Returns, want)
	}
	if st.ReturnsMade != 1 || st.ReturnsSeen != 0 {
		t.Errorf("loaded counts: made %d, seen %d; want 1, 0", st.ReturnsMade, st.ReturnsSeen)
	}
}

// TestActiveMissionKnowsItsPlaceAndLength verifies a mission under way
// tells the panel where it goes and how long it is in all.
func TestActiveMissionKnowsItsPlaceAndLength(t *testing.T) {
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")
	if err := ge.LaunchExpedition(keyOffer(ge, "scouting_party")); err != nil {
		t.Fatal(err)
	}
	active := ge.GetState().Military.ActiveShort
	if active == nil || active.Place == "" || active.Label == "" {
		t.Fatalf("the mission under way has no place or label: %+v", active)
	}
	if active.TotalTicks < active.TicksLeft || active.TotalTicks <= 0 {
		t.Errorf("total ticks %d with %d left", active.TotalTicks, active.TicksLeft)
	}
}
