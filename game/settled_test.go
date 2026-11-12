package game

import (
	"slices"
	"testing"
)

// The places the player settled on the map survive a save and a load, reach
// the snapshot the map reads, and a reset forgets them.
func TestSettledPlacesRoundTripThroughTheSave(t *testing.T) {
	isolateAccountDir(t)
	ge := NewGameEngine()
	if ge.SettlePlaces(nil) {
		t.Fatal("nothing settled reported a change")
	}
	if !ge.SettlePlaces([]string{"andromeda", "~Ithtaris 1873067"}) || ge.SettlePlaces([]string{"andromeda"}) {
		t.Fatal("SettlePlaces should report only a new place")
	}
	if err := ge.SaveGame("settled_round_trip"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("settled_round_trip"); err != nil {
		t.Fatal(err)
	}
	s := loaded.GetState().Settled
	if len(s) != 2 || !s["andromeda"] || !s["~Ithtaris 1873067"] {
		t.Fatalf("loaded settled set %v", s)
	}
	if got := loaded.settledForSave(); !slices.IsSorted(got) {
		t.Errorf("saved ids %v are not in a fixed order", got)
	}
}
