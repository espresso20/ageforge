package game

import "testing"

// TestSessionMark: loading a save records the state the save left (the
// maps' "since your last visit" baseline), the snapshot hands out a copy,
// and a new game has none.
func TestSessionMark(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(1)
	ge.mu.Lock()
	ge.tick = 1000
	ge.Buildings.counts["hut"] = 3
	ge.mu.Unlock()
	if st := ge.GetState(); st.SessionStart != nil {
		t.Fatal("a game that was never loaded has a session mark")
	}
	if err := ge.SaveGame("mark"); err != nil {
		t.Fatal(err)
	}
	b := NewGameEngine()
	if err := b.LoadGame("mark"); err != nil {
		t.Fatal(err)
	}
	st := b.GetState()
	m := st.SessionStart
	if m == nil {
		t.Fatal("no session mark after a load")
	}
	if m.Tick != 1000 || m.Age != "primitive_age" || m.Buildings["hut"] != 3 || m.SavedAt.IsZero() {
		t.Errorf("session mark %+v", m)
	}
	m.Buildings["hut"] = 99
	if b.GetState().SessionStart.Buildings["hut"] != 3 {
		t.Error("the snapshot's session mark shares its map with the engine")
	}
	b.Reset()
	if b.GetState().SessionStart != nil {
		t.Error("Reset kept the session mark")
	}
}
